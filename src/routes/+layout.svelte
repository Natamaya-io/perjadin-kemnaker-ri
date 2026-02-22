<script>
  import '../app.css';
  import { userStore } from '$lib/stores/auth';
  import { page, navigating } from '$app/stores';
  import { fly } from 'svelte/transition';
  import { onMount } from 'svelte';
  
  import GlobalLoader from '$lib/components/ui/loader/GlobalLoader.svelte';
  import Toaster from '$lib/components/ui/toast/Toaster.svelte';
  
  // Layout Components
  import Sidebar from '$lib/components/layout/Sidebar.svelte';
  import MobileHeader from '$lib/components/layout/MobileHeader.svelte';
  import MobileNavigation from '$lib/components/layout/MobileNavigation.svelte';

  $: activeRoute = $page.url?.pathname || '';
  
  let isLoading = true; // Start true for initial hydration
  let navLoading = false;
  let navTimer;

  // Handle navigation loading with debounce
  $: if ($navigating) {
      navTimer = setTimeout(() => {
          navLoading = true;
      }, 100); // 100ms delay before showing
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

  $: showLoader = isLoading || navLoading;
</script>

{#if showLoader}
    <GlobalLoader />
{/if}

<Toaster />

<div class="h-[100dvh] w-screen flex font-sans antialiased text-slate-900 overflow-hidden {activeRoute.startsWith('/login') ? 'bg-slate-900' : 'bg-slate-50'}">
  {#if $userStore.loggedIn && !activeRoute.startsWith('/login')}
    <Sidebar />
  {/if}

  <div class="flex-1 flex flex-col h-full overflow-hidden w-full relative">
      {#if $userStore.loggedIn && !activeRoute.startsWith('/login')}
        <MobileHeader />
      {/if}

      <!-- Main Content Area (Scrollable) -->
      <main class="flex-1 overflow-y-auto overflow-x-hidden no-scrollbar w-full relative">
        <div class="{$userStore.loggedIn && !activeRoute.startsWith('/login') ? 'max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8 pb-20 md:pb-8' : ''} min-h-full">
            {#key activeRoute}
                <div in:fly={{ y: 10, duration: 300, delay: 150 }} out:fly={{ y: -10, duration: 150 }} class="min-h-full">
                    <slot />
                </div>
            {/key}
        </div>
      </main>

      {#if $userStore.loggedIn && !activeRoute.startsWith('/login')}
        <MobileNavigation />
      {/if}
  </div>
</div>
