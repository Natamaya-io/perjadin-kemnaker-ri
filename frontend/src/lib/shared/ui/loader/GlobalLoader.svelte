<script>
    import { fly, fade } from 'svelte/transition';
    import LottieLoader from './LottieLoader.svelte';
</script>

<div class="fixed inset-0 z-[100] flex items-center justify-center bg-white/80 backdrop-blur-sm" transition:fade={{ duration: 200 }}>
    <div class="flex flex-col items-center gap-1">
        <!-- Lottie Animation -->
        <LottieLoader size="200px" />
        
        <!-- Text -->
        <div class="flex flex-col items-center animate-pulse -mt-4" transition:fly={{ y: 5, duration: 300 }}>
            <span class="text-sm font-semibold text-slate-700 tracking-wide uppercase">Memuat</span>
            <span class="text-[10px] text-slate-500 font-medium">Mohon Tunggu...</span>
        </div>
    </div>
</div>
