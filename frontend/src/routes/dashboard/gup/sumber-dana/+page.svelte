<script lang="ts">
    export let data: any;
    $: masterData = data?.masterData || { procurementTypes: [], fundingSources: [] };
</script>

<div class="space-y-6 pb-20 max-w-7xl mx-auto">
    <!-- Header -->
    <div class="flex flex-col md:flex-row md:items-center justify-between gap-4 bg-white p-6 rounded-2xl shadow-sm border border-slate-100">
        <div>
            <h1 class="text-2xl font-bold text-slate-800 tracking-tight">Rincian Sumber Dana</h1>
            <p class="text-sm text-slate-500 mt-1">Lihat alokasi sumber dana GUP dan riwayat LS.</p>
        </div>
    </div>

    <!-- Content -->
    <section class="grid grid-cols-1 gap-6 no-print">
        <div class="bg-white rounded-2xl shadow-sm border border-slate-200 p-6">
            <h3 class="text-lg font-bold text-slate-900 mb-4">Rincian Sumber Dana Bulanan</h3>
            <div class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4 gap-4 text-sm">
                {#each masterData.fundingSources as src}
                    <div class="rounded-xl border border-slate-200 p-4 bg-slate-50 shadow-sm transition-shadow hover:shadow-md">
                        <div class="flex items-center justify-between gap-3 mb-2">
                            <p class="text-slate-500 font-medium">{src.monthName}</p>
                            <span class="rounded-full bg-sky-100 px-2 py-1 text-[11px] font-bold text-sky-700">LS</span>
                        </div>
                        <p class="font-bold text-slate-900 text-lg">{src.gupLabel}</p>
                        <p class="mt-2 text-xs font-semibold text-slate-400">
                            <!-- Backend LS logic can be injected here -->
                            Data LS belum dihubungkan
                        </p>
                    </div>
                {/each}
            </div>
            
            {#if masterData.fundingSources.length === 0}
                <div class="p-12 text-center text-slate-500 w-full border border-dashed border-slate-200 rounded-xl">
                    <p>Belum ada rincian sumber dana yang tersedia.</p>
                </div>
            {/if}
        </div>
    </section>
</div>
