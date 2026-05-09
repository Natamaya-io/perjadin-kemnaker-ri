<script>
    import { page } from '$app/stores';
    import { recordsStore, loadRecords } from '$lib/features/pengajuan/store';
    import { onMount, onDestroy } from 'svelte';
    import { toast } from '$lib/shared/stores/toast';
    import { api } from '$lib/shared/api';
    import DocumentViewer from '$lib/shared/ui/document-viewer/DocumentViewer.svelte';

    let type = $page.url.searchParams.get('type'); // 'spd', 'rincian', 'laporan'
    let spd = $page.url.searchParams.get('spd');
    let recordId = $page.url.searchParams.get('id');
    
    $: allRecordsForSpd = $recordsStore.filter(r => r.spd === spd);
    $: record = recordId 
        ? $recordsStore.find(r => r.id === recordId) || (allRecordsForSpd.length > 0 ? allRecordsForSpd[0] : null)
        : (allRecordsForSpd.length > 0 ? allRecordsForSpd[0] : null);

    let pdfUrl = '';
    let isGeneratingPdf = false;
    let isDataLoaded = false;
    let pdfError = '';
    let hasAttemptedLoad = false;

    $: if (isDataLoaded && record && !hasAttemptedLoad) {
        hasAttemptedLoad = true;
        loadPdfPreview();
    }

    async function loadPdfPreview() {
        if (!record) return;

        isGeneratingPdf = true;
        pdfError = '';
        try {
            // Kembali gunakan PDF karena Gotenberg sudah terintegrasi komprehensif
            let pdfBlob;
            if (type === 'spd') {
                pdfBlob = await api.exportSpdPdf(record.id);
            } else if (type === 'laporan') {
                pdfBlob = await api.exportLaporanPdf(record.id);
            } else if (type === 'rincian') {
                pdfBlob = await api.exportRincianPdf(record.id);
            }

            if (pdfUrl && pdfUrl.startsWith('blob:')) window.URL.revokeObjectURL(pdfUrl);
            pdfUrl = window.URL.createObjectURL(pdfBlob);
            
            isGeneratingPdf = false;
        } catch (e) {
            console.error('Error loading PDF preview:', e);
            pdfError = e.message || 'Terjadi kesalahan saat memuat dokumen.';
            toast.error('Gagal membuat pratinjau PDF.');
            isGeneratingPdf = false;
        }
    }

    onMount(async () => {
        try {
            if ($recordsStore.length === 0) {
                await loadRecords(spd);
            }
            isDataLoaded = true;
        } catch (e) {
            console.error(e);
            toast.error('Gagal memuat data perjalanan.');
        }
    });

    onDestroy(() => {
        if (pdfUrl && pdfUrl.startsWith('blob:')) window.URL.revokeObjectURL(pdfUrl);
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

</script>

<svelte:head>
    <title>Pratinjau {type ? type.toUpperCase() : 'Dokumen'} - {spd || ''}</title>
</svelte:head>

<div class="min-h-screen bg-slate-100 flex flex-col">
    <!-- Toolbar -->
    <div class="h-16 bg-white border-b border-slate-200 flex items-center justify-between px-6 shadow-sm z-10">
        <div class="flex items-center gap-4">
            <div>
                <h2 class="text-sm font-bold text-slate-800">Pratinjau Dokumen Resmi</h2>
                <p class="text-xs text-slate-500">{spd || 'Memuat...'}</p>
            </div>
        </div>
        
        <div class="flex items-center gap-3">
            {#if isGeneratingPdf}
                <div class="flex items-center gap-2 text-amber-600 text-xs font-medium animate-pulse">
                    <svg class="animate-paper-flight h-3 w-3" xmlns="http://www.w3.org/2000/svg" fill="currentColor" viewBox="0 0 24 24">
                        <path d="M2.01 21L23 12 2.01 3 2 10l15 2-15 2z"/>
                    </svg>
                    Menghasilkan Dokumen...
                </div>
            {/if}

        </div>
    </div>

    <!-- Document Viewer Area -->
    <div class="flex-1 overflow-hidden relative">
        {#if isGeneratingPdf && !pdfUrl}
            <div class="absolute inset-0 flex flex-col items-center justify-center bg-slate-50 z-20">
                <div class="relative flex items-center justify-center overflow-hidden w-24 h-24 mb-4">
                    <svg xmlns="http://www.w3.org/2000/svg" class="w-12 h-12 text-blue-600 animate-paper-flight drop-shadow-md" fill="currentColor" viewBox="0 0 24 24">
                        <path d="M2.01 21L23 12 2.01 3 2 10l15 2-15 2z"/>
                    </svg>
                </div>
                <p class="text-slate-600 font-medium">Sedang menyiapkan pratinjau PDF...</p>
                <p class="text-slate-400 text-sm mt-2">Ini mungkin memakan waktu beberapa detik karena merender via LibreOffice.</p>
            </div>
        {/if}

        {#if pdfUrl}
            <DocumentViewer url={pdfUrl} type="pdf" title="Pratinjau {type.toUpperCase()}" />
        {:else if !isGeneratingPdf && isDataLoaded}
            <div class="flex flex-col items-center justify-center h-full text-slate-400 bg-white">
                <svg xmlns="http://www.w3.org/2000/svg" class="h-16 w-16 mb-4 opacity-20" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z" />
                </svg>
                <p class="text-lg font-medium">Pratinjau tidak tersedia.</p>
                {#if pdfError}
                    <p class="text-sm text-red-500 mt-2 max-w-md text-center">{pdfError}</p>
                {/if}
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
