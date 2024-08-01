package db

import (
	"fmt"
	"time"

	"github.com/Bengkelin/bengkelin-service/internal/pkg/config"
	"github.com/Bengkelin/bengkelin-service/internal/pkg/models"
	"github.com/Bengkelin/bengkelin-service/pkg/helpers"
	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var (
	DB  *gorm.DB
	err error
)

// Database instance
type Database struct {
	*gorm.DB
}

// SetupDB is a function to open connection to database
func SetupDB() {
	var db = DB

	configuration := config.GetConfig()

	// Viper Config
	driver := configuration.Database.Driver
	database := configuration.Database.Dbname
	username := configuration.Database.Username
	password := configuration.Database.Password
	host := configuration.Database.Host
	port := configuration.Database.Port

	// Gorm config
	gormConfig := &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	}

	switch driver {
	case "mysql":
		db, err = gorm.Open(mysql.Open(username+":"+password+"@tcp("+host+":"+port+")/"+database+"?charset=utf8&parseTime=True&loc=Local"), gormConfig)
		if err != nil {
			fmt.Println("db err:", err)
		}
	case "postgres":
		db, err = gorm.Open(postgres.Open("host="+host+" port="+port+" user="+username+" dbname="+database+"  sslmode=disable password="+password), gormConfig)
		if err != nil {
			fmt.Println("db err:", err)
		}
	}

	// Set up the connection pools
	sqlDb, _ := db.DB()
	sqlDb.SetMaxIdleConns(configuration.Database.MaxIdleConns)
	sqlDb.SetMaxOpenConns(configuration.Database.MaxOpenConns)
	sqlDb.SetConnMaxLifetime(time.Duration(configuration.Database.MaxLifetime))

	DB = db

	err := migrateTable()
	if err != nil {
		fmt.Println(err)
	}
	//CreateMitraSeeder()
	//createCreatedAtUpdatedAtBengkelModel()
}

// AutoMigrate project models
func migrateTable() error {
	err := DB.AutoMigrate(
		&models.User{},
		&models.Mitra{},
		&models.AddressUser{},
		&models.Vehicle{},
		&models.VehiclePhoto{},
		&models.Bengkel{},
		&models.BengkelPhoto{},
		&models.BengkelOperasional{},
		&models.BengkelAddress{},
		&models.BengkelService{},
	)
	if err != nil {
		return err
	}
	return nil
}

func CreateMitraSeeder() {

	var id []string
	// create 10 mitra using batch insert
	for i := 1; i <= 10; i++ {
		mitra := models.Mitra{
			ID:          helpers.GenerateUUID(),
			FirstName:   "Mitra" + fmt.Sprint(i),
			LastName:    "Bengkelin" + fmt.Sprint(i),
			Email:       "bengkelin" + fmt.Sprint(i) + "@gmail.com",
			Password:    "password",
			PhoneNumber: "08123456789",
			BankName:    "BCA",
			BankNumber:  "1234567890",
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		}
		id = append(id, mitra.ID)
		DB.Create(&mitra)
	}

	var isOpen = true
	var idBengkel []string
	// create 10 bengkel using batch insert
	for i := 1; i <= 10; i++ {
		bengkel := models.Bengkel{
			ID:           helpers.GenerateUUID(),
			MitraID:      id[i-1],
			BengkelName:  "Honda Mitra" + fmt.Sprint(i),
			BengkelPhone: "08123456789",
			JumlahMontir: 3,
			HomeService:  &isOpen,
			StoreService: &isOpen,
			IsOpen:       &isOpen,
		}
		idBengkel = append(idBengkel, bengkel.ID)
		DB.Create(&bengkel)
	}

	// create 10 bengkel operasional using batch insert
	for i := 1; i <= 10; i++ {
		bengkelOperasional := models.BengkelOperasional{
			BengkelID: idBengkel[i-1],
			Hari:      "Senin",
			JamBuka:   "08:00",
		}
		DB.Create(&bengkelOperasional)
	}

	// create 10 bengkel address using batch insert
	for i := 1; i <= 10; i++ {
		bengkelAddress := models.BengkelAddress{
			BengkelID:    idBengkel[i-1],
			Latitude:     -6.193125,
			Longitude:    106.821808,
			AddressLabel: "Jl. Jend. Sudirman",
			FullAddress:  "Jl. Jend. Sudirman No.Kav 54-55, RT.5/RW.3, Senayan, Kec. Kby. Baru, Kota Jakarta Selatan, Daerah Khusus Ibukota Jakarta 12190",
			Note:         "Sebelah Bank BCA",
		}
		DB.Create(&bengkelAddress)
	}

	// create 10 bengkel photo using batch insert
	for i := 1; i <= 10; i++ {
		bengkelPhoto := models.BengkelPhoto{
			BengkelID: idBengkel[i-1],
			PhotoURL:  "http://103.127.136.198/api/v1/static/bengkel/post-1722173668.png",
		}
		DB.Create(&bengkelPhoto)
	}

	// create 10 bengkel service using batch insert
	for i := 1; i <= 10; i++ {
		bengkelService := models.BengkelService{
			BengkelID:   idBengkel[i-1],
			NamaService: "Service" + fmt.Sprint(i),
		}
		DB.Create(&bengkelService)
	}

}

func createCreatedAtUpdatedAtBengkelModel() {
	var bengkels []models.Bengkel
	DB.Find(&bengkels)
	for _, bengkel := range bengkels {
		if bengkel.CreatedAt.IsZero() && bengkel.UpdatedAt.IsZero() {
			bengkel.CreatedAt = time.Now()
			bengkel.UpdatedAt = time.Now().Add(10 * time.Second)
			DB.Save(&bengkel)
		}
	}
}

func GetDB() *gorm.DB {
	return DB
}
