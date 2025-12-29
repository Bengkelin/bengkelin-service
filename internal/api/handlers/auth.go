package handlers

import (
	"github.com/Bengkelin/bengkelin-service/internal/api/middleware"
	"github.com/Bengkelin/bengkelin-service/internal/pkg/container"
	"github.com/Bengkelin/bengkelin-service/internal/pkg/dto"
	appErrors "github.com/Bengkelin/bengkelin-service/internal/pkg/errors"
	"github.com/Bengkelin/bengkelin-service/internal/pkg/service"
	"github.com/Bengkelin/bengkelin-service/internal/pkg/validator"
	applog "github.com/Bengkelin/bengkelin-service/pkg/log"
	"github.com/gin-gonic/gin"
)

// AuthHandler handles authentication-related requests
type AuthHandler struct {
	BaseHandler
	authService service.AuthServiceInterface
}

// AuthHandlerInterface defines the auth handler contract
type AuthHandlerInterface interface {
	// User authentication
	UsersAuthLogin(c *gin.Context)
	UsersAuthRegister(c *gin.Context)
	UsersAuthGoogle(c *gin.Context)
	
	// Mitra authentication
	MitrasAuthLogin(c *gin.Context)
	MitrasAuthRegister(c *gin.Context)
	MitrasAuthGoogle(c *gin.Context)
	
	// Token management
	UsersRefreshToken(c *gin.Context)
	MitrasRefreshToken(c *gin.Context)
	UsersLogout(c *gin.Context)
	MitrasLogout(c *gin.Context)
	UsersLogoutAll(c *gin.Context)
	MitrasLogoutAll(c *gin.Context)
	
	// Legacy methods (deprecated)
	UsersNewAddress(c *gin.Context)
	UsersNewVehicle(c *gin.Context)
	MitrasNewBank(c *gin.Context)
	MitrasUpdateBank(c *gin.Context)
	MitrasUpdateProfile(c *gin.Context)
}

// NewAuthHandler creates a new auth handler
func NewAuthHandler() AuthHandlerInterface {
	container := container.GetContainer()
	return &AuthHandler{
		authService: container.AuthService,
	}
}

// User Authentication Methods

// UsersAuthLogin godoc
// @Summary User login
// @Description Authenticate a user with email and password
// @Tags Auth
// @Accept json
// @Produce json
// @Param request body validator.LoginRequest true "Login credentials"
// @Success 200 {object} response.Response{data=dto.AuthResponse} "Login successful"
// @Failure 400 {object} response.Response "Invalid request"
// @Failure 401 {object} response.Response "Invalid credentials"
// @Failure 429 {object} response.Response "Too many requests"
// @Failure 500 {object} response.Response "Internal server error"
// @Router /api/v1/users/auth/login [post]
func (h *AuthHandler) UsersAuthLogin(c *gin.Context) {
	h.LogRequest(c, "UsersAuthLogin")
	
	var req validator.LoginRequest
	if !middleware.ValidateRequest(c, &req) {
		return
	}
	
	// Convert validator to DTO
	loginReq := dto.LoginRequest{
		Email:    req.Email,
		Password: req.Password,
	}
	
	authResp, err := h.authService.LoginUser(c.Request.Context(), loginReq)
	if err != nil {
		h.HandleError(c, err)
		return
	}
	
	h.HandleSuccess(c, "Login successful", authResp)
}

// UsersAuthRegister godoc
// @Summary User registration
// @Description Register a new user account
// @Tags Auth
// @Accept json
// @Produce json
// @Param request body validator.RegisterNewUserRequest true "Registration details"
// @Success 201 {object} response.Response{data=dto.AuthResponse} "Registration successful"
// @Failure 400 {object} response.Response "Invalid request"
// @Failure 409 {object} response.Response "User already exists"
// @Failure 429 {object} response.Response "Too many requests"
// @Failure 500 {object} response.Response "Internal server error"
// @Router /api/v1/users/auth/register [post]
func (h *AuthHandler) UsersAuthRegister(c *gin.Context) {
	h.LogRequest(c, "UsersAuthRegister")
	
	var req validator.RegisterNewUserRequest
	if !middleware.ValidateRequest(c, &req) {
		return
	}
	
	// Convert validator to DTO
	registerReq := dto.RegisterUserRequest{
		FirstName:       req.FirstName,
		LastName:        req.LastName,
		Email:           req.Email,
		PhoneNumber:     req.PhoneNumber,
		Password:        req.Password,
		ConfirmPassword: req.ConfirmPassword,
	}
	
	authResp, err := h.authService.RegisterUser(c.Request.Context(), registerReq)
	if err != nil {
		h.HandleError(c, err)
		return
	}
	
	h.HandleCreated(c, "Registration successful", authResp)
}

func (h *AuthHandler) UsersAuthGoogle(c *gin.Context) {
	h.LogRequest(c, "UsersAuthGoogle")
	
	var req validator.GoogleAuthRequest
	if !middleware.ValidateRequest(c, &req) {
		return
	}
	
	// Convert validator to DTO
	googleReq := dto.GoogleAuthRequest{
		Email:     req.Email,
		FirstName: req.FirstName,
	}
	
	authResp, err := h.authService.LoginUserWithGoogle(c.Request.Context(), googleReq)
	if err != nil {
		h.HandleError(c, err)
		return
	}
	
	h.HandleSuccess(c, "Google login successful", authResp)
}

// Mitra Authentication Methods

// MitrasAuthLogin godoc
// @Summary Mitra login
// @Description Authenticate a mitra (bengkel owner) with email and password
// @Tags Auth
// @Accept json
// @Produce json
// @Param request body validator.LoginRequest true "Login credentials"
// @Success 200 {object} response.Response{data=dto.AuthResponse} "Login successful"
// @Failure 400 {object} response.Response "Invalid request"
// @Failure 401 {object} response.Response "Invalid credentials"
// @Failure 429 {object} response.Response "Too many requests"
// @Failure 500 {object} response.Response "Internal server error"
// @Router /api/v1/mitras/auth/login [post]
func (h *AuthHandler) MitrasAuthLogin(c *gin.Context) {
	h.LogRequest(c, "MitrasAuthLogin")
	
	var req validator.LoginRequest
	if !middleware.ValidateRequest(c, &req) {
		return
	}
	
	// Convert validator to DTO
	loginReq := dto.LoginRequest{
		Email:    req.Email,
		Password: req.Password,
	}
	
	authResp, err := h.authService.LoginMitra(c.Request.Context(), loginReq)
	if err != nil {
		h.HandleError(c, err)
		return
	}
	
	h.HandleSuccess(c, "Login successful", authResp)
}

func (h *AuthHandler) MitrasAuthRegister(c *gin.Context) {
	h.LogRequest(c, "MitrasAuthRegister")
	
	var req validator.RegisterNewMitraRequest
	if !middleware.ValidateRequest(c, &req) {
		return
	}
	
	// Convert validator to DTO
	registerReq := dto.RegisterMitraRequest{
		FirstName:       req.FirstName,
		LastName:        req.LastName,
		Email:           req.Email,
		PhoneNumber:     req.PhoneNumber,
		Password:        req.Password,
		ConfirmPassword: req.ConfirmPassword,
	}
	
	authResp, err := h.authService.RegisterMitra(c.Request.Context(), registerReq)
	if err != nil {
		h.HandleError(c, err)
		return
	}
	
	h.HandleCreated(c, "Registration successful", authResp)
}

func (h *AuthHandler) MitrasAuthGoogle(c *gin.Context) {
	h.LogRequest(c, "MitrasAuthGoogle")
	
	var req validator.GoogleAuthRequest
	if !middleware.ValidateRequest(c, &req) {
		return
	}
	
	// Convert validator to DTO
	googleReq := dto.GoogleAuthRequest{
		Email:     req.Email,
		FirstName: req.FirstName,
	}
	
	authResp, err := h.authService.LoginMitraWithGoogle(c.Request.Context(), googleReq)
	if err != nil {
		h.HandleError(c, err)
		return
	}
	
	h.HandleSuccess(c, "Google login successful", authResp)
}

// Token Management Methods

func (h *AuthHandler) UsersRefreshToken(c *gin.Context) {
	h.LogRequest(c, "UsersRefreshToken")
	
	var req validator.RefreshTokenRequest
	if !middleware.ValidateRequest(c, &req) {
		return
	}
	
	authResp, err := h.authService.RefreshUserToken(c.Request.Context(), req.RefreshToken)
	if err != nil {
		h.HandleError(c, err)
		return
	}
	
	h.HandleSuccess(c, "Token refreshed successfully", authResp)
}

func (h *AuthHandler) MitrasRefreshToken(c *gin.Context) {
	h.LogRequest(c, "MitrasRefreshToken")
	
	var req validator.RefreshTokenRequest
	if !middleware.ValidateRequest(c, &req) {
		return
	}
	
	authResp, err := h.authService.RefreshMitraToken(c.Request.Context(), req.RefreshToken)
	if err != nil {
		h.HandleError(c, err)
		return
	}
	
	h.HandleSuccess(c, "Token refreshed successfully", authResp)
}

func (h *AuthHandler) UsersLogout(c *gin.Context) {
	h.LogRequest(c, "UsersLogout")
	
	var req validator.LogoutRequest
	if !middleware.ValidateRequest(c, &req) {
		return
	}
	
	err := h.authService.LogoutUser(c.Request.Context(), req.RefreshToken)
	if err != nil {
		h.HandleError(c, err)
		return
	}
	
	h.HandleSuccess(c, "Logout successful", nil)
}

func (h *AuthHandler) MitrasLogout(c *gin.Context) {
	h.LogRequest(c, "MitrasLogout")
	
	var req validator.LogoutRequest
	if !middleware.ValidateRequest(c, &req) {
		return
	}
	
	err := h.authService.LogoutMitra(c.Request.Context(), req.RefreshToken)
	if err != nil {
		h.HandleError(c, err)
		return
	}
	
	h.HandleSuccess(c, "Logout successful", nil)
}

func (h *AuthHandler) UsersLogoutAll(c *gin.Context) {
	h.LogRequest(c, "UsersLogoutAll")
	
	userID, err := h.GetUserID(c)
	if err != nil {
		h.HandleError(c, err)
		return
	}
	
	err = h.authService.LogoutAllUserDevices(c.Request.Context(), userID)
	if err != nil {
		h.HandleError(c, err)
		return
	}
	
	h.HandleSuccess(c, "Logout from all devices successful", nil)
}

func (h *AuthHandler) MitrasLogoutAll(c *gin.Context) {
	h.LogRequest(c, "MitrasLogoutAll")
	
	mitraID, err := h.GetMitraID(c)
	if err != nil {
		h.HandleError(c, err)
		return
	}
	
	err = h.authService.LogoutAllMitraDevices(c.Request.Context(), mitraID)
	if err != nil {
		h.HandleError(c, err)
		return
	}
	
	h.HandleSuccess(c, "Logout from all devices successful", nil)
}

// Legacy methods for backward compatibility (will be removed)

func (h *AuthHandler) UsersNewAddress(c *gin.Context) {
	applog.Info("Deprecated method called: UsersNewAddress - use UserHandler instead")
	h.HandleError(c, appErrors.ErrInternalServer.WithDetails("This endpoint has been moved to /api/v1/users/address"))
}

func (h *AuthHandler) UsersNewVehicle(c *gin.Context) {
	applog.Info("Deprecated method called: UsersNewVehicle - use UserHandler instead")
	h.HandleError(c, appErrors.ErrInternalServer.WithDetails("This endpoint has been moved to /api/v1/users/vehicle"))
}

func (h *AuthHandler) MitrasNewBank(c *gin.Context) {
	applog.Info("Deprecated method called: MitrasNewBank - use MitraHandler instead")
	h.HandleError(c, appErrors.ErrInternalServer.WithDetails("This endpoint has been moved to /api/v1/mitras/bank"))
}

func (h *AuthHandler) MitrasUpdateBank(c *gin.Context) {
	applog.Info("Deprecated method called: MitrasUpdateBank - use MitraHandler instead")
	h.HandleError(c, appErrors.ErrInternalServer.WithDetails("This endpoint has been moved to /api/v1/mitras/bank"))
}

func (h *AuthHandler) MitrasUpdateProfile(c *gin.Context) {
	applog.Info("Deprecated method called: MitrasUpdateProfile - use MitraHandler instead")
	h.HandleError(c, appErrors.ErrInternalServer.WithDetails("This endpoint has been moved to /api/v1/mitras/profile"))
}