<script lang="ts">
    import { goto } from '$app/navigation';
    import { page } from '$app/stores';
    import { api } from '$lib/shared/api';
    import { formatCurrency } from '$lib/shared/utils/utils';
    import Button from '$lib/shared/ui/button/Button.svelte';
    import Select from '$lib/shared/ui/select/Select.svelte';
    import Datepicker from '$lib/shared/ui/datepicker/Datepicker.svelte';

    export let data: any;

    let masterData = data.masterData || { fundingSources: [], procurementTypes: [] };
    let fundingSources = masterData.fundingSources || [];
    let procurementTypes = masterData.procurementTypes || [];

    let form = {
        businessId: '', // Auto-generate on mount, user can edit
        paymentDescription: '',
        procurementTypeId: '',
        fundingSourceId: '',
        valueAmount: 0,
        paidAmount: 0,
        taxAmount: 0,
        receiptDate: new Date().toISOString().split('T')[0],
        recipient: '',
        pum: ''
    };

    let isEditMode = false;
    let editId = '';

    let isSubmitting = false;
    let errorMessage = '';

    // Formatted strings for input display
    let formattedValueAmount = '';
    let formattedPaidAmount = '';
    let formattedTaxAmount = '';

    function handleCurrencyInput(e: Event, field: 'valueAmount' | 'paidAmount' | 'taxAmount') {
        const target = e.target as HTMLInputElement;
        let val = target.value.replace(/\D/g, '');
        const num = val !== '' ? parseInt(val, 10) : 0;
        form[field] = num;
        
        const formatted = val !== '' ? num.toLocaleString('id-ID') : '';
        if (field === 'valueAmount') formattedValueAmount = formatted;
        if (field === 'paidAmount') formattedPaidAmount = formatted;
        if (field === 'taxAmount') formattedTaxAmount = formatted;
    }

    let supportingDocumentFile: File | { name: string } | null = null;
    let existingDocumentFile: any = null;
    let fileInputRef: HTMLInputElement;

    function handleFileChange(event: Event) {
        const input = event.target as HTMLInputElement;
        if (input.files && input.files.length > 0) {
            supportingDocumentFile = input.files[0];
        } else {
            supportingDocumentFile = null;
        }
    }

    $: selisih = form.valueAmount - form.paidAmount - form.taxAmount;
    $: selectedProcurementType = procurementTypes.find(t => t.id === form.procurementTypeId);
    $: autoKodeAkun = selectedProcurementType?.accountCode || '-';
    $: autoMak = selectedProcurementType?.accountMak || '-';

    import { onMount } from 'svelte';
    onMount(async () => {
        const idParam = $page.url.searchParams.get('id');
        if (idParam) {
            isEditMode = true;
            editId = idParam;
            try {
                const data = await api.getGupPengajuanById(editId);
                form = {
                    businessId: data.businessId,
                    paymentDescription: data.paymentDescription,
                    procurementTypeId: data.procurementTypeId,
                    fundingSourceId: data.fundingSourceId || '',
                    valueAmount: data.valueAmount,
                    paidAmount: data.paidAmount,
                    taxAmount: data.taxAmount,
                    receiptDate: data.receiptDate ? new Date(data.receiptDate).toISOString().split('T')[0] : new Date().toISOString().split('T')[0],
                    recipient: data.recipient,
                    pum: data.pum
                };
                if (data.documentFile) {
                    existingDocumentFile = data.documentFile;
                    supportingDocumentFile = { name: data.documentFile.originalName };
                }
                formattedValueAmount = form.valueAmount.toLocaleString('id-ID');
                formattedPaidAmount = form.paidAmount.toLocaleString('id-ID');
                formattedTaxAmount = form.taxAmount.toLocaleString('id-ID');
            } catch (e) {
                console.error("Failed to fetch edit data:", e);
                errorMessage = "Gagal memuat data pengajuan untuk diubah.";
            }
        } else {
            try {
                const res = await api.getNextBusinessID(window.fetch);
                if (res && res.nextId) {
                    form.businessId = res.nextId;
                }
            } catch (e) {
                console.warn("Failed to fetch next business ID:", e);
                form.businessId = 'gup_001';
            }
        }
    });

    async function handleSubmit() {
        if (!form.businessId || !form.paymentDescription || !form.procurementTypeId || !form.receiptDate || !form.recipient || !form.pum) {
            errorMessage = 'Mohon lengkapi semua field yang wajib diisi.';
            return;
        }

        if (form.valueAmount < 0 || form.paidAmount < 0 || form.taxAmount < 0) {
            errorMessage = 'Nilai nominal tidak boleh negatif.';
            return;
        }

        isSubmitting = true;
        errorMessage = '';

        try {
            let documentFile = existingDocumentFile;
            if (supportingDocumentFile && supportingDocumentFile instanceof File) {
                const uploadRes = await api.uploadFile(supportingDocumentFile);
                documentFile = { path: uploadRes.path, originalName: supportingDocumentFile.name };
            }

            if (!supportingDocumentFile) {
                documentFile = undefined;
            }

            const payload = {
                ...form,
                fundingSourceId: form.fundingSourceId || undefined,
                receiptDate: form.receiptDate ? new Date(form.receiptDate).toISOString() : undefined,
                valueAmount: Number(form.valueAmount),
                paidAmount: Number(form.paidAmount),
                taxAmount: Number(form.taxAmount),
                documentFile
            };

            if (isEditMode) {
                await api.updateGupPengajuan(editId, payload);
            } else {
                await api.createGupPengajuan(payload);
            }
            goto('/dashboard/gup/pengajuan');
        } catch (e: any) {
            console.error('Submit error:', e);
            errorMessage = e.message || 'Terjadi kesalahan saat menyimpan data.';
            isSubmitting = false;
        }
    }
</script>

<div class="max-w-4xl mx-auto space-y-6 pb-20 p-4 sm:p-6 lg:p-8">
    <div class="flex items-center gap-4 mb-6">
        <Button 
            variant="outline"
            size="icon"
            on:click={() => goto('/dashboard/gup/pengajuan')}
            aria-label="Kembali"
        >
            <svg xmlns="http://www.w3.org/2000/svg" class="h-6 w-6" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M10 19l-7-7m0 0l7-7m-7 7h18" />
            </svg>
        </Button>
        <div>
            <h1 class="text-3xl font-bold text-slate-900">{isEditMode ? 'Ubah Pengajuan' : 'Buat Pengajuan Baru'}</h1>
            <p class="text-sm text-slate-500 mt-1">{isEditMode ? 'Ubah formulir di bawah ini untuk mengedit data transaksi GUP.' : 'Isi formulir di bawah ini untuk mencatat transaksi GUP.'}</p>
        </div>
    </div>

    {#if errorMessage}
        <div class="p-4 bg-rose-50 text-rose-600 rounded-2xl border border-rose-200 font-medium flex gap-3 items-start shadow-sm">
            <svg xmlns="http://www.w3.org/2000/svg" class="h-6 w-6 flex-shrink-0" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-3L13.732 4c-.77-1.333-2.694-1.333-3.464 0L3.34 16c-.77 1.333.192 3 1.732 3z" />
            </svg>
            <p>{errorMessage}</p>
        </div>
    {/if}

    <div class="rounded-2xl bg-white shadow-sm border border-slate-200 p-6 md:p-8">
        
        <form on:submit|preventDefault={handleSubmit} class="space-y-6">
            
            <div class="grid grid-cols-1 md:grid-cols-2 gap-6">
                <!-- ID Transaksi -->
                <div class="space-y-2">
                    <label for="businessId" class="block text-sm font-bold uppercase tracking-wide text-slate-500">ID Transaksi *</label>
                    <input 
                        type="text" 
                        id="businessId" 
                        bind:value={form.businessId}
                        required
                        class="w-full rounded-xl border border-slate-300 bg-white px-4 py-3 text-sm text-slate-700 focus:outline-none focus:ring-2 focus:ring-indigo-500"
                        placeholder="Contoh: GUP-12345"
                    />
                </div>

                <!-- Tanggal Kwitansi -->
                <div class="space-y-2">
                    <label for="receiptDate" class="block text-sm font-bold uppercase tracking-wide text-slate-500">Tanggal Kwitansi *</label>
                    <Datepicker 
                        bind:value={form.receiptDate}
                        class="border-slate-300"
                    />
                </div>
            </div>

            <!-- Uraian Pembayaran -->
            <div class="space-y-2">
                <label for="paymentDescription" class="block text-sm font-bold uppercase tracking-wide text-slate-500">Uraian Pembayaran *</label>
                <textarea 
                    id="paymentDescription" 
                    bind:value={form.paymentDescription}
                    required
                    rows="3"
                    class="w-full rounded-xl border border-slate-300 bg-white px-4 py-3 text-sm text-slate-700 focus:outline-none focus:ring-2 focus:ring-indigo-500 resize-none"
                    placeholder="Masukkan uraian detail pembayaran..."
                ></textarea>
            </div>

            <div class="grid grid-cols-1 md:grid-cols-2 gap-6">
                <!-- Jenis Pengadaan -->
                <div class="space-y-2">
                    <label for="procurementTypeId" class="block text-sm font-bold uppercase tracking-wide text-slate-500">Jenis Pengadaan *</label>
                    <Select 
                        id="procurementTypeId" 
                        bind:value={form.procurementTypeId}
                        required
                        class="h-[46px] border-slate-300"
                        options={[
                            {value: '', label: 'Pilih Jenis Pengadaan'},
                            ...procurementTypes.map(t => ({value: t.id, label: `${t.name} (Akun: ${t.accountCode || '-'})`}))
                        ]}
                    />
                </div>

                <!-- Sumber Dana -->
                <div class="space-y-2">
                    <label for="fundingSourceId" class="block text-sm font-bold uppercase tracking-wide text-slate-500">Sumber Dana</label>
                    <Select 
                        id="fundingSourceId" 
                        bind:value={form.fundingSourceId}
                        class="h-[46px] border-slate-300"
                        options={[
                            {value: '', label: 'Tidak ada sumber dana (Opsional)'},
                            ...fundingSources.map(s => ({value: s.id, label: `${s.gupLabel} - ${s.monthName} (Sisa: ${formatCurrency(s.remainingBudget)})`}))
                        ]}
                    />
                </div>
            </div>

            <div class="grid grid-cols-1 md:grid-cols-2 gap-6">
                <!-- Kode Akun (Auto) -->
                <div class="space-y-2">
                    <span class="block text-sm font-bold uppercase tracking-wide text-slate-500">Kode Akun</span>
                    <div class="w-full rounded-xl border border-slate-200 bg-slate-50 px-4 py-3 text-sm text-slate-500 font-medium opacity-70 cursor-not-allowed">
                        {autoKodeAkun}
                    </div>
                </div>

                <!-- MAK (Auto) -->
                <div class="space-y-2">
                    <span class="block text-sm font-bold uppercase tracking-wide text-slate-500">MAK</span>
                    <div class="w-full rounded-xl border border-slate-200 bg-slate-50 px-4 py-3 text-sm text-slate-500 font-medium opacity-70 cursor-not-allowed">
                        {autoMak}
                    </div>
                </div>
            </div>

            <div class="grid grid-cols-1 md:grid-cols-2 gap-6">
                <!-- Penerima -->
                <div class="space-y-2">
                    <label for="recipient" class="block text-sm font-bold uppercase tracking-wide text-slate-500">Penerima *</label>
                    <input 
                        type="text" 
                        id="recipient" 
                        bind:value={form.recipient}
                        required
                        class="w-full rounded-xl border border-slate-300 bg-white px-4 py-3 text-sm text-slate-700 focus:outline-none focus:ring-2 focus:ring-indigo-500"
                        placeholder="Nama Penerima"
                    />
                </div>

                <!-- PUM -->
                <div class="space-y-2">
                    <label for="pum" class="block text-sm font-bold uppercase tracking-wide text-slate-500">PUM *</label>
                    <input 
                        type="text" 
                        id="pum" 
                        bind:value={form.pum}
                        required
                        class="w-full rounded-xl border border-slate-300 bg-white px-4 py-3 text-sm text-slate-700 focus:outline-none focus:ring-2 focus:ring-indigo-500"
                        placeholder="Nama PUM"
                    />
                </div>
            </div>

            <div class="p-5 rounded-2xl bg-slate-50 border border-slate-200">
                <h3 class="font-bold text-slate-900 mb-4 border-b border-slate-200 pb-2">Rincian Nominal</h3>
                <div class="grid grid-cols-1 md:grid-cols-3 gap-6">
                    <!-- Nilai Total -->
                    <div class="space-y-2">
                        <label for="valueAmount" class="block text-sm font-bold uppercase tracking-wide text-slate-500">Nilai (Rp) *</label>
                        <input 
                            type="text" 
                            id="valueAmount" 
                            value={formattedValueAmount}
                            on:input={(e) => handleCurrencyInput(e, 'valueAmount')}
                            required
                            class="w-full rounded-xl border border-slate-300 bg-white px-4 py-3 text-sm text-slate-900 focus:outline-none focus:ring-2 focus:ring-indigo-500 font-mono"
                            placeholder="0"
                        />
                    </div>

                    <!-- Dibayarkan -->
                    <div class="space-y-2">
                        <label for="paidAmount" class="block text-sm font-bold uppercase tracking-wide text-slate-500">Dibayarkan (Rp) *</label>
                        <input 
                            type="text" 
                            id="paidAmount" 
                            value={formattedPaidAmount}
                            on:input={(e) => handleCurrencyInput(e, 'paidAmount')}
                            required
                            class="w-full rounded-xl border border-slate-300 bg-white px-4 py-3 text-sm text-emerald-700 focus:outline-none focus:ring-2 focus:ring-emerald-500 font-mono"
                            placeholder="0"
                        />
                    </div>

                    <!-- Pajak -->
                    <div class="space-y-2">
                        <label for="taxAmount" class="block text-sm font-bold uppercase tracking-wide text-slate-500">Pajak (Rp) *</label>
                        <input 
                            type="text" 
                            id="taxAmount" 
                            value={formattedTaxAmount}
                            on:input={(e) => handleCurrencyInput(e, 'taxAmount')}
                            required
                            class="w-full rounded-xl border border-slate-300 bg-white px-4 py-3 text-sm text-rose-600 focus:outline-none focus:ring-2 focus:ring-rose-500 font-mono"
                            placeholder="0"
                        />
                    </div>
                </div>

                <div class="mt-4 pt-4 border-t border-slate-200 flex justify-between items-center">
                    <span class="font-bold text-slate-600">Selisih:</span>
                    <span class="font-bold font-mono text-lg {selisih < 0 ? 'text-rose-500' : selisih === 0 ? 'text-emerald-600' : 'text-amber-500'}">
                        {formatCurrency(selisih)}
                    </span>
                </div>
            </div>

            <!-- Dokumen Pendukung -->
            <div class="space-y-2">
                <label for="supportingDocument" class="block text-sm font-bold uppercase tracking-wide text-slate-500">Dokumen Pendukung (Opsional)</label>
                <div class="flex items-center gap-4 p-4 border border-slate-200 rounded-xl bg-slate-50">
                    {#if supportingDocumentFile}
                        <div class="flex items-center gap-3 overflow-hidden flex-1">
                            <svg xmlns="http://www.w3.org/2000/svg" class="h-6 w-6 text-blue-500 shrink-0" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M7 21h10a2 2 0 002-2V9.414a1 1 0 00-.293-.707l-5.414-5.414A1 1 0 0012.586 3H7a2 2 0 00-2 2v14a2 2 0 002 2z" />
                            </svg>
                            <span class="text-sm font-medium text-slate-700 truncate">{supportingDocumentFile.name}</span>
                        </div>
                        <button type="button" class="text-sm text-red-500 font-semibold hover:text-red-700 hover:bg-red-50 px-3 py-1.5 rounded-lg transition-colors" on:click={() => { supportingDocumentFile = null; if(fileInputRef) fileInputRef.value = ''; }}>Hapus</button>
                    {:else}
                        <div class="flex-1">
                            <p class="text-sm text-slate-500">Format: PDF atau Gambar (Maks 5MB)</p>
                        </div>
                    {/if}
                    <div class="{supportingDocumentFile ? 'hidden' : ''}">
                        <label class="cursor-pointer inline-flex items-center justify-center px-4 py-2 text-sm font-semibold text-slate-700 bg-white border border-slate-300 hover:bg-slate-50 rounded-lg shadow-sm transition-all">
                            Pilih File
                            <input 
                                type="file" 
                                id="supportingDocument"
                                bind:this={fileInputRef}
                                on:change={handleFileChange}
                                accept="application/pdf,image/*" 
                                class="hidden" 
                            />
                        </label>
                    </div>
                </div>
            </div>

            <div class="pt-4 flex justify-end border-t border-slate-200">
                <Button 
                    type="submit" 
                    disabled={isSubmitting}
                    class="w-full sm:w-auto px-8 gap-2"
                >
                    {#if isSubmitting}
                        <svg class="animate-spin -ml-1 mr-2 h-5 w-5 text-white" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24">
                            <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle>
                            <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
                        </svg>
                        Menyimpan...
                    {:else}
                        <svg xmlns="http://www.w3.org/2000/svg" class="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2.5" d="M5 13l4 4L19 7" />
                        </svg>
                        {isEditMode ? 'Perbarui Pengajuan' : 'Simpan Pengajuan'}
                    {/if}
                </Button>
            </div>

        </form>
    </div>
</div>
