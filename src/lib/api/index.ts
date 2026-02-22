import { MockApiClient } from './mock';
import { RealApiClient } from './real';

// Toggle this via .env
const USE_REAL_API = import.meta.env.VITE_USE_REAL_API === 'true';

export const api = USE_REAL_API ? new RealApiClient() : new MockApiClient();
