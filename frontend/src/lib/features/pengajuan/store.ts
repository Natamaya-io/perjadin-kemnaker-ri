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
        recordsStore.set(data || []);
    } catch (e: any) {
        if (e.message === 'Unauthorized') return;
        console.warn("Failed to load records (backend might be starting):", e.message);
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

export async function updateMultipleRecords(updates: Array<{id: string, data: Partial<TravelRecord>}>) {
    try {
        // Run updates sequentially to avoid potential race conditions in simple backends
        const results = [];
        for (const update of updates) {
            const res = await api.updateRecord(update.id, update.data);
            results.push(res);
        }
        
        // Refresh full store to ensure consistency
        await loadRecords();
        return results;
    } catch (e) {
        console.error("Failed to update multiple records", e);
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
    // Optimistic update: remove from UI immediately
    recordsStore.update(current => current.filter(r => r.spd !== spd));

    try {
        await api.deleteRecordsBySpd(spd);
    } catch (e: any) {
        // If it's a 502, it might have actually succeeded in the backend
        // We check if the records are actually gone
        if (e.message?.includes('502')) {
             console.warn("Detected 502 during delete, verifying data status...");
             await loadRecords();
             
             let exists = false;
             recordsStore.subscribe(recs => {
                 exists = recs.some(r => r.spd === spd);
             })();

             if (!exists) {
                 // Success! Data is gone despite the 502 error
                 return;
             }
        }

        console.error(`Failed to delete records for SPD ${spd}`, e);
        // Rollback: reload from server if it actually failed
        await loadRecords();
        throw e;
    }
}
