package validator

type MitraBankRequest struct {
	BankName   string `binding:"required" json:"bank_name"`
	BankNumber string `binding:"required" json:"bank_number"`
}
