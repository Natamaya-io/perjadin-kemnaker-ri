<script>
    import { fade } from 'svelte/transition';
</script>

<!-- Non-blocking top progress bar instead of a full-screen overlay -->
<div class="fixed top-0 left-0 right-0 z-[100] h-1 bg-transparent overflow-hidden" transition:fade={{ duration: 200 }}>
    <div class="h-full bg-blue-600 animate-progress origin-left"></div>
</div>

<style>
    @keyframes progress {
        0% { width: 0%; opacity: 1; }
        50% { width: 70%; opacity: 1; }
        100% { width: 100%; opacity: 0; }
    }
    .animate-progress {
        animation: progress 1.5s infinite cubic-bezier(0.4, 0, 0.2, 1);
    }
</style>
