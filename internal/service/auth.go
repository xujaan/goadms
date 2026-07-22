package service

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/jan/goadms/internal/config"
	"github.com/jan/goadms/internal/model"
	"github.com/jan/goadms/internal/repository"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrUserInactive       = errors.New("user is inactive")
	ErrInvalidToken       = errors.New("invalid token")
)

type AuthService struct {
	userRepo      *repository.AppUserRepo
	refreshRepo   *repository.RefreshTokenRepo
	accessTTL     time.Duration
	refreshTTL    time.Duration
	jwtSecret     []byte
}

func NewAuthService(userRepo *repository.AppUserRepo, refreshRepo *repository.RefreshTokenRepo, cfg config.AuthConfig) *AuthService {
	return &AuthService{
		userRepo:    userRepo,
		refreshRepo: refreshRepo,
		accessTTL:   cfg.AccessTTLDuration(),
		refreshTTL:  cfg.RefreshTTLDuration(),
		jwtSecret:   []byte(cfg.JWTSecret),
	}
}

type TokenPair struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int    `json:"expires_in"`
}

// Login verifies credentials and returns JWT token pair.
func (s *AuthService) Login(ctx context.Context, username, password string) (*TokenPair, *model.AppUser, error) {
	user, err := s.userRepo.GetByUsername(ctx, username)
	if err != nil || user == nil {
		return nil, nil, ErrInvalidCredentials
	}
	if !user.IsActive {
		return nil, nil, ErrUserInactive
	}
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return nil, nil, ErrInvalidCredentials
	}

	return s.generateTokens(ctx, user)
}

// Refresh validates refresh token and issues new token pair.
func (s *AuthService) Refresh(ctx context.Context, refreshToken string) (*TokenPair, *model.AppUser, error) {
	// We don't know the user from the token alone — try looking up by iterating.
	// Instead, we decode the token claims. For simplicity, store user ID in refresh token metadata.
	// Here we just try to validate against all user tokens.
	// Better approach: embed user ID in refresh token (stateless refresh).
	// For now, we iterate to find matching token.
	users, err := s.userRepo.List(ctx)
	if err != nil {
		return nil, nil, ErrInvalidToken
	}
	for _, u := range users {
		valid, _ := s.refreshRepo.ValidateAndDelete(ctx, u.ID, refreshToken)
		if valid {
			return s.generateTokens(ctx, &u)
		}
	}
	return nil, nil, ErrInvalidToken
}

// Logout invalidates all refresh tokens for the user.
func (s *AuthService) Logout(ctx context.Context, userID uuid.UUID) error {
	return s.refreshRepo.DeleteByUser(ctx, userID)
}

// GetUser returns user by ID.
func (s *AuthService) GetUser(ctx context.Context, userID uuid.UUID) (*model.AppUser, error) {
	return s.userRepo.GetByID(ctx, userID)
}

func (s *AuthService) generateTokens(ctx context.Context, user *model.AppUser) (*TokenPair, *model.AppUser, error) {
	now := time.Now()
	accessExp := now.Add(s.accessTTL)
	refreshExp := now.Add(s.refreshTTL)

	// Access token.
	accessClaims := jwt.MapClaims{
		"sub":  user.ID.String(),
		"role": user.Role,
		"iat":  now.Unix(),
		"exp":  accessExp.Unix(),
	}
	accessToken, err := jwt.NewWithClaims(jwt.SigningMethodHS256, accessClaims).SignedString(s.jwtSecret)
	if err != nil {
		return nil, nil, fmt.Errorf("sign access token: %w", err)
	}

	// Refresh token (opaque random string).
	refreshBytes := make([]byte, 32)
	rand.Read(refreshBytes)
	refreshToken := hex.EncodeToString(refreshBytes)

	if err := s.refreshRepo.Create(ctx, user.ID, refreshToken, refreshExp); err != nil {
		return nil, nil, fmt.Errorf("store refresh token: %w", err)
	}

	return &TokenPair{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresIn:    int(s.accessTTL.Seconds()),
	}, user, nil
}
