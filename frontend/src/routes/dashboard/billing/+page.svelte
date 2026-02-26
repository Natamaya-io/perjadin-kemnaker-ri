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

    // Derived Records - Filter based on role
    $: filteredRecords = $recordsStore
        .filter(r => ($userStore.role === 'super_admin' || $userStore.role === 'kasubag') ? true : r.employee?.email === $userStore.email) // ONLY show own records for protokol
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
            if (sortOption === 'date-desc') return new Date(b.startDate) - new Date(a.startDate);
            if (sortOption === 'date-asc') return new Date(a.startDate) - new Date(b.startDate);
            if (sortOption === 'cost-desc') return b.totalCost - a.totalCost;
            if (sortOption === 'cost-asc') return a.totalCost - b.totalCost;
            return 0;
        });



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
        window.open(`/print?type=spd&spd=${encodeURIComponent(record.spd)}`, '_blank');
    }

    function generateRincian(record) {
        window.open(`/print?type=rincian&spd=${encodeURIComponent(record.spd)}`, '_blank');
    }
</script>

<div class="space-y-6 pb-20">
    <div class="flex flex-col sm:flex-row justify-between items-start sm:items-center gap-4 mb-6">
        <div>
            <h1 class="text-2xl font-bold text-slate-900 tracking-tight">Input Perjalanan</h1>
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

    <div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
        {#each filteredRecords as record (record.id)}
            <div class="bg-white rounded-xl border border-slate-200 shadow-sm hover:shadow-md transition-all duration-200 overflow-hidden flex flex-col">
                <div class="p-6 flex-1 space-y-4">
                    <div class="flex justify-between items-start">
                        <div class="flex flex-wrap gap-1.5">
                            <span class="inline-flex items-center rounded-full px-2.5 py-0.5 text-[10px] font-bold uppercase tracking-wide border {record.status === 'Draft' ? 'bg-slate-50 text-slate-600 border-slate-200' : (record.status === 'Submitted' ? 'bg-amber-50 text-amber-700 border-amber-200' : 'bg-emerald-50 text-emerald-700 border-emerald-200')}">
                                {record.status}
                            </span>
                            <span class="inline-flex items-center rounded-full px-2 py-0.5 text-[10px] font-bold capitalize tracking-wide bg-indigo-50 text-indigo-700 border border-indigo-200">
                                {record.type ? record.type.split('_').join(' ') : 'Dalam Kota'}
                            </span>
                        </div>
                        <span class="text-xs font-mono text-slate-400">{record.spd}</span>
                    </div>

                    <div>
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

                <div class="px-6 py-4 bg-slate-50 border-t border-slate-100 flex items-center justify-between">
                    <span class="font-mono font-semibold text-slate-700 text-sm">{new Intl.NumberFormat('id-ID', { style: 'currency', currency: 'IDR' }).format(record.totalCost || 0)}</span>
                    <button class="bg-white border border-slate-200 text-blue-600 hover:bg-blue-50 hover:border-blue-200 px-3 py-1.5 rounded-md text-xs font-medium transition-all shadow-sm" on:click={() => openEditModal(record)}>
                        {$userStore.role === 'protokol' ? 'Input Dokumen & Biaya' : 'Review Biaya'}
                    </button>
                </div>
            </div>
        {:else}
            <div class="col-span-full p-12 text-center text-slate-500 bg-slate-50/50 rounded-xl border-2 border-dashed border-slate-200">
                Belum ada data perjalanan dinas untuk Anda.
            </div>
        {/each}
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
