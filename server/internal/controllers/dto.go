package controllers

import (
	"time"

	"github.com/sdedovic/wgsltoy-server/internal/domain"
)

// DTOs are the JSON bodies of requests and responses. Domain types are never encoded directly, a User
// carries its password hash.

// RFC 3339 to the second.
const timestampLayout = time.RFC3339

// timestamp is a time in JSON, always in UTC, e.g. 2026-11-24T19:54:18Z.
type timestamp time.Time

func (t timestamp) MarshalJSON() ([]byte, error) {
	return []byte(`"` + time.Time(t).UTC().Format(timestampLayout) + `"`), nil
}

//==== Requests ====\\

type loginDTO struct {
	Username string `json:"username"`
	Password string `json:"password"`
} // @name Login

type registerDTO struct {
	Username string `json:"username"`
	Email    string `json:"email"`
	Password string `json:"password"`
} // @name Register

type updateMeDTO struct {
	Username   string `json:"username"`
	Email      string `json:"email"`
	Visibility string `json:"visibility"`
} // @name UpdateMe

type shaderInputDTO struct {
	Name        string   `json:"name"`
	Visibility  string   `json:"visibility"`
	Description string   `json:"description"`
	Tags        []string `json:"tags"`
	Content     string   `json:"content"`
} // @name ShaderInput

//==== Responses ====\\

type healthDTO struct {
	Status string `json:"status"`
} // @name Health

type authorDTO struct {
	Username string `json:"username"`
} // @name Author

type shaderSummaryDTO struct {
	ID          string     `json:"id"`
	Author      *authorDTO `json:"author"`
	CreatedAt   timestamp  `json:"createdAt" format:"date-time" pattern:"^[0-9]{4}-[0-9]{2}-[0-9]{2}T[0-9]{2}:[0-9]{2}:[0-9]{2}Z$"`
	UpdatedAt   timestamp  `json:"updatedAt" format:"date-time" pattern:"^[0-9]{4}-[0-9]{2}-[0-9]{2}T[0-9]{2}:[0-9]{2}:[0-9]{2}Z$"`
	Name        string     `json:"name"`
	Visibility  string     `json:"visibility"`
	Description string     `json:"description"`
	Tags        []string   `json:"tags"`
} // @name ShaderSummary

type shaderDTO struct {
	shaderSummaryDTO
	Content string `json:"content"`
} // @name Shader

type galleryDTO struct {
	Shaders []shaderSummaryDTO `json:"shaders"`
} // @name Gallery

type profileDTO struct {
	Username  string     `json:"username"`
	Private   bool       `json:"private"`
	CreatedAt *timestamp `json:"createdAt,omitempty" format:"date-time" pattern:"^[0-9]{4}-[0-9]{2}-[0-9]{2}T[0-9]{2}:[0-9]{2}:[0-9]{2}Z$"`
} // @name Profile

type meDTO struct {
	Username          string    `json:"username"`
	Email             string    `json:"email"`
	EmailVerification string    `json:"emailVerification"`
	Visibility        string    `json:"visibility"`
	IsAdmin           bool      `json:"isAdmin"`
	CreatedAt         timestamp `json:"createdAt" format:"date-time" pattern:"^[0-9]{4}-[0-9]{2}-[0-9]{2}T[0-9]{2}:[0-9]{2}:[0-9]{2}Z$"`
	UpdatedAt         timestamp `json:"updatedAt" format:"date-time" pattern:"^[0-9]{4}-[0-9]{2}-[0-9]{2}T[0-9]{2}:[0-9]{2}:[0-9]{2}Z$"`
} // @name Me

func newShaderSummaryDTO(s domain.ShaderSummary) shaderSummaryDTO {
	var author *authorDTO
	if s.Author != nil {
		author = &authorDTO{Username: string(s.Author.Username)}
	}
	tags := make([]string, len(s.Tags))
	copy(tags, s.Tags)
	return shaderSummaryDTO{
		ID:          string(s.ID),
		Author:      author,
		CreatedAt:   timestamp(s.CreatedAt),
		UpdatedAt:   timestamp(s.UpdatedAt),
		Name:        string(s.Name),
		Visibility:  string(s.Visibility),
		Description: string(s.Description),
		Tags:        tags,
	}
}

func newShaderDTO(s domain.Shader) shaderDTO {
	return shaderDTO{newShaderSummaryDTO(s.ShaderSummary), string(s.Content)}
}

func newGalleryDTO(shaders []domain.ShaderSummary) galleryDTO {
	dtos := make([]shaderSummaryDTO, len(shaders))
	for i, s := range shaders {
		dtos[i] = newShaderSummaryDTO(s)
	}
	return galleryDTO{dtos}
}

func newProfileDTO(p domain.Profile) profileDTO {
	var createdAt *timestamp
	if p.CreatedAt != nil {
		t := timestamp(*p.CreatedAt)
		createdAt = &t
	}
	return profileDTO{Username: string(p.Username), Private: p.Private, CreatedAt: createdAt}
}

func newMeDTO(u domain.User) meDTO {
	return meDTO{
		Username:          string(u.Username),
		Email:             string(u.Email),
		EmailVerification: string(u.EmailVerification),
		Visibility:        string(u.Visibility),
		IsAdmin:           u.IsAdmin,
		CreatedAt:         timestamp(u.CreatedAt),
		UpdatedAt:         timestamp(u.UpdatedAt),
	}
}
