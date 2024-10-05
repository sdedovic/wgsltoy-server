package service

import (
	"context"
	"errors"
	"fmt"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/sdedovic/wgsltoy-server/src/go/infra"
	"github.com/sdedovic/wgsltoy-server/src/go/repository"
	"regexp"
	"strings"
	"time"
	"unicode/utf8"
)

var tagRegex = regexp.MustCompile(`^[a-z][a-z0-9]+$`)
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

type CreateShader struct {
	Name        string
	Visibility  string
	Description string
	Content     string
	Tags        []string
	ForkedFrom  string
}

func ShaderCreate(ctx context.Context, pgPool *pgxpool.Pool, shader *CreateShader) (string, error) {
	userInfo := ExtractUserInfoFromContext(ctx)
	if userInfo == nil {
		return "", infra.UnauthorizedError
	}

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

	if shader.ForkedFrom != "" && !infra.ValidateGUID(shader.ForkedFrom) {
		return "", infra.NewValidationError("Field 'forkedFrom' is invalid!")
	}

	guid, err := repository.ShaderInsertOne(ctx, pgPool, &repository.ShaderInsertCommand{
		CreatedBy:   userInfo.UserID(),
		Visibility:  shader.Visibility,
		Name:        shader.Name,
		Description: shader.Description,
		Content:     shader.Content,
		Tags:        shader.Tags,
		ForkedFrom:  shader.ForkedFrom,
	})
	if err != nil {
		return "", err
	}

	return guid, nil
}

type UpdateShader struct {
	Name        string
	Visibility  string
	Description string
	Tags        []string
	Content     string
}

func ShaderUpdate(ctx context.Context, pgPool *pgxpool.Pool, shaderId string, shader UpdateShader) error {
	userInfo := ExtractUserInfoFromContext(ctx)
	if userInfo == nil {
		return infra.UnauthorizedError
	}

	var query strings.Builder
	args := make([]any, 0, 8)

	query.WriteString("UPDATE shaders SET updated_at = $1 ")
	args = append(args, time.Now())

	if shader.Name != "" {
		if err := validateShaderName(shader.Name); err != nil {
			return err
		}
	}

	if shader.Visibility != "" {
		if err := validateShaderVisibility(shader.Visibility); err != nil {
			return err
		}
	}

	if shader.Description != "" {
		if err := validateShaderDescription(shader.Description); err != nil {
			return err
		}
	}

	if shader.Content != "" {
		if err := validateShaderContent(shader.Content); err != nil {
			return err
		}
	}

	if shader.Tags != nil {
		if err := validateShaderTags(shader.Tags); err != nil {
			return err
		}
	}

	err := repository.ShaderUpdateByCreatedByAndShaderId(ctx, pgPool, userInfo.UserID(), shaderId, &repository.ShaderUpdateCommand{
		Name:        shader.Name,
		Description: shader.Description,
		Visibility:  shader.Visibility,
		Content:     shader.Content,
		Tags:        shader.Tags,
	})
	if err != nil {
		return err
	}

	return nil
}

// ShaderInfo contains all information about a shader, commiting code (content) for a smaller network footprint
type ShaderInfo struct {
	Id        string
	CreatedAt time.Time
	UpdatedAt time.Time

	Name        string
	Visibility  string
	Description string
	ForkedFrom  string
	Tags        []string
}

func ShaderInfoListCurrentUser(ctx context.Context, pgPool *pgxpool.Pool) ([]*ShaderInfo, error) {
	userInfo := ExtractUserInfoFromContext(ctx)
	if userInfo == nil {
		return nil, infra.UnauthorizedError
	}

	all, err := repository.ShaderInfoFindAllByCreatedBy(ctx, pgPool, userInfo.UserID())
	if err != nil {
		return nil, err
	}

	var shaders = make([]*ShaderInfo, len(all))
	for idx, stored := range all {
		shaders[idx] = &ShaderInfo{
			stored.Id,
			stored.CreatedAt,
			stored.UpdatedAt,
			stored.Name,
			stored.Visibility,
			stored.Description,
			stored.ForkedFrom,
			stored.Tags,
		}
	}

	return shaders, nil
}

type Shader struct {
	Id        string
	CreatedAt time.Time
	UpdatedAt time.Time

	Name        string
	Visibility  string
	Description string
	ForkedFrom  string
	Tags        []string
	Content     string
}

func ShaderGetOneById(ctx context.Context, pgPool *pgxpool.Pool, shaderId string) (*Shader, error) {
	userInfo := ExtractUserInfoFromContext(ctx)

	var shader *repository.ShaderDTO
	var err error
	if userInfo == nil {
		shader, err = repository.ShaderFindOneById(ctx, pgPool, shaderId)
	} else {
		shader, err = repository.ShaderFindOneByIdAndCreatedBy(ctx, pgPool, shaderId, userInfo.UserID())
	}

	if err != nil {
		return nil, err
	}

	return &Shader{
		shader.Id,
		shader.CreatedAt,
		shader.UpdatedAt,
		shader.Name,
		shader.Visibility,
		shader.Description,
		shader.ForkedFrom,
		shader.Tags,
		shader.Content,
	}, nil
}
