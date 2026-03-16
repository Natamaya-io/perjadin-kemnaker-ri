<script>
    import { recordsStore, updateRecord } from '$lib/features/pengajuan/store';
    import { userStore } from '$lib/features/auth/store';
    import { onMount } from 'svelte';
    import { goto } from '$app/navigation';
    import { toast } from '$lib/shared/stores/toast';
    import { getStatusBadge } from '$lib/shared/utils/utils';
    
    // Components
    import AdminHeader from '$lib/features/admin/ui/AdminHeader.svelte';
    import AdminTableFilters from '$lib/features/admin/ui/AdminTableFilters.svelte';

    import AdminTable from '$lib/features/admin/ui/AdminTable.svelte';
    import AdminTableRow from '$lib/features/admin/ui/AdminTableRow.svelte';
    import CostModal from '$lib/features/admin/ui/CostModal.svelte';
    
    // UI Helpers
    import { ConfirmationModal } from '$lib/shared/ui/confirmation-modal';

    // Filter & Sort State
    let searchQuery = '';
    let statusFilter = 'all'; // 'all', 'Submitted', 'Approved'
    let sortOption = 'date-desc'; // 'date-desc', 'date-asc', 'cost-desc', 'cost-asc'
    let startDate = '';
    let endDate = '';

    // Modal State
    let isModalOpen = false;
    let selectedRecord = null;
    let editingCosts = {};
    let pendingGrandTotal = 0; // To store calculation result for confirmation

    let isConfirmOpen = false;

    // Accordion State
    let expandedGroups = {};

    function toggleGroup(spd) {
        expandedGroups[spd] = !expandedGroups[spd];
        expandedGroups = expandedGroups; // Trigger reactivity
    }

    // Derived Records
    $: filteredRecords = $recordsStore
        .filter(r => {
            const query = searchQuery.toLowerCase();
            const matchSearch = 
                (r.employee?.name?.toLowerCase() || '').includes(query) ||
                (r.spd?.toLowerCase() || '').includes(query) ||
                (r.location?.toLowerCase() || '').includes(query);
            
            const badge = getStatusBadge(r);
            const matchStatus = statusFilter === 'all' || badge.label === statusFilter;
            
            let matchDate = true;
            if (startDate || endDate) {
                const recordDate = new Date(r.startDate).setHours(0,0,0,0);
                const start = startDate ? new Date(startDate).setHours(0,0,0,0) : null;
                const end = endDate ? new Date(endDate).setHours(0,0,0,0) : null;

                if (start && end) {
                    matchDate = recordDate >= start && recordDate <= end;
                } else if (start) {
                    matchDate = recordDate >= start;
                } else if (end) {
                    matchDate = recordDate <= end;
                }
            }

            return matchSearch && matchStatus && matchDate;
        })
        .sort((a, b) => {
            if (sortOption === 'date-desc') return new Date(b.startDate).getTime() - new Date(a.startDate).getTime();
            if (sortOption === 'date-asc') return new Date(a.startDate).getTime() - new Date(b.startDate).getTime();
            if (sortOption === 'cost-desc') return b.totalCost - a.totalCost;
            if (sortOption === 'cost-asc') return a.totalCost - b.totalCost;
            return 0;
        });

    $: groupedRecords = filteredRecords.reduce((acc, record) => {
        if (!acc[record.spd]) {
            acc[record.spd] = { ...record, employeesList: [record] };
        } else {
            acc[record.spd].employeesList.push(record);
            // DO NOT accumulate total cost. The record totalCost is per-SPD, not per-employee.
            // Ensure we use the latest/highest cost or just the base one.
            acc[record.spd].totalCost = record.totalCost || acc[record.spd].totalCost || 0;
        }
        return acc;
    }, {});

    $: uniqueRecords = Object.values(groupedRecords);

    function openEditModal(record) {
        selectedRecord = record;
        editingCosts = { ...record.costs }; // Clone costs
        isModalOpen = true;
    }

    function handleModalSave(event) {
        const { editingCosts: newCosts, grandTotal } = event.detail;
        
        if (newCosts.localTransport > 500000) {
            toast.warning('Transport Lokal maksimal Rp 500.000');
            return;
        }

        // Store temp state for confirmation
        editingCosts = newCosts;
        pendingGrandTotal = grandTotal;
        isConfirmOpen = true;
    }

    async function processSave() {
        if (!selectedRecord) return;

        try {
            await updateRecord(selectedRecord.id, {
                costs: { ...editingCosts },
                totalCost: pendingGrandTotal,
                status: 'Approved'
            });
            
            toast.success('Rincian biaya berhasil disimpan!');
        } catch (e) {
            toast.error('Gagal menyimpan perubahan.');
        } finally {
            isModalOpen = false;
            isConfirmOpen = false; // Close confirmation modal too
        }
    }

    function generateSPD(record) {
        // Find the specific travel record ID or use the SPD string?
        // Since SPD is shared, passing spd string will just pick the first one. We should pass ID in the real app, but for now we'll pass SPD and the employee id if possible.
        // Actually, we must change the print route to use record id, but to avoid breaking print, we'll open it with the specific record.
        // For now, keeping the original behavior but warning that it prints the first match.
        // Wait, I can pass &id=record.id to print
        window.open(`/print?type=spd&id=${record.id}&spd=${encodeURIComponent(record.spd)}`, '_blank');
    }

    function generateRincian(record) {
        window.open(`/print?type=rincian&id=${record.id}&spd=${encodeURIComponent(record.spd)}`, '_blank');
    }
    onMount(() => {
        if ($userStore.role !== 'super_admin' && $userStore.role !== 'kasubag' && $userStore.role !== 'keuangan') {
            goto('/dashboard');
        }
    });
</script>

<div class="space-y-6 pb-20">
    <div class="flex flex-col sm:flex-row justify-between items-start sm:items-center gap-4 mb-6">
        <div>
            <h1 class="text-2xl font-bold text-slate-900 tracking-tight">Rekap & Kalkulasi</h1>
            <p class="text-sm text-slate-500 mt-1">Review dan verifikasi rincian biaya perjalanan dinas.</p>
        </div>
    </div>

    <div class="bg-white rounded-xl shadow-sm border border-slate-200 p-4 mb-6">
        <AdminTableFilters
            bind:searchQuery
            bind:statusFilter
            bind:sortOption
            bind:startDate
            bind:endDate
        />
    </div>
    {#if $userStore.role === 'super_admin' || $userStore.role === 'keuangan' || $userStore.role === 'kasubag'}

        <div class="bg-white rounded-xl shadow-sm border border-slate-200 overflow-hidden">
            <!-- Desktop Table View -->
            <div class="hidden md:block overflow-x-auto w-full">
                <table class="w-full text-left text-sm border-collapse min-w-[800px]">
                    <thead class="bg-slate-50 border-b border-slate-200 text-xs uppercase font-semibold text-slate-500">
                        <tr>
                            <th class="px-6 py-4 whitespace-nowrap">ID SPJ</th>
                            <th class="px-6 py-4 whitespace-nowrap">Lokasi</th>
                            <th class="px-6 py-4 whitespace-nowrap">Tanggal</th>
                            <th class="px-6 py-4 whitespace-nowrap text-right">Total Biaya Akhir</th>
                            <th class="px-6 py-4 whitespace-nowrap text-center">Status</th>
                            <th class="px-6 py-4 whitespace-nowrap text-right">Aksi</th>
                        </tr>
                    </thead>
                    <tbody class="divide-y divide-slate-100">
                        {#each uniqueRecords as record (record.spd)}
                            <!-- Group Header Row -->
                            <tr class="bg-slate-50/80 border-b border-slate-200 cursor-pointer hover:bg-slate-100 transition-colors select-none" on:click={() => toggleGroup(record.spd)}>
                                <td class="px-6 py-3 whitespace-nowrap">
                                    <div class="flex items-center gap-3">
                                        <button class="p-1 rounded-md hover:bg-slate-200 transition-colors focus:outline-none shrink-0" aria-label="Toggle details">
                                            <svg xmlns="http://www.w3.org/2000/svg" class="h-5 w-5 text-slate-500 transition-transform duration-200 {expandedGroups[record.spd] ? 'rotate-180' : ''}" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                                                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 9l-7 7-7-7" />
                                            </svg>
                                        </button>
                                        <span class="font-mono text-xs text-slate-700 font-bold">{record.spd}</span>
                                    </div>
                                </td>
                                <td class="px-6 py-3">
                                    <div class="text-sm text-slate-800 font-medium line-clamp-1" title="{record.location}, {record.province}">{record.location}, {record.province}</div>
                                </td>
                                <td class="px-6 py-3 whitespace-nowrap">
                                    <div class="text-xs text-slate-500 flex items-center gap-1.5">
                                        <svg xmlns="http://www.w3.org/2000/svg" class="h-3.5 w-3.5 text-slate-400" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                                            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M8 7V3m8 4V3m-9 8h10M5 21h14a2 2 0 002-2V7a2 2 0 00-2-2H5a2 2 0 00-2 2v12a2 2 0 002 2z" />
                                        </svg>
                                        <span>{new Date(record.startDate).toLocaleDateString('id-ID', { day: 'numeric', month: 'short', year: 'numeric'})} &mdash; {new Date(record.endDate).toLocaleDateString('id-ID', { day: 'numeric', month: 'short', year: 'numeric'})}</span>
                                    </div>
                                </td>
                                <td class="px-6 py-3 whitespace-nowrap text-right">
                                    <div class="px-3 py-1 inline-flex bg-blue-50 text-blue-700 font-mono font-bold text-xs rounded-lg border border-blue-100">
                                        {new Intl.NumberFormat('id-ID', { style: 'currency', currency: 'IDR' }).format(record.employeesList.reduce((sum, e) => sum + (e.totalCost || 0), 0))}
                                    </div>
                                </td>
                                <td class="px-6 py-3 whitespace-nowrap text-center">
                                    <span class="inline-flex items-center rounded-full px-2.5 py-1 text-[10px] font-bold uppercase tracking-wide border {getStatusBadge(record).class}">
                                        {getStatusBadge(record).label}
                                    </span>
                                </td>
                                <td class="px-6 py-3 whitespace-nowrap text-right">
                                    <div class="flex items-center justify-end gap-2">
                                        <!-- Keep empty or add action later if requested -->
                                    </div>
                                </td>
                            </tr>
                            <!-- Employee Rows -->
                            {#if expandedGroups[record.spd]}
                                <!-- Nested Table Head for Employee Data -->
                                <tr class="bg-slate-100 border-y border-slate-200 text-[11px] uppercase tracking-wider font-semibold text-slate-500 shadow-inner">
                                    <th class="px-6 py-3 text-left font-semibold">NO. SPD</th>
                                    <th class="px-6 py-3 text-left font-semibold">Nama Pegawai</th>
                                    <th class="px-6 py-3 text-left font-semibold">Nomor Telepon</th>
                                    <th class="px-6 py-3 text-right font-semibold">Total Biaya</th>
                                    <th class="px-6 py-3 text-center font-semibold">Kelengkapan</th>
                                    <th class="px-6 py-3 text-right font-semibold">Aksi</th>
                                </tr>

                                {#each record.employeesList as empRecord, index}
                                    <tr class="hover:bg-slate-50 transition-colors bg-white">
                                        <td class="px-6 py-4 align-middle">
                                            <div class="flex items-center gap-3">
                                                <div class="w-4 h-6 border-l-2 border-b-2 border-slate-200 rounded-bl-lg opacity-60 shrink-0"></div>
                                                <div class="font-mono font-bold text-slate-700">{String(index + 1).padStart(3, '0')}</div>
                                            </div>
                                        </td>
                                        <td class="px-6 py-4 align-middle">
                                            <div class="font-medium text-slate-900">{empRecord.employee?.name || '-'}</div>
                                            <div class="text-xs text-slate-500 mb-2">
                                                {#if empRecord.employee?.jabatan && empRecord.employee.jabatan !== '-'}
                                                    <div class="font-medium text-slate-600">{empRecord.employee.jabatan}</div>
                                                {/if}
                                                {#if empRecord.employee?.pangkat && empRecord.employee.pangkat !== '-' && empRecord.employee?.golongan && empRecord.employee.golongan !== '-'}
                                                    <div class="mt-0.5">{empRecord.employee.pangkat} ({empRecord.employee.golongan})</div>
                                                {:else if empRecord.employee?.pangkat && empRecord.employee.pangkat !== '-'}
                                                    <div class="mt-0.5">{empRecord.employee.pangkat}</div>
                                                {:else if empRecord.employee?.golongan && empRecord.employee.golongan !== '-'}
                                                    <div class="mt-0.5">{empRecord.employee.golongan}</div>
                                                {:else if empRecord.employee?.rank && empRecord.employee.rank !== '-'}
                                                    <div class="mt-0.5">{empRecord.employee.rank}</div>
                                                {/if}
                                            </div>
                                        </td>
                                        <td class="px-6 py-4 align-middle text-sm text-slate-600 whitespace-nowrap">
                                            <div class="flex items-center gap-1.5">
                                                <svg xmlns="http://www.w3.org/2000/svg" class="h-4 w-4 text-slate-400" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M3 5a2 2 0 012-2h3.28a1 1 0 01.948.684l1.498 4.493a1 1 0 01-.502 1.21l-2.257 1.13a11.042 11.042 0 005.516 5.516l1.13-2.257a1 1 0 011.21-.502l4.493 1.498a1 1 0 01.684.949V19a2 2 0 01-2 2h-1C9.716 21 3 14.284 3 6V5z" /></svg>
                                                {empRecord.employee?.nomorHp || '-'}
                                            </div>
                                        </td>
                                        <td class="px-6 py-4 align-middle text-right font-mono font-medium text-blue-600">
                                            {new Intl.NumberFormat('id-ID', { style: 'currency', currency: 'IDR' }).format(empRecord.totalCost || 0)}
                                        </td>
                                        <td class="px-6 py-4 align-middle text-center">
                                            <span class="inline-flex items-center rounded-full px-2.5 py-1 text-[10px] font-bold uppercase tracking-wide border {empRecord.status === 'Draft' ? 'bg-yellow-50 text-yellow-700 border-yellow-200' : 'bg-emerald-50 text-emerald-700 border-emerald-200'}">
                                                {empRecord.status === 'Draft' ? 'Belum Lengkap' : 'Lengkap'}
                                            </span>
                                        </td>
                                        <td class="px-6 py-4 align-middle">
                                            <div class="flex items-center justify-end gap-2">
                                                <a href={`/print?type=spd&id=${empRecord.id}&spd=${encodeURIComponent(empRecord.spd)}`} target="_blank" class="p-1.5 hover:bg-slate-100 rounded-md text-slate-400 hover:text-slate-700 transition-colors bg-white border border-slate-200 shadow-sm inline-flex items-center justify-center" title="Cetak SPD" on:click={(e) => e.stopPropagation()}>
                                                    <svg xmlns="http://www.w3.org/2000/svg" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M14.5 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V7.5L14.5 2z"/><polyline points="14 2 14 8 20 8"/><line x1="16" x2="8" y1="13" y2="13"/><line x1="16" x2="8" y1="17" y2="17"/><line x1="10" x2="8" y1="9" y2="9"/></svg>
                                                </a>
                                                <a href={`/print?type=rincian&id=${empRecord.id}&spd=${encodeURIComponent(empRecord.spd)}`} target="_blank" class="p-1.5 hover:bg-slate-100 rounded-md text-slate-400 hover:text-slate-700 transition-colors bg-white border border-slate-200 shadow-sm inline-flex items-center justify-center" title="Cetak Rincian" on:click={(e) => e.stopPropagation()}>
                                                    <svg xmlns="http://www.w3.org/2000/svg" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><rect width="16" height="20" x="4" y="2" rx="2"/><line x1="8" x2="16" y1="6" y2="6"/><line x1="16" x2="16" y1="14" y2="18"/><path d="M16 10h.01"/><path d="M12 10h.01"/><path d="M8 10h.01"/><path d="M12 14h.01"/><path d="M8 14h.01"/><path d="M12 18h.01"/><path d="M8 18h.01"/></svg>
                                                </a>
                                                {#if $userStore.role !== 'kasubag'}
                                                    {#if empRecord.status === 'Submitted'}
                                                        <button class="bg-white border border-slate-200 text-blue-600 hover:bg-blue-50 hover:border-blue-200 px-3 py-1.5 rounded-md text-xs font-medium transition-all shadow-sm whitespace-nowrap" on:click={(e) => { e.stopPropagation(); openEditModal(empRecord); }}>
                                                            Review Biaya
                                                        </button>
                                                    {:else if empRecord.status === 'Approved'}
                                                        <button class="bg-white border border-emerald-200 text-emerald-600 hover:bg-emerald-50 hover:border-emerald-300 px-3 py-1.5 rounded-md text-xs font-medium transition-all shadow-sm whitespace-nowrap" on:click={(e) => { e.stopPropagation(); openEditModal(empRecord); }}>
                                                            Edit Review
                                                        </button>
                                                    {/if}
                                                {/if}
                                            </div>
                                        </td>
                                    </tr>
                                {/each}
                            {/if}
                        {/each}
                        {#if uniqueRecords.length === 0}
                            <tr>
                                <td colspan="5" class="p-12 text-center text-slate-500 bg-slate-50/50">
                                    Belum ada pengajuan yang masuk.
                                </td>
                            </tr>
                        {/if}
                    </tbody>
                </table>
            </div>

            <!-- Mobile Stacked/Card View -->
            <div class="md:hidden flex flex-col divide-y divide-slate-100 bg-slate-50">
                {#each uniqueRecords as record (record.spd)}
                    <div class="flex flex-col">
                        <!-- Group Header (Mobile) -->
                        <div class="p-4 bg-slate-100/80 border-b border-slate-200 cursor-pointer hover:bg-slate-200 transition-colors select-none" on:click={() => toggleGroup(record.spd)}>
                            <div class="flex justify-between items-start mb-2 gap-2">
                                <div class="flex items-center flex-wrap gap-2">
                                    <span class="font-mono text-xs text-slate-600 font-semibold bg-white px-1.5 py-0.5 rounded border border-slate-200 shadow-sm">{record.spd}</span>
                                    <span class="inline-flex items-center rounded-full px-2 py-0.5 text-[10px] font-bold capitalize tracking-wide bg-indigo-50 text-indigo-700 border border-indigo-200">
                                        {record.type ? record.type.replace(/_/g, ' ') : 'Dalam Kota'}
                                    </span>
                                </div>
                                <button class="p-1 rounded-md bg-white border border-slate-200 shadow-sm hover:bg-slate-50 transition-colors focus:outline-none" aria-label="Toggle details">
                                    <svg xmlns="http://www.w3.org/2000/svg" class="h-4 w-4 text-slate-500 transition-transform duration-200 {expandedGroups[record.spd] ? 'rotate-180' : ''}" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                                        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 9l-7 7-7-7" />
                                    </svg>
                                </button>
                            </div>
                            <div class="text-sm font-semibold text-slate-800 leading-tight mb-2 pr-6">{record.purpose}</div>
                            <div class="flex flex-col gap-1 text-xs text-slate-600">
                                <div class="flex items-center gap-1.5">
                                    <svg xmlns="http://www.w3.org/2000/svg" class="h-3.5 w-3.5 text-slate-400 shrink-0" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                                        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M17.657 16.657L13.414 20.9a1.998 1.998 0 01-2.827 0l-4.244-4.243a8 8 0 1111.314 0z" />
                                        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 11a3 3 0 11-6 0 3 3 0 016 0z" />
                                    </svg>
                                    <span class="truncate">{record.location}, {record.province}</span>
                                </div>
                                <div class="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-1.5 sm:gap-0 mt-1">
                                    <div class="flex items-center gap-1.5 text-slate-500 min-w-0">
                                        <svg xmlns="http://www.w3.org/2000/svg" class="h-3.5 w-3.5 shrink-0" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                                            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M8 7V3m8 4V3m-9 8h10M5 21h14a2 2 0 002-2V7a2 2 0 00-2-2H5a2 2 0 00-2 2v12a2 2 0 002 2z" />
                                        </svg>
                                        <span class="truncate">{new Date(record.startDate).toLocaleDateString('id-ID', { day: 'numeric', month: 'short', year: 'numeric'})} &mdash; {new Date(record.endDate).toLocaleDateString('id-ID', { day: 'numeric', month: 'short', year: 'numeric'})}</span>
                                    </div>
                                    <span class="font-semibold text-[10px] text-slate-500 uppercase tracking-wider shrink-0">{record.employeesList.length} Petugas</span>
                                </div>
                            </div>
                        </div>

                        <!-- Employees (Mobile) -->
                        {#if expandedGroups[record.spd]}
                            <div class="flex flex-col divide-y divide-slate-100 bg-white">
                                {#each record.employeesList as empRecord}
                                    <div class="p-4 flex flex-col gap-3">
                                        <div class="flex justify-between items-start gap-2">
                                            <div class="flex flex-col">
                                                <span class="text-[10px] font-semibold text-slate-400 uppercase tracking-wider mb-0.5">Pegawai</span>
                                                <div class="font-medium text-sm text-slate-900 leading-tight">{empRecord.employee?.name || '-'}</div>
                                                <div class="text-xs text-slate-500 mb-2">
                                                    {#if empRecord.employee?.jabatan}
                                                        <div class="font-medium text-slate-700">{empRecord.employee.jabatan}</div>
                                                    {/if}
                                                    {#if (empRecord.employee?.pangkat && empRecord.employee.pangkat !== '-') || (empRecord.employee?.golongan && empRecord.employee.golongan !== '-') || (empRecord.employee?.rank && empRecord.employee.rank !== '-')}
                                                        <div class="mt-0.5">
                                                            {#if (empRecord.employee?.pangkat && empRecord.employee.pangkat !== '-') && (empRecord.employee?.golongan && empRecord.employee.golongan !== '-')}
                                                                {empRecord.employee.pangkat} ({empRecord.employee.golongan})
                                                            {:else}
                                                                {(empRecord.employee?.pangkat !== '-' ? empRecord.employee?.pangkat : null) || (empRecord.employee?.golongan !== '-' ? empRecord.employee?.golongan : null) || (empRecord.employee?.rank !== '-' ? empRecord.employee?.rank : '')}
                                                            {/if}
                                                        </div>
                                                    {/if}
                                                </div>
                                            </div>
                                            <span class="shrink-0 inline-flex items-center rounded-full px-2 py-0.5 text-[9px] font-bold uppercase tracking-wide border {getStatusBadge(empRecord).class}">
                                                {getStatusBadge(empRecord).label}
                                            </span>
                                        </div>

                                        <div class="flex flex-col gap-3 border-t border-slate-50 pt-3">
                                            <div class="flex justify-between items-center">
                                                <span class="text-[10px] font-semibold text-slate-400 uppercase tracking-wider mb-0.5">Total Biaya</span>
                                                <span class="font-mono font-bold text-blue-600 text-sm">
                                                    {new Intl.NumberFormat('id-ID', { style: 'currency', currency: 'IDR' }).format(empRecord.totalCost || 0)}
                                                </span>
                                            </div>
                                            <div class="flex items-center gap-2 w-full">
                                                <a href={`/print?type=spd&id=${empRecord.id}&spd=${encodeURIComponent(empRecord.spd)}`} target="_blank" class="p-1.5 hover:bg-slate-100 rounded-md text-slate-400 hover:text-slate-700 transition-colors bg-white border border-slate-200 shadow-sm inline-flex items-center justify-center" title="Cetak SPD" on:click={(e) => e.stopPropagation()}>
                                                    <svg xmlns="http://www.w3.org/2000/svg" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M14.5 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V7.5L14.5 2z"/><polyline points="14 2 14 8 20 8"/><line x1="16" x2="8" y1="13" y2="13"/><line x1="16" x2="8" y1="17" y2="17"/><line x1="10" x2="8" y1="9" y2="9"/></svg>
                                                </a>
                                                <a href={`/print?type=rincian&id=${empRecord.id}&spd=${encodeURIComponent(empRecord.spd)}`} target="_blank" class="p-1.5 hover:bg-slate-100 rounded-md text-slate-400 hover:text-slate-700 transition-colors bg-white border border-slate-200 shadow-sm inline-flex items-center justify-center" title="Cetak Rincian" on:click={(e) => e.stopPropagation()}>
                                                    <svg xmlns="http://www.w3.org/2000/svg" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><rect width="16" height="20" x="4" y="2" rx="2"/><line x1="8" x2="16" y1="6" y2="6"/><line x1="16" x2="16" y1="14" y2="18"/><path d="M16 10h.01"/><path d="M12 10h.01"/><path d="M8 10h.01"/><path d="M12 14h.01"/><path d="M8 14h.01"/><path d="M12 18h.01"/><path d="M8 18h.01"/></svg>
                                                </a>
                                                {#if $userStore.role !== 'kasubag'}
                                                    {#if empRecord.status === 'Submitted'}
                                                        <button class="bg-white border border-slate-200 text-blue-600 hover:bg-blue-50 hover:border-blue-200 px-3 py-1.5 rounded-md text-xs font-medium transition-all shadow-sm flex-1" on:click={(e) => { e.stopPropagation(); openEditModal(empRecord); }}>
                                                            Review Biaya
                                                        </button>
                                                    {:else if empRecord.status === 'Approved'}
                                                        <button class="bg-white border border-emerald-200 text-emerald-600 hover:bg-emerald-50 hover:border-emerald-300 px-3 py-1.5 rounded-md text-xs font-medium transition-all shadow-sm flex-1" on:click={(e) => { e.stopPropagation(); openEditModal(empRecord); }}>
                                                            Edit Review
                                                        </button>
                                                    {/if}
                                                {/if}
                                            </div>
                                        </div>
                                    </div>
                                {/each}
                            </div>
                        {/if}
                    </div>
                {/each}
                {#if uniqueRecords.length === 0}
                    <div class="p-12 text-center text-slate-500 bg-white">
                        Belum ada pengajuan yang masuk.
                    </div>
                {/if}
            </div>
        </div>
    {:else}
        <div class="flex flex-col items-center justify-center p-12 text-center border-2 border-dashed border-slate-200 rounded-xl bg-slate-50">
            <svg xmlns="http://www.w3.org/2000/svg" class="h-12 w-12 text-slate-300 mb-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 15v2m-6 4h12a2 2 0 002-2v-6a2 2 0 00-2-2H6a2 2 0 00-2 2v6a2 2 0 002 2zm10-10V7a4 4 0 00-8 0v4h8z" />
            </svg>
            <h3 class="text-lg font-medium text-slate-900">Akses Dibatasi</h3>
            <p class="text-slate-500 max-w-sm mt-1">Halaman ini khusus untuk Admin Keuangan.</p>
        </div>
    {/if}

    <CostModal 
        bind:open={isModalOpen} 
        record={selectedRecord}
        bind:editingCosts={editingCosts}
        on:close={() => isModalOpen = false}
        on:save={handleModalSave}
    />

    <ConfirmationModal
        bind:open={isConfirmOpen}
        title="Simpan Rincian Biaya"
        description="Apakah Anda yakin data rincian biaya ini sudah sesuai? Status akan diubah menjadi Approved."
        confirmText="Ya, Simpan & Approve"
        onConfirm={processSave}
    />
</div>
