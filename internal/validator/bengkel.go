package validator

type BengkelRegisterRequest struct {
	BengkelName  string   `json:"bengkel_name" binding:"required,min=2,max=100" validate:"no_xss,no_sql_injection"`
	BengkelPhone string   `json:"bengkel_phone" binding:"required" validate:"phone"`
	JumlahMontir uint     `json:"jumlah_montir" binding:"required,min=1,max=50"`
	Hari         []string `json:"hari" binding:"required,min=1,max=7" validate:"dive,day_name"`
	JamBuka      []string `json:"jam_buka" binding:"required,min=1,max=7" validate:"dive,time_format"`
}

type BengkelRegisterRequestV2 struct {
	BengkelName   string                    `json:"bengkel_name" binding:"required,min=2,max=100" validate:"no_xss,no_sql_injection"`
	BengkelPhone  string                    `json:"bengkel_phone" binding:"required" validate:"phone"`
	JumlahMontir  uint                      `json:"jumlah_montir" binding:"required,min=1,max=50"`
	Operasionals  []BengkelOperationalItem  `json:"operasionals" binding:"required,min=1,max=7" validate:"dive"`
}

type BengkelUpdateRequest struct {
	BengkelName  string `json:"bengkel_name" binding:"min=2,max=100" validate:"no_xss,no_sql_injection"`
	BengkelPhone string `json:"bengkel_phone" validate:"phone"`
	JumlahMontir uint   `json:"jumlah_montir" binding:"min=1,max=50"`
}

type BengkelMontirUpdateRequest struct {
	JumlahMontir uint `json:"jumlah_montir" binding:"required,min=1,max=50"`
}

type BengkelOperasionalUpdateRequest struct {
	Hari    []string `json:"hari" binding:"required,min=1,max=7" validate:"dive,day_name"`
	JamBuka []string `json:"jam_buka" binding:"required,min=1,max=7" validate:"dive,time_format"`
}

type BengkelOperationalItem struct {
	ID       uint   `json:"id"`
	Hari     string `json:"hari" binding:"required" validate:"day_name"`
	JamBuka  string `json:"jam_buka" binding:"required" validate:"time_format"`
	JamTutup string `json:"jam_tutup" binding:"required" validate:"time_format"`
	IsActive *bool  `json:"is_active"`
}

type BengkelOperationalUpdateRequestV2 struct {
	Operasionals []BengkelOperationalItem `json:"operasionals" binding:"required,min=1,max=7" validate:"dive"`
}

type BengkelAddressRequest struct {
	Latitude     float64 `json:"latitude" binding:"required" validate:"latitude"`
	Longitude    float64 `json:"longitude" binding:"required" validate:"longitude"`
	AddressLabel string  `json:"address_label" binding:"required,min=1,max=100" validate:"no_xss,no_sql_injection"`
	FullAddress  string  `json:"full_address" binding:"required,min=10,max=500" validate:"no_xss,no_sql_injection"`
	Note         string  `json:"note" binding:"max=200" validate:"no_xss,no_sql_injection"`
}

type BengkelServiceRequest struct {
	NamaService []string `json:"nama_service" binding:"required,min=1,max=20" validate:"dive,required,min=1,max=100,alpha_numeric_space,no_xss"`
}

type BengkelServiceItem struct {
	ID          uint     `json:"id"`
	NamaService string   `json:"nama_service" binding:"required,min=1,max=100" validate:"alpha_numeric_space,no_xss"`
	Description string   `json:"description" binding:"max=500" validate:"no_xss,no_sql_injection"`
	Price       float64  `json:"price" binding:"required,min=0" validate:"min=0"`
	IsAvailable *bool    `json:"is_available"`
}

type BengkelServiceCreateRequest struct {
	Services []BengkelServiceItem `json:"services" binding:"required,min=1,max=20" validate:"dive"`
}

type BengkelServiceUpdateRequest struct {
	Services []BengkelServiceItem `json:"services" binding:"required,min=1,max=50" validate:"dive"`
}

type BengkelServiceOptionRequest struct {
	HomeService  bool `json:"home_service"`
	StoreService bool `json:"store_service"`
	IsOpen       bool `json:"is_open"`
}

type BengkelTestimoniRequest struct {
	Testimoni string `json:"testimoni" binding:"required,min=10,max=1000" validate:"no_xss,no_sql_injection"`
	Rating    int    `json:"rating" binding:"required" validate:"rating"`
}
