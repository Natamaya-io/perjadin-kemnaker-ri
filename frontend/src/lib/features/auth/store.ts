import { writable, get } from 'svelte/store';
import { api } from '$lib/shared/api';
import type { User } from '$lib/shared/api/types';
import { browser } from '$app/environment';
import { loadRecords, clearStores } from '$lib/features/pengajuan/store';
import { loadMasterData } from '$lib/shared/stores/master-data';

const isBrowser = typeof window !== 'undefined';

// --- Session Management ---
// We initialize from localStorage directly for immediate hydration.
// But ideally, we should check `api.getCurrentUser()` on mount.
const storedUser = isBrowser ? localStorage.getItem('user_session_v2') : null;
const initialUser = storedUser ? JSON.parse(storedUser) : {
    email: null,
    role: null,
    loggedIn: false
};

export const userStore = writable(initialUser);

// Persist to localStorage whenever userStore changes
userStore.subscribe(val => {
    if (isBrowser) {
        localStorage.setItem('user_session_v2', JSON.stringify(val));
    }
});

// Auto-load data if session exists on init
if (initialUser.loggedIn && isBrowser) {
    loadMasterData();
    loadRecords();
}

export const login = async (email: string, password?: string) => {
    try {
        const { user } = await api.login(email, password);
        userStore.set({ ...user, loggedIn: true });
        
        // Load data on login
        loadMasterData();
        loadRecords();
        
        return { success: true, user };
    } catch (e: any) {
        console.error("Login failed", e);
        return { success: false, error: e.message };
    }
};

export const logout = async () => {
    await api.logout();
    userStore.set({ email: null, role: null, loggedIn: false });
    clearStores();
};

// Auto-logout on 401 Unauthorized from any API call
api.setUnauthorizedHandler(() => {
    console.warn("Session expired or unauthorized. Logging out...");
    userStore.set({ email: null, role: null, loggedIn: false });
    clearStores();
    if (isBrowser) {
        window.location.href = '/login';
    }
});

// --- User Database Management (Admin) ---
export const usersStore = writable<User[]>([]);

export const loadUsers = async () => {
    if (!isBrowser || !localStorage.getItem('auth_token')) return;
    try {
        const users = await api.getUsers();
        usersStore.set(users);
    } catch (e) {
        console.error("Failed to load users", e);
    }
};

// Auto-load users if we are admin or kasubag so they can select protocols
userStore.subscribe(u => {
    if (u.loggedIn && (u.role === 'super_admin' || u.role === 'kasubag')) {
        loadUsers();
    }
});

export async function addUser(user: Omit<User, 'id'>) {
    try {
        const newUser = await api.createUser(user);
        usersStore.update(users => [...users, newUser]);
        return newUser;
    } catch (e) {
        console.error("Failed to add user", e);
        throw e;
    }
}

export async function updateUser(id: string | number, updatedData: Partial<User>) {
    try {
        const updated = await api.updateUser(id, updatedData);
        usersStore.update(users => users.map(u => u.id === id ? updated : u));
        return updated;
    } catch (e) {
        console.error("Failed to update user", e);
        throw e;
    }
}

export async function removeUser(id: string | number) {
    try {
        await api.deleteUser(id);
        usersStore.update(users => users.filter(u => u.id !== id));
    } catch (e) {
        console.error("Failed to remove user", e);
        throw e;
    }
}
