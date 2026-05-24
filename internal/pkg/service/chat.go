package service

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/Bengkelin/bengkelin-service/internal/pkg/config"
	"github.com/Bengkelin/bengkelin-service/internal/pkg/dto"
	"github.com/Bengkelin/bengkelin-service/internal/pkg/models"
	"github.com/Bengkelin/bengkelin-service/internal/pkg/repository"
	"github.com/Bengkelin/bengkelin-service/pkg/helpers"
)

type ChatServiceImpl struct {
	chatHistoryRepo repository.ChatHistoryRepositoryInterface
	userRepo        repository.UserRepositoryInterface
	mitraRepo       repository.MitraRepositoryInterface
}

func NewChatService(deps ServiceDependencies) ChatServiceInterface {
	return &ChatServiceImpl{
		chatHistoryRepo: deps.ChatHistoryRepo,
		userRepo:        deps.UserRepo,
		mitraRepo:       deps.MitraRepo,
	}
}

func (s *ChatServiceImpl) CreateUserAppToken(ctx context.Context, userID string) (*dto.ChatTokenResponse, error) {
	return nil, fmt.Errorf("not implemented - use Agora directly")
}

func (s *ChatServiceImpl) CreateMitraAppToken(ctx context.Context, mitraID string) (*dto.ChatTokenResponse, error) {
	return nil, fmt.Errorf("not implemented - use Agora directly")
}

func (s *ChatServiceImpl) CreateUserChatToken(ctx context.Context, userID string) (*dto.ChatTokenResponse, error) {
	return nil, fmt.Errorf("not implemented - use Agora directly")
}

func (s *ChatServiceImpl) CreateMitraChatToken(ctx context.Context, mitraID string) (*dto.ChatTokenResponse, error) {
	return nil, fmt.Errorf("not implemented - use Agora directly")
}

func (s *ChatServiceImpl) SaveChatHistory(ctx context.Context, req dto.ChatHistoryRequest) error {
	chatModel := models.ChatHistory{
		ID:             helpers.GenerateUUID(),
		MessageText:    req.Message,
		SenderUserId:   req.UserID,
		ReceiverUserId: req.BengkelID,
	}

	_, err := s.chatHistoryRepo.CreateChatHistory(ctx, chatModel)
	if err != nil {
		return fmt.Errorf("failed to save chat history: %w", err)
	}

	return nil
}

func (s *ChatServiceImpl) GetChatHistory(ctx context.Context, userID string, userType string) ([]dto.ChatHistoryResponse, error) {
	res, _, err := s.chatHistoryRepo.GetAllChatHistoryPaginate(ctx, 1, 100, "", "")
	if err != nil {
		return nil, err
	}

	var responses []dto.ChatHistoryResponse
	for _, ch := range res {
		responses = append(responses, dto.ChatHistoryResponse{
			ID:        ch.ID,
			UserID:    ch.SenderUserId,
			BengkelID: ch.ReceiverUserId,
			Message:   ch.MessageText,
			CreatedAt: ch.CreatedAt,
		})
	}

	return responses, nil
}

func (s *ChatServiceImpl) GenerateRtmParams(ctx context.Context, userID string) (string, uint32, error) {
	if _, err := s.userRepo.GetDetailUser(ctx, userID); err != nil {
		return "", 0, fmt.Errorf("user not found: %w", err)
	}

	cfg := config.GetConfig()
	expireTime64, err := strconv.ParseUint(cfg.Agore.ExpiryTime, 10, 64)
	if err != nil {
		return "", 0, fmt.Errorf("failed to parse expireTime: %s, causing error: %s", cfg.Agore.ExpiryTime, err)
	}

	currentTimestamp := uint32(time.Now().UTC().Unix())
	expireTimestamp := currentTimestamp + uint32(expireTime64)

	return userID, expireTimestamp, nil
}

func (s *ChatServiceImpl) ValidateUserExists(ctx context.Context, userID string) error {
	if _, err := s.userRepo.GetDetailUser(ctx, userID); err != nil {
		return fmt.Errorf("user not found: %w", err)
	}
	return nil
}

func (s *ChatServiceImpl) ValidateMitraExists(ctx context.Context, mitraID string) error {
	if _, err := s.mitraRepo.FindMitraByID(ctx, mitraID); err != nil {
		return fmt.Errorf("mitra not found: %w", err)
	}
	return nil
}

func (s *ChatServiceImpl) GetChatHistoryPaginate(ctx context.Context, page, limit int, senderID, receiverID string) ([]models.ChatHistory, int, error) {
	return s.chatHistoryRepo.GetAllChatHistoryPaginate(ctx, page, limit, senderID, receiverID)
}
