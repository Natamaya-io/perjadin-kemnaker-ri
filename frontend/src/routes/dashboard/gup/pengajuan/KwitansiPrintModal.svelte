<script lang="ts">
    import { createEventDispatcher } from 'svelte';
    import { fade, scale } from 'svelte/transition';
    import Button from '$lib/shared/ui/button/Button.svelte';

    export let isOpen = false;
    export let data: any = null;

    const dispatch = createEventDispatcher();

    // Default names (in real app, this might come from settings)
    let ppkName = 'Arief Hadfiyanto';
    let ppkNip = '19720827 200312 1 002';
    let bendaharaName = 'Ardimas Perdana Putra';
    let bendaharaNip = '19970722 201812 1 002';

    // Editable reactive values
    let printPpkName = ppkName;
    let printPpkNip = ppkNip;
    let printBendaharaName = bendaharaName;
    let printBendaharaNip = bendaharaNip;
    let printPumName = data?.pum || 'Superadmin';
    let printPumNip = '';

    $: if (data) {
        printPumName = data.pum || 'Superadmin';
    }

    function close() {
        isOpen = false;
        dispatch('close');
    }

    function handlePrint() {
        // Hide normal content, show print content via body class
        document.body.classList.add('print-kwitansi');
        window.print();
        setTimeout(() => {
            document.body.classList.remove('print-kwitansi');
        }, 500);
    }

    function formatCurrency(amount: number) {
        return amount.toLocaleString('id-ID');
    }

    function getYear(dateStr: string) {
        if (!dateStr) return new Date().getFullYear();
        return new Date(dateStr).getFullYear();
    }

    function formatBulanTahun(dateStr: string) {
        if (!dateStr) return '-';
        return new Date(dateStr).toLocaleDateString('id-ID', { month: 'long', year: 'numeric' });
    }

    function terbilang(value: number): string {
        const number = Math.floor(Math.abs(Number(value || 0)));
        const units = ['', 'Satu', 'Dua', 'Tiga', 'Empat', 'Lima', 'Enam', 'Tujuh', 'Delapan', 'Sembilan', 'Sepuluh', 'Sebelas'];

        function convert(n: number): string {
            if (n < 12) return units[n];
            if (n < 20) return `${convert(n - 10)} Belas`;
            if (n < 100) return `${convert(Math.floor(n / 10))} Puluh ${convert(n % 10)}`;
            if (n < 200) return `Seratus ${convert(n - 100)}`;
            if (n < 1000) return `${convert(Math.floor(n / 100))} Ratus ${convert(n % 100)}`;
            if (n < 2000) return `Seribu ${convert(n - 1000)}`;
            if (n < 1000000) return `${convert(Math.floor(n / 1000))} Ribu ${convert(n % 1000)}`;
            if (n < 1000000000) return `${convert(Math.floor(n / 1000000))} Juta ${convert(n % 1000000)}`;
            if (n < 1000000000000) return `${convert(Math.floor(n / 1000000000))} Miliar ${convert(n % 1000000000)}`;
            if (n < 1000000000000000) return `${convert(Math.floor(n / 1000000000000))} Triliun ${convert(n % 1000000000000)}`;
            return String(n);
        }

        if (number === 0) return 'Nol Rupiah';
        return `${convert(number).replace(/\s+/g, ' ').trim()} Rupiah`;
    }

    function formatBkuNumber(businessId: string) {
        if (!businessId) return '-';
        // Extract all digits (e.g., from 'gup_001' or 'gup_1005')
        const digits = String(businessId).replace(/\D/g, '');
        if (!digits) return '000';
        // Pad to at least 3 digits, but allow more if length > 3
        return digits.padStart(3, '0');
    }
</script>

{#if isOpen}
    <!-- Backdrop -->
    <!-- svelte-ignore a11y-click-events-have-key-events -->
    <!-- svelte-ignore a11y-no-static-element-interactions -->
    <div class="fixed inset-0 z-[100] bg-black/60 flex items-start justify-center overflow-y-auto p-4 sm:p-8 print:p-0 print:bg-white" transition:fade={{ duration: 150 }}>
        
        <!-- Modal Content Container -->
        <div class="bg-white rounded-2xl w-full max-w-4xl shadow-2xl relative flex flex-col print:shadow-none print:rounded-none" transition:scale={{ duration: 200, start: 0.95 }}>
            
            <!-- Modal Header (No Print) -->
            <div class="flex items-center justify-between p-6 border-b border-slate-100 no-print">
                <div>
                    <h3 class="text-xl font-bold text-slate-900">Cetak Kwitansi GUP</h3>
                    <p class="text-sm text-slate-500">Pratinjau mengikuti format kwitansi dan dapat diedit sebelum dicetak.</p>
                </div>
                <div class="flex items-center gap-3 no-print">
                    <Button variant="outline" on:click={close}>Tutup</Button>
                    <Button variant="warning" on:click={handlePrint}>
                        <svg class="h-4 w-4 mr-2" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M17 17h2a2 2 0 002-2v-4a2 2 0 00-2-2H5a2 2 0 00-2 2v4a2 2 0 002 2h2m2 4h6a2 2 0 002-2v-4a2 2 0 00-2-2H9a2 2 0 00-2 2v4a2 2 0 002 2zm8-12V5a2 2 0 00-2-2H9a2 2 0 00-2 2v4h10z" />
                        </svg>
                        Cetak Kwitansi
                    </Button>
                </div>
            </div>

            <!-- Printable Area -->
            <div class="p-8 print:p-0">
                <div id="kwitansiPrintArea" class="kwitansi-sheet">
                    <h1 class="gup-kwitansi-title" contenteditable="true">KWITANSI</h1>
                    <table class="gup-kwitansi-meta no-border">
                        <tbody>
                            <tr><td>No. BKU</td><td>:</td><td contenteditable="true">{formatBkuNumber(data?.businessId)}</td></tr>
                            <tr><td>Tahun Anggaran</td><td>:</td><td contenteditable="true">{getYear(data?.receiptDate)}</td></tr>
                            <tr><td>Akun</td><td>:</td><td contenteditable="true">{data?.recipient || '-'}</td></tr>
                        </tbody>
                    </table>

                    <table class="gup-kwitansi-body no-border">
                        <tbody>
                            <tr>
                                <td class="label">Sudah Terima Dari</td><td class="colon">:</td>
                                <td contenteditable="true">KPA DIPA - Sekretariat Jenderal Kementerian Ketenagakerjaan</td>
                            </tr>
                            <tr>
                                <td class="label"></td><td class="colon"></td>
                                <td><strong>Rp.</strong> <span class="gup-kwitansi-amount" contenteditable="true">{formatCurrency(data?.paidAmount || data?.valueAmount || 0)}</span></td>
                            </tr>
                            <tr>
                                <td class="label">Banyaknya Uang</td><td class="colon">:</td>
                                <td><span class="gup-kwitansi-amount" contenteditable="true">{terbilang(data?.paidAmount || data?.valueAmount || 0)}</span></td>
                            </tr>
                            <tr>
                                <td class="label">Untuk Pembayaran</td><td class="colon">:</td>
                                <td><span class="gup-kwitansi-textbox" contenteditable="true">{data?.procurementType?.name || '-'}</span></td>
                            </tr>
                        </tbody>
                    </table>

                    <div class="gup-kwitansi-divider"></div>
                    <div class="gup-kwitansi-signatures">
                        <div class="gup-kwitansi-signature">
                            <div>Setuju dibayar<br>Pejabat Pembuat Komitmen<br>Biro Umum Unit Setjen<br>DIPA Setjen Kemnaker</div>
                            <div><span class="name" contenteditable="true" bind:textContent={printPpkName}></span><br>NIP. <span contenteditable="true" bind:textContent={printPpkNip}></span></div>
                        </div>
                        <div class="gup-kwitansi-signature">
                            <div>Lunas dibayar<br>Bendahara Pengeluaran Pembantu<br>Biro Umum Unit Setjen<br>DIPA Setjen Kemnaker</div>
                            <div><span class="name" contenteditable="true" bind:textContent={printBendaharaName}></span><br>NIP. <span contenteditable="true" bind:textContent={printBendaharaNip}></span></div>
                        </div>
                        <div class="gup-kwitansi-signature">
                            <div>Jakarta, &nbsp;&nbsp; <span contenteditable="true">{formatBulanTahun(data?.receiptDate)}</span><br>Yang Membayarkan</div>
                            <div><span class="name" contenteditable="true" bind:textContent={printPumName}></span><br>NIP. <span contenteditable="true" bind:textContent={printPumNip}></span></div>
                        </div>
                    </div>
                </div>
            </div>
        </div>
    </div>
{/if}

<style>
    /* Styling khusus yang aktif saat dicetak atau pratinjau */
    .kwitansi-sheet {
        width: 100%;
        max-width: 1120px;
        margin: 0 auto;
        font-family: Arial, Helvetica, sans-serif;
        font-size: 14px;
        line-height: 1.35;
        color: #000;
    }
    
    .kwitansi-sheet table {
        width: 100%;
        border-collapse: collapse;
    }
    
    .kwitansi-sheet td,
    .kwitansi-sheet th {
        border: 1px solid #111827;
        padding: 6px 8px;
        vertical-align: top;
    }
    
    .kwitansi-sheet .no-border td,
    .kwitansi-sheet .no-border th {
        border: 0;
    }
    
    .kwitansi-sheet [contenteditable="true"] {
        outline: 1px dashed transparent;
        min-height: 18px;
    }
    
    .kwitansi-sheet [contenteditable="true"]:hover {
        outline-color: #f59e0b;
        background: #fffbeb;
    }

    .gup-kwitansi-title {
        margin: 0 0 2px;
        text-align: center;
        font-size: 20px;
        font-style: italic;
        font-weight: 800;
        text-decoration: underline;
    }

    .gup-kwitansi-meta { border-top: 2px solid #000; border-bottom: 3px double #000; }
    .gup-kwitansi-meta td { padding: 2px 4px; vertical-align: top; }
    .gup-kwitansi-meta td:first-child { width: 130px; white-space: nowrap; }
    .gup-kwitansi-meta td:nth-child(2) { width: 12px; text-align: center; }

    .gup-kwitansi-body { margin-top: 8px; }
    .gup-kwitansi-body td { padding: 4px; vertical-align: top; }
    .gup-kwitansi-body .label { width: 150px; }
    .gup-kwitansi-body .colon { width: 12px; text-align: center; }
    .gup-kwitansi-amount {
        display: block;
        min-height: 25px;
        border: 1px solid #000;
        padding: 5px 7px;
        font-weight: 700;
    }
    .gup-kwitansi-textbox {
        display: block;
        min-height: 58px;
        border: 1px solid #000;
        padding: 6px 7px;
    }
    .gup-kwitansi-divider { margin: 2px 0 8px; border-top: 2px solid #000; }
    .gup-kwitansi-signatures {
        display: grid;
        grid-template-columns: repeat(3, minmax(0, 1fr));
        gap: 24px;
        text-align: center;
    }
    .gup-kwitansi-signature { min-height: 180px; display: flex; flex-direction: column; justify-content: space-between; }
    .gup-kwitansi-signature .name { display: inline-block; min-width: 165px; border-bottom: 1px solid #000; }

    @media (max-width: 640px) {
        .gup-kwitansi-signatures { grid-template-columns: 1fr; }
    }

    @media print {
        @page {
            margin: 0; /* Menghilangkan header & footer bawaan browser (URL, Tanggal, Hal) */
            size: auto;
        }

        :global(body.print-kwitansi *) { visibility: hidden !important; }
        :global(body.print-kwitansi .kwitansi-sheet),
        :global(body.print-kwitansi .kwitansi-sheet *) { visibility: visible !important; }
        
        :global(body.print-kwitansi .kwitansi-sheet) {
            position: absolute !important;
            left: 0 !important;
            top: 0 !important;
            width: 100% !important;
            max-width: none !important;
            font-size: 12px !important;
            padding: 1.5cm !important; /* Memberikan margin manual agar tidak menempel ke tepi kertas */
            margin: 0 !important;
        }
        
        .kwitansi-sheet [contenteditable="true"] { outline: none !important; }
    }
</style>
