<script>
    import { userStore } from '$lib/features/auth/store';
    import { onMount } from 'svelte';
    import { api } from '$lib/shared/api';
    import { formatCurrency, getStatusBadge } from '$lib/shared/utils/utils';
    import LottieLoader from '$lib/shared/ui/loader/LottieLoader.svelte';
    import Button from '$lib/shared/ui/button/Button.svelte';
    import BaseModal from '$lib/shared/ui/base-modal/BaseModal.svelte';
    import Select from '$lib/shared/ui/select/Select.svelte';

    let isInfoModalOpen = false;
    let infoModalTitle = '';
    let infoModalContent = '';
    let infoModalNotes = '';

    function openInfoModal(title, content, notes = '') {
        infoModalTitle = title;
        infoModalContent = content;
        infoModalNotes = notes;
        isInfoModalOpen = true;
    }

    let isLoading = true;
    let records = [];
    let searchQuery = '';

    const currentYear = new Date().getFullYear();
    let filterTahun = currentYear.toString();

    const appLaunchYear = 2025;
    const endYear = Math.max(currentYear + 1, appLaunchYear + 1);
    const tahunOptions = [
        { value: '', label: 'Semua Tahun' },
        ...Array.from({ length: endYear - appLaunchYear + 1 }, (_, i) => ({
            value: (appLaunchYear + i).toString(),
            label: (appLaunchYear + i).toString()
        })).reverse()
    ];

    let filterBulan = '';
    const bulanOptions = [
        { value: '', label: 'Semua Bulan' },
        { value: '01', label: 'Januari' },
        { value: '02', label: 'Februari' },
        { value: '03', label: 'Maret' },
        { value: '04', label: 'April' },
        { value: '05', label: 'Mei' },
        { value: '06', label: 'Juni' },
        { value: '07', label: 'Juli' },
        { value: '08', label: 'Agustus' },
        { value: '09', label: 'September' },
        { value: '10', label: 'Oktober' },
        { value: '11', label: 'November' },
        { value: '12', label: 'Desember' }
    ];

    function resetFilters() {
        searchQuery = '';
        filterTahun = currentYear.toString();
        filterBulan = '';
    }

    onMount(async () => {
        try {
            // Gunakan endpoint paginated dengan filter type=luar_kota agar server
            // hanya mengirim data yang relevan saja (~267KB vs ~1.2MB sebelumnya).
            const res = await api.getPaginatedRecords({ type: 'luar_kota', limit: 9999 });
            const rawRecords = res?.data ?? [];
            
            const grouped = {};
            const uniqueRecords = [];
            
            rawRecords.forEach(r => {
                const spd = r.spd;
                const cost = r.totalCost || 0;
                
                if (spd) {
                    if (!grouped[spd]) {
                        grouped[spd] = { ...r, spd, totalSpjCost: cost, totalActualCost: 0 };
                        uniqueRecords.push(grouped[spd]);
                    } else {
                        grouped[spd].totalSpjCost += cost;
                    }
                } else {
                    uniqueRecords.push({ ...r, spd: '-', totalSpjCost: cost, totalActualCost: 0 });
                }
            });
            
            records = uniqueRecords;
        } catch (e) {
            console.error('Failed to fetch luar kota records:', e);
        } finally {
            isLoading = false;
        }
    });

    $: recordsByYear = records.filter(r => {
        if (!filterTahun) return true;
        const dateStr = r.executionDate || r.startDate || '';
        return dateStr ? dateStr.split('-')[0] === filterTahun : false;
    });

    $: submitted = recordsByYear.filter(r => {
        const l = statusLabel(r.status).label;
        return l === 'Ajukan' || r.status === 'Submitted' || (typeof r.status === 'string' && r.status.toLowerCase() === 'submitted');
    });
    $: approved  = recordsByYear.filter(r => {
        const l = statusLabel(r.status).label;
        return l === 'Setujui' || l === 'Selesai' || r.status === 'Approved' || r.status === 'Completed' || r.status === 'Paid' || (typeof r.status === 'string' && (r.status.toLowerCase() === 'approved' || r.status.toLowerCase() === 'completed' || r.status.toLowerCase() === 'paid'));
    });
    // [ANTI-MISS COUNT]: Pastikan total mutlak sama dengan Total Pengajuan dengan menjadikan pending sebagai catch-all.
    $: pending = recordsByYear.filter(r => !submitted.includes(r) && !approved.includes(r));

    $: totalSpj      = recordsByYear.reduce((s, r) => s + (r.totalSpjCost || 0), 0);
    $: totalRiil     = recordsByYear.reduce((s, r) => s + (r.totalActualCost || 0), 0);
    // [ANTI-MISS COST]: Pending = semua yang belum disetujui (Draft + Ajukan), sehingga Total = Pending + Disetujui
    $: totalPending  = [...pending, ...submitted].reduce((s, r) => s + (r.totalSpjCost || 0), 0);
    $: totalApproved = approved.reduce((s, r) => s + (r.totalSpjCost || 0), 0);

    $: filtered = recordsByYear.filter(r => {

        const dateStr = r.executionDate || r.startDate || '';
        if (filterBulan && dateStr) {
            if (dateStr.split('-')[1] !== filterBulan) return false;
        }
        if (searchQuery) {
            const q = searchQuery.toLowerCase();
            return (
                (r.spd || '').toLowerCase().includes(q) ||
                (r.purpose || '').toLowerCase().includes(q) ||
                (r.location || '').toLowerCase().includes(q)
            );
        }
        return true;
    });

    let currentPage = 1;
    const itemsPerPage = 10;
    
    $: totalPages = Math.ceil(filtered.length / itemsPerPage) || 1;
    $: if (currentPage > totalPages && totalPages > 0) currentPage = totalPages;
    $: paginatedRecords = filtered.slice((currentPage - 1) * itemsPerPage, currentPage * itemsPerPage);

    $: if (searchQuery || filterTahun || filterBulan) currentPage = 1;

    function formatDate(dateString) {
        if (!dateString) return '-';
        return new Date(dateString).toLocaleDateString('id-ID', { day: '2-digit', month: 'short', year: 'numeric' });
    }

    function statusLabel(status) {
        const map = {
            'Draft': { label: 'Draft', class: 'bg-slate-100 text-slate-600 border-slate-200' },
            'Pending': { label: 'Draft', class: 'bg-slate-100 text-slate-600 border-slate-200' },
            'Submitted': { label: 'Ajukan', class: 'bg-orange-50 text-orange-700 border-orange-200' },
            'Approved': { label: 'Setujui', class: 'bg-emerald-50 text-emerald-700 border-emerald-200' },
            'Rejected': { label: 'Draft', class: 'bg-slate-100 text-slate-600 border-slate-200' },
            'Completed': { label: 'Selesai', class: 'bg-indigo-50 text-indigo-700 border-indigo-200' },
            'Paid': { label: 'Selesai', class: 'bg-indigo-50 text-indigo-700 border-indigo-200' }
        };
        return map[status] || { label: status || 'Draft', class: 'bg-slate-100 text-slate-600 border-slate-200' };
    }
</script>

<svelte:head>
    <title>Rekapitulasi Perdin Luar Kota | Dashboard</title>
</svelte:head>

{#if isLoading}
    <div class="flex items-center justify-center min-h-[60vh]">
        <LottieLoader text="Memuat Rekapitulasi Perdin Luar Kota..." />
    </div>
{:else}
    <div class="space-y-6 pb-20 w-full animate-in fade-in duration-500">

        <!-- Header -->
        <div class="flex flex-col md:flex-row md:items-center justify-between gap-4 bg-white p-6 rounded-2xl shadow-sm border border-slate-100">
            <div>
                <div class="flex items-center gap-2 mb-2">
                    <a href="/dashboard" class="text-slate-500 hover:text-slate-800 transition-colors text-sm flex items-center gap-1">
                        <svg xmlns="http://www.w3.org/2000/svg" class="h-4 w-4" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><polyline points="15 18 9 12 15 6"/></svg>
                        Master Dashboard
                    </a>
                    <span class="text-slate-300">/</span>
                    <span class="text-slate-600 text-sm">Rekapitulasi Perdin Luar Kota</span>
                </div>
                <h1 class="text-2xl font-bold text-slate-800 tracking-tight">Rekapitulasi Perdin Luar Kota</h1>
                <p class="text-sm text-slate-500 mt-1 max-w-xl">Pantau seluruh status, realisasi biaya, dan riwayat Perjalanan Dinas Luar Kota secara terpusat.</p>
            </div>
            <div class="flex flex-col sm:flex-row items-stretch sm:items-center gap-3">
                <Select
                    id="filterTahun"
                    bind:value={filterTahun}
                    options={tahunOptions}
                    class="w-full sm:w-40"
                />
                <Button variant="default" class="w-full sm:w-auto flex items-center justify-center gap-2" on:click={() => window.location.href = '/dashboard/admin/perdin'}>
                    Kelola Perdin &rarr;
                </Button>
            </div>
        </div>

        <!-- Summary Cards -->
        <div class="grid grid-cols-2 lg:grid-cols-4 gap-5">
            <!-- Total Pengajuan -->
            <div class="bg-white rounded-2xl p-5 shadow-sm border border-slate-100 hover:border-emerald-200 transition-colors">
                <div class="flex items-start justify-between gap-3 mb-4">
                    <div class="flex items-center gap-1.5">
                        <p class="text-xs sm:text-sm text-slate-500 font-medium">Total Pengajuan</p>
                        <button type="button" title="Informasi" class="text-blue-500 hover:text-blue-700 bg-blue-50 hover:bg-blue-100 rounded-full p-0.5 transition-colors" on:click|preventDefault={() => openInfoModal('Informasi: Total Pengajuan', 'Menampilkan keseluruhan dokumen Perjalanan Dinas Luar Kota (Luar Kota) yang pernah dibuat, baik yang masih berupa draft maupun yang sudah disetujui.')}>
                            <svg xmlns="http://www.w3.org/2000/svg" class="h-3.5 w-3.5" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13 16h-1v-4h-1m1-4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z" /></svg>
                        </button>
                    </div>
                    <div class="w-9 h-9 rounded-xl bg-emerald-50 text-emerald-600 flex items-center justify-center shrink-0">
                        <svg xmlns="http://www.w3.org/2000/svg" class="h-4 w-4" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M14 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V8z"/><polyline points="14 2 14 8 20 8"/></svg>
                    </div>
                </div>
                <h3 class="text-2xl font-bold text-slate-800">{recordsByYear.length} <span class="text-sm font-normal text-slate-500">Berkas</span></h3>
                <p class="text-xs text-slate-400 mt-1">Keseluruhan data Luar Kota</p>
            </div>

            <!-- Pending (Draft) -->
            <div class="bg-white rounded-2xl p-5 shadow-sm border border-slate-100 hover:border-slate-300 transition-colors">
                <div class="flex items-start justify-between gap-3 mb-4">
                    <div class="flex items-center gap-1.5">
                        <p class="text-xs sm:text-sm text-slate-500 font-medium">Draft</p>
                        <button type="button" title="Informasi" class="text-blue-500 hover:text-blue-700 bg-blue-50 hover:bg-blue-100 rounded-full p-0.5 transition-colors" on:click|preventDefault={() => openInfoModal('Informasi: Draft', 'Menampilkan dokumen yang baru dibuat (Draft).', 'Periksa kembali kelengkapan dokumen yang berstatus Draft sebelum diajukan.')}>
                            <svg xmlns="http://www.w3.org/2000/svg" class="h-3.5 w-3.5" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13 16h-1v-4h-1m1-4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z" /></svg>
                        </button>
                    </div>
                    <div class="w-9 h-9 rounded-xl bg-slate-100 text-slate-600 flex items-center justify-center shrink-0">
                        <svg xmlns="http://www.w3.org/2000/svg" class="h-4 w-4" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M12 22c5.523 0 10-4.477 10-10S17.523 2 12 2 2 6.477 2 12s4.477 10 10 10z"/><circle cx="12" cy="12" r="2"/></svg>
                    </div>
                </div>
                <h3 class="text-2xl font-bold text-slate-800">{pending.length} <span class="text-sm font-normal text-slate-500">Berkas</span></h3>
            </div>



            <!-- Ajukan -->
            <div class="bg-white rounded-2xl p-5 shadow-sm border border-slate-100 hover:border-orange-200 transition-colors">
                <div class="flex items-start justify-between gap-3 mb-4">
                    <div class="flex items-center gap-1.5">
                        <p class="text-xs sm:text-sm text-slate-500 font-medium">Ajukan</p>
                        <button type="button" title="Informasi" class="text-blue-500 hover:text-blue-700 bg-blue-50 hover:bg-blue-100 rounded-full p-0.5 transition-colors" on:click|preventDefault={() => openInfoModal('Informasi: Ajukan', 'Menampilkan dokumen yang sudah dikirim dan sedang menunggu persetujuan (Submitted) oleh pihak yang berwenang.')}>
                            <svg xmlns="http://www.w3.org/2000/svg" class="h-3.5 w-3.5" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13 16h-1v-4h-1m1-4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z" /></svg>
                        </button>
                    </div>
                    <div class="w-9 h-9 rounded-xl bg-orange-50 text-orange-600 flex items-center justify-center shrink-0">
                        <svg xmlns="http://www.w3.org/2000/svg" class="h-4 w-4" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><polyline points="22 2 15 22 11 13 2 9 22 2"/></svg>
                    </div>
                </div>
                <h3 class="text-2xl font-bold text-slate-800">{submitted.length} <span class="text-sm font-normal text-slate-500">Berkas</span></h3>
            </div>

            <!-- Disetujui -->
            <div class="bg-white rounded-2xl p-5 shadow-sm border border-slate-100 hover:border-emerald-200 transition-colors">
                <div class="flex items-start justify-between gap-3 mb-4">
                    <div class="flex items-center gap-1.5">
                        <p class="text-xs sm:text-sm text-slate-500 font-medium">Disetujui</p>
                        <button type="button" title="Informasi" class="text-blue-500 hover:text-blue-700 bg-blue-50 hover:bg-blue-100 rounded-full p-0.5 transition-colors" on:click|preventDefault={() => openInfoModal('Informasi: Disetujui', 'Menampilkan dokumen yang sudah final, disahkan (Approved), atau telah selesai sepenuhnya (Completed).')}>
                            <svg xmlns="http://www.w3.org/2000/svg" class="h-3.5 w-3.5" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13 16h-1v-4h-1m1-4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z" /></svg>
                        </button>
                    </div>
                    <div class="w-9 h-9 rounded-xl bg-emerald-50 text-emerald-600 flex items-center justify-center shrink-0">
                        <svg xmlns="http://www.w3.org/2000/svg" class="h-4 w-4" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><polyline points="20 6 9 17 4 12"/></svg>
                    </div>
                </div>
                <h3 class="text-2xl font-bold text-slate-800">{approved.length} <span class="text-sm font-normal text-slate-500">Berkas</span></h3>
                {#if $userStore?.role !== 'protokol'}
                <p class="text-xs text-emerald-600 font-medium mt-1">{formatCurrency(totalApproved)}</p>
                {/if}
            </div>
        </div>

        <!-- Biaya Summary -->
        {#if $userStore?.role !== 'protokol'}
        <div class="grid grid-cols-1 sm:grid-cols-3 gap-5">
            <div class="bg-white rounded-2xl p-5 shadow-sm border border-slate-100 hover:border-blue-200 transition-colors">
                <div class="flex items-start justify-between gap-3 mb-3">
                    <div class="flex items-center gap-1.5">
                        <p class="text-sm text-slate-500 font-medium">Total Biaya Keseluruhan</p>
                        <button type="button" title="Informasi" class="text-blue-500 hover:text-blue-700 bg-blue-50 hover:bg-blue-100 rounded-full p-0.5 transition-colors" on:click|preventDefault={() => openInfoModal('Informasi: Total Biaya', 'Total keseluruhan biaya Perjalanan Dinas Luar Kota.')}>
                            <svg xmlns="http://www.w3.org/2000/svg" class="h-3.5 w-3.5" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13 16h-1v-4h-1m1-4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z" /></svg>
                        </button>
                    </div>
                    <div class="w-9 h-9 rounded-xl bg-blue-50 text-blue-600 flex items-center justify-center shrink-0">
                        <svg xmlns="http://www.w3.org/2000/svg" class="h-4 w-4" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><rect width="20" height="14" x="2" y="5" rx="2"/><line x1="2" x2="22" y1="10" y2="10"/></svg>
                    </div>
                </div>
                <h3 class="text-xl font-bold text-slate-800">{formatCurrency(totalSpj)}</h3>
                <p class="text-xs text-slate-400 mt-1">Akumulasi seluruh biaya</p>
            </div>
            <div class="bg-white rounded-2xl p-5 shadow-sm border border-slate-100 hover:border-yellow-200 transition-colors">
                <div class="flex items-start justify-between gap-3 mb-3">
                    <div class="flex items-center gap-1.5">
                        <p class="text-sm text-slate-500 font-medium">Biaya Pending</p>
                        <button type="button" title="Informasi" class="text-blue-500 hover:text-blue-700 bg-blue-50 hover:bg-blue-100 rounded-full p-0.5 transition-colors" on:click|preventDefault={() => openInfoModal('Informasi: Biaya Pending', 'Potensi pengeluaran dari dokumen-dokumen yang statusnya belum disetujui (masih Draft atau Ajukan).', 'Nilai ini dapat berubah setelah dokumen disetujui.')}>
                            <svg xmlns="http://www.w3.org/2000/svg" class="h-3.5 w-3.5" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13 16h-1v-4h-1m1-4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z" /></svg>
                        </button>
                    </div>
                    <div class="w-9 h-9 rounded-xl bg-yellow-50 text-yellow-600 flex items-center justify-center shrink-0">
                        <svg xmlns="http://www.w3.org/2000/svg" class="h-4 w-4" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><circle cx="12" cy="12" r="10"/><line x1="12" y1="8" x2="12" y2="12"/><line x1="12" y1="16" x2="12.01" y2="16"/></svg>
                    </div>
                </div>
                <h3 class="text-xl font-bold text-yellow-700">{formatCurrency(totalPending)}</h3>
                <p class="text-xs text-slate-400 mt-1">Draft &amp; Ajukan / Belum disetujui</p>
            </div>
            <div class="bg-white rounded-2xl p-5 shadow-sm border border-slate-100 hover:border-emerald-200 transition-colors">
                <div class="flex items-start justify-between gap-3 mb-3">
                    <div class="flex items-center gap-1.5">
                        <p class="text-sm text-slate-500 font-medium">Biaya Disetujui</p>
                        <button type="button" title="Informasi" class="text-blue-500 hover:text-blue-700 bg-blue-50 hover:bg-blue-100 rounded-full p-0.5 transition-colors" on:click|preventDefault={() => openInfoModal('Informasi: Biaya Disetujui', 'Total biaya keseluruhan yang sah atau sudah disetujui untuk dicairkan.')}>
                            <svg xmlns="http://www.w3.org/2000/svg" class="h-3.5 w-3.5" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13 16h-1v-4h-1m1-4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z" /></svg>
                        </button>
                    </div>
                    <div class="w-9 h-9 rounded-xl bg-emerald-50 text-emerald-600 flex items-center justify-center shrink-0">
                        <svg xmlns="http://www.w3.org/2000/svg" class="h-4 w-4" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><polyline points="20 6 9 17 4 12"/></svg>
                    </div>
                </div>
                <h3 class="text-xl font-bold text-emerald-700">{formatCurrency(totalApproved)}</h3>
                <p class="text-xs text-slate-400 mt-1">Status Approved</p>
            </div>
        </div>
        {/if}

        <!-- Alert Pending -->
        {#if pending.length > 0}
        <div class="bg-yellow-50 border border-yellow-200 rounded-2xl p-4 flex items-center gap-3">
            <svg xmlns="http://www.w3.org/2000/svg" class="h-5 w-5 text-yellow-600 shrink-0" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M10.29 3.86L1.82 18a2 2 0 0 0 1.71 3h16.94a2 2 0 0 0 1.71-3L13.71 3.86a2 2 0 0 0-3.42 0z"/><line x1="12" y1="9" x2="12" y2="13"/><line x1="12" y1="17" x2="12.01" y2="17"/></svg>
            <p class="text-sm font-medium text-yellow-800">Terdapat <span class="font-bold">{pending.length} berkas</span> Perdin Luar Kota yang masih berstatus Draft/Pending dan perlu segera ditindak lanjuti.</p>
        </div>
        {/if}

        <!-- Table -->
        <div class="bg-white rounded-2xl shadow-sm border border-slate-100 overflow-hidden">
            <div class="px-6 py-5 border-b border-slate-100">
                <div class="flex flex-col sm:flex-row sm:items-center justify-between gap-3 mb-4">
                    <div>
                        <h3 class="text-lg font-bold text-slate-800">Daftar Perjalanan Dinas Luar Kota</h3>
                        <p class="text-sm text-slate-500">{filtered.length} dari {recordsByYear.length} berkas ditemukan (Tahun {filterTahun || 'Semua'}).</p>
                    </div>
                </div>
                <!-- Filter Bar -->
                <div class="grid grid-cols-1 sm:grid-cols-[1fr_minmax(160px,200px)_auto] gap-3 items-end">
                    <!-- Search -->
                    <div>
                        <label for="searchQuery" class="mb-1.5 block text-xs font-bold uppercase tracking-widest text-slate-500">Pencarian</label>
                        <div class="relative">
                            <svg xmlns="http://www.w3.org/2000/svg" class="absolute left-3 top-1/2 -translate-y-1/2 h-4 w-4 text-slate-400 pointer-events-none" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><circle cx="11" cy="11" r="8"/><line x1="21" y1="21" x2="16.65" y2="16.65"/></svg>
                            <input
                                id="searchQuery"
                                bind:value={searchQuery}
                                type="text"
                                placeholder="Cari SPD, kegiatan, lokasi..."
                                class="w-full pl-9 pr-4 py-2.5 text-sm border border-slate-200 rounded-xl bg-slate-50 focus:outline-none focus:ring-2 focus:ring-emerald-500/30 focus:border-emerald-400 focus:bg-white transition-colors"
                            />
                        </div>
                    </div>
                    <!-- Filter Bulan -->
                    <div>
                        <label for="filterBulan" class="mb-1.5 block text-xs font-bold uppercase tracking-widest text-slate-500">Bulan</label>
                        <Select id="filterBulan" bind:value={filterBulan} options={bulanOptions} class="border-slate-200 w-full" />
                    </div>
                    <!-- Reset -->
                    <div class="flex items-end">
                        <Button variant="outline" on:click={resetFilters} class="h-[42px] hover:text-red-600 hover:border-red-100 w-full">
                            <svg class="h-4 w-4 mr-2" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15" /></svg>
                            Reset
                        </Button>
                    </div>
                </div>
            </div>
            <div class="overflow-x-auto">
                <table class="w-full text-sm text-left">
                    <thead class="bg-slate-50 border-b border-slate-200 text-xs uppercase font-semibold text-slate-500 sticky top-0 z-10 shadow-sm">
                        <tr>
                            <th class="px-6 py-4 whitespace-nowrap w-[15%]">ID Luar Kota</th>
                            <th class="px-6 py-4 whitespace-nowrap w-[30%]">Lokasi</th>
                            <th class="px-6 py-4 whitespace-nowrap w-[20%]">Tanggal</th>
                            {#if $userStore?.role !== 'protokol'}
                            <th class="px-6 py-4 whitespace-nowrap text-right w-[20%]">Total Biaya Akhir</th>
                            {/if}
                            <th class="px-6 py-4 whitespace-nowrap w-[15%]">
                                <div class="flex items-center justify-center gap-1.5">
                                    Status
                                    <button type="button" title="Informasi Status" class="text-slate-400 hover:text-slate-600 bg-slate-100 hover:bg-slate-200 rounded-full p-0.5 transition-colors" on:click|preventDefault={() => openInfoModal('Legenda Status', 'Berikut adalah tahapan status beserta pemicunya:<br><br><ul class=\'list-disc pl-5 space-y-2 text-slate-600\'><li><strong>Draft:</strong> Dokumen awal baru dibuat di sistem.</li><li><strong>Pending:</strong> Dokumen perlu direvisi atau ditindaklanjuti.</li><li><strong>Submitted:</strong> Dokumen diajukan untuk reviu.</li><li><strong>Approved:</strong> Dokumen telah disetujui secara sah.</li><li><strong>Completed:</strong> Semua kegiatan telah selesai dan pelaporan SPJ ditutup.</li></ul>')}>
                                        <svg xmlns="http://www.w3.org/2000/svg" class="h-3.5 w-3.5" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13 16h-1v-4h-1m1-4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z" /></svg>
                                    </button>
                                </div>
                            </th>
                        </tr>
                    </thead>
                    <tbody class="divide-y divide-slate-100">
                        {#if filtered.length === 0}
                            <tr>
                                <td colspan="5" class="px-6 py-12 text-center text-slate-500">
                                    {#if searchQuery}
                                        Tidak ada data yang cocok dengan pencarian "<span class="font-medium">{searchQuery}</span>".
                                    {:else}
                                        Belum ada data Perjalanan Dinas Luar Kota.
                                    {/if}
                                </td>
                            </tr>
                        {:else}
                            {#each paginatedRecords as r}
                                {@const s = statusLabel(r.status)}
                                <tr class="hover:bg-slate-50/50 transition-colors">
                                    <td class="px-6 py-4 font-mono font-bold text-slate-700 text-[13px] tracking-widest">{r.spd || '-'}</td>
                                    <td class="px-6 py-4 text-slate-800 font-medium max-w-[250px]">
                                        <div class="truncate" title={r.location}>{r.location || '-'}</div>
                                        {#if r.activityName}
                                            <div class="text-xs text-slate-400 truncate mt-0.5" title={r.activityName}>{r.activityName}</div>
                                        {/if}
                                    </td>
                                    <td class="px-6 py-4 text-slate-600 whitespace-nowrap">{formatDate(r.executionDate || r.startDate)}</td>
                                    {#if $userStore?.role !== 'protokol'}
                                    <td class="px-6 py-4 text-right font-mono font-semibold text-slate-800">
                                        {formatCurrency(r.totalSpjCost || 0)}
                                    </td>
                                    {/if}
                                    <td class="px-6 py-4 text-center">
                                        <span class="inline-flex items-center rounded-full px-2.5 py-0.5 text-[10px] font-bold uppercase tracking-wide border {s.class}">
                                            {s.label}
                                        </span>
                                    </td>
                                </tr>
                            {/each}
                        {/if}
                    </tbody>
                </table>
            </div>
        </div>

        {#if totalPages > 1}
            <div class="flex items-center justify-between px-4 py-3 bg-white border border-slate-200 mt-6 rounded-xl shadow-sm no-print">
                <div class="flex flex-1 justify-between sm:hidden">
                    <Button variant="outline" size="sm" disabled={currentPage === 1} on:click={() => currentPage--}>
                        Sebelumnya
                    </Button>
                    <Button variant="outline" size="sm" disabled={currentPage === totalPages} on:click={() => currentPage++}>
                        Selanjutnya
                    </Button>
                </div>

                <div class="hidden sm:flex sm:flex-1 sm:items-center sm:justify-between">
                    <div>
                        <p class="text-sm text-slate-700">
                            Menampilkan <span class="font-medium">{(currentPage - 1) * itemsPerPage + 1}</span>
                            hingga <span class="font-medium">{Math.min(currentPage * itemsPerPage, filtered.length)}</span>
                            dari <span class="font-medium">{filtered.length}</span> hasil
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

    <!-- Info Modal -->
    <BaseModal bind:open={isInfoModalOpen} maxWidth="max-w-md">
        <svelte:fragment slot="header">
            <h3 class="text-lg font-bold text-slate-800">{infoModalTitle}</h3>
        </svelte:fragment>
        
        <svelte:fragment slot="body">
            <div class="space-y-4 text-sm text-slate-600">
                <div class="leading-relaxed">{@html infoModalContent}</div>
                
                {#if infoModalNotes}
                <div class="mt-2 bg-amber-50 border border-amber-100 p-3 rounded-lg text-amber-800">
                    <strong class="block mb-1">Catatan Penting:</strong>
                    {infoModalNotes}
                </div>
                {/if}
            </div>
        </svelte:fragment>
    </BaseModal>
{/if}
