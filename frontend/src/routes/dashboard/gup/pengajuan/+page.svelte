<script lang="ts">
    import { formatCurrency } from '$lib/shared/utils/utils';
    import { goto } from '$app/navigation';
    
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

    function resetFilters() {
        searchQuery = '';
        filterBulan = '';
        filterJenis = '';
    }
</script>

<<div class="space-y-6 pb-20 max-w-7xl mx-auto">
    <!-- Header halaman dan tombol aksi -->
    <div class="flex flex-col md:flex-row md:items-center justify-between gap-4 bg-white p-6 rounded-2xl shadow-sm border border-slate-100">
        <div>
            <h1 class="text-2xl font-bold text-slate-800 tracking-tight">Pengajuan GUP</h1>
            <p class="text-sm text-slate-500 mt-1">Tabel menampilkan ringkasan pengajuan. Gunakan tombol View untuk rincian.</p>
        </div>
        <div class="flex items-center gap-3 no-print">
            <button 
                class="inline-flex items-center justify-center gap-2 rounded-xl bg-amber-500 px-6 py-3 text-sm font-bold text-slate-950 border border-amber-400"
                style="box-shadow: inset 2px 2px 5px rgba(255,255,255,0.4), inset -3px -3px 7px rgba(0,0,0,0.15);"
            >
                <svg xmlns="http://www.w3.org/2000/svg" class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M17 17h2a2 2 0 002-2v-4a2 2 0 00-2-2H5a2 2 0 00-2 2v4a2 2 0 002 2h2m2 4h6a2 2 0 002-2v-4a2 2 0 00-2-2H9a2 2 0 00-2 2v4a2 2 0 002 2zm8-12V5a2 2 0 00-2-2H9a2 2 0 00-2 2v4h10z" />
                </svg>
                Cetak Rekapitulasi
            </button>
            <button 
                class="inline-flex items-center justify-center gap-2 rounded-xl bg-indigo-600 px-6 py-3 text-sm font-bold text-white shadow-sm hover:bg-indigo-700 transition-colors"
                on:click={() => goto('/dashboard/gup/pengajuan/new')}
            >
                <svg xmlns="http://www.w3.org/2000/svg" class="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4v16m8-8H4" />
                </svg>
                Tambah GUP
            </button>
        </div>
    </div>

    <!-- Filter Section -->
    <div class="bg-white p-4 rounded-xl shadow-sm border border-slate-100 no-print">
        <div class="grid grid-cols-1 md:grid-cols-2 xl:grid-cols-4 gap-4">
            <div>
                <label for="filterJenisPengadaan" class="mb-1.5 block text-xs font-bold uppercase tracking-widest text-slate-500">Jenis Pengadaan</label>
                <select id="filterJenisPengadaan" bind:value={filterJenis} class="w-full rounded-lg border border-slate-200 bg-slate-50 px-3 py-2.5 text-sm text-slate-700 focus:outline-none focus:ring-2 focus:ring-indigo-500 focus:bg-white transition-colors">
                    <option value="">Semua Jenis</option>
                    {#each masterData.procurementTypes as type}
                        <option value={type.id}>{type.name}</option>
                    {/each}
                </select>
            </div>
            <div>
                <label for="filterBulan" class="mb-1.5 block text-xs font-bold uppercase tracking-widest text-slate-500">Bulan Transaksi</label>
                <select id="filterBulan" bind:value={filterBulan} class="w-full rounded-lg border border-slate-200 bg-slate-50 px-3 py-2.5 text-sm text-slate-700 focus:outline-none focus:ring-2 focus:ring-indigo-500 focus:bg-white transition-colors">
                    <option value="">Semua Bulan</option>
                    <option value="01">Januari</option>
                    <option value="02">Februari</option>
                    <option value="03">Maret</option>
                    <option value="04">April</option>
                    <option value="05">Mei</option>
                    <option value="06">Juni</option>
                    <option value="07">Juli</option>
                    <option value="08">Agustus</option>
                    <option value="09">September</option>
                    <option value="10">Oktober</option>
                    <option value="11">November</option>
                    <option value="12">Desember</option>
                </select>
            </div>
            <div>
                <label for="searchQuery" class="mb-1.5 block text-xs font-bold uppercase tracking-widest text-slate-500">Pencarian</label>
                <div class="relative">
                    <svg class="absolute left-3 top-1/2 -translate-y-1/2 h-4 w-4 text-slate-400" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z" />
                    </svg>
                    <input id="searchQuery" type="search" bind:value={searchQuery} placeholder="Cari pembayaran..." class="w-full rounded-lg border border-slate-200 bg-slate-50 py-2.5 pl-9 pr-4 text-sm text-slate-700 focus:outline-none focus:ring-2 focus:ring-indigo-500 focus:bg-white transition-colors">
                </div>
            </div>
            <div class="flex items-end gap-3">
                <button type="button" on:click={resetFilters} class="inline-flex flex-1 items-center justify-center gap-2 rounded-lg border border-slate-200 bg-white px-4 py-2.5 text-sm font-semibold text-slate-700 transition hover:bg-slate-50 shadow-sm">
                    Reset Filter
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
                        <th class="font-semibold text-slate-700 pl-4 py-3 bg-slate-50 w-12 text-center">No</th>
                        <th class="font-semibold text-slate-700 py-3 bg-slate-50 min-w-[250px]">Pembayaran</th>
                        <th class="font-semibold text-slate-700 py-3 bg-slate-50 text-right">Nilai</th>
                        <th class="font-semibold text-slate-700 py-3 bg-slate-50 text-right">Dibayarkan</th>
                        <th class="font-semibold text-slate-700 py-3 bg-slate-50 text-right">Pajak</th>
                        <th class="font-semibold text-slate-700 py-3 bg-slate-50 text-right">Selisih</th>
                        <th class="font-semibold text-slate-700 py-3 bg-slate-50 min-w-[140px]">Kwitansi</th>
                        <th class="font-semibold text-slate-700 py-3 bg-slate-50 text-center pr-4 w-[160px] no-print">Aksi</th>
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
                                <td class="pl-4 py-4 align-middle text-center font-medium text-slate-500">{index + 1}</td>
                                <td class="py-4 align-middle">
                                    <span class="inline-flex items-center font-mono text-[13px] font-bold tracking-widest text-indigo-700 mb-1">{trx.businessId}</span>
                                    <div class="font-medium text-slate-800 text-sm line-clamp-2">{trx.paymentDescription}</div>
                                </td>
                                <td class="py-4 align-middle text-right font-medium text-slate-800">{formatCurrency(trx.valueAmount)}</td>
                                <td class="py-4 align-middle text-right font-semibold text-emerald-600">{formatCurrency(trx.paidAmount)}</td>
                                <td class="py-4 align-middle text-right text-rose-500">{formatCurrency(trx.taxAmount)}</td>
                                <td class="py-4 align-middle text-right font-medium text-slate-800">{formatCurrency(trx.valueAmount - trx.paidAmount - trx.taxAmount)}</td>
                                <td class="py-4 align-middle text-xs text-slate-600">
                                    <div class="font-medium">{formatDate(trx.receiptDate)}</div>
                                    <div class="text-[10px] text-slate-400 mt-0.5">{trx.recipient}</div>
                                </td>
                                <td class="pr-4 py-4 align-middle text-center no-print">
                                    <div class="flex items-center justify-center gap-1.5">
                                        <button class="text-blue-600 hover:text-blue-800 bg-blue-50 hover:bg-blue-100 p-1.5 rounded transition-colors" title="View Detail">
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
    <section class="grid grid-cols-1 xl:grid-cols-3 gap-6 no-print">
        <div class="bg-white rounded-2xl shadow-sm border border-slate-200 p-6 xl:col-span-2">
            <h3 class="text-lg font-bold text-slate-900 mb-4">Integrasi Jenis Pengadaan, Kode Akun, dan MAK</h3>
            <p class="text-sm text-slate-500 mb-4">
                Pada form Tambah GUP, pengguna cukup memilih <strong>Jenis Pengadaan</strong>. Sistem akan mengisi <strong>Kode Akun</strong> dan <strong>MAK</strong> secara otomatis berdasarkan data master.
            </p>
            <div class="overflow-x-auto">
                <table class="w-full text-sm">
                    <thead class="bg-slate-50 text-slate-600">
                        <tr>
                            <th class="px-4 py-3 text-left font-semibold">Jenis Pengadaan</th>
                            <th class="px-4 py-3 text-left font-semibold">Kode Akun</th>
                            <th class="px-4 py-3 text-left font-semibold">Format MAK</th>
                        </tr>
                    </thead>
                    <tbody class="divide-y divide-slate-100">
                        {#each masterData.procurementTypes as type}
                            <tr>
                                <td class="px-4 py-3 text-slate-700">{type.name}</td>
                                <td class="px-4 py-3 font-semibold text-slate-900 whitespace-nowrap">{type.accountCode || '-'}</td>
                                <td class="px-4 py-3 text-slate-600 min-w-[320px]">{type.accountMak || '-'}</td>
                            </tr>
                        {/each}
                    </tbody>
                </table>
            </div>
        </div>

        <div class="bg-white rounded-2xl shadow-sm border border-slate-200 p-6">
            <h3 class="text-lg font-bold text-slate-900 mb-4">Rincian Sumber Dana</h3>
            <div class="grid grid-cols-1 sm:grid-cols-2 xl:grid-cols-1 gap-3 text-sm">
                {#each masterData.fundingSources as src}
                    <div class="rounded-xl border border-slate-200 p-3 bg-slate-50">
                        <div class="flex items-center justify-between gap-3">
                            <p class="text-slate-500">{src.monthName}</p>
                            <span class="rounded-full bg-sky-100 px-2 py-1 text-[11px] font-bold text-sky-700">LS</span>
                        </div>
                        <p class="font-bold text-slate-900">{src.gupLabel}</p>
                        <p class="mt-1 text-xs font-semibold text-slate-400">
                            <!-- Backend LS logic can be injected here -->
                            Data LS belum dihubungkan
                        </p>
                    </div>
                {/each}
            </div>
        </div>
    </section>
</div>
