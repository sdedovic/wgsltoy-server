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

type UpdateShaderCommand struct {
	Name        string
	Visibility  string
	Description string
	Tags        []string
	Content     string
}

func (s *ShaderService) Create(ctx context.Context, cmd CreateShaderCommand) (Shader, error) {
	principal, ok := PrincipalFromContext(ctx)
	if !ok {
		return Shader{}, ErrUnauthorized
	}

	name, err := NewShaderName(cmd.Name)
	if err != nil {
		return Shader{}, err
	}
	visibility, err := NewVisibility(cmd.Visibility)
	if err != nil {
		return Shader{}, err
	}
	description, err := NewShaderDescription(cmd.Description)
	if err != nil {
		return Shader{}, err
	}
	content, err := NewShaderContent(cmd.Content)
	if err != nil {
		return Shader{}, err
	}
	tags, err := NewShaderTags(cmd.Tags)
	if err != nil {
		return Shader{}, err
	}

	createdAt := time.Now()
	shader := Shader{
		ShaderSummary: ShaderSummary{
			ID:          GenerateShaderID(),
			Author:      &Author{ID: principal.User.ID, Username: principal.User.Username},
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
		return Shader{}, err
	}
	return shader, nil
}

func (s *ShaderService) Get(ctx context.Context, id ShaderID) (Shader, error) {
	shader, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return Shader{}, err
	}

	principal, _ := PrincipalFromContext(ctx)
	if !shader.IsVisibleTo(principal) {
		return Shader{}, ErrNotFound
	}
	return shader, nil
}

func (s *ShaderService) Update(ctx context.Context, id ShaderID, cmd UpdateShaderCommand) (Shader, error) {
	principal, ok := PrincipalFromContext(ctx)
	if !ok {
		return Shader{}, ErrUnauthorized
	}

	name, err := NewShaderName(cmd.Name)
	if err != nil {
		return Shader{}, err
	}
	visibility, err := NewVisibility(cmd.Visibility)
	if err != nil {
		return Shader{}, err
	}
	description, err := NewShaderDescription(cmd.Description)
	if err != nil {
		return Shader{}, err
	}
	content, err := NewShaderContent(cmd.Content)
	if err != nil {
		return Shader{}, err
	}
	tags, err := NewShaderTags(cmd.Tags)
	if err != nil {
		return Shader{}, err
	}

	shader, err := s.repo.FindByIDAndOwner(ctx, id, principal.User.ID)
	if err != nil {
		return Shader{}, err
	}

	shader.Apply(ShaderChanges{
		Name:        name,
		Visibility:  visibility,
		Description: description,
		Tags:        tags,
		Content:     content,
	}, time.Now())
	if err := s.repo.Save(ctx, shader); err != nil {
		return Shader{}, err
	}
	return shader, nil
}

func (s *ShaderService) Delete(ctx context.Context, id ShaderID) error {
	principal, ok := PrincipalFromContext(ctx)
	if !ok {
		return ErrUnauthorized
	}
	if principal.User.IsAdmin {
		return s.repo.DeleteByID(ctx, id)
	}
	return s.repo.DeleteByIDAndOwner(ctx, id, principal.User.ID)
}

func (s *ShaderService) ListMine(ctx context.Context) ([]ShaderSummary, error) {
	principal, ok := PrincipalFromContext(ctx)
	if !ok {
		return nil, ErrUnauthorized
	}
	return s.repo.FindTop100ByCreatedByOrderByUpdatedAtDesc(ctx, principal.User.ID)
}

func (s *ShaderService) ListPublic(ctx context.Context) ([]ShaderSummary, error) {
	return s.repo.FindTop100PublicOrderByUpdatedAtDesc(ctx)
}

// ListRandomPublic picks up to n public shaders at random, with their content so they can be previewed.
func (s *ShaderService) ListRandomPublic(ctx context.Context, n int) ([]Shader, error) {
	return s.repo.FindRandomPublic(ctx, n)
}
