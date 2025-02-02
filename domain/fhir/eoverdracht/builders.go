package eoverdracht

import (
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/monarko/fhirgo/STU3/datatypes"
	"github.com/monarko/fhirgo/STU3/resources"
	"github.com/nuts-foundation/nuts-demo-ehr/domain/fhir"
	"github.com/nuts-foundation/nuts-demo-ehr/domain/types"
	"github.com/sirupsen/logrus"
)

type TransferFHIRBuilder interface {
	BuildTask(props fhir.TaskProperties) resources.Task
	BuildAdvanceNotice(createRequest types.CreateTransferRequest, patient *types.Patient) AdvanceNotice
	BuildNursingHandoffComposition(patient *types.Patient, advanceNotice AdvanceNotice) (fhir.Composition, error)
}

type FHIRBuilder struct {
	IDGenerator IDGenerator
}

func NewFHIRBuilder() TransferFHIRBuilder {
	return FHIRBuilder{IDGenerator: UUIDGenerator{}}
}

// BuildTask builds a task from the TaskProperties struct. If no ID is set, a new uuid is generated.
func (b FHIRBuilder) BuildTask(props fhir.TaskProperties) resources.Task {
	var id string
	if props.ID != nil {
		id = *props.ID
	} else {
		id = b.IDGenerator.GenerateID()
	}
	return resources.Task{
		Domain: resources.Domain{
			Base: resources.Base{
				ResourceType: "Task",
				ID:           fhir.ToIDPtr(id),
			},
		},
		Status: fhir.ToCodePtr(props.Status),
		Code:   &SnomedTransferType,
		Requester: &resources.TaskRequester{
			Agent: &datatypes.Reference{
				Identifier: &datatypes.Identifier{
					System: &fhir.NutsCodingSystem,
					Value:  fhir.ToStringPtr(props.RequesterID),
				},
			},
		},
		Owner: &datatypes.Reference{
			Identifier: &datatypes.Identifier{
				System: &fhir.NutsCodingSystem,
				Value:  fhir.ToStringPtr(props.OwnerID),
			}},
		// TODO: patient seems mandatory in the spec, but can only be sent when placed already
		// has patient in care to protect the identity of the patient during the negotiation phase.
		//"for": map[string]string{
		//	"reference": fmt.Sprintf("Patient/%s", domainTask.PatientID),
		//},
		Input:  props.Input,
		Output: props.Output,
	}
}

func (b FHIRBuilder) BuildAdvanceNotice(createRequest types.CreateTransferRequest, patient *types.Patient) AdvanceNotice {
	problems, interventions, careplan := b.buildCarePlan(createRequest.CarePlan)
	administrativeData := b.buildAdministrativeData(createRequest)
	anonymousPatient := b.buildAnonymousPatient(patient)

	an := AdvanceNotice{
		Patient:       anonymousPatient,
		Problems:      problems,
		Interventions: interventions,
	}

	composition := b.buildAdvanceNoticeComposition(anonymousPatient, administrativeData, careplan)
	an.Composition = composition

	return an
}

// buildAnonymousPatient only contains address information so the receiving organisation can
// decide if they can deliver the requested care
func (b FHIRBuilder) buildAnonymousPatient(patient *types.Patient) resources.Patient {
	return resources.Patient{
		Domain: resources.Domain{
			Base: resources.Base{
				ResourceType: "Patient",
				ID:           fhir.ToIDPtr(b.IDGenerator.GenerateID()),
			},
		},
		Address: []datatypes.Address{{PostalCode: fhir.ToStringPtr(patient.Zipcode)}},
	}
}

// buildAdministrativeData constructs the Administrative Data segment of the transfer as defined by the Nictiz:
// https://decor.nictiz.nl/pub/eoverdracht/e-overdracht-html-20210510T093529/tr-2.16.840.1.113883.2.4.3.11.60.30.4.63-2021-01-27T000000.html#_2.16.840.1.113883.2.4.3.11.60.30.22.4.1_20210126000000
func (FHIRBuilder) buildAdministrativeData(request types.CreateTransferRequest) fhir.CompositionSection {
	transferDate := request.TransferDate.Format(time.RFC3339)
	return fhir.CompositionSection{
		BackboneElement: datatypes.BackboneElement{
			Element: datatypes.Element{
				Extension: []datatypes.Extension{{
					URL:           (*datatypes.URI)(fhir.ToStringPtr("http://nictiz.nl/fhir/StructureDefinition/eOverdracht-TransferDate")),
					ValueDateTime: (*datatypes.DateTime)(fhir.ToStringPtr(transferDate)),
				}},
			},
		},
		Title: fhir.ToStringPtr("Administrative data"),
		Code:  AdministrativeDocConcept,
	}

}

func (b FHIRBuilder) buildNursingHandoffComposition(administrativeData, careplan fhir.CompositionSection, patient resources.Patient) fhir.Composition {
	return fhir.Composition{
		Base: resources.Base{
			ResourceType: "Composition",
			ID:           fhir.ToIDPtr(b.IDGenerator.GenerateID()),
		},
		Type: datatypes.CodeableConcept{
			Coding: []datatypes.Coding{{System: &fhir.SnomedCodingSystem, Code: fhir.ToCodePtr("371535009"), Display: fhir.ToStringPtr("verslag van overdracht")}},
		},
		Subject: datatypes.Reference{Reference: fhir.ToStringPtr("Patient/" + fhir.FromIDPtr(patient.ID))},
		Title:   "Nursing handoff",
		Section: []fhir.CompositionSection{administrativeData, careplan},
	}
}

func (b FHIRBuilder) buildAdvanceNoticeComposition(patient resources.Patient, administrativeData, careplan fhir.CompositionSection) fhir.Composition {

	return fhir.Composition{
		Base: resources.Base{
			ResourceType: "Composition",
			ID:           fhir.ToIDPtr(b.IDGenerator.GenerateID()),
		},
		Type: datatypes.CodeableConcept{
			Coding: []datatypes.Coding{{System: &fhir.LoincCodingSystem, Code: fhir.ToCodePtr("57830-2")}},
		},
		Title:   "Advance notice",
		Subject: datatypes.Reference{Reference: fhir.ToStringPtr(fmt.Sprintf("Patient/%s", fhir.FromIDPtr(patient.ID)))},
		Section: []fhir.CompositionSection{administrativeData, careplan},
	}
}

func (b FHIRBuilder) buildCarePlan(carePlan types.CarePlan) (problems []resources.Condition, interventions []fhir.Procedure, section fhir.CompositionSection) {
	for _, cpPatientProblems := range carePlan.PatientProblems {
		newProblem := b.buildConditionFromProblem(cpPatientProblems.Problem)
		problems = append(problems, newProblem)

		for _, i := range cpPatientProblems.Interventions {
			if strings.TrimSpace(i.Comment) == "" {
				continue
			}
			interventions = append(interventions, b.buildProcedureFromIntervention(i, fhir.FromIDPtr(newProblem.ID)))
		}
	}

	// new patientProblems
	patientProblems := fhir.CompositionSection{
		Title: fhir.ToStringPtr("Current patient problems"),
		Code: datatypes.CodeableConcept{
			Coding: []datatypes.Coding{{
				System:  &fhir.SnomedCodingSystem,
				Code:    fhir.ToCodePtr("86644006"),
				Display: fhir.ToStringPtr("Nursing diagnosis"),
			}},
		},
	}

	// Add the problems as a section
	for _, p := range problems {
		patientProblems.Entry = append(patientProblems.Entry, datatypes.Reference{Reference: fhir.ToStringPtr("Condition/" + fhir.FromIDPtr(p.ID))})
	}
	for _, i := range interventions {
		patientProblems.Entry = append(patientProblems.Entry, datatypes.Reference{Reference: fhir.ToStringPtr("Procedure/" + fhir.FromIDPtr(i.ID))})
	}

	// Start with empty care plan
	careplan := fhir.CompositionSection{
		Code: CarePlanConcept,
		Section: []fhir.CompositionSection{
			patientProblems,
		},
	}
	return problems, interventions, careplan
}

func (b FHIRBuilder) buildProcedureFromIntervention(intervention types.Intervention, problemID string) fhir.Procedure {
	return fhir.Procedure{
		Domain: resources.Domain{
			Base: resources.Base{
				ResourceType: "Procedure",
				ID:           fhir.ToIDPtr(b.IDGenerator.GenerateID()),
			},
		},
		ReasonReference: []datatypes.Reference{{Reference: fhir.ToStringPtr("Condition/" + problemID)}},
		Note:            []datatypes.Annotation{{Text: fhir.ToStringPtr(intervention.Comment)}},
	}
}

func (b FHIRBuilder) buildConditionFromProblem(problem types.Problem) resources.Condition {
	return resources.Condition{
		Domain: resources.Domain{
			Base: resources.Base{
				ResourceType: "Condition",
				ID:           fhir.ToIDPtr(b.IDGenerator.GenerateID()),
			},
		},
		Note: []datatypes.Annotation{{Text: fhir.ToStringPtr(problem.Name)}},
	}
}

func (b FHIRBuilder) BuildNursingHandoffComposition(patient *types.Patient, advanceNotice AdvanceNotice) (fhir.Composition, error) {

	careplan, err := FilterCompositionSectionByType(advanceNotice.Composition.Section, CarePlanCode)
	if err != nil {
		logrus.Warn("unable to get CarePlan from composition")
		// Don't fail when the transfer is incomplete to allow increment development.
		//return eoverdracht.Composition{}, err
	}

	administrativeData, err := FilterCompositionSectionByType(advanceNotice.Composition.Section, AdministrativeDocCode)
	if err != nil {
		logrus.Warn("unable to get AdministrativeDocument from composition")
		// Don't fail when the transfer is incomplete to allow increment development.
		//return eoverdracht.Composition{}, err
	}

	fhirPatient := resources.Patient{Domain: resources.Domain{Base: resources.Base{ID: fhir.ToIDPtr(string(patient.ObjectID))}}}

	return b.buildNursingHandoffComposition(administrativeData, careplan, fhirPatient), nil
}

type IDGenerator interface {
	GenerateID() string
}

type UUIDGenerator struct{}

func (UUIDGenerator) GenerateID() string {
	return uuid.NewString()
}
