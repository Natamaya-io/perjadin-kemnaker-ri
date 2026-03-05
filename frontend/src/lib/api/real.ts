import type { ApiClient, TravelRecord, User } from './types';
import { browser } from '$app/environment';

const BASE_URL = import.meta.env.VITE_API_URL || '/api/v1';

export class RealApiClient implements ApiClient {
    private token: string | null = null;
    private unauthorizedHandler: (() => void) | null = null;

    constructor() {
        if (browser) {
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
    async login(email: string, password?: string): Promise<{ user: User; token?: string }> {
        const res = await this.request<{ user: User; token: string }>('/auth/login', {
            method: 'POST',
            body: JSON.stringify({ email, password })
        });
        
        this.token = res.token;
        if (browser) localStorage.setItem('auth_token', res.token);
        
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
        if (browser) localStorage.removeItem('auth_token');
        // Optional: Call backend logout endpoint if it exists
    }

    async getCurrentUser(): Promise<User | null> {
        // If we have a token, we might want to validate it or fetch user profile
        // For now, let's assume the frontend state manages the user object after login
        // Or implement a /me endpoint in backend
        return null; 
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

    async createRecord(record: any): Promise<TravelRecord[]> {
         const employees = record.employees || [record.employee];
         const createdRecords: TravelRecord[] = [];

         for (const emp of employees) {
             const safeDate = (d: any) => {
                 if (!d) return undefined;
                 const date = new Date(d);
                 return isNaN(date.getTime()) ? undefined : date.toISOString();
             };

             const singleRecord = {
                 ...record,
                 employeeId: emp.id,
                 employee: undefined,
                 employees: undefined,
                 startDate: safeDate(record.startDate),
                 endDate: safeDate(record.endDate)
             };

             // backend expects 'employeeId' (UUID).
             // Ideally we need the UUID. Assuming frontend sends proper Employee objects with IDs.

             const res = await this.request<TravelRecord>('/records', {
                 method: 'POST',
                 body: JSON.stringify(singleRecord)
             });

             // Manually attach employee to response because Create response might not preload it
             createdRecords.push({ ...res, employee: emp });
         }
         return createdRecords;
    }
    updateRecord(id: string, record: Partial<TravelRecord>): Promise<TravelRecord> {
        return this.request<TravelRecord>(`/records/${id}`, {
            method: 'PUT',
            body: JSON.stringify(record)
        });
    }

    deleteRecord(id: string): Promise<void> {
        return this.request<void>(`/records/${id}`, {
            method: 'DELETE'
        });
    }
}
