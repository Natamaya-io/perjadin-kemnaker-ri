<script>
    import { cn, getInitials } from '$lib/utils';
    import Button from '$lib/components/ui/button/Button.svelte';

    export let record;
</script>

<div class="group bg-white rounded-xl border border-slate-200 shadow-sm hover:shadow-md transition-all duration-200 overflow-hidden flex flex-col">
    <div class="p-6 flex-1 space-y-4">
        <div class="flex justify-between items-start">
            <span class={cn("px-2.5 py-1 rounded-full text-[10px] font-bold uppercase tracking-wide border", 
                record.reportStatus === 'Completed' ? "bg-emerald-50 text-emerald-700 border-emerald-100" : "bg-amber-50 text-amber-700 border-amber-100")}>
                {record.reportStatus === 'Completed' ? 'Selesai' : 'Pending'}
            </span>
            <span class="text-xs font-mono text-slate-400">{record.spd}</span>
        </div>
        
        <div>
            <h3 class="font-bold text-lg text-slate-800 line-clamp-2 group-hover:text-blue-600 transition-colors">{record.purpose}</h3>
            <p class="text-sm text-slate-500 mt-1 flex items-center gap-1">
                <svg xmlns="http://www.w3.org/2000/svg" class="h-4 w-4 text-slate-400" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M17.657 16.657L13.414 20.9a1.998 1.998 0 01-2.827 0l-4.244-4.243a8 8 0 1111.314 0z" />
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 11a3 3 0 11-6 0 3 3 0 016 0z" />
                </svg>
                {record.location}, {record.province}
            </p>
        </div>
        
        <div class="flex items-center gap-3 text-xs text-slate-500 pt-2 border-t border-slate-50">
            <div class="flex items-center gap-1">
                <svg xmlns="http://www.w3.org/2000/svg" class="h-4 w-4 text-slate-400" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M8 7V3m8 4V3m-9 8h10M5 21h14a2 2 0 002-2V7a2 2 0 00-2-2H5a2 2 0 00-2 2v12a2 2 0 002 2z" />
                </svg>
                {record.startDate}
            </div>
            <span>&mdash;</span>
            <div>{record.endDate}</div>
        </div>
    </div>
    
    <div class="px-6 py-4 bg-slate-50 border-t border-slate-100 flex items-center justify-between">
         <div class="flex -space-x-2 overflow-hidden">
            <!-- Avatar placeholders -->
            <div class="h-6 w-6 rounded-full ring-2 ring-white bg-slate-200 flex items-center justify-center text-[10px] font-bold text-slate-500 leading-none">
                {getInitials(record.employee.name)}
            </div>
         </div>
         <div class="flex gap-2">
             {#if record.reportStatus === 'Completed'}
                <Button variant="ghost" size="sm" class="h-8 w-8 p-0 rounded-full hover:bg-slate-200 text-slate-500" title="Cetak Laporan" on:click={() => window.open(`/print?type=laporan&spd=${encodeURIComponent(record.spd)}`, '_blank')}>
                    <svg xmlns="http://www.w3.org/2000/svg" class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M17 17h2a2 2 0 002-2v-4a2 2 0 00-2-2H5a2 2 0 00-2 2v4a2 2 0 002 2h2m2 4h6a2 2 0 002-2v-4a2 2 0 00-2-2H9a2 2 0 00-2 2v4a2 2 0 002 2zm8-12V5a2 2 0 00-2-2H9a2 2 0 00-2 2v4h10z" />
                    </svg>
                </Button>
             {/if}
             <a href={`/dashboard/laporan/${encodeURIComponent(record.spd)}`}>
                <Button size="sm" class={cn("h-8 text-xs font-medium shadow-sm transition-all", 
                    record.reportStatus === 'Completed' 
                    ? "bg-white text-slate-700 border border-slate-200 hover:bg-slate-50 hover:text-blue-600" 
                    : "bg-blue-600 hover:bg-blue-700 text-white shadow-blue-500/20")}>
                    {record.reportStatus === 'Completed' ? 'Edit Laporan' : 'Buat Laporan'}
                </Button>
             </a>
         </div>
    </div>
</div>
