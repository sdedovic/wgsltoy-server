package domain

import (
	"context"
	"errors"
	"fmt"
	"time"
)

type UserService struct {
	repo   UserRepository
	hasher PasswordHasher
	tokens TokenIssuer
}

func NewUserService(repo UserRepository, hasher PasswordHasher, tokens TokenIssuer) *UserService {
	return &UserService{repo, hasher, tokens}
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
	})
}

func (s *UserService) Login(ctx context.Context, username string, password string) (string, error) {
	if len(username) == 0 {
		return "", NewValidationError("Field 'username' is required!")
	}

	if len(password) == 0 {
		return "", NewValidationError("Field 'password' is required!")
	}

	user, err := s.repo.FindByUsername(ctx, Username(username))
	if errors.Is(err, ErrNotFound) {
		return "", ErrBadLogin
	}
	if err != nil {
		return "", err
	}

	isMatch, err := s.hasher.Verify(Password(password), user.PasswordHash)
	if err != nil {
		return "", fmt.Errorf("failed verifying user password: %w", err)
	}
	if !isMatch {
		return "", ErrBadLogin
	}

	return s.tokens.Issue(Identity{UserID: user.ID})
}

func (s *UserService) GetCurrent(ctx context.Context) (User, error) {
	identity, ok := IdentityFromContext(ctx)
	if !ok {
		return User{}, ErrUnauthorized
	}

	return s.repo.FindByID(ctx, identity.UserID)
}
