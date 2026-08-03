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
                fundingSourceId: form.fundingSourceId || undefined,
                receiptDate: form.receiptDate ? new Date(form.receiptDate).toISOString() : undefined,
                valueAmount: Number(form.valueAmount),
                paidAmount: Number(form.paidAmount),
                taxAmount: Number(form.taxAmount)
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

<div class="max-w-4xl mx-auto space-y-6 pb-20 p-4 sm:p-6 lg:p-8">
    <div class="flex items-center gap-4 mb-6">
        <button 
            class="p-2 rounded-xl text-slate-500 bg-white hover:bg-slate-50 transition-colors shadow-sm border border-slate-200"
            on:click={() => goto('/dashboard/gup/pengajuan')}
        >
            <svg xmlns="http://www.w3.org/2000/svg" class="h-6 w-6" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M10 19l-7-7m0 0l7-7m-7 7h18" />
            </svg>
        </button>
        <div>
            <h1 class="text-3xl font-bold text-slate-900">Buat Pengajuan Baru</h1>
            <p class="text-sm text-slate-500 mt-1">Isi formulir di bawah ini untuk mencatat transaksi GUP.</p>
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
                    <label for="businessId" class="block text-sm font-bold uppercase tracking-wide text-slate-500">ID Transaksi <span class="text-rose-500">*</span></label>
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
                    <label for="receiptDate" class="block text-sm font-bold uppercase tracking-wide text-slate-500">Tanggal Kwitansi <span class="text-rose-500">*</span></label>
                    <input 
                        type="date" 
                        id="receiptDate" 
                        bind:value={form.receiptDate}
                        required
                        class="w-full rounded-xl border border-slate-300 bg-white px-4 py-3 text-sm text-slate-700 focus:outline-none focus:ring-2 focus:ring-indigo-500"
                    />
                </div>
            </div>

            <!-- Uraian Pembayaran -->
            <div class="space-y-2">
                <label for="paymentDescription" class="block text-sm font-bold uppercase tracking-wide text-slate-500">Uraian Pembayaran <span class="text-rose-500">*</span></label>
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
                    <label for="procurementTypeId" class="block text-sm font-bold uppercase tracking-wide text-slate-500">Jenis Pengadaan <span class="text-rose-500">*</span></label>
                    <select 
                        id="procurementTypeId" 
                        bind:value={form.procurementTypeId}
                        required
                        class="w-full rounded-xl border border-slate-300 bg-white px-4 py-3 text-sm text-slate-700 focus:outline-none focus:ring-2 focus:ring-indigo-500"
                    >
                        <option value="" disabled selected>Pilih Jenis Pengadaan</option>
                        {#each procurementTypes as type}
                            <option value={type.id}>{type.name} (MAK: {type.accountMak})</option>
                        {/each}
                    </select>
                </div>

                <!-- Sumber Dana -->
                <div class="space-y-2">
                    <label for="fundingSourceId" class="block text-sm font-bold uppercase tracking-wide text-slate-500">Sumber Dana</label>
                    <select 
                        id="fundingSourceId" 
                        bind:value={form.fundingSourceId}
                        class="w-full rounded-xl border border-slate-300 bg-white px-4 py-3 text-sm text-slate-700 focus:outline-none focus:ring-2 focus:ring-indigo-500"
                    >
                        <option value="">Tidak ada sumber dana (Opsional)</option>
                        {#each fundingSources as src}
                            <option value={src.id}>{src.gupLabel} - {src.monthName}</option>
                        {/each}
                    </select>
                </div>
            </div>

            <div class="grid grid-cols-1 md:grid-cols-2 gap-6">
                <!-- Penerima -->
                <div class="space-y-2">
                    <label for="recipient" class="block text-sm font-bold uppercase tracking-wide text-slate-500">Penerima <span class="text-rose-500">*</span></label>
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
                    <label for="pum" class="block text-sm font-bold uppercase tracking-wide text-slate-500">PUM <span class="text-rose-500">*</span></label>
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
                        <label for="valueAmount" class="block text-sm font-bold uppercase tracking-wide text-slate-500">Nilai (Rp) <span class="text-rose-500">*</span></label>
                        <input 
                            type="number" 
                            id="valueAmount" 
                            bind:value={form.valueAmount}
                            min="0"
                            required
                            class="w-full rounded-xl border border-slate-300 bg-white px-4 py-3 text-sm text-slate-900 focus:outline-none focus:ring-2 focus:ring-indigo-500 font-mono"
                        />
                    </div>

                    <!-- Dibayarkan -->
                    <div class="space-y-2">
                        <label for="paidAmount" class="block text-sm font-bold uppercase tracking-wide text-slate-500">Dibayarkan (Rp) <span class="text-rose-500">*</span></label>
                        <input 
                            type="number" 
                            id="paidAmount" 
                            bind:value={form.paidAmount}
                            min="0"
                            required
                            class="w-full rounded-xl border border-slate-300 bg-white px-4 py-3 text-sm text-emerald-700 focus:outline-none focus:ring-2 focus:ring-emerald-500 font-mono"
                        />
                    </div>

                    <!-- Pajak -->
                    <div class="space-y-2">
                        <label for="taxAmount" class="block text-sm font-bold uppercase tracking-wide text-slate-500">Pajak (Rp) <span class="text-rose-500">*</span></label>
                        <input 
                            type="number" 
                            id="taxAmount" 
                            bind:value={form.taxAmount}
                            min="0"
                            required
                            class="w-full rounded-xl border border-slate-300 bg-white px-4 py-3 text-sm text-rose-600 focus:outline-none focus:ring-2 focus:ring-rose-500 font-mono"
                        />
                    </div>
                </div>

                <div class="mt-4 pt-4 border-t border-slate-200 flex justify-between items-center">
                    <span class="font-bold text-slate-600">Selisih:</span>
                    <span class="font-bold font-mono text-lg {selisih < 0 ? 'text-rose-500' : 'text-slate-900'}">
                        {formatCurrency(selisih)}
                    </span>
                </div>
            </div>

            <div class="pt-4 flex justify-end border-t border-slate-200">
                <button 
                    type="submit" 
                    disabled={isSubmitting}
                    class="inline-flex items-center justify-center gap-2 rounded-xl bg-indigo-500 hover:bg-indigo-400 px-8 py-3 text-sm font-bold text-white border-2 border-indigo-400 transition-all hover:-translate-y-0.5 hover:shadow-lg hover:shadow-indigo-500/30 active:scale-95 disabled:opacity-50 disabled:pointer-events-none w-full sm:w-auto"
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
