package receiver

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/monarko/fhirgo/STU3/datatypes"
	"github.com/monarko/fhirgo/STU3/resources"
	"github.com/nuts-foundation/go-did/vc"
	"github.com/nuts-foundation/nuts-demo-ehr/domain"
	"github.com/nuts-foundation/nuts-demo-ehr/domain/customers"
	"github.com/nuts-foundation/nuts-demo-ehr/domain/fhir"
	"github.com/nuts-foundation/nuts-demo-ehr/domain/fhir/eoverdracht"
	"github.com/nuts-foundation/nuts-demo-ehr/domain/transfer"
	"github.com/nuts-foundation/nuts-demo-ehr/http/auth"
	"github.com/nuts-foundation/nuts-demo-ehr/nuts/registry"
)

type TransferService interface {
	CreateOrUpdate(ctx context.Context, status string, customerID int, customerDID, senderDID, fhirTaskID string) error
	UpdateTransferRequestState(ctx context.Context, customerID int, requesterDID, fhirTaskID string, newState string) error
	GetTransferRequest(ctx context.Context, customerID int, requesterDID string, fhirTaskID string) (*domain.TransferRequest, error)
}

type service struct {
	transferRepo           TransferRepository
	notifier               transfer.Notifier
	auth                   auth.Service
	localFHIRClientFactory fhir.Factory // client for interacting with the local FHIR server
	customerRepo           customers.Repository
	registry               registry.OrganizationRegistry
	vcr                    registry.VerifiableCredentialRegistry
}

func NewTransferService(authService auth.Service, localFHIRClientFactory fhir.Factory, transferRepository TransferRepository, customerRepository customers.Repository, organizationRegistry registry.OrganizationRegistry, vcr registry.VerifiableCredentialRegistry) TransferService {
	return &service{
		auth:                   authService,
		localFHIRClientFactory: localFHIRClientFactory,
		transferRepo:           transferRepository,
		customerRepo:           customerRepository,
		registry:               organizationRegistry,
		vcr:                    vcr,
		notifier:               transfer.FireAndForgetNotifier{},
	}
}

func verifyStateUpdate(from, to string) error {
	// Contains all possible status-updates from: https://informatiestandaarden.nictiz.nl/wiki/vpk:V4.0_FHIR_eOverdracht#Using_Task_to_manage_the_workflow
	allowed := []struct {
		from string
		to   string
	}{
		{transfer.RequestedState, transfer.AcceptedState},
		{transfer.RequestedState, transfer.RejectedState},
		{transfer.RequestedState, transfer.OnHoldState},
		{transfer.OnHoldState, transfer.RequestedState},
		{transfer.OnHoldState, transfer.CancelledState},
		{transfer.AcceptedState, transfer.InProgressState},
		{transfer.InProgressState, transfer.CompletedState},
	}

	for _, update := range allowed {
		if from == update.from && to == update.to {
			return nil
		}
	}

	return fmt.Errorf("invalid state change from %s to %s", from, to)
}

func (s service) completeTransfer(ctx context.Context, incomingTransfer *domain.IncomingTransfer, customerDID, senderDID string) error {
	taskPath := fmt.Sprintf("Task/%s", incomingTransfer.FhirTaskID)
	client, err := s.getRemoteFHIRClient(ctx, senderDID, customerDID, taskPath)
	if err != nil {
		return err
	}

	// First, get the task
	task, err := s.getRemoteTransferTask(ctx, client, incomingTransfer.FhirTaskID)
	if err != nil {
		return err
	}

	// Then, get the composition reference
	var compositionRef string

	for _, input := range task.Input {
		if *input.Type.Coding[0].Code == fhir.SnomedNursingHandoffCode {
			compositionRef = fhir.FromStringPtr(input.ValueReference.Reference)
			break
		}
	}

	if compositionRef == "" {
		return errors.New("unable to find nursing handoff input")
	}

	// Read the subject from the composition
	composition := &eoverdracht.Composition{}

	if err := client().ReadOne(ctx, fmt.Sprintf("/%s", compositionRef), composition); err != nil {
		return fmt.Errorf("failed to read composition: %w", err)
	}

	// And at last, retrieve the patient and create it on the local FHIR server
	patient := &resources.Patient{}

	if err := client().ReadOne(ctx, fmt.Sprintf("/%s", fhir.FromStringPtr(composition.Subject.Reference)), patient); err != nil {
		return fmt.Errorf("failed to read the patient: %w", err)
	}

	if err := s.localFHIRClientFactory().CreateOrUpdate(ctx, patient); err != nil {
		return fmt.Errorf("failed to create/update the patient: %w", err)
	}

	return nil
}

func (s service) CreateOrUpdate(ctx context.Context, status string, customerID int, customerDID, senderDID, fhirTaskID string) error {
	incomingTransfer, err := s.transferRepo.FindByFHIRTaskID(ctx, customerID, fhirTaskID)
	switch err {
	case sql.ErrNoRows:
	case nil:
		negotiationStatus := domain.TransferNegotiationStatus{Status: domain.TransferNegotiationStatusStatus(status)}

		// Return early if there is nothing to be updated
		if incomingTransfer.Status == negotiationStatus {
			return nil
		}

		if err := verifyStateUpdate(string(incomingTransfer.Status.Status), status); err != nil {
			return err
		}
	default:
		return err
	}

	incomingTransfer, err = s.transferRepo.CreateOrUpdate(ctx, status, fhirTaskID, customerID, senderDID)
	if err != nil {
		return err
	}

	if status == transfer.CompletedState {
		return s.completeTransfer(ctx, incomingTransfer, customerDID, senderDID)
	}

	return nil
}

func (s service) UpdateTransferRequestState(ctx context.Context, customerID int, requesterDID, fhirTaskID string, newState string) error {
	customer, err := s.customerRepo.FindByID(customerID)
	if err != nil || customer.Did == nil {
		return err
	}

	taskPath := fmt.Sprintf("/Task/%s", fhirTaskID)
	client, err := s.getRemoteFHIRClient(ctx, requesterDID, *customer.Did, taskPath)
	if err != nil {
		return err
	}

	task, err := s.getRemoteTransferTask(ctx, client, fhirTaskID)
	if err != nil {
		return err
	}

	if err := verifyStateUpdate(fhir.FromCodePtr(task.Status), newState); err != nil {
		return err
	}

	task.Status = fhir.ToCodePtr(newState)

	err = client().CreateOrUpdate(ctx, task)
	if err != nil {
		return err
	}

	// update was a success. Get the remote task again and update the local transfer_request
	task, err = s.getRemoteTransferTask(ctx, client, fhirTaskID)
	if err != nil {
		return err
	}

	_, err = s.transferRepo.CreateOrUpdate(ctx, fhir.FromCodePtr(task.Status), fhirTaskID, customerID, requesterDID)
	if err != nil {
		return fmt.Errorf("could update incomming transfers with new state")
	}

	return nil
}

func (s service) taskContainsCode(task resources.Task, code datatypes.Code) bool {
	for _, input := range task.Input {
		if fhir.FromCodePtr(input.Type.Coding[0].Code) == string(code) {
			return true
		}
	}

	return false
}

func (s service) GetTransferRequest(ctx context.Context, customerID int, requesterDID string, fhirTaskID string) (*domain.TransferRequest, error) {
	customer, err := s.customerRepo.FindByID(customerID)
	if err != nil || customer.Did == nil {
		return nil, fmt.Errorf("unable to find customer: %w", err)
	}

	taskPath := fmt.Sprintf("/Task/%s", fhirTaskID)
	fhirClient, err := s.getRemoteFHIRClient(ctx, requesterDID, *customer.Did, taskPath)
	if err != nil {
		return nil, err
	}

	task, err := s.getRemoteTransferTask(ctx, fhirClient, fhirTaskID)
	if err != nil {
		return nil, fmt.Errorf("unable to get remote transfer: %w", err)
	}

	if !s.taskContainsCode(task, fhir.LoincAdvanceNoticeCode) {
		return nil, fmt.Errorf("invalid task, expected an advanceNotice composition")
	}

	advanceNotice, err := s.getAdvanceNotice(ctx, fhirClient(), fhir.FromStringPtr(task.Input[0].ValueReference.Reference))
	if err != nil {
		return nil, fmt.Errorf("unable to get advance notice: %w", err)
	}
	domainAdvanceNotice, err := domain.FHIRAdvanceNoticeToDomainTransfer(advanceNotice)
	if err != nil {
		return nil, err
	}

	organization, err := s.registry.Get(ctx, requesterDID)
	if err != nil {
		return nil, fmt.Errorf("unable to get organization from registry: %w", err)
	}

	// TODO: Do we need nil checks?
	transferRequest := domain.TransferRequest{
		Sender:        *organization,
		AdvanceNotice: domainAdvanceNotice,
		Status:        fhir.FromCodePtr(task.Status),
	}

	// If the task input contains the nursing handoff, add that one too.
	if len(task.Input) == 2 {
		nursingHandoff, err := s.getNursingHandoff(ctx, fhirClient(), fhir.FromStringPtr(task.Input[1].ValueReference.Reference))
		if err != nil {
			return nil, fmt.Errorf("unable to get nursing handoff: %w", err)
		}
		domainTransfer, err := domain.FHIRNursingHandoffToDomainTransfer(nursingHandoff)
		if err != nil {
			return nil, fmt.Errorf("unable to convert fhir nursing handoff to domain transfer: %w", err)
		}
		transferRequest.NursingHandoff = &domainTransfer
	}

	return &transferRequest, nil
}

func (s service) getRemoteFHIRClient(ctx context.Context, custodianDID string, localActorDID string, resource string) (fhir.Factory, error) {
	fhirServer, err := s.registry.GetCompoundServiceEndpoint(ctx, custodianDID, transfer.SenderServiceName, "fhir")
	if err != nil {
		return nil, fmt.Errorf("error while looking up custodian's FHIR server (did=%s): %w", custodianDID, err)
	}
	credentials, err := s.vcr.FindAuthorizationCredentials(ctx, transfer.SenderServiceName, localActorDID, resource)

	var transformed = make([]vc.VerifiableCredential, len(credentials))
	for i, c := range credentials {
		bytes, err := json.Marshal(c)
		if err != nil {
			return nil, err
		}
		tCred := vc.VerifiableCredential{}
		if err = json.Unmarshal(bytes, &tCred); err != nil {
			return nil, err
		}
		transformed[i] = tCred
	}

	accessToken, err := s.auth.RequestAccessToken(ctx, localActorDID, custodianDID, transfer.SenderServiceName, transformed)
	if err != nil {
		return nil, err
	}

	return fhir.NewFactory(fhir.WithURL(fhirServer), fhir.WithAuthToken(accessToken.AccessToken)), nil
}

func (s service) getRemoteTransferTask(ctx context.Context, client fhir.Factory, fhirTaskID string) (resources.Task, error) {
	// TODO: Read AdvanceNotification here instead of the transfer task
	task := resources.Task{}
	err := client().ReadOne(ctx, "/Task/"+fhirTaskID, &task)
	if err != nil {
		return resources.Task{}, fmt.Errorf("error while looking up transfer task remotely(task-id=%s): %w", fhirTaskID, err)
	}
	return task, nil
}

// getAdvanceNotice fetches a complete nursing handoff from a FHIR server
func (s service) getNursingHandoff(ctx context.Context, fhirClient fhir.Client, fhirCompositionPath string) (eoverdracht.NursingHandoff, error) {
	nursingHandoff := eoverdracht.NursingHandoff{}

	// Fetch the composition
	err := fhirClient.ReadOne(ctx, "/"+fhirCompositionPath, &nursingHandoff.Composition)
	if err != nil {
		return eoverdracht.NursingHandoff{}, fmt.Errorf("error while fetching the advance notice composition(composition-id=%s): %w", fhirCompositionPath, err)
	}

	// Fetch the Patient
	err = fhirClient.ReadOne(ctx, "/"+fhir.FromStringPtr(nursingHandoff.Composition.Subject.Reference), &nursingHandoff.Patient)
	if err != nil {
		return eoverdracht.NursingHandoff{}, fmt.Errorf("error while fetching the transfer subject (patient): %w", err)
	}

	// Fetch the careplan
	careplan, err := eoverdracht.FilterCompositionSectionByType(nursingHandoff.Composition.Section, eoverdracht.CarePlanCode)
	if err != nil {
		return eoverdracht.NursingHandoff{}, err
	}

	// Fetchh the nursing diagnosis
	nursingDiagnosis, err := eoverdracht.FilterCompositionSectionByType(careplan.Section, eoverdracht.NursingDiagnosisCode)
	if err != nil {
		return eoverdracht.NursingHandoff{}, err
	}

	// the nursing diagnosis contains both conditions and procedures
	for _, entry := range nursingDiagnosis.Entry {
		if strings.HasPrefix(fhir.FromStringPtr(entry.Reference), "Condition") {
			conditionID := fhir.FromStringPtr(entry.Reference)
			condition := resources.Condition{}
			err = fhirClient.ReadOne(ctx, "/"+conditionID, &condition)
			if err != nil {
				return eoverdracht.NursingHandoff{}, fmt.Errorf("error while fetching a advance notice condition (condition-id=%s): %w", conditionID, err)
			}
			nursingHandoff.Problems = append(nursingHandoff.Problems, condition)
		}
		if strings.HasPrefix(fhir.FromStringPtr(entry.Reference), "Procedure") {
			procedureID := fhir.FromStringPtr(entry.Reference)
			procedure := eoverdracht.Procedure{}
			err = fhirClient.ReadOne(ctx, "/"+procedureID, &procedure)
			if err != nil {
				return eoverdracht.NursingHandoff{}, fmt.Errorf("error while fetching a advance notice procedure (procedure-id=%s): %w", procedureID, err)
			}
			nursingHandoff.Interventions = append(nursingHandoff.Interventions, procedure)
		}
	}

	return nursingHandoff, nil
}

// getAdvanceNotice fetches a complete advance notice from a FHIR server
func (s service) getAdvanceNotice(ctx context.Context, fhirClient fhir.Client, fhirCompositionPath string) (eoverdracht.AdvanceNotice, error) {
	advanceNotice := eoverdracht.AdvanceNotice{}

	err := fhirClient.ReadOne(ctx, "/"+fhirCompositionPath, &advanceNotice.Composition)
	if err != nil {
		return eoverdracht.AdvanceNotice{}, fmt.Errorf("error while fetching the advance notice composition(composition-id=%s): %w", fhirCompositionPath, err)
	}

	if advanceNotice.Composition.Subject.Reference != nil {
		err = fhirClient.ReadOne(ctx, "/"+fhir.FromStringPtr(advanceNotice.Composition.Subject.Reference), &advanceNotice.Patient)
		if err != nil {
			return eoverdracht.AdvanceNotice{}, fmt.Errorf("error while fetching the transfer subject (patient): %w", err)
		}
	}

	careplan, err := eoverdracht.FilterCompositionSectionByType(advanceNotice.Composition.Section, eoverdracht.CarePlanCode)
	if err != nil {
		return eoverdracht.AdvanceNotice{}, err
	}

	nursingDiagnosis, err := eoverdracht.FilterCompositionSectionByType(careplan.Section, eoverdracht.NursingDiagnosisCode)
	if err != nil {
		return eoverdracht.AdvanceNotice{}, err
	}

	// the nursing diagnosis contains both conditions and procedures
	for _, entry := range nursingDiagnosis.Entry {
		if strings.HasPrefix(fhir.FromStringPtr(entry.Reference), "Condition") {
			conditionID := fhir.FromStringPtr(entry.Reference)
			condition := resources.Condition{}
			err = fhirClient.ReadOne(ctx, "/"+conditionID, &condition)
			if err != nil {
				return eoverdracht.AdvanceNotice{}, fmt.Errorf("error while fetching a advance notice condition (condition-id=%s): %w", conditionID, err)
			}
			advanceNotice.Problems = append(advanceNotice.Problems, condition)
		}
		if strings.HasPrefix(fhir.FromStringPtr(entry.Reference), "Procedure") {
			procedureID := fhir.FromStringPtr(entry.Reference)
			procedure := eoverdracht.Procedure{}
			err = fhirClient.ReadOne(ctx, "/"+procedureID, &procedure)
			if err != nil {
				return eoverdracht.AdvanceNotice{}, fmt.Errorf("error while fetching a advance notice procedure (procedure-id=%s): %w", procedureID, err)
			}
			advanceNotice.Interventions = append(advanceNotice.Interventions, procedure)
		}
	}

	return advanceNotice, nil
}
