package validator

type RegisterBengkelRequest struct {
	BengkelName  string   `json:"bengkel_name"`
	BengkelPhone string   `json:"bengkel_phone"`
	JumlahMontir uint     `json:"jumlah_montir"`
	Hari         []string `json:"hari"`
	JamBuka      []string `json:"jam_buka"`
}
