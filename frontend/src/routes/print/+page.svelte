<script>
    import { page } from '$app/stores';
    import { recordsStore } from '$lib/stores/records';
    import { onMount, tick } from 'svelte';
    import Button from '$lib/components/ui/button/Button.svelte';
    import { terbilang } from '$lib/utils/terbilang';

    let type = $page.url.searchParams.get('type'); // 'spd', 'rincian', 'laporan'
    let spd = $page.url.searchParams.get('spd');
    
    $: record = $recordsStore.find(r => r.spd === spd);

    const defaultCosts = {
        dailyAllowanceDays: 0, dailyAllowanceRate: 0,
        hotelDays: 0, hotelRate: 0,
        ticketGo: 0, ticketBack: 0,
        localTransport: 0, regionalTransport: 0,
        transportMode: '-'
    };
    $: costs = { ...defaultCosts, ...(record?.costs || {}) };

    let isPagedLoaded = false;

    async function initPaged() {
        // @ts-ignore
        if (!record || !window.Paged) return;
        
        // Wait for DOM
        await tick();
        
        const source = document.querySelector('#source-content');
        const target = document.querySelector('#paged-preview');

        if (source && target) {
            target.innerHTML = ''; // Clear previous
            // @ts-ignore
            const previewer = new window.Paged.Previewer();
            
            // Get all stylesheets to pass to Paged.js
            const stylesheets = Array.from(document.styleSheets)
                .map(sheet => sheet.href)
                .filter(href => href);

            await previewer.preview(source.innerHTML, stylesheets, target);
            isPagedLoaded = true;
        }
    }

    onMount(() => {
        // Poll for Paged.js availability (since it's loaded via CDN async)
        const interval = setInterval(() => {
            // @ts-ignore
            if (window.Paged) {
                clearInterval(interval);
                initPaged();
            }
        }, 100);
        
        return () => clearInterval(interval);
    });

    function print() {
        window.print();
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
            margin: 20mm;
        }
    </style>
</svelte:head>

<div class="min-h-screen font-serif text-black">
    <!-- Toolbar -->
    <div class="print-toolbar fixed top-0 left-0 right-0 h-16 bg-white/90 backdrop-blur border-b border-slate-200 z-50 flex items-center justify-between px-4 sm:px-8 shadow-sm print:hidden">
        <Button variant="outline" class="bg-white border-slate-300 hover:bg-slate-50 text-xs sm:text-sm px-3 sm:px-4" on:click={handleBack}>
            <svg xmlns="http://www.w3.org/2000/svg" class="h-4 w-4 sm:mr-2" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M10 19l-7-7m0 0l7-7m-7 7h18" />
            </svg>
            <span class="hidden sm:inline">Kembali</span>
        </Button>
        <div class="flex gap-2 sm:gap-4 items-center">
            <span class="text-xs sm:text-sm font-medium text-slate-500 hidden sm:inline">{isPagedLoaded ? 'Siap Dicetak' : 'Memuat...'}</span>
            <Button on:click={print} disabled={!isPagedLoaded} class="bg-blue-600 hover:bg-blue-700 text-white shadow-sm disabled:opacity-50 text-xs sm:text-sm px-3 sm:px-4">
                <svg xmlns="http://www.w3.org/2000/svg" class="h-4 w-4 sm:mr-2" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M17 17h2a2 2 0 002-2v-4a2 2 0 00-2-2H5a2 2 0 00-2 2v4a2 2 0 002 2h2m2 4h6a2 2 0 002-2v-4a2 2 0 00-2-2H9a2 2 0 00-2 2v4a2 2 0 002 2zm8-12V5a2 2 0 00-2-2H9a2 2 0 00-2 2v4h10z" />
                </svg>
                <span class="hidden sm:inline">Cetak / Simpan PDF</span>
                <span class="sm:hidden">PDF</span>
            </Button>
        </div>
    </div>

    <!-- Hidden Source Content (Vue/Svelte renders here, Paged.js copies from here) -->
    <div id="source-content" class="hidden">
        {#if !record}
            <div class="flex flex-col items-center justify-center h-96 text-slate-400">
                <p class="text-lg font-medium">Data dokumen tidak ditemukan.</p>
            </div>
        {:else}
            <!-- Kop Surat -->
            <div class="flex items-start justify-center gap-6 border-b-[3px] border-black pb-4 mb-8">
                <img src="/kemnaker-ri.webp" alt="Logo Kemnaker" class="h-24 w-auto object-contain" />
                <div class="text-center flex-1 pt-1">
                    <h2 class="text-xl font-bold uppercase tracking-wide leading-tight">Kementerian Ketenagakerjaan<br/>Republik Indonesia</h2>
                    <h3 class="text-base font-bold uppercase mt-1">Biro Protokol dan Hubungan Masyarakat</h3>
                    <p class="text-[11px] mt-1 leading-tight">Jalan Jenderal Gatot Subroto Kaveling 51, Jakarta Selatan 12950</p>
                    <p class="text-[11px] leading-tight">Telepon: (021) 525 5733 • Faksimili: (021) 525 5733</p>
                    <p class="text-[11px] leading-tight font-medium text-blue-900">Laman: www.kemnaker.go.id</p>
                </div>
            </div>

            {#if type === 'spd'}
                <div class="space-y-6">
                    <div class="grid grid-cols-2 text-sm">
                        <div></div>
                        <div class="space-y-1">
                            <div class="flex">
                                <span class="w-24">Lembar ke</span>
                                <span>: 1</span>
                            </div>
                            <div class="flex">
                                <span class="w-24">Kode No</span>
                                <span>: ....................</span>
                            </div>
                            <div class="flex">
                                <span class="w-24">Nomor</span>
                                <span>: {record.spd}</span>
                            </div>
                        </div>
                    </div>

                    <div class="text-center space-y-1 mt-4">
                        <h1 class="text-xl font-bold uppercase underline decoration-2 underline-offset-4">Surat Perjalanan Dinas (SPD)</h1>
                    </div>

                    <div class="space-y-2 mt-8 text-sm leading-relaxed">
                        <div class="flex">
                            <span class="w-8 font-bold">1.</span>
                            <span class="w-1/3">Pejabat Pembuat Komitmen</span>
                            <span class="flex-1">: Kepala Biro Protokol dan Humas</span>
                        </div>
                        <div class="flex">
                            <span class="w-8 font-bold">2.</span>
                            <span class="w-1/3">Nama Pegawai yang diperintah</span>
                            <span class="flex-1 font-semibold uppercase">: {record.employee.name}</span>
                        </div>
                        <div class="flex">
                            <span class="w-8 font-bold">3.</span>
                            <div class="w-1/3 space-y-1">
                                <p>a. Pangkat dan Golongan</p>
                                <p>b. Jabatan/Instansi</p>
                                <p>c. Tingkat Biaya Perjalanan Dinas</p>
                            </div>
                            <div class="flex-1 space-y-1">
                                <p>: {record.employee.pangkat} ({record.employee.golongan})</p>
                                <p>: Staf Protokol / Kementerian Ketenagakerjaan</p>
                                <p>: C</p>
                            </div>
                        </div>
                        <div class="flex">
                            <span class="w-8 font-bold">4.</span>
                            <span class="w-1/3">Maksud Perjalanan Dinas</span>
                            <span class="flex-1 text-justify">: {record.purpose}</span>
                        </div>
                        <div class="flex">
                            <span class="w-8 font-bold">5.</span>
                            <span class="w-1/3">Alat Angkutan yang dipergunakan</span>
                            <span class="flex-1">: {costs.transportMode}</span>
                        </div>
                        <div class="flex">
                            <span class="w-8 font-bold">6.</span>
                            <div class="w-1/3 space-y-1">
                                <p>a. Tempat Berangkat</p>
                                <p>b. Tempat Tujuan</p>
                            </div>
                            <div class="flex-1 space-y-1">
                                <p>: Jakarta</p>
                                <p>: {record.location}, {record.province}</p>
                            </div>
                        </div>
                        <div class="flex">
                            <span class="w-8 font-bold">7.</span>
                            <div class="w-1/3 space-y-1">
                                <p>a. Lama Perjalanan Dinas</p>
                                <p>b. Tanggal Berangkat</p>
                                <p>c. Tanggal Harus Kembali</p>
                            </div>
                            <div class="flex-1 space-y-1">
                                <p>: {Math.ceil((new Date(record.endDate).getTime() - new Date(record.startDate).getTime()) / (1000 * 60 * 60 * 24)) + 1} (....................) hari</p>
                                <p>: {new Date(record.startDate).toLocaleDateString('id-ID', { day: 'numeric', month: 'long', year: 'numeric'})}</p>
                                <p>: {new Date(record.endDate).toLocaleDateString('id-ID', { day: 'numeric', month: 'long', year: 'numeric'})}</p>
                            </div>
                        </div>
                        <div class="flex">
                            <span class="w-8 font-bold">8.</span>
                            <span class="w-1/3">Pembebanan Anggaran</span>
                            <span class="flex-1">: DIPA Biro Protokol dan Humas TA {new Date().getFullYear()}</span>
                        </div>
                        <div class="flex">
                            <span class="w-8 font-bold">9.</span>
                            <span class="w-1/3">Keterangan Lain-lain</span>
                            <span class="flex-1">: -</span>
                        </div>
                    </div>

                    <div class="mt-16 grid grid-cols-2 gap-8 break-inside-avoid">
                        <div></div> <!-- Empty left column -->
                        <div class="text-center space-y-24">
                            <div>
                                <p>Dikeluarkan di: Jakarta</p>
                                <p>Pada tanggal: {new Date().toLocaleDateString('id-ID', { day: 'numeric', month: 'long', year: 'numeric'})}</p>
                                <p class="font-bold mt-4">Pejabat Pembuat Komitmen</p>
                            </div>
                            <div>
                                <p class="font-bold underline uppercase">Nama Pejabat</p>
                                <p>NIP. 19700101 199002 1 001</p>
                            </div>
                        </div>
                    </div>
                </div>

            {:else if type === 'rincian'}
                <div class="space-y-6">
                    <div class="text-center space-y-1 mb-8">
                        <h1 class="text-xl font-bold uppercase underline decoration-2 underline-offset-4">Rincian Biaya Perjalanan Dinas</h1>
                        <p class="text-sm font-medium">Lampiran SPD Nomor: {record.spd}</p>
                    </div>

                    <div class="space-y-1 text-sm mb-6">
                         <div class="flex">
                            <span class="w-48">Tanggal</span>
                            <span>: {new Date(record.startDate).toLocaleDateString('id-ID', { day: 'numeric', month: 'long', year: 'numeric'})}</span>
                         </div>
                    </div>

                    <table class="w-full border-collapse border border-black text-sm">
                        <thead>
                            <tr class="bg-slate-100">
                                <th class="border border-black p-2 text-center w-12">No</th>
                                <th class="border border-black p-2 text-left">Perincian Biaya</th>
                                <th class="border border-black p-2 text-right w-40">Jumlah</th>
                                <th class="border border-black p-2 text-left w-1/3">Keterangan</th>
                            </tr>
                        </thead>
                        <tbody>
                            <tr>
                                <td class="border border-black p-2 text-center align-top">1</td>
                                <td class="border border-black p-2 align-top">
                                    <div class="font-bold mb-1">Biaya Transportasi</div>
                                    <div class="pl-4 text-xs space-y-1">
                                        <div class="flex justify-between">
                                            <span>Tiket Berangkat</span>
                                            <span>{new Intl.NumberFormat('id-ID', { style: 'currency', currency: 'IDR' }).format(costs.ticketGo)}</span>
                                        </div>
                                        <div class="flex justify-between">
                                            <span>Tiket Pulang</span>
                                            <span>{new Intl.NumberFormat('id-ID', { style: 'currency', currency: 'IDR' }).format(costs.ticketBack)}</span>
                                        </div>
                                    </div>
                                </td>
                                <td class="border border-black p-2 text-right align-top font-medium">
                                    {new Intl.NumberFormat('id-ID', { style: 'currency', currency: 'IDR' }).format(Number(costs.ticketGo) + Number(costs.ticketBack))}
                                </td>
                                <td class="border border-black p-2 text-xs align-top">Sesuai bukti tiket terlampir.</td>
                            </tr>
                            <tr>
                                <td class="border border-black p-2 text-center align-top">2</td>
                                <td class="border border-black p-2 align-top">
                                    <div class="font-bold mb-1">Uang Harian</div>
                                    <div class="pl-4 text-xs">
                                        {costs.dailyAllowanceDays} hari x {new Intl.NumberFormat('id-ID').format(costs.dailyAllowanceRate)}
                                    </div>
                                </td>
                                <td class="border border-black p-2 text-right align-top font-medium">
                                    {new Intl.NumberFormat('id-ID', { style: 'currency', currency: 'IDR' }).format(costs.dailyAllowanceDays * costs.dailyAllowanceRate)}
                                </td>
                                <td class="border border-black p-2 text-xs align-top">Uang saku, uang makan, dan transport lokal.</td>
                            </tr>
                            <tr>
                                <td class="border border-black p-2 text-center align-top">3</td>
                                <td class="border border-black p-2 align-top">
                                    <div class="font-bold mb-1">Biaya Penginapan</div>
                                    <div class="pl-4 text-xs">
                                        {costs.hotelDays} malam x {new Intl.NumberFormat('id-ID').format(costs.hotelRate)}
                                    </div>
                                </td>
                                <td class="border border-black p-2 text-right align-top font-medium">
                                    {new Intl.NumberFormat('id-ID', { style: 'currency', currency: 'IDR' }).format(costs.hotelDays * costs.hotelRate)}
                                </td>
                                <td class="border border-black p-2 text-xs align-top">Sesuai bill hotel terlampir (30%).</td>
                            </tr>
                            <tr>
                                <td class="border border-black p-2 text-center align-top">4</td>
                                <td class="border border-black p-2 align-top">
                                    <div class="font-bold mb-1">Lain-lain</div>
                                    <div class="pl-4 text-xs space-y-1">
                                        {#if costs.localTransport > 0}
                                        <div class="flex justify-between">
                                            <span>Transport Lokal</span>
                                            <span>{new Intl.NumberFormat('id-ID', { style: 'currency', currency: 'IDR' }).format(costs.localTransport)}</span>
                                        </div>
                                        {/if}
                                        {#if costs.regionalTransport > 0}
                                        <div class="flex justify-between">
                                            <span>Transport Daerah</span>
                                            <span>{new Intl.NumberFormat('id-ID', { style: 'currency', currency: 'IDR' }).format(costs.regionalTransport)}</span>
                                        </div>
                                        {/if}
                                    </div>
                                </td>
                                <td class="border border-black p-2 text-right align-top font-medium">
                                    {new Intl.NumberFormat('id-ID', { style: 'currency', currency: 'IDR' }).format(Number(costs.localTransport) + Number(costs.regionalTransport))}
                                </td>
                                <td class="border border-black p-2 text-xs align-top">Dukungan transport kegiatan.</td>
                            </tr>
                            <tr class="bg-slate-50">
                                <td class="border border-black p-2 font-bold text-center" colspan="2">Jumlah Total</td>
                                <td class="border border-black p-2 text-right font-bold text-base">
                                    {new Intl.NumberFormat('id-ID', { style: 'currency', currency: 'IDR' }).format(record.totalCost)}
                                </td>
                                <td class="border border-black p-2 bg-slate-50"></td>
                            </tr>
                            <tr>
                                <td class="border border-black p-3" colspan="4">
                                    <span class="font-bold">Terbilang:</span> 
                                    <span class="italic capitalize bg-slate-100 px-2 py-0.5 rounded">{terbilang(record.totalCost)} Rupiah</span>
                                </td>
                            </tr>
                        </tbody>
                    </table>

                    <div class="mt-16 grid grid-cols-2 gap-12 text-center text-sm break-inside-avoid">
                        <div class="space-y-20">
                            <div>
                                <p>Telah dibayar sejumlah</p>
                                <p class="font-bold">{new Intl.NumberFormat('id-ID', { style: 'currency', currency: 'IDR' }).format(record.totalCost)}</p>
                            </div>
                            <div>
                                <p class="font-bold">Bendahara Pengeluaran</p>
                                <br/><br/><br/>
                                <p class="font-bold underline">(......................................)</p>
                                <p>NIP. ..................................</p>
                            </div>
                        </div>
                        <div class="space-y-20">
                            <div>
                                <p>Jakarta, {new Date().toLocaleDateString('id-ID', { day: 'numeric', month: 'long', year: 'numeric'})}</p>
                                <p>Telah menerima jumlah uang sebesar</p>
                                <p class="font-bold">{new Intl.NumberFormat('id-ID', { style: 'currency', currency: 'IDR' }).format(record.totalCost)}</p>
                            </div>
                            <div>
                                <p class="font-bold">Yang Menerima</p>
                                <br/><br/><br/>
                                <p class="font-bold underline">{record.employee.name}</p>
                                <p>NIP. {record.employee.nip}</p>
                            </div>
                        </div>
                    </div>
                    
                    <div class="border-t-2 border-black border-dashed mt-8 pt-4 text-center break-inside-avoid">
                        <p class="font-bold mb-8">PERHITUNGAN RAMPUNG</p>
                        <div class="grid grid-cols-2 gap-12 text-sm">
                             <div class="space-y-20">
                                <div>
                                    <p>Ditetapkan sejumlah</p>
                                    <p class="font-bold">{new Intl.NumberFormat('id-ID', { style: 'currency', currency: 'IDR' }).format(record.totalCost)}</p>
                                </div>
                                <div>
                                    <p class="font-bold">Pejabat Pembuat Komitmen</p>
                                    <br/><br/><br/>
                                    <p class="font-bold underline">(......................................)</p>
                                    <p>NIP. ..................................</p>
                                </div>
                            </div>
                            <div class="space-y-20">
                                <div>
                                    <p>Yang telah dibayar semula</p>
                                    <p class="font-bold">Rp 0</p>
                                </div>
                                <div>
                                    <p>Sisa Kurang / Lebih</p>
                                    <p class="font-bold">{new Intl.NumberFormat('id-ID', { style: 'currency', currency: 'IDR' }).format(record.totalCost)}</p>
                                </div>
                            </div>
                        </div>
                    </div>
                </div>

            {:else if type === 'laporan'}
                <div class="space-y-6 font-sans">
                    <div class="text-center space-y-1 mb-8 border-b pb-4">
                        <h1 class="text-xl font-bold uppercase">Laporan Pelaksanaan Perjalanan Dinas</h1>
                        <p class="text-sm font-normal text-slate-500 mt-1">Berdasarkan SPD Nomor: {record.spd}</p>
                    </div>

                    <div class="grid grid-cols-2 gap-4 mb-6 text-sm bg-slate-50 p-4 rounded-lg border border-slate-200">
                        <div>
                            <span class="font-bold block text-slate-700">Nama Pelaksana:</span>
                            <span class="text-lg">{record.employee.name}</span>
                            <span class="block text-xs text-slate-500">NIP. {record.employee.nip}</span>
                            <span class="block text-xs text-slate-500 mt-1">{record.employee.pangkat}</span>
                        </div>
                        <div class="text-right">
                            <span class="font-bold block text-slate-700">Waktu & Tempat:</span>
                            <span>{new Date(record.startDate).toLocaleDateString('id-ID')} s.d {new Date(record.endDate).toLocaleDateString('id-ID')}</span>
                            <span class="block text-sm font-semibold text-blue-800 mt-1">{record.location}, {record.province}</span>
                        </div>
                    </div>

                    <div class="mb-8">
                        <h3 class="font-bold border-b border-slate-300 mb-2 pb-1 text-lg">I. Hasil Kegiatan</h3>
                        <div class="text-justify whitespace-pre-wrap leading-relaxed text-sm">
                            {record.reportData?.text || 'Belum ada laporan yang disubmit oleh pegawai.'}
                        </div>
                    </div>

                    <div class="mb-8 break-inside-avoid">
                        <h3 class="font-bold border-b border-slate-300 mb-4 pb-1 text-lg">II. Dokumentasi</h3>
                        {#if record.reportData?.files && record.reportData.files.length > 0}
                            <div class="grid grid-cols-2 gap-4">
                                {#each record.reportData.files as file}
                                    {#if file.type.startsWith('image/')}
                                        <div class="border p-2 rounded-lg break-inside-avoid">
                                            <img src={file.data} alt={file.name} class="w-full h-48 object-cover mb-2 rounded bg-slate-100"/>
                                            <p class="text-[10px] text-center text-slate-500">{file.name} - {new Date(file.timestamp).toLocaleDateString()}</p>
                                        </div>
                                    {/if}
                                {/each}
                            </div>
                        {:else}
                            <div class="p-8 text-center bg-slate-50 rounded-lg border border-dashed border-slate-300">
                                <p class="text-sm text-slate-500 italic">Tidak ada dokumentasi terlampir.</p>
                            </div>
                        {/if}
                    </div>

                    <div class="mt-16 flex justify-end text-center break-inside-avoid">
                        <div>
                            <p>Jakarta, {new Date().toLocaleDateString('id-ID', { day: 'numeric', month: 'long', year: 'numeric'})}</p>
                            <p class="font-bold">Pelaksana Perjalanan Dinas</p>
                            <br/><br/><br/>
                            <p class="font-bold underline uppercase">{record.employee.name}</p>
                            <p>NIP. {record.employee.nip}</p>
                        </div>
                    </div>
                </div>
            {/if}
        {/if}
    </div>

    <!-- Paged.js Render Output -->
    <div id="paged-preview" class="pt-20 pb-12"></div>
</div>