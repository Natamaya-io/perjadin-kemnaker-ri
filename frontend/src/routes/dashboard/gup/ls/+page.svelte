<script lang="ts">
    import { goto } from '$app/navigation';
    import BaseModal from '$lib/shared/ui/base-modal/BaseModal.svelte';
    import Button from '$lib/shared/ui/button/Button.svelte';
    import Input from '$lib/shared/ui/input/Input.svelte';
    import Select from '$lib/shared/ui/select/Select.svelte';
    import { api } from '$lib/shared/api';
    import { toast } from '$lib/shared/stores/toast';
    import { formatCurrency } from '$lib/shared/utils/utils';

    export let data: any;

    let isModalOpen = false;
    let showInfoModal = false;
    let isLoading = false;
    let editForm = {
        fundingSourceId: '',
        amount: 0
    };

    $: lsData = data?.lsData || [];
    $: transactions = data?.transactions || [];
    $: fundingSources = data?.fundingSources || [];
    $: accountCodes = data?.accountCodes || [];
    let pendingYear: string | null = null;
    $: currentYear = pendingYear || data?.year || new Date().getFullYear().toString();

    // Mapping LS data (Funding Source ID -> Amount)
    $: lsMap = new Map(lsData.map(ls => [ls.fundingSourceId, ls.amount]));

    // Hitung realisasi dari GUP Transaction berdasarkan Funding Source ID
    $: realisasiMap = transactions.reduce((acc, trx) => {
        if (trx.fundingSourceId) {
            const current = acc.get(trx.fundingSourceId) || 0;
            acc.set(trx.fundingSourceId, current + (trx.paidAmount || trx.valueAmount || 0));
        }
        return acc;
    }, new Map());

    // Summary data
    $: totalLs = Array.from(lsMap.values()).reduce((a, b) => a + Number(b), 0);
    $: totalRealisasi = Array.from(realisasiMap.values()).reduce((a, b) => a + Number(b), 0);
    $: totalSisa = totalLs - totalRealisasi;
    $: bulanTerisi = lsMap.size;

    // Filter status
    let isFilterOpen = false;
    let selectedFilter = 'all';

    // Opsi untuk filter & form
    $: sumberDanaOptions = fundingSources.map(fs => ({
        value: fs.id,
        label: `${fs.monthName} — ${fs.gupLabel}`
    }));

    $: accountCodeOptions = [
        { value: '', label: '-- Pilih Kode Akun (Opsional) --' },
        ...(accountCodes || []).map(ac => ({
            value: ac.id,
            label: ac.code
        }))
    ];

    $: filteredSources = selectedFilter === 'all' 
        ? fundingSources 
        : fundingSources.filter(f => f.id === selectedFilter);

    // Form logic
    $: selectedFs = fundingSources.find(f => f.id === editForm.fundingSourceId);
    $: currentRealisasi = selectedFs ? (realisasiMap.get(selectedFs.id) || 0) : 0;
    $: currentSisa = (editForm.amount || 0) - currentRealisasi;
    
    // Custom Dropdown State for Form & Year
    let isFormDropdownOpen = false;
    let isYearDropdownOpen = false;
    
    // Tahun yang tersedia (dinamis dari 2026 hingga tahun saat ini + 1)
    const startYear = 2026;
    const currentSystemYear = new Date().getFullYear();
    const availableYears = Array.from(
        { length: Math.max(currentSystemYear - startYear + 2, 3) }, 
        (_, i) => (startYear + i).toString()
    );

    function openEditModal(fundingSourceId?: string) {
        const id = fundingSourceId || fundingSources[0]?.id || '';
        const existingLs = lsData?.find(ls => ls.fundingSourceId === id);
        editForm = {
            fundingSourceId: id,
            accountCodeId: existingLs?.accountCodeId || '',
            amount: id ? (lsMap.get(id) || 0) : 0
        };
        isModalOpen = true;
    }

    async function handleSaveLS() {
        if (!editForm.fundingSourceId || editForm.amount < 0) return;

        isLoading = true;
        try {
            await api.saveGupLs([{
                fundingSourceId: editForm.fundingSourceId,
                accountCodeId: editForm.accountCodeId || undefined,
                amount: Number(editForm.amount),
                year: Number(currentYear)
            } as any]);
            toast.success('Data LS berhasil disimpan');
            isModalOpen = false;
            // Invalidate the data to trigger a reload
            await goto(`?year=${currentYear}`, { invalidateAll: true });
        } catch (err: any) {
            toast.error(err.message || 'Terjadi kesalahan saat menyimpan data');
        } finally {
            isLoading = false;
        }
    }
    
    function handleYearChange(e: Event) {
        const target = e.target as HTMLSelectElement;
        const newYear = target.value;
        goto(`?year=${newYear}`, { invalidateAll: true });
    }
</script>

<div class="space-y-6 pb-20 w-full">
    <!-- Header Halaman -->
    <div class="flex flex-col md:flex-row md:items-center justify-between gap-4 bg-white p-6 rounded-2xl shadow-sm border border-slate-100">
        <div>
            <h1 class="text-2xl font-bold text-slate-800 tracking-tight flex items-center gap-2">
                LS Bulanan
                <button 
                    type="button" 
                    title="Informasi LS Bulanan" 
                    on:click={() => showInfoModal = true} 
                    class="text-blue-500 hover:text-blue-700 bg-blue-50 hover:bg-blue-100 rounded-full p-1 transition-colors"
                >
                    <svg xmlns="http://www.w3.org/2000/svg" class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13 16h-1v-4h-1m1-4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
                    </svg>
                </button>
            </h1>
            <p class="text-sm text-slate-500 mt-1">Sumber dana GUP (LS Januari untuk GUP 1, dst).</p>
        </div>
        <div class="flex flex-col sm:flex-row items-center gap-3">
            <!-- Global Year Filter (Custom Dropdown) -->
            <div class="relative w-full sm:w-32">
                <button 
                    type="button" 
                    on:click={() => isYearDropdownOpen = !isYearDropdownOpen}
                    class="flex w-full items-center justify-between rounded-xl border border-slate-200 bg-slate-50 px-4 py-2.5 text-sm font-semibold text-slate-700 shadow-sm focus:outline-none focus:ring-2 focus:ring-indigo-500 focus:bg-white transition-colors"
                >
                    <span class="truncate">{currentYear}</span>
                    <svg class="h-4 w-4 text-slate-400" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 9l-7 7-7-7" />
                    </svg>
                </button>
                
                {#if isYearDropdownOpen}
                    <!-- svelte-ignore a11y-click-events-have-key-events -->
                    <!-- svelte-ignore a11y-no-static-element-interactions -->
                    <div class="fixed inset-0 z-40" on:click={() => isYearDropdownOpen = false}></div>
                    
                    <div class="absolute z-50 mt-2 w-full origin-top-right rounded-xl border border-slate-100 bg-white shadow-lg overflow-hidden animate-in fade-in zoom-in-95 duration-100">
                        <ul class="py-1">
                            {#each availableYears as year}
                                <!-- svelte-ignore a11y-click-events-have-key-events -->
                                <!-- svelte-ignore a11y-no-noninteractive-element-interactions -->
                                <li 
                                    class="relative cursor-pointer select-none py-2.5 pl-4 pr-4 text-sm transition-colors {currentYear === year ? 'font-semibold text-indigo-700 bg-indigo-50/50 hover:bg-indigo-50 flex items-center justify-between' : 'text-slate-700 hover:bg-slate-50 hover:text-indigo-600'}"
                                    on:click={() => { 
                                        pendingYear = year;
                                        isYearDropdownOpen = false;
                                        goto(`?year=${year}`, { invalidateAll: true }).then(() => {
                                            pendingYear = null;
                                        });
                                    }}
                                >
                                    <span>{year}</span>
                                    {#if currentYear === year}
                                        <svg class="h-4 w-4 text-indigo-600 shrink-0" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 13l4 4L19 7" /></svg>
                                    {/if}
                                </li>
                            {/each}
                        </ul>
                    </div>
                {/if}
            </div>

            <!-- Tombol Aksi Utama -->
            <Button variant="default" class="w-full sm:w-auto gap-2" on:click={() => openEditModal()}>
                <svg xmlns="http://www.w3.org/2000/svg" class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4v16m8-8H4" />
                </svg>
                Input LS Bulanan
            </Button>
        </div>
    </div>

    <!-- Summary Cards -->
    <div class="grid grid-cols-1 gap-4 sm:grid-cols-2 xl:grid-cols-4">
        <!-- Card 1 -->
        <div class="bg-white p-5 rounded-xl shadow-sm border border-slate-100 flex items-center justify-between gap-3">
            <div>
                <p class="text-xs font-bold uppercase tracking-wider text-slate-400">Total LS</p>
                <h2 class="mt-1 text-2xl font-black text-slate-900">{formatCurrency(totalLs)}</h2>
            </div>
            <div class="grid h-10 w-10 shrink-0 place-items-center rounded-xl bg-slate-50 text-slate-600 border border-slate-100">
                <svg xmlns="http://www.w3.org/2000/svg" class="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M8 7h12m0 0l-4-4m4 4l-4 4m0 6H4m0 0l4 4m-4-4l4-4" /></svg>
            </div>
        </div>
        <!-- Card 2 -->
        <div class="bg-white p-5 rounded-xl shadow-sm border border-slate-100 flex items-center justify-between gap-3">
            <div>
                <p class="text-xs font-bold uppercase tracking-wider text-slate-400">Realisasi GUP</p>
                <h2 class="mt-1 text-2xl font-black text-emerald-600">{formatCurrency(totalRealisasi)}</h2>
            </div>
            <div class="grid h-10 w-10 shrink-0 place-items-center rounded-xl bg-emerald-50 text-emerald-600 border border-emerald-100">
                <svg xmlns="http://www.w3.org/2000/svg" class="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13 17h8m0 0V9m0 8l-8-8-4 4-6-6" /></svg>
            </div>
        </div>
        <!-- Card 3 -->
        <div class="bg-white p-5 rounded-xl shadow-sm border border-slate-100 flex items-center justify-between gap-3">
            <div>
                <p class="text-xs font-bold uppercase tracking-wider text-slate-400">Sisa Dana</p>
                <h2 class="mt-1 text-2xl font-black text-amber-600">{formatCurrency(totalSisa)}</h2>
            </div>
            <div class="grid h-10 w-10 shrink-0 place-items-center rounded-xl bg-amber-50 text-amber-600 border border-amber-100">
                <svg xmlns="http://www.w3.org/2000/svg" class="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 15v2m-6 4h12a2 2 0 002-2v-6a2 2 0 00-2-2H6a2 2 0 00-2 2v6a2 2 0 002 2zm10-10V7a4 4 0 00-8 0v4h8z" /></svg>
            </div>
        </div>
        <!-- Card 4 -->
        <div class="bg-white p-5 rounded-xl shadow-sm border border-slate-100 flex items-center justify-between gap-3">
            <div>
                <p class="text-xs font-bold uppercase tracking-wider text-slate-400">Bulan Terisi</p>
                <h2 class="mt-1 text-2xl font-black text-indigo-600">{bulanTerisi} <span class="text-lg text-slate-400 font-semibold">/ {fundingSources.length || 12}</span></h2>
            </div>
            <div class="grid h-10 w-10 shrink-0 place-items-center rounded-xl bg-indigo-50 text-indigo-600 border border-indigo-100">
                <svg xmlns="http://www.w3.org/2000/svg" class="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M8 7V3m8 4V3m-9 8h10M5 21h14a2 2 0 002-2V7a2 2 0 00-2-2H5a2 2 0 00-2 2v12a2 2 0 002 2z" /></svg>
            </div>
        </div>
    </div>

    <!-- Filter & Grid -->
    <div class="bg-white p-5 sm:p-6 rounded-xl shadow-sm border border-slate-100">
        <div class="flex flex-col sm:flex-row sm:items-end justify-between gap-4 border-b border-slate-100 pb-5 mb-5">
            <div>
                <h2 class="text-lg font-bold text-slate-800">Pemetaan LS ke GUP</h2>
                <p class="text-sm text-slate-500 mt-1">Realisasi penggunaan dana dihitung otomatis dari riwayat transaksi GUP.</p>
            </div>
            
            <!-- Custom Dropdown untuk Filter (Sesuai Guideline) -->
            <div class="w-full sm:w-64 relative">
                <span class="block text-xs font-bold uppercase tracking-wider text-slate-500 mb-2">Filter Bulan</span>
                <button 
                    type="button" 
                    on:click={() => isFilterOpen = !isFilterOpen}
                    class="flex w-full items-center justify-between rounded-xl border border-slate-200 bg-slate-50 px-4 py-2.5 text-sm text-slate-700 focus:outline-none focus:ring-2 focus:ring-indigo-500 focus:bg-white transition-colors"
                >
                    <span class="truncate">{selectedFilter === 'all' ? 'Semua Bulan' : (fundingSources.find(f => f.id === selectedFilter)?.monthName || 'Pilih...')}</span>
                    <svg class="h-4 w-4 text-slate-400" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 9l-7 7-7-7" />
                    </svg>
                </button>
                
                {#if isFilterOpen}
                    <!-- svelte-ignore a11y-click-events-have-key-events -->
                    <!-- svelte-ignore a11y-no-static-element-interactions -->
                    <div class="fixed inset-0 z-40" on:click={() => isFilterOpen = false}></div>
                    
                    <div class="absolute z-50 mt-2 w-full origin-top-right rounded-xl border border-slate-100 bg-white shadow-lg overflow-hidden animate-in fade-in zoom-in-95 duration-100">
                        <ul class="max-h-60 overflow-y-auto custom-scrollbar py-1">
                            <!-- svelte-ignore a11y-click-events-have-key-events -->
                            <!-- svelte-ignore a11y-no-noninteractive-element-interactions -->
                            <li 
                                class="relative cursor-pointer select-none py-2.5 pl-4 pr-4 text-sm transition-colors {selectedFilter === 'all' ? 'font-semibold text-indigo-700 bg-indigo-50/50 hover:bg-indigo-50 flex items-center justify-between' : 'text-slate-700 hover:bg-slate-50 hover:text-indigo-600'}"
                                on:click={() => { selectedFilter = 'all'; isFilterOpen = false; }}
                            >
                                <span>Semua Bulan</span>
                                {#if selectedFilter === 'all'}
                                    <svg class="h-4 w-4 text-indigo-600 shrink-0" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 13l4 4L19 7" /></svg>
                                {/if}
                            </li>
                            {#each sumberDanaOptions as opt}
                                <!-- svelte-ignore a11y-click-events-have-key-events -->
                                <!-- svelte-ignore a11y-no-noninteractive-element-interactions -->
                                <li 
                                    class="relative cursor-pointer select-none py-2.5 pl-4 pr-4 text-sm transition-colors {selectedFilter === opt.value ? 'font-semibold text-indigo-700 bg-indigo-50/50 hover:bg-indigo-50 flex items-center justify-between' : 'text-slate-700 hover:bg-slate-50 hover:text-indigo-600'}"
                                    on:click={() => { selectedFilter = opt.value; isFilterOpen = false; }}
                                >
                                    <span class="truncate pr-2">{opt.label}</span>
                                    {#if selectedFilter === opt.value}
                                        <svg class="h-4 w-4 text-indigo-600 shrink-0" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 13l4 4L19 7" /></svg>
                                    {/if}
                                </li>
                            {/each}
                        </ul>
                    </div>
                {/if}
            </div>
        </div>

        <div class="grid grid-cols-1 gap-4 sm:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4">
            {#each filteredSources as fs (fs.id)}
                {@const lsAmount = lsMap.get(fs.id) || 0}
                {@const realisasi = realisasiMap.get(fs.id) || 0}
                {@const sisa = lsAmount - realisasi}
                
                <!-- Clean Modern Grid Card -->
                <div class="rounded-xl border border-slate-200 bg-white p-4 shadow-sm hover:shadow-md transition-shadow group flex flex-col justify-between">
                    <div>
                        <div class="flex items-start justify-between gap-3 mb-4">
                            <div>
                                <span class="inline-flex items-center rounded-full px-2 py-0.5 text-[9px] font-bold capitalize tracking-wide bg-indigo-50 text-indigo-700 border border-indigo-200">{fs.monthName}</span>
                                <h3 class="mt-2 text-lg font-black text-slate-900 leading-tight">{fs.gupLabel}</h3>
                            </div>
                            <button on:click={() => openEditModal(fs.id)} type="button" class="text-indigo-600 hover:text-indigo-800 bg-indigo-50 hover:bg-indigo-100 p-2 rounded-lg transition-colors border border-indigo-100 flex items-center justify-center shrink-0" title="Input LS {fs.monthName}">
                                <svg xmlns="http://www.w3.org/2000/svg" class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15.232 5.232l3.536 3.536m-2.036-5.036a2.5 2.5 0 113.536 3.536L6.5 21.036H3v-3.572L16.732 3.732z" /></svg>
                            </button>
                        </div>
                        
                        <div class="space-y-2.5 text-sm border-t border-slate-100 pt-3">
                            <div class="flex justify-between gap-3"><span class="text-slate-500 text-xs">Pemasukan (LS)</span><strong class="text-slate-700">{formatCurrency(lsAmount)}</strong></div>
                            <div class="flex justify-between gap-3"><span class="text-slate-500 text-xs">Realisasi (GUP)</span><strong class="text-emerald-600">{formatCurrency(realisasi)}</strong></div>
                            <div class="flex justify-between gap-3">
                                <span class="text-slate-500 text-xs">Sisa Saldo</span>
                                <strong class={sisa < 0 ? 'text-rose-600' : 'text-slate-900'}>{formatCurrency(sisa)}</strong>
                            </div>
                        </div>
                    </div>
                    
                    <div class="mt-4 flex items-center justify-between pt-3 border-t border-slate-50">
                        {#if lsAmount <= 0}
                            <span class="inline-flex items-center rounded-full px-2 py-0.5 text-[10px] font-bold tracking-wide bg-slate-100 text-slate-500">Belum diinput</span>
                        {:else if sisa < 0}
                            <span class="inline-flex items-center rounded-full px-2 py-0.5 text-[10px] font-bold tracking-wide bg-red-50 text-red-700 border border-red-100">Melebihi Batas</span>
                        {:else}
                            <span class="inline-flex items-center rounded-full px-2 py-0.5 text-[10px] font-bold tracking-wide bg-emerald-50 text-emerald-700 border border-emerald-100">Dana Tersedia</span>
                        {/if}
                    </div>
                </div>
            {/each}
        </div>
    </div>
</div>

<!-- Modal Informasi -->
<BaseModal bind:open={showInfoModal} maxWidth="max-w-md">
    <div slot="header">
        <h2 class="text-lg font-bold text-slate-800">Informasi LS Bulanan</h2>
    </div>
    <div slot="body">
        <div class="space-y-4 text-sm text-slate-600">
            <p>Menu <strong>LS Bulanan</strong> digunakan untuk mencatat pemasukan awal (dana LS) yang akan menjadi pagu/batas anggaran bagi operasional GUP (Ganti Uang Persediaan).</p>
            <ul class="list-disc pl-5 space-y-2 text-slate-600">
                <li>Siklus bulanan berjalan berurutan: LS Januari mendanai GUP 1, LS Februari mendanai GUP 2, dst.</li>
                <li>Setiap realisasi yang terjadi pada GUP 1 akan otomatis mengurangi saldo dari dana LS Januari.</li>
            </ul>
            <div class="mt-2 bg-amber-50 border border-amber-100 p-3 rounded-lg text-amber-800">
                <strong class="block mb-1">Catatan Penting:</strong>
                Pastikan Anda selalu memasukkan nilai LS sebelum operasional GUP bulan terkait dimulai untuk menghindari saldo sisa yang tercatat minus (Melebihi Batas).
            </div>
        </div>
    </div>
    <div slot="footer" class="flex justify-end w-full">
        <Button variant="default" on:click={() => showInfoModal = false}>Mengerti</Button>
    </div>
</BaseModal>

<!-- Modal Input LS -->
<BaseModal bind:open={isModalOpen} maxWidth="max-w-2xl">
    <div slot="header" class="flex items-center justify-between">
        <h2 class="text-lg font-bold text-slate-800 flex items-center gap-2">
            Input Pemasukan LS
            <button 
                type="button" 
                title="Informasi Input" 
                on:click={() => showInfoModal = true} 
                class="text-blue-500 hover:text-blue-700 bg-blue-50 hover:bg-blue-100 rounded-full p-1 transition-colors"
            >
                <svg xmlns="http://www.w3.org/2000/svg" class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13 16h-1v-4h-1m1-4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
                </svg>
            </button>
        </h2>
    </div>
    <div slot="body">
        <p class="text-sm text-slate-500 mb-6">Input nilai pemasukan LS yang menjadi pagu dana GUP untuk bulan tersebut.</p>
        
        <form id="lsForm" on:submit|preventDefault={handleSaveLS} class="space-y-6">
            <div class="grid grid-cols-1 gap-6 md:grid-cols-2">
                <!-- Custom Dropdown untuk Modal -->
                <div class="relative">
                    <span class="block text-sm font-semibold text-slate-700 mb-2">Bulan</span>
                    <button 
                        type="button" 
                        on:click={() => isFormDropdownOpen = !isFormDropdownOpen}
                        class="flex w-full items-center justify-between rounded-xl border border-slate-300 bg-white px-4 py-2.5 text-sm text-slate-900 focus:outline-none focus:ring-2 focus:ring-indigo-500 transition-colors"
                    >
                        <span class="truncate font-medium">{sumberDanaOptions.find(o => o.value === editForm.fundingSourceId)?.label || 'Pilih Bulan'}</span>
                        <svg class="h-4 w-4 text-slate-400 shrink-0 ml-2" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 9l-7 7-7-7" />
                        </svg>
                    </button>
                    
                    {#if isFormDropdownOpen}
                        <!-- svelte-ignore a11y-click-events-have-key-events -->
                        <!-- svelte-ignore a11y-no-static-element-interactions -->
                        <div class="fixed inset-0 z-[105]" on:click={() => isFormDropdownOpen = false}></div>
                        
                        <div class="absolute z-[110] mt-2 w-full origin-top-right rounded-xl border border-slate-100 bg-white shadow-lg overflow-hidden animate-in fade-in zoom-in-95 duration-100">
                            <ul class="max-h-60 overflow-y-auto custom-scrollbar py-1">
                                {#each sumberDanaOptions as opt}
                                    <!-- svelte-ignore a11y-click-events-have-key-events -->
                                    <!-- svelte-ignore a11y-no-noninteractive-element-interactions -->
                                    <li 
                                        class="relative cursor-pointer select-none py-2.5 pl-4 pr-4 text-sm transition-colors {editForm.fundingSourceId === opt.value ? 'font-semibold text-indigo-700 bg-indigo-50/50 hover:bg-indigo-50 flex items-center justify-between' : 'text-slate-700 hover:bg-slate-50 hover:text-indigo-600'}"
                                        on:click={() => { 
                                            editForm.fundingSourceId = opt.value; 
                                            editForm.amount = lsMap.get(opt.value) || 0;
                                            isFormDropdownOpen = false; 
                                        }}
                                    >
                                        <span class="truncate pr-2">{opt.label}</span>
                                        {#if editForm.fundingSourceId === opt.value}
                                            <svg class="h-4 w-4 text-indigo-600 shrink-0" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 13l4 4L19 7" /></svg>
                                        {/if}
                                    </li>
                                {/each}
                            </ul>
                        </div>
                    {/if}
                </div>
                
                <!-- Sumber Dana GUP -->
                <div>
                    <span class="block text-sm font-semibold text-slate-700 mb-2">Sumber Dana GUP</span>
                    <Input
                        value={selectedFs?.gupLabel || ''}
                        disabled
                        class="bg-slate-50 font-semibold text-slate-700 border-slate-200"
                    />
                </div>
            </div>

            <div class="p-6 rounded-2xl border border-indigo-100 bg-indigo-50/40 space-y-6">
                <!-- Dropdown Kode Akun LS (Dinonaktifkan sementara karena belum memiliki fungsi spesifik) -->
                <!--
                <div>
                    <span class="block text-sm font-semibold text-slate-700 mb-2">Kode Akun LS</span>
                    <Select
                        options={accountCodeOptions}
                        bind:value={editForm.accountCodeId}
                        class="w-full h-12 text-sm bg-white border-indigo-200 focus:ring-indigo-500 focus:border-indigo-500"
                    />
                </div>
                -->
                <div>
                    <span class="block text-sm font-semibold text-slate-700 mb-2">Nilai Pemasukan (Rp) <span class="text-rose-500">*</span></span>
                    <Input
                        type="text"
                        value={editForm.amount ? editForm.amount.toLocaleString('id-ID') : ''}
                        on:input={(e) => {
                            let raw = e.target.value.replace(/\D/g, '');
                            editForm.amount = raw ? parseInt(raw, 10) : 0;
                            e.target.value = editForm.amount ? editForm.amount.toLocaleString('id-ID') : '';
                        }}
                        required
                        class="text-xl font-black font-mono border-indigo-200 focus:ring-indigo-500 focus:border-indigo-500 bg-white h-12"
                        placeholder="Contoh: 50.000.000"
                    />
                </div>
                
                <!-- Realisasi GUP & Sisa Sumber Dana -->
                <div class="grid grid-cols-2 gap-4 mt-2">
                    <div class="bg-emerald-50 rounded-xl p-4 border border-emerald-100">
                        <span class="block text-xs font-semibold text-emerald-800 mb-1">Realisasi GUP</span>
                        <p class="text-xl font-black text-emerald-600 truncate" title={formatCurrency(currentRealisasi)}>{formatCurrency(currentRealisasi)}</p>
                    </div>
                    <div class="bg-indigo-50 rounded-xl p-4 border border-indigo-100">
                        <span class="block text-xs font-semibold text-indigo-800 mb-1">Sisa Sumber Dana</span>
                        <p class="text-xl font-black text-indigo-600 truncate" title={formatCurrency(currentSisa)}>{formatCurrency(currentSisa)}</p>
                    </div>
                </div>
            </div>
        </form>
    </div>
    
    <div slot="footer" class="flex justify-end gap-3 w-full">
        <Button variant="outline" on:click={() => isModalOpen = false}>Batal</Button>
        <Button 
            type="submit" 
            form="lsForm"
            variant="default"
            disabled={isLoading}
            class="flex items-center gap-2"
        >
            {#if isLoading}
                <svg class="animate-spin h-4 w-4" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24">
                    <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle>
                    <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
                </svg>
                Menyimpan...
            {:else}
                <svg xmlns="http://www.w3.org/2000/svg" class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 13l4 4L19 7" /></svg>
                Simpan Data
            {/if}
        </Button>
    </div>
</BaseModal>
