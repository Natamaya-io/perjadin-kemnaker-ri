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
	log.Println("Seeding users...")

	// Clear out old mock protokol users to prevent duplicates/garbage
	// Ignore errors since there might be foreign key constraints if travel records exist.
	if err := db.Unscoped().Where("role = ?", "protokol").Delete(&models.User{}).Error; err != nil {
		log.Printf("Could not delete existing protokol users (might have travel records): %v", err)
	}

	// Default Password "123"
	password := "123"
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	hashedPwdStr := string(hashedPassword)

	users := []models.User{
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

	type rawUser struct {
		Name         string
		NIP          string
		TingkatBiaya string
		Pangkat      string
		Golongan     string
		Jabatan      string
	}

	protokolUsers := []rawUser{
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

	for i, raw := range protokolUsers {
		email := strings.ToLower(strings.ReplaceAll(raw.Name, " ", "")) + "@kemnaker.go.id"
		// Remove dots in email if any to make it a standard alias
		email = strings.ReplaceAll(email, ".", "")

		users = append(users, models.User{
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
		})
	}

	for _, u := range users {
		var existing models.User
		if err := db.Where("email = ?", u.Email).First(&existing).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				if err := db.Create(&u).Error; err != nil {
					log.Printf("Failed to create user %s: %v", u.Email, err)
				} else {
					log.Printf("Created user: %s", u.Email)
				}
			} else {
				log.Printf("Error checking user %s: %v", u.Email, err)
			}
		} else {
			// Update existing user details just in case they were modified
			db.Model(&existing).Updates(u)
			log.Printf("Updated user: %s", u.Email)
		}
	}

	log.Println("Seeding complete.")
}
