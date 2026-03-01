<script>
    import { page } from '$app/stores';
    import { userStore, logout } from '$lib/stores/auth';
    import { createEventDispatcher } from 'svelte';
    import { getInitials } from '$lib/utils';

    const dispatch = createEventDispatcher();

    export let mobileOpen = false;

    // Use $page directly for reactivity
    $: activeRoute = $page.url.pathname;
    
    let isSidebarOpen = true;

    function toggleSidebar() {
        isSidebarOpen = !isSidebarOpen;
    }

    function handleLinkClick() {
        dispatch('close');
    }

    // Icon Paths (Heroicons v2 Outline)
    const icons = {
        dashboard: "M4 5a1 1 0 011-1h4a1 1 0 011 1v4a1 1 0 01-1 1H5a1 1 0 01-1-1V5zM14 5a1 1 0 011-1h4a1 1 0 011 1v4a1 1 0 01-1 1h-4a1 1 0 01-1-1V5zM4 15a1 1 0 011-1h4a1 1 0 011 1v4a1 1 0 01-1 1H5a1 1 0 01-1-1v-4zM14 15a1 1 0 011-1h4a1 1 0 011 1v4a1 1 0 01-1 1h-4a1 1 0 01-1-1v-4z",
        chartPie: "M10.5 6a1.5 1.5 0 113 0 1.5 1.5 0 01-3 0zM10.5 6h9.75M10.5 6a1.5 1.5 0 11-3 0m3 0a1.5 1.5 0 10-3 0M3.75 6H7.5m3 12h9.75m-9.75 0a1.5 1.5 0 01-3 0m3 0a1.5 1.5 0 00-3 0m-3.75 0H7.5m9-6h3.75m-3.75 0a1.5 1.5 0 01-3 0m3 0a1.5 1.5 0 00-3 0m-9.75 0h9.75",
        receipt: "M9 14l6-6m-5.5.5h.01m4.99 5h.01M19 21l-2-2l-2 2l-2-2l-2 2l-2-2l-2 2l-2-2l-2 2l-2-2l-2 2l-2-2l-2 2L3 21V5a2 2 0 012-2h14a2 2 0 012 2v16z",
        map: "M9 20l-5.447-2.724A1 1 0 013 16.382V5.618a1 1 0 011.447-.894L9 7m0 13l6-3m-6 3V7m6 10l4.553 2.276A1 1 0 0021 18.382V7.618a1 1 0 00-.553-.894L15 4m0 13V4m0 0L9 7",
        building: "M3.75 21h16.5M4.5 3h15M5.25 3v18m13.5-18v18M9 6.75h1.5m-1.5 3h1.5m-1.5 3h1.5m3-6H15m-1.5 3H15m-1.5 3H15M9 21v-3.375c0-.621.504-1.125 1.125-1.125h3.75c.621 0 1.125.504 1.125 1.125V21",
        globe: "M12 21a9.004 9.004 0 008.716-6.747M12 21a9.004 9.004 0 01-8.716-6.747M12 21c2.485 0 4.5-4.03 4.5-9S14.485 3 12 3m0 18c-2.485 0-4.5-4.03-4.5-9S9.515 3 12 3m0 0a8.997 8.997 0 017.843 4.582M12 3a8.997 8.997 0 00-7.843 4.582m15.686 0A11.953 11.953 0 0112 10.5c-2.998 0-5.74-1.1-7.843-2.918m15.686 0A8.959 8.959 0 0121 12c0 .778-.099 1.533-.284 2.253m0 0A17.919 17.919 0 0112 16.5c-3.162 0-6.133-.815-8.716-2.247m0 0A9.015 9.015 0 013 12c0-1.605.42-3.113 1.157-4.418",
        star: "M11.48 3.499a.562.562 0 011.04 0l2.125 5.111a.563.563 0 00.475.345l5.518.442c.499.04.701.663.321.988l-4.204 3.602a.563.563 0 00-.182.557l1.285 5.385a.562.562 0 01-.84.61l-4.725-2.885a.563.563 0 00-.586 0L6.982 20.54a.562.562 0 01-.84-.61l1.285-5.386a.562.562 0 00-.182-.557l-4.204-3.602a.563.563 0 01.321-.988l5.518-.442a.563.563 0 00.475-.345L11.48 3.5z",
        cash: "M12 6v12m-3-2.818l.879.659c1.171.879 3.07.879 4.242 0 1.172-.879 1.172-2.303 0-3.182C13.536 12.219 12.768 12 12 12c-.725 0-1.45-.22-2.003-.659-1.106-.879-1.106-2.303 0-3.182s2.9-.879 4.006 0l.415.33M21 12a9 9 0 11-18 0 9 9 0 0118 0z",
        car: "M8.25 18.75a1.5 1.5 0 01-3 0m3 0a1.5 1.5 0 00-3 0m3 0h6m-9 0H3.375a1.125 1.125 0 01-1.125-1.125V14.25m17.25 4.5a1.5 1.5 0 01-3 0m3 0a1.5 1.5 0 00-3 0m3 0h1.125c.621 0 1.129-.504 1.09-1.124a17.902 17.902 0 00-3.213-9.193 2.056 2.056 0 00-1.58-.86H14.25M16.5 18.75h-2.25m0-11.177v-.958c0-.568-.422-1.048-.987-1.106a48.554 48.554 0 00-10.026 0 1.106 1.106 0 00-.987 1.106v7.635m12-6.677v6.677m0 4.5v-4.5m0 0h-12",
        wrench: "M11.42 15.17L17.25 21A2.652 2.652 0 0021 17.25l-5.83-5.83M3 3l5.83 5.83M3 3l5.83 5.83",
        users: "M15 19.128a9.38 9.38 0 002.625.372 9.337 9.337 0 004.121-.952 4.125 4.125 0 00-7.533-2.493M15 19.128v-.003c0-1.113-.285-2.16-.786-3.07M15 19.128v.106A12.318 12.318 0 018.624 21c-2.331 0-4.512-.645-6.374-1.766l-.001-.109a6.375 6.375 0 0111.964-3.07M12 6.375a3.375 3.375 0 11-6.75 0 3.375 3.375 0 016.75 0zm8.25 2.25a2.625 2.625 0 11-5.25 0 2.625 2.625 0 015.25 0z",
        userCircle: "M17.982 18.725A7.488 7.488 0 0012 15.75a7.488 7.488 0 00-5.982 2.975m11.963 0a9 9 0 10-11.963 0m11.963 0A8.966 8.966 0 0112 21a8.966 8.966 0 01-5.982-2.275M15 9.75a3 3 0 11-6 0 3 3 0 016 0z",
        plus: "M12 4.5v15m7.5-7.5h-15"
    };

    // Menu Structure
    const menuGroups = [
        {
            items: [
                { label: "Dashboard", href: "/dashboard", icon: icons.dashboard, role: ['super_admin', 'keuangan', 'kasubag', 'protokol'] }
            ]
        },
        {
            header: "Perjalanan Dinas",
            items: [
                { label: "Pengajuan Perjalanan", href: "/dashboard/pengajuan", icon: icons.plus, role: ['super_admin', 'kasubag'] },
                { label: "Input Perjalanan", href: "/dashboard/billing", icon: icons.receipt, role: ['protokol', 'super_admin', 'kasubag'] },
                { label: "Rekap & Kalkulasi", href: "/dashboard/admin/perdin", icon: icons.cash, role: ['super_admin', 'keuangan', 'kasubag'] },
                { label: "Laporan", href: "/dashboard/laporan", icon: icons.map, role: ['protokol', 'super_admin', 'kasubag'] }
            ]
        },
        {
            header: "Keuangan & Umum",
            items: [
                { label: "VIP Bandara Halim", href: "/dashboard/vip-bandara", icon: icons.star, role: ['super_admin', 'keuangan', 'kasubag'] },
                { label: "Total Penarikan", href: "/dashboard/total-penarikan", icon: icons.cash, role: ['super_admin', 'keuangan', 'kasubag'] },
                { label: "Sewa Kendaraan", href: "/dashboard/sewa-kendaraan", icon: icons.car, role: ['super_admin', 'keuangan', 'kasubag'] },
                { label: "Pemeliharaan", href: "/dashboard/pemeliharaan", icon: icons.wrench, role: ['super_admin', 'keuangan', 'kasubag'] }
            ]
        },
        {
            header: "Zona Admin",
            role: "super_admin",
            items: [
                 { label: "User", href: "/dashboard/admin/users", icon: icons.userCircle }
            ]
        }
    ];

    function isActive(href, currentPath) {
        // Normalize paths by removing trailing slashes for consistent comparison
        const normalizedCurrent = (currentPath || '').replace(/\/+$/, '') || '/';
        const targetPath = href.replace(/\/+$/, '') || '/';

        // 1. Exact match is always active
        if (normalizedCurrent === targetPath) return true;
        
        // 2. Dashboard home (/dashboard) is STRICT exact match only.
        if (targetPath === '/dashboard') {
            return false;
        }

        // 3. For other items, allow sub-path matching
        if (normalizedCurrent.startsWith(targetPath + '/')) return true;
        
        return false;
    }

    function hasAccess(itemRole) {
        if (!itemRole) return true;
        if (Array.isArray(itemRole)) {
            return itemRole.includes($userStore.role);
        }
        return $userStore.role === itemRole;
    }

    function hasVisibleItems(group) {
        return group.items.some(item => !item.role || hasAccess(item.role));
    }
</script>

<aside class="sidebar-root fixed md:sticky inset-y-0 left-0 z-[60] flex flex-col h-[100dvh] max-w-[85vw] bg-white border-r border-slate-200 shadow-xl md:shadow-[2px_0_8px_-3px_rgba(0,0,0,0.05)] print:hidden transition-all duration-300 ease-[cubic-bezier(0.25,0.1,0.25,1)] transform {mobileOpen ? 'translate-x-0' : '-translate-x-full md:translate-x-0'} {isSidebarOpen || mobileOpen ? 'w-[280px]' : 'w-[80px]'}" style="will-change: width, transform;">
    <!-- Logo Section Wrapper -->
    <div class="sidebar-header-wrapper w-full relative flex-none z-50">
        <!-- Header Container: Transitions padding-left to center logo -->
        <div class="sidebar-header-container border-b border-slate-100/80 flex items-center h-[88px] transition-all duration-300 ease-in-out {isSidebarOpen || mobileOpen ? 'pl-6 pr-5 justify-between' : 'justify-center px-0'}">
            <div class="logo-group flex items-center transition-all duration-300 {isSidebarOpen || mobileOpen ? 'gap-3' : 'gap-0'} overflow-hidden">
                <div class="logo-image-wrapper flex-shrink-0">
                    <img src="/kemnaker-ri.webp" alt="Logo" class="h-10 w-auto object-contain" />
                </div>
                <div class="logo-text-wrapper flex flex-col leading-none transition-all duration-300 {isSidebarOpen || mobileOpen ? 'opacity-100 translate-x-0 w-auto' : 'opacity-0 -translate-x-4 w-0 overflow-hidden'}">
                    <span class="logo-title font-serif font-bold text-lg text-slate-900 tracking-tight whitespace-nowrap">Perjadin</span>
                    <span class="logo-subtitle font-sans text-[10px] font-medium text-slate-500 tracking-[0.2em] uppercase whitespace-nowrap">Protokol</span>
                </div>
            </div>
        </div>
        
        <!-- Toggle Button (Desktop Only) -->
        <button 
            class="absolute top-[34px] -right-3 hidden md:flex items-center justify-center h-6 w-6 bg-white border border-slate-200 rounded-full shadow-sm text-slate-400 hover:text-white hover:bg-slate-800 hover:border-slate-800 transition-all duration-200 z-50 focus:outline-none focus:ring-2 focus:ring-slate-200 cursor-pointer"
            on:click={toggleSidebar}
            title={isSidebarOpen ? "Collapse Sidebar" : "Expand Sidebar"}
            aria-label={isSidebarOpen ? "Collapse Sidebar" : "Expand Sidebar"}
        >
            <svg xmlns="http://www.w3.org/2000/svg" class="h-3.5 w-3.5 transition-transform duration-300 {isSidebarOpen ? 'rotate-0' : 'rotate-180'}" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2.5">
                <path stroke-linecap="round" stroke-linejoin="round" d="M15 19l-7-7 7-7" />
            </svg>
        </button>

        <!-- Close Button (Mobile Only) -->
        <button 
            class="absolute top-[30px] right-4 md:hidden flex items-center justify-center h-8 w-8 text-slate-400 hover:text-slate-600 focus:outline-none"
            on:click={() => dispatch('close')}
            title="Close Sidebar"
        >
             <svg xmlns="http://www.w3.org/2000/svg" class="h-6 w-6" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
            </svg>
        </button>
    </div>
      
    <!-- Scrollable Content Wrapper -->
    <div class="flex-1 flex flex-col overflow-y-auto overflow-x-hidden custom-scrollbar">
        <!-- Navigation Links Wrapper -->
        <nav class="sidebar-nav-root flex flex-col flex-1 px-3 py-6 space-y-1.5" aria-label="Main Navigation">
            {#each menuGroups as group}
                {#if hasAccess(group.role) && hasVisibleItems(group)}
                    {#if group.header}
                        <!-- Section Label -->
                        <div class="nav-item-wrapper w-full pt-4 pb-2 relative h-8 flex items-center justify-center transition-all duration-300 {isSidebarOpen || mobileOpen ? 'px-3' : 'px-0'}">
                             <span class="absolute text-[10px] font-bold text-slate-400 uppercase tracking-widest transition-all duration-300 {isSidebarOpen || mobileOpen ? 'opacity-100 left-3' : 'opacity-0 left-0'} whitespace-nowrap">
                                {group.header}
                             </span>
                             <div class="h-px bg-slate-200 transition-all duration-300 {isSidebarOpen || mobileOpen ? 'w-0 opacity-0' : 'w-8 opacity-100'}"></div>
                        </div>
                    {/if}

                    {#each group.items as item}
                        {#if !item.role || hasAccess(item.role)}
                            <div class="nav-item-wrapper w-full">
                                <div class="nav-item-container" title={!isSidebarOpen && !mobileOpen ? item.label : ""}>
                                    <a 
                                        class="nav-link flex items-center {isSidebarOpen || mobileOpen ? '' : 'justify-center'} px-3 py-2.5 rounded-xl transition-all duration-200 group whitespace-nowrap {isActive(item.href, activeRoute) ? 'bg-gradient-to-r from-blue-50 to-blue-50/50 text-blue-700 shadow-sm ring-1 ring-blue-100' : 'text-slate-600 hover:bg-slate-50 hover:text-slate-900'}" 
                                        href={item.href}
                                        on:click={handleLinkClick}
                                    >
                                        <span class="icon-wrapper flex-shrink-0 transition-all duration-300 ml-0 {isActive(item.href, activeRoute) ? 'text-blue-600' : 'text-slate-400 group-hover:text-slate-600'}">
                                             <svg xmlns="http://www.w3.org/2000/svg" class="h-6 w-6" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
                                                <path stroke-linecap="round" stroke-linejoin="round" d={item.icon} />
                                            </svg>
                                        </span>
                                        <span class="text-wrapper ml-3 transition-all duration-300 {isSidebarOpen || mobileOpen ? 'opacity-100 translate-x-0' : 'opacity-0 -translate-x-2 w-0 hidden'}">
                                            <span class="text-inner font-medium text-sm">{item.label}</span>
                                        </span>
                                    </a>
                                </div>
                            </div>
                        {/if}
                    {/each}
                {/if}
            {/each}
        </nav>
    </div>

    <!-- User Profile & Logout (Pinned to Bottom) -->
    <div class="sidebar-footer-wrapper border-t border-slate-100 bg-slate-50/50 flex-none z-50">
        <div class="sidebar-footer-content p-4">
            <div class="user-session-card flex items-center {isSidebarOpen || mobileOpen ? 'justify-between' : 'justify-center'} px-1 py-1 transition-all duration-300">
                <!-- User Identity Section -->
                <div class="user-identity-section flex items-center overflow-hidden transition-all duration-300 {isSidebarOpen || mobileOpen ? 'opacity-100 w-auto ml-0 gap-3' : 'opacity-0 w-0 hidden gap-0'}">
                    
                    <!-- Avatar Wrapper -->
                    <div class="avatar-wrapper flex-shrink-0">
                        <div class="avatar-container relative">
                            <div class="avatar-circle h-9 w-9 rounded-full bg-gradient-to-tr from-blue-600 to-indigo-600 text-white flex items-center justify-center shadow-md shadow-blue-500/20 ring-2 ring-white">
                                <span class="avatar-text font-bold text-xs select-none uppercase">
                                    {getInitials($userStore.name || $userStore.role || 'US')}
                                </span>
                            </div>
                            <!-- Online Status Indicator (Optional/Decorative) -->
                            <div class="status-indicator-wrapper absolute bottom-0 right-0">
                                <div class="status-dot h-2.5 w-2.5 bg-green-500 rounded-full border-2 border-white"></div>
                            </div>
                        </div>
                    </div>

                    <!-- User Details Wrapper -->
                    <div class="user-details-wrapper flex flex-col justify-center min-w-0">
                        <div class="user-label-row flex items-center">
                            <span class="user-label text-[10px] font-semibold text-slate-400 uppercase tracking-wider whitespace-nowrap leading-tight">
                                Logged in as
                            </span>
                        </div>
                        <div class="user-role-row flex items-center mt-0.5">
                            <span class="user-role text-sm font-bold text-slate-800 truncate max-w-[120px] leading-tight capitalize" title={$userStore.email}>
                                {$userStore.name || $userStore.role?.replace('_', ' ') || 'User'}
                            </span>
                        </div>
                    </div>

                </div>
                
                <!-- Session Actions Section -->
                <div class="session-actions-section flex items-center {isSidebarOpen || mobileOpen ? 'ml-0' : 'mx-auto w-full justify-center'}">
                    <div class="logout-button-container">
                        <button class="logout-btn group relative flex items-center justify-center p-2 rounded-lg text-slate-400 hover:text-red-600 hover:bg-red-50 transition-all duration-200 focus:outline-none focus:ring-2 focus:ring-red-100" title="Logout" on:click={logout}>
                            <div class="icon-wrapper transition-transform duration-200 group-hover:scale-110">
                                <svg xmlns="http://www.w3.org/2000/svg" class="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
                                    <path stroke-linecap="round" stroke-linejoin="round" d="M17 16l4-4m0 0l-4-4m4 4H7m6 4v1a3 3 0 01-3 3H6a3 3 0 01-3-3V7a3 3 0 013-3h4a3 3 0 013 3v1" />
                                </svg>
                            </div>
                        </button>
                    </div>
                </div>

            </div>
        </div>
    </div>
</aside>
