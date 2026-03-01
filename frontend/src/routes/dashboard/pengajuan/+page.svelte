<script>
    import { recordsStore, deleteRecordBySpd } from '$lib/stores/records';
    import { userStore } from '$lib/stores/auth';
    import { provincesStore, stakeholdersStore } from '$lib/stores/master-data';
    import { toast } from '$lib/stores/toast';
    
    import Table from '$lib/components/ui/table/Table.svelte';
    import TableHeader from '$lib/components/ui/table/TableHeader.svelte';
    import TableRow from '$lib/components/ui/table/TableRow.svelte';
    import TableHead from '$lib/components/ui/table/TableHead.svelte';
    import TableBody from '$lib/components/ui/table/TableBody.svelte';
    import TableCell from '$lib/components/ui/table/TableCell.svelte';
    import ReviewModal from '$lib/components/dashboard/pengajuan/ReviewModal.svelte';
    import { ConfirmationModal } from '$lib/components/ui/confirmation-modal';
    
    // Derived Records: Filter only records created by the current user
    // Group by SPD number since one submission can have multiple employees
    $: myRecords = $recordsStore
        .filter(r => {
            const creatorEmail = r.creator?.email || r.email; // fallback to email if it was stored that way
            const creatorId = r.creatorId || r.creator?.id;
            return creatorEmail === $userStore?.email || creatorId === $userStore?.id;
        })
        .sort((a, b) => new Date(b.startDate).getTime() - new Date(a.startDate).getTime());
        
    // Group records by SPD for display
    $: groupedRecords = myRecords.reduce((/** @type {Record<string, any>} */ acc, record) => {
        if (!acc[record.spd]) {
            acc[record.spd] = { ...record, employeesList: [record.employee], totalCost: record.totalCost || 0 };
        } else {
            acc[record.spd].employeesList.push(record.employee);
            acc[record.spd].totalCost += record.totalCost || 0;
        }
        return acc;
    }, {});
    
    $: uniqueRecords = Object.values(groupedRecords).sort((a, b) => new Date(b.startDate).getTime() - new Date(a.startDate).getTime());

    let isReviewOpen = false;
    /** @type {any[]} */
    let selectedRecords = [];
    
    let isDeleteModalOpen = false;
    let spdToDelete = '';

    /** @param {string} spd */
    function handleReview(spd) {
        selectedRecords = myRecords.filter(r => r.spd === spd);
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
</script>

<div class="space-y-6 pb-20 max-w-7xl mx-auto">
    <div class="flex flex-col md:flex-row md:items-center justify-between gap-4 bg-white p-6 rounded-2xl shadow-sm border border-slate-100">
        <div>
            <h1 class="text-2xl font-bold text-slate-800 tracking-tight">Daftar Pengajuan Saya</h1>
            <p class="text-sm text-slate-500 mt-1">Pantau status dan riwayat perjalanan dinas yang telah Anda ajukan.</p>
        </div>
        <a href="/dashboard/pengajuan/new" class="inline-flex items-center justify-center bg-blue-600 hover:bg-blue-700 text-white font-medium px-5 py-2.5 rounded-xl shadow-sm shadow-blue-500/30 transition-all duration-200">
            <svg xmlns="http://www.w3.org/2000/svg" class="h-5 w-5 mr-2" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
                <path stroke-linecap="round" stroke-linejoin="round" d="M12 4v16m8-8H4" />
            </svg>
            Buat Pengajuan Baru
        </a>
    </div>

    <div class="rounded-xl border border-slate-200 shadow-sm bg-white overflow-hidden">
        <div class="overflow-x-auto w-full">
            <Table class="w-full text-sm text-left">
                <TableHeader class="bg-slate-50 border-b border-slate-200">
                    <TableRow class="hover:bg-slate-50/50">
                        <TableHead class="min-w-[120px] font-semibold text-slate-700 pl-4 py-3">No. SPD</TableHead>
                        <TableHead class="min-w-[250px] font-semibold text-slate-700 py-3">Tujuan & Lokasi</TableHead>
                        <TableHead class="min-w-[160px] font-semibold text-slate-700 py-3">Tanggal</TableHead>
                        <TableHead class="w-[120px] min-w-[120px] font-semibold text-slate-700 py-3">Status</TableHead>
                        <TableHead class="min-w-[140px] font-semibold text-slate-700 text-right py-3">Total SBM</TableHead>
                        <TableHead class="w-[140px] min-w-[140px] font-semibold text-slate-700 text-center pr-4 py-3">Aksi</TableHead>
                    </TableRow>
                </TableHeader>
                <TableBody>
                    {#if uniqueRecords.length === 0}
                        <TableRow>
                            <TableCell colspan="6" class="p-12 text-center text-slate-500">
                                <svg xmlns="http://www.w3.org/2000/svg" class="h-12 w-12 mx-auto text-slate-300 mb-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z" />
                                </svg>
                                Belum ada pengajuan yang Anda buat.
                            </TableCell>
                        </TableRow>
                    {:else}
                        {#each uniqueRecords as record (record.id)}
                            <TableRow class="hover:bg-slate-50/50 border-b border-slate-100 last:border-0 transition-colors">
                                <TableCell class="font-mono text-xs text-slate-500 pl-4 py-4 align-top">
                                    {record.spd}
                                    <div class="mt-1">
                                        <span class="inline-flex items-center rounded-full px-2 py-0.5 text-[9px] font-bold capitalize tracking-wide bg-indigo-50 text-indigo-700 border border-indigo-200">
                                            {record.type ? record.type.replace(/_/g, ' ') : 'Dalam Kota'}
                                        </span>
                                    </div>
                                    <div class="mt-2 text-[10px] text-slate-400">
                                        {record.employeesList.length} Petugas
                                    </div>
                                </TableCell>
                                <TableCell class="py-4 align-top">
                                    <div class="font-medium text-slate-800 text-sm line-clamp-2">{record.purpose}</div>
                                    <div class="text-xs text-slate-500 mt-1 flex items-center gap-1">
                                        <svg xmlns="http://www.w3.org/2000/svg" class="h-3.5 w-3.5 text-slate-400" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                                            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M17.657 16.657L13.414 20.9a1.998 1.998 0 01-2.827 0l-4.244-4.243a8 8 0 1111.314 0z" />
                                            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 11a3 3 0 11-6 0 3 3 0 016 0z" />
                                        </svg>
                                        {record.location}, {record.province}
                                    </div>
                                </TableCell>
                                <TableCell class="py-4 align-top text-xs text-slate-600">
                                    <div class="whitespace-nowrap">
                                        {new Date(record.startDate).toLocaleDateString('id-ID', { day: 'numeric', month: 'short', year: 'numeric'})}
                                    </div>
                                    <div class="text-slate-400 my-0.5 text-[10px]">s/d</div>
                                    <div class="whitespace-nowrap">
                                        {new Date(record.endDate).toLocaleDateString('id-ID', { day: 'numeric', month: 'short', year: 'numeric'})}
                                    </div>
                                </TableCell>
                                <TableCell class="py-4 align-top">
                                    <div class="flex flex-col gap-1.5 items-start">
                                        <span class="inline-flex items-center rounded-full px-2.5 py-0.5 text-[10px] font-bold uppercase tracking-wide border {record.status === 'Draft' ? 'bg-slate-50 text-slate-600 border-slate-200' : (record.status === 'Submitted' ? 'bg-amber-50 text-amber-700 border-amber-200' : 'bg-emerald-50 text-emerald-700 border-emerald-200')}">
                                            {record.status === 'Draft' ? 'Menunggu Protokol' : record.status}
                                        </span>
                                        {#if record.status === 'Draft'}
                                            <span class="text-[10px] text-slate-400 italic">Belum Diproses</span>
                                        {:else}
                                            <span class="text-[10px] text-emerald-600 font-medium bg-emerald-50 px-1.5 py-0.5 rounded border border-emerald-100">Diproses</span>
                                        {/if}
                                    </div>
                                </TableCell>
                                <TableCell class="text-right py-4 align-top">
                                    <span class="font-mono font-semibold text-slate-700 text-sm whitespace-nowrap">
                                        {new Intl.NumberFormat('id-ID', { style: 'currency', currency: 'IDR', minimumFractionDigits: 0 }).format(record.totalCost || 0)}
                                    </span>
                                </TableCell>
                                <TableCell class="text-center pr-4 py-4 align-top">
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
                                        <button 
                                            class="text-red-600 hover:text-red-800 bg-red-50 hover:bg-red-100 p-1.5 rounded transition-colors" 
                                            title="Hapus"
                                            on:click={() => handleDelete(record.spd)}
                                        >
                                            <svg xmlns="http://www.w3.org/2000/svg" class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                                                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16" />
                                            </svg>
                                        </button>
                                    </div>
                                </TableCell>
                            </TableRow>
                        {/each}
                    {/if}
                </TableBody>
            </Table>
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
    description="Apakah Anda yakin ingin membatalkan dan menghapus pengajuan SPD {spdToDelete}? Tindakan ini tidak dapat dibatalkan."
    confirmText="Ya, Hapus"
    cancelText="Batal"
    onConfirm={processDelete}
    variant="destructive"
/>
