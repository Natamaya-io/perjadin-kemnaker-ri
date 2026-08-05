<script>
    import { formatCurrency } from '$lib/shared/utils/utils';
    import { onMount } from 'svelte';
    import Chart from 'chart.js/auto';
    import Button from '$lib/shared/ui/button/Button.svelte';
    import { api } from '$lib/shared/api';

    export let data;

    // Data dari API (reaktif)
    $: laporan = data?.laporan || { summary: {}, rows: [] };
    $: laporanRows = laporan.rows || [];
    $: laporanSummary = laporan.summary || {};
    $: currentYear = data?.year || new Date().getFullYear();

    const colors = ['#0ea5e9', '#f59e0b', '#8b5cf6', '#ef4444', '#10b981', '#f97316', '#06b6d4'];

    // Chart refs
    let totalCanvas;
    let komposisiCanvas;
    let detailCanvases = [];
    let totalChart, komposisiChart;
    let detailCharts = [];

    // Modal Input Anggaran
    let showAnggaranModal = false;
    let selectedRow = null;
    let anggaranInput = '';
    let savingAnggaran = false;
    let saveError = '';

    // Filter state
    let filterSearch = '';
    let filterStatus = 'semua'; // semua | baik | progress | melebihi | belum
    let filterSort = 'nama'; // nama | realisasi_desc | realisasi_asc | serapan_desc
    let showStatusDropdown = false;
    let showSortDropdown = false;

    const statusOptions = [
        { value: 'semua',    label: 'Semua Status' },
        { value: 'baik',     label: 'Baik (≥ 80%)' },
        { value: 'progress', label: 'Dalam Progress (< 80%)' },
        { value: 'melebihi', label: 'Melebihi Anggaran (> 100%)' },
        { value: 'belum',    label: 'Belum Diinput' }
    ];
    const sortOptions = [
        { value: 'nama',          label: 'Nama A–Z' },
        { value: 'realisasi_desc', label: 'Realisasi Tertinggi' },
        { value: 'realisasi_asc',  label: 'Realisasi Terendah' },
        { value: 'serapan_desc',   label: 'Serapan Tertinggi' }
    ];

    $: statusLabel = statusOptions.find(o => o.value === filterStatus)?.label ?? 'Semua Status';
    $: sortLabel   = sortOptions.find(o => o.value === filterSort)?.label ?? 'Nama A–Z';

    // Baris yang sudah difilter dan diurutkan
    $: filteredRows = laporanRows
        .filter(row => {
            const matchSearch = filterSearch.trim() === '' ||
                row.jenisPengadaan.toLowerCase().includes(filterSearch.trim().toLowerCase()) ||
                row.kodeAkun.toLowerCase().includes(filterSearch.trim().toLowerCase()) ||
                row.mak.toLowerCase().includes(filterSearch.trim().toLowerCase());

            const pct = row.persentaseSerapan || 0;
            const matchStatus = filterStatus === 'semua' ? true
                : filterStatus === 'belum'    ? (row.anggaran === 0)
                : filterStatus === 'melebihi' ? (row.anggaran > 0 && pct > 100)
                : filterStatus === 'baik'     ? (row.anggaran > 0 && pct >= 80 && pct <= 100)
                : filterStatus === 'progress' ? (row.anggaran > 0 && pct < 80)
                : true;

            return matchSearch && matchStatus;
        })
        .sort((a, b) => {
            if (filterSort === 'realisasi_desc') return b.realisasi - a.realisasi;
            if (filterSort === 'realisasi_asc')  return a.realisasi - b.realisasi;
            if (filterSort === 'serapan_desc')   return b.persentaseSerapan - a.persentaseSerapan;
            return a.jenisPengadaan.localeCompare(b.jenisPengadaan, 'id');
        });

    // Paginasi — Section 11 UI Styling Guidelines
    const ITEMS_PER_PAGE = 4;
    let currentPage = 1;
    $: totalPages = Math.ceil(filteredRows.length / ITEMS_PER_PAGE);
    $: pagedRows = filteredRows.slice((currentPage - 1) * ITEMS_PER_PAGE, currentPage * ITEMS_PER_PAGE);
    $: globalOffset = (currentPage - 1) * ITEMS_PER_PAGE;

    // Reset halaman jika filter/data berubah
    $: if (filteredRows || filterSearch || filterStatus || filterSort) { currentPage = 1; }

    function resetFilters() {
        filterSearch = '';
        filterStatus = 'semua';
        filterSort = 'nama';
    }

    const commonOptions = {
        responsive: true,
        maintainAspectRatio: false,
        plugins: {
            legend: {
                display: true,
                position: 'bottom',
                labels: {
                    font: { size: 11, family: 'Inter, sans-serif' },
                    color: '#64748b',
                    padding: 12,
                    boxWidth: 12,
                    boxHeight: 12,
                    borderRadius: 4,
                    useBorderRadius: true
                }
            },
            tooltip: {
                backgroundColor: '#0f172a',
                padding: 12,
                cornerRadius: 12,
                displayColors: false,
                titleFont: { size: 12, family: 'Inter, sans-serif' },
                bodyFont: { size: 14, weight: 'bold', family: 'Inter, sans-serif' }
            }
        }
    };

    function buildTotalChart() {
        if (!totalCanvas) return;
        if (totalChart) totalChart.destroy();
        totalChart = new Chart(totalCanvas, {
            type: 'doughnut',
            data: {
                labels: ['Realisasi', 'Sisa Anggaran'],
                datasets: [{
                    data: [laporanSummary.totalRealisasi || 0, laporanSummary.sisaAnggaran || 0],
                    backgroundColor: ['#10b981', '#e2e8f0'],
                    borderWidth: 2,
                    borderColor: '#ffffff',
                    hoverOffset: 8
                }]
            },
            options: {
                ...commonOptions,
                cutout: '70%',
                plugins: { ...commonOptions.plugins, tooltip: { ...commonOptions.plugins.tooltip,
                    callbacks: { label: (ctx) => ctx.label + ': ' + formatCurrency(ctx.raw) }
                }}
            }
        });
    }

    function buildKomposisiChart() {
        if (!komposisiCanvas) return;
        if (komposisiChart) komposisiChart.destroy();
        komposisiChart = new Chart(komposisiCanvas, {
            type: 'pie',
            data: {
                labels: laporanRows.map(r => r.jenisPengadaan),
                datasets: [{
                    data: laporanRows.map(r => r.realisasi || 0),
                    backgroundColor: colors.slice(0, laporanRows.length),
                    borderWidth: 2,
                    borderColor: '#ffffff',
                    hoverOffset: 10
                }]
            },
            options: {
                ...commonOptions,
                plugins: { ...commonOptions.plugins, tooltip: { ...commonOptions.plugins.tooltip,
                    callbacks: {
                        title: (ctx) => laporanRows[ctx[0].dataIndex]?.jenisPengadaan,
                        label: (ctx) => 'Realisasi: ' + formatCurrency(ctx.raw)
                    }
                }}
            }
        });
    }

    function buildDetailCharts() {
        detailCharts.forEach(c => c?.destroy());
        detailCharts = [];
        laporanRows.forEach((row, i) => {
            if (!detailCanvases[i]) return;
            const sisa = Math.max(0, (row.anggaran || 0) - (row.realisasi || 0));
            const chart = new Chart(detailCanvases[i], {
                type: 'pie',
                data: {
                    labels: ['Realisasi', 'Sisa Anggaran'],
                    datasets: [{
                        data: [row.realisasi || 0, sisa],
                        backgroundColor: [colors[i % colors.length], '#e2e8f0'],
                        borderWidth: 2,
                        borderColor: '#ffffff',
                        hoverOffset: 8
                    }]
                },
                options: {
                    ...commonOptions,
                    plugins: { ...commonOptions.plugins, tooltip: { ...commonOptions.plugins.tooltip,
                        callbacks: { label: (ctx) => ctx.label + ': ' + formatCurrency(ctx.raw) }
                    }}
                }
            });
            detailCharts.push(chart);
        });
    }

    onMount(() => {
        buildTotalChart();
        buildKomposisiChart();
        // Detail charts dibuat setelah DOM tersedia
        setTimeout(buildDetailCharts, 50);
    });

    // Reactive rebuild saat data berubah (misal setelah save anggaran)
    $: if (totalCanvas && laporanSummary) { buildTotalChart(); }
    $: if (komposisiCanvas && laporanRows.length) { buildKomposisiChart(); }
    // Rebuild detail charts saat halaman berubah (canvas baru di-render setelah tick)
    $: if (pagedRows) { setTimeout(buildDetailCharts, 50); }


    function openAnggaranModal(row) {
        selectedRow = row;
        anggaranInput = row.anggaran > 0 ? String(row.anggaran) : '';
        saveError = '';
        showAnggaranModal = true;
    }

    function closeAnggaranModal() {
        showAnggaranModal = false;
        selectedRow = null;
        anggaranInput = '';
    }

    async function handleSaveAnggaran() {
        if (!selectedRow || !anggaranInput) return;
        savingAnggaran = true;
        saveError = '';
        try {
            await api.saveGupBudget({
                procurementTypeId: selectedRow.procurementTypeId,
                year: currentYear,
                amount: parseFloat(anggaranInput)
            });
            // Reload data
            const fresh = await api.getGupLaporan(currentYear);
            laporan = fresh || { summary: {}, rows: [] };
            closeAnggaranModal();
            setTimeout(buildDetailCharts, 50);
        } catch (e) {
            saveError = 'Gagal menyimpan anggaran. Coba lagi.';
        } finally {
            savingAnggaran = false;
        }
    }
</script>

<div class="space-y-6 pb-20 w-full">
    <!-- Hero halaman -->
    <section class="relative overflow-hidden rounded-3xl bg-gradient-to-br from-slate-950 via-slate-800 to-sky-900 p-6 sm:p-8 text-white shadow-xl no-print">
        <div class="absolute -right-24 -top-24 h-64 w-64 rounded-full bg-sky-400/20 blur-3xl"></div>
        <div class="absolute -bottom-28 left-20 h-72 w-72 rounded-full bg-amber-300/10 blur-3xl"></div>

        <div class="relative z-10 flex flex-col gap-6 xl:flex-row xl:items-end xl:justify-between">
            <div class="max-w-3xl">
                <div class="mb-4 inline-flex items-center gap-2 rounded-full border border-white/15 bg-white/10 px-4 py-2 text-xs font-semibold uppercase tracking-[0.2em] text-sky-100">
                    <svg xmlns="http://www.w3.org/2000/svg" class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M11 3.055A9.001 9.001 0 1020.945 13H11V3.055z" />
                        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M20.488 9H15V3.512A9.025 9.025 0 0120.488 9z" />
                    </svg>
                    Visualisasi Serapan Anggaran GUP — {currentYear}
                </div>
                <h1 class="text-3xl font-black tracking-tight sm:text-4xl">Laporan dan Rekapitulasi</h1>
                <p class="mt-3 max-w-2xl text-sm leading-6 text-slate-200 sm:text-base">
                    Pantau Anggaran, Realisasi, Sisa Anggaran, dan Persentase Serapan dalam bentuk visualisasi per Jenis Pengadaan.
                </p>
            </div>

            <div class="flex flex-col gap-3 sm:flex-row">
                <Button variant="warning" class="gap-2" on:click={() => openAnggaranModal(laporanRows[0])}>
                    <svg xmlns="http://www.w3.org/2000/svg" class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15.232 5.232l3.536 3.536m-2.036-5.036a2.5 2.5 0 113.536 3.536L6.5 21.036H3v-3.572L16.732 3.732z" />
                    </svg>
                    Input Anggaran
                </Button>
                <Button variant="default" class="gap-2 bg-white/10 text-white border border-white/20 hover:bg-white/20" on:click={() => window.print()}>
                    <svg xmlns="http://www.w3.org/2000/svg" class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M17 17h2a2 2 0 002-2v-4a2 2 0 00-2-2H5a2 2 0 00-2 2v4a2 2 0 002 2h2m2 4h6a2 2 0 002-2v-4a2 2 0 00-2-2H9a2 2 0 00-2 2v4a2 2 0 002 2zm8-12V5a2 2 0 00-2-2H9a2 2 0 00-2 2v4h10z" />
                    </svg>
                    Cetak Laporan
                </Button>
            </div>
        </div>
    </section>

    <!-- Overview -->
    <section class="grid grid-cols-1 gap-6 xl:grid-cols-5">
        <article class="overflow-hidden rounded-3xl border border-slate-200 bg-white p-5 shadow-sm xl:col-span-2">
            <div class="flex flex-col gap-4 sm:flex-row sm:items-start sm:justify-between">
                <div>
                    <p class="text-xs font-bold uppercase tracking-[0.2em] text-slate-400">Overview</p>
                    <h2 class="mt-2 text-xl font-black text-slate-900">Total Serapan Anggaran</h2>
                    <p class="mt-1 text-sm text-slate-500">Komposisi total Realisasi dan Sisa Anggaran.</p>
                </div>
                <div class="rounded-2xl bg-slate-900 px-4 py-3 text-right text-white">
                    <p class="text-xs text-slate-300">Serapan</p>
                    <p class="text-2xl font-black">{(laporanSummary.persentaseSerapan || 0).toFixed(1)}%</p>
                </div>
            </div>

            <div class="mt-5 grid grid-cols-1 gap-4 md:grid-cols-2">
                <div class="relative h-64 rounded-3xl bg-slate-50 p-4">
                    <canvas bind:this={totalCanvas}></canvas>
                </div>
                <div class="grid grid-cols-1 gap-3 content-start">
                    <div class="rounded-2xl border border-slate-100 bg-slate-50 p-4">
                        <p class="text-xs font-semibold text-slate-500">Total Anggaran</p>
                        <h3 class="mt-1 text-xl font-black text-slate-900">{formatCurrency(laporanSummary.totalAnggaran || 0)}</h3>
                    </div>
                    <div class="rounded-2xl border border-emerald-100 bg-emerald-50 p-4">
                        <p class="text-xs font-semibold text-emerald-700">Total Realisasi</p>
                        <h3 class="mt-1 text-xl font-black text-emerald-800">{formatCurrency(laporanSummary.totalRealisasi || 0)}</h3>
                    </div>
                    <div class="rounded-2xl border border-amber-100 bg-amber-50 p-4">
                        <p class="text-xs font-semibold text-amber-700">Sisa Anggaran</p>
                        <h3 class="mt-1 text-xl font-black text-amber-800">{formatCurrency(laporanSummary.sisaAnggaran || 0)}</h3>
                    </div>
                </div>
            </div>
        </article>

        <article class="overflow-hidden rounded-3xl border border-slate-200 bg-white p-5 shadow-sm xl:col-span-3">
            <div class="flex flex-col gap-4 sm:flex-row sm:items-start sm:justify-between">
                <div>
                    <p class="text-xs font-bold uppercase tracking-[0.2em] text-slate-400">Komposisi</p>
                    <h2 class="mt-2 text-xl font-black text-slate-900">Realisasi per Jenis Pengadaan</h2>
                    <p class="mt-1 text-sm text-slate-500">Pie chart ini ditarik dari data Realisasi pada menu Pengajuan GUP.</p>
                </div>
                <div class="inline-flex items-center gap-2 rounded-2xl bg-sky-50 px-4 py-3 text-sm font-bold text-sky-700 border border-sky-100">
                    <svg xmlns="http://www.w3.org/2000/svg" class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 11H5m14 0a2 2 0 012 2v6a2 2 0 01-2 2H5a2 2 0 01-2-2v-6a2 2 0 012-2m14 0V9a2 2 0 00-2-2M5 11V9a2 2 0 002-2m0 0V5a2 2 0 012-2h6a2 2 0 012 2v2M7 7h10" />
                    </svg>
                    <span>{laporanSummary.totalJenisPengadaan || 0}</span> Jenis
                </div>
            </div>

            <div class="mt-5 grid grid-cols-1 gap-5 lg:grid-cols-5">
                <div class="relative h-72 rounded-3xl bg-slate-50 p-4 lg:col-span-3">
                    <canvas bind:this={komposisiCanvas}></canvas>
                </div>
                <div class="space-y-3 lg:col-span-2">
                    <div class="rounded-2xl border border-slate-100 bg-slate-50 p-4">
                        <p class="text-xs font-semibold text-slate-500">Jumlah Transaksi GUP</p>
                        <h3 class="mt-1 text-2xl font-black text-slate-900">{laporanSummary.totalTransaksi || 0}</h3>
                    </div>
                    <div class="rounded-2xl border border-slate-100 bg-slate-50 p-4">
                        <p class="text-xs font-semibold text-slate-500">Nilai Pengajuan</p>
                        <h3 class="mt-1 text-2xl font-black text-slate-900">{formatCurrency(laporanSummary.totalRealisasi || 0)}</h3>
                    </div>
                    <div class="rounded-2xl border border-blue-100 bg-blue-50 p-4 text-sm text-blue-700 flex items-start gap-2">
                        <svg xmlns="http://www.w3.org/2000/svg" class="h-5 w-5 shrink-0 mt-0.5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13 16h-1v-4h-1m1-4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
                        </svg>
                        <span>Jenis Pengadaan dengan realisasi nol tetap muncul pada kartu detail, tetapi tidak mendominasi pie komposisi.</span>
                    </div>
                </div>
            </div>
        </article>
    </section>

    <!-- Pie chart per jenis pengadaan -->
    <section>
        <!-- Header section — Section 1.1 -->
        <div class="flex flex-col gap-4 bg-white p-6 rounded-2xl shadow-sm border border-slate-100 sm:flex-row sm:items-center sm:justify-between">
            <div>
                <h2 class="text-2xl font-bold text-slate-800 tracking-tight">Detail Serapan</h2>
                <p class="text-sm text-slate-500 mt-1">Setiap kartu menampilkan komposisi Realisasi dan Sisa Anggaran berdasarkan anggaran yang diinput.</p>
            </div>
            <div class="flex items-center gap-3 no-print">
                <Button variant="default" class="w-full sm:w-auto flex items-center justify-center gap-2" on:click={() => openAnggaranModal(laporanRows[0])}>
                    <svg xmlns="http://www.w3.org/2000/svg" class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4v16m8-8H4" />
                    </svg>
                    Input Anggaran
                </Button>
            </div>
        </div>

        <!-- Filter section — Section 1.2 -->
        <div class="bg-white p-4 rounded-xl shadow-sm border border-slate-100 mt-4 no-print">
            <div class="flex flex-col gap-3 sm:flex-row sm:items-center">

                <!-- Search input -->
                <div class="relative flex-1">
                    <svg class="absolute left-3 top-1/2 -translate-y-1/2 h-4 w-4 text-slate-400 pointer-events-none" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z" />
                    </svg>
                    <input
                        type="text"
                        placeholder="Cari jenis pengadaan, kode akun, MAK..."
                        bind:value={filterSearch}
                        class="w-full rounded-xl border border-slate-200 bg-slate-50 pl-9 pr-4 py-2.5 text-sm text-slate-700 placeholder-slate-400 focus:outline-none focus:ring-2 focus:ring-indigo-500 focus:bg-white transition-colors"
                    />
                </div>

                <!-- Dropdown Status Serapan — Section 6 -->
                <div class="relative sm:w-56">
                    <button
                        type="button"
                        id="filter-status-btn"
                        class="flex w-full items-center justify-between rounded-xl border border-slate-200 bg-slate-50 px-4 py-2.5 text-sm text-slate-700 focus:outline-none focus:ring-2 focus:ring-indigo-500 focus:bg-white transition-colors"
                        on:click={() => { showStatusDropdown = !showStatusDropdown; showSortDropdown = false; }}
                    >
                        <span class="truncate {filterStatus !== 'semua' ? 'font-semibold text-indigo-700' : ''}">{statusLabel}</span>
                        <svg class="h-4 w-4 text-slate-400 shrink-0 ml-2 transition-transform {showStatusDropdown ? 'rotate-180' : ''}" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 9l-7 7-7-7" />
                        </svg>
                    </button>

                    {#if showStatusDropdown}
                        <div
                            class="absolute z-50 mt-2 w-full origin-top-right rounded-xl border border-slate-100 bg-white shadow-lg overflow-hidden animate-in fade-in zoom-in-95 duration-100"
                            role="listbox"
                            aria-labelledby="filter-status-btn"
                        >
                            <!-- svelte-ignore a11y_click_events_have_key_events -->
                            <ul class="max-h-60 overflow-y-auto py-1">
                                {#each statusOptions as opt}
                                    <!-- svelte-ignore a11y_click_events_have_key_events -->
                                    <!-- svelte-ignore a11y_interactive_supports_focus -->
                                    <li
                                        role="option"
                                        aria-selected={filterStatus === opt.value}
                                        class="cursor-pointer select-none py-2.5 pl-4 pr-4 text-sm transition-colors
                                            {filterStatus === opt.value
                                                ? 'font-semibold text-indigo-700 bg-indigo-50/50 hover:bg-indigo-50 flex items-center justify-between'
                                                : 'text-slate-700 hover:bg-slate-50 hover:text-indigo-600'}"
                                        on:click={() => { filterStatus = opt.value; showStatusDropdown = false; }}
                                    >
                                        <span>{opt.label}</span>
                                        {#if filterStatus === opt.value}
                                            <svg class="h-4 w-4 text-indigo-600" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                                                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 13l4 4L19 7" />
                                            </svg>
                                        {/if}
                                    </li>
                                {/each}
                            </ul>
                        </div>
                    {/if}
                </div>

                <!-- Dropdown Urutan — Section 6 -->
                <div class="relative sm:w-52">
                    <button
                        type="button"
                        id="filter-sort-btn"
                        class="flex w-full items-center justify-between rounded-xl border border-slate-200 bg-slate-50 px-4 py-2.5 text-sm text-slate-700 focus:outline-none focus:ring-2 focus:ring-indigo-500 focus:bg-white transition-colors"
                        on:click={() => { showSortDropdown = !showSortDropdown; showStatusDropdown = false; }}
                    >
                        <span class="truncate {filterSort !== 'nama' ? 'font-semibold text-indigo-700' : ''}">{sortLabel}</span>
                        <svg class="h-4 w-4 text-slate-400 shrink-0 ml-2 transition-transform {showSortDropdown ? 'rotate-180' : ''}" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 9l-7 7-7-7" />
                        </svg>
                    </button>

                    {#if showSortDropdown}
                        <div
                            class="absolute z-50 mt-2 w-full origin-top-right rounded-xl border border-slate-100 bg-white shadow-lg overflow-hidden animate-in fade-in zoom-in-95 duration-100"
                            role="listbox"
                            aria-labelledby="filter-sort-btn"
                        >
                            <!-- svelte-ignore a11y_click_events_have_key_events -->
                            <ul class="max-h-60 overflow-y-auto py-1">
                                {#each sortOptions as opt}
                                    <!-- svelte-ignore a11y_click_events_have_key_events -->
                                    <!-- svelte-ignore a11y_interactive_supports_focus -->
                                    <li
                                        role="option"
                                        aria-selected={filterSort === opt.value}
                                        class="cursor-pointer select-none py-2.5 pl-4 pr-4 text-sm transition-colors
                                            {filterSort === opt.value
                                                ? 'font-semibold text-indigo-700 bg-indigo-50/50 hover:bg-indigo-50 flex items-center justify-between'
                                                : 'text-slate-700 hover:bg-slate-50 hover:text-indigo-600'}"
                                        on:click={() => { filterSort = opt.value; showSortDropdown = false; }}
                                    >
                                        <span>{opt.label}</span>
                                        {#if filterSort === opt.value}
                                            <svg class="h-4 w-4 text-indigo-600" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                                                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 13l4 4L19 7" />
                                            </svg>
                                        {/if}
                                    </li>
                                {/each}
                            </ul>
                        </div>
                    {/if}
                </div>

                <!-- Tombol Reset filter -->
                {#if filterSearch || filterStatus !== 'semua' || filterSort !== 'nama'}
                    <button
                        type="button"
                        class="shrink-0 flex items-center gap-1.5 rounded-xl border border-slate-200 bg-white px-3 py-2.5 text-sm text-slate-600 hover:bg-slate-50 hover:text-red-600 hover:border-red-200 transition-colors"
                        on:click={resetFilters}
                    >
                        <svg class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
                        </svg>
                        Reset
                    </button>
                {/if}
            </div>

            <!-- Info hasil filter -->
            {#if filterSearch || filterStatus !== 'semua'}
                <div class="mt-3 flex items-center gap-2 text-xs text-slate-500">
                    <svg class="h-4 w-4 text-indigo-400" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13 16h-1v-4h-1m1-4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
                    </svg>
                    Menampilkan <span class="font-semibold text-slate-700">{filteredRows.length}</span> dari <span class="font-semibold text-slate-700">{laporanRows.length}</span> jenis pengadaan
                    {#if filterSearch}<span>· pencarian: &ldquo;<em class="text-slate-700">{filterSearch}</em>&rdquo;</span>{/if}
                    {#if filterStatus !== 'semua'}<span>· status: <em class="text-indigo-600">{statusLabel}</em></span>{/if}
                </div>
            {/if}
        </div>

        <div class="mt-6">
            {#if laporanRows.length === 0}
                <div class="p-16 text-center text-slate-500 bg-white rounded-2xl border border-slate-200 shadow-sm">
                    <svg xmlns="http://www.w3.org/2000/svg" class="h-12 w-12 mx-auto text-slate-300 mb-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z" />
                    </svg>
                    <p class="font-semibold">Belum ada data laporan untuk tahun {currentYear}.</p>
                    <p class="text-sm mt-1 text-slate-400">Pastikan Jenis Pengadaan sudah diinput di master data dan ada transaksi GUP.</p>
                </div>
            {:else if filteredRows.length === 0}
                <div class="p-12 text-center text-slate-500 bg-white rounded-2xl border border-slate-200 shadow-sm">
                    <svg xmlns="http://www.w3.org/2000/svg" class="h-12 w-12 mx-auto text-slate-300 mb-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z" />
                    </svg>
                    <p class="font-semibold text-slate-700">Tidak ada hasil yang cocok.</p>
                    <p class="text-sm mt-1 text-slate-400">Coba ubah kata kunci pencarian atau filter status.</p>
                    <button type="button" on:click={resetFilters} class="mt-4 inline-flex items-center gap-2 rounded-xl border border-slate-200 bg-white px-4 py-2 text-sm text-slate-600 hover:bg-slate-50 transition-colors">
                        <svg class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15" /></svg>
                        Reset Filter
                    </button>
                </div>
            {:else}
                <div class="grid grid-cols-1 gap-5 sm:grid-cols-2">
                    {#each pagedRows as row, i}
                        {@const globalI = globalOffset + i}
                        <article class="bg-white p-5 md:p-6 rounded-2xl border border-slate-200 hover:border-blue-300 shadow-sm hover:shadow-lg transition-all duration-300 flex flex-col relative overflow-hidden">
                            <!-- Garis gradasi atas sesuai Section 5 guidelines -->
                            <div class="absolute top-0 left-0 w-full h-1.5" style="background: linear-gradient(to right, {colors[globalI % colors.length]}, {colors[(globalI + 1) % colors.length]})"></div>

                            <div class="flex items-start gap-3 mb-4 pb-3 border-b border-slate-100/80">
                                <div class="p-2.5 rounded-xl shadow-sm border border-slate-100 shrink-0" style="background-color: {colors[globalI % colors.length]}20; color: {colors[globalI % colors.length]}">
                                    <svg xmlns="http://www.w3.org/2000/svg" class="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                                        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M11 3.055A9.001 9.001 0 1020.945 13H11V3.055z" />
                                    </svg>
                                </div>
                                <div class="min-w-0">
                                    <p class="text-[10px] font-bold uppercase tracking-[0.2em] text-slate-400">{row.kodeAkun}</p>
                                    <h3 class="mt-0.5 text-sm font-black leading-snug text-slate-900 truncate">{row.jenisPengadaan}</h3>
                                    <p class="text-[10px] text-slate-400 mt-0.5">{row.jumlahTransaksi} transaksi</p>
                                </div>
                            </div>

                            <!-- Pie chart canvas -->
                            <div class="relative h-52 rounded-xl bg-slate-50 p-3">
                                <canvas bind:this={detailCanvases[globalI]}></canvas>
                            </div>

                            <!-- Stats grid -->
                            <div class="mt-4 grid grid-cols-2 gap-2 text-sm">
                                <div class="rounded-xl bg-slate-50 border border-slate-100 p-3">
                                    <p class="text-[10px] font-semibold uppercase tracking-wider text-slate-500">Anggaran</p>
                                    <p class="mt-1 font-black text-slate-900 text-sm">{formatCurrency(row.anggaran || 0)}</p>
                                </div>
                                <div class="rounded-xl bg-emerald-50 border border-emerald-100 p-3">
                                    <p class="text-[10px] font-semibold uppercase tracking-wider text-emerald-700">Realisasi</p>
                                    <p class="mt-1 font-black text-emerald-800 text-sm">{formatCurrency(row.realisasi || 0)}</p>
                                </div>
                                <div class="rounded-xl bg-amber-50 border border-amber-100 p-3">
                                    <p class="text-[10px] font-semibold uppercase tracking-wider text-amber-700">Sisa</p>
                                    <p class="mt-1 font-black text-amber-800 text-sm">{formatCurrency(row.sisaAnggaran || 0)}</p>
                                </div>
                                <div class="rounded-xl bg-violet-50 border border-violet-100 p-3">
                                    <p class="text-[10px] font-semibold uppercase tracking-wider text-violet-700">Serapan</p>
                                    <p class="mt-1 font-black text-violet-800 text-sm">{(row.persentaseSerapan || 0).toFixed(1)}%</p>
                                </div>
                            </div>

                            <!-- Progress bar + badge -->
                            <div class="mt-4">
                                <div class="h-2 overflow-hidden rounded-full bg-slate-100">
                                    <div class="h-full rounded-full transition-all duration-500"
                                        style="width: {Math.min(row.persentaseSerapan || 0, 100)}%; background-color: {colors[globalI % colors.length]}"></div>
                                </div>
                                <div class="mt-3 flex items-center justify-between gap-2">
                                    {#if row.anggaran > 0}
                                        {#if (row.persentaseSerapan || 0) > 100}
                                            <span class="inline-flex rounded-full border border-rose-200 bg-rose-50 px-2.5 py-0.5 text-[9px] font-bold capitalize tracking-wide text-rose-700">Melebihi Anggaran</span>
                                        {:else if (row.persentaseSerapan || 0) >= 80}
                                            <span class="inline-flex rounded-full border border-emerald-200 bg-emerald-50 px-2.5 py-0.5 text-[9px] font-bold capitalize tracking-wide text-emerald-700">Baik</span>
                                        {:else}
                                            <span class="inline-flex rounded-full border border-amber-200 bg-amber-50 px-2.5 py-0.5 text-[9px] font-bold capitalize tracking-wide text-amber-700">Dalam Progress</span>
                                        {/if}
                                    {:else}
                                        <span class="inline-flex rounded-full border border-slate-200 bg-slate-100 px-2.5 py-0.5 text-[9px] font-bold capitalize tracking-wide text-slate-600">Belum Diinput</span>
                                    {/if}
                                    <span class="text-[10px] text-slate-400 font-mono">MAK: {row.mak}</span>
                                </div>
                            </div>
                        </article>
                    {/each}
                </div>
            {/if}
        </div>

        <!-- Paginasi — Section 11 UI Styling Guidelines -->
        {#if totalPages > 1}
            <div class="flex items-center justify-between px-4 py-3 bg-white border border-slate-200 mt-6 rounded-xl shadow-sm">

                <!-- MOBILE: Hanya dua tombol -->
                <div class="flex flex-1 justify-between sm:hidden">
                    <button
                        on:click={() => currentPage--}
                        disabled={currentPage === 1}
                        class="relative inline-flex items-center rounded-xl border border-slate-300 bg-white px-4 py-2 text-sm font-medium text-slate-700 hover:bg-slate-50 disabled:opacity-50 disabled:cursor-not-allowed"
                    >Sebelumnya</button>
                    <button
                        on:click={() => currentPage++}
                        disabled={currentPage === totalPages}
                        class="relative inline-flex items-center rounded-xl border border-slate-300 bg-white px-4 py-2 text-sm font-medium text-slate-700 hover:bg-slate-50 disabled:opacity-50 disabled:cursor-not-allowed"
                    >Selanjutnya</button>
                </div>

                <!-- DESKTOP: Info data + navigasi lengkap -->
                <div class="hidden sm:flex sm:flex-1 sm:items-center sm:justify-between">
                    <div>
                        <p class="text-sm text-slate-700">
                            Menampilkan <span class="font-medium">{(currentPage - 1) * ITEMS_PER_PAGE + 1}</span>
                            hingga <span class="font-medium">{Math.min(currentPage * ITEMS_PER_PAGE, laporanRows.length)}</span>
                            dari <span class="font-medium">{laporanRows.length}</span> jenis pengadaan
                        </p>
                    </div>
                    <div>
                        <nav class="isolate inline-flex -space-x-px rounded-md shadow-sm" aria-label="Pagination">
                            <!-- Chevron kiri -->
                            <button
                                on:click={() => currentPage--}
                                disabled={currentPage === 1}
                                class="relative inline-flex items-center rounded-l-md px-2 py-2 text-slate-400 ring-1 ring-inset ring-slate-300 hover:bg-slate-50 focus:z-20 focus:outline-offset-0 disabled:opacity-50 disabled:cursor-not-allowed"
                            >
                                <span class="sr-only">Previous</span>
                                <svg class="h-5 w-5" viewBox="0 0 20 20" fill="currentColor" aria-hidden="true">
                                    <path fill-rule="evenodd" d="M12.79 5.23a.75.75 0 01-.02 1.06L8.832 10l3.938 3.71a.75.75 0 11-1.04 1.08l-4.5-4.25a.75.75 0 010-1.08l4.5-4.25a.75.75 0 011.06.02z" clip-rule="evenodd" />
                                </svg>
                            </button>

                            <!-- Nomor halaman dengan ellipsis cerdas -->
                            {#each Array(totalPages) as _, i}
                                {#if totalPages <= 7 || (i === 0 || i === totalPages - 1 || (i >= currentPage - 2 && i <= currentPage))}
                                    <button
                                        on:click={() => currentPage = i + 1}
                                        class="relative inline-flex items-center px-4 py-2 text-sm font-semibold
                                            {currentPage === i + 1
                                                ? 'z-10 bg-blue-600 text-white focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-blue-600'
                                                : 'text-slate-900 ring-1 ring-inset ring-slate-300 hover:bg-slate-50 focus:z-20 focus:outline-offset-0'}"
                                    >{i + 1}</button>
                                {:else if i === 1 || i === totalPages - 2}
                                    <span class="relative inline-flex items-center px-4 py-2 text-sm font-semibold text-slate-700 ring-1 ring-inset ring-slate-300 focus:outline-offset-0">...</span>
                                {/if}
                            {/each}

                            <!-- Chevron kanan -->
                            <button
                                on:click={() => currentPage++}
                                disabled={currentPage === totalPages}
                                class="relative inline-flex items-center rounded-r-md px-2 py-2 text-slate-400 ring-1 ring-inset ring-slate-300 hover:bg-slate-50 focus:z-20 focus:outline-offset-0 disabled:opacity-50 disabled:cursor-not-allowed"
                            >
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
    </section>

    <!-- Legenda -->
    <section class="grid grid-cols-1 gap-5 lg:grid-cols-3 no-print">
        <div class="rounded-2xl border border-slate-200 bg-white p-5 shadow-sm lg:col-span-2">
            <h3 class="text-lg font-black text-slate-900">Cara baca pie chart</h3>
            <div class="mt-4 grid grid-cols-1 gap-4 md:grid-cols-3">
                <div class="rounded-2xl bg-slate-50 p-4 border border-slate-100">
                    <div class="mb-3 flex h-10 w-10 items-center justify-center rounded-xl bg-emerald-600 text-white">
                        <svg xmlns="http://www.w3.org/2000/svg" class="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z" /></svg>
                    </div>
                    <p class="font-bold text-slate-900">Realisasi</p>
                    <p class="mt-1 text-sm text-slate-500">Bagian anggaran yang sudah digunakan berdasarkan Pengajuan GUP.</p>
                </div>
                <div class="rounded-2xl bg-slate-50 p-4 border border-slate-100">
                    <div class="mb-3 flex h-10 w-10 items-center justify-center rounded-xl bg-slate-500 text-white">
                        <svg xmlns="http://www.w3.org/2000/svg" class="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M3 10h18M7 15h1m4 0h1m-7 4h12a3 3 0 003-3V8a3 3 0 00-3-3H6a3 3 0 00-3 3v8a3 3 0 003 3z" /></svg>
                    </div>
                    <p class="font-bold text-slate-900">Sisa Anggaran</p>
                    <p class="mt-1 text-sm text-slate-500">Selisih Anggaran dikurangi Realisasi untuk jenis pengadaan tersebut.</p>
                </div>
                <div class="rounded-2xl bg-slate-50 p-4 border border-slate-100">
                    <div class="mb-3 flex h-10 w-10 items-center justify-center rounded-xl bg-rose-600 text-white">
                        <svg xmlns="http://www.w3.org/2000/svg" class="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-3L13.732 4c-.77-1.333-2.694-1.333-3.464 0L3.34 16c-.77 1.333.192 3 1.732 3z" /></svg>
                    </div>
                    <p class="font-bold text-slate-900">Melebihi Anggaran</p>
                    <p class="mt-1 text-sm text-slate-500">Muncul jika Realisasi lebih besar daripada Anggaran yang diinput.</p>
                </div>
            </div>
        </div>

        <div class="rounded-2xl border border-slate-200 bg-white p-5 shadow-sm">
            <h3 class="text-lg font-black text-slate-900">Data Sumber</h3>
            <dl class="mt-4 space-y-3 text-sm">
                <div class="flex items-center justify-between rounded-xl bg-slate-50 px-4 py-3 border border-slate-100">
                    <dt class="text-slate-500">Jenis Pengadaan</dt>
                    <dd class="font-black text-slate-900">{laporanSummary.totalJenisPengadaan || 0}</dd>
                </div>
                <div class="flex items-center justify-between rounded-xl bg-slate-50 px-4 py-3 border border-slate-100">
                    <dt class="text-slate-500">Transaksi GUP</dt>
                    <dd class="font-black text-slate-900">{laporanSummary.totalTransaksi || 0}</dd>
                </div>
                <div class="flex items-center justify-between rounded-xl bg-slate-50 px-4 py-3 border border-slate-100">
                    <dt class="text-slate-500">Realisasi</dt>
                    <dd class="font-black text-slate-900">{formatCurrency(laporanSummary.totalRealisasi || 0)}</dd>
                </div>
            </dl>
        </div>
    </section>
</div>

<!-- Modal Input Anggaran -->
{#if showAnggaranModal}
    <div
        class="fixed inset-0 z-[100] bg-slate-900/80 flex items-center justify-center p-4 no-print"
        role="dialog"
        aria-modal="true"
        aria-labelledby="modal-anggaran-title"
        on:click|self={closeAnggaranModal}
        on:keydown={(e) => e.key === 'Escape' && closeAnggaranModal()}
    >
        <div class="w-full max-w-lg bg-white rounded-2xl shadow-2xl animate-in fade-in zoom-in-95 duration-200">
            <div class="px-6 py-5 border-b border-slate-100 flex items-center justify-between">
                <div>
                    <h3 id="modal-anggaran-title" class="text-lg font-bold text-slate-900">Input Anggaran</h3>
                    <p class="text-sm text-slate-500 mt-0.5">{selectedRow?.jenisPengadaan || ''}</p>
                </div>
                <button
                    class="rounded-full p-2 hover:bg-slate-100 text-slate-400 hover:text-slate-600 transition-colors"
                    aria-label="Tutup modal"
                    on:click={closeAnggaranModal}
                >
                    <svg xmlns="http://www.w3.org/2000/svg" class="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" /></svg>
                </button>
            </div>

            <div class="p-6 space-y-4">
                <div class="grid grid-cols-2 gap-3 text-sm">
                    <div class="rounded-xl bg-slate-50 px-4 py-3 border border-slate-100">
                        <p class="text-xs text-slate-500">Kode Akun</p>
                        <p class="font-bold text-slate-800 mt-0.5">{selectedRow?.kodeAkun || '-'}</p>
                    </div>
                    <div class="rounded-xl bg-emerald-50 px-4 py-3 border border-emerald-100">
                        <p class="text-xs text-emerald-600">Realisasi Saat Ini</p>
                        <p class="font-bold text-emerald-800 mt-0.5">{formatCurrency(selectedRow?.realisasi || 0)}</p>
                    </div>
                </div>

                <div class="space-y-2">
                    <label for="input-anggaran" class="text-sm font-semibold text-slate-700">Nominal Anggaran (Rp)</label>
                    <input
                        id="input-anggaran"
                        type="number"
                        min="0"
                        bind:value={anggaranInput}
                        placeholder="Contoh: 50000000"
                        class="w-full rounded-xl border border-slate-300 px-4 py-3 text-lg font-black text-slate-900 focus:outline-none focus:ring-2 focus:ring-indigo-500"
                    />
                </div>

                {#if saveError}
                    <p class="text-sm text-rose-600 font-medium">{saveError}</p>
                {/if}
            </div>

            <div class="px-6 pb-6 flex justify-end gap-3 border-t border-slate-100 pt-4">
                <button on:click={closeAnggaranModal} class="px-4 py-2 text-sm font-medium border border-slate-200 text-slate-700 bg-white hover:bg-slate-50 rounded-xl transition-colors">
                    Batal
                </button>
                <button on:click={handleSaveAnggaran} disabled={savingAnggaran || !anggaranInput} class="px-5 py-2 text-sm font-bold bg-indigo-600 hover:bg-indigo-700 disabled:opacity-50 text-white rounded-xl shadow-lg transition-colors">
                    {savingAnggaran ? 'Menyimpan...' : 'Simpan Anggaran'}
                </button>
            </div>
        </div>
    </div>
{/if}
