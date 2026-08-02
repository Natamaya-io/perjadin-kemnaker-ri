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

<div class="p-4 sm:p-6 lg:p-8 space-y-6 max-w-screen-2xl mx-auto pb-20">
    <!-- Header halaman dan tombol aksi -->
    <section class="bg-white rounded-2xl shadow-sm border border-slate-200 p-5 sm:p-6">
        <div class="flex flex-col xl:flex-row xl:items-center xl:justify-between gap-4">
            <div>
                <p class="text-sm text-slate-500 mb-1">Modul Pengajuan Ganti Uang Persediaan</p>
                <h1 class="text-2xl sm:text-3xl font-bold text-slate-900">Pengajuan GUP</h1>
                <p class="text-sm text-slate-500 mt-2 max-w-3xl">
                    Tabel menampilkan ringkasan pengajuan. Gunakan tombol <strong>View</strong> untuk rincian, <strong>Cetak Kwitansi</strong> untuk bukti pembayaran, dan <strong>Cetak SPBY</strong> untuk Surat Perintah Bayar.
                </p>
            </div>

            <div class="flex flex-col sm:flex-row gap-3 no-print">
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
                    class="inline-flex items-center justify-center gap-2 rounded-xl bg-indigo-500 px-6 py-3 text-sm font-bold text-white border border-indigo-400"
                    style="box-shadow: inset 2px 2px 5px rgba(255,255,255,0.4), inset -3px -3px 7px rgba(0,0,0,0.15);"
                    on:click={() => goto('/dashboard/gup/pengajuan/new')}
                >
                    <svg xmlns="http://www.w3.org/2000/svg" class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2.5" d="M12 4v16m8-8H4" />
                    </svg>
                    Tambah GUP
                </button>
            </div>
        </div>
    </section>

    <!-- Tabel Pengajuan GUP -->
    <section class="bg-white rounded-2xl shadow-sm border border-slate-200 overflow-hidden">
        <div class="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-3 px-6 py-5 border-b border-slate-200">
            <div>
                <h2 class="text-lg font-bold text-slate-900">Tabel Pengajuan GUP</h2>
                <p class="text-sm text-slate-500">Kode Akun dan MAK dibuat otomatis berdasarkan Jenis Pengadaan.</p>
            </div>
            <div class="flex items-center gap-2 text-sm text-slate-500 no-print">
                <svg xmlns="http://www.w3.org/2000/svg" class="h-4 w-4 text-blue-500" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13 16h-1v-4h-1m1-4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
                </svg>
                <span>Tersedia View, Cetak Kwitansi, dan Cetak SPBY.</span>
            </div>
        </div>

        <div class="grid grid-cols-1 gap-3 border-b border-slate-200 bg-slate-50/70 px-6 py-4 md:grid-cols-2 xl:grid-cols-4 no-print">
            <div>
                <label for="filterJenisPengadaan" class="mb-2 block text-xs font-bold uppercase tracking-wide text-slate-500">Jenis Pengadaan</label>
                <select id="filterJenisPengadaan" bind:value={filterJenis} class="w-full rounded-xl border border-slate-300 bg-white px-4 py-3 text-sm text-slate-700 focus:outline-none focus:ring-2 focus:ring-indigo-500">
                    <option value="">Semua Jenis Pengadaan</option>
                    {#each masterData.procurementTypes as type}
                        <option value={type.id}>{type.name}</option>
                    {/each}
                </select>
            </div>
            <div>
                <label for="filterBulan" class="mb-2 block text-xs font-bold uppercase tracking-wide text-slate-500">Bulan Transaksi</label>
                <select id="filterBulan" bind:value={filterBulan} class="w-full rounded-xl border border-slate-300 bg-white px-4 py-3 text-sm text-slate-700 focus:outline-none focus:ring-2 focus:ring-indigo-500">
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
                <label for="searchQuery" class="mb-2 block text-xs font-bold uppercase tracking-wide text-slate-500">Search</label>
                <div class="relative">
                    <svg class="absolute left-4 top-1/2 -translate-y-1/2 h-4 w-4 text-slate-400" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z" />
                    </svg>
                    <input id="searchQuery" type="search" bind:value={searchQuery} placeholder="Cari pembayaran, penerima..." class="w-full rounded-xl border border-slate-300 bg-white py-3 pl-11 pr-4 text-sm text-slate-700 focus:outline-none focus:ring-2 focus:ring-indigo-500">
                </div>
            </div>
            <div class="flex items-end gap-3">
                <button type="button" on:click={resetFilters} class="inline-flex flex-1 items-center justify-center gap-2 rounded-xl border border-slate-300 bg-white px-4 py-3 text-sm font-semibold text-slate-700 transition hover:bg-slate-100">
                    <svg class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15" />
                    </svg>
                    Reset Filter
                </button>
                <div class="min-w-[6rem] rounded-xl bg-slate-900 px-3 py-3 text-center text-xs font-bold text-white" title="Jumlah data yang ditampilkan">
                    <span>{filteredTransactions.length}</span> data
                </div>
            </div>
        </div>

        <div class="overflow-x-auto">
            <table class="w-full text-sm">
                <thead class="bg-slate-800 text-white">
                    <tr>
                        <th class="px-4 py-4 text-left font-semibold whitespace-nowrap">No</th>
                        <th class="px-4 py-4 text-left font-semibold whitespace-nowrap">Pembayaran</th>
                        <th class="px-4 py-4 text-right font-semibold whitespace-nowrap">Nilai</th>
                        <th class="px-4 py-4 text-right font-semibold whitespace-nowrap">Jumlah Dibayarkan</th>
                        <th class="px-4 py-4 text-right font-semibold whitespace-nowrap">Pajak</th>
                        <th class="px-4 py-4 text-right font-semibold whitespace-nowrap">Selisih</th>
                        <th class="px-4 py-4 text-left font-semibold whitespace-nowrap">Tanggal Kwitansi</th>
                        <th class="px-4 py-4 text-center font-semibold whitespace-nowrap no-print">Aksi</th>
                    </tr>
                </thead>
                <tbody class="divide-y divide-slate-100">
                    {#if filteredTransactions.length === 0}
                        <tr>
                            <td colspan="8" class="px-6 py-10 text-center text-sm text-slate-500">
                                <svg class="h-12 w-12 mx-auto text-slate-300 mb-3" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9.172 16.172a4 4 0 015.656 0M9 10h.01M15 10h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
                                </svg>
                                Tidak ada data Pengajuan GUP yang sesuai.
                            </td>
                        </tr>
                    {:else}
                        {#each filteredTransactions as trx, index (trx.id)}
                            <tr class="hover:bg-slate-50 transition">
                                <td class="px-4 py-4 font-semibold text-slate-900 whitespace-nowrap">{index + 1}</td>
                                <td class="px-4 py-4 text-slate-700 min-w-[220px]">
                                    <div class="font-mono text-[10px] text-indigo-600 font-bold mb-1">{trx.businessId}</div>
                                    <div class="line-clamp-2">{trx.paymentDescription}</div>
                                </td>
                                <td class="px-4 py-4 text-right font-semibold text-slate-900 whitespace-nowrap">{formatCurrency(trx.valueAmount)}</td>
                                <td class="px-4 py-4 text-right font-semibold text-emerald-600 whitespace-nowrap">{formatCurrency(trx.paidAmount)}</td>
                                <td class="px-4 py-4 text-right text-rose-500 whitespace-nowrap">{formatCurrency(trx.taxAmount)}</td>
                                <td class="px-4 py-4 text-right font-semibold text-slate-900 whitespace-nowrap">{formatCurrency(trx.valueAmount - trx.paidAmount - trx.taxAmount)}</td>
                                <td class="px-4 py-4 text-slate-700 whitespace-nowrap">
                                    <div>{formatDate(trx.receiptDate)}</div>
                                    <div class="text-xs text-slate-500 mt-1">{trx.recipient}</div>
                                </td>
                                <td class="px-4 py-4 text-center whitespace-nowrap no-print">
                                    <div class="flex items-center justify-center gap-2">
                                        <button 
                                            class="inline-flex items-center justify-center gap-1.5 rounded-lg bg-slate-800 px-3 py-2 text-xs font-semibold text-white border border-slate-700"
                                            style="box-shadow: inset 1px 1px 3px rgba(255,255,255,0.2), inset -1px -1px 3px rgba(0,0,0,0.5);"
                                            title="View Detail"
                                        >
                                            <svg class="h-3.5 w-3.5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                                                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2.5" d="M15 12a3 3 0 11-6 0 3 3 0 016 0z" />
                                                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2.5" d="M2.458 12C3.732 7.943 7.523 5 12 5c4.478 0 8.268 2.943 9.542 7-1.274 4.057-5.064 7-9.542 7-4.477 0-8.268-2.943-9.542-7z" />
                                            </svg>
                                            View
                                        </button>
                                        <button 
                                            class="inline-flex items-center justify-center gap-1.5 rounded-lg bg-amber-500 px-3 py-2 text-xs font-semibold text-slate-950 border border-amber-400"
                                            style="box-shadow: inset 1px 1px 3px rgba(255,255,255,0.4), inset -1px -1px 3px rgba(0,0,0,0.15);"
                                            title="Cetak Kwitansi"
                                        >
                                            <svg class="h-3.5 w-3.5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                                                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z" />
                                            </svg>
                                            Kwitansi
                                        </button>
                                        <button 
                                            class="inline-flex items-center justify-center gap-1.5 rounded-lg bg-sky-600 px-3 py-2 text-xs font-semibold text-white border border-sky-500"
                                            style="box-shadow: inset 1px 1px 3px rgba(255,255,255,0.3), inset -1px -1px 3px rgba(0,0,0,0.4);"
                                            title="Cetak SPBY"
                                        >
                                            <svg class="h-3.5 w-3.5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                                                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 5H7a2 2 0 00-2 2v12a2 2 0 002 2h10a2 2 0 002-2V7a2 2 0 00-2-2h-2M9 5a2 2 0 002 2h2a2 2 0 002-2M9 5a2 2 0 012-2h2a2 2 0 012 2" />
                                            </svg>
                                            SPBY
                                        </button>
                                    </div>
                                </td>
                            </tr>
                        {/each}
                    {/if}
                </tbody>
            </table>
        </div>
    </section>

    <!-- Informasi Integrasi Otomatis -->
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
