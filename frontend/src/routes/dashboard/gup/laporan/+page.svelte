<script>
    import { formatCurrency } from '$lib/shared/utils/utils';
    
    export let data;
    
    $: reports = data?.reports || [];

    function formatDate(dateStr) {
        if (!dateStr) return '-';
        return new Date(dateStr).toLocaleDateString('id-ID', {
            day: 'numeric', month: 'short', year: 'numeric'
        });
    }
</script>

<div class="space-y-6 max-w-7xl mx-auto pb-20">
    <!-- Header Section -->
    <div class="flex flex-col md:flex-row md:items-center justify-between gap-4 p-6 rounded-3xl bg-slate-100 border-2 border-white"
         style="box-shadow: inset 4px 4px 10px rgba(0,0,0,0.05), inset -4px -4px 10px rgba(255,255,255,0.8);">
        <div>
            <h1 class="text-3xl font-extrabold text-slate-800 tracking-tight">Laporan & Rekapitulasi GUP</h1>
            <p class="text-sm text-slate-500 mt-2 font-medium">Rekapitulasi transaksi GUP untuk keperluan laporan.</p>
        </div>
        <div class="flex items-center gap-3">
            <button 
                class="w-full sm:w-auto flex items-center justify-center gap-2 px-6 py-3 rounded-2xl font-bold text-slate-600 bg-slate-100 border-2 border-white"
                style="box-shadow: inset 2px 2px 5px rgba(0,0,0,0.05), inset -3px -3px 7px rgba(255,255,255,0.9);"
            >
                <svg xmlns="http://www.w3.org/2000/svg" class="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2.5" d="M12 10v6m0 0l-3-3m3 3l3-3m2 8H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z" />
                </svg>
                Export Excel
            </button>
        </div>
    </div>

    <!-- Data Table Section -->
    <div class="rounded-3xl bg-slate-100 border-2 border-white p-2"
         style="box-shadow: inset 4px 4px 10px rgba(0,0,0,0.05), inset -4px -4px 10px rgba(255,255,255,0.8);">
        
        <div class="overflow-x-auto w-full relative min-h-[400px] bg-slate-50 rounded-2xl p-1"
             style="box-shadow: inset 2px 2px 6px rgba(0,0,0,0.04), inset -2px -2px 6px rgba(255,255,255,1);">
            <table class="w-full text-sm text-left border-collapse">
                <thead class="text-slate-600 bg-slate-100 rounded-t-xl">
                    <tr>
                        <th class="min-w-[150px] font-bold py-4 px-5 rounded-tl-xl">ID Transaksi / Uraian</th>
                        <th class="min-w-[150px] font-bold py-4 px-5">Jenis Pengadaan</th>
                        <th class="min-w-[120px] font-bold py-4 px-5 text-right">Nilai (Rp)</th>
                        <th class="min-w-[120px] font-bold py-4 px-5 text-right">Dibayar (Rp)</th>
                        <th class="min-w-[120px] font-bold py-4 px-5 text-right">Pajak (Rp)</th>
                        <th class="min-w-[130px] font-bold py-4 px-5 rounded-tr-xl">Penerima / Tanggal</th>
                    </tr>
                </thead>
                <tbody class="divide-y divide-slate-200/50">
                    {#if reports.length === 0}
                        <tr>
                            <td colspan="6" class="p-12 text-center text-slate-500 font-medium">
                                <svg xmlns="http://www.w3.org/2000/svg" class="h-12 w-12 mx-auto text-slate-300 mb-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z" />
                                </svg>
                                Belum ada rekapitulasi laporan.
                            </td>
                        </tr>
                    {:else}
                        {#each reports as trx (trx.id)}
                            <tr class="hover:bg-slate-100/50 transition-colors">
                                <td class="py-4 px-5 align-middle">
                                    <div class="font-mono text-xs font-bold text-indigo-600 bg-indigo-50 px-2 py-1 rounded inline-block mb-1">
                                        {trx.businessId}
                                    </div>
                                    <div class="font-semibold text-slate-800 line-clamp-2">
                                        {trx.paymentDescription}
                                    </div>
                                </td>
                                <td class="py-4 px-5 align-middle font-medium text-slate-700">
                                    {trx.procurementTypeName || '-'}
                                </td>
                                <td class="py-4 px-5 align-middle text-right font-bold text-slate-800">
                                    {formatCurrency(trx.valueAmount)}
                                </td>
                                <td class="py-4 px-5 align-middle text-right font-bold text-emerald-600">
                                    {formatCurrency(trx.paidAmount)}
                                </td>
                                <td class="py-4 px-5 align-middle text-right font-bold text-rose-500">
                                    {formatCurrency(trx.taxAmount)}
                                </td>
                                <td class="py-4 px-5 align-middle">
                                    <div class="font-semibold text-slate-700">{trx.recipient || '-'}</div>
                                    <div class="text-xs text-slate-500 mt-0.5">{formatDate(trx.receiptDate)}</div>
                                </td>
                            </tr>
                        {/each}
                    {/if}
                </tbody>
            </table>
        </div>
    </div>
</div>
