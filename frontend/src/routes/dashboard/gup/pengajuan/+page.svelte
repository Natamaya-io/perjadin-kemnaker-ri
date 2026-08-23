<script lang="ts">
    import { formatCurrency } from '$lib/shared/utils/utils';
    import { goto, invalidateAll } from '$app/navigation';
    import { api } from '$lib/shared/api';
    import Button from '$lib/shared/ui/button/Button.svelte';
    import Select from '$lib/shared/ui/select/Select.svelte';
    import BaseModal from '$lib/shared/ui/base-modal/BaseModal.svelte';
    import KwitansiPrintModal from './KwitansiPrintModal.svelte';
    import SpbyPrintModal from './SpbyPrintModal.svelte';
    import RekapitulasiPrintModal from './RekapitulasiPrintModal.svelte';
    import { userStore } from '$lib/features/auth/store';
    import { toast } from '$lib/shared/stores/toast';
    import { ConfirmationModal } from '$lib/shared/ui/confirmation-modal';
    
    export let data: any;
    
    $: transactions = data?.transactions || [];
    $: masterData = data?.masterData || { procurementTypes: [], fundingSources: [] };

    let searchQuery = '';
    let filterTahun = new Date().getFullYear().toString();
    let filterBulan = '';
    let filterJenis = '';

    const currentYear = new Date().getFullYear();
    const startYear = 2026;
    const endYear = Math.max(currentYear + 1, startYear + 1);
    const yearOptions = [
        {value: '', label: 'Semua Tahun'},
        ...Array.from({ length: endYear - startYear + 1 }, (_, i) => ({
            value: (startYear + i).toString(),
            label: (startYear + i).toString()
        })).reverse()
    ];

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

        if (filterTahun) {
            if (trx.receiptDate) {
                const year = trx.receiptDate.split('-')[0];
                match = match && year === filterTahun;
            } else {
                match = false;
            }
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

    // Pagination state
    let currentPage = 1;
    const itemsPerPage = 10;
    $: totalPages = Math.ceil(filteredTransactions.length / itemsPerPage) || 1;
    $: if (currentPage > totalPages && totalPages > 0) currentPage = totalPages;
    $: paginatedTransactions = filteredTransactions.slice((currentPage - 1) * itemsPerPage, currentPage * itemsPerPage);

    // Reset ke halaman 1 jika filter berubah
    $: if (searchQuery || filterTahun || filterBulan || filterJenis) currentPage = 1;

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
        filterTahun = new Date().getFullYear().toString();
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

    // --- Print Modals Logic ---
    let isKwitansiOpen = false;
    let isSpbyOpen = false;
    let isRekapitulasiOpen = false;
    let printData: any = null;

    async function handlePrintKwitansi(id: string) {
        try {
            const res = await api.getGupPengajuanById(id);
            printData = { 
                ...res, 
                procurementType: masterData.procurementTypes.find(t => t.id === res.procurementTypeId) 
            };
            isKwitansiOpen = true;
        } catch (error) {
            console.error("Gagal mengambil data Kwitansi:", error);
            toast.send("Gagal memuat data untuk dicetak", 'error');
        }
    }

    async function handlePrintSpby(id: string) {
        try {
            const res = await api.getGupPengajuanById(id);
            printData = { 
                ...res, 
                procurementType: masterData.procurementTypes.find(t => t.id === res.procurementTypeId) 
            };
            isSpbyOpen = true;
        } catch (error) {
            console.error("Gagal mengambil data SPBY:", error);
            toast.send("Gagal memuat data untuk dicetak", 'error');
        }
    }

    let isDeleteModalOpen = false;
    let transactionToDelete: string | null = null;

    function handleDelete(id: string) {
        transactionToDelete = id;
        isDeleteModalOpen = true;
    }

    async function processDelete() {
        if (!transactionToDelete) return;
        try {
            await api.deleteGupPengajuan(transactionToDelete);
            toast.send("Pengajuan berhasil dihapus", 'success');
            await invalidateAll();
        } catch (error) {
            console.error("Gagal menghapus pengajuan:", error);
            toast.send("Gagal menghapus pengajuan", 'error');
        } finally {
            isDeleteModalOpen = false;
            transactionToDelete = null;
        }
    }
</script>

<div class="space-y-6 pb-20 w-full">
    <!-- Header halaman dan tombol aksi -->
    <div class="flex flex-col md:flex-row md:items-center justify-between gap-4 bg-white p-6 rounded-2xl shadow-sm border border-slate-100">
        <div>
            <h1 class="text-2xl font-bold text-slate-800 tracking-tight">Pengajuan GUP</h1>
            <p class="text-sm text-slate-500 mt-1">Tabel menampilkan ringkasan pengajuan. Gunakan tombol View untuk rincian.</p>
        </div>
        <div class="flex items-center gap-3 no-print">
            <Button variant="warning" class="gap-2" on:click={() => isRekapitulasiOpen = true}>
                <svg xmlns="http://www.w3.org/2000/svg" class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M17 17h2a2 2 0 002-2v-4a2 2 0 00-2-2H5a2 2 0 00-2 2v4a2 2 0 002 2h2m2 4h6a2 2 0 002-2v-4a2 2 0 00-2-2H9a2 2 0 00-2 2v4a2 2 0 002 2zm8-12V5a2 2 0 00-2-2H9a2 2 0 00-2 2v4h10z" />
                </svg>
                Cetak Rekapitulasi
            </Button>
            {#if $userStore?.role !== 'kasubag'}
            <Button variant="default" class="gap-2" on:click={() => goto('/dashboard/gup/pengajuan/new')}>
                <svg xmlns="http://www.w3.org/2000/svg" class="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4v16m8-8H4" />
                </svg>
                Tambah GUP
            </Button>
            {/if}
        </div>
    </div>

    <!-- Filter Section -->
    <div class="bg-white p-4 rounded-xl shadow-sm border border-slate-100 no-print">
        <div class="grid grid-cols-1 md:grid-cols-[minmax(180px,250px)_1fr_1fr_1fr_auto] gap-3 items-end w-full">
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
            <div class="w-full">
                <label for="filterJenisPengadaan" class="mb-1.5 block text-xs font-bold uppercase tracking-widest text-slate-500">Jenis Pengadaan</label>
                <Select id="filterJenisPengadaan" bind:value={filterJenis} class="border-slate-200 w-full" options={[
                    {value: '', label: 'Semua Jenis'},
                    ...masterData.procurementTypes.map(t => ({value: t.id, label: `${t.name} (Akun: ${t.accountCode || '-'})`}))
                ]} />
            </div>
            <!-- Tahun -->
            <div class="w-full">
                <label for="filterTahun" class="mb-1.5 block text-xs font-bold uppercase tracking-widest text-slate-500">Tahun</label>
                <Select id="filterTahun" bind:value={filterTahun} class="border-slate-200 w-full" options={yearOptions} />
            </div>
            <!-- Bulan -->
            <div class="w-full">
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
                <Button variant="outline" on:click={resetFilters} class="h-[42px] hover:text-red-600 hover:border-red-100 w-full md:w-auto">
                    <svg class="h-4 w-4 mr-2" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15" /></svg>
                    Reset
                </Button>
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
                        <th class="font-semibold text-slate-700 px-4 py-3 bg-slate-50 w-[5%] text-center">No</th>
                        <th class="font-semibold text-slate-700 px-4 py-3 bg-slate-50 w-[20%]">Pembayaran</th>
                        <th class="font-semibold text-slate-700 px-4 py-3 bg-slate-50 text-right w-[12%]">Nilai</th>
                        <th class="font-semibold text-slate-700 px-4 py-3 bg-slate-50 text-right w-[15%]">Jumlah Dibayarkan</th>
                        <th class="font-semibold text-slate-700 px-4 py-3 bg-slate-50 text-right w-[12%]">Pajak</th>
                        <th class="font-semibold text-slate-700 px-4 py-3 bg-slate-50 text-right w-[12%]">Selisih</th>
                        <th class="font-semibold text-slate-700 px-4 py-3 bg-slate-50 text-center w-[12%]">Tanggal Kwitansi</th>
                        <th class="font-semibold text-slate-700 px-4 py-3 bg-slate-50 text-center w-[12%] no-print">Aksi</th>
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
                        {#each paginatedTransactions as trx, index (trx.id + '-' + index)}
                            <tr class="hover:bg-slate-50/50 border-b border-slate-100 transition-colors bg-white">
                                <td class="px-4 py-4 align-middle text-center text-sm text-slate-600">{(currentPage - 1) * itemsPerPage + index + 1}</td>
                                <td class="px-4 py-4 align-middle">
                                    <div class="font-medium text-slate-800 text-sm leading-relaxed">{trx.paymentDescription || '-'}</div>
                                    <div class="mt-1 flex items-center gap-2">
                                        <span class="text-[10px] text-slate-500 font-bold bg-slate-100 px-2 py-0.5 rounded-md border border-slate-200">
                                            {getProcurementTypeName(trx.procurementTypeId)}
                                        </span>
                                    </div>
                                </td>
                                <td class="px-4 py-4 align-middle text-right font-medium text-slate-800">{formatCurrency(trx.valueAmount || 0)}</td>
                                <td class="px-4 py-4 align-middle text-right font-bold text-emerald-600">{formatCurrency(trx.paidAmount || 0)}</td>
                                <td class="px-4 py-4 align-middle text-right font-medium text-rose-600">{formatCurrency(trx.taxAmount || 0)}</td>
                                <td class="px-4 py-4 align-middle text-right font-bold {(trx.valueAmount || 0) - (trx.paidAmount || 0) - (trx.taxAmount || 0) < 0 ? 'text-rose-500' : ((trx.valueAmount || 0) - (trx.paidAmount || 0) - (trx.taxAmount || 0) === 0 ? 'text-emerald-600' : 'text-amber-500')}">{formatCurrency((trx.valueAmount || 0) - (trx.paidAmount || 0) - (trx.taxAmount || 0))}</td>
                                <td class="px-4 py-4 align-middle text-center text-sm text-slate-600">{formatDate(trx.receiptDate)}</td>
                                <td class="px-4 py-4 align-middle text-center no-print">
                                    <div class="flex items-center justify-center gap-1.5">
                                        <button class="text-blue-600 hover:text-blue-800 bg-blue-50 hover:bg-blue-100 p-1.5 rounded transition-colors" title="View Detail" on:click={() => handleReview(trx.id)}>
                                            <svg class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                                                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 12a3 3 0 11-6 0 3 3 0 016 0z" />
                                                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M2.458 12C3.732 7.943 7.523 5 12 5c4.478 0 8.268 2.943 9.542 7-1.274 4.057-5.064 7-9.542 7-4.477 0-8.268-2.943-9.542-7z" />
                                            </svg>
                                        </button>
                                        <button class="text-amber-600 hover:text-amber-800 bg-amber-50 hover:bg-amber-100 p-1.5 rounded transition-colors" title="Cetak Kwitansi" on:click={() => handlePrintKwitansi(trx.id)}>
                                            <svg class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                                                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z" />
                                            </svg>
                                        </button>
                                        <button class="text-emerald-600 hover:text-emerald-800 bg-emerald-50 hover:bg-emerald-100 p-1.5 rounded transition-colors" title="Cetak SPBY" on:click={() => handlePrintSpby(trx.id)}>
                                            <svg class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                                                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 5H7a2 2 0 00-2 2v12a2 2 0 002 2h10a2 2 0 002-2V7a2 2 0 00-2-2h-2M9 5a2 2 0 002 2h2a2 2 0 002-2M9 5a2 2 0 012-2h2a2 2 0 012 2" />
                                            </svg>
                                        </button>
                                        {#if $userStore?.role !== 'kasubag'}
                                        <button class="text-indigo-600 hover:text-indigo-800 bg-indigo-50 hover:bg-indigo-100 p-1.5 rounded transition-colors" title="Ubah Pengajuan" on:click={() => goto(`/dashboard/gup/pengajuan/new?id=${trx.id}`)}>
                                            <svg class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                                                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15.232 5.232l3.536 3.536m-2.036-5.036a2.5 2.5 0 113.536 3.536L6.5 21.036H3v-3.572L16.732 3.732z" />
                                            </svg>
                                        </button>
                                        <button class="text-rose-600 hover:text-rose-800 bg-rose-50 hover:bg-rose-100 p-1.5 rounded transition-colors" title="Hapus Pengajuan" on:click={() => handleDelete(trx.id)}>
                                            <svg class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                                                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16" />
                                            </svg>
                                        </button>
                                        {/if}
                                    </div>
                                </td>
                            </tr>
                        {/each}
                    {/if}
                </tbody>
            </table>
        </div>
    </div>

    <!-- Pagination (Canonical Pattern) -->
    {#if totalPages > 1}
        <div class="flex items-center justify-between px-4 py-3 bg-white border border-slate-200 mt-6 rounded-xl shadow-sm no-print">
            <!-- MOBILE: Hanya dua tombol -->
            <div class="flex flex-1 justify-between sm:hidden">
                <Button variant="outline" size="sm" disabled={currentPage === 1} on:click={() => currentPage--}>
                    Sebelumnya
                </Button>
                <Button variant="outline" size="sm" disabled={currentPage === totalPages} on:click={() => currentPage++}>
                    Selanjutnya
                </Button>
            </div>

            <!-- DESKTOP: Info data + nav lengkap -->
            <div class="hidden sm:flex sm:flex-1 sm:items-center sm:justify-between">
                <div>
                    <p class="text-sm text-slate-700">
                        Menampilkan <span class="font-medium">{(currentPage - 1) * itemsPerPage + 1}</span>
                        hingga <span class="font-medium">{Math.min(currentPage * itemsPerPage, filteredTransactions.length)}</span>
                        dari <span class="font-medium">{filteredTransactions.length}</span> hasil
                    </p>
                </div>
                <div>
                    <nav class="isolate inline-flex -space-x-px rounded-md shadow-sm" aria-label="Pagination">
                        <button on:click={() => currentPage--} disabled={currentPage === 1} class="relative inline-flex items-center rounded-l-md px-2 py-2 text-slate-400 ring-1 ring-inset ring-slate-300 hover:bg-slate-50 focus:z-20 focus:outline-offset-0 disabled:opacity-50 disabled:cursor-not-allowed">
                            <span class="sr-only">Previous</span>
                            <svg class="h-5 w-5" viewBox="0 0 20 20" fill="currentColor" aria-hidden="true">
                                <path fill-rule="evenodd" d="M12.79 5.23a.75.75 0 01-.02 1.06L8.832 10l3.938 3.71a.75.75 0 11-1.04 1.08l-4.5-4.25a.75.75 0 010-1.08l4.5-4.25a.75.75 0 011.06.02z" clip-rule="evenodd" />
                            </svg>
                        </button>
                        {#each Array(totalPages) as _, i}
                            {#if totalPages <= 7 || (i === 0 || i === totalPages - 1 || (i >= currentPage - 2 && i <= currentPage))}
                                <button on:click={() => currentPage = i + 1} class="relative inline-flex items-center px-4 py-2 text-sm font-semibold {currentPage === i + 1 ? 'z-10 bg-blue-600 text-white focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-blue-600' : 'text-slate-900 ring-1 ring-inset ring-slate-300 hover:bg-slate-50 focus:z-20 focus:outline-offset-0'}">
                                    {i + 1}
                                </button>
                            {:else if i === 1 || i === totalPages - 2}
                                <span class="relative inline-flex items-center px-4 py-2 text-sm font-semibold text-slate-700 ring-1 ring-inset ring-slate-300 focus:outline-offset-0">...</span>
                            {/if}
                        {/each}
                        <button on:click={() => currentPage++} disabled={currentPage === totalPages} class="relative inline-flex items-center rounded-r-md px-2 py-2 text-slate-400 ring-1 ring-inset ring-slate-300 hover:bg-slate-50 focus:z-20 focus:outline-offset-0 disabled:opacity-50 disabled:cursor-not-allowed">
                            <span class="sr-only">Next</span>
                            <svg class="h-5 w-5" viewBox="0 0 20 20" fill="currentColor" aria-hidden="true">
                                <path fill-rule="evenodd" d="M7.21 14.77a.75.75 0 01.02-1.06L11.168 10 7.23 6.29a.75.75 0 111.04-1.08l4.5 4.25a.75.75 0 010 1.08l-4.5 4.25a.75.75 0 01-1.06-.02z" clip-rule="evenodd" />
                            </svg>
                        </button>
                    </nav>
                </div>
            </div>
        </div>
    {/if}
</div>

<!-- Modal Detail GUP -->
{#if isReviewOpen}
<BaseModal bind:open={isReviewOpen} maxWidth="max-w-3xl">
    <svelte:fragment slot="header">
        <div class="flex flex-wrap items-center gap-2 mb-1">
            <h2 class="text-lg sm:text-xl font-bold text-slate-800 leading-tight">Detail Pengajuan GUP</h2>
        </div>
        <div class="flex items-center gap-2">
            <span class="font-mono text-indigo-600 bg-indigo-50 px-2 py-0.5 rounded border border-indigo-100 text-xs sm:text-sm">
                {selectedGup?.businessId || 'Memuat...'}
            </span>
        </div>
    </svelte:fragment>

    <svelte:fragment slot="body">
        {#if isFetchingDetail}
            <div class="flex flex-col items-center justify-center h-52 space-y-4">
                <div class="w-10 h-10 border-4 border-indigo-200 border-t-indigo-600 rounded-full animate-spin"></div>
                <p class="text-sm font-medium text-slate-500 animate-pulse">Mengambil data detail...</p>
            </div>
        {:else if selectedGup}
            <div class="space-y-5">

                <!-- Kartu: Informasi Dasar -->
                <div class="bg-white p-5 rounded-2xl border border-slate-200 shadow-sm flex flex-col relative overflow-hidden group">
                    <div class="flex items-center gap-3 mb-4 pb-3 border-b border-slate-100/80">
                        <div class="p-2.5 bg-slate-100 text-slate-600 rounded-xl shadow-sm border border-slate-200">
                            <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13 16h-1v-4h-1m1-4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z"></path></svg>
                        </div>
                        <h4 class="text-xs font-bold text-slate-700 uppercase tracking-widest">Informasi Dasar</h4>
                    </div>
                    <div class="grid grid-cols-1 sm:grid-cols-2 gap-x-8 gap-y-4">
                        <div class="space-y-0.5">
                            <p class="text-[10px] font-bold text-slate-400 uppercase tracking-widest">ID Transaksi / SPBY</p>
                            <p class="font-bold text-slate-800 font-mono text-sm">{selectedGup.businessId || '-'}</p>
                        </div>
                        <div class="space-y-0.5">
                            <p class="text-[10px] font-bold text-slate-400 uppercase tracking-widest">Tanggal Kuitansi</p>
                            <p class="font-semibold text-slate-800 text-sm">{formatDate(selectedGup.receiptDate)}</p>
                        </div>
                        <div class="space-y-0.5">
                            <p class="text-[10px] font-bold text-slate-400 uppercase tracking-widest">Penerima Uang (PUM)</p>
                            <p class="font-semibold text-slate-800 text-sm">{selectedGup.pum || '-'}</p>
                        </div>
                        <div class="space-y-0.5">
                            <p class="text-[10px] font-bold text-slate-400 uppercase tracking-widest">Penerima Barang/Jasa</p>
                            <p class="font-semibold text-slate-800 text-sm">{selectedGup.recipient || '-'}</p>
                        </div>
                        <div class="sm:col-span-2 space-y-1">
                            <p class="text-[10px] font-bold text-slate-400 uppercase tracking-widest">Uraian Pembayaran</p>
                            <p class="font-medium text-slate-700 bg-slate-50 p-3 rounded-xl border border-slate-100 text-sm leading-relaxed">{selectedGup.paymentDescription || '-'}</p>
                        </div>
                    </div>
                </div>

                <!-- Kartu: Rincian Nilai Transaksi -->
                <div class="bg-white p-5 rounded-2xl border border-slate-200 shadow-sm flex flex-col relative overflow-hidden group">
                    <div class="flex items-center gap-3 mb-4 pb-3 border-b border-slate-100/80">
                        <div class="p-2.5 bg-slate-100 text-slate-600 rounded-xl shadow-sm border border-slate-200">
                            <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 8c-1.657 0-3 .895-3 2s1.343 2 3 2 3 .895 3 2-1.343 2-3 2m0-8c1.11 0 2.08.402 2.599 1M12 8V7m0 1v8m0 0v1m0-1c-1.11 0-2.08-.402-2.599-1M21 12a9 9 0 11-18 0 9 9 0 0118 0z"></path></svg>
                        </div>
                        <h4 class="text-xs font-bold text-slate-700 uppercase tracking-widest">Rincian Nilai Transaksi</h4>
                    </div>
                    
                    <div class="mb-4 flex items-center justify-between px-4 py-3 bg-indigo-50/50 rounded-xl border border-indigo-100/50">
                        <span class="text-sm font-semibold text-indigo-700">Sumber Dana Pengajuan</span>
                        <span class="text-sm font-bold text-indigo-900 bg-indigo-100/80 px-3 py-1.5 rounded-lg uppercase tracking-wide">{masterData.fundingSources.find(f => f.id === selectedGup.fundingSourceId)?.gupLabel || '-'}</span>
                    </div>

                    <div class="grid grid-cols-1 sm:grid-cols-3 gap-4">
                        <div class="bg-slate-50 rounded-xl p-4 border border-slate-100 space-y-1">
                            <p class="text-[10px] font-bold text-slate-500 uppercase tracking-wider">Nilai Kuitansi</p>
                            <p class="text-lg font-bold text-slate-900">{formatCurrency(selectedGup.valueAmount)}</p>
                        </div>
                        <div class="bg-emerald-50 rounded-xl p-4 border border-emerald-100 space-y-1">
                            <p class="text-[10px] font-bold text-emerald-700 uppercase tracking-wider">Yang Dibayarkan (Netto)</p>
                            <p class="text-lg font-bold text-emerald-600">{formatCurrency(selectedGup.paidAmount)}</p>
                        </div>
                        <div class="bg-rose-50 rounded-xl p-4 border border-rose-100 space-y-1">
                            <p class="text-[10px] font-bold text-rose-700 uppercase tracking-wider">Total Pajak</p>
                            <p class="text-lg font-bold text-rose-600">{formatCurrency(selectedGup.taxAmount)}</p>
                        </div>
                    </div>
                    <!-- Selisih -->
                    <div class="mt-3 flex items-center justify-between px-4 py-3 bg-slate-50 rounded-xl border border-slate-200">
                        <span class="text-sm font-semibold text-slate-600">Selisih (Nilai − Dibayar − Pajak)</span>
                        <span class="text-sm font-bold text-slate-800">{formatCurrency((selectedGup.valueAmount || 0) - (selectedGup.paidAmount || 0) - (selectedGup.taxAmount || 0))}</span>
                    </div>
                </div>

                <!-- Kartu: Detail Klasifikasi (MAK) -->
                <div class="bg-white p-5 rounded-2xl border border-slate-200 shadow-sm flex flex-col relative overflow-hidden group">
                    <div class="flex items-center gap-3 mb-4 pb-3 border-b border-slate-100/80">
                        <div class="p-2.5 bg-amber-50 text-amber-600 rounded-xl shadow-sm border border-amber-100">
                            <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 7h6m0 10v-3m-3 3h.01M9 17h.01M9 14h.01M12 14h.01M15 11h.01M12 11h.01M9 11h.01M7 21h10a2 2 0 002-2V5a2 2 0 00-2-2H7a2 2 0 00-2 2v14a2 2 0 002 2z"></path></svg>
                        </div>
                        <h4 class="text-xs font-bold text-slate-700 uppercase tracking-widest">Detail Klasifikasi (MAK)</h4>
                    </div>
                    <div class="space-y-3">
                        <div class="flex flex-col sm:flex-row sm:justify-between sm:items-center gap-1.5 py-2.5 border-b border-slate-50">
                            <span class="text-xs font-semibold text-slate-500">Jenis Pengadaan</span>
                            <span class="text-sm font-bold text-slate-800">{getProcurementTypeName(selectedGup.procurementTypeId)}</span>
                        </div>
                        <div class="flex flex-col sm:flex-row sm:justify-between sm:items-start gap-1.5 py-2.5 border-b border-slate-50">
                            <span class="text-xs font-semibold text-slate-500 shrink-0">Program / Kegiatan / KRO / RO / Komponen / Sub</span>
                            <span class="text-sm font-bold text-slate-800 sm:text-right">{masterData.procurementTypes.find(t => t.id === selectedGup.procurementTypeId)?.accountMak || '-'}</span>
                        </div>
                        <div class="flex flex-col sm:flex-row sm:justify-between sm:items-center gap-1.5 py-2.5">
                            <span class="text-xs font-semibold text-slate-500">Kode Akun</span>
                            <span class="text-sm font-bold text-slate-800 font-mono bg-slate-100 px-2.5 py-1 rounded-lg inline-block">{masterData.procurementTypes.find(t => t.id === selectedGup.procurementTypeId)?.accountCode || '-'}</span>
                        </div>
                    </div>
                </div>

            </div>
        {:else}
            <div class="flex flex-col items-center justify-center h-48 text-center">
                <div class="inline-flex items-center justify-center w-16 h-16 rounded-full bg-slate-100 text-slate-300 mb-4">
                    <svg class="h-8 w-8" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9.172 16.172a4 4 0 015.656 0M9 10h.01M15 10h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
                    </svg>
                </div>
                <p class="text-slate-500 font-medium">Gagal memuat data detail pengajuan GUP.</p>
            </div>
        {/if}
    </svelte:fragment>

    <svelte:fragment slot="footer">
        <button
            class="px-4 py-2 text-sm font-medium border border-slate-200 text-slate-700 bg-white hover:bg-slate-50 rounded-xl transition-colors"
            on:click={() => isReviewOpen = false}
        >
            Tutup
        </button>
        <Button variant="default" class="gap-2" disabled={!selectedGup} on:click={() => { isReviewOpen = false; handlePrintSpby(selectedGup.id); }}>
            <svg xmlns="http://www.w3.org/2000/svg" class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M17 17h2a2 2 0 002-2v-4a2 2 0 00-2-2H5a2 2 0 00-2 2v4a2 2 0 002 2h2m2 4h6a2 2 0 002-2v-4a2 2 0 00-2-2H9a2 2 0 00-2 2v4a2 2 0 002 2zm8-12V5a2 2 0 00-2-2H9a2 2 0 00-2 2v4h10z" />
            </svg>
            Cetak SPBY
        </Button>
    </svelte:fragment>
</BaseModal>
{/if}

<KwitansiPrintModal bind:isOpen={isKwitansiOpen} data={printData} />
<SpbyPrintModal bind:isOpen={isSpbyOpen} data={printData} />
<RekapitulasiPrintModal bind:isOpen={isRekapitulasiOpen} transactions={transactions} {masterData} />

<ConfirmationModal 
    bind:open={isDeleteModalOpen}
    title="Hapus Pengajuan GUP"
    description="Apakah Anda yakin ingin menghapus pengajuan GUP ini secara permanen? Tindakan ini tidak dapat dibatalkan."
    confirmText="Ya, Hapus"
    cancelText="Batal"
    onConfirm={processDelete}
    variant="destructive"
/>
