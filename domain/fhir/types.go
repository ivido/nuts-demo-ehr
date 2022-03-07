package fhir

import (
	"github.com/monarko/fhirgo/STU3/datatypes"
	"github.com/monarko/fhirgo/STU3/resources"
)

// Coding systems
var (
	SnomedCodingSystem datatypes.URI = "http://snomed.info/sct"
	LoincCodingSystem  datatypes.URI = "http://loinc.org"
	NutsCodingSystem   datatypes.URI = "http://nuts.nl"
	UZICodingSystem    datatypes.URI = "http://fhir.nl/fhir/NamingSystem/uzi-nr-pers"
)

// Codes for the status of an EpisodeOfCare
const (
	EpisodeStatusPlanned        = datatypes.Code("planned")
	EpisodeStatusWaitlist       = datatypes.Code("waitlist")
	EpisodeStatusActive         = datatypes.Code("active")
	EpisodeStatusOnHold         = datatypes.Code("onhold")
	EpisodeStatusFinished       = datatypes.Code("finished")
	EpisodeStatusCancelled      = datatypes.Code("cancelled")
	EpisodeStatusEnteredInError = datatypes.Code("entered-in-error")
)

// Codes for the status of an MedicationRequest
const (
	MedicationRequestStatusActive         = datatypes.Code("active")
	MedicationRequestStatusOnHold         = datatypes.Code("on-hold")
	MedicationRequestStatusCancelled      = datatypes.Code("cancelled")
	MedicationRequestStatusEnteredInError = datatypes.Code("entered-in-error")
	MedicationRequestStatusCompleted 	  = datatypes.Code("completed")
	MedicationRequestStatusStopped        = datatypes.Code("stopped")
	MedicationRequestStatusDraft          = datatypes.Code("draft")
	MedicationRequestStatusUnknown        = datatypes.Code("unknown")
)

// Codes for the intent of an MedicationRequest
const (
	MedicationRequestIntentProposal      = datatypes.Code("proposal")
	MedicationRequestIntentPlan        	 = datatypes.Code("plan")
	MedicationRequestIntentOrder         = datatypes.Code("order")
	MedicationRequestIntentInstanceOrder = datatypes.Code("instance-order")
)

// Codes for the priority of an MedicationRequest
const (
	MedicationRequestPriorityRoutine     = datatypes.Code("routine")
	MedicationRequestPriorityUrgent      = datatypes.Code("urgent")
	MedicationRequestPriorityStat        = datatypes.Code("stat")
	MedicationRequestPriorityAsap        = datatypes.Code("asap")
)

type TaskProperties struct {
	ID        *string
	Status    string
	PatientID string
	// nuts DID of the placer
	RequesterID string
	// nuts DID of the filler
	OwnerID string
	Input   []resources.TaskInputOutput
	Output  []resources.TaskInputOutput
}

// Procedure defines a basic FHIR STU3 Procedure resource which is currently not included in the FHIR library.
type Procedure struct {
	resources.Domain
	Identifier      []datatypes.Identifier `json:"identifier,omitempty"`
	Code            datatypes.Code         `json:"code,omitempty"`
	Subject         datatypes.Reference    `json:"subject,omitempty"`
	ReasonReference []datatypes.Reference  `json:"reasonReference,omitempty"`
	Note            []datatypes.Annotation `json:"note,omitempty"`
}

type CompositionSection struct {
	datatypes.BackboneElement
	Code    datatypes.CodeableConcept `json:"code"`
	Title   *datatypes.String         `json:"title,omitempty"`
	Section []CompositionSection      `json:"section,omitempty"`
	Entry   []datatypes.Reference     `json:"entry,omitempty"`
}

// Composition defines a basic FHIR STU3 Composition resource which is currently not included in the FHIR library.
type Composition struct {
	resources.Base
	Identifier []datatypes.Identifier    `json:"identifier,omitempty"`
	Type       datatypes.CodeableConcept `json:"type"`
	Status     datatypes.Code            `json:"status,omitempty"`
	Subject    datatypes.Reference       `json:"subject"`
	Date       datatypes.DateTime        `json:"date,omitempty"`
	Author     []datatypes.Reference     `json:"author,omitempty"`
	Title      datatypes.String          `json:"title,omitempty"`
	Section    []CompositionSection      `json:"section,omitempty"`
}

// EpisodeOfCare defines a basic FHIR STU3 EpisodeOfCare resource which is currently not included in the FHIR library.
type EpisodeOfCare struct {
	resources.Base
	Identifier           []datatypes.Identifier      `json:"identifier,omitempty"`
	Status               datatypes.Code              `json:"status"`
	Type                 []datatypes.CodeableConcept `json:"type,omitempty"`
	Patient              datatypes.Reference         `json:"patient"`
	ManagingOrganization *datatypes.Reference        `json:"managingOrganization,omitempty"`
	Period               *datatypes.Period           `json:"period,omitempty"`
	ReferralRequest      []datatypes.Reference       `json:"referralRequest,omitempty"`
	CareManager          *datatypes.Reference        `json:"careManager,omitempty"`
	Team                 []datatypes.Reference       `json:"team,omitempty"`
	Account              []datatypes.Reference       `json:"account,omitempty"`
}


// MedicationRequest defines a basic FHIR STU3 MedicationRequest resource which is currently not included in the FHIR library.
type MedicationRequest struct {
	resources.Base
	Identifier            []datatypes.Identifier      `json:"identifier,omitempty"`
	Definition			  []datatypes.Reference  	 `json:"definition,omitempty"`
	BasedOn			 	  []datatypes.Reference  	 `json:"basedOn,omitempty"`
	GroupIdentifier       []datatypes.Identifier      `json:"groupIdentifier,omitempty"`
	Status                datatypes.Code              `json:"status"`
	Intent                datatypes.Code              `json:"intent"`
	Category              *datatypes.CodeableConcept `json:"category,omitempty"`
	Priority              datatypes.Code              `json:"priority"`
	MedicationReference   datatypes.Reference         `json:"medicationReference"`
	Subject               datatypes.Reference         `json:"subject"`
	Context               *datatypes.Reference         `json:"context,omitempty"`
	SupportingInformation *datatypes.Reference        `json:"supportingInformation,omitempty"`
	AuthoredOn            datatypes.DateTime          `json:"authoredOn,omitempty"`
	Requester 			  datatypes.BackboneElement   `json:"requester,omitempty"`
	Recorder              datatypes.Reference         `json:"recorder,omitempty"`
	ReasonCode		 	  datatypes.CodeableConcept	  `json:"reasonCode,omitempty"`
	ReasonReference       []datatypes.Reference       `json:"reasonReference,omitempty"`
	Note				  datatypes.Annotation		  `json:"note,omitempty"`
	DosageInstruction	  datatypes.Dosage			  `json:"dosageInstruction,omitempty"`		
	DispenseRequest		  datatypes.BackboneElement	  `json:"dispenseRequest,omitempty"`
	Substitution		  datatypes.BackboneElement	  `json:"substitution,omitempty"`
	PriorPrescription     datatypes.Reference         `json:"priorPrescription,omitempty"`
	/*
	detectedIssue
	eventHistory
	*/
}


// Medication defines a basic FHIR STU3 Medication resource which is currently not included in the FHIR library.
type Medication struct {
	resources.Base
	Identifier            []datatypes.Identifier      `json:"identifier,omitempty"`
	Code                 *datatypes.CodeableConcept          `json:"code,omitempty"`
}

