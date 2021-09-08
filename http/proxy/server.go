package proxy

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"

	"github.com/nuts-foundation/go-did/vc"
	"github.com/nuts-foundation/nuts-demo-ehr/domain/transfer"
	"github.com/nuts-foundation/nuts-demo-ehr/http/auth"
	nutsAuthClient "github.com/nuts-foundation/nuts-demo-ehr/nuts/client/auth"
	"github.com/nuts-foundation/nuts-demo-ehr/nuts/client/vcr"
	"github.com/nuts-foundation/nuts-demo-ehr/nuts/registry"

	"github.com/nuts-foundation/nuts-demo-ehr/domain/customers"
	"github.com/nuts-foundation/nuts-node/vcr/credential"
	"github.com/sirupsen/logrus"

	"github.com/labstack/echo/v4"
)

var fhirServerTenant = struct{}{}

type Server struct {
	proxy               *httputil.ReverseProxy
	auth                auth.Service
	path                string
	customerRepository  customers.Repository
	vcRegistry          registry.VerifiableCredentialRegistry
	multiTenancyEnabled bool
}

func NewServer(authService auth.Service, customerRepository customers.Repository, vcRegistry registry.VerifiableCredentialRegistry, targetURL url.URL, path string, multiTenancyEnabled bool) *Server {
	server := &Server{
		path:                path,
		auth:                authService,
		customerRepository:  customerRepository,
		vcRegistry: 		 vcRegistry,
		multiTenancyEnabled: multiTenancyEnabled,
	}

	server.proxy = &httputil.ReverseProxy{
		// Does not support query parameters in targetURL
		Director: func(req *http.Request) {
			requestURL := &url.URL{}
			*requestURL = *req.URL
			requestURL.Scheme = targetURL.Scheme
			requestURL.Host = targetURL.Host
			requestURL.RawPath = "" // Not required?

			if server.multiTenancyEnabled {
				tenant := req.Context().Value(fhirServerTenant).(string) // this shouldn't/can't fail, because the middleware handler should've set it.
				requestURL.Path = targetURL.Path + "/" + tenant + req.URL.Path[len(path):]
			} else {
				requestURL.Path = targetURL.Path + req.URL.Path[len(path):]
			}

			req.URL = requestURL
			req.Host = requestURL.Host

			logrus.Debugf("Rewritten to: %s", req.URL.Path)

			if _, ok := req.Header["User-Agent"]; !ok {
				// explicitly disable User-Agent so it's not set to default value
				req.Header.Set("User-Agent", "")
			}
		},
	}

	return server
}

func (server *Server) AuthMiddleware() echo.MiddlewareFunc {
	config := auth.Config{
		Skipper: server.skipper,
		ErrorF:  errorFunc,
		AccessF: server.verifyAccess,
	}

	return auth.SecurityFilter{Auth: server.auth}.AuthWithConfig(config)
}

func (server *Server) skipper(ctx echo.Context) bool {
	requestURI := ctx.Request().RequestURI
	return !strings.HasPrefix(requestURI, server.path)
}

func errorFunc(ctx echo.Context, err error) error {
	return ctx.JSON(http.StatusUnauthorized, NewOperationOutcome(err, "access denied", CodeSecurity, SeverityError))
}

func (server *Server) Handler(_ echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		c.Logger().Debugf("FHIR Proxy: proxying %s %s", c.Request().Method, c.Request().RequestURI)
		accessToken := c.Get(auth.AccessToken).(nutsAuthClient.TokenIntrospectionResponse)

		if server.multiTenancyEnabled {
			// Enrich request with resource owner's FHIR server tenant, which is the customer's ID
			tenant, err := server.getTenant(*accessToken.Iss)
			if err != nil {
				return c.JSON(http.StatusBadRequest, NewOperationOutcome(err, err.Error(), CodeSecurity, SeverityError))
			}

			c.SetRequest(c.Request().WithContext(context.WithValue(
				c.Request().Context(),
				fhirServerTenant,
				tenant,
			)))
		}

		server.proxy.ServeHTTP(c.Response(), c.Request())

		return nil
	}
}

// verifyAccess checks the access policy rules. The token has already been checked and the introspected token is used.
func (server *Server) verifyAccess(request *http.Request, token *nutsAuthClient.TokenIntrospectionResponse) error {
	// todo: delegate to access_policy.go

	route := parseRoute(request)

	// check purposeOfUse/service according to §6.2 eOverdracht-sender policy
	service := token.Service
	if service == nil {
		return errors.New("access-token doesn't contain 'service' claim")
	}
	if *service != transfer.SenderServiceName {
		return fmt.Errorf("access-token contains incorrect 'service' claim: %s, must be %s", *service, transfer.SenderServiceName)
	}

	// task specific access
	serverTaskPath := server.path+"/Task"
	if route.path() == serverTaskPath {
		// §6.2.1.1 retrieving tasks via a search operation
		// validate GET [base]/Task?code=http://snomed.info/sct|308292007&_lastUpdated=[time of last request]
		if route.operation != "read" {
			return fmt.Errorf("incorrect operation %s on: %s, must be read", route.operation, serverTaskPath)
		}
		// query params(code and _lastUpdated) are optional
		// ok for Task search
		return nil
	}

	// §6.2.1.2 Updating the Task
	// and
	// §6.2.2 other resources that require a credential and a user contract
	// the existence of the user contract is validated by validateWithNutsAuthorizationCredential
	if err := server.validateWithNutsAuthorizationCredential(request.Context(), token, *route); err!= nil {
		fmt.Errorf("access denied for %s on %s: %w", route.operation, route.path(), err)
	}
	return nil
}

func (server *Server) validateWithNutsAuthorizationCredential(ctx context.Context, token *nutsAuthClient.TokenIntrospectionResponse, route fhirRoute) error {
	hasUser := token.Usi != nil
	if token.Vcs != nil {
		for _, credentialID := range *token.Vcs {
			// resolve credential. NutsAuthCredential must be resolved with the untrusted flag
			vc, err := server.vcRegistry.ResolveVerifiableCredential(ctx, credentialID)
			if err != nil {
				return fmt.Errorf("invalid credential: %w", err)
			}

			didVC, err := convertCredential(*vc)
			if err != nil {
				return fmt.Errorf("invalid credential format: %w", err)
			}
			if !validCredentialType(*didVC) {
				continue
			}

			subject := &credential.NutsAuthorizationCredentialSubject{}
			if err := didVC.UnmarshalCredentialSubject(subject); err != nil {
				return fmt.Errorf("invalid content for NutsAuthorizationCredential credentialSubject: %w", err)
			}
			for _, resource := range subject.Resources {
				// path should match
				if route.path() != resource.Path {
					continue
				}

				// usi must be present when resource requires user context
				if resource.UserContext && !hasUser {
					continue
				}

				// operation must match
				for _, operation := range resource.Operations {
					if operation == route.operation {
						// all is ok, no need to continue after a match
						return nil
					}
				}
			}
		}
		return errors.New("no matching NutsAuthorizationCredential found in access-token")
	}

	return errors.New("no NutsAuthorizationCredential in access-token")
}

func (server *Server) getTenant(requesterDID string) (int, error) {
	customer, err := server.customerRepository.FindByDID(requesterDID)
	if err != nil {
		return 0, err
	}
	if customer == nil {
		return 0, errors.New("unknown tenant")
	}
	return customer.Id, nil
}

func convertCredential(verifiableCredential vcr.VerifiableCredential) (*vc.VerifiableCredential, error) {
	data, err := json.Marshal(verifiableCredential)
	if err != nil {
		return nil, err
	}
	didVC := vc.VerifiableCredential{}
	if err = json.Unmarshal(data, &didVC); err != nil {
		return nil, err
	}
	return &didVC, nil
}

func validCredentialType(verifiableCredential vc.VerifiableCredential) bool {
	return verifiableCredential.IsType(*credential.NutsAuthorizationCredentialTypeURI)
}
