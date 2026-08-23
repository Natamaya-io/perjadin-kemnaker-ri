<script>
    import { formatCurrency } from '$lib/shared/utils/utils';
    import { createEventDispatcher } from 'svelte';
    import { fade, scale } from 'svelte/transition';

    const dispatch = createEventDispatcher();

    /** @type {any} */
    export let record = null;
    export let open = false;

    $: spjEmployees = record?.assignments?.filter(a => a.assignmentType === 'SPJ' || a.assignmentType === 'RIIL') || [];
    $: onlySpj = record?.assignments?.filter(a => a.assignmentType === 'SPJ') || [];
    $: onlyRiil = record?.assignments?.filter(a => a.assignmentType === 'RIIL') || [];
    $: hasSplit = onlySpj.length > 0 && onlyRiil.length > 0;

    $: totalSpjCost  = onlySpj.reduce((s, a) => s + (a.spjCost || 0), 0);
    $: totalRiilCost = onlyRiil.reduce((s, a) => s + (a.actualCost || 0), 0);
    $: totalAll = totalSpjCost + totalRiilCost;

    $: executionDateStr = record?.executionDate
        ? new Date(record.executionDate).toLocaleDateString('id-ID', { weekday: 'long', day: 'numeric', month: 'long', year: 'numeric' })
        : '-';

    function close() {
        dispatch('close');
        open = false;
    }

    function handleBackdrop(e) {
        if (e.target === e.currentTarget) close();
    }
</script>

{#if open && record}
    <!-- Backdrop -->
    <!-- svelte-ignore a11y-click-events-have-key-events -->
    <!-- svelte-ignore a11y-no-static-element-interactions -->
    <div
        class="fixed inset-0 z-50 flex items-center justify-center p-4"
        style="background:rgba(15,23,42,0.55);backdrop-filter:blur(4px)"
        on:click={handleBackdrop}
        transition:fade={{ duration: 180 }}
    >
        <!-- Modal Panel -->
        <div
            class="relative w-full max-w-2xl max-h-[92vh] overflow-y-auto bg-white rounded-2xl shadow-2xl flex flex-col"
            transition:scale={{ start: 0.97, duration: 200 }}
        >
            <!-- Header -->
            <div class="flex items-start justify-between px-6 py-5 border-b border-slate-100 sticky top-0 bg-white z-10 rounded-t-2xl">
                <div>
                    <div class="flex items-center gap-2.5 mb-1">
                        <span class="bg-indigo-100 text-indigo-700 text-[11px] font-bold uppercase tracking-widest px-2.5 py-1 rounded-full border border-indigo-200">
                            Dalam Kota
                        </span>
                        <span class="font-mono font-bold text-slate-700 text-sm tracking-widest">{record.spdNumber || '-'}</span>
                    </div>
                    <h2 class="text-lg font-bold text-slate-900 leading-tight">Rincian Biaya Perjalanan Dinas</h2>
                    <p class="text-xs text-slate-500 mt-0.5">{executionDateStr}</p>
                </div>
                <button
                    id="dalkot-cost-modal-close"
                    aria-label="Tutup modal"
                    class="p-2 rounded-lg hover:bg-slate-100 text-slate-400 hover:text-slate-700 transition-colors shrink-0 ml-4"
                    on:click={close}
                >
                    <svg xmlns="http://www.w3.org/2000/svg" class="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
                    </svg>
                </button>
            </div>

            <!-- Body -->
            <div class="px-6 py-5 space-y-5">

                <!-- Info Card -->
                <div class="grid grid-cols-2 gap-3">
                    <div class="bg-slate-50 rounded-xl px-4 py-3 border border-slate-100">
                        <p class="text-[10px] font-semibold uppercase tracking-wider text-slate-400 mb-1">Kegiatan</p>
                        <p class="text-sm font-semibold text-slate-800">{record.activityName || '-'}</p>
                    </div>
                    <div class="bg-slate-50 rounded-xl px-4 py-3 border border-slate-100">
                        <p class="text-[10px] font-semibold uppercase tracking-wider text-slate-400 mb-1">Pejabat Didampingi</p>
                        <p class="text-sm font-semibold text-slate-800">{record.official || '-'}</p>
                    </div>
                    <div class="bg-slate-50 rounded-xl px-4 py-3 border border-slate-100">
                        <p class="text-[10px] font-semibold uppercase tracking-wider text-slate-400 mb-1">Lokasi</p>
                        <p class="text-sm font-semibold text-slate-800">{record.location || '-'}</p>
                    </div>
                    <div class="bg-slate-50 rounded-xl px-4 py-3 border border-slate-100">
                        <p class="text-[10px] font-semibold uppercase tracking-wider text-slate-400 mb-1">Kategori / Tipe</p>
                        <p class="text-sm font-semibold text-slate-800">{record.category || '-'} &mdash; <span class="text-indigo-600">{record.dalkotType || '-'}</span></p>
                    </div>
                </div>

                <!-- Divider -->
                <div class="border-t border-dashed border-slate-200"></div>

                <!-- SPJ Employees Table -->
                {#if onlySpj.length > 0}
                <div>
                    <div class="flex items-center justify-between mb-2">
                        <h3 class="text-sm font-bold text-slate-700 flex items-center gap-2">
                            <span class="w-2 h-2 rounded-full bg-blue-500 inline-block"></span>
                            Petugas SPJ
                            <span class="text-xs font-medium text-slate-400">({onlySpj.length} orang)</span>
                        </h3>
                        <span class="text-xs font-bold text-blue-700 bg-blue-50 px-2.5 py-1 rounded-lg border border-blue-100">{formatCurrency(totalSpjCost)}</span>
                    </div>
                    <div class="rounded-xl border border-slate-200 overflow-hidden">
                        <table class="w-full text-sm">
                            <thead class="bg-slate-50 text-[11px] uppercase tracking-wider text-slate-500 border-b border-slate-200">
                                <tr>
                                    <th class="px-4 py-2.5 text-left font-semibold">ID</th>
                                    <th class="px-4 py-2.5 text-left font-semibold">Nama Petugas</th>
                                    <th class="px-4 py-2.5 text-right font-semibold">Biaya SPJ</th>
                                </tr>
                            </thead>
                            <tbody class="divide-y divide-slate-100">
                                {#each onlySpj.sort((a,b) => (a.sequenceNumber||0)-(b.sequenceNumber||0)) as a}
                                <tr class="hover:bg-slate-50/50 transition-colors">
                                    <td class="px-4 py-3">
                                        <span class="font-mono font-bold text-slate-600 text-xs">{String(a.sequenceNumber || 0).padStart(3, '0')}</span>
                                    </td>
                                    <td class="px-4 py-3">
                                        <div class="font-medium text-slate-800">{a.user?.name || '-'}</div>
                                        {#if a.user?.jabatan}
                                            <div class="text-xs text-slate-500">{a.user.jabatan}</div>
                                        {/if}
                                    </td>
                                    <td class="px-4 py-3 text-right font-mono font-semibold text-blue-600">
                                        {formatCurrency(a.spjCost || 0)}
                                    </td>
                                </tr>
                                {/each}
                            </tbody>
                        </table>
                    </div>
                </div>
                {/if}

                <!-- RIIL Employees Table -->
                {#if onlyRiil.length > 0}
                <div>
                    <div class="flex items-center justify-between mb-2">
                        <h3 class="text-sm font-bold text-slate-700 flex items-center gap-2">
                            <span class="w-2 h-2 rounded-full bg-emerald-500 inline-block"></span>
                            Petugas Riil
                            <span class="text-xs font-medium text-slate-400">({onlyRiil.length} orang)</span>
                        </h3>
                        <span class="text-xs font-bold text-emerald-700 bg-emerald-50 px-2.5 py-1 rounded-lg border border-emerald-100">{formatCurrency(totalRiilCost)}</span>
                    </div>
                    <div class="rounded-xl border border-slate-200 overflow-hidden">
                        <table class="w-full text-sm">
                            <thead class="bg-slate-50 text-[11px] uppercase tracking-wider text-slate-500 border-b border-slate-200">
                                <tr>
                                    <th class="px-4 py-2.5 text-left font-semibold">ID</th>
                                    <th class="px-4 py-2.5 text-left font-semibold">Nama Petugas</th>
                                    <th class="px-4 py-2.5 text-right font-semibold">Biaya Riil</th>
                                </tr>
                            </thead>
                            <tbody class="divide-y divide-slate-100">
                                {#each onlyRiil.sort((a,b) => (a.sequenceNumber||0)-(b.sequenceNumber||0)) as a}
                                <tr class="hover:bg-slate-50/50 transition-colors">
                                    <td class="px-4 py-3">
                                        <span class="font-mono font-bold text-slate-600 text-xs">{String(a.sequenceNumber || 0).padStart(3, '0')}</span>
                                    </td>
                                    <td class="px-4 py-3">
                                        <div class="font-medium text-slate-800">{a.user?.name || '-'}</div>
                                        {#if a.user?.jabatan}
                                            <div class="text-xs text-slate-500">{a.user.jabatan}</div>
                                        {/if}
                                    </td>
                                    <td class="px-4 py-3 text-right font-mono font-semibold text-emerald-600">
                                        {formatCurrency(a.actualCost || 0)}
                                    </td>
                                </tr>
                                {/each}
                            </tbody>
                        </table>
                    </div>
                </div>
                {/if}

                <!-- Empty state -->
                {#if onlySpj.length === 0 && onlyRiil.length === 0}
                <div class="text-center py-8 text-slate-400 text-sm">
                    Belum ada data petugas untuk pengajuan ini.
                </div>
                {/if}

                <!-- Divider -->
                <div class="border-t border-dashed border-slate-200"></div>

                <!-- Grand Total -->
                <div class="flex items-center justify-between bg-gradient-to-r from-indigo-50 to-blue-50 rounded-xl px-5 py-4 border border-indigo-100">
                    <div>
                        <p class="text-xs font-semibold uppercase tracking-wider text-indigo-500 mb-0.5">Total Biaya Keseluruhan</p>
                        {#if hasSplit}
                        <p class="text-xs text-slate-500">
                            SPJ: {formatCurrency(totalSpjCost)} + Riil: {formatCurrency(totalRiilCost)}
                        </p>
                        {/if}
                    </div>
                    <p class="text-xl font-black text-indigo-700 font-mono">{formatCurrency(totalAll)}</p>
                </div>
            </div>

            <!-- Footer -->
            <div class="px-6 py-4 border-t border-slate-100 bg-slate-50/60 rounded-b-2xl flex justify-between items-center">
                {#if record.status === 'Submitted' || record.status === 'Draft' || !record.status}
                    <div></div>
                    <div class="flex gap-2">
                        <button
                            class="px-5 py-2 rounded-lg bg-white border border-slate-200 hover:bg-slate-50 text-slate-700 text-sm font-semibold transition-colors shadow-sm"
                            on:click={close}
                        >
                            Tutup
                        </button>
                        <button
                            class="px-5 py-2 rounded-lg bg-emerald-600 hover:bg-emerald-700 text-white text-sm font-semibold transition-colors shadow-sm shadow-emerald-200"
                            on:click={() => dispatch('approve', { record })}
                        >
                            Setujui
                        </button>
                    </div>
                {:else if record.status === 'Approved'}
                    <div></div>
                    <div class="flex gap-2">
                        <button
                            class="px-5 py-2 rounded-lg bg-white border border-slate-200 hover:bg-slate-50 text-slate-700 text-sm font-semibold transition-colors shadow-sm"
                            on:click={close}
                        >
                            Tutup
                        </button>
                        <button
                            class="px-5 py-2 rounded-lg bg-indigo-600 hover:bg-indigo-700 text-white text-sm font-semibold transition-colors shadow-sm shadow-indigo-200"
                            on:click={() => dispatch('complete', { record })}
                        >
                            Selesai
                        </button>
                    </div>
                {:else}
                    <div class="flex-1"></div>
                    <button
                        id="dalkot-cost-modal-close-btn"
                        class="px-5 py-2 rounded-lg bg-indigo-600 hover:bg-indigo-700 text-white text-sm font-semibold transition-colors shadow-sm shadow-indigo-200"
                        on:click={close}
                    >
                        Tutup
                    </button>
                {/if}
            </div>
        </div>
    </div>
{/if}
