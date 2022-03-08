package medications

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/monarko/fhirgo/STU3/datatypes"
	"github.com/monarko/fhirgo/STU3/resources"
	"github.com/nuts-foundation/nuts-demo-ehr/domain/fhir"
	"github.com/nuts-foundation/nuts-demo-ehr/domain/types"
)

type Repository interface {
	Create(ctx context.Context, customerID int, medication types.Medication)  (*types.Medication, error)
	All(ctx context.Context, customerID int) ([]types.Medication, error)
}

type fhirRepository struct {
	factory fhir.Factory
}

func NewFHIRRepository(factory fhir.Factory) *fhirRepository {
	return &fhirRepository{
		factory: factory,
	}
}

func (repo *fhirRepository) Create(ctx context.Context, customerID int, medication types.Medication)  (*types.Medication, error) {
	if medication.Id == "" {
		medication.Id = types.ObjectID(uuid.NewString())
	}
	fhirMedication, err := convertToFHIR(medication)
	if err != nil {
		return nil,fmt.Errorf("unable to convert medication to FHIR Medication: %w", err)
	}
	err = repo.factory(fhir.WithTenant(customerID)).CreateOrUpdate(ctx, fhirMedication)
	if err != nil {
		return nil,fmt.Errorf("unable to write Medication to FHIR store: %w", err)
	}
	return &types.Medication{
		
	},nil
}

func (r *fhirRepository) All(ctx context.Context, customerID int) ([]types.Medication, error) {
	var params map[string]string

	fhirMedications := []fhir.Medication{}
	
	err := r.factory(fhir.WithTenant(customerID)).ReadMultiple(ctx, "Medication", params, &fhirMedications)
	if err != nil {
		return nil, err
	}

	medications := make([]types.Medication, 0)
	for _, medication := range fhirMedications {
		medications = append(medications, ConvertToDomain(&medication))
	}

	return medications, nil
}


func convertToFHIR(medication types.Medication) (*fhir.Medication, error) {
		fhirMedication := &fhir.Medication{
			Base: resources.Base{
					ID:           fhir.ToIDPtr(string(medication.Id)),
					ResourceType: "Medication",
			},

			Code: &datatypes.CodeableConcept{
				Coding: []datatypes.Coding{{
					System:  fhir.ToUriPtr("urn:oid:2.16.840.1.113883.2.4.4.10"),
					Code:    fhir.ToCodePtr("29998"),
					Display: fhir.ToStringPtr("INSULINE INSULATARD INJ 100IE/ML FLACON 10M"),
				},
				{	System:   fhir.ToUriPtr("urn:oid:2.16.840.1.113883.2.4.4.1"),
					Code:    fhir.ToCodePtr("111325"),}},
				},
		}
		return fhirMedication, nil
}

func ConvertToDomain(fhirMedication *fhir.Medication) types.Medication {
	//var value string

	source := "Unknown"

	medication := types.Medication{
		Id:        types.ObjectID(fhir.FromIDPtr(fhirMedication.ID)),
		Name:		string(*fhirMedication.Code.Coding[0].Display),
		Source : &source,
	}


	return medication
}
