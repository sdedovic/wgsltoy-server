package domain

import (
	"context"
	"errors"
	"fmt"
	"time"
)

type ShaderDisposition string

const (
	ShaderDispositionDefault   ShaderDisposition = "default"
	ShaderDispositionRemoveAll ShaderDisposition = "remove_all"
	ShaderDispositionDonateAll ShaderDisposition = "donate_all"
)

func NewShaderDisposition(s string) (ShaderDisposition, error) {
	switch ShaderDisposition(s) {
	case ShaderDispositionDefault, ShaderDispositionRemoveAll, ShaderDispositionDonateAll:
		return ShaderDisposition(s), nil
	case "":
		return ShaderDispositionDefault, nil
	default:
		return "", NewValidationError("Field 'shaderDisposition' must be one of 'default', 'remove_all', or 'donate_all'!")
	}
}

type UserService struct {
	repo     UserRepository
	shaders  ShaderRepository
	tx       Transactor
	hasher   PasswordHasher
	sessions SessionRepository
}

func NewUserService(repo UserRepository, shaders ShaderRepository, tx Transactor, hasher PasswordHasher, sessions SessionRepository) *UserService {
	return &UserService{repo, shaders, tx, hasher, sessions}
}

type UpdateUserCommand struct {
	Username   string
	Email      string
	Visibility string
}

func (s *UserService) Register(ctx context.Context, username string, email string, password string) error {
	validUsername, err := NewUsername(username)
	if err != nil {
		return err
	}
	validEmail, err := NewEmail(email)
	if err != nil {
		return err
	}
	validPassword, err := NewPassword(password)
	if err != nil {
		return err
	}

	hashedPassword, err := s.hasher.Hash(validPassword)
	if err != nil {
		return fmt.Errorf("failed password hashing caused by: %w", err)
	}

	createdAt := time.Now()
	return s.repo.Save(ctx, User{
		ID:                GenerateUserID(),
		CreatedAt:         createdAt,
		UpdatedAt:         createdAt,
		Username:          validUsername,
		Email:             validEmail,
		EmailVerification: EmailVerificationPending,
		PasswordHash:      hashedPassword,
		Visibility:        VisibilityPublic,
	})
}

func (s *UserService) Login(ctx context.Context, username string, password string) (Session, error) {
	if len(username) == 0 {
		return Session{}, NewValidationError("Field 'username' is required!")
	}
	if len(password) == 0 {
		return Session{}, NewValidationError("Field 'password' is required!")
	}

	user, err := s.repo.FindByUsername(ctx, Username(username))
	if err != nil {
		if isNotFound(err) {
			return Session{}, ErrBadLogin
		}
		return Session{}, err
	}

	isMatch, err := s.hasher.Verify(Password(password), user.PasswordHash)
	if err != nil {
		return Session{}, fmt.Errorf("failed verifying user password: %w", err)
	}
	if !isMatch {
		return Session{}, ErrBadLogin
	}

	session, err := NewSession(user.ID)
	if err != nil {
		return Session{}, err
	}
	if err := s.sessions.Save(ctx, session); err != nil {
		return Session{}, err
	}
	return session, nil
}

func (s *UserService) Logout(ctx context.Context, sessionID SessionID) error {
	return s.sessions.DeleteByID(ctx, sessionID)
}

func (s *UserService) GetCurrent(ctx context.Context) (User, error) {
	principal, ok := PrincipalFromContext(ctx)
	if !ok {
		return User{}, ErrUnauthorized
	}
	return principal.User, nil
}

func (s *UserService) GetProfile(ctx context.Context, username string) (Profile, error) {
	if len(username) == 0 {
		return Profile{}, ErrNotFound
	}

	user, err := s.repo.FindByUsername(ctx, Username(username))
	if err != nil {
		return Profile{}, err
	}

	principal, _ := PrincipalFromContext(ctx)
	return user.ProfileFor(principal), nil
}

func (s *UserService) Update(ctx context.Context, cmd UpdateUserCommand) (User, error) {
	principal, ok := PrincipalFromContext(ctx)
	if !ok {
		return User{}, ErrUnauthorized
	}

	username, err := NewUsername(cmd.Username)
	if err != nil {
		return User{}, err
	}
	email, err := NewEmail(cmd.Email)
	if err != nil {
		return User{}, err
	}
	visibility, err := NewUserVisibility(cmd.Visibility)
	if err != nil {
		return User{}, err
	}

	user := principal.User
	user.Apply(UserChanges{
		Username:   username,
		Email:      email,
		Visibility: visibility,
	}, time.Now())

	if err := s.repo.Save(ctx, user); err != nil {
		return User{}, err
	}
	return user, nil
}

func (s *UserService) Delete(ctx context.Context, disposition ShaderDisposition) error {
	principal, ok := PrincipalFromContext(ctx)
	if !ok {
		return ErrUnauthorized
	}

	return s.tx.InTransaction(ctx, func(ctx context.Context) error {
		switch disposition {
		case ShaderDispositionRemoveAll:
			if err := s.shaders.DeleteByOwner(ctx, principal.User.ID); err != nil {
				return err
			}
		case ShaderDispositionDefault:
			if err := s.shaders.DeletePrivateByOwner(ctx, principal.User.ID); err != nil {
				return err
			}
		case ShaderDispositionDonateAll:
			// ON DELETE SET NULL handles orphaning when user row is deleted
		}
		return s.repo.DeleteByID(ctx, principal.User.ID)
	})
}

func isNotFound(err error) bool {
	return errors.Is(err, ErrNotFound)
}
