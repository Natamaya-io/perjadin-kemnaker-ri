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

    const commonOptions = {
        responsive: true,
        maintainAspectRatio: false,
        cutout: '75%',
        plugins: {
            legend: { display: false },
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
                    backgroundColor: ['#10b981', '#f1f5f9'],
                    borderWidth: 0
                }]
            },
            options: {
                ...commonOptions,
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
            type: 'doughnut',
            data: {
                labels: laporanRows.map(r => r.kodeAkun),
                datasets: [{
                    data: laporanRows.map(r => r.realisasi || 0),
                    backgroundColor: colors.slice(0, laporanRows.length),
                    borderWidth: 0
                }]
            },
            options: {
                ...commonOptions,
                cutout: '65%',
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
        // Destroy old charts
        detailCharts.forEach(c => c?.destroy());
        detailCharts = [];
        laporanRows.forEach((row, i) => {
            if (!detailCanvases[i]) return;
            const sisa = Math.max(0, (row.anggaran || 0) - (row.realisasi || 0));
            const chart = new Chart(detailCanvases[i], {
                type: 'doughnut',
                data: {
                    labels: ['Realisasi', 'Sisa Anggaran'],
                    datasets: [{
                        data: [row.realisasi || 0, sisa],
                        backgroundColor: [colors[i % colors.length], '#f1f5f9'],
                        borderWidth: 0
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

<div class="space-y-6 pb-20 max-w-7xl mx-auto">
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
    <section class="rounded-3xl border border-slate-200 bg-white shadow-sm">
        <div class="flex flex-col gap-4 border-b border-slate-200 p-5 sm:flex-row sm:items-center sm:justify-between sm:p-6">
            <div>
                <p class="text-xs font-bold uppercase tracking-[0.2em] text-slate-400">Detail Serapan</p>
                <h2 class="mt-2 text-xl font-black text-slate-900">Pie Chart Tiap Jenis Pengadaan</h2>
                <p class="mt-1 text-sm text-slate-500">Setiap kartu menampilkan komposisi Realisasi dan Sisa Anggaran berdasarkan anggaran yang diinput.</p>
            </div>
            <div class="no-print">
                <Button variant="default" class="gap-2" on:click={() => openAnggaranModal(laporanRows[0])}>
                    <svg xmlns="http://www.w3.org/2000/svg" class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4v16m8-8H4" />
                    </svg>
                    Input Anggaran
                </Button>
            </div>
        </div>

        {#if laporanRows.length === 0}
            <div class="p-16 text-center text-slate-500">
                <svg xmlns="http://www.w3.org/2000/svg" class="h-12 w-12 mx-auto text-slate-300 mb-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z" />
                </svg>
                <p class="font-semibold">Belum ada data laporan untuk tahun {currentYear}.</p>
                <p class="text-sm mt-1 text-slate-400">Pastikan Jenis Pengadaan sudah diinput di master data dan ada transaksi GUP.</p>
            </div>
        {:else}
            <div class="grid grid-cols-1 gap-5 p-5 sm:grid-cols-2 sm:p-6">
                {#each laporanRows as row, i}
                    <article class="rounded-3xl border border-slate-200 bg-white p-5 shadow-sm transition hover:-translate-y-1 hover:shadow-lg">
                        <div class="flex items-start justify-between gap-4">
                            <div>
                                <p class="text-xs font-bold uppercase tracking-[0.16em] text-slate-400">{row.kodeAkun}</p>
                                <h3 class="mt-2 text-base font-black leading-snug text-slate-900">{row.jenisPengadaan}</h3>
                                <p class="mt-1 text-xs text-slate-400">{row.jumlahTransaksi} transaksi pengajuan</p>
                            </div>
                            <button type="button" class="no-print text-blue-600 hover:text-blue-800 bg-blue-50 hover:bg-blue-100 p-2 rounded transition-colors" title="Input Anggaran" on:click={() => openAnggaranModal(row)}>
                                <svg xmlns="http://www.w3.org/2000/svg" class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15.232 5.232l3.536 3.536m-2.036-5.036a2.5 2.5 0 113.536 3.536L6.5 21.036H3v-3.572L16.732 3.732z" />
                                </svg>
                            </button>
                        </div>

                        <div class="relative mt-5 h-52 rounded-3xl bg-slate-50 p-4">
                            <canvas bind:this={detailCanvases[i]}></canvas>
                        </div>

                        <div class="mt-5 grid grid-cols-2 gap-3 text-sm">
                            <div class="rounded-2xl bg-slate-50 p-3">
                                <p class="text-xs font-semibold text-slate-500">Anggaran</p>
                                <p class="mt-1 font-black text-slate-900">{formatCurrency(row.anggaran || 0)}</p>
                            </div>
                            <div class="rounded-2xl bg-emerald-50 p-3 border border-emerald-100">
                                <p class="text-xs font-semibold text-emerald-700">Realisasi</p>
                                <p class="mt-1 font-black text-emerald-800">{formatCurrency(row.realisasi || 0)}</p>
                            </div>
                            <div class="rounded-2xl bg-amber-50 p-3 border border-amber-100">
                                <p class="text-xs font-semibold text-amber-700">Sisa</p>
                                <p class="mt-1 font-black text-amber-800">{formatCurrency(row.sisaAnggaran || 0)}</p>
                            </div>
                            <div class="rounded-2xl bg-violet-50 p-3 border border-violet-100">
                                <p class="text-xs font-semibold text-violet-700">Serapan</p>
                                <p class="mt-1 font-black text-violet-800">{(row.persentaseSerapan || 0).toFixed(1)}%</p>
                            </div>
                        </div>

                        <div class="mt-4">
                            <div class="h-2.5 overflow-hidden rounded-full bg-slate-100">
                                <div class="h-full rounded-full transition-all duration-500"
                                    style="width: {Math.min(row.persentaseSerapan || 0, 100)}%; background-color: {colors[i % colors.length]}">
                                </div>
                            </div>
                            <div class="mt-3 flex items-center justify-between gap-3">
                                {#if row.anggaran > 0}
                                    {#if (row.persentaseSerapan || 0) > 100}
                                        <span class="inline-flex rounded-full border border-rose-200 bg-rose-100 px-3 py-1 text-[10px] uppercase tracking-wider font-bold text-rose-700">Melebihi Anggaran</span>
                                    {:else if (row.persentaseSerapan || 0) >= 80}
                                        <span class="inline-flex rounded-full border border-emerald-200 bg-emerald-100 px-3 py-1 text-[10px] uppercase tracking-wider font-bold text-emerald-700">Baik</span>
                                    {:else}
                                        <span class="inline-flex rounded-full border border-amber-200 bg-amber-100 px-3 py-1 text-[10px] uppercase tracking-wider font-bold text-amber-700">Dalam Progress</span>
                                    {/if}
                                {:else}
                                    <span class="inline-flex rounded-full border border-slate-200 bg-slate-100 px-3 py-1 text-[10px] uppercase tracking-wider font-bold text-slate-600">Belum Diinput</span>
                                {/if}
                                <span class="text-xs text-slate-400 font-mono">MAK: {row.mak}</span>
                            </div>
                        </div>
                    </article>
                {/each}
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
    <div class="fixed inset-0 z-[100] bg-slate-900/80 flex items-center justify-center p-4 no-print" on:click|self={closeAnggaranModal}>
        <div class="w-full max-w-lg bg-white rounded-2xl shadow-2xl animate-in fade-in zoom-in-95 duration-200">
            <div class="px-6 py-5 border-b border-slate-100 flex items-center justify-between">
                <div>
                    <h3 class="text-lg font-bold text-slate-900">Input Anggaran</h3>
                    <p class="text-sm text-slate-500 mt-0.5">{selectedRow?.jenisPengadaan || ''}</p>
                </div>
                <button class="rounded-full p-2 hover:bg-slate-100 text-slate-400 hover:text-slate-600 transition-colors" on:click={closeAnggaranModal}>
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
                    <label class="text-sm font-semibold text-slate-700">Nominal Anggaran (Rp)</label>
                    <input
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
