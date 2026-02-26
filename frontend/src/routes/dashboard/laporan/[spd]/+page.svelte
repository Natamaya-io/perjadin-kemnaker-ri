<script>
    import { page } from '$app/stores';
    import { recordsStore, updateRecord } from '$lib/stores/records';
    import { userStore } from '$lib/stores/auth';
    import { toast } from '$lib/stores/toast';
    import { onMount } from 'svelte';
    import { goto } from '$app/navigation';
    import { fade } from 'svelte/transition';
    
    // UI
    import Button from '$lib/components/ui/button/Button.svelte';
    import Label from '$lib/components/ui/label/Label.svelte';
    import Textarea from '$lib/components/ui/textarea/Textarea.svelte';
    import Input from '$lib/components/ui/input/Input.svelte';
    import DocumentViewer from '$lib/components/ui/document-viewer/DocumentViewer.svelte';
    import Dialog from '$lib/components/ui/dialog/Dialog.svelte';

    let spd = $page.params.spd || ''; 
    
    $: record = spd ? $recordsStore.find(r => r.spd === decodeURIComponent(spd)) : undefined;

    let reportText = '';
    /** @type {Array<{name: string, type: string, data: string, timestamp: string}>} */
    let uploadedFiles = []; 
    let isDragging = false;
    
    // Preview Modal State
    let isPreviewOpen = false;
    /** @type {{name: string, type: string, data: string} | null} */
    let previewFile = null;

    /** @param {{name: string, type: string, data: string}} file */
    function openPreview(file) {
        previewFile = file;
        isPreviewOpen = true;
    }

    onMount(() => {
        if (record && record.reportData) {
            reportText = record.reportData.text;
            uploadedFiles = record.reportData.files || [];
        }
    });

    /** @param {File[]} selectedFiles */
    function processFiles(selectedFiles) {
        if (uploadedFiles.length + selectedFiles.length > 6) {
            toast.error('Maksimal 6 file yang diperbolehkan.');
            return;
        }

        selectedFiles.forEach(file => {
            if (file.size > 10 * 1024 * 1024) { 
                toast.error(`File ${file.name} melebihi batas 10MB.`);
                return;
            }
            
            // Image Processing: Resize & Metadata
            if (file.type.startsWith('image/')) {
                const reader = new FileReader();
                reader.onload = (e) => {
                    const img = new Image();
                    img.onload = () => {
                        const canvas = document.createElement('canvas');
                        const ctx = canvas.getContext('2d');
                        if (!ctx) return;
                        
                        // Resize logic (Max width 800px)
                        const maxWidth = 800;
                        let width = img.width;
                        let height = img.height;
                        
                        if (width > maxWidth) {
                            height = height * (maxWidth / width);
                            width = maxWidth;
                        }
                        
                        canvas.width = width;
                        canvas.height = height;
                        ctx.drawImage(img, 0, 0, width, height);
                        
                        // Watermark (Draw on canvas)
                        ctx.font = '12px sans-serif';
                        ctx.fillStyle = 'rgba(0, 0, 0, 0.5)';
                        ctx.fillRect(width - 200, height - 30, 190, 20);
                        ctx.fillStyle = 'white';
                        ctx.fillText(`${new Date(file.lastModified).toLocaleDateString()} • ${record.location}`, width - 190, height - 15);

                        const resizedData = canvas.toDataURL('image/jpeg', 0.7);
                        
                        uploadedFiles = [...uploadedFiles, { 
                            name: file.name, 
                            type: file.type, 
                            data: resizedData, 
                            timestamp: new Date(file.lastModified).toISOString()
                        }];
                    };
                    img.src = e.target.result;
                };
                reader.readAsDataURL(file);
            } else {
                const reader = new FileReader();
                reader.onload = (e) => {
                    uploadedFiles = [...uploadedFiles, { 
                        name: file.name, 
                        type: file.type, 
                        data: e.target.result, 
                        timestamp: new Date(file.lastModified).toISOString()
                    }];
                };
                reader.readAsDataURL(file);
            }
        });
    }

    /** @param {Event} event */
    function handleFileChange(event) {
        // @ts-ignore
        const selectedFiles = Array.from(event.target.files);
        processFiles(selectedFiles);
        event.target.value = ''; 
    }

    function handleDragOver(e) {
        e.preventDefault();
        isDragging = true;
    }

    /** @param {DragEvent} e */
    function handleDragLeave(e) {
        e.preventDefault();
        isDragging = false;
    }

    /** @param {DragEvent} e */
    function handleDrop(e) {
        e.preventDefault();
        isDragging = false;
        if (e.dataTransfer.files && e.dataTransfer.files.length > 0) {
            const droppedFiles = Array.from(e.dataTransfer.files).filter(f => 
                f.type.startsWith('image/') || f.type === 'application/pdf'
            );
            if (droppedFiles.length !== e.dataTransfer.files.length) {
                toast.warning('Beberapa file diabaikan. Hanya gambar dan PDF yang diperbolehkan.');
            }
            processFiles(droppedFiles);
        }
    }

    /** @param {number} index */
    function removeFile(index) {
        uploadedFiles = uploadedFiles.filter((_, i) => i !== index);
    }

    async function handleSubmit() {
        if (!reportText) {
             toast.error('Isi laporan tidak boleh kosong.');
             return;
        }

        const words = reportText.trim().split(/\s+/).length;
        if (words > 200) {
            toast.error(`Laporan terlalu panjang (${words} kata). Maksimal 200 kata.`);
            return;
        }

        if (uploadedFiles.length === 0) {
            toast.error('Harap unggah minimal 1 dokumentasi.');
            return;
        }

        if (!record) return;

        try {
            await updateRecord(record.id, {
                reportStatus: 'Completed',
                reportData: {
                    text: reportText,
                    files: uploadedFiles,
                    submittedAt: new Date().toISOString()
                }
            });

            toast.success('Laporan berhasil disimpan!');
            goto('/dashboard/laporan');
        } catch (error) {
            toast.error('Gagal menyimpan laporan. Silakan coba lagi.');
            console.error('Submit report error:', error);
        }
    }
</script>

<div class="max-w-4xl mx-auto pb-20 space-y-8">
    <div class="flex flex-col sm:flex-row sm:items-center justify-between gap-4">
        <div>
            <h2 class="text-2xl sm:text-3xl font-bold tracking-tight text-slate-900">Form Laporan Kegiatan</h2>
            <p class="text-sm sm:text-base text-slate-500">Isi detail laporan dan unggah dokumentasi kegiatan.</p>
        </div>
        <Button variant="outline" class="w-full sm:w-auto border-slate-200 text-slate-600 hover:text-slate-900" on:click={() => goto('/dashboard/laporan')}>
            <svg xmlns="http://www.w3.org/2000/svg" class="h-4 w-4 mr-2" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M10 19l-7-7m0 0l7-7m-7 7h18" />
            </svg>
            Kembali
        </Button>
    </div>

    {#if !record}
        <div class="text-red-500 p-4 border border-red-200 rounded bg-red-50">Data perjalanan dinas tidak ditemukan.</div>
    {:else}
        <div class="bg-white rounded-xl shadow-sm border border-slate-200 overflow-hidden">
            <!-- Header Info -->
            <div class="bg-slate-50/50 border-b border-slate-100 p-6 grid grid-cols-1 md:grid-cols-2 gap-6 text-sm">
                <div>
                    <span class="block text-xs font-semibold uppercase text-slate-400 tracking-wider mb-1">Nomor SPD</span>
                    <span class="font-mono text-slate-700 font-medium bg-white px-2 py-1 rounded border border-slate-200 inline-block">{record.spd}</span>
                </div>
                <div>
                    <span class="block text-xs font-semibold uppercase text-slate-400 tracking-wider mb-1">Pegawai Pelaksana</span>
                    <span class="font-medium text-slate-800">{record.employee.name}</span>
                    <span class="text-slate-500 block text-xs">{record.employee.pangkat}</span>
                </div>
                <div>
                    <span class="block text-xs font-semibold uppercase text-slate-400 tracking-wider mb-1">Lokasi & Tujuan</span>
                    <span class="font-medium text-slate-800">{record.location}, {record.province}</span>
                </div>
                <div>
                    <span class="block text-xs font-semibold uppercase text-slate-400 tracking-wider mb-1">Periode Perjalanan</span>
                    <span class="font-medium text-slate-800">{new Date(record.startDate).toLocaleDateString('id-ID', { day: 'numeric', month: 'long', year: 'numeric'})} &mdash; {new Date(record.endDate).toLocaleDateString('id-ID', { day: 'numeric', month: 'long', year: 'numeric'})}</span>
                </div>
            </div>

            <div class="p-6 md:p-8 space-y-8">
                <!-- Report Text -->
                <div class="space-y-3">
                    <div class="flex justify-between items-center">
                        <Label class="text-base font-semibold text-slate-800">Isi Laporan Kegiatan</Label>
                        <span class="text-xs text-slate-400 font-medium px-2 py-1 bg-slate-50 rounded-full border border-slate-100">
                            {reportText.trim().split(/\s+/).filter(w => w.length > 0).length} / 200 Kata
                        </span>
                    </div>
                    <Textarea 
                        rows="12" 
                        class="resize-y min-h-[200px] text-base leading-relaxed p-4 border-slate-200 focus:border-blue-300 focus:ring-blue-100 placeholder:text-slate-300"
                        placeholder="Deskripsikan hasil kegiatan, kendala yang dihadapi, dan tindak lanjut yang diperlukan..." 
                        bind:value={reportText} 
                        disabled={$userStore.role === 'kasubag'}
                    />
                    <p class="text-xs text-slate-400 italic">Maksimal 200 kata. Gunakan bahasa yang baku dan jelas.</p>
                </div>

                <!-- File Upload -->
                <div class="space-y-4 pt-6 border-t border-slate-100">
                    <div class="flex justify-between items-center">
                         <Label class="text-base font-semibold text-slate-800">Dokumentasi Kegiatan</Label>
                         {#if $userStore.role !== 'kasubag'}
                             <span class="text-xs text-slate-400">Max 6 File (JPG/PDF), Max 10MB</span>
                         {/if}
                    </div>
                    
                    {#if $userStore.role !== 'kasubag'}
                    <label 
                        class="flex flex-col items-center justify-center w-full h-32 border-2 border-dashed rounded-lg cursor-pointer transition-all group relative
                        {isDragging ? 'border-blue-500 bg-blue-50' : 'border-slate-200 bg-slate-50 hover:bg-slate-100 hover:border-blue-300'}"
                        on:dragover={handleDragOver}
                        on:dragleave={handleDragLeave}
                        on:drop={handleDrop}
                    >
                        <div class="flex flex-col items-center justify-center pt-5 pb-6 pointer-events-none">
                            <svg xmlns="http://www.w3.org/2000/svg" class="w-8 h-8 mb-3 text-slate-400 group-hover:text-blue-500 transition-colors" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M7 16a4 4 0 01-.88-7.903A5 5 0 1115.9 6L16 6a5 5 0 011 9.9M15 13l-3-3m0 0l-3 3m3-3v12" />
                            </svg>
                            <p class="mb-1 text-sm text-slate-500"><span class="font-semibold text-blue-600">Klik untuk unggah</span> atau seret file ke sini</p>
                            <p class="text-xs text-slate-400">PDF, PNG, JPG (Maks. 10MB)</p>
                        </div>
                        {#if isDragging}
                            <div class="absolute inset-0 flex items-center justify-center bg-blue-50/90 rounded-lg pointer-events-none">
                                <p class="text-blue-600 font-bold text-lg animate-pulse">Lepaskan file di sini</p>
                            </div>
                        {/if}
                        <input type="file" multiple accept="image/*,application/pdf" class="hidden" on:change={handleFileChange} />
                    </label>
                    {/if}
                    
                    {#if uploadedFiles.length > 0}
                        <div class="grid grid-cols-2 sm:grid-cols-3 md:grid-cols-4 gap-4 mt-6">
                            {#each uploadedFiles as file, i}
                                <div transition:fade class="group relative aspect-square bg-slate-100 rounded-lg overflow-hidden border border-slate-200 shadow-sm cursor-pointer" on:click={() => openPreview(file)}>
                                    {#if file.type.startsWith('image/')}
                                        <img src={file.data} alt="Preview" class="object-cover w-full h-full transition-transform duration-500 group-hover:scale-110" />
                                    {:else if file.type === 'application/pdf'}
                                        <div class="flex flex-col items-center justify-center h-full text-red-500 bg-red-50 p-4 text-center group-hover:bg-red-100 transition-colors">
                                            <svg xmlns="http://www.w3.org/2000/svg" class="h-10 w-10 mb-2" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                                                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M7 21h10a2 2 0 002-2V9.414a1 1 0 00-.293-.707l-5.414-5.414A1 1 0 0012.586 3H7a2 2 0 00-2 2v14a2 2 0 002 2z" />
                                            </svg>
                                            <span class="text-[10px] font-bold uppercase tracking-wider text-red-700">PDF Document</span>
                                            <span class="text-[9px] truncate w-full mt-1 text-red-600/70">{file.name}</span>
                                        </div>
                                    {:else if file.type === 'application/vnd.openxmlformats-officedocument.wordprocessingml.document'}
                                        <div class="flex flex-col items-center justify-center h-full text-blue-600 bg-blue-50 p-4 text-center group-hover:bg-blue-100 transition-colors">
                                            <svg xmlns="http://www.w3.org/2000/svg" class="h-10 w-10 mb-2" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                                                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z" />
                                            </svg>
                                            <span class="text-[10px] font-bold uppercase tracking-wider text-blue-700">Word Document</span>
                                            <span class="text-[9px] truncate w-full mt-1 text-blue-600/70">{file.name}</span>
                                        </div>
                                    {:else}
                                        <div class="flex flex-col items-center justify-center h-full text-slate-400 p-4 text-center">
                                            <svg xmlns="http://www.w3.org/2000/svg" class="h-8 w-8 mb-2" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                                                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z" />
                                            </svg>
                                            <span class="text-[10px] truncate w-full">{file.name}</span>
                                        </div>
                                    {/if}
                                    
                                    <div class="absolute inset-0 bg-black/40 opacity-0 group-hover:opacity-100 transition-opacity flex items-center justify-center z-10" on:click|stopPropagation>
                                        <div class="flex gap-2">
                                             <button 
                                                class="bg-white/20 backdrop-blur-md text-white p-2 rounded-full hover:bg-white/40 transform hover:scale-110 transition-all shadow-lg"
                                                on:click|stopPropagation={() => openPreview(file)}
                                                title="Lihat"
                                            >
                                                <svg xmlns="http://www.w3.org/2000/svg" class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                                                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 12a3 3 0 11-6 0 3 3 0 016 0z" />
                                                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M2.458 12C3.732 7.943 7.523 5 12 5c4.478 0 8.268 2.943 9.542 7-1.274 4.057-5.064 7-9.542 7-4.477 0-8.268-2.943-9.542-7z" />
                                                </svg>
                                            </button>
                                            {#if $userStore.role !== 'kasubag'}
                                            <button 
                                                class="bg-red-500 text-white p-2 rounded-full hover:bg-red-600 transform hover:scale-110 transition-all shadow-lg"
                                                on:click|stopPropagation={() => removeFile(i)}
                                                title="Hapus file"
                                            >
                                                <svg xmlns="http://www.w3.org/2000/svg" class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                                                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16" />
                                                </svg>
                                            </button>
                                            {/if}
                                        </div>
                                    </div>

                                    <!-- Watermark Badge -->
                                    <div class="absolute bottom-2 right-2 bg-black/60 backdrop-blur-[2px] text-white text-[9px] px-1.5 py-0.5 rounded pointer-events-none z-0">
                                        {new Date(file.timestamp).toLocaleDateString()}
                                    </div>
                                </div>
                            {/each}
                        </div>
                    {/if}
                </div>
            </div>

            {#if $userStore.role !== 'kasubag'}
            <div class="bg-slate-50 border-t border-slate-100 p-6 flex justify-end">
                <Button size="lg" class="bg-blue-600 hover:bg-blue-700 text-white shadow-lg shadow-blue-500/20 px-8" on:click={handleSubmit}>Simpan Laporan Akhir</Button>
            </div>
            {/if}
        </div>
    {/if}

    <!-- Document Preview Modal -->
    <Dialog open={isPreviewOpen} on:close={() => isPreviewOpen = false} class="max-w-5xl h-[85vh] p-0 overflow-hidden">
        <div class="h-full flex flex-col">
            <div class="flex items-center justify-between px-6 py-4 border-b border-slate-100 bg-slate-50/50 flex-none">
                <div class="flex flex-col">
                    <h3 class="text-lg font-bold text-slate-800 tracking-tight">Pratinjau Dokumen</h3>
                    {#if previewFile}
                        <p class="text-xs text-slate-500 truncate max-w-md">{previewFile.name}</p>
                    {/if}
                </div>
                <button class="p-2 hover:bg-slate-200 rounded-full text-slate-400 hover:text-slate-600 transition-colors" on:click={() => isPreviewOpen = false}>
                    <svg xmlns="http://www.w3.org/2000/svg" class="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
                    </svg>
                </button>
            </div>
            
            <div class="flex-1 overflow-hidden bg-slate-100 relative p-0">
                {#if previewFile}
                    <DocumentViewer 
                        url={previewFile.data} 
                        type={previewFile.type.startsWith('image/') ? 'image' : (previewFile.type === 'application/pdf' ? 'pdf' : 'docx')} 
                        filename={previewFile.name} 
                    />
                {/if}
            </div>
        </div>
    </Dialog>
</div>
