import { api } from '$lib/shared/api';

export async function load({ fetch }) {
    try {
        const response = await api.getGupLaporan(fetch);
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
