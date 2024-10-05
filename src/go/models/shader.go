package models

import "time"

type ShaderCreate struct {
	Name        string   `json:"name"`
	Visibility  string   `json:"visibility"`
	Description string   `json:"description"`
	Tags        []string `json:"tags"`
	Content     string   `json:"content"`
}

type ShaderPartialUpdate struct {
	Name        *string   `json:"name"`
	Visibility  *string   `json:"visibility"`
	Description *string   `json:"description"`
	Tags        *[]string `json:"tags"`
	Content     *string   `json:"content"`
}

// ShaderInfo represents just the information about a shader, excluding the actual code to keep it lighter.
type ShaderInfo struct {
	Id        string    `json:"id" db:"shader_id"`
	Location  string    `json:"location" db:"-"`
	CreatedBy string    `json:"createdBy" db:"created_by"`
	CreatedAt time.Time `json:"createdAt" db:"created_at"`
	UpdatedAt time.Time `json:"updatedAt" db:"updated_at"`

	Name        string   `json:"name" db:"name"`
	Visibility  string   `json:"visibility" db:"visibility"`
	Description string   `json:"description" db:"description"`
	Tags        []string `json:"tags" db:"tags"`
}

// Shader represents the information about a shader and the code.
type Shader struct {
	Id        string    `json:"id" db:"shader_id"`
	Location  string    `json:"location" db:"-"`
	CreatedBy string    `json:"createdBy" db:"created_by"`
	CreatedAt time.Time `json:"createdAt" db:"created_at"`
	UpdatedAt time.Time `json:"updatedAt" db:"updated_at"`

	Name        string   `json:"name" db:"name"`
	Visibility  string   `json:"visibility" db:"visibility"`
	Description string   `json:"description" db:"description"`
	Tags        []string `json:"tags" db:"tags"`
	Content     string   `json:"content" db:"content"`
}
