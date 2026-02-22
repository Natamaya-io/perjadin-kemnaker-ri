<script>
    import { logout } from '$lib/stores/auth';
</script>

<header class="md:hidden sticky top-0 z-50 w-full bg-white/90 backdrop-blur border-b border-slate-200 px-4 py-3 flex justify-between items-center shadow-sm flex-none print:hidden">
    <div class="flex items-center space-x-2">
        <img src="/kemnaker-ri.webp" alt="Logo" class="h-8 w-auto object-contain" />
        <div class="flex flex-col leading-none">
          <span class="font-serif font-bold text-base text-slate-900 tracking-tight">Perjadin</span>
          <span class="font-sans text-[10px] italic text-slate-500 tracking-widest uppercase">Protokol</span>
        </div>
    </div>
    <button class="text-slate-500" on:click={logout}>
      <svg xmlns="http://www.w3.org/2000/svg" class="h-6 w-6" fill="none" viewBox="0 0 24 24" stroke="currentColor">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M17 16l4-4m0 0l-4-4m4 4H7m6 4v1a3 3 0 01-3 3H6a3 3 0 01-3-3V7a3 3 0 013-3h4a3 3 0 013 3v1" />
      </svg>
    </button>
</header>
