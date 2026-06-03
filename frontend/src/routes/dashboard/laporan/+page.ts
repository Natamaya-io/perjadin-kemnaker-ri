import { api } from '$lib/shared/api';

export async function load() {
    try {
        if (typeof window !== 'undefined' && localStorage.getItem('auth_token')) {
            // Because we don't know the exact user_id easily outside of the Svelte store,
            // we let the backend handle the user scoping if we don't pass user_id.
            const response = await api.getPaginatedRecords({ limit: 50, sort_by: 'spj-desc' });
            return {
                recordsResponse: response
            };
        }
    } catch (e) {
        console.warn("Failed to prefetch laporan records:", e);
    }
    return { recordsResponse: null };
}
