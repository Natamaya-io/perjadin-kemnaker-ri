<script>
    import { page } from '$app/stores';
    import { recordsStore, loadRecords } from '$lib/features/pengajuan/store';
    import { onMount, tick } from 'svelte';
    import Button from '$lib/shared/ui/button/Button.svelte';
    import { terbilang } from '$lib/shared/utils/terbilang';
    import { toast } from '$lib/shared/stores/toast';

    let type = $page.url.searchParams.get('type'); // 'spd', 'rincian', 'laporan'
    let spd = $page.url.searchParams.get('spd');
    
    $: allRecordsForSpd = $recordsStore.filter(r => r.spd === spd);
    $: record = allRecordsForSpd.length > 0 ? allRecordsForSpd[0] : null;
    $: allFiles = allRecordsForSpd.flatMap(r => r.reportData?.files || []);

    const defaultCosts = {
        dailyAllowanceDays: 0, dailyAllowanceRate: 0,
        hotelDays: 0, hotelRate: 0,
        ticketGo: 0, ticketBack: 0,
        localTransport: 0, regionalTransport: 0,
        transportMode: '-'
    };
    $: costs = { ...defaultCosts, ...(record?.costs || {}) };

    let isPagedLoaded = false;
    let pagedRendering = false;
    let isDataLoaded = false;
    let loadError = false;

    async function initPaged() {
        // @ts-ignore
        if (!window.Paged || pagedRendering || isPagedLoaded) return;
        
        pagedRendering = true;
        // Wait for DOM to reflect the 'record' state (either found or not found)
        await tick();
        
        const source = document.querySelector('#source-content');
        const target = document.querySelector('#paged-preview');

        if (source && target) {
            target.innerHTML = ''; // Clear previous
            try {
                // @ts-ignore
                const previewer = new window.Paged.Previewer();
                
                // Get all stylesheets to pass to Paged.js
                const stylesheets = Array.from(document.styleSheets)
                    .map(sheet => sheet.href)
                    .filter(href => href);

                await previewer.preview(source.innerHTML, stylesheets, target);
                isPagedLoaded = true;
            } catch (e) {
                console.error('PagedJS error:', e);
                target.innerHTML = '<div class="p-8 text-red-500 font-bold text-center">Gagal memuat pratinjau cetak.</div>';
                isPagedLoaded = true; // Set true so it stops loading state
                loadError = true;
            }
        }
        pagedRendering = false;
    }

    onMount(async () => {
        try {
            if ($recordsStore.length === 0) {
                await loadRecords();
            }
        } catch (e) {
            console.error(e);
        } finally {
            isDataLoaded = true;
        }

        // Poll for Paged.js availability AND data loaded state
        const interval = setInterval(() => {
            // @ts-ignore
            if (window.Paged && isDataLoaded && !isPagedLoaded && !pagedRendering) {
                initPaged();
            }
        }, 100);
        
        return () => clearInterval(interval);
    });

    let isGeneratingPdf = false;

    async function print() {
        // Fallback to browser print if Gotenberg fails or record is missing
        if (!record || loadError) {
            window.print();
            return;
        }

        isGeneratingPdf = true;
        toast.info('Mohon tunggu, sedang memproses dan mencetak PDF resolusi tinggi...', 5000);
        
        try {
            // Clone the entire current document to preserve all styles (Tailwind, Paged.js, etc.)
            const htmlDoc = document.documentElement.cloneNode(true);
            
            // 1. Remove all script tags to prevent any re-hydration or JS execution in Gotenberg
            const scripts = htmlDoc.querySelectorAll('script');
            scripts.forEach(s => s.remove());

            // 2. Isolate the exact Paged.js output
            const pagedPages = htmlDoc.querySelector('.pagedjs_pages');
            if (!pagedPages) throw new Error('Paged.js output not found');
            
            // Reconstruct a clean HTML document string from scratch containing ONLY the Paged.js output
            const styles = Array.from(htmlDoc.querySelectorAll('style, link[rel="stylesheet"]'))
                .map(el => el.outerHTML)
                .join('\n');

            // 3. Inject specific print overrides to ensure Gotenberg's Chromium engine prints the Paged.js DOM perfectly 1:1
            const printOverrides = `
            <link href="https://fonts.googleapis.com/css2?family=Tinos:ital,wght@0,400;0,700;1,400;1,700&display=swap" rel="stylesheet">
            <style>
                @media print {
                    @page { size: A4; margin: 0 !important; }
                    html, body {
                        background: white !important;
                        margin: 0 !important;
                        padding: 0 !important;
                        -webkit-print-color-adjust: exact !important;
                        print-color-adjust: exact !important;
                    }
                    .pagedjs_pages {
                        transform: none !important;
                        width: 100% !important;
                        margin: 0 !important;
                        padding: 0 !important;
                    }
                    .pagedjs_page {
                        margin: 0 !important;
                        border: none !important;
                        box-shadow: none !important;
                        page-break-after: always !important;
                        break-after: page !important;
                    }
                    .print-toolbar { display: none !important; }
                }
            </style>`;

            // Construct the ultimate exact HTML for Gotenberg
            const html = `<!DOCTYPE html>
<html lang="id">
<head>
    <meta charset="UTF-8">
    <title>Cetak Dokumen - ${spd}</title>
    ${styles}
    ${printOverrides}
</head>
<body class="bg-white text-black" style="font-family: 'Times New Roman', Times, serif;">
    ${pagedPages.outerHTML}
</body>
</html>`;
            
            const filename = `Dokumen-${type}-${spd}`;

            const response = await fetch('/export-pdf', {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json'
                },
                body: JSON.stringify({ html, filename })
            });

            if (!response.ok) {
                throw new Error('Gagal menghasilkan PDF dari server');
            }

            const blob = await response.blob();
            const url = window.URL.createObjectURL(blob);
            const a = document.createElement('a');
            a.style.display = 'none';
            a.href = url;
            a.download = `${filename}.pdf`;
            document.body.appendChild(a);
            a.click();
            window.URL.revokeObjectURL(url);
            document.body.removeChild(a);
            
            toast.success('PDF berhasil diunduh!');
        } catch (e) {
            console.error(e);
            toast.error('Gagal mengunduh PDF. Menggunakan print browser bawaan.');
            window.print(); // Fallback
        } finally {
            isGeneratingPdf = false;
        }
    }

    function handleBack() {
        if (window.history.length > 1) {
            window.history.back();
        } else {
            window.close();
            setTimeout(() => {
                window.location.href = '/dashboard/admin/perdin';
            }, 100);
        }
    }
</script>

<svelte:head>
    <title>Cetak {type ? type.toUpperCase() : 'Dokumen'} - {spd || ''}</title>
    <script>
        window.PagedConfig = { auto: false };
    </script>
    <script src="https://unpkg.com/pagedjs/dist/paged.polyfill.js"></script>
    <style>
        /* Paged.js Native-like Styles with Mobile Responsiveness */
        @media screen {
            body {
                background-color: #f1f5f9; /* Slate-100 */
                margin: 0;
                padding: 0;
                overflow-x: hidden; /* Prevent horizontal scroll on body */
            }
            .pagedjs_pages {
                display: flex;
                flex-direction: column;
                align-items: center;
                width: 100%;
                padding: 2rem 0;
                transform-origin: top center;
            }
            .pagedjs_page {
                background-color: white;
                box-shadow: 0 4px 6px -1px rgb(0 0 0 / 0.1), 0 2px 4px -2px rgb(0 0 0 / 0.1);
                margin-bottom: 2rem;
                flex: none;
            }
            
            /* Mobile Scaling for A4 Preview */
            @media (max-width: 768px) {
                .pagedjs_pages {
                    transform: scale(0.6); /* Scale down to 60% on mobile */
                    margin-top: -2rem; /* Adjust for whitespace caused by scaling */
                    margin-bottom: -20rem; /* Adjust for whitespace at bottom */
                    width: 166.66%; /* Compensate width (100 / 0.6) */
                    margin-left: -33.33%; /* Center the scaled content */
                }
                
                #paged-preview {
                    padding-top: 1rem;
                    overflow-x: hidden;
                }
            }
            
            @media (max-width: 480px) {
                .pagedjs_pages {
                    transform: scale(0.45); /* Scale down further for very small screens */
                    width: 222.22%; /* 100 / 0.45 */
                    margin-left: -61.11%;
                }
            }
        }
        
        /* Print Styles */
        @media print {
            body {
                background-color: white;
            }
            .pagedjs_pages {
                display: block;
            }
            .pagedjs_page {
                box-shadow: none;
                margin: 0;
            }
            /* Hide toolbar explicitly */
            .print-toolbar {
                display: none !important;
            }
        }

        /* Page Setup */
        @page {
            size: A4;
            margin: 15mm;
        }

        /* Document Section Breaks - CRITICAL FOR PAGED.JS TO PARSE */
        .document-section {
            break-before: page;
            page-break-before: always;
        }
        .document-section:first-child {
            break-before: auto;
            page-break-before: auto;
        }
    </style>
</svelte:head>

<div class="min-h-screen text-black" style="font-family: 'Times New Roman', Times, serif;">
    <!-- Toolbar -->
    <div class="print-toolbar fixed top-0 left-0 right-0 h-[72px] bg-white/80 backdrop-blur-md border-b border-slate-200/80 z-50 flex items-center justify-between px-4 sm:px-8 shadow-sm transition-all duration-300">
        <div class="flex items-center gap-4">
            <button class="flex items-center justify-center h-10 w-10 rounded-full hover:bg-slate-100 transition-colors text-slate-600" on:click={handleBack} title="Kembali">
                <svg xmlns="http://www.w3.org/2000/svg" class="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M10 19l-7-7m0 0l7-7m-7 7h18" />
                </svg>
            </button>
            <div class="hidden sm:flex flex-col">
                <h2 class="text-sm font-bold text-slate-800 leading-tight">Pratinjau Dokumen</h2>
                <p class="text-xs text-slate-500">{spd || 'Memuat...'}</p>
            </div>
        </div>
        
        <div class="flex items-center gap-3 sm:gap-5">
            {#if isPagedLoaded}
                <div class="hidden sm:flex items-center gap-2 px-3 py-1.5 bg-green-50 text-green-700 rounded-full text-xs font-medium border border-green-200/50">
                    <span class="relative flex h-2 w-2">
                      <span class="animate-ping absolute inline-flex h-full w-full rounded-full bg-green-400 opacity-75"></span>
                      <span class="relative inline-flex rounded-full h-2 w-2 bg-green-500"></span>
                    </span>
                    Siap Diunduh
                </div>
            {:else}
                <div class="hidden sm:flex items-center gap-2 px-3 py-1.5 bg-amber-50 text-amber-700 rounded-full text-xs font-medium border border-amber-200/50">
                    <svg class="animate-spin h-3 w-3 text-amber-600" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24">
                        <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle>
                        <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
                    </svg>
                    Memuat Pratinjau...
                </div>
            {/if}

            <button 
                on:click={print} 
                disabled={!isPagedLoaded || isGeneratingPdf} 
                class="group relative flex items-center justify-center gap-2.5 h-10 px-5 sm:px-6 bg-gradient-to-r from-blue-600 to-indigo-600 hover:from-blue-700 hover:to-indigo-700 text-white rounded-full font-medium text-sm shadow-md hover:shadow-lg disabled:opacity-70 disabled:cursor-not-allowed transition-all duration-200 overflow-hidden"
            >
                {#if isGeneratingPdf}
                    <svg class="animate-spin -ml-1 mr-2 h-4 w-4 text-white" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24">
                        <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle>
                        <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
                    </svg>
                    <span>Memproses PDF...</span>
                {:else}
                    <svg xmlns="http://www.w3.org/2000/svg" class="h-4 w-4 group-hover:-translate-y-0.5 transition-transform duration-200" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 16v1a3 3 0 003 3h10a3 3 0 003-3v-1m-4-4l-4 4m0 0l-4-4m4 4V4" />
                    </svg>
                    <span>Unduh PDF Resmi</span>
                {/if}
            </button>
        </div>
    </div>

    <!-- Hidden Source Content (Vue/Svelte renders here, Paged.js copies from here) -->
    <div id="source-content" class="hidden">
        {#if !record}
            <div class="flex flex-col items-center justify-center h-96 text-slate-400">
                <p class="text-lg font-medium">Data dokumen tidak ditemukan.</p>
            </div>
        {:else}
            <!-- 1. LAPORAN SECTION -->
            {#if type === 'laporan' || type === 'gabungan'}
                <div class="document-section" style="page-break-after: always; break-after: page;">
                    <!-- Kop Surat Laporan -->
                    <div class="flex items-center justify-center gap-4 border-b-[3px] border-black pb-4 mb-8">
                        <div class="flex flex-col items-center">
                            <img src="/kemnaker-ri.webp" alt="Logo Kemnaker" class="h-20 w-auto object-contain" />
                            <span class="text-[10px] font-bold mt-1 tracking-wider text-[#1e3a5f]">KEMNAKER</span>
                        </div>
                        
                        <!-- Vertical Divider Line -->
                        <div class="w-[2.5px] h-24 bg-[#1e3a5f]"></div>
                        
                        <div class="flex-1 pt-1">
                            <h2 class="text-[22px] font-bold uppercase tracking-tight leading-none text-black">KEMENTERIAN KETENAGAKERJAAN REPUBLIK INDONESIA</h2>
                            <h3 class="text-[26px] font-bold uppercase tracking-tight leading-tight text-[#005792] mb-1">SEKRETARIAT JENDERAL</h3>
                            <p class="text-[11px] font-medium leading-tight text-black italic">Jalan Jenderal Gatot Subroto Kaveling 51, Kelurahan Kuningan Timur, Kecamatan Setiabudi, Kota Jakarta Selatan, Provinsi DKI Jakarta 12950</p>
                            
                            <div class="flex items-center gap-6 mt-1">
                                <div class="flex items-center gap-1.5">
                                    <div class="w-4 h-4 rounded-full bg-[#005792] flex items-center justify-center text-white">
                                        <svg xmlns="http://www.w3.org/2000/svg" width="10" height="10" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="3" stroke-linecap="round" stroke-linejoin="round"><circle cx="12" cy="12" r="10"/><line x1="2" y1="12" x2="22" y2="12"/><path d="M12 2a15.3 15.3 0 0 1 4 10 15.3 15.3 0 0 1-4 10 15.3 15.3 0 0 1-4-10 15.3 15.3 0 0 1 4-10z"/></svg>
                                    </div>
                                    <span class="text-[11px] font-medium text-black">www.kemnaker.go.id</span>
                                </div>
                                <div class="flex items-center gap-1.5">
                                    <div class="w-4 h-4 rounded-full bg-[#005792] flex items-center justify-center text-white">
                                        <svg xmlns="http://www.w3.org/2000/svg" width="10" height="10" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="3" stroke-linecap="round" stroke-linejoin="round"><path d="m22 2-7 20-4-9-9-4Z"/><path d="M22 2 11 13"/></svg>
                                    </div>
                                    <span class="text-[11px] font-medium text-black">persuratan@kemnaker.go.id</span>
                                </div>
                                <div class="flex items-center gap-1.5">
                                    <div class="w-4 h-4 rounded-full bg-[#005792] flex items-center justify-center text-white">
                                        <svg xmlns="http://www.w3.org/2000/svg" width="10" height="10" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="3" stroke-linecap="round" stroke-linejoin="round"><path d="M22 16.92v3a2 2 0 0 1-2.18 2 19.79 19.79 0 0 1-8.63-3.07 19.5 19.5 0 0 1-6-6 19.79 19.79 0 0 1-3.07-8.67A2 2 0 0 1 4.11 2h3a2 2 0 0 1 2 1.72 12.84 12.84 0 0 0 .7 2.81 2 2 0 0 1-.45 2.11L8.09 9.91a16 16 0 0 0 6 6l1.27-1.27a2 2 0 0 1 2.11-.45 12.84 12.84 0 0 0 2.81.7A2 2 0 0 1 22 16.92z"/></svg>
                                    </div>
                                    <span class="text-[11px] font-medium text-black">1500630</span>
                                </div>
                            </div>
                        </div>
                    </div>

                    <div class="space-y-6">
                        <div class="text-center space-y-1 mb-8">
                            <h1 class="text-lg font-bold uppercase">LAPORAN</h1>
                            <p class="text-base font-normal">Perjalanan Dinas ke {record.location}, {record.province}</p>
                            <p class="text-base font-normal">{new Date(record.startDate).toLocaleDateString('id-ID', { day: 'numeric', month: 'long', year: 'numeric'})} - {new Date(record.endDate).toLocaleDateString('id-ID', { day: 'numeric', month: 'long', year: 'numeric'})}</p>
                        </div>

                        <div class="mb-4">
                            <h3 class="font-bold mb-2 text-base">I. PENDAHULUAN</h3>
                            <div class="flex text-sm leading-relaxed text-justify">
                                <div class="w-24 shrink-0">Dasar</div>
                                <div class="w-4 shrink-0">:</div>
                                <div class="flex-1">
                                    <ol class="list-decimal pl-4 space-y-1">
                                        <li>Surat Tugas Kepala Biro Umum Sekretariat Jenderal Kementerian Ketenagakerjaan Nomor {record.suratTugasNumber || '1/ /UM.06.00/...../2026'} Tanggal {new Date(record.startDate).toLocaleDateString('id-ID', { month: 'long'})} {new Date().getFullYear()} dalam rangka {record.purpose};</li>
                                        <li>Peraturan Menteri Keuangan No. 119 Tahun 2023 tentang Perubahan atas PMK No. 113/PMK.05/2012 tentang Perjalanan Dinas dalam negeri bagi Pejabat Negara, Pegawai Negeri dan Pegawai Tidak Tetap;</li>
                                        <li>Peraturan Menteri Ketenagakerjaan Nomor 20 Tahun 2024 tanggal 30 Desember 2024 tentang Organisasi dan Tata Kerja Kementerian Ketenagakerjaan;</li>
                                        <li>Keputusan Menteri Ketenagakerjaan Nomor 460 tanggal 23 Desember 2025, tentang Pengangkatan Pejabat Perbendaharaan Negara selaku Kuasa Pengguna Anggaran (KPA)/Kuasa Pengguna Barang (KPB) Daftar Isian Pelaksanaan Anggaran (DIPA) Bidang Ketenagakerjaan pada Kantor Pusat Kementerian Ketenagakerjaan;</li>
                                        <li>Surat Keputusan Kuasa Pengguna Anggaran (KPA) /Kuasa Pengguna Barang (KPB) Sekretariat Jenderal NOMOR 1/02/KU.04/I/2026 tanggal 2 Januari 2026 tentang Pejabat Pembuat Komitmen DIPA Unit Sekretariat Jenderal Kementerian Ketenagakerjaan Tahun 2026;</li>
                                        <li>DIPA Unit Sekretariat Jenderal Kementerian Ketenagakerjaan Tahun 2026, NOMOR : SP DIPA-026.01.1.450938/2026 tanggal 1 Desember 2025.</li>
                                    </ol>
                                </div>
                            </div>
                        </div>

                        <div class="mb-4">
                            <h3 class="font-bold mb-2 text-base">II. ISI LAPORAN</h3>
                            <div class="text-justify whitespace-pre-wrap leading-relaxed text-sm px-8">
                                {record.reportData?.text || 'Belum ada laporan yang disubmit oleh pegawai.'}
                            </div>
                        </div>

                        <div class="mb-8">
                            <h3 class="font-bold mb-2 text-base">III. PENUTUP</h3>
                            <div class="text-justify leading-relaxed text-sm px-8">
                                Demikian laporan ini dibuat untuk digunakan sebagaimana mestinya.
                            </div>
                        </div>

                        <div class="mt-8 text-sm break-inside-avoid px-8">
                            <div class="flex justify-end mb-8">
                                <p>Jakarta, {record.reportData?.submittedAt ? new Date(record.reportData.submittedAt).toLocaleDateString('id-ID', { day: 'numeric', month: 'long', year: 'numeric'}) : new Date().toLocaleDateString('id-ID', { day: 'numeric', month: 'long', year: 'numeric'})}</p>
                            </div>

                            <p class="font-bold mb-6 underline">Pelaksana Perjalanan Dinas</p>

                            <table class="w-full">
                                <tbody>
                                    {#each allRecordsForSpd as empRecord, i}
                                        <tr>
                                            <td class="w-12 align-top py-6 text-right pr-4">{i + 1}.</td>
                                            <td class="align-top py-6 w-1/2">
                                                <span class="block">{empRecord.employee?.name || '-'}</span>
                                                <span class="block">NIP. {empRecord.employee?.nip || '-'}</span>
                                            </td>
                                            <td class="align-bottom py-6">
                                                {i + 1}........................................
                                            </td>
                                        </tr>
                                    {/each}
                                </tbody>
                            </table>
                        </div>
                        </div>
                        </div>
                {#if allFiles.length > 0}
                    <div class="document-section" style="page-break-before: always; break-before: page;">
                        <div class="text-center space-y-1 mb-8">
                            <h1 class="text-lg font-bold uppercase">DOKUMENTASI</h1>
                        </div>
                        <div class="flex flex-col items-center gap-8">
                            {#each allFiles as file, i}
                                {#if file.type.startsWith('image/')}
                                    <img src={file.data} alt="Dokumentasi {i + 1}" class="max-w-[80%] max-h-[500px] w-auto h-auto object-contain break-inside-avoid mb-8" />
                                {/if}
                            {/each}
                        </div>
                    </div>
                {/if}

                <!-- Page break handled by document-section -->
            {/if}

            <!-- 2. RINCIAN BIAYA SECTION -->
            {#if type === 'rincian' || type === 'laporan'}
                <div class="document-section" style="page-break-after: always; break-after: page;">
                    <div class="space-y-6 px-4 py-8">
                        <div class="text-center mb-8">
                            <h1 class="text-xl font-bold uppercase">RINCIAN BIAYA PERJALANAN DINAS</h1>
                        </div>

                        <div class="space-y-1 text-sm mb-6">
                            <div class="flex">
                                <span class="w-48">Lampiran SPPD Nomor</span>
                                <span>: {record.spd}</span>
                            </div>
                            <div class="flex">
                                <span class="w-48">Tanggal</span>
                                <span>: {new Date().toLocaleDateString('id-ID', { month: 'long', year: 'numeric'})}</span>
                            </div>
                        </div>

                        <table class="w-full border-collapse border-2 border-black text-sm">
                            <thead>
                                <tr>
                                    <th class="border border-black p-2 text-center w-10 font-bold">NO</th>
                                    <th class="border border-black p-2 text-center font-bold">PERINCIAN BIAYA</th>
                                    <th class="border border-black p-2 text-center w-48 font-bold">JUMLAH</th>
                                    <th class="border border-black p-2 text-center w-[30%] font-bold">KETERANGAN</th>
                                </tr>
                            </thead>
                            <tbody>
                                <tr>
                                    <td class="border-l border-r border-black p-2 text-center font-bold align-top">A</td>
                                    <td class="border-l border-r border-black p-2 font-bold align-top">Transport PP</td>
                                    <td class="border-l border-r border-black p-2 align-top"></td>
                                    <td class="border-l border-r border-black p-2 align-top" rowspan="5">
                                        Biaya Perjalanan Dinas dalam rangka {record.purpose} kunjungan kerja {record.employee.name} ke {record.location}, Provinsi {record.province} selama {Math.ceil((new Date(record.endDate).getTime() - new Date(record.startDate).getTime()) / (1000 * 60 * 60 * 24)) + 1} ({terbilang(Math.ceil((new Date(record.endDate).getTime() - new Date(record.startDate).getTime()) / (1000 * 60 * 60 * 24)) + 1)}) Hari pada tanggal {new Date(record.startDate).toLocaleDateString('id-ID', { day: 'numeric', month: 'long', year: 'numeric'})} - {new Date(record.endDate).toLocaleDateString('id-ID', { day: 'numeric', month: 'long', year: 'numeric'})};
                                    </td>
                                </tr>
                                <tr>
                                    <td class="border-l border-r border-black p-2 text-center align-top">1</td>
                                    <td class="border-l border-r border-black p-2 align-top">
                                        <div class="space-y-1">
                                            <div>- Tiket Pesawat/Kereta/Bus</div>
                                            <div>- Transport Lokal Jakarta</div>
                                            <div>- Transport Daerah</div>
                                        </div>
                                    </td>
                                    <td class="border-l border-r border-black p-2 align-top">
                                        <div class="space-y-1 flex flex-col h-full">
                                            <div class="flex justify-between">
                                                <span>Rp</span>
                                                <span>{new Intl.NumberFormat('id-ID').format(Number(costs.ticketGo) + Number(costs.ticketBack))}</span>
                                            </div>
                                            <div class="flex justify-between">
                                                <span>Rp</span>
                                                <span>{new Intl.NumberFormat('id-ID').format(costs.localTransport)}</span>
                                            </div>
                                            <div class="flex justify-between">
                                                <span>Rp</span>
                                                <span>{new Intl.NumberFormat('id-ID').format(costs.regionalTransport)}</span>
                                            </div>
                                        </div>
                                    </td>
                                </tr>
                                <tr>
                                    <td class="border-l border-r border-black p-2 text-center align-top">2</td>
                                    <td class="border-l border-r border-black p-2 align-top">
                                        <div class="space-y-1">
                                            <div>- Uang Harian</div>
                                            <div class="pl-3">{costs.dailyAllowanceDays} hari x Rp {new Intl.NumberFormat('id-ID').format(costs.dailyAllowanceRate)}</div>
                                        </div>
                                    </td>
                                    <td class="border-l border-r border-black p-2 align-bottom pb-3">
                                        <div class="flex justify-between">
                                            <span>Rp</span>
                                            <span>{new Intl.NumberFormat('id-ID').format(costs.dailyAllowanceDays * costs.dailyAllowanceRate)}</span>
                                        </div>
                                    </td>
                                </tr>
                                <tr>
                                    <td class="border-l border-r border-black p-2 text-center align-top"></td>
                                    <td class="border-l border-r border-black p-2 align-top">
                                        <div class="space-y-1">
                                            <div>- Penginapan</div>
                                            <div class="pl-3">{costs.hotelDays} hari x Rp {new Intl.NumberFormat('id-ID').format(costs.hotelRate)}</div>
                                        </div>
                                    </td>
                                    <td class="border-l border-r border-black p-2 align-bottom pb-3">
                                        <div class="flex justify-between">
                                            <span>Rp</span>
                                            <span>{new Intl.NumberFormat('id-ID').format(costs.hotelDays * costs.hotelRate)}</span>
                                        </div>
                                    </td>
                                </tr>
                                <tr>
                                    <td class="border-l border-r border-black p-2 text-center align-top"></td>
                                    <td class="border-l border-r border-black p-2 align-top pb-8"></td>
                                    <td class="border-l border-r border-black p-2 align-top"></td>
                                </tr>
                                <tr>
                                    <td class="border border-black p-2 font-bold text-center" colspan="2">JUMLAH :</td>
                                    <td class="border border-black p-2 font-bold">
                                        <div class="flex justify-between">
                                            <span>Rp</span>
                                            <span>{new Intl.NumberFormat('id-ID').format(record.totalCost)}</span>
                                        </div>
                                    </td>
                                    <td class="border border-black p-2 bg-slate-50"></td>
                                </tr>
                                <tr>
                                    <td class="border border-black p-3" colspan="4">
                                        <div class="flex gap-4">
                                            <span class="font-bold whitespace-nowrap">Terbilang :</span> 
                                            <span class="uppercase font-bold">{terbilang(record.totalCost)} RUPIAH</span>
                                        </div>
                                    </td>
                                </tr>
                            </tbody>
                        </table>

                        <div class="mt-8 grid grid-cols-2 gap-12 text-sm break-inside-avoid">
                            <div class="space-y-24">
                                <div>
                                    <p>Telah dibayar sejumlah</p>
                                    <p>Rp. <span class="underline decoration-black underline-offset-4">{new Intl.NumberFormat('id-ID').format(record.totalCost)}</span></p>
                                    <p class="font-bold mt-4 text-center">Bendahara Pengeluaran Pembantu<br/>Biro Umum Tahun {new Date().getFullYear()}</p>
                                </div>
                                <div class="text-center">
                                    <p class="font-bold underline decoration-black">Ardimas Perdana Putra</p>
                                    <p>NIP. 19970722 201812 1 002</p>
                                </div>
                            </div>
                            <div class="space-y-24">
                                <div class="text-center">
                                    <p>Jakarta, <span class="inline-block w-40 border-b border-black"></span></p>
                                    <p class="mt-4">Telah menerima jumlah uang sebesar</p>
                                    <p>Rp. <span class="underline decoration-black underline-offset-4">{new Intl.NumberFormat('id-ID').format(record.totalCost)}</span></p>
                                    <p class="font-bold mt-4">Yang menerima,</p>
                                </div>
                                <div class="text-center">
                                    <p class="font-bold underline decoration-black uppercase">{record.employee.name}</p>
                                    <p>NIP. {record.employee.nip || '-'}</p>
                                </div>
                            </div>
                        </div>
                        
                        <div class="border-t-2 border-black mt-8 pt-4 break-inside-avoid text-sm">
                            <p class="mb-4">PERHITUNGAN SPPD RAMPUNG</p>
                            <div class="grid grid-cols-2 gap-12">
                                 <div>
                                    <div class="flex">
                                        <span class="w-48">Ditetapkan sejumlah</span>
                                        <span>: Rp {new Intl.NumberFormat('id-ID').format(record.totalCost)} ,-</span>
                                    </div>
                                    <div class="flex">
                                        <span class="w-48">Yang telah dibayar semula</span>
                                        <span>: Rp 0 ,-</span>
                                    </div>
                                    <div class="flex">
                                        <span class="w-48">Sisa kurang/lebih</span>
                                        <span>: Rp {new Intl.NumberFormat('id-ID').format(record.totalCost)} ,-</span>
                                    </div>
                                </div>
                                <div class="text-center space-y-24 mt-[-1.5rem]">
                                    <div>
                                        <p class="font-bold">Pejabat Pembuat Komitmen<br/>Biro Umum Sekretariat Jenderal</p>
                                    </div>
                                    <div>
                                        <p class="font-bold underline decoration-black">Arief Hafidiyanto</p>
                                        <p>NIP. 19720827 200312 1 002</p>
                                    </div>
                                </div>
                            </div>
                        </div>
                    </div>
                </div>

                <!-- Page break handled by document-section -->
            {/if}

            <!-- 3. SURAT PERJALANAN DINAS (SPD) SECTION -->
            {#if type === 'spd' || type === 'laporan'}
                <div class="document-section px-8 py-10" style="page-break-after: always; break-after: page;">
                    <div class="flex justify-between items-start mb-6">
                        <div class="font-bold text-sm">SEKRETARIAT JENDERAL</div>
                        <div class="text-sm">
                            <table class="w-full">
                                <tbody>
                                    <tr>
                                        <td class="w-24 font-bold">Lembar Ke</td>
                                        <td class="w-4">:</td>
                                        <td></td>
                                    </tr>
                                    <tr>
                                        <td class="w-24 font-bold">Kode Nomor</td>
                                        <td class="w-4">:</td>
                                        <td></td>
                                    </tr>
                                    <tr>
                                        <td class="w-24 font-bold tracking-widest">Nomor</td>
                                        <td class="w-4">:</td>
                                        <td>{record.spd}</td>
                                    </tr>
                                </tbody>
                            </table>
                        </div>
                    </div>

                    <div class="text-center space-y-1 mb-6">
                        <h1 class="text-base font-bold uppercase underline decoration-2 underline-offset-4">SURAT PERJALANAN DINAS (SPD)</h1>
                    </div>

                    <table class="w-full border-collapse border-2 border-black text-sm">
                        <tbody>
                            <tr>
                                <td class="border-b border-r border-black p-2 text-center w-8 align-top">1.</td>
                                <td class="border-b border-r border-black p-2 align-top w-[35%]">Pejabat berwenang yang memberi perintah</td>
                                <td class="border-b border-black p-2 align-top">KPA Biro Umum Sekretariat Jenderal Kemnaker</td>
                            </tr>
                            <tr>
                                <td class="border-b border-r border-black p-2 text-center w-8 align-top">2.</td>
                                <td class="border-b border-r border-black p-2 align-top">Nama /NIP pegawai yang diperintahkan</td>
                                <td class="border-b border-black p-2 align-top">
                                    <span class="inline-block min-w-[200px]">{record.employee.name}</span> / NIP. {record.employee.nip || '-'}
                                </td>
                            </tr>
                            <tr>
                                <td class="border-b border-r border-black p-2 text-center w-8 align-top">3.</td>
                                <td class="border-b border-r border-black p-2 align-top">
                                    <div class="space-y-1">
                                        <div class="flex"><span class="w-4">a.</span> Pangkat dan Golongan</div>
                                        <div class="flex"><span class="w-4">b.</span> Jabatan / Instansi</div>
                                        <div class="flex"><span class="w-4">c.</span> Tingkat Biaya Perjalanan Dinas</div>
                                    </div>
                                </td>
                                <td class="border-b border-black p-2 align-top">
                                    <div class="space-y-1">
                                        <div class="flex"><span class="w-6">a.</span> {#if record.employee.pangkat && record.employee.pangkat !== '-' && record.employee.golongan && record.employee.golongan !== '-'}{record.employee.pangkat} ({record.employee.golongan}){:else}{record.employee.jabatan || '-'}{/if}</div>
                                        <div class="flex"><span class="w-6">b.</span> {record.employee.jabatan || 'Staf Protokol'} / Kementerian Ketenagakerjaan</div>
                                        <div class="flex"><span class="w-6">c.</span> {record.employee.tingkatBiaya || 'C'}</div>
                                    </div>
                                </td>
                            </tr>
                            <tr>
                                <td class="border-b border-r border-black p-2 text-center w-8 align-top">4</td>
                                <td class="border-b border-r border-black p-2 align-top">
                                    <div class="flex"><span class="w-4">a.</span> Maksud Perjalanan Dinas</div>
                                </td>
                                <td class="border-b border-black p-2 align-top">
                                    Biaya Perjalanan Dinas dalam rangka {record.purpose} kunjungan kerja {record.employee.name} di {record.location}, Provinsi {record.province};
                                </td>
                            </tr>
                            <tr>
                                <td class="border-b border-r border-black p-2 text-center w-8 align-top">5.</td>
                                <td class="border-b border-r border-black p-2 align-top">Alat Angkutan yang dipergunakan</td>
                                <td class="border-b border-black p-2 align-top">{costs.transportMode}</td>
                            </tr>
                            <tr>
                                <td class="border-b border-r border-black p-2 text-center w-8 align-top">6.</td>
                                <td class="border-b border-r border-black p-2 align-top">
                                    <div class="space-y-1">
                                        <div class="flex"><span class="w-4">a.</span> Tempat berangkat</div>
                                        <div class="flex"><span class="w-4">b.</span> Tempat tujuan</div>
                                    </div>
                                </td>
                                <td class="border-b border-black p-2 align-top">
                                    <div class="space-y-1">
                                        <div class="flex"><span class="w-6">a.</span> Jakarta</div>
                                        <div class="flex"><span class="w-6">b.</span> {record.location}</div>
                                    </div>
                                </td>
                            </tr>
                            <tr>
                                <td class="border-b border-r border-black p-2 text-center w-8 align-top">7.</td>
                                <td class="border-b border-r border-black p-2 align-top">
                                    <div class="space-y-1">
                                        <div class="flex"><span class="w-4">a.</span> Lamanya Perjalanan Dinas</div>
                                        <div class="flex"><span class="w-4">b.</span> Tanggal berangkat</div>
                                        <div class="flex"><span class="w-4">c.</span> Tanggal harus kembali</div>
                                    </div>
                                </td>
                                <td class="border-b border-black p-2 align-top">
                                    <div class="space-y-1">
                                        <div class="flex"><span class="w-6">a.</span> {Math.ceil((new Date(record.endDate).getTime() - new Date(record.startDate).getTime()) / (1000 * 60 * 60 * 24)) + 1} ({terbilang(Math.ceil((new Date(record.endDate).getTime() - new Date(record.startDate).getTime()) / (1000 * 60 * 60 * 24)) + 1)}) Hari</div>
                                        <div class="flex"><span class="w-6">b.</span> {new Date(record.startDate).toLocaleDateString('id-ID', { day: 'numeric', month: 'long', year: 'numeric'})}</div>
                                        <div class="flex"><span class="w-6">c.</span> {new Date(record.endDate).toLocaleDateString('id-ID', { day: 'numeric', month: 'long', year: 'numeric'})}</div>
                                    </div>
                                </td>
                            </tr>
                            <tr>
                                <td class="border-b border-r border-black p-2 text-center w-8 align-top">8.</td>
                                <td class="border-b border-r border-black p-2 align-top">Pengikut : Nama</td>
                                <td class="border-b border-black p-2 align-top text-center">Keterangan</td>
                            </tr>
                            <tr>
                                <td class="border-b border-r border-black p-2 text-center w-8 align-top"></td>
                                <td class="border-b border-r border-black p-2 align-top h-16">
                                    <div class="space-y-1">
                                        <div>1.</div>
                                        <div>2.</div>
                                        <div>3.</div>
                                    </div>
                                </td>
                                <td class="border-b border-black p-2 align-top"></td>
                            </tr>
                            <tr>
                                <td class="border-b border-r border-black p-2 text-center w-8 align-top">9.</td>
                                <td class="border-b border-r border-black p-2 align-top">
                                    <div class="space-y-1">
                                        <div>Pembebanan Anggaran</div>
                                        <div class="flex"><span class="w-4">a.</span> Instansi</div>
                                        <div class="flex"><span class="w-4">b.</span> Mata Anggaran</div>
                                    </div>
                                </td>
                                <td class="border-b border-black p-2 align-bottom">
                                    <div class="space-y-1">
                                        <div class="flex"><span class="w-6">a.</span> KEMENTERIAN KETENAGAKERJAAN R.I.</div>
                                        <div class="flex"><span class="w-6">b.</span> 026.01.WA.2158.EBA.994.002  S  524111</div>
                                    </div>
                                </td>
                            </tr>
                            <tr>
                                <td class="border-r border-black p-2 text-center w-8 align-top">10.</td>
                                <td class="border-r border-black p-2 align-top h-12">Keterangan</td>
                                <td class="border-black p-2 align-top"></td>
                            </tr>
                        </tbody>
                    </table>

                    <div class="mt-8 grid grid-cols-2 gap-8 text-sm break-inside-avoid">
                        <div></div> <!-- Empty left column -->
                        <div class="flex flex-col">
                            <table class="w-full mb-4">
                                <tbody>
                                    <tr>
                                        <td class="w-32">Dikeluarkan di</td>
                                        <td class="w-4">:</td>
                                        <td>Jakarta</td>
                                    </tr>
                                    <tr>
                                        <td class="w-32">Pada Tanggal</td>
                                        <td class="w-4">:</td>
                                        <td></td>
                                    </tr>
                                </tbody>
                            </table>
                            <div class="text-center space-y-24">
                                <div>
                                    <p class="font-bold">Pejabat Pembuat Komitmen<br/>Biro Umum Sekretariat Jenderal</p>
                                </div>
                                <div>
                                    <p class="font-bold underline decoration-black">Arief Hafidiyanto</p>
                                    <p>NIP. 19720827 200312 1 002</p>
                                </div>
                            </div>
                        </div>
                    </div>
                </div>
            {/if}
        {/if}
    </div>

    <!-- Paged.js Render Output -->
    <div id="paged-preview" class="pt-20 pb-12"></div>
</div>