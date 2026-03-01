<script>
    import { recordsStore } from '$lib/stores/records';
    import { userStore } from '$lib/stores/auth';
    
    // Derived Records: Filter only records created by the current user
    $: myRecords = $recordsStore
        .filter(r => {
            const creatorEmail = r.creator?.email || r.email; // fallback to email if it was stored that way
            const creatorId = r.creatorId || r.creator?.id;
            return creatorEmail === $userStore?.email || creatorId === $userStore?.id;
        })
        .sort((a, b) => new Date(b.startDate).getTime() - new Date(a.startDate).getTime());
</script>

<div class="space-y-6 pb-20 max-w-6xl mx-auto">
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

    <div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
        {#each myRecords as record (record.id)}
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
                        <div class="font-semibold text-slate-900 text-sm">{record.employee?.name || '-'}</div>
                        <div class="text-xs text-slate-500 mb-2">{record.employee?.rank || record.employee?.golongan || '-'}</div>
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
                    <span class="font-mono font-semibold text-slate-700 text-sm">
                        {new Intl.NumberFormat('id-ID', { style: 'currency', currency: 'IDR' }).format(record.totalCost || 0)}
                    </span>
                    <div class="flex items-center gap-2">
                        {#if record.status === 'Draft'}
                            <span class="text-xs text-slate-400 italic">Belum Diproses</span>
                        {:else}
                            <span class="text-xs text-emerald-600 font-medium bg-emerald-50 px-2 py-1 rounded border border-emerald-100">Sedang/Selesai Diproses</span>
                        {/if}
                    </div>
                </div>
            </div>
        {:else}
            <div class="col-span-full p-12 text-center text-slate-500 bg-slate-50/50 rounded-xl border-2 border-dashed border-slate-200">
                <svg xmlns="http://www.w3.org/2000/svg" class="h-12 w-12 mx-auto text-slate-300 mb-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z" />
                </svg>
                Belum ada pengajuan yang Anda buat.
            </div>
        {/each}
    </div>
</div>
