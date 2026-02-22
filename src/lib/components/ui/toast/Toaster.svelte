<script>
    import { toast } from '$lib/stores/toast';
    import ToastItem from './Toast.svelte';
</script>

<div class="fixed bottom-4 right-4 z-[100] flex flex-col items-end gap-2 pointer-events-none">
    <div class="flex flex-col items-end pointer-events-auto">
        {#each $toast as item (item.id)}
            <ToastItem {item} />
        {/each}
    </div>
</div>
