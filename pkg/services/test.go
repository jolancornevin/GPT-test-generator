package services_test

import (
	"context"
	"errors"
	"testing"

	"github.com/Work4Labs/go_framework/sdk/keycloak"
	"github.com/Work4Labs/go_models/models"
	"github.com/Work4Labs/uservice-applications/pkg/services"
	"github.com/Work4Labs/uservice-applications/pkg/services/mocks"
	"github.com/Work4Labs/uservice-applications/restapi/operations/applications"
	"github.com/stretchr/testify/mock"
)

func TestApplicationCreateCommentService(t *testing.T) {
	ctx := context.Background()

	var (
		user = &keycloak.JWTUser{
			ID: "user_id",
		}
		errDAO = errors.New("failed to create comment")
	)

	flagTestCreateHistory := []struct {
		name string

		params applications.CreateApplicationCommentParams

		resDao int64
		errDao error

		expectedErr error
	}{
		{
			name: "ok",

			params: applications.CreateApplicationCommentParams{
				ApplicationID: "application_uuid",
				Comment: &models.ApplicationCommentCreation{
					Content:      "content",
					IsBulkAction: false,
					Kind:         "COMMENT",
				},
			},

			resDao: 1,
		},
		{
			name: "err invalid kind",

			params: applications.CreateApplicationCommentParams{
				ApplicationID: "application_uuid",
				Comment: &models.ApplicationCommentCreation{
					Content:      "content",
					IsBulkAction: false,
					Kind:         "string",
				},
			},

			expectedErr: services.ErrInvalidCommentKind,
		},
		{
			name: "err DAO ko",

			params: applications.CreateApplicationCommentParams{
				ApplicationID: "application_uuid",
				Comment: &models.ApplicationCommentCreation{
					Content:      "content",
					IsBulkAction: false,
					Kind:         "COMMENT",
				},
			},

			resDao: 0,
			errDao: errDAO,

			expectedErr: errDAO,
		},
	}

	for _, tt := range flagTestCreateHistory {
		tt := tt

		t.Run(tt.name, func(t *testing.T) {
			assert := tassert.New(t)

			commentDAO := &mocks.ApplicationCommentCreator{}
			commentDAO.On(
				"CreateApplicationComment",
				mock.Anything,
				tt.params.ApplicationID,
				tt.params.Comment.Content,
				"user_id",
				tt.params.Comment.Kind,
				tt.params.Comment.IsBulkAction,
			).Return(tt.resDao, tt.errDao)

			service := services.NewApplicationCreateCommentService(commentDAO)

			err := service.CreateApplicationComment(ctx, tt.params, user)

			assert.ErrorIs(err, tt.expectedErr)
			commentDAO.AssertExpectations(t)
		})
	}
}
