import type { ApiClient, Employee, TravelRecord, User } from './types';
import { browser } from '$app/environment';

// Mock Data Generators (copied from original stores)
const LOCATIONS = [
    { city: "Surabaya", prov: "Jawa Timur" }, { city: "Bandung", prov: "Jawa Barat" },
    { city: "Semarang", prov: "Jawa Tengah" }, { city: "Yogyakarta", prov: "DI Yogyakarta" },
    { city: "Medan", prov: "Sumatera Utara" }, { city: "Makassar", prov: "Sulawesi Selatan" },
    { city: "Denpasar", prov: "Bali" }, { city: "Balikpapan", prov: "Kalimantan Timur" },
    { city: "Banjarmasin", prov: "Kalimantan Selatan" }, { city: "Palembang", prov: "Sumatera Selatan" },
    { city: "Padang", prov: "Sumatera Barat" }, { city: "Manado", prov: "Sulawesi Utara" },
    { city: "Ambon", prov: "Maluku" }, { city: "Jayapura", prov: "Papua" },
    { city: "Mataram", prov: "Nusa Tenggara Barat" }
];

const PURPOSES = [
    "Persiapan dan Pendampingan Kunjungan Kerja", "Koordinasi dan Konsultasi Kunjungan Kerja",
    "Monitoring dan Evaluasi Program", "Bimbingan Teknis Ketenagakerjaan", "Rapat Koordinasi Nasional"
];

const DEFAULT_EMPLOYEES: Employee[] = [
    { id: 1, name: "Budi Santoso", nip: "198501012010011001", rank: "Penata", golongan: "III/c" },
    { id: 2, name: "Siti Aminah", nip: "199002022015022002", rank: "Penata Muda", golongan: "III/a" },
    { id: 3, name: "Ahmad Fauzi", nip: "198003032005031003", rank: "Pembina", golongan: "IV/a" },
    { id: 4, name: "Dewi Ratna", nip: "199504042020042004", rank: "Pengatur", golongan: "II/c" },
    { id: 5, name: "Eko Prasetyo", nip: "198805052012051005", rank: "Penata Tingkat I", golongan: "III/d" },
    { id: 6, name: "Rina Wati", nip: "199206062016062006", rank: "Penata Muda Tk. I", golongan: "III/b" },
    { id: 7, name: "Joko Widodo", nip: "197807072002071007", rank: "Pembina Utama Muda", golongan: "IV/c" },
    { id: 8, name: "Megawati", nip: "198308082008082008", rank: "Pembina Tk. I", golongan: "IV/b" },
    { id: 9, name: "Susilo Bambang", nip: "197509091999091009", rank: "Pembina Utama Madya", golongan: "IV/d" },
    { id: 10, name: "Ani Yudhoyono", nip: "198610102011102010", rank: "Penata", golongan: "III/c" },
    { id: 20, name: "Anies Baswedan", nip: "198008202005081020", rank: "Pembina", golongan: "IV/a" }
];

const DEFAULT_USERS: User[] = [
    { id: 1, email: 'user@kemnaker.go.id', role: 'user', name: 'Staf Pengaju' },
    { id: 2, email: 'admin@kemnaker.go.id', role: 'super_admin', name: 'Super Admin' },
    { id: 3, email: 'ppk@kemnaker.go.id', role: 'ppk', name: 'Pejabat Pembuat Komitmen' },
    { id: 4, email: 'keuangan@kemnaker.go.id', role: 'keuangan', name: 'Admin Keuangan' }
];

function generateMockRecords(count: number, employees: Employee[]): TravelRecord[] {
    const records: TravelRecord[] = [];
    const now = new Date();

    for (let i = 0; i < count; i++) {
        const isPast = Math.random() > 0.3;
        const dateOffset = isPast ? -Math.floor(Math.random() * 300) : Math.floor(Math.random() * 60);
        const startDate = new Date(now);
        startDate.setDate(now.getDate() + dateOffset);
        
        const duration = Math.floor(Math.random() * 4) + 2; 
        const endDate = new Date(startDate);
        endDate.setDate(startDate.getDate() + duration - 1);

        const loc = LOCATIONS[Math.floor(Math.random() * LOCATIONS.length)];
        const empList = employees.length > 0 ? employees : DEFAULT_EMPLOYEES;
        const emp = empList[Math.floor(Math.random() * empList.length)];
        const purpose = PURPOSES[Math.floor(Math.random() * PURPOSES.length)];
        
        let status: 'Submitted' | 'Approved' | 'Rejected' = 'Submitted';
        let reportStatus: 'Pending' | 'Completed' = 'Pending';
        let reportData = null;
        let costs: any = {
             ticketGo: 0, ticketBack: 0, dailyAllowanceDays: 0, dailyAllowanceRate: 0,
             hotelDays: 0, hotelRate: 0, localTransport: 0, regionalTransport: 0, transportMode: "Pesawat"
        };
        let totalCost = 0;

        if (isPast) {
            status = Math.random() > 0.1 ? 'Approved' : 'Submitted';
            if (status === 'Approved') {
                costs = {
                    ticketGo: Math.floor(Math.random() * 2000000) + 1000000,
                    ticketBack: Math.floor(Math.random() * 2000000) + 1000000,
                    dailyAllowanceDays: duration,
                    dailyAllowanceRate: 530000,
                    hotelDays: duration - 1,
                    hotelRate: Math.floor(Math.random() * 800000) + 400000,
                    localTransport: Math.floor(Math.random() * 300000) + 100000,
                    regionalTransport: 0,
                    transportMode: "Pesawat"
                };
                totalCost = costs.ticketGo + costs.ticketBack + (costs.dailyAllowanceDays * costs.dailyAllowanceRate) + (costs.hotelDays * costs.hotelRate) + costs.localTransport;

                if (Math.random() > 0.4) {
                    reportStatus = 'Completed';
                    reportData = {
                        text: `Kegiatan ${purpose} di ${loc.city} berjalan dengan lancar.`,
                        submittedAt: endDate.toISOString(),
                        files: []
                    };
                }
            }
        }

        const spdNumber = Math.floor(Math.random() * 1000).toString().padStart(3, '0');
        const month = startDate.getMonth() + 1;
        const romanMonth = ["I", "II", "III", "IV", "V", "VI", "VII", "VIII", "IX", "X", "XI", "XII"][month - 1];
        const spd = `SPD-${spdNumber}/${romanMonth}/${startDate.getFullYear()}`;

        records.push({
            id: `PERDIN-${startDate.toISOString().slice(0,10).replace(/-/g,'')}-${Math.floor(Math.random()*1000)}`,
            spd: spd,
            employee: emp,
            email: "user@kemnaker.go.id",
            startDate: startDate.toISOString().split('T')[0],
            endDate: endDate.toISOString().split('T')[0],
            location: loc.city,
            province: loc.prov,
            purpose: purpose,
            stakeholder: "Menteri Ketenagakerjaan",
            agenda: `Melakukan ${purpose.toLowerCase()} bersama tim protokol dan pemda setempat.`,
            status: status,
            totalCost: totalCost,
            costs: costs,
            reportStatus: reportStatus,
            reportData: reportData
        });
    }
    return records.sort((a, b) => new Date(b.startDate).getTime() - new Date(a.startDate).getTime());
}

export class MockApiClient implements ApiClient {
    private employees: Employee[] = [];
    private records: TravelRecord[] = [];
    private users: User[] = [];
    private currentUser: User | null = null;

    constructor() {
        if (browser) {
            this.loadFromStorage();
        } else {
            // Server-side / Build time fallback
            this.employees = DEFAULT_EMPLOYEES;
            this.users = DEFAULT_USERS;
            this.records = generateMockRecords(50, this.employees);
        }
    }

    private loadFromStorage() {
        try {
            const storedEmployees = localStorage.getItem('perjadin_employees');
            this.employees = storedEmployees ? JSON.parse(storedEmployees) : DEFAULT_EMPLOYEES;

            const storedUsers = localStorage.getItem('app_users');
            this.users = storedUsers ? JSON.parse(storedUsers) : DEFAULT_USERS;

            const storedRecords = localStorage.getItem('perjadin_records');
            if (storedRecords) {
                this.records = JSON.parse(storedRecords);
            } else {
                this.records = generateMockRecords(20, this.employees);
                this.saveRecords();
            }

            const storedSession = localStorage.getItem('user_session');
            if (storedSession) {
                const session = JSON.parse(storedSession);
                if (session.loggedIn) {
                    this.currentUser = { ...session, id: 0 }; // Mock ID
                }
            }
        } catch (e) {
            console.error("MockClient Load Error", e);
        }
    }

    private saveRecords() {
        if (browser) localStorage.setItem('perjadin_records', JSON.stringify(this.records));
    }
    private saveEmployees() {
        if (browser) localStorage.setItem('perjadin_employees', JSON.stringify(this.employees));
    }
    private saveSession() {
        if (browser) {
            if (this.currentUser) {
                localStorage.setItem('user_session', JSON.stringify({ ...this.currentUser, loggedIn: true }));
            } else {
                localStorage.removeItem('user_session');
            }
        }
    }

    // --- Auth ---
    async login(email: string, password?: string): Promise<{ user: User; token?: string }> {
        const user = this.users.find(u => u.email === email);
        // Mock password check: in real app, we check hash. Here we assume '123456' for all mocks or skip
        if (user) {
            this.currentUser = user;
            this.saveSession();
            return { user, token: 'mock-jwt-token' };
        }
        throw new Error("Invalid credentials");
    }

    async register(email: string, password: string, name: string, role?: string): Promise<User> {
         const newUser: User = { 
             id: this.users.length + 1, 
             email, 
             name, 
             role: (role as any) || 'user' 
        };
        this.users.push(newUser);
        if(browser) localStorage.setItem('app_users', JSON.stringify(this.users));
        return newUser;
    }

    async logout(): Promise<void> {
        this.currentUser = null;
        this.saveSession();
    }

    async getCurrentUser(): Promise<User | null> {
        return this.currentUser;
    }

    // --- User Management ---
    async getUsers(): Promise<User[]> {
        return this.users;
    }

    async createUser(data: Omit<User, 'id'>): Promise<User> {
        const newId = this.users.length > 0 ? Math.max(...this.users.map(u => Number(u.id))) + 1 : 1;
        const newUser = { ...data, id: newId };
        this.users = [...this.users, newUser];
        if (browser) localStorage.setItem('app_users', JSON.stringify(this.users));
        return newUser;
    }

    async updateUser(id: string | number, data: Partial<User>): Promise<User> {
        this.users = this.users.map(u => u.id == id ? { ...u, ...data } : u);
        if (browser) localStorage.setItem('app_users', JSON.stringify(this.users));
        return this.users.find(u => u.id == id)!;
    }

    async deleteUser(id: string | number): Promise<void> {
        this.users = this.users.filter(u => u.id != id);
        if (browser) localStorage.setItem('app_users', JSON.stringify(this.users));
    }

    // --- Employees ---
    async getEmployees(): Promise<Employee[]> {
        return this.employees;
    }
    
    async createEmployee(data: Omit<Employee, 'id'>): Promise<Employee> {
        const newId = this.employees.length > 0 ? Math.max(...this.employees.map(e => Number(e.id))) + 1 : 1;
        const emp = { ...data, id: newId };
        this.employees = [...this.employees, emp];
        this.saveEmployees();
        return emp;
    }

    async updateEmployee(id: string | number, data: Partial<Employee>): Promise<Employee> {
        this.employees = this.employees.map(e => e.id == id ? { ...e, ...data } : e);
        this.saveEmployees();
        return this.employees.find(e => e.id == id)!;
    }

    async deleteEmployee(id: string | number): Promise<void> {
        this.employees = this.employees.filter(e => e.id != id);
        this.saveEmployees();
    }

    // --- Records ---
    async getRecords(filters?: Record<string, any>): Promise<TravelRecord[]> {
        let result = [...this.records];
        if (filters?.status) {
            result = result.filter(r => r.status === filters.status);
        }
        return result.sort((a, b) => new Date(b.startDate).getTime() - new Date(a.startDate).getTime());
    }

    async getRecordById(id: string): Promise<TravelRecord | null> {
        return this.records.find(r => r.id === id) || null;
    }

    async createRecord(data: any): Promise<TravelRecord[]> {
        // Handle bulk creation (one trip, multiple employees)
        const newRecords: TravelRecord[] = (data.employees || [data.employee]).map((emp: Employee, index: number) => {
             const random = Math.floor(Math.random() * 1000).toString().padStart(3, '0');
             const spd = `SPD-${random}-${index}/XII/2026`;
             
             return {
                 ...data,
                 id: `${Date.now()}-${index}`,
                 spd,
                 employee: emp,
                 status: 'Submitted',
                 reportStatus: 'Pending',
                 totalCost: 0,
                 costs: {
                    ticketGo: 0, ticketBack: 0, dailyAllowanceDays: 0, dailyAllowanceRate: 0,
                    hotelDays: 0, hotelRate: 0, localTransport: 0, regionalTransport: 0, transportMode: 'Pesawat'
                },
             };
        });

        this.records = [...newRecords, ...this.records];
        this.saveRecords();
        return newRecords;
    }

    async updateRecord(id: string, data: Partial<TravelRecord>): Promise<TravelRecord> {
        this.records = this.records.map(r => r.id === id ? { ...r, ...data } : r);
        this.saveRecords();
        return this.records.find(r => r.id === id)!;
    }
}
