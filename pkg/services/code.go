package services

import (
	"context"
	"fmt"

	"github.com/Work4Labs/go_framework/sdk/keycloak"
	"github.com/Work4Labs/uservice-applications/restapi/operations/applications"
)

type ApplicationCommentCreator interface {
	CreateApplicationComment(ctx context.Context, applicationExternalID, content, userEntityID, kind string, isBulkAction bool) (int64, error)
}

type ApplicationCreateCommentService struct {
	commentDAO ApplicationCommentCreator
}

func NewApplicationCreateCommentService(commentDAO ApplicationCommentCreator) *ApplicationCreateCommentService {
	return &ApplicationCreateCommentService{
		commentDAO: commentDAO,
	}
}

func (s *ApplicationCreateCommentService) CreateApplicationComment(ctx context.Context, params applications.CreateApplicationCommentParams, principal *keycloak.JWTUser) error {
	if !lo.Contains(allowedCommentKinds, params.Comment.Kind) {
		return fmt.Errorf("invalid comment kind '%s': %w", params.Comment.Kind, ErrInvalidCommentKind)
	}

	_, err := s.commentDAO.CreateApplicationComment(
		ctx,
		params.ApplicationID,
		params.Comment.Content,
		principal.ID,
		params.Comment.Kind,
		params.Comment.IsBulkAction,
	)

	return err
}
