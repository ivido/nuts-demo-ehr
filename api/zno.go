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

	jwt, err := w.ZnoService.CreateSsoJwt(*patient)
	if err != nil {
		return err
	}

	info := types.ZnoSsoInfo{
		Jwt: jwt,
		Url: w.ZnoService.GetSsoUrl(),
	}

	return ctx.JSON(http.StatusOK, info)
}
