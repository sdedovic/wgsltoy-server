package service

import (
	"context"
	"errors"
	"fmt"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/sdedovic/wgsltoy-server/src/go/infra"
	"regexp"
	"time"
	"unicode/utf8"
)

type CreateShader struct {
	Name        string
	Visibility  string
	Description string
	Content     string
	Tags        []string
	CreatedBy   string
	ForkedFrom  string
}

var displayTextRegex = regexp.MustCompile(`^[\pL\pM\pN\pP\pS]+$`)

const VisibilityPrivate = "private"
const VisibilityUnlisted = "unlisted"
const VisibilityPublic = "public"

func ShaderCreate(ctx context.Context, pgPool *pgxpool.Pool, shader *CreateShader) (string, error) {
	if shader == nil {
		return "", errors.New("shader is nil")
	}

	if shader.Name == "" {
		return "", infra.NewValidationError("Field 'name' is required!")
	}
	if utf8.RuneCountInString(shader.Name) > 160 {
		return "", infra.NewValidationError("Field 'name' is too long!")
	}
	if !displayTextRegex.MatchString(shader.Name) {
		return "", infra.NewValidationError("Field 'name' contains invalid characters!")
	}

	if shader.Visibility == "" {
		return "", infra.NewValidationError("Field 'visibility' is required!")
	}
	if shader.Visibility != VisibilityUnlisted && shader.Visibility != VisibilityPrivate && shader.Visibility != VisibilityPublic {
		return "", infra.NewValidationError("Field 'visibility' must be one of 'private', 'unlisted' or 'public'!")
	}

	if utf8.RuneCountInString(shader.Description) > 480 {
		return "", infra.NewValidationError("Field 'description' is too long!")
	}
	if shader.Description != "" && !displayTextRegex.MatchString(shader.Name) {
		return "", infra.NewValidationError("Field 'description' contains invalid characters!")
	}

	if utf8.RuneCountInString(shader.Content) > 5250 {
		return "", infra.NewValidationError("Field 'content' is too long!")
	}
	if shader.Content != "" && !displayTextRegex.MatchString(shader.Content) {
		return "", infra.NewValidationError("Field 'content' contains invalid characters!")
	}

	for idx, tag := range shader.Tags {
		if tag == "" {
			return "", infra.NewValidationError(fmt.Sprintf("Field 'tags[%d]' is empty!", idx))
		}
		if utf8.RuneCountInString(tag) < 3 {
			return "", infra.NewValidationError(fmt.Sprintf("Field 'tags[%d]' is too short!", idx))
		}
		if utf8.RuneCountInString(tag) > 10 {
			return "", infra.NewValidationError(fmt.Sprintf("Field 'tags[%d]' is too long!", idx))
		}
		if !displayTextRegex.MatchString(tag) {
			return "", infra.NewValidationError(fmt.Sprintf("Field 'tags[%d]' contains invalid characters!", idx))
		}
	}
	tags := shader.Tags
	if tags == nil || len(tags) != 0 {
		tags = make([]string, 0)
	}

	if utf8.RuneCountInString(shader.ForkedFrom) > 50 {
		return "", infra.NewValidationError("Field 'forkedFrom' is too long!")
	}
	if shader.ForkedFrom != "" && !displayTextRegex.MatchString(shader.ForkedFrom) {
		return "", infra.NewValidationError("Field 'forkedFrom' contains invalid characters!")
	}
	var forkedFrom *string
	if shader.ForkedFrom == "" {
		forkedFrom = nil
	}

	conn, err := pgPool.Acquire(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to aquire connection to database caused by: %w", err)
	}

	now := time.Now()
	_, err = conn.Exec(ctx, "INSERT INTO shaders (created_at, updated_at, created_by, visibility, name, description, content, tags, forked_from) VALUES ($1, $1, $2, $3, $4, $5, $6, $7, $8)",
		now, shader.CreatedBy, shader.Visibility, shader.Name, shader.Description, shader.Content, tags, forkedFrom)
	if err != nil {
		return "", fmt.Errorf("failed inserting user caused by: %w", err)
	}

	return shader.Name, nil
}
