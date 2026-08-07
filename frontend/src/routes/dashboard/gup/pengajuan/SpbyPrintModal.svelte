<script lang="ts">
    import { createEventDispatcher } from 'svelte';
    import { fade, scale } from 'svelte/transition';
    import Button from '$lib/shared/ui/button/Button.svelte';

    export let isOpen = false;
    export let data: any = null;

    const dispatch = createEventDispatcher();

    // Default names (in real app, this might come from settings)
    let ppkName = 'Yuda Susanto';
    let ppkNip = '19810516 200901 1 003';
    let bendaharaName = 'Ardimas Perdana Putra';
    let bendaharaNip = '19970722 201812 1 002';
    
    // Editable reactive values
    let printPpkName = ppkName;
    let printPpkNip = ppkNip;
    let printBendaharaName = bendaharaName;
    let printBendaharaNip = bendaharaNip;
    let printPumName = data?.pum || 'Mary Dona Mailoa';
    let printPumNip = '19890430 201212 2 006';

    $: if (data) {
        printPumName = data.pum || 'Mary Dona Mailoa';
    }

    function close() {
        isOpen = false;
        dispatch('close');
    }

    function handlePrint() {
        // Hide normal content, show print content via body class
        document.body.classList.add('print-spby');
        window.print();
        setTimeout(() => {
            document.body.classList.remove('print-spby');
        }, 500);
    }

    function formatCurrency(amount: number) {
        return amount.toLocaleString('id-ID');
    }

    function formatDate(dateStr: string) {
        if (!dateStr) return '-';
        return new Date(dateStr).toLocaleDateString('id-ID', { day: 'numeric', month: 'long', year: 'numeric' });
    }

    function buildNomorSpby(gupData: any) {
        if (!gupData) return '';
        // Mock generation just like gup-dalkot
        const roman = ['I', 'II', 'III', 'IV', 'V', 'VI', 'VII', 'VIII', 'IX', 'X', 'XI', 'XII'];
        const date = new Date(gupData.receiptDate || new Date());
        const romanMonth = roman[date.getMonth()] || 'I';
        const year = date.getFullYear();
        // Since we don't have the exact index easily, use a placeholder random or businessId
        const prefix = (gupData.businessId || '000').slice(-3);
        return `${prefix}/001/KU.02.01/PROT/${romanMonth}/${year}`;
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
                    <h3 class="text-xl font-bold text-slate-900">Cetak SPBY GUP</h3>
                    <p class="text-sm text-slate-500">Pratinjau mengikuti format SPBY dan dapat diedit sebelum dicetak.</p>
                </div>
                <div class="flex items-center gap-3 no-print">
                    <Button variant="outline" on:click={close}>Tutup</Button>
                    <Button variant="primary" on:click={handlePrint}>
                        <svg class="w-4 h-4 mr-2" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M17 17h2a2 2 0 002-2v-4a2 2 0 00-2-2H5a2 2 0 00-2 2v4a2 2 0 002 2h2m2 4h6a2 2 0 002-2v-4a2 2 0 00-2-2H9a2 2 0 00-2 2v4a2 2 0 002 2zm8-12V5a2 2 0 00-2-2H9a2 2 0 00-2-2v4h10z" /></svg>
                        Cetak SPBY
                    </Button>
                </div>
            </div>

            <!-- Printable Area -->
            <div class="p-8 print:p-0">
                <div id="spbyPrintArea" class="spby-sheet">
                    <table>
                        <tbody>
                            <tr><td colspan="4" class="spby-heading"><span contenteditable="true">KEMENTERIAN KETENAGAKERJAAN R.I</span><br><span contenteditable="true">SATUAN KERJA BIRO UMUM</span></td></tr>
                            <tr><td colspan="4" class="spby-title" contenteditable="true">SURAT PERINTAH BAYAR</td></tr>
                            <tr>
                                <td style="width:20%"></td>
                                <td><strong>Tanggal:</strong> <span contenteditable="true">{formatDate(data?.receiptDate)}</span></td>
                                <td style="width:8%"></td>
                                <td><strong>Nomor:</strong> <span contenteditable="true">{buildNomorSpby(data)}</span></td>
                            </tr>
                            <tr><td colspan="4" contenteditable="true">Saya yang bertandatangan di bawah ini selaku Pejabat Pembuat Komitmen memerintahkan Bendahara Pengeluaran melakukan pembayaran sejumlah:</td></tr>
                            <tr><td><strong>Rp</strong></td><td colspan="3"><strong><span contenteditable="true">{formatCurrency(data?.paidAmount || data?.valueAmount || 0)}</span></strong></td></tr>
                        </tbody>
                    </table>

                    <table>
                        <tbody>
                            <tr><td style="width:20%">Kepada</td><td style="width:4%">:</td><td contenteditable="true">{data?.recipient || '-'}</td></tr>
                            <tr><td>Untuk Pembayaran</td><td>:</td><td contenteditable="true">{data?.paymentDescription || '-'}</td></tr>
                            <tr><td colspan="3" style="height:60px"></td></tr>
                            <tr>
                                <td colspan="3">
                                    <strong>Atas Dasar:</strong><br>
                                    1. <span contenteditable="true">Kuitansi/bukti pembelian</span> : <span contenteditable="true"></span><br>
                                    2. <span contenteditable="true">Nota/bukti penerimaan barang/jasa</span> : <span contenteditable="true"></span><br>
                                    &nbsp;&nbsp;&nbsp;<span contenteditable="true">(bukti lainnya)</span>
                                </td>
                            </tr>
                            <tr><td>Dibebankan pada</td><td>:</td><td></td></tr>
                            <tr><td>Kegiatan, Output, MAK</td><td>:</td><td contenteditable="true">{data?.mak || '-'}</td></tr>
                            <tr><td>Kode</td><td>:</td><td contenteditable="true">{data?.accountCode || '-'}</td></tr>
                        </tbody>
                    </table>

                    <table>
                        <tbody>
                            <tr>
                                <td style="width:33.33%"><strong>Setuju/Lunas dibayar, tgl<br>Bendahara Pengeluaran Pembantu<br>Biro Umum Unit Setjen<br>DIPA Setjen Kemnaker</strong></td>
                                <td style="width:33.33%"><strong>Diterima tgl <span contenteditable="true">{formatDate(data?.receiptDate)}</span><br>Penerima Uang/Uang Muka Kerja</strong></td>
                                <td style="width:33.33%"><strong>Jakarta, <span contenteditable="true">{formatDate(data?.receiptDate)}</span><br>Pejabat Pembuat Komitmen<br>Biro Umum Unit Setjen<br>DIPA Setjen Kemnaker</strong></td>
                            </tr>
                            <tr><td class="spby-signature-space"></td><td class="spby-signature-space"></td><td class="spby-signature-space"></td></tr>
                            <tr>
                                <td><strong><span contenteditable="true" bind:textContent={printBendaharaName}></span><br>NIP. <span contenteditable="true" bind:textContent={printBendaharaNip}></span></strong></td>
                                <td><strong><span contenteditable="true" bind:textContent={printPumName}></span><br>NIP. <span contenteditable="true" bind:textContent={printPumNip}></span></strong></td>
                                <td><strong><span contenteditable="true" bind:textContent={printPpkName}></span><br>NIP. <span contenteditable="true" bind:textContent={printPpkNip}></span></strong></td>
                            </tr>
                        </tbody>
                    </table>
                </div>
            </div>
        </div>
    </div>
{/if}

<style>
    .spby-sheet {
        width: 100%;
        max-width: 1120px;
        margin: 0 auto;
        font-family: Arial, Helvetica, sans-serif;
        font-size: 14px;
        line-height: 1.35;
        color: #000;
    }
    
    .spby-sheet table {
        width: 100%;
        border-collapse: collapse;
        margin-top: 8px;
    }
    
    .spby-sheet td,
    .spby-sheet th {
        border: 1px solid #000;
        padding: 6px 8px;
        vertical-align: top;
    }
    
    .spby-sheet [contenteditable="true"] {
        outline: 1px dashed transparent;
        min-height: 18px;
    }
    
    .spby-sheet [contenteditable="true"]:hover {
        outline-color: #f59e0b;
        background: #fffbeb;
    }

    .spby-heading { text-align: center; font-size: 18px; font-weight: 800; }
    .spby-title { text-align: center; font-size: 16px; font-weight: 800; text-decoration: underline; }
    .spby-signature-space { height: 140px; }

    @media print {
        @page {
            margin: 0; /* Menghilangkan header & footer bawaan browser (URL, Tanggal, Hal) */
            size: auto;
        }

        :global(body.print-spby *) { visibility: hidden !important; }
        :global(body.print-spby .spby-sheet),
        :global(body.print-spby .spby-sheet *) { visibility: visible !important; }
        
        :global(body.print-spby .spby-sheet) {
            position: absolute !important;
            left: 0 !important;
            top: 0 !important;
            width: 100% !important;
            max-width: none !important;
            font-size: 12px !important;
            padding: 1.5cm !important; /* Memberikan margin manual agar tidak menempel ke tepi kertas */
            margin: 0 !important;
        }
        
        .spby-sheet [contenteditable="true"] { outline: none !important; }
    }
</style>
