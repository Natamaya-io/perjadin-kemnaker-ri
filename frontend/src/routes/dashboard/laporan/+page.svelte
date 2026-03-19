<script>
    import { recordsStore, updateRecord } from '$lib/features/pengajuan/store';
    import { onMount } from 'svelte';
    import { userStore } from '$lib/features/auth/store';
    import { getInitials, getStatusBadge } from '$lib/shared/utils/utils';
    import { cn } from '$lib/shared/utils/utils';

    // Components
    import AdminTableFilters from '$lib/features/admin/ui/AdminTableFilters.svelte';
    import EmptyState from '$lib/features/laporan/ui/EmptyState.svelte';
    import Table from '$lib/shared/ui/table/Table.svelte';
    import TableHeader from '$lib/shared/ui/table/TableHeader.svelte';
    import TableRow from '$lib/shared/ui/table/TableRow.svelte';
    import TableHead from '$lib/shared/ui/table/TableHead.svelte';
    import TableBody from '$lib/shared/ui/table/TableBody.svelte';
    import TableCell from '$lib/shared/ui/table/TableCell.svelte';
    import Button from '$lib/shared/ui/button/Button.svelte';
    import Dialog from '$lib/shared/ui/dialog/Dialog.svelte';
    import DialogHeader from '$lib/shared/ui/dialog/DialogHeader.svelte';
    import DialogTitle from '$lib/shared/ui/dialog/DialogTitle.svelte';
    import TripStepper from '$lib/features/dashboard/ui/roadmap/TripStepper.svelte';
    import DocumentViewer from '$lib/shared/ui/document-viewer/DocumentViewer.svelte';

    $: myRecords = $recordsStore.filter(r => r.email === $userStore.email || (r.employee && r.employee.email === $userStore.email) || $userStore.role === 'super_admin' || $userStore.role === 'keuangan' || $userStore.role === 'kasubag' || $userStore.role === 'protokol');

    // Filter & Sort State
    let searchQuery = '';
    let statusFilter = 'all'; // 'all', 'Completed', 'Pending'
    let sortOption = 'date-desc'; // 'date-desc', 'date-asc'
    let startDate = '';
    let endDate = '';

    // Detail Modal State
    let isDetailModalOpen = false;
    let selectedDetailRecord = null;
    
    // Preview Modal State
    let isPreviewOpen = false;
    let previewUrl = '';
    let previewType = '';
    let previewFilename = '';

    function openPreview(url, type, filename) {
        previewUrl = url;
        previewType = type;
        previewFilename = filename;
        isPreviewOpen = true;
    }

    function openDetailModal(record) {
        selectedDetailRecord = record;
        isDetailModalOpen = true;
    }

    function getDays(record) {
        if (!record?.startDate || !record?.endDate) return 0;
        const start = new Date(record.startDate);
        const end = new Date(record.endDate);
        start.setHours(0,0,0,0);
        end.setHours(0,0,0,0);
        const diffTime = end.getTime() - start.getTime();
        const diffDays = Math.ceil(diffTime / (1000 * 60 * 60 * 24)) + 1; 
        return diffDays > 0 ? diffDays : 0;
    }

    $: filteredRecords = myRecords
        .filter(r => r.status === 'Approved' || r.status === 'Submitted' || r.status === 'Draft') // Allow report creation from Draft status
        .filter(r => {
            const query = searchQuery.toLowerCase();
            const matchSearch =
                (r.purpose?.toLowerCase() || '').includes(query) ||
                (r.location?.toLowerCase() || '').includes(query) ||
                (r.spd?.toLowerCase() || '').includes(query) ||
                (r.employee?.name?.toLowerCase() || '').includes(query);

            const matchStatus = statusFilter === 'all' || r.reportStatus === statusFilter;

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
            return 0;
        });

    $: allRecordsSorted = [...$recordsStore].sort((a, b) => new Date(a.createdAt).getTime() - new Date(b.createdAt).getTime());
    $: recordToIndexMap = new Map(allRecordsSorted.map((r, i) => [r.id, i + 1]));

    $: groupedRecords = filteredRecords.reduce((acc, record) => {
        const isMyRecord = record.email === $userStore.email || (record.employee && record.employee.email === $userStore.email);

        if (!acc[record.spd] || isMyRecord) {
            const allEmployeesForSpd = $recordsStore.filter(r => r.spd === record.spd);
            const globalIndex = recordToIndexMap.get(record.id) || 0;
            const nomorSpdPetugas = String(globalIndex).padStart(3, '0');
            
            // Create group if not exists, or overwrite if it's MY record (to show my number)
            if (!acc[record.spd] || isMyRecord) {
                acc[record.spd] = { ...record, employeesList: allEmployeesForSpd, nomorSpdPetugas };
            }
        }
        return acc;
    }, {});

    $: uniqueRecords = Object.values(groupedRecords);
    onMount(() => {
        // Automatically mark all unviewed draft records as viewed when entering this page
        const unviewed = myRecords.filter(r => r.status === 'Draft' && !r.isViewed);
        if (unviewed.length > 0) {
            unviewed.forEach(r => {
                updateRecord(r.id, { isViewed: true });
            });
        }
    });
</script>

<div class="space-y-6 pb-20">
    <div class="flex flex-col sm:flex-row justify-between items-start sm:items-center gap-4 mb-6">
        <div>
            <h1 class="text-2xl font-bold text-slate-900 tracking-tight">Laporan Pasca Dinas</h1>
            <p class="text-sm text-slate-500 mt-1">Upload dan kelola dokumen laporan hasil perjalanan dinas Anda.</p>
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

    {#if uniqueRecords.length === 0}
        <EmptyState />
    {:else}
        <!-- Desktop Table View -->
        <div class="hidden md:block rounded-xl border border-slate-200 shadow-sm bg-white overflow-hidden">
            <div class="overflow-x-auto w-full">
                <Table class="w-full text-sm text-left">
                    <TableHeader class="bg-slate-50 border-b border-slate-200">
                        <TableRow class="hover:bg-slate-50/50">
                            <TableHead class="min-w-[120px] font-semibold text-slate-700 pl-4 py-3">{$userStore.role !== 'protokol' ? 'ID SPD' : 'No. SPD'}</TableHead>
                            <TableHead class="min-w-[250px] font-semibold text-slate-700 py-3">Tujuan & Lokasi</TableHead>
                            <TableHead class="min-w-[160px] font-semibold text-slate-700 py-3">Tanggal</TableHead>
                            <TableHead class="w-[120px] min-w-[120px] font-semibold text-slate-700 py-3">Status Laporan</TableHead>
                            <TableHead class="w-[200px] min-w-[200px] font-semibold text-slate-700 text-center pr-4 py-3">Aksi</TableHead>
                        </TableRow>
                    </TableHeader>
                    <TableBody>
                        {#each uniqueRecords as record (record.id || record.spd)}
                            <TableRow class="hover:bg-slate-50/50 border-b border-slate-100 last:border-0 transition-colors">
                                <TableCell class="font-mono text-xs text-slate-500 pl-4 py-4 align-top">
                                    {$userStore.role === 'protokol' && (record.email === $userStore.email || record.employee?.email === $userStore.email) ? record.nomorSpdPetugas : record.spd}
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
                                        <span class={cn("inline-flex items-center rounded-full px-2.5 py-0.5 text-[10px] font-bold uppercase tracking-wide border", 
                                            record.reportStatus === 'Completed' ? "bg-emerald-50 text-emerald-700 border-emerald-100" : "bg-amber-50 text-amber-700 border-amber-100")}>
                                            {record.reportStatus === 'Completed' ? 'Selesai' : 'Pending'}
                                        </span>
                                    </div>
                                </TableCell>
                                <TableCell class="text-center pr-4 py-4 align-top">
                                    <div class="flex flex-col gap-2">
                                        <div class="flex items-center justify-center gap-2">
                                            <a href={`/dashboard/laporan/${encodeURIComponent(record.spd)}`} class="flex-1">
                                                <button 
                                                    class={cn("w-full px-2 py-1.5 rounded-lg text-[11px] font-medium shadow-sm transition-all border flex items-center justify-center gap-1.5", 
                                                        record.reportStatus === 'Completed' || $userStore.role === 'kasubag' || $userStore.role === 'keuangan'
                                                        ? "bg-white text-slate-700 border-slate-200 hover:bg-slate-50 hover:text-blue-600" 
                                                        : "bg-blue-600 border-blue-600 hover:bg-blue-700 text-white shadow-blue-500/20")}
                                                >
                                                    <svg xmlns="http://www.w3.org/2000/svg" class="h-3.5 w-3.5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                                                        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z" />
                                                    </svg>
                                                    {$userStore.role === 'kasubag' || $userStore.role === 'keuangan' ? 'Lihat Laporan' : (record.reportStatus === 'Completed' ? 'Edit Laporan' : 'Input Laporan & Rincian Biaya')}
                                                </button>
                                            </a>
                                        </div>

                                    </div>
                                </TableCell>
                            </TableRow>
                        {/each}
                    </TableBody>
                </Table>
            </div>
        </div>

        <!-- Mobile Card View -->
        <div class="grid grid-cols-1 gap-4 md:hidden">
            {#each uniqueRecords as record (record.id || record.spd)}
                <div class="bg-white p-4 rounded-xl border border-slate-200 shadow-sm space-y-3">
                    <div class="flex justify-between items-start gap-2">
                        <div class="flex-1 min-w-0">
                            <span class="font-mono text-xs font-bold text-slate-800 break-all">{$userStore.role === 'protokol' && (record.email === $userStore.email || record.employee?.email === $userStore.email) ? record.nomorSpdPetugas : record.spd}</span>
                            <div class="mt-1.5 flex flex-wrap gap-1.5">
                                <span class="inline-flex items-center rounded-full px-2 py-0.5 text-[9px] font-bold capitalize tracking-wide bg-indigo-50 text-indigo-700 border border-indigo-200">
                                    {record.type ? record.type.replace(/_/g, ' ') : 'Dalam Kota'}
                                </span>
                                <span class={cn("inline-flex items-center rounded-full px-2 py-0.5 text-[9px] font-bold uppercase tracking-wide border", 
                                    record.reportStatus === 'Completed' ? "bg-emerald-50 text-emerald-700 border-emerald-100" : "bg-amber-50 text-amber-700 border-amber-100")}>
                                    {record.reportStatus === 'Completed' ? 'Selesai' : 'Pending'}
                                </span>
                            </div>
                        </div>
                        <span class="text-[9px] font-medium text-slate-600 bg-slate-100 px-2 py-1 rounded border border-slate-200 shrink-0 whitespace-nowrap">
                            {record.employeesList.length} Petugas
                        </span>
                    </div>

                    <div class="pt-2.5 border-t border-slate-100">
                        <div class="font-medium text-slate-800 text-[13px] leading-snug line-clamp-2">{record.purpose}</div>
                        <div class="text-[11px] text-slate-500 mt-1.5 flex items-start gap-1.5">
                            <svg xmlns="http://www.w3.org/2000/svg" class="h-3.5 w-3.5 text-slate-400 shrink-0 mt-0.5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M17.657 16.657L13.414 20.9a1.998 1.998 0 01-2.827 0l-4.244-4.243a8 8 0 1111.314 0z" />
                                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 11a3 3 0 11-6 0 3 3 0 016 0z" />
                            </svg>
                            <span class="line-clamp-2 leading-relaxed">{record.location}, {record.province}</span>
                        </div>
                    </div>

                    <div class="flex items-center gap-2 text-[11px] text-slate-600 bg-slate-50 p-2.5 rounded-lg border border-slate-100">
                        <svg xmlns="http://www.w3.org/2000/svg" class="h-3.5 w-3.5 text-slate-400 shrink-0" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M8 7V3m8 4V3m-9 8h10M5 21h14a2 2 0 002-2V7a2 2 0 00-2-2H5a2 2 0 00-2 2v12a2 2 0 002 2z" />
                        </svg>
                        <span>{new Date(record.startDate).toLocaleDateString('id-ID', { day: '2-digit', month: 'short', year: 'numeric'})}</span>
                        <span class="text-slate-400 text-[10px]">s/d</span>
                        <span>{new Date(record.endDate).toLocaleDateString('id-ID', { day: '2-digit', month: 'short', year: 'numeric'})}</span>
                    </div>

                    <div class="flex flex-col gap-2 pt-1.5">
                        <div class="flex items-center gap-2">
                            <a href={`/dashboard/laporan/${encodeURIComponent(record.spd)}`} class="flex-1">
                                <button 
                                    class={cn("w-full py-2.5 rounded-lg text-[10px] sm:text-[11px] font-semibold shadow-sm transition-all flex items-center justify-center gap-1.5 border", 
                                        record.reportStatus === 'Completed' || $userStore.role === 'kasubag' || $userStore.role === 'keuangan'
                                        ? "bg-white text-slate-700 border-slate-200 hover:bg-slate-50" 
                                        : "bg-blue-600 border-blue-600 hover:bg-blue-700 text-white shadow-blue-500/20")}
                                >
                                    <svg xmlns="http://www.w3.org/2000/svg" class="h-3.5 w-3.5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                                        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z" />
                                    </svg>
                                    {$userStore.role === 'kasubag' || $userStore.role === 'keuangan' ? 'Lihat Laporan' : (record.reportStatus === 'Completed' ? 'Edit Laporan' : 'Input Laporan & Rincian Biaya')}
                                </button>
                            </a>
                        </div>

                    </div>
                </div>
            {/each}
        </div>
    {/if}
</div>

<!-- Detail Modal (Surat Tugas & Roadmap) -->
<Dialog bind:open={isDetailModalOpen} class="!w-[95vw] !max-w-5xl !p-0 overflow-hidden flex flex-col rounded-2xl shadow-2xl !max-h-[85vh]">
    <!-- Header -->
    <div class="border-b border-slate-100 p-6 bg-white sticky top-0 z-20 flex items-center shrink-0">
        <DialogTitle class="text-xl font-bold text-slate-800 m-0">Detail Perjalanan Dinas</DialogTitle>
    </div>
    
    <!-- Body Content -->
    <div class="p-4 md:p-8 overflow-y-auto bg-slate-50 flex-1 space-y-6 md:space-y-8">
        {#if selectedDetailRecord}
            <!-- Info Cards Container -->
            <div class="grid grid-cols-1 md:grid-cols-3 gap-5 md:gap-6">
                <!-- Card 1: Informasi Dasar -->
                <div class="bg-white p-5 md:p-6 rounded-2xl border border-slate-200 hover:border-blue-300 shadow-sm hover:shadow-lg transition-all duration-300 flex flex-col h-full relative overflow-hidden group">
                    <div class="absolute top-0 left-0 w-full h-1.5 bg-gradient-to-r from-blue-500 to-cyan-400"></div>
                    <div class="absolute -right-6 -top-6 w-24 h-24 bg-blue-50 rounded-full opacity-50 group-hover:scale-150 transition-transform duration-700 ease-in-out"></div>
                    
                    <div class="flex items-center gap-3 mb-4 md:mb-5 pb-3 border-b border-slate-100/80 relative z-10">
                        <div class="p-2.5 bg-gradient-to-br from-blue-50 to-blue-100/50 text-blue-600 rounded-xl shadow-sm border border-blue-100">
                            <svg xmlns="http://www.w3.org/2000/svg" class="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13 16h-1v-4h-1m1-4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z" /></svg>
                        </div>
                        <h4 class="text-xs font-bold text-slate-700 uppercase tracking-widest">Informasi Dasar</h4>
                    </div>
                    <div class="space-y-4.5 flex-1 relative z-10">
                        <div>
                            <span class="text-[11px] font-bold text-slate-400 uppercase tracking-widest block mb-1.5">Nomor SPD</span>
                            <span class="font-mono text-slate-800 font-bold bg-slate-50 px-3 py-1.5 rounded-lg border border-slate-200 inline-block text-sm shadow-sm">{selectedDetailRecord.spd}</span>
                        </div>
                        <div>
                            <span class="text-[11px] font-bold text-slate-400 uppercase tracking-widest block mb-1.5">Periode</span>
                            <span class="font-semibold text-slate-700 block text-[13px] md:text-sm bg-slate-50/50 p-2 rounded-lg border border-slate-100">
                                <span class="text-blue-600 font-bold">{new Date(selectedDetailRecord.startDate).toLocaleDateString('id-ID', { day: 'numeric', month: 'short', year: 'numeric'})}</span> 
                                <span class="text-slate-400 mx-1">&rarr;</span> 
                                <span class="text-blue-600 font-bold">{new Date(selectedDetailRecord.endDate).toLocaleDateString('id-ID', { day: 'numeric', month: 'short', year: 'numeric'})}</span>
                            </span>
                        </div>
                        <div>
                            <span class="text-[11px] font-bold text-slate-400 uppercase tracking-widest block mb-1.5">Total Durasi</span>
                            <div class="flex items-baseline gap-1.5">
                                <span class="font-black text-transparent bg-clip-text bg-gradient-to-r from-blue-600 to-cyan-500 text-2xl tracking-tight">{getDays(selectedDetailRecord)}</span>
                                <span class="font-bold text-slate-500 text-sm uppercase tracking-wide">Hari</span>
                            </div>
                        </div>
                    </div>
                </div>

                <!-- Card 2: Lokasi & Agenda -->
                <div class="bg-white p-5 md:p-6 rounded-2xl border border-slate-200 hover:border-emerald-300 shadow-sm hover:shadow-lg transition-all duration-300 flex flex-col h-full relative overflow-hidden group">
                    <div class="absolute top-0 left-0 w-full h-1.5 bg-gradient-to-r from-emerald-500 to-teal-400"></div>
                    <div class="absolute -right-6 -top-6 w-24 h-24 bg-emerald-50 rounded-full opacity-50 group-hover:scale-150 transition-transform duration-700 ease-in-out"></div>
                    
                    <div class="flex items-center gap-3 mb-4 md:mb-5 pb-3 border-b border-slate-100/80 relative z-10">
                        <div class="p-2.5 bg-gradient-to-br from-emerald-50 to-emerald-100/50 text-emerald-600 rounded-xl shadow-sm border border-emerald-100">
                            <svg xmlns="http://www.w3.org/2000/svg" class="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M17.657 16.657L13.414 20.9a1.998 1.998 0 01-2.827 0l-4.244-4.243a8 8 0 1111.314 0z" /><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 11a3 3 0 11-6 0 3 3 0 016 0z" /></svg>
                        </div>
                        <h4 class="text-xs font-bold text-slate-700 uppercase tracking-widest">Lokasi & Agenda</h4>
                    </div>
                    <div class="space-y-4.5 flex-1 relative z-10">
                        <div>
                            <span class="text-[11px] font-bold text-slate-400 uppercase tracking-widest block mb-1.5">Destinasi Utama</span>
                            <div class="bg-slate-50 p-3 rounded-lg border border-slate-100">
                                <span class="font-bold text-slate-800 block text-sm leading-snug">
                                    {selectedDetailRecord.province}
                                </span>
                                <span class="text-emerald-600 font-semibold text-xs mt-0.5 block flex items-center gap-1">
                                    <svg xmlns="http://www.w3.org/2000/svg" class="h-3 w-3" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M17.657 16.657L13.414 20.9a1.998 1.998 0 01-2.827 0l-4.244-4.243a8 8 0 1111.314 0z" /></svg>
                                    {selectedDetailRecord.location}
                                </span>
                            </div>
                        </div>
                        <div>
                            <span class="text-[11px] font-bold text-slate-400 uppercase tracking-widest block mb-1.5">Nama Kegiatan</span>
                            <p class="font-medium text-slate-600 text-[13px] md:text-sm leading-relaxed p-3 bg-slate-50/50 rounded-lg border border-slate-100/50">{selectedDetailRecord.purpose || '-'}</p>
                        </div>
                    </div>
                </div>

                <!-- Card 3: Tim Pelaksana -->
                <div class="bg-white p-5 md:p-6 rounded-2xl border border-slate-200 hover:border-indigo-300 shadow-sm hover:shadow-lg transition-all duration-300 flex flex-col h-full relative overflow-hidden group">
                    <div class="absolute top-0 left-0 w-full h-1.5 bg-gradient-to-r from-indigo-500 to-purple-400"></div>
                    <div class="absolute -right-6 -top-6 w-24 h-24 bg-indigo-50 rounded-full opacity-50 group-hover:scale-150 transition-transform duration-700 ease-in-out"></div>
                    
                    <div class="flex items-center justify-between mb-4 md:mb-5 pb-3 border-b border-slate-100/80 relative z-10">
                        <div class="flex items-center gap-3">
                            <div class="p-2.5 bg-gradient-to-br from-indigo-50 to-indigo-100/50 text-indigo-600 rounded-xl shadow-sm border border-indigo-100">
                                <svg xmlns="http://www.w3.org/2000/svg" class="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4.354a4 4 0 110 5.292M15 21H3v-1a6 6 0 0112 0v1zm0 0h6v-1a6 6 0 00-9-5.197M13 7a4 4 0 11-8 0 4 4 0 018 0z" /></svg>
                            </div>
                            <h4 class="text-xs font-bold text-slate-700 uppercase tracking-widest">Tim Pelaksana</h4>
                        </div>
                        <span class="text-[11px] font-black text-white bg-gradient-to-r from-indigo-600 to-purple-500 px-2.5 py-1.5 rounded-md shadow-sm">
                            {selectedDetailRecord.employeesList?.length || 0} Petugas
                        </span>
                    </div>
                    <div class="border border-slate-200/60 rounded-xl bg-slate-50/50 overflow-hidden flex-1 h-[180px] overflow-y-auto custom-scrollbar relative z-10 shadow-inner">
                        {#if selectedDetailRecord.employeesList && selectedDetailRecord.employeesList.length > 0}
                            <div class="divide-y divide-slate-100">
                                {#each selectedDetailRecord.employeesList as emp, idx}
                                    <div class="px-4 py-3.5 flex items-center gap-3.5 hover:bg-white transition-colors cursor-default group/item">
                                        <div class="h-10 w-10 rounded-full bg-gradient-to-br from-indigo-100 to-purple-100 border border-white text-indigo-700 shadow-sm flex items-center justify-center font-black text-sm shrink-0 group-hover/item:scale-105 transition-transform">
                                            {getInitials(emp.employee?.name || '-')}
                                        </div>
                                        <div class="flex-1 min-w-0">
                                            <p class="font-bold text-slate-800 text-[13px] md:text-sm truncate">{emp.employee?.name || '-'}</p>
                                            <p class="text-[11px] text-slate-500 font-semibold truncate mt-0.5 uppercase tracking-wide">{emp.employee?.jabatan || '-'}</p>
                                        </div>
                                        <span class="text-[10px] text-slate-400 font-mono font-bold bg-white px-2 py-1 rounded shadow-sm border border-slate-100 shrink-0">{String(idx + 1).padStart(3, '0')}</span>
                                    </div>
                                {/each}
                            </div>
                        {:else}
                            <div class="h-full flex flex-col items-center justify-center p-4 text-slate-400 text-sm italic font-medium gap-2">
                                <svg xmlns="http://www.w3.org/2000/svg" class="h-8 w-8 text-slate-300" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M17 20h5v-2a3 3 0 00-5.356-1.857M17 20H7m10 0v-2c0-.656-.126-1.283-.356-1.857M7 20H2v-2a3 3 0 015.356-1.857M7 20v-2c0-.656.126-1.283.356-1.857m0 0a5.002 5.002 0 019.288 0M15 7a3 3 0 11-6 0 3 3 0 016 0zm6 3a2 2 0 11-4 0 2 2 0 014 0zM7 10a2 2 0 11-4 0 2 2 0 014 0z" /></svg>
                                <span>Tidak ada data tim</span>
                            </div>
                        {/if}
                    </div>
                </div>
            </div>

            <!-- Documents & Roadmap Blocks -->
            <div class="space-y-4 md:space-y-6">
                <!-- Surat Tugas Box -->
                <div class="bg-white border border-slate-200 rounded-xl shadow-sm overflow-hidden">
                    <div class="p-4 md:p-5 border-b border-slate-100 bg-slate-50 flex items-center gap-3">
                        <div class="p-2 bg-rose-50 text-rose-600 rounded-lg shrink-0">
                            <svg xmlns="http://www.w3.org/2000/svg" class="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z" /></svg>
                        </div>
                        <h4 class="text-sm font-bold uppercase tracking-wide text-slate-800">Dokumen Surat Tugas</h4>
                    </div>
                    <div class="p-4 md:p-6">
                        {#if selectedDetailRecord?.suratTugasPath}
                            <div class="flex flex-col md:flex-row md:items-center justify-between gap-4 p-4 md:p-6 bg-white border border-slate-200 rounded-xl transition-all hover:border-blue-200 hover:shadow-md">
                                <div class="flex items-center gap-4 md:gap-5 overflow-hidden min-w-0">
                                    <div class="h-12 w-12 md:h-14 md:w-14 rounded-xl bg-red-50 text-red-500 flex items-center justify-center border border-red-100 shrink-0">
                                        <svg xmlns="http://www.w3.org/2000/svg" class="h-6 w-6 md:h-7 md:w-7" viewBox="0 0 20 20" fill="currentColor">
                                            <path fill-rule="evenodd" d="M4 4a2 2 0 012-2h4.586A2 2 0 0112 2.586L15.414 6A2 2 0 0116 7.414V16a2 2 0 01-2 2H6a2 2 0 01-2-2V4zm2 6a1 1 0 011-1h6a1 1 0 110 2H7a1 1 0 01-1-1zm1 3a1 1 0 100 2h6a1 1 0 100-2H7z" clip-rule="evenodd" />
                                        </svg>
                                    </div>
                                    <div class="flex-1 min-w-0 space-y-1">
                                        <p class="text-sm md:text-base font-bold text-slate-800 truncate" title={selectedDetailRecord.suratTugasPath}>{selectedDetailRecord.suratTugasPath}</p>
                                        <p class="text-xs md:text-sm font-medium text-slate-500">PDF Document &bull; Bukti Penugasan Resmi</p>
                                    </div>
                                </div>
                                <button type="button" on:click={() => openPreview(`/uploads/${selectedDetailRecord.suratTugasPath}`, 'pdf', selectedDetailRecord.suratTugasPath)} class="w-full md:w-auto shrink-0 inline-flex items-center justify-center px-6 py-3 text-sm font-bold text-white bg-blue-600 hover:bg-blue-700 shadow-md shadow-blue-600/20 rounded-lg transition-all hover:-translate-y-0.5">
                                    Buka Dokumen
                                    <svg xmlns="http://www.w3.org/2000/svg" class="h-4 w-4 ml-2" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2.5" d="M10 6H6a2 2 0 00-2 2v10a2 2 0 002 2h10a2 2 0 002-2v-4M14 4h6m0 0v6m0-6L10 14" /></svg>
                                </button>
                            </div>
                        {:else}
                            <div class="text-sm text-slate-500 font-medium flex flex-col items-center justify-center gap-4 p-8 md:p-10 bg-slate-50 border border-dashed border-slate-300 rounded-xl">
                                <div class="p-3 md:p-4 bg-white rounded-full shadow-sm border border-slate-200 text-slate-400">
                                    <svg xmlns="http://www.w3.org/2000/svg" class="h-6 w-6 md:h-8 md:w-8" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                                        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-3L13.732 4c-.77-1.333-2.694-1.333-3.464 0L3.34 16c-.77 1.333.192 3 1.732 3z" />
                                    </svg>
                                </div>
                                <span class="text-center">Surat Tugas belum dilampirkan atau tidak tersedia di server.</span>
                            </div>
                        {/if}
                    </div>
                </div>

                <!-- Roadmap Box -->
                <div class="bg-white border border-slate-200 rounded-xl shadow-sm overflow-hidden">
                    <div class="p-4 md:p-5 border-b border-slate-100 bg-slate-50 flex items-center gap-3">
                        <div class="p-2 bg-amber-50 text-amber-600 rounded-lg shrink-0">
                            <svg xmlns="http://www.w3.org/2000/svg" class="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 20l-5.447-2.724A1 1 0 013 16.382V5.618a1 1 0 011.447-.894L9 7m0 13l6-3m-6 3V7m6 10l4.553 2.276A1 1 0 0021 18.382V7.618a1 1 0 00-.553-.894L15 4m0 13V4m0 0L9 7" /></svg>
                        </div>
                        <h4 class="text-sm font-bold uppercase tracking-wide text-slate-800">Roadmap Perjalanan</h4>
                    </div>
                    <div class="p-4 md:p-8 overflow-x-auto custom-scrollbar">
                        <TripStepper record={selectedDetailRecord} />
                    </div>
                </div>
            </div>
        {/if}
    </div>
    
    <!-- Footer -->
    <div class="p-4 md:p-6 border-t border-slate-200 bg-white shrink-0 sticky bottom-0 z-20 flex justify-end">
        <Button class="w-full md:w-auto min-w-[140px] h-12 bg-white text-slate-700 hover:bg-slate-50 hover:text-slate-900 border border-slate-300 font-bold text-sm rounded-lg transition-colors shadow-sm" variant="outline" on:click={() => isDetailModalOpen = false}>
            Tutup Detail
        </Button>
    </div>
</Dialog>

<!-- Document Preview Modal -->
<Dialog open={isPreviewOpen} hideCloseButton={true} on:close={() => isPreviewOpen = false} class="!w-[95vw] md:!w-[90vw] !max-w-6xl !h-[90vh] md:!h-[85vh] !p-0 overflow-hidden rounded-xl shadow-2xl z-[60]">
    <div class="h-full flex flex-col">
        <div class="flex items-center justify-between px-4 md:px-6 py-3 md:py-4 border-b border-slate-100 bg-slate-50/50 flex-none">
            <div class="flex flex-col min-w-0 pr-4">
                <h3 class="text-base md:text-lg font-bold text-slate-800 tracking-tight truncate">Pratinjau Dokumen</h3>
                {#if previewFilename}
                    <p class="text-[10px] md:text-xs text-slate-500 truncate max-w-[200px] sm:max-w-xs md:max-w-md">{previewFilename}</p>
                {/if}
            </div>
            <button type="button" class="p-2 -mr-2 text-slate-400 hover:text-red-500 hover:bg-slate-100 rounded-full transition-colors flex-shrink-0" on:click={() => isPreviewOpen = false}>
                <svg xmlns="http://www.w3.org/2000/svg" class="h-5 w-5 md:h-6 md:w-6" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
                </svg>
            </button>
        </div>
        
        <div class="flex-1 overflow-hidden bg-slate-100 relative p-0">
            {#if previewUrl}
                <DocumentViewer 
                    url={previewUrl} 
                    type={previewType} 
                    filename={previewFilename} 
                />
            {/if}
        </div>
    </div>
</Dialog>
