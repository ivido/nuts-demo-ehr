package api

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/nuts-foundation/nuts-demo-ehr/domain/types"
)

type GetZnoSsoInfoParams = types.GetZnoSsoInfoParams

func (w Wrapper) GetZnoSsoInfo(ctx echo.Context, params GetZnoSsoInfoParams) error {
	cid, err := w.getCustomerID(ctx)
	if err != nil {
		return err
	}

	patient, err := w.PatientRepository.FindByID(ctx.Request().Context(), cid, params.PatientID)
	if err != nil {
		return err
	}

	authJwt, err := w.APIAuth.extractJWTFromHeader(ctx)
	if err != nil {
		return err
	}

	sid, _ := authJwt.Get("sid")
	sess := w.APIAuth.GetSessions()[sid.(string)]

	jwt, err := w.ZnoService.CreateSsoJwt(*patient, sess.Credential)
	if err != nil {
		return err
	}

	info := types.ZnoSsoInfo{
		Jwt: jwt,
		Url: w.ZnoService.GetSsoUrl(),
	}

	return ctx.JSON(http.StatusOK, info)
}
