<script>
    import { recordsStore } from '$lib/stores/records';
    import { userStore } from '$lib/stores/auth';
    import { getInitials, getStatusBadge } from '$lib/utils';
    import { cn } from '$lib/utils';

    // Components
    import AdminTableFilters from '$lib/components/dashboard/admin/AdminTableFilters.svelte';
    import EmptyState from '$lib/components/dashboard/laporan/EmptyState.svelte';
    import Table from '$lib/components/ui/table/Table.svelte';
    import TableHeader from '$lib/components/ui/table/TableHeader.svelte';
    import TableRow from '$lib/components/ui/table/TableRow.svelte';
    import TableHead from '$lib/components/ui/table/TableHead.svelte';
    import TableBody from '$lib/components/ui/table/TableBody.svelte';
    import TableCell from '$lib/components/ui/table/TableCell.svelte';

    $: myRecords = $recordsStore.filter(r => r.email === $userStore.email || (r.employee && r.employee.email === $userStore.email) || $userStore.role === 'super_admin' || $userStore.role === 'keuangan' || $userStore.role === 'kasubag');

    // Filter & Sort State
    let searchQuery = '';
    let statusFilter = 'all'; // 'all', 'Completed', 'Pending'
    let sortOption = 'date-desc'; // 'date-desc', 'date-asc'
    let startDate = '';
    let endDate = '';

    $: filteredRecords = myRecords
        .filter(r => r.status === 'Approved' || r.status === 'Submitted') // Requirement: SPJ must be Submitted or Approved to make a report
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

    $: groupedRecords = filteredRecords.reduce((acc, record) => {
        if (!acc[record.spd]) {
            acc[record.spd] = { ...record, employeesList: [record] };
        } else {
            acc[record.spd].employeesList.push(record);
        }
        return acc;
    }, {});

    $: uniqueRecords = Object.values(groupedRecords);
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
        <div class="rounded-xl border border-slate-200 shadow-sm bg-white overflow-hidden">
            <div class="overflow-x-auto w-full">
                <Table class="w-full text-sm text-left">
                    <TableHeader class="bg-slate-50 border-b border-slate-200">
                        <TableRow class="hover:bg-slate-50/50">
                            <TableHead class="min-w-[120px] font-semibold text-slate-700 pl-4 py-3">No. SPD</TableHead>
                            <TableHead class="min-w-[250px] font-semibold text-slate-700 py-3">Tujuan & Lokasi</TableHead>
                            <TableHead class="min-w-[160px] font-semibold text-slate-700 py-3">Tanggal</TableHead>
                            <TableHead class="w-[120px] min-w-[120px] font-semibold text-slate-700 py-3">Status Laporan</TableHead>
                            <TableHead class="w-[140px] min-w-[140px] font-semibold text-slate-700 text-center pr-4 py-3">Aksi</TableHead>
                        </TableRow>
                    </TableHeader>
                    <TableBody>
                        {#each uniqueRecords as record (record.id || record.spd)}
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
                                        <span class={cn("inline-flex items-center rounded-full px-2.5 py-0.5 text-[10px] font-bold uppercase tracking-wide border", 
                                            record.reportStatus === 'Completed' ? "bg-emerald-50 text-emerald-700 border-emerald-100" : "bg-amber-50 text-amber-700 border-amber-100")}>
                                            {record.reportStatus === 'Completed' ? 'Selesai' : 'Pending'}
                                        </span>
                                    </div>
                                </TableCell>
                                <TableCell class="text-center pr-4 py-4 align-top">
                                    <div class="flex items-center justify-center gap-2">
                                        <a href={`/dashboard/laporan/${encodeURIComponent(record.spd)}`}>
                                            <button 
                                                class={cn("px-3 py-1.5 rounded-lg text-xs font-medium shadow-sm transition-all whitespace-nowrap border flex items-center gap-1.5", 
                                                    record.reportStatus === 'Completed' || $userStore.role === 'kasubag' || $userStore.role === 'keuangan' || $userStore.role === 'super_admin'
                                                    ? "bg-white text-slate-700 border-slate-200 hover:bg-slate-50 hover:text-blue-600" 
                                                    : "bg-blue-600 border-blue-600 hover:bg-blue-700 text-white shadow-blue-500/20")}
                                            >
                                                {#if record.reportStatus !== 'Completed' && $userStore.role !== 'kasubag' && $userStore.role !== 'keuangan' && $userStore.role !== 'super_admin'}
                                                    <svg xmlns="http://www.w3.org/2000/svg" class="h-3.5 w-3.5" viewBox="0 0 20 20" fill="currentColor">
                                                        <path fill-rule="evenodd" d="M10 3a1 1 0 011 1v5h5a1 1 0 110 2h-5v5a1 1 0 11-2 0v-5H4a1 1 0 110-2h5V4a1 1 0 011-1z" clip-rule="evenodd" />
                                                    </svg>
                                                {:else}
                                                    <svg xmlns="http://www.w3.org/2000/svg" class="h-3.5 w-3.5" viewBox="0 0 20 20" fill="currentColor">
                                                        <path d="M10 12a2 2 0 100-4 2 2 0 000 4z" />
                                                        <path fill-rule="evenodd" d="M.458 10C1.732 5.943 5.522 3 10 3s8.268 2.943 9.542 7c-1.274 4.057-5.064 7-9.542 7S1.732 14.057.458 10zM14 10a4 4 0 11-8 0 4 4 0 018 0z" clip-rule="evenodd" />
                                                    </svg>
                                                {/if}
                                                {$userStore.role === 'kasubag' || $userStore.role === 'keuangan' || $userStore.role === 'super_admin' ? 'Lihat Laporan' : (record.reportStatus === 'Completed' ? 'Edit Laporan' : 'Buat Laporan')}
                                            </button>
                                        </a>
                                        {#if record.reportStatus === 'Completed'}
                                            <button 
                                                class="text-blue-600 hover:text-blue-800 bg-blue-50 hover:bg-blue-100 p-1.5 rounded-lg transition-colors border border-blue-200 shadow-sm" 
                                                title="Cetak Laporan"
                                                on:click={() => window.open(`/print?type=laporan&spd=${encodeURIComponent(record.spd)}`, '_blank')}
                                            >
                                                <svg xmlns="http://www.w3.org/2000/svg" class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                                                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M17 17h2a2 2 0 002-2v-4a2 2 0 00-2-2H5a2 2 0 00-2 2v4a2 2 0 002 2h2m2 4h6a2 2 0 002-2v-4a2-2 0 00-2-2H9a2 2 0 00-2 2v4a2 2 0 002 2zm8-12V5a2 2 0 00-2-2H9a2 2 0 00-2 2v4h10z" />
                                                </svg>
                                            </button>
                                        {/if}
                                    </div>
                                </TableCell>
                            </TableRow>
                        {/each}
                    </TableBody>
                </Table>
            </div>
        </div>
    {/if}
</div>
