import { api } from '$lib/shared/api';

export async function load() {
    try {
        if (typeof window !== 'undefined' && localStorage.getItem('auth_token')) {
            // Default parameters for the initial load of the admin table
            const response = await api.getPaginatedRecords({ limit: 50, sort_by: 'spj-desc' });
            return {
                recordsResponse: response
            };
        }
    } catch (e) {
        console.warn("Failed to prefetch admin records:", e);
    }
    return { recordsResponse: null };
}
