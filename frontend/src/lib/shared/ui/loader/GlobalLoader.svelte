<script>
    import { fly, fade } from 'svelte/transition';
</script>

<div class="fixed inset-0 z-[100] flex items-center justify-center bg-white/80 backdrop-blur-sm" transition:fade={{ duration: 200 }}>
    <div class="flex flex-col items-center gap-4">
        <!-- Modern Spinner -->
        <div class="relative h-16 w-16">
            <div class="absolute inset-0 rounded-full border-4 border-slate-200"></div>
            <div class="absolute inset-0 rounded-full border-4 border-blue-600 border-t-transparent animate-spin"></div>
            <!-- Inner Pulse -->
            <div class="absolute inset-4 rounded-full bg-blue-50 animate-pulse"></div>
        </div>
        
        <!-- Text -->
        <div class="flex flex-col items-center animate-pulse" transition:fly={{ y: 5, duration: 300 }}>
            <span class="text-sm font-semibold text-slate-700 tracking-wide uppercase">Memuat</span>
            <span class="text-[10px] text-slate-500 font-medium">Mohon Tunggu...</span>
        </div>
    </div>
</div>
