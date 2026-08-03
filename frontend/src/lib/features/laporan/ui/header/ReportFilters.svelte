<script>
    import Input from '$lib/shared/ui/input/Input.svelte';
    import Select from '$lib/shared/ui/select/Select.svelte';
    
    export let searchQuery = '';
    export let statusFilter = 'all';
    export let sortOption = 'date-desc';
</script>

<div class="flex flex-col sm:flex-row gap-2 w-full md:w-auto">
    <div class="relative w-full md:w-64">
        <div class="absolute inset-y-0 left-0 pl-3 flex items-center pointer-events-none text-slate-400">
            <svg xmlns="http://www.w3.org/2000/svg" class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z" />
            </svg>
        </div>
        <Input type="text" placeholder="Cari Kegiatan, Lokasi, SPD, Nama..." bind:value={searchQuery} class="pl-9 h-10 bg-white border-slate-200 w-full" />
    </div>
    
    <div class="flex gap-2 w-full sm:w-auto">
        <div class="w-1/2 sm:w-32">
            <Select bind:value={statusFilter} class="h-10 text-xs w-full bg-white border-slate-200" options={[
				{value: 'all', label: 'Semua Status'},
				{value: 'Completed', label: 'Selesai'},
				{value: 'Pending', label: 'Pending'}
			]} />
        </div>
        <div class="w-1/2 sm:w-36">
            <Select bind:value={sortOption} class="h-10 text-xs w-full bg-white border-slate-200" options={[
				{value: 'date-desc', label: 'Terbaru Pertama'},
				{value: 'date-asc', label: 'Terlama Pertama'},
				{value: 'spj-asc', label: 'ID SPJ (A-Z)'},
				{value: 'spj-desc', label: 'ID SPJ (Z-A)'}
			]} />
        </div>
    </div>
</div>
