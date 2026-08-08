<script>
    import { createEventDispatcher } from 'svelte';
    import BaseModal from '$lib/shared/ui/base-modal/BaseModal.svelte';
    import { formatCurrency, getStatusBadge } from '$lib/shared/utils/utils';
    
    export let open = false;
    export let record = null;
    
    const dispatch = createEventDispatcher();
    
    function close() {
        open = false;
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
                <div>
                    <span class="block text-[11px] font-bold text-slate-500 uppercase tracking-wider mb-1">Tanggal Pelaksanaan</span>
                    <span class="font-medium text-slate-900">
                        {record.startDate && !isNaN(new Date(record.startDate).getTime()) 
                            ? new Date(record.startDate).toLocaleDateString('id-ID', { day: 'numeric', month: 'long', year: 'numeric'}) 
                            : '-'}
                    </span>
                </div>
                <div>
                    <span class="block text-[11px] font-bold text-slate-500 uppercase tracking-wider mb-1">Jenis Dalkot</span>
                    <span class="font-medium text-slate-900 capitalize">{record.type ? record.type.replace(/_/g, ' ') : '-'}</span>
                </div>
                <div>
                    <span class="block text-[11px] font-bold text-slate-500 uppercase tracking-wider mb-1">Pejabat / Stakeholder</span>
                    <span class="font-medium text-slate-900">{record.stakeholder || '-'}</span>
                </div>
                <div>
                    <span class="block text-[11px] font-bold text-slate-500 uppercase tracking-wider mb-1">Lokasi</span>
                    <span class="font-medium text-slate-900">{record.location || '-'}</span>
                </div>
                <div class="md:col-span-2">
                    <span class="block text-[11px] font-bold text-slate-500 uppercase tracking-wider mb-1">Nama Kegiatan</span>
                    <span class="font-medium text-slate-900 leading-relaxed">{record.purpose || '-'}</span>
                </div>
                <div class="md:col-span-2 flex items-center gap-3">
                    <span class="block text-[11px] font-bold text-slate-500 uppercase tracking-wider">Status:</span>
                    <span class="inline-flex items-center rounded-full px-2.5 py-0.5 text-[10px] font-bold uppercase tracking-wide border {getStatusBadge(record).class}">
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
                                <th class="px-4 py-3 text-right">Biaya SPJ</th>
                                <th class="px-4 py-3 text-right">Biaya Riil</th>
                                <th class="px-4 py-3 text-right text-indigo-900 bg-indigo-50/80">Total Dibayarkan</th>
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
                                        <td class="px-4 py-3 text-right font-mono text-[13px] text-slate-600">
                                            {formatCurrency(emp.spjCost || 0)}
                                        </td>
                                        <td class="px-4 py-3 text-right font-mono text-[13px] text-slate-600">
                                            {formatCurrency(emp.actualCost || 0)}
                                        </td>
                                        <td class="px-4 py-3 text-right font-mono text-[13px] font-bold text-indigo-700 bg-indigo-50/30">
                                            {formatCurrency(emp.totalCost || 0)}
                                        </td>
                                    </tr>
                                {/each}
                                <tr class="bg-slate-50 font-bold border-t border-slate-200">
                                    <td colspan="4" class="px-4 py-3 text-right text-slate-700 uppercase text-[11px] tracking-wider">Total Keseluruhan</td>
                                    <td class="px-4 py-3 text-right font-mono text-[14px] text-indigo-700 bg-indigo-100/50 border-l border-indigo-100">
                                        {formatCurrency(record.employeesList.reduce((sum, e) => sum + (e.totalCost || 0), 0))}
                                    </td>
                                </tr>
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
