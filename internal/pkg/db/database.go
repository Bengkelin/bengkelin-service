package db

import (
	"fmt"
	"time"

	"github.com/Bengkelin/bengkelin-service/internal/pkg/config"
	"github.com/Bengkelin/bengkelin-service/internal/pkg/models"
	"github.com/Bengkelin/bengkelin-service/pkg/helpers"
	applog "github.com/Bengkelin/bengkelin-service/pkg/log"
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

	applog.Info("Setting up database connection", "driver", driver, "host", host, "port", port, "database", database)

	switch driver {
	case "mysql":
		applog.Debug("Attempting MySQL connection", "host", host, "port", port, "database", database)
		db, err = gorm.Open(mysql.Open(username+":"+password+"@tcp("+host+":"+port+")/"+database+"?charset=utf8&parseTime=True&loc=Local"), gormConfig)
		if err != nil {
			applog.Error("Failed to connect to MySQL database", "error", err.Error(), "host", host, "database", database)
			return
		}
		applog.Info("✓ Successfully connected to MySQL database", "host", host, "port", port, "database", database)
	case "postgres":
		applog.Debug("Attempting PostgreSQL connection", "host", host, "port", port, "database", database)
		db, err = gorm.Open(postgres.Open("host="+host+" port="+port+" user="+username+" dbname="+database+"  sslmode=disable password="+password), gormConfig)
		if err != nil {
			applog.Error("Failed to connect to PostgreSQL database", "error", err.Error(), "host", host, "database", database)
			return
		}
		applog.Info("✓ Successfully connected to PostgreSQL database", "host", host, "port", port, "database", database)
	default:
		applog.Error("Unsupported database driver", "driver", driver)
		return
	}

	// Verify DB connection is valid
	if db == nil {
		applog.Error("Database connection is nil after opening")
		return
	}
	applog.Debug("Database connection object created successfully")

	// Set up the connection pools
	sqlDb, _ := db.DB()
	if sqlDb != nil {
		sqlDb.SetMaxIdleConns(configuration.Database.MaxIdleConns)
		sqlDb.SetMaxOpenConns(configuration.Database.MaxOpenConns)
		sqlDb.SetConnMaxLifetime(time.Duration(configuration.Database.MaxLifetime))

		applog.Info("✓ Database connection pool configured",
			"maxIdleConns", configuration.Database.MaxIdleConns,
			"maxOpenConns", configuration.Database.MaxOpenConns,
			"maxLifetime", configuration.Database.MaxLifetime)
	} else {
		applog.Error("Failed to get underlying SQL database")
		return
	}

	DB = db

	err = migrateTable()
	if err != nil {
		applog.Error("Failed to migrate database tables", "error", err.Error())
		return
	}
	applog.Info("✓ Database tables migrated successfully")
	//CreateMitraSeeder()
	//CreateCreatedAtUpdatedAtBengkelModel()
}

// AutoMigrate project models
func migrateTable() error {
	applog.Info("🔄 Starting database table auto-migration...")
	applog.Debug("Tables to migrate: User, Mitra, AddressUser, Vehicle, VehiclePhoto, Bengkel, BengkelPhoto, BengkelOperasional, BengkelAddress, BengkelService, BengkelTestimoni, ChatHistory, Pesanan, PesananService, AdminFee")

	// Verify DB is not nil
	if DB == nil {
		applog.Error("❌ Cannot migrate: Database connection (DB) is nil")
		return fmt.Errorf("database connection is nil")
	}

	err := DB.AutoMigrate(
		&models.User{},
		&models.Mitra{},
		&models.UserAddress{},
		&models.Vehicle{},
		&models.VehiclePhoto{},
		&models.Bengkel{},
		&models.BengkelPhoto{},
		&models.BengkelOperational{},
		&models.BengkelAddress{},
		&models.BengkelService{},
		&models.BengkelTestimonial{},
		&models.ChatHistory{},
		&models.Order{},
		&models.OrderService{},
		&models.AdminFee{},
		&models.RefreshToken{},
	)

	if err != nil {
		applog.Error("❌ Database auto-migration FAILED", "error", err.Error())
		return err
	}

	applog.Info("✅ Database auto-migration completed successfully")
	return nil
}

func CreateMitraSeeder() {
	applog.Info("Starting Mitra seeder")
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
		result := DB.Create(&mitra)
		if result.Error != nil {
			applog.Error("Failed to create mitra", "mitra_id", mitra.ID, "error", result.Error.Error())
		}
	}
	applog.Info("Mitra seeder completed", "count", len(id))

	var isOpen = true
	var idBengkel []string
	// create 10 bengkel using batch insert
	applog.Info("Starting Bengkel seeder")
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
		result := DB.Create(&bengkel)
		if result.Error != nil {
			applog.Error("Failed to create bengkel", "bengkel_id", bengkel.ID, "error", result.Error.Error())
		}
	}
	applog.Info("Bengkel seeder completed", "count", len(idBengkel))

	// create 10 bengkel operasional using batch insert
	applog.Info("Starting Bengkel Operasional seeder")
	for i := 1; i <= 10; i++ {
		bengkelOperational := models.BengkelOperational{
			BengkelID: idBengkel[i-1],
			Hari:      "Senin",
			JamBuka:   "08:00",
		}
		result := DB.Create(&bengkelOperational)
		if result.Error != nil {
			applog.Error("Failed to create bengkel operasional", "bengkel_id", idBengkel[i-1], "error", result.Error.Error())
		}
	}
	applog.Info("Bengkel Operasional seeder completed")

	// create 10 bengkel address using batch insert
	applog.Info("Starting Bengkel Address seeder")
	for i := 1; i <= 10; i++ {
		bengkelAddress := models.BengkelAddress{
			BengkelID:    idBengkel[i-1],
			Latitude:     -6.193125,
			Longitude:    106.821808,
			AddressLabel: "Jl. Jend. Sudirman",
			FullAddress:  "Jl. Jend. Sudirman No.Kav 54-55, RT.5/RW.3, Senayan, Kec. Kby. Baru, Kota Jakarta Selatan, Daerah Khusus Ibukota Jakarta 12190",
			Note:         "Sebelah Bank BCA",
		}
		result := DB.Create(&bengkelAddress)
		if result.Error != nil {
			applog.Error("Failed to create bengkel address", "bengkel_id", idBengkel[i-1], "error", result.Error.Error())
		}
	}
	applog.Info("Bengkel Address seeder completed")

	// create 10 bengkel photo using batch insert
	applog.Info("Starting Bengkel Photo seeder")
	for i := 1; i <= 10; i++ {
		bengkelPhoto := models.BengkelPhoto{
			BengkelID: idBengkel[i-1],
			PhotoURL:  "http://103.127.136.198/api/v1/static/bengkel/post-1722173668.png",
		}
		result := DB.Create(&bengkelPhoto)
		if result.Error != nil {
			applog.Error("Failed to create bengkel photo", "bengkel_id", idBengkel[i-1], "error", result.Error.Error())
		}
	}
	applog.Info("Bengkel Photo seeder completed")

	// create 10 bengkel service using batch insert
	applog.Info("Starting Bengkel Service seeder")
	for i := 1; i <= 10; i++ {
		bengkelService := models.BengkelService{
			BengkelID:   idBengkel[i-1],
			NamaService: "Service" + fmt.Sprint(i),
		}
		result := DB.Create(&bengkelService)
		if result.Error != nil {
			applog.Error("Failed to create bengkel service", "bengkel_id", idBengkel[i-1], "error", result.Error.Error())
		}
	}
	applog.Info("Bengkel Service seeder completed")
}

func CreateCreatedAtUpdatedAtBengkelModel() {
	applog.Info("Starting CreateCreatedAtUpdatedAtBengkelModel")
	var bengkels []models.Bengkel
	result := DB.Find(&bengkels)
	if result.Error != nil {
		applog.Error("Failed to fetch bengkels", "error", result.Error.Error())
		return
	}
	applog.Debug("Fetched bengkels for timestamp update", "count", len(bengkels))

	updated := 0
	for _, bengkel := range bengkels {
		if bengkel.CreatedAt.IsZero() && bengkel.UpdatedAt.IsZero() {
			bengkel.CreatedAt = time.Now()
			bengkel.UpdatedAt = time.Now().Add(10 * time.Second)
			saveResult := DB.Save(&bengkel)
			if saveResult.Error != nil {
				applog.Error("Failed to update bengkel timestamps", "bengkel_id", bengkel.ID, "error", saveResult.Error.Error())
			} else {
				updated++
			}
		}
	}
	applog.Info("CreateCreatedAtUpdatedAtBengkelModel completed", "updated", updated, "total", len(bengkels))
}

func GetDB() *gorm.DB {
	return DB
}
