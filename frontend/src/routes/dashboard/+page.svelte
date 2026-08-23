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
        <div class="flex flex-col md:flex-row md:items-center justify-between gap-4 bg-white p-6 sm:p-8 rounded-2xl shadow-sm border border-slate-100">
            <div>
                {#if $userStore.role === 'protokol'}
                    <p class="text-sm text-slate-500 mb-1">Sistem Informasi Perjalanan Dinas</p>
                    <h1 class="text-2xl sm:text-3xl font-bold text-slate-800 tracking-tight">Dashboard Perjalanan Dinas</h1>
                    <p class="max-w-3xl text-slate-500 text-sm sm:text-base mt-1">
                        Kelola dan pantau seluruh pengajuan perjalanan dinas Anda secara cepat dan terpusat.
                    </p>
                {:else}
                    <p class="text-sm text-slate-500 mb-1">Sistem Monitoring Ganti Uang Persediaan</p>
                    <h1 class="text-2xl sm:text-3xl font-bold text-slate-800 tracking-tight">{$userStore?.role === 'super_admin' || $userStore?.role === 'kasubag' ? 'Master Dashboard' : 'Dashboard GUP & Dalkot'}</h1>
                    <p class="max-w-3xl text-slate-500 text-sm sm:text-base mt-1">
                        Pantau pagu anggaran, saldo uang persediaan, status pengajuan, serta riwayat transaksi GUP secara cepat dan terpusat.
                    </p>
                {/if}
            </div>
        </div>



        <!-- Rekapitulasi Cards -->
        <div class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-5">
            <!-- Rekap GUP -->
            {#if $userStore?.role !== 'protokol'}
            <a href="/dashboard/rekapitulasi-gup" class="bg-white rounded-2xl p-5 shadow-sm border border-slate-100 flex flex-col justify-between hover:border-blue-400 hover:shadow-md transition-all group">
                <div class="flex items-start justify-between gap-4 mb-4">
                    <p class="text-sm text-slate-600 font-bold group-hover:text-blue-600 transition-colors">Rekapitulasi GUP</p>
                    <div class="w-10 h-10 rounded-xl bg-blue-50 text-blue-600 flex items-center justify-center shrink-0 group-hover:scale-110 transition-transform">
                        <svg xmlns="http://www.w3.org/2000/svg" class="h-5 w-5" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M14 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V8z"/><polyline points="14 2 14 8 20 8"/><line x1="16" x2="8" y1="13" y2="13"/><line x1="16" x2="8" y1="17" y2="17"/><polyline points="10 9 9 9 8 9"/></svg>
                    </div>
                </div>
                <div class="flex items-center text-xs font-medium text-blue-600 mt-2">
                    Lihat Dokumen <svg xmlns="http://www.w3.org/2000/svg" class="h-3.5 w-3.5 ml-1 transition-transform group-hover:translate-x-1" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><polyline points="9 18 15 12 9 6"/></svg>
                </div>
            </a>
            {/if}

            <!-- Rekap Dalam Kota -->
            <a href="/dashboard/rekapitulasi-dalkot" class="bg-white rounded-2xl p-5 shadow-sm border border-slate-100 flex flex-col justify-between hover:border-emerald-400 hover:shadow-md transition-all group">
                <div class="flex items-start justify-between gap-4 mb-4">
                    <p class="text-sm text-slate-600 font-bold group-hover:text-emerald-600 transition-colors">Rekapitulasi Perdin Dalam Kota</p>
                    <div class="w-10 h-10 rounded-xl bg-emerald-50 text-emerald-600 flex items-center justify-center shrink-0 group-hover:scale-110 transition-transform">
                        <svg xmlns="http://www.w3.org/2000/svg" class="h-5 w-5" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><rect width="16" height="20" x="4" y="2" rx="2" ry="2"/><path d="M9 22v-4h6v4"/><path d="M8 6h.01"/><path d="M16 6h.01"/><path d="M12 6h.01"/><path d="M12 10h.01"/><path d="M12 14h.01"/><path d="M16 10h.01"/><path d="M16 14h.01"/><path d="M8 10h.01"/><path d="M8 14h.01"/></svg>
                    </div>
                </div>
                <div class="flex items-center text-xs font-medium text-emerald-600 mt-2">
                    Lihat Dokumen <svg xmlns="http://www.w3.org/2000/svg" class="h-3.5 w-3.5 ml-1 transition-transform group-hover:translate-x-1" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><polyline points="9 18 15 12 9 6"/></svg>
                </div>
            </a>

            <!-- Rekap Luar Kota -->
            <a href="/dashboard/rekapitulasi-luar-kota" class="bg-white rounded-2xl p-5 shadow-sm border border-slate-100 flex flex-col justify-between hover:border-amber-400 hover:shadow-md transition-all group">
                <div class="flex items-start justify-between gap-4 mb-4">
                    <p class="text-sm text-slate-600 font-bold group-hover:text-amber-600 transition-colors">Rekapitulasi Perdin Luar Kota</p>
                    <div class="w-10 h-10 rounded-xl bg-amber-50 text-amber-600 flex items-center justify-center shrink-0 group-hover:scale-110 transition-transform">
                        <svg xmlns="http://www.w3.org/2000/svg" class="h-5 w-5" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><rect x="2" y="7" width="20" height="15" rx="2" ry="2"/><polyline points="17 2 12 7 7 2"/></svg>
                    </div>
                </div>
                <div class="flex items-center text-xs font-medium text-amber-600 mt-2">
                    Lihat Dokumen <svg xmlns="http://www.w3.org/2000/svg" class="h-3.5 w-3.5 ml-1 transition-transform group-hover:translate-x-1" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><polyline points="9 18 15 12 9 6"/></svg>
                </div>
            </a>

            <!-- Rekap Luar Negeri -->
            <a href="/dashboard/laporan" class="bg-white rounded-2xl p-5 shadow-sm border border-slate-100 flex flex-col justify-between hover:border-purple-400 hover:shadow-md transition-all group">
                <div class="flex items-start justify-between gap-4 mb-4">
                    <p class="text-sm text-slate-600 font-bold group-hover:text-purple-600 transition-colors">Rekapitulasi Perdin Luar Negeri</p>
                    <div class="w-10 h-10 rounded-xl bg-purple-50 text-purple-600 flex items-center justify-center shrink-0 group-hover:scale-110 transition-transform">
                        <svg xmlns="http://www.w3.org/2000/svg" class="h-5 w-5" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M17.8 19.2 16 11l3.5-3.5C21 6 21.5 4 21 3c-1-.5-3 0-4.5 1.5L13 8 4.8 6.2c-.5-.1-.9.2-1.1.6L3 8l6 5-4 4-3-1-1 1 2 4 4 2 1-1-1-3 4-4 5 6l1.2-.7c.4-.2.7-.6.6-1.1z"/></svg>
                    </div>
                </div>
                <div class="flex items-center text-xs font-medium text-purple-600 mt-2">
                    Lihat Dokumen <svg xmlns="http://www.w3.org/2000/svg" class="h-3.5 w-3.5 ml-1 transition-transform group-hover:translate-x-1" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><polyline points="9 18 15 12 9 6"/></svg>
                </div>
            </a>
        </div>



    </div>
{/if}
