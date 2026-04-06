// Shared Types for API Interaction

export interface User {
    id: string | number;
    email: string;
    name: string;
    role: 'super_admin' | 'kasubag' | 'protokol';
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
    transportAmount?: number;
    otherCost?: number;
    otherCostDesc?: string;
    additionalCosts?: { name: string; amount: number; file?: any }[];
    boardingPassFiles?: any[];
    details?: TravelCost[]; // Per-location cost breakdown
}

export interface TravelReport {
    text: string;
    submittedAt: string;
    files: any[];
}

export interface TravelLocation {
    id: string;
    travelRecordId: string;
    location: string;
    province: string;
    startDate: string;
    endDate: string;
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
    locations?: TravelLocation[];
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

export interface ImportDetail {
    row: number;
    spjId: string;
    name: string;
    status: 'imported' | 'skipped_duplicate' | 'skipped_thr' | 'skipped_no_user' | 'failed';
    message?: string;
}

export interface ImportResult {
    totalRows: number;
    imported: number;
    skipped: number;
    failed: number;
    details: ImportDetail[];
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
    exportSpdPdf(id: string): Promise<Blob>;
    exportSpdDocx(id: string): Promise<Blob>;
    exportLaporanPdf(id: string): Promise<Blob>;
    exportLaporanDocx(id: string): Promise<Blob>;
    exportRincianPdf(id: string): Promise<Blob>;
    exportRincianDocx(id: string): Promise<Blob>;
    createRecord(record: Omit<TravelRecord, 'id' | 'spd' | 'status' | 'reportStatus'>): Promise<TravelRecord[]>; // Returns array because one request can create multiple records (bulk)
    updateRecord(id: string, record: Partial<TravelRecord>): Promise<TravelRecord>;
    deleteRecord(id: string): Promise<void>;
    deleteRecordsBySpd(spd: string): Promise<void>;

    // Import
    importExcel(file: File): Promise<ImportResult>;

    // Master Data
    getProvinces(): Promise<any[]>;
    getSBMRates(): Promise<any[]>;
    getSettings(): Promise<any>;
    updateSettings(settings: any): Promise<void>;
}
