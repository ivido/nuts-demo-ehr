// Package vcr provides primitives to interact with the openapi HTTP API.
//
// Code generated by github.com/deepmap/oapi-codegen version v1.8.2 DO NOT EDIT.
package vcr

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"

	"github.com/deepmap/oapi-codegen/pkg/runtime"
)

// Defines values for ResolutionResultCurrentStatus.
const (
	ResolutionResultCurrentStatusRevoked ResolutionResultCurrentStatus = "revoked"

	ResolutionResultCurrentStatusTrusted ResolutionResultCurrentStatus = "trusted"

	ResolutionResultCurrentStatusUntrusted ResolutionResultCurrentStatus = "untrusted"
)

// CredentialIssuer defines model for CredentialIssuer.
type CredentialIssuer struct {
	// a credential type
	CredentialType string `json:"credentialType"`

	// the DID of an issuer
	Issuer string `json:"issuer"`
}

// Subject of a Verifiable Credential identifying the holder and expressing claims.
type CredentialSubject map[string]interface{}

// DID according to Nuts specification
type DID string

// A request for issuing a new Verifiable Credential.
type IssueVCRequest struct {
	// Subject of a Verifiable Credential identifying the holder and expressing claims.
	CredentialSubject CredentialSubject `json:"credentialSubject"`

	// rfc3339 time string until when the credential is valid.
	ExpirationDate *string `json:"expirationDate,omitempty"`

	// DID according to Nuts specification.
	Issuer string `json:"issuer"`

	// Embedded proof (optional)
	Proof *[]interface{} `json:"proof,omitempty"`

	// Type definition for the credential.
	Type string `json:"type"`
}

// used search params
type KeyValuePair struct {
	// Fields from VCs to search on. Concept specific keys must be prepended with the concept name and a '.'. Default fields like: issuer, subject, type do not require a prefix since they are a mandatory part of each VC.
	Key   string `json:"key"`
	Value string `json:"value"`
}

// result of a Resolve operation.
type ResolutionResult struct {
	// Only credentials with with "trusted" state are valid. If a revoked credential is also untrusted, revoked will be returned.
	CurrentStatus ResolutionResultCurrentStatus `json:"currentStatus"`

	// A credential according to the W3C and Nuts specs.
	VerifiableCredential VerifiableCredential `json:"verifiableCredential"`
}

// Only credentials with with "trusted" state are valid. If a revoked credential is also untrusted, revoked will be returned.
type ResolutionResultCurrentStatus string

// Credential revocation record
type Revocation struct {
	// date is a rfc3339 formatted datetime.
	Date string `json:"date"`

	// DID according to Nuts specification
	Issuer DID `json:"issuer"`

	// Proof contains the cryptographic proof(s).
	Proof *map[string]interface{} `json:"proof,omitempty"`

	// reason describes why the VC has been revoked
	Reason *string `json:"reason,omitempty"`

	// subject refers to the credential identifier that is revoked
	Subject string `json:"subject"`
}

// Input for a search call. Parameters are entered as key/value pairs. Concept specific query params need to be prepended with the concept name.
type SearchRequest struct {
	// limit number of return values to x, default 10
	Limit *float32 `json:"limit,omitempty"`

	// skips first x results, default 0
	Offset *float32 `json:"offset,omitempty"`

	// key/value pairs
	Params []KeyValuePair `json:"params"`
}

// A credential according to the W3C and Nuts specs.
type VerifiableCredential struct {
	// List of URIs
	Context []string `json:"@context"`

	// Subject of a Verifiable Credential identifying the holder and expressing claims.
	CredentialSubject CredentialSubject `json:"credentialSubject"`

	// rfc3339 time string untill when the credential is valid.
	ExpirationDate *string `json:"expirationDate,omitempty"`

	// credential ID. A Nuts DID followed by a large number.
	Id *string `json:"id,omitempty"`

	// rfc3339 time string when the credential was issued.
	IssuanceDate string `json:"issuanceDate"`

	// DID according to Nuts specification
	Issuer DID `json:"issuer"`

	// one or multiple cryptographic proofs
	Proof map[string]interface{} `json:"proof"`

	// List of type definitions for the credential. Always includes 'VerifiableCredential'
	Type []string `json:"type"`
}

// UntrustIssuerJSONBody defines parameters for UntrustIssuer.
type UntrustIssuerJSONBody CredentialIssuer

// TrustIssuerJSONBody defines parameters for TrustIssuer.
type TrustIssuerJSONBody CredentialIssuer

// CreateJSONBody defines parameters for Create.
type CreateJSONBody IssueVCRequest

// ResolveParams defines parameters for Resolve.
type ResolveParams struct {
	// a rfc3339 time string for resolving a VC at a specific moment in time
	ResolveTime *string `json:"resolveTime,omitempty"`
}

// SearchJSONBody defines parameters for Search.
type SearchJSONBody SearchRequest

// SearchParams defines parameters for Search.
type SearchParams struct {
	// when true, the search also returns untrusted credentials. Default false
	Untrusted *bool `json:"untrusted,omitempty"`
}

// UntrustIssuerJSONRequestBody defines body for UntrustIssuer for application/json ContentType.
type UntrustIssuerJSONRequestBody UntrustIssuerJSONBody

// TrustIssuerJSONRequestBody defines body for TrustIssuer for application/json ContentType.
type TrustIssuerJSONRequestBody TrustIssuerJSONBody

// CreateJSONRequestBody defines body for Create for application/json ContentType.
type CreateJSONRequestBody CreateJSONBody

// SearchJSONRequestBody defines body for Search for application/json ContentType.
type SearchJSONRequestBody SearchJSONBody

// RequestEditorFn  is the function signature for the RequestEditor callback function
type RequestEditorFn func(ctx context.Context, req *http.Request) error

// Doer performs HTTP requests.
//
// The standard http.Client implements this interface.
type HttpRequestDoer interface {
	Do(req *http.Request) (*http.Response, error)
}

// Client which conforms to the OpenAPI3 specification for this service.
type Client struct {
	// The endpoint of the server conforming to this interface, with scheme,
	// https://api.deepmap.com for example. This can contain a path relative
	// to the server, such as https://api.deepmap.com/dev-test, and all the
	// paths in the swagger spec will be appended to the server.
	Server string

	// Doer for performing requests, typically a *http.Client with any
	// customized settings, such as certificate chains.
	Client HttpRequestDoer

	// A list of callbacks for modifying requests which are generated before sending over
	// the network.
	RequestEditors []RequestEditorFn
}

// ClientOption allows setting custom parameters during construction
type ClientOption func(*Client) error

// Creates a new Client, with reasonable defaults
func NewClient(server string, opts ...ClientOption) (*Client, error) {
	// create a client with sane default values
	client := Client{
		Server: server,
	}
	// mutate client and add all optional params
	for _, o := range opts {
		if err := o(&client); err != nil {
			return nil, err
		}
	}
	// ensure the server URL always has a trailing slash
	if !strings.HasSuffix(client.Server, "/") {
		client.Server += "/"
	}
	// create httpClient, if not already present
	if client.Client == nil {
		client.Client = &http.Client{}
	}
	return &client, nil
}

// WithHTTPClient allows overriding the default Doer, which is
// automatically created using http.Client. This is useful for tests.
func WithHTTPClient(doer HttpRequestDoer) ClientOption {
	return func(c *Client) error {
		c.Client = doer
		return nil
	}
}

// WithRequestEditorFn allows setting up a callback function, which will be
// called right before sending the request. This can be used to mutate the request.
func WithRequestEditorFn(fn RequestEditorFn) ClientOption {
	return func(c *Client) error {
		c.RequestEditors = append(c.RequestEditors, fn)
		return nil
	}
}

// The interface specification for the client above.
type ClientInterface interface {
	// UntrustIssuer request with any body
	UntrustIssuerWithBody(ctx context.Context, contentType string, body io.Reader, reqEditors ...RequestEditorFn) (*http.Response, error)

	UntrustIssuer(ctx context.Context, body UntrustIssuerJSONRequestBody, reqEditors ...RequestEditorFn) (*http.Response, error)

	// TrustIssuer request with any body
	TrustIssuerWithBody(ctx context.Context, contentType string, body io.Reader, reqEditors ...RequestEditorFn) (*http.Response, error)

	TrustIssuer(ctx context.Context, body TrustIssuerJSONRequestBody, reqEditors ...RequestEditorFn) (*http.Response, error)

	// Create request with any body
	CreateWithBody(ctx context.Context, contentType string, body io.Reader, reqEditors ...RequestEditorFn) (*http.Response, error)

	Create(ctx context.Context, body CreateJSONRequestBody, reqEditors ...RequestEditorFn) (*http.Response, error)

	// Revoke request
	Revoke(ctx context.Context, id string, reqEditors ...RequestEditorFn) (*http.Response, error)

	// Resolve request
	Resolve(ctx context.Context, id string, params *ResolveParams, reqEditors ...RequestEditorFn) (*http.Response, error)

	// Search request with any body
	SearchWithBody(ctx context.Context, concept string, params *SearchParams, contentType string, body io.Reader, reqEditors ...RequestEditorFn) (*http.Response, error)

	Search(ctx context.Context, concept string, params *SearchParams, body SearchJSONRequestBody, reqEditors ...RequestEditorFn) (*http.Response, error)

	// ListTrusted request
	ListTrusted(ctx context.Context, credentialType string, reqEditors ...RequestEditorFn) (*http.Response, error)

	// ListUntrusted request
	ListUntrusted(ctx context.Context, credentialType string, reqEditors ...RequestEditorFn) (*http.Response, error)
}

func (c *Client) UntrustIssuerWithBody(ctx context.Context, contentType string, body io.Reader, reqEditors ...RequestEditorFn) (*http.Response, error) {
	req, err := NewUntrustIssuerRequestWithBody(c.Server, contentType, body)
	if err != nil {
		return nil, err
	}
	req = req.WithContext(ctx)
	if err := c.applyEditors(ctx, req, reqEditors); err != nil {
		return nil, err
	}
	return c.Client.Do(req)
}

func (c *Client) UntrustIssuer(ctx context.Context, body UntrustIssuerJSONRequestBody, reqEditors ...RequestEditorFn) (*http.Response, error) {
	req, err := NewUntrustIssuerRequest(c.Server, body)
	if err != nil {
		return nil, err
	}
	req = req.WithContext(ctx)
	if err := c.applyEditors(ctx, req, reqEditors); err != nil {
		return nil, err
	}
	return c.Client.Do(req)
}

func (c *Client) TrustIssuerWithBody(ctx context.Context, contentType string, body io.Reader, reqEditors ...RequestEditorFn) (*http.Response, error) {
	req, err := NewTrustIssuerRequestWithBody(c.Server, contentType, body)
	if err != nil {
		return nil, err
	}
	req = req.WithContext(ctx)
	if err := c.applyEditors(ctx, req, reqEditors); err != nil {
		return nil, err
	}
	return c.Client.Do(req)
}

func (c *Client) TrustIssuer(ctx context.Context, body TrustIssuerJSONRequestBody, reqEditors ...RequestEditorFn) (*http.Response, error) {
	req, err := NewTrustIssuerRequest(c.Server, body)
	if err != nil {
		return nil, err
	}
	req = req.WithContext(ctx)
	if err := c.applyEditors(ctx, req, reqEditors); err != nil {
		return nil, err
	}
	return c.Client.Do(req)
}

func (c *Client) CreateWithBody(ctx context.Context, contentType string, body io.Reader, reqEditors ...RequestEditorFn) (*http.Response, error) {
	req, err := NewCreateRequestWithBody(c.Server, contentType, body)
	if err != nil {
		return nil, err
	}
	req = req.WithContext(ctx)
	if err := c.applyEditors(ctx, req, reqEditors); err != nil {
		return nil, err
	}
	return c.Client.Do(req)
}

func (c *Client) Create(ctx context.Context, body CreateJSONRequestBody, reqEditors ...RequestEditorFn) (*http.Response, error) {
	req, err := NewCreateRequest(c.Server, body)
	if err != nil {
		return nil, err
	}
	req = req.WithContext(ctx)
	if err := c.applyEditors(ctx, req, reqEditors); err != nil {
		return nil, err
	}
	return c.Client.Do(req)
}

func (c *Client) Revoke(ctx context.Context, id string, reqEditors ...RequestEditorFn) (*http.Response, error) {
	req, err := NewRevokeRequest(c.Server, id)
	if err != nil {
		return nil, err
	}
	req = req.WithContext(ctx)
	if err := c.applyEditors(ctx, req, reqEditors); err != nil {
		return nil, err
	}
	return c.Client.Do(req)
}

func (c *Client) Resolve(ctx context.Context, id string, params *ResolveParams, reqEditors ...RequestEditorFn) (*http.Response, error) {
	req, err := NewResolveRequest(c.Server, id, params)
	if err != nil {
		return nil, err
	}
	req = req.WithContext(ctx)
	if err := c.applyEditors(ctx, req, reqEditors); err != nil {
		return nil, err
	}
	return c.Client.Do(req)
}

func (c *Client) SearchWithBody(ctx context.Context, concept string, params *SearchParams, contentType string, body io.Reader, reqEditors ...RequestEditorFn) (*http.Response, error) {
	req, err := NewSearchRequestWithBody(c.Server, concept, params, contentType, body)
	if err != nil {
		return nil, err
	}
	req = req.WithContext(ctx)
	if err := c.applyEditors(ctx, req, reqEditors); err != nil {
		return nil, err
	}
	return c.Client.Do(req)
}

func (c *Client) Search(ctx context.Context, concept string, params *SearchParams, body SearchJSONRequestBody, reqEditors ...RequestEditorFn) (*http.Response, error) {
	req, err := NewSearchRequest(c.Server, concept, params, body)
	if err != nil {
		return nil, err
	}
	req = req.WithContext(ctx)
	if err := c.applyEditors(ctx, req, reqEditors); err != nil {
		return nil, err
	}
	return c.Client.Do(req)
}

func (c *Client) ListTrusted(ctx context.Context, credentialType string, reqEditors ...RequestEditorFn) (*http.Response, error) {
	req, err := NewListTrustedRequest(c.Server, credentialType)
	if err != nil {
		return nil, err
	}
	req = req.WithContext(ctx)
	if err := c.applyEditors(ctx, req, reqEditors); err != nil {
		return nil, err
	}
	return c.Client.Do(req)
}

func (c *Client) ListUntrusted(ctx context.Context, credentialType string, reqEditors ...RequestEditorFn) (*http.Response, error) {
	req, err := NewListUntrustedRequest(c.Server, credentialType)
	if err != nil {
		return nil, err
	}
	req = req.WithContext(ctx)
	if err := c.applyEditors(ctx, req, reqEditors); err != nil {
		return nil, err
	}
	return c.Client.Do(req)
}

// NewUntrustIssuerRequest calls the generic UntrustIssuer builder with application/json body
func NewUntrustIssuerRequest(server string, body UntrustIssuerJSONRequestBody) (*http.Request, error) {
	var bodyReader io.Reader
	buf, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}
	bodyReader = bytes.NewReader(buf)
	return NewUntrustIssuerRequestWithBody(server, "application/json", bodyReader)
}

// NewUntrustIssuerRequestWithBody generates requests for UntrustIssuer with any type of body
func NewUntrustIssuerRequestWithBody(server string, contentType string, body io.Reader) (*http.Request, error) {
	var err error

	serverURL, err := url.Parse(server)
	if err != nil {
		return nil, err
	}

	operationPath := fmt.Sprintf("/internal/vcr/v1/trust")
	if operationPath[0] == '/' {
		operationPath = "." + operationPath
	}

	queryURL, err := serverURL.Parse(operationPath)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("DELETE", queryURL.String(), body)
	if err != nil {
		return nil, err
	}

	req.Header.Add("Content-Type", contentType)

	return req, nil
}

// NewTrustIssuerRequest calls the generic TrustIssuer builder with application/json body
func NewTrustIssuerRequest(server string, body TrustIssuerJSONRequestBody) (*http.Request, error) {
	var bodyReader io.Reader
	buf, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}
	bodyReader = bytes.NewReader(buf)
	return NewTrustIssuerRequestWithBody(server, "application/json", bodyReader)
}

// NewTrustIssuerRequestWithBody generates requests for TrustIssuer with any type of body
func NewTrustIssuerRequestWithBody(server string, contentType string, body io.Reader) (*http.Request, error) {
	var err error

	serverURL, err := url.Parse(server)
	if err != nil {
		return nil, err
	}

	operationPath := fmt.Sprintf("/internal/vcr/v1/trust")
	if operationPath[0] == '/' {
		operationPath = "." + operationPath
	}

	queryURL, err := serverURL.Parse(operationPath)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", queryURL.String(), body)
	if err != nil {
		return nil, err
	}

	req.Header.Add("Content-Type", contentType)

	return req, nil
}

// NewCreateRequest calls the generic Create builder with application/json body
func NewCreateRequest(server string, body CreateJSONRequestBody) (*http.Request, error) {
	var bodyReader io.Reader
	buf, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}
	bodyReader = bytes.NewReader(buf)
	return NewCreateRequestWithBody(server, "application/json", bodyReader)
}

// NewCreateRequestWithBody generates requests for Create with any type of body
func NewCreateRequestWithBody(server string, contentType string, body io.Reader) (*http.Request, error) {
	var err error

	serverURL, err := url.Parse(server)
	if err != nil {
		return nil, err
	}

	operationPath := fmt.Sprintf("/internal/vcr/v1/vc")
	if operationPath[0] == '/' {
		operationPath = "." + operationPath
	}

	queryURL, err := serverURL.Parse(operationPath)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", queryURL.String(), body)
	if err != nil {
		return nil, err
	}

	req.Header.Add("Content-Type", contentType)

	return req, nil
}

// NewRevokeRequest generates requests for Revoke
func NewRevokeRequest(server string, id string) (*http.Request, error) {
	var err error

	var pathParam0 string

	pathParam0, err = runtime.StyleParamWithLocation("simple", false, "id", runtime.ParamLocationPath, id)
	if err != nil {
		return nil, err
	}

	serverURL, err := url.Parse(server)
	if err != nil {
		return nil, err
	}

	operationPath := fmt.Sprintf("/internal/vcr/v1/vc/%s", pathParam0)
	if operationPath[0] == '/' {
		operationPath = "." + operationPath
	}

	queryURL, err := serverURL.Parse(operationPath)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("DELETE", queryURL.String(), nil)
	if err != nil {
		return nil, err
	}

	return req, nil
}

// NewResolveRequest generates requests for Resolve
func NewResolveRequest(server string, id string, params *ResolveParams) (*http.Request, error) {
	var err error

	var pathParam0 string

	pathParam0, err = runtime.StyleParamWithLocation("simple", false, "id", runtime.ParamLocationPath, id)
	if err != nil {
		return nil, err
	}

	serverURL, err := url.Parse(server)
	if err != nil {
		return nil, err
	}

	operationPath := fmt.Sprintf("/internal/vcr/v1/vc/%s", pathParam0)
	if operationPath[0] == '/' {
		operationPath = "." + operationPath
	}

	queryURL, err := serverURL.Parse(operationPath)
	if err != nil {
		return nil, err
	}

	queryValues := queryURL.Query()

	if params.ResolveTime != nil {

		if queryFrag, err := runtime.StyleParamWithLocation("form", true, "resolveTime", runtime.ParamLocationQuery, *params.ResolveTime); err != nil {
			return nil, err
		} else if parsed, err := url.ParseQuery(queryFrag); err != nil {
			return nil, err
		} else {
			for k, v := range parsed {
				for _, v2 := range v {
					queryValues.Add(k, v2)
				}
			}
		}

	}

	queryURL.RawQuery = queryValues.Encode()

	req, err := http.NewRequest("GET", queryURL.String(), nil)
	if err != nil {
		return nil, err
	}

	return req, nil
}

// NewSearchRequest calls the generic Search builder with application/json body
func NewSearchRequest(server string, concept string, params *SearchParams, body SearchJSONRequestBody) (*http.Request, error) {
	var bodyReader io.Reader
	buf, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}
	bodyReader = bytes.NewReader(buf)
	return NewSearchRequestWithBody(server, concept, params, "application/json", bodyReader)
}

// NewSearchRequestWithBody generates requests for Search with any type of body
func NewSearchRequestWithBody(server string, concept string, params *SearchParams, contentType string, body io.Reader) (*http.Request, error) {
	var err error

	var pathParam0 string

	pathParam0, err = runtime.StyleParamWithLocation("simple", false, "concept", runtime.ParamLocationPath, concept)
	if err != nil {
		return nil, err
	}

	serverURL, err := url.Parse(server)
	if err != nil {
		return nil, err
	}

	operationPath := fmt.Sprintf("/internal/vcr/v1/%s", pathParam0)
	if operationPath[0] == '/' {
		operationPath = "." + operationPath
	}

	queryURL, err := serverURL.Parse(operationPath)
	if err != nil {
		return nil, err
	}

	queryValues := queryURL.Query()

	if params.Untrusted != nil {

		if queryFrag, err := runtime.StyleParamWithLocation("form", true, "untrusted", runtime.ParamLocationQuery, *params.Untrusted); err != nil {
			return nil, err
		} else if parsed, err := url.ParseQuery(queryFrag); err != nil {
			return nil, err
		} else {
			for k, v := range parsed {
				for _, v2 := range v {
					queryValues.Add(k, v2)
				}
			}
		}

	}

	queryURL.RawQuery = queryValues.Encode()

	req, err := http.NewRequest("POST", queryURL.String(), body)
	if err != nil {
		return nil, err
	}

	req.Header.Add("Content-Type", contentType)

	return req, nil
}

// NewListTrustedRequest generates requests for ListTrusted
func NewListTrustedRequest(server string, credentialType string) (*http.Request, error) {
	var err error

	var pathParam0 string

	pathParam0, err = runtime.StyleParamWithLocation("simple", false, "credentialType", runtime.ParamLocationPath, credentialType)
	if err != nil {
		return nil, err
	}

	serverURL, err := url.Parse(server)
	if err != nil {
		return nil, err
	}

	operationPath := fmt.Sprintf("/internal/vcr/v1/%s/trusted", pathParam0)
	if operationPath[0] == '/' {
		operationPath = "." + operationPath
	}

	queryURL, err := serverURL.Parse(operationPath)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("GET", queryURL.String(), nil)
	if err != nil {
		return nil, err
	}

	return req, nil
}

// NewListUntrustedRequest generates requests for ListUntrusted
func NewListUntrustedRequest(server string, credentialType string) (*http.Request, error) {
	var err error

	var pathParam0 string

	pathParam0, err = runtime.StyleParamWithLocation("simple", false, "credentialType", runtime.ParamLocationPath, credentialType)
	if err != nil {
		return nil, err
	}

	serverURL, err := url.Parse(server)
	if err != nil {
		return nil, err
	}

	operationPath := fmt.Sprintf("/internal/vcr/v1/%s/untrusted", pathParam0)
	if operationPath[0] == '/' {
		operationPath = "." + operationPath
	}

	queryURL, err := serverURL.Parse(operationPath)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("GET", queryURL.String(), nil)
	if err != nil {
		return nil, err
	}

	return req, nil
}

func (c *Client) applyEditors(ctx context.Context, req *http.Request, additionalEditors []RequestEditorFn) error {
	for _, r := range c.RequestEditors {
		if err := r(ctx, req); err != nil {
			return err
		}
	}
	for _, r := range additionalEditors {
		if err := r(ctx, req); err != nil {
			return err
		}
	}
	return nil
}

// ClientWithResponses builds on ClientInterface to offer response payloads
type ClientWithResponses struct {
	ClientInterface
}

// NewClientWithResponses creates a new ClientWithResponses, which wraps
// Client with return type handling
func NewClientWithResponses(server string, opts ...ClientOption) (*ClientWithResponses, error) {
	client, err := NewClient(server, opts...)
	if err != nil {
		return nil, err
	}
	return &ClientWithResponses{client}, nil
}

// WithBaseURL overrides the baseURL.
func WithBaseURL(baseURL string) ClientOption {
	return func(c *Client) error {
		newBaseURL, err := url.Parse(baseURL)
		if err != nil {
			return err
		}
		c.Server = newBaseURL.String()
		return nil
	}
}

// ClientWithResponsesInterface is the interface specification for the client with responses above.
type ClientWithResponsesInterface interface {
	// UntrustIssuer request with any body
	UntrustIssuerWithBodyWithResponse(ctx context.Context, contentType string, body io.Reader, reqEditors ...RequestEditorFn) (*UntrustIssuerResponse, error)

	UntrustIssuerWithResponse(ctx context.Context, body UntrustIssuerJSONRequestBody, reqEditors ...RequestEditorFn) (*UntrustIssuerResponse, error)

	// TrustIssuer request with any body
	TrustIssuerWithBodyWithResponse(ctx context.Context, contentType string, body io.Reader, reqEditors ...RequestEditorFn) (*TrustIssuerResponse, error)

	TrustIssuerWithResponse(ctx context.Context, body TrustIssuerJSONRequestBody, reqEditors ...RequestEditorFn) (*TrustIssuerResponse, error)

	// Create request with any body
	CreateWithBodyWithResponse(ctx context.Context, contentType string, body io.Reader, reqEditors ...RequestEditorFn) (*CreateResponse, error)

	CreateWithResponse(ctx context.Context, body CreateJSONRequestBody, reqEditors ...RequestEditorFn) (*CreateResponse, error)

	// Revoke request
	RevokeWithResponse(ctx context.Context, id string, reqEditors ...RequestEditorFn) (*RevokeResponse, error)

	// Resolve request
	ResolveWithResponse(ctx context.Context, id string, params *ResolveParams, reqEditors ...RequestEditorFn) (*ResolveResponse, error)

	// Search request with any body
	SearchWithBodyWithResponse(ctx context.Context, concept string, params *SearchParams, contentType string, body io.Reader, reqEditors ...RequestEditorFn) (*SearchResponse, error)

	SearchWithResponse(ctx context.Context, concept string, params *SearchParams, body SearchJSONRequestBody, reqEditors ...RequestEditorFn) (*SearchResponse, error)

	// ListTrusted request
	ListTrustedWithResponse(ctx context.Context, credentialType string, reqEditors ...RequestEditorFn) (*ListTrustedResponse, error)

	// ListUntrusted request
	ListUntrustedWithResponse(ctx context.Context, credentialType string, reqEditors ...RequestEditorFn) (*ListUntrustedResponse, error)
}

type UntrustIssuerResponse struct {
	Body         []byte
	HTTPResponse *http.Response
}

// Status returns HTTPResponse.Status
func (r UntrustIssuerResponse) Status() string {
	if r.HTTPResponse != nil {
		return r.HTTPResponse.Status
	}
	return http.StatusText(0)
}

// StatusCode returns HTTPResponse.StatusCode
func (r UntrustIssuerResponse) StatusCode() int {
	if r.HTTPResponse != nil {
		return r.HTTPResponse.StatusCode
	}
	return 0
}

type TrustIssuerResponse struct {
	Body         []byte
	HTTPResponse *http.Response
}

// Status returns HTTPResponse.Status
func (r TrustIssuerResponse) Status() string {
	if r.HTTPResponse != nil {
		return r.HTTPResponse.Status
	}
	return http.StatusText(0)
}

// StatusCode returns HTTPResponse.StatusCode
func (r TrustIssuerResponse) StatusCode() int {
	if r.HTTPResponse != nil {
		return r.HTTPResponse.StatusCode
	}
	return 0
}

type CreateResponse struct {
	Body         []byte
	HTTPResponse *http.Response
}

// Status returns HTTPResponse.Status
func (r CreateResponse) Status() string {
	if r.HTTPResponse != nil {
		return r.HTTPResponse.Status
	}
	return http.StatusText(0)
}

// StatusCode returns HTTPResponse.StatusCode
func (r CreateResponse) StatusCode() int {
	if r.HTTPResponse != nil {
		return r.HTTPResponse.StatusCode
	}
	return 0
}

type RevokeResponse struct {
	Body         []byte
	HTTPResponse *http.Response
}

// Status returns HTTPResponse.Status
func (r RevokeResponse) Status() string {
	if r.HTTPResponse != nil {
		return r.HTTPResponse.Status
	}
	return http.StatusText(0)
}

// StatusCode returns HTTPResponse.StatusCode
func (r RevokeResponse) StatusCode() int {
	if r.HTTPResponse != nil {
		return r.HTTPResponse.StatusCode
	}
	return 0
}

type ResolveResponse struct {
	Body         []byte
	HTTPResponse *http.Response
}

// Status returns HTTPResponse.Status
func (r ResolveResponse) Status() string {
	if r.HTTPResponse != nil {
		return r.HTTPResponse.Status
	}
	return http.StatusText(0)
}

// StatusCode returns HTTPResponse.StatusCode
func (r ResolveResponse) StatusCode() int {
	if r.HTTPResponse != nil {
		return r.HTTPResponse.StatusCode
	}
	return 0
}

type SearchResponse struct {
	Body         []byte
	HTTPResponse *http.Response
	JSON200      *[]map[string]interface{}
}

// Status returns HTTPResponse.Status
func (r SearchResponse) Status() string {
	if r.HTTPResponse != nil {
		return r.HTTPResponse.Status
	}
	return http.StatusText(0)
}

// StatusCode returns HTTPResponse.StatusCode
func (r SearchResponse) StatusCode() int {
	if r.HTTPResponse != nil {
		return r.HTTPResponse.StatusCode
	}
	return 0
}

type ListTrustedResponse struct {
	Body         []byte
	HTTPResponse *http.Response
	JSON200      *[]DID
}

// Status returns HTTPResponse.Status
func (r ListTrustedResponse) Status() string {
	if r.HTTPResponse != nil {
		return r.HTTPResponse.Status
	}
	return http.StatusText(0)
}

// StatusCode returns HTTPResponse.StatusCode
func (r ListTrustedResponse) StatusCode() int {
	if r.HTTPResponse != nil {
		return r.HTTPResponse.StatusCode
	}
	return 0
}

type ListUntrustedResponse struct {
	Body         []byte
	HTTPResponse *http.Response
	JSON200      *[]DID
}

// Status returns HTTPResponse.Status
func (r ListUntrustedResponse) Status() string {
	if r.HTTPResponse != nil {
		return r.HTTPResponse.Status
	}
	return http.StatusText(0)
}

// StatusCode returns HTTPResponse.StatusCode
func (r ListUntrustedResponse) StatusCode() int {
	if r.HTTPResponse != nil {
		return r.HTTPResponse.StatusCode
	}
	return 0
}

// UntrustIssuerWithBodyWithResponse request with arbitrary body returning *UntrustIssuerResponse
func (c *ClientWithResponses) UntrustIssuerWithBodyWithResponse(ctx context.Context, contentType string, body io.Reader, reqEditors ...RequestEditorFn) (*UntrustIssuerResponse, error) {
	rsp, err := c.UntrustIssuerWithBody(ctx, contentType, body, reqEditors...)
	if err != nil {
		return nil, err
	}
	return ParseUntrustIssuerResponse(rsp)
}

func (c *ClientWithResponses) UntrustIssuerWithResponse(ctx context.Context, body UntrustIssuerJSONRequestBody, reqEditors ...RequestEditorFn) (*UntrustIssuerResponse, error) {
	rsp, err := c.UntrustIssuer(ctx, body, reqEditors...)
	if err != nil {
		return nil, err
	}
	return ParseUntrustIssuerResponse(rsp)
}

// TrustIssuerWithBodyWithResponse request with arbitrary body returning *TrustIssuerResponse
func (c *ClientWithResponses) TrustIssuerWithBodyWithResponse(ctx context.Context, contentType string, body io.Reader, reqEditors ...RequestEditorFn) (*TrustIssuerResponse, error) {
	rsp, err := c.TrustIssuerWithBody(ctx, contentType, body, reqEditors...)
	if err != nil {
		return nil, err
	}
	return ParseTrustIssuerResponse(rsp)
}

func (c *ClientWithResponses) TrustIssuerWithResponse(ctx context.Context, body TrustIssuerJSONRequestBody, reqEditors ...RequestEditorFn) (*TrustIssuerResponse, error) {
	rsp, err := c.TrustIssuer(ctx, body, reqEditors...)
	if err != nil {
		return nil, err
	}
	return ParseTrustIssuerResponse(rsp)
}

// CreateWithBodyWithResponse request with arbitrary body returning *CreateResponse
func (c *ClientWithResponses) CreateWithBodyWithResponse(ctx context.Context, contentType string, body io.Reader, reqEditors ...RequestEditorFn) (*CreateResponse, error) {
	rsp, err := c.CreateWithBody(ctx, contentType, body, reqEditors...)
	if err != nil {
		return nil, err
	}
	return ParseCreateResponse(rsp)
}

func (c *ClientWithResponses) CreateWithResponse(ctx context.Context, body CreateJSONRequestBody, reqEditors ...RequestEditorFn) (*CreateResponse, error) {
	rsp, err := c.Create(ctx, body, reqEditors...)
	if err != nil {
		return nil, err
	}
	return ParseCreateResponse(rsp)
}

// RevokeWithResponse request returning *RevokeResponse
func (c *ClientWithResponses) RevokeWithResponse(ctx context.Context, id string, reqEditors ...RequestEditorFn) (*RevokeResponse, error) {
	rsp, err := c.Revoke(ctx, id, reqEditors...)
	if err != nil {
		return nil, err
	}
	return ParseRevokeResponse(rsp)
}

// ResolveWithResponse request returning *ResolveResponse
func (c *ClientWithResponses) ResolveWithResponse(ctx context.Context, id string, params *ResolveParams, reqEditors ...RequestEditorFn) (*ResolveResponse, error) {
	rsp, err := c.Resolve(ctx, id, params, reqEditors...)
	if err != nil {
		return nil, err
	}
	return ParseResolveResponse(rsp)
}

// SearchWithBodyWithResponse request with arbitrary body returning *SearchResponse
func (c *ClientWithResponses) SearchWithBodyWithResponse(ctx context.Context, concept string, params *SearchParams, contentType string, body io.Reader, reqEditors ...RequestEditorFn) (*SearchResponse, error) {
	rsp, err := c.SearchWithBody(ctx, concept, params, contentType, body, reqEditors...)
	if err != nil {
		return nil, err
	}
	return ParseSearchResponse(rsp)
}

func (c *ClientWithResponses) SearchWithResponse(ctx context.Context, concept string, params *SearchParams, body SearchJSONRequestBody, reqEditors ...RequestEditorFn) (*SearchResponse, error) {
	rsp, err := c.Search(ctx, concept, params, body, reqEditors...)
	if err != nil {
		return nil, err
	}
	return ParseSearchResponse(rsp)
}

// ListTrustedWithResponse request returning *ListTrustedResponse
func (c *ClientWithResponses) ListTrustedWithResponse(ctx context.Context, credentialType string, reqEditors ...RequestEditorFn) (*ListTrustedResponse, error) {
	rsp, err := c.ListTrusted(ctx, credentialType, reqEditors...)
	if err != nil {
		return nil, err
	}
	return ParseListTrustedResponse(rsp)
}

// ListUntrustedWithResponse request returning *ListUntrustedResponse
func (c *ClientWithResponses) ListUntrustedWithResponse(ctx context.Context, credentialType string, reqEditors ...RequestEditorFn) (*ListUntrustedResponse, error) {
	rsp, err := c.ListUntrusted(ctx, credentialType, reqEditors...)
	if err != nil {
		return nil, err
	}
	return ParseListUntrustedResponse(rsp)
}

// ParseUntrustIssuerResponse parses an HTTP response from a UntrustIssuerWithResponse call
func ParseUntrustIssuerResponse(rsp *http.Response) (*UntrustIssuerResponse, error) {
	bodyBytes, err := ioutil.ReadAll(rsp.Body)
	defer rsp.Body.Close()
	if err != nil {
		return nil, err
	}

	response := &UntrustIssuerResponse{
		Body:         bodyBytes,
		HTTPResponse: rsp,
	}

	return response, nil
}

// ParseTrustIssuerResponse parses an HTTP response from a TrustIssuerWithResponse call
func ParseTrustIssuerResponse(rsp *http.Response) (*TrustIssuerResponse, error) {
	bodyBytes, err := ioutil.ReadAll(rsp.Body)
	defer rsp.Body.Close()
	if err != nil {
		return nil, err
	}

	response := &TrustIssuerResponse{
		Body:         bodyBytes,
		HTTPResponse: rsp,
	}

	return response, nil
}

// ParseCreateResponse parses an HTTP response from a CreateWithResponse call
func ParseCreateResponse(rsp *http.Response) (*CreateResponse, error) {
	bodyBytes, err := ioutil.ReadAll(rsp.Body)
	defer rsp.Body.Close()
	if err != nil {
		return nil, err
	}

	response := &CreateResponse{
		Body:         bodyBytes,
		HTTPResponse: rsp,
	}

	return response, nil
}

// ParseRevokeResponse parses an HTTP response from a RevokeWithResponse call
func ParseRevokeResponse(rsp *http.Response) (*RevokeResponse, error) {
	bodyBytes, err := ioutil.ReadAll(rsp.Body)
	defer rsp.Body.Close()
	if err != nil {
		return nil, err
	}

	response := &RevokeResponse{
		Body:         bodyBytes,
		HTTPResponse: rsp,
	}

	return response, nil
}

// ParseResolveResponse parses an HTTP response from a ResolveWithResponse call
func ParseResolveResponse(rsp *http.Response) (*ResolveResponse, error) {
	bodyBytes, err := ioutil.ReadAll(rsp.Body)
	defer rsp.Body.Close()
	if err != nil {
		return nil, err
	}

	response := &ResolveResponse{
		Body:         bodyBytes,
		HTTPResponse: rsp,
	}

	return response, nil
}

// ParseSearchResponse parses an HTTP response from a SearchWithResponse call
func ParseSearchResponse(rsp *http.Response) (*SearchResponse, error) {
	bodyBytes, err := ioutil.ReadAll(rsp.Body)
	defer rsp.Body.Close()
	if err != nil {
		return nil, err
	}

	response := &SearchResponse{
		Body:         bodyBytes,
		HTTPResponse: rsp,
	}

	switch {
	case strings.Contains(rsp.Header.Get("Content-Type"), "json") && rsp.StatusCode == 200:
		var dest []map[string]interface{}
		if err := json.Unmarshal(bodyBytes, &dest); err != nil {
			return nil, err
		}
		response.JSON200 = &dest

	}

	return response, nil
}

// ParseListTrustedResponse parses an HTTP response from a ListTrustedWithResponse call
func ParseListTrustedResponse(rsp *http.Response) (*ListTrustedResponse, error) {
	bodyBytes, err := ioutil.ReadAll(rsp.Body)
	defer rsp.Body.Close()
	if err != nil {
		return nil, err
	}

	response := &ListTrustedResponse{
		Body:         bodyBytes,
		HTTPResponse: rsp,
	}

	switch {
	case strings.Contains(rsp.Header.Get("Content-Type"), "json") && rsp.StatusCode == 200:
		var dest []DID
		if err := json.Unmarshal(bodyBytes, &dest); err != nil {
			return nil, err
		}
		response.JSON200 = &dest

	}

	return response, nil
}

// ParseListUntrustedResponse parses an HTTP response from a ListUntrustedWithResponse call
func ParseListUntrustedResponse(rsp *http.Response) (*ListUntrustedResponse, error) {
	bodyBytes, err := ioutil.ReadAll(rsp.Body)
	defer rsp.Body.Close()
	if err != nil {
		return nil, err
	}

	response := &ListUntrustedResponse{
		Body:         bodyBytes,
		HTTPResponse: rsp,
	}

	switch {
	case strings.Contains(rsp.Header.Get("Content-Type"), "json") && rsp.StatusCode == 200:
		var dest []DID
		if err := json.Unmarshal(bodyBytes, &dest); err != nil {
			return nil, err
		}
		response.JSON200 = &dest

	}

	return response, nil
}
