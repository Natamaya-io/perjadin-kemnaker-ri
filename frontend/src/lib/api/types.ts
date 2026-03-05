// Shared Types for API Interaction

export interface User {
    id: string | number;
    email: string;
    name: string;
    role: 'super_admin' | 'keuangan' | 'kasubag' | 'protokol';
    token?: string; // JWT for real API
    nip?: string;
    pangkat?: string;
    golongan?: string;
    jabatan?: string;
    tingkatBiaya?: string;
    nomorHp?: string;
}

export interface TravelCost {
    ticketGo: number;
    ticketBack: number;
    dailyAllowanceDays: number;
    dailyAllowanceRate: number;
    hotelDays: number;
    hotelRate: number;
    localTransport: number;
    regionalTransport: number;
    transportMode: string;
}

export interface TravelReport {
    text: string;
    submittedAt: string;
    files: any[];
}

export interface TravelRecord {
    id: string;
    spd: string;
    employee: User;
    creator?: User;
    creatorId?: string;
    email: string; // Creator email
    startDate: string;
    endDate: string;
    suratTugasPath?: string;
    suratTugasNumber?: string;
    location: string;
    province: string;
    type?: string;
    purpose: string;
    stakeholder: string;
    agenda: string;
    status: 'Draft' | 'Submitted' | 'Approved' | 'Rejected';
    reportStatus: 'Pending' | 'Completed';
    totalCost: number;
    costs?: TravelCost;
    reportData?: TravelReport | null;
}

export interface ApiClient {
    // Auth
    login(email: string, password?: string): Promise<{ user: User; token?: string }>;
    register(email: string, password: string, name: string, role?: string): Promise<User>;
    logout(): Promise<void>;
    getCurrentUser(): Promise<User | null>;
    
    // User Management (Admin)
    getUsers(): Promise<User[]>;
    createUser(user: Omit<User, 'id'>): Promise<User>;
    updateUser(id: string | number, user: Partial<User>): Promise<User>;
    deleteUser(id: string | number): Promise<void>;

    // Files
    uploadFile(file: File): Promise<{path: string}>;

    // Records
    getRecords(filters?: Record<string, any>): Promise<TravelRecord[]>;
    getRecordById(id: string): Promise<TravelRecord | null>;
    createRecord(record: Omit<TravelRecord, 'id' | 'spd' | 'status' | 'reportStatus'>): Promise<TravelRecord[]>; // Returns array because one request can create multiple records (bulk)
    updateRecord(id: string, record: Partial<TravelRecord>): Promise<TravelRecord>;
    deleteRecord(id: string): Promise<void>;
}
