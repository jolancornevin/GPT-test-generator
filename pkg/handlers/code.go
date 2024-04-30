package handlers

import (
	"context"

	"github.com/mitchellh/mapstructure"
	"github.com/samber/lo"

	"github.com/Work4Labs/go_framework/db"
	"github.com/Work4Labs/go_framework/helpers"
	"github.com/Work4Labs/go_framework/sdk/keycloak"

	"github.com/Work4Labs/uservice-applications/models"
	"github.com/Work4Labs/uservice-applications/pkg/entities"
	"github.com/Work4Labs/uservice-applications/restapi/operations/applications"

	"github.com/go-openapi/runtime/middleware"
	log "github.com/sirupsen/logrus"
)

type CreateApplicationServiceFactory func(ctx context.Context) (*CreateApplicationDependencies, error)

type CreateApplication struct {
	ServiceFactory CreateApplicationServiceFactory
}

func NewCreateApplication() *CreateApplication {
	return &CreateApplication{ServiceFactory: InitCreateApplicationDependencies}
}

func (h *CreateApplication) Handle(params applications.CreateApplicationByJobIDParams, principal *keycloak.JWTUser) middleware.Responder {
	requestID := lo.FromPtr(params.RequestID)
	ctx, logger := helpers.ContextWithLog(
		params.HTTPRequest.Context(),
		log.Fields{"request_id": requestID, "user_groups": principal.Groups},
	)

	deps, err := h.ServiceFactory(ctx)
	if err != nil {
		logger.WithError(err).Error("failed to create dependencies for create application handler")
		return applications.NewUpdateApplicationStatusInternalServerError().WithPayload("failed to create application")
	}

	defer func() {
		db.CleanTx(ctx, deps.Tx, err)
	}()

	applicationCreation := &entities.ApplicationCreation{}
	if err = mapstructure.Decode(params.Application, applicationCreation); err != nil {
		logger.WithError(err).Error("failed to create entities to create application handler")
		return applications.NewUpdateApplicationStatusInternalServerError().WithPayload("failed to create application")
	}

	if params.UtmCampaign != nil {
		applicationCreation.UTMParameters = append(applicationCreation.UTMParameters, entities.UTMParameters{
			Label: "utm_campaign", Value: *params.UtmCampaign,
		})
	}

	application, err := deps.Service.CreateApplication(ctx, applicationCreation, principal)
	if err != nil {
		logger.WithError(err).Error("failed to handle creating application")
		return applications.NewCreateApplicationInternalServerError().WithPayload(models.ApplicationsError("failed to create application"))
	}

	return applications.NewGetApplicationOK().WithPayload(&models.ApplicationDetails{
		ID:               application.ExternalID.String(),
		JobID:            application.JobID,
		OrganizationName: application.OrganizationName,
		CampaignID:       application.CampaignID,
		CandidateID:      *application.CandidateID,
	})
}
