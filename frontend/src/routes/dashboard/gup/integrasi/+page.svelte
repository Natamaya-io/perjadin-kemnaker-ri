<script>
    import { api } from '$lib/shared/api';
    import { invalidateAll } from '$app/navigation';
    import Button from '$lib/shared/ui/button/Button.svelte';
    import { ConfirmationModal } from '$lib/shared/ui/confirmation-modal';

    import Select from '$lib/shared/ui/select/Select.svelte';

    export let data;

    let showModal = false;
    let isSubmitting = false;
    let errorMessage = '';

    let showAcModal = false;
    let isAcSubmitting = false;
    let acErrorMessage = '';

    let formMode = 'add'; // 'add' or 'edit'
    let currentPtId = null;

    let isDeleteConfirmOpen = false;
    let itemToDelete = null;

    let isSaveConfirmOpen = false;

    let activeTab = 'jenisPengadaan'; // 'jenisPengadaan' or 'kodeAkun'

    // Pagination state for Jenis Pengadaan
    let currentPage = 1;
    let itemsPerPage = 10;

    // Filter state
    let searchQuery = '';
    let filterAccountCode = '';

    $: ptList = (data.masterData?.procurementTypes || []).filter(pt => {
        const matchName = !searchQuery || pt.name.toLowerCase().includes(searchQuery.toLowerCase());
        const matchCode = !filterAccountCode || (pt.accountCode || '').toLowerCase().includes(filterAccountCode.toLowerCase());
        return matchName && matchCode;
    });

    $: accountCodeOptions = (data.masterData?.accountCodes || []).map(ac => ({
        value: ac.id,
        label: ac.code
    }));

    // Reset to page 1 when filter changes
    $: if (searchQuery || filterAccountCode) currentPage = 1;
    $: totalPages = Math.ceil(ptList.length / itemsPerPage) || 1;
    $: if (currentPage > totalPages && totalPages > 0) currentPage = totalPages;
    $: paginatedData = ptList.slice((currentPage - 1) * itemsPerPage, currentPage * itemsPerPage);

    function goToPage(page) {
        if (page >= 1 && page <= totalPages) currentPage = page;
    }
    
    // Pagination & Filter state for Kode Akun
    let acCurrentPage = 1;
    let searchAcQuery = '';
    
    $: acList = (data.masterData?.accountCodes || []).filter(ac => {
        const matchCode = !searchAcQuery || ac.code.toLowerCase().includes(searchAcQuery.toLowerCase());
        const matchMak = !searchAcQuery || (ac.mak || '').toLowerCase().includes(searchAcQuery.toLowerCase());
        const matchDesc = !searchAcQuery || (ac.description || '').toLowerCase().includes(searchAcQuery.toLowerCase());
        return matchCode || matchMak || matchDesc;
    });
    
    $: if (searchAcQuery) acCurrentPage = 1;
    $: acTotalPages = Math.ceil(acList.length / itemsPerPage) || 1;
    $: if (acCurrentPage > acTotalPages && acTotalPages > 0) acCurrentPage = acTotalPages;
    $: acPaginatedData = acList.slice((acCurrentPage - 1) * itemsPerPage, acCurrentPage * itemsPerPage);

    function goToAcPage(page) {
        if (page >= 1 && page <= acTotalPages) acCurrentPage = page;
    }
    
    let formData = {
        name: '',
        accountCodeId: ''
    };

    let acFormData = {
        code: '',
        mak: '',
        description: ''
    };
    
    let formAcMode = 'add'; // 'add' or 'edit'
    let currentAcId = null;

    let isDeleteAcConfirmOpen = false;
    let itemAcToDelete = null;

    function getMakFormat(code) {
        if (!code) return '-';
        
        // 1. Prioritaskan data aktual dari database (hasil CRUD) jika formatnya panjang & valid
        const accountCode = data.masterData?.accountCodes?.find(ac => ac.code === code);
        if (accountCode && accountCode.mak && accountCode.mak.length > 10) {
            return accountCode.mak;
        }

        // 2. Jika di database ternyata kosong/salah, hasilkan format baku secara dinamis
        return `2158.01.WA.2158.EBA.994.002.${code}`;
    }

    function openAddModal() {
        formMode = 'add';
        currentPtId = null;
        formData = { name: '', accountCodeId: '' };
        errorMessage = '';
        showModal = true;
    }

    function openAddAcModal() {
        formAcMode = 'add';
        currentAcId = null;
        acFormData = { code: '', mak: '', description: '' };
        acErrorMessage = '';
        showAcModal = true;
    }

    function openEditAcModal(ac) {
        formAcMode = 'edit';
        currentAcId = ac.id;
        acFormData = {
            code: ac.code,
            mak: ac.mak,
            description: ac.description || ''
        };
        acErrorMessage = '';
        showAcModal = true;
    }

    function confirmDeleteAc(acId) {
        itemAcToDelete = acId;
        isDeleteAcConfirmOpen = true;
    }

    async function processDeleteAc() {
        if (!itemAcToDelete) return;
        
        try {
            await api.deleteAccountCode(itemAcToDelete);
            await invalidateAll();
        } catch (e) {
            console.error(e);
            alert('Gagal menghapus data. Pastikan kode akun tidak sedang digunakan.');
        } finally {
            isDeleteAcConfirmOpen = false;
            itemAcToDelete = null;
        }
    }

    function openEditModal(pt) {
        formMode = 'edit';
        currentPtId = pt.id;
        
        let acId = '';
        const existingAc = data.masterData?.accountCodes?.find(ac => ac.code === pt.accountCode);
        if (existingAc) {
            acId = existingAc.id;
        }

        formData = {
            name: pt.name,
            accountCodeId: acId
        };
        errorMessage = '';
        showModal = true;
    }

    function confirmDelete(ptId) {
        itemToDelete = ptId;
        isDeleteConfirmOpen = true;
    }

    async function processDelete() {
        if (!itemToDelete) return;
        
        try {
            await api.deleteProcurementType(itemToDelete);
            await invalidateAll();
        } catch (e) {
            console.error(e);
            alert('Gagal menghapus data.');
        } finally {
            isDeleteConfirmOpen = false;
            itemToDelete = null;
        }
    }

    function confirmSave() {
        if (!formData.name || !formData.accountCodeId) {
            errorMessage = 'Semua field harus diisi.';
            return;
        }
        isSaveConfirmOpen = true;
    }

    async function saveAccountCode() {
        if (!acFormData.code || !acFormData.mak) {
            acErrorMessage = 'Kode Akun dan MAK wajib diisi.';
            return;
        }

        isAcSubmitting = true;
        acErrorMessage = '';

        try {
            if (formAcMode === 'add') {
                await api.createAccountCode({
                    code: acFormData.code,
                    mak: acFormData.mak,
                    description: acFormData.description || `Dibuat secara manual`
                });
            } else {
                await api.updateAccountCode(currentAcId, {
                    code: acFormData.code,
                    mak: acFormData.mak,
                    description: acFormData.description || `Dibuat secara manual`
                });
            }
            showAcModal = false;
            await invalidateAll();
        } catch (e) {
            console.error(e);
            acErrorMessage = 'Gagal menyimpan Kode Akun. Pastikan Kode Akun tidak duplikat.';
        } finally {
            isAcSubmitting = false;
        }
    }

    async function saveIntegration() {
        isSaveConfirmOpen = false;

        isSubmitting = true;
        errorMessage = '';

        try {
            if (formMode === 'add') {
                const names = formData.name.split(',').map(n => n.trim()).filter(n => n);
                for (const n of names) {
                    await api.createProcurementType({
                        name: n,
                        accountCodeId: formData.accountCodeId,
                        isActive: true
                    });
                }
            } else {
                await api.updateProcurementType(currentPtId, {
                    name: formData.name,
                    accountCodeId: formData.accountCodeId,
                    isActive: true
                });
            }

            showModal = false;
            await invalidateAll();
        } catch (e) {
            console.error(e);
            errorMessage = 'Gagal menyimpan data. Pastikan nama tidak duplikat.';
        } finally {
            isSubmitting = false;
        }
    }
</script>

<svelte:head>
    <title>Integrasi MAK - GUP</title>
</svelte:head>

<div class="space-y-6 pb-20 w-full">
    <!-- Header Halaman -->
    <div class="flex flex-col md:flex-row md:items-center justify-between gap-4 bg-white p-6 rounded-2xl shadow-sm border border-slate-100">
        <div>
            <h1 class="text-2xl font-bold text-slate-800 tracking-tight">Integrasi Jenis Pengadaan, Kode Akun, dan MAK</h1>
            <p class="text-sm text-slate-500 mt-1">Kelola referensi kode akun dan format MAK untuk pengadaan GUP.</p>
        </div>
        <div class="flex flex-col sm:flex-row items-center gap-3 w-full md:w-auto">
            <Button variant="outline" on:click={openAddAcModal} class="w-full sm:w-auto flex items-center justify-center gap-2">
                <svg xmlns="http://www.w3.org/2000/svg" class="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M3 10h18M7 15h1m4 0h1m-7 4h12a3 3 0 003-3V8a3 3 0 00-3-3H6a3 3 0 00-3 3v8a3 3 0 003 3z" />
                </svg>
                Buat Kode Akun
            </Button>
            <Button variant="default" on:click={openAddModal} class="w-full sm:w-auto flex items-center justify-center gap-2">
                <svg xmlns="http://www.w3.org/2000/svg" class="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4v16m8-8H4" />
                </svg>
                Tambah Jenis Pengadaan
            </Button>
        </div>
    </div>

    <!-- Tabs Header -->
    <div class="flex border-b border-slate-200 mb-6">
        <button 
            on:click={() => activeTab = 'jenisPengadaan'}
            class="px-6 py-3 font-semibold text-sm transition-colors border-b-2 {activeTab === 'jenisPengadaan' ? 'text-indigo-600 border-indigo-600 bg-indigo-50/30' : 'text-slate-500 border-transparent hover:text-slate-700 hover:border-slate-300'}"
        >
            Daftar Jenis Pengadaan
        </button>
        <button 
            on:click={() => activeTab = 'kodeAkun'}
            class="px-6 py-3 font-semibold text-sm transition-colors border-b-2 {activeTab === 'kodeAkun' ? 'text-indigo-600 border-indigo-600 bg-indigo-50/30' : 'text-slate-500 border-transparent hover:text-slate-700 hover:border-slate-300'}"
        >
            Daftar Kode Akun
        </button>
    </div>

    {#if activeTab === 'jenisPengadaan'}
    <!-- Filter Card -->
    <div class="bg-white p-4 rounded-xl shadow-sm border border-slate-100">
        <div class="flex flex-col sm:flex-row items-stretch sm:items-center gap-3">
            <!-- Search Nama -->
            <div class="relative flex-1">
                <svg xmlns="http://www.w3.org/2000/svg" class="h-4 w-4 text-slate-400 absolute left-3 top-1/2 -translate-y-1/2 pointer-events-none" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z" />
                </svg>
                <input
                    type="text"
                    id="filter-nama"
                    bind:value={searchQuery}
                    placeholder="Cari nama pengadaan..."
                    class="w-full pl-9 pr-4 py-2.5 text-sm rounded-xl border border-slate-200 bg-slate-50 text-slate-700 placeholder-slate-400 focus:outline-none focus:ring-2 focus:ring-indigo-500/20 focus:border-indigo-400 focus:bg-white transition-all"
                />
            </div>

            <!-- Filter Kode Akun -->
            <div class="relative sm:w-56">
                <svg xmlns="http://www.w3.org/2000/svg" class="h-4 w-4 text-slate-400 absolute left-3 top-1/2 -translate-y-1/2 pointer-events-none" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M7 20l4-16m2 16l4-16M6 9h14M4 15h14" />
                </svg>
                <input
                    type="text"
                    id="filter-kode-akun"
                    bind:value={filterAccountCode}
                    placeholder="Filter kode akun..."
                    class="w-full pl-9 pr-4 py-2.5 text-sm rounded-xl border border-slate-200 bg-slate-50 text-slate-700 placeholder-slate-400 focus:outline-none focus:ring-2 focus:ring-indigo-500/20 focus:border-indigo-400 focus:bg-white transition-all"
                />
            </div>

            <!-- Tombol Reset -->
            {#if searchQuery || filterAccountCode}
                <button
                    type="button"
                    on:click={() => { searchQuery = ''; filterAccountCode = ''; }}
                    class="flex items-center justify-center gap-1.5 px-3 py-2.5 text-sm font-medium text-slate-500 hover:text-red-600 bg-slate-100 hover:bg-red-50 border border-slate-200 hover:border-red-200 rounded-xl transition-colors whitespace-nowrap"
                >
                    <svg xmlns="http://www.w3.org/2000/svg" class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
                    </svg>
                    Reset
                </button>
            {/if}
        </div>

        <!-- Info Hasil Filter -->
        {#if searchQuery || filterAccountCode}
            <p class="mt-3 text-xs text-slate-500">
                Menampilkan <span class="font-semibold text-indigo-600">{ptList.length}</span> hasil dari 
                <span class="font-semibold">{data.masterData?.procurementTypes?.length || 0}</span> total data
            </p>
        {/if}
    </div>

    <!-- Tabel -->
    <div class="rounded-xl border border-slate-200 shadow-sm bg-white overflow-hidden relative flex flex-col">
        <div class="overflow-x-auto w-full relative min-h-[400px]">
            <table class="w-full text-sm text-left relative border-collapse">
                <thead class="bg-slate-50 sticky top-0 z-20 shadow-sm border-b border-slate-200">
                    <tr>
                        <th class="font-semibold text-slate-700 pl-4 py-3 bg-slate-50 whitespace-nowrap w-1/3">Jenis Pengadaan</th>
                        <th class="font-semibold text-slate-700 py-3 bg-slate-50 whitespace-nowrap w-1/4">Kode Akun</th>
                        <th class="font-semibold text-slate-700 py-3 bg-slate-50 whitespace-nowrap">Format MAK</th>
                        <th class="font-semibold text-slate-700 py-3 pr-4 bg-slate-50 whitespace-nowrap text-right">Aksi</th>
                    </tr>
                </thead>
                <tbody>
                    {#if paginatedData && paginatedData.length > 0}
                        {#each paginatedData as type (type.id)}
                            <tr class="hover:bg-slate-50/50 border-b border-slate-100 transition-colors bg-white">
                                <td class="pl-4 py-4 align-middle whitespace-nowrap font-medium text-slate-800">
                                    {type.name}
                                </td>
                                <td class="py-4 align-middle whitespace-nowrap">
                                    <span class="inline-flex items-center rounded-full px-2 py-0.5 text-[11px] font-bold tracking-wide bg-indigo-50 text-indigo-700 border border-indigo-200">
                                        {type.accountCode}
                                    </span>
                                </td>
                                <td class="py-4 align-middle whitespace-nowrap font-mono text-[13px] tracking-widest text-slate-700">
                                    {getMakFormat(type.accountCode)}
                                </td>
                                <td class="py-4 pr-4 align-middle whitespace-nowrap text-right">
                                    <div class="flex items-center justify-end gap-2">
                                        <button aria-label="Edit integrasi" on:click={() => openEditModal(type)} class="p-1.5 text-slate-400 hover:text-indigo-600 hover:bg-indigo-50 rounded-lg transition-colors">
                                            <svg xmlns="http://www.w3.org/2000/svg" class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M11 5H6a2 2 0 00-2 2v11a2 2 0 002 2h11a2 2 0 002-2v-5m-1.414-9.414a2 2 0 112.828 2.828L11.828 15H9v-2.828l8.586-8.586z" /></svg>
                                        </button>
                                        <button aria-label="Hapus integrasi" on:click={() => confirmDelete(type.id)} class="p-1.5 text-slate-400 hover:text-red-600 hover:bg-red-50 rounded-lg transition-colors">
                                            <svg xmlns="http://www.w3.org/2000/svg" class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16" /></svg>
                                        </button>
                                    </div>
                                </td>
                            </tr>
                        {/each}
                    {:else}
                        <tr>
                            <td colspan="4" class="py-8 text-center text-slate-500">
                                Tidak ada data jenis pengadaan.
                            </td>
                        </tr>
                    {/if}
                </tbody>
            </table>
        </div>
    </div>

    <!-- Pagination (Canonical Pattern) -->
    {#if totalPages > 1}
        <div class="flex items-center justify-between px-4 py-3 bg-white border border-slate-200 mt-6 rounded-xl shadow-sm">
            <!-- MOBILE: Hanya dua tombol -->
            <div class="flex flex-1 justify-between sm:hidden">
                <Button variant="outline" size="sm" disabled={currentPage === 1} on:click={() => currentPage--}>
                    Sebelumnya
                </Button>
                <Button variant="outline" size="sm" disabled={currentPage === totalPages} on:click={() => currentPage++}>
                    Selanjutnya
                </Button>
            </div>

            <!-- DESKTOP: Info data + nav lengkap -->
            <div class="hidden sm:flex sm:flex-1 sm:items-center sm:justify-between">
                <div>
                    <p class="text-sm text-slate-700">
                        Menampilkan <span class="font-medium">{(currentPage - 1) * itemsPerPage + 1}</span>
                        hingga <span class="font-medium">{Math.min(currentPage * itemsPerPage, ptList.length)}</span>
                        dari <span class="font-medium">{ptList.length}</span> hasil
                    </p>
                </div>
                <div>
                    <nav class="isolate inline-flex -space-x-px rounded-md shadow-sm" aria-label="Pagination">
                        <button on:click={() => currentPage--} disabled={currentPage === 1} class="relative inline-flex items-center rounded-l-md px-2 py-2 text-slate-400 ring-1 ring-inset ring-slate-300 hover:bg-slate-50 focus:z-20 focus:outline-offset-0 disabled:opacity-50 disabled:cursor-not-allowed">
                            <span class="sr-only">Previous</span>
                            <svg class="h-5 w-5" viewBox="0 0 20 20" fill="currentColor" aria-hidden="true">
                                <path fill-rule="evenodd" d="M12.79 5.23a.75.75 0 01-.02 1.06L8.832 10l3.938 3.71a.75.75 0 11-1.04 1.08l-4.5-4.25a.75.75 0 010-1.08l4.5-4.25a.75.75 0 011.06.02z" clip-rule="evenodd" />
                            </svg>
                        </button>
                        {#each Array(totalPages) as _, i}
                            {#if totalPages <= 7 || (i === 0 || i === totalPages - 1 || (i >= currentPage - 2 && i <= currentPage))}
                                <button on:click={() => currentPage = i + 1} class="relative inline-flex items-center px-4 py-2 text-sm font-semibold {currentPage === i + 1 ? 'z-10 bg-blue-600 text-white focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-blue-600' : 'text-slate-900 ring-1 ring-inset ring-slate-300 hover:bg-slate-50 focus:z-20 focus:outline-offset-0'}">
                                    {i + 1}
                                </button>
                            {:else if i === 1 || i === totalPages - 2}
                                <span class="relative inline-flex items-center px-4 py-2 text-sm font-semibold text-slate-700 ring-1 ring-inset ring-slate-300 focus:outline-offset-0">...</span>
                            {/if}
                        {/each}
                        <button on:click={() => currentPage++} disabled={currentPage === totalPages} class="relative inline-flex items-center rounded-r-md px-2 py-2 text-slate-400 ring-1 ring-inset ring-slate-300 hover:bg-slate-50 focus:z-20 focus:outline-offset-0 disabled:opacity-50 disabled:cursor-not-allowed">
                            <span class="sr-only">Next</span>
                            <svg class="h-5 w-5" viewBox="0 0 20 20" fill="currentColor" aria-hidden="true">
                                <path fill-rule="evenodd" d="M7.21 14.77a.75.75 0 01.02-1.06L11.168 10 7.23 6.29a.75.75 0 111.04-1.08l4.5 4.25a.75.75 0 010 1.08l-4.5 4.25a.75.75 0 01-1.06-.02z" clip-rule="evenodd" />
                            </svg>
                        </button>
                    </nav>
                </div>
            </div>
        </div>
    {/if}
    
    {:else}
    <!-- Filter Kode Akun -->
    <div class="bg-white p-4 rounded-xl shadow-sm border border-slate-100 mb-6">
        <div class="relative w-full">
            <svg xmlns="http://www.w3.org/2000/svg" class="h-4 w-4 text-slate-400 absolute left-3 top-1/2 -translate-y-1/2 pointer-events-none" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z" />
            </svg>
            <input
                type="text"
                bind:value={searchAcQuery}
                placeholder="Cari kode akun, MAK, atau keterangan..."
                class="w-full pl-9 pr-4 py-2.5 text-sm rounded-xl border border-slate-200 bg-slate-50 text-slate-700 placeholder-slate-400 focus:outline-none focus:ring-2 focus:ring-indigo-500/20 focus:border-indigo-400 focus:bg-white transition-all"
            />
        </div>
    </div>

    <!-- Tabel Kode Akun -->
    <div class="rounded-xl border border-slate-200 shadow-sm bg-white overflow-hidden relative flex flex-col">
        <div class="overflow-x-auto w-full relative min-h-[400px]">
            <table class="w-full text-sm text-left relative border-collapse">
                <thead class="bg-slate-50 sticky top-0 z-20 shadow-sm border-b border-slate-200">
                    <tr>
                        <th class="font-semibold text-slate-700 pl-4 py-3 bg-slate-50 whitespace-nowrap w-1/5">Kode Akun</th>
                        <th class="font-semibold text-slate-700 py-3 bg-slate-50 whitespace-nowrap">Format MAK</th>
                        <th class="font-semibold text-slate-700 py-3 bg-slate-50 whitespace-nowrap">Keterangan</th>
                        <th class="font-semibold text-slate-700 py-3 pr-4 bg-slate-50 whitespace-nowrap text-right">Aksi</th>
                    </tr>
                </thead>
                <tbody>
                    {#if acPaginatedData && acPaginatedData.length > 0}
                        {#each acPaginatedData as ac (ac.id)}
                            <tr class="hover:bg-slate-50/50 border-b border-slate-100 transition-colors bg-white">
                                <td class="pl-4 py-4 align-middle whitespace-nowrap">
                                    <span class="inline-flex items-center rounded-full px-2 py-0.5 text-[12px] font-bold tracking-wide bg-sky-50 text-sky-700 border border-sky-200">
                                        {ac.code}
                                    </span>
                                </td>
                                <td class="py-4 align-middle whitespace-nowrap font-mono text-[13px] tracking-widest text-slate-700">
                                    {ac.mak || '-'}
                                </td>
                                <td class="py-4 align-middle whitespace-nowrap text-slate-600">
                                    {ac.description || '-'}
                                </td>
                                <td class="py-4 pr-4 align-middle whitespace-nowrap text-right">
                                    <div class="flex items-center justify-end gap-2">
                                        <button aria-label="Edit Kode Akun" on:click={() => openEditAcModal(ac)} class="p-1.5 text-slate-400 hover:text-indigo-600 hover:bg-indigo-50 rounded-lg transition-colors">
                                            <svg xmlns="http://www.w3.org/2000/svg" class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M11 5H6a2 2 0 00-2 2v11a2 2 0 002 2h11a2 2 0 002-2v-5m-1.414-9.414a2 2 0 112.828 2.828L11.828 15H9v-2.828l8.586-8.586z" /></svg>
                                        </button>
                                        <button aria-label="Hapus Kode Akun" on:click={() => confirmDeleteAc(ac.id)} class="p-1.5 text-slate-400 hover:text-red-600 hover:bg-red-50 rounded-lg transition-colors">
                                            <svg xmlns="http://www.w3.org/2000/svg" class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16" /></svg>
                                        </button>
                                    </div>
                                </td>
                            </tr>
                        {/each}
                    {:else}
                        <tr>
                            <td colspan="4" class="py-8 text-center text-slate-500">
                                Tidak ada data kode akun.
                            </td>
                        </tr>
                    {/if}
                </tbody>
            </table>
        </div>
    </div>

    <!-- Pagination Kode Akun -->
    {#if acTotalPages > 1}
        <div class="flex items-center justify-between px-4 py-3 bg-white border border-slate-200 mt-6 rounded-xl shadow-sm">
            <div class="flex flex-1 justify-between sm:hidden">
                <Button variant="outline" size="sm" disabled={acCurrentPage === 1} on:click={() => acCurrentPage--}>Sebelumnya</Button>
                <Button variant="outline" size="sm" disabled={acCurrentPage === acTotalPages} on:click={() => acCurrentPage++}>Selanjutnya</Button>
            </div>
            <div class="hidden sm:flex sm:flex-1 sm:items-center sm:justify-between">
                <div>
                    <p class="text-sm text-slate-700">
                        Menampilkan <span class="font-medium">{(acCurrentPage - 1) * itemsPerPage + 1}</span>
                        hingga <span class="font-medium">{Math.min(acCurrentPage * itemsPerPage, acList.length)}</span>
                        dari <span class="font-medium">{acList.length}</span> hasil
                    </p>
                </div>
                <div>
                    <nav class="isolate inline-flex -space-x-px rounded-md shadow-sm" aria-label="Pagination">
                        <button on:click={() => acCurrentPage--} disabled={acCurrentPage === 1} class="relative inline-flex items-center rounded-l-md px-2 py-2 text-slate-400 ring-1 ring-inset ring-slate-300 hover:bg-slate-50 focus:z-20 focus:outline-offset-0 disabled:opacity-50 disabled:cursor-not-allowed">
                            <span class="sr-only">Previous</span>
                            <svg class="h-5 w-5" viewBox="0 0 20 20" fill="currentColor" aria-hidden="true">
                                <path fill-rule="evenodd" d="M12.79 5.23a.75.75 0 01-.02 1.06L8.832 10l3.938 3.71a.75.75 0 11-1.04 1.08l-4.5-4.25a.75.75 0 010-1.08l4.5-4.25a.75.75 0 011.06.02z" clip-rule="evenodd" />
                            </svg>
                        </button>
                        {#each Array(acTotalPages) as _, i}
                            {#if acTotalPages <= 7 || (i === 0 || i === acTotalPages - 1 || (i >= acCurrentPage - 2 && i <= acCurrentPage))}
                                <button on:click={() => acCurrentPage = i + 1} class="relative inline-flex items-center px-4 py-2 text-sm font-semibold {acCurrentPage === i + 1 ? 'z-10 bg-blue-600 text-white focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-blue-600' : 'text-slate-900 ring-1 ring-inset ring-slate-300 hover:bg-slate-50 focus:z-20 focus:outline-offset-0'}">
                                    {i + 1}
                                </button>
                            {:else if i === 1 || i === acTotalPages - 2}
                                <span class="relative inline-flex items-center px-4 py-2 text-sm font-semibold text-slate-700 ring-1 ring-inset ring-slate-300 focus:outline-offset-0">...</span>
                            {/if}
                        {/each}
                        <button on:click={() => acCurrentPage++} disabled={acCurrentPage === acTotalPages} class="relative inline-flex items-center rounded-r-md px-2 py-2 text-slate-400 ring-1 ring-inset ring-slate-300 hover:bg-slate-50 focus:z-20 focus:outline-offset-0 disabled:opacity-50 disabled:cursor-not-allowed">
                            <span class="sr-only">Next</span>
                            <svg class="h-5 w-5" viewBox="0 0 20 20" fill="currentColor" aria-hidden="true">
                                <path fill-rule="evenodd" d="M7.21 14.77a.75.75 0 01.02-1.06L11.168 10 7.23 6.29a.75.75 0 111.04-1.08l4.5 4.25a.75.75 0 010 1.08l-4.5 4.25a.75.75 0 01-1.06-.02z" clip-rule="evenodd" />
                            </svg>
                        </button>
                    </nav>
                </div>
            </div>
        </div>
    {/if}
    {/if}
</div>

<!-- Modal Form -->
{#if showModal}
    <div class="fixed inset-0 z-50 flex items-center justify-center p-4 sm:p-0">
        <!-- svelte-ignore a11y_click_events_have_key_events -->
        <!-- svelte-ignore a11y_no_static_element_interactions -->
        <div role="button" tabindex="-1" class="fixed inset-0 bg-slate-900/40 backdrop-blur-sm transition-opacity" on:click={() => showModal = false}></div>
        
        <div class="bg-white rounded-2xl shadow-xl border border-slate-100 w-full max-w-md relative z-10 animate-in fade-in zoom-in-95 duration-200 overflow-visible">
            <div class="px-6 py-4 border-b border-slate-100 flex justify-between items-center bg-slate-50/50 rounded-t-2xl">
                <h3 class="text-lg font-bold text-slate-800">
                    {formMode === 'add' ? 'Tambah Integrasi Baru' : 'Edit Integrasi'}
                </h3>
                <button aria-label="Tutup modal" on:click={() => showModal = false} class="text-slate-400 hover:text-slate-600 transition-colors">
                    <svg xmlns="http://www.w3.org/2000/svg" class="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
                    </svg>
                </button>
            </div>
            
            <div class="p-6 space-y-4">
                {#if errorMessage}
                    <div class="p-3 bg-red-50 border border-red-100 rounded-lg text-red-600 text-sm">
                        {errorMessage}
                    </div>
                {/if}

                <div>
                    <label for="integrasi-name" class="block text-sm font-medium text-slate-700 mb-1">Jenis Pengadaan</label>
                    {#if formMode === 'add'}
                        <textarea id="integrasi-name" rows="2" bind:value={formData.name} placeholder="Contoh: VIP Halim, Pass Bandara, Sewa Kendaraan" class="w-full px-3 py-2 bg-white border border-slate-200 rounded-xl text-sm focus:outline-none focus:ring-2 focus:ring-indigo-500/20 focus:border-indigo-500 transition-all shadow-[inset_-4px_-4px_10px_rgba(0,0,0,0.05),inset_4px_4px_10px_rgba(255,255,255,0.45)]"></textarea>
                        <p class="text-[11px] text-slate-500 mt-1">Pisahkan dengan koma (,) untuk menambahkan banyak jenis pengadaan sekaligus.</p>
                    {:else}
                        <input id="integrasi-name" type="text" bind:value={formData.name} placeholder="Contoh: VIP Halim" class="w-full px-3 py-2 bg-white border border-slate-200 rounded-xl text-sm focus:outline-none focus:ring-2 focus:ring-indigo-500/20 focus:border-indigo-500 transition-all shadow-[inset_-4px_-4px_10px_rgba(0,0,0,0.05),inset_4px_4px_10px_rgba(255,255,255,0.45)]" />
                    {/if}
                </div>
                
                <div>
                    <label for="integrasi-account-code" class="block text-sm font-medium text-slate-700 mb-1">Pilih Kode Akun (Tujuan Integrasi)</label>
                    <Select
                        options={accountCodeOptions}
                        bind:value={formData.accountCodeId}
                        class="w-full h-10 text-sm bg-white border-slate-200 focus:ring-indigo-500 focus:border-indigo-500"
                    />
                </div>
            </div>
            
            <div class="px-6 py-4 bg-slate-50/80 border-t border-slate-100 flex justify-end gap-3 rounded-b-2xl">
                <Button variant="default" on:click={() => showModal = false} disabled={isSubmitting}>
                    Batal
                </Button>
                <Button variant="default" on:click={confirmSave} disabled={isSubmitting}>
                    {isSubmitting ? 'Menyimpan...' : 'Simpan Integrasi'}
                </Button>
            </div>
        </div>
    </div>
{/if}

<ConfirmationModal
    bind:open={isSaveConfirmOpen}
    title={formMode === 'add' ? 'Simpan Integrasi Baru' : 'Simpan Perubahan Integrasi'}
    description="Apakah Anda yakin data integrasi ini sudah benar?"
    confirmText="Ya, Simpan"
    onConfirm={saveIntegration}
/>

<ConfirmationModal
    bind:open={isDeleteConfirmOpen}
    title="Hapus Integrasi"
    description="Apakah Anda yakin ingin menghapus jenis pengadaan ini? Tindakan ini tidak dapat dibatalkan."
    confirmText="Ya, Hapus"
    onConfirm={processDelete}
    variant="destructive"
/>

<!-- Modal Tambah Kode Akun -->
{#if showAcModal}
    <div class="fixed inset-0 z-50 flex items-center justify-center p-4 sm:p-0">
        <!-- svelte-ignore a11y_click_events_have_key_events -->
        <!-- svelte-ignore a11y_no_static_element_interactions -->
        <div role="button" tabindex="-1" class="fixed inset-0 bg-slate-900/40 backdrop-blur-sm transition-opacity" on:click={() => showAcModal = false}></div>
        
        <div class="bg-white rounded-2xl shadow-xl border border-slate-100 w-full max-w-md relative z-10 animate-in fade-in zoom-in-95 duration-200 overflow-visible">
            <div class="px-6 py-4 border-b border-slate-100 flex justify-between items-center bg-slate-50/50 rounded-t-2xl">
                <h3 class="text-lg font-bold text-slate-800">
                    {formAcMode === 'add' ? 'Buat Kode Akun Baru' : 'Edit Kode Akun'}
                </h3>
                <button aria-label="Tutup modal" on:click={() => showAcModal = false} class="text-slate-400 hover:text-slate-600 transition-colors">
                    <svg xmlns="http://www.w3.org/2000/svg" class="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
                    </svg>
                </button>
            </div>
            
            <div class="p-6 space-y-4">
                {#if acErrorMessage}
                    <div class="p-3 bg-red-50 border border-red-100 rounded-lg text-red-600 text-sm">
                        {acErrorMessage}
                    </div>
                {/if}

                <div>
                    <label for="ac-code" class="block text-sm font-medium text-slate-700 mb-1">Kode Akun</label>
                    <input id="ac-code" type="text" bind:value={acFormData.code} placeholder="Contoh: S.521119" class="w-full px-3 py-2 bg-white border border-slate-200 rounded-xl text-sm focus:outline-none focus:ring-2 focus:ring-indigo-500/20 focus:border-indigo-500 transition-all shadow-[inset_-4px_-4px_10px_rgba(0,0,0,0.05),inset_4px_4px_10px_rgba(255,255,255,0.45)]" />
                </div>
                
                <div>
                    <label for="ac-mak" class="block text-sm font-medium text-slate-700 mb-1">Format MAK Lengkap</label>
                    <input id="ac-mak" type="text" bind:value={acFormData.mak} placeholder="Contoh: 2158.01.WA.2158.EBA.994.002.S.521119" class="w-full px-3 py-2 bg-white border border-slate-200 rounded-xl text-sm font-mono focus:outline-none focus:ring-2 focus:ring-indigo-500/20 focus:border-indigo-500 transition-all shadow-[inset_-4px_-4px_10px_rgba(0,0,0,0.05),inset_4px_4px_10px_rgba(255,255,255,0.45)]" />
                </div>

                <div>
                    <label for="ac-description" class="block text-sm font-medium text-slate-700 mb-1">Keterangan Singkat</label>
                    <input id="ac-description" type="text" bind:value={acFormData.description} placeholder="Contoh: Biaya Pemeliharaan Peralatan" class="w-full px-3 py-2 bg-white border border-slate-200 rounded-xl text-sm focus:outline-none focus:ring-2 focus:ring-indigo-500/20 focus:border-indigo-500 transition-all shadow-[inset_-4px_-4px_10px_rgba(0,0,0,0.05),inset_4px_4px_10px_rgba(255,255,255,0.45)]" />
                </div>
            </div>
            
            <div class="px-6 py-4 bg-slate-50/80 border-t border-slate-100 flex justify-end gap-3 rounded-b-2xl">
                <Button variant="default" on:click={() => showAcModal = false} disabled={isAcSubmitting}>
                    Batal
                </Button>
                <Button variant="default" on:click={saveAccountCode} disabled={isAcSubmitting}>
                    {isAcSubmitting ? 'Menyimpan...' : 'Simpan Kode Akun'}
                </Button>
            </div>
        </div>
    </div>
{/if}

<ConfirmationModal
    bind:open={isDeleteAcConfirmOpen}
    title="Hapus Kode Akun"
    description="Apakah Anda yakin ingin menghapus kode akun ini? Tindakan ini tidak dapat dibatalkan dan akan gagal jika kode akun masih digunakan oleh jenis pengadaan."
    confirmText="Ya, Hapus"
    onConfirm={processDeleteAc}
    variant="destructive"
/>
