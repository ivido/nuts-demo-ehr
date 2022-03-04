package zno

import (
	"time"

	"github.com/lestrrat-go/jwx/jwa"
	"github.com/lestrrat-go/jwx/jwt"
	"github.com/lestrrat-go/jwx/jwt/openid"
)

const user = "user"
const patient = "patient"

type jwtPatient struct {
	Bsn string `json:"bsn"`
}

type jwtUser struct {
	FirstName string `json:"firstName"`
	LastName  string `json:"lastName"`
}

type Service interface {
	GetSsoUrl() string
	CreateSsoJwt(patientBSN string) (string, error)
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

func (service *service) CreateSsoJwt(patientBSN string) (string, error) {
	t := openid.New()

	t.Set(jwt.ExpirationKey, time.Now().Add(time.Hour*24*365).Unix())
	t.Set(jwt.IssuedAtKey, time.Now().Unix())
	t.Set(patient, &jwtPatient{
		Bsn: patientBSN,
	})
	t.Set(user, &jwtUser{
		FirstName: "Jane",
		LastName:  "the Doctor",
	})

	ts, err := jwt.Sign(t, jwa.HS256, []byte(service.ssoSecret))
	if err != nil {
		return "", err
	}

	return string(ts), nil
}
