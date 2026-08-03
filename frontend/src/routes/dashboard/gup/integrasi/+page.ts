import { api } from '$lib/shared/api';

export async function load({ fetch }) {
    try {
        const masterData = await api.getGupMasterData(new Date().getFullYear(), fetch);

        return {
            masterData: masterData || { fundingSources: [], procurementTypes: [], accountCodes: [] }
        };
    } catch (e) {
        console.error('Failed to load GUP master data', e);
        return {
            masterData: { fundingSources: [], procurementTypes: [], accountCodes: [] }
        };
    }
}
