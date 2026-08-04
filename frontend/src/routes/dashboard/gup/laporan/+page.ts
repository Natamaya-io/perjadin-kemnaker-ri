import { api } from '$lib/shared/api';

export async function load({ fetch }) {
    const year = new Date().getFullYear();
    try {
        const response = await api.getGupLaporan(year, fetch);
        return {
            laporan: response || { summary: {}, rows: [] },
            year
        };
    } catch (e) {
        console.error('Failed to load GUP laporan', e);
        return {
            laporan: { summary: {}, rows: [] },
            year
        };
    }
}
