import { api } from '$lib/api';

export async function load({ fetch }) {
    try {
        const response = await api.get('/gup/laporan', { fetch });
        return {
            reports: response || []
        };
    } catch (e) {
        console.error('Failed to load GUP laporan', e);
        return {
            reports: []
        };
    }
}
