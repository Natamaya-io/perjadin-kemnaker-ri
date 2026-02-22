import { writable, get } from 'svelte/store';
import { api } from '$lib/api';
import type { User } from '$lib/api/types';
import { browser } from '$app/environment';

// --- Session Management ---
// We initialize from localStorage directly for immediate hydration if available (mock mode mainly)
// But ideally, we should check `api.getCurrentUser()` on mount.
const storedUser = browser ? localStorage.getItem('user_session') : null;
const initialUser = storedUser ? JSON.parse(storedUser) : {
    email: null,
    role: null,
    loggedIn: false
};

export const userStore = writable(initialUser);

export const login = async (email: string, password?: string) => {
    try {
        const { user } = await api.login(email, password);
        userStore.set({ ...user, loggedIn: true });
        return { success: true, user };
    } catch (e) {
        console.error("Login failed", e);
        return { success: false, error: e.message };
    }
};

export const logout = async () => {
    await api.logout();
    userStore.set({ email: null, role: null, loggedIn: false });
};

// --- User Database Management (Admin) ---
export const usersStore = writable<User[]>([]);

export const loadUsers = async () => {
    try {
        const users = await api.getUsers();
        usersStore.set(users);
    } catch (e) {
        console.error("Failed to load users", e);
    }
};

// Auto-load users if we are admin (simplified check)
userStore.subscribe(u => {
    if (u.loggedIn && (u.role === 'super_admin')) {
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
