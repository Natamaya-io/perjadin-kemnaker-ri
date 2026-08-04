<script lang="ts">
    export let data: any;
    $: masterData = data?.masterData || { procurementTypes: [], fundingSources: [] };
</script>

<div class="space-y-6 pb-20 max-w-7xl mx-auto">
    <!-- Header -->
    <div class="flex flex-col md:flex-row md:items-center justify-between gap-4 bg-white p-6 rounded-2xl shadow-sm border border-slate-100">
        <div>
            <h1 class="text-2xl font-bold text-slate-800 tracking-tight">Rincian Sumber Dana</h1>
            <p class="text-sm text-slate-500 mt-1">Lihat alokasi sumber dana GUP dan riwayat LS.</p>
        </div>
    </div>

    <!-- Content -->
    <section class="no-print">
        <div class="rounded-xl border border-slate-200 shadow-sm bg-white overflow-hidden relative flex flex-col">
            <div class="px-6 py-4 border-b border-slate-100 bg-white">
                <h2 class="text-sm font-bold text-slate-800 uppercase tracking-widest">Rincian Sumber Dana Bulanan</h2>
                <p class="text-xs text-slate-500 mt-0.5">{masterData.fundingSources.length} bulan terdaftar.</p>
            </div>
            
            <div class="overflow-x-auto w-full relative min-h-[400px]">
                <table class="w-full text-sm text-left relative border-collapse">
                    <thead class="bg-slate-50 sticky top-0 z-20 shadow-sm border-b border-slate-200">
                        <tr>
                            <th class="font-semibold text-slate-700 px-6 py-3 bg-slate-50 w-24">Bulan</th>
                            <th class="font-semibold text-slate-700 px-6 py-3 bg-slate-50">Label GUP / Sumber Dana</th>
                            <th class="font-semibold text-slate-700 px-6 py-3 bg-slate-50">Status LS</th>
                            <th class="font-semibold text-slate-700 px-6 py-3 bg-slate-50 text-right w-32">Aksi</th>
                        </tr>
                    </thead>
                    <tbody>
                        {#if masterData.fundingSources.length === 0}
                            <tr>
                                <td colspan="4" class="p-12 text-center text-slate-500">
                                    Belum ada rincian sumber dana yang tersedia.
                                </td>
                            </tr>
                        {:else}
                            {#each masterData.fundingSources as src}
                                <tr class="hover:bg-slate-50/50 border-b border-slate-100 transition-colors bg-white">
                                    <td class="px-6 py-4 align-middle">
                                        <div class="font-medium text-slate-800">{src.monthName}</div>
                                    </td>
                                    <td class="px-6 py-4 align-middle">
                                        <div class="font-bold text-slate-900">{src.gupLabel}</div>
                                    </td>
                                    <td class="px-6 py-4 align-middle">
                                        <div class="flex items-center gap-2">
                                            <span class="inline-flex items-center rounded-full px-2 py-0.5 text-[9px] font-bold tracking-wide bg-sky-50 text-sky-700 border border-sky-100 uppercase">
                                                LS
                                            </span>
                                            <span class="text-xs font-semibold text-slate-400">
                                                Data LS belum dihubungkan
                                            </span>
                                        </div>
                                    </td>
                                    <td class="px-6 py-4 align-middle text-right">
                                        <button class="text-blue-600 hover:text-blue-800 bg-blue-50 hover:bg-blue-100 p-1.5 rounded transition-colors" title="Lihat Rincian">
                                            <svg xmlns="http://www.w3.org/2000/svg" class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                                                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 12a3 3 0 11-6 0 3 3 0 016 0z" />
                                                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M2.458 12C3.732 7.943 7.523 5 12 5c4.478 0 8.268 2.943 9.542 7-1.274 4.057-5.064 7-9.542 7-4.477 0-8.268-2.943-9.542-7z" />
                                            </svg>
                                        </button>
                                    </td>
                                </tr>
                            {/each}
                        {/if}
                    </tbody>
                </table>
            </div>
        </div>
    </section>
</div>
