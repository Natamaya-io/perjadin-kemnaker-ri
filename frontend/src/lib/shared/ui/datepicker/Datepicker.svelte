<script>
    import { createEventDispatcher } from 'svelte';
    import { clickOutside } from '$lib/shared/actions/clickOutside.js';
    import { cn } from '$lib/shared/utils/utils';

    export let value = ''; // 'YYYY-MM-DD'
    export let min = '';
    export let max = '';
    export let disabled = false;
    export let placeholder = 'Pilih Tanggal';
    let className = undefined;
    export { className as class };

    const dispatch = createEventDispatcher();
    let isOpen = false;

    // View state
    let currentMonth = new Date();
    
    // Synchronize currentMonth with value when opened
    function toggle() {
        if (disabled) return;
        isOpen = !isOpen;
        if (isOpen && value) {
            const parsed = new Date(value);
            if (!isNaN(parsed)) currentMonth = new Date(parsed);
        }
    }

    function selectDate(dateObj) {
        if (disabled) return;
        const year = dateObj.getFullYear();
        const month = String(dateObj.getMonth() + 1).padStart(2, '0');
        const day = String(dateObj.getDate()).padStart(2, '0');
        const newValue = `${year}-${month}-${day}`;
        value = newValue;
        isOpen = false;
        dispatch('change', { detail: { value: newValue } });
    }

    // Navigation
    function prevMonth() {
        currentMonth = new Date(currentMonth.getFullYear(), currentMonth.getMonth() - 1, 1);
    }
    
    function nextMonth() {
        currentMonth = new Date(currentMonth.getFullYear(), currentMonth.getMonth() + 1, 1);
    }

    const monthNames = ['Januari', 'Februari', 'Maret', 'April', 'Mei', 'Juni', 'Juli', 'Agustus', 'September', 'Oktober', 'November', 'Desember'];

    // Generate calendar grid
    $: days = (() => {
        const year = currentMonth.getFullYear();
        const month = currentMonth.getMonth();
        
        // Start day of month (0 = Sun, 1 = Mon, ..., 6 = Sat)
        let firstDayIndex = new Date(year, month, 1).getDay();
        // Adjust for Monday start (0 = Mon, 6 = Sun)
        firstDayIndex = firstDayIndex === 0 ? 6 : firstDayIndex - 1;
        
        const daysInMonth = new Date(year, month + 1, 0).getDate();
        const daysInPrevMonth = new Date(year, month, 0).getDate();
        
        const grid = [];
        
        // Prev month days
        for (let i = firstDayIndex - 1; i >= 0; i--) {
            grid.push({
                date: new Date(year, month - 1, daysInPrevMonth - i),
                isCurrentMonth: false
            });
        }
        
        // Current month days
        for (let i = 1; i <= daysInMonth; i++) {
            grid.push({
                date: new Date(year, month, i),
                isCurrentMonth: true
            });
        }
        
        // Next month days to fill grid (up to 42)
        const remaining = 42 - grid.length;
        for (let i = 1; i <= remaining; i++) {
            grid.push({
                date: new Date(year, month + 1, i),
                isCurrentMonth: false
            });
        }
        
        return grid;
    })();

    function formatDateDisplay(val) {
        if (!val) return placeholder;
        const d = new Date(val);
        if (isNaN(d)) return placeholder;
        return `${d.getDate()} ${monthNames[d.getMonth()]} ${d.getFullYear()}`;
    }

    function isSameDate(d1, d2) {
        if (!d1 || !d2) return false;
        return d1.getFullYear() === d2.getFullYear() && d1.getMonth() === d2.getMonth() && d1.getDate() === d2.getDate();
    }

    function isDisabled(dateObj) {
        const dateStr = `${dateObj.getFullYear()}-${String(dateObj.getMonth() + 1).padStart(2, '0')}-${String(dateObj.getDate()).padStart(2, '0')}`;
        if (min && dateStr < min) return true;
        if (max && dateStr > max) return true;
        return false;
    }

    $: today = new Date();
    $: selectedDate = value ? new Date(value) : null;
    
    $: triggerClass = cn(
        'flex w-full items-center justify-start rounded-xl border border-slate-200 bg-slate-50 px-4 py-2.5 md:py-3 text-sm md:text-[0.9375rem] text-slate-700 focus:outline-none focus:ring-2 focus:ring-indigo-500 focus:bg-white transition-colors',
        disabled && 'opacity-50 cursor-not-allowed pointer-events-none bg-slate-100',
        className
    );
</script>

<div class="relative w-full" use:clickOutside on:click_outside={() => isOpen = false}>
    <input type="hidden" {value} {...$$restProps} />
    
    <button type="button" on:click={toggle} {disabled} class={triggerClass}>
        <svg class="h-4 w-4 text-slate-400 mr-2 shrink-0" fill="none" viewBox="0 0 24 24" stroke="currentColor">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M8 7V3m8 4V3m-9 8h10M5 21h14a2 2 0 002-2V7a2 2 0 00-2-2H5a2 2 0 00-2 2v12a2 2 0 002 2z" />
        </svg>
        <span class="truncate">{formatDateDisplay(value)}</span>
    </button>
    
    {#if isOpen}
        <div class="absolute z-50 mt-2 p-3 w-72 origin-top-left rounded-xl border border-slate-100 bg-white shadow-lg animate-in fade-in zoom-in-95 duration-100">
            <div class="flex items-center justify-between mb-4">
                <button type="button" aria-label="Bulan sebelumnya" on:click|preventDefault={prevMonth} class="p-1 hover:bg-slate-100 rounded-lg text-slate-500 transition-colors">
                    <svg class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 19l-7-7 7-7" /></svg>
                </button>
                <span class="text-sm font-semibold text-slate-700">{monthNames[currentMonth.getMonth()]} {currentMonth.getFullYear()}</span>
                <button type="button" aria-label="Bulan berikutnya" on:click|preventDefault={nextMonth} class="p-1 hover:bg-slate-100 rounded-lg text-slate-500 transition-colors">
                    <svg class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 5l7 7-7 7" /></svg>
                </button>
            </div>
            
            <div class="grid grid-cols-7 gap-1 mb-1 text-center">
                <div class="text-[10px] font-bold text-slate-400 uppercase tracking-wider">Sn</div>
                <div class="text-[10px] font-bold text-slate-400 uppercase tracking-wider">Sl</div>
                <div class="text-[10px] font-bold text-slate-400 uppercase tracking-wider">Rb</div>
                <div class="text-[10px] font-bold text-slate-400 uppercase tracking-wider">Km</div>
                <div class="text-[10px] font-bold text-slate-400 uppercase tracking-wider">Jm</div>
                <div class="text-[10px] font-bold text-slate-400 uppercase tracking-wider text-rose-400">Sb</div>
                <div class="text-[10px] font-bold text-slate-400 uppercase tracking-wider text-rose-400">Mn</div>
            </div>
            
            <div class="grid grid-cols-7 gap-1 text-center">
                {#each days as day}
                    {#if !day.isCurrentMonth}
                        <button type="button" class="h-8 w-full text-xs flex items-center justify-center rounded-lg text-slate-300 hover:bg-slate-50 transition-colors" disabled>
                            {day.date.getDate()}
                        </button>
                    {:else if isDisabled(day.date)}
                        <button type="button" class="h-8 w-full text-xs flex items-center justify-center rounded-lg text-slate-300 bg-slate-50 opacity-50 cursor-not-allowed" disabled>
                            {day.date.getDate()}
                        </button>
                    {:else if selectedDate && isSameDate(day.date, selectedDate)}
                        <button type="button" on:click|preventDefault={() => selectDate(day.date)} class="h-8 w-full text-xs flex items-center justify-center rounded-lg bg-indigo-600 text-white font-semibold shadow-sm hover:bg-indigo-700 transition-colors">
                            {day.date.getDate()}
                        </button>
                    {:else if isSameDate(day.date, today)}
                        <button type="button" on:click|preventDefault={() => selectDate(day.date)} class="h-8 w-full text-xs flex items-center justify-center rounded-lg border border-indigo-200 text-indigo-700 font-semibold hover:bg-indigo-50 transition-colors">
                            {day.date.getDate()}
                        </button>
                    {:else}
                        <button type="button" on:click|preventDefault={() => selectDate(day.date)} class="h-8 w-full text-xs flex items-center justify-center rounded-lg text-slate-700 hover:bg-slate-100 hover:text-indigo-600 transition-colors">
                            {day.date.getDate()}
                        </button>
                    {/if}
                {/each}
            </div>
        </div>
    {/if}
</div>
