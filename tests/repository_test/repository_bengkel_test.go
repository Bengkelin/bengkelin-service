package repository_test

import (
	"testing"

	"github.com/Bengkelin/bengkelin-service/internal/pkg/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// Mock Repository
type MockBengkelRepository struct {
	mock.Mock
}

func (m *MockBengkelRepository) GetBengkelById(bengkelId string) (*models.Bengkel, error) {
	args := m.Called(bengkelId)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Bengkel), args.Error(1)
}

func (m *MockBengkelRepository) CreateBengkel(bengkel models.Bengkel) (models.Bengkel, error) {
	args := m.Called(bengkel)
	return args.Get(0).(models.Bengkel), args.Error(1)
}

func TestGetBengkelById(t *testing.T) {
	mockRepo := new(MockBengkelRepository)

	expectedBengkel := &models.Bengkel{
		ID:          "test-id-123",
		BengkelName: "Test Bengkel",
	}

	mockRepo.On("GetBengkelById", "test-id-123").Return(expectedBengkel, nil)

	result, err := mockRepo.GetBengkelById("test-id-123")

	assert.NoError(t, err)
	assert.Equal(t, expectedBengkel.ID, result.ID)
	assert.Equal(t, expectedBengkel.BengkelName, result.BengkelName)
	mockRepo.AssertExpectations(t)
}

func TestCreateBengkel(t *testing.T) {
	mockRepo := new(MockBengkelRepository)

	newBengkel := models.Bengkel{
		BengkelName:  "New Bengkel",
		BengkelPhone: "08123456789",
	}

	mockRepo.On("CreateBengkel", newBengkel).Return(newBengkel, nil)

	result, err := mockRepo.CreateBengkel(newBengkel)

	assert.NoError(t, err)
	assert.Equal(t, newBengkel.BengkelName, result.BengkelName)
}
