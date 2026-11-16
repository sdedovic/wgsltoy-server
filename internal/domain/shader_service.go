package domain

import (
	"context"
	"time"
)

type ShaderService struct {
	repo ShaderRepository
}

func NewShaderService(repo ShaderRepository) *ShaderService {
	return &ShaderService{repo}
}

type CreateShaderCommand struct {
	Name        string
	Visibility  string
	Description string
	Tags        []string
	Content     string
}

// Nil fields are left unchanged.
type UpdateShaderCommand struct {
	Name        *string
	Visibility  *string
	Description *string
	Tags        *[]string
	Content     *string
}

func (s *ShaderService) Create(ctx context.Context, cmd CreateShaderCommand) (ShaderID, error) {
	identity, ok := IdentityFromContext(ctx)
	if !ok {
		return "", ErrUnauthorized
	}

	name, err := NewShaderName(cmd.Name)
	if err != nil {
		return "", err
	}

	visibility, err := NewVisibility(cmd.Visibility)
	if err != nil {
		return "", err
	}

	description, err := NewShaderDescription(cmd.Description)
	if err != nil {
		return "", err
	}

	content, err := NewShaderContent(cmd.Content)
	if err != nil {
		return "", err
	}

	tags, err := NewShaderTags(cmd.Tags)
	if err != nil {
		return "", err
	}

	createdAt := time.Now()
	shader := Shader{
		ShaderSummary: ShaderSummary{
			ID:          GenerateShaderID(),
			CreatedBy:   identity.UserID,
			CreatedAt:   createdAt,
			UpdatedAt:   createdAt,
			Name:        name,
			Visibility:  visibility,
			Description: description,
			Tags:        tags,
		},
		Content: content,
	}
	if err := s.repo.Save(ctx, shader); err != nil {
		return "", err
	}

	return shader.ID, nil
}

func (s *ShaderService) Update(ctx context.Context, id ShaderID, cmd UpdateShaderCommand) (Shader, error) {
	identity, ok := IdentityFromContext(ctx)
	if !ok {
		return Shader{}, ErrUnauthorized
	}

	var changes ShaderChanges
	var err error
	if changes.Name, err = optional(cmd.Name, NewShaderName); err != nil {
		return Shader{}, err
	}
	if changes.Visibility, err = optional(cmd.Visibility, NewVisibility); err != nil {
		return Shader{}, err
	}
	if changes.Description, err = optional(cmd.Description, NewShaderDescription); err != nil {
		return Shader{}, err
	}
	if changes.Content, err = optional(cmd.Content, NewShaderContent); err != nil {
		return Shader{}, err
	}
	if changes.Tags, err = optional(cmd.Tags, NewShaderTags); err != nil {
		return Shader{}, err
	}

	shader, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return Shader{}, err
	}
	// not found rather than forbidden, so other users' shader IDs aren't revealed
	if shader.CreatedBy != identity.UserID {
		return Shader{}, ErrNotFound
	}

	shader.Apply(changes, time.Now())
	if err := s.repo.Save(ctx, shader); err != nil {
		return Shader{}, err
	}
	return shader, nil
}

func (s *ShaderService) ListMine(ctx context.Context) ([]ShaderSummary, error) {
	identity, ok := IdentityFromContext(ctx)
	if !ok {
		return nil, ErrUnauthorized
	}

	return s.repo.FindTop100ByCreatedByOrderByUpdatedAtDesc(ctx, identity.UserID)
}

func (s *ShaderService) Get(ctx context.Context, id ShaderID) (Shader, error) {
	shader, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return Shader{}, err
	}

	identity, _ := IdentityFromContext(ctx)
	if !shader.IsVisibleTo(identity.UserID) {
		return Shader{}, ErrNotFound
	}
	return shader, nil
}

func optional[In, Out any](value *In, newFn func(In) (Out, error)) (*Out, error) {
	if value == nil {
		return nil, nil
	}
	out, err := newFn(*value)
	if err != nil {
		return nil, err
	}
	return &out, nil
}
