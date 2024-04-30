package dao

import (
	"context"
	_ "embed"

	"github.com/georgysavva/scany/v2/pgxscan"
	"github.com/jackc/pgx/v5"
	"github.com/samber/lo"

	"github.com/Work4Labs/uservice-applications/pkg/entities"
)

type Application struct {
	DB pgx.Tx
}

func NewApplication(tx pgx.Tx) *Application {
	return &Application{
		DB: tx,
	}
}

var getApplicationQuery string = `WITH
application AS (
    SELECT 
        a.id,
        a.external_id,
        a.organization_name,
        a.job_id,
        a.campaign_id,
        a.candidate_id,
        a.created_at,
        a.updated_at,
        a.last_interaction_date
    FROM applications.applications a
    WHERE external_id = $1 AND ($2::TEXT[] IS NULL OR organization_name = ANY ($2::TEXT[]))
),
answers AS (
    SELECT answers.application_id, to_jsonb(array_remove(array_agg(answers), NULL)) AS agg_answers
    FROM applications.answers
    WHERE application_id = (SELECT id FROM application)
    GROUP BY answers.application_id
),
comments AS (
    SELECT comments.application_id, to_jsonb(array_remove(array_agg(comments), NULL)) AS agg_comments
    FROM applications.comments
    WHERE application_id = (SELECT id FROM application)
    GROUP BY comments.application_id
),
statuses AS (
    SELECT statusesb.application_id, to_jsonb(array_remove(array_agg(statusesb), NULL)) AS agg_statuses
    FROM (
        SELECT a_as.external_id AS id, a_as.application_id, a_as.status_id, a_as.user_entity_id, a_as.created_at, a_as.is_bulk_action, s AS status
        FROM applications.application_statuses a_as
            INNER JOIN applications.statuses s ON s.id = a_as.status_id
        WHERE application_id = (SELECT id FROM application)
        ORDER BY a_as.created_at ASC
    ) statusesb
    GROUP BY statusesb.application_id
)
SELECT
    a.id,
    a.external_id,
    a.organization_name,
    a.job_id,
    a.campaign_id,
    a.candidate_id,
    a.created_at,
    a.updated_at,
    a.last_interaction_date,
    answers.agg_answers as answers,
    comments.agg_comments as comments,
    statuses.agg_statuses as statuses
FROM application a
LEFT JOIN answers ON a.id = answers.application_id
LEFT JOIN comments ON a.id = comments.application_id
LEFT JOIN statuses ON statuses.application_id = a.id
`

func (a *Application) GetApplication(ctx context.Context, id string, organizations []string) (*entities.Application, error) {
	application := new(entities.Application)
	err := pgxscan.Get(
		ctx, a.DB, application, getApplicationQuery,
		id, lo.Ternary(len(organizations) > 0, organizations, nil),
	)

	return application, err
}
