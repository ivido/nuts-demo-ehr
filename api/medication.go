package api

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/nuts-foundation/nuts-demo-ehr/domain/types"
)

func (w Wrapper) GetMedication(ctx echo.Context, medicationID string) error {
	//customer := w.getCustomer(ctx)

	// medications, err := w.MedicationRepository.All(ctx.Request().Context(), customer.Id)
	// if err != nil {
	// 	return err
	// }

	// return ctx.JSON(http.StatusOK, medications)
	return ctx.NoContent(http.StatusOK)
}

func (w Wrapper) CreateMedication(ctx echo.Context) error {
	cid, err := w.getCustomerID(ctx)
	if err != nil {
		return err
	}

	medicationToCreate := types.Medication{}
	if err := ctx.Bind(&medicationToCreate); err != nil {
		return err
	}


	medication, err := w.MedicationRepository.Create(ctx.Request().Context(), cid,  medicationToCreate)
	if err != nil {
		return err
	}

	return ctx.JSON(http.StatusOK, medication)
}


