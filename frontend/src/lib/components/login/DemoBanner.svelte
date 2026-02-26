<script>
    import { createEventDispatcher, onMount } from 'svelte';
    import { browser } from '$app/environment';
    
    const dispatch = createEventDispatcher();
    // Using relative path, assumes proxy or CORS configured, or full URL from env
    const API_URL = import.meta.env.VITE_API_URL || 'http://localhost:8081/api/v1';

    function fill(email, password) {
        dispatch('fill', { email, password: password || '123' });
    }

    let demoUsers = [
        { name: 'Super Admin', email: 'superadmin@kemnaker.go.id', password: '123' },
        { name: 'Keuangan', email: 'keuangan@kemnaker.go.id', password: '123' },
        { name: 'Kasubag', email: 'kasubag@kemnaker.go.id', password: '123' },
        { name: 'Protokol Vito', email: 'vito@kemnaker.go.id', password: '123' },
        { name: 'Protokol Keneth', email: 'keneth@kemnaker.go.id', password: '123' }
    ];

    onMount(async () => {
        if (browser) {
            try {
                const res = await fetch(`${API_URL}/auth/demo-users`);
                if (res.ok) {
                    const users = await res.json();
                    if (Array.isArray(users) && users.length > 0) {
                        demoUsers = users.map(u => ({
                            name: u.name,
                            email: u.email,
                            password: u.password // From backend
                        }));
                    }
                }
            } catch (e) {
                console.error("Failed to fetch demo users", e);
            }
        }
    });
</script>

<div class="bg-blue-50/60 border border-blue-100/60 rounded-xl p-3.5 flex flex-col gap-1.5">
    <div class="flex items-center gap-2 text-xs font-bold text-blue-800">
        <svg xmlns="http://www.w3.org/2000/svg" class="h-3.5 w-3.5" viewBox="0 0 20 20" fill="currentColor">
            <path fill-rule="evenodd" d="M18 10a8 8 0 11-16 0 8 8 0 0116 0zm-7-4a1 1 0 11-2 0 1 1 0 012 0zM9 9a1 1 0 000 2v3a1 1 0 001 1h1a1 1 0 100-2v-3a1 1 0 00-1-1H9z" clip-rule="evenodd" />
        </svg>
        Akun Demo (Klik untuk isi otomatis)
    </div>
    <div class="text-[10px] text-blue-700/80 font-mono pl-5 grid grid-cols-1 gap-0.5 max-h-[200px] overflow-y-auto">
        {#each demoUsers as user}
            <button class="flex justify-between w-full hover:bg-blue-100/50 px-1 rounded cursor-pointer text-left transition-colors" on:click={() => fill(user.email, user.password)}>
                <span class="truncate mr-2">{user.name}:</span> <span class="font-semibold">{user.email}</span>
            </button>
        {/each}
    </div>
</div>
