<script>
    import { recordsStore, updateRecord } from '$lib/stores/records';
    import { userStore } from '$lib/stores/auth';
    import { toast } from '$lib/stores/toast';
    
    // Components
    import AdminHeader from '$lib/components/dashboard/admin/AdminHeader.svelte';
    import AdminTableFilters from '$lib/components/dashboard/admin/AdminTableFilters.svelte';

    import AdminTable from '$lib/components/dashboard/admin/AdminTable.svelte';
    import AdminTableRow from '$lib/components/dashboard/admin/AdminTableRow.svelte';
    import CostModal from '$lib/components/dashboard/admin/CostModal.svelte';
    
    // UI Helpers
    import { ConfirmationModal } from '$lib/components/ui/confirmation-modal';

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

    // Derived Records - Filter based on role
    $: filteredRecords = $recordsStore
        .filter(r => {
            if ($userStore.role === 'super_admin' || $userStore.role === 'kasubag') return true;
            // Show if the user is the employee OR if the user created the record
            const isEmployee = r.employee?.email === $userStore.email;
            const isCreator = (r.creator?.email || r.email) === $userStore.email || r.creatorId === $userStore.id;
            return isEmployee || isCreator;
        })
        .filter(r => {
            const query = searchQuery.toLowerCase();
            const matchSearch = 
                (r.spd?.toLowerCase() || '').includes(query) ||
                (r.location?.toLowerCase() || '').includes(query);
            
            const matchStatus = statusFilter === 'all' || r.status === statusFilter;
            
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
                <thead class="bg-slate-50 border-b border-slate-200 text-slate-500 uppercase tracking-wider text-xs">
                    <tr>
                        <th class="px-6 py-4 font-semibold w-1/3">Info Perjalanan</th>
                        <th class="px-6 py-4 font-semibold w-1/4">Pegawai</th>
                        <th class="px-6 py-4 font-semibold text-right">Total Biaya</th>
                        <th class="px-6 py-4 font-semibold text-center">Status</th>
                        <th class="px-6 py-4 font-semibold text-right">Aksi</th>
                    </tr>
                </thead>
                <tbody class="divide-y divide-slate-100">
                    {#each uniqueRecords as record (record.spd)}
                        <!-- Group Header Row -->
                        <tr class="bg-slate-50/80 border-b border-slate-200 cursor-pointer hover:bg-slate-100 transition-colors select-none" on:click={() => toggleGroup(record.spd)}>
                            <td colspan="5" class="px-6 py-3">
                                <div class="flex justify-between items-center gap-4">
                                    <div class="flex items-center gap-3">
                                        <button class="p-1 rounded-md hover:bg-slate-200 transition-colors focus:outline-none" aria-label="Toggle details">
                                            <svg xmlns="http://www.w3.org/2000/svg" class="h-5 w-5 text-slate-500 transition-transform duration-200 {expandedGroups[record.spd] ? 'rotate-180' : ''}" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                                                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 9l-7 7-7-7" />
                                            </svg>
                                        </button>
                                        <span class="font-mono text-xs text-slate-600 font-semibold">{record.spd}</span>
                                        <span class="inline-flex items-center rounded-full px-2 py-0.5 text-[10px] font-bold capitalize tracking-wide bg-indigo-50 text-indigo-700 border border-indigo-200">
                                            {record.type ? record.type.replace(/_/g, ' ') : 'Dalam Kota'}
                                        </span>
                                        <span class="text-slate-800 font-medium line-clamp-1">{record.purpose}</span>
                                        <span class="text-xs font-semibold text-slate-400 ml-2">({record.employeesList.length} Pegawai)</span>
                                    </div>
                                    <div class="text-xs text-slate-500 flex items-center gap-1 shrink-0">
                                        <svg xmlns="http://www.w3.org/2000/svg" class="h-3.5 w-3.5 text-slate-400" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                                            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M8 7V3m8 4V3m-9 8h10M5 21h14a2 2 0 002-2V7a2 2 0 00-2-2H5a2 2 0 00-2 2v12a2 2 0 002 2z" />
                                        </svg>
                                        <span>{new Date(record.startDate).toLocaleDateString('id-ID', { day: 'numeric', month: 'short', year: 'numeric'})} &mdash; {new Date(record.endDate).toLocaleDateString('id-ID', { day: 'numeric', month: 'short', year: 'numeric'})}</span>
                                    </div>
                                </div>
                            </td>
                        </tr>
                        <!-- Employee Rows -->
                        {#if expandedGroups[record.spd]}
                            {#each record.employeesList as empRecord}
                                <tr class="hover:bg-slate-50 transition-colors bg-white">
                                    <td class="px-6 py-4 align-middle">
                                        <div class="flex items-center gap-2">
                                            <div class="w-4 h-4 border-l-2 border-b-2 border-slate-200 rounded-bl-lg mb-4 ml-2 mr-3 opacity-60"></div>
                                            <div class="text-xs text-slate-500 truncate max-w-[200px]" title="{record.location}, {record.province}">
                                                <svg xmlns="http://www.w3.org/2000/svg" class="h-3.5 w-3.5 inline-block text-slate-400 mr-1" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                                                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M17.657 16.657L13.414 20.9a1.998 1.998 0 01-2.827 0l-4.244-4.243a8 8 0 1111.314 0z" />
                                                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 11a3 3 0 11-6 0 3 3 0 016 0z" />
                                                </svg>
                                                {record.location}, {record.province}
                                            </div>
                                        </div>
                                    </td>
                                    <td class="px-6 py-4 align-middle">
                                        <div class="font-medium text-slate-900">{empRecord.employee?.name || '-'}</div>
                                        <div class="text-xs text-slate-500">{empRecord.employee?.rank || empRecord.employee?.golongan || '-'}</div>
                                    </td>
                                    <td class="px-6 py-4 align-middle text-right font-mono font-medium text-blue-600">
                                        {new Intl.NumberFormat('id-ID', { style: 'currency', currency: 'IDR' }).format(empRecord.totalCost || 0)}
                                    </td>
                                    <td class="px-6 py-4 align-middle text-center">
                                        <span class="inline-flex items-center rounded-full px-2.5 py-0.5 text-[10px] font-bold uppercase tracking-wide border {empRecord.status === 'Draft' ? 'bg-slate-50 text-slate-600 border-slate-200' : (empRecord.status === 'Submitted' ? 'bg-amber-50 text-amber-700 border-amber-200' : 'bg-emerald-50 text-emerald-700 border-emerald-200')}">
                                            {empRecord.status === 'Draft' ? 'Menunggu Protokol' : empRecord.status}
                                        </span>
                                    </td>
                                    <td class="px-6 py-4 align-middle text-right">
                                        <button class="bg-white border border-slate-200 text-blue-600 hover:bg-blue-50 hover:border-blue-200 px-3 py-1.5 rounded-md text-xs font-medium transition-all shadow-sm whitespace-nowrap" on:click={(e) => { e.stopPropagation(); openEditModal(empRecord); }}>
                                            {$userStore.role === 'protokol' ? 'Input Dokumen & Biaya' : 'Review Biaya'}
                                        </button>
                                    </td>
                                </tr>
                            {/each}
                        {/if}
                    {/each}
                    {#if uniqueRecords.length === 0}
                        <tr>
                            <td colspan="5" class="p-12 text-center text-slate-500 bg-slate-50/50">
                                Belum ada data perjalanan dinas untuk Anda.
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
                            <div class="flex items-center justify-between">
                                <div class="flex items-center gap-1.5">
                                    <svg xmlns="http://www.w3.org/2000/svg" class="h-3.5 w-3.5 text-slate-400 shrink-0" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                                        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M8 7V3m8 4V3m-9 8h10M5 21h14a2 2 0 002-2V7a2 2 0 00-2-2H5a2 2 0 00-2 2v12a2 2 0 002 2z" />
                                    </svg>
                                    <span>{new Date(record.startDate).toLocaleDateString('id-ID', { day: 'numeric', month: 'short', year: 'numeric'})} &mdash; {new Date(record.endDate).toLocaleDateString('id-ID', { day: 'numeric', month: 'short', year: 'numeric'})}</span>
                                </div>
                                <span class="font-semibold text-[10px] text-slate-500 uppercase tracking-wider">{record.employeesList.length} Pegawai</span>
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
                                            <div class="text-xs text-slate-500 mt-0.5">{empRecord.employee?.rank || empRecord.employee?.golongan || '-'}</div>
                                        </div>
                                        <span class="shrink-0 inline-flex items-center rounded-full px-2 py-0.5 text-[9px] font-bold uppercase tracking-wide border {empRecord.status === 'Draft' ? 'bg-slate-50 text-slate-600 border-slate-200' : (empRecord.status === 'Submitted' ? 'bg-amber-50 text-amber-700 border-amber-200' : 'bg-emerald-50 text-emerald-700 border-emerald-200')}">
                                            {empRecord.status === 'Draft' ? 'Menunggu Protokol' : empRecord.status}
                                        </span>
                                    </div>

                                    <div class="flex justify-between items-end border-t border-slate-50 pt-3">
                                        <div class="flex flex-col">
                                            <span class="text-[10px] font-semibold text-slate-400 uppercase tracking-wider mb-0.5">Total Biaya</span>
                                            <span class="font-mono font-bold text-blue-600 text-sm">
                                                {new Intl.NumberFormat('id-ID', { style: 'currency', currency: 'IDR' }).format(empRecord.totalCost || 0)}
                                            </span>
                                        </div>
                                        <button class="bg-white border border-slate-200 text-blue-600 hover:bg-blue-50 hover:border-blue-200 px-3 py-1.5 rounded-md text-xs font-medium transition-all shadow-sm" on:click={(e) => { e.stopPropagation(); openEditModal(empRecord); }}>
                                            {$userStore.role === 'protokol' ? 'Input Dokumen & Biaya' : 'Review Biaya'}
                                        </button>
                                    </div>
                                </div>
                            {/each}
                        </div>
                    {/if}
                </div>
            {/each}
            {#if uniqueRecords.length === 0}
                <div class="p-12 text-center text-slate-500 bg-white">
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
</div>
