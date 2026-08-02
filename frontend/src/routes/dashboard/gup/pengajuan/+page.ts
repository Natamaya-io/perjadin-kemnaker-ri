import { api } from '$lib/api';

export async function load({ fetch }) {
    try {
        const response = await api.get('/gup/pengajuan', { fetch });
        return {
            transactions: response || []
        };
    } catch (e) {
        console.error('Failed to load GUP transactions', e);
        return {
            transactions: []
        };
    }
}
