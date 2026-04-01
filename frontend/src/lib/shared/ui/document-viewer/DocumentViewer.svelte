<script>
    import { onMount, onDestroy } from 'svelte';
    import Button from '$lib/shared/ui/button/Button.svelte';
    import pdfWorkerUrl from 'pdfjs-dist/build/pdf.worker.min.mjs?url';
    
    export let url; // URL or Data URI
    export let type; // 'pdf' | 'docx' | 'image'
    export let filename = 'Dokumen';

    let container;
    let canvas;
    let loading = true;
    let error = null;
    let pdfDoc = null;
    let pageNum = 1;
    let pageRendering = false;
    let pageNumPending = null;
    let scale = 1.5; // Default scale for better readability
    let totalPages = 0;
    let currentRenderTask = null;
    $: finalUrl = (typeof url === 'string' && url.startsWith('/uploads')) ? window.location.origin + url + '?t=' + new Date().getTime() : url;

    // Pan & Zoom CSS state
    let cssScale = 1.0;
    let panX = 0;
    let panY = 0;
    
    // Dynamic libraries
    let pdfjsLib;
    let renderAsync;

    async function loadDocument() {
        if (!finalUrl) return;
        
        loading = true;
        error = null;
        pdfDoc = null;
        totalPages = 0;
        resetCssZoom();

        try {
            if (type === 'pdf') {
                if (!pdfjsLib) {
                    const mod = await import('pdfjs-dist');
                    pdfjsLib = mod;
                    pdfjsLib.GlobalWorkerOptions.workerSrc = pdfWorkerUrl;
                }
                
                let docParams;
                
                // Fetch the PDF robustly
                const res = await fetch(finalUrl);
                if (!res.ok) throw new Error(`Gagal mengambil file: ${res.statusText}`);
                
                const buf = await res.arrayBuffer();
                if (buf.byteLength === 0) {
                    throw new Error("Data dokumen PDF kosong (0 byte) diterima dari server.");
                }
                
                docParams = { data: new Uint8Array(buf) };

                const loadingTask = pdfjsLib.getDocument(docParams);
                pdfDoc = await loadingTask.promise;
                totalPages = pdfDoc.numPages;
                await renderPage(pageNum);
            } else if (type === 'docx') {
                if (!renderAsync) {
                    const mod = await import('docx-preview');
                    renderAsync = mod.renderAsync;
                }

                const response = await fetch(finalUrl);
                const blob = await response.blob();
                if (container) {
                    container.innerHTML = '';
                    await renderAsync(blob, container, container, {
                        className: 'docx-viewer',
                        inWrapper: true
                    });
                }
                loading = false;
            } else if (type === 'image') {
                loading = false;
            }
        } catch (err) {
            console.error("Error loading document:", err);
            error = "Gagal memuat dokumen. " + (err.message || err.toString());
            loading = false;
        }
    }

    async function renderPage(num) {
        if (!pdfDoc || !canvas) return;
        
        // Jika sedang merender, antrikan halaman ini dan keluar
        if (pageRendering) {
            pageNumPending = num;
            return;
        }
        
        pageRendering = true;
        
        try {
            const page = await pdfDoc.getPage(num);
            const viewport = page.getViewport({ scale });
            
            const ctx = canvas.getContext('2d', { willReadFrequently: true });
            canvas.height = viewport.height;
            canvas.width = viewport.width;

            const renderContext = {
                canvasContext: ctx,
                viewport: viewport
            };
            
            // Batalkan task sebelumnya jika ada
            if (currentRenderTask) {
                await currentRenderTask.cancel();
            }

            currentRenderTask = page.render(renderContext);
            await currentRenderTask.promise;
            
            currentRenderTask = null;
            pageRendering = false;
            loading = false;

            // Jika ada antrian halaman baru saat kita sedang merender tadi, jalankan sekarang
            if (pageNumPending !== null) {
                const nextNum = pageNumPending;
                pageNumPending = null;
                renderPage(nextNum);
            }
        } catch (err) {
            if (err.name !== 'RenderingCancelledException') {
                console.error("Page render error:", err);
            }
            pageRendering = false;
            currentRenderTask = null;
        }
    }

    function queueRenderPage(num) {
        renderPage(num);
    }

    function onPrevPage() {
        if (pageNum <= 1) return;
        pageNum--;
        resetCssZoom();
        renderPage(pageNum);
    }

    function onNextPage() {
        if (pageNum >= totalPages) return;
        pageNum++;
        resetCssZoom();
        renderPage(pageNum);
    }

    function onZoomIn() {
        scale += 0.2;
        renderPage(pageNum);
    }

    function onZoomOut() {
        if (scale <= 0.4) return;
        scale -= 0.2;
        renderPage(pageNum);
    }

    function resetCssZoom() {
        cssScale = 1.0;
        panX = 0;
        panY = 0;
    }

    function panzoom(node, triggerState) {
        let isDraggingContent = false;
        let dragStartX = 0, dragStartY = 0;
        let initialDistance = null;
        let initialScale = 1;

        function updateTransform() {
            node.style.transform = `translate(${panX}px, ${panY}px) scale(${cssScale})`;
            node.style.transition = isDraggingContent ? 'none' : 'transform 0.1s ease-out';
        }

        function handleTouchStart(e) {
            if (e.touches.length === 2) {
                initialDistance = Math.hypot(
                    e.touches[0].clientX - e.touches[1].clientX,
                    e.touches[0].clientY - e.touches[1].clientY
                );
                initialScale = cssScale;
                isDraggingContent = false;
            } else if (e.touches.length === 1) {
                isDraggingContent = true;
                dragStartX = e.touches[0].clientX - panX;
                dragStartY = e.touches[0].clientY - panY;
                node.style.transition = 'none';
            }
        }

        function handleTouchMove(e) {
            if (e.touches.length === 2) {
                e.preventDefault();
                const currentDistance = Math.hypot(
                    e.touches[0].clientX - e.touches[1].clientX,
                    e.touches[0].clientY - e.touches[1].clientY
                );
                cssScale = Math.min(Math.max(0.5, initialScale * (currentDistance / initialDistance)), 6);
                updateTransform();
            } else if (e.touches.length === 1 && isDraggingContent) {
                if (cssScale > 1) {
                    e.preventDefault();
                    panX = e.touches[0].clientX - dragStartX;
                    panY = e.touches[0].clientY - dragStartY;
                    updateTransform();
                }
            }
        }

        function handleTouchEnd() {
            isDraggingContent = false;
            initialDistance = null;
            node.style.transition = 'transform 0.1s ease-out';
        }

        function handleMouseDown(e) {
            isDraggingContent = true;
            dragStartX = e.clientX - panX;
            dragStartY = e.clientY - panY;
            node.style.cursor = 'grabbing';
            node.style.transition = 'none';
        }

        function handleMouseMove(e) {
            if (isDraggingContent) {
                if (cssScale > 1 || type === 'image') {
                    e.preventDefault();
                }
                panX = e.clientX - dragStartX;
                panY = e.clientY - dragStartY;
                updateTransform();
            }
        }

        function handleMouseUp() {
            isDraggingContent = false;
            node.style.cursor = 'grab';
            node.style.transition = 'transform 0.1s ease-out';
        }

        function handleWheel(e) {
            if (e.ctrlKey || e.metaKey) {
                e.preventDefault();
                const delta = e.deltaY * -0.01;
                cssScale = Math.min(Math.max(0.5, cssScale + delta), 6);
                updateTransform();
            }
        }

        const containerElem = node.parentElement;
        containerElem.addEventListener('touchstart', handleTouchStart, { passive: false });
        containerElem.addEventListener('touchmove', handleTouchMove, { passive: false });
        containerElem.addEventListener('touchend', handleTouchEnd);
        containerElem.addEventListener('mousedown', handleMouseDown);
        window.addEventListener('mousemove', handleMouseMove, { passive: false });
        window.addEventListener('mouseup', handleMouseUp);
        containerElem.addEventListener('wheel', handleWheel, { passive: false });
        node.style.cursor = 'grab';
        updateTransform();

        return {
            update() { updateTransform(); },
            destroy() {
                containerElem.removeEventListener('touchstart', handleTouchStart);
                containerElem.removeEventListener('touchmove', handleTouchMove);
                containerElem.removeEventListener('touchend', handleTouchEnd);
                containerElem.removeEventListener('mousedown', handleMouseDown);
                window.removeEventListener('mousemove', handleMouseMove);
                window.removeEventListener('mouseup', handleMouseUp);
                containerElem.removeEventListener('wheel', handleWheel);
            }
        };
    }

    onMount(() => {
        // Redundant call removed to prevent race conditions with the reactive block below
    });

    $: if (url) {
        pageNum = 1;
        loadDocument();
    }
</script>

<div class="flex flex-col h-full bg-slate-200/50 rounded-lg overflow-hidden border border-slate-300 shadow-inner relative group select-none">
    
    <!-- Toolbar -->
    <div class="flex flex-col sm:flex-row sm:items-center justify-between gap-3 sm:gap-4 px-3 sm:px-4 py-3 bg-slate-800 text-white shadow-md z-20">
        <div class="flex items-center justify-between w-full sm:w-auto gap-4">
            <div class="flex items-center gap-2 sm:gap-3 truncate">
                <div class="bg-white/10 p-1.5 rounded-md shrink-0">
                    <svg xmlns="http://www.w3.org/2000/svg" class="h-3.5 w-3.5 sm:h-4 sm:w-4 text-white/90" viewBox="0 0 20 20" fill="currentColor">
                        <path fill-rule="evenodd" d="M4 4a2 2 0 012-2h4.586A2 2 0 0112 2.586L15.414 6A2 2 0 0116 7.414V16a2 2 0 01-2 2H6a2 2 0 01-2-2V4zm2 6a1 1 0 011-1h6a1 1 0 110 2H7a1 1 0 01-1-1zm1 3a1 1 0 100 2h6a1 1 0 100-2H7z" clip-rule="evenodd" />
                    </svg>
                </div>
                <span class="text-xs sm:text-sm font-medium text-slate-100 truncate tracking-wide" title={filename}>{filename}</span>
            </div>
        </div>
        
        <div class="flex items-center justify-center gap-1.5 sm:gap-2 bg-slate-700/50 rounded-lg p-1 border border-slate-600/50 w-full sm:w-auto overflow-x-auto">
            {#if type === 'pdf'}
                <button class="p-1.5 sm:p-2 hover:bg-slate-600 rounded text-slate-300 hover:text-white transition-colors disabled:opacity-30 shrink-0" on:click={onPrevPage} disabled={pageNum <= 1}>
                    <svg xmlns="http://www.w3.org/2000/svg" class="h-4 w-4" viewBox="0 0 20 20" fill="currentColor"><path fill-rule="evenodd" d="M12.707 5.293a1 1 0 010 1.414L9.414 10l3.293 3.293a1 1 0 01-1.414 1.414l-4-4a1 1 0 010-1.414l4-4a1 1 0 011.414 0z" clip-rule="evenodd" /></svg>
                </button>
                <span class="text-[10px] sm:text-xs font-mono text-slate-200 min-w-[3.5rem] sm:min-w-[4rem] text-center font-semibold select-none shrink-0">{pageNum} / {totalPages}</span>
                <button class="p-1.5 sm:p-2 hover:bg-slate-600 rounded text-slate-300 hover:text-white transition-colors disabled:opacity-30 shrink-0" on:click={onNextPage} disabled={pageNum >= totalPages}>
                    <svg xmlns="http://www.w3.org/2000/svg" class="h-4 w-4" viewBox="0 0 20 20" fill="currentColor"><path fill-rule="evenodd" d="M7.293 14.707a1 1 0 010-1.414L10.586 10 7.293 6.707a1 1 0 011.414-1.414l4 4a1 1 0 010 1.414l-4 4a1 1 0 01-1.414 0z" clip-rule="evenodd" /></svg>
                </button>
                <div class="h-3 sm:h-4 w-px bg-slate-600 mx-1 shrink-0"></div>
                <button class="p-1.5 sm:p-2 hover:bg-slate-600 rounded text-slate-300 hover:text-white transition-colors shrink-0" on:click={onZoomOut}>
                    <svg xmlns="http://www.w3.org/2000/svg" class="h-4 w-4" viewBox="0 0 20 20" fill="currentColor"><path fill-rule="evenodd" d="M5 10a1 1 0 011-1h8a1 1 0 110 2H6a1 1 0 01-1-1z" clip-rule="evenodd" /></svg>
                </button>
                <button class="p-1.5 sm:p-2 hover:bg-slate-600 rounded text-slate-300 hover:text-white transition-colors shrink-0" on:click={onZoomIn}>
                    <svg xmlns="http://www.w3.org/2000/svg" class="h-4 w-4" viewBox="0 0 20 20" fill="currentColor"><path fill-rule="evenodd" d="M10 3a1 1 0 011 1v5h5a1 1 0 110 2h-5v5a1 1 0 11-2 0v-5H4a1 1 0 110-2h5V4a1 1 0 011-1z" clip-rule="evenodd" /></svg>
                </button>
            {/if}
            <div class="h-3 sm:h-4 w-px bg-slate-600 mx-1 shrink-0"></div>
            <button class="flex items-center gap-1 p-1.5 sm:p-2 hover:bg-slate-600 rounded text-amber-400 hover:text-amber-300 transition-colors shrink-0" on:click={resetCssZoom}>
                <span class="text-[10px] font-bold">Reset</span>
            </button>
        </div>

        <div class="hidden sm:flex items-center gap-2">
            <!-- Tombol unduh di sini dihapus karena sudah ada di header utama -->
        </div>
    </div>

    <!-- Content -->
    <div class="flex-1 overflow-hidden relative flex justify-center items-center bg-slate-200/50 touch-none">
        {#if loading}
            <div class="absolute inset-0 flex items-center justify-center bg-white/80 z-20 backdrop-blur-[2px]">
                <div class="flex flex-col items-center gap-3">
                    <div class="w-12 h-12 border-4 border-slate-200 border-t-blue-600 rounded-full animate-spin"></div>
                    <span class="text-sm font-semibold text-slate-600 animate-pulse">Memproses Dokumen...</span>
                </div>
            </div>
        {/if}

        {#if error}
            <div class="flex flex-col items-center justify-center h-full text-slate-400 z-10">
                <p class="text-sm font-medium text-slate-600">{error}</p>
                <button class="mt-4 px-4 py-2 text-xs font-bold text-blue-600 hover:bg-blue-50 rounded-lg" on:click={loadDocument}>Coba Lagi</button>
            </div>
        {:else}
            <div use:panzoom class="relative flex justify-center items-center z-10 transform-origin-center will-change-transform w-full h-full p-4 md:p-8">
                {#if type === 'pdf'}
                    <div class="shadow-2xl shadow-slate-400/20 rounded-sm overflow-hidden bg-white flex max-h-full">
                        <canvas bind:this={canvas} class="block max-w-full max-h-[85vh] object-contain"></canvas>
                    </div>
                {:else if type === 'docx'}
                    <div bind:this={container} class="bg-white shadow-2xl p-8 min-h-[800px] w-full max-w-[800px] docx-wrapper"></div>
                {:else if type === 'image'}
                    <img src={finalUrl} alt={filename} class="max-w-full h-auto shadow-xl rounded-lg" />
                {/if}
            </div>
        {/if}
    </div>
</div>

<style>
    :global(.docx-wrapper) { background: white !important; }
    :global(.docx-wrapper section) { box-shadow: none !important; margin-bottom: 0 !important; }
</style>
