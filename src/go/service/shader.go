package service

import (
	"context"
	"fmt"
	"github.com/sdedovic/wgsltoy-server/src/go/db"
	"github.com/sdedovic/wgsltoy-server/src/go/infra"
	"github.com/sdedovic/wgsltoy-server/src/go/models"
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
		return infra.NewValidationError("Field 'name' may not be empty!")
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
		return infra.NewValidationError("Field 'visibility' may not be empty!")
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

func ShaderCreate(ctx context.Context, shader models.ShaderCreate) (string, error) {
	userInfo := ExtractUserInfoFromContext(ctx)
	if userInfo == nil {
		return "", infra.UnauthorizedError
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

	storedShader, err := db.ShaderCreate(shader.Name, shader.Visibility, shader.Description, shader.Tags, shader.Content, userInfo.Id)
	if err != nil {
		return "", err
	}

	return storedShader.Id, nil
}

func ShaderUpdate(ctx context.Context, shaderId string, shader models.ShaderPartialUpdate) (models.Shader, error) {
	userInfo := ExtractUserInfoFromContext(ctx)
	if userInfo == nil {
		return models.Shader{}, infra.UnauthorizedError
	}

	var query strings.Builder
	args := make([]any, 0, 8)

	query.WriteString("UPDATE shaders SET updated_at = $1 ")
	args = append(args, time.Now())

	if shader.Name != nil {
		if err := validateShaderName(*shader.Name); err != nil {
			return models.Shader{}, err
		}
	}

	if shader.Visibility != nil {
		if err := validateShaderVisibility(*shader.Visibility); err != nil {
			return models.Shader{}, err
		}
	}

	if shader.Description != nil {
		if err := validateShaderDescription(*shader.Description); err != nil {
			return models.Shader{}, err
		}
	}

	if shader.Content != nil {
		if err := validateShaderContent(*shader.Content); err != nil {
			return models.Shader{}, err
		}
	}

	if shader.Tags != nil {
		if err := validateShaderTags(*shader.Tags); err != nil {
			return models.Shader{}, err
		}
	}

	updatedShader, err := db.ShaderPartialUpdate(shaderId, userInfo.Id, shader.Name, shader.Visibility, shader.Description, shader.Tags, shader.Content)
	if err != nil {
		return models.Shader{}, err
	}

	return updatedShader, nil
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

func ShaderInfoListCurrentUser(ctx context.Context) ([]models.ShaderInfo, error) {
	userInfo := ExtractUserInfoFromContext(ctx)
	if userInfo == nil {
		return nil, infra.UnauthorizedError
	}

	return db.ShaderInfoListByCreatedBy(userInfo.Id)
}

func ShaderGet(ctx context.Context, shaderId string) (models.Shader, error) {
	userInfo := ExtractUserInfoFromContext(ctx)

	if userInfo == nil {
		return db.ShaderGetPubliclyVisibleById(shaderId)
	} else {
		return db.ShaderGetVisibleByIdAndLoggedInUser(shaderId, userInfo.Id)
	}
}
