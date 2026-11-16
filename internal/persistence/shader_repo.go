package persistence

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v5"
	"github.com/sdedovic/wgsltoy-server/internal/domain"
)

var shaderColumns = []string{"shader_id", "created_by", "created_at", "updated_at", "name", "visibility", "description", "tags", "content"}
var shaderSummaryColumns = shaderColumns[:len(shaderColumns)-1]

type shaderRow struct {
	ID          string    `db:"shader_id"`
	CreatedBy   string    `db:"created_by"`
	CreatedAt   time.Time `db:"created_at"`
	UpdatedAt   time.Time `db:"updated_at"`
	Name        string    `db:"name"`
	Visibility  string    `db:"visibility"`
	Description string    `db:"description"`
	Tags        []string  `db:"tags"`
	Content     string    `db:"content"`
}

func (r shaderRow) toDomain() domain.Shader {
	return domain.Shader{
		ShaderSummary: domain.ShaderSummary{
			ID:          domain.ShaderID(r.ID),
			CreatedBy:   domain.UserID(r.CreatedBy),
			CreatedAt:   r.CreatedAt,
			UpdatedAt:   r.UpdatedAt,
			Name:        domain.ShaderName(r.Name),
			Visibility:  domain.Visibility(r.Visibility),
			Description: domain.ShaderDescription(r.Description),
			Tags:        domain.ShaderTags(r.Tags),
		},
		Content: domain.ShaderContent(r.Content),
	}
}

type shaderSummaryRow struct {
	ID          string    `db:"shader_id"`
	CreatedBy   string    `db:"created_by"`
	CreatedAt   time.Time `db:"created_at"`
	UpdatedAt   time.Time `db:"updated_at"`
	Name        string    `db:"name"`
	Visibility  string    `db:"visibility"`
	Description string    `db:"description"`
	Tags        []string  `db:"tags"`
}

func (r shaderSummaryRow) toDomain() domain.ShaderSummary {
	return domain.ShaderSummary{
		ID:          domain.ShaderID(r.ID),
		CreatedBy:   domain.UserID(r.CreatedBy),
		CreatedAt:   r.CreatedAt,
		UpdatedAt:   r.UpdatedAt,
		Name:        domain.ShaderName(r.Name),
		Visibility:  domain.Visibility(r.Visibility),
		Description: domain.ShaderDescription(r.Description),
		Tags:        domain.ShaderTags(r.Tags),
	}
}

type ShaderRepository struct {
	db PGXProvider
}

var _ domain.ShaderRepository = (*ShaderRepository)(nil)

func NewShaderRepository(db PGXProvider) *ShaderRepository {
	return &ShaderRepository{db}
}

func (r *ShaderRepository) Save(ctx context.Context, shader domain.Shader) error {
	ctx, cancel := context.WithTimeout(ctx, queryTimeout)
	defer cancel()

	sql, args, err := psql.
		Insert("shaders").
		Columns(shaderColumns...).
		Values(string(shader.ID), string(shader.CreatedBy), shader.CreatedAt, shader.UpdatedAt, string(shader.Name),
			string(shader.Visibility), string(shader.Description), []string(shader.Tags), string(shader.Content)).
		Suffix("ON CONFLICT (shader_id) DO UPDATE SET " +
			"updated_at = EXCLUDED.updated_at, name = EXCLUDED.name, visibility = EXCLUDED.visibility, " +
			"description = EXCLUDED.description, tags = EXCLUDED.tags, content = EXCLUDED.content").
		ToSql()
	if err != nil {
		return fmt.Errorf("failed building sql caused by: %w", err)
	}

	_, err = r.db.GetPGX(ctx).Exec(ctx, sql, args...)
	if err != nil {
		return fmt.Errorf("failed saving shader caused by: %w", err)
	}
	return nil
}

func (r *ShaderRepository) FindByID(ctx context.Context, id domain.ShaderID) (domain.Shader, error) {
	ctx, cancel := context.WithTimeout(ctx, queryTimeout)
	defer cancel()

	sql, args, err := psql.Select(shaderColumns...).From("shaders").Where(squirrel.Eq{"shader_id": string(id)}).ToSql()
	if err != nil {
		return domain.Shader{}, fmt.Errorf("failed building sql caused by: %w", err)
	}

	rows, err := r.db.GetPGX(ctx).Query(ctx, sql, args...)
	if err != nil {
		return domain.Shader{}, fmt.Errorf("failed querying shader caused by: %w", err)
	}

	row, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[shaderRow])
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.Shader{}, domain.ErrNotFound
	}
	if err != nil {
		return domain.Shader{}, fmt.Errorf("failed querying shader caused by: %w", err)
	}

	return row.toDomain(), nil
}

func (r *ShaderRepository) FindTop100ByCreatedByOrderByUpdatedAtDesc(ctx context.Context, createdBy domain.UserID) ([]domain.ShaderSummary, error) {
	ctx, cancel := context.WithTimeout(ctx, queryTimeout)
	defer cancel()

	sql, args, err := psql.
		Select(shaderSummaryColumns...).
		From("shaders").
		Where(squirrel.Eq{"created_by": string(createdBy)}).
		OrderBy("updated_at DESC").
		Limit(100).
		ToSql()
	if err != nil {
		return nil, fmt.Errorf("failed building sql caused by: %w", err)
	}

	rows, err := r.db.GetPGX(ctx).Query(ctx, sql, args...)
	if err != nil {
		return nil, fmt.Errorf("failed querying shaders caused by: %w", err)
	}

	shaderRows, err := pgx.CollectRows(rows, pgx.RowToStructByName[shaderSummaryRow])
	if err != nil {
		return nil, fmt.Errorf("failed querying shaders caused by: %w", err)
	}

	summaries := make([]domain.ShaderSummary, len(shaderRows))
	for i, row := range shaderRows {
		summaries[i] = row.toDomain()
	}
	return summaries, nil
}
