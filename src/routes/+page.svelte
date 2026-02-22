<script>
    import Button from '$lib/components/ui/button/Button.svelte';
    import { userStore } from '$lib/stores/auth';
    import { fly, fade } from 'svelte/transition';
    import { onMount } from 'svelte';

    let mounted = false;
    onMount(() => mounted = true);
</script>

<div class="relative min-h-screen w-full flex flex-col items-center justify-center overflow-hidden bg-[#0f172a] text-white selection:bg-blue-500/30">
    
    <!-- Cinematic Background -->
    <div class="absolute inset-0 z-0">
        <!-- Deep radial gradient for depth -->
        <div class="absolute inset-0 bg-[radial-gradient(circle_at_center,_var(--tw-gradient-stops))] from-[#1e293b] via-[#0f172a] to-[#020617] opacity-80"></div>
        
        <!-- Subtle noise texture for premium feel -->
        <div class="absolute inset-0 bg-[url('https://grainy-gradients.vercel.app/noise.svg')] opacity-[0.03] mix-blend-overlay"></div>
        
        <!-- Animated Ethereal Glow -->
        <div class="absolute top-1/2 left-1/2 -translate-x-1/2 -translate-y-1/2 w-[800px] h-[800px] bg-blue-900/20 rounded-full blur-[120px] animate-pulse" style="animation-duration: 4s;"></div>
    </div>

    <!-- Minimalist Navbar (Absolute Top) -->
    <nav class="absolute top-0 w-full z-20 px-8 py-8 flex justify-between items-center max-w-7xl mx-auto">
        <div class="flex items-center gap-4 opacity-0 animate-[fadeIn_1s_ease-out_forwards]">
            <img src="/kemnaker-ri.webp" alt="Logo Kemnaker" class="h-12 w-auto drop-shadow-lg" />
        </div>
        <div class="opacity-0 animate-[fadeIn_1s_ease-out_0.2s_forwards]">
            {#if !$userStore.loggedIn}
                <a href="/login" class="text-sm font-medium tracking-wide text-slate-300 hover:text-white transition-colors duration-300">
                    Masuk Akun
                </a>
            {:else}
                <a href="/dashboard/pengajuan/new" class="text-sm font-medium tracking-wide text-slate-300 hover:text-white transition-colors duration-300">
                    Dashboard
                </a>
            {/if}
        </div>
    </nav>

    <!-- Main Content (Centered) -->
    <main class="relative z-10 container max-w-4xl mx-auto px-4 text-center">
        {#if mounted}
            <div in:fly={{ y: 20, duration: 1200, delay: 100 }} class="space-y-8 flex flex-col items-center">
                
                <!-- Center Logo -->
                <img src="/kemnaker-ri.webp" alt="Logo Kemnaker RI" class="h-32 w-auto mb-4 drop-shadow-2xl animate-[fadeIn_1.5s_ease-out_forwards]" />

                <!-- Elegant Typography -->
                <div class="space-y-4">
                    <h1 class="text-5xl md:text-7xl lg:text-8xl font-serif font-medium tracking-tight leading-tight text-white/90">
                        Perjadin <br />
                        <span class="italic text-white/60">Protokol</span>
                    </h1>
                    <div class="h-px w-24 bg-gradient-to-r from-transparent via-white/20 to-transparent mx-auto my-6"></div>
                    <p class="text-lg md:text-xl font-light text-slate-400 max-w-lg mx-auto leading-relaxed tracking-wide">
                        Sistem manajemen perjalanan dinas dan protokol Kementerian Ketenagakerjaan Republik Indonesia.
                    </p>
                </div>

                <!-- Minimalist Action Buttons -->
                <div class="flex flex-col sm:flex-row items-center justify-center gap-5 pt-8">
                    <a href={$userStore.loggedIn ? "/dashboard/pengajuan/new" : "/login"}>
                        <button class="group relative px-8 py-3 bg-white text-slate-900 rounded-full font-medium text-sm transition-all duration-300 hover:bg-slate-200 hover:shadow-[0_0_20px_rgba(255,255,255,0.3)] hover:scale-105 active:scale-95">
                            <span class="relative z-10">Buat Pengajuan</span>
                        </button>
                    </a>
                    
                    <a href={$userStore.loggedIn ? "/dashboard/admin/perdin" : "/login"}>
                        <button class="group px-8 py-3 bg-transparent border border-white/10 text-white rounded-full font-medium text-sm transition-all duration-300 hover:bg-white/5 hover:border-white/30 active:scale-95">
                            Portal Admin
                        </button>
                    </a>
                </div>

            </div>
        {/if}
    </main>

    <!-- Minimal Footer (Absolute Bottom) -->
    <footer class="absolute bottom-0 w-full py-6 text-center z-10 opacity-0 animate-[fadeIn_1s_ease-out_1s_forwards]">
        <p class="text-[10px] uppercase tracking-[0.2em] text-slate-600">
            &copy; {new Date().getFullYear()} Biro Umum Kemnaker RI
        </p>
    </footer>

</div>

<style>
    /* Custom Animations */
    @keyframes fadeIn {
        from { opacity: 0; transform: translateY(-10px); }
        to { opacity: 1; transform: translateY(0); }
    }
</style>
