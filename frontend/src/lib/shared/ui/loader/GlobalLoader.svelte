<script>
    import { fly, fade } from 'svelte/transition';
</script>

<div class="fixed inset-0 z-[100] flex items-center justify-center bg-white/80 backdrop-blur-sm" transition:fade={{ duration: 200 }}>
    <div class="flex flex-col items-center gap-4">
        <!-- Modern Spinner -->
        <div class="relative flex items-center justify-center overflow-hidden w-24 h-24">
            <svg xmlns="http://www.w3.org/2000/svg" class="w-12 h-12 text-blue-600 animate-paper-flight drop-shadow-md" fill="currentColor" viewBox="0 0 24 24">
                <path d="M2.01 21L23 12 2.01 3 2 10l15 2-15 2z"/>
            </svg>
        </div>
        
        <!-- Text -->
        <div class="flex flex-col items-center animate-pulse" transition:fly={{ y: 5, duration: 300 }}>
            <span class="text-sm font-semibold text-slate-700 tracking-wide uppercase">Memuat</span>
            <span class="text-[10px] text-slate-500 font-medium">Mohon Tunggu...</span>
        </div>
    </div>
</div>
