<script>
    import { formatCurrency } from '$lib/shared/utils/utils';
    import Button from '$lib/shared/ui/button/Button.svelte';
    
    export let data;
    
    $: budgets = data?.budgets || [];
</script>

<div class="space-y-6 pb-20 w-full">
    <!-- Header Section -->
    <div class="flex flex-col md:flex-row md:items-center justify-between gap-4 bg-white p-6 rounded-2xl shadow-sm border border-slate-100">
        <div>
            <h1 class="text-2xl font-bold text-slate-800 tracking-tight">Data Pagu Anggaran</h1>
            <p class="text-sm text-slate-500 mt-1">Manajemen pagu anggaran tahunan berdasarkan jenis pengadaan GUP.</p>
        </div>
        <div class="flex items-center gap-3">
            <Button variant="default" class="w-full sm:w-auto gap-2">
                <svg xmlns="http://www.w3.org/2000/svg" class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4v16m8-8H4" />
                </svg>
                Tambah Anggaran
            </Button>
        </div>
    </div>

    <!-- Data Table Section -->
    <div class="rounded-xl border border-slate-200 shadow-sm bg-white overflow-hidden relative flex flex-col">
        <div class="overflow-x-auto w-full relative min-h-[400px]">
            <table class="w-full text-sm text-left relative border-collapse">
                <thead class="bg-slate-50 sticky top-0 z-20 shadow-sm border-b border-slate-200">
                    <tr>
                        <th class="min-w-[150px] font-semibold text-slate-700 pl-4 py-3 bg-slate-50">Tahun</th>
                        <th class="min-w-[150px] font-semibold text-slate-700 py-3 bg-slate-50">Jenis Pengadaan</th>
                        <th class="min-w-[150px] font-semibold text-slate-700 pr-4 py-3 bg-slate-50 text-right">Pagu Anggaran (Rp)</th>
                    </tr>
                </thead>
                <tbody>
                    {#if budgets.length === 0}
                        <tr>
                            <td colspan="3" class="p-12 text-center text-slate-500 w-full">
                                <svg xmlns="http://www.w3.org/2000/svg" class="h-12 w-12 mx-auto text-slate-300 mb-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z" />
                                </svg>
                                Belum ada data pagu anggaran.
                            </td>
                        </tr>
                    {:else}
                        {#each budgets as budget (budget.id)}
                            <tr class="hover:bg-slate-50/50 border-b border-slate-100 transition-colors bg-white">
                                <td class="pl-4 py-4 align-middle">
                                    <div class="font-medium text-slate-800 text-sm">{budget.year}</div>
                                </td>
                                <td class="py-4 align-middle font-medium text-slate-700">
                                    {budget.procurementTypeName || '-'}
                                </td>
                                <td class="pr-4 py-4 align-middle text-right font-medium text-slate-800">
                                    {formatCurrency(budget.amount)}
                                </td>
                            </tr>
                        {/each}
                    {/if}
                </tbody>
            </table>
        </div>
    </div>
</div>
