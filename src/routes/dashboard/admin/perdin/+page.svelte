<script>
    import { recordsStore } from '$lib/stores/records';
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

            return matchSearch && matchStatus;
        })
        .sort((a, b) => {
            if (sortOption === 'date-desc') return new Date(b.startDate) - new Date(a.startDate);
            if (sortOption === 'date-asc') return new Date(a.startDate) - new Date(b.startDate);
            if (sortOption === 'cost-desc') return b.totalCost - a.totalCost;
            if (sortOption === 'cost-asc') return a.totalCost - b.totalCost;
            return 0;
        });



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

    function processSave() {
        recordsStore.update(currentRecords => {
            return currentRecords.map(r => {
                if (r.spd === selectedRecord.spd) {
                    return {
                        ...r,
                        costs: { ...editingCosts },
                        totalCost: pendingGrandTotal,
                        status: 'Approved' 
                    };
                }
                return r;
            });
        });
        isModalOpen = false;
        toast.success('Rincian biaya berhasil disimpan!');
    }

    function generateSPD(record) {
        window.open(`/print?type=spd&spd=${encodeURIComponent(record.spd)}`, '_blank');
    }

    function generateRincian(record) {
        window.open(`/print?type=rincian&spd=${encodeURIComponent(record.spd)}`, '_blank');
    }
</script>

<div class="space-y-6 pb-20">
    <AdminHeader>
        <AdminTableFilters 
            bind:searchQuery 
            bind:statusFilter 
            bind:sortOption 
        />
    </AdminHeader>

    {#if $userStore.role === 'super_admin' || $userStore.role === 'keuangan'}


        <AdminTable>
            {#each filteredRecords as record (record.id)}
                <AdminTableRow 
                    {record} 
                    on:edit={() => openEditModal(record)}
                    on:printSPD={() => generateSPD(record)}
                    on:printRincian={() => generateRincian(record)}
                />
            {/each}
        </AdminTable>
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
