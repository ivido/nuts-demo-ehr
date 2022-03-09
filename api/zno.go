package api

import (
	"errors"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/nuts-foundation/nuts-demo-ehr/domain/types"
)

type GetZnoSsoInfoParams = types.GetZnoSsoInfoParams

func (w Wrapper) GetZnoSsoInfo(ctx echo.Context, params GetZnoSsoInfoParams) error {
	c := w.getCustomer(ctx)
	if c == nil {
		return errors.New("Customer not found")
	}

	patient, err := w.PatientRepository.FindByID(ctx.Request().Context(), c.Id, params.PatientID)
	if err != nil {
		return err
	}

	sid := ctx.Get("sid")
	sess := w.APIAuth.GetSessions()[sid.(string)]

	jwt, err := w.ZnoService.CreateSsoJwt(*patient, sess.Credential, *c)
	if err != nil {
		return err
	}

	info := types.ZnoSsoInfo{
		Jwt: jwt,
		Url: w.ZnoService.GetSsoUrl(),
	}

	return ctx.JSON(http.StatusOK, info)
}
