<script>
    import { page } from '$app/stores';
    import { api } from '$lib/shared/api';
    import { onMount } from 'svelte';
    import { goto } from '$app/navigation';
    import { startLoading, stopLoading } from '$lib/shared/stores/loading';
    import { toast } from '$lib/shared/stores/toast';
    import { userStore } from '$lib/features/auth/store';
    import { formatCurrency, compressImage, toTitleCase } from '$lib/shared/utils/utils';

    // UI
    import Button from '$lib/shared/ui/button/Button.svelte';
    import Label from '$lib/shared/ui/label/Label.svelte';
    import Textarea from '$lib/shared/ui/textarea/Textarea.svelte';
    import DocumentViewer from '$lib/shared/ui/document-viewer/DocumentViewer.svelte';
    import Dialog from '$lib/shared/ui/dialog/Dialog.svelte';

    let id = $page.params.id;
    let record = null;
    let reportContent = '';
    let documentationFiles = [];

    let isPreviewOpen = false;
    let previewFile = null;

    // Confirm Submit Modal State
    let showSubmitConfirmModal = false;

    function promptSubmit() {
        showSubmitConfirmModal = true;
    }

    function confirmSubmitAction() {
        showSubmitConfirmModal = false;
        saveLaporan();
    }

    onMount(async () => {
        await loadRecord();
    });

    async function loadRecord() {
        startLoading('Memuat data Dalkot...');
        try {
            record = await api.getDalkotRecordById(id);
            if (record) {
                reportContent = record.reportContent || '';
                if (Array.isArray(record.documentationFile)) {
                    documentationFiles = record.documentationFile;
                } else if (record.documentationFile) {
                    documentationFiles = [record.documentationFile];
                } else {
                    documentationFiles = [];
                }
            } else {
                toast.error("Data tidak ditemukan");
                goto('/dashboard/laporan');
            }
        } catch(err) {
            console.error(err);
            toast.error("Gagal memuat data dalkot");
        } finally {
            stopLoading();
        }
    }

    async function uploadAndGetPath(file) {
        if (!file) return null;
        startLoading('Mengupload file...');
        try {
            const res = await api.uploadFile(file);
            return res.path;
        } catch(err) {
            toast.error("Gagal mengupload file");
            return null;
        } finally {
            stopLoading();
        }
    }

    async function handleFileSelect(e) {
        const files = Array.from(e.target.files);
        if (files.length === 0) return;
        
        for (let file of files) {
            if (file.size > 10 * 1024 * 1024) {
                toast.error(`Ukuran file ${file.name} melebihi 10MB.`);
                continue;
            }
            
            if (file.type.startsWith('image/')) {
                file = await compressImage(file);
            }
            
            const path = await uploadAndGetPath(file);
            if (path) {
                documentationFiles = [...documentationFiles, {
                    name: file.name,
                    size: file.size,
                    type: file.type,
                    path: path
                }];
            }
        }
        e.target.value = '';
    }

    function removeFile(index) {
        documentationFiles = documentationFiles.filter((_, i) => i !== index);
    }

    function openPreview(file) {
        if (!file) return;
        previewFile = file;
        isPreviewOpen = true;
    }

    async function saveLaporan() {
        if (!reportContent && documentationFiles.length === 0) {
            toast.error("Silakan isi laporan kegiatan atau upload file dokumentasi terlebih dahulu.");
            return;
        }

        startLoading('Menyimpan laporan...');
        try {
            await api.request(`/dalkot/${id}`, {
                method: 'PUT',
                body: JSON.stringify({
                    ...record,
                    reportContent: reportContent,
                    documentationFile: documentationFiles.length > 0 ? documentationFiles : null,
                    status: 'pending' // Update status to pending (Ajukan) upon Laporan submission
                })
            });
            toast.success("Laporan berhasil disimpan!");
            goto('/dashboard/laporan');
        } catch(err) {
            console.error(err);
            toast.error("Gagal menyimpan laporan");
        } finally {
            stopLoading();
        }
    }

    async function printHTML(endpoint) {
        startLoading('Menyiapkan dokumen...');
        try {
            const token = localStorage.getItem('auth_token');
            const res = await fetch(`/api/v1${endpoint}`, {
                headers: {
                    'Authorization': `Bearer ${token}`
                }
            });
            
            if (!res.ok) throw new Error('Gagal mengambil dokumen');
            const html = await res.text();
            
            const printWindow = window.open('', '_blank');
            if (printWindow) {
                printWindow.document.open();
                printWindow.document.write(html);
                printWindow.document.close();
            } else {
                toast.error("Gagal membuka tab baru. Pastikan pop-up diizinkan.");
            }
        } catch (e) {
            console.error(e);
            toast.error("Gagal mencetak dokumen.");
        } finally {
            stopLoading();
        }
    }

    $: onlySpj = record?.assignments?.filter(a => a.assignmentType === 'SPJ') || [];
    $: onlyRiil = record?.assignments?.filter(a => a.assignmentType === 'RIIL') || [];
    $: executionDateStr = record?.executionDate
        ? new Date(record.executionDate).toLocaleDateString('id-ID', { weekday: 'long', day: 'numeric', month: 'long', year: 'numeric' })
        : '-';

</script>

<div class="space-y-6 pb-24 max-w-5xl mx-auto w-full">
    <!-- Header Area -->
    <div class="flex flex-col sm:flex-row justify-between items-start sm:items-center gap-4 bg-white p-6 rounded-2xl shadow-sm border border-slate-100 mt-6 mx-4 md:mx-0">
        <div class="flex items-center gap-4">
            <button 
                on:click={() => goto('/dashboard/laporan')}
                class="p-2 rounded-full hover:bg-slate-100 text-slate-500 transition-colors"
            >
                <svg xmlns="http://www.w3.org/2000/svg" class="h-5 w-5" viewBox="0 0 20 20" fill="currentColor">
                    <path fill-rule="evenodd" d="M9.707 16.707a1 1 0 01-1.414 0l-6-6a1 1 0 010-1.414l6-6a1 1 0 011.414 1.414L5.414 9H17a1 1 0 110 2H5.414l4.293 4.293a1 1 0 010 1.414z" clip-rule="evenodd" />
                </svg>
            </button>
            <div>
                <div class="flex items-center gap-2 mb-1">
                    <span class="bg-indigo-100 text-indigo-700 text-xs font-bold uppercase tracking-widest px-2 py-0.5 rounded border border-indigo-200">
                        Dalam Kota
                    </span>
                    <span class="text-xs text-slate-500 font-mono bg-slate-100 px-2 py-0.5 rounded">{record?.spdNumber || '-'}</span>
                </div>
                <h1 class="text-xl md:text-2xl font-bold text-slate-800 tracking-tight leading-tight">Laporan Kegiatan Dalkot</h1>
            </div>
        </div>
        <div class="flex items-center gap-3">
            <Button variant="outline" class="border-indigo-200 text-indigo-700 hover:bg-indigo-50 font-semibold px-4 py-2 rounded-xl shadow-sm hidden md:flex items-center gap-2" on:click={() => printHTML(`/dalkot/${id}/laporan-stream`)}>
                <svg xmlns="http://www.w3.org/2000/svg" class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M17 17h2a2 2 0 002-2v-4a2 2 0 00-2-2H5a2 2 0 00-2 2v4a2 2 0 002 2h2m2 4h6a2 2 0 002-2v-4a2 2 0 00-2-2H9a2 2 0 00-2 2v4a2 2 0 002 2zm8-12V5a2 2 0 00-2-2H9a2 2 0 00-2 2v4h10z" /></svg>
                Cetak Laporan
            </Button>
            <Button variant="outline" class="border-emerald-200 text-emerald-700 hover:bg-emerald-50 font-semibold px-4 py-2 rounded-xl shadow-sm hidden md:flex items-center gap-2" on:click={() => printHTML(`/dalkot/${id}/dpr-stream`)}>
                <svg xmlns="http://www.w3.org/2000/svg" class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z" /></svg>
                Cetak Semua DPR
            </Button>
            <Button class="bg-indigo-600 hover:bg-indigo-700 text-white font-semibold px-6 py-2 rounded-xl shadow-sm" on:click={promptSubmit}>
                Simpan Laporan
            </Button>
        </div>
    </div>

    {#if record}
    <div class="grid grid-cols-1 lg:grid-cols-3 gap-6 mx-4 md:mx-0">
        <!-- Sidebar Info -->
        <div class="space-y-6">
            <div class="bg-white rounded-2xl shadow-sm border border-slate-100 overflow-hidden">
                <div class="bg-slate-50/80 border-b border-slate-100 px-5 py-4">
                    <h2 class="text-sm font-bold text-slate-700">Informasi Perjalanan</h2>
                </div>
                <div class="p-5 space-y-4">
                    <div>
                        <Label class="text-xs text-slate-400 uppercase tracking-wider mb-1 block">Kegiatan</Label>
                        <div class="text-sm font-medium text-slate-800">{record.activityName || '-'}</div>
                    </div>
                    <div>
                        <Label class="text-xs text-slate-400 uppercase tracking-wider mb-1 block">Lokasi</Label>
                        <div class="text-sm font-medium text-slate-800">{record.location || '-'}</div>
                    </div>
                    <div>
                        <Label class="text-xs text-slate-400 uppercase tracking-wider mb-1 block">Tanggal</Label>
                        <div class="text-sm font-medium text-slate-800">{executionDateStr}</div>
                    </div>
                    <div>
                        <Label class="text-xs text-slate-400 uppercase tracking-wider mb-1 block">Pejabat / Tipe</Label>
                        <div class="text-sm font-medium text-slate-800">{record.official || '-'} &mdash; {record.dalkotType || '-'}</div>
                    </div>
                </div>
            </div>

            <div class="bg-white rounded-2xl shadow-sm border border-slate-100 overflow-hidden">
                <div class="bg-slate-50/80 border-b border-slate-100 px-5 py-4">
                    <h2 class="text-sm font-bold text-slate-700">Petugas</h2>
                </div>
                <div class="p-0">
                    <div class="divide-y divide-slate-100 max-h-64 overflow-y-auto">
                        {#each [...onlySpj, ...onlyRiil].sort((a,b) => (a.sequenceNumber||0)-(b.sequenceNumber||0)) as a}
                            <div class="p-4 hover:bg-slate-50/50 transition-colors flex items-center gap-3">
                                <div class="w-8 h-8 rounded-full bg-indigo-100 text-indigo-600 flex items-center justify-center font-bold text-xs uppercase shrink-0">
                                    {a.user?.name ? a.user.name.charAt(0) : '?'}
                                </div>
                                <div class="flex-1 min-w-0">
                                    <div class="text-sm font-semibold text-slate-800 truncate">{a.user?.name || '-'}</div>
                                    <div class="text-[10px] text-slate-500 truncate flex gap-2 mt-0.5">
                                        <span class="font-mono text-slate-400">ID: {String(a.sequenceNumber || 0).padStart(3, '0')}</span>
                                        <span class={a.assignmentType === 'SPJ' ? 'text-blue-500 font-medium' : 'text-emerald-500 font-medium'}>{a.assignmentType}</span>
                                    </div>
                                </div>
                            </div>
                        {/each}
                    </div>
                </div>
            </div>
        </div>

        <!-- Main Form Area -->
        <div class="lg:col-span-2 space-y-6">
            <div class="bg-white rounded-2xl shadow-sm border border-slate-100 overflow-hidden">
                <div class="p-6 md:p-8 space-y-8">
                    <!-- Text Laporan -->
                    <div class="space-y-3">
                        <Label class="text-base font-bold text-slate-700 flex items-center gap-2">
                            <span class="w-6 h-6 rounded bg-indigo-100 text-indigo-600 flex items-center justify-center text-xs">1</span>
                            Teks Laporan Kegiatan
                        </Label>
                        <p class="text-sm text-slate-500 pl-8">Uraikan laporan ringkas hasil kegiatan dinas dalam kota ini.</p>
                        
                        <div class="pl-8">
                            <Textarea 
                                bind:value={reportContent} 
                                placeholder="Ketikan laporan kegiatan di sini..." 
                                class="min-h-[250px] w-full resize-y text-sm rounded-xl border-slate-200 focus:border-indigo-500 focus:ring-indigo-500"
                            />
                        </div>
                    </div>

                    <hr class="border-slate-100 ml-8" />

                    <!-- File Upload -->
                    <div class="space-y-3">
                        <Label class="text-base font-bold text-slate-700 flex items-center gap-2">
                            <span class="w-6 h-6 rounded bg-emerald-100 text-emerald-600 flex items-center justify-center text-xs">2</span>
                            Dokumentasi (Opsional)
                        </Label>
                        <p class="text-sm text-slate-500 pl-8">Upload file pendukung kegiatan (Foto/PDF).</p>

                        <div class="pl-8 space-y-3">
                            {#if documentationFiles.length > 0}
                                <div class="space-y-2">
                                {#each documentationFiles as file, idx}
                                    <div class="flex items-center justify-between p-3 border border-slate-200 rounded-xl bg-slate-50">
                                        <div class="flex items-center gap-3 overflow-hidden">
                                            <div class="w-10 h-10 rounded bg-white border border-slate-200 flex items-center justify-center shrink-0">
                                                {#if file.type.startsWith('image/')}
                                                    <svg class="w-5 h-5 text-blue-500" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 16l4.586-4.586a2 2 0 012.828 0L16 16m-2-2l1.586-1.586a2 2 0 012.828 0L20 14m-6-6h.01M6 20h12a2 2 0 002-2V6a2 2 0 00-2-2H6a2 2 0 00-2 2v12a2 2 0 002 2z"/></svg>
                                                {:else}
                                                    <svg class="w-5 h-5 text-red-500" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M7 21h10a2 2 0 002-2V9.414a1 1 0 00-.293-.707l-5.414-5.414A1 1 0 0012.586 3H7a2 2 0 00-2 2v14a2 2 0 002 2z"/></svg>
                                                {/if}
                                            </div>
                                            <div class="truncate">
                                                <p class="text-sm font-medium text-slate-700 truncate">{file.name}</p>
                                                <div class="flex items-center gap-2 mt-0.5">
                                                    <span class="text-[10px] text-slate-500 font-mono">{(file.size / 1024 / 1024).toFixed(2)} MB</span>
                                                    <button on:click={() => openPreview({ data: `/api/v1/files/${file.path}`, type: file.type, name: file.name })} class="text-[10px] font-semibold text-indigo-600 hover:underline">Lihat Preview</button>
                                                </div>
                                            </div>
                                        </div>
                                        <button on:click={() => removeFile(idx)} class="p-2 text-red-500 hover:bg-red-50 rounded-lg transition-colors shrink-0">
                                            <svg xmlns="http://www.w3.org/2000/svg" class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                                                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16" />
                                            </svg>
                                        </button>
                                    </div>
                                {/each}
                                </div>
                            {/if}
                            
                            <label class="flex flex-col items-center justify-center w-full h-32 border-2 border-slate-200 border-dashed rounded-xl cursor-pointer hover:bg-slate-50 hover:border-indigo-300 transition-colors bg-white">
                                <div class="flex flex-col items-center justify-center pt-5 pb-6">
                                    <svg class="w-8 h-8 mb-3 text-slate-400" aria-hidden="true" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 20 16">
                                        <path stroke="currentColor" stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13 13h3a3 3 0 0 0 0-6h-.025A5.56 5.56 0 0 0 16 6.5 5.5 5.5 0 0 0 5.207 5.021C5.137 5.017 5.071 5 5 5a4 4 0 0 0 0 8h2.167M10 15V6m0 0L8 8m2-2 2 2"/>
                                    </svg>
                                    <p class="mb-1 text-sm text-slate-500"><span class="font-bold text-indigo-600">Klik untuk upload dokumentasi</span></p>
                                    <p class="text-xs text-slate-400">PDF, JPG, PNG (Max 10MB per file)</p>
                                </div>
                                <input type="file" class="hidden" accept=".pdf,image/*" multiple on:change={handleFileSelect} />
                            </label>
                        </div>
                    </div>
                </div>
            </div>
        </div>
    </div>
    {/if}
</div>

<!-- Document Preview Modal -->
<Dialog bind:open={isPreviewOpen} hideCloseButton={true} class="!w-[95vw] md:!w-[90vw] !max-w-6xl !h-[90vh] md:!h-[85vh] !p-0 overflow-hidden rounded-xl shadow-2xl z-[60]">
    <div class="h-full flex flex-col">
        <div class="flex items-center justify-between px-4 md:px-6 py-3 md:py-4 border-b border-slate-100 bg-slate-50/50 flex-none">
            <div class="flex flex-col min-w-0 pr-4">
                <h3 class="text-base md:text-lg font-bold text-slate-800 tracking-tight truncate">Pratinjau Dokumen</h3>
                {#if previewFile}
                    <p class="text-[10px] md:text-xs text-slate-500 truncate max-w-[200px] sm:max-w-xs md:max-w-md" title={previewFile.name}>{previewFile.name}</p>
                {/if}
            </div>
            
            <button type="button" title="Tutup Preview" aria-label="Tutup Preview" class="p-2 -mr-2 text-slate-400 hover:text-red-500 hover:bg-slate-100 rounded-full transition-colors flex-shrink-0" on:click={() => isPreviewOpen = false}>
                <svg xmlns="http://www.w3.org/2000/svg" class="h-5 w-5 md:h-6 md:w-6" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
                </svg>
            </button>
        </div>
        
        <div class="flex-1 overflow-hidden bg-slate-100 relative p-0 flex items-center justify-center">
            {#if previewFile}
                {#if previewFile.type && previewFile.type.startsWith('image/')}
                    <div class="p-6 h-full w-full flex items-center justify-center overflow-auto">
                        <img src={previewFile.data} alt={previewFile.name} class="max-w-full max-h-[80vh] object-contain rounded-lg shadow-2xl" />
                    </div>
                {:else if previewFile.type && previewFile.type.includes('pdf')}
                    <div class="w-full h-full p-2 md:p-6">
                        <iframe src={previewFile.data} title={previewFile.name} class="w-full h-[85vh] rounded-lg bg-white shadow-xl" frameborder="0"></iframe>
                    </div>
                {:else}
                    <div class="p-6 h-full w-full overflow-y-auto">
                        <DocumentViewer 
                            url={previewFile.data} 
                            type={previewFile.type}
                            filename={previewFile.name}
                        />
                    </div>
                {/if}
            {/if}
        </div>
    </div>
</Dialog>

<!-- Urgent Confirmation Modal for Submit -->
<Dialog bind:open={showSubmitConfirmModal} title="Konfirmasi Submit Laporan Dalkot">
    <div class="space-y-4">
        <div class="p-4 bg-amber-50 border border-amber-200 rounded-xl flex items-start gap-3">
            <svg xmlns="http://www.w3.org/2000/svg" class="h-6 w-6 text-amber-600 flex-shrink-0 mt-0.5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-3L13.732 4c-.77-1.333-2.694-1.333-3.464 0L3.34 16c-.77 1.333.192 3 1.732 3z" />
            </svg>
            <div class="text-sm text-amber-800">
                <p class="font-bold mb-1">Peringatan Penting!</p>
                <p>Setelah Anda men-submit dokumen ini, statusnya akan langsung berubah menjadi <span class="font-bold uppercase tracking-wider">Selesai</span> dan dikirimkan ke server. <strong>Data yang telah disubmit tetap bisa Anda edit kembali</strong> jika terdapat kekeliruan di kemudian hari.</p>
                <p class="mt-2">Mohon pastikan seluruh kelengkapan laporan kegiatan dan bukti dokumentasi sudah benar dan tidak ada yang terlewat.</p>
            </div>
        </div>
        <p class="text-slate-600 text-sm mt-4 font-medium">Apakah Anda yakin ingin men-submit Laporan Kegiatan ini sekarang?</p>
        <div class="flex justify-end gap-3 mt-6 pt-4 border-t border-slate-100">
            <Button variant="outline" on:click={() => showSubmitConfirmModal = false}>Batal, Periksa Lagi</Button>
            <Button class="bg-indigo-600 hover:bg-indigo-700 text-white" on:click={confirmSubmitAction}>Ya, Submit Sekarang</Button>
        </div>
    </div>
</Dialog>
