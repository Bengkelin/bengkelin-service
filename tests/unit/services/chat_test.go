package services_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/Bengkelin/bengkelin-service/internal/dto"
	"github.com/Bengkelin/bengkelin-service/internal/models"
	"github.com/Bengkelin/bengkelin-service/internal/service"
	"github.com/Bengkelin/bengkelin-service/tests/fixtures/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func setupChatService() (*mocks.MockUserRepository, *mocks.MockMitraRepository, *mocks.MockChatHistoryRepository, service.ChatServiceInterface) {
	userRepo := new(mocks.MockUserRepository)
	mitraRepo := new(mocks.MockMitraRepository)
	chatRepo := new(mocks.MockChatHistoryRepository)
	svc := service.NewChatService(service.ServiceDependencies{
		UserRepo:        userRepo,
		MitraRepo:       mitraRepo,
		ChatHistoryRepo: chatRepo,
	})
	return userRepo, mitraRepo, chatRepo, svc
}

// --- ValidateUserExists ---

func TestChatValidateUserExists_Success(t *testing.T) {
	userRepo, _, _, svc := setupChatService()
	ctx := context.Background()

	user := &models.User{ID: "user-1"}
	userRepo.On("GetDetailUser", ctx, "user-1").Return(user, nil)

	err := svc.ValidateUserExists(ctx, "user-1")

	assert.NoError(t, err)
	userRepo.AssertExpectations(t)
}

func TestChatValidateUserExists_NotFound(t *testing.T) {
	userRepo, _, _, svc := setupChatService()
	ctx := context.Background()

	userRepo.On("GetDetailUser", ctx, "user-999").Return(nil, errors.New("not found"))

	err := svc.ValidateUserExists(ctx, "user-999")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "user not found")
	userRepo.AssertExpectations(t)
}

// --- ValidateMitraExists ---

func TestChatValidateMitraExists_Success(t *testing.T) {
	_, mitraRepo, _, svc := setupChatService()
	ctx := context.Background()

	mitra := &models.Mitra{ID: "mitra-1"}
	mitraRepo.On("FindMitraByID", ctx, "mitra-1").Return(mitra, nil)

	err := svc.ValidateMitraExists(ctx, "mitra-1")

	assert.NoError(t, err)
	mitraRepo.AssertExpectations(t)
}

func TestChatValidateMitraExists_NotFound(t *testing.T) {
	_, mitraRepo, _, svc := setupChatService()
	ctx := context.Background()

	mitraRepo.On("FindMitraByID", ctx, "mitra-999").Return(nil, errors.New("not found"))

	err := svc.ValidateMitraExists(ctx, "mitra-999")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "mitra not found")
	mitraRepo.AssertExpectations(t)
}

// --- GetChatHistoryPaginate ---

func TestGetChatHistoryPaginate_Success(t *testing.T) {
	_, _, chatRepo, svc := setupChatService()
	ctx := context.Background()

	history := []models.ChatHistory{
		{ID: "msg-1", SenderUserId: "user-1", ReceiverUserId: "mitra-1", MessageText: "Hello", CreatedAt: time.Now()},
		{ID: "msg-2", SenderUserId: "mitra-1", ReceiverUserId: "user-1", MessageText: "Hi there", CreatedAt: time.Now()},
	}
	chatRepo.On("GetAllChatHistoryPaginate", ctx, 1, 10, "user-1", "mitra-1").Return(history, 2, nil)

	result, count, err := svc.GetChatHistoryPaginate(ctx, 1, 10, "user-1", "mitra-1")

	assert.NoError(t, err)
	assert.Len(t, result, 2)
	assert.Equal(t, 2, count)
	assert.Equal(t, "msg-1", result[0].ID)
	assert.Equal(t, "Hello", result[0].MessageText)
	chatRepo.AssertExpectations(t)
}

func TestGetChatHistoryPaginate_Empty(t *testing.T) {
	_, _, chatRepo, svc := setupChatService()
	ctx := context.Background()

	chatRepo.On("GetAllChatHistoryPaginate", ctx, 1, 10, "user-1", "mitra-1").Return([]models.ChatHistory{}, 0, nil)

	result, count, err := svc.GetChatHistoryPaginate(ctx, 1, 10, "user-1", "mitra-1")

	assert.NoError(t, err)
	assert.Len(t, result, 0)
	assert.Equal(t, 0, count)
	chatRepo.AssertExpectations(t)
}

func TestGetChatHistoryPaginate_Error(t *testing.T) {
	_, _, chatRepo, svc := setupChatService()
	ctx := context.Background()

	chatRepo.On("GetAllChatHistoryPaginate", ctx, 1, 10, "user-1", "mitra-1").Return([]models.ChatHistory{}, 0, errors.New("db error"))

	_, count, err := svc.GetChatHistoryPaginate(ctx, 1, 10, "user-1", "mitra-1")

	assert.Error(t, err)
	assert.Equal(t, 0, count)
}

// --- SaveChatHistory ---

func TestSaveChatHistory_Success(t *testing.T) {
	_, _, chatRepo, svc := setupChatService()
	ctx := context.Background()

	chatRepo.On("CreateChatHistory", ctx, mock.Anything).Return(models.ChatHistory{ID: "msg-1"}, nil)

	req := dto.ChatHistoryRequest{
		UserID:    "user-1",
		BengkelID: "mitra-1",
		Message:   "Hello bengkel!",
	}

	err := svc.SaveChatHistory(ctx, req)

	assert.NoError(t, err)
	chatRepo.AssertExpectations(t)
}

func TestSaveChatHistory_RepoError(t *testing.T) {
	_, _, chatRepo, svc := setupChatService()
	ctx := context.Background()

	chatRepo.On("CreateChatHistory", ctx, mock.Anything).Return(models.ChatHistory{}, errors.New("db error"))

	req := dto.ChatHistoryRequest{
		UserID:    "user-1",
		BengkelID: "mitra-1",
		Message:   "Hello!",
	}

	err := svc.SaveChatHistory(ctx, req)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to save chat history")
}

// --- GetChatHistory ---

func TestGetChatHistory_Success(t *testing.T) {
	_, _, chatRepo, svc := setupChatService()
	ctx := context.Background()

	history := []models.ChatHistory{
		{ID: "msg-1", SenderUserId: "user-1", ReceiverUserId: "mitra-1", MessageText: "Hello", CreatedAt: time.Now()},
		{ID: "msg-2", SenderUserId: "mitra-1", ReceiverUserId: "user-1", MessageText: "Hi!", CreatedAt: time.Now()},
	}
	chatRepo.On("GetAllChatHistoryPaginate", ctx, 1, 100, "", "").Return(history, 2, nil)

	result, err := svc.GetChatHistory(ctx, "user-1", "user")

	assert.NoError(t, err)
	assert.Len(t, result, 2)
	assert.Equal(t, "msg-1", result[0].ID)
	assert.Equal(t, "Hello", result[0].Message)
	assert.Equal(t, "user-1", result[0].UserID)
	assert.Equal(t, "mitra-1", result[0].BengkelID)
	chatRepo.AssertExpectations(t)
}

func TestGetChatHistory_Empty(t *testing.T) {
	_, _, chatRepo, svc := setupChatService()
	ctx := context.Background()

	chatRepo.On("GetAllChatHistoryPaginate", ctx, 1, 100, "", "").Return([]models.ChatHistory{}, 0, nil)

	result, err := svc.GetChatHistory(ctx, "user-1", "user")

	assert.NoError(t, err)
	assert.Len(t, result, 0)
}

func TestGetChatHistory_Error(t *testing.T) {
	_, _, chatRepo, svc := setupChatService()
	ctx := context.Background()

	chatRepo.On("GetAllChatHistoryPaginate", ctx, 1, 100, "", "").Return([]models.ChatHistory{}, 0, errors.New("db error"))

	result, err := svc.GetChatHistory(ctx, "user-1", "user")

	assert.Error(t, err)
	assert.Nil(t, result)
}

// --- GenerateRtmParams ---

func TestGenerateRtmParams_Success(t *testing.T) {
	t.Skip("Requires config.GetConfig() — panics without initialized config")

	userRepo, _, _, svc := setupChatService()
	ctx := context.Background()

	user := &models.User{ID: "user-1"}
	userRepo.On("GetDetailUser", ctx, "user-1").Return(user, nil)

	token, expireTime, err := svc.GenerateRtmParams(ctx, "user-1")

	assert.NoError(t, err)
	assert.Equal(t, "user-1", token)
	assert.Greater(t, expireTime, uint32(0))
	userRepo.AssertExpectations(t)
}

func TestGenerateRtmParams_UserNotFound(t *testing.T) {
	userRepo, _, _, svc := setupChatService()
	ctx := context.Background()

	userRepo.On("GetDetailUser", ctx, "user-999").Return(nil, errors.New("not found"))

	_, _, err := svc.GenerateRtmParams(ctx, "user-999")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "user not found")
}
