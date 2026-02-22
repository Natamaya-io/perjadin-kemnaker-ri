<script>
    import { createEventDispatcher } from 'svelte';
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
    export let editingCosts = {};

    const dispatch = createEventDispatcher();

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

    function handleSave() {
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
                    <Input type="number" bind:value={editingCosts.ticketGo} class="pl-9 bg-slate-50 border-slate-200 focus:bg-white" />
                </div>
            </div>
            <div class="space-y-2">
                <Label class="text-xs font-semibold uppercase text-slate-500 tracking-wider">Tiket Pulang</Label>
                <div class="relative">
                    <span class="absolute left-3 top-2.5 text-slate-400 text-sm">Rp</span>
                    <Input type="number" bind:value={editingCosts.ticketBack} class="pl-9 bg-slate-50 border-slate-200 focus:bg-white" />
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
                    <Input type="number" bind:value={editingCosts.dailyAllowanceDays} class="bg-white border-slate-200" />
                </div>
                <div class="space-y-2 col-span-1 sm:col-span-2">
                    <Label class="text-xs text-slate-500">Rate per Hari</Label>
                    <div class="relative">
                        <span class="absolute left-3 top-2.5 text-slate-400 text-sm">Rp</span>
                        <Input type="number" bind:value={editingCosts.dailyAllowanceRate} class="pl-9 bg-white border-slate-200" />
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
                    <Input type="number" bind:value={editingCosts.hotelDays} class="bg-white border-slate-200" />
                </div>
                <div class="space-y-2 col-span-1 sm:col-span-2">
                    <Label class="text-xs text-slate-500">Rate per Malam</Label>
                    <div class="relative">
                        <span class="absolute left-3 top-2.5 text-slate-400 text-sm">Rp</span>
                        <Input type="number" bind:value={editingCosts.hotelRate} class="pl-9 bg-white border-slate-200" />
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
                    <Input type="number" bind:value={editingCosts.localTransport} class="pl-9 bg-slate-50 border-slate-200 focus:bg-white" />
                </div>
                <p class="text-[10px] text-slate-400">Maks. Rp 500.000</p>
            </div>
            <div class="space-y-2">
                <Label class="text-xs font-semibold uppercase text-slate-500 tracking-wider">Transport Daerah</Label>
                <div class="relative">
                    <span class="absolute left-3 top-2.5 text-slate-400 text-sm">Rp</span>
                    <Input type="number" bind:value={editingCosts.regionalTransport} class="pl-9 bg-slate-50 border-slate-200 focus:bg-white" />
                </div>
            </div>
        </div>

        <div class="space-y-2 pt-2">
            <Label class="text-xs font-semibold uppercase text-slate-500 tracking-wider">Mode Transportasi</Label>
            <Select bind:value={editingCosts.transportMode} class="bg-slate-50 border-slate-200">
                <option value="Pesawat">Pesawat Udara</option>
                <option value="Kendaraan Umum">Kendaraan Umum / Kereta</option>
                <option value="Kendaraan Dinas">Kendaraan Dinas</option>
            </Select>
        </div>
    </div>

    <DialogFooter>
        <div class="w-full flex flex-col sm:flex-row items-center justify-between gap-4 border-t border-slate-100 pt-4">
            <div class="w-full sm:w-auto flex justify-between sm:block text-left">
                <span class="block text-xs text-slate-500 self-center sm:self-auto">Total Estimasi</span>
                <span class="text-lg font-bold text-blue-600">{formatCurrency(grandTotal)}</span>
            </div>
            <div class="flex gap-2 w-full sm:w-auto">
                <Button variant="outline" class="flex-1 sm:flex-none border-slate-200 text-slate-600" on:click={() => dispatch('close')}>Batal</Button>
                <Button class="flex-1 sm:flex-none bg-blue-600 hover:bg-blue-700 text-white shadow-lg shadow-blue-500/20" on:click={handleSave}>Simpan</Button>
            </div>
        </div>
    </DialogFooter>
</Dialog>
