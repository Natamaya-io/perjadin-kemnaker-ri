import { api } from '$lib/shared/api';

export async function load({ fetch }) {
    try {
        const response = await api.get('/gup/ls?year=2024', { fetch });
        return {
            lsData: response || []
        };
    } catch (e) {
        console.error('Failed to load GUP LS', e);
        return {
            lsData: []
        };
    }
}
