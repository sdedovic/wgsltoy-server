package db

import (
	"context"
	"errors"
	"fmt"
	"github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/sdedovic/wgsltoy-server/src/go/guid"
	"github.com/sdedovic/wgsltoy-server/src/go/infra"
	"github.com/sdedovic/wgsltoy-server/src/go/models"
	"os"
	"time"
)

var psql = squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar)
var Storage DataStorage

const OperationTimeout = 40

type DataStorage struct {
	Client *pgxpool.Pool
	Ctx    context.Context
	Cancel context.CancelFunc
}

func CloseStorageDb(db DataStorage) {
	defer db.Cancel()
	defer func() {
		db.Client.Close()
	}()
}

func InitializeDataDbConnection() (DataStorage, error) {
	ctx, cancelFunc := context.WithTimeout(context.Background(), OperationTimeout*time.Second)
	defer cancelFunc()

	// TODO: replace with configs
	storageDbUrl := os.Getenv("DATABASE_URL")

	config, err := pgxpool.ParseConfig(storageDbUrl)
	if err != nil {
		return DataStorage{}, err
	}

	// TODO: handle TLS

	pool, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		return DataStorage{}, err
	}

	err = pool.Ping(ctx)
	if err != nil {
		return DataStorage{}, err
	}

	Storage = DataStorage{Client: pool, Ctx: ctx, Cancel: cancelFunc}
	return Storage, nil
}

func UserCreate(username string, email string, hashedPassword string) (models.User, error) {
	ctx, cancelFunc := context.WithTimeout(context.Background(), OperationTimeout*time.Second)
	defer cancelFunc()

	createdAt := time.Now()
	userId := guid.New()
	emailVerification := "pending"

	sql, args, err := psql.
		Insert("users").Columns("user_id", "username", "email", "email_verification", "password", "created_at", "updated_at").
		Values(userId, username, email, emailVerification, hashedPassword, createdAt, createdAt).
		ToSql()
	if err != nil {
		return models.User{}, fmt.Errorf("failed building sql caused by: %w", err)
	}

	_, err = Storage.Client.Exec(ctx, sql, args...)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {

			// uniqueness constraint violation
			if pgErr.Code == "23505" {
				if pgErr.ConstraintName == "unique_email" {
					return models.User{}, infra.NewValidationError("Email is already taken!")
				}

				if pgErr.ConstraintName == "unique_username" {
					return models.User{}, infra.NewValidationError("Username is already taken!")
				}
			}
		}

		// catchall
		return models.User{}, fmt.Errorf("failed inserting user caused by: %w", err)
	}

	return models.User{
		Id:                userId,
		CreatedAt:         createdAt,
		UpdatedAt:         createdAt,
		Username:          username,
		Email:             email,
		EmailVerification: emailVerification,
	}, nil
}

func UserGetByUsername(username string) (models.User, error) {
	ctx, cancelFunc := context.WithTimeout(context.Background(), OperationTimeout*time.Second)
	defer cancelFunc()

	sql, args, err := psql.Select("*").
		From("users").
		Where("username = ?", username).
		ToSql()
	if err != nil {
		return models.User{}, fmt.Errorf("failed building sql caused by: %w", err)
	}

	rows, _ := Storage.Client.Query(ctx, sql, args...)
	user, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[models.User])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return models.User{}, infra.BadLoginError
		}
		return models.User{}, fmt.Errorf("failed querying user caused by: %w", err)
	}

	return user, nil
}

func UserGetById(userId string) (models.User, error) {
	ctx, cancelFunc := context.WithTimeout(context.Background(), OperationTimeout*time.Second)
	defer cancelFunc()

	sql, args, err := psql.Select("*").
		From("users").
		Where("user_id = ?", userId).
		ToSql()
	if err != nil {
		return models.User{}, fmt.Errorf("failed building sql caused by: %w", err)
	}

	rows, _ := Storage.Client.Query(ctx, sql, args...)
	user, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[models.User])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return models.User{}, infra.BadLoginError
		}
		return models.User{}, fmt.Errorf("failed querying user caused by: %w", err)
	}

	return user, nil
}

func ShaderCreate(name string, visibility string, description string, tags []string, content string, createdBy string) (models.Shader, error) {
	ctx, cancelFunc := context.WithTimeout(context.Background(), OperationTimeout*time.Second)
	defer cancelFunc()

	createdAt := time.Now()
	shaderId := guid.New()

	sql, args, err := psql.
		Insert("shaders").
		Columns("created_at", "updated_at", "created_by", "visibility", "name", "description", "content", "tags", "shader_id").
		Values(createdAt, createdAt, createdBy, visibility, name, description, content, tags, shaderId).
		ToSql()
	if err != nil {
		return models.Shader{}, err
	}

	_, err = Storage.Client.Exec(ctx, sql, args...)
	if err != nil {
		return models.Shader{}, fmt.Errorf("failed inserting shader caused by: %w", err)
	}

	return models.Shader{
		Id:          shaderId,
		CreatedAt:   createdAt,
		UpdatedAt:   createdAt,
		Name:        name,
		Visibility:  visibility,
		Description: description,
		Tags:        tags,
	}, nil
}

func ShaderPartialUpdate(shaderId string, createdBy string, name *string, visibility *string, description *string, tags *[]string, content *string) (models.Shader, error) {
	ctx, cancelFunc := context.WithTimeout(context.Background(), OperationTimeout*time.Second)
	defer cancelFunc()

	updatedAt := time.Now()

	builder := psql.
		Update("shaders").
		Set("updated_at", updatedAt)

	if name != nil {
		builder = builder.Set("name", name)
	}

	if description != nil {
		builder = builder.Set("description", description)
	}

	if visibility != nil {
		builder = builder.Set("visibility", visibility)
	}

	if content != nil {
		builder = builder.Set("content", content)
	}

	// nil means do not change, empty means set to empty
	if tags != nil {
		builder = builder.Set("tags", tags)
	}

	sql, args, err := builder.
		Where(squirrel.Eq{"shader_id": shaderId, "created_by": createdBy}).
		Suffix("RETURNING *").
		ToSql()

	rows, err := Storage.Client.Query(ctx, sql, args...)
	if err != nil {
		return models.Shader{}, fmt.Errorf("failed updating shader caused by: %w", err)
	}

	shader, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[models.Shader])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return models.Shader{}, infra.NotFoundError
		}
		return models.Shader{}, fmt.Errorf("failed updating shader caused by: %w", err)
	}

	return shader, nil
}

func ShaderGetPubliclyVisibleById(shaderId string) (models.Shader, error) {
	ctx, cancelFunc := context.WithTimeout(context.Background(), OperationTimeout*time.Second)
	defer cancelFunc()

	sql, args, err := psql.
		Select("*").
		From("shaders").
		Where(squirrel.Eq{
			"shader_id":  shaderId,
			"visibility": []string{"public", "unlisted"},
		}).
		Limit(1).
		ToSql()
	if err != nil {
		return models.Shader{}, err
	}

	rows, _ := Storage.Client.Query(ctx, sql, args...)
	shader, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[models.Shader])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return models.Shader{}, infra.NotFoundError
		}
		return models.Shader{}, fmt.Errorf("failed deserializing database rows caused by: %w", err)
	}

	return shader, nil
}

func ShaderGetVisibleByIdAndLoggedInUser(shaderId string, currentUser string) (models.Shader, error) {
	ctx, cancelFunc := context.WithTimeout(context.Background(), OperationTimeout*time.Second)
	defer cancelFunc()

	sql, args, err := psql.
		Select("*").
		From("shaders").
		Where(squirrel.And{
			squirrel.Eq{"shader_id": shaderId},
			squirrel.Or{
				squirrel.Eq{"visibility": []string{"public", "unlisted"}},
				squirrel.Eq{"visibility": "private", "created_by": currentUser},
			},
		}).
		Limit(1).
		ToSql()
	if err != nil {
		return models.Shader{}, err
	}

	rows, _ := Storage.Client.Query(ctx, sql, args...)
	shader, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[models.Shader])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return models.Shader{}, infra.NotFoundError
		}
		return models.Shader{}, fmt.Errorf("failed deserializing database rows caused by: %w", err)
	}

	return shader, nil
}

func ShaderInfoListByCreatedBy(createdBy string) ([]models.ShaderInfo, error) {
	ctx, cancelFunc := context.WithTimeout(context.Background(), OperationTimeout*time.Second)
	defer cancelFunc()

	sql, args, err := psql.
		Select("shader_id", "created_at", "updated_at", "created_by", "name", "visibility", "description", "tags").
		From("shaders").
		Where("created_by = ?", createdBy).
		OrderBy("updated_at DESC").
		Limit(100).
		ToSql()
	if err != nil {
		return nil, err
	}

	rows, err := Storage.Client.Query(ctx, sql, args...)
	if err != nil {
		return nil, fmt.Errorf("failed querying shaders by user caused by: %w", err)
	}

	shaders, err := pgx.CollectRows(rows, pgx.RowToStructByName[models.ShaderInfo])
	if err != nil {
		return nil, fmt.Errorf("failed deserializing database rows caused by: %w", err)
	}

	return shaders, nil
}
