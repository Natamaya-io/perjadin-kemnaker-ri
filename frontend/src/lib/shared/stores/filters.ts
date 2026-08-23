import { writable } from 'svelte/store';
import { browser } from '$app/environment';

function createPersistentStore(key: string, initialValue: any) {
    let storedValue;
    if (browser) {
        const item = sessionStorage.getItem(key);
        if (item) {
            try {
                storedValue = JSON.parse(item);
            } catch (e) {
                storedValue = initialValue;
            }
        }
    }
    
    const store = writable(storedValue ?? initialValue);
    
    if (browser) {
        store.subscribe(value => {
            sessionStorage.setItem(key, JSON.stringify(value));
        });
    }
    
    return store;
}

// Pengajuan Filters
export const pengajuanFilters = createPersistentStore('pengajuanFilters', {
    searchQuery: '',
    statusFilter: 'all',
    sortOption: 'spj-desc',
    startDate: '',
    endDate: '',
    typeFilter: 'luar_kota'
});

// Admin Perdin Filters
export const adminPerdinFilters = createPersistentStore('adminPerdinFilters', {
    searchQuery: '',
    statusFilter: 'all',
    sortOption: 'spj-desc',
    startDate: '',
    endDate: '',
    typeFilter: 'all'
});

// Laporan Filters
export const laporanFilters = createPersistentStore('laporanFilters', {
    searchQuery: '',
    statusFilter: 'all',
    sortOption: 'spj-desc',
    startDate: '',
    endDate: '',
    typeFilter: 'all'
});
