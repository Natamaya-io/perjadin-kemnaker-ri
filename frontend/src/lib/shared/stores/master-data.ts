import { writable } from 'svelte/store';

export const PROVINCES = [
    { name: "ACEH", luarKota: 360000 },
    { name: "SUMATERA UTARA", luarKota: 370000 },
    { name: "SUMATERA BARAT", luarKota: 380000 },
    { name: "RIAU", luarKota: 370000 },
    { name: "JAMBI", luarKota: 370000 },
    { name: "SUMATERA SELATAN", luarKota: 380000 },
    { name: "BENGKULU", luarKota: 380000 },
    { name: "LAMPUNG", luarKota: 380000 },
    { name: "KEPULAUAN BANGKA BELITUNG", luarKota: 410000 },
    { name: "KEPULAUAN RIAU", luarKota: 370000 },
    { name: "DKI JAKARTA", luarKota: 530000 },
    { name: "JAWA BARAT", luarKota: 430000 },
    { name: "JAWA TENGAH", luarKota: 370000 },
    { name: "DI YOGYAKARTA", luarKota: 420000 },
    { name: "JAWA TIMUR", luarKota: 410000 },
    { name: "BANTEN", luarKota: 370000 },
    { name: "BALI", luarKota: 480000 },
    { name: "NUSA TENGGARA BARAT", luarKota: 440000 },
    { name: "NUSA TENGGARA TIMUR", luarKota: 430000 },
    { name: "KALIMANTAN BARAT", luarKota: 380000 },
    { name: "KALIMANTAN TENGAH", luarKota: 360000 },
    { name: "KALIMANTAN SELATAN", luarKota: 380000 },
    { name: "KALIMANTAN TIMUR", luarKota: 430000 },
    { name: "KALIMANTAN UTARA", luarKota: 430000 },
    { name: "SULAWESI UTARA", luarKota: 370000 },
    { name: "SULAWESI TENGAH", luarKota: 370000 },
    { name: "SULAWESI SELATAN", luarKota: 430000 },
    { name: "SULAWESI TENGGARA", luarKota: 380000 },
    { name: "GORONTALO", luarKota: 370000 },
    { name: "SULAWESI BARAT", luarKota: 410000 },
    { name: "MALUKU", luarKota: 380000 },
    { name: "MALUKU UTARA", luarKota: 430000 },
    { name: "PAPUA BARAT", luarKota: 480000 },
    { name: "PAPUA", luarKota: 580000 },
    { name: "PAPUA SELATAN", luarKota: 580000 },
    { name: "PAPUA TENGAH", luarKota: 580000 },
    { name: "PAPUA PEGUNUNGAN", luarKota: 580000 },
    { name: "PAPUA BARAT DAYA", luarKota: 480000 }
];

export const STAKEHOLDERS = [
    "Menteri Ketenagakerjaan",
    "Wakil Menteri Ketenagakerjaan",
    "Pendamping Menteri Ketenagakerjaan",
    "Pendamping Wakil Menteri Ketenagakerjaan"
];

export const provincesStore = writable(PROVINCES);
export const stakeholdersStore = writable(STAKEHOLDERS);
