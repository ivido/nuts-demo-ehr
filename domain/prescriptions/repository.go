package prescriptions

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/monarko/fhirgo/STU3/datatypes"
	"github.com/monarko/fhirgo/STU3/resources"
	"github.com/nuts-foundation/nuts-demo-ehr/domain/fhir"
	"github.com/nuts-foundation/nuts-demo-ehr/domain/fhir/zorginzage"
	"github.com/nuts-foundation/nuts-demo-ehr/domain/types"
	"github.com/sirupsen/logrus"
)

type Repository interface {
	AllByPatient(ctx context.Context, customerID int, patientID string, episodeID *string) ([]types.Prescription, error)
	Create(ctx context.Context, customerID int, patientID string, prescription types.Prescription) (*types.Prescription, error)
}

type fhirRepository struct {
	factory fhir.Factory
}

func NewFHIRRepository(factory fhir.Factory) *fhirRepository {
	return &fhirRepository{
		factory: factory,
	}
}

func (repo *fhirRepository) Create(ctx context.Context, customerID int, patientID string, prescription types.Prescription) (*types.Prescription, error) {
	if prescription.Id == "" {
		prescription.Id = types.ObjectID(uuid.NewString())
	}

	medicationRequest, err := convertToFHIR(prescription)

	if err != nil {
		return nil, fmt.Errorf("unable to convert prescription to FHIR MedicationRequest: %w", err)
	}
	err = repo.factory(fhir.WithTenant(customerID)).CreateOrUpdate(ctx, medicationRequest)
	if err != nil {
		return nil, fmt.Errorf("unable to write MedicationRequest to FHIR store: %w", err)
	}
	return &types.Prescription{
		PatientID: types.ObjectID(patientID),
	}, nil
}

func convertToFHIR(prescription types.Prescription) (*fhir.MedicationRequest, error) {
	/*
	* @TODO: add fields needed
	*
	 */
	dosage_value := datatypes.Decimal(*prescription.Dosage.Quantity)

	dosage_frequency := datatypes.Integer(*prescription.Dosage.Frequency)

	dosage_unit := datatypes.String(*prescription.Dosage.Unit)

	dosage_amount := datatypes.Decimal(*prescription.Dosage.Amount)

	dosage_period := datatypes.Decimal(*prescription.Dosage.Period)

	medicationRequest := &fhir.MedicationRequest{
		Base: resources.Base{
			ID:           fhir.ToIDPtr(string(prescription.Id)),
			ResourceType: "MedicationRequest",
		},
		// Code: &datatypes.CodeableConcept{
		// 	Coding: []datatypes.Coding{{
		// 		System:  &fhir.LoincCodingSystem,
		// 		Code:    fhir.ToCodePtr("8893-0"),
		// 		Display: fhir.ToStringPtr("Heart rate Peripheral artery by Palpation"),
		// 	}},
		// },
		Subject: datatypes.Reference{Reference: fhir.ToStringPtr("Patient/" + string(prescription.PatientID))},
		MedicationReference: datatypes.Reference{
			Reference: fhir.ToStringPtr("Medication/" + string(*prescription.MedicationID)),
			Display:   fhir.ToStringPtr(*prescription.MedicationName),
		},
		Status: "active",
		Intent: "order",
		Category: &datatypes.CodeableConcept{
			Coding: []datatypes.Coding{{
				System:  &fhir.SnomedCodingSystem,
				Code:    fhir.ToCodePtr("16076005"),
				Display: fhir.ToStringPtr("Prescription (procedure)"),
			}},
		},

		DosageInstruction: []datatypes.Dosage{{
			Text: fhir.ToStringPtr(*prescription.Dosage.Instructions),
			Timing: &datatypes.Timing{
				Repeat: &datatypes.Repeat{
					BoundsDuration: &datatypes.Duration{
						Value:  &dosage_value,
						Unit:   fhir.ToStringPtr("day"),
						System: fhir.ToUriPtr("http://unitsofmeasure.org"),
					},
					Frequency:  &dosage_frequency,
					Period:     &dosage_period,
					PeriodUnit: fhir.ToCodePtr("d"),
				},
			},
			DoseQuantity: &datatypes.SimpleQuantity{
				Value:  &dosage_amount,
				Unit:   &dosage_unit,
				System: fhir.ToUriPtr("urn:oid:2.16.840.1.113883.2.4.4.1.900.2"),
				Code:   fhir.ToCodePtr("245"),
			},
		}},
	}
	if prescription.EpisodeID != nil {
		medicationRequest.Context = &datatypes.Reference{Reference: fhir.ToStringPtr("EpisodeOfCare/" + string(*prescription.EpisodeID))}
	}
	return medicationRequest, nil
}

func ConvertToDomain(medicationrequest *fhir.MedicationRequest, patientID string) types.Prescription {
	//var value string

	source := "Unknown"

	prescription := types.Prescription{
		Id:        types.ObjectID(fhir.FromIDPtr(medicationrequest.ID)),
		PatientID: types.ObjectID(patientID),
		Source:    &source,
	}

	if medicationrequest.MedicationReference.Reference != nil {
		id := types.ObjectID(strings.Split(fhir.FromStringPtr(medicationrequest.MedicationReference.Reference), "/")[1])
		prescription.MedicationID = &id
	}

	if medicationrequest.Context != nil {
		id := types.ObjectID(strings.Split(fhir.FromStringPtr(medicationrequest.Context.Reference), "/")[1])
		prescription.EpisodeID = &id
	}

	//	medication.MedicationID =  types.ObjectID(strings.Split(fhir.FromStringPtr(medicationrequest.MedicationReference.Reference), "/")[1])
	// switch {
	// case observation.ValueString != nil:
	// 	value = fhir.FromStringPtr(observation.ValueString)
	// case observation.ValueQuantity != nil:
	// 	value = renderQuantity(observation.ValueQuantity)
	// case observation.Component != nil:
	// 	var values []string
	// 	for _, component := range observation.Component {
	// 		if component.ValueString != nil {
	// 			values = append(values, fhir.FromStringPtr(component.ValueString))
	// 		} else if component.ValueQuantity != nil {
	// 			values = append(values, renderQuantity(component.ValueQuantity))
	// 		}
	// 	}
	// 	value = strings.Join(values, ", ")
	// }

	// if len(observation.Performer) > 0 {
	// 	source = fhir.FromStringPtr(observation.Performer[0].Display)
	// }

	// report := types.Report{
	// 	Type:      fhir.FromStringPtr(observation.Code.Coding[0].Display),
	// 	Id:        types.ObjectID(fhir.FromIDPtr(observation.ID)),
	// 	Source:    source,
	// 	PatientID: types.ObjectID(patientID),
	// 	Value:     value,
	// }

	//  if medicationrequest.Context != nil {
	// 	id := types.ObjectID(strings.Split(fhir.FromStringPtr(medicationrequest.Context.Reference), "/")[1])
	// 	medication.EpisodeID = &id
	//  }

	return prescription
}

func CreateMedication(ctx context.Context, fhirClient fhir.Client) {
	// @TODO: hacky way to put the default medication in the fhir db, will fix this later

	fhirMedication := &fhir.Medication{
		Base: resources.Base{
			ID:           fhir.ToIDPtr("2fba9b2c-7c46-45a2-acb3-8f653fa9e52a"),
			ResourceType: "Medication",
		},
		Code: &datatypes.CodeableConcept{
			Coding: []datatypes.Coding{{
				System:  fhir.ToUriPtr("urn:oid:2.16.840.1.113883.2.4.4.10"),
				Code:    fhir.ToCodePtr("29998"),
				Display: fhir.ToStringPtr("INSULINE INSULATARD INJ 100IE/ML FLACON 10M"),
			},
				{System: fhir.ToUriPtr("urn:oid:2.16.840.1.113883.2.4.4.1"),
					Code: fhir.ToCodePtr("111325")}},
		},
	}
	err := fhirClient.ReadOne(ctx, "Medication/2fba9b2c-7c46-45a2-acb3-8f653fa9e52a", &fhirMedication)
	if err != nil {
		fhirClient.CreateOrUpdate(ctx, fhirMedication)
	}

	fhirMedication = &fhir.Medication{
		Base: resources.Base{
			ID:           fhir.ToIDPtr("c913371c-3a89-4125-a77b-fe50fd9c5b5b"),
			ResourceType: "Medication",
		},
		Code: &datatypes.CodeableConcept{
			Coding: []datatypes.Coding{{
				System:  fhir.ToUriPtr("urn:oid:2.16.840.1.113883.2.4.4.10"),
				Code:    fhir.ToCodePtr("6920"),
				Display: fhir.ToStringPtr("Ibuprofen 400mg tablet"),
			},
			},
			Text: fhir.ToStringPtr("Ibuprofen 400mg tablet"),
		},
	}
	err = fhirClient.ReadOne(ctx, "Medication/c913371c-3a89-4125-a77b-fe50fd9c5b5b", &fhirMedication)
	if err != nil {
		fhirClient.CreateOrUpdate(ctx, fhirMedication)
	}

	fhirMedication = &fhir.Medication{
		Base: resources.Base{
			ID:           fhir.ToIDPtr("c913371c-3a89-4125-a77b-fe50fd9cbbbb"),
			ResourceType: "Medication",
		},
		Code: &datatypes.CodeableConcept{
			Coding: []datatypes.Coding{{
				System:  fhir.ToUriPtr("urn:oid:2.16.840.1.113883.2.4.4.10"),
				Code:    fhir.ToCodePtr("6920"),
				Display: fhir.ToStringPtr("Acenocoumarol 400mg tablet"),
			},
			},
			Text: fhir.ToStringPtr("Acenocoumarol 400mg tablet"),
		},
	}
	err = fhirClient.ReadOne(ctx, "Medication/c913371c-3a89-4125-a77b-fe50fd9cbbbb", &fhirMedication)
	if err != nil {
		fhirClient.CreateOrUpdate(ctx, fhirMedication)
	}
}

func (repo *fhirRepository) AllByPatient(ctx context.Context, customerID int, patientID string, episodeID *string) ([]types.Prescription, error) {
	medicationrequests := []fhir.MedicationRequest{}

	queryMap := map[string]string{
		"subject": fmt.Sprintf("Patient/%s", patientID),
	}

	if episodeID != nil {
		queryMap["context"] = fmt.Sprintf("EpisodeOfCare/%s", *episodeID)
	}

	fhirClient := repo.factory(fhir.WithTenant(customerID))
	if err := fhirClient.ReadMultiple(ctx, "MedicationRequest", queryMap, &medicationrequests); err != nil {
		return nil, err
	}

	// too late to figure out a normal way to add the medication to the fhir server
	CreateMedication(ctx, fhirClient)

	prescriptions := []types.Prescription{}
	episodeCache := map[string]types.Episode{}

	for _, medicationrequest := range medicationrequests {

		ref := fhir.FromStringPtr(medicationrequest.Subject.Reference)

		prescription := ConvertToDomain(&medicationrequest, ref[len("Patient/"):])
		source := "Local"
		prescription.Source = &source
		if prescription.MedicationID != nil {
			medicationID := string(*prescription.MedicationID)

			fhirMedication := &fhir.Medication{}
			err := fhirClient.ReadOne(ctx, "Medication/"+medicationID, &fhirMedication)
			if err != nil {
				// A failure is not fatal for this request
				logrus.StandardLogger().WithError(err).Warn("could not fetch medication for local Prescription")
				continue
			}
			medicationName := fhir.FromStringPtr(fhirMedication.Code.Coding[0].Display)
			prescription.MedicationName = &medicationName
		}

		if prescription.EpisodeID != nil {
			episodeID := string(*prescription.EpisodeID)
			if _, ok := episodeCache[episodeID]; !ok {
				fhirEpisode := &fhir.EpisodeOfCare{}
				err := fhirClient.ReadOne(ctx, "EpisodeOfCare/"+string(*prescription.EpisodeID), &fhirEpisode)
				if err != nil {
					// A failure is not fatal for this request
					logrus.StandardLogger().WithError(err).Warn("could not fetch episode for local prescription")
					continue
				}
				episode := zorginzage.ToEpisode(fhirEpisode)
				episodeCache[episodeID] = *episode
			}
			diagnosis := episodeCache[episodeID].Diagnosis
			prescription.EpisodeName = &diagnosis
		}
		prescriptions = append(prescriptions, prescription)
	}

	return prescriptions, nil
}
