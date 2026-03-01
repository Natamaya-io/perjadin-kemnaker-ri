package seeder

import (
	"fmt"
	"log"
	"strings"

	"github.com/kemnaker/perjadin-backend/internal/models"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

func Seed(db *gorm.DB) {
	log.Println("Starting Database Seeding...")

	seedProvincesAndRates(db)
	seedUsers(db)

	log.Println("Database Seeding Completed Successfully.")
}

func seedProvincesAndRates(db *gorm.DB) {
	log.Println("Seeding Provinces and SBM Rates...")

	provinces := []struct {
		Name string
		Code string
		// Base Rates for generating dummy SBM data (Uang Harian / Hotel)
		BaseRate float64 
	}{
		{"ACEH", "11", 360000},
		{"SUMATERA UTARA", "12", 370000},
		{"SUMATERA BARAT", "13", 380000},
		{"RIAU", "14", 370000},
		{"JAMBI", "15", 370000},
		{"SUMATERA SELATAN", "16", 380000},
		{"BENGKULU", "17", 380000},
		{"LAMPUNG", "18", 380000},
		{"KEPULAUAN BANGKA BELITUNG", "19", 410000},
		{"KEPULAUAN RIAU", "21", 420000},
		{"DKI JAKARTA", "31", 530000},
		{"JAWA BARAT", "32", 430000},
		{"JAWA TENGAH", "33", 370000},
		{"DI YOGYAKARTA", "34", 420000},
		{"JAWA TIMUR", "35", 410000},
		{"BANTEN", "36", 370000},
		{"BALI", "51", 480000},
		{"NUSA TENGGARA BARAT", "52", 440000},
		{"NUSA TENGGARA TIMUR", "53", 430000},
		{"KALIMANTAN BARAT", "61", 380000},
		{"KALIMANTAN TENGAH", "62", 360000},
		{"KALIMANTAN SELATAN", "63", 380000},
		{"KALIMANTAN TIMUR", "64", 430000},
		{"KALIMANTAN UTARA", "65", 430000},
		{"SULAWESI UTARA", "71", 370000},
		{"SULAWESI TENGAH", "72", 370000},
		{"SULAWESI SELATAN", "73", 430000},
		{"SULAWESI TENGGARA", "74", 380000},
		{"GORONTALO", "75", 370000},
		{"SULAWESI BARAT", "76", 410000},
		{"MALUKU", "81", 380000},
		{"MALUKU UTARA", "82", 430000},
		{"PAPUA BARAT", "91", 480000},
		{"PAPUA", "92", 580000},
		{"PAPUA SELATAN", "93", 580000},
		{"PAPUA TENGAH", "94", 580000},
		{"PAPUA PEGUNUNGAN", "95", 580000},
		{"PAPUA BARAT DAYA", "96", 480000},
	}

	for _, p := range provinces {
		var province models.Province
		// Upsert Province
		if err := db.Where("code = ?", p.Code).First(&province).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				province = models.Province{Name: p.Name, Code: p.Code}
				db.Create(&province)
			}
		} else {
			if province.Name != p.Name {
				province.Name = p.Name
				db.Save(&province)
			}
		}

		// Seed SBM Rate for 2025
		year := 2025
		var sbm models.SBMRate
		if err := db.Where("province_id = ? AND year = ?", province.ID, year).First(&sbm).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				// Create simplified logic for rates based on BaseRate
				// These are APPROXIMATIONS for demo/staging purposes.
				
				base := p.BaseRate
				
				sbm = models.SBMRate{
					ProvinceID:      province.ID,
					Year:            year,
					
					// Uang Harian
					OutsideCityRate: base, // Luar Kota
					InsideCityRate:  base * 0.4, // Dalam Kota > 8 Jam ~40%
					DiklatRate:      base * 0.3, // Diklat ~30%
					FullboardRate:   base * 0.4, // Fullboard (Paket Meeting)
					FullhalfRate:    base * 0.6, // Fullhalf
					
					// Hotel (Pagu Tertinggi) - Estimasi
					HotelEchelon1:   base * 10,  // ~4jt - 5jt
					HotelEchelon2:   base * 5,   // ~2jt - 3jt
					HotelEchelon3:   base * 3,   // ~1jt - 2jt
					HotelEchelon4:   base * 2.5, // ~800k - 1jt
					HotelStaff:      base * 2,   // ~600k - 800k
					
					TaxiRate:        150000, // Flat average
				}
				db.Create(&sbm)
			}
		}
	}
}

func seedUsers(db *gorm.DB) {
	log.Println("Seeding Users...")

	// Default Password "123"
	password := "123"
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	hashedPwdStr := string(hashedPassword)

	// Admin Users
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

	// Protokol Users Data
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
		{"Amsari B Dulmuti", "-", "-", "-", "-", "Staf Tata Usaha"},
		{"Sigit Santoso", "-", "-", "-", "-", "Staf Tata Usaha"},
		{"Ramadhan Putra Herdian", "-", "-", "-", "-", "Staf Tata Usaha"},
		{"Beni Sanjaya", "-", "-", "-", "-", "Staf Tata Usaha"},
		{"Suratno", "-", "-", "-", "-", "Tenaga Administrasi"},
		{"M. Muhtadin", "-", "-", "-", "-", "Tenaga Administrasi"},
		{"Muhammad Isa", "19871017 201902 1 003", "C", "Penata Muda Tk.I", "III/c", "Analis Perencanaan Evaluasi dan Pelaporan"},
		{"Perananta Purba", "19910306 201902 1 003", "C", "Penata Muda Tk. I", "II/b", "Analis Protokoler"},
		{"Mochamad Gufron", "19940526 201902 1 003", "C", "Penata Muda Tk. I", "III/b", "Kepala Subbagian Protokol"},
		{"Rezky Aries Munandar", "19960407 201812 1 001", "D", "Penata Muda", "III/a", "Penata Protokoler"},
		{"Imelda Anggraeni Sibarani", "19941004 201902 2 008", "C", "Penata Muda Tk. I", "III/b", "Analis Protokoler"},
		{"Efi Kurniawati", "19920106 201503 2004", "C", "Penata", "III/c", "Analis Protokoler"},
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
		{"Syamazka Zakirni", "19950512 202521 2 042", "D", "-", "IX", "Penata Layanan Operasional"},
		{"Zainal Hafit", "19930609 202521 1 068", "D", "-", "IX", "Pengadministrasi Perkantoran"},
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
		{"Hendi Rionaldo", "19870518 202521 1010", "-", "-", "IX", "Penata Layanan Operasional"},
		{"Chandra Hakim", "-", "D", "-", "-", "Pengadministrasi"},
		{"Doni Renaldi", "-", "-", "-", "-", "Staf Tata Usaha"},
		{"Dhika Nur Khaliffa", "-", "-", "-", "-", "Pramu Pimpinan"},
	}

	// Seed Admins
	for _, u := range adminUsers {
		upsertUser(db, u)
	}

	// Seed Protokol Users
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
		upsertUser(db, user)
	}
}

func upsertUser(db *gorm.DB, u models.User) {
	var existing models.User
	if err := db.Where("email = ?", u.Email).First(&existing).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			if err := db.Create(&u).Error; err != nil {
				log.Printf("Failed to create user %s: %v", u.Email, err)
			} else {
				log.Printf("Created user: %s", u.Email)
			}
		}
	} else {
		// Update existing user details
		updates := map[string]interface{}{
			"Name":         u.Name,
			"Role":         u.Role,
			"Jabatan":      u.Jabatan,
			"Pangkat":      u.Pangkat,
			"Golongan":     u.Golongan,
			"TingkatBiaya": u.TingkatBiaya,
			"NIP":          u.NIP,
			// Do NOT update Password to prevent locking out real users if they changed it
			// "Password": u.Password, 
		}
		db.Model(&existing).Updates(updates)
		// log.Printf("Updated user: %s", u.Email)
	}
}
