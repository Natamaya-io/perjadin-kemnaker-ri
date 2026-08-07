import { api } from '$lib/shared/api';

export async function load({ fetch, url }) {
    const yearStr = url.searchParams.get('year') || '2026';
    const year = parseInt(yearStr, 10);

    try {
        const [lsData, transactions, masterData] = await Promise.all([
            api.getGupLs(year, fetch),
            api.getGupPengajuan(fetch),
            api.getGupMasterData(year, fetch)
        ]);

        return {
            year: yearStr,
            lsData: lsData || [],
            transactions: transactions || [],
            fundingSources: masterData?.fundingSources || [],
            accountCodes: masterData?.accountCodes || []
        };
    } catch (e) {
        console.error('Failed to load GUP LS', e);
        return {
            year: yearStr,
            lsData: [],
            transactions: [],
            fundingSources: [],
            accountCodes: []
        };
    }
}
