package seeder

import (
	"database/sql"
	"fmt"
	"log"
	"strings"
	"sync"

	"github.com/google/uuid"
	"github.com/kemnaker/perjadin-backend/internal/domain/user"
	"github.com/kemnaker/perjadin-backend/internal/models"
	"golang.org/x/crypto/bcrypt"
)

type rawUser struct {
	Name         string
	NIP          string
	TingkatBiaya string
	Pangkat      string
	Golongan     string
	Jabatan      string
}

func getOriginalUsers() ([]models.User, []rawUser) {
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
		{
			Name:         "Kasubag",
			Email:        "kasubag@kemnaker.go.id",
			Password:     hashedPwdStr,
			Role:         "kasubag",
			DemoPassword: password,
			NomorHP:      "081200000002",
		},
	}

	protokolData := []rawUser{
		{"Jiyanto", "-", "-", "-", "-", "Petugas Pamwal Menteri Ketenagakerjaan"},
		{"Fathan Asyraf", "-", "-", "-", "-", "Staf Tata Usaha"},
		{"Auditya Hermawan", "19880920 201403 1 001", "C", "Penata Tk.I", "III/d", "Kabag TU Pimpinan dan Protokol"},
		{"Amsari B Dulmuti", "-", "-", "-", "-", "Staf Tata Usaha"},
		{"Sigit Santoso", "-", "-", "-", "-", "Staf Tata Usaha"},
		{"Ramadhan Putra Herdian", "-", "-", "-", "-", "Staf Tata Usaha"},
		{"Beni Sanjaya", "-", "-", "-", "-", "Staf Tata Usaha"},
		{"Suratno", "-", "-", "-", "-", "Tenaga Administrasi"},
		{"M. Muhtadin", "-", "-", "-", "-", "Tenaga Administrasi"},
		{"Muhammad Isa", "19871017 201902 1 003", "C", "Penata Muda Tk.I", "III/c", "Analis Perencanaan Evaluasi dan Pelaporan"},
		{"Perananta Purba", "19910306 201902 1 003", "C", "Penata Muda Tk. I", "III/b", "Analis Protokoler"},
		{"Mochamad Gufron", "19940526 201902 1 003", "C", "Penata Muda Tk. I", "III/b", "Kepala Subbagian Protokol"},
		{"Rezky Aries Munandar", "19960407 201812 1 001", "D", "Penata Muda", "III/a", "Penata Protokoler"},
		{"Imelda Anggraeni Sibarani", "19941004 201902 2 008", "C", "Penata Muda Tk. I", "III/b", "Analis Protokoler"},
		{"Efi Kurniawati", "19920106 201503 2 004", "C", "Penata", "III/c", "Analis Protokoler"},
		{"Bobby Rizky", "19940929 201902 1 005", "C", "Penata Muda Tk. I", "III/b", "Analis Protokoler"},
		{"Nurcahyo Purnomo", "19890404 201503 1 007", "D", "Penata Muda", "III/a", "Petugas Protokoler"},
		{"Bagas Winektu", "-", "-", "-", "-", "Staf Tata Usaha"},
		{"Yudi Santoso", "80090538", "-", "AIPTU", "-", "Petugas Pamwal Menteri Ketenagakerjaan"},
		{"M. Choirul Hidayat", "-", "-", "-", "-", "Petugas Pamwal Menteri Ketenagakerjaan"},
		{"Widada", "75120659", "-", "AIPDA", "-", "Petugas Pamwal Menteri Ketenagakerjaan"},
		{"Nanang", "77060070", "-", "AIPDA", "-", "Petugas Pamwal Menteri Ketenagakerjaan"},
		{"Beny Sanjaya", "-", "-", "-", "-", "Tenaga Administrasi"},
		{"Muhammad Dienul Islami", "-", "D", "-", "-", "Pramu Pimpinan"},
		{"Wahyu Nino Prasangka", "-", "D", "-", "-", "Pramu Pimpinan"},
		{"Firman Andriansyah", "-", "D", "-", "-", "Pramu Pimpinan"},
		{"Muhammad Afendrianto", "-", "D", "-", "-", "Pramu Pimpinan"},
		{"Dudi Erwanto", "-", "D", "-", "-", "Pramu Pimpinan"},
		{"Afriyanti", "-", "D", "-", "-", "Pramu Pimpinan"},
		{"Muhammad Farras Fadhilsyah", "-", "D", "-", "-", "Pramu Pimpinan"},
		{"Siti Munzayanah", "19930812 202012 2 021", "C", "Penata Muda", "III/a", "Penelaah Teknis Kebijakan"},
		{"Syamazka Zakirni", "19950512 202521 2 042", "D", "IX", "-", "Penata Layanan Operasional"},
		{"Zainal Hafit", "19930609 202521 1 068", "D", "IX", "-", "Pengadministrasi Perkantoran"},
		{"Taufik Hidayat Sitompul", "198603272009121003", "C", "Penata Muda Tk. I", "III/b", "Analis Persuratan"},
		{"Mark Hermawan", "-", "D", "-", "-", "Tenaga Administrasi"},
		{"Widianto", "-", "D", "-", "-", "Tenaga Administrasi"},
		{"Imam Wahyu Sucipto", "-", "-", "-", "-", "ADC Menteri Ketenagakerjaan"},
		{"Muhammad Nuril Anwar", "-", "D", "-", "-", "Tenaga Administrasi"},
		{"Adria Jabartaru Putra", "19890920 201503 1 003", "-", "-", "-", "Auditor muda inspektorat 1"},
		{"Heru Anggara Tri Susila", "-", "-", "-", "-", "Tenaga Administrasi"},
		{"Nurin Nashfati", "20010115 202505 2 002", "C", "Penata Muda", "III/a", "Penata Keprotokolan"},
		{"Citra Anastasya", "20010728 202505 2 007", "C", "Penata Muda", "III/a", "Penata Keprotokolan"},
		{"Riki Nurkamal Arsandi", "-", "D", "-", "-", "Staf Biro Umum"},
		{"Mohamad Abdul Baasith", "-", "D", "-", "-", "Petugas Administrasi"},
		{"Regina Dwita Sari", "20020627 202505 2 004", "C", "Penata Muda", "III/a", "Penata Protokoler"},
		{"Hendi Rionaldo", "19870518 202521 1 010", "-", "IX", "-", "Penata Layanan Operasional"},
		{"Chandra Hakim", "-", "D", "-", "-", "Pengadministrasi"},
		{"Doni Renaldi", "-", "-", "-", "-", "Staf Tata Usaha"},
		{"Dhika Nur Khaliffa", "-", "-", "-", "-", "Pramu Pimpinan"},
	}
	return adminUsers, protokolData
}

func Seed(db *sql.DB) {
	log.Println("Starting Database Seeding...")

	adminUsers, protokolData := getOriginalUsers()
	
	userRepo := user.NewRepository(db)

	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		defer wg.Done()
		seedProvincesAndRates(db)
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		seedUsers(userRepo, adminUsers, protokolData)
	}()

	wg.Wait()

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
		{"ACEH", "01", 360000, 140000, 110000},
		{"SUMATRA UTARA", "02", 370000, 150000, 110000},
		{"RIAU", "03", 370000, 150000, 110000},
		{"KEPULAUAN RIAU", "04", 370000, 150000, 110000},
		{"JAMBI", "05", 370000, 150000, 110000},
		{"SUMATRA BARAT", "06", 380000, 150000, 110000},
		{"SUMATRA SELATAN", "07", 380000, 150000, 110000},
		{"LAMPUNG", "08", 380000, 150000, 110000},
		{"BENGKULU", "09", 380000, 150000, 110000},
		{"BANGKA BELITUNG", "10", 410000, 160000, 120000},
		{"BANTEN", "11", 370000, 150000, 110000},
		{"JAWA BARAT", "12", 430000, 170000, 130000},
		{"DKI JAKARTA", "13", 530000, 210000, 160000},
		{"JAWA TENGAH", "14", 370000, 150000, 110000},
		{"DI YOGYAKARTA", "15", 420000, 170000, 130000},
		{"JAWA TIMUR", "16", 410000, 160000, 120000},
		{"BALI", "17", 480000, 190000, 140000},
		{"NUSA TENGGARA BARAT", "18", 440000, 180000, 130000},
		{"NUSA TENGGARA TIMUR", "19", 430000, 170000, 130000},
		{"KALIMANTAN BARAT", "20", 380000, 150000, 110000},
		{"KALIMANTAN TENGAH", "21", 360000, 140000, 110000},
		{"KALIMANTAN SELATAN", "22", 380000, 150000, 110000},
		{"KALIMANTAN TIMUR", "23", 430000, 170000, 130000},
		{"KALIMANTAN UTARA", "24", 430000, 170000, 130000},
		{"SULAWESI UTARA", "25", 370000, 150000, 110000},
		{"GORONTALO", "26", 370000, 150000, 110000},
		{"SULAWESI BARAT", "27", 410000, 160000, 120000},
		{"SULAWESI SELATAN", "28", 430000, 170000, 130000},
		{"SULAWESI TENGAH", "29", 370000, 150000, 110000},
		{"SULAWESI TENGGARA", "30", 380000, 150000, 110000},
		{"MALUKU", "31", 380000, 150000, 110000},
		{"MALUKU UTARA", "32", 430000, 170000, 130000},
		{"PAPUA", "33", 580000, 230000, 170000},
		{"PAPUA BARAT", "34", 480000, 190000, 140000},
		{"PAPUA BARAT DAYA", "35", 480000, 190000, 140000},
		{"PAPUA TENGAH", "36", 580000, 230000, 170000},
		{"PAPUA SELATAN", "37", 580000, 230000, 170000},
		{"PAPUA PEGUNUNGAN", "38", 580000, 230000, 170000},
	}

	for _, p := range provinces {
		var provinceID uuid.UUID
		err := db.QueryRow("SELECT id FROM provinces WHERE name = $1", p.Name).Scan(&provinceID)
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
		} else {
			// Update code if necessary
			_, err = db.Exec("UPDATE provinces SET code = $1 WHERE id = $2", p.Code, provinceID)
			if err != nil {
				log.Printf("Failed to update province code %s: %v", p.Name, err)
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
		} else {
			// Update SBM rate if necessary
			_, err = db.Exec(`
				UPDATE sbm_rates SET 
					fullboard_rate=$1, fullhalf_rate=$2, outside_city_rate=$3, inside_city_rate=$4, diklat_rate=$5, hotel_echelon1=$6, hotel_echelon2=$7, hotel_echelon3=$8, hotel_echelon4=$9, hotel_staff=$10, taxi_rate=$11
				WHERE id = $12`,
				p.LuarKota*0.4, p.LuarKota*0.6, p.LuarKota, p.DalamKota, p.Diklat, p.LuarKota*10, p.LuarKota*5, p.LuarKota*3, p.LuarKota*2.5, p.LuarKota*2, 150000, sbmID,
			)
			if err != nil {
				log.Printf("Failed to update SBM rate for %s: %v", p.Name, err)
			}
		}
	}
}

func seedUsers(userRepo user.Repository, adminUsers []models.User, protokolData []rawUser) {
	log.Println("Seeding Users...")
	password := "123"
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	hashedPwdStr := string(hashedPassword)

	for _, u := range adminUsers {
		upsertUser(userRepo, u)
	}

	for i, raw := range protokolData {
		namePart := strings.ToLower(strings.ReplaceAll(raw.Name, " ", ""))
		namePart = strings.ReplaceAll(namePart, ".", "")
		email := namePart + "@kemnaker.go.id"

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
