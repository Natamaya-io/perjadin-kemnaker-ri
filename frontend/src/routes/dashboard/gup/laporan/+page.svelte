<script>
    import { formatCurrency } from '$lib/shared/utils/utils';
    import { onMount } from 'svelte';
    import { goto } from '$app/navigation';
    import Chart from 'chart.js/auto';
    import Button from '$lib/shared/ui/button/Button.svelte';
    import BaseModal from '$lib/shared/ui/base-modal/BaseModal.svelte';
    import { api } from '$lib/shared/api';
    import { userStore } from '$lib/features/auth/store';

    export let data;

    // Data dari API (reaktif)
    $: laporan = data?.laporan || { summary: {}, rows: [] };
    $: laporanRows = laporan.rows || [];
    $: laporanSummary = laporan.summary || {};
    $: currentYear = data?.year || new Date().getFullYear();

    // Filter Tahun Global
    let showYearDropdown = false;
    const startYear = 2026;
    const endYear = Math.max(new Date().getFullYear() + 1, startYear + 1);
    const yearOptions = Array.from({ length: endYear - startYear + 1 }, (_, i) => startYear + i).reverse();
    function handleYearChange(selectedYear) {
        showYearDropdown = false;
        goto(`?year=${selectedYear}&month=${currentMonth}`, { invalidateAll: true });
    }

    // Filter Bulan Global
    $: currentMonth = data?.month || 0;
    let showMonthDropdown = false;
    const monthOptions = [
        { value: 0, label: 'Sepanjang Tahun' },
        { value: 1, label: 'Januari' },
        { value: 2, label: 'Februari' },
        { value: 3, label: 'Maret' },
        { value: 4, label: 'April' },
        { value: 5, label: 'Mei' },
        { value: 6, label: 'Juni' },
        { value: 7, label: 'Juli' },
        { value: 8, label: 'Agustus' },
        { value: 9, label: 'September' },
        { value: 10, label: 'Oktober' },
        { value: 11, label: 'November' },
        { value: 12, label: 'Desember' }
    ];
    function handleMonthChange(selectedMonth) {
        showMonthDropdown = false;
        goto(`?year=${currentYear}&month=${selectedMonth}`, { invalidateAll: true });
    }

    const baseColors = [
        '#3b82f6', '#10b981', '#f59e0b', '#ef4444', '#8b5cf6', '#ec4899', '#14b8a6', '#f97316', '#6366f1', '#84cc16'
    ];
    
    function getColors(length) {
        return Array.from({length}, (_, i) => baseColors[i % baseColors.length]);
    }
    
    // Chart refs
    let totalCanvas;
    let komposisiCanvas;
    let detailCanvases = [];
    let totalChart, komposisiChart;
    let detailCharts = [];

    let showAnggaranModal = false;
    let selectedRow = null;
    let anggaranInput = '';
    let savingAnggaran = false;
    let saveError = '';
    let isAnggaranDropdownOpen = false;
    let showInfoModal = false;

    // Modal Chart Detail
    let showChartModal = false;
    let chartModalRow = null;
    let chartModalIndex = 0;
    let modalCanvas;
    let modalChart;

    // Modal Total & Komposisi
    let showTotalModal = false;
    let totalModalCanvas;
    let totalModalChart;

    let showKomposisiModal = false;
    let komposisiModalCanvas;
    let komposisiModalChart;

    function openTotalModal() {
        showTotalModal = true;
        setTimeout(buildTotalModalChart, 80);
    }

    function buildTotalModalChart() {
        if (!totalModalCanvas) return;
        if (totalModalChart) totalModalChart.destroy();
        totalModalChart = new Chart(totalModalCanvas, {
            type: 'doughnut',
            data: {
                labels: ['Realisasi', 'Sisa Anggaran'],
                datasets: [{
                    data: [laporanSummary.totalRealisasi || 0, laporanSummary.sisaAnggaran || 0],
                    backgroundColor: ['#10b981', '#e2e8f0'],
                    borderWidth: 3,
                    borderColor: '#ffffff',
                    hoverOffset: 12
                }]
            },
            options: {
                responsive: true,
                maintainAspectRatio: false,
                cutout: '70%',
                plugins: {
                    legend: {
                        display: true,
                        position: 'bottom',
                        labels: {
                            font: { size: 13, family: 'Inter, sans-serif' },
                            color: '#475569',
                            padding: 16,
                            boxWidth: 14,
                            boxHeight: 14,
                            borderRadius: 4,
                            useBorderRadius: true
                        }
                    },
                    tooltip: {
                        backgroundColor: '#0f172a',
                        padding: 14,
                        cornerRadius: 12,
                        displayColors: false,
                        titleFont: { size: 12, family: 'Inter, sans-serif' },
                        bodyFont: { size: 15, weight: 'bold', family: 'Inter, sans-serif' },
                        callbacks: { label: (ctx) => ctx.label + ': ' + formatCurrency(ctx.raw) }
                    }
                }
            }
        });
    }

    function openKomposisiModal() {
        showKomposisiModal = true;
        setTimeout(buildKomposisiModalChart, 80);
    }

    function buildKomposisiModalChart() {
        if (!komposisiModalCanvas) return;
        if (komposisiModalChart) komposisiModalChart.destroy();
        komposisiModalChart = new Chart(komposisiModalCanvas, {
            type: 'pie',
            data: {
                labels: laporanRows.map(r => r.jenisPengadaan || 'Tidak Diketahui'),
                datasets: [{
                    data: laporanRows.map(r => r.realisasi || 0),
                    backgroundColor: getColors(laporanRows.length),
                    borderWidth: 3,
                    borderColor: '#ffffff',
                    hoverOffset: 14
                }]
            },
            options: {
                responsive: true,
                maintainAspectRatio: false,
                plugins: {
                    legend: {
                        display: true,
                        position: 'right', // Posisi di kanan untuk komposisi karena itemnya banyak
                        labels: {
                            font: { size: 12, family: 'Inter, sans-serif' },
                            color: '#475569',
                            padding: 14,
                            boxWidth: 12,
                            boxHeight: 12,
                            borderRadius: 4,
                            useBorderRadius: true
                        }
                    },
                    tooltip: {
                        backgroundColor: '#0f172a',
                        padding: 14,
                        cornerRadius: 12,
                        displayColors: false,
                        titleFont: { size: 12, family: 'Inter, sans-serif' },
                        bodyFont: { size: 15, weight: 'bold', family: 'Inter, sans-serif' },
                        callbacks: {
                            title: (ctx) => laporanRows[ctx[0].dataIndex]?.jenisPengadaan || 'Tidak Diketahui',
                            label: (ctx) => 'Realisasi: ' + formatCurrency(ctx.raw)
                        }
                    }
                }
            }
        });
    }

    function openChartModal(row, globalI) {
        chartModalRow = row;
        chartModalIndex = globalI;
        showChartModal = true;
        // Render chart setelah DOM muncul
        setTimeout(buildModalChart, 80);
    }

    function buildModalChart() {
        if (!modalCanvas || !chartModalRow) return;
        if (modalChart) modalChart.destroy();
        const row = chartModalRow;
        const i = chartModalIndex;
        const sisa = Math.max(0, (row.anggaran || 0) - (row.realisasi || 0));
        modalChart = new Chart(modalCanvas, {
            type: 'pie',
            data: {
                labels: ['Realisasi', 'Sisa Anggaran'],
                datasets: [{
                    data: [row.realisasi || 0, sisa],
                    backgroundColor: [baseColors[i % baseColors.length], '#e2e8f0'],
                    borderWidth: 3,
                    borderColor: '#ffffff',
                    hoverOffset: 12
                }]
            },
            options: {
                responsive: true,
                maintainAspectRatio: false,
                plugins: {
                    legend: {
                        display: true,
                        position: 'bottom',
                        labels: {
                            font: { size: 13, family: 'Inter, sans-serif' },
                            color: '#475569',
                            padding: 16,
                            boxWidth: 14,
                            boxHeight: 14,
                            borderRadius: 4,
                            useBorderRadius: true
                        }
                    },
                    tooltip: {
                        backgroundColor: '#0f172a',
                        padding: 14,
                        cornerRadius: 12,
                        displayColors: false,
                        titleFont: { size: 12, family: 'Inter, sans-serif' },
                        bodyFont: { size: 15, weight: 'bold', family: 'Inter, sans-serif' },
                        callbacks: { label: (ctx) => ctx.label + ': ' + formatCurrency(ctx.raw) }
                    }
                }
            }
        });
    }

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
        
        const totalRealisasi = laporanSummary.totalRealisasi || 0;
        const totalSisa = laporanSummary.sisaAnggaran || 0;
        const hasData = totalRealisasi > 0 || totalSisa > 0;
        
        totalChart = new Chart(totalCanvas, {
            type: 'doughnut',
            data: {
                labels: ['Realisasi', 'Sisa Anggaran'],
                datasets: [{
                    data: hasData ? [totalRealisasi, totalSisa] : [0, 1],
                    backgroundColor: hasData ? ['#10b981', '#e2e8f0'] : ['#10b981', '#f1f5f9'],
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
        
        const realisasiData = laporanRows.map(r => r.realisasi || 0);
        const hasData = realisasiData.some(val => val > 0);

        komposisiChart = new Chart(komposisiCanvas, {
            type: 'pie',
            data: {
                labels: hasData ? laporanRows.map(r => r.jenisPengadaan || 'Tidak Diketahui') : ['Belum Ada Realisasi'],
                datasets: [{
                    data: hasData ? realisasiData : [1],
                    backgroundColor: hasData ? getColors(laporanRows.length) : ['#f1f5f9'],
                    borderWidth: 2,
                    borderColor: '#ffffff',
                    hoverOffset: 10
                }]
            },
            options: {
                ...commonOptions,
                plugins: { ...commonOptions.plugins, tooltip: { ...commonOptions.plugins.tooltip,
                    callbacks: {
                        title: (ctx) => hasData ? (laporanRows[ctx[0].dataIndex]?.jenisPengadaan || 'Tidak Diketahui') : 'Tidak Diketahui',
                        label: (ctx) => 'Realisasi: ' + formatCurrency(hasData ? ctx.raw : 0)
                    }
                }}
            }
        });
    }

    function buildDetailCharts() {
        detailCharts.forEach(c => c?.destroy());
        detailCharts = [];
        pagedRows.forEach((row, i) => {
            const globalI = globalOffset + i;
            if (!detailCanvases[globalI]) return;
            const sisa = Math.max(0, (row.anggaran || 0) - (row.realisasi || 0));
            const realisasi = row.realisasi || 0;
            const hasData = realisasi > 0 || sisa > 0;
            
            const chart = new Chart(detailCanvases[globalI], {
                type: 'pie',
                data: {
                    labels: ['Realisasi', 'Sisa Anggaran'],
                    datasets: [{
                        data: hasData ? [realisasi, sisa] : [0, 1],
                        backgroundColor: hasData ? [baseColors[globalI % baseColors.length], '#e2e8f0'] : [baseColors[globalI % baseColors.length], '#f1f5f9'],
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


    let selectedModalMonth = 0;
    let showModalMonthDropdown = false;
    function openAnggaranModal(row = null) {
        selectedRow = row;
        anggaranInput = row && row.anggaran > 0 ? Number(row.anggaran).toLocaleString('id-ID') : '';
        selectedModalMonth = currentMonth; // Default ke filter bulan saat ini
        saveError = '';
        showAnggaranModal = true;
    }

    function closeAnggaranModal() {
        showAnggaranModal = false;
        selectedRow = null;
        anggaranInput = '';
        isAnggaranDropdownOpen = false;
        showModalMonthDropdown = false;
    }

    async function handleSaveAnggaran() {
        if (!selectedRow || !anggaranInput) return;
        savingAnggaran = true;
        saveError = '';
        try {
            await api.saveGupBudget({
                procurementTypeId: selectedRow.procurementTypeId,
                year: currentYear,
                monthNumber: selectedModalMonth,
                amount: parseFloat(anggaranInput.replace(/\./g, ''))
            });
            // Reload data
            const fresh = await api.getGupLaporan(currentYear, currentMonth);
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
    <!-- Header halaman dan tombol aksi -->
    <div class="flex flex-col md:flex-row md:items-center justify-between gap-4 bg-white p-6 rounded-2xl shadow-sm border border-slate-100 no-print">
        <div>
            <h1 class="text-2xl text-slate-800 tracking-tight flex items-center gap-2">
                Laporan dan Rekapitulasi
                <button type="button" title="Informasi Global & Legenda" on:click={() => showInfoModal = true} class="text-blue-500 hover:text-blue-700 bg-blue-50 hover:bg-blue-100 rounded-full p-1 transition-colors">
                    <svg xmlns="http://www.w3.org/2000/svg" class="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13 16h-1v-4h-1m1-4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
                    </svg>
                </button>
            </h1>
            <p class="text-sm text-slate-500 mt-1">Pantau Anggaran, Realisasi, Sisa Anggaran, dan Persentase Serapan.</p>
        </div>
        <div class="flex items-center gap-3 no-print">
            <!-- Dropdown Tahun -->
            <div class="relative inline-block w-full sm:w-40">
                <button 
                    type="button" 
                    on:click={() => { showYearDropdown = !showYearDropdown; showMonthDropdown = false; }}
                    class="flex w-full items-center justify-between rounded-xl border border-slate-200 bg-white px-4 py-2 text-sm font-medium text-slate-700 shadow-sm hover:bg-slate-50 focus:outline-none focus:ring-2 focus:ring-slate-100 transition-all cursor-pointer h-10"
                >
                    <span>Tahun {currentYear}</span>
                    <svg class="h-4 w-4 text-slate-500" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 9l-7 7-7-7" />
                    </svg>
                </button>
                
                {#if showYearDropdown}
                    <!-- svelte-ignore a11y-click-events-have-key-events -->
                    <!-- svelte-ignore a11y-no-static-element-interactions -->
                    <div class="fixed inset-0 z-40" on:click={() => showYearDropdown = false}></div>
                    
                    <div class="absolute right-0 z-50 mt-2 w-full origin-top-right rounded-xl border border-slate-100 bg-white shadow-lg overflow-hidden animate-in fade-in zoom-in-95 duration-100">
                        <ul class="max-h-60 overflow-y-auto custom-scrollbar py-1">
                            {#each yearOptions as yr}
                                <!-- svelte-ignore a11y-click-events-have-key-events -->
                                <!-- svelte-ignore a11y-no-noninteractive-element-interactions -->
                                <li 
                                    class="relative cursor-pointer select-none py-2.5 pl-4 pr-4 text-sm transition-colors {currentYear === yr ? 'font-semibold text-indigo-700 bg-indigo-50/50 hover:bg-indigo-50 flex items-center justify-between' : 'text-slate-700 hover:bg-slate-50 hover:text-indigo-600'}"
                                    on:click={() => { showYearDropdown = false; handleYearChange(yr); }}
                                >
                                    <span>Tahun {yr}</span>
                                    {#if currentYear === yr}
                                        <svg class="h-4 w-4 text-indigo-600 shrink-0" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 13l4 4L19 7" /></svg>
                                    {/if}
                                </li>
                            {/each}
                        </ul>
                    </div>
                {/if}
            </div>

            <!-- Dropdown Bulan -->
            <div class="relative inline-block w-full sm:w-40">
                <button 
                    type="button" 
                    on:click={() => { showMonthDropdown = !showMonthDropdown; showYearDropdown = false; }}
                    class="flex w-full items-center justify-between rounded-xl border border-slate-200 bg-white px-4 py-2 text-sm font-medium text-slate-700 shadow-sm hover:bg-slate-50 focus:outline-none focus:ring-2 focus:ring-slate-100 transition-all cursor-pointer h-10"
                >
                    <span class="truncate">{monthOptions.find(m => m.value === currentMonth)?.label || 'Bulan'}</span>
                    <svg class="h-4 w-4 text-slate-500" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 9l-7 7-7-7" />
                    </svg>
                </button>
                
                {#if showMonthDropdown}
                    <!-- svelte-ignore a11y-click-events-have-key-events -->
                    <!-- svelte-ignore a11y-no-static-element-interactions -->
                    <div class="fixed inset-0 z-40" on:click={() => showMonthDropdown = false}></div>
                    
                    <div class="absolute right-0 z-50 mt-2 w-full min-w-[12rem] origin-top-right rounded-xl border border-slate-100 bg-white shadow-lg overflow-hidden animate-in fade-in zoom-in-95 duration-100">
                        <ul class="max-h-60 overflow-y-auto custom-scrollbar py-1">
                            {#each monthOptions as opt}
                                <!-- svelte-ignore a11y-click-events-have-key-events -->
                                <!-- svelte-ignore a11y-no-noninteractive-element-interactions -->
                                <li 
                                    class="relative cursor-pointer select-none py-2.5 pl-4 pr-4 text-sm transition-colors {currentMonth === opt.value ? 'font-semibold text-indigo-700 bg-indigo-50/50 hover:bg-indigo-50 flex items-center justify-between' : 'text-slate-700 hover:bg-slate-50 hover:text-indigo-600'}"
                                    on:click={() => { showMonthDropdown = false; handleMonthChange(opt.value); }}
                                >
                                    <span class="truncate pr-2">{opt.label}</span>
                                    {#if currentMonth === opt.value}
                                        <svg class="h-4 w-4 text-indigo-600 shrink-0" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 13l4 4L19 7" /></svg>
                                    {/if}
                                </li>
                            {/each}
                        </ul>
                    </div>
                {/if}
            </div>

            {#if $userStore?.role !== 'kasubag'}
            <Button variant="warning" class="gap-2 h-10" on:click={() => openAnggaranModal(null)}>
                <svg xmlns="http://www.w3.org/2000/svg" class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15.232 5.232l3.536 3.536m-2.036-5.036a2.5 2.5 0 113.536 3.536L6.5 21.036H3v-3.572L16.732 3.732z" />
                </svg>
                Input Anggaran
            </Button>
            {/if}
            <Button variant="outline" class="gap-2 h-10" on:click={() => window.print()}>
                <svg xmlns="http://www.w3.org/2000/svg" class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M17 17h2a2 2 0 002-2v-4a2 2 0 00-2-2H5a2 2 0 00-2 2v4a2 2 0 002 2h2m2 4h6a2 2 0 002-2v-4a2 2 0 00-2-2H9a2 2 0 00-2 2v4a2 2 0 002 2zm8-12V5a2 2 0 00-2-2H9a2 2 0 00-2 2v4h10z" />
                </svg>
                Cetak
            </Button>
        </div>
    </div>

    <!-- Overview -->
    <section class="grid grid-cols-1 gap-6 xl:grid-cols-5">
        <article class="overflow-hidden rounded-2xl border border-slate-100 bg-white p-5 shadow-sm xl:col-span-2">
            <div class="flex flex-col gap-4 sm:flex-row sm:items-start sm:justify-between">
                <div>
                    <h3 class="text-lg text-slate-800">Total Serapan Anggaran</h3>
                    <p class="mt-1 text-sm text-slate-500">Komposisi total Realisasi dan Sisa Anggaran.</p>
                </div>
                <div class="flex items-center gap-2">
                    <button
                        type="button"
                        title="Lihat chart lebih besar"
                        class="text-slate-500 hover:text-indigo-700 bg-slate-50 hover:bg-indigo-50 border border-slate-200 hover:border-indigo-200 p-2 rounded-xl transition-colors shrink-0 no-print self-stretch flex items-center justify-center"
                        on:click={openTotalModal}
                    >
                        <svg xmlns="http://www.w3.org/2000/svg" class="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 8V4m0 0h4M4 4l5 5m11-1V4m0 0h-4m4 0l-5 5M4 16v4m0 0h4m-4 0l5-5m11 5l-5-5m5 5v-4m0 4h-4" />
                        </svg>
                    </button>
                </div>
            </div>

            <div class="mt-5 grid grid-cols-1 gap-3 sm:grid-cols-2">
                <div class="rounded-2xl border border-slate-100 bg-slate-50 p-4">
                    <p class="text-xs font-semibold text-slate-500">Total Anggaran</p>
                    <h3 class="mt-1 text-xl font-black text-slate-900">{formatCurrency(laporanSummary.totalAnggaran || 0)}</h3>
                </div>
                <div class="rounded-2xl border border-emerald-100 bg-emerald-50 p-4">
                    <p class="text-xs font-semibold text-emerald-700">Total Realisasi</p>
                    <h3 class="mt-1 text-xl font-black text-emerald-800">{formatCurrency(laporanSummary.totalRealisasi || 0)}</h3>
                </div>
                <div class="rounded-2xl border border-amber-100 bg-amber-50 p-4 sm:col-span-2">
                    <p class="text-xs font-semibold text-amber-700">Sisa Anggaran</p>
                    <h3 class="mt-1 text-xl font-black text-amber-800">{formatCurrency(laporanSummary.sisaAnggaran || 0)}</h3>
                </div>
            </div>

            <div class="mt-4 relative h-64 rounded-3xl bg-slate-50 p-4">
                <canvas bind:this={totalCanvas}></canvas>
            </div>
        </article>

        <article class="overflow-hidden rounded-2xl border border-slate-100 bg-white p-5 shadow-sm xl:col-span-3 flex flex-col">
            <div class="flex flex-col gap-4 sm:flex-row sm:items-start sm:justify-between">
                <div>
                    <h3 class="text-lg text-slate-800 flex items-center gap-2">
                        Realisasi per Jenis Pengadaan
                        <button type="button" title="Informasi Laporan" on:click={() => showInfoModal = true} class="text-blue-500 hover:text-blue-700 bg-blue-50 hover:bg-blue-100 rounded-full p-1 transition-colors">
                            <svg xmlns="http://www.w3.org/2000/svg" class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13 16h-1v-4h-1m1-4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
                            </svg>
                        </button>
                    </h3>
                    <p class="mt-1 text-sm text-slate-500">Pie chart ini ditarik dari data Realisasi pada menu Pengajuan GUP.</p>
                </div>
                <div class="flex items-center gap-2">
                    <div class="inline-flex items-center gap-2 rounded-2xl bg-sky-50 px-4 py-3 text-sm font-bold text-sky-700 border border-sky-100">
                        <svg xmlns="http://www.w3.org/2000/svg" class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 11H5m14 0a2 2 0 012 2v6a2 2 0 01-2 2H5a2 2 0 01-2-2v-6a2 2 0 012-2m14 0V9a2 2 0 00-2-2M5 11V9a2 2 0 002-2m0 0V5a2 2 0 012-2h6a2 2 0 012 2v2M7 7h10" />
                        </svg>
                        <span>{laporanSummary.totalJenisPengadaan || 0}</span> Jenis
                    </div>
                    <button
                        type="button"
                        title="Lihat chart lebih besar"
                        class="text-slate-500 hover:text-indigo-700 bg-slate-50 hover:bg-indigo-50 border border-slate-200 hover:border-indigo-200 p-2 rounded-xl transition-colors shrink-0 no-print self-stretch flex items-center justify-center"
                        on:click={openKomposisiModal}
                    >
                        <svg xmlns="http://www.w3.org/2000/svg" class="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 8V4m0 0h4M4 4l5 5m11-1V4m0 0h-4m4 0l-5 5M4 16v4m0 0h4m-4 0l5-5m11 5l-5-5m5 5v-4m0 4h-4" />
                        </svg>
                    </button>
                </div>
            </div>

            <div class="mt-5 grid grid-cols-1 gap-3 sm:grid-cols-2">
                <div class="rounded-2xl border border-slate-100 bg-slate-50 p-4 flex flex-col justify-between">
                    <p class="text-xs font-semibold text-slate-500">Jumlah Transaksi GUP</p>
                    <div class="mt-1 flex items-center justify-between">
                        <h3 class="text-2xl font-black text-slate-900">{laporanSummary.totalTransaksi || 0}</h3>
                        <a
                            href="/dashboard/gup/pengajuan"
                            title="Lihat Rincian"
                            class="text-indigo-600 hover:text-indigo-800 bg-indigo-50 hover:bg-indigo-100 p-2 rounded-lg transition-colors border border-indigo-100 flex items-center justify-center shrink-0"
                        >
                            <svg xmlns="http://www.w3.org/2000/svg" class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 5l7 7-7 7" />
                            </svg>
                        </a>
                    </div>
                </div>
                <div class="rounded-2xl border border-emerald-100 bg-emerald-50 p-4 flex flex-col justify-between">
                    <p class="text-xs font-semibold text-emerald-700">Nilai Pengajuan</p>
                    <h3 class="mt-1 text-2xl font-black text-emerald-800">{formatCurrency(laporanSummary.totalRealisasi || 0)}</h3>
                </div>
            </div>

            <div class="mt-4 relative rounded-3xl bg-slate-50 p-4 flex-1 min-h-[16rem]">
                <canvas bind:this={komposisiCanvas}></canvas>
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
                {#if $userStore?.role !== 'kasubag'}
                <Button variant="default" class="w-full sm:w-auto flex items-center justify-center gap-2" on:click={() => openAnggaranModal(null)}>
                    <svg xmlns="http://www.w3.org/2000/svg" class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4v16m8-8H4" />
                    </svg>
                    Input Pagu Anggaran
                </Button>
                {/if}
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
                <div class="grid grid-cols-1 gap-4 sm:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4">
                    {#each pagedRows as row, i}
                        {@const globalI = globalOffset + i}
                        <article class="bg-white p-4 rounded-2xl border border-slate-200 hover:border-blue-300 shadow-sm hover:shadow-lg transition-all duration-300 flex flex-col relative overflow-hidden">
                            <!-- Garis atas sesuai Section 5 guidelines -->
                            <div class="absolute top-0 left-0 w-full h-1.5" style="background-color: {baseColors[globalI % baseColors.length]}"></div>

                            <div class="flex items-start gap-3 mb-4 pb-3 border-b border-slate-100/80">
                                <div class="p-2 rounded-xl shadow-sm border border-slate-100 shrink-0" style="background-color: {baseColors[globalI % baseColors.length]}20; color: {baseColors[globalI % baseColors.length]}">
                                    <svg xmlns="http://www.w3.org/2000/svg" class="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                                        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M11 3.055A9.001 9.001 0 1020.945 13H11V3.055z" />
                                    </svg>
                                </div>
                                <div class="min-w-0 flex-1">
                                    <p class="text-[10px] font-bold uppercase tracking-[0.2em] text-slate-400">{row.kodeAkun}</p>
                                    <h3 class="mt-0.5 text-sm font-black leading-snug text-slate-900 truncate">{row.jenisPengadaan}</h3>
                                    <p class="text-[10px] text-slate-400 mt-0.5">{row.jumlahTransaksi} transaksi</p>
                                </div>
                                <!-- Tombol expand chart — Section 4.2 Flat & Clean -->
                                <button
                                    type="button"
                                    title="Lihat chart lebih besar"
                                    class="text-slate-500 hover:text-indigo-700 bg-slate-50 hover:bg-indigo-50 border border-slate-200 hover:border-indigo-200 p-1.5 rounded transition-colors shrink-0"
                                    on:click={() => openChartModal(row, globalI)}
                                >
                                    <svg class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                                        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 8V4m0 0h4M4 4l5 5m11-1V4m0 0h-4m4 0l-5 5M4 16v4m0 0h4m-4 0l5-5m11 5l-5-5m5 5v-4m0 4h-4" />
                                    </svg>
                                </button>
                            </div>

                            <!-- Pie chart canvas -->
                            <div class="relative h-36 rounded-xl bg-slate-50 p-2">
                                <canvas bind:this={detailCanvases[globalI]}></canvas>
                            </div>

                            <!-- Stats grid -->
                            <div class="mt-4 grid grid-cols-2 gap-2 text-sm">
                                <div class="rounded-xl bg-slate-50 border border-slate-100 p-2.5 flex flex-col justify-center">
                                    <p class="text-[9px] font-bold uppercase tracking-wider text-slate-500">Anggaran</p>
                                    <p class="mt-0.5 font-black text-slate-900 text-sm truncate">{formatCurrency(row.anggaran || 0)}</p>
                                </div>
                                <div class="rounded-xl bg-emerald-50 border border-emerald-100 p-2.5 flex flex-col justify-center">
                                    <p class="text-[9px] font-bold uppercase tracking-wider text-emerald-700">Realisasi</p>
                                    <p class="mt-0.5 font-black text-emerald-800 text-sm truncate">{formatCurrency(row.realisasi || 0)}</p>
                                </div>
                                <div class="rounded-xl bg-amber-50 border border-amber-100 p-2.5 flex flex-col justify-center">
                                    <p class="text-[9px] font-bold uppercase tracking-wider text-amber-700">Sisa</p>
                                    <p class="mt-0.5 font-black text-amber-800 text-sm truncate">{formatCurrency(row.sisaAnggaran || 0)}</p>
                                </div>
                                <div class="rounded-xl bg-violet-50 border border-violet-100 p-2.5 flex flex-col justify-center">
                                    <p class="text-[9px] font-bold uppercase tracking-wider text-violet-700">Serapan</p>
                                    <p class="mt-0.5 font-black text-violet-800 text-sm truncate">{(row.persentaseSerapan || 0).toFixed(1)}%</p>
                                </div>
                            </div>

                            <!-- Progress bar + badge -->
                            <div class="mt-4">
                                <div class="h-2 overflow-hidden rounded-full bg-slate-100">
                                    <div class="h-full rounded-full transition-all duration-500"
                                        style="width: {Math.min(row.persentaseSerapan || 0, 100)}%; background-color: {baseColors[globalI % baseColors.length]}"></div>
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
                    <Button variant="outline" size="sm" disabled={currentPage === 1} on:click={() => currentPage--}>
                        Sebelumnya
                    </Button>
                    <Button variant="outline" size="sm" disabled={currentPage === totalPages} on:click={() => currentPage++}>
                        Selanjutnya
                    </Button>
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

    <!-- Legenda & Data Sumber telah dipindahkan ke modal Informasi Global -->
</div>

<!-- Modal Informasi Global -->
<BaseModal
    bind:open={showInfoModal}
    maxWidth="max-w-lg"
    on:close={() => showInfoModal = false}
>
    <div slot="header">
        <h2 class="text-lg font-bold text-slate-800">Informasi Global & Legenda</h2>
    </div>

    <div slot="body">
        <div class="space-y-4 text-sm text-slate-600">
            <p><strong>Cara Baca Grafik & Indikator Warna:</strong></p>
            <ul class="list-disc pl-5 space-y-2 text-slate-600">
                <li><strong class="text-emerald-700">Realisasi (Hijau):</strong> Bagian anggaran yang sudah digunakan berdasarkan Pengajuan GUP.</li>
                <li><strong class="text-slate-500">Sisa Anggaran (Abu-abu):</strong> Selisih Anggaran dikurangi Realisasi untuk jenis pengadaan tersebut.</li>
                <li><strong class="text-rose-700">Melebihi Anggaran (Merah):</strong> Muncul jika Realisasi lebih besar daripada Anggaran yang diinput.</li>
            </ul>

            <p><strong>Ringkasan Data Sumber:</strong></p>
            <ul class="list-disc pl-5 space-y-2 text-slate-600">
                <li>Total Jenis Pengadaan: <strong>{laporanSummary.totalJenisPengadaan || 0}</strong></li>
                <li>Jumlah Transaksi GUP: <strong>{laporanSummary.totalTransaksi || 0}</strong></li>
                <li>Total Realisasi: <strong>{formatCurrency(laporanSummary.totalRealisasi || 0)}</strong></li>
            </ul>

            <div class="mt-2 bg-amber-50 border border-amber-100 p-3 rounded-lg text-amber-800">
                <strong class="block mb-1">Catatan Perhitungan Laporan:</strong>
                Jenis Pengadaan dengan realisasi nol tetap muncul pada daftar laporan (kartu detail) agar ketersediaan pagu anggarannya bisa dipantau, tetapi jenis tersebut <strong>tidak akan mendominasi grafik pie komposisi</strong> untuk menjaga akurasi representasi pengeluaran riil.
            </div>
        </div>
    </div>

    <div slot="footer" class="flex justify-end w-full">
        <Button variant="default" on:click={() => showInfoModal = false}>Mengerti</Button>
    </div>
</BaseModal>


<!-- Modal Input Anggaran -->
{#if showAnggaranModal}
    <div
        class="fixed inset-0 z-[100] bg-slate-900/80 flex items-center justify-center p-4 no-print"
        role="dialog"
        aria-modal="true"
        tabindex="-1"
        aria-labelledby="modal-anggaran-title"
        on:click|self={closeAnggaranModal}
        on:keydown={(e) => e.key === 'Escape' && closeAnggaranModal()}
    >
        <div class="w-full max-w-lg bg-white rounded-2xl shadow-2xl animate-in fade-in zoom-in-95 duration-200">
            <div class="px-6 py-5 border-b border-slate-100 flex items-center justify-between">
                <div>
                    <h3 id="modal-anggaran-title" class="text-lg font-bold text-slate-900">Input Anggaran</h3>
                    <p class="text-sm text-slate-500 mt-0.5">Atur pagu anggaran untuk jenis pengadaan tertentu</p>
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
                <div class="space-y-2 relative">
                    <span class="block text-sm font-semibold text-slate-700">Jenis Pengadaan</span>
                    <div class="relative w-full">
                        <button 
                            type="button" 
                            on:click={() => isAnggaranDropdownOpen = !isAnggaranDropdownOpen}
                            class="flex w-full items-center justify-between rounded-xl border border-slate-200 bg-slate-50 px-4 py-2.5 text-sm text-slate-700 focus:outline-none focus:ring-2 focus:ring-indigo-500 focus:bg-white transition-colors"
                        >
                            <span class="truncate">{selectedRow ? `${selectedRow.kodeAkun} - ${selectedRow.jenisPengadaan}` : 'Pilih Item...'}</span>
                            <svg class="h-4 w-4 text-slate-400" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 9l-7 7-7-7" />
                            </svg>
                        </button>
                        
                        {#if isAnggaranDropdownOpen}
                            <!-- svelte-ignore a11y-click-events-have-key-events -->
                            <!-- svelte-ignore a11y-no-static-element-interactions -->
                            <div class="fixed inset-0 z-40" on:click={() => isAnggaranDropdownOpen = false}></div>
                            
                            <div class="absolute z-50 mt-2 w-full origin-top-right rounded-xl border border-slate-100 bg-white shadow-lg overflow-hidden animate-in fade-in zoom-in-95 duration-100">
                                <ul class="max-h-60 overflow-y-auto custom-scrollbar py-1">
                                    {#each laporanRows as row}
                                        {@const isSelected = selectedRow?.procurementTypeId === row.procurementTypeId}
                                        <!-- svelte-ignore a11y-click-events-have-key-events -->
                                        <!-- svelte-ignore a11y-no-noninteractive-element-interactions -->
                                        <li 
                                            class="relative cursor-pointer select-none py-2.5 pl-4 pr-4 text-sm transition-colors {isSelected ? 'font-semibold text-indigo-700 bg-indigo-50/50 hover:bg-indigo-50 flex items-center justify-between' : 'text-slate-700 hover:bg-slate-50 hover:text-indigo-600'}"
                                            on:click={() => {
                                                selectedRow = row;
                                                // Jika memilih row dari modal, tidak mengambil otomatis pagu, karena pagu tergantung bulan
                                                // Tapi agar simple, biarkan input kosong atau tetap apa adanya
                                                anggaranInput = '';
                                                saveError = '';
                                                isAnggaranDropdownOpen = false;
                                            }}
                                        >
                                            <span class="truncate pr-2">{row.kodeAkun} - {row.jenisPengadaan}</span>
                                            {#if isSelected}
                                                <svg class="h-4 w-4 text-indigo-600 shrink-0" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                                                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 13l4 4L19 7" />
                                                </svg>
                                            {/if}
                                        </li>
                                    {/each}
                                </ul>
                            </div>
                        {/if}
                    </div>
                </div>

                {#if selectedRow}
                    {@const numericAnggaran = parseInt(String(anggaranInput).replace(/\D/g, ''), 10) || 0}
                    {@const numericRealisasi = selectedRow.realisasi || 0}
                    {@const sisaAnggaran = numericAnggaran - numericRealisasi}
                    {@const persentase = numericAnggaran > 0 ? ((numericRealisasi / numericAnggaran) * 100).toFixed(2) : 0}
                    
                    <div class="grid grid-cols-2 gap-3 text-sm animate-in fade-in slide-in-from-top-2 duration-300">
                        <div class="rounded-xl bg-slate-50 px-4 py-3 border border-slate-100">
                            <p class="text-xs text-slate-500">Kode Akun</p>
                            <p class="font-bold text-slate-800 mt-0.5">{selectedRow.kodeAkun || '-'}</p>
                        </div>
                        <div class="rounded-xl bg-emerald-50 px-4 py-3 border border-emerald-100">
                            <p class="text-xs text-emerald-600">Realisasi Saat Ini</p>
                            <p class="font-bold text-emerald-800 mt-0.5">{formatCurrency(numericRealisasi)}</p>
                        </div>
                        <div class="rounded-xl bg-amber-50 px-4 py-3 border border-amber-100">
                            <p class="text-xs text-amber-600">Sisa Anggaran</p>
                            <p class="font-bold text-amber-800 mt-0.5">
                                {sisaAnggaran < 0 ? '-' : ''}{formatCurrency(Math.abs(sisaAnggaran))}
                            </p>
                        </div>
                        <div class="rounded-xl bg-sky-50 px-4 py-3 border border-sky-100">
                            <p class="text-xs text-sky-600">Persentase Terpakai</p>
                            <p class="font-bold text-sky-800 mt-0.5">{persentase}%</p>
                        </div>
                    </div>
                {/if}

                <div class="space-y-2">
                    <label for="input-bulan" class="text-sm font-semibold text-slate-700">Untuk Bulan</label>
                    <div class="relative w-full">
                        <button 
                            type="button" 
                            id="input-bulan"
                            on:click={() => showModalMonthDropdown = !showModalMonthDropdown}
                            class="flex w-full items-center justify-between appearance-none rounded-xl border border-slate-200 bg-slate-50 px-4 py-2.5 text-sm font-medium text-slate-700 hover:border-slate-300 focus:outline-none focus:ring-2 focus:ring-indigo-500/30 focus:border-indigo-400 focus:bg-white transition-colors cursor-pointer"
                        >
                            <span class="truncate">{monthOptions.find(m => m.value === selectedModalMonth)?.label || 'Pilih Bulan'}</span>
                            <svg class="h-4 w-4 text-slate-400" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 9l-7 7-7-7" />
                            </svg>
                        </button>
                        
                        {#if showModalMonthDropdown}
                            <!-- svelte-ignore a11y-click-events-have-key-events -->
                            <!-- svelte-ignore a11y-no-static-element-interactions -->
                            <div class="fixed inset-0 z-40" on:click={() => showModalMonthDropdown = false}></div>
                            <div class="absolute left-0 right-0 z-50 mt-2 origin-top rounded-xl border border-slate-100 bg-white shadow-xl overflow-hidden animate-in fade-in zoom-in-95 duration-100 ring-1 ring-black/5">
                                <ul class="max-h-60 overflow-y-auto custom-scrollbar py-1.5">
                                    {#each monthOptions as opt}
                                        <!-- svelte-ignore a11y-click-events-have-key-events -->
                                        <!-- svelte-ignore a11y-no-noninteractive-element-interactions -->
                                        <li 
                                            class="relative cursor-pointer select-none py-2 px-4 text-sm transition-colors {selectedModalMonth === opt.value ? 'font-semibold text-indigo-700 bg-indigo-50/70 hover:bg-indigo-50 flex items-center justify-between' : 'text-slate-700 hover:bg-slate-50 hover:text-indigo-600'}"
                                            on:click={() => { selectedModalMonth = opt.value; showModalMonthDropdown = false; }}
                                        >
                                            <span>{opt.label}</span>
                                            {#if selectedModalMonth === opt.value}
                                                <svg class="h-4 w-4 text-indigo-600 shrink-0" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 13l4 4L19 7" /></svg>
                                            {/if}
                                        </li>
                                    {/each}
                                </ul>
                            </div>
                        {/if}
                    </div>
                </div>

                <div class="space-y-2">
                    <label for="input-anggaran" class="text-sm font-semibold text-slate-700">Nominal Anggaran (Rp)</label>
                    <input
                        id="input-anggaran"
                        type="text"
                        value={anggaranInput}
                        on:input={(e) => {
                            let val = e.target.value.replace(/\D/g, '');
                            anggaranInput = val ? parseInt(val, 10).toLocaleString('id-ID') : '';
                        }}
                        disabled={!selectedRow}
                        placeholder={selectedRow ? "Contoh: 50.000.000" : "Pilih jenis pengadaan dahulu"}
                        class="w-full rounded-xl border border-slate-300 px-4 py-3 text-lg font-black text-slate-900 focus:outline-none focus:ring-2 focus:ring-indigo-500 disabled:opacity-50 disabled:bg-slate-50"
                    />
                </div>

                {#if saveError}
                    <p class="text-sm text-rose-600 font-medium">{saveError}</p>
                {/if}
            </div>

            <div class="px-6 pb-6 flex justify-end gap-3 border-t border-slate-100 pt-4">
                <Button variant="outline" on:click={closeAnggaranModal}>
                    Batal
                </Button>
                <Button variant="default" on:click={handleSaveAnggaran} disabled={savingAnggaran || !anggaranInput}>
                    {savingAnggaran ? 'Menyimpan...' : 'Simpan Anggaran'}
                </Button>
            </div>
        </div>
    </div>
{/if}

<!-- Modal Chart Detail — Section 9 BaseModal -->
<BaseModal
    bind:open={showChartModal}
    maxWidth="max-w-6xl"
    on:close={() => { if (modalChart) { modalChart.destroy(); modalChart = null; } chartModalRow = null; }}
>
    <!-- Header — Section 9.2 -->
    <svelte:fragment slot="header">
        {#if chartModalRow}
            <div>
                <p class="text-[10px] font-bold uppercase tracking-[0.2em] text-slate-400 mb-0.5">
                    {chartModalRow.kodeAkun} · MAK {chartModalRow.mak}
                </p>
                <h2 class="text-xl font-bold text-slate-800 leading-tight">
                    {chartModalRow.jenisPengadaan}
                </h2>
                <p class="text-sm text-slate-500 mt-1">Visualisasi serapan anggaran per jenis pengadaan tahun {currentYear}</p>
            </div>
        {/if}
    </svelte:fragment>

    <!-- Body — Section 9.2 (bg-slate-50/50 diatur BaseModal) -->
    <svelte:fragment slot="body">
        {#if chartModalRow}
            {@const sisa = Math.max(0, (chartModalRow.anggaran || 0) - (chartModalRow.realisasi || 0))}
            {@const pct = chartModalRow.persentaseSerapan || 0}
            {@const ac = baseColors[chartModalIndex % baseColors.length]}
            {@const ac2 = baseColors[(chartModalIndex + 1) % baseColors.length]}

            <div class="grid grid-cols-1 gap-5 lg:grid-cols-5">

                <!-- ===== KIRI: Chart canvas — lg:col-span-3 ===== -->
                <div class="lg:col-span-3">
                    <!-- Kartu chart — Section 5 pattern -->
                    <div class="bg-white p-5 md:p-6 rounded-2xl border border-slate-200 hover:border-blue-300 shadow-sm hover:shadow-lg transition-all duration-300 flex flex-col relative overflow-hidden group h-full">
                        <!-- Garis Atas — Section 5 -->
                        <div class="absolute top-0 left-0 w-full h-1.5 bg-blue-500"></div>

                        <!-- Header Kartu — Section 5 -->
                        <div class="flex items-center gap-3 mb-4 pb-3 border-b border-slate-100/80">
                            <div class="p-2.5 bg-blue-50 text-blue-600 rounded-xl shadow-sm border border-blue-100">
                                <svg xmlns="http://www.w3.org/2000/svg" class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M11 3.055A9.001 9.001 0 1020.945 13H11V3.055z" />
                                </svg>
                            </div>
                            <h4 class="text-xs font-bold text-slate-700 uppercase tracking-widest">Visualisasi Pie Chart</h4>
                        </div>

                        <!-- Canvas chart -->
                        <div class="relative flex-1 min-h-[400px]">
                            <canvas bind:this={modalCanvas}></canvas>
                        </div>
                    </div>
                </div>

                <!-- ===== KANAN: Info cards — lg:col-span-2 ===== -->
                <div class="lg:col-span-2 flex flex-col gap-4">

                    <!-- Kartu 1: Realisasi — Section 5, gradasi emerald -->
                    <div class="bg-white p-5 md:p-6 rounded-2xl border border-slate-200 hover:border-blue-300 shadow-sm hover:shadow-lg transition-all duration-300 flex flex-col relative overflow-hidden group">
                        <div class="absolute top-0 left-0 w-full h-1.5 bg-emerald-500"></div>

                        <div class="flex items-center gap-3 mb-4 pb-3 border-b border-slate-100/80">
                            <div class="p-2.5 bg-emerald-50 text-emerald-600 rounded-xl shadow-sm border border-emerald-100">
                                <svg xmlns="http://www.w3.org/2000/svg" class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z" />
                                </svg>
                            </div>
                            <div class="flex items-center gap-2">
                                <span class="inline-block h-3 w-3 rounded-sm shrink-0" style="background-color: {ac}"></span>
                                <h4 class="text-xs font-bold text-slate-700 uppercase tracking-widest">Realisasi</h4>
                            </div>
                        </div>

                        <div class="space-y-3">
                            <p class="text-2xl font-black text-emerald-800">{formatCurrency(chartModalRow.realisasi || 0)}</p>
                            <p class="text-xs text-slate-400 leading-relaxed">Anggaran yang telah terealisasi berdasarkan total pengajuan GUP yang disetujui pada tahun ini.</p>
                            <div class="pt-2 border-t border-slate-100 flex items-center justify-between">
                                <span class="text-xs text-slate-500">Proporsi dari Anggaran</span>
                                <span class="text-sm font-black" style="color: {ac}">{pct.toFixed(1)}%</span>
                            </div>
                            <div class="h-2 overflow-hidden rounded-full bg-slate-100">
                                <div class="h-full rounded-full transition-all duration-700" style="width: {Math.min(pct, 100)}%; background-color: {ac}"></div>
                            </div>
                        </div>
                    </div>

                    <!-- Kartu 2: Sisa Anggaran — Section 5, gradasi amber -->
                    <div class="bg-white p-5 md:p-6 rounded-2xl border border-slate-200 hover:border-blue-300 shadow-sm hover:shadow-lg transition-all duration-300 flex flex-col relative overflow-hidden group">
                        <div class="absolute top-0 left-0 w-full h-1.5 bg-amber-400"></div>

                        <div class="flex items-center gap-3 mb-4 pb-3 border-b border-slate-100/80">
                            <div class="p-2.5 bg-amber-50 text-amber-600 rounded-xl shadow-sm border border-amber-100">
                                <svg xmlns="http://www.w3.org/2000/svg" class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M3 10h18M7 15h1m4 0h1m-7 4h12a3 3 0 003-3V8a3 3 0 00-3-3H6a3 3 0 00-3 3v8a3 3 0 003 3z" />
                                </svg>
                            </div>
                            <div class="flex items-center gap-2">
                                <span class="inline-block h-3 w-3 rounded-sm shrink-0 bg-[#e2e8f0] border border-slate-300"></span>
                                <h4 class="text-xs font-bold text-slate-700 uppercase tracking-widest">Sisa Anggaran</h4>
                            </div>
                        </div>

                        <div class="space-y-3">
                            <p class="text-2xl font-black text-amber-800">{formatCurrency(sisa)}</p>
                            <p class="text-xs text-slate-400 leading-relaxed">Selisih antara anggaran yang diinput dengan total realisasi. Tersisa dan belum digunakan.</p>
                            <div class="pt-2 border-t border-slate-100">
                                <dl class="space-y-2">
                                    <div class="flex items-center justify-between">
                                        <dt class="text-xs text-slate-500">Anggaran Ditetapkan</dt>
                                        <dd class="text-xs font-black text-slate-900">{formatCurrency(chartModalRow.anggaran || 0)}</dd>
                                    </div>
                                    <div class="flex items-center justify-between border-t border-slate-100 pt-2">
                                        <dt class="text-xs text-violet-700">Total Pajak</dt>
                                        <dd class="text-xs font-black text-violet-800">{formatCurrency(chartModalRow.totalPajak || 0)}</dd>
                                    </div>
                                    <div class="flex items-center justify-between border-t border-slate-100 pt-2">
                                        <dt class="text-xs text-slate-500">Jumlah Transaksi</dt>
                                        <dd class="text-xs font-black text-slate-900">{chartModalRow.jumlahTransaksi} transaksi</dd>
                                    </div>
                                </dl>
                            </div>
                        </div>
                    </div>

                    <!-- Kartu 3: Status Serapan — Section 5, warna solid -->
                    <div class="bg-white p-5 rounded-2xl border border-slate-200 hover:border-blue-300 shadow-sm hover:shadow-lg transition-all duration-300 flex flex-col relative overflow-hidden group">
                        <div class="absolute top-0 left-0 w-full h-1.5" style="background-color: {ac}"></div>

                        <div class="flex items-center gap-3 mb-4 pb-3 border-b border-slate-100/80">
                            <div class="p-2.5 rounded-xl shadow-sm border border-slate-100" style="background-color: {ac}15; color: {ac}; border-color: {ac}30">
                                <svg xmlns="http://www.w3.org/2000/svg" class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 19v-6a2 2 0 00-2-2H5a2 2 0 00-2 2v6a2 2 0 002 2h2a2 2 0 002-2zm0 0V9a2 2 0 012-2h2a2 2 0 012 2v10m-6 0a2 2 0 002 2h2a2 2 0 002-2m0 0V5a2 2 0 012-2h2a2 2 0 012 2v14a2 2 0 01-2 2h-2a2 2 0 01-2-2z" />
                                </svg>
                            </div>
                            <h4 class="text-xs font-bold text-slate-700 uppercase tracking-widest">Status Serapan</h4>
                        </div>

                        <div class="space-y-3">
                            <div class="flex items-end justify-between">
                                <span class="text-3xl font-black" style="color: {ac}">{pct.toFixed(1)}%</span>
                                {#if chartModalRow.anggaran > 0}
                                    {#if pct > 100}
                                        <span class="inline-flex rounded-full border border-rose-200 bg-rose-50 px-3 py-1 text-[9px] font-bold capitalize tracking-wide text-rose-700">Melebihi Anggaran</span>
                                    {:else if pct >= 80}
                                        <span class="inline-flex rounded-full border border-emerald-200 bg-emerald-50 px-3 py-1 text-[9px] font-bold capitalize tracking-wide text-emerald-700">Baik</span>
                                    {:else}
                                        <span class="inline-flex rounded-full border border-amber-200 bg-amber-50 px-3 py-1 text-[9px] font-bold capitalize tracking-wide text-amber-700">Dalam Progress</span>
                                    {/if}
                                {:else}
                                    <span class="inline-flex rounded-full border border-slate-200 bg-slate-100 px-3 py-1 text-[9px] font-bold capitalize tracking-wide text-slate-600">Belum Diinput</span>
                                {/if}
                            </div>
                            <div class="h-3 overflow-hidden rounded-full bg-slate-100">
                                <div class="h-full rounded-full transition-all duration-700" style="width: {Math.min(pct, 100)}%; background-color: {ac}"></div>
                            </div>
                            <p class="text-xs text-slate-400 leading-relaxed">Persentase serapan dihitung dari Realisasi dibagi Anggaran yang telah ditetapkan.</p>
                        </div>
                    </div>

                </div>
            </div>
        {/if}
    </svelte:fragment>

    <!-- Footer — Section 8.2 -->
    <svelte:fragment slot="footer">
        {#if chartModalRow}
            <p class="text-xs text-slate-400 w-full text-center sm:text-left">
                Data serapan anggaran tahun <span class="font-semibold text-slate-600">{currentYear}</span>
            </p>
        {/if}
    </svelte:fragment>
</BaseModal>

<!-- Modal Expand Total Serapan -->
<BaseModal
    bind:open={showTotalModal}
    maxWidth="max-w-6xl"
    on:close={() => { if (totalModalChart) { totalModalChart.destroy(); totalModalChart = null; } }}
>
    <svelte:fragment slot="header">
        <div>
            <h2 class="text-xl font-bold text-slate-800 leading-tight">
                Total Serapan Anggaran
            </h2>
            <p class="text-sm text-slate-500 mt-1">Komposisi total Realisasi dan Sisa Anggaran tahun {currentYear}</p>
        </div>
    </svelte:fragment>

    <svelte:fragment slot="body">
        <div class="grid grid-cols-1 gap-5 lg:grid-cols-5">
            <!-- ===== KIRI: Chart canvas ===== -->
            <div class="lg:col-span-3">
                <div class="bg-white p-5 md:p-6 rounded-2xl border border-slate-200 shadow-sm flex flex-col relative overflow-hidden h-full">
                    <div class="absolute top-0 left-0 w-full h-1.5 bg-emerald-500"></div>
                    <div class="flex items-center gap-3 mb-4 pb-3 border-b border-slate-100/80">
                        <div class="p-2.5 bg-emerald-50 text-emerald-600 rounded-xl shadow-sm border border-emerald-100">
                            <svg xmlns="http://www.w3.org/2000/svg" class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M11 3.055A9.001 9.001 0 1020.945 13H11V3.055z" />
                            </svg>
                        </div>
                        <h4 class="text-xs font-bold text-slate-700 uppercase tracking-widest">Visualisasi Doughnut</h4>
                    </div>
                    <div class="relative flex-1 min-h-[400px]">
                        <canvas bind:this={totalModalCanvas}></canvas>
                    </div>
                </div>
            </div>

            <!-- ===== KANAN: Info cards ===== -->
            <div class="lg:col-span-2 flex flex-col gap-4">
                <div class="bg-white p-5 md:p-6 rounded-2xl border border-slate-200 shadow-sm relative overflow-hidden group hover:border-slate-300 transition-colors">
                    <div class="absolute top-0 left-0 w-full h-1.5 bg-slate-300"></div>
                    <p class="text-xs font-bold uppercase tracking-widest text-slate-500 mb-2">Total Anggaran</p>
                    <p class="text-3xl font-black text-slate-900">{formatCurrency(laporanSummary.totalAnggaran || 0)}</p>
                </div>
                
                <div class="bg-white p-5 md:p-6 rounded-2xl border border-emerald-200 shadow-sm relative overflow-hidden bg-emerald-50/30 group hover:border-emerald-300 transition-colors">
                    <div class="absolute top-0 left-0 w-full h-1.5 bg-emerald-500"></div>
                    <p class="text-xs font-bold uppercase tracking-widest text-emerald-700 mb-2">Total Realisasi</p>
                    <p class="text-3xl font-black text-emerald-800">{formatCurrency(laporanSummary.totalRealisasi || 0)}</p>
                </div>

                <div class="rounded-2xl bg-slate-900 p-5 md:p-6 shadow-sm border border-slate-800 flex items-center justify-between gap-4">
                    <div>
                        <p class="text-xs font-bold uppercase tracking-widest text-slate-400">Total Serapan</p>
                        <h3 class="mt-2 text-3xl font-black text-white">
                            {(laporanSummary.persentaseSerapan || 0).toFixed(1)}<span class="text-lg font-bold text-slate-500 ml-1">%</span>
                        </h3>
                    </div>
                    <div class="relative w-16 h-16 shrink-0">
                        <svg class="w-full h-full transform -rotate-90" viewBox="0 0 36 36">
                            <path class="text-slate-800" stroke-width="4" stroke="currentColor" fill="none" d="M18 2.0845 a 15.9155 15.9155 0 0 1 0 31.831 a 15.9155 15.9155 0 0 1 0 -31.831" />
                            <path class="text-indigo-400 transition-all duration-1000 ease-out" stroke-dasharray="{Math.min(laporanSummary.persentaseSerapan || 0, 100)}, 100" stroke-width="4" stroke-linecap="round" stroke="currentColor" fill="none" d="M18 2.0845 a 15.9155 15.9155 0 0 1 0 31.831 a 15.9155 15.9155 0 0 1 0 -31.831" />
                        </svg>
                    </div>
                </div>

                <div class="bg-white p-5 md:p-6 rounded-2xl border border-amber-200 shadow-sm relative overflow-hidden bg-amber-50/30 group hover:border-amber-300 transition-colors">
                    <div class="absolute top-0 left-0 w-full h-1.5 bg-amber-400"></div>
                    <p class="text-xs font-bold uppercase tracking-widest text-amber-700 mb-2">Sisa Anggaran</p>
                    <p class="text-3xl font-black text-amber-800">{formatCurrency(laporanSummary.sisaAnggaran || 0)}</p>
                </div>
            </div>
        </div>
    </svelte:fragment>

    <svelte:fragment slot="footer">
        <p class="text-xs text-slate-400 w-full text-center sm:text-left">
            Data serapan anggaran tahun <span class="font-semibold text-slate-600">{currentYear}</span>
        </p>
    </svelte:fragment>
</BaseModal>

<!-- Modal Expand Komposisi -->
<BaseModal
    bind:open={showKomposisiModal}
    maxWidth="max-w-6xl"
    on:close={() => { if (komposisiModalChart) { komposisiModalChart.destroy(); komposisiModalChart = null; } }}
>
    <svelte:fragment slot="header">
        <div>
            <h2 class="text-xl font-bold text-slate-800 leading-tight">
                Realisasi per Jenis Pengadaan
            </h2>
            <p class="text-sm text-slate-500 mt-1">Komposisi Realisasi berdasarkan Jenis Pengadaan tahun {currentYear}</p>
        </div>
    </svelte:fragment>

    <svelte:fragment slot="body">
        <div class="grid grid-cols-1 gap-5 lg:grid-cols-5">
            <!-- ===== KIRI: Chart canvas ===== -->
            <div class="lg:col-span-3">
                <div class="bg-white p-5 md:p-6 rounded-2xl border border-slate-200 shadow-sm flex flex-col relative overflow-hidden h-full">
                    <div class="absolute top-0 left-0 w-full h-1.5 bg-sky-500"></div>
                    <div class="flex items-center gap-3 mb-4 pb-3 border-b border-slate-100/80">
                        <div class="p-2.5 bg-sky-50 text-sky-600 rounded-xl shadow-sm border border-sky-100">
                            <svg xmlns="http://www.w3.org/2000/svg" class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M11 3.055A9.001 9.001 0 1020.945 13H11V3.055z" />
                            </svg>
                        </div>
                        <h4 class="text-xs font-bold text-slate-700 uppercase tracking-widest">Visualisasi Pie</h4>
                    </div>
                    <div class="relative flex-1 min-h-[400px]">
                        <canvas bind:this={komposisiModalCanvas}></canvas>
                    </div>
                </div>
            </div>

            <!-- ===== KANAN: Info cards ===== -->
            <div class="lg:col-span-2 flex flex-col gap-4">
                <div class="bg-white p-5 md:p-6 rounded-2xl border border-slate-200 shadow-sm relative overflow-hidden group hover:border-slate-300 transition-colors">
                    <div class="absolute top-0 left-0 w-full h-1.5 bg-sky-500"></div>
                    <p class="text-xs font-bold uppercase tracking-widest text-slate-500 mb-2">Total Jenis Pengadaan</p>
                    <p class="text-3xl font-black text-slate-900">{laporanSummary.totalJenisPengadaan || 0}</p>
                </div>
                
                <div class="bg-white p-5 md:p-6 rounded-2xl border border-slate-200 shadow-sm relative overflow-hidden group hover:border-slate-300 transition-colors">
                    <div class="absolute top-0 left-0 w-full h-1.5 bg-indigo-500"></div>
                    <p class="text-xs font-bold uppercase tracking-widest text-slate-500 mb-2">Jumlah Transaksi GUP</p>
                    <p class="text-3xl font-black text-slate-900">{laporanSummary.totalTransaksi || 0}</p>
                </div>

                <div class="bg-white p-5 md:p-6 rounded-2xl border border-emerald-200 shadow-sm relative overflow-hidden bg-emerald-50/30 group hover:border-emerald-300 transition-colors">
                    <div class="absolute top-0 left-0 w-full h-1.5 bg-emerald-500"></div>
                    <p class="text-xs font-bold uppercase tracking-widest text-emerald-700 mb-2">Nilai Pengajuan (Total Realisasi)</p>
                    <p class="text-3xl font-black text-emerald-800">{formatCurrency(laporanSummary.totalRealisasi || 0)}</p>
                </div>
            </div>
        </div>
    </svelte:fragment>

    <svelte:fragment slot="footer">
        <p class="text-xs text-slate-400 w-full text-center sm:text-left">
            Data serapan anggaran tahun <span class="font-semibold text-slate-600">{currentYear}</span>
        </p>
    </svelte:fragment>
</BaseModal>
