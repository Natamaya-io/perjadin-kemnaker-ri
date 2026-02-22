<script>
    import { fade } from 'svelte/transition';
</script>

<div class="fixed inset-0 z-0 bg-slate-900 pointer-events-none">
    <!-- Gradient Mesh -->
    <div class="absolute inset-0 bg-gradient-to-br from-slate-900 via-[#0f172a] to-[#1e293b]"></div>
    
    <!-- Ambient Orbs -->
    <div class="absolute top-[-10%] left-[-10%] w-[60vw] h-[60vw] bg-blue-600/20 rounded-full blur-[120px] animate-pulse"></div>
    <div class="absolute bottom-[-10%] right-[-10%] w-[60vw] h-[60vw] bg-indigo-600/20 rounded-full blur-[120px] animate-pulse" style="animation-delay: 2s;"></div>
    
    <!-- Grid Pattern Overlay (Optional Texture) -->
    <div class="absolute inset-0 bg-[url('/grid.svg')] bg-center [mask-image:linear-gradient(180deg,white,rgba(255,255,255,0))] opacity-10"></div>
</div>

<!-- Back Button -->
<a href="/" class="fixed top-6 left-6 z-20 p-2.5 rounded-full bg-white/5 hover:bg-white/10 text-white/60 hover:text-white transition-all backdrop-blur-md border border-white/5" title="Kembali ke Beranda" in:fade={{ duration: 300 }}>
    <svg xmlns="http://www.w3.org/2000/svg" class="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M10 19l-7-7m0 0l7-7m-7 7h18" />
    </svg>
</a>
