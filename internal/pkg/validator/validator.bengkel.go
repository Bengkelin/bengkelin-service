package validator

type BengkelRegisterRequest struct {
	BengkelName  string   `json:"bengkel_name"`
	BengkelPhone string   `json:"bengkel_phone"`
	JumlahMontir uint     `json:"jumlah_montir"`
	Hari         []string `json:"hari"`
	JamBuka      []string `json:"jam_buka"`
}

type BengkelAddressRequest struct {
	Latitude     float64 `json:"latitude" binding:"required"`
	Longitude    float64 `json:"longitude" binding:"required"`
	AddressLabel string  `json:"address_label" binding:"required"`
	FullAddress  string  `json:"full_address" binding:"required"`
	Note         string  `json:"note"`
}

type BengkelServiceRequest struct {
	NamaService []string `json:"nama_service"`
}

type BengkelServiceOptionRequest struct {
	HomeService  bool `json:"home_service"`
	StoreService bool `json:"store_service"`
}
