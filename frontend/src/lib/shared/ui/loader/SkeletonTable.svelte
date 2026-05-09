<script>
    export let rows = 5;
    export let columns = 5;
</script>

<div class="w-full space-y-4 animate-pulse p-4">
    <!-- Header Skeleton -->
    <div class="flex space-x-4 border-b border-slate-100 pb-4">
        {#each Array(columns) as _}
            <div class="h-4 bg-slate-200 rounded w-full"></div>
        {/each}
    </div>

    <!-- Rows Skeleton -->
    {#each Array(rows) as _}
        <div class="flex space-x-4 py-3 border-b border-slate-50 last:border-0">
            {#each Array(columns) as _, i}
                {#if i === 0}
                    <div class="h-10 w-10 bg-slate-100 rounded-lg shrink-0"></div>
                {:else}
                    <div class="space-y-2 w-full pt-1">
                        <div class="h-3 bg-slate-100 rounded w-3/4"></div>
                        <div class="h-2 bg-slate-50 rounded w-1/2"></div>
                    </div>
                {/if}
            {/each}
        </div>
    {/each}
</div>
