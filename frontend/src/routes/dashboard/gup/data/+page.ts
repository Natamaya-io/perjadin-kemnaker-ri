import { api } from '$lib/api';

export async function load({ fetch }) {
    try {
        const response = await api.get('/gup/data?year=2024', { fetch });
        return {
            budgets: response || []
        };
    } catch (e) {
        console.error('Failed to load GUP budgets', e);
        return {
            budgets: []
        };
    }
}
