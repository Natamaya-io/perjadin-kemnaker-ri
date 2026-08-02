<script>
    import { userStore } from '$lib/features/auth/store';
    import { goto } from '$app/navigation';
    import { onMount, onDestroy } from 'svelte';
    import { browser } from '$app/environment';
    import { page, navigating } from '$app/stores';
    import ChangePasswordModal from '$lib/features/auth/ui/ChangePasswordModal.svelte';

    let unsubscribeUser;
    let unsubscribePage;
    let isPasswordModalVisible = false;

    // Trigger modal visibility only once per session when requirement is detected
    $: if ($userStore.requirePasswordChange && !$userStore.passwordModalDismissed && !isPasswordModalVisible && browser) {
        isPasswordModalVisible = true;
    }

    // Persist the dismissed state to the store so it survives navigation
    $: if (!isPasswordModalVisible && $userStore.requirePasswordChange && browser) {
        userStore.update(u => ({ ...u, passwordModalDismissed: true }));
    }

    const accessRules = [
        { path: '/dashboard/admin/users', roles: ['super_admin', 'kasubag', 'protokol'] },
        { path: '/dashboard/pengajuan', roles: ['super_admin', 'kasubag'] },
        { path: '/dashboard/pengajuan/new', roles: ['super_admin', 'kasubag'] },
        { path: '/dashboard/admin/perdin', roles: ['super_admin', 'kasubag'] },
        { path: '/dashboard/billing-protokol', roles: ['protokol', 'super_admin', 'kasubag'] },
        { path: '/dashboard/laporan', roles: ['protokol', 'super_admin', 'kasubag'] },
        { path: '/dashboard/vip-bandara', roles: ['super_admin', 'kasubag'] },
        { path: '/dashboard/total-penarikan', roles: ['super_admin', 'kasubag'] },
        { path: '/dashboard/sewa-kendaraan', roles: ['super_admin', 'kasubag'] },
        { path: '/dashboard/pemeliharaan', roles: ['super_admin', 'kasubag'] }
    ];

    function checkAccess(pathname, role) {
        // Find if any rule matches the current path
        for (const rule of accessRules) {
            if (pathname.startsWith(rule.path)) {
                if (!rule.roles.includes(role)) {
                    return false; // Access denied
                }
            }
        }
        return true; // No rule restricting this path, or role is allowed
    }

    onMount(() => {
        if (browser && !$userStore.loggedIn) {
            goto('/login');
            return;
        }

        if (browser && $userStore.loggedIn) {
             if (!checkAccess($page.url.pathname, $userStore.role)) {
                 goto('/dashboard');
             }
        }

        unsubscribeUser = userStore.subscribe(user => {
            if (browser) {
                if (!user.loggedIn) {
                    goto('/login');
                } else if (!checkAccess($page.url.pathname, user.role)) {
                    goto('/dashboard');
                }
            }
        });

        unsubscribePage = page.subscribe(p => {
             if (browser && $userStore.loggedIn) {
                  if (!checkAccess(p.url.pathname, $userStore.role)) {
                      goto('/dashboard');
                  }
             }
        });
    });

    onDestroy(() => {
        if (unsubscribeUser) unsubscribeUser();
        if (unsubscribePage) unsubscribePage();
    });
    $: currentPath = $page.url.pathname;
    $: isUnderDevelopment = 
        currentPath.startsWith('/dashboard/gup/pengajuan') || 
        currentPath.startsWith('/dashboard/gup/laporan') || 
        currentPath.startsWith('/dashboard/gup/ls') || 
        currentPath.startsWith('/dashboard/gup/data') ||
        (currentPath.startsWith('/dashboard/pengajuan/new') && $page.url.searchParams.get('type') === 'dalam_kota');
</script>

{#if isUnderDevelopment}
    <div class="flex flex-col items-center justify-center min-h-[70vh] p-8 text-center bg-white rounded-3xl border border-slate-200 shadow-sm m-6">
        <svg xmlns="http://www.w3.org/2000/svg" class="h-24 w-24 text-slate-300 mb-6" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="1.5">
            <path stroke-linecap="round" stroke-linejoin="round" d="M10.325 4.317c.426-1.756 2.924-1.756 3.35 0a1.724 1.724 0 002.573 1.066c1.543-.94 3.31.826 2.37 2.37a1.724 1.724 0 001.065 2.572c1.756.426 1.756 2.924 0 3.35a1.724 1.724 0 00-1.066 2.573c.94 1.543-.826 3.31-2.37 2.37a1.724 1.724 0 00-2.572 1.065c-.426 1.756-2.924 1.756-3.35 0a1.724 1.724 0 00-2.573-1.066c-1.543.94-3.31-.826-2.37-2.37a1.724 1.724 0 00-1.065-2.572c-1.756-.426-1.756-2.924 0-3.35a1.724 1.724 0 001.066-2.573c-.94-1.543.826-3.31 2.37-2.37.996.608 2.296.07 2.572-1.065z" />
            <path stroke-linecap="round" stroke-linejoin="round" d="M15 12a3 3 0 11-6 0 3 3 0 016 0z" />
        </svg>
        <h2 class="text-3xl font-bold text-slate-800 mb-3">Segera Hadir</h2>
        <p class="text-slate-500 max-w-md mx-auto text-lg">Fitur ini masih dalam tahap pengembangan aktif dan belum siap digunakan untuk produksi.</p>
        <button on:click={() => history.back()} class="mt-8 px-6 py-3 bg-indigo-600 text-white rounded-xl font-medium hover:bg-indigo-700 transition-colors shadow-sm">
            Kembali ke Halaman Sebelumnya
        </button>
    </div>
{:else}
    <slot />
{/if}

<ChangePasswordModal 
    bind:isOpen={isPasswordModalVisible} 
    on:close={() => {
        userStore.update(u => ({ ...u, passwordModalDismissed: true }));
        isPasswordModalVisible = false;
    }}
    on:success={() => {
        userStore.update(u => ({ ...u, requirePasswordChange: false }));
        isPasswordModalVisible = false;
    }} 
/>

{#if $navigating}
    <div class="global-progress-bar"></div>
{/if}

<style>
    .global-progress-bar {
        position: fixed;
        top: 0;
        left: 0;
        right: 0;
        height: 3px;
        background-color: #3b82f6; /* Tailwind blue-500 */
        z-index: 99999;
        animation: loader-progress 1.5s cubic-bezier(0.4, 0, 0.2, 1) infinite;
        transform-origin: left;
    }

    @keyframes loader-progress {
        0% { transform: scaleX(0); opacity: 1; }
        50% { transform: scaleX(0.7); opacity: 1; }
        100% { transform: scaleX(1); opacity: 0; }
    }
</style>
