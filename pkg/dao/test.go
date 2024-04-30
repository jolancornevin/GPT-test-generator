package dao_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/samber/lo"
	tassert "github.com/stretchr/testify/assert"

	"github.com/Work4Labs/uservice-applications/pkg/dao"
	"github.com/Work4Labs/uservice-applications/pkg/database"
	"github.com/Work4Labs/uservice-applications/pkg/entities"
)

func TestGetApplication(t *testing.T) {
	const (
		applicationID  = 3
		answerID       = 4
		secondAnswerID = 40

		applicationExternalID = "00000000-0000-0000-0000-000000000010"
		organizationName      = "organization_name"

		jobID = "00000000-0000-0000-0000-000000000110"

		lastInteractionDate = "2023-04-06 13:14:51"

		questionLabel       = "questionLabel"
		answerLabel         = "answer"
		secondQuestionLabel = "questionLabel"
		secondAnswerLabel   = "answer"

		commentID       = 5
		secondCommentID = 50
		content         = "this is a comment"
		secondContent   = "this is a comment"
		userEntityID    = "00000000-0000-0000-0000-000000000011"

		statusID               = 6
		statusExternalID       = "00000000-0000-0000-0000-000000000012"
		secondStatusID         = 60
		secondStatusExternalID = "00000000-0000-0000-0000-000000000112"

		statusLabel       = "hired"
		secondStatusLabel = "rejected"

		applicationStatusID               = 7
		applicationStatusExternalID       = "00000000-0000-0000-0000-000000000015"
		applicationSecondStatusID         = 70
		applicationStatusSecondExternalID = "00000000-0000-0000-0000-000000000016"

		campaignID  = "00000000-0000-0000-0000-000000000013"
		candidateID = "00000000-0000-0000-0000-000000000014"

		isBulkAction = true
	)

	createdAt := time.Now().UTC().Truncate(1 * time.Second)
	secondCreatedAt := time.Now().UTC().Add(-10 * time.Minute).Truncate(1 * time.Second)
	updatedAt := time.Now().UTC().Truncate(1 * time.Second)

	_ = entities.Candidate{
		ID:               "00000000-0000-0000-0000-000000000015",
		Email:            "test@seiza.co",
		EmailConstraint:  "",
		EmailVerified:    true,
		Enabled:          true,
		FirstName:        "test",
		LastName:         "last",
		RealmID:          "1",
		Username:         "testlast",
		CreatedTimestamp: lo.ToPtr(createdAt.Unix()),
	}

	application := entities.Application{
		JobID:            jobID,
		ExternalID:       uuid.MustParse(applicationExternalID),
		OrganizationName: organizationName,
		CampaignID:       campaignID,
		CandidateID:      lo.ToPtr(candidateID),
	}

	firstApplicationStatus := entities.ApplicationStatus{
		ID:            applicationStatusExternalID,
		ApplicationID: applicationID,
		StatusID:      statusID,
		UserEntityID:  userEntityID,
		IsBulkAction:  isBulkAction,

		Status: &entities.Status{
			ID: statusID,
			// ExternalID: statusExternalID, // TODO return the external id, not the id
			Label: statusLabel,
		},
		CreatedAt: database.CustomTime(createdAt),
	}

	secondApplicationStatus := entities.ApplicationStatus{
		ID:            applicationStatusSecondExternalID,
		ApplicationID: applicationID,
		StatusID:      secondStatusID,
		UserEntityID:  userEntityID,
		IsBulkAction:  isBulkAction,

		Status: &entities.Status{
			ID: secondStatusID,
			// ExternalID: statusExternalID, // TODO return the external id, not the id
			Label: secondStatusLabel,
		},
		CreatedAt: database.CustomTime(secondCreatedAt),
	}

	firstAnswer := entities.Answer{
		ID: answerID,
		// ExternalID: statusExternalID, // TODO return the external id, not the id
		ApplicationID: applicationID,
		QuestionLabel: questionLabel,
		Answer:        answerLabel,
	}

	secondAnswer := entities.Answer{
		ID: secondAnswerID,
		// ExternalID: statusExternalID, // TODO return the external id, not the id
		ApplicationID: applicationID,
		QuestionLabel: secondQuestionLabel,
		Answer:        secondAnswerLabel,
	}

	firstComment := entities.Comment{
		ID: commentID,
		// ExternalID: statusExternalID, // TODO return the external id, not the id
		ApplicationID: applicationID,
		UserEntityID:  userEntityID,
		IsBulkAction:  isBulkAction,
		Kind:          "COMMENT",
		Content:       content,
		CreatedAt:     database.CustomTime(createdAt),
	}

	secondComment := entities.Comment{
		ID: secondCommentID,
		// ExternalID: statusExternalID, // TODO return the external id, not the id
		ApplicationID: applicationID,
		UserEntityID:  userEntityID,
		IsBulkAction:  isBulkAction,
		Kind:          "COMMENT",
		Content:       secondContent,
		CreatedAt:     database.CustomTime(secondCreatedAt),
	}

	ctx := context.Background()

	initDBForApplication := func(ctx context.Context, tx pgx.Tx) (err error) {
		// start the tx
		_, err = tx.Exec(ctx, `
			CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
			CREATE SCHEMA IF NOT EXISTS applications;

			CREATE TABLE applications.applications (
				id SERIAL PRIMARY KEY,
				external_id TEXT,

				job_id TEXT NOT NULL,
				organization_name TEXT,

				campaign_id TEXT NOT NULL,
				candidate_id CHARACTER VARYING(36),

				created_at TIMESTAMP NOT NULL DEFAULT NOW(),
				updated_at TIMESTAMP NOT NULL DEFAULT NOW(),

				last_interaction_date timestamptz NOT NULL DEFAULT NOW()
			);
			
			CREATE TABLE applications.answers (
				id BIGINT PRIMARY KEY,
				application_id BIGINT NOT NULL,
				question_label TEXT NOT NULL,
				answer TEXT NOT NULL,
				CONSTRAINT answers FOREIGN KEY (application_id) REFERENCES applications.applications (id)
			);

			CREATE TYPE comment_kind AS ENUM (
				'COMMENT',
				'INTERVIEW_CANCELLED'
			);

			CREATE TABLE applications.comments (
				id BIGINT  PRIMARY KEY,
				application_id BIGINT NOT NULL,
				content TEXT NOT NULL,
				user_entity_id CHARACTER VARYING(36) NOT NULL,
				kind comment_kind NOT NULL DEFAULT 'COMMENT',
				"is_bulk_action" BOOLEAN NOT NULL DEFAULT FALSE,

				created_at TIMESTAMP NOT NULL DEFAULT NOW(),
				CONSTRAINT comments FOREIGN KEY (application_id) REFERENCES applications.applications (id)
			);

			CREATE TABLE applications.statuses (
				id BIGINT PRIMARY KEY,
				external_id uuid DEFAULT uuid_generate_v4() NOT NULL UNIQUE,
				label TEXT NOT NULL
			);

			CREATE TABLE applications.application_statuses (
				id BIGINT PRIMARY KEY,
				external_id TEXT,

				application_id BIGINT NOT NULL,
				status_id BIGINT NOT NULL,
				user_entity_id CHARACTER VARYING(36) NOT NULL,
				
				"is_bulk_action" BOOLEAN NOT NULL DEFAULT FALSE,
			
				created_at TIMESTAMP NOT NULL DEFAULT NOW(),
			
				CONSTRAINT application_fk FOREIGN KEY (application_id) REFERENCES applications.applications (id),
				CONSTRAINT status_fk FOREIGN KEY (status_id) REFERENCES applications.statuses (id) ON DELETE SET NULL
			);
		`)
		if err != nil {
			return err
		}

		return nil
	}

	fixturesForApplication := func(ctx context.Context, tx pgx.Tx) (err error) {
		_, err = tx.Exec(
			ctx,
			`INSERT INTO applications.applications (
				id,
				external_id,
				organization_name,
				job_id,
				campaign_id,
				candidate_id,
				last_interaction_date,
				created_at,
				updated_at
			) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9);`,
			applicationID,
			applicationExternalID,
			organizationName,
			jobID,
			campaignID,
			candidateID,
			lastInteractionDate,
			createdAt,
			updatedAt,
		)
		if err != nil {
			return err
		}

		_, err = tx.Exec(
			ctx,
			`INSERT INTO applications.answers (
				id, application_id, question_label, answer
			) VALUES ($1, $2, $3, $4);`,
			answerID,
			applicationID,
			questionLabel,
			answerLabel,
		)
		if err != nil {
			return err
		}

		_, err = tx.Exec(
			ctx,
			`INSERT INTO applications.answers (
				id, application_id, question_label, answer
			) VALUES ($1, $2, $3, $4);`,
			secondAnswerID,
			applicationID,
			secondQuestionLabel,
			secondAnswerLabel,
		)
		if err != nil {
			return err
		}

		_, err = tx.Exec(
			ctx,
			`INSERT INTO applications.comments (
				id, application_id, content, is_bulk_action, user_entity_id, created_at
			) VALUES ($1, $2, $3, $4, $5, $6);`,
			commentID,
			applicationID,
			content,
			isBulkAction,
			userEntityID,
			createdAt,
		)
		if err != nil {
			return err
		}

		_, err = tx.Exec(
			ctx,
			`INSERT INTO applications.comments (
				id, application_id, content, is_bulk_action, user_entity_id, created_at
			) VALUES ($1, $2, $3, $4, $5, $6);`,
			secondCommentID,
			applicationID,
			secondContent,
			isBulkAction,
			userEntityID,
			secondCreatedAt,
		)
		if err != nil {
			return err
		}

		_, err = tx.Exec(
			ctx,
			`INSERT INTO applications.statuses (
				id, external_id, label
			) VALUES ($1, $2, $3);`,
			statusID,
			statusExternalID,
			statusLabel,
		)
		if err != nil {
			return err
		}

		_, err = tx.Exec(
			ctx,
			`INSERT INTO applications.statuses (
				id, external_id, label
			) VALUES ($1, $2, $3);`,
			secondStatusID,
			secondStatusExternalID,
			secondStatusLabel,
		)
		if err != nil {
			return err
		}

		_, err = tx.Exec(
			ctx,
			`INSERT INTO applications.application_statuses (
				id, external_id, application_id, status_id, user_entity_id, is_bulk_action, created_at
			) VALUES ($1, $2, $3, $4, $5, $6, $7);`,
			applicationStatusID,
			applicationStatusExternalID,
			applicationID,
			statusID,
			userEntityID,
			isBulkAction,
			createdAt,
		)
		if err != nil {
			return err
		}

		_, err = tx.Exec(
			ctx,
			`INSERT INTO applications.application_statuses (
				id, external_id, application_id, status_id, user_entity_id, is_bulk_action, created_at
			) VALUES ($1, $2, $3, $4, $5, $6, $7);`,
			applicationSecondStatusID,
			applicationStatusSecondExternalID,
			applicationID,
			secondStatusID,
			userEntityID,
			isBulkAction,
			secondCreatedAt,
		)
		if err != nil {
			return err
		}

		return nil
	}

	db, err := connect(ctx)
	if err != nil {
		t.Fatal(err)
	}

	defer db.Close()

	flagTestGetApplication := []struct {
		name string

		getFunc func(*dao.Application) (*entities.Application, error)

		application         entities.Application
		applicationStatuses []entities.ApplicationStatus
		answers             []entities.Answer
		comments            []entities.Comment

		expectedErr error
	}{
		{
			name: "ok GetApplicationByID",

			getFunc: func(applicationDAO *dao.Application) (*entities.Application, error) {
				return applicationDAO.GetApplication(ctx, applicationExternalID, []string{organizationName})
			},

			application:         application,
			applicationStatuses: []entities.ApplicationStatus{secondApplicationStatus, firstApplicationStatus},
			answers:             []entities.Answer{firstAnswer, secondAnswer},
			comments:            []entities.Comment{firstComment, secondComment},

			expectedErr: nil,
		},
		{
			name: "ok GetApplicationByCandidate",

			getFunc: func(applicationDAO *dao.Application) (*entities.Application, error) {
				return applicationDAO.GetApplicationByCandidate(ctx, candidateID, campaignID, jobID)
			},

			application:         application,
			applicationStatuses: []entities.ApplicationStatus{secondApplicationStatus, firstApplicationStatus},
			answers:             []entities.Answer{firstAnswer, secondAnswer},
			comments:            []entities.Comment{firstComment, secondComment},

			expectedErr: nil,
		},
	}

	for _, tt := range flagTestGetApplication {
		tt := tt

		t.Run(tt.name, func(t *testing.T) {
			assert := tassert.New(t)

			tx, err := db.Begin(ctx)
			if err != nil {
				t.Fatal(err)
			}

			// rollback to clean the DB
			defer func() {
				err = tx.Rollback(ctx)
				if err != nil {
					t.Fatal(err)
				}
			}()

			if err = initDBForApplication(ctx, tx); err != nil {
				t.Error(err)
				return
			}

			err = fixturesForApplication(ctx, tx)
			if err != nil {
				t.Error(err)
				return
			}

			// Call the DAO to do the insert
			applicationDAO := dao.NewApplication(tx)

			app, err := tt.getFunc(applicationDAO)
			assert.ErrorIs(err, tt.expectedErr)

			if err == nil {
				assert.Equal(tt.application.JobID, app.JobID)
				assert.Equal(tt.application.ExternalID.String(), app.ExternalID.String())
				assert.Equal(tt.application.OrganizationName, app.OrganizationName)
				assert.Equal(tt.application.CampaignID, app.CampaignID)
				assert.Equal(*tt.application.CandidateID, *app.CandidateID)

				assert.Nil(app.Candidate)

				for i, applicationStatus := range tt.applicationStatuses {
					assert.Equal(applicationStatus, *app.Statuses[i])
				}

				for i, answer := range tt.answers {
					assert.Equal(answer, *app.Answers[i])
				}

				for i, comment := range tt.comments {
					assert.Equal(comment, *app.Comments[i])
				}
				assert.WithinRange(time.Now(), time.Time(app.LastInteractionDate), time.Now().UTC().Add(3*time.Second))
			}
		})
	}
}
