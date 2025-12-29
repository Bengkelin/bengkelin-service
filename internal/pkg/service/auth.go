package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/Bengkelin/bengkelin-service/internal/pkg/dto"
	"github.com/Bengkelin/bengkelin-service/internal/pkg/models"
	"github.com/Bengkelin/bengkelin-service/internal/pkg/repository"
	"github.com/Bengkelin/bengkelin-service/pkg/crypto"
	"github.com/Bengkelin/bengkelin-service/pkg/helpers"
	applog "github.com/Bengkelin/bengkelin-service/pkg/log"
	"gorm.io/gorm"
)

type AuthService struct {
	userRepo         repository.UserRepositoryInterface
	mitraRepo        repository.MitraRepositoryInterface
	refreshTokenRepo repository.RefreshTokenRepositoryInterface
	jwtHelper        crypto.JWTCryptoHelper
	passwordHelper   crypto.PasswordCryptoHelper
}

func NewAuthService(deps ServiceDependencies) AuthServiceInterface {
	return &AuthService{
		userRepo:         deps.UserRepo,
		mitraRepo:        deps.MitraRepo,
		refreshTokenRepo: deps.RefreshTokenRepo,
		jwtHelper:        deps.JWTHelper,
		passwordHelper:   deps.PasswordHelper,
	}
}

// User authentication methods

func (s *AuthService) LoginUser(ctx context.Context, req dto.LoginRequest) (*dto.AuthResponse, error) {
	applog.InfoCtx(ctx, "User login attempt", "email", req.Email)
	
	user, err := s.userRepo.FindUserByEmail(req.Email)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			applog.InfoCtx(ctx, "User login failed - user not found", "email", req.Email)
			return nil, errors.New("invalid credentials")
		}
		applog.LogErrorCtx(ctx, err, "User login failed - database error", "email", req.Email)
		return nil, fmt.Errorf("login failed: %w", err)
	}

	if !s.passwordHelper.ComparePassword(user.Password, []byte(req.Password)) {
		applog.InfoCtx(ctx, "User login failed - invalid password", "email", req.Email)
		return nil, errors.New("invalid credentials")
	}

	tokenPair, err := s.jwtHelper.GenerateTokenPair(user.ID)
	if err != nil {
		applog.LogErrorCtx(ctx, err, "User login failed - token generation error", "user_id", user.ID)
		return nil, fmt.Errorf("failed to generate tokens: %w", err)
	}

	applog.InfoCtx(ctx, "User login successful", "user_id", user.ID, "email", req.Email)
	
	return &dto.AuthResponse{
		AccessToken:  tokenPair.AccessToken,
		RefreshToken: tokenPair.RefreshToken,
		ExpiresIn:    tokenPair.ExpiresIn,
		TokenType:    tokenPair.TokenType,
		User: &dto.UserInfo{
			ID:          user.ID,
			FirstName:   user.FirstName,
			LastName:    user.LastName,
			Email:       user.Email,
			PhoneNumber: user.PhoneNumber,
			AvatarURL:   user.AvatarUrl,
		},
	}, nil
}

func (s *AuthService) RegisterUser(ctx context.Context, req dto.RegisterUserRequest) (*dto.AuthResponse, error) {
	applog.InfoCtx(ctx, "User registration attempt", "email", req.Email)
	
	// Check if passwords match
	if req.Password != req.ConfirmPassword {
		return nil, errors.New("passwords do not match")
	}

	// Check if user already exists
	existingUser, err := s.userRepo.FindUserByEmail(req.Email)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		applog.LogErrorCtx(ctx, err, "User registration failed - database error", "email", req.Email)
		return nil, fmt.Errorf("registration failed: %w", err)
	}
	if existingUser != nil {
		applog.InfoCtx(ctx, "User registration failed - email already exists", "email", req.Email)
		return nil, errors.New("email already registered")
	}

	// Hash password
	hashedPassword, err := s.passwordHelper.HashAndSalt([]byte(req.Password))
	if err != nil {
		applog.LogErrorCtx(ctx, err, "User registration failed - password hashing error", "email", req.Email)
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	// Create user
	user := models.User{
		ID:          helpers.GenerateUUID(),
		FirstName:   req.FirstName,
		LastName:    req.LastName,
		Email:       req.Email,
		PhoneNumber: req.PhoneNumber,
		Password:    hashedPassword,
	}

	createdUser, err := s.userRepo.CreateUser(user)
	if err != nil {
		applog.LogErrorCtx(ctx, err, "User registration failed - create user error", "email", req.Email)
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	// Generate tokens
	tokenPair, err := s.jwtHelper.GenerateTokenPair(createdUser.ID)
	if err != nil {
		applog.LogErrorCtx(ctx, err, "User registration failed - token generation error", "user_id", createdUser.ID)
		return nil, fmt.Errorf("failed to generate tokens: %w", err)
	}

	applog.InfoCtx(ctx, "User registration successful", "user_id", createdUser.ID, "email", req.Email)
	
	return &dto.AuthResponse{
		AccessToken:  tokenPair.AccessToken,
		RefreshToken: tokenPair.RefreshToken,
		ExpiresIn:    tokenPair.ExpiresIn,
		TokenType:    tokenPair.TokenType,
		User: &dto.UserInfo{
			ID:          createdUser.ID,
			FirstName:   createdUser.FirstName,
			LastName:    createdUser.LastName,
			Email:       createdUser.Email,
			PhoneNumber: createdUser.PhoneNumber,
		},
	}, nil
}

func (s *AuthService) LoginUserWithGoogle(ctx context.Context, req dto.GoogleAuthRequest) (*dto.AuthResponse, error) {
	applog.InfoCtx(ctx, "User Google login attempt", "email", req.Email)
	
	user, err := s.userRepo.FindUserByEmail(req.Email)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		applog.LogErrorCtx(ctx, err, "User Google login failed - database error", "email", req.Email)
		return nil, fmt.Errorf("login failed: %w", err)
	}

	// Create user if doesn't exist
	if user == nil {
		user = &models.User{
			ID:        helpers.GenerateUUID(),
			FirstName: req.FirstName,
			Email:     req.Email,
		}
		
		createdUser, err := s.userRepo.CreateUser(*user)
		if err != nil {
			applog.LogErrorCtx(ctx, err, "User Google login failed - create user error", "email", req.Email)
			return nil, fmt.Errorf("failed to create user: %w", err)
		}
		user = &createdUser
		applog.InfoCtx(ctx, "New user created via Google login", "user_id", user.ID, "email", req.Email)
	}

	// Generate tokens
	tokenPair, err := s.jwtHelper.GenerateTokenPair(user.ID)
	if err != nil {
		applog.LogErrorCtx(ctx, err, "User Google login failed - token generation error", "user_id", user.ID)
		return nil, fmt.Errorf("failed to generate tokens: %w", err)
	}

	applog.InfoCtx(ctx, "User Google login successful", "user_id", user.ID, "email", req.Email)
	
	return &dto.AuthResponse{
		AccessToken:  tokenPair.AccessToken,
		RefreshToken: tokenPair.RefreshToken,
		ExpiresIn:    tokenPair.ExpiresIn,
		TokenType:    tokenPair.TokenType,
		User: &dto.UserInfo{
			ID:        user.ID,
			FirstName: user.FirstName,
			LastName:  user.LastName,
			Email:     user.Email,
			AvatarURL: user.AvatarUrl,
		},
	}, nil
}

// Mitra authentication methods

func (s *AuthService) LoginMitra(ctx context.Context, req dto.LoginRequest) (*dto.AuthResponse, error) {
	applog.InfoCtx(ctx, "Mitra login attempt", "email", req.Email)
	
	mitra, err := s.mitraRepo.FindMitraByEmail(req.Email)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			applog.InfoCtx(ctx, "Mitra login failed - mitra not found", "email", req.Email)
			return nil, errors.New("invalid credentials")
		}
		applog.LogErrorCtx(ctx, err, "Mitra login failed - database error", "email", req.Email)
		return nil, fmt.Errorf("login failed: %w", err)
	}

	if !s.passwordHelper.ComparePassword(mitra.Password, []byte(req.Password)) {
		applog.InfoCtx(ctx, "Mitra login failed - invalid password", "email", req.Email)
		return nil, errors.New("invalid credentials")
	}

	tokenPair, err := s.jwtHelper.GenerateTokenPairMitra(mitra.ID)
	if err != nil {
		applog.LogErrorCtx(ctx, err, "Mitra login failed - token generation error", "mitra_id", mitra.ID)
		return nil, fmt.Errorf("failed to generate tokens: %w", err)
	}

	applog.InfoCtx(ctx, "Mitra login successful", "mitra_id", mitra.ID, "email", req.Email)
	
	return &dto.AuthResponse{
		AccessToken:  tokenPair.AccessToken,
		RefreshToken: tokenPair.RefreshToken,
		ExpiresIn:    tokenPair.ExpiresIn,
		TokenType:    tokenPair.TokenType,
		Mitra: &dto.MitraInfo{
			ID:          mitra.ID,
			FirstName:   mitra.FirstName,
			LastName:    mitra.LastName,
			Email:       mitra.Email,
			PhoneNumber: mitra.PhoneNumber,
			BankName:    mitra.BankName,
			BankNumber:  mitra.BankNumber,
		},
	}, nil
}

func (s *AuthService) RegisterMitra(ctx context.Context, req dto.RegisterMitraRequest) (*dto.AuthResponse, error) {
	applog.InfoCtx(ctx, "Mitra registration attempt", "email", req.Email)
	
	// Check if passwords match
	if req.Password != req.ConfirmPassword {
		return nil, errors.New("passwords do not match")
	}

	// Check if mitra already exists
	existingMitra, err := s.mitraRepo.FindMitraByEmail(req.Email)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		applog.LogErrorCtx(ctx, err, "Mitra registration failed - database error", "email", req.Email)
		return nil, fmt.Errorf("registration failed: %w", err)
	}
	if existingMitra != nil {
		applog.InfoCtx(ctx, "Mitra registration failed - email already exists", "email", req.Email)
		return nil, errors.New("email already registered")
	}

	// Hash password
	hashedPassword, err := s.passwordHelper.HashAndSalt([]byte(req.Password))
	if err != nil {
		applog.LogErrorCtx(ctx, err, "Mitra registration failed - password hashing error", "email", req.Email)
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	// Create mitra
	mitra := models.Mitra{
		ID:          helpers.GenerateUUID(),
		FirstName:   req.FirstName,
		LastName:    req.LastName,
		Email:       req.Email,
		PhoneNumber: req.PhoneNumber,
		Password:    hashedPassword,
	}

	createdMitra, err := s.mitraRepo.CreateMitra(mitra)
	if err != nil {
		applog.LogErrorCtx(ctx, err, "Mitra registration failed - create mitra error", "email", req.Email)
		return nil, fmt.Errorf("failed to create mitra: %w", err)
	}

	// Generate tokens
	tokenPair, err := s.jwtHelper.GenerateTokenPairMitra(createdMitra.ID)
	if err != nil {
		applog.LogErrorCtx(ctx, err, "Mitra registration failed - token generation error", "mitra_id", createdMitra.ID)
		return nil, fmt.Errorf("failed to generate tokens: %w", err)
	}

	applog.InfoCtx(ctx, "Mitra registration successful", "mitra_id", createdMitra.ID, "email", req.Email)
	
	return &dto.AuthResponse{
		AccessToken:  tokenPair.AccessToken,
		RefreshToken: tokenPair.RefreshToken,
		ExpiresIn:    tokenPair.ExpiresIn,
		TokenType:    tokenPair.TokenType,
		Mitra: &dto.MitraInfo{
			ID:          createdMitra.ID,
			FirstName:   createdMitra.FirstName,
			LastName:    createdMitra.LastName,
			Email:       createdMitra.Email,
			PhoneNumber: createdMitra.PhoneNumber,
		},
	}, nil
}

func (s *AuthService) LoginMitraWithGoogle(ctx context.Context, req dto.GoogleAuthRequest) (*dto.AuthResponse, error) {
	applog.InfoCtx(ctx, "Mitra Google login attempt", "email", req.Email)
	
	mitra, err := s.mitraRepo.FindMitraByEmail(req.Email)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		applog.LogErrorCtx(ctx, err, "Mitra Google login failed - database error", "email", req.Email)
		return nil, fmt.Errorf("login failed: %w", err)
	}

	// Create mitra if doesn't exist
	if mitra == nil {
		mitra = &models.Mitra{
			ID:        helpers.GenerateUUID(),
			FirstName: req.FirstName,
			Email:     req.Email,
		}
		
		createdMitra, err := s.mitraRepo.CreateMitra(*mitra)
		if err != nil {
			applog.LogErrorCtx(ctx, err, "Mitra Google login failed - create mitra error", "email", req.Email)
			return nil, fmt.Errorf("failed to create mitra: %w", err)
		}
		mitra = &createdMitra
		applog.InfoCtx(ctx, "New mitra created via Google login", "mitra_id", mitra.ID, "email", req.Email)
	}

	// Generate tokens
	tokenPair, err := s.jwtHelper.GenerateTokenPairMitra(mitra.ID)
	if err != nil {
		applog.LogErrorCtx(ctx, err, "Mitra Google login failed - token generation error", "mitra_id", mitra.ID)
		return nil, fmt.Errorf("failed to generate tokens: %w", err)
	}

	applog.InfoCtx(ctx, "Mitra Google login successful", "mitra_id", mitra.ID, "email", req.Email)
	
	return &dto.AuthResponse{
		AccessToken:  tokenPair.AccessToken,
		RefreshToken: tokenPair.RefreshToken,
		ExpiresIn:    tokenPair.ExpiresIn,
		TokenType:    tokenPair.TokenType,
		Mitra: &dto.MitraInfo{
			ID:        mitra.ID,
			FirstName: mitra.FirstName,
			LastName:  mitra.LastName,
			Email:     mitra.Email,
		},
	}, nil
}

// Token management methods

func (s *AuthService) RefreshUserToken(ctx context.Context, refreshToken string) (*dto.AuthResponse, error) {
	applog.DebugCtx(ctx, "User token refresh attempt")
	
	tokenPair, err := s.jwtHelper.RefreshAccessToken(refreshToken)
	if err != nil {
		applog.InfoCtx(ctx, "User token refresh failed", "error", err.Error())
		return nil, fmt.Errorf("failed to refresh token: %w", err)
	}

	applog.DebugCtx(ctx, "User token refresh successful")
	
	return &dto.AuthResponse{
		AccessToken:  tokenPair.AccessToken,
		RefreshToken: tokenPair.RefreshToken,
		ExpiresIn:    tokenPair.ExpiresIn,
		TokenType:    tokenPair.TokenType,
	}, nil
}

func (s *AuthService) RefreshMitraToken(ctx context.Context, refreshToken string) (*dto.AuthResponse, error) {
	applog.DebugCtx(ctx, "Mitra token refresh attempt")
	
	tokenPair, err := s.jwtHelper.RefreshAccessTokenMitra(refreshToken)
	if err != nil {
		applog.InfoCtx(ctx, "Mitra token refresh failed", "error", err.Error())
		return nil, fmt.Errorf("failed to refresh token: %w", err)
	}

	applog.DebugCtx(ctx, "Mitra token refresh successful")
	
	return &dto.AuthResponse{
		AccessToken:  tokenPair.AccessToken,
		RefreshToken: tokenPair.RefreshToken,
		ExpiresIn:    tokenPair.ExpiresIn,
		TokenType:    tokenPair.TokenType,
	}, nil
}

func (s *AuthService) LogoutUser(ctx context.Context, refreshToken string) error {
	applog.DebugCtx(ctx, "User logout attempt")
	
	err := s.jwtHelper.RevokeRefreshToken(refreshToken)
	if err != nil {
		applog.LogErrorCtx(ctx, err, "User logout failed")
		return fmt.Errorf("failed to logout: %w", err)
	}

	applog.DebugCtx(ctx, "User logout successful")
	return nil
}

func (s *AuthService) LogoutMitra(ctx context.Context, refreshToken string) error {
	applog.DebugCtx(ctx, "Mitra logout attempt")
	
	err := s.jwtHelper.RevokeRefreshToken(refreshToken)
	if err != nil {
		applog.LogErrorCtx(ctx, err, "Mitra logout failed")
		return fmt.Errorf("failed to logout: %w", err)
	}

	applog.DebugCtx(ctx, "Mitra logout successful")
	return nil
}

func (s *AuthService) LogoutAllUserDevices(ctx context.Context, userID string) error {
	applog.InfoCtx(ctx, "User logout all devices", "user_id", userID)
	
	err := s.jwtHelper.RevokeAllUserTokens(userID)
	if err != nil {
		applog.LogErrorCtx(ctx, err, "User logout all devices failed", "user_id", userID)
		return fmt.Errorf("failed to logout from all devices: %w", err)
	}

	applog.InfoCtx(ctx, "User logout all devices successful", "user_id", userID)
	return nil
}

func (s *AuthService) LogoutAllMitraDevices(ctx context.Context, mitraID string) error {
	applog.InfoCtx(ctx, "Mitra logout all devices", "mitra_id", mitraID)
	
	err := s.jwtHelper.RevokeAllMitraTokens(mitraID)
	if err != nil {
		applog.LogErrorCtx(ctx, err, "Mitra logout all devices failed", "mitra_id", mitraID)
		return fmt.Errorf("failed to logout from all devices: %w", err)
	}

	applog.InfoCtx(ctx, "Mitra logout all devices successful", "mitra_id", mitraID)
	return nil
}
