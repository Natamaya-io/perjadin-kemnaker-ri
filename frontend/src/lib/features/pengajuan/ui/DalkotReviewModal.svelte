<script>
    import { createEventDispatcher } from 'svelte';
    import BaseModal from '$lib/shared/ui/base-modal/BaseModal.svelte';
    import Button from '$lib/shared/ui/button/Button.svelte';
    import { formatCurrency, getStatusBadge } from '$lib/shared/utils/utils';
    import { userStore } from '$lib/features/auth/store';
    
    export let open = false;
    export let record = null;
    
    const dispatch = createEventDispatcher();
    
    function close() {
        open = false;
    }

    function updateStatus(newStatus) {
        dispatch('updateStatus', { id: record.id, status: newStatus });
    }
</script>

<BaseModal bind:open size="lg">
    <div slot="header">
        <h2 class="text-xl font-bold text-slate-800 tracking-tight">Detail Pengajuan Dalam Kota ({record?.spd || '-'})</h2>
    </div>

    <div slot="body">
    {#if record}
        <div class="space-y-6 text-sm text-slate-700 pb-16 px-2">
            <div class="grid grid-cols-1 md:grid-cols-2 gap-4 bg-slate-50 p-4 rounded-xl border border-slate-200 mt-2">
                <div class="flex flex-col">
                    <div class="flex items-center gap-1.5 mb-1.5 text-slate-500">
                        <svg class="w-3.5 h-3.5" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M8 7V3m8 4V3m-9 8h10M5 21h14a2 2 0 002-2V7a2 2 0 00-2-2H5a2 2 0 00-2 2v12a2 2 0 002 2z"/></svg>
                        <span class="block text-[11px] font-bold uppercase tracking-wider">Tanggal Pelaksanaan</span>
                    </div>
                    <span class="font-medium text-slate-900 bg-white px-3 py-2 rounded-lg border border-slate-100/60 shadow-sm">
                        {record.startDate && !isNaN(new Date(record.startDate).getTime()) 
                            ? new Date(record.startDate).toLocaleDateString('id-ID', { day: 'numeric', month: 'long', year: 'numeric'}) 
                            : '-'}
                    </span>
                </div>
                <div class="flex flex-col">
                    <div class="flex items-center gap-1.5 mb-1.5 text-slate-500">
                        <svg class="w-3.5 h-3.5" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 20l-5.447-2.724A1 1 0 013 16.382V5.618a1 1 0 011.447-.894L9 7m0 13l6-3m-6 3V7m6 10l4.553 2.276A1 1 0 0021 18.382V7.618a1 1 0 00-.553-.894L15 4m0 13V4m0 0L9 7"/></svg>
                        <span class="block text-[11px] font-bold uppercase tracking-wider">Jenis Dalkot</span>
                    </div>
                    <span class="font-medium text-slate-900 capitalize bg-white px-3 py-2 rounded-lg border border-slate-100/60 shadow-sm">{record.type ? record.type.replace(/_/g, ' ') : '-'}</span>
                </div>
                <div class="flex flex-col">
                    <div class="flex items-center gap-1.5 mb-1.5 text-slate-500">
                        <svg class="w-3.5 h-3.5" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M16 7a4 4 0 11-8 0 4 4 0 018 0zM12 14a7 7 0 00-7 7h14a7 7 0 00-7-7z"/></svg>
                        <span class="block text-[11px] font-bold uppercase tracking-wider">Pejabat / Stakeholder</span>
                    </div>
                    <span class="font-medium text-slate-900 bg-white px-3 py-2 rounded-lg border border-slate-100/60 shadow-sm">{record.stakeholder || '-'}</span>
                </div>
                <div class="flex flex-col">
                    <div class="flex items-center gap-1.5 mb-1.5 text-slate-500">
                        <svg class="w-3.5 h-3.5" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M17.657 16.657L13.414 20.9a1.998 1.998 0 01-2.827 0l-4.243-4.243a8 8 0 1111.314 0z"/><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 11a3 3 0 11-6 0 3 3 0 016 0z"/></svg>
                        <span class="block text-[11px] font-bold uppercase tracking-wider">Lokasi</span>
                    </div>
                    <span class="font-medium text-slate-900 bg-white px-3 py-2 rounded-lg border border-slate-100/60 shadow-sm">{record.location || '-'}</span>
                </div>
                <div class="md:col-span-2 flex flex-col">
                    <div class="flex items-center gap-1.5 mb-1.5 text-slate-500">
                        <svg class="w-3.5 h-3.5" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z"/></svg>
                        <span class="block text-[11px] font-bold uppercase tracking-wider">Nama Kegiatan</span>
                    </div>
                    <span class="font-medium text-slate-900 leading-relaxed bg-white px-3 py-2 rounded-lg border border-slate-100/60 shadow-sm">{record.purpose || '-'}</span>
                </div>
                <div class="md:col-span-2 flex items-center justify-between bg-white px-4 py-3 rounded-lg border border-slate-100/60 shadow-sm mt-2">
                    <div class="flex items-center gap-2 text-slate-500">
                        <svg class="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13 16h-1v-4h-1m1-4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z"/></svg>
                        <span class="block text-[11px] font-bold uppercase tracking-wider">Status Pengajuan</span>
                    </div>
                    <span class="inline-flex items-center rounded-full px-3 py-1 text-xs font-bold uppercase tracking-wide border shadow-sm {getStatusBadge(record).class}">
                        {getStatusBadge(record).label}
                    </span>
                </div>
            </div>

            <!-- Assignments Table -->
            <div class="mt-8">
                <h3 class="text-[13px] font-bold text-slate-800 mb-3 uppercase tracking-wider">Daftar Petugas & Rincian Biaya</h3>
                <div class="overflow-x-auto rounded-xl border border-slate-200 shadow-sm">
                    <table class="w-full text-left whitespace-nowrap">
                        <thead class="bg-slate-100/80 text-[11px] font-bold text-slate-600 uppercase tracking-wider border-b border-slate-200">
                            <tr>
                                <th class="px-4 py-3">Nama Petugas</th>
                                <th class="px-4 py-3 text-center">Jenis Penugasan</th>
                                {#if $userStore?.role !== 'protokol'}
                                <th class="px-4 py-3 text-right">Biaya SPJ</th>
                                <th class="px-4 py-3 text-right">Biaya Riil</th>
                                <th class="px-4 py-3 text-right text-indigo-900 bg-indigo-50/80">Total Dibayarkan</th>
                                {/if}
                            </tr>
                        </thead>
                        <tbody class="divide-y divide-slate-100 bg-white">
                            {#if record.employeesList && record.employeesList.length > 0}
                                {#each record.employeesList as emp}
                                    <tr class="hover:bg-slate-50/50 transition-colors">
                                        <td class="px-4 py-3">
                                            <div class="font-bold text-sm text-slate-800">{emp.employee?.name || '-'}</div>
                                            <div class="text-[11px] text-slate-500">{emp.employee?.jabatan || 'Protokol'}</div>
                                        </td>
                                        <td class="px-4 py-3 text-center align-middle">
                                            <span class="inline-flex items-center justify-center min-w-[50px] rounded-full px-2 py-0.5 text-[9px] font-bold uppercase tracking-wide border {emp.assignmentType === 'SPJ' ? 'bg-blue-50 text-blue-700 border-blue-200' : 'bg-emerald-50 text-emerald-700 border-emerald-200'}">
                                                {emp.assignmentType || '-'}
                                            </span>
                                        </td>
                                        {#if $userStore?.role !== 'protokol'}
                                        <td class="px-4 py-3 text-right font-mono text-[13px] text-slate-600">
                                            {formatCurrency(emp.spjCost || 0)}
                                        </td>
                                        <td class="px-4 py-3 text-right font-mono text-[13px] text-slate-600">
                                            {formatCurrency(emp.actualCost || 0)}
                                        </td>
                                        <td class="px-4 py-3 text-right font-mono text-[13px] font-bold text-indigo-700 bg-indigo-50/30">
                                            {formatCurrency(emp.totalCost || 0)}
                                        </td>
                                        {/if}
                                    </tr>
                                {/each}
                                {#if $userStore?.role !== 'protokol'}
                                <tr class="bg-slate-50 font-bold border-t border-slate-200">
                                    <td colspan="2" class="px-4 py-3 text-right text-slate-700 uppercase text-[11px] tracking-wider">Total Keseluruhan</td>
                                    <td colspan="3" class="px-4 py-3 text-right font-mono text-[14px] text-indigo-700 bg-indigo-100/50 border-l border-indigo-100">
                                        {formatCurrency(record.employeesList.reduce((sum, e) => sum + (e.totalCost || 0), 0))}
                                    </td>
                                </tr>
                                {/if}
                            {:else}
                                <tr>
                                    <td colspan="5" class="px-4 py-8 text-center">
                                        <div class="inline-flex flex-col items-center justify-center">
                                            <svg xmlns="http://www.w3.org/2000/svg" class="h-8 w-8 text-slate-300 mb-2" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                                                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M17 20h5v-2a3 3 0 00-5.356-1.857M17 20H7m10 0v-2c0-.656-.126-1.283-.356-1.857M7 20H2v-2a3 3 0 015.356-1.857M7 20v-2c0-.656.126-1.283.356-1.857m0 0a5.002 5.002 0 019.288 0M15 7a3 3 0 11-6 0 3 3 0 016 0zm6 3a2 2 0 11-4 0 2 2 0 014 0zM7 10a2 2 0 11-4 0 2 2 0 014 0z" />
                                            </svg>
                                            <span class="text-sm text-slate-500 font-medium">Tidak ada daftar petugas.</span>
                                        </div>
                                    </td>
                                </tr>
                            {/if}
                        </tbody>
                    </table>
                </div>
            </div>
        </div>
    {/if}
    </div>

</BaseModal>
