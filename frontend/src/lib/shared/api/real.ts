import type { ApiClient, TravelRecord, User } from './types';
import { browser } from '$app/environment';

// Hardcode BASE_URL to always use the proxy to prevent any CORS or Direct-Backend 0 byte body issues
const BASE_URL = '/api/v1';

export class RealApiClient implements ApiClient {
    private token: string | null = null;
    private unauthorizedHandler: (() => void) | null = null;

    constructor() {
        if (typeof window !== 'undefined') {
            this.token = localStorage.getItem('auth_token');
        }
    }

    setUnauthorizedHandler(handler: () => void) {
        this.unauthorizedHandler = handler;
    }

    private async request<T>(endpoint: string, options: RequestInit = {}): Promise<T> {
        const headers: HeadersInit = {
            'Content-Type': 'application/json',
            ...(this.token ? { 'Authorization': `Bearer ${this.token}` } : {})
        };

        const response = await fetch(`${BASE_URL}${endpoint}`, {
            ...options,
            headers: {
                ...headers,
                ...options.headers
            }
        });

        if (response.status === 204) {
            return null as unknown as T;
        }

        if (!response.ok) {
            if (response.status === 401) {
                if (this.unauthorizedHandler) {
                    this.unauthorizedHandler();
                }
                throw new Error("Unauthorized");
            }
            const errorText = await response.text();
            let errorMessage = `API Error: ${response.status}`;
            try {
                const errorJson = JSON.parse(errorText);
                if (errorJson.message) errorMessage = errorJson.message;
            } catch (e) {
                // ignore
            }
            throw new Error(errorMessage);
        }

        return response.json();
    }

    // --- Auth ---
    async login(email: string, password?: string): Promise<{ user: User; token?: string; require_password_change?: boolean }> {
        const res = await this.request<{ user: User; token: string; require_password_change: boolean }>('/auth/login', {
            method: 'POST',
            body: JSON.stringify({ email, password })
        });
        
        this.token = res.token;
        if (typeof window !== 'undefined') localStorage.setItem('auth_token', res.token);
        
        return res;
    }

    async register(email: string, password: string, name: string, role?: string): Promise<User> {
        return this.request<User>('/auth/register', {
            method: 'POST',
            body: JSON.stringify({ email, password, name, role })
        });
    }

    async logout(): Promise<void> {
        this.token = null;
        if (typeof window !== 'undefined') localStorage.removeItem('auth_token');
        // Optional: Call backend logout endpoint if it exists
    }

    async getCurrentUser(): Promise<User | null> {
        if (!this.token) return null;
        try {
            return await this.request<User>('/auth/me');
        } catch (e) {
            return null;
        }
    }

    async changePassword(newPassword: string): Promise<void> {
        return this.request<void>('/auth/change-password', {
            method: 'PUT',
            body: JSON.stringify({ new_password: newPassword })
        });
    }

    // --- User Management ---
    async getUsers(): Promise<User[]> {
        return this.request<User[]>('/users');
    }

    async createUser(user: Omit<User, 'id'>): Promise<User> {
        return this.request<User>('/users', {
            method: 'POST',
            body: JSON.stringify(user)
        });
    }

    async updateUser(id: string | number, user: Partial<User>): Promise<User> {
        return this.request<User>(`/users/${id}`, {
            method: 'PUT',
            body: JSON.stringify(user)
        });
    }

    async deleteUser(id: string | number): Promise<void> {
        return this.request<void>(`/users/${id}`, {
            method: 'DELETE'
        });
    }

    // --- Files ---
    async uploadFile(file: File): Promise<{path: string}> {
        const formData = new FormData();
        formData.append('file', file);

        const headers: HeadersInit = {
            ...(this.token ? { 'Authorization': `Bearer ${this.token}` } : {})
        };

        const response = await fetch(`${BASE_URL}/upload`, {
            method: 'POST',
            headers,
            body: formData
        });

        if (!response.ok) {
            if (response.status === 401 && this.unauthorizedHandler) {
                this.unauthorizedHandler();
            }
            throw new Error(`Upload failed: ${response.statusText}`);
        }

        return response.json();
    }

    // --- Records ---
    async getRecords(filters?: Record<string, any>): Promise<TravelRecord[]> {
        const query = filters ? '?' + new URLSearchParams(filters).toString() : '';
        return this.request<TravelRecord[]>(`/records${query}`);
    }

    async getRecordById(id: string): Promise<TravelRecord | null> {
        return this.request<TravelRecord>(`/records/${id}`);
    }

    private async fetchDocument(id: string, type: string, format: 'pdf' | 'docx' = 'pdf'): Promise<Blob> {
        const headers: HeadersInit = {
            ...(this.token ? { 'Authorization': `Bearer ${this.token}` } : {})
        };

        const formatPath = format === 'pdf' ? 'stream' : 'docx';
        const response = await fetch(`${BASE_URL}/records/${id}/${type}-${formatPath}`, {
            method: 'GET',
            headers
        });


        if (!response.ok) {
            if (response.status === 401 && this.unauthorizedHandler) {
                this.unauthorizedHandler();
            }
            const errorText = await response.text();
            let errorMessage = response.statusText;
            try {
                const errorJson = JSON.parse(errorText);
                if (errorJson.message) errorMessage = errorJson.message;
            } catch (e) {
                if (errorText) errorMessage = errorText;
            }
            throw new Error(`Export ${type.toUpperCase()} ${format.toUpperCase()} failed: ${errorMessage}`);
        }

        const buffer = await response.arrayBuffer();
        if (buffer.byteLength === 0) {
            throw new Error(`Data dokumen ${format.toUpperCase()} kosong (0 byte) diterima dari server untuk tipe ${type}.`);
        }
        
        const mimeType = format === 'pdf' 
            ? 'application/pdf' 
            : 'application/vnd.openxmlformats-officedocument.wordprocessingml.document';
            
        return new Blob([buffer], { type: mimeType });
    }

    async exportSpdPdf(id: string): Promise<Blob> {
        return this.fetchDocument(id, 'spd', 'pdf');
    }

    async exportSpdDocx(id: string): Promise<Blob> {
        return this.fetchDocument(id, 'spd', 'docx');
    }

    async exportLaporanPdf(id: string): Promise<Blob> {
        return this.fetchDocument(id, 'laporan', 'pdf');
    }

    async exportLaporanDocx(id: string): Promise<Blob> {
        return this.fetchDocument(id, 'laporan', 'docx');
    }

    async exportRincianPdf(id: string): Promise<Blob> {
        return this.fetchDocument(id, 'rincian', 'pdf');
    }

    async exportRincianDocx(id: string): Promise<Blob> {
        return this.fetchDocument(id, 'rincian', 'docx');
    }

    async createRecord(record: any): Promise<TravelRecord[]> {
         const employees = record.employees || [record.employee];
         const createdRecords: TravelRecord[] = [];

         const safeDate = (d: any) => {
             if (!d) return undefined;
             const date = new Date(d);
             return isNaN(date.getTime()) ? undefined : date.toISOString();
         };

         // Map locations to have ISO dates
         const locations = (record.locations || []).map((loc: any) => ({
             ...loc,
             startDate: safeDate(loc.startDate),
             endDate: safeDate(loc.endDate)
         }));

         for (const emp of employees) {
             const singleRecord = {
                 ...record,
                 employeeId: emp.id,
                 employee: undefined,
                 employees: undefined,
                 startDate: safeDate(record.startDate),
                 endDate: safeDate(record.endDate),
                 locations: locations
             };

             const res = await this.request<TravelRecord>('/records', {
                 method: 'POST',
                 body: JSON.stringify(singleRecord)
             });

             createdRecords.push({ ...res, employee: emp });
         }
         return createdRecords;
    }
    async updateRecord(id: string, record: Partial<TravelRecord>): Promise<TravelRecord> {
        const safeDate = (d: any) => {
            if (!d) return undefined;
            const date = new Date(d);
            return isNaN(date.getTime()) ? undefined : date.toISOString();
        };

        const payload: any = {
            ...record,
        };
        if (record.startDate !== undefined) payload.startDate = safeDate(record.startDate);
        if (record.endDate !== undefined) payload.endDate = safeDate(record.endDate);
        if (record.locations !== undefined) {
            payload.locations = record.locations.map((loc: any) => ({
                ...loc,
                startDate: safeDate(loc.startDate),
                endDate: safeDate(loc.endDate)
            }));
        }

        return this.request<TravelRecord>(`/records/${id}`, {
            method: 'PUT',
            body: JSON.stringify(payload)
        });
    }

    deleteRecord(id: string): Promise<void> {
        return this.request<void>(`/records/${id}`, {
            method: 'DELETE'
        });
    }

    deleteRecordsBySpd(spd: string): Promise<void> {
        return this.request<void>(`/records/spd/${spd}`, {
            method: 'DELETE'
        });
    }

    async importExcel(file: File): Promise<import('./types').ImportResult> {
        const formData = new FormData();
        formData.append('file', file);

        const headers: HeadersInit = {
            ...(this.token ? { 'Authorization': `Bearer ${this.token}` } : {})
        };

        const response = await fetch(`${BASE_URL}/records/import`, {
            method: 'POST',
            headers,
            body: formData
        });

        if (!response.ok) {
            if (response.status === 401 && this.unauthorizedHandler) {
                this.unauthorizedHandler();
            }
            const errorText = await response.text();
            let errorMessage = `Import gagal: ${response.status}`;
            try {
                const errorJson = JSON.parse(errorText);
                if (errorJson.message) errorMessage = errorJson.message;
            } catch (e) { /* ignore */ }
            throw new Error(errorMessage);
        }

        return response.json();
    }

    // Master Data
    getProvinces(): Promise<any[]> {
        return this.request<any[]>('/master/provinces');
    }

    getSBMRates(): Promise<any[]> {
        return this.request<any[]>('/master/sbm-rates');
    }

    getSettings(): Promise<any> {
        return this.request<any>('/master/settings');
    }

    updateSettings(settings: any): Promise<void> {
        return this.request<void>('/master/settings', {
            method: 'PUT',
            body: JSON.stringify(settings)
        });
    }
}
