import { api } from '$lib/shared/api';

export async function load({ fetch, url }) {
    const queryYear = url.searchParams.get('year');
    const year = queryYear ? parseInt(queryYear, 10) : new Date().getFullYear();
    
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
