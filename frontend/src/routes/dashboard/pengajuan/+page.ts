import { api } from '$lib/shared/api';

export async function load() {
    try {
        if (typeof window !== 'undefined' && localStorage.getItem('auth_token')) {
            // Because we don't know the exact user_id easily outside of the Svelte store,
            // we let the backend handle the user scoping if we don't pass user_id.
            // Wait, the API automatically scopes to the logged-in user if they are not super_admin/kasubag.
            // So we just call it normally!
            const response = await api.getPaginatedRecords({ limit: 50, sort_by: 'spj-desc' });
            return {
                recordsResponse: response
            };
        }
    } catch (e) {
        console.warn("Failed to prefetch pengajuan records:", e);
    }
    return { recordsResponse: null };
}
