package repository

import (
	"context"
	"sync"

	"github.com/Bengkelin/bengkelin-service/internal/pkg/db"
	"github.com/Bengkelin/bengkelin-service/internal/pkg/models"
)

var (
	chatHistoryRepository *ChatHistoryRepository
	chatHistoryOnce       sync.Once
)

type ChatHistoryRepositoryInterface interface {
	CreateChatHistory(ctx context.Context, chatHistory models.ChatHistory) (models.ChatHistory, error)
	UpdateChatHistoryById(ctx context.Context, chatHistoryId string, chatHistory *models.ChatHistory) error
	GetAllChatHistory(ctx context.Context) ([]models.ChatHistory, error)
	GetAllChatHistoryPaginate(ctx context.Context, page int, limit int, senderId, receiverId string) ([]models.ChatHistory, int, error)
}

type ChatHistoryRepository struct{}

func GetChatHistoryRepository() ChatHistoryRepositoryInterface {
	chatHistoryOnce.Do(func() {
		chatHistoryRepository = &ChatHistoryRepository{}
	})
	return chatHistoryRepository
}

// CreateChatHistory implements ChatHistoryRepositoryInterface.
func (repo *ChatHistoryRepository) CreateChatHistory(ctx context.Context, chatHistory models.ChatHistory) (models.ChatHistory, error) {
	err := db.GetDB().WithContext(ctx).Create(&chatHistory).Error
	if err != nil {
		return models.ChatHistory{}, err
	}
	return chatHistory, nil
}

// UpdateChatHistoryById implements ChatHistoryRepositoryInterface.
func (*ChatHistoryRepository) UpdateChatHistoryById(ctx context.Context, chatHistoryId string, chatHistory *models.ChatHistory) error {
	err := db.GetDB().WithContext(ctx).Model(&models.ChatHistory{}).Where("id = ?", chatHistoryId).Updates(chatHistory).Error

	if err != nil {
		return err
	}
	return nil
}

// GetAllChatHistory implements ChatHistoryRepositoryInterface.
func (*ChatHistoryRepository) GetAllChatHistory(ctx context.Context) ([]models.ChatHistory, error) {
	var chatHistory []models.ChatHistory
	err := db.GetDB().WithContext(ctx).Find(&chatHistory).Error
	if err != nil {
		return nil, err
	}
	return chatHistory, nil
}

// GetAllChatHistoryPaginate implements ChatHistoryRepositoryInterface.
func (*ChatHistoryRepository) GetAllChatHistoryPaginate(ctx context.Context, page int, limit int, senderId, receiverId string) ([]models.ChatHistory, int, error) {
	var chatHistory []models.ChatHistory
	var total int64
	err := db.GetDB().WithContext(ctx).
		Model(&models.ChatHistory{}).
		Where("sender_user_id = ? and receiver_user_id = ?", senderId, receiverId).
		Count(&total).Error

	if err != nil {
		return nil, 0, err
	}
	err = db.GetDB().WithContext(ctx).
		Model(&models.ChatHistory{}).
		Where("sender_user_id = ? and receiver_user_id = ?", senderId, receiverId).
		Offset((page - 1) * limit).
		Limit(limit).
		Find(&chatHistory).Error
	if err != nil {
		return nil, 0, err
	}
	return chatHistory, int(total), nil
}
