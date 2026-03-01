<script>
    import Label from '$lib/components/ui/label/Label.svelte';
    import Input from '$lib/components/ui/input/Input.svelte';
    import Select from '$lib/components/ui/select/Select.svelte';
    import Textarea from '$lib/components/ui/textarea/Textarea.svelte';

    export let location = '';
    export let province = '';
    export let purpose = '';
    export let agenda = '';
    export let provinces = [];
    export let readonly = false;
</script>

<div class="bg-white rounded-xl shadow-sm border border-slate-200 overflow-hidden">
    <div class="px-6 py-4 border-b border-slate-100 bg-slate-50/50">
        <h3 class="font-semibold text-slate-800 flex items-center gap-2">
            <svg xmlns="http://www.w3.org/2000/svg" class="h-5 w-5 text-blue-600" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M17.657 16.657L13.414 20.9a1.998 1.998 0 01-2.827 0l-4.244-4.243a8 8 0 1111.314 0z" />
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 11a3 3 0 11-6 0 3 3 0 016 0z" />
            </svg>
            Lokasi & Tujuan
        </h3>
    </div>
    <div class="p-6 space-y-6">
        <div class="grid grid-cols-1 sm:grid-cols-2 gap-6">
            <div class="space-y-2">
                <Label class="text-slate-600">Lokasi Dinas (Kab/Kota)</Label>
                <Input placeholder="Contoh: Surabaya" bind:value={location} class="h-11 {readonly ? 'opacity-70 cursor-not-allowed' : ''}" disabled={readonly} />
            </div>
            <div class="space-y-2">
                <Label class="text-slate-600">Provinsi <span class="text-red-500">*</span></Label>
                <Select bind:value={province} class="h-11 {readonly ? 'opacity-70 cursor-not-allowed' : ''}" disabled={readonly}>
                    <option value="" disabled selected>Pilih Provinsi</option>
                    {#each provinces as prov}
                        <option value={prov.name}>{prov.name}</option>
                    {/each}
                </Select>
            </div>
        </div>

        <div class="space-y-3">
            <Label class="text-slate-600">Tujuan Perjalanan</Label>
            <div class="grid grid-cols-1 gap-3">
                <label class="relative flex items-center p-4 rounded-lg border border-slate-200 {readonly ? 'cursor-not-allowed opacity-70' : 'cursor-pointer hover:bg-slate-50'} transition-colors has-[:checked]:border-blue-500 has-[:checked]:bg-blue-50/30">
                    <input type="radio" name="purpose" value="persiapan" bind:group={purpose} disabled={readonly} class="accent-blue-600 h-4 w-4 mr-3" />
                    <span class="text-sm font-medium text-slate-700">Persiapan dan Pendampingan Kunjungan Kerja</span>
                </label>
                <label class="relative flex items-center p-4 rounded-lg border border-slate-200 {readonly ? 'cursor-not-allowed opacity-70' : 'cursor-pointer hover:bg-slate-50'} transition-colors has-[:checked]:border-blue-500 has-[:checked]:bg-blue-50/30">
                    <input type="radio" name="purpose" value="koordinasi" bind:group={purpose} disabled={readonly} class="accent-blue-600 h-4 w-4 mr-3" />
                    <span class="text-sm font-medium text-slate-700">Koordinasi dan Konsultasi Kunjungan Kerja</span>
                </label>
            </div>
        </div>

        <div class="space-y-2">
            <Label class="text-slate-600">Detail Agenda</Label>
            <Textarea placeholder="Jelaskan detail agenda kegiatan secara singkat..." bind:value={agenda} class="min-h-[100px] resize-y {readonly ? 'opacity-70 cursor-not-allowed' : ''}" disabled={readonly} />
        </div>
    </div>
</div>
