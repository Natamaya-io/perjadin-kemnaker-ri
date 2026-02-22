import { writable } from 'svelte/store';
import { api } from '$lib/api';
import type { TravelRecord, Employee } from '$lib/api/types';

// --- Travel Records ---
export const recordsStore = writable<TravelRecord[]>([]);

export async function loadRecords() {
    try {
        const data = await api.getRecords();
        recordsStore.set(data);
    } catch (e) {
        console.error("Failed to load records", e);
    }
}

// Initial Load for Records
if (typeof window !== 'undefined') {
    loadRecords();
}

export async function addRecord(tripData: any) {
    try {
        const newRecords = await api.createRecord(tripData);
        recordsStore.update(current => [...newRecords, ...current]);
        return newRecords;
    } catch (e) {
        console.error("Failed to add record", e);
        throw e;
    }
}

export async function updateRecord(id: string, data: Partial<TravelRecord>) {
    try {
        const updated = await api.updateRecord(id, data);
        recordsStore.update(current => 
            current.map(r => r.id === id ? updated : r)
        );
        return updated;
    } catch (e) {
        console.error("Failed to update record", e);
        throw e;
    }
}

// --- Employees ---
export const employeesStore = writable<Employee[]>([]);

export async function loadEmployees() {
    try {
        const data = await api.getEmployees();
        employeesStore.set(data);
    } catch (e) {
        console.error("Failed to load employees", e);
    }
}

// Initial Load for Employees
if (typeof window !== 'undefined') {
    loadEmployees();
}

export async function addEmployee(employeeData: Omit<Employee, 'id'>) {
    try {
        const newEmployee = await api.createEmployee(employeeData);
        employeesStore.update(current => [...current, newEmployee]);
        return newEmployee;
    } catch (e) {
        console.error("Failed to add employee", e);
        throw e;
    }
}

export async function updateEmployee(id: string | number, data: Partial<Employee>) {
    try {
        const updated = await api.updateEmployee(id, data);
        employeesStore.update(current => 
            current.map(e => e.id == id ? updated : e)
        );
        return updated;
    } catch (e) {
        console.error("Failed to update employee", e);
        throw e;
    }
}

export async function removeEmployee(id: string | number) {
    try {
        await api.deleteEmployee(id);
        employeesStore.update(current => current.filter(e => e.id != id));
    } catch (e) {
        console.error("Failed to remove employee", e);
        throw e;
    }
}
