<script>
  import { createEventDispatcher } from 'svelte';
  import { portal } from '$lib/actions/portal';
  export let open = false;
  const dispatch = createEventDispatcher();
  
  function close() {
    dispatch('close');
    open = false;
  }
</script>

{#if open}
  <div use:portal>
    <div class="fixed inset-0 z-50 bg-black/80 backdrop-blur-sm" role="button" tabindex="0" on:click={close} on:keydown={(e) => e.key === 'Escape' && close()}></div>
    <div class="fixed left-[50%] top-[50%] z-50 grid w-full max-w-lg translate-x-[-50%] translate-y-[-50%] gap-4 border bg-background p-6 shadow-lg duration-200 sm:rounded-lg bg-white">
      <slot />
      <button class="absolute right-4 top-4 rounded-sm opacity-70 ring-offset-background transition-opacity hover:opacity-100 focus:outline-none disabled:pointer-events-none data-[state=open]:bg-accent data-[state=open]:text-muted-foreground" on:click={close}>
        <svg xmlns="http://www.w3.org/2000/svg" width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" class="h-4 w-4"><line x1="18" x2="6" y1="6" y2="18"></line><line x1="6" x2="18" y1="6" y2="18"></line></svg>
        <span class="sr-only">Close</span>
      </button>
    </div>
  </div>
{/if}
