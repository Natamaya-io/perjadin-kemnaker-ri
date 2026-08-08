<script lang="ts">
    import { createEventDispatcher } from 'svelte';
    import { fade, scale } from 'svelte/transition';
    import Button from '$lib/shared/ui/button/Button.svelte';
    import Select from '$lib/shared/ui/select/Select.svelte';
    import { formatCurrency } from '$lib/shared/utils/utils';

    export let isOpen = false;
    export let transactions: any[] = [];
    export let masterData: any = { procurementTypes: [], fundingSources: [] };

    const dispatch = createEventDispatcher();

    let filterSumberDana = 'all';

    // Get unique funding sources for the current year / month based on transactions
    // In gup-dalkot, it was Month - Funding Source, but we can just list Funding Sources or use masterData
    $: sumberDanaOptions = [
        { value: 'all', label: 'Semua GUP' },
        ...masterData.fundingSources.map((s: any) => ({
            value: s.id,
            label: s.gupLabel
        }))
    ];

    $: filteredTransactions = filterSumberDana === 'all' 
        ? transactions 
        : transactions.filter(t => t.fundingSourceId === filterSumberDana);

    // Group by Procurement Type
    $: groupedData = filteredTransactions.reduce((acc: any, trx: any) => {
        const typeObj = masterData.procurementTypes.find((p: any) => p.id === trx.procurementTypeId);
        const typeName = typeObj ? typeObj.name : 'Tanpa Jenis Pengadaan';
        
        if (!acc[typeName]) {
            acc[typeName] = { name: typeName, total: 0, items: [] };
        }
        acc[typeName].items.push(trx);
        acc[typeName].total += (trx.netValue || trx.amountPaid || 0);
        return acc;
    }, {});

    $: groups = Object.values(groupedData).sort((a: any, b: any) => a.name.localeCompare(b.name));

    function close() {
        isOpen = false;
        dispatch('close');
    }

    function handlePrint() {
        document.body.classList.add('print-rekap');
        window.print();
        setTimeout(() => {
            document.body.classList.remove('print-rekap');
        }, 500);
    }
</script>

<style>
    /* Styling khusus print rekapitulasi, setara dengan app/Views/layouts/main.php */
    :global(body.print-rekap) {
        visibility: hidden;
    }
    :global(body.print-rekap .rekap-modal-content) {
        visibility: visible;
        position: absolute;
        left: 0;
        top: 0;
        width: 100%;
        margin: 0;
        padding: 0;
    }
    :global(body.print-rekap .no-print) {
        display: none !important;
    }

    .rekap-table {
        width: 100%;
        border-collapse: collapse;
        font-size: 11px;
        font-family: Arial, sans-serif;
    }
    .rekap-table th, .rekap-table td {
        border: 1px solid #000;
        padding: 6px 8px;
    }
    .rekap-table th {
        background-color: #f3f4f6;
        font-weight: bold;
        text-align: left;
    }
    .rekap-table .group-row {
        background-color: #f9fafb;
        font-weight: bold;
    }
    .rekap-table .amount {
        text-align: right;
    }
    .rekap-table .signature {
        text-align: center;
        min-width: 120px;
    }
</style>

{#if isOpen}
    <div class="fixed inset-0 z-[100] flex items-center justify-center bg-black/50 p-4" transition:fade={{ duration: 200 }}>
        <div class="w-full max-w-5xl bg-white rounded-2xl shadow-2xl flex flex-col max-h-[90vh] relative" transition:scale={{ duration: 200, start: 0.95 }}>
            <!-- Header (No Print) -->
            <div class="flex items-center justify-between p-6 border-b border-slate-100 no-print">
                <div>
                    <h3 class="text-xl font-bold text-slate-800">Cetak Rekapitulasi GUP</h3>
                    <p class="text-sm text-slate-500">Tabel tanda terima GUP dikelompokkan berdasarkan Jenis Pengadaan.</p>
                </div>
                <div class="flex items-center gap-4">
                    <div class="w-48">
                        <Select 
                            bind:value={filterSumberDana} 
                            options={sumberDanaOptions} 
                            class="border-slate-200 text-sm"
                        />
                    </div>
                    <Button variant="warning" class="gap-2 shadow-sm font-bold" on:click={handlePrint}>
                        <svg class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M17 17h2a2 2 0 002-2v-4a2 2 0 00-2-2H5a2 2 0 00-2 2v4a2 2 0 002 2h2m2 4h6a2 2 0 002-2v-4a2 2 0 00-2-2H9a2 2 0 00-2 2v4a2 2 0 002 2zm8-12V5a2 2 0 00-2-2H9a2 2 0 00-2 2v4h10z" />
                        </svg>
                        Cetak
                    </Button>
                    <button class="w-10 h-10 rounded-full bg-slate-100 hover:bg-slate-200 flex items-center justify-center transition-colors text-slate-500" on:click={close}>
                        <svg class="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
                        </svg>
                    </button>
                </div>
            </div>

            <!-- Content Area -->
            <div class="p-8 overflow-y-auto bg-slate-100/50 rekap-modal-content">
                <div class="bg-white p-10 max-w-4xl mx-auto shadow-sm print:shadow-none print:p-0 print:max-w-none w-full border border-slate-200 print:border-none">
                    
                    <h2 class="text-base font-bold mb-4 font-serif">
                        Tanda Terima GUP {filterSumberDana !== 'all' ? (masterData.fundingSources.find(s => s.id === filterSumberDana)?.name || '') : ''} Subbagian Protokol
                    </h2>

                    {#if groups.length === 0}
                        <div class="p-8 text-center text-slate-500 border border-dashed border-slate-300 no-print">
                            Tidak ada data transaksi GUP untuk sumber dana yang dipilih.
                        </div>
                    {:else}
                        <table class="rekap-table">
                            <thead>
                                <tr>
                                    <th>Row Labels</th>
                                    <th>Sum of Nilai Bersih<br>(Potong Pajak)</th>
                                    <th>Tanda Tangan<br>Penerima</th>
                                </tr>
                            </thead>
                            <tbody>
                                {#each groups as group}
                                    <tr class="group-row">
                                        <td>⊟ {group.name}</td>
                                        <td class="amount">{formatCurrency(group.total).replace('Rp', '').trim()}</td>
                                        <td></td>
                                    </tr>
                                    {#each group.items as item}
                                        <tr>
                                            <td>{item.paymentDescription || '-'}</td>
                                            <td class="amount">{formatCurrency(item.netValue || item.amountPaid || 0).replace('Rp', '').trim()}</td>
                                            <td class="signature">{item.recipient || '-'}</td>
                                        </tr>
                                    {/each}
                                {/each}
                            </tbody>
                        </table>
                    {/if}
                </div>
            </div>
        </div>
    </div>
{/if}
