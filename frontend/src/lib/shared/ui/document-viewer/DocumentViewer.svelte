<script>
    import { onMount, onDestroy } from 'svelte';
    import Button from '$lib/shared/ui/button/Button.svelte';
    
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
    let scale = 1.0;
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
        loading = true;
        error = null;
        resetCssZoom();

        try {
            if (type === 'pdf') {
                if (!pdfjsLib) {
                    const mod = await import('pdfjs-dist');
                    pdfjsLib = mod;
                    // Configure PDF Worker - using unpkg as a reliable fallback for pdf.js worker
                    pdfjsLib.GlobalWorkerOptions.workerSrc = `https://unpkg.com/pdfjs-dist@${pdfjsLib.version}/build/pdf.worker.min.mjs`;
                }
                
                const loadingTask = pdfjsLib.getDocument(finalUrl);
                pdfDoc = await loadingTask.promise;
                totalPages = pdfDoc.numPages;
                renderPage(pageNum);
            } else if (type === 'docx') {
                if (!renderAsync) {
                    const mod = await import('docx-preview');
                    renderAsync = mod.renderAsync;
                }

                // docx-preview expects a Blob or ArrayBuffer
                const response = await fetch(finalUrl);
                const blob = await response.blob();
                if (container) {
                    container.innerHTML = ''; // Clear previous
                    await renderAsync(blob, container, container, {
                        className: 'docx-viewer',
                        inWrapper: true,
                        ignoreWidth: false,
                        ignoreHeight: false,
                        ignoreFonts: false,
                        breakPages: true,
                        ignoreLastRenderedPageBreak: true,
                        experimental: false,
                        trimXmlDeclaration: true,
                        useBase64URL: false,
                        renderChanges: false,
                        debug: false,
                    });
                }
            } else if (type === 'image') {
                // Image handling is simple via <img> tag in HTML
            }
        } catch (err) {
            console.error("Error loading document:", err);
            error = "Gagal memuat dokumen. Pastikan file valid.";
        } finally {
            if (type !== 'pdf') loading = false; // PDF loading state handled in renderPage
        }
    }

    async function renderPage(num) {
        if (currentRenderTask) {
            try {
                await currentRenderTask.cancel();
            } catch (err) {
                // Ignore cancel exception
            }
        }
        
        pageRendering = true;
        
        try {
            const page = await pdfDoc.getPage(num);
            const viewport = page.getViewport({ scale });
            
            if (canvas) {
                const ctx = canvas.getContext('2d');
                canvas.height = viewport.height;
                canvas.width = viewport.width;

                const renderContext = {
                    canvasContext: ctx,
                    viewport: viewport
                };
                
                currentRenderTask = page.render(renderContext);

                await currentRenderTask.promise;
                currentRenderTask = null;
                pageRendering = false;
                loading = false;

                if (pageNumPending !== null) {
                    renderPage(pageNumPending);
                    pageNumPending = null;
                }
            }
        } catch (err) {
            if (err.name !== 'RenderingCancelledException') {
                console.error("Page render error:", err);
            }
            pageRendering = false;
        }
    }

    function queueRenderPage(num) {
        if (pageRendering) {
            pageNumPending = num;
        } else {
            renderPage(num);
        }
    }

    function onPrevPage() {
        if (pageNum <= 1) return;
        pageNum--;
        resetCssZoom();
        queueRenderPage(pageNum);
    }

    function onNextPage() {
        if (pageNum >= totalPages) return;
        pageNum++;
        resetCssZoom();
        queueRenderPage(pageNum);
    }

    function onZoomIn() {
        scale += 0.2;
        resetCssZoom();
        queueRenderPage(pageNum);
    }

    function onZoomOut() {
        if (scale <= 0.4) return;
        scale -= 0.2;
        resetCssZoom();
        queueRenderPage(pageNum);
    }

    function resetCssZoom() {
        cssScale = 1.0;
        panX = 0;
        panY = 0;
    }

    // Custom Svelte Action for Pinch-to-Zoom and Panning
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
                // Only prevent default scrolling if we are zoomed in, allowing native scroll otherwise
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

        // Attach listeners to the parent viewport
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
            update(newState) {
                // When triggerState (cssScale, panX, panY) changes from outside (e.g. Reset button)
                updateTransform();
            },
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
        if (url) loadDocument();
    });

    // React to URL changes if the viewer is reused
    $: if (url) {
        // Reset state
        pageNum = 1;
        scale = 1.0;
        loadDocument();
    }
</script>

<div class="flex flex-col h-full bg-slate-200/50 rounded-lg overflow-hidden border border-slate-300 shadow-inner relative group select-none">
    
    <!-- Modern Native-like Toolbar -->
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
            
            <div class="flex items-center gap-2 sm:hidden shrink-0">
                {#if cssScale !== 1.0 || panX !== 0 || panY !== 0}
                    <button class="flex items-center justify-center p-2 bg-amber-500 hover:bg-amber-400 text-white rounded-lg transition-all shadow-sm" on:click={resetCssZoom} title="Reset Pan/Zoom">
                        <svg xmlns="http://www.w3.org/2000/svg" class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 8V4m0 0h4M4 4l5 5m11-1V4m0 0h-4m4 0l-5 5M4 16v4m0 0h4m-4 0l5-5m11 5l-5-5m5 5v-4m0 4h-4" />
                        </svg>
                    </button>
                {/if}
                {#if type === 'image' || type === 'pdf'}
                    <a href={finalUrl} download={filename} class="flex items-center justify-center p-2 bg-blue-600 hover:bg-blue-500 text-white rounded-lg transition-all shadow-sm" title="Download">
                        <svg xmlns="http://www.w3.org/2000/svg" class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 16v1a3 3 0 003 3h10a3 3 0 003-3v-1m-4-4l-4 4m0 0l-4-4m4 4V4" />
                        </svg>
                    </a>
                {/if}
            </div>
        </div>
        
        <div class="flex items-center justify-center gap-1.5 sm:gap-2 bg-slate-700/50 rounded-lg p-1 border border-slate-600/50 w-full sm:w-auto overflow-x-auto">
            {#if type === 'pdf'}
                <button class="p-1.5 sm:p-2 hover:bg-slate-600 rounded text-slate-300 hover:text-white transition-colors disabled:opacity-30 shrink-0" on:click={onPrevPage} disabled={pageNum <= 1} title="Halaman Sebelumnya">
                    <svg xmlns="http://www.w3.org/2000/svg" class="h-4 w-4" viewBox="0 0 20 20" fill="currentColor"><path fill-rule="evenodd" d="M12.707 5.293a1 1 0 010 1.414L9.414 10l3.293 3.293a1 1 0 01-1.414 1.414l-4-4a1 1 0 010-1.414l4-4a1 1 0 011.414 0z" clip-rule="evenodd" /></svg>
                </button>
                <span class="text-[10px] sm:text-xs font-mono text-slate-200 min-w-[3.5rem] sm:min-w-[4rem] text-center font-semibold select-none shrink-0">{pageNum} / {totalPages}</span>
                <button class="p-1.5 sm:p-2 hover:bg-slate-600 rounded text-slate-300 hover:text-white transition-colors disabled:opacity-30 shrink-0" on:click={onNextPage} disabled={pageNum >= totalPages} title="Halaman Selanjutnya">
                    <svg xmlns="http://www.w3.org/2000/svg" class="h-4 w-4" viewBox="0 0 20 20" fill="currentColor"><path fill-rule="evenodd" d="M7.293 14.707a1 1 0 010-1.414L10.586 10 7.293 6.707a1 1 0 011.414-1.414l4 4a1 1 0 010 1.414l-4 4a1 1 0 01-1.414 0z" clip-rule="evenodd" /></svg>
                </button>
                <div class="h-3 sm:h-4 w-px bg-slate-600 mx-1 shrink-0"></div>
                <button class="p-1.5 sm:p-2 hover:bg-slate-600 rounded text-slate-300 hover:text-white transition-colors shrink-0" on:click={onZoomOut} title="Zoom Out (Resolution)">
                    <svg xmlns="http://www.w3.org/2000/svg" class="h-4 w-4" viewBox="0 0 20 20" fill="currentColor"><path fill-rule="evenodd" d="M5 10a1 1 0 011-1h8a1 1 0 110 2H6a1 1 0 01-1-1z" clip-rule="evenodd" /></svg>
                </button>
                <button class="p-1.5 sm:p-2 hover:bg-slate-600 rounded text-slate-300 hover:text-white transition-colors shrink-0" on:click={onZoomIn} title="Zoom In (Resolution)">
                    <svg xmlns="http://www.w3.org/2000/svg" class="h-4 w-4" viewBox="0 0 20 20" fill="currentColor"><path fill-rule="evenodd" d="M10 3a1 1 0 011 1v5h5a1 1 0 110 2h-5v5a1 1 0 11-2 0v-5H4a1 1 0 110-2h5V4a1 1 0 011-1z" clip-rule="evenodd" /></svg>
                </button>
            {/if}
            {#if type === 'image'}
                 <span class="text-[10px] sm:text-xs text-slate-400 px-3 py-1.5 shrink-0">Image Viewer</span>
            {/if}
            {#if type === 'docx'}
                 <span class="text-[10px] sm:text-xs text-slate-400 px-3 py-1.5 shrink-0">Document Preview</span>
            {/if}
            
            {#if cssScale !== 1.0 || panX !== 0 || panY !== 0}
                 <div class="h-3 sm:h-4 w-px bg-slate-600 mx-1 hidden sm:block shrink-0"></div>
                 <button class="hidden sm:flex items-center gap-1 p-1.5 sm:p-2 hover:bg-slate-600 rounded text-amber-400 hover:text-amber-300 transition-colors shrink-0" on:click={resetCssZoom} title="Reset Pan & Zoom">
                    <svg xmlns="http://www.w3.org/2000/svg" class="h-3.5 w-3.5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 8V4m0 0h4M4 4l5 5m11-1V4m0 0h-4m4 0l-5 5M4 16v4m0 0h4m-4 0l5-5m11 5l-5-5m5 5v-4m0 4h-4" />
                    </svg>
                    <span class="text-[10px] font-bold">Reset</span>
                 </button>
            {/if}
        </div>

        <div class="hidden sm:flex items-center gap-2">
            {#if type === 'image' || type === 'pdf'}
                <a href={finalUrl} download={filename} class="flex items-center gap-2 px-3 py-1.5 bg-blue-600 hover:bg-blue-500 text-white rounded text-xs font-bold transition-all shadow-sm hover:shadow" title="Download">
                    <svg xmlns="http://www.w3.org/2000/svg" class="h-3.5 w-3.5" viewBox="0 0 20 20" fill="currentColor"><path fill-rule="evenodd" d="M3 17a1 1 0 011-1h12a1 1 0 110 2H4a1 1 0 01-1-1zm3.293-7.707a1 1 0 011.414 0L9 10.586V3a1 1 0 112 0v7.586l1.293-1.293a1 1 0 111.414 1.414l-3 3a1 1 0 01-1.414 0l-3-3a1 1 0 010-1.414z" clip-rule="evenodd" /></svg>
                    <span>Unduh</span>
                </a>
            {/if}
        </div>
    </div>

    <!-- Viewer Content with Native-like Background -->
    <div class="flex-1 overflow-hidden relative flex justify-center items-center bg-slate-200/50 touch-none">
        <!-- Paper Texture Pattern -->
        <div class="absolute inset-0 opacity-[0.03] pointer-events-none z-0" style="background-image: radial-gradient(#000 1px, transparent 1px); background-size: 20px 20px;"></div>

        {#if loading}
            <div class="absolute inset-0 flex items-center justify-center bg-white/80 z-20 backdrop-blur-[2px]">
                <div class="flex flex-col items-center gap-3">
                    <div class="relative">
                        <div class="w-12 h-12 border-4 border-slate-200 border-t-blue-600 rounded-full animate-spin"></div>
                        <div class="absolute inset-0 flex items-center justify-center">
                            <div class="w-2 h-2 bg-blue-600 rounded-full"></div>
                        </div>
                    </div>
                    <span class="text-sm font-semibold text-slate-600 tracking-wide animate-pulse">Memproses Dokumen...</span>
                </div>
            </div>
        {/if}

        {#if error}
            <div class="flex flex-col items-center justify-center h-full text-slate-400 z-10">
                <div class="bg-white p-6 rounded-full shadow-sm mb-4">
                    <svg xmlns="http://www.w3.org/2000/svg" class="h-10 w-10 text-red-400" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-3L13.732 4c-.77-1.333-2.694-1.333-3.464 0L3.34 16c-.77 1.333.192 3 1.732 3z" />
                    </svg>
                </div>
                <p class="text-sm font-medium text-slate-600">{error}</p>
                <button class="mt-4 px-4 py-2 text-xs font-bold text-blue-600 hover:bg-blue-50 rounded-lg transition-colors" on:click={loadDocument}>Coba Lagi</button>
            </div>
        {:else}
            <!-- Pan/Zoom Wrapper -->
            <div use:panzoom={{cssScale, panX, panY}} class="relative flex justify-center items-center z-10 transform-origin-center will-change-transform max-w-full max-h-full p-4 md:p-8">
                {#if type === 'pdf'}
                    <div class="shadow-2xl shadow-slate-400/20 rounded-sm overflow-hidden ring-1 ring-black/5 bg-white flex max-h-full">
                        <canvas bind:this={canvas} class="block max-w-full max-h-[80vh] object-contain pointer-events-none"></canvas>
                    </div>
                {:else if type === 'docx'}
                    <div bind:this={container} class="bg-white shadow-2xl shadow-slate-400/20 p-8 min-h-[800px] w-full max-w-[800px] docx-wrapper ring-1 ring-black/5"></div>
                {:else if type === 'image'}
                    <img src={finalUrl} alt={filename} class="max-w-full h-auto shadow-xl rounded-lg ring-1 ring-black/5 pointer-events-none" draggable="false" />
                {/if}
            </div>
        {/if}
    </div>
</div>

<style>
    /* docx-preview styling overrides if needed */
    :global(.docx-wrapper) {
        background: white !important;
    }
    :global(.docx-wrapper section) {
        box-shadow: none !important;
        margin-bottom: 0 !important;
    }
</style>
