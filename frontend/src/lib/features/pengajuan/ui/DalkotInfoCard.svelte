<script>
    import Label from '$lib/shared/ui/label/Label.svelte';
    import Input from '$lib/shared/ui/input/Input.svelte';
    import Select from '$lib/shared/ui/select/Select.svelte';
    import Textarea from '$lib/shared/ui/textarea/Textarea.svelte';

    export let category = 'Hari Libur';
    export let official = '';
    export let dalkotType = 'SPJ RIIL';
    export let activityName = '';
    export let location = '';
    export let executionDate = '';
    export let suratTugasNumber = '';
    export let suratTugasDate = '';
    export let spjCostPerPerson = 170000;
    export let actualCostPerPerson = 250000;
    export let reportContent = '';
    export let documentationFile = null;
    export let readonly = false;

    let fileInput;
    function handleFileChange(event) {
        const file = event.target.files[0];
        if (file) {
            documentationFile = file;
        } else {
            documentationFile = null;
        }
    }
</script>

<div class="bg-white rounded-xl shadow-sm border border-slate-200 overflow-hidden">
    <!-- Header -->
    <div class="px-6 py-4 border-b border-slate-100 bg-slate-50/50">
        <h3 class="font-semibold text-slate-800 flex items-center gap-2">
            <svg xmlns="http://www.w3.org/2000/svg" class="h-5 w-5 text-blue-500" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 21V5a2 2 0 00-2-2H7a2 2 0 00-2 2v16m14 0h2m-2 0h-5m-9 0H3m2 0h5M9 7h1m-1 4h1m4-4h1m-1 4h1m-5 10v-5a1 1 0 011-1h2a1 1 0 011 1v5m-4 0h4" />
            </svg>
            Detail Kegiatan Dalam Kota
        </h3>
    </div>

    <!-- Body -->
    <div class="p-6 space-y-6">
        <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
            <!-- Tanggal Pelaksanaan -->
            <div class="space-y-2">
                <Label class="text-slate-600 text-sm">Tanggal Pelaksanaan Dinas *</Label>
                <Input 
                    type="date" 
                    bind:value={executionDate}
                    disabled={readonly}
                    class="h-10 text-sm {readonly ? 'opacity-70 cursor-not-allowed' : ''}"
                />
            </div>

            <!-- Kategori -->
            <div class="space-y-2">
                <Label class="text-slate-600 text-sm">Kategori Dalkot *</Label>
                <Select 
                    bind:value={category}
                    disabled={readonly}
                    class="h-10 text-sm {readonly ? 'opacity-70 cursor-not-allowed' : ''}"
                >
                    <option value="Jam Kerja">Jam Kerja</option>
                    <option value="Overtime">Overtime (Luar Jam Kerja)</option>
                    <option value="Hari Libur">Hari Libur</option>
                </Select>
            </div>
        </div>

        <div class="space-y-2">
            <Label class="text-slate-600 text-sm">Pejabat yang Didampingi *</Label>
            <Input 
                type="text" 
                bind:value={official}
                placeholder="Cth: Sekretaris Jenderal"
                disabled={readonly}
                class="h-10 text-sm {readonly ? 'opacity-70 cursor-not-allowed' : ''}"
            />
        </div>

        <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
            <div class="space-y-2">
                <Label class="text-slate-600 text-sm">Nomor Surat Tugas <span class="text-xs text-slate-400 font-normal">(Opsional)</span></Label>
                <Input 
                    type="text" 
                    bind:value={suratTugasNumber}
                    placeholder="Cth: ST/123/VI/2026"
                    disabled={readonly}
                    class="h-10 text-sm {readonly ? 'opacity-70 cursor-not-allowed' : ''}"
                />
                <p class="text-[10px] text-slate-400">Nomor ini akan ditampilkan pada rekap bulanan SPJ.</p>
            </div>
            
            <div class="space-y-2">
                <Label class="text-slate-600 text-sm">Tanggal Surat Tugas <span class="text-xs text-slate-400 font-normal">(Opsional)</span></Label>
                <Input 
                    type="date" 
                    bind:value={suratTugasDate}
                    disabled={readonly}
                    class="h-10 text-sm {readonly ? 'opacity-70 cursor-not-allowed' : ''}"
                />
                <p class="text-[10px] text-slate-400">Boleh dikosongkan apabila Surat Tugas belum diterbitkan.</p>
            </div>
        </div>

        <hr class="border-slate-100" />

        <div class="space-y-3">
            <Label class="text-slate-600 font-medium">Jenis Dalkot *</Label>
            <div class="grid grid-cols-1 gap-2">
                <label class="relative flex items-center p-3 rounded-xl border border-slate-200 {readonly ? 'cursor-not-allowed opacity-70' : 'cursor-pointer hover:bg-slate-50'} transition-colors has-[:checked]:border-blue-500 has-[:checked]:bg-blue-50/30">
                    <input type="radio" bind:group={dalkotType} value="SPJ RIIL" disabled={readonly} class="accent-blue-600 h-4 w-4 mr-3">
                    <span class="text-sm font-medium text-slate-700">SPJ RIIL</span>
                </label>
                <label class="relative flex items-center p-3 rounded-xl border border-slate-200 {readonly ? 'cursor-not-allowed opacity-70' : 'cursor-pointer hover:bg-slate-50'} transition-colors has-[:checked]:border-blue-500 has-[:checked]:bg-blue-50/30">
                    <input type="radio" bind:group={dalkotType} value="Kebijakan Protokol" disabled={readonly} class="accent-blue-600 h-4 w-4 mr-3">
                    <span class="text-sm font-medium text-slate-700">Kebijakan Protokol</span>
                </label>
            </div>
        </div>

        {#if dalkotType === 'SPJ RIIL'}
        <div class="p-4 bg-slate-50/50 rounded-xl border border-slate-200 space-y-4">
            <div>
                <h4 class="font-medium text-slate-700 text-sm">Form SPJ RIIL</h4>
                <p class="text-xs text-slate-500">Petugas SPJ dan petugas Riil dapat dipilih secara terpisah di panel sebelah kanan.</p>
            </div>
            
            <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
                <div class="space-y-2">
                    <Label class="text-slate-600 text-sm">SPJ per petugas</Label>
                    <div class="flex items-center">
                        <span class="inline-flex items-center px-3 border border-r-0 border-slate-300 bg-slate-100 text-slate-500 text-sm rounded-l-md h-10">
                            Rp
                        </span>
                        <Input 
                            type="number" 
                            bind:value={spjCostPerPerson}
                            disabled={readonly}
                            class="flex-1 rounded-none rounded-r-md h-10 {readonly ? 'bg-slate-50 opacity-70 cursor-not-allowed' : ''}"
                        />
                    </div>
                </div>
                <div class="space-y-2">
                    <Label class="text-slate-600 text-sm">Biaya riil per petugas</Label>
                    <div class="flex items-center">
                        <span class="inline-flex items-center px-3 border border-r-0 border-slate-300 bg-slate-100 text-slate-500 text-sm rounded-l-md h-10">
                            Rp
                        </span>
                        <Input 
                            type="number" 
                            bind:value={actualCostPerPerson}
                            disabled={readonly}
                            class="flex-1 rounded-none rounded-r-md h-10 {readonly ? 'bg-slate-50 opacity-70 cursor-not-allowed' : ''}"
                        />
                    </div>
                </div>
            </div>
        </div>
        {/if}

        <hr class="border-slate-100" />

        <div class="space-y-2">
            <Label class="text-slate-600 font-medium">Nama Kegiatan *</Label>
            <Input 
                type="text" 
                bind:value={activityName}
                placeholder="Cth: Pendampingan Kunjungan Kerja"
                disabled={readonly}
                class="h-10 text-sm {readonly ? 'opacity-70 cursor-not-allowed' : ''}"
            />
        </div>

        <div class="space-y-2">
            <Label class="text-slate-600 font-medium">Lokasi *</Label>
            <Select 
                bind:value={location}
                disabled={readonly}
                class="h-10 text-sm {readonly ? 'opacity-70 cursor-not-allowed' : ''}"
            >
                <option value="" disabled selected>Pilih Lokasi</option>
                <option value="Jakarta Pusat">Jakarta Pusat</option>
                <option value="Jakarta Selatan">Jakarta Selatan</option>
                <option value="Jakarta Barat">Jakarta Barat</option>
                <option value="Jakarta Utara">Jakarta Utara</option>
                <option value="Jakarta Timur">Jakarta Timur</option>
            </Select>
        </div>

        <div class="space-y-2">
            <Label class="text-slate-600 font-medium">Isi Laporan *</Label>
            <Textarea 
                bind:value={reportContent}
                placeholder="Cth: Pendampingan kegiatan kunjungan kerja dan pengaturan alur keprotokolan..."
                disabled={readonly}
                class="min-h-[100px] text-sm resize-y {readonly ? 'opacity-70 cursor-not-allowed' : ''}"
            />
        </div>

        <div class="space-y-2">
            <Label class="text-slate-600 font-medium">Dokumentasi *</Label>
            <div class="flex flex-col sm:flex-row items-center gap-4 p-4 border border-slate-200 rounded-xl bg-slate-50/30">
                <input 
                    type="file" 
                    accept=".jpg,.jpeg,.png,.pdf"
                    on:change={handleFileChange}
                    disabled={readonly}
                    class="block w-full text-sm text-slate-500 file:mr-4 file:py-2.5 file:px-4 file:rounded-xl file:border-0 file:text-sm file:font-semibold file:bg-blue-50 file:text-blue-700 hover:file:bg-blue-100 cursor-pointer focus:outline-none"
                    required
                />
            </div>
            <p class="text-[10px] text-slate-400 mt-1">Pilih JPG, PNG, atau PDF. Maksimal 5 MB.</p>
        </div>
    </div>
</div>
