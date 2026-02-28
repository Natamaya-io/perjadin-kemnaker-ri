<script>
    import Label from '$lib/components/ui/label/Label.svelte';
    import Input from '$lib/components/ui/input/Input.svelte';

    export let email = '';
    export let startDate = '';
    export let endDate = '';
    export let readonly = false;
</script>

<div class="bg-white rounded-xl shadow-sm border border-slate-200 overflow-hidden">
    <div class="px-6 py-4 border-b border-slate-100 bg-slate-50/50">
        <h3 class="font-semibold text-slate-800 flex items-center gap-2">
            <svg xmlns="http://www.w3.org/2000/svg" class="h-5 w-5 text-blue-600" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13 16h-1v-4h-1m1-4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
            </svg>
            Informasi Dasar
        </h3>
    </div>
    <div class="p-6 space-y-6">
        <div class="space-y-2">
            <Label class="text-slate-600">Email Pengaju</Label>
            <Input value={email} readonly disabled class="bg-slate-50 text-slate-500 border-slate-200" />
        </div>
        
        <div class="grid grid-cols-1 sm:grid-cols-2 gap-6">
            <div class="space-y-2">
                <Label class="text-slate-600">Tanggal Mulai <span class="text-red-500">*</span></Label>
                <Input type="date" bind:value={startDate} required class="h-11 {readonly ? 'opacity-70 cursor-not-allowed' : ''}" disabled={readonly} />
            </div>
            <div class="space-y-2">
                <Label class="text-slate-600">Tanggal Selesai <span class="text-red-500">*</span></Label>
                <Input type="date" bind:value={endDate} required class="h-11 {readonly ? 'opacity-70 cursor-not-allowed' : ''}" disabled={readonly} />
            </div>
        </div>
    </div>
</div>
