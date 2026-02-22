import type { ApiClient, Employee, TravelRecord, User } from './types';
import { browser } from '$app/environment';

const BASE_URL = import.meta.env.VITE_API_URL || 'http://localhost:8080/api/v1';

export class RealApiClient implements ApiClient {
    private token: string | null = null;

    constructor() {
        if (browser) {
            this.token = localStorage.getItem('auth_token');
        }
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

        if (!response.ok) {
            if (response.status === 401) {
                this.logout();
                throw new Error("Unauthorized");
            }
            const error = await response.json().catch(() => ({ message: 'Unknown error' }));
            throw new Error(error.message || `API Error: ${response.status}`);
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

    // --- Employees ---
    async getEmployees(): Promise<Employee[]> {
        // Assuming backend has this endpoint
        return this.request<Employee[]>('/employees');
    }

    async createEmployee(employee: Omit<Employee, 'id'>): Promise<Employee> {
        return this.request<Employee>('/employees', {
            method: 'POST',
            body: JSON.stringify(employee)
        });
    }

    async updateEmployee(id: string | number, employee: Partial<Employee>): Promise<Employee> {
        return this.request<Employee>(`/employees/${id}`, {
            method: 'PUT',
            body: JSON.stringify(employee)
        });
    }

    async deleteEmployee(id: string | number): Promise<void> {
        return this.request<void>(`/employees/${id}`, {
            method: 'DELETE'
        });
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
         // Adjust backend to accept bulk or loop here
         // For now, assume backend accepts single and we loop if needed or backend updated
         const res = await this.request<TravelRecord>('/records', {
             method: 'POST',
             body: JSON.stringify(record)
         });
         return [res];
    }

    async updateRecord(id: string, record: Partial<TravelRecord>): Promise<TravelRecord> {
        return this.request<TravelRecord>(`/records/${id}`, {
            method: 'PUT',
            body: JSON.stringify(record)
        });
    }
}
