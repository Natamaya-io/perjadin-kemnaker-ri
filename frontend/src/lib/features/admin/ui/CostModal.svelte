<script>
    import { createEventDispatcher } from 'svelte';
    import { toast } from '$lib/shared/stores/toast';
    import { userStore } from '$lib/features/auth/store';
    import { provincesStore } from '$lib/shared/stores/master-data';
    import Dialog from '$lib/shared/ui/dialog/Dialog.svelte';
    import DialogHeader from '$lib/shared/ui/dialog/DialogHeader.svelte';
    import DialogTitle from '$lib/shared/ui/dialog/DialogTitle.svelte';
    import DialogFooter from '$lib/shared/ui/dialog/DialogFooter.svelte';
    import Label from '$lib/shared/ui/label/Label.svelte';
    import Input from '$lib/shared/ui/input/Input.svelte';
    import Select from '$lib/shared/ui/select/Select.svelte';
    import Button from '$lib/shared/ui/button/Button.svelte';
    import DocumentViewer from '$lib/shared/ui/document-viewer/DocumentViewer.svelte';

    export let open = false;
    export let record = null;
    
    export let editingCosts = {};
    export let isReadOnlyOverride = false;

    const dispatch = createEventDispatcher();

    // The modal is read-only if the user is a kasubag or if the record is already Approved (meaning it's just for review/printing) or if overridden
    // Allow 'keuangan' and 'super_admin' to edit even if Approved
    $: isReadOnly = isReadOnlyOverride || $userStore.role === 'kasubag' || (record?.status === 'Approved' && $userStore.role !== 'keuangan' && $userStore.role !== 'super_admin');

    // Derived Calculations for SBM Uang Harian
	$: costBreakdown = (record?.locations && record.locations.length > 0 
    	? record.locations 
 	   : [{ startDate: record?.startDate, endDate: record?.endDate, province: record?.province }]
		).map(loc => {
    		const provData = $provincesStore.find(p => p.name === loc.province);
    		const rate = provData ? provData.luarKota : 0;
    		const start = new Date(loc.startDate);
    		const end = new Date(loc.endDate);
    		let locDays = 0;
    		if (!isNaN(start.getTime()) && !isNaN(end.getTime())) {
        	const diffTime = end.getTime() - start.getTime();
        	const diffDays = Math.ceil(diffTime / (1000 * 60 * 60 * 24)) + 1;
        	locDays = diffDays > 0 ? diffDays : 0;
    	}
    		return { days: locDays, rate: rate, province: loc.province };
		});

	$: days = costBreakdown.reduce((sum, item) => sum + item.days, 0);
	$: totalDailyAllowance = costBreakdown.reduce((sum, item) => sum + (item.rate * item.days), 0);
	$: sbmRateAvg = days > 0 ? totalDailyAllowance / days : 0;
    // Default arrays if undefined
    $: if (open && !editingCosts.additionalCosts) {
        editingCosts.additionalCosts = [];
    }

    $: totalHotel = (editingCosts.hotelDays || 0) * (editingCosts.hotelRate || 0);
    $: totalTicket = Number(editingCosts.ticketGo || 0) + Number(editingCosts.ticketBack || 0);
    $: totalTransportAmount = Number(editingCosts.transportAmount || 0);
    $: totalAdditional = (editingCosts.additionalCosts || []).reduce((sum, cost) => sum + (cost.amount || 0), 0);
    
    $: grandTotal = totalTicket + totalDailyAllowance + totalHotel + totalTransportAmount + totalAdditional;

    // Preview State
    let previewFile = null;
    let showPreview = false;

    function openPreview(file) {
        previewFile = file;
        showPreview = true;
    }

    function closePreview() {
        showPreview = false;
        previewFile = null;
    }

    function formatFileSize(bytes) {
        if (!bytes) return '0 Bytes';
        const k = 1024;
        const sizes = ['Bytes', 'KB', 'MB', 'GB'];
        const i = Math.floor(Math.log(bytes) / Math.log(k));
        return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i];
    }

    function formatCurrency(amount) {
        return new Intl.NumberFormat('id-ID', { style: 'currency', currency: 'IDR' }).format(amount);
    }

    function formatInputNumber(value) {
        if (value === undefined || value === null || value === '') return '';
        const numStr = value.toString().replace(/[^0-9]/g, '');
        if (!numStr) return '';
        return parseInt(numStr, 10).toLocaleString('id-ID');
    }

    function updateCost(field, event) {
        const raw = event.target.value.replace(/[^0-9]/g, '');
        const num = parseInt(raw, 10);
        editingCosts[field] = isNaN(num) ? undefined : num;
        editingCosts = editingCosts;
    }

    function handleSpecificFileSelect(e, field) {
        const file = e.target.files[0];
        if (!file) return;
        if (file.size > 5 * 1024 * 1024) {
            toast.error(`Ukuran file melebihi 5MB.`);
            e.target.value = '';
            return;
        }
        const reader = new FileReader();
        reader.onload = (ev) => {
            editingCosts[field] = {
                name: file.name,
                size: file.size,
                type: file.type,
                data: ev.target.result
            };
        };
        reader.readAsDataURL(file);
        e.target.value = '';
    }

    function handleBoardingPassFileSelect(e) {
        const file = e.target.files[0];
        if (!file) return;
        if (file.size > 5 * 1024 * 1024) {
            toast.error(`Ukuran file melebihi 5MB.`);
            e.target.value = '';
            return;
        }

        if (!editingCosts.boardingPassFiles) {
            editingCosts.boardingPassFiles = [];
            if (editingCosts.boardingPassFile) {
                editingCosts.boardingPassFiles.push(editingCosts.boardingPassFile);
                delete editingCosts.boardingPassFile;
            }
        }

        const reader = new FileReader();
        reader.onload = (ev) => {
            editingCosts.boardingPassFiles = [...editingCosts.boardingPassFiles, {
                name: file.name,
                size: file.size,
                type: file.type,
                data: ev.target.result
            }];
        };
        reader.readAsDataURL(file);
        e.target.value = '';
    }

    function removeBoardingPassFile(index) {
        if (editingCosts.boardingPassFiles) {
            editingCosts.boardingPassFiles = editingCosts.boardingPassFiles.filter((_, i) => i !== index);
        }
    }

    function removeSpecificFile(field) {
        editingCosts[field] = null;
        editingCosts = editingCosts;
    }

    function addAdditionalCost() {
        if (!editingCosts.additionalCosts) editingCosts.additionalCosts = [];
        editingCosts.additionalCosts = [...editingCosts.additionalCosts, { name: '', amount: undefined, file: null }];
        editingCosts = editingCosts;
    }

    function removeAdditionalCost(index) {
        editingCosts.additionalCosts = editingCosts.additionalCosts.filter((_, i) => i !== index);
        editingCosts = editingCosts;
    }

    function updateAdditionalCostAmount(index, event) {
        const raw = event.target.value.replace(/[^0-9]/g, '');
        const num = parseInt(raw, 10);
        editingCosts.additionalCosts[index].amount = isNaN(num) ? undefined : num;
        editingCosts = editingCosts;
    }

    function handleAdditionalFileSelect(e, index) {
        const file = e.target.files[0];
        if (!file) return;
        if (file.size > 5 * 1024 * 1024) {
            toast.error(`Ukuran file melebihi 5MB.`);
            e.target.value = '';
            return;
        }
        const reader = new FileReader();
        reader.onload = (ev) => {
            editingCosts.additionalCosts[index].file = {
                name: file.name,
                size: file.size,
                type: file.type,
                data: ev.target.result
            };
            editingCosts = editingCosts;
        };
        reader.readAsDataURL(file);
        e.target.value = '';
    }

    function validateCosts() {
        if ((editingCosts.ticketGo || 0) < 0) return "Biaya Tiket Berangkat tidak boleh negatif";
        if ((editingCosts.ticketBack || 0) < 0) return "Biaya Tiket Pulang tidak boleh negatif";
        if ((editingCosts.hotelDays || 0) < 0) return "Durasi Penginapan tidak boleh negatif";
        if ((editingCosts.hotelRate || 0) < 0) return "Rate Penginapan tidak boleh negatif";
        if ((editingCosts.transportAmount || 0) < 0) return "Biaya Transportasi tidak boleh negatif";
        for (const cost of (editingCosts.additionalCosts || [])) {
            if ((cost.amount || 0) < 0) return "Biaya Tambahan tidak boleh negatif";
        }
        return null;
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

<Dialog bind:open={open} class="w-[calc(100vw-2rem)] md:w-full max-w-[95vw] md:max-w-2xl lg:max-w-3xl overflow-hidden flex flex-col p-0 h-[calc(100vh-2rem)] max-h-[90vh] md:max-h-[85vh] mx-auto my-auto md:mt-10 rounded-2xl shadow-2xl border-0" on:close={() => dispatch('close')}>
    <DialogHeader class="border-b border-slate-100 p-4 md:p-6 shrink-0 bg-white/95 backdrop-blur z-10 sticky top-0">
        <DialogTitle class="text-lg md:text-xl font-bold text-slate-800">
            {$userStore.role === 'protokol' 
                ? (record?.status === 'Draft' ? 'Input Rincian Biaya' : 'Edit Rincian Biaya') 
                : 'Review Rincian Biaya'}
        </DialogTitle>
        <p class="text-xs md:text-sm text-slate-500 mt-1">Rincian komponen biaya untuk pegawai <span class="font-semibold text-blue-600 bg-blue-50 px-1.5 py-0.5 rounded">{record?.employee?.name}</span>.</p>
    </DialogHeader>
    
    <div class="grid gap-4 md:gap-6 p-4 md:p-6 overflow-y-auto overflow-x-hidden bg-slate-50/50 custom-scrollbar flex-1 min-h-0 relative">
        
        <!-- Mode Transportasi -->
        <div class="p-3 md:p-4 bg-white rounded-xl border border-slate-200 shadow-sm w-full space-y-1.5">
            <Label class="text-[10px] md:text-xs font-semibold uppercase text-slate-500 tracking-wider">Mode Transportasi</Label>
            <Select bind:value={editingCosts.transportMode} disabled={isReadOnly} class="bg-slate-50 border-slate-200 h-9 md:h-10 text-sm {isReadOnly ? 'opacity-70 cursor-not-allowed pointer-events-none' : ''}">
                <option value="Pesawat">Pesawat Udara</option>
                <option value="Kendaraan Umum">Kendaraan Umum / Kereta</option>
                <option value="Kendaraan Dinas">Kendaraan Dinas</option>
            </Select>
        </div>

        <!-- Uang Harian SBM -->
        <div class="p-3 md:p-4 bg-blue-50/50 rounded-xl border border-blue-100 space-y-2.5 md:space-y-3 w-full">
            <div class="flex flex-wrap justify-between items-center gap-2">
                <h4 class="text-xs md:text-sm font-semibold text-blue-800 flex items-center gap-1.5">
                    <svg xmlns="http://www.w3.org/2000/svg" class="h-3.5 w-3.5 md:h-4 md:w-4 text-blue-500" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 8v4l3 3m6-3a9 9 0 11-18 0 9 9 0 0118 0z" />
                    </svg>
                    Uang Harian (SBM)
                </h4>
                <span class="text-base md:text-lg font-bold text-blue-700">{formatCurrency(totalDailyAllowance)}</span>
            </div>
            
            <div class="space-y-2">
                {#each costBreakdown as item}
                    <div class="flex flex-wrap md:flex-nowrap items-center gap-2 md:gap-3 text-xs md:text-sm text-slate-600 bg-white p-2 md:p-3 rounded-lg border border-blue-50/50 shadow-sm w-full">
                        <div class="flex-none pr-2 border-r border-slate-100 min-w-[80px]">
                             <span class="block text-[9px] md:text-[10px] uppercase tracking-wider text-slate-400 font-semibold mb-0.5">Provinsi</span>
                             <span class="font-medium whitespace-nowrap text-blue-600">{item.province}</span>
                        </div>
                        <div class="flex-none pr-2 border-r border-slate-100">
                            <span class="block text-[9px] md:text-[10px] uppercase tracking-wider text-slate-400 font-semibold mb-0.5">Durasi</span>
                            <span class="font-medium whitespace-nowrap">{item.days} Hari</span>
                        </div>
                        <div class="flex-1 min-w-0 overflow-hidden">
                            <span class="block text-[9px] md:text-[10px] uppercase tracking-wider text-slate-400 font-semibold mb-0.5">Rate SBM</span>
                            <span class="font-medium truncate block w-full">{formatCurrency(item.rate)} <span class="text-[10px] md:text-xs text-slate-400 font-normal">/ hari</span></span>
                        </div>
                        <div class="flex-none pl-2 border-l border-slate-100 text-right">
                             <span class="block text-[9px] md:text-[10px] uppercase tracking-wider text-slate-400 font-semibold mb-0.5">Total</span>
                             <span class="font-bold text-slate-800">{formatCurrency(item.rate * item.days)}</span>
                        </div>
                    </div>
                {/each}
            </div>
        </div>

        <!-- Tiket & Boarding Pass -->
        <div class="p-3 md:p-4 bg-white rounded-xl border border-slate-200 shadow-sm space-y-4 w-full">
            <h4 class="text-xs md:text-sm font-semibold text-slate-700 flex items-center gap-1.5">
                <svg xmlns="http://www.w3.org/2000/svg" class="h-3.5 w-3.5 md:h-4 md:w-4 text-blue-500" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M3.055 11H5a2 2 0 012 2v1a2 2 0 002 2 2 2 0 012 2v2.945M8 3.935V5.5A2.5 2.5 0 0010.5 8h.5a2 2 0 012 2 2 2 0 104 0 2 2 0 012-2h1.064M15 20.488V18a2 2 0 012-2h3.064M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
                </svg>
                Tiket & Boarding Pass
            </h4>
            
            <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
                <!-- Tiket Berangkat -->
                <div class="space-y-1.5 p-3 border border-slate-100 bg-slate-50 rounded-lg">
                    <div class="flex justify-between items-center">
                        <Label class="text-[10px] md:text-xs font-semibold uppercase text-slate-500 tracking-wider">Tiket Berangkat</Label>
                        {#if !editingCosts.ticketGoFile && !isReadOnly}
                            <label class="cursor-pointer text-[9px] md:text-[10px] text-blue-700 bg-blue-50 hover:bg-blue-100 px-2 py-1 md:px-2.5 md:py-1 rounded-md border border-blue-200 font-medium transition-colors flex items-center gap-1">
                                <svg xmlns="http://www.w3.org/2000/svg" class="h-3 w-3" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 16v1a3 3 0 003 3h10a3 3 0 003-3v-1m-4-8l-4-4m0 0L8 8m4-4v12" /></svg>
                                Upload Kwitansi
                                <input type="file" class="hidden" accept=".pdf,.jpg,.jpeg,.png" on:change={(e) => handleSpecificFileSelect(e, 'ticketGoFile')} />
                            </label>
                        {/if}
                    </div>
                    <div class="relative">
                        <span class="absolute left-2.5 top-2 md:top-2.5 text-slate-400 text-xs md:text-sm">Rp</span>
                        <Input type="text" value={formatInputNumber(editingCosts.ticketGo)} on:input={(e) => updateCost('ticketGo', e)} disabled={isReadOnly} class="pl-8 md:pl-9 h-9 md:h-10 text-sm bg-white border-slate-200 focus:bg-white {isReadOnly ? 'opacity-70 cursor-not-allowed pointer-events-none' : ''}" />
                    </div>
                    {#if editingCosts.ticketGoFile}
                        <div class="flex items-center justify-between p-2 mt-2 bg-white border border-slate-200 rounded-md">
                            <span class="text-[10px] md:text-xs text-slate-700 truncate mr-2 flex-1">{editingCosts.ticketGoFile.name}</span>
                            <div class="flex gap-2 shrink-0 text-[10px]">
                                <button type="button" class="inline-flex items-center justify-center px-2 py-1 md:px-2.5 md:py-1 bg-blue-50 hover:bg-blue-100 text-blue-700 font-medium rounded-md border border-blue-200 transition-colors" on:click={() => openPreview(editingCosts.ticketGoFile)}>Lihat</button>
                                {#if !isReadOnly}<button type="button" class="inline-flex items-center justify-center px-2 py-1 md:px-2.5 md:py-1 bg-red-50 hover:bg-red-100 text-red-600 font-medium rounded-md border border-red-200 transition-colors" on:click={() => removeSpecificFile('ticketGoFile')}>Hapus</button>{/if}
                            </div>
                        </div>
                    {/if}
                </div>

                <!-- Tiket Pulang -->
                <div class="space-y-1.5 p-3 border border-slate-100 bg-slate-50 rounded-lg">
                    <div class="flex justify-between items-center">
                        <Label class="text-[10px] md:text-xs font-semibold uppercase text-slate-500 tracking-wider">Tiket Pulang</Label>
                        {#if !editingCosts.ticketBackFile && !isReadOnly}
                            <label class="cursor-pointer text-[9px] md:text-[10px] text-blue-700 bg-blue-50 hover:bg-blue-100 px-2 py-1 md:px-2.5 md:py-1 rounded-md border border-blue-200 font-medium transition-colors flex items-center gap-1">
                                <svg xmlns="http://www.w3.org/2000/svg" class="h-3 w-3" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 16v1a3 3 0 003 3h10a3 3 0 003-3v-1m-4-8l-4-4m0 0L8 8m4-4v12" /></svg>
                                Upload Kwitansi
                                <input type="file" class="hidden" accept=".pdf,.jpg,.jpeg,.png" on:change={(e) => handleSpecificFileSelect(e, 'ticketBackFile')} />
                            </label>
                        {/if}
                    </div>
                    <div class="relative">
                        <span class="absolute left-2.5 top-2 md:top-2.5 text-slate-400 text-xs md:text-sm">Rp</span>
                        <Input type="text" value={formatInputNumber(editingCosts.ticketBack)} on:input={(e) => updateCost('ticketBack', e)} disabled={isReadOnly} class="pl-8 md:pl-9 h-9 md:h-10 text-sm bg-white border-slate-200 focus:bg-white {isReadOnly ? 'opacity-70 cursor-not-allowed' : ''}" />
                    </div>
                    {#if editingCosts.ticketBackFile}
                        <div class="flex items-center justify-between p-2 mt-2 bg-white border border-slate-200 rounded-md">
                            <span class="text-[10px] md:text-xs text-slate-700 truncate mr-2 flex-1">{editingCosts.ticketBackFile.name}</span>
                            <div class="flex gap-2 shrink-0 text-[10px]">
                                <button type="button" class="inline-flex items-center justify-center px-2 py-1 md:px-2.5 md:py-1 bg-blue-50 hover:bg-blue-100 text-blue-700 font-medium rounded-md border border-blue-200 transition-colors" on:click={() => openPreview(editingCosts.ticketBackFile)}>Lihat</button>
                                {#if !isReadOnly}<button type="button" class="inline-flex items-center justify-center px-2 py-1 md:px-2.5 md:py-1 bg-red-50 hover:bg-red-100 text-red-600 font-medium rounded-md border border-red-200 transition-colors" on:click={() => removeSpecificFile('ticketBackFile')}>Hapus</button>{/if}
                            </div>
                        </div>
                    {/if}
                </div>

                <!-- Boarding Pass -->
                <div class="space-y-1.5 p-3 border border-slate-100 bg-slate-50 rounded-lg md:col-span-2 flex flex-col justify-center">
                    <div class="flex flex-wrap justify-between items-center gap-2">
                        <Label class="text-[10px] md:text-xs font-semibold uppercase text-slate-500 tracking-wider">Boarding Pass</Label>
                        {#if !isReadOnly}
                            <label class="cursor-pointer text-[10px] text-blue-600 font-medium hover:underline bg-white px-2 py-1.5 rounded border border-blue-200 shadow-sm transition-colors">
                                + Tambah Boarding Pass (.pdf / gambar)
                                <input type="file" class="hidden" accept=".pdf,.jpg,.jpeg,.png" on:change={handleBoardingPassFileSelect} />
                            </label>
                        {/if}
                    </div>
                    
                    {#if editingCosts.boardingPassFiles && editingCosts.boardingPassFiles.length > 0}
                        <div class="grid grid-cols-1 md:grid-cols-2 gap-2 mt-2">
                            {#each editingCosts.boardingPassFiles as bpFile, idx}
                                <div class="flex items-center justify-between p-2 bg-white border border-slate-200 rounded-md w-full">
                                    <span class="text-[10px] md:text-xs text-slate-700 truncate mr-2 flex-1">{bpFile.name}</span>
                                    <div class="flex gap-2 shrink-0 text-[10px]">
                                        <button type="button" class="inline-flex items-center justify-center px-2 py-1 md:px-2.5 md:py-1 bg-blue-50 hover:bg-blue-100 text-blue-700 font-medium rounded-md border border-blue-200 transition-colors" on:click={() => openPreview(bpFile)}>Lihat</button>
                                        {#if !isReadOnly}<button type="button" class="inline-flex items-center justify-center px-2 py-1 md:px-2.5 md:py-1 bg-red-50 hover:bg-red-100 text-red-600 font-medium rounded-md border border-red-200 transition-colors" on:click={() => removeBoardingPassFile(idx)}>Hapus</button>{/if}
                                    </div>
                                </div>
                            {/each}
                        </div>
                    {/if}

                    {#if editingCosts.boardingPassFile}
                        <div class="flex items-center justify-between p-2 mt-2 bg-white border border-slate-200 rounded-md w-full">
                            <span class="text-[10px] md:text-xs text-slate-700 truncate mr-2 flex-1">{editingCosts.boardingPassFile.name}</span>
                            <div class="flex gap-2 shrink-0 text-[10px]">
                                <button type="button" class="inline-flex items-center justify-center px-2 py-1 md:px-2.5 md:py-1 bg-blue-50 hover:bg-blue-100 text-blue-700 font-medium rounded-md border border-blue-200 transition-colors" on:click={() => openPreview(editingCosts.boardingPassFile)}>Lihat</button>
                                {#if !isReadOnly}<button type="button" class="inline-flex items-center justify-center px-2 py-1 md:px-2.5 md:py-1 bg-red-50 hover:bg-red-100 text-red-600 font-medium rounded-md border border-red-200 transition-colors" on:click={() => removeSpecificFile('boardingPassFile')}>Hapus</button>{/if}
                            </div>
                        </div>
                    {/if}
                </div>
            </div>
        </div>

        <!-- Hotel -->
        <div class="p-3 md:p-4 bg-white rounded-xl border border-slate-200 shadow-sm space-y-3 w-full">
            <div class="flex justify-between items-center mb-1">
                <h4 class="text-xs md:text-sm font-semibold text-slate-700 flex items-center gap-1.5 md:gap-2">
                    <svg xmlns="http://www.w3.org/2000/svg" class="h-3.5 w-3.5 md:h-4 md:w-4 text-blue-500" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 21V5a2 2 0 00-2-2H7a2 2 0 00-2 2v16m14 0h2m-2 0h-5m-9 0H3m2 0h5M9 7h1m-1 4h1m4-4h1m-1 4h1m-5 10v-5a1 1 0 011-1h2a1 1 0 011 1v5m-4 0h4" />
                    </svg>
                    Penginapan (Hotel)
                </h4>
                {#if !editingCosts.hotelFile && !isReadOnly}
                    <label class="cursor-pointer text-[10px] text-blue-600 font-medium hover:underline bg-slate-50 px-2 py-1.5 rounded border border-slate-200 transition-colors">
                        + Upload Kwitansi
                        <input type="file" class="hidden" accept=".pdf,.jpg,.jpeg,.png" on:change={(e) => handleSpecificFileSelect(e, 'hotelFile')} />
                    </label>
                {/if}
            </div>

            <div class="grid grid-cols-3 gap-3 md:gap-4 w-full">
                <div class="space-y-1.5 col-span-1">
                    <Label class="text-[10px] md:text-xs font-semibold uppercase text-slate-500 tracking-wider">Malam</Label>
                    <Input type="number" bind:value={editingCosts.hotelDays} disabled={isReadOnly} class="h-9 md:h-10 text-sm bg-slate-50 border-slate-200 px-2 {isReadOnly ? 'opacity-70 cursor-not-allowed pointer-events-none' : ''}" />
                </div>
                <div class="space-y-1.5 col-span-2">
                    <Label class="text-[10px] md:text-xs font-semibold uppercase text-slate-500 tracking-wider">Rate per Malam</Label>
                    <div class="relative">
                        <span class="absolute left-2.5 top-2 md:top-2.5 text-slate-400 text-xs md:text-sm">Rp</span>
                        <Input type="text" value={formatInputNumber(editingCosts.hotelRate)} on:input={(e) => updateCost('hotelRate', e)} disabled={isReadOnly} class="pl-8 md:pl-9 h-9 md:h-10 text-sm bg-slate-50 border-slate-200 {isReadOnly ? 'opacity-70 cursor-not-allowed pointer-events-none' : ''}" />
                    </div>
                    </div>            </div>
            
            {#if editingCosts.hotelFile}
                <div class="flex items-center justify-between p-2 mt-2 bg-slate-50 border border-slate-200 rounded-md w-full">
                    <span class="text-[10px] md:text-xs text-slate-700 truncate mr-2 flex-1">{editingCosts.hotelFile.name}</span>
                    <div class="flex gap-2 shrink-0 text-[10px]">
                        <button type="button" class="inline-flex items-center justify-center px-2 py-1 md:px-2.5 md:py-1 bg-blue-50 hover:bg-blue-100 text-blue-700 font-medium rounded-md border border-blue-200 transition-colors" on:click={() => openPreview(editingCosts.hotelFile)}>Lihat</button>
                        {#if !isReadOnly}<button type="button" class="inline-flex items-center justify-center px-2 py-1 md:px-2.5 md:py-1 bg-red-50 hover:bg-red-100 text-red-600 font-medium rounded-md border border-red-200 transition-colors" on:click={() => removeSpecificFile('hotelFile')}>Hapus</button>{/if}
                    </div>
                </div>
            {/if}

            <div class="text-right text-xs md:text-sm font-mono font-medium text-slate-600 border-t border-slate-100 pt-2 mt-1">
                Subtotal Hotel: <span class="text-slate-800">{formatCurrency(totalHotel)}</span>
            </div>
        </div>

        <!-- Bukti Transportasi atau Rental -->
        <div class="p-3 md:p-4 bg-white rounded-xl border border-slate-200 shadow-sm space-y-3 w-full">
            <div class="flex justify-between items-center mb-1">
                <h4 class="text-xs md:text-sm font-semibold text-slate-700 flex items-center gap-1.5">
                    <svg xmlns="http://www.w3.org/2000/svg" class="h-3.5 w-3.5 md:h-4 md:w-4 text-blue-500" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M8 7h12m0 0l-4-4m4 4l-4 4m0 6H4m0 0l4 4m-4-4l4-4" />
                    </svg>
                    Transport Darat / Rental
                </h4>
                {#if !editingCosts.transportFile && !isReadOnly}
                    <label class="cursor-pointer text-[10px] text-blue-600 font-medium hover:underline bg-slate-50 px-2 py-1.5 rounded border border-slate-200 transition-colors">
                        + Upload Bukti
                        <input type="file" class="hidden" accept=".pdf,.jpg,.jpeg,.png" on:change={(e) => handleSpecificFileSelect(e, 'transportFile')} />
                    </label>
                {/if}
            </div>

            <div class="space-y-1.5 w-full">
                <Label class="text-[10px] md:text-xs font-semibold uppercase text-slate-500 tracking-wider">Total Biaya (Opsional)</Label>
                <div class="relative">
                    <span class="absolute left-2.5 top-2 md:top-2.5 text-slate-400 text-xs md:text-sm">Rp</span>
                    <Input type="text" value={formatInputNumber(editingCosts.transportAmount)} on:input={(e) => updateCost('transportAmount', e)} disabled={isReadOnly} class="pl-8 md:pl-9 h-9 md:h-10 text-sm bg-slate-50 border-slate-200 {isReadOnly ? 'opacity-70 cursor-not-allowed' : ''}" />
                </div>
            </div>

            {#if editingCosts.transportFile}
                <div class="flex items-center justify-between p-2 mt-2 bg-slate-50 border border-slate-200 rounded-md w-full">
                    <span class="text-[10px] md:text-xs text-slate-700 truncate mr-2 flex-1">{editingCosts.transportFile.name}</span>
                    <div class="flex gap-2 shrink-0 text-[10px]">
                        <button type="button" class="inline-flex items-center justify-center px-2 py-1 md:px-2.5 md:py-1 bg-blue-50 hover:bg-blue-100 text-blue-700 font-medium rounded-md border border-blue-200 transition-colors" on:click={() => openPreview(editingCosts.transportFile)}>Lihat</button>
                        {#if !isReadOnly}<button type="button" class="inline-flex items-center justify-center px-2 py-1 md:px-2.5 md:py-1 bg-red-50 hover:bg-red-100 text-red-600 font-medium rounded-md border border-red-200 transition-colors" on:click={() => removeSpecificFile('transportFile')}>Hapus</button>{/if}
                    </div>
                </div>
            {/if}
        </div>

        <!-- Add Cost / Additional Costs -->
        <div class="p-3 md:p-4 bg-slate-50 rounded-xl border border-slate-200 shadow-sm space-y-4 w-full">
            <div class="flex justify-between items-center border-b border-slate-200 pb-2">
                <h4 class="text-xs md:text-sm font-semibold text-slate-700 flex items-center gap-1.5">
                    <svg xmlns="http://www.w3.org/2000/svg" class="h-3.5 w-3.5 md:h-4 md:w-4 text-indigo-500" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 6v6m0 0v6m0-6h6m-6 0H6" />
                    </svg>
                    Biaya Tambahan Lainnya
                </h4>
                {#if !isReadOnly}
                    <Button size="sm" class="h-7 px-3 text-[10px] font-bold tracking-wider bg-indigo-50 hover:bg-indigo-100 text-indigo-700 border border-indigo-200 shadow-sm transition-colors rounded-md flex items-center gap-1" on:click={addAdditionalCost}>
                        <svg xmlns="http://www.w3.org/2000/svg" class="h-3 w-3" viewBox="0 0 20 20" fill="currentColor"><path fill-rule="evenodd" d="M10 3a1 1 0 011 1v5h5a1 1 0 110 2h-5v5a1 1 0 11-2 0v-5H4a1 1 0 110-2h5V4a1 1 0 011-1z" clip-rule="evenodd" /></svg>
                        ADD COST
                    </Button>
                {/if}
            </div>

            <div class="space-y-3">
                {#if editingCosts.additionalCosts && editingCosts.additionalCosts.length > 0}
                    {#each editingCosts.additionalCosts as cost, index}
                        <div class="bg-white p-3 border border-slate-200 rounded-lg relative group">
                            {#if !isReadOnly}
                                <button type="button" class="absolute -top-2 -right-2 bg-red-100 text-red-600 rounded-full p-1 border border-red-200 hover:bg-red-200 transition-colors shadow-sm" on:click={() => removeAdditionalCost(index)}>
                                    <svg xmlns="http://www.w3.org/2000/svg" class="h-3 w-3" viewBox="0 0 20 20" fill="currentColor">
                                        <path fill-rule="evenodd" d="M4.293 4.293a1 1 0 011.414 0L10 8.586l4.293-4.293a1 1 0 111.414 1.414L11.414 10l4.293 4.293a1 1 0 01-1.414 1.414L10 11.414l-4.293 4.293a1 1 0 01-1.414-1.414L8.586 10 4.293 5.707a1 1 0 010-1.414z" clip-rule="evenodd" />
                                    </svg>
                                </button>
                            {/if}
                            <div class="grid grid-cols-1 md:grid-cols-2 gap-3">
                                <div class="space-y-1.5">
                                    <Label class="text-[10px] font-semibold uppercase text-slate-500 tracking-wider">Nama Biaya</Label>
                                    <Input type="text" placeholder="Cth: Taksi Bandara" bind:value={cost.name} disabled={isReadOnly} class="h-8 text-xs bg-slate-50 border-slate-200 {isReadOnly ? 'opacity-70 cursor-not-allowed pointer-events-none' : ''}" />
                                </div>
                                <div class="space-y-1.5">
                                    <Label class="text-[10px] font-semibold uppercase text-slate-500 tracking-wider">Nominal</Label>
                                    <div class="relative">
                                        <span class="absolute left-2 top-1.5 text-slate-400 text-xs">Rp</span>
                                        <Input type="text" value={formatInputNumber(cost.amount)} on:input={(e) => updateAdditionalCostAmount(index, e)} disabled={isReadOnly} class="pl-7 h-8 text-xs bg-slate-50 border-slate-200 {isReadOnly ? 'opacity-70 cursor-not-allowed pointer-events-none' : ''}" />
                                    </div>
                                </div>
                                <div class="md:col-span-2 flex flex-col sm:flex-row sm:items-center justify-between mt-1 pt-2 border-t border-slate-100 gap-2">
                                    <span class="text-[10px] font-medium text-slate-500">Kwitansi / Bukti (.pdf):</span>
                                    {#if !cost.file && !isReadOnly}
                                        <label class="cursor-pointer text-[10px] text-blue-600 font-medium hover:underline bg-slate-50 px-2 py-1.5 rounded border border-slate-200 w-fit">
                                            Upload Kwitansi
                                            <input type="file" class="hidden" accept=".pdf" on:change={(e) => handleAdditionalFileSelect(e, index)} />
                                        </label>
                                    {:else if cost.file}
                                        <div class="flex items-center gap-2 bg-slate-50 p-1.5 rounded border border-slate-200 w-full sm:w-auto">
                                            <span class="text-[10px] text-slate-700 truncate max-w-[150px]">{cost.file.name}</span>
                                            <button type="button" class="inline-flex items-center justify-center px-2 py-1 md:px-2.5 md:py-1 bg-blue-50 hover:bg-blue-100 text-blue-700 font-medium rounded-md border border-blue-200 transition-colors" on:click={() => openPreview(cost.file)}>Lihat</button>
                                            {#if !isReadOnly}
                                                <button type="button" class="inline-flex items-center justify-center px-2 py-1 md:px-2.5 md:py-1 bg-red-50 hover:bg-red-100 text-red-600 font-medium rounded-md border border-red-200 transition-colors" on:click={() => { cost.file = null; editingCosts = editingCosts; }}>Hapus</button>
                                            {/if}
                                        </div>
                                    {/if}
                                </div>
                            </div>
                        </div>
                    {/each}
                    <div class="text-right text-xs md:text-sm font-mono font-medium text-slate-600 pt-2">
                        Subtotal Tambahan: <span class="text-slate-800">{formatCurrency(totalAdditional)}</span>
                    </div>
                {:else}
                    <div class="text-center p-4 border border-dashed border-slate-300 rounded-lg text-xs text-slate-500">
                        Belum ada biaya tambahan diinputkan.
                    </div>
                {/if}
            </div>
        </div>
    </div>

    <DialogFooter class="p-4 md:p-6 pt-0 bg-white shrink-0 z-10 border-t border-slate-100 shadow-[0_-10px_15px_-3px_rgba(0,0,0,0.05)]">
        <div class="w-full flex flex-col md:flex-row md:items-center justify-between gap-4 pt-4">
            <div class="w-full md:w-auto text-left">
                <span class="block text-[10px] md:text-xs text-slate-500 font-semibold uppercase tracking-wider mb-1">Total Estimasi</span>
                <span class="text-xl font-bold text-blue-700 font-mono tracking-tight">{formatCurrency(grandTotal)}</span>
            </div>
            <div class="grid grid-cols-2 md:flex gap-2 w-full md:w-auto">
                <Button variant="outline" class="w-full h-10 border-slate-200 text-slate-600 bg-white px-3 flex justify-center items-center" on:click={() => {
                    window.open(`/print?type=spd&id=${record.id}&spd=${encodeURIComponent(record.spd)}`, '_blank');
                }}>
                    <svg xmlns="http://www.w3.org/2000/svg" class="w-4 h-4 mr-1.5 shrink-0" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M14.5 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V7.5L14.5 2z"/><polyline points="14 2 14 8 20 8"/><line x1="16" x2="8" y1="13" y2="13"/><line x1="16" x2="8" y1="17" y2="17"/><line x1="10" x2="8" y1="9" y2="9"/></svg>
                    <span class="text-xs">Cetak SPD</span>
                </Button>
                <Button variant="outline" class="w-full h-10 border-slate-200 text-slate-600 bg-white px-3 flex justify-center items-center" on:click={() => {
                    window.open(`/print?type=rincian&id=${record.id}&spd=${encodeURIComponent(record.spd)}`, '_blank');
                }}>
                    <svg xmlns="http://www.w3.org/2000/svg" class="w-4 h-4 mr-1.5 shrink-0" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><rect width="16" height="20" x="4" y="2" rx="2"/><line x1="8" x2="16" y1="6" y2="6"/><line x1="16" x2="16" y1="14" y2="18"/><path d="M16 10h.01"/><path d="M12 10h.01"/><path d="M8 10h.01"/><path d="M12 14h.01"/><path d="M8 14h.01"/><path d="M12 18h.01"/><path d="M8 18h.01"/></svg>
                    <span class="text-xs">Cetak Rincian</span>
                </Button>
                {#if isReadOnly}
                    <div class="col-span-2">
                        <Button variant="outline" class="w-full h-10 border-slate-200 text-slate-600 px-3" on:click={() => dispatch('close')}>
                            <span class="text-xs">Tutup</span>
                        </Button>
                    </div>
                {:else}
                    <Button variant="outline" class="w-full h-10 border-slate-200 text-slate-600 px-3" on:click={() => dispatch('close')}>
                        <span class="text-xs">Tutup</span>
                    </Button>
                    <Button class="w-full h-10 bg-blue-600 hover:bg-blue-700 text-white shadow-lg shadow-blue-500/20 px-3" on:click={handleSave}>
                        <span class="text-xs">Simpan Rincian</span>
                    </Button>
                {/if}
            </div>
        </div>
    </DialogFooter>

    {#if showPreview && previewFile}
        <div class="absolute inset-0 z-50 bg-white flex flex-col">
            <div class="flex items-center justify-between px-4 py-3 border-b border-slate-100 bg-slate-50/50 flex-none">
                <div class="flex flex-col min-w-0 pr-4">
                    <h3 class="text-sm font-bold text-slate-800 tracking-tight truncate">Preview Dokumen</h3>
                    <p class="text-[10px] text-slate-500 truncate">{previewFile.name}</p>
                </div>
                <button type="button" class="p-2 -mr-2 text-slate-400 hover:text-red-500 hover:bg-slate-100 rounded-full transition-colors flex-shrink-0" on:click={closePreview}>
                    <svg xmlns="http://www.w3.org/2000/svg" class="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
                    </svg>
                </button>
            </div>
            <div class="flex-1 overflow-hidden bg-slate-100 relative">
                <DocumentViewer 
                    url={previewFile.data} 
                    type={previewFile.type && previewFile.type.startsWith('image/') ? 'image' : (previewFile.type === 'application/pdf' ? 'pdf' : 'docx')} 
                    filename={previewFile.name} 
                />
            </div>
        </div>
    {/if}
</Dialog>
