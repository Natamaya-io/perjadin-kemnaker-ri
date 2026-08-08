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
    dailyAllowanceDays?: number;
    dailyAllowanceRate?: number;
    hotelDays?: number;
    hotelRate?: number;
    localTransport?: number;
    regionalTransport?: number;
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
    submittedAt?: string;
    files: any[];
    sppdFile?: any;
    suratTugasFile?: any;
    tanggalMerah?: string | string[];
    lastDraftSavedAt?: string;
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
    employeeId?: string;
    creator?: User;
    creatorId?: string;
    email: string; // Creator email
    startDate: string;
    endDate: string;
    createdAt: string;
    suratTugasPath?: string;
    suratTugasNumber?: string;
    suratTugasDate?: string;
    isViewed?: boolean;
    location: string;
    province: string;
    locations?: TravelLocation[];
    type?: string;
    purpose: string;
    stakeholder: string;
    agenda: string;
    status: 'Draft' | 'Submitted' | 'Approved' | 'Rejected';
    reportStatus: 'Pending' | 'Completed' | 'Draft';
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

export interface DashboardBudget {
    year: number;
    month: number;
    total: number;
}

export interface DashboardSummary {
    totalTrips: number;
    activeTrips: number;
    statusCompleted: number;
    statusInProgress: number;
    statusAssigned: number;
    statusRejected: number;
    reportCompleted: number;
    reportPending: number;
    recentRecords: TravelRecord[];
    budgets: DashboardBudget[];
}

export interface PaginatedParams {
    cursor?: string;
    limit?: number;
    search?: string;
    status?: string;
    report_status?: string;
    payment_status?: string;
    sort_by?: string;
    start_date?: string; // ISO string
    end_date?: string;   // ISO string
    type?: string;
    user_id?: string;
}

export interface PaginatedResponse {
    data: TravelRecord[];
    totalItems: number;
    totalRecords: number;
    nextCursor: string;
    limit: number;
}

export interface ApiClient {
    // Auth
    login(email: string, password?: string): Promise<{ user: User; token?: string; require_password_change?: boolean }>;
    register(email: string, password: string, name: string, role?: string): Promise<User>;
    logout(): Promise<void>;
    getCurrentUser(): Promise<User | null>;
    changePassword(newPassword: string): Promise<void>;
    getCurrentUser(): Promise<User | null>;

    // User Management (Admin)
    getUsers(filters?: { role?: string; search?: string }): Promise<User[]>;
    createUser(user: Omit<User, 'id'>): Promise<User>;
    updateUser(id: string | number, user: Partial<User>): Promise<User>;
    deleteUser(id: string | number): Promise<void>;

    // Files
    uploadFile(file: File): Promise<{path: string}>;

    // Records
    getDashboardSummary(): Promise<DashboardSummary>;
    getRecords(filters?: Record<string, any>): Promise<TravelRecord[]>;
    getPaginatedRecords(params: PaginatedParams): Promise<PaginatedResponse>;
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

    // GUP
    getGupMasterData(year?: number, customFetch?: typeof fetch): Promise<any>;
    createAccountCode(data: any): Promise<any>;
    updateAccountCode(id: string, data: any): Promise<any>;
    deleteAccountCode(id: string): Promise<any>;
    createProcurementType(data: any): Promise<any>;
    updateProcurementType(id: string, data: any): Promise<any>;
    deleteProcurementType(id: string): Promise<any>;
    getGupPengajuan(customFetch?: typeof fetch): Promise<any[]>;
    getGupPengajuanById(id: string, customFetch?: typeof fetch): Promise<any>;
    createGupPengajuan(data: any): Promise<any>;
    getGupLaporan(year?: number, customFetch?: typeof fetch): Promise<any>;
    saveGupBudget(data: { procurementTypeId: string; year: number; amount: number }): Promise<any>;
    getGupLs(year: number, customFetch?: typeof fetch): Promise<any[]>;
    saveGupLs(data: any[]): Promise<any>;
    getGupData(year: number, customFetch?: typeof fetch): Promise<any[]>;

    // Chatbot
    askChatbot(message: string): Promise<any>;
    getChatbotSnapshot(): Promise<any>;
}
