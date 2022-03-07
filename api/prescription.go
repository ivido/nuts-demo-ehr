package api

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/nuts-foundation/nuts-demo-ehr/domain/types"
)

// GetReportsParams defines parameters for GetPrescriptions.
type GetPrescriptionsParams struct {
	// The identifier of episode the report must be part of.
	EpisodeID *string `json:"episodeID,omitempty"`
}

func (w Wrapper) GetPrescriptions(ctx echo.Context, patientID string, params GetPrescriptionsParams) error {
	customer := w.getCustomer(ctx)

	// Get the local reports for the patient
	prescriptions, err := w.PrescriptionRepository.AllByPatient(ctx.Request().Context(), customer.Id, patientID, params.EpisodeID)
	if err != nil {
		return err
	}

	// // Get the remote reports for the patient
	// patient, err := w.PatientRepository.FindByID(ctx.Request().Context(), customer.Id, patientID)
	// if err != nil {
	// 	return err
	// }

	// if patient.Ssn != nil {
	// 	remoteReports, err := w.EpisodeService.GetReports(ctx.Request().Context(), *customer.Did, *patient.Ssn)
	// 	if err != nil {
	// 		return err
	// 	}

	// 	medications = append(medications, remoteReports...)
	// }

	return ctx.JSON(http.StatusOK, prescriptions)
}

func (w Wrapper) CreatePrescription(ctx echo.Context, patientID string) error {
	cid, err := w.getCustomerID(ctx)
	if err != nil {
		return err
	}

	prescriptionToCreate := types.Prescription{}
	if err := ctx.Bind(&prescriptionToCreate); err != nil {
		return err
	}

	prescription, err := w.PrescriptionRepository.Create(ctx.Request().Context(), cid, patientID, prescriptionToCreate)
	if err != nil {
		return err
	}

	return ctx.JSON(http.StatusOK, prescription)
}
