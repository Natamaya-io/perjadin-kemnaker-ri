package seeder

import (
	"database/sql"
	"fmt"
	"log"
	"strings"
	"sync"

	"github.com/brianvoe/gofakeit/v6"
	"github.com/google/uuid"
	"github.com/kemnaker/perjadin-backend/internal/domain/record"
	"github.com/kemnaker/perjadin-backend/internal/domain/user"
	"github.com/kemnaker/perjadin-backend/internal/models"
	"golang.org/x/crypto/bcrypt"
)

func Seed(db *sql.DB) {
	log.Println("Starting Database Seeding...")

	userRepo := user.NewRepository(db)
	recordRepo := record.NewRepository(db)

	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		defer wg.Done()
		seedProvincesAndRates(db)
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		seedUsers(userRepo)
	}()

	wg.Wait()

	// Runs after users are seeded because it depends on them
	seedFakerRecords(userRepo, recordRepo)

	log.Println("Database Seeding Completed Successfully.")
}

func seedProvincesAndRates(db *sql.DB) {
	log.Println("Seeding Provinces and SBM Rates...")

	provinces := []struct {
		Name      string
		Code      string
		LuarKota  float64
		DalamKota float64
		Diklat    float64
	}{
		{"ACEH", "11", 360000, 140000, 110000},
		{"DKI JAKARTA", "31", 530000, 210000, 160000},
		{"JAWA BARAT", "32", 430000, 170000, 130000},
		{"JAWA TENGAH", "33", 370000, 150000, 110000},
		{"DI YOGYAKARTA", "34", 420000, 170000, 130000},
		{"JAWA TIMUR", "35", 410000, 160000, 120000},
		{"BALI", "51", 480000, 190000, 140000},
		{"PAPUA", "92", 580000, 230000, 170000},
	}

	for _, p := range provinces {
		var provinceID uuid.UUID
		err := db.QueryRow("SELECT id FROM provinces WHERE code = $1", p.Code).Scan(&provinceID)
		if err != nil {
			if err == sql.ErrNoRows {
				provinceID = uuid.New()
				_, err = db.Exec("INSERT INTO provinces (id, name, code) VALUES ($1, $2, $3)", provinceID, p.Name, p.Code)
				if err != nil {
					log.Printf("Failed to insert province %s: %v", p.Name, err)
					continue
				}
			} else {
				log.Printf("Error querying province %s: %v", p.Name, err)
				continue
			}
		}

		year := 2025
		var sbmID uuid.UUID
		err = db.QueryRow("SELECT id FROM sbm_rates WHERE province_id = $1 AND year = $2", provinceID, year).Scan(&sbmID)
		if err != nil {
			if err == sql.ErrNoRows {
				_, err = db.Exec(`
					INSERT INTO sbm_rates (
						id, province_id, year, fullboard_rate, fullhalf_rate, outside_city_rate, inside_city_rate, diklat_rate, hotel_echelon1, hotel_echelon2, hotel_echelon3, hotel_echelon4, hotel_staff, taxi_rate
					) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14)`,
					uuid.New(), provinceID, year, p.LuarKota*0.4, p.LuarKota*0.6, p.LuarKota, p.DalamKota, p.Diklat, p.LuarKota*10, p.LuarKota*5, p.LuarKota*3, p.LuarKota*2.5, p.LuarKota*2, 150000,
				)
				if err != nil {
					log.Printf("Failed to insert SBM rate for %s: %v", p.Name, err)
				}
			}
		}
	}
}

func seedUsers(userRepo user.Repository) {
	log.Println("Seeding Users...")

	password := "123"
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	hashedPwdStr := string(hashedPassword)

	adminUsers := []models.User{
		{
			Name:         "Super Admin",
			Email:        "superadmin@kemnaker.go.id",
			Password:     hashedPwdStr,
			Role:         "super_admin",
			DemoPassword: password,
			NomorHP:      "081200000000",
		},
		{
			Name:         "Keuangan",
			Email:        "keuangan@kemnaker.go.id",
			Password:     hashedPwdStr,
			Role:         "keuangan",
			DemoPassword: password,
			NomorHP:      "081200000001",
		},
	}

	type rawUser struct {
		Name         string
		NIP          string
		TingkatBiaya string
		Pangkat      string
		Golongan     string
		Jabatan      string
	}

	protokolData := []rawUser{
		{"Jiyanto", "-", "-", "-", "-", "Petugas Pamwal Menteri Ketenagakerjaan"},
		{"Fathan Asyraf", "-", "-", "-", "-", "Staf Tata Usaha"},
		{"Auditya Hermawan", "19880920 201403 1 001", "C", "Penata Tk.I", "III/d", "Kabag TU Pimpinan dan Protokol"},
	}

	for _, u := range adminUsers {
		upsertUser(userRepo, u)
	}

	for i, raw := range protokolData {
		email := strings.ToLower(strings.ReplaceAll(raw.Name, " ", "")) + "@kemnaker.go.id"
		email = strings.ReplaceAll(email, ".", "")

		user := models.User{
			Name:         raw.Name,
			Email:        email,
			Password:     hashedPwdStr,
			Role:         "protokol",
			DemoPassword: password,
			Jabatan:      raw.Jabatan,
			Pangkat:      raw.Pangkat,
			Golongan:     raw.Golongan,
			TingkatBiaya: raw.TingkatBiaya,
			NIP:          raw.NIP,
			NomorHP:      fmt.Sprintf("081234567%03d", i+1),
		}
		upsertUser(userRepo, user)
	}

	gofakeit.Seed(0)
	for i := 0; i < 5; i++ {
		email := gofakeit.Email()
		user := models.User{
			Name:         gofakeit.Name(),
			Email:        email,
			Password:     hashedPwdStr,
			Role:         "protokol",
			DemoPassword: password,
			Jabatan:      gofakeit.JobTitle(),
			NIP:          gofakeit.DigitN(18),
			NomorHP:      gofakeit.Phone(),
		}
		upsertUser(userRepo, user)
	}
}

func upsertUser(userRepo user.Repository, u models.User) {
	existing, err := userRepo.GetUserByEmail(u.Email)
	if err != nil || existing == nil {
		if err := userRepo.CreateUser(&u); err != nil {
			log.Printf("Failed to create user %s: %v", u.Email, err)
		} else {
			log.Printf("Created user: %s", u.Email)
		}
	} else {
		u.ID = existing.ID
		u.Password = existing.Password // Keep existing password
		if err := userRepo.UpdateUser(&u); err != nil {
			log.Printf("Failed to update user %s: %v", u.Email, err)
		}
	}
}

func seedFakerRecords(userRepo user.Repository, recordRepo record.Repository) {
	log.Println("Seeding Faker Travel Records...")
	gofakeit.Seed(0)

	users, err := userRepo.GetUsers()
	if err != nil || len(users) < 2 {
		return
	}

	var creatorID uuid.UUID
	var empID uuid.UUID

	for _, u := range users {
		if u.Role == "super_admin" {
			creatorID = u.ID
		}
		if u.Role == "protokol" && empID == uuid.Nil {
			empID = u.ID
		}
	}

	var wg sync.WaitGroup
	for i := 0; i < 10; i++ {
		wg.Add(1)
		
		// To avoid race conditions in gofakeit and ensure unique generation,
		// we generate data sequentially but insert concurrently.
		record := models.TravelRecord{
			SPDNumber:   gofakeit.UUID(),
			EmployeeID:  empID,
			CreatorID:   creatorID,
			Location:    gofakeit.City(),
			Province:    gofakeit.State(),
			Type:        "luar_kota",
			Purpose:     gofakeit.Sentence(5),
			Stakeholder: gofakeit.Company(),
			Agenda:      gofakeit.Paragraph(1, 2, 5, " "),
			Status:      "Draft",
		}
		
		go func(rec models.TravelRecord) {
			defer wg.Done()
			recordRepo.CreateTravelRecord(&rec)
		}(record)
	}
	wg.Wait()
}
