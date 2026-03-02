<script>
    import { createEventDispatcher } from 'svelte';
    import { toast } from '$lib/stores/toast';
    import { userStore } from '$lib/stores/auth';
    import { provincesStore } from '$lib/stores/master-data';
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
    
    /** @type {{ hotelDays?: number, hotelRate?: number, ticketGo?: number, ticketBack?: number, localTransport?: number, regionalTransport?: number, transportMode?: string, uangRepresentasi?: number, sewaKendaraan?: number }} */
    export let editingCosts = {};

    const dispatch = createEventDispatcher();

    $: isReadOnly = $userStore.role === 'kasubag';

    // Derived Calculations for SBM Uang Harian
    $: days = (() => {
        if (!record?.startDate || !record?.endDate) return 0;
        const start = new Date(record.startDate);
        const end = new Date(record.endDate);
        // Reset time to ignore timezone differences for day calculation
        start.setHours(0,0,0,0);
        end.setHours(0,0,0,0);
        const diffTime = end.getTime() - start.getTime();
        const diffDays = Math.ceil(diffTime / (1000 * 60 * 60 * 24)) + 1; 
        return diffDays > 0 ? diffDays : 0;
    })();
    
    $: sbmRate = $provincesStore.find(p => p.name === record?.province)?.luarKota || 0;
    $: totalDailyAllowance = days * sbmRate;

    // Derived Calculations for Other Costs
    $: totalHotel = (editingCosts.hotelDays || 0) * (editingCosts.hotelRate || 0);
    $: totalTicket = Number(editingCosts.ticketGo || 0) + Number(editingCosts.ticketBack || 0);
    $: totalLocal = Number(editingCosts.localTransport || 0);
    $: totalRegional = Number(editingCosts.regionalTransport || 0);
    $: totalRepresentasi = Number(editingCosts.uangRepresentasi || 0);
    $: totalSewa = Number(editingCosts.sewaKendaraan || 0);
    $: grandTotal = totalTicket + totalDailyAllowance + totalHotel + totalLocal + totalRegional + totalRepresentasi + totalSewa;

    // --- File Upload State & Logic ---
    let uploadedFiles = [];
    let isDragging = false;

    // Reactively initialize files when modal opens
    $: if (open && editingCosts) {
        uploadedFiles = editingCosts.receiptFiles || [];
    }

    function processFiles(selectedFiles) {
        if (uploadedFiles.length + selectedFiles.length > 10) {
            toast.error('Maksimal 10 file yang dapat diunggah.');
            return;
        }

        selectedFiles.forEach(file => {
            if (file.size > 5 * 1024 * 1024) {
                toast.error(`Ukuran file ${file.name} melebihi 5MB.`);
                return;
            }
            const reader = new FileReader();
            reader.onload = (e) => {
                uploadedFiles = [...uploadedFiles, {
                    name: file.name,
                    size: file.size,
                    type: file.type,
                    data: e.target.result // base64
                }];
                editingCosts.receiptFiles = uploadedFiles; // Update bound costs
                editingCosts = editingCosts;
            };
            reader.readAsDataURL(file);
        });
    }

    function handleFileSelect(e) {
        const files = Array.from(e.target.files);
        processFiles(files);
        e.target.value = ''; // Reset
    }

    function handleDrop(e) {
        e.preventDefault();
        isDragging = false;
        if (e.dataTransfer.files && e.dataTransfer.files.length > 0) {
            const files = Array.from(e.dataTransfer.files).filter(f => f.type.match(/(image.*|application\/pdf)/));
            processFiles(files);
        }
    }

    function handleDragOver(e) {
        e.preventDefault();
        isDragging = true;
    }

    function handleDragLeave() {
        isDragging = false;
    }

    function removeFile(index) {
        uploadedFiles = uploadedFiles.filter((_, i) => i !== index);
        editingCosts.receiptFiles = uploadedFiles;
        editingCosts = editingCosts;
    }

    function formatFileSize(bytes) {
        if (bytes === 0) return '0 Bytes';
        const k = 1024;
        const sizes = ['Bytes', 'KB', 'MB', 'GB'];
        const i = Math.floor(Math.log(bytes) / Math.log(k));
        return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i];
    }

    function viewFile(file) {
        if (file.data) {
            const win = window.open();
            if (win) {
                win.document.write(`<iframe src="${file.data}" frameborder="0" style="border:0; top:0px; left:0px; bottom:0px; right:0px; width:100%; height:100%;" allowfullscreen></iframe>`);
            }
        }
    }
    // ---------------------------------

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
        editingCosts = editingCosts; // trigger reactivity
    }

    function validateCosts() {
        if ((editingCosts.ticketGo || 0) < 0) return "Biaya Tiket Berangkat tidak boleh negatif";
        if ((editingCosts.ticketBack || 0) < 0) return "Biaya Tiket Pulang tidak boleh negatif";
        
        if ((editingCosts.hotelDays || 0) < 0) return "Durasi Penginapan tidak boleh negatif";
        // Allow 0 hotel rate if stayed at non-paid accommodation, but strictly check negative
        if ((editingCosts.hotelRate || 0) < 0) return "Rate Penginapan tidak boleh negatif";

        if ((editingCosts.localTransport || 0) < 0) return "Transport Lokal tidak boleh negatif";
        
        if ((editingCosts.regionalTransport || 0) < 0) return "Transport Daerah tidak boleh negatif";

        if ((editingCosts.uangRepresentasi || 0) < 0) return "Uang Representasi tidak boleh negatif";
        if ((editingCosts.sewaKendaraan || 0) < 0) return "Biaya Sewa Kendaraan tidak boleh negatif";

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

<Dialog bind:open={open} class="md:max-w-2xl overflow-hidden flex flex-col p-0 md:p-0" on:close={() => dispatch('close')}>
    <DialogHeader class="border-b border-slate-100 p-4 md:p-6 pb-4 shrink-0 bg-white">
        <DialogTitle class="text-xl">
            {$userStore.role === 'protokol' 
                ? (record?.status === 'Draft' ? 'Input Rincian Biaya' : 'Edit Rincian Biaya') 
                : 'Review Rincian Biaya'}
        </DialogTitle>
        <p class="text-sm text-slate-500 mt-1">Lengkapi komponen biaya untuk <span class="font-semibold text-slate-800">{record?.employee?.name}</span>.</p>
    </DialogHeader>
    
    <div class="grid gap-6 p-4 md:p-6 overflow-y-auto bg-slate-50/30 custom-scrollbar">
        <!-- Surat Tugas Display -->
        <div class="p-4 bg-white rounded-xl border border-slate-200 shadow-sm">
            <Label class="text-xs font-semibold uppercase text-slate-500 tracking-wider block mb-2">Surat Tugas (Dasar Perjalanan)</Label>
            {#if record?.suratTugasPath}
                <div class="flex items-center justify-between gap-3">
                    <div class="flex items-center gap-3 overflow-hidden">
                        <div class="h-10 w-10 rounded-lg bg-rose-50 flex items-center justify-center border border-rose-100 shrink-0">
                            <svg xmlns="http://www.w3.org/2000/svg" class="h-5 w-5 text-rose-500" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z" />
                            </svg>
                        </div>
                        <div class="flex-1 min-w-0">
                            <p class="text-sm font-semibold text-slate-800 truncate" title={record.suratTugasPath}>{record.suratTugasPath}</p>
                            <p class="text-[10px] text-slate-500 mt-0.5">Dokumen pendukung pengajuan SPD</p>
                        </div>
                    </div>
                    <a href={`/uploads/${record.suratTugasPath}`} target="_blank" rel="noopener noreferrer" class="shrink-0 inline-flex items-center justify-center px-3 py-1.5 text-xs font-medium text-blue-700 bg-blue-50 hover:bg-blue-100 border border-blue-200 rounded-md transition-colors">
                        Lihat Dokumen
                    </a>
                </div>
            {:else}
                <div class="text-xs text-slate-500 italic flex items-center gap-2 p-3 bg-slate-50 border border-dashed border-slate-200 rounded-lg">
                    <svg xmlns="http://www.w3.org/2000/svg" class="h-4 w-4 text-slate-400" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-3L13.732 4c-.77-1.333-2.694-1.333-3.464 0L3.34 16c-.77 1.333.192 3 1.732 3z" />
                    </svg>
                    Tidak ada Surat Tugas yang dilampirkan.
                </div>
            {/if}
        </div>

        <!-- Transportation -->
        <div class="grid grid-cols-1 sm:grid-cols-2 gap-4">
            <div class="space-y-2">
                <Label class="text-xs font-semibold uppercase text-slate-500 tracking-wider">Tiket Berangkat</Label>
                <div class="relative">
                    <span class="absolute left-3 top-2.5 text-slate-400 text-sm">Rp</span>
                    <Input type="text" value={formatInputNumber(editingCosts.ticketGo)} on:input={(e) => updateCost('ticketGo', e)} disabled={isReadOnly} class="pl-9 bg-slate-50 border-slate-200 focus:bg-white {isReadOnly ? 'opacity-70 cursor-not-allowed' : ''}" />
                </div>
            </div>
            <div class="space-y-2">
                <Label class="text-xs font-semibold uppercase text-slate-500 tracking-wider">Tiket Pulang</Label>
                <div class="relative">
                    <span class="absolute left-3 top-2.5 text-slate-400 text-sm">Rp</span>
                    <Input type="text" value={formatInputNumber(editingCosts.ticketBack)} on:input={(e) => updateCost('ticketBack', e)} disabled={isReadOnly} class="pl-9 bg-slate-50 border-slate-200 focus:bg-white {isReadOnly ? 'opacity-70 cursor-not-allowed' : ''}" />
                </div>
            </div>
        </div>

        <!-- Daily Allowance (SBM - Read Only) -->
        <div class="p-4 bg-blue-50/50 rounded-xl border border-blue-100 space-y-3">
            <div class="flex flex-col sm:flex-row sm:justify-between sm:items-center gap-2">
                <h4 class="text-sm font-semibold text-blue-800 flex items-center gap-2">
                    <svg xmlns="http://www.w3.org/2000/svg" class="h-4 w-4 text-blue-500" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 8v4l3 3m6-3a9 9 0 11-18 0 9 9 0 0118 0z" />
                    </svg>
                    Uang Harian (Otomatis SBM)
                </h4>
                <span class="text-lg font-bold text-blue-700">{formatCurrency(totalDailyAllowance)}</span>
            </div>
            <div class="flex items-center gap-3 text-sm text-slate-600 bg-white p-3 rounded-lg border border-blue-50/50 shadow-sm">
                <div class="flex-1">
                    <span class="block text-[10px] uppercase tracking-wider text-slate-400 font-semibold mb-0.5">Durasi</span>
                    <span class="font-medium whitespace-nowrap">{days} Hari</span>
                </div>
                <div class="w-px h-8 bg-slate-100"></div>
                <div class="flex-[2]">
                    <span class="block text-[10px] uppercase tracking-wider text-slate-400 font-semibold mb-0.5">Rate Estimasi</span>
                    <span class="font-medium whitespace-nowrap">{formatCurrency(sbmRate)} <span class="text-xs text-slate-400 font-normal hidden sm:inline">/ hari</span></span>
                </div>
            </div>
        </div>

        <!-- Hotel -->
        <div class="p-4 bg-slate-50 rounded-xl border border-slate-100 space-y-4">
            <h4 class="text-sm font-semibold text-slate-700 flex items-center gap-2">
                <svg xmlns="http://www.w3.org/2000/svg" class="h-4 w-4 text-blue-500" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 21V5a2 2 0 00-2-2H7a2 2 0 00-2 2v16m14 0h2m-2 0h-5m-9 0H3m2 0h5M9 7h1m-1 4h1m4-4h1m-1 4h1m-5 10v-5a1 1 0 011-1h2a1 1 0 011 1v5m-4 0h4" />
                </svg>
                Penginapan
            </h4>
            <div class="grid grid-cols-1 sm:grid-cols-3 gap-4">
                <div class="space-y-2 col-span-1">
                    <Label class="text-xs font-semibold uppercase text-slate-500 tracking-wider">Durasi (Malam)</Label>
                    <Input type="number" bind:value={editingCosts.hotelDays} disabled={isReadOnly} class="bg-white border-slate-200 {isReadOnly ? 'opacity-70 cursor-not-allowed' : ''}" />
                </div>
                <div class="space-y-2 col-span-1 sm:col-span-2">
                    <Label class="text-xs font-semibold uppercase text-slate-500 tracking-wider">Rate per Malam</Label>
                    <div class="relative">
                        <span class="absolute left-3 top-2.5 text-slate-400 text-sm">Rp</span>
                        <Input type="text" value={formatInputNumber(editingCosts.hotelRate)} on:input={(e) => updateCost('hotelRate', e)} disabled={isReadOnly} class="pl-9 bg-white border-slate-200 {isReadOnly ? 'opacity-70 cursor-not-allowed' : ''}" />
                    </div>
                </div>
            </div>
            <div class="text-right text-sm font-mono font-medium text-slate-600 border-t border-slate-200 pt-3 mt-2">
                Subtotal: {formatCurrency(totalHotel)}
            </div>
        </div>

        <!-- Local Transport -->
        <div class="grid grid-cols-1 sm:grid-cols-2 gap-4">
            <div class="space-y-2">
                <div class="flex justify-between items-center">
                    <Label class="text-xs font-semibold uppercase text-slate-500 tracking-wider">Transport Lokal</Label>
                </div>
                <div class="relative">
                    <span class="absolute left-3 top-2.5 text-slate-400 text-sm">Rp</span>
                    <Input type="text" value={formatInputNumber(editingCosts.localTransport)} on:input={(e) => updateCost('localTransport', e)} disabled={isReadOnly} class="pl-9 bg-slate-50 border-slate-200 focus:bg-white {isReadOnly ? 'opacity-70 cursor-not-allowed' : ''}" />
                </div>
            </div>
            <div class="space-y-2">
                <Label class="text-xs font-semibold uppercase text-slate-500 tracking-wider">Transport Daerah</Label>
                <div class="relative">
                    <span class="absolute left-3 top-2.5 text-slate-400 text-sm">Rp</span>
                    <Input type="text" value={formatInputNumber(editingCosts.regionalTransport)} on:input={(e) => updateCost('regionalTransport', e)} disabled={isReadOnly} class="pl-9 bg-slate-50 border-slate-200 focus:bg-white {isReadOnly ? 'opacity-70 cursor-not-allowed' : ''}" />
                </div>
            </div>
        </div>

        <!-- Additional Components for Protokol -->
        <div class="grid grid-cols-1 sm:grid-cols-2 gap-4 pt-2 border-t border-slate-100">
            <div class="space-y-2 pt-2">
                <Label class="text-xs font-semibold uppercase text-slate-500 tracking-wider">Uang Representasi</Label>
                <div class="relative">
                    <span class="absolute left-3 top-2.5 text-slate-400 text-sm">Rp</span>
                    <Input type="text" value={formatInputNumber(editingCosts.uangRepresentasi)} on:input={(e) => updateCost('uangRepresentasi', e)} disabled={isReadOnly} class="pl-9 bg-slate-50 border-slate-200 focus:bg-white {isReadOnly ? 'opacity-70 cursor-not-allowed' : ''}" />
                </div>
            </div>
            <div class="space-y-2 pt-2">
                <Label class="text-xs font-semibold uppercase text-slate-500 tracking-wider">Sewa Kendaraan</Label>
                <div class="relative">
                    <span class="absolute left-3 top-2.5 text-slate-400 text-sm">Rp</span>
                    <Input type="text" value={formatInputNumber(editingCosts.sewaKendaraan)} on:input={(e) => updateCost('sewaKendaraan', e)} disabled={isReadOnly} class="pl-9 bg-slate-50 border-slate-200 focus:bg-white {isReadOnly ? 'opacity-70 cursor-not-allowed' : ''}" />
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
                <div class="border-2 {isDragging ? 'border-indigo-500 bg-indigo-50' : 'border-dashed border-slate-200'} rounded-lg p-6 text-center hover:bg-slate-50 transition-colors"
                    on:drop={handleDrop}
                    on:dragover={handleDragOver}
                    on:dragleave={handleDragLeave}
                    role="region"
                    aria-label="File upload area"
                >
                    <svg xmlns="http://www.w3.org/2000/svg" class="mx-auto h-8 w-8 {isDragging ? 'text-indigo-500' : 'text-slate-400'} mb-2" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M7 16a4 4 0 01-.88-7.903A5 5 0 1115.9 6L16 6a5 5 0 011 9.9M15 13l-3-3m0 0l-3 3m3-3v12" />
                    </svg>
                    <div class="text-sm text-slate-600">
                        <label for="file-upload" class="relative cursor-pointer bg-transparent rounded-md font-medium text-blue-600 hover:text-blue-500 focus-within:outline-none">
                            <span>Upload File</span>
                            <input id="file-upload" name="file-upload" type="file" class="sr-only" multiple accept=".pdf,.jpg,.jpeg,.png" on:change={handleFileSelect}>
                        </label>
                        <p class="pl-1 text-slate-500 text-xs mt-1">PDF, PNG, JPG hingga 5MB (Bisa pilih lebih dari satu)</p>
                    </div>
                </div>
            {:else if isReadOnly && (uploadedFiles.length === 0)}
                 <div class="border border-slate-200 rounded-lg p-4 bg-slate-50 text-center text-xs text-slate-500">
                    Tidak ada dokumen yang dilampirkan.
                 </div>
            {/if}

            {#if uploadedFiles.length > 0}
                <div class="mt-4 space-y-2">
                    {#each uploadedFiles as file, index}
                        <div class="flex items-center justify-between p-3 bg-white border border-slate-200 rounded-lg shadow-sm">
                            <div class="flex items-center space-x-3 overflow-hidden">
                                <div class="flex-shrink-0">
                                    {#if file.type?.includes('pdf') || file.name?.endsWith('.pdf')}
                                        <svg class="h-6 w-6 text-rose-500" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                                            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M7 21h10a2 2 0 002-2V9.414a1 1 0 00-.293-.707l-5.414-5.414A1 1 0 0012.586 3H7a2 2 0 00-2 2v14a2 2 0 002 2z" />
                                        </svg>
                                    {:else}
                                        <svg class="h-6 w-6 text-blue-500" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                                            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 16l4.586-4.586a2 2 0 012.828 0L16 16m-2-2l1.586-1.586a2 2 0 012.828 0L20 14m-6-6h.01M6 20h12a2 2 0 002-2V6a2 2 0 00-2-2H6a2 2 0 00-2 2v12a2 2 0 002 2z" />
                                        </svg>
                                    {/if}
                                </div>
                                <div class="flex-1 min-w-0 text-left">
                                    <p class="text-sm font-medium text-slate-900 truncate" title={file.name}>{file.name}</p>
                                    <p class="text-xs text-slate-500">{formatFileSize(file.size)}</p>
                                </div>
                            </div>
                            <div class="flex items-center space-x-2 flex-shrink-0 ml-4">
                                <button type="button" class="text-slate-400 hover:text-blue-600 transition-colors" title="Lihat Dokumen" on:click={() => viewFile(file)}>
                                    <svg xmlns="http://www.w3.org/2000/svg" class="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                                        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 12a3 3 0 11-6 0 3 3 0 016 0z" />
                                        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M2.458 12C3.732 7.943 7.523 5 12 5c4.478 0 8.268 2.943 9.542 7-1.274 4.057-5.064 7-9.542 7-4.477 0-8.268-2.943-9.542-7z" />
                                    </svg>
                                </button>
                                {#if !isReadOnly}
                                    <button type="button" class="text-slate-400 hover:text-red-500 transition-colors" title="Hapus Dokumen" on:click={() => removeFile(index)}>
                                        <svg xmlns="http://www.w3.org/2000/svg" class="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                                            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16" />
                                        </svg>
                                    </button>
                                {/if}
                            </div>
                        </div>
                    {/each}
                </div>
            {/if}
        </div>
    </div>

    <DialogFooter class="p-4 md:p-6 pt-0 bg-white shrink-0">
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
