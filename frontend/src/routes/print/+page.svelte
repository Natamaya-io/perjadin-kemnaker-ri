<script>
    import { page } from '$app/stores';
    import { recordsStore, loadRecords } from '$lib/features/pengajuan/store';
    import { onMount } from 'svelte';
    import { toast } from '$lib/shared/stores/toast';
    import { api } from '$lib/shared/api';
    import DocumentViewer from '$lib/shared/ui/document-viewer/DocumentViewer.svelte';

    let type = $page.url.searchParams.get('type'); // 'spd', 'rincian', 'laporan'
    let spd = $page.url.searchParams.get('spd');
    
    $: allRecordsForSpd = $recordsStore.filter(r => r.spd === spd);
    $: record = allRecordsForSpd.length > 0 ? allRecordsForSpd[0] : null;

    let pdfUrl = '';
    let isGeneratingPdf = false;
    let isDataLoaded = false;

    async function loadPdfPreview() {
        if (!record) return;
        
        isGeneratingPdf = true;
        try {
            let pdfBlob;
            if (type === 'spd') {
                pdfBlob = await api.exportSpdPdf(record.id);
            } else if (type === 'laporan') {
                pdfBlob = await api.exportLaporanPdf(record.id);
            } else if (type === 'rincian') {
                pdfBlob = await api.exportRincianPdf(record.id);
            } else {
                throw new Error('Jenis dokumen tidak dikenal');
            }

            // Create a URL for the blob to be displayed in the viewer
            if (pdfUrl) window.URL.revokeObjectURL(pdfUrl);
            pdfUrl = window.URL.createObjectURL(pdfBlob);
        } catch (e) {
            console.error('Error loading PDF preview:', e);
            toast.error(`Gagal memuat pratinjau dokumen: ${e.message}`);
        } finally {
            isGeneratingPdf = false;
        }
    }

    onMount(async () => {
        try {
            if ($recordsStore.length === 0) {
                await loadRecords();
            }
            isDataLoaded = true;
            await loadPdfPreview();
        } catch (e) {
            console.error(e);
            toast.error('Gagal memuat data perjalanan.');
        }
    });

    async function downloadPdf() {
        if (!pdfUrl || !record) return;
        
        const filename = type === 'spd' 
            ? `SPD_${record.employee.name}_${spd.replace(/\//g, '_')}` 
            : type === 'laporan' 
                ? `Laporan_${record.employee.name}_${spd.replace(/\//g, '_')}`
                : `Rincian_Biaya_${record.employee.name}_${spd.replace(/\//g, '_')}`;

        const a = document.createElement('a');
        a.href = pdfUrl;
        a.download = `${filename}.pdf`;
        document.body.appendChild(a);
        a.click();
        document.body.removeChild(a);
        
        toast.success('PDF berhasil diunduh!');
    }

    function handleBack() {
        if (window.history.length > 1) {
            window.history.back();
        } else {
            window.location.href = '/dashboard/admin/perdin';
        }
    }
</script>

<svelte:head>
    <title>Pratinjau {type ? type.toUpperCase() : 'Dokumen'} - {spd || ''}</title>
</svelte:head>

<div class="min-h-screen bg-slate-100 flex flex-col">
    <!-- Toolbar -->
    <div class="h-16 bg-white border-b border-slate-200 flex items-center justify-between px-6 shadow-sm z-10">
        <div class="flex items-center gap-4">
            <button class="p-2 rounded-full hover:bg-slate-100 transition-colors text-slate-600" on:click={handleBack}>
                <svg xmlns="http://www.w3.org/2000/svg" class="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M10 19l-7-7m0 0l7-7m-7 7h18" />
                </svg>
            </button>
            <div>
                <h2 class="text-sm font-bold text-slate-800">Pratinjau Dokumen Resmi</h2>
                <p class="text-xs text-slate-500">{spd || 'Memuat...'}</p>
            </div>
        </div>
        
        <div class="flex items-center gap-3">
            {#if isGeneratingPdf}
                <div class="flex items-center gap-2 text-amber-600 text-xs font-medium animate-pulse">
                    <svg class="animate-spin h-3 w-3" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24">
                        <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle>
                        <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
                    </svg>
                    Menghasilkan Dokumen...
                </div>
            {/if}

            <button 
                on:click={downloadPdf} 
                disabled={!pdfUrl || isGeneratingPdf} 
                class="flex items-center gap-2 h-10 px-6 bg-blue-600 hover:bg-blue-700 text-white rounded-lg font-medium text-sm shadow-sm disabled:opacity-50 transition-all"
            >
                <svg xmlns="http://www.w3.org/2000/svg" class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 16v1a3 3 0 003 3h10a3 3 0 003-3v-1m-4-4l-4 4m0 0l-4-4m4 4V4" />
                </svg>
                <span>Unduh PDF</span>
            </button>
        </div>
    </div>

    <!-- Document Viewer Area -->
    <div class="flex-1 overflow-hidden relative">
        {#if isGeneratingPdf && !pdfUrl}
            <div class="absolute inset-0 flex flex-col items-center justify-center bg-slate-50 z-20">
                <div class="w-16 h-16 border-4 border-blue-200 border-t-blue-600 rounded-full animate-spin mb-4"></div>
                <p class="text-slate-600 font-medium">Sedang menyiapkan pratinjau dari template DOCX...</p>
                <p class="text-slate-400 text-sm mt-2">Ini mungkin memakan waktu beberapa detik karena merender via LibreOffice.</p>
            </div>
        {/if}

        {#if pdfUrl}
            <DocumentViewer url={pdfUrl} type="pdf" title="Pratinjau {type.toUpperCase()}" />
        {:else if !isGeneratingPdf && isDataLoaded}
            <div class="flex flex-col items-center justify-center h-full text-slate-400 bg-white">
                <svg xmlns="http://www.w3.org/2000/center" class="h-16 w-16 mb-4 opacity-20" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z" />
                </svg>
                <p class="text-lg font-medium">Pratinjau tidak tersedia.</p>
            </div>
        {/if}
    </div>
</div>

<style>
    /* Remove Paged.js global styles that might interfere */
    :global(.pagedjs_pages) {
        display: none !important;
    }
</style>
