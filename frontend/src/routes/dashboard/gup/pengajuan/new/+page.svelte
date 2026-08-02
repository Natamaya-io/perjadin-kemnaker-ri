<script lang="ts">
    import { goto } from '$app/navigation';
    import { api } from '$lib/shared/api';
    import { formatCurrency } from '$lib/shared/utils/utils';

    export let data: any;

    let masterData = data.masterData || { fundingSources: [], procurementTypes: [] };
    let fundingSources = masterData.fundingSources || [];
    let procurementTypes = masterData.procurementTypes || [];

    let form = {
        businessId: 'GUP-' + Date.now(), // Auto-generate a default, user can edit
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

    let isSubmitting = false;
    let errorMessage = '';

    $: selisih = form.valueAmount - form.paidAmount - form.taxAmount;

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
            const payload = {
                ...form,
                fundingSourceId: form.fundingSourceId || undefined
            };

            await api.createGupPengajuan(payload);
            goto('/dashboard/gup/pengajuan');
        } catch (e: any) {
            console.error('Submit error:', e);
            errorMessage = e.message || 'Terjadi kesalahan saat menyimpan data.';
            isSubmitting = false;
        }
    }
</script>

<div class="max-w-4xl mx-auto space-y-6 pb-20">
    <div class="flex items-center gap-4 mb-6">
        <button 
            class="p-2 rounded-xl text-slate-500 bg-slate-100 hover:bg-slate-200 transition-colors border border-white"
            style="box-shadow: inset 1px 1px 3px rgba(0,0,0,0.05), inset -1px -1px 3px rgba(255,255,255,0.9);"
            on:click={() => goto('/dashboard/gup/pengajuan')}
        >
            <svg xmlns="http://www.w3.org/2000/svg" class="h-6 w-6" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M10 19l-7-7m0 0l7-7m-7 7h18" />
            </svg>
        </button>
        <div>
            <h1 class="text-3xl font-extrabold text-slate-800 tracking-tight">Buat Pengajuan Baru</h1>
            <p class="text-sm text-slate-500 mt-1 font-medium">Isi formulir di bawah ini untuk mencatat transaksi GUP.</p>
        </div>
    </div>

    {#if errorMessage}
        <div class="p-4 bg-rose-50 text-rose-600 rounded-2xl border border-rose-200 font-medium flex gap-3 items-start">
            <svg xmlns="http://www.w3.org/2000/svg" class="h-6 w-6 flex-shrink-0" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-3L13.732 4c-.77-1.333-2.694-1.333-3.464 0L3.34 16c-.77 1.333.192 3 1.732 3z" />
            </svg>
            <p>{errorMessage}</p>
        </div>
    {/if}

    <div class="rounded-3xl bg-slate-100 border-2 border-white p-6 md:p-8"
         style="box-shadow: inset 4px 4px 10px rgba(0,0,0,0.05), inset -4px -4px 10px rgba(255,255,255,0.8);">
        
        <form on:submit|preventDefault={handleSubmit} class="space-y-6">
            
            <div class="grid grid-cols-1 md:grid-cols-2 gap-6">
                <!-- ID Transaksi -->
                <div class="space-y-2">
                    <label for="businessId" class="block text-sm font-bold text-slate-700">ID Transaksi <span class="text-rose-500">*</span></label>
                    <input 
                        type="text" 
                        id="businessId" 
                        bind:value={form.businessId}
                        required
                        class="w-full px-4 py-3 rounded-2xl bg-slate-50 border-transparent text-slate-800 focus:ring-2 focus:ring-indigo-400 focus:outline-none transition-all"
                        style="box-shadow: inset 2px 2px 5px rgba(0,0,0,0.05), inset -2px -2px 5px rgba(255,255,255,1);"
                        placeholder="Contoh: GUP-12345"
                    />
                </div>

                <!-- Tanggal Kwitansi -->
                <div class="space-y-2">
                    <label for="receiptDate" class="block text-sm font-bold text-slate-700">Tanggal Kwitansi <span class="text-rose-500">*</span></label>
                    <input 
                        type="date" 
                        id="receiptDate" 
                        bind:value={form.receiptDate}
                        required
                        class="w-full px-4 py-3 rounded-2xl bg-slate-50 border-transparent text-slate-800 focus:ring-2 focus:ring-indigo-400 focus:outline-none transition-all"
                        style="box-shadow: inset 2px 2px 5px rgba(0,0,0,0.05), inset -2px -2px 5px rgba(255,255,255,1);"
                    />
                </div>
            </div>

            <!-- Uraian Pembayaran -->
            <div class="space-y-2">
                <label for="paymentDescription" class="block text-sm font-bold text-slate-700">Uraian Pembayaran <span class="text-rose-500">*</span></label>
                <textarea 
                    id="paymentDescription" 
                    bind:value={form.paymentDescription}
                    required
                    rows="3"
                    class="w-full px-4 py-3 rounded-2xl bg-slate-50 border-transparent text-slate-800 focus:ring-2 focus:ring-indigo-400 focus:outline-none transition-all resize-none"
                    style="box-shadow: inset 2px 2px 5px rgba(0,0,0,0.05), inset -2px -2px 5px rgba(255,255,255,1);"
                    placeholder="Masukkan uraian detail pembayaran..."
                ></textarea>
            </div>

            <div class="grid grid-cols-1 md:grid-cols-2 gap-6">
                <!-- Jenis Pengadaan -->
                <div class="space-y-2">
                    <label for="procurementTypeId" class="block text-sm font-bold text-slate-700">Jenis Pengadaan <span class="text-rose-500">*</span></label>
                    <div class="relative">
                        <select 
                            id="procurementTypeId" 
                            bind:value={form.procurementTypeId}
                            required
                            class="w-full px-4 py-3 rounded-2xl bg-slate-50 border-transparent text-slate-800 focus:ring-2 focus:ring-indigo-400 focus:outline-none transition-all appearance-none"
                            style="box-shadow: inset 2px 2px 5px rgba(0,0,0,0.05), inset -2px -2px 5px rgba(255,255,255,1);"
                        >
                            <option value="" disabled selected>Pilih Jenis Pengadaan</option>
                            {#each procurementTypes as type}
                                <option value={type.id}>{type.name} (MAK: {type.accountMak})</option>
                            {/each}
                        </select>
                        <div class="pointer-events-none absolute inset-y-0 right-0 flex items-center px-4 text-slate-500">
                            <svg class="h-4 w-4" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 9l-7 7-7-7"></path></svg>
                        </div>
                    </div>
                </div>

                <!-- Sumber Dana -->
                <div class="space-y-2">
                    <label for="fundingSourceId" class="block text-sm font-bold text-slate-700">Sumber Dana</label>
                    <div class="relative">
                        <select 
                            id="fundingSourceId" 
                            bind:value={form.fundingSourceId}
                            class="w-full px-4 py-3 rounded-2xl bg-slate-50 border-transparent text-slate-800 focus:ring-2 focus:ring-indigo-400 focus:outline-none transition-all appearance-none"
                            style="box-shadow: inset 2px 2px 5px rgba(0,0,0,0.05), inset -2px -2px 5px rgba(255,255,255,1);"
                        >
                            <option value="">Tidak ada sumber dana (Opsional)</option>
                            {#each fundingSources as src}
                                <option value={src.id}>{src.gupLabel} - {src.monthName}</option>
                            {/each}
                        </select>
                        <div class="pointer-events-none absolute inset-y-0 right-0 flex items-center px-4 text-slate-500">
                            <svg class="h-4 w-4" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 9l-7 7-7-7"></path></svg>
                        </div>
                    </div>
                </div>
            </div>

            <div class="grid grid-cols-1 md:grid-cols-2 gap-6">
                <!-- Penerima -->
                <div class="space-y-2">
                    <label for="recipient" class="block text-sm font-bold text-slate-700">Penerima <span class="text-rose-500">*</span></label>
                    <input 
                        type="text" 
                        id="recipient" 
                        bind:value={form.recipient}
                        required
                        class="w-full px-4 py-3 rounded-2xl bg-slate-50 border-transparent text-slate-800 focus:ring-2 focus:ring-indigo-400 focus:outline-none transition-all"
                        style="box-shadow: inset 2px 2px 5px rgba(0,0,0,0.05), inset -2px -2px 5px rgba(255,255,255,1);"
                        placeholder="Nama Penerima"
                    />
                </div>

                <!-- PUM -->
                <div class="space-y-2">
                    <label for="pum" class="block text-sm font-bold text-slate-700">PUM <span class="text-rose-500">*</span></label>
                    <input 
                        type="text" 
                        id="pum" 
                        bind:value={form.pum}
                        required
                        class="w-full px-4 py-3 rounded-2xl bg-slate-50 border-transparent text-slate-800 focus:ring-2 focus:ring-indigo-400 focus:outline-none transition-all"
                        style="box-shadow: inset 2px 2px 5px rgba(0,0,0,0.05), inset -2px -2px 5px rgba(255,255,255,1);"
                        placeholder="Nama PUM"
                    />
                </div>
            </div>

            <div class="p-5 rounded-2xl bg-slate-50 border border-slate-100" style="box-shadow: inset 2px 2px 6px rgba(0,0,0,0.04), inset -2px -2px 6px rgba(255,255,255,1);">
                <h3 class="font-bold text-slate-700 mb-4 border-b border-slate-200 pb-2">Rincian Nominal</h3>
                <div class="grid grid-cols-1 md:grid-cols-3 gap-6">
                    <!-- Nilai Total -->
                    <div class="space-y-2">
                        <label for="valueAmount" class="block text-sm font-bold text-slate-700">Nilai (Rp) <span class="text-rose-500">*</span></label>
                        <input 
                            type="number" 
                            id="valueAmount" 
                            bind:value={form.valueAmount}
                            min="0"
                            required
                            class="w-full px-4 py-3 rounded-2xl bg-white border-transparent text-slate-800 focus:ring-2 focus:ring-indigo-400 focus:outline-none font-mono"
                            style="box-shadow: inset 2px 2px 5px rgba(0,0,0,0.05), inset -2px -2px 5px rgba(255,255,255,1);"
                        />
                    </div>

                    <!-- Dibayarkan -->
                    <div class="space-y-2">
                        <label for="paidAmount" class="block text-sm font-bold text-slate-700">Dibayarkan (Rp) <span class="text-rose-500">*</span></label>
                        <input 
                            type="number" 
                            id="paidAmount" 
                            bind:value={form.paidAmount}
                            min="0"
                            required
                            class="w-full px-4 py-3 rounded-2xl bg-white border-transparent text-emerald-700 focus:ring-2 focus:ring-emerald-400 focus:outline-none font-mono"
                            style="box-shadow: inset 2px 2px 5px rgba(0,0,0,0.05), inset -2px -2px 5px rgba(255,255,255,1);"
                        />
                    </div>

                    <!-- Pajak -->
                    <div class="space-y-2">
                        <label for="taxAmount" class="block text-sm font-bold text-slate-700">Pajak (Rp) <span class="text-rose-500">*</span></label>
                        <input 
                            type="number" 
                            id="taxAmount" 
                            bind:value={form.taxAmount}
                            min="0"
                            required
                            class="w-full px-4 py-3 rounded-2xl bg-white border-transparent text-rose-600 focus:ring-2 focus:ring-rose-400 focus:outline-none font-mono"
                            style="box-shadow: inset 2px 2px 5px rgba(0,0,0,0.05), inset -2px -2px 5px rgba(255,255,255,1);"
                        />
                    </div>
                </div>

                <div class="mt-4 pt-4 border-t border-slate-200 flex justify-between items-center">
                    <span class="font-bold text-slate-600">Selisih:</span>
                    <span class="font-bold font-mono text-lg {selisih < 0 ? 'text-rose-500' : 'text-slate-800'}">
                        {formatCurrency(selisih)}
                    </span>
                </div>
            </div>

            <div class="pt-4 flex justify-end">
                <button 
                    type="submit" 
                    disabled={isSubmitting}
                    class="w-full sm:w-auto px-8 py-3 rounded-2xl font-bold text-white bg-indigo-500 border-2 border-indigo-400 disabled:opacity-70 disabled:cursor-not-allowed flex justify-center items-center gap-2"
                    style="box-shadow: inset 2px 2px 5px rgba(255,255,255,0.4), inset -3px -3px 7px rgba(0,0,0,0.15);"
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
                        Simpan Pengajuan
                    {/if}
                </button>
            </div>

        </form>
    </div>
</div>
