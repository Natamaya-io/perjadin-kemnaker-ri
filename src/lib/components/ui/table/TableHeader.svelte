<script>
    import { cn } from '$lib/utils';
    export let className = '';
</script>

<thead class={cn('[&_tr]:border-b', className)} {...$$restProps}>
    <slot />
</thead>
