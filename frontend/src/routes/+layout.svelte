<script>
  import '../app.css';
  import { userStore } from '$lib/features/auth/store';
  import { loadingStore } from '$lib/shared/stores/loading';
  import { page, navigating } from '$app/stores';
  import { browser } from '$app/environment';
  import { fly, fade } from 'svelte/transition';
  import { onMount } from 'svelte';
  import { goto } from '$app/navigation';
  
  import GlobalLoader from '$lib/shared/ui/loader/GlobalLoader.svelte';
  import Toaster from '$lib/shared/ui/toast/Toaster.svelte';
  
  // Layout Components
  import Sidebar from '$lib/shared/ui/layout/Sidebar.svelte';
  import MobileHeader from '$lib/shared/ui/layout/MobileHeader.svelte';

  $: activeRoute = $page.url?.pathname || '';
  $: isBlankPage = activeRoute.startsWith('/login') || activeRoute.startsWith('/print');
  
  // Auth Redirect
  $: if (!$userStore.loggedIn && !isBlankPage) {
      if (browser) goto('/login');
  }
  
  let isLoading = true; // Start true for initial hydration
  let navLoading = false;
  let navTimer;
  let mobileSidebarOpen = false;

  function toggleMobileSidebar() {
      mobileSidebarOpen = !mobileSidebarOpen;
  }

  function closeMobileSidebar() {
      mobileSidebarOpen = false;
  }

  // Handle navigation loading with debounce
  $: if ($navigating) {
      navTimer = setTimeout(() => {
          navLoading = true;
      }, 100); // 100ms delay before showing
      // Close sidebar on navigation (redundant if sidebar links dispatch close, but good for safety)
      mobileSidebarOpen = false;
  } else {
      clearTimeout(navTimer);
      navLoading = false;
  }

  // Handle initial hydration loading
  onMount(() => {
      // Simulate a brief "hard refresh" loading if needed, or just turn off immediately
      // A small delay feels more "app-like" on hard refresh
      setTimeout(() => {
          isLoading = false;
      }, 500);
  });

  $: showLoader = isLoading || navLoading || $loadingStore;
</script>

{#if showLoader}
    <GlobalLoader />
{/if}

<Toaster />

<div class="h-[100dvh] w-screen flex font-sans antialiased text-slate-900 overflow-hidden {activeRoute.startsWith('/login') ? 'bg-slate-900' : 'bg-slate-50'}">
  {#if $userStore.loggedIn && !isBlankPage}
    {#if mobileSidebarOpen}
        <div 
            class="fixed inset-0 z-40 bg-slate-900/50 backdrop-blur-sm md:hidden" 
            transition:fade={{ duration: 200 }} 
            on:click={closeMobileSidebar} 
            role="button" 
            tabindex="0" 
            on:keydown={(e) => e.key === 'Escape' && closeMobileSidebar()}
        ></div>
    {/if}
    <Sidebar mobileOpen={mobileSidebarOpen} on:close={closeMobileSidebar} />
  {/if}

  <div class="flex-1 flex flex-col h-full overflow-hidden w-full relative">
      {#if $userStore.loggedIn && !isBlankPage}
        <MobileHeader on:toggleSidebar={toggleMobileSidebar} />
      {/if}

      <!-- Main Content Area (Scrollable) -->
      <main class="flex-1 overflow-y-auto overflow-x-hidden no-scrollbar w-full relative">
        <div class="{$userStore.loggedIn && !isBlankPage ? 'max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8' : ''} min-h-full">
            {#key activeRoute}
                <div in:fly={{ y: 10, duration: 300, delay: 150 }} out:fly={{ y: -10, duration: 150 }} class="min-h-full">
                    <slot />
                </div>
            {/key}
        </div>
      </main>
  </div>
</div>
