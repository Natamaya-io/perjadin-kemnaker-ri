<script>
    import { cn } from '$lib/shared/utils/utils';
    import Button from '$lib/shared/ui/button/Button.svelte';

    export let data;

    function getMakFormat(code) {
        const accountCode = data.masterData?.accountCodes?.find(ac => ac.code === code);
        return accountCode ? accountCode.mak : '-';
    }
</script>

<svelte:head>
    <title>Integrasi MAK - GUP</title>
</svelte:head>

<div class="space-y-6 pb-20 max-w-7xl mx-auto">
    <!-- Header Halaman -->
    <div class="flex flex-col md:flex-row md:items-center justify-between gap-4 bg-white p-6 rounded-2xl shadow-sm border border-slate-100">
        <div>
            <h1 class="text-2xl font-bold text-slate-800 tracking-tight">Integrasi Jenis Pengadaan, Kode Akun, dan MAK</h1>
            <p class="text-sm text-slate-500 mt-1">Daftar referensi kode akun dan format MAK untuk setiap jenis pengadaan GUP.</p>
        </div>
        <div class="flex items-center gap-3">
            <Button variant="default" class="w-full sm:w-auto gap-2">
                <svg xmlns="http://www.w3.org/2000/svg" class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 16v1a3 3 0 003 3h10a3 3 0 003-3v-1m-4-4l-4 4m0 0l-4-4m4 4V4" />
                </svg>
                Ekspor Data
            </Button>
        </div>
    </div>

    <!-- Tabel -->
    <div class="rounded-xl border border-slate-200 shadow-sm bg-white overflow-hidden relative flex flex-col">
        <div class="overflow-x-auto w-full relative min-h-[400px]">
            <table class="w-full text-sm text-left relative border-collapse">
                <thead class="bg-slate-50 sticky top-0 z-20 shadow-sm border-b border-slate-200">
                    <tr>
                        <th class="font-semibold text-slate-700 pl-4 py-3 bg-slate-50 whitespace-nowrap w-1/3">Jenis Pengadaan</th>
                        <th class="font-semibold text-slate-700 py-3 bg-slate-50 whitespace-nowrap w-1/4">Kode Akun</th>
                        <th class="font-semibold text-slate-700 py-3 pr-4 bg-slate-50 whitespace-nowrap">Format MAK</th>
                    </tr>
                </thead>
                <tbody>
                    {#if data.masterData?.procurementTypes?.length > 0}
                        {#each data.masterData.procurementTypes as type}
                            <tr class="hover:bg-slate-50/50 border-b border-slate-100 transition-colors bg-white">
                                <td class="pl-4 py-4 align-middle whitespace-nowrap font-medium text-slate-800">
                                    {type.name}
                                </td>
                                <td class="py-4 align-middle whitespace-nowrap">
                                    <span class="inline-flex items-center rounded-full px-2 py-0.5 text-[11px] font-bold tracking-wide bg-indigo-50 text-indigo-700 border border-indigo-200">
                                        {type.accountCode}
                                    </span>
                                </td>
                                <td class="py-4 pr-4 align-middle whitespace-nowrap font-mono text-[13px] tracking-widest text-slate-700">
                                    {getMakFormat(type.accountCode)}
                                </td>
                            </tr>
                        {/each}
                    {:else}
                        <tr>
                            <td colspan="3" class="py-8 text-center text-slate-500">
                                Tidak ada data jenis pengadaan.
                            </td>
                        </tr>
                    {/if}
                </tbody>
            </table>
        </div>
    </div>
</div>
