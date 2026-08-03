<script>
    import Input from '$lib/shared/ui/input/Input.svelte';
    import Select from '$lib/shared/ui/select/Select.svelte';
    
    export let searchQuery = '';
    export let statusFilter = 'all';
    export let sortOption = 'date-desc';
</script>

<div class="grid grid-cols-1 md:grid-cols-[1fr_auto_auto] gap-3 w-full items-center">

    <!-- Search Bar -->
    <div class="relative w-full min-w-0">
        <div class="absolute inset-y-0 left-0 pl-3 flex items-center pointer-events-none text-slate-400">
            <svg xmlns="http://www.w3.org/2000/svg" class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z" />
            </svg>
        </div>
        <Input type="text" placeholder="Cari Kegiatan, Lokasi, SPD, Nama..." bind:value={searchQuery} class="pl-9 bg-white border-slate-200 w-full" />
    </div>

    <!-- Filter Status -->
    <div class="w-full md:w-36">
        <Select bind:value={statusFilter} class="bg-white border-slate-200 w-full" options={[
            { value: 'all', label: 'Semua Status' },
            { value: 'Completed', label: 'Selesai' },
            { value: 'Pending', label: 'Pending' }
        ]} />
    </div>

    <!-- Sort -->
    <div class="w-full md:w-44">
        <Select bind:value={sortOption} class="bg-white border-slate-200 w-full" options={[
            { value: 'date-desc', label: 'Terbaru Pertama' },
            { value: 'date-asc', label: 'Terlama Pertama' },
            { value: 'spj-asc', label: 'ID SPJ (A-Z)' },
            { value: 'spj-desc', label: 'ID SPJ (Z-A)' }
        ]} />
    </div>
</div>
