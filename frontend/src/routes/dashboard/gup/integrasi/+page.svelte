<script>
    import { api } from '$lib/shared/api';
    import { invalidateAll } from '$app/navigation';
    import Button from '$lib/shared/ui/button/Button.svelte';

    export let data;

    let showModal = false;
    let isSubmitting = false;
    let errorMessage = '';

    let formMode = 'add'; // 'add' or 'edit'
    let currentPtId = null;

    // Pagination state
    let currentPage = 1;
    let itemsPerPage = 10;
    
    $: ptList = data.masterData?.procurementTypes || [];
    $: totalPages = Math.ceil(ptList.length / itemsPerPage) || 1;
    $: if (currentPage > totalPages && totalPages > 0) currentPage = totalPages;
    $: paginatedData = ptList.slice((currentPage - 1) * itemsPerPage, currentPage * itemsPerPage);

    function goToPage(page) {
        if (page >= 1 && page <= totalPages) currentPage = page;
    }
    
    // We combine them in one simple form for better UX
    let formData = {
        name: '',
        accountCode: '',
        mak: ''
    };

    let prevName = '';
    $: if (formData.name !== prevName) {
        prevName = formData.name;
        if (showModal && formMode === 'add') {
            const existingPt = data.masterData?.procurementTypes?.find(pt => pt.name.toLowerCase() === formData.name.toLowerCase());
            if (existingPt) {
                formData.accountCode = existingPt.accountCode;
                formData.mak = getMakFormat(existingPt.accountCode);
            }
        }
    }

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
        formData = { name: '', accountCode: '', mak: '' };
        errorMessage = '';
        showModal = true;
    }

    function openEditModal(pt) {
        formMode = 'edit';
        currentPtId = pt.id;
        
        formData = {
            name: pt.name,
            accountCode: pt.accountCode || '',
            mak: getMakFormat(pt.accountCode)
        };
        errorMessage = '';
        showModal = true;
    }

    async function handleDelete(ptId) {
        if (!confirm('Apakah Anda yakin ingin menghapus jenis pengadaan ini?')) return;
        
        try {
            await api.deleteProcurementType(ptId);
            await invalidateAll();
        } catch (e) {
            console.error(e);
            alert('Gagal menghapus data.');
        }
    }

    async function saveIntegration() {
        if (!formData.name || !formData.accountCode || !formData.mak) {
            errorMessage = 'Semua field harus diisi.';
            return;
        }

        isSubmitting = true;
        errorMessage = '';

        try {
            // 1. Resolve Account Code
            let acId = null;
            const existingAc = data.masterData?.accountCodes?.find(ac => ac.code === formData.accountCode);
            
            if (existingAc) {
                acId = existingAc.id;
                // Update the MAK if it changed
                if (existingAc.mak !== formData.mak) {
                    await api.updateAccountCode(acId, {
                        code: formData.accountCode,
                        mak: formData.mak,
                        description: existingAc.description || ''
                    });
                }
            } else {
                // Create new account code
                const newAc = await api.createAccountCode({
                    code: formData.accountCode,
                    mak: formData.mak,
                    description: `Dibuat dari Integrasi: ${formData.name}`
                });
                acId = newAc.id;
            }

            // 2. Save Procurement Type
            if (formMode === 'add') {
                await api.createProcurementType({
                    name: formData.name,
                    accountCodeId: acId,
                    isActive: true
                });
            } else {
                await api.updateProcurementType(currentPtId, {
                    name: formData.name,
                    accountCodeId: acId,
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
        <div class="flex items-center gap-3">
            <Button variant="default" on:click={openAddModal} class="w-full sm:w-auto flex items-center justify-center gap-2">
                <svg xmlns="http://www.w3.org/2000/svg" class="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4v16m8-8H4" />
                </svg>
                Tambah Integrasi
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
                                        <button aria-label="Hapus integrasi" on:click={() => handleDelete(type.id)} class="p-1.5 text-slate-400 hover:text-red-600 hover:bg-red-50 rounded-lg transition-colors">
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
        
        <!-- Bagian Pagination (Clean & Modern) -->
        {#if totalPages > 1}
        <div class="px-4 sm:px-6 py-4 border-t border-slate-100 flex flex-col sm:flex-row sm:items-center justify-between gap-4 bg-slate-50/50">
            <div class="text-sm text-slate-500 text-center sm:text-left">
                Menampilkan <span class="font-medium text-slate-700">{(currentPage - 1) * itemsPerPage + 1}</span> 
                hingga <span class="font-medium text-slate-700">{Math.min(currentPage * itemsPerPage, ptList.length)}</span> 
                dari <span class="font-medium text-slate-700">{ptList.length}</span> data
            </div>
            <div class="flex items-center justify-center gap-1 sm:gap-2">
                <button 
                    type="button" 
                    class="px-3 py-1.5 text-sm font-medium rounded-lg border border-slate-200 bg-white text-slate-600 hover:bg-slate-50 hover:text-indigo-600 disabled:opacity-50 disabled:cursor-not-allowed transition-colors"
                    disabled={currentPage === 1}
                    on:click={() => goToPage(currentPage - 1)}
                >
                    Sebelumnya
                </button>
                
                <div class="hidden sm:flex items-center gap-1">
                    {#each Array(totalPages) as _, i}
                        <button 
                            type="button"
                            class="w-8 h-8 flex items-center justify-center text-sm font-medium rounded-lg transition-colors {currentPage === i + 1 ? 'bg-indigo-600 text-white border border-indigo-600 shadow-sm' : 'border border-slate-200 bg-white text-slate-600 hover:bg-slate-50 hover:text-indigo-600'}"
                            on:click={() => goToPage(i + 1)}
                        >
                            {i + 1}
                        </button>
                    {/each}
                </div>
                
                <!-- Mobile page indicator -->
                <div class="sm:hidden text-sm font-medium text-slate-700 px-2">
                    Hal {currentPage} / {totalPages}
                </div>

                <button 
                    type="button" 
                    class="px-3 py-1.5 text-sm font-medium rounded-lg border border-slate-200 bg-white text-slate-600 hover:bg-slate-50 hover:text-indigo-600 disabled:opacity-50 disabled:cursor-not-allowed transition-colors"
                    disabled={currentPage === totalPages}
                    on:click={() => goToPage(currentPage + 1)}
                >
                    Selanjutnya
                </button>
            </div>
        </div>
        {/if}
    </div>
</div>

<!-- Modal Form -->
{#if showModal}
    <div class="fixed inset-0 z-50 flex items-center justify-center p-4 sm:p-0">
        <!-- svelte-ignore a11y_click_events_have_key_events -->
        <!-- svelte-ignore a11y_no_static_element_interactions -->
        <div role="button" tabindex="-1" class="fixed inset-0 bg-slate-900/40 backdrop-blur-sm transition-opacity" on:click={() => showModal = false}></div>
        
        <div class="bg-white rounded-2xl shadow-xl border border-slate-100 w-full max-w-md overflow-hidden relative z-10 animate-in fade-in zoom-in-95 duration-200">
            <div class="px-6 py-4 border-b border-slate-100 flex justify-between items-center bg-slate-50/50">
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
                    <input id="integrasi-name" type="text" bind:value={formData.name} placeholder="Contoh: VIP Halim" class="w-full px-3 py-2 bg-white border border-slate-200 rounded-xl text-sm focus:outline-none focus:ring-2 focus:ring-indigo-500/20 focus:border-indigo-500 transition-all shadow-[inset_-4px_-4px_10px_rgba(0,0,0,0.05),inset_4px_4px_10px_rgba(255,255,255,0.45)]" />
                </div>
                
                <div>
                    <label for="integrasi-account-code" class="block text-sm font-medium text-slate-700 mb-1">Kode Akun</label>
                    <input id="integrasi-account-code" type="text" bind:value={formData.accountCode} placeholder="Contoh: S.521119" class="w-full px-3 py-2 bg-white border border-slate-200 rounded-xl text-sm focus:outline-none focus:ring-2 focus:ring-indigo-500/20 focus:border-indigo-500 transition-all shadow-[inset_-4px_-4px_10px_rgba(0,0,0,0.05),inset_4px_4px_10px_rgba(255,255,255,0.45)]" />
                    <p class="text-[11px] text-slate-500 mt-1">Jika kode belum ada, sistem otomatis membuat referensi baru.</p>
                </div>
                
                <div>
                    <label for="integrasi-mak" class="block text-sm font-medium text-slate-700 mb-1">Format MAK Lengkap</label>
                    <input id="integrasi-mak" type="text" bind:value={formData.mak} placeholder="Contoh: 2158.01.WA.2158.EBA.994.002.S.521119" class="w-full px-3 py-2 bg-white border border-slate-200 rounded-xl text-sm font-mono focus:outline-none focus:ring-2 focus:ring-indigo-500/20 focus:border-indigo-500 transition-all shadow-[inset_-4px_-4px_10px_rgba(0,0,0,0.05),inset_4px_4px_10px_rgba(255,255,255,0.45)]" />
                </div>
            </div>
            
            <div class="px-6 py-4 bg-slate-50/80 border-t border-slate-100 flex justify-end gap-3">
                <Button variant="default" on:click={() => showModal = false} disabled={isSubmitting}>
                    Batal
                </Button>
                <Button variant="default" on:click={saveIntegration} disabled={isSubmitting}>
                    {isSubmitting ? 'Menyimpan...' : 'Simpan Integrasi'}
                </Button>
            </div>
        </div>
    </div>
{/if}
