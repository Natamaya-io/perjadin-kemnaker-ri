<script>
    import { paginatedRecordsStore, paginatedMetadataStore, loadPaginatedRecords, deleteRecordBySpd, isFetchingRecords } from '$lib/features/pengajuan/store';
    import { userStore } from '$lib/features/auth/store';
    import { onMount, onDestroy } from 'svelte';
    import { goto } from '$app/navigation';
    import { provincesStore, stakeholdersStore } from '$lib/shared/stores/master-data';
    import { toast } from '$lib/shared/stores/toast';
    import { getStatusBadge, toTitleCase, formatLocations, formatCurrency } from '$lib/shared/utils/utils';
    
    import { createVirtualizer } from '@tanstack/svelte-virtual';

    import Table from '$lib/shared/ui/table/Table.svelte';
    import TableHeader from '$lib/shared/ui/table/TableHeader.svelte';
    import TableRow from '$lib/shared/ui/table/TableRow.svelte';
    import TableHead from '$lib/shared/ui/table/TableHead.svelte';
    import TableBody from '$lib/shared/ui/table/TableBody.svelte';
    import TableCell from '$lib/shared/ui/table/TableCell.svelte';
    import ReviewModal from '$lib/features/pengajuan/ui/ReviewModal.svelte';
    import { ConfirmationModal } from '$lib/shared/ui/confirmation-modal';
    import AdminTableFilters from '$lib/features/admin/ui/AdminTableFilters.svelte';

    // Filter & Sort State
    let searchQuery = '';
    let statusFilter = 'all';
    let sortOption = 'spj-desc';
    let startDate = '';
    let endDate = '';
    
    let statusOptions = [
        { value: 'all', label: 'Semua Status' },
        { value: 'Draft', label: 'Draft' },
        { value: 'Submitted', label: 'Menunggu Persetujuan' },
        { value: 'Approved', label: 'Disetujui' },
        { value: 'Rejected', label: 'Ditolak' }
    ];

    let debounceTimer;
    let currentCursor = '';

    // Reactively refetch when filters change (Resetting)
    $: {
        const s = searchQuery;
        const st = statusFilter;
        const so = sortOption;
        const sd = startDate;
        const ed = endDate;
        
        if (typeof window !== 'undefined') {
            clearTimeout(debounceTimer);
            debounceTimer = setTimeout(() => {
                currentCursor = '';
                fetchRecords(false);
            }, 300);
        }
    }

    onMount(() => {
        // Automatically fetch fresh records every time the page is visited
        if (typeof window !== 'undefined') {
            currentCursor = '';
            fetchRecords(false);
        }
    });

    onDestroy(() => {
        if (typeof window !== 'undefined') {
            clearTimeout(debounceTimer);
        }
    });

    function fetchRecords(append = false) {
        if (!$userStore) return;

        let statusParam = statusFilter === 'all' ? undefined : statusFilter;
        let sortByParam = sortOption;

        let startIso, endIso;        if (startDate) {
            const start = new Date(startDate);
            start.setHours(0,0,0,0);
            startIso = start.toISOString();
        }
        if (endDate) {
            const end = new Date(endDate);
            end.setHours(23,59,59,999);
            endIso = end.toISOString();
        }

        const targetUserId = ($userStore.role === 'super_admin' || $userStore.role === 'kasubag') ? undefined : $userStore.id;

        const params = {
            limit: 20,
            search: searchQuery,
            status: statusParam,
            sort_by: sortByParam,
            start_date: startIso,
            end_date: endIso,
            user_id: targetUserId
        };

        if (append && currentCursor) {
            params.cursor = currentCursor;
        }

        loadPaginatedRecords(params, append);
    }

    // Group records by SPD for display
    $: groupedRecordsMap = $paginatedRecordsStore.reduce((/** @type {Record<string, any>} */ acc, record) => {
        if (!acc[record.spd]) {
            acc[record.spd] = { ...record, employeesList: [record], totalCost: record.totalCost || 0 };
        } else {
            acc[record.spd].employeesList.push(record);
            acc[record.spd].totalCost += record.totalCost || 0;
        }
        return acc;
    }, {});
    
    // Maintain SPD order
    $: uniqueSPDs = [...new Set($paginatedRecordsStore.map(r => r.spd))];
    $: uniqueRecords = uniqueSPDs.map(spd => groupedRecordsMap[spd]).filter(Boolean);

    let scrollContainer;
    let virtualizer;
    $: virtualizer = createVirtualizer({
        count: uniqueRecords.length,
        getScrollElement: () => scrollContainer,
        estimateSize: () => 100, // Estimated row height
        overscan: 5, // Render 5 items outside viewport for smooth scrolling
    });

    // Infinite Scroll logic: fetch more when reaching bottom
    function handleScroll() {
        if (!scrollContainer || $isFetchingRecords) return;
        const { scrollTop, scrollHeight, clientHeight } = scrollContainer;
        if (scrollHeight - scrollTop - clientHeight < 200) {
            if ($paginatedMetadataStore.nextCursor) {
                currentCursor = $paginatedMetadataStore.nextCursor;
                fetchRecords(true);
            }
        }
    }

    let isReviewOpen = false;
    /** @type {any[]} */
    let selectedRecords = [];
    
    let isDeleteModalOpen = false;
    let spdToDelete = '';

    /** @param {string} spd */
    function handleReview(spd) {
        selectedRecords = $paginatedRecordsStore.filter(r => r.spd === spd);
        isReviewOpen = true;
    }

    /** @param {string} spd */
    function handleDelete(spd) {
        spdToDelete = spd;
        isDeleteModalOpen = true;
    }
    
    async function processDelete() {
        if (spdToDelete) {
            try {
                await deleteRecordBySpd(spdToDelete);
                toast.success(`Pengajuan ${spdToDelete} berhasil dihapus.`);
            } catch (e) {
                toast.error(`Gagal menghapus pengajuan ${spdToDelete}.`);
            } finally {
                spdToDelete = '';
                isDeleteModalOpen = false;
            }
        }
    }

    onMount(() => {
        if ($userStore.role !== 'super_admin' && $userStore.role !== 'kasubag') {
            goto('/dashboard');
        } else {
            fetchRecords(false);
        }
    });
</script>

<div class="space-y-6 pb-20 max-w-7xl mx-auto">
    <div class="flex flex-col md:flex-row md:items-center justify-between gap-4 bg-white p-6 rounded-2xl shadow-sm border border-slate-100">
        <div>
            <h1 class="text-2xl font-bold text-slate-800 tracking-tight">Daftar Pengajuan Saya</h1>
            <p class="text-sm text-slate-500 mt-1">Pantau status dan riwayat perjalanan dinas yang telah Anda ajukan.</p>
        </div>
        <div class="flex items-center gap-3">
            <a href="/dashboard/pengajuan/import" class="inline-flex items-center justify-center bg-white border border-slate-200 hover:bg-slate-50 text-slate-700 font-medium px-4 py-2.5 rounded-xl shadow-sm transition-all duration-200">
                <svg xmlns="http://www.w3.org/2000/svg" class="h-5 w-5 mr-2 text-green-600" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
                    <path stroke-linecap="round" stroke-linejoin="round" d="M12 10v6m0 0l-3-3m3 3l3-3m2 8H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z" />
                </svg>
                Import Excel
            </a>
            <a href="/dashboard/pengajuan/new" class="inline-flex items-center justify-center bg-blue-600 hover:bg-blue-700 text-white font-medium px-5 py-2.5 rounded-xl shadow-sm shadow-blue-500/30 transition-all duration-200">
                <svg xmlns="http://www.w3.org/2000/svg" class="h-5 w-5 mr-2" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
                    <path stroke-linecap="round" stroke-linejoin="round" d="M12 4v16m8-8H4" />
                </svg>
                Buat Pengajuan Baru
            </a>
        </div>
    </div>

    <div class="bg-white p-4 rounded-xl shadow-sm border border-slate-100">
        <AdminTableFilters 
            bind:searchQuery 
            bind:statusFilter 
            bind:sortOption 
            bind:startDate
            bind:endDate
            statusOptions={statusOptions}
        />
    </div>

    <div class="rounded-xl border border-slate-200 shadow-sm bg-white overflow-hidden relative">
        <!-- Filler for scrollbar header gap -->
        <div class="absolute top-0 right-0 w-[12px] h-[45px] bg-slate-50 border-b border-slate-200 z-30"></div>
        <!-- VIRTUAL SCROLL CONTAINER -->
        <div 
            bind:this={scrollContainer} 
            on:scroll={handleScroll}
            class="overflow-auto max-h-[70vh] w-full relative scroll-smooth table-scrollbar table-scroll-shadows"
        >
            <table class="w-full text-sm text-left relative">
                <thead class="bg-slate-50 sticky top-0 z-10 shadow-sm">
                    <tr>
                        <th class="min-w-[120px] font-semibold text-slate-700 pl-4 py-3 bg-slate-50 border-b border-slate-200">ID SPJ</th>
                        <th class="min-w-[250px] font-semibold text-slate-700 py-3 bg-slate-50 border-b border-slate-200">Tujuan & Lokasi</th>
                        <th class="min-w-[130px] font-semibold text-slate-700 py-3 bg-slate-50 border-b border-slate-200">Tanggal</th>
                        <th class="min-w-[160px] font-semibold text-slate-700 text-center py-3 bg-slate-50 border-b border-slate-200">Status</th>
                        <th class="min-w-[180px] font-semibold text-slate-700 text-center py-3 bg-slate-50 border-b border-slate-200">Total SBM</th>
                        <th class="w-[140px] min-w-[140px] font-semibold text-slate-700 text-center pr-4 py-3 bg-slate-50 border-b border-slate-200">Aksi</th>
                    </tr>
                </thead>
                <tbody>
                    {#if uniqueRecords.length === 0}
                        <tr>
                            <td colspan="6" class="p-12 text-center text-slate-500">
                                {#if $isFetchingRecords}
                                    <div class="flex flex-col items-center justify-center py-6 gap-3">
                                        <div class="relative flex items-center justify-center overflow-hidden w-24 h-24">
                                            <svg xmlns="http://www.w3.org/2000/svg" class="w-10 h-10 text-blue-500 animate-paper-flight drop-shadow-md" fill="currentColor" viewBox="0 0 24 24">
                                                <path d="M2.01 21L23 12 2.01 3 2 10l15 2-15 2z"/>
                                            </svg>
                                        </div>
                                        <span class="text-sm font-semibold text-slate-500 tracking-wide animate-pulse">Sedang memuat data...</span>
                                    </div>
                                {:else}
                                    <svg xmlns="http://www.w3.org/2000/svg" class="h-12 w-12 mx-auto text-slate-300 mb-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                                        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z" />
                                    </svg>
                                    Belum ada pengajuan.
                                {/if}
                            </td>
                        </tr>
                    {:else}
                        <!-- Virtual Window Top Spacer -->
                        {#if $virtualizer.getVirtualItems().length > 0}
                            <tr style="height: {$virtualizer.getVirtualItems()[0].start}px"></tr>
                            
                            <!-- Virtual Rendered Rows -->
                            {#each $virtualizer.getVirtualItems() as virtualRow (virtualRow.index)}
                                {@const record = uniqueRecords[virtualRow.index]}
                                <tr class="hover:bg-slate-50/50 border-b border-slate-100 transition-colors bg-white">
                                    <td class="font-mono text-xs text-slate-500 pl-4 py-4 align-middle">
                                        {record.spd}
                                        <div class="mt-1">
                                            <span class="inline-flex items-center rounded-full px-2 py-0.5 text-[9px] font-bold capitalize tracking-wide bg-indigo-50 text-indigo-700 border border-indigo-200">
                                                {record.type ? record.type.replace(/_/g, ' ') : 'Dalam Kota'}
                                            </span>
                                        </div>
                                        <div class="mt-2 text-[10px] text-slate-400">
                                            {record.employeesList.length} Petugas
                                        </div>
                                    </td>
                                    <td class="py-4 align-middle">
                                        <div class="font-medium text-slate-800 text-sm line-clamp-2">
                                            {record.stakeholder ? `${record.purpose} ${record.stakeholder}` : record.purpose}
                                        </div>
                                        <div class="text-xs text-slate-500 mt-1 flex items-center gap-1">
                                            <svg xmlns="http://www.w3.org/2000/svg" class="h-3.5 w-3.5 text-slate-400" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                                                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M17.657 16.657L13.414 20.9a1.998 1.998 0 01-2.827 0l-4.244-4.243a8 8 0 1111.314 0z" />
                                                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 11a3 3 0 11-6 0 3 3 0 016 0z" />
                                            </svg>
                                            {formatLocations(record)}
                                        </div>
                                    </td>
                                    <td class="py-4 align-middle text-xs text-slate-600">
                                        <div class="whitespace-nowrap">
                                            {new Date(record.startDate).toLocaleDateString('id-ID', { day: 'numeric', month: 'short', year: 'numeric'})}
                                        </div>
                                        <div class="text-slate-400 my-0.5 text-[10px]">s/d</div>
                                        <div class="whitespace-nowrap">
                                            {new Date(record.endDate).toLocaleDateString('id-ID', { day: 'numeric', month: 'short', year: 'numeric'})}
                                        </div>
                                    </td>
                                    <td class="py-4 align-middle text-center">
                                        <div class="flex flex-col gap-1.5 items-center justify-center">
                                            <span class="inline-flex items-center rounded-full px-2.5 py-0.5 text-[10px] font-bold uppercase tracking-wide border {getStatusBadge(record).class}">
                                                {getStatusBadge(record).label}
                                            </span>
                                        </div>
                                    </td>
                                    <td class="text-center py-4 align-middle">
                                        <span class="font-mono font-semibold text-slate-700 text-sm whitespace-nowrap">
                                            {formatCurrency(record.totalCost || 0)}
                                        </span>
                                    </td>
                                    <td class="text-center pr-4 py-4 align-middle">
                                        <div class="flex items-center justify-center gap-2">
                                            <button 
                                                class="text-blue-600 hover:text-blue-800 bg-blue-50 hover:bg-blue-100 p-1.5 rounded transition-colors" 
                                                title="Review"
                                                on:click={() => handleReview(record.spd)}
                                            >
                                                <svg xmlns="http://www.w3.org/2000/svg" class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                                                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 12a3 3 0 11-6 0 3 3 0 016 0z" />
                                                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M2.458 12C3.732 7.943 7.523 5 12 5c4.478 0 8.268 2.943 9.542 7-1.274 4.057-5.064 7-9.542 7-4.477 0-8.268-2.943-9.542-7z" />
                                                </svg>
                                            </button>
                                            {#if $userStore?.role === 'super_admin' || $userStore?.role === 'kasubag' || (record.employeesList[0].creator?.email || record.employeesList[0].email) === $userStore?.email || (record.employeesList[0].creatorId || record.employeesList[0].creator?.id) === $userStore?.id}
                                                <button 
                                                    class="text-red-600 hover:text-red-800 bg-red-50 hover:bg-red-100 p-1.5 rounded transition-colors" 
                                                    title="Hapus"
                                                    on:click={() => handleDelete(record.spd)}
                                                >
                                                    <svg xmlns="http://www.w3.org/2000/svg" class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                                                        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16" />
                                                    </svg>
                                                </button>
                                            {/if}
                                        </div>
                                    </td>
                                </tr>
                            {/each}
                            
                            <!-- Virtual Window Bottom Spacer -->
                            <tr style="height: {$virtualizer.getTotalSize() - $virtualizer.getVirtualItems()[$virtualizer.getVirtualItems().length - 1].end}px"></tr>
                            
                            <!-- Loading Indicator when scrolling near bottom -->
                            {#if $isFetchingRecords && currentCursor}
                                <tr>
                                    <td colspan="6" class="p-4 text-center">
                                        <div class="flex items-center justify-center gap-3">
                                            <div class="relative flex items-center justify-center overflow-hidden w-8 h-8">
                                                <svg xmlns="http://www.w3.org/2000/svg" class="w-4 h-4 text-blue-500 animate-paper-flight drop-shadow-sm" fill="currentColor" viewBox="0 0 24 24">
                                                    <path d="M2.01 21L23 12 2.01 3 2 10l15 2-15 2z"/>
                                                </svg>
                                            </div>
                                            <span class="text-sm font-semibold text-slate-500 animate-pulse tracking-wide">Memuat data selanjutnya...</span>
                                        </div>
                                    </td>
                                </tr>
                            {/if}
                        {/if}
                    {/if}
                </tbody>
            </table>
        </div>
    </div>
</div>

<ReviewModal 
    bind:open={isReviewOpen} 
    records={selectedRecords} 
    provinces={$provincesStore} 
    stakeholders={$stakeholdersStore} 
/>

<ConfirmationModal 
    bind:open={isDeleteModalOpen}
    title="Hapus Pengajuan"
    description="Apakah Anda yakin ingin membatalkan dan menghapus pengajuan SPJ {spdToDelete}? Tindakan ini tidak dapat dibatalkan."
    confirmText="Ya, Hapus"
    cancelText="Batal"
    onConfirm={processDelete}
    variant="destructive"
/>
