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

<div class="space-y-6 pb-20 max-w-7xl mx-auto">
    <!-- Header Section -->
    <div class="flex flex-col md:flex-row md:items-center justify-between gap-4 bg-white p-6 rounded-2xl shadow-sm border border-slate-100">
        <div>
            <h1 class="text-2xl font-bold text-slate-800 tracking-tight">Laporan & Rekapitulasi GUP</h1>
            <p class="text-sm text-slate-500 mt-1">Rekapitulasi transaksi GUP untuk keperluan laporan.</p>
        </div>
        <div class="flex items-center gap-3">
            <button class="inline-flex items-center justify-center gap-2 rounded-xl bg-emerald-500 hover:bg-emerald-400 px-6 py-2.5 text-sm font-bold text-white border-2 border-emerald-400 transition-all hover:-translate-y-0.5 hover:shadow-lg hover:shadow-emerald-500/30 active:scale-95 w-full sm:w-auto" style="box-shadow: inset 2px 2px 5px rgba(255,255,255,0.4), inset -3px -3px 7px rgba(0,0,0,0.15);">
                <svg xmlns="http://www.w3.org/2000/svg" class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 10v6m0 0l-3-3m3 3l3-3m2 8H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z" />
                </svg>
                Export Excel
            </button>
        </div>
    </div>

    <!-- Data Table Section -->
    <div class="rounded-xl border border-slate-200 shadow-sm bg-white overflow-hidden relative flex flex-col">
        <div class="overflow-x-auto w-full relative min-h-[400px]">
            <table class="w-full text-sm text-left relative border-collapse">
                <thead class="bg-slate-50 sticky top-0 z-20 shadow-sm border-b border-slate-200">
                    <tr>
                        <th class="min-w-[150px] font-semibold text-slate-700 pl-4 py-3 bg-slate-50">ID Transaksi / Uraian</th>
                        <th class="min-w-[150px] font-semibold text-slate-700 py-3 bg-slate-50">Jenis Pengadaan</th>
                        <th class="min-w-[120px] font-semibold text-slate-700 py-3 bg-slate-50 text-right">Nilai (Rp)</th>
                        <th class="min-w-[120px] font-semibold text-slate-700 py-3 bg-slate-50 text-right">Dibayar (Rp)</th>
                        <th class="min-w-[120px] font-semibold text-slate-700 py-3 bg-slate-50 text-right">Pajak (Rp)</th>
                        <th class="min-w-[130px] font-semibold text-slate-700 pr-4 py-3 bg-slate-50">Penerima / Tanggal</th>
                    </tr>
                </thead>
                <tbody>
                    {#if reports.length === 0}
                        <tr>
                            <td colspan="6" class="p-12 text-center text-slate-500 w-full">
                                <svg xmlns="http://www.w3.org/2000/svg" class="h-12 w-12 mx-auto text-slate-300 mb-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z" />
                                </svg>
                                Belum ada rekapitulasi laporan.
                            </td>
                        </tr>
                    {:else}
                        {#each reports as trx (trx.id)}
                            <tr class="hover:bg-slate-50/50 border-b border-slate-100 transition-colors bg-white">
                                <td class="pl-4 py-4 align-middle">
                                    <span class="inline-flex items-center font-mono text-[13px] font-bold tracking-widest text-indigo-700 mb-1">{trx.businessId}</span>
                                    <div class="font-medium text-slate-800 text-sm line-clamp-2">{trx.paymentDescription}</div>
                                </td>
                                <td class="py-4 align-middle font-medium text-slate-700">
                                    {trx.procurementTypeName || '-'}
                                </td>
                                <td class="py-4 align-middle text-right font-medium text-slate-800">
                                    {formatCurrency(trx.valueAmount)}
                                </td>
                                <td class="py-4 align-middle text-right font-semibold text-emerald-600">
                                    {formatCurrency(trx.paidAmount)}
                                </td>
                                <td class="py-4 align-middle text-right text-rose-500">
                                    {formatCurrency(trx.taxAmount)}
                                </td>
                                <td class="pr-4 py-4 align-middle text-xs text-slate-600">
                                    <div class="font-medium">{trx.recipient || '-'}</div>
                                    <div class="text-[10px] text-slate-400 mt-0.5">{formatDate(trx.receiptDate)}</div>
                                </td>
                            </tr>
                        {/each}
                    {/if}
                </tbody>
            </table>
        </div>
    </div>
</div>
