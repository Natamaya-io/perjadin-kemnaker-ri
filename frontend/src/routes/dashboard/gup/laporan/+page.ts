import { api } from '$lib/shared/api';

export async function load({ fetch, url }) {
    const queryYear = url.searchParams.get('year');
    const queryMonth = url.searchParams.get('month');
    const year = queryYear ? parseInt(queryYear, 10) : new Date().getFullYear();
    const month = queryMonth ? parseInt(queryMonth, 10) : 0;
    
    try {
        const response = await api.getGupLaporan(year, month, fetch);
        return {
            laporan: response || { summary: {}, rows: [] },
            year,
            month
        };
    } catch (e) {
        console.error('Failed to load GUP laporan', e);
        return {
            laporan: { summary: {}, rows: [] },
            year,
            month
        };
    }
}
