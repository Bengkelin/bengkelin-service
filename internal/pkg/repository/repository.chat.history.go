package repository

import (
	"github.com/Bengkelin/bengkelin-service/internal/pkg/db"
	"github.com/Bengkelin/bengkelin-service/internal/pkg/models"
)

var (
	chatHistoryRepository *ChatHistoryRepository
)

type ChatHistoryRepositoryInterface interface {
	CreateChatHistory(chatHistory models.ChatHistory) (models.ChatHistory, error)
	UpdateChatHistoryById(chatHistoryId string, chatHistory *models.ChatHistory) error
	GetAllChatHistory() ([]models.ChatHistory, error)
	GetAllChatHistoryPaginate(page int, limit int) ([]models.ChatHistory, int, error)
}

type ChatHistoryRepository struct{}

func GetChatHistoryRepository() ChatHistoryRepositoryInterface {
	if chatHistoryRepository == nil {
		chatHistoryRepository = &ChatHistoryRepository{}
	}
	return chatHistoryRepository
}

// CreateChatHistory implements ChatHistoryRepositoryInterface.
func (repo *ChatHistoryRepository) CreateChatHistory(chatHistory models.ChatHistory) (models.ChatHistory, error) {
	err := db.GetDB().Create(&chatHistory).Error
	if err != nil {
		return models.ChatHistory{}, err
	}
	return chatHistory, nil
}

// UpdateChatHistoryById implements ChatHistoryRepositoryInterface.
func (*ChatHistoryRepository) UpdateChatHistoryById(chatHistoryId string, chatHistory *models.ChatHistory) error {
	err := db.GetDB().Model(&models.ChatHistory{}).Where("id = ?", chatHistoryId).Updates(chatHistory).Error

	if err != nil {
		return err
	}
	return nil
}

// GetAllChatHistory implements ChatHistoryRepositoryInterface.
func (*ChatHistoryRepository) GetAllChatHistory() ([]models.ChatHistory, error) {
	var chatHistory []models.ChatHistory
	err := db.GetDB().Find(&chatHistory).Error
	if err != nil {
		return nil, err
	}
	return chatHistory, nil
}

// GetAllChatHistoryPaginate implements ChatHistoryRepositoryInterface.
func (*ChatHistoryRepository) GetAllChatHistoryPaginate(page int, limit int) ([]models.ChatHistory, int, error) {
	var chatHistory []models.ChatHistory
	var total int64
	err := db.GetDB().Model(&models.ChatHistory{}).Count(&total).Error
	if err != nil {
		return nil, 0, err
	}
	err = db.GetDB().Offset((page - 1) * limit).Limit(limit).Find(&chatHistory).Error
	if err != nil {
		return nil, 0, err
	}
	return chatHistory, int(total), nil
}
