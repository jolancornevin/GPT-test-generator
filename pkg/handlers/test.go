package handlers_test

import (
	"context"
	"errors"
	"net/http"
	"testing"

	"github.com/go-openapi/runtime/middleware"
	"github.com/go-openapi/strfmt"
	"github.com/google/uuid"
	"github.com/samber/lo"
	tassert "github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/Work4Labs/go_framework/sdk/keycloak"

	"github.com/Work4Labs/uservice-applications/models"
	"github.com/Work4Labs/uservice-applications/pkg/entities"
	"github.com/Work4Labs/uservice-applications/pkg/handlers"
	"github.com/Work4Labs/uservice-applications/pkg/handlers/mocks"
	"github.com/Work4Labs/uservice-applications/restapi/operations/applications"
)

func TestCreateApplication(t *testing.T) {
	ctx := context.Background()

	var (
		user = &keycloak.JWTUser{
			IsAuthenticated: true,
			JWT:             "user-token",
		}

		requestID   = "request_id"
		validParams = applications.CreateApplicationByJobIDParams{
			HTTPRequest: (&http.Request{}).WithContext(ctx),
			RequestID:   &requestID,
			Application: &models.ApplicationCreation{
				Answers: []*models.ApplicationAnswerCreation{
					{
						QuestionLabel: "question one",
						Answer:        "oui",
					},
					{
						QuestionLabel: "question two",
						Answer:        "non",
					},
				},
				CampaignID:       "00000000-0000-0000-0000-000000000011",
				CandidateID:      "00000000-0000-0000-0000-000000000012",
				JobID:            strfmt.UUID("00000000-0000-0000-0000-000000000013"),
				OrganizationName: "seiza",
			},
			UtmCampaign: lo.ToPtr("12"),
			UtmMedium:   lo.ToPtr("lead"),
			UtmSource:   lo.ToPtr("facebook"),
		}

		application = &entities.Application{
			ExternalID:       uuid.MustParse("00000000-0000-0000-0000-000000000113"),
			JobID:            "00000000-0000-0000-0000-000000000013",
			OrganizationName: "seiza",
			CampaignID:       "00000000-0000-0000-0000-000000000011",
			CandidateID:      lo.ToPtr("00000000-0000-0000-0000-000000000012"),
		}
	)

	flagTestCreateApplication := []struct {
		name string

		depsErr error

		queryParams   applications.CreateApplicationByJobIDParams
		serviceParams *entities.ApplicationCreation

		serviceRes *entities.Application
		serviceErr error

		expectedRes middleware.Responder

		commitCalled *bool
	}{
		{
			name: "ok",

			depsErr: nil,

			queryParams: validParams,
			serviceParams: &entities.ApplicationCreation{
				Answers: []entities.AnswerCreation{
					{
						QuestionLabel: "question one",
						Answer:        "oui",
					},
					{
						QuestionLabel: "question two",
						Answer:        "non",
					},
				},
				CampaignID:       "00000000-0000-0000-0000-000000000011",
				CandidateID:      "00000000-0000-0000-0000-000000000012",
				JobID:            strfmt.UUID("00000000-0000-0000-0000-000000000013"),
				OrganizationName: "seiza",
				UTMParameters: []entities.UTMParameters{
					{Label: "utm_campaign", Value: "12"},
					{Label: "utm_medium", Value: "lead"},
					{Label: "utm_source", Value: "facebook"},
				},
			},

			serviceRes: application,
			serviceErr: nil,

			expectedRes: applications.NewGetApplicationOK().WithPayload(&models.ApplicationDetails{
				ID:               application.ExternalID.String(),
				JobID:            application.JobID,
				OrganizationName: application.OrganizationName,
				CampaignID:       application.CampaignID,
				CandidateID:      *application.CandidateID,
			}),

			commitCalled: lo.ToPtr(true),
		},
		{
			name: "ok - with utm nil",

			depsErr: nil,

			queryParams: applications.CreateApplicationByJobIDParams{
				HTTPRequest: (&http.Request{}).WithContext(ctx),
				RequestID:   &requestID,
				Application: &models.ApplicationCreation{
					Answers: []*models.ApplicationAnswerCreation{
						{
							QuestionLabel: "question one",
							Answer:        "oui",
						},
						{
							QuestionLabel: "question two",
							Answer:        "non",
						},
					},
					CampaignID:       "00000000-0000-0000-0000-000000000011",
					CandidateID:      "00000000-0000-0000-0000-000000000012",
					JobID:            strfmt.UUID("00000000-0000-0000-0000-000000000013"),
					OrganizationName: "seiza",
				},

				UtmCampaign: nil,
				UtmMedium:   nil,
				UtmSource:   nil,
			},

			serviceParams: &entities.ApplicationCreation{
				Answers: []entities.AnswerCreation{
					{
						QuestionLabel: "question one",
						Answer:        "oui",
					},
					{
						QuestionLabel: "question two",
						Answer:        "non",
					},
				},
				CampaignID:       "00000000-0000-0000-0000-000000000011",
				CandidateID:      "00000000-0000-0000-0000-000000000012",
				JobID:            strfmt.UUID("00000000-0000-0000-0000-000000000013"),
				OrganizationName: "seiza",
				UTMParameters:    nil,
			},

			serviceRes: application,
			serviceErr: nil,

			expectedRes: applications.NewGetApplicationOK().WithPayload(&models.ApplicationDetails{
				ID:               application.ExternalID.String(),
				JobID:            application.JobID,
				OrganizationName: application.OrganizationName,
				CampaignID:       application.CampaignID,
				CandidateID:      *application.CandidateID,
			}),

			commitCalled: lo.ToPtr(true),
		},
		{
			name: "ko - init service",

			depsErr: errors.New("failed to init service"),

			queryParams:   validParams,
			serviceParams: nil,

			serviceRes: application,
			serviceErr: nil,

			expectedRes:  applications.NewUpdateApplicationStatusInternalServerError().WithPayload("failed to create application"),
			commitCalled: nil,
		},
		{
			name: "ko - service error",

			depsErr: nil,

			queryParams: validParams,
			serviceParams: &entities.ApplicationCreation{
				Answers: []entities.AnswerCreation{
					{
						QuestionLabel: "question one",
						Answer:        "oui",
					},
					{
						QuestionLabel: "question two",
						Answer:        "non",
					},
				},
				CampaignID:       "00000000-0000-0000-0000-000000000011",
				CandidateID:      "00000000-0000-0000-0000-000000000012",
				JobID:            strfmt.UUID("00000000-0000-0000-0000-000000000013"),
				OrganizationName: "seiza",
				UTMParameters: []entities.UTMParameters{
					{Label: "utm_campaign", Value: "12"},
					{Label: "utm_medium", Value: "lead"},
					{Label: "utm_source", Value: "facebook"},
				},
			},

			serviceRes: nil,
			serviceErr: errors.New("fail during service"),

			expectedRes:  applications.NewCreateApplicationInternalServerError().WithPayload(models.ApplicationsError("failed to create application")),
			commitCalled: lo.ToPtr(false),
		},
	}

	for _, tt := range flagTestCreateApplication {
		tt := tt

		t.Run(tt.name, func(t *testing.T) {
			assert := tassert.New(t)

			applicationCommentService := &mocks.CreateApplicationService{}
			applicationCommentService.On("CreateApplication", mock.Anything, tt.serviceParams, user).Return(tt.serviceRes, tt.serviceErr)

			tx := &mocks.Tx{}
			tx.On("Commit", mock.Anything).Return(nil)
			tx.On("Rollback", mock.Anything).Return(nil)

			handler := &handlers.CreateApplication{
				ServiceFactory: func(ctx context.Context) (*handlers.CreateApplicationDependencies, error) {
					return &handlers.CreateApplicationDependencies{
						Service: applicationCommentService,
						Tx:      tx,
					}, tt.depsErr
				},
			}

			res := handler.Handle(tt.queryParams, user)
			assert.Equal(tt.expectedRes, res)

			if tt.commitCalled != nil {
				if *tt.commitCalled {
					tx.AssertCalled(t, "Commit", mock.Anything)
					applicationCommentService.AssertExpectations(t)
				} else {
					tx.AssertCalled(t, "Rollback", mock.Anything)
				}
			}
		})
	}
}
