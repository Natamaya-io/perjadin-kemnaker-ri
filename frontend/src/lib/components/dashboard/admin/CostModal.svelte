<script>
    import { createEventDispatcher } from 'svelte';
    import { toast } from '$lib/stores/toast';
    import { userStore } from '$lib/stores/auth';
    import Dialog from '$lib/components/ui/dialog/Dialog.svelte';
    import DialogHeader from '$lib/components/ui/dialog/DialogHeader.svelte';
    import DialogTitle from '$lib/components/ui/dialog/DialogTitle.svelte';
    import DialogFooter from '$lib/components/ui/dialog/DialogFooter.svelte';
    import Label from '$lib/components/ui/label/Label.svelte';
    import Input from '$lib/components/ui/input/Input.svelte';
    import Select from '$lib/components/ui/select/Select.svelte';
    import Button from '$lib/components/ui/button/Button.svelte';

    export let open = false;
    export let record = null;
    
    /** @type {{ dailyAllowanceDays?: number, dailyAllowanceRate?: number, hotelDays?: number, hotelRate?: number, ticketGo?: number, ticketBack?: number, localTransport?: number, regionalTransport?: number, transportMode?: string }} */
    export let editingCosts = {};

    const dispatch = createEventDispatcher();

    $: isReadOnly = $userStore.role === 'kasubag';

    // Derived Calculations
    $: totalDailyAllowance = (editingCosts.dailyAllowanceDays || 0) * (editingCosts.dailyAllowanceRate || 0);
    $: totalHotel = (editingCosts.hotelDays || 0) * (editingCosts.hotelRate || 0);
    $: totalTicket = Number(editingCosts.ticketGo || 0) + Number(editingCosts.ticketBack || 0);
    $: totalLocal = Number(editingCosts.localTransport || 0);
    $: totalRegional = Number(editingCosts.regionalTransport || 0);
    $: grandTotal = totalTicket + totalDailyAllowance + totalHotel + totalLocal + totalRegional;

    function formatCurrency(amount) {
        return new Intl.NumberFormat('id-ID', { style: 'currency', currency: 'IDR' }).format(amount);
    }

    function validateCosts() {
        if ((editingCosts.ticketGo || 0) < 0) return "Biaya Tiket Berangkat tidak boleh negatif";
        if ((editingCosts.ticketBack || 0) < 0) return "Biaya Tiket Pulang tidak boleh negatif";
        
        if ((editingCosts.dailyAllowanceDays || 0) <= 0) return "Durasi Uang Harian harus lebih dari 0 hari";
        if ((editingCosts.dailyAllowanceRate || 0) <= 0) return "Rate Uang Harian harus lebih dari Rp 0";
        
        if ((editingCosts.hotelDays || 0) < 0) return "Durasi Penginapan tidak boleh negatif";
        // Allow 0 hotel rate if stayed at non-paid accommodation, but strictly check negative
        if ((editingCosts.hotelRate || 0) < 0) return "Rate Penginapan tidak boleh negatif";

        if ((editingCosts.localTransport || 0) < 0) return "Transport Lokal tidak boleh negatif";
        if ((editingCosts.localTransport || 0) > 500000) return "Transport Lokal maksimal Rp 500.000";
        
        if ((editingCosts.regionalTransport || 0) < 0) return "Transport Daerah tidak boleh negatif";

        return null; // Valid
    }

    function handleSave() {
        if (isReadOnly) return;
        const error = validateCosts();
        if (error) {
            toast.error(error);
            return;
        }
        dispatch('save', { editingCosts, grandTotal });
    }
</script>

<Dialog bind:open={open} on:close={() => dispatch('close')}>
    <DialogHeader class="border-b border-slate-100 pb-4">
        <DialogTitle class="text-xl">Input Rincian Biaya</DialogTitle>
        <p class="text-sm text-slate-500">Lengkapi komponen biaya untuk <span class="font-semibold text-slate-800">{record?.employee?.name}</span>.</p>
    </DialogHeader>
    
    <div class="grid gap-6 py-6 max-h-[60vh] overflow-y-auto pr-2">
        <!-- Transportation -->
        <div class="grid grid-cols-1 sm:grid-cols-2 gap-4">
            <div class="space-y-2">
                <Label class="text-xs font-semibold uppercase text-slate-500 tracking-wider">Tiket Berangkat</Label>
                <div class="relative">
                    <span class="absolute left-3 top-2.5 text-slate-400 text-sm">Rp</span>
                    <Input type="number" bind:value={editingCosts.ticketGo} disabled={isReadOnly} class="pl-9 bg-slate-50 border-slate-200 focus:bg-white {isReadOnly ? 'opacity-70 cursor-not-allowed' : ''}" />
                </div>
            </div>
            <div class="space-y-2">
                <Label class="text-xs font-semibold uppercase text-slate-500 tracking-wider">Tiket Pulang</Label>
                <div class="relative">
                    <span class="absolute left-3 top-2.5 text-slate-400 text-sm">Rp</span>
                    <Input type="number" bind:value={editingCosts.ticketBack} disabled={isReadOnly} class="pl-9 bg-slate-50 border-slate-200 focus:bg-white {isReadOnly ? 'opacity-70 cursor-not-allowed' : ''}" />
                </div>
            </div>
        </div>

        <!-- Daily Allowance -->
        <div class="p-4 bg-slate-50 rounded-lg border border-slate-100 space-y-4">
            <h4 class="text-sm font-semibold text-slate-700 flex items-center gap-2">
                <svg xmlns="http://www.w3.org/2000/svg" class="h-4 w-4 text-blue-500" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 8v4l3 3m6-3a9 9 0 11-18 0 9 9 0 0118 0z" />
                </svg>
                Uang Harian
            </h4>
            <div class="grid grid-cols-1 sm:grid-cols-3 gap-4">
                <div class="space-y-2 col-span-1">
                    <Label class="text-xs text-slate-500">Durasi (Hari)</Label>
                    <Input type="number" bind:value={editingCosts.dailyAllowanceDays} disabled={isReadOnly} class="bg-white border-slate-200 {isReadOnly ? 'opacity-70 cursor-not-allowed' : ''}" />
                </div>
                <div class="space-y-2 col-span-1 sm:col-span-2">
                    <Label class="text-xs text-slate-500">Rate per Hari</Label>
                    <div class="relative">
                        <span class="absolute left-3 top-2.5 text-slate-400 text-sm">Rp</span>
                        <Input type="number" bind:value={editingCosts.dailyAllowanceRate} disabled={isReadOnly} class="pl-9 bg-white border-slate-200 {isReadOnly ? 'opacity-70 cursor-not-allowed' : ''}" />
                    </div>
                </div>
            </div>
            <div class="text-right text-sm font-mono font-medium text-slate-600 border-t border-slate-200 pt-2 mt-2">
                Subtotal: {formatCurrency(totalDailyAllowance)}
            </div>
        </div>

        <!-- Hotel -->
        <div class="p-4 bg-slate-50 rounded-lg border border-slate-100 space-y-4">
            <h4 class="text-sm font-semibold text-slate-700 flex items-center gap-2">
                <svg xmlns="http://www.w3.org/2000/svg" class="h-4 w-4 text-blue-500" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 21V5a2 2 0 00-2-2H7a2 2 0 00-2 2v16m14 0h2m-2 0h-5m-9 0H3m2 0h5M9 7h1m-1 4h1m4-4h1m-1 4h1m-5 10v-5a1 1 0 011-1h2a1 1 0 011 1v5m-4 0h4" />
                </svg>
                Penginapan
            </h4>
            <div class="grid grid-cols-1 sm:grid-cols-3 gap-4">
                <div class="space-y-2 col-span-1">
                    <Label class="text-xs text-slate-500">Durasi (Malam)</Label>
                    <Input type="number" bind:value={editingCosts.hotelDays} disabled={isReadOnly} class="bg-white border-slate-200 {isReadOnly ? 'opacity-70 cursor-not-allowed' : ''}" />
                </div>
                <div class="space-y-2 col-span-1 sm:col-span-2">
                    <Label class="text-xs text-slate-500">Rate per Malam</Label>
                    <div class="relative">
                        <span class="absolute left-3 top-2.5 text-slate-400 text-sm">Rp</span>
                        <Input type="number" bind:value={editingCosts.hotelRate} disabled={isReadOnly} class="pl-9 bg-white border-slate-200 {isReadOnly ? 'opacity-70 cursor-not-allowed' : ''}" />
                    </div>
                </div>
            </div>
            <div class="text-right text-sm font-mono font-medium text-slate-600 border-t border-slate-200 pt-2 mt-2">
                Subtotal: {formatCurrency(totalHotel)}
            </div>
        </div>

        <!-- Local Transport -->
        <div class="grid grid-cols-1 sm:grid-cols-2 gap-4 pt-2">
            <div class="space-y-2">
                <Label class="text-xs font-semibold uppercase text-slate-500 tracking-wider">Transport Lokal</Label>
                <div class="relative">
                    <span class="absolute left-3 top-2.5 text-slate-400 text-sm">Rp</span>
                    <Input type="number" bind:value={editingCosts.localTransport} disabled={isReadOnly} class="pl-9 bg-slate-50 border-slate-200 focus:bg-white {isReadOnly ? 'opacity-70 cursor-not-allowed' : ''}" />
                </div>
                <p class="text-[10px] text-slate-400">Maks. Rp 500.000</p>
            </div>
            <div class="space-y-2">
                <Label class="text-xs font-semibold uppercase text-slate-500 tracking-wider">Transport Daerah</Label>
                <div class="relative">
                    <span class="absolute left-3 top-2.5 text-slate-400 text-sm">Rp</span>
                    <Input type="number" bind:value={editingCosts.regionalTransport} disabled={isReadOnly} class="pl-9 bg-slate-50 border-slate-200 focus:bg-white {isReadOnly ? 'opacity-70 cursor-not-allowed' : ''}" />
                </div>
            </div>
        </div>

        <div class="space-y-2 pt-2">
            <Label class="text-xs font-semibold uppercase text-slate-500 tracking-wider">Mode Transportasi</Label>
            <Select bind:value={editingCosts.transportMode} disabled={isReadOnly} class="bg-slate-50 border-slate-200 {isReadOnly ? 'opacity-70 cursor-not-allowed' : ''}">
                <option value="Pesawat">Pesawat Udara</option>
                <option value="Kendaraan Umum">Kendaraan Umum / Kereta</option>
                <option value="Kendaraan Dinas">Kendaraan Dinas</option>
            </Select>
        </div>

        <!-- Document Upload / Review -->
        <div class="space-y-2 pt-2 border-t border-slate-100 mt-4">
            <Label class="text-xs font-semibold uppercase text-slate-500 tracking-wider">Dokumen, Kwitansi, & Tagihan</Label>
            
            {#if $userStore.role === 'protokol' && !isReadOnly}
                <div class="border-2 border-dashed border-slate-200 rounded-lg p-6 text-center hover:bg-slate-50 transition-colors">
                    <svg xmlns="http://www.w3.org/2000/svg" class="mx-auto h-8 w-8 text-slate-400 mb-2" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M7 16a4 4 0 01-.88-7.903A5 5 0 1115.9 6L16 6a5 5 0 011 9.9M15 13l-3-3m0 0l-3 3m3-3v12" />
                    </svg>
                    <div class="text-sm text-slate-600">
                        <label for="file-upload" class="relative cursor-pointer bg-white rounded-md font-medium text-blue-600 hover:text-blue-500 focus-within:outline-none">
                            <span>Upload File</span>
                            <input id="file-upload" name="file-upload" type="file" class="sr-only" multiple accept=".pdf,.jpg,.jpeg,.png">
                        </label>
                        <p class="pl-1 text-slate-500 text-xs mt-1">PDF, PNG, JPG hingga 5MB</p>
                    </div>
                </div>
            {:else}
                <!-- Keuangan Review View -->
                <div class="border border-slate-200 rounded-lg p-4 bg-slate-50 flex items-center justify-between">
                    <div class="flex items-center gap-3">
                        <div class="h-10 w-10 bg-white border border-slate-200 rounded flex items-center justify-center text-blue-500">
                            <svg xmlns="http://www.w3.org/2000/svg" class="h-6 w-6" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z" />
                            </svg>
                        </div>
                        <div>
                            <p class="text-sm font-medium text-slate-800">Berkas Tagihan Protokol</p>
                            <p class="text-xs text-slate-500">3 File dilampirkan</p>
                        </div>
                    </div>
                    <Button variant="outline" size="sm" class="text-xs bg-white text-blue-600 hover:bg-blue-50 border-slate-200" on:click={() => toast.info('Fitur preview dokumen sedang dalam pengembangan.')}>
                        Review Dokumen
                    </Button>
                </div>
            {/if}
        </div>
    </div>

    <DialogFooter>
        <div class="w-full flex flex-col sm:flex-row items-center justify-between gap-4 border-t border-slate-100 pt-4">
            <div class="w-full sm:w-auto flex justify-between sm:block text-left">
                <span class="block text-xs text-slate-500 self-center sm:self-auto">Total Estimasi</span>
                <span class="text-lg font-bold text-blue-600">{formatCurrency(grandTotal)}</span>
            </div>
            <div class="flex gap-2 w-full sm:w-auto">
                <Button variant="outline" class="flex-1 sm:flex-none border-slate-200 text-slate-600" on:click={() => dispatch('close')}>Tutup</Button>
                {#if !isReadOnly}
                    <Button class="flex-1 sm:flex-none bg-blue-600 hover:bg-blue-700 text-white shadow-lg shadow-blue-500/20" on:click={handleSave}>Simpan</Button>
                {/if}
            </div>
        </div>
    </DialogFooter>
</Dialog>
