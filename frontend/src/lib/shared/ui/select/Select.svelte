<script>
    import { createEventDispatcher } from 'svelte';
    import { clickOutside } from '$lib/shared/actions/clickOutside.js';
    import { cn } from '$lib/shared/utils/utils';

    export let options = []; // Array of { value, label }
    export let value = '';
    export let placeholder = 'Pilih Item...';
    export let disabled = false;
    export let id = undefined;
    let className = undefined;
    export { className as class };

    const dispatch = createEventDispatcher();
    let isOpen = false;

    function toggle() {
        if (!disabled) isOpen = !isOpen;
    }

    function select(val, optDisabled = false) {
        if (disabled || optDisabled) return;
        value = val;
        isOpen = false;
        dispatch('change', { detail: { value: val } });
    }

    $: selectedLabel = options.find(o => String(o.value) === String(value))?.label || placeholder;

    $: triggerClass = cn(
        'flex w-full items-center justify-between rounded-xl border border-slate-200 bg-slate-50 px-4 py-2.5 md:py-3 text-sm md:text-[0.9375rem] text-slate-700 focus:outline-none focus:ring-2 focus:ring-indigo-500 focus:bg-white transition-colors',
        disabled && 'opacity-50 cursor-not-allowed pointer-events-none bg-slate-100',
        className
    );
</script>

<div class="relative w-full" use:clickOutside on:click_outside={() => isOpen = false}>
    <!-- Hidden input to maintain native form submission behavior if needed -->
    <input type="hidden" {value} {...$$restProps} />
    
    <button {id} type="button" on:click={toggle} {disabled} class={triggerClass}>
        <span class="truncate pr-4">{selectedLabel}</span>
        {#if !disabled}
        <svg class="h-4 w-4 text-slate-400 shrink-0" fill="none" viewBox="0 0 24 24" stroke="currentColor">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 9l-7 7-7-7" />
        </svg>
        {/if}
    </button>
    
    {#if isOpen}
        <div class="absolute z-50 mt-2 w-full origin-top-right rounded-xl border border-slate-100 bg-white shadow-lg overflow-hidden animate-in fade-in zoom-in-95 duration-100">
            <ul class="max-h-60 overflow-y-auto custom-scrollbar py-1">
                {#if options.length === 0}
                    <li class="py-2.5 px-4 text-sm text-slate-400 text-center select-none">Tidak ada opsi</li>
                {/if}
                {#each options as option}
                    <!-- svelte-ignore a11y-click-events-have-key-events -->
                    <!-- svelte-ignore a11y-no-noninteractive-element-interactions -->
                    <li 
                        on:click={() => select(option.value, option.disabled)} 
                        role="option" 
                        aria-selected={String(value) === String(option.value)}
                        class="relative select-none py-2.5 pl-4 pr-4 text-sm transition-colors flex items-center justify-between {option.disabled ? 'text-slate-400 cursor-not-allowed bg-slate-50/50' : (String(value) === String(option.value) ? 'font-semibold text-indigo-700 bg-indigo-50/50 hover:bg-indigo-50 cursor-pointer' : 'text-slate-700 hover:bg-slate-50 hover:text-indigo-600 cursor-pointer')}"
                    >
                        <span class="truncate">{option.label} {option.disabled ? '(Belum Tersedia)' : ''}</span>
                        {#if String(value) === String(option.value)}
                            <svg class="h-4 w-4 text-indigo-600 shrink-0" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 13l4 4L19 7" />
                            </svg>
                        {/if}
                    </li>
                {/each}
            </ul>
        </div>
    {/if}
</div>
