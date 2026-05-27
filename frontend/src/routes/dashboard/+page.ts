import { api } from '$lib/shared/api';

export async function load() {
    try {
        // We only prefetch if running in the browser (since SSR is disabled)
        // and if there's a token
        if (typeof window !== 'undefined' && localStorage.getItem('auth_token')) {
            const summary = await api.getDashboardSummary();
            return {
                summary
            };
        }
    } catch (e) {
        console.warn("Failed to prefetch dashboard summary:", e);
    }
    return { summary: null };
}
