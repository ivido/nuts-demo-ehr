package zno

import (
	"time"

	"github.com/lestrrat-go/jwx/jwa"
	"github.com/lestrrat-go/jwx/jwt"
	"github.com/lestrrat-go/jwx/jwt/openid"
	"github.com/nuts-foundation/nuts-demo-ehr/domain/types"
	"github.com/nuts-foundation/nuts-demo-ehr/nuts/client/auth"
)

const userClaim = "user"
const patientClaim = "patient"

type jwtPatient struct {
	Bsn       string `json:"bsn"`
	FirstName string `json:"firstName"`
	LastName  string `json:"lastName"`
}

type jwtUser struct {
	FirstName string                      `json:"firstName"`
	LastName  string                      `json:"lastName"`
	Contract  auth.VerifiablePresentation `json:"contract"`
}

type Service interface {
	GetSsoUrl() string
	CreateSsoJwt(types.Patient, auth.VerifiablePresentation) (string, error)
}

type service struct {
	ssoAddress string
	ssoSecret  string
}

func NewService(ssoAddress string, ssoSecret string) Service {
	return &service{ssoAddress: ssoAddress, ssoSecret: ssoSecret}
}

func (service *service) GetSsoUrl() string {
	return service.ssoAddress
}

func (service *service) CreateSsoJwt(patient types.Patient, vp auth.VerifiablePresentation) (string, error) {
	t := openid.New()

	t.Set(jwt.ExpirationKey, time.Now().Add(time.Hour*24*365).Unix())
	t.Set(jwt.IssuedAtKey, time.Now().Unix())

	t.Set(patientClaim, &jwtPatient{
		Bsn:       *patient.Ssn,
		FirstName: patient.FirstName,
		LastName:  patient.Surname,
	})

	t.Set(userClaim, &jwtUser{
		FirstName: "Jane",
		LastName:  "the Doctor",
		// Contract:  vp,
	})

	ts, err := jwt.Sign(t, jwa.HS256, []byte(service.ssoSecret))
	if err != nil {
		return "", err
	}

	return string(ts), nil
}
