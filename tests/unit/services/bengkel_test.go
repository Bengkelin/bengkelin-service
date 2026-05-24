package services_test

import (
	"context"
	"errors"
	"testing"

	"github.com/Bengkelin/bengkelin-service/internal/pkg/dto"
	"github.com/Bengkelin/bengkelin-service/internal/pkg/models"
	"github.com/Bengkelin/bengkelin-service/internal/pkg/service"
	"github.com/Bengkelin/bengkelin-service/tests/fixtures/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func setupBengkelService() (*mocks.MockBengkelRepository, *mocks.MockMitraRepository, *mocks.MockBengkelOperationalRepository, *mocks.MockBengkelTestimonialRepository, *mocks.MockOrderRepository, service.BengkelServiceInterface) {
	bengkelRepo := new(mocks.MockBengkelRepository)
	mitraRepo := new(mocks.MockMitraRepository)
	bengkelOpRepo := new(mocks.MockBengkelOperationalRepository)
	bengkelTestRepo := new(mocks.MockBengkelTestimonialRepository)
	orderRepo := new(mocks.MockOrderRepository)

	svc := service.NewBengkelService(service.ServiceDependencies{
		BengkelRepo:             bengkelRepo,
		MitraRepo:               mitraRepo,
		BengkelOperationalRepo:  bengkelOpRepo,
		BengkelTestimonialRepo:  bengkelTestRepo,
		OrderRepo:               orderRepo,
		BengkelAddressRepo:      new(mocks.MockBengkelAddressRepository),
		BengkelServiceRepo:      new(mocks.MockBengkelServiceRepository),
		BengkelPhotoRepo:        new(mocks.MockBengkelPhotoRepository),
	})
	return bengkelRepo, mitraRepo, bengkelOpRepo, bengkelTestRepo, orderRepo, svc
}

// --- CreateTestimonial ---

func TestCreateTestimonial_Success(t *testing.T) {
	bengkelRepo, _, _, bengkelTestRepo, orderRepo, svc := setupBengkelService()
	ctx := context.Background()

	order := &models.Order{ID: "order-1"}
	orderRepo.On("GetOrderById", ctx, "order-1").Return(order, nil)

	bengkel := &models.Bengkel{ID: "bengkel-1"}
	bengkelRepo.On("GetBengkelById", ctx, "bengkel-1").Return(bengkel, nil)

	testimonial := models.BengkelTestimonial{
		BengkelID: "bengkel-1",
		UserID:    "user-1",
		OrderID:   "order-1",
		Testimoni: "Great service!",
		Rating:    5,
	}
	bengkelTestRepo.On("CreateBengkelTestimonial", ctx, mock.Anything).Return(testimonial, nil)

	err := svc.CreateTestimonial(ctx, "user-1", "bengkel-1", "order-1", "Great service!", 5)

	assert.NoError(t, err)
	orderRepo.AssertExpectations(t)
	bengkelRepo.AssertExpectations(t)
	bengkelTestRepo.AssertExpectations(t)
}

func TestCreateTestimonial_OrderNotFound(t *testing.T) {
	_, _, _, _, orderRepo, svc := setupBengkelService()
	ctx := context.Background()

	orderRepo.On("GetOrderById", ctx, "order-999").Return(nil, errors.New("not found"))

	err := svc.CreateTestimonial(ctx, "user-1", "bengkel-1", "order-999", "Great!", 5)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "order not found")
}

func TestCreateTestimonial_BengkelNotFound(t *testing.T) {
	bengkelRepo, _, _, _, orderRepo, svc := setupBengkelService()
	ctx := context.Background()

	order := &models.Order{ID: "order-1"}
	orderRepo.On("GetOrderById", ctx, "order-1").Return(order, nil)
	bengkelRepo.On("GetBengkelById", ctx, "bengkel-999").Return(nil, errors.New("not found"))

	err := svc.CreateTestimonial(ctx, "user-1", "bengkel-999", "order-1", "Great!", 5)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "bengkel not found")
}

// --- GetMitraWithBengkel ---

func TestGetMitraWithBengkel_Success(t *testing.T) {
	_, mitraRepo, _, _, _, svc := setupBengkelService()
	ctx := context.Background()

	mitra := &models.Mitra{
		ID: "mitra-1",
		Bengkel: []models.Bengkel{
			{ID: "bengkel-1", BengkelName: "Honda Mitra"},
		},
	}
	mitraRepo.On("FindMitraByID", ctx, "mitra-1").Return(mitra, nil)

	result, err := svc.GetMitraWithBengkel(ctx, "mitra-1")

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "mitra-1", result.ID)
	assert.Len(t, result.Bengkel, 1)
	mitraRepo.AssertExpectations(t)
}

// --- GetAllBengkels ---

func TestGetAllBengkels_Success(t *testing.T) {
	bengkelRepo, _, _, _, _, svc := setupBengkelService()
	ctx := context.Background()

	bengkels := []models.Bengkel{
		{ID: "bengkel-1", BengkelName: "Honda Mitra", MitraID: "mitra-1"},
		{ID: "bengkel-2", BengkelName: "Yamaha Mitra", MitraID: "mitra-2"},
	}
	bengkelRepo.On("GetAllBengkel", ctx).Return(bengkels, nil)

	result, err := svc.GetAllBengkels(ctx)

	assert.NoError(t, err)
	assert.Len(t, result, 2)
	assert.Equal(t, "Honda Mitra", result[0].BengkelName)
	assert.Equal(t, "Yamaha Mitra", result[1].BengkelName)
	bengkelRepo.AssertExpectations(t)
}

func TestGetAllBengkels_Empty(t *testing.T) {
	bengkelRepo, _, _, _, _, svc := setupBengkelService()
	ctx := context.Background()

	bengkelRepo.On("GetAllBengkel", ctx).Return([]models.Bengkel{}, nil)

	result, err := svc.GetAllBengkels(ctx)

	assert.NoError(t, err)
	assert.Len(t, result, 0)
}

// --- GetAllBengkelsPaginate ---

func TestGetAllBengkelsPaginate_Success(t *testing.T) {
	bengkelRepo, _, _, _, _, svc := setupBengkelService()
	ctx := context.Background()

	bengkels := []models.Bengkel{
		{ID: "bengkel-1", BengkelName: "Honda Mitra", MitraID: "mitra-1"},
	}
	bengkelRepo.On("GetAllBengkelPaginate", ctx, 1, 10).Return(bengkels, 1, nil)

	result, count, err := svc.GetAllBengkelsPaginate(ctx, 1, 10)

	assert.NoError(t, err)
	assert.Len(t, result, 1)
	assert.Equal(t, 1, count)
	assert.Equal(t, "Honda Mitra", result[0].BengkelName)
	bengkelRepo.AssertExpectations(t)
}

// --- SearchBengkelsV2 ---

func TestSearchBengkelsV2_Success(t *testing.T) {
	bengkelRepo, _, _, _, _, svc := setupBengkelService()
	ctx := context.Background()

	bengkels := []models.Bengkel{
		{ID: "bengkel-1", BengkelName: "Honda Mitra", MitraID: "mitra-1"},
	}
	bengkelRepo.On("GetBengkelSearchV2", ctx, "home", "honda", 1, 10).Return(bengkels, 1, nil)

	result, count, err := svc.SearchBengkelsV2(ctx, "home", "honda", 1, 10)

	assert.NoError(t, err)
	assert.Len(t, result, 1)
	assert.Equal(t, 1, count)
	bengkelRepo.AssertExpectations(t)
}

// --- GetBengkelDetailWithTestimonials ---

func TestGetBengkelDetailWithTestimonials_Success(t *testing.T) {
	bengkelRepo, _, _, _, _, svc := setupBengkelService()
	ctx := context.Background()

	bengkel := &models.Bengkel{ID: "bengkel-1", BengkelName: "Honda Mitra"}
	testimonials := []models.BengkelTestimonial{
		{ID: 1, BengkelID: "bengkel-1", Testimoni: "Great!", Rating: 5},
	}
	bengkelRepo.On("FindBengkelById", ctx, "bengkel-1", 1, 10).Return(bengkel, testimonials, 1, nil)

	resultBengkel, resultTestimonials, count, err := svc.GetBengkelDetailWithTestimonials(ctx, "bengkel-1", 1, 10)

	assert.NoError(t, err)
	assert.NotNil(t, resultBengkel)
	assert.Equal(t, "bengkel-1", resultBengkel.ID)
	assert.Len(t, resultTestimonials, 1)
	assert.Equal(t, 1, count)
	bengkelRepo.AssertExpectations(t)
}

// --- GetBengkelOperationalTimeSlots ---

func TestGetBengkelOperationalTimeSlots_Success(t *testing.T) {
	_, _, bengkelOpRepo, _, _, svc := setupBengkelService()
	ctx := context.Background()

	operational := &models.BengkelOperational{
		BengkelID: "bengkel-1",
		Hari:      "Senin",
		JamBuka:   "08:00 - 17:00",
	}
	bengkelOpRepo.On("GetBengkelOperationalByIdAndDay", ctx, "bengkel-1", "Senin").Return(operational, nil)

	result, err := svc.GetBengkelOperationalTimeSlots(ctx, "bengkel-1", "Senin")

	assert.NoError(t, err)
	assert.NotEmpty(t, result)
	// Should have slots from 08:00 to 17:00 = 9 slots
	assert.Len(t, result, 9)
	assert.Equal(t, "08:00 - 09:00", result[0])
	bengkelOpRepo.AssertExpectations(t)
}

func TestGetBengkelOperationalTimeSlots_NoHours(t *testing.T) {
	_, _, bengkelOpRepo, _, _, svc := setupBengkelService()
	ctx := context.Background()

	operational := &models.BengkelOperational{
		BengkelID: "bengkel-1",
		Hari:      "Senin",
		JamBuka:   "",
	}
	bengkelOpRepo.On("GetBengkelOperationalByIdAndDay", ctx, "bengkel-1", "Senin").Return(operational, nil)

	result, err := svc.GetBengkelOperationalTimeSlots(ctx, "bengkel-1", "Senin")

	assert.NoError(t, err)
	assert.Empty(t, result)
	bengkelOpRepo.AssertExpectations(t)
}

func TestGetBengkelOperationalTimeSlots_NotFound(t *testing.T) {
	_, _, bengkelOpRepo, _, _, svc := setupBengkelService()
	ctx := context.Background()

	bengkelOpRepo.On("GetBengkelOperationalByIdAndDay", ctx, "bengkel-1", "Minggu").Return(nil, errors.New("not found"))

	result, err := svc.GetBengkelOperationalTimeSlots(ctx, "bengkel-1", "Minggu")

	assert.Error(t, err)
	assert.Nil(t, result)
}

// --- CreateBengkel ---

func TestCreateBengkel_Success(t *testing.T) {
	bengkelRepo, mitraRepo, bengkelOpRepo, _, _, svc := setupBengkelService()
	ctx := context.Background()

	mitra := &models.Mitra{ID: "mitra-1"}
	mitraRepo.On("FindMitraByID", ctx, "mitra-1").Return(mitra, nil)

	createdBengkel := models.Bengkel{
		ID:           "bengkel-1",
		MitraID:      "mitra-1",
		BengkelName:  "Honda Mitra",
		BengkelPhone: "08123456789",
		JumlahMontir: 3,
	}
	bengkelRepo.On("CreateBengkel", ctx, mock.Anything).Return(createdBengkel, nil)

	createdOp := models.BengkelOperational{BengkelID: "bengkel-1", Hari: "Senin", JamBuka: "08:00 - 17:00"}
	bengkelOpRepo.On("CreateBengkelOperational", ctx, mock.Anything).Return(createdOp, nil)

	req := dto.CreateBengkelRequest{
		BengkelName:  "Honda Mitra",
		BengkelPhone: "08123456789",
		JumlahMontir: 3,
		Hari:         []string{"Senin"},
		JamBuka:      []string{"08:00 - 17:00"},
	}

	result, err := svc.CreateBengkel(ctx, "mitra-1", req)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "Honda Mitra", result.BengkelName)
	mitraRepo.AssertExpectations(t)
	bengkelRepo.AssertExpectations(t)
}

func TestCreateBengkel_MitraNotFound(t *testing.T) {
	_, mitraRepo, _, _, _, svc := setupBengkelService()
	ctx := context.Background()

	mitraRepo.On("FindMitraByID", ctx, "mitra-999").Return(nil, errors.New("not found"))

	req := dto.CreateBengkelRequest{BengkelName: "Test"}
	result, err := svc.CreateBengkel(ctx, "mitra-999", req)

	assert.Error(t, err)
	assert.Nil(t, result)
}

// --- GetBengkelProfile ---

func TestGetBengkelProfile_Success(t *testing.T) {
	_, mitraRepo, _, _, _, svc := setupBengkelService()
	ctx := context.Background()

	mitra := &models.Mitra{
		ID: "mitra-1",
		Bengkel: []models.Bengkel{
			{ID: "bengkel-1", BengkelName: "Honda Mitra", MitraID: "mitra-1"},
		},
	}
	mitraRepo.On("FindMitraByID", ctx, "mitra-1").Return(mitra, nil)

	result, err := svc.GetBengkelProfile(ctx, "mitra-1")

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "Honda Mitra", result.BengkelName)
	mitraRepo.AssertExpectations(t)
}

func TestGetBengkelProfile_NoBengkel(t *testing.T) {
	_, mitraRepo, _, _, _, svc := setupBengkelService()
	ctx := context.Background()

	mitra := &models.Mitra{
		ID:       "mitra-1",
		Bengkel:  []models.Bengkel{},
	}
	mitraRepo.On("FindMitraByID", ctx, "mitra-1").Return(mitra, nil)

	result, err := svc.GetBengkelProfile(ctx, "mitra-1")

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "mitra has no bengkel")
}

// --- UpdateBengkelMontir ---

func TestUpdateBengkelMontir_Success(t *testing.T) {
	bengkelRepo, mitraRepo, _, _, _, svc := setupBengkelService()
	ctx := context.Background()

	mitra := &models.Mitra{
		ID: "mitra-1",
		Bengkel: []models.Bengkel{
			{ID: "bengkel-1"},
		},
	}
	mitraRepo.On("FindMitraByID", ctx, "mitra-1").Return(mitra, nil)
	bengkelRepo.On("UpdateBengkelById", ctx, "bengkel-1", mock.Anything).Return(nil)

	err := svc.UpdateBengkelMontir(ctx, "mitra-1", 5)

	assert.NoError(t, err)
	mitraRepo.AssertExpectations(t)
	bengkelRepo.AssertExpectations(t)
}

func TestUpdateBengkelMontir_MitraNotFound(t *testing.T) {
	_, mitraRepo, _, _, _, svc := setupBengkelService()
	ctx := context.Background()

	mitraRepo.On("FindMitraByID", ctx, "mitra-999").Return(nil, errors.New("not found"))

	err := svc.UpdateBengkelMontir(ctx, "mitra-999", 5)

	assert.Error(t, err)
}

func TestUpdateBengkelMontir_NoBengkel(t *testing.T) {
	_, mitraRepo, _, _, _, svc := setupBengkelService()
	ctx := context.Background()

	mitra := &models.Mitra{ID: "mitra-1", Bengkel: []models.Bengkel{}}
	mitraRepo.On("FindMitraByID", ctx, "mitra-1").Return(mitra, nil)

	err := svc.UpdateBengkelMontir(ctx, "mitra-1", 5)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "mitra has no bengkel")
}

// --- UpdateBengkelStatusOptions ---

func TestUpdateBengkelStatusOptions_Success(t *testing.T) {
	bengkelRepo, mitraRepo, _, _, _, svc := setupBengkelService()
	ctx := context.Background()

	mitra := &models.Mitra{
		ID: "mitra-1",
		Bengkel: []models.Bengkel{
			{ID: "bengkel-1"},
		},
	}
	mitraRepo.On("FindMitraByID", ctx, "mitra-1").Return(mitra, nil)
	bengkelRepo.On("UpdateBengkelById", ctx, "bengkel-1", mock.Anything).Return(nil)

	err := svc.UpdateBengkelStatusOptions(ctx, "mitra-1", true, true, true)

	assert.NoError(t, err)
	mitraRepo.AssertExpectations(t)
	bengkelRepo.AssertExpectations(t)
}

func TestUpdateBengkelStatusOptions_MitraNotFound(t *testing.T) {
	_, mitraRepo, _, _, _, svc := setupBengkelService()
	ctx := context.Background()

	mitraRepo.On("FindMitraByID", ctx, "mitra-999").Return(nil, errors.New("not found"))

	err := svc.UpdateBengkelStatusOptions(ctx, "mitra-999", true, true, true)

	assert.Error(t, err)
}
