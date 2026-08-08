<script>
    import { userStore } from '$lib/features/auth/store';
    import { onMount, onDestroy } from 'svelte';
    import { dashboardSummaryStore } from '$lib/features/dashboard/store';
    import Button from '$lib/shared/ui/button/Button.svelte';
    import { formatCurrency } from '$lib/shared/utils/utils';
    import LottieLoader from '$lib/shared/ui/loader/LottieLoader.svelte';
    import { Chart, registerables } from 'chart.js';
    import Select from '$lib/shared/ui/select/Select.svelte';
    import { api } from '$lib/shared/api';

    Chart.register(...registerables);

    const currentYear = new Date().getFullYear();
    let selectedYear = currentYear;
    
    const startYear = 2026;
    const endYear = Math.max(currentYear + 1, startYear + 1);
    const yearOptions = Array.from({ length: endYear - startYear + 1 }, (_, i) => ({
        value: startYear + i,
        label: (startYear + i).toString()
    })).reverse();

    async function handleYearChange(e) {
        // Mendapatkan value terbaru langsung dari event karena bind:value Svelte berjalan asinkron
        const newYear = e?.detail?.detail?.value || selectedYear;
        selectedYear = newYear;

        isFetchingStats = true;
        try {
            const data = await api.getGupDashboardSummary(window.fetch, selectedYear);
            dashboardSummaryStore.set(data);
        } catch (err) {
            console.error("Failed to fetch dashboard summary:", err);
        } finally {
            isFetchingStats = false;
        }
    }

    let isFetchingStats = !$dashboardSummaryStore;
    let chartInstances = [];

    let lineChartCanvas;
    let doughNutChartCanvas;

    $: stats = $dashboardSummaryStore || {
        totalPaguAnggaran: 0,
        totalRealisasiGup: 0,
        sisaSaldoUp: 0,
        totalGupBulanIni: 0,
        statusDalkotPending: 0,
        monthlyRealisasi: [],
        compositionUp: [],
        recentTransactions: []
    };

    $: if ($dashboardSummaryStore) {
        isFetchingStats = false;
        if (typeof window !== 'undefined') {
            setTimeout(renderCharts, 100);
        }
    }

    onMount(() => {
        if ($dashboardSummaryStore) {
            isFetchingStats = false;
            setTimeout(renderCharts, 100);
        }
    });

    onDestroy(() => {
        chartInstances.forEach(c => c.destroy());
    });

    const monthNames = ['Jan', 'Feb', 'Mar', 'Apr', 'Mei', 'Jun', 'Jul', 'Agu', 'Sep', 'Okt', 'Nov', 'Des'];

    function renderCharts() {
        if (!lineChartCanvas || !doughNutChartCanvas) return;
        
        // Destroy existing
        chartInstances.forEach(c => c.destroy());
        chartInstances = [];

        // 1. Line Chart
        const lineCtx = lineChartCanvas.getContext('2d');
        const monthsData = new Array(12).fill(0);
        (stats.monthlyRealisasi || []).forEach(item => {
            if (item.month >= 1 && item.month <= 12) {
                monthsData[item.month - 1] = item.total;
            }
        });

        const lineChart = new Chart(lineCtx, {
            type: 'line',
            data: {
                labels: monthNames,
                datasets: [{
                    label: 'Realisasi GUP (Rp)',
                    data: monthsData,
                    borderColor: '#3b82f6',
                    backgroundColor: 'rgba(59, 130, 246, 0.1)',
                    borderWidth: 2,
                    tension: 0.4,
                    fill: true
                }]
            },
            options: {
                responsive: true,
                maintainAspectRatio: false,
                plugins: { legend: { display: false } },
                scales: {
                    y: {
                        beginAtZero: true,
                        ticks: {
                            callback: function(value) {
                                if (value >= 1000000) return (value / 1000000) + 'Jt';
                                return value;
                            }
                        }
                    }
                }
            }
        });
        chartInstances.push(lineChart);

        // 2. Doughnut Chart
        const doughCtx = doughNutChartCanvas.getContext('2d');
        const compLabels = (stats.compositionUp || []).map(c => c.label.replace(/^Belanja\s+/i, ''));
        const compData = (stats.compositionUp || []).map(c => c.value);
        const compCounts = (stats.compositionUp || []).map(c => c.count || 0);
        
        const doughChart = new Chart(doughCtx, {
            type: 'doughnut',
            data: {
                labels: compLabels.length ? compLabels : ['Belum ada data'],
                datasets: [{
                    data: compData.length ? compData : [1],
                    backgroundColor: compData.length ? ['#3b82f6', '#10b981', '#f59e0b', '#8b5cf6', '#ec4899'] : ['#e2e8f0'],
                    borderWidth: 0
                }]
            },
            options: {
                responsive: true,
                maintainAspectRatio: false,
                cutout: '75%',
                plugins: {
                    legend: {
                        position: 'bottom',
                        labels: { padding: 20, usePointStyle: true, pointStyle: 'circle' }
                    },
                    tooltip: {
                        callbacks: {
                            label: function(context) {
                                if (!compData.length) return 'Belum ada data';
                                let val = context.raw || 0;
                                let count = compCounts[context.dataIndex] || 0;
                                return ` ${formatCurrency(val)} (${count} Transaksi)`;
                            }
                        }
                    }
                }
            }
        });
        chartInstances.push(doughChart);
    }

    function formatDate(dateString) {
        if (!dateString) return '-';
        const d = new Date(dateString);
        return d.toLocaleDateString('id-ID', { day: '2-digit', month: 'short', year: 'numeric' });
    }
</script>

{#if isFetchingStats}
    <div class="flex items-center justify-center min-h-[60vh]">
        <LottieLoader text="Memuat Dashboard GUP..." />
    </div>
{:else}
    <div class="space-y-6 pb-20 w-full animate-in fade-in duration-500">
        
        <!-- Welcome Banner -->
        <div class="bg-slate-900 rounded-2xl p-6 sm:p-8 text-white shadow-lg relative flex flex-col sm:flex-row sm:items-center sm:justify-between gap-6">
            <div class="relative z-10">
                {#if $userStore.role === 'protokol'}
                    <p class="text-sm text-slate-300 mb-2">Sistem Informasi Perjalanan Dinas</p>
                    <h1 class="text-2xl sm:text-3xl font-bold mb-3">Dashboard Perjalanan Dinas</h1>
                    <p class="max-w-3xl text-slate-300 text-sm sm:text-base">
                        Kelola dan pantau seluruh pengajuan perjalanan dinas Anda secara cepat dan terpusat.
                    </p>
                {:else}
                    <p class="text-sm text-slate-300 mb-2">Sistem Monitoring Ganti Uang Persediaan</p>
                    <h1 class="text-2xl sm:text-3xl font-bold mb-3">Dashboard GUP & Dalkot</h1>
                    <p class="max-w-3xl text-slate-300 text-sm sm:text-base">
                        Pantau pagu anggaran, saldo uang persediaan, status pengajuan, serta riwayat transaksi GUP secara cepat dan terpusat.
                    </p>
                {/if}
            </div>
            <div class="absolute inset-0 overflow-hidden rounded-2xl pointer-events-none">
                <div class="absolute -right-10 -top-10 w-40 h-40 rounded-full bg-white/5"></div>
                <div class="absolute right-20 -bottom-16 w-52 h-52 rounded-full bg-blue-500/10"></div>
            </div>
        </div>

        <!-- Summary Cards -->
        <div class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-5">
            {#if $userStore.role !== 'protokol'}
            <!-- Pagu -->
            <div class="bg-white rounded-2xl p-5 shadow-sm border border-slate-100 flex flex-col justify-between hover:border-blue-100 transition-colors">
                <div class="flex items-start justify-between gap-4 mb-4">
                    <p class="text-sm text-slate-500 font-medium">Total Pagu Anggaran UP</p>
                    <div class="w-10 h-10 rounded-xl bg-blue-50 text-blue-600 flex items-center justify-center shrink-0">
                        <svg xmlns="http://www.w3.org/2000/svg" class="h-5 w-5" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M21 12V7H5a2 2 0 0 1 0-4h14v4"/><path d="M3 5v14a2 2 0 0 0 2 2h16v-5"/><path d="M18 12a2 2 0 0 0 0 4h4v-4Z"/></svg>
                    </div>
                </div>
                <div>
                    <h3 class="text-2xl font-bold text-slate-800 tracking-tight">{formatCurrency(stats.totalPaguAnggaran)}</h3>
                    <p class="text-xs text-slate-400 mt-1">Tahun Anggaran {new Date().getFullYear()}</p>
                </div>
            </div>

            <!-- Saldo -->
            <div class="bg-white rounded-2xl p-5 shadow-sm border border-slate-100 flex flex-col justify-between hover:border-emerald-100 transition-colors">
                <div class="flex items-start justify-between gap-4 mb-4">
                    <p class="text-sm text-slate-500 font-medium">Sisa Saldo UP Saat Ini</p>
                    <div class="w-10 h-10 rounded-xl bg-emerald-50 text-emerald-600 flex items-center justify-center shrink-0">
                        <svg xmlns="http://www.w3.org/2000/svg" class="h-5 w-5" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><rect width="20" height="14" x="2" y="5" rx="2"/><line x1="2" x2="22" y1="10" y2="10"/></svg>
                    </div>
                </div>
                <div>
                    <h3 class="text-2xl font-bold text-slate-800 tracking-tight">{formatCurrency(stats.sisaSaldoUp)}</h3>
                    <p class="text-xs text-slate-400 mt-1">Pagu dikurangi realisasi GUP</p>
                </div>
            </div>

            <!-- GUP Bulan ini -->
            <div class="bg-white rounded-2xl p-5 shadow-sm border border-slate-100 flex flex-col justify-between hover:border-amber-100 transition-colors">
                <div class="flex items-start justify-between gap-4 mb-4">
                    <p class="text-sm text-slate-500 font-medium">GUP Dibayar Bulan Ini</p>
                    <div class="w-10 h-10 rounded-xl bg-amber-50 text-amber-600 flex items-center justify-center shrink-0">
                        <svg xmlns="http://www.w3.org/2000/svg" class="h-5 w-5" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M12 2v20"/><path d="M17 5H9.5a3.5 3.5 0 0 0 0 7h5a3.5 3.5 0 0 1 0 7H6"/></svg>
                    </div>
                </div>
                <div>
                    <h3 class="text-2xl font-bold text-slate-800 tracking-tight">{formatCurrency(stats.totalGupBulanIni)}</h3>
                    <p class="text-xs text-slate-400 mt-1">Periode bulan berjalan</p>
                </div>
            </div>
            {/if}

            <!-- Dalkot Status -->
            <div class="bg-white rounded-2xl p-5 shadow-sm border border-slate-100 flex flex-col justify-between hover:border-rose-100 transition-colors">
                <div class="flex items-start justify-between gap-4 mb-4">
                    <p class="text-sm text-slate-500 font-medium">Status Dalkot (Pending)</p>
                    <div class="w-10 h-10 rounded-xl bg-rose-50 text-rose-600 flex items-center justify-center shrink-0">
                        <svg xmlns="http://www.w3.org/2000/svg" class="h-5 w-5" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M12 22c5.523 0 10-4.477 10-10S17.523 2 12 2 2 6.477 2 12s4.477 10 10 10z"/><polyline points="12 6 12 12 16 14"/></svg>
                    </div>
                </div>
                <div>
                    <h3 class="text-2xl font-bold text-slate-800 tracking-tight">{stats.statusDalkotPending} <span class="text-sm font-normal text-slate-500">Berkas</span></h3>
                    <p class="text-xs text-slate-400 mt-1">Perlu ditindak lanjuti</p>
                </div>
            </div>
        </div>

        {#if $userStore.role !== 'protokol'}
        <!-- Charts Section -->
        <div class="grid grid-cols-1 xl:grid-cols-3 gap-6">
            <!-- Line Chart -->
            <div class="xl:col-span-2 bg-white rounded-2xl p-6 shadow-sm border border-slate-100">
                <div class="flex flex-col sm:flex-row sm:items-start justify-between gap-4 mb-6">
                    <div>
                        <h3 class="text-lg font-bold text-slate-800">Grafik Realisasi GUP Per Bulan</h3>
                        <p class="text-sm text-slate-500">Visualisasi tren pengajuan GUP tahun berjalan.</p>
                    </div>
                    <div class="w-full sm:w-32 shrink-0">
                        <Select 
                            options={yearOptions}
                            bind:value={selectedYear}
                            on:change={handleYearChange}
                            placeholder="Tahun"
                            class="border-slate-200 bg-slate-50"
                        />
                    </div>
                </div>
                <div class="h-[300px] w-full">
                    <canvas bind:this={lineChartCanvas}></canvas>
                </div>
            </div>

            <!-- Doughnut Chart -->
            <div class="bg-white rounded-2xl p-6 shadow-sm border border-slate-100 flex flex-col">
                <div class="mb-4">
                    <h3 class="text-lg font-bold text-slate-800">Komposisi Penggunaan UP</h3>
                    <p class="text-sm text-slate-500">Berdasarkan klasifikasi Belanja (MAK).</p>
                </div>
                <div class="flex-1 min-h-[250px] flex items-center justify-center relative">
                    <canvas bind:this={doughNutChartCanvas}></canvas>
                    {#if !(stats.compositionUp || []).length}
                        <div class="absolute inset-0 flex items-center justify-center text-sm text-slate-400">Belum ada data komposisi</div>
                    {/if}
                </div>
            </div>
        </div>

        <!-- Recent Transactions -->
        <div class="bg-white rounded-2xl shadow-sm border border-slate-100 overflow-hidden">
            <div class="px-6 py-5 border-b border-slate-100 flex justify-between items-center">
                <div>
                    <h3 class="text-lg font-bold text-slate-800">Data Pengajuan GUP Terakhir</h3>
                    <p class="text-sm text-slate-500">Ringkasan transaksi pengajuan GUP terbaru.</p>
                </div>
                <Button variant="default" class="text-sm bg-slate-900 text-white hover:bg-slate-800" on:click={() => window.location.href = '/dashboard/gup/pengajuan'}>
                    Lihat Semua &rarr;
                </Button>
            </div>
            <div class="overflow-x-auto">
                <table class="w-full text-sm text-left">
                    <thead class="bg-slate-50 text-slate-600 font-medium border-b border-slate-100">
                        <tr>
                            <th class="px-6 py-4">No. SPM/SPP</th>
                            <th class="px-6 py-4">Tanggal</th>
                            <th class="px-6 py-4">Deskripsi Keperluan</th>
                            <th class="px-6 py-4 text-right">Jumlah</th>
                            <th class="px-6 py-4 text-center">Status</th>
                        </tr>
                    </thead>
                    <tbody class="divide-y divide-slate-100">
                        {#if !(stats.recentTransactions || []).length}
                            <tr>
                                <td colspan="5" class="px-6 py-12 text-center text-slate-500">Belum ada transaksi GUP.</td>
                            </tr>
                        {:else}
                            {#each (stats.recentTransactions || []) as trx}
                                <tr class="hover:bg-slate-50/50 transition-colors">
                                    <td class="px-6 py-4 font-mono font-medium text-sm text-slate-700">{trx.businessId || '-'}</td>
                                    <td class="px-6 py-4 text-sm text-slate-600">{formatDate(trx.receiptDate)}</td>
                                    <td class="px-6 py-4 text-sm text-slate-800 leading-relaxed font-medium">{trx.paymentDescription || '-'}</td>
                                    <td class="px-6 py-4 text-right font-bold text-slate-800">{formatCurrency(trx.valueAmount || 0)}</td>
                                    <td class="px-6 py-4 text-center">
                                        <span class="inline-flex items-center px-2.5 py-1 rounded-full text-xs font-bold {trx.paidAmount > 0 ? 'bg-emerald-100 text-emerald-700' : 'bg-amber-100 text-amber-700'}">
                                            {trx.paidAmount > 0 ? 'Lunas' : 'Belum Lunas'}
                                        </span>
                                    </td>
                                </tr>
                            {/each}
                        {/if}
                    </tbody>
                </table>
            </div>
        </div>
        {/if}

    </div>
{/if}
