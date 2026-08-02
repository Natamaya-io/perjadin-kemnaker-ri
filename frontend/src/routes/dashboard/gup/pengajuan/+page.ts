import { api } from '$lib/shared/api';

export async function load({ fetch }) {
    try {
        const [transactions, masterData] = await Promise.all([
            api.getGupPengajuan(fetch),
            api.getGupMasterData(2024, fetch) // For now, hardcode 2024 or get current year
        ]);

        return {
            transactions: transactions || [],
            masterData: masterData || { fundingSources: [], procurementTypes: [] }
        };
    } catch (e) {
        console.error('Failed to load GUP transactions', e);
        return {
            transactions: [],
            masterData: { fundingSources: [], procurementTypes: [] }
        };
    }
}
