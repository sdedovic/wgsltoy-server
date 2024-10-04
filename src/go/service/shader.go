package service

import (
	"context"
	"errors"
	"fmt"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/sdedovic/wgsltoy-server/src/go/infra"
	"regexp"
	"strings"
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

var tagRegex = regexp.MustCompile(`^[[:alnum:]]+$`)
var displayRegex = regexp.MustCompile(`^[\pL\pM\pN\pP\pS ]+$`)
var displayMultilineRegex = regexp.MustCompile(`^[\pL\pM\pN\pP\pS\s]+$`)

const VisibilityPrivate = "private"
const VisibilityUnlisted = "unlisted"
const VisibilityPublic = "public"

func validateShaderName(name string) error {
	if name == "" {
		return infra.NewValidationError("Field 'name' is required!")
	}
	if utf8.RuneCountInString(name) > 160 {
		return infra.NewValidationError("Field 'name' is too long!")
	}
	if !displayRegex.MatchString(name) {
		return infra.NewValidationError("Field 'name' contains invalid characters!")
	}
	return nil
}

func validateShaderVisibility(visibility string) error {
	if visibility == "" {
		return infra.NewValidationError("Field 'visibility' is required!")
	}
	if visibility != VisibilityUnlisted && visibility != VisibilityPrivate && visibility != VisibilityPublic {
		return infra.NewValidationError("Field 'visibility' must be one of 'private', 'unlisted' or 'public'!")
	}
	return nil
}

func validateShaderDescription(description string) error {
	if utf8.RuneCountInString(description) > 480 {
		return infra.NewValidationError("Field 'description' is too long!")
	}
	if description != "" && !displayMultilineRegex.MatchString(description) {
		return infra.NewValidationError("Field 'description' contains invalid characters!")
	}
	return nil
}

func validateShaderContent(content string) error {
	if utf8.RuneCountInString(content) > 5250 {
		return infra.NewValidationError("Field 'content' is too long!")
	}
	if content != "" && !displayMultilineRegex.MatchString(content) {
		return infra.NewValidationError("Field 'content' contains invalid characters!")
	}
	return nil
}

func validateShaderTags(tags []string) error {
	for idx, tag := range tags {
		if tag == "" {
			return infra.NewValidationError(fmt.Sprintf("Field 'tags[%d]' is empty!", idx))
		}
		if utf8.RuneCountInString(tag) < 3 {
			return infra.NewValidationError(fmt.Sprintf("Field 'tags[%d]' is too short!", idx))
		}
		if utf8.RuneCountInString(tag) > 10 {
			return infra.NewValidationError(fmt.Sprintf("Field 'tags[%d]' is too long!", idx))
		}
		if !tagRegex.MatchString(tag) {
			return infra.NewValidationError(fmt.Sprintf("Field 'tags[%d]' contains invalid characters!", idx))
		}
	}
	return nil
}

func ShaderCreate(ctx context.Context, pgPool *pgxpool.Pool, shader *CreateShader) (string, error) {
	if shader == nil {
		return "", errors.New("shader is nil")
	}

	if err := validateShaderName(shader.Name); err != nil {
		return "", err
	}

	if err := validateShaderVisibility(shader.Visibility); err != nil {
		return "", err
	}

	if err := validateShaderDescription(shader.Description); err != nil {
		return "", err
	}

	if err := validateShaderContent(shader.Content); err != nil {
		return "", err
	}

	if err := validateShaderTags(shader.Tags); err != nil {
		return "", err
	}
	tags := shader.Tags
	if tags == nil {
		tags = make([]string, 0)
	}

	if shader.ForkedFrom != "" && !ValidateGUID(shader.ForkedFrom) {
		return "", infra.NewValidationError("Field 'forkedFrom' is invalid!")
	}
	var forkedFrom *string
	if shader.ForkedFrom != "" {
		forkedFrom = &shader.ForkedFrom
	}

	now := time.Now()
	guid := NewGUID()
	_, err := pgPool.Exec(ctx, "INSERT INTO shaders (created_at, updated_at, created_by, visibility, name, description, content, tags, forked_from, shader_id) VALUES ($1, $1, $2, $3, $4, $5, $6, $7, $8, $9)",
		now, shader.CreatedBy, shader.Visibility, shader.Name, shader.Description, shader.Content, tags, forkedFrom, guid)
	if err != nil {
		return "", fmt.Errorf("failed inserting user caused by: %w", err)
	}

	return guid, nil
}

type Shader struct {
	Id        string    `db:"shader_id"`
	CreatedAt time.Time `db:"created_at"`
	UpdatedAt time.Time `db:"updated_at"`

	Name        string   `db:"name"`
	Visibility  string   `db:"visibility"`
	Description string   `db:"description"`
	ForkedFrom  string   `db:"forked_from"`
	Tags        []string `db:"tags"`

	Content string `db:"content"`
}

type UpdateShader struct {
	Name        string
	Visibility  string
	Description string
	Tags        []string
	Content     string
}

func ShaderUpdate(ctx context.Context, pgPool *pgxpool.Pool, userId string, shaderId string, shader UpdateShader) error {
	var query strings.Builder
	args := make([]any, 0, 8)

	query.WriteString("UPDATE shaders SET updated_at = $1 ")
	args = append(args, time.Now())

	if shader.Name != "" {
		if err := validateShaderName(shader.Name); err != nil {
			return err
		}

		args = append(args, shader.Name)
		_, err := fmt.Fprintf(&query, ", name = $%d", len(args))
		if err != nil {
			return err
		}
	}

	if shader.Visibility != "" {
		if err := validateShaderVisibility(shader.Visibility); err != nil {
			return err
		}

		args = append(args, shader.Visibility)
		_, err := fmt.Fprintf(&query, ", visibility = $%d", len(args))
		if err != nil {
			return err
		}
	}

	if shader.Description != "" {
		if err := validateShaderDescription(shader.Description); err != nil {
			return err
		}

		args = append(args, shader.Description)
		_, err := fmt.Fprintf(&query, ", description = $%d", len(args))
		if err != nil {
			return err
		}
	}

	if shader.Content != "" {
		if err := validateShaderContent(shader.Content); err != nil {
			return err
		}

		args = append(args, shader.Content)
		_, err := fmt.Fprintf(&query, ", content = $%d", len(args))
		if err != nil {
			return err
		}
	}

	if shader.Tags != nil {
		if err := validateShaderTags(shader.Tags); err != nil {
			return err
		}

		args = append(args, shader.Tags)
		_, err := fmt.Fprintf(&query, ", tags = $%d", len(args))
		if err != nil {
			return err
		}
	}

	_, err := fmt.Fprintf(&query, " WHERE shader_id = $%d AND created_by = $%d RETURNING 1", len(args)+1, len(args)+2)
	args = append(args, shaderId, userId)

	rows, err := pgPool.Query(ctx, query.String(), args...)
	if err != nil {
		return fmt.Errorf("failed updating shader caused by: %w", err)
	}

	if _, err = pgx.CollectExactlyOneRow(rows, pgx.RowTo[int]); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return infra.NotFoundError
		}
		return fmt.Errorf("failed updating shader caused by: %w", err)
	}

	return nil
}

func ShaderListByUser(ctx context.Context, pgPool *pgxpool.Pool, userId string) ([]*Shader, error) {
	rows, err := pgPool.Query(ctx,
		"SELECT shader_id, created_at, updated_at, name, visibility, description, COALESCE(forked_from, '') as forked_from, tags, content FROM shaders WHERE created_by = $1 ORDER BY updated_at DESC LIMIT 25", userId)
	if err != nil {
		return nil, fmt.Errorf("failed querying shaders by user caused by: %w", err)
	}

	shaders, err := pgx.CollectRows(rows, pgx.RowToAddrOfStructByName[Shader])
	if err != nil {
		return nil, fmt.Errorf("failed deserializing database rows caused by: %w", err)
	}

	return shaders, nil
}
