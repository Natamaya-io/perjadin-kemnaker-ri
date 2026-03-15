import { writable } from 'svelte/store';
import { api } from '$lib/shared/api';
import type { TravelRecord } from '$lib/shared/api/types';

import { browser } from '$app/environment';

// --- Travel Records ---
export const recordsStore = writable<TravelRecord[]>([]);

export async function loadRecords() {
    if (typeof window === 'undefined' || !localStorage.getItem('auth_token')) return;
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
            current.map(r => r.id === id ? { ...r, ...updated, employee: updated.employee?.id ? updated.employee : r.employee } : r)
        );
        return updated;
    } catch (e) {
        console.error("Failed to update record", e);
        throw e;
    }
}

export async function deleteRecord(id: string) {
    try {
        await api.deleteRecord(id);
        recordsStore.update(current => current.filter(r => r.id !== id));
    } catch (e) {
        console.error(`Failed to delete record ${id}`, e);
        throw e;
    }
}

export async function deleteRecordBySpd(spd: string) {
    try {
        // Find all records with this SPD
        let recordsToDelete: TravelRecord[] = [];
        recordsStore.subscribe(records => {
            recordsToDelete = records.filter(r => r.spd === spd);
        })();

        // Delete each via API
        for (const record of recordsToDelete) {
            await api.deleteRecord(record.id);
        }

        // Update local store
        recordsStore.update(current => current.filter(r => r.spd !== spd));
    } catch (e) {
        console.error(`Failed to delete records for SPD ${spd}`, e);
        throw e;
    }
}
