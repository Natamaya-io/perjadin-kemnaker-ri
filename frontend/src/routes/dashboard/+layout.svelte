<script>
    import { userStore } from '$lib/features/auth/store';
    import { goto } from '$app/navigation';
    import { onMount, onDestroy } from 'svelte';
    import { browser } from '$app/environment';
    import { page } from '$app/stores';

    let unsubscribeUser;
    let unsubscribePage;

    const accessRules = [
        { path: '/dashboard/admin/users', roles: ['super_admin'] },
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
</script>

<slot />
