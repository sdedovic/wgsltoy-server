package repository

import (
	"context"
	"errors"
	"fmt"
	"github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype/zeronull"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/sdedovic/wgsltoy-server/src/go/infra"
	"time"
)

type ShaderInsertCommand struct {
	CreatedBy   string
	Visibility  string
	Name        string
	Description string
	Content     string
	Tags        []string
	ForkedFrom  string
}

type ShaderUpdateCommand struct {
	Name        string
	Description string
	Visibility  string
	Content     string
	Tags        []string
}

type zeronullShaderInfoDTO struct {
	Id        string    `db:"shader_id"`
	CreatedAt time.Time `db:"created_at"`
	UpdatedAt time.Time `db:"updated_at"`

	Name        string        `db:"name"`
	Visibility  string        `db:"visibility"`
	Description string        `db:"description"`
	ForkedFrom  zeronull.Text `db:"forked_from"`
	Tags        []string      `db:"tags"`
}

type ShaderInfoDTO struct {
	Id        string
	CreatedAt time.Time
	UpdatedAt time.Time

	Name        string
	Visibility  string
	Description string
	ForkedFrom  string
	Tags        []string
}

type zeronullShaderDTO struct {
	zeronullShaderInfoDTO
	content string `db:"content"`
}

type ShaderDTO struct {
	ShaderInfoDTO
	content string
}

func ShaderInsertOne(ctx context.Context, pgPool *pgxpool.Pool, shader *ShaderInsertCommand) (string, error) {
	now := time.Now()
	guid := infra.NewGUID()

	// tags is NOT NULL
	if shader.Tags == nil {
		shader.Tags = make([]string, 0)
	}

	sql, args, err := psql.
		Insert("shaders").
		Columns("created_at", "updated_at", "created_by", "visibility", "name", "description", "content", "tags", "forked_from", "shader_id").
		Values(now, now, shader.CreatedBy, shader.Visibility, shader.Name, shader.Description, shader.Content, shader.Tags, zeronull.Text(shader.ForkedFrom), guid).
		ToSql()
	if err != nil {
		return "", err
	}

	_, err = pgPool.Exec(ctx, sql, args...)
	if err != nil {
		return "", fmt.Errorf("failed inserting shader caused by: %w", err)
	}

	return guid, nil
}

func ShaderUpdateByCreatedByAndShaderId(ctx context.Context, pgPool *pgxpool.Pool, createdBy string, shaderId string, shader *ShaderUpdateCommand) error {
	now := time.Now()

	builder := psql.
		Update("shaders").
		Set("updated_at", now)

	if shader.Name != "" {
		builder = builder.Set("name", shader.Name)
	}

	if shader.Description != "" {
		builder = builder.Set("description", shader.Description)
	}

	if shader.Visibility != "" {
		builder.Set("visibility", shader.Visibility)
	}

	if shader.Content != "" {
		builder.Set("content", shader.Content)
	}

	if shader.Tags != nil {
		builder.Set("tags", shader.Tags)
	}

	sql, args, err := builder.
		Where(squirrel.Eq{"shader_id": shaderId, "created_by": createdBy}).
		Suffix("RETURNING 1").
		ToSql()

	rows, err := pgPool.Query(ctx, sql, args...)
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

func ShaderInfoFindAllByCreatedBy(ctx context.Context, pgPool *pgxpool.Pool, createdBy string) ([]*ShaderInfoDTO, error) {
	sql, args, err := psql.
		Select("shader_id", "created_at", "updated_at", "name", "visibility", "description", "forked_from", "tags").
		From("shaders").
		Where("created_by = ?", createdBy).
		OrderBy("updated_at DESC").
		Limit(100).
		ToSql()
	if err != nil {
		return nil, err
	}

	rows, err := pgPool.Query(ctx, sql, args...)
	if err != nil {
		return nil, fmt.Errorf("failed querying shaders by user caused by: %w", err)
	}

	shaderDTOs, err := pgx.CollectRows(rows, pgx.RowToAddrOfStructByName[zeronullShaderInfoDTO])
	if err != nil {
		return nil, fmt.Errorf("failed deserializing database rows caused by: %w", err)
	}

	shaders := make([]*ShaderInfoDTO, len(shaderDTOs))
	for idx, dto := range shaderDTOs {
		shaders[idx] = &ShaderInfoDTO{
			dto.Id,
			dto.CreatedAt,
			dto.UpdatedAt,
			dto.Name,
			dto.Visibility,
			dto.Description,
			string(dto.ForkedFrom),
			dto.Tags,
		}
	}

	return shaders, nil
}

const visibilityPublic = "public"
const visibilityUnlisted = "unlisted"

func ShaderFindOneById(ctx context.Context, pgPool *pgxpool.Pool, shaderId string) (*ShaderDTO, error) {
	sql, args, err := psql.
		Select("created_at", "updated_at", "created_by", "visibility", "name", "description", "content", "tags", "forked_from", "shader_id", "content").
		From("shaders").
		Where(squirrel.Eq{
			"shader_id":  shaderId,
			"visibility": []string{visibilityPublic, visibilityUnlisted},
		}).
		Limit(1).
		ToSql()
	if err != nil {
		return nil, err
	}

	rows, err := pgPool.Query(ctx, sql, args...)
	if err != nil {
		return nil, fmt.Errorf("failed querying shader caused by: %w", err)
	}
	shader, err := pgx.CollectExactlyOneRow(rows, pgx.RowToAddrOfStructByName[zeronullShaderDTO])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, infra.NotFoundError
		}
		return nil, fmt.Errorf("failed deserializing database rows caused by: %w", err)
	}

	return &ShaderDTO{
		ShaderInfoDTO{
			shader.Id,
			shader.CreatedAt,
			shader.UpdatedAt,
			shader.Name,
			shader.Visibility,
			shader.Description,
			string(shader.ForkedFrom),
			shader.Tags,
		},
		shader.content,
	}, nil
}

func ShaderFindOneByIdAndCreatedBy(ctx context.Context, pgPool *pgxpool.Pool, shaderId string, createdBy string) (*ShaderDTO, error) {
	sql, args, err := psql.
		Select("created_at", "updated_at", "created_by", "visibility", "name", "description", "content", "tags", "forked_from", "shader_id", "content").
		From("shaders").
		Where(squirrel.And{
			squirrel.Eq{"shader_id": shaderId},
			squirrel.Or{
				squirrel.Eq{"visibility": []string{visibilityPublic, visibilityUnlisted}},
				squirrel.Eq{"visibility": "private", "created_by": createdBy},
			},
		}).
		Limit(1).
		ToSql()
	if err != nil {
		return nil, err
	}

	rows, err := pgPool.Query(ctx, sql, args...)
	if err != nil {
		return nil, fmt.Errorf("failed querying shader caused by: %w", err)
	}
	shader, err := pgx.CollectExactlyOneRow(rows, pgx.RowToAddrOfStructByName[zeronullShaderDTO])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, infra.NotFoundError
		}
		return nil, fmt.Errorf("failed deserializing database rows caused by: %w", err)
	}

	return &ShaderDTO{
		ShaderInfoDTO{
			shader.Id,
			shader.CreatedAt,
			shader.UpdatedAt,
			shader.Name,
			shader.Visibility,
			shader.Description,
			string(shader.ForkedFrom),
			shader.Tags,
		},
		shader.content,
	}, nil
}
