package api

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/nuts-foundation/nuts-demo-ehr/domain/types"
)

type GetZnoSsoInfoParams = types.GetZnoSsoInfoParams

func (w Wrapper) GetZnoSsoInfo(ctx echo.Context, params GetZnoSsoInfoParams) error {
	jwt, err := w.ZnoService.CreateSsoJwt(params.Bsn)
	if err != nil {
		return err
	}

	info := types.ZnoSsoInfo{
		Jwt: jwt,
		Url: w.ZnoService.GetSsoUrl(),
	}

	return ctx.JSON(http.StatusOK, info)
}
