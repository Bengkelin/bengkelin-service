package validator

type AdminFeeRequest struct {
	AdminFee float64 `json:"admin_fee" binding:"required"`
}