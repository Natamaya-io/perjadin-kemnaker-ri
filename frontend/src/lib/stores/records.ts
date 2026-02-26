import { writable } from 'svelte/store';
import { api } from '$lib/api';
import type { TravelRecord } from '$lib/api/types';

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

// Export a reset function
export function clearStores() {
    recordsStore.set([]);
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
