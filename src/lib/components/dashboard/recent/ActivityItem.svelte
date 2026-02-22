<script>
    import { getInitials } from '$lib/utils';
    
    export let record;
    export let formatIDR;

    const statusColors = {
        'Approved': 'bg-emerald-100 text-emerald-700 border-emerald-200 ring-emerald-500/20',
        'Submitted': 'bg-amber-100 text-amber-700 border-amber-200 ring-amber-500/20',
        'Rejected': 'bg-red-100 text-red-700 border-red-200 ring-red-500/20'
    };
</script>

<tr class="group hover:bg-slate-50/80 transition-all duration-200 cursor-default">
    <!-- Employee Column -->
    <td class="px-6 py-4 align-top">
        <div class="flex items-center gap-3">
            <div class="relative flex-none">
                <div class="h-10 w-10 rounded-full bg-gradient-to-br from-blue-500 to-indigo-600 text-white flex items-center justify-center text-xs font-bold shadow-md ring-2 ring-white">
                    {getInitials(record.employee.name)}
                </div>
                <div class="absolute -bottom-0.5 -right-0.5 h-3.5 w-3.5 bg-green-500 border-2 border-white rounded-full"></div>
            </div>
            <div class="flex flex-col min-w-0">
                <span class="font-bold text-slate-800 truncate group-hover:text-blue-700 transition-colors">{record.employee.name}</span>
                <span class="text-[10px] text-slate-500 font-mono tracking-tight bg-slate-100 px-1.5 py-0.5 rounded-md w-fit mt-0.5 border border-slate-200/50">
                    {record.spd}
                </span>
            </div>
        </div>
    </td>

    <!-- Destination Column -->
    <td class="px-6 py-4 align-top">
        <div class="flex flex-col space-y-1">
            <div class="flex items-center gap-1.5 font-medium text-slate-700">
                <svg xmlns="http://www.w3.org/2000/svg" class="h-3.5 w-3.5 text-slate-400" viewBox="0 0 20 20" fill="currentColor">
                    <path fill-rule="evenodd" d="M5.05 4.05a7 7 0 119.9 9.9L10 18.9l-4.95-4.95a7 7 0 010-9.9zM10 11a2 2 0 100-4 2 2 0 000 4z" clip-rule="evenodd" />
                </svg>
                {record.location}
            </div>
            <span class="text-xs text-slate-500 pl-5">{record.province}</span>
        </div>
    </td>

    <!-- Date Column -->
    <td class="px-6 py-4 align-top">
        <div class="flex flex-col space-y-1">
            <span class="text-sm text-slate-700 font-medium">
                {new Date(record.startDate).toLocaleDateString('id-ID', { day: 'numeric', month: 'short' })}
            </span>
            <span class="text-xs text-slate-400">
                {new Date(record.startDate).getFullYear()}
            </span>
        </div>
    </td>

    <!-- Status Column -->
    <td class="px-6 py-4 align-top text-center">
        <span class="inline-flex items-center px-2.5 py-1 rounded-full text-[10px] font-bold uppercase tracking-wider border ring-1 ring-inset {statusColors[record.status] || 'bg-slate-100 text-slate-600'}">
            {#if record.status === 'Approved'}
                <svg class="mr-1 h-3 w-3" fill="currentColor" viewBox="0 0 20 20"><path fill-rule="evenodd" d="M16.707 5.293a1 1 0 010 1.414l-8 8a1 1 0 01-1.414 0l-4-4a1 1 0 011.414-1.414L8 12.586l7.293-7.293a1 1 0 011.414 0z" clip-rule="evenodd"/></svg>
            {:else if record.status === 'Submitted'}
                 <svg class="mr-1 h-3 w-3" fill="currentColor" viewBox="0 0 20 20"><path fill-rule="evenodd" d="M10 18a8 8 0 100-16 8 8 0 000 16zm1-12a1 1 0 10-2 0v4a1 1 0 00.293.707l2.828 2.829a1 1 0 101.415-1.415L11 9.586V6z" clip-rule="evenodd"/></svg>
            {/if}
            {record.status}
        </span>
    </td>

    <!-- Cost Column -->
    <td class="px-6 py-4 align-top text-right">
        <div class="flex flex-col items-end gap-0.5">
            <span class="font-bold text-slate-700 font-feature-settings-tnum">
                {formatIDR(record.totalCost)}
            </span>
            {#if record.status === 'Approved'}
                 <span class="text-[10px] text-emerald-600 font-medium bg-emerald-50 px-1 rounded">Realisasi</span>
            {:else}
                 <span class="text-[10px] text-slate-400 font-medium">Estimasi</span>
            {/if}
        </div>
    </td>
</tr>

<style>
    .font-feature-settings-tnum {
        font-feature-settings: "tnum";
        font-variant-numeric: tabular-nums;
    }
</style>
