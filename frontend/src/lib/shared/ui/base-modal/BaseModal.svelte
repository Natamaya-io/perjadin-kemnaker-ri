<script>
    import { createEventDispatcher } from 'svelte';
    import { portal } from '$lib/shared/actions/portal';
    
    export let open = false;
    export let maxWidth = 'max-w-4xl'; // Default width for Detail Modal (bisa diganti misal max-w-2xl)
    
    const dispatch = createEventDispatcher();
    
    function close() {
        if (!open) return;
        open = false;
        dispatch('close');
    }
</script>

{#if open}
  <div use:portal>
    <!-- Background Overlay -->
    <div 
        class="fixed inset-0 z-[100] bg-slate-900/90" 
        role="button" 
        tabindex="0" 
        aria-label="Close modal" 
        on:click={close} 
        on:keydown={(e) => e.key === 'Escape' && close()}
    ></div>
    
    <!-- Modal Dialog -->
    <div class="fixed left-[50%] top-[50%] z-[100] w-full {maxWidth} translate-x-[-50%] translate-y-[-50%] border border-slate-200 bg-white shadow-2xl sm:rounded-2xl overflow-hidden max-h-[90vh] flex flex-col animate-in fade-in zoom-in-95 duration-200">
        <!-- Header -->
        <div class="px-6 py-5 border-b border-slate-100 flex justify-between items-start bg-white">
            <div class="flex-1 min-w-0 pr-4">
                <slot name="header" />
            </div>
            <button 
                class="rounded-full p-2 hover:bg-slate-100 text-slate-400 hover:text-slate-600 transition-colors" 
                on:click={close} 
                aria-label="Close"
            >
                <svg xmlns="http://www.w3.org/2000/svg" width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" class="h-5 w-5">
                    <line x1="18" x2="6" y1="6" y2="18"></line>
                    <line x1="6" x2="18" y1="6" y2="18"></line>
                </svg>
            </button>
        </div>
        
        <!-- Body -->
        <div class="flex-1 overflow-y-auto p-6 bg-slate-50/50">
            <slot name="body" />
        </div>
        
        <!-- Footer (Opsional) -->
        {#if $$slots.footer}
        <div class="px-6 py-5 border-t border-slate-100 bg-white flex flex-col sm:flex-row justify-between items-center gap-4">
            <slot name="footer" />
        </div>
        {/if}
    </div>
  </div>
{/if}
