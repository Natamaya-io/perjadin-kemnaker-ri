<script>
    /** @type {string} */
    export let text = '';
    /** @type {string} */
    export let className = '';
</script>

<div class="flex flex-col items-center justify-center gap-3 p-4 {className}">
    <!-- Spinner Minimalis: Cincin tipis dengan 1/4 bagian berwarna primer -->
    <div class="w-8 h-8 rounded-full border-2 border-slate-200 border-t-indigo-600 animate-spin"></div>
    
    {#if text}
        <span class="text-xs font-medium text-slate-500 tracking-wide">
            {text}
        </span>
    {/if}
</div>
