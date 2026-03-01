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

    // Derived Records
    $: filteredRecords = $recordsStore
        .filter(r => {
            const query = searchQuery.toLowerCase();
            const matchSearch = 
                (r.employee?.name?.toLowerCase() || '').includes(query) ||
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
</script>

<div class="space-y-6 pb-20">
    <AdminHeader>
        <AdminTableFilters 
            bind:searchQuery 
            bind:statusFilter 
            bind:sortOption 
            bind:startDate
            bind:endDate
        />
    </AdminHeader>

    {#if $userStore.role === 'super_admin' || $userStore.role === 'keuangan' || $userStore.role === 'kasubag'}

        <div class="flex flex-col gap-6">
            {#each uniqueRecords as record (record.spd)}
                <div class="bg-white rounded-xl border border-slate-200 shadow-sm hover:shadow-md transition-all duration-200 overflow-hidden flex flex-col">
                    <div class="p-6 flex-1 space-y-4">
                        <div class="flex justify-between items-start">
                            <div class="flex flex-wrap gap-1.5">
                                <span class="inline-flex items-center rounded-full px-2.5 py-0.5 text-[10px] font-bold uppercase tracking-wide border {record.status === 'Draft' ? 'bg-slate-50 text-slate-600 border-slate-200' : (record.status === 'Submitted' ? 'bg-amber-50 text-amber-700 border-amber-200' : 'bg-emerald-50 text-emerald-700 border-emerald-200')}">
                                    {record.status === 'Draft' ? 'Menunggu Protokol' : record.status}
                                </span>
                                <span class="inline-flex items-center rounded-full px-2 py-0.5 text-[10px] font-bold capitalize tracking-wide bg-indigo-50 text-indigo-700 border border-indigo-200">
                                    {record.type ? record.type.replace(/_/g, ' ') : 'Dalam Kota'}
                                </span>
                            </div>
                            <span class="text-xs font-mono text-slate-400">{record.spd}</span>
                        </div>

                        <div>
                            <div class="font-semibold text-slate-900 text-sm mb-2">{record.employeesList.length} Petugas</div>
                            <h3 class="font-bold text-lg text-slate-800 line-clamp-2">{record.purpose}</h3>
                            <p class="text-sm text-slate-500 mt-1 flex items-center gap-1">
                                <svg xmlns="http://www.w3.org/2000/svg" class="h-4 w-4 text-slate-400" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M17.657 16.657L13.414 20.9a1.998 1.998 0 01-2.827 0l-4.244-4.243a8 8 0 1111.314 0z" />
                                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 11a3 3 0 11-6 0 3 3 0 016 0z" />
                                </svg>
                                {record.location}, {record.province}
                            </p>
                        </div>

                        <div class="flex items-center gap-3 text-xs text-slate-500 pt-2 border-t border-slate-50">
                            <div class="flex items-center gap-1">
                                <svg xmlns="http://www.w3.org/2000/svg" class="h-4 w-4 text-slate-400" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M8 7V3m8 4V3m-9 8h10M5 21h14a2 2 0 002-2V7a2 2 0 00-2-2H5a2 2 0 00-2 2v12a2 2 0 002 2z" />
                                </svg>
                                {new Date(record.startDate).toLocaleDateString('id-ID', { day: 'numeric', month: 'short', year: 'numeric'})}
                            </div>
                            <span>&mdash;</span>
                            <div>{new Date(record.endDate).toLocaleDateString('id-ID', { day: 'numeric', month: 'short', year: 'numeric'})}</div>
                        </div>
                    </div>

                    <div class="px-6 py-4 bg-slate-50 border-t border-slate-100 flex flex-col gap-3">
                        <div class="text-sm font-semibold text-slate-700">Daftar Pegawai & Biaya</div>
                        <div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
                            {#each record.employeesList as empRecord}
                                <div class="bg-white p-4 rounded-lg border border-slate-200 shadow-sm flex flex-col justify-between h-full">
                                    <div class="mb-4">
                                        <div class="font-medium text-sm text-slate-800 line-clamp-1" title={empRecord.employee?.name}>{empRecord.employee?.name || '-'}</div>
                                        <div class="text-xs text-slate-500 mb-2">{empRecord.employee?.rank || empRecord.employee?.golongan || '-'}</div>
                                        <span class="font-mono font-semibold text-blue-600 text-sm">{new Intl.NumberFormat('id-ID', { style: 'currency', currency: 'IDR' }).format(empRecord.totalCost || 0)}</span>
                                        <div class="mt-2">
                                            <span class="inline-flex items-center rounded-full px-2 py-0.5 text-[9px] font-bold uppercase tracking-wide border {empRecord.status === 'Draft' ? 'bg-slate-50 text-slate-600 border-slate-200' : (empRecord.status === 'Submitted' ? 'bg-amber-50 text-amber-700 border-amber-200' : 'bg-emerald-50 text-emerald-700 border-emerald-200')}">
                                                {empRecord.status === 'Draft' ? 'Menunggu Protokol' : empRecord.status}
                                            </span>
                                        </div>
                                    </div>
                                    <div class="flex items-center gap-2 pt-3 border-t border-slate-100">
                                        <button class="p-1.5 hover:bg-slate-100 rounded-md text-slate-400 hover:text-slate-700 transition-colors" title="Cetak SPD" on:click={() => generateSPD(empRecord)}>
                                            <svg xmlns="http://www.w3.org/2000/svg" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M14.5 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V7.5L14.5 2z"/><polyline points="14 2 14 8 20 8"/><line x1="16" x2="8" y1="13" y2="13"/><line x1="16" x2="8" y1="17" y2="17"/><line x1="10" x2="8" y1="9" y2="9"/></svg>
                                        </button>
                                        <button class="p-1.5 hover:bg-slate-100 rounded-md text-slate-400 hover:text-slate-700 transition-colors" title="Cetak Rincian" on:click={() => generateRincian(empRecord)}>
                                            <svg xmlns="http://www.w3.org/2000/svg" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><rect width="16" height="20" x="4" y="2" rx="2"/><line x1="8" x2="16" y1="6" y2="6"/><line x1="16" x2="16" y1="14" y2="18"/><path d="M16 10h.01"/><path d="M12 10h.01"/><path d="M8 10h.01"/><path d="M12 14h.01"/><path d="M8 14h.01"/><path d="M12 18h.01"/><path d="M8 18h.01"/></svg>
                                        </button>
                                        {#if $userStore.role !== 'kasubag'}
                                            {#if empRecord.status === 'Submitted'}
                                                <button class="ml-auto bg-white border border-slate-200 text-blue-600 hover:bg-blue-50 hover:border-blue-200 px-3 py-1.5 rounded-md text-xs font-medium transition-all shadow-sm flex-1" on:click={() => openEditModal(empRecord)}>
                                                    Review Biaya
                                                </button>
                                            {:else if empRecord.status === 'Approved'}
                                                <button class="ml-auto bg-white border border-emerald-200 text-emerald-600 hover:bg-emerald-50 hover:border-emerald-300 px-3 py-1.5 rounded-md text-xs font-medium transition-all shadow-sm flex-1" on:click={() => openEditModal(empRecord)}>
                                                    Edit Review
                                                </button>
                                            {/if}
                                        {/if}
                                    </div>
                                </div>
                            {/each}
                        </div>
                    </div>
                </div>
            {:else}
                <div class="col-span-full p-12 text-center text-slate-500 bg-slate-50/50 rounded-xl border-2 border-dashed border-slate-200">
                    Belum ada pengajuan yang masuk.
                </div>
            {/each}
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
