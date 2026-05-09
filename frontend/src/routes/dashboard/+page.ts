import { api } from '$lib/shared/api';
import type { PageLoad } from './$types';

export const load: PageLoad = async () => {
    try {
        const stats = await api.getDashboardSummary();
        return {
            stats
        };
    } catch (error) {
        console.error("Failed to load dashboard summary:", error);
        return {
            stats: null
        };
    }
};
