<script>
    import { formatCurrency } from '$lib/utils';

    export let days = 0;
    export let rate = 0;
    export let employeeCount = 0;
    export let totalCost = 0;
    export let readonly = false;
</script>

<div class="bg-white rounded-xl shadow-sm border border-slate-200 overflow-hidden mt-6">
    <div class="px-6 py-4 border-b border-slate-100 bg-slate-50/50">
        <h3 class="font-semibold text-slate-800 flex items-center gap-2">
            <svg xmlns="http://www.w3.org/2000/svg" class="h-5 w-5 text-emerald-600" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 8c-1.657 0-3 .895-3 2s1.343 2 3 2 3 .895 3 2-1.343 2-3 2m0-8c1.11 0 2.08.402 2.599 1M12 8V7m0 1v8m0 0v1m0-1c-1.11 0-2.08-.402-2.599-1M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
            </svg>
            Estimasi Ongkos Harian (Luar Kota)
        </h3>
    </div>
    <div class="p-6 space-y-4">
        <div class="flex justify-between items-center text-sm">
            <span class="text-slate-600">Durasi Perjalanan</span>
            <span class="font-medium text-slate-800">{days} Hari</span>
        </div>
        <div class="flex justify-between items-center text-sm">
            <span class="text-slate-600">Tarif per Hari</span>
            <span class="font-medium text-slate-800">{formatCurrency(rate)}</span>
        </div>
        <div class="flex justify-between items-center text-sm">
            <span class="text-slate-600">Jumlah Petugas</span>
            <span class="font-medium text-slate-800">{employeeCount} Orang</span>
        </div>
        
        <div class="pt-4 border-t border-slate-100">
            <div class="flex justify-between items-center">
                <span class="font-semibold text-slate-700">Total Estimasi</span>
                <span class="text-lg font-bold text-emerald-600">{formatCurrency(totalCost)}</span>
            </div>
            <p class="text-xs text-slate-500 mt-2 leading-relaxed">
                *Estimasi ini hanya mencakup uang harian berdasarkan SBM Luar Kota provinsi tujuan. Biaya transportasi dan penginapan dihitung terpisah.
            </p>
        </div>
    </div>
</div>
