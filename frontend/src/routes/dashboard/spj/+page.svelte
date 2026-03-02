<script>
    import { recordsStore, updateRecord, deleteRecordBySpd } from '$lib/stores/records';
    import { userStore } from '$lib/stores/auth';
    import { provincesStore, stakeholdersStore } from '$lib/stores/master-data';
    import { toast } from '$lib/stores/toast';
    import { getStatusBadge } from '$lib/utils';

    // Components
    import AdminHeader from '$lib/components/dashboard/admin/AdminHeader.svelte';
    import AdminTableFilters from '$lib/components/dashboard/admin/AdminTableFilters.svelte';

    import AdminTable from '$lib/components/dashboard/admin/AdminTable.svelte';
    import AdminTableRow from '$lib/components/dashboard/admin/AdminTableRow.svelte';
    import CostModal from '$lib/components/dashboard/admin/CostModal.svelte';
    import ReviewModal from '$lib/components/dashboard/pengajuan/ReviewModal.svelte';    
    // UI Helpers
    import { ConfirmationModal } from '$lib/components/ui/confirmation-modal';
    
    import TripStepper from '$lib/components/dashboard/roadmap/TripStepper.svelte';

    // Filter & Sort State
    let searchQuery = '';
    let statusFilter = 'all'; // 'all', 'In Progress', 'Completed'
    let sortOption = 'date-desc'; // 'date-desc', 'date-asc', 'cost-desc', 'cost-asc'
    let startDate = '';
    let endDate = '';

    // Modal State
    let isModalOpen = false;
    let selectedRecord = null;
    let editingCosts = {};
    let pendingGrandTotal = 0; // To store calculation result for confirmation

    let isConfirmOpen = false;
    let isDeleteModalOpen = false;
    let selectedSpdToDelete = null;

    let isReviewModalOpen = false;
    let selectedReviewRecords = [];

    // Accordion State
    let expandedGroups = {};

    function toggleGroup(spd) {
        expandedGroups[spd] = !expandedGroups[spd];
        expandedGroups = expandedGroups; // Trigger reactivity
    }

    function openReviewModal(spd) {
        selectedReviewRecords = groupedRecords[spd].employeesList;
        isReviewModalOpen = true;
    }

    // Derived Records - Filter based on role
    $: myRecords = $recordsStore.filter(r => {
        if ($userStore.role === 'super_admin' || $userStore.role === 'kasubag') return true;
        // Show if the user is the employee OR if the user created the record
        const isEmployee = r.employee?.email === $userStore.email;
        const isCreator = (r.creator?.email || r.email) === $userStore.email || r.creatorId === $userStore.id;
        return isEmployee || isCreator;
    });

    $: filteredRecords = myRecords
        .filter(r => {
            const query = searchQuery.toLowerCase();
            const matchSearch = 
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
        // Protokol can edit if status is 'Draft' or 'Submitted'
        if (record.status === 'Approved') {
            toast.warning('Pengajuan sudah di-approve oleh Keuangan dan tidak dapat diedit lagi.');
            return;
        }

        selectedRecord = record;
        editingCosts = { ...record.costs }; // Clone costs
        isModalOpen = true;
    }

    function handleModalSave(event) {
        const { editingCosts: newCosts, grandTotal } = event.detail;

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
                status: 'Submitted' // Ensure status stays 'Submitted' so Keuangan can review
            });
            
            toast.success('Rincian biaya berhasil disimpan dan siap direview Keuangan!');
        } catch (e) {
            toast.error('Gagal menyimpan perubahan.');
        } finally {
            isModalOpen = false;
            isConfirmOpen = false; // Close confirmation modal too
        }
    }

    function confirmDeleteGroup(spd, e) {
        e.stopPropagation();
        selectedSpdToDelete = spd;
        isDeleteModalOpen = true;
    }

    async function processDeleteGroup() {
        if (!selectedSpdToDelete) return;
        try {
            await deleteRecordBySpd(selectedSpdToDelete);
            toast.success('Perjalanan dinas berhasil dihapus.');
        } catch (e) {
            toast.error('Gagal menghapus perjalanan dinas.');
        } finally {
            isDeleteModalOpen = false;
            selectedSpdToDelete = null;
        }
    }

    function generateSPD(record) {
        window.open(`/print?type=spd&id=${record.id}&spd=${encodeURIComponent(record.spd)}`, '_blank');
    }

    function generateRincian(record) {
        window.open(`/print?type=rincian&id=${record.id}&spd=${encodeURIComponent(record.spd)}`, '_blank');
    }
</script>

<div class="space-y-6 pb-20">
    <div class="flex flex-col sm:flex-row justify-between items-start sm:items-center gap-4 mb-6">
        <div>
            <h1 class="text-2xl font-bold text-slate-900 tracking-tight">Input SPJ</h1>
            <p class="text-sm text-slate-500 mt-1">Input rincian biaya, dokumen bukti, dan tagihan perjalanan dinas Anda di sini.</p>
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

    <div class="bg-white rounded-xl shadow-sm border border-slate-200 overflow-hidden">
        <!-- Desktop Table View -->
        <div class="hidden md:block overflow-x-auto w-full">
            <table class="w-full text-left text-sm border-collapse min-w-[800px]">
                <thead class="bg-slate-50 border-b border-slate-200 text-xs uppercase font-semibold text-slate-500">
                    <tr>
                        <th class="px-6 py-4 whitespace-nowrap">ID SPD</th>
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
                                    <button class="bg-white border border-slate-200 text-blue-600 hover:bg-blue-50 hover:border-blue-200 px-3 py-1.5 rounded-lg text-xs font-medium transition-all shadow-sm whitespace-nowrap inline-flex items-center gap-1.5" on:click|stopPropagation={() => openReviewModal(record.spd)}>
                                        <svg xmlns="http://www.w3.org/2000/svg" class="h-3.5 w-3.5 shrink-0" viewBox="0 0 20 20" fill="currentColor">
                                            <path d="M10 12a2 2 0 100-4 2 2 0 000 4z" />
                                            <path fill-rule="evenodd" d="M.458 10C1.732 5.943 5.522 3 10 3s8.268 2.943 9.542 7c-1.274 4.057-5.064 7-9.542 7S1.732 14.057.458 10zM14 10a4 4 0 11-8 0 4 4 0 018 0z" clip-rule="evenodd" />
                                        </svg>
                                        Review
                                    </button>
                                    {#if $userStore.role === 'super_admin'}
                                        <button class="bg-white border border-rose-200 text-rose-600 hover:bg-rose-50 hover:border-rose-300 px-3 py-1.5 rounded-lg text-xs font-medium transition-all shadow-sm whitespace-nowrap inline-flex items-center gap-1.5" on:click={(e) => confirmDeleteGroup(record.spd, e)}>
                                            <svg xmlns="http://www.w3.org/2000/svg" class="h-3.5 w-3.5 shrink-0" viewBox="0 0 20 20" fill="currentColor">
                                                <path fill-rule="evenodd" d="M9 2a1 1 0 00-.894.553L7.382 4H4a1 1 0 000 2v10a2 2 0 002 2h8a2 2 0 002-2V6a1 1 0 100-2h-3.382l-.724-1.447A1 1 0 0011 2H9zM7 8a1 1 0 012 0v6a1 1 0 11-2 0V8zm5-1a1 1 0 00-1 1v6a1 1 0 102 0V8a1 1 0 00-1-1z" clip-rule="evenodd" />
                                            </svg>
                                            Hapus
                                        </button>
                                    {/if}
                                </div>
                            </td>
                        </tr>
                        <!-- Employee Rows -->
                        {#if expandedGroups[record.spd]}
                            {#if $userStore.role === 'protokol'}
                                <!-- Roadmap / Stepper (Group level for Protokol) -->
                                <tr class="bg-white border-b border-slate-100">
                                    <td colspan="6" class="px-6 pb-2 pt-0">
                                        <div class="text-[10px] uppercase font-bold text-slate-400 tracking-wider mb-2 mt-4">Roadmap Perjalanan</div>
                                        <TripStepper record={record.employeesList[0]} on:input-spj={(e) => { e.stopPropagation(); openEditModal(record.employeesList[0]); }} />
                                    </td>
                                </tr>
                            {/if}
                            
                            <!-- Nested Table Head for Employee Data -->
                            <tr class="bg-slate-100 border-y border-slate-200 text-[11px] uppercase tracking-wider font-semibold text-slate-500 shadow-inner">
                                <th class="px-6 py-3 text-left font-semibold">NO. SPD</th>
                                <th class="px-6 py-3 text-left font-semibold">Nama Pegawai</th>
                                <th class="px-6 py-3 text-left font-semibold">Nomor Telepon</th>
                                <th class="px-6 py-3 text-right font-semibold">Total Biaya</th>
                                <th class="px-6 py-3 text-center font-semibold">Kelengkapan</th>
                                <th class="px-6 py-3 text-right font-semibold">Aksi</th>
                            </tr>

                            <!-- Employees List Header (optional, but keep design simple) -->
                            {#each record.employeesList as empRecord, index}
                                {#if $userStore.role === 'super_admin' || $userStore.role === 'kasubag'}
                                    <tr class="bg-white border-b border-slate-50">
                                        <td colspan="6" class="px-6 pb-0 pt-3">
                                            <div class="text-[10px] uppercase font-bold text-slate-400 tracking-wider mb-1">Roadmap Pegawai</div>
                                            <TripStepper record={empRecord} />
                                        </td>
                                    </tr>
                                {/if}
                                <tr class="hover:bg-slate-50 transition-colors bg-white">
                                    <td class="px-6 py-4 align-middle">
                                        <div class="flex items-center gap-3">
                                            <div class="w-4 h-6 border-l-2 border-b-2 border-slate-200 rounded-bl-lg opacity-60 shrink-0"></div>
                                            <div class="font-mono font-bold text-slate-700">{String(index + 1).padStart(3, '0')}</div>
                                        </div>
                                    </td>
                                    <td class="px-6 py-4 align-middle">
                                        <div class="font-medium text-slate-900">{empRecord.employee?.name || '-'}</div>
                                        <div class="text-xs text-slate-500 mt-0.5">
                                            {#if empRecord.employee?.jabatan && empRecord.employee.jabatan !== '-'}
                                                <div class="font-medium text-slate-600">{empRecord.employee.jabatan}</div>
                                            {/if}
                                            {#if empRecord.employee?.pangkat && empRecord.employee.pangkat !== '-' && empRecord.employee?.golongan && empRecord.employee.golongan !== '-'}
                                                <div>{empRecord.employee.pangkat} ({empRecord.employee.golongan})</div>
                                            {:else if empRecord.employee?.pangkat && empRecord.employee.pangkat !== '-'}
                                                <div>{empRecord.employee.pangkat}</div>
                                            {:else if empRecord.employee?.golongan && empRecord.employee.golongan !== '-'}
                                                <div>{empRecord.employee.golongan}</div>
                                            {:else if empRecord.employee?.rank && empRecord.employee.rank !== '-'}
                                                <div>{empRecord.employee.rank}</div>
                                            {/if}
                                        </div>
                                    </td>
                                    <td class="px-6 py-4 align-middle text-sm text-slate-600 whitespace-nowrap">
                                        <div class="flex items-center gap-1.5">
                                            <svg xmlns="http://www.w3.org/2000/svg" class="h-4 w-4 text-slate-400" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M3 5a2 2 0 012-2h3.28a1 1 0 01.948.684l1.498 4.493a1 1 0 01-.502 1.21l-2.257 1.13a11.042 11.042 0 005.516 5.516l1.13-2.257a1 1 0 011.21-.502l4.493 1.498a1 1 0 01.684.949V19a2 2 0 01-2 2h-1C9.716 21 3 14.284 3 6V5z" /></svg>
                                            {empRecord.employee?.nomorHp || '-'}
                                        </div>
                                    </td>
                                    <td class="px-6 py-4 align-middle text-right font-mono font-medium text-slate-800">
                                        {new Intl.NumberFormat('id-ID', { style: 'currency', currency: 'IDR' }).format(empRecord.totalCost || 0)}
                                    </td>
                                    <td class="px-6 py-4 align-middle text-center">
                                        <span class="inline-flex items-center rounded-full px-2.5 py-1 text-[10px] font-bold uppercase tracking-wide border {empRecord.status === 'Draft' ? 'bg-yellow-50 text-yellow-700 border-yellow-200' : 'bg-emerald-50 text-emerald-700 border-emerald-200'}">
                                            {empRecord.status === 'Draft' ? 'Belum Lengkap' : 'Lengkap'}
                                        </span>
                                    </td>
                                    <td class="px-6 py-4 align-middle text-right">
                                        {#if $userStore.role !== 'kasubag'}
                                            <button class="bg-white border border-slate-200 text-blue-600 hover:bg-blue-50 hover:border-blue-200 px-3 py-1.5 rounded-lg text-xs font-medium transition-all shadow-sm whitespace-nowrap inline-flex items-center justify-end gap-1.5" on:click={(e) => { e.stopPropagation(); openEditModal(empRecord); }}>
                                                <svg xmlns="http://www.w3.org/2000/svg" class="h-3.5 w-3.5 shrink-0" viewBox="0 0 20 20" fill="currentColor">
                                                    <path d="M13.586 3.586a2 2 0 112.828 2.828l-.793.793-2.828-2.828.793-.793zM11.379 5.793L3 14.172V17h2.828l8.38-8.379-2.83-2.828z" />
                                                </svg>
                                                {$userStore.role === 'protokol' ? (empRecord.status === 'Draft' ? 'Input Rincian Biaya' : 'Edit Rincian Biaya') : 'Review'}
                                            </button>
                                        {/if}
                                    </td>
                                </tr>
                            {/each}
                        {/if}
                    {/each}
                    {#if uniqueRecords.length === 0}
                        <tr>
                            <td colspan="6" class="p-12 text-center text-slate-500 bg-slate-50/50">
                                <div class="flex flex-col items-center justify-center">
                                    <svg xmlns="http://www.w3.org/2000/svg" class="h-10 w-10 text-slate-300 mb-3" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                                        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z" />
                                    </svg>
                                    Belum ada data perjalanan dinas untuk Anda.
                                </div>
                            </td>
                        </tr>
                    {/if}
                </tbody>
            </table>
        </div>

        <!-- Mobile Stacked/Card View -->
        <div class="md:hidden flex flex-col gap-4 bg-slate-50/50 p-4">
            {#each uniqueRecords as record (record.spd)}
                <div class="bg-white rounded-xl shadow-sm border border-slate-200 overflow-hidden">
                    <!-- Group Header (Mobile) -->
                    <div class="p-4 bg-slate-50/80 cursor-pointer hover:bg-slate-100 transition-colors select-none" on:click={() => toggleGroup(record.spd)}>
                        <div class="flex justify-between items-start mb-3 gap-2">
                            <div class="flex flex-col gap-1.5">
                                <div class="flex items-center flex-wrap gap-2">
                                    <span class="font-mono text-xs text-slate-700 font-bold">{record.spd}</span>
                                </div>
                                <div class="text-sm font-semibold text-slate-800 leading-snug line-clamp-2">{record.location}, {record.province}</div>
                            </div>
                            <button class="p-1.5 rounded-lg bg-white border border-slate-200 shadow-sm hover:bg-slate-50 transition-colors focus:outline-none shrink-0" aria-label="Toggle details">
                                <svg xmlns="http://www.w3.org/2000/svg" class="h-4 w-4 text-slate-500 transition-transform duration-300 {expandedGroups[record.spd] ? 'rotate-180' : ''}" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 9l-7 7-7-7" />
                                </svg>
                            </button>
                        </div>

                        <div class="grid grid-cols-2 gap-3 mb-3">
                            <div class="flex flex-col gap-1">
                                <span class="text-[9px] font-semibold text-slate-400 uppercase tracking-wider">Status Terkini</span>
                                <div>
                                    <span class="inline-flex items-center rounded-full px-2 py-0.5 text-[9px] font-bold uppercase tracking-wide border {getStatusBadge(record).class}">
                                        {getStatusBadge(record).label}
                                    </span>
                                </div>
                            </div>
                            <div class="flex flex-col gap-1">
                                <span class="text-[9px] font-semibold text-slate-400 uppercase tracking-wider">Tanggal</span>
                                <div class="text-xs text-slate-700 flex items-start gap-1">
                                    <svg xmlns="http://www.w3.org/2000/svg" class="h-3.5 w-3.5 text-slate-400 shrink-0 mt-0.5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                                        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M8 7V3m8 4V3m-9 8h10M5 21h14a2 2 0 002-2V7a2 2 0 00-2-2H5a2 2 0 00-2 2v12a2 2 0 002 2z" />
                                    </svg>
                                    <div class="flex flex-col leading-tight gap-0.5">
                                        <span>{new Date(record.startDate).toLocaleDateString('id-ID', { day: 'numeric', month: 'short', year: 'numeric'})}</span>
                                        <span class="text-[9px] text-slate-400">s/d</span>
                                        <span>{new Date(record.endDate).toLocaleDateString('id-ID', { day: 'numeric', month: 'short', year: 'numeric'})}</span>
                                    </div>
                                </div>
                            </div>
                        </div>

                        <div class="flex items-center justify-between pt-3 border-t border-slate-200/60">
                            <div class="flex gap-2">
                                <button class="bg-white border border-slate-200 text-blue-600 hover:bg-blue-50 hover:border-blue-200 px-3 py-1.5 rounded-lg text-xs font-medium transition-all shadow-sm whitespace-nowrap inline-flex items-center gap-1.5" on:click|stopPropagation={() => openReviewModal(record.spd)}>
                                    <svg xmlns="http://www.w3.org/2000/svg" class="h-3.5 w-3.5 shrink-0" viewBox="0 0 20 20" fill="currentColor">
                                        <path d="M10 12a2 2 0 100-4 2 2 0 000 4z" />
                                        <path fill-rule="evenodd" d="M.458 10C1.732 5.943 5.522 3 10 3s8.268 2.943 9.542 7c-1.274 4.057-5.064 7-9.542 7S1.732 14.057.458 10zM14 10a4 4 0 11-8 0 4 4 0 018 0z" clip-rule="evenodd" />
                                    </svg>
                                    Review
                                </button>
                                {#if $userStore.role === 'super_admin' || $userStore.role === 'kasubag'}
                                    <button class="bg-white border border-rose-200 text-rose-600 hover:bg-rose-50 hover:border-rose-300 px-3 py-1.5 rounded-lg text-xs font-medium transition-all shadow-sm whitespace-nowrap inline-flex items-center gap-1.5" on:click={(e) => confirmDeleteGroup(record.spd, e)}>
                                        <svg xmlns="http://www.w3.org/2000/svg" class="h-3.5 w-3.5 shrink-0" viewBox="0 0 20 20" fill="currentColor">
                                            <path fill-rule="evenodd" d="M9 2a1 1 0 00-.894.553L7.382 4H4a1 1 0 000 2v10a2 2 0 002 2h8a2 2 0 002-2V6a1 1 0 100-2h-3.382l-.724-1.447A1 1 0 0011 2H9zM7 8a1 1 0 012 0v6a1 1 0 11-2 0V8zm5-1a1 1 0 00-1 1v6a1 1 0 102 0V8a1 1 0 00-1-1z" clip-rule="evenodd" />
                                        </svg>
                                        Hapus
                                    </button>
                                {/if}
                            </div>
                            <div class="flex flex-col items-end">
                                <span class="text-[9px] font-semibold text-slate-500 uppercase tracking-wider mb-0.5">Total Biaya Grup</span>
                                <span class="font-mono font-bold text-blue-700 text-sm">
                                    {new Intl.NumberFormat('id-ID', { style: 'currency', currency: 'IDR' }).format(record.employeesList.reduce((sum, e) => sum + (e.totalCost || 0), 0))}
                                </span>
                            </div>
                        </div>
                    </div>

                    <!-- Employees (Mobile) -->
                    {#if expandedGroups[record.spd]}
                        <div class="flex flex-col divide-y divide-slate-100 bg-white border-t border-slate-200">
                            <!-- Roadmap / Stepper -->
                            <div class="p-4 bg-slate-50/50">
                                <div class="text-[10px] uppercase font-bold text-slate-400 tracking-wider mb-2">Roadmap Perjalanan</div>
                                <TripStepper record={record.employeesList[0]} />
                            </div>

                            {#each record.employeesList as empRecord, index}
                                <div class="p-4 flex flex-col gap-3 hover:bg-slate-50 transition-colors">
                                    <div class="flex justify-between items-start gap-2">
                                        <div class="flex flex-col">
                                            <div class="flex items-center gap-2 mb-1.5">
                                                <span class="text-[9px] font-mono font-bold bg-slate-100 text-slate-500 px-1.5 py-0.5 rounded border border-slate-200 shadow-sm">NO. SPD: {String(index + 1).padStart(3, '0')}</span>
                                            </div>
                                            <div class="font-semibold text-sm text-slate-900 leading-tight">{empRecord.employee?.name || '-'}</div>
                                            <div class="text-xs text-slate-500 mt-1">
                                                {#if empRecord.employee?.jabatan && empRecord.employee.jabatan !== '-'}
                                                    <div class="font-medium text-slate-600">{empRecord.employee.jabatan}</div>
                                                {/if}
                                                {#if empRecord.employee?.pangkat && empRecord.employee.pangkat !== '-' && empRecord.employee?.golongan && empRecord.employee.golongan !== '-'}
                                                    <div>{empRecord.employee.pangkat} ({empRecord.employee.golongan})</div>
                                                {:else if empRecord.employee?.pangkat && empRecord.employee.pangkat !== '-'}
                                                    <div>{empRecord.employee.pangkat}</div>
                                                {:else if empRecord.employee?.golongan && empRecord.employee.golongan !== '-'}
                                                    <div>{empRecord.employee.golongan}</div>
                                                {:else if empRecord.employee?.rank && empRecord.employee.rank !== '-'}
                                                    <div>{empRecord.employee.rank}</div>
                                                {/if}
                                            </div>
                                            <div class="flex items-center gap-1.5 mt-1.5 text-xs text-slate-600">
                                                <svg xmlns="http://www.w3.org/2000/svg" class="h-3.5 w-3.5 text-slate-400" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M3 5a2 2 0 012-2h3.28a1 1 0 01.948.684l1.498 4.493a1 1 0 01-.502 1.21l-2.257 1.13a11.042 11.042 0 005.516 5.516l1.13-2.257a1 1 0 011.21-.502l4.493 1.498a1 1 0 01.684.949V19a2 2 0 01-2 2h-1C9.716 21 3 14.284 3 6V5z" /></svg>
                                                {empRecord.employee?.nomorHp || '-'}
                                            </div>
                                        </div>
                                        <div class="flex flex-col items-end gap-2">
                                            <span class="shrink-0 inline-flex items-center rounded-full px-2.5 py-1 text-[9px] font-bold uppercase tracking-wide border {empRecord.status === 'Draft' ? 'bg-yellow-50 text-yellow-700 border-yellow-200' : 'bg-emerald-50 text-emerald-700 border-emerald-200'}">
                                                {empRecord.status === 'Draft' ? 'Belum Lengkap' : 'Lengkap'}
                                            </span>
                                        </div>
                                    </div>

                                    <div class="flex justify-between items-end bg-slate-50 rounded-xl p-3 border border-slate-100 mt-1 shadow-sm">
                                        <div class="flex flex-col">
                                            <span class="text-[9px] font-semibold text-slate-400 uppercase tracking-wider mb-1">Total Biaya Pegawai</span>
                                            <span class="font-mono font-bold text-slate-800 text-sm">
                                                {new Intl.NumberFormat('id-ID', { style: 'currency', currency: 'IDR' }).format(empRecord.totalCost || 0)}
                                            </span>
                                        </div>
                                        {#if $userStore.role !== 'kasubag'}
                                            <button class="bg-white border border-slate-200 text-blue-600 hover:bg-blue-50 hover:border-blue-200 px-3 py-1.5 rounded-lg text-xs font-medium transition-all shadow-sm flex items-center gap-1.5" on:click={(e) => { e.stopPropagation(); openEditModal(empRecord); }}>
                                                <svg xmlns="http://www.w3.org/2000/svg" class="h-3.5 w-3.5" viewBox="0 0 20 20" fill="currentColor">
                                                    <path d="M13.586 3.586a2 2 0 112.828 2.828l-.793.793-2.828-2.828.793-.793zM11.379 5.793L3 14.172V17h2.828l8.38-8.379-2.83-2.828z" />
                                                </svg>
                                                {$userStore.role === 'protokol' ? (empRecord.status === 'Draft' ? 'Input Rincian Biaya' : 'Edit Rincian Biaya') : 'Review'}
                                            </button>
                                        {/if}
                                    </div>
                                </div>
                            {/each}
                        </div>
                    {/if}
                </div>
            {/each}
            {#if uniqueRecords.length === 0}
                <div class="p-8 text-center text-slate-500 bg-white rounded-xl shadow-sm border border-slate-200">
                    <svg xmlns="http://www.w3.org/2000/svg" class="h-10 w-10 mx-auto text-slate-300 mb-3" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z" />
                    </svg>
                    Belum ada data perjalanan dinas untuk Anda.
                </div>
            {/if}
        </div>
    </div>

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
        description="Apakah Anda yakin rincian biaya yang Anda input sudah sesuai? Data akan siap direview oleh Keuangan."
        confirmText="Ya, Simpan"
        onConfirm={processSave}
    />

    <ConfirmationModal
        bind:open={isDeleteModalOpen}
        title="Hapus Perjalanan Dinas"
        description="Apakah Anda yakin ingin menghapus seluruh data perjalanan dinas ini? Tindakan ini tidak dapat dibatalkan."
        confirmText="Ya, Hapus"
        onConfirm={processDeleteGroup}
    />

    <ReviewModal
        bind:open={isReviewModalOpen}
        records={selectedReviewRecords}
        provinces={$provincesStore}
        stakeholders={$stakeholdersStore}
    />
</div>
