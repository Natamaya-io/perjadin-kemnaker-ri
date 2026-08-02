import { api } from '$lib/shared/api';

export async function load({ fetch }) {
    try {
        const response = await api.getGupPengajuan(fetch);
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
