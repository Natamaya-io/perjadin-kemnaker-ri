<script>
    import { recordsStore, updateRecord } from '$lib/stores/records';
    import { userStore } from '$lib/stores/auth';
    import { toast } from '$lib/stores/toast';
    
    // Components
    import AdminHeader from '$lib/components/dashboard/admin/AdminHeader.svelte';
    import AdminTableFilters from '$lib/components/dashboard/admin/AdminTableFilters.svelte';
    
    // UI Helpers
    import { ConfirmationModal } from '$lib/components/ui/confirmation-modal';

    // Filter & Sort State
    let searchQuery = '';
    let statusFilter = 'all'; // We might use this for PaymentStatus later: 'all', 'Unpaid', 'Paid'
    let sortOption = 'date-desc';
    let startDate = '';
    let endDate = '';

    let isConfirmOpen = false;
    let selectedRecord = null;
    let confirmAction = ''; // 'mark_paid' or 'mark_unpaid'

    // Accordion State
    let expandedGroups = {};

    function toggleGroup(spd) {
        expandedGroups[spd] = !expandedGroups[spd];
        expandedGroups = expandedGroups;
    }

    // Derived Records - Filter based on role
    $: myRecords = $recordsStore.filter(r => {
        if ($userStore.role === 'super_admin' || $userStore.role === 'keuangan' || $userStore.role === 'kasubag') return true;
        const isEmployee = r.employee?.email === $userStore.email;
        const isCreator = (r.creator?.email || r.email) === $userStore.email || r.creatorId === $userStore.id;
        return isEmployee || isCreator;
    });

    // Derived Records - Only those with reportStatus === 'Completed'
    $: filteredRecords = myRecords
        .filter(r => r.reportStatus === 'Completed')
        .filter(r => {
            const query = searchQuery.toLowerCase();
            const matchSearch = 
                (r.employee?.name?.toLowerCase() || '').includes(query) ||
                (r.spd?.toLowerCase() || '').includes(query) ||
                (r.location?.toLowerCase() || '').includes(query);
            
            // Re-purposing statusFilter for PaymentStatus
            const paymentStat = r.paymentStatus || 'Unpaid';
            const matchStatus = statusFilter === 'all' || 
                               (statusFilter === 'Submitted' && paymentStat === 'Unpaid') || 
                               (statusFilter === 'Approved' && paymentStat === 'Paid');
            
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
            acc[record.spd].totalCost = record.totalCost || acc[record.spd].totalCost || 0;
        }
        return acc;
    }, {});

    $: uniqueRecords = Object.values(groupedRecords);

    function openConfirmModal(record, action) {
        selectedRecord = record;
        confirmAction = action;
        isConfirmOpen = true;
    }

    async function processPaymentStatus() {
        if (!selectedRecord) return;

        try {
            const newStatus = confirmAction === 'mark_paid' ? 'Paid' : 'Unpaid';
            await updateRecord(selectedRecord.id, {
                paymentStatus: newStatus
            });
            
            toast.success(`Status pembayaran berhasil diubah menjadi ${newStatus === 'Paid' ? 'Lunas' : 'Belum Dibayar'}!`);
        } catch (e) {
            toast.error('Gagal menyimpan perubahan.');
        } finally {
            isConfirmOpen = false;
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
    <div>
        <h2 class="text-2xl font-bold tracking-tight text-slate-900">Billing Protokol</h2>
        <p class="text-sm text-slate-500 mt-1">Kelola pembayaran petugas yang telah menyelesaikan laporan perjalanan.</p>
    </div>
    
    <div class="bg-white p-4 rounded-xl border border-slate-200 shadow-sm">
        <AdminTableFilters 
            bind:searchQuery 
            bind:statusFilter 
            bind:sortOption 
            bind:startDate
            bind:endDate
        />
        <!-- Note: statusFilter options in AdminTableFilters are hardcoded. We repurpose them visually: 
             'Submitted' -> 'Unpaid', 'Approved' -> 'Paid'. We could make AdminTableFilters more generic later, 
             but for now it filters effectively using our logic above. -->
    </div>

    {#if $userStore.role === 'super_admin' || $userStore.role === 'keuangan' || $userStore.role === 'kasubag' || $userStore.role === 'protokol'}

        <div class="bg-white rounded-xl shadow-sm border border-slate-200 overflow-hidden">
            <!-- Desktop Table View -->
            <div class="hidden md:block overflow-x-auto w-full">
                <table class="w-full text-left text-sm border-collapse min-w-[800px]">
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
                                            <span class="text-xs font-semibold text-slate-400 ml-2">({record.employeesList.length} Petugas)</span>
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
                                <!-- Nested Table Head for Employee Data -->
                                <tr class="bg-slate-100 border-y border-slate-200 text-[11px] uppercase tracking-wider font-semibold text-slate-500 shadow-inner">
                                    <th class="px-6 py-3 text-left font-semibold">Lokasi Dinas</th>
                                    <th class="px-6 py-3 text-left font-semibold">Petugas</th>
                                    <th class="px-6 py-3 text-right font-semibold">Total Akhir</th>
                                    <th class="px-6 py-3 text-center font-semibold">Status Pembayaran</th>
                                    <th class="px-6 py-3 text-right font-semibold">Aksi</th>
                                </tr>
                                {#each record.employeesList as empRecord}
                                    <tr class="hover:bg-slate-50 transition-colors bg-white">
                                        <td class="px-6 py-4 align-middle w-1/3">
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
                                        <td class="px-6 py-4 align-middle w-1/4">
                                            <div class="font-medium text-slate-900">{empRecord.employee?.name || '-'}</div>
                                            <div class="text-xs text-slate-500 mb-2">
                                                {#if empRecord.employee?.jabatan}
                                                    <div class="font-medium text-slate-700">{empRecord.employee.jabatan}</div>
                                                {/if}
                                                {#if (empRecord.employee?.pangkat && empRecord.employee.pangkat !== '-') || (empRecord.employee?.golongan && empRecord.employee.golongan !== '-')}
                                                    <div class="mt-0.5">
                                                        {#if (empRecord.employee?.pangkat && empRecord.employee.pangkat !== '-') && (empRecord.employee?.golongan && empRecord.employee.golongan !== '-')}
                                                            {empRecord.employee.pangkat} ({empRecord.employee.golongan})
                                                        {:else}
                                                            {(empRecord.employee?.pangkat !== '-' ? empRecord.employee?.pangkat : null) || (empRecord.employee?.golongan !== '-' ? empRecord.employee?.golongan : null) || ''}
                                                        {/if}
                                                    </div>
                                                {/if}
                                            </div>
                                        </td>
                                        <td class="px-6 py-4 align-middle text-right font-mono font-medium text-blue-600">
                                            {new Intl.NumberFormat('id-ID', { style: 'currency', currency: 'IDR' }).format(empRecord.totalCost || 0)}
                                        </td>
                                        <td class="px-6 py-4 align-middle text-center">
                                            <span class="inline-flex items-center rounded-full px-2.5 py-0.5 text-[10px] font-bold uppercase tracking-wide border {(empRecord.paymentStatus === 'Paid') ? 'bg-emerald-50 text-emerald-700 border-emerald-200' : 'bg-rose-50 text-rose-700 border-rose-200'}">
                                                {empRecord.paymentStatus === 'Paid' ? 'Lunas' : 'Belum Dibayar'}
                                            </span>
                                        </td>
                                        <td class="px-6 py-4 align-middle">
                                            <div class="flex items-center justify-end gap-2">
                                                <button class="p-1.5 hover:bg-slate-100 rounded-md text-slate-400 hover:text-slate-700 transition-colors" title="Cetak Rincian" on:click={(e) => { e.stopPropagation(); generateRincian(empRecord); }}>
                                                    <svg xmlns="http://www.w3.org/2000/svg" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><rect width="16" height="20" x="4" y="2" rx="2"/><line x1="8" x2="16" y1="6" y2="6"/><line x1="16" x2="16" y1="14" y2="18"/><path d="M16 10h.01"/><path d="M12 10h.01"/><path d="M8 10h.01"/><path d="M12 14h.01"/><path d="M8 14h.01"/><path d="M12 18h.01"/><path d="M8 18h.01"/></svg>
                                                </button>
                                                {#if $userStore.role === 'keuangan' || $userStore.role === 'super_admin'}
                                                    {#if empRecord.paymentStatus !== 'Paid'}
                                                        <button class="bg-indigo-600 hover:bg-indigo-700 text-white px-3 py-1.5 rounded-md text-xs font-medium transition-all shadow-sm shadow-indigo-500/20 whitespace-nowrap" on:click={(e) => { e.stopPropagation(); openConfirmModal(empRecord, 'mark_paid'); }}>
                                                            Bayar
                                                        </button>
                                                    {:else}
                                                        <button class="bg-white border border-slate-200 text-slate-600 hover:bg-slate-50 hover:text-slate-900 px-3 py-1.5 rounded-md text-xs font-medium transition-all shadow-sm whitespace-nowrap" on:click={(e) => { e.stopPropagation(); openConfirmModal(empRecord, 'mark_unpaid'); }}>
                                                            Batal Bayar
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
                                    Belum ada tagihan protokol yang masuk.
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
                                    <span class="font-semibold text-[10px] text-slate-500 uppercase tracking-wider">{record.employeesList.length} Petugas</span>
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
                                                </div>
                                            </div>
                                            <span class="shrink-0 inline-flex items-center rounded-full px-2 py-0.5 text-[9px] font-bold uppercase tracking-wide border {(empRecord.paymentStatus === 'Paid') ? 'bg-emerald-50 text-emerald-700 border-emerald-200' : 'bg-rose-50 text-rose-700 border-rose-200'}">
                                                {empRecord.paymentStatus === 'Paid' ? 'Lunas' : 'Belum Dibayar'}
                                            </span>
                                        </div>

                                        <div class="flex flex-col gap-3 border-t border-slate-50 pt-3">
                                            <div class="flex justify-between items-center">
                                                <span class="text-[10px] font-semibold text-slate-400 uppercase tracking-wider mb-0.5">Total Akhir</span>
                                                <span class="font-mono font-bold text-blue-600 text-sm">
                                                    {new Intl.NumberFormat('id-ID', { style: 'currency', currency: 'IDR' }).format(empRecord.totalCost || 0)}
                                                </span>
                                            </div>
                                            <div class="flex items-center gap-2 w-full">
                                                <button class="p-1.5 hover:bg-slate-100 rounded-md text-slate-400 hover:text-slate-700 transition-colors bg-white border border-slate-200 shadow-sm" title="Cetak Rincian" on:click={(e) => { e.stopPropagation(); generateRincian(empRecord); }}>
                                                    <svg xmlns="http://www.w3.org/2000/svg" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><rect width="16" height="20" x="4" y="2" rx="2"/><line x1="8" x2="16" y1="6" y2="6"/><line x1="16" x2="16" y1="14" y2="18"/><path d="M16 10h.01"/><path d="M12 10h.01"/><path d="M8 10h.01"/><path d="M12 14h.01"/><path d="M8 14h.01"/><path d="M12 18h.01"/><path d="M8 18h.01"/></svg>
                                                </button>
                                                {#if $userStore.role === 'keuangan' || $userStore.role === 'super_admin'}
                                                    {#if empRecord.paymentStatus !== 'Paid'}
                                                        <button class="bg-indigo-600 hover:bg-indigo-700 text-white px-3 py-1.5 rounded-md text-xs font-medium transition-all shadow-sm shadow-indigo-500/20 flex-1" on:click={(e) => { e.stopPropagation(); openConfirmModal(empRecord, 'mark_paid'); }}>
                                                            Bayar
                                                        </button>
                                                    {:else}
                                                        <button class="bg-white border border-slate-200 text-slate-600 hover:bg-slate-50 hover:text-slate-900 px-3 py-1.5 rounded-md text-xs font-medium transition-all shadow-sm flex-1" on:click={(e) => { e.stopPropagation(); openConfirmModal(empRecord, 'mark_unpaid'); }}>
                                                            Batal Bayar
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
                        Belum ada tagihan protokol yang masuk.
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

    <ConfirmationModal
        bind:open={isConfirmOpen}
        title={confirmAction === 'mark_paid' ? "Konfirmasi Pembayaran" : "Batalkan Pembayaran"}
        description={confirmAction === 'mark_paid' ? "Apakah Anda yakin ingin menandai tagihan ini sebagai Lunas?" : "Apakah Anda yakin ingin membatalkan status lunas tagihan ini?"}
        confirmText={confirmAction === 'mark_paid' ? "Ya, Tandai Lunas" : "Ya, Batalkan"}
        onConfirm={processPaymentStatus}
    />
</div>