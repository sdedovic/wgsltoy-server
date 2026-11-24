package domain

import (
	"fmt"
	"regexp"
	"time"
	"unicode/utf8"
	"uuid"
)

//==== ShaderID ====\\

type ShaderID string

func NewShaderID(s string) (ShaderID, error) {
	id, err := uuid.Parse(s)
	if err != nil {
		return "", NewValidationError("Supplied shader id is not valid!")
	}
	return ShaderID(id.String()), nil
}

func GenerateShaderID() ShaderID {
	return ShaderID(uuid.New().String())
}

//==== Fields ====\\

var tagRegex = regexp.MustCompile(`^[a-z][a-z0-9]+$`)
var displayRegex = regexp.MustCompile(`^[\pL\pM\pN\pP\pS ]+$`)
var displayMultilineRegex = regexp.MustCompile(`^[\pL\pM\pN\pP\pS\s]+$`)

type Visibility string

const (
	VisibilityPrivate  Visibility = "private"
	VisibilityUnlisted Visibility = "unlisted"
	VisibilityPublic   Visibility = "public"
)

func NewVisibility(s string) (Visibility, error) {
	switch Visibility(s) {
	case VisibilityPrivate, VisibilityUnlisted, VisibilityPublic:
		return Visibility(s), nil
	case "":
		return "", NewValidationError("Field 'visibility' may not be empty!")
	default:
		return "", NewValidationError("Field 'visibility' must be one of 'private', 'unlisted' or 'public'!")
	}
}

type ShaderName string

func NewShaderName(s string) (ShaderName, error) {
	if s == "" {
		return "", NewValidationError("Field 'name' may not be empty!")
	}
	if utf8.RuneCountInString(s) > 160 {
		return "", NewValidationError("Field 'name' is too long!")
	}
	if !displayRegex.MatchString(s) {
		return "", NewValidationError("Field 'name' contains invalid characters!")
	}
	return ShaderName(s), nil
}

type ShaderDescription string

func NewShaderDescription(s string) (ShaderDescription, error) {
	if utf8.RuneCountInString(s) > 480 {
		return "", NewValidationError("Field 'description' is too long!")
	}
	if s != "" && !displayMultilineRegex.MatchString(s) {
		return "", NewValidationError("Field 'description' contains invalid characters!")
	}
	return ShaderDescription(s), nil
}

type ShaderContent string

func NewShaderContent(s string) (ShaderContent, error) {
	if utf8.RuneCountInString(s) > 5250 {
		return "", NewValidationError("Field 'content' is too long!")
	}
	if s != "" && !displayMultilineRegex.MatchString(s) {
		return "", NewValidationError("Field 'content' contains invalid characters!")
	}
	return ShaderContent(s), nil
}

type ShaderTags []string

func NewShaderTags(tags []string) (ShaderTags, error) {
	for idx, tag := range tags {
		if tag == "" {
			return nil, NewValidationError(fmt.Sprintf("Field 'tags[%d]' is empty!", idx))
		}
		if utf8.RuneCountInString(tag) < 3 {
			return nil, NewValidationError(fmt.Sprintf("Field 'tags[%d]' is too short!", idx))
		}
		if utf8.RuneCountInString(tag) > 10 {
			return nil, NewValidationError(fmt.Sprintf("Field 'tags[%d]' is too long!", idx))
		}
		if !tagRegex.MatchString(tag) {
			return nil, NewValidationError(fmt.Sprintf("Field 'tags[%d]' contains invalid characters!", idx))
		}
	}
	return append(ShaderTags{}, tags...), nil
}

//==== Shader ====\\

type Author struct {
	ID       UserID
	Username Username
}

// Listings leave out Content, which can be large.
type ShaderSummary struct {
	ID          ShaderID
	Author      *Author
	CreatedAt   time.Time
	UpdatedAt   time.Time
	Name        ShaderName
	Visibility  Visibility
	Description ShaderDescription
	Tags        ShaderTags
}

// An anonymous principal has an empty User.ID, so only sees public and unlisted shaders.
func (s ShaderSummary) IsVisibleTo(p Principal) bool {
	if p.User.IsAdmin {
		return true
	}
	if s.Visibility != VisibilityPrivate {
		return true
	}
	return s.Author != nil && s.Author.ID == p.User.ID
}

func (s ShaderSummary) IsEditableBy(p Principal) bool {
	if p.User.IsAdmin {
		return true
	}
	return s.Author != nil && s.Author.ID == p.User.ID
}

type Shader struct {
	ShaderSummary
	Content ShaderContent
}

type ShaderChanges struct {
	Name        ShaderName
	Visibility  Visibility
	Description ShaderDescription
	Tags        ShaderTags
	Content     ShaderContent
}

func (s *Shader) Apply(changes ShaderChanges, updatedAt time.Time) {
	s.Name = changes.Name
	s.Visibility = changes.Visibility
	s.Description = changes.Description
	s.Tags = changes.Tags
	s.Content = changes.Content
	s.UpdatedAt = updatedAt
}
