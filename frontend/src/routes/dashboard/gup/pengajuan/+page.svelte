<script lang="ts">
    import { formatCurrency } from '$lib/shared/utils/utils';
    import { goto } from '$app/navigation';
    import { api } from '$lib/shared/api';
    import Button from '$lib/shared/ui/button/Button.svelte';
    import Select from '$lib/shared/ui/select/Select.svelte';
    
    export let data: any;
    
    $: transactions = data?.transactions || [];
    $: masterData = data?.masterData || { procurementTypes: [], fundingSources: [] };

    let searchQuery = '';
    let filterBulan = '';
    let filterJenis = '';

    $: filteredTransactions = transactions.filter((trx: any) => {
        let match = true;
        
        if (searchQuery) {
            const q = searchQuery.toLowerCase();
            match = match && (
                (trx.paymentDescription && trx.paymentDescription.toLowerCase().includes(q)) ||
                (trx.recipient && trx.recipient.toLowerCase().includes(q)) ||
                (trx.pum && trx.pum.toLowerCase().includes(q)) ||
                (trx.businessId && trx.businessId.toLowerCase().includes(q))
            );
        }

        if (filterBulan) {
            // trx.receiptDate is ISO string like "2024-08-02T00:00:00Z"
            if (trx.receiptDate) {
                const month = trx.receiptDate.split('-')[1];
                match = match && month === filterBulan;
            } else {
                match = false;
            }
        }

        if (filterJenis) {
            match = match && trx.procurementTypeId === filterJenis;
        }

        return match;
    });

    function formatDate(dateStr: string) {
        if (!dateStr) return '-';
        return new Date(dateStr).toLocaleDateString('id-ID', {
            day: '2-digit', month: 'short', year: 'numeric'
        });
    }

    function getProcurementTypeName(id: string) {
        if (!id) return '-';
        const pt = masterData.procurementTypes.find((p: any) => p.id === id);
        return pt ? pt.name : '-';
    }

    function resetFilters() {
        searchQuery = '';
        filterBulan = '';
        filterJenis = '';
    }

    // --- Modal Detail Logic ---
    let isReviewOpen = false;
    let selectedGup: any = null;
    let isFetchingDetail = false;

    async function handleReview(id: string) {
        selectedGup = null;
        isFetchingDetail = true;
        isReviewOpen = true;
        try {
            const res = await api.getGupPengajuanById(id);
            selectedGup = res;
        } catch (error) {
            console.error("Gagal mengambil detail GUP:", error);
        } finally {
            isFetchingDetail = false;
        }
    }
</script>

<div class="space-y-6 pb-20 max-w-7xl mx-auto">
    <!-- Header halaman dan tombol aksi -->
    <div class="flex flex-col md:flex-row md:items-center justify-between gap-4 bg-white p-6 rounded-2xl shadow-sm border border-slate-100">
        <div>
            <h1 class="text-2xl font-bold text-slate-800 tracking-tight">Pengajuan GUP</h1>
            <p class="text-sm text-slate-500 mt-1">Tabel menampilkan ringkasan pengajuan. Gunakan tombol View untuk rincian.</p>
        </div>
        <div class="flex items-center gap-3 no-print">
            <Button variant="warning" class="gap-2">
                <svg xmlns="http://www.w3.org/2000/svg" class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M17 17h2a2 2 0 002-2v-4a2 2 0 00-2-2H5a2 2 0 00-2 2v4a2 2 0 002 2h2m2 4h6a2 2 0 002-2v-4a2 2 0 00-2-2H9a2 2 0 00-2 2v4a2 2 0 002 2zm8-12V5a2 2 0 00-2-2H9a2 2 0 00-2 2v4h10z" />
                </svg>
                Cetak Rekapitulasi
            </Button>
            <Button variant="default" class="gap-2" on:click={() => goto('/dashboard/gup/pengajuan/new')}>
                <svg xmlns="http://www.w3.org/2000/svg" class="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4v16m8-8H4" />
                </svg>
                Tambah GUP
            </Button>
        </div>
    </div>

    <!-- Filter Section -->
    <div class="bg-white p-4 rounded-xl shadow-sm border border-slate-100 no-print">
        <div class="grid grid-cols-1 md:grid-cols-[1fr_auto_auto_auto] gap-3 items-end w-full">
            <!-- Search Bar -->
            <div class="min-w-0">
                <label for="searchQuery" class="mb-1.5 block text-xs font-bold uppercase tracking-widest text-slate-500">Pencarian</label>
                <div class="relative">
                    <svg class="absolute left-3 top-1/2 -translate-y-1/2 h-4 w-4 text-slate-400 pointer-events-none" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z" />
                    </svg>
                    <input id="searchQuery" type="search" bind:value={searchQuery} placeholder="Cari pembayaran..." class="w-full rounded-xl border border-slate-200 bg-slate-50 py-2.5 pl-9 pr-4 text-sm text-slate-700 focus:outline-none focus:ring-2 focus:ring-indigo-500 focus:bg-white transition-colors md:text-[0.9375rem]">
                </div>
            </div>
            <!-- Jenis Pengadaan -->
            <div class="w-full md:w-48">
                <label for="filterJenisPengadaan" class="mb-1.5 block text-xs font-bold uppercase tracking-widest text-slate-500">Jenis Pengadaan</label>
                <Select id="filterJenisPengadaan" bind:value={filterJenis} class="border-slate-200 w-full" options={[
                    {value: '', label: 'Semua Jenis'},
                    ...masterData.procurementTypes.map(t => ({value: t.id, label: t.name}))
                ]} />
            </div>
            <!-- Bulan -->
            <div class="w-full md:w-40">
                <label for="filterBulan" class="mb-1.5 block text-xs font-bold uppercase tracking-widest text-slate-500">Bulan Transaksi</label>
                <Select id="filterBulan" bind:value={filterBulan} class="border-slate-200 w-full" options={[
                    {value: '', label: 'Semua Bulan'},
                    {value: '01', label: 'Januari'},
                    {value: '02', label: 'Februari'},
                    {value: '03', label: 'Maret'},
                    {value: '04', label: 'April'},
                    {value: '05', label: 'Mei'},
                    {value: '06', label: 'Juni'},
                    {value: '07', label: 'Juli'},
                    {value: '08', label: 'Agustus'},
                    {value: '09', label: 'September'},
                    {value: '10', label: 'Oktober'},
                    {value: '11', label: 'November'},
                    {value: '12', label: 'Desember'}
                ]} />
            </div>
            <!-- Reset -->
            <div class="flex items-end">
                <button type="button" on:click={resetFilters} class="inline-flex h-full min-h-[42px] items-center justify-center gap-2 rounded-xl border border-slate-200 bg-white px-4 py-2.5 text-sm font-semibold text-slate-600 transition hover:bg-slate-50 hover:text-red-600 hover:border-red-100 shadow-sm whitespace-nowrap w-full md:w-auto">
                    <svg class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15" /></svg>
                    Reset
                </button>
            </div>
        </div>
    </div>


    <!-- Tabel Pengajuan GUP -->
    <div class="rounded-xl border border-slate-200 shadow-sm bg-white overflow-hidden relative flex flex-col">
        <div class="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-3 px-6 py-4 border-b border-slate-100 bg-white">
            <div>
                <h2 class="text-sm font-bold text-slate-800 uppercase tracking-widest">Daftar Pengajuan GUP</h2>
                <p class="text-xs text-slate-500 mt-0.5">{filteredTransactions.length} data ditemukan.</p>
            </div>
        </div>

        <div class="overflow-x-auto w-full relative min-h-[400px]">
            <table class="w-full text-sm text-left relative border-collapse">
                <thead class="bg-slate-50 sticky top-0 z-20 shadow-sm border-b border-slate-200">
                    <tr>
                        <th class="font-semibold text-slate-700 px-4 py-3 bg-slate-50 w-12 text-center">No</th>
                        <th class="font-semibold text-slate-700 px-4 py-3 bg-slate-50 min-w-[280px]">Pembayaran</th>
                        <th class="font-semibold text-slate-700 px-4 py-3 bg-slate-50 text-right min-w-[130px]">Nilai</th>
                        <th class="font-semibold text-slate-700 px-4 py-3 bg-slate-50 text-right min-w-[130px]">Dibayarkan</th>
                        <th class="font-semibold text-slate-700 px-4 py-3 bg-slate-50 text-right min-w-[120px]">Pajak</th>
                        <th class="font-semibold text-slate-700 px-4 py-3 bg-slate-50 text-right min-w-[130px]">Selisih</th>
                        <th class="font-semibold text-slate-700 px-4 py-3 bg-slate-50 min-w-[160px]">Kwitansi</th>
                        <th class="font-semibold text-slate-700 px-4 py-3 bg-slate-50 text-center w-[140px] no-print">Aksi</th>
                    </tr>
                </thead>
                <tbody>
                    {#if filteredTransactions.length === 0}
                        <tr>
                            <td colspan="8" class="p-12 text-center text-slate-500 w-full">
                                <svg class="h-12 w-12 mx-auto text-slate-300 mb-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9.172 16.172a4 4 0 015.656 0M9 10h.01M15 10h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
                                </svg>
                                Belum ada pengajuan GUP yang sesuai.
                            </td>
                        </tr>
                    {:else}
                        {#each filteredTransactions as trx, index (trx.id)}
                            <tr class="hover:bg-slate-50/50 border-b border-slate-100 transition-colors bg-white">
                                <td class="px-4 py-4 align-middle text-center font-medium text-slate-500">{index + 1}</td>
                                <td class="px-4 py-4 align-middle">
                                    <span class="inline-flex items-center font-mono text-[13px] font-bold tracking-widest text-slate-700 mb-1">{trx.businessId}</span>
                                    <div class="font-medium text-slate-800 text-sm line-clamp-2">{trx.paymentDescription}</div>
                                    <span class="inline-flex items-center rounded-full px-2 py-0.5 text-[9px] font-bold capitalize tracking-wide bg-indigo-50 text-indigo-700 border border-indigo-200 mt-1.5">
                                        {getProcurementTypeName(trx.procurementTypeId)}
                                    </span>
                                </td>
                                <td class="px-4 py-4 align-middle text-right font-medium text-slate-800">{formatCurrency(trx.valueAmount || 0)}</td>
                                <td class="px-4 py-4 align-middle text-right font-semibold text-emerald-600">{formatCurrency(trx.paidAmount || 0)}</td>
                                <td class="px-4 py-4 align-middle text-right text-rose-500">{formatCurrency(trx.taxAmount || 0)}</td>
                                <td class="px-4 py-4 align-middle text-right font-medium text-slate-800">{formatCurrency((trx.valueAmount || 0) - (trx.paidAmount || 0) - (trx.taxAmount || 0))}</td>
                                <td class="px-4 py-4 align-middle text-xs text-slate-600">
                                    <div class="font-medium text-slate-800">{formatDate(trx.receiptDate)}</div>
                                    <div class="text-[10px] text-slate-500 mt-0.5 max-w-[140px] truncate" title={trx.recipient}>Penerima: {trx.recipient || '-'}</div>
                                    {#if trx.pum}
                                        <div class="text-[10px] text-slate-400 max-w-[140px] truncate" title={trx.pum}>PUM: {trx.pum}</div>
                                    {/if}
                                </td>
                                <td class="px-4 py-4 align-middle text-center no-print">
                                    <div class="flex items-center justify-center gap-1.5">
                                        <button class="text-blue-600 hover:text-blue-800 bg-blue-50 hover:bg-blue-100 p-1.5 rounded transition-colors" title="View Detail" on:click={() => handleReview(trx.id)}>
                                            <svg class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                                                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 12a3 3 0 11-6 0 3 3 0 016 0z" />
                                                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M2.458 12C3.732 7.943 7.523 5 12 5c4.478 0 8.268 2.943 9.542 7-1.274 4.057-5.064 7-9.542 7-4.477 0-8.268-2.943-9.542-7z" />
                                            </svg>
                                        </button>
                                        <button class="text-amber-600 hover:text-amber-800 bg-amber-50 hover:bg-amber-100 p-1.5 rounded transition-colors" title="Cetak Kwitansi">
                                            <svg class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                                                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z" />
                                            </svg>
                                        </button>
                                        <button class="text-emerald-600 hover:text-emerald-800 bg-emerald-50 hover:bg-emerald-100 p-1.5 rounded transition-colors" title="Cetak SPBY">
                                            <svg class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                                                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 5H7a2 2 0 00-2 2v12a2 2 0 002 2h10a2 2 0 002-2V7a2 2 0 00-2-2h-2M9 5a2 2 0 002 2h2a2 2 0 002-2M9 5a2 2 0 012-2h2a2 2 0 012 2" />
                                            </svg>
                                        </button>
                                    </div>
                                </td>
                            </tr>
                        {/each}
                    {/if}
                </tbody>
            </table>
        </div>
    </div>
</div>

<!-- Modal Detail GUP -->
{#if isReviewOpen}
<div class="fixed inset-0 z-50 flex items-center justify-center p-4 sm:p-6 bg-slate-900/40 backdrop-blur-sm transition-opacity" on:click|self={() => isReviewOpen = false} on:keydown|self={(e) => e.key === 'Escape' && (isReviewOpen = false)} role="dialog" aria-modal="true" tabindex="-1">
    <div class="bg-slate-50 w-full max-w-4xl max-h-[90vh] rounded-2xl shadow-2xl overflow-hidden flex flex-col relative animate-in fade-in zoom-in-95 duration-200">
        
        <!-- Header Modal -->
        <div class="flex items-center justify-between px-6 py-5 bg-white border-b border-slate-200 relative">
            <div class="absolute top-0 left-0 w-full h-1 bg-gradient-to-r from-indigo-500 to-sky-400"></div>
            <div>
                <h3 class="text-xl font-bold text-slate-800 tracking-tight">Detail Pengajuan GUP</h3>
                <p class="text-xs text-slate-500 mt-1 uppercase tracking-widest font-semibold">{selectedGup?.businessId || 'Memuat...'}</p>
            </div>
            <button class="p-2 -mr-2 text-slate-400 hover:text-rose-500 hover:bg-rose-50 rounded-full transition-colors" on:click={() => isReviewOpen = false} title="Tutup" aria-label="Tutup">
                <svg xmlns="http://www.w3.org/2000/svg" class="h-6 w-6" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
                </svg>
            </button>
        </div>

        <!-- Body Modal -->
        <div class="p-6 overflow-y-auto flex-1 space-y-6">
            {#if isFetchingDetail}
                <div class="flex flex-col items-center justify-center h-48 space-y-4">
                    <div class="w-10 h-10 border-4 border-indigo-200 border-t-indigo-600 rounded-full animate-spin"></div>
                    <p class="text-sm font-medium text-slate-500 animate-pulse">Mengambil data...</p>
                </div>
            {:else if selectedGup}
                <div class="bg-white p-5 md:p-6 rounded-2xl border border-slate-200 hover:border-blue-300 shadow-sm hover:shadow-lg transition-all duration-300 relative overflow-hidden group">
                    <div class="absolute top-0 left-0 w-full h-1.5 bg-gradient-to-r from-blue-500 to-cyan-400"></div>
                    <div class="flex items-center gap-3 mb-4 pb-3 border-b border-slate-100">
                        <div class="p-2.5 bg-gradient-to-br from-blue-50 to-blue-100/50 text-blue-600 rounded-xl shadow-sm border border-blue-100">
                            <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13 16h-1v-4h-1m1-4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z"></path></svg>
                        </div>
                        <h4 class="text-xs font-bold text-slate-700 uppercase tracking-widest">Informasi Dasar</h4>
                    </div>
                    <div class="grid grid-cols-1 md:grid-cols-2 gap-6">
                        <div>
                            <p class="text-xs font-semibold text-slate-400 uppercase mb-1">ID Transaksi / SPBY</p>
                            <p class="font-bold text-slate-800 font-mono text-sm">{selectedGup.businessId || '-'}</p>
                        </div>
                        <div>
                            <p class="text-xs font-semibold text-slate-400 uppercase mb-1">Tanggal Kuitansi</p>
                            <p class="font-bold text-slate-800 text-sm">{formatDate(selectedGup.receiptDate)}</p>
                        </div>
                        <div class="md:col-span-2">
                            <p class="text-xs font-semibold text-slate-400 uppercase mb-1">Uraian Pembayaran</p>
                            <p class="font-medium text-slate-700 bg-slate-50 p-3 rounded-lg border border-slate-100 text-sm leading-relaxed">{selectedGup.paymentDescription || '-'}</p>
                        </div>
                        <div>
                            <p class="text-xs font-semibold text-slate-400 uppercase mb-1">Penerima Uang (PUM)</p>
                            <p class="font-bold text-slate-800 text-sm">{selectedGup.pum || '-'}</p>
                        </div>
                        <div>
                            <p class="text-xs font-semibold text-slate-400 uppercase mb-1">Penerima Barang/Jasa</p>
                            <p class="font-bold text-slate-800 text-sm">{selectedGup.recipient || '-'}</p>
                        </div>
                    </div>
                </div>

                <div class="bg-white p-5 md:p-6 rounded-2xl border border-slate-200 hover:border-emerald-300 shadow-sm hover:shadow-lg transition-all duration-300 relative overflow-hidden group">
                    <div class="absolute top-0 left-0 w-full h-1.5 bg-gradient-to-r from-emerald-500 to-teal-400"></div>
                    <div class="flex items-center gap-3 mb-4 pb-3 border-b border-slate-100">
                        <div class="p-2.5 bg-gradient-to-br from-emerald-50 to-emerald-100/50 text-emerald-600 rounded-xl shadow-sm border border-emerald-100">
                            <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 8c-1.657 0-3 .895-3 2s1.343 2 3 2 3 .895 3 2-1.343 2-3 2m0-8c1.11 0 2.08.402 2.599 1M12 8V7m0 1v8m0 0v1m0-1c-1.11 0-2.08-.402-2.599-1M21 12a9 9 0 11-18 0 9 9 0 0118 0z"></path></svg>
                        </div>
                        <h4 class="text-xs font-bold text-slate-700 uppercase tracking-widest">Rincian Nilai Transaksi</h4>
                    </div>
                    <div class="grid grid-cols-1 sm:grid-cols-3 gap-4">
                        <div class="bg-slate-50 rounded-xl p-4 border border-slate-100 flex flex-col justify-center">
                            <p class="text-xs font-semibold text-slate-500 mb-1">Nilai Kuitansi</p>
                            <p class="text-xl font-bold text-slate-900">{formatCurrency(selectedGup.valueAmount)}</p>
                        </div>
                        <div class="bg-emerald-50 rounded-xl p-4 border border-emerald-100 flex flex-col justify-center">
                            <p class="text-xs font-semibold text-emerald-700 mb-1">Yang Dibayarkan (Netto)</p>
                            <p class="text-xl font-bold text-emerald-600">{formatCurrency(selectedGup.paidAmount)}</p>
                        </div>
                        <div class="bg-rose-50 rounded-xl p-4 border border-rose-100 flex flex-col justify-center">
                            <p class="text-xs font-semibold text-rose-700 mb-1">Total Pajak</p>
                            <p class="text-xl font-bold text-rose-600">{formatCurrency(selectedGup.taxAmount)}</p>
                        </div>
                    </div>
                </div>

                <div class="bg-white p-5 md:p-6 rounded-2xl border border-slate-200 hover:border-amber-300 shadow-sm hover:shadow-lg transition-all duration-300 relative overflow-hidden group">
                    <div class="absolute top-0 left-0 w-full h-1.5 bg-gradient-to-r from-amber-500 to-orange-400"></div>
                    <div class="flex items-center gap-3 mb-4 pb-3 border-b border-slate-100">
                        <div class="p-2.5 bg-gradient-to-br from-amber-50 to-amber-100/50 text-amber-600 rounded-xl shadow-sm border border-amber-100">
                            <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 7h6m0 10v-3m-3 3h.01M9 17h.01M9 14h.01M12 14h.01M15 11h.01M12 11h.01M9 11h.01M7 21h10a2 2 0 002-2V5a2 2 0 00-2-2H7a2 2 0 00-2 2v14a2 2 0 002 2z"></path></svg>
                        </div>
                        <h4 class="text-xs font-bold text-slate-700 uppercase tracking-widest">Detail Klasifikasi (MAK)</h4>
                    </div>
                    <div class="space-y-4">
                        <div class="flex flex-col sm:flex-row sm:justify-between sm:items-center py-2 border-b border-slate-50 gap-2">
                            <span class="text-sm font-semibold text-slate-500">Program / Kegiatan / KRO / RO / Komponen / Sub</span>
                            <span class="text-sm font-bold text-slate-800 sm:text-right max-w-md">{selectedGup.mak || '-'}</span>
                        </div>
                        <div class="flex flex-col sm:flex-row sm:justify-between sm:items-center py-2 border-b border-slate-50 gap-2">
                            <span class="text-sm font-semibold text-slate-500">Kode Akun</span>
                            <span class="text-sm font-bold text-slate-800 font-mono bg-slate-100 px-2 py-1 rounded inline-block">{selectedGup.accountCode || '-'}</span>
                        </div>
                    </div>
                </div>

            {:else}
                <div class="text-center py-12">
                    <div class="inline-flex items-center justify-center w-16 h-16 rounded-full bg-slate-100 text-slate-400 mb-4">
                        <svg class="h-8 w-8" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9.172 16.172a4 4 0 015.656 0M9 10h.01M15 10h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
                        </svg>
                    </div>
                    <p class="text-slate-500 font-medium">Gagal memuat data detail pengajuan GUP.</p>
                </div>
            {/if}
        </div>
        
        <!-- Footer Modal -->
        <div class="px-6 py-4 bg-slate-50 border-t border-slate-200 flex justify-end gap-3 rounded-b-2xl">
            <Button variant="warning" class="w-full sm:w-auto" on:click={() => isReviewOpen = false}>Tutup</Button>
            <Button variant="default" class="gap-2" disabled={!selectedGup}>
                <svg xmlns="http://www.w3.org/2000/svg" class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M17 17h2a2 2 0 002-2v-4a2 2 0 00-2-2H5a2 2 0 00-2 2v4a2 2 0 002 2h2m2 4h6a2 2 0 002-2v-4a2 2 0 00-2-2H9a2 2 0 00-2 2v4a2 2 0 002 2zm8-12V5a2 2 0 00-2-2H9a2 2 0 00-2 2v4h10z" />
                </svg>
                Cetak SPBY
            </Button>
        </div>
    </div>
</div>
{/if}
