<script>
    import { toast } from '$lib/shared/stores/toast';
    import Button from '$lib/shared/ui/button/Button.svelte';
    import Input from '$lib/shared/ui/input/Input.svelte';

    export let employees = [];
    export let selectedSpjEmployees = [];
    export let selectedRiilEmployees = [];
    export let isLoading = false;
    export let readonly = false;
    export let disabledIds = [];
    export let dalkotType = 'SPJ RIIL';
    export let spjCostPerPerson = 170000;
    export let actualCostPerPerson = 250000;
    
    import { createEventDispatcher } from 'svelte';
    import { formatCurrency } from '$lib/shared/utils/utils';
    const dispatch = createEventDispatcher();

    let searchQuerySpj = '';
    let searchQueryRiil = '';

    $: filteredSpjEmployees = employees.filter(e => 
        (e.name || '').toLowerCase().includes(searchQuerySpj.toLowerCase()) ||
        (e.nip || '').includes(searchQuerySpj) ||
        (e.jabatan || '').toLowerCase().includes(searchQuerySpj.toLowerCase())
    );

    $: filteredRiilEmployees = employees.filter(e => 
        (e.name || '').toLowerCase().includes(searchQueryRiil.toLowerCase()) ||
        (e.nip || '').includes(searchQueryRiil) ||
        (e.jabatan || '').toLowerCase().includes(searchQueryRiil.toLowerCase())
    );

    function toggleSpjEmployee(id) {
        if (readonly || disabledIds.includes(id)) return;
        if (selectedSpjEmployees.includes(id)) {
            selectedSpjEmployees = selectedSpjEmployees.filter(e => e !== id);
        } else {
            selectedSpjEmployees = [...selectedSpjEmployees, id];
        }
    }

    function toggleRiilEmployee(id) {
        if (readonly || disabledIds.includes(id)) return;
        if (selectedRiilEmployees.includes(id)) {
            selectedRiilEmployees = selectedRiilEmployees.filter(e => e !== id);
        } else {
            selectedRiilEmployees = [...selectedRiilEmployees, id];
        }
    }

    function formatToElegantStyle(amount) {
        return formatCurrency(amount);
    }
</script>

<div class="space-y-6">
    <!-- SPJ Employees Selection -->
    {#if dalkotType === 'SPJ RIIL'}
    <div class="bg-white rounded-xl shadow-sm border border-slate-200 overflow-hidden flex flex-col h-[400px]">
        <div class="px-6 py-4 border-b border-slate-100 bg-slate-50/50 flex justify-between items-center">
            <h3 class="font-semibold text-slate-800">Petugas SPJ</h3>
            <span class="text-xs font-medium px-2 py-1 bg-blue-100 text-blue-700 rounded-full">
                {selectedSpjEmployees.length} dipilih
            </span>
        </div>
        <div class="p-3 border-b border-slate-100">
            <div class="relative">
                <div class="absolute inset-y-0 left-0 pl-3 flex items-center pointer-events-none text-slate-400">
                    <svg xmlns="http://www.w3.org/2000/svg" class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z" />
                    </svg>
                </div>
                <Input type="text" placeholder="Cari nama, NIP, jabatan..." bind:value={searchQuerySpj} class="pl-9 bg-slate-50/50 border-slate-200 text-sm h-9" />
            </div>
        </div>
        <div class="p-4 overflow-y-auto flex-1 space-y-2">
            {#if isLoading}
                <div class="space-y-3 animate-pulse">
                    {#each Array(3) as _}
                    <div class="flex items-center space-x-3 p-3 rounded-lg border border-slate-100 bg-slate-50/50">
                        <div class="h-5 w-5 bg-slate-200 rounded shrink-0"></div>
                        <div class="grid gap-2 w-full">
                            <div class="h-4 bg-slate-200 rounded w-1/2"></div>
                            <div class="h-3 bg-slate-200 rounded w-3/4"></div>
                        </div>
                    </div>
                    {/each}
                </div>
            {:else if filteredSpjEmployees.length === 0}
                <div class="text-center py-6 text-sm text-slate-500 italic">Tidak ada petugas yang cocok.</div>
            {:else}
                {#each filteredSpjEmployees as employee (employee.id)}
                    {@const isDisabled = disabledIds.includes(employee.id)}
                    <label class="flex items-center space-x-3 p-3 rounded-lg border border-transparent {readonly || isDisabled ? 'cursor-default opacity-50 bg-slate-50' : 'hover:border-slate-200 hover:bg-slate-50 cursor-pointer'} transition-all has-[:checked]:bg-blue-50/30 has-[:checked]:border-blue-200">
                        <input type="checkbox"
                            checked={selectedSpjEmployees.includes(employee.id)}
                            on:change={() => !isDisabled && toggleSpjEmployee(employee.id)}
                            disabled={readonly || isDisabled}
                            class="accent-blue-600 h-5 w-5 rounded border-slate-300 shrink-0 {readonly || isDisabled ? 'opacity-50' : ''}"
                        />
                        <div class="grid gap-0.5 w-full">
                            <span class="text-sm font-semibold text-slate-800 flex items-center justify-between gap-2">
                                <span class="flex items-center gap-2">
                                    {employee.name}
                                </span>
                                {#if isDisabled}
                                    <span class="text-[10px] text-red-500 bg-red-50 px-1.5 py-0.5 rounded border border-red-100 font-medium">Sibuk</span>
                                {/if}
                            </span>
                            {#if employee.jabatan}
                                <span class="text-xs font-medium text-slate-700">{employee.jabatan}</span>
                            {/if}
                            {#if (employee.pangkat && employee.pangkat !== '-') || (employee.golongan && employee.golongan !== '-')}
                                <span class="text-xs text-slate-500">
                                    {#if (employee.pangkat && employee.pangkat !== '-') && (employee.golongan && employee.golongan !== '-')}
                                        {employee.pangkat} ({employee.golongan})
                                    {:else}
                                        {(employee.pangkat !== '-' ? employee.pangkat : '') || (employee.golongan !== '-' ? employee.golongan : '')}
                                    {/if}
                                </span>
                            {/if}
                            <span class="text-[10px] text-slate-400 font-mono">NIP. {employee.nip || '-'}</span>
                        </div>
                    </label>
                {/each}
            {/if}
        </div>
    </div>
    
    <!-- Riil Employees Selection -->
    <div class="bg-white rounded-xl shadow-sm border border-slate-200 overflow-hidden flex flex-col h-[400px]">
        <div class="px-6 py-4 border-b border-slate-100 bg-slate-50/50 flex justify-between items-center">
            <h3 class="font-semibold text-slate-800">Petugas Riil</h3>
            <span class="text-xs font-medium px-2 py-1 bg-blue-100 text-blue-700 rounded-full">
                {selectedRiilEmployees.length} dipilih
            </span>
        </div>
        <div class="p-3 border-b border-slate-100">
            <div class="relative">
                <div class="absolute inset-y-0 left-0 pl-3 flex items-center pointer-events-none text-slate-400">
                    <svg xmlns="http://www.w3.org/2000/svg" class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z" />
                    </svg>
                </div>
                <Input type="text" placeholder="Cari nama, NIP, jabatan..." bind:value={searchQueryRiil} class="pl-9 bg-slate-50/50 border-slate-200 text-sm h-9" />
            </div>
        </div>
        <div class="p-4 overflow-y-auto flex-1 space-y-2">
            {#if isLoading}
                <div class="space-y-3 animate-pulse">
                    {#each Array(3) as _}
                    <div class="flex items-center space-x-3 p-3 rounded-lg border border-slate-100 bg-slate-50/50">
                        <div class="h-5 w-5 bg-slate-200 rounded shrink-0"></div>
                        <div class="grid gap-2 w-full">
                            <div class="h-4 bg-slate-200 rounded w-1/2"></div>
                            <div class="h-3 bg-slate-200 rounded w-3/4"></div>
                        </div>
                    </div>
                    {/each}
                </div>
            {:else if filteredRiilEmployees.length === 0}
                <div class="text-center py-6 text-sm text-slate-500 italic">Tidak ada petugas yang cocok.</div>
            {:else}
                {#each filteredRiilEmployees as employee (employee.id)}
                    {@const isDisabled = disabledIds.includes(employee.id)}
                    <label class="flex items-center space-x-3 p-3 rounded-lg border border-transparent {readonly || isDisabled ? 'cursor-default opacity-50 bg-slate-50' : 'hover:border-slate-200 hover:bg-slate-50 cursor-pointer'} transition-all has-[:checked]:bg-blue-50/30 has-[:checked]:border-blue-200">
                        <input type="checkbox"
                            checked={selectedRiilEmployees.includes(employee.id)}
                            on:change={() => !isDisabled && toggleRiilEmployee(employee.id)}
                            disabled={readonly || isDisabled}
                            class="accent-blue-600 h-5 w-5 rounded border-slate-300 shrink-0 {readonly || isDisabled ? 'opacity-50' : ''}"
                        />
                        <div class="grid gap-0.5 w-full">
                            <span class="text-sm font-semibold text-slate-800 flex items-center justify-between gap-2">
                                <span class="flex items-center gap-2">
                                    {employee.name}
                                </span>
                                {#if isDisabled}
                                    <span class="text-[10px] text-red-500 bg-red-50 px-1.5 py-0.5 rounded border border-red-100 font-medium">Sibuk</span>
                                {/if}
                            </span>
                            {#if employee.jabatan}
                                <span class="text-xs font-medium text-slate-700">{employee.jabatan}</span>
                            {/if}
                            {#if (employee.pangkat && employee.pangkat !== '-') || (employee.golongan && employee.golongan !== '-')}
                                <span class="text-xs text-slate-500">
                                    {#if (employee.pangkat && employee.pangkat !== '-') && (employee.golongan && employee.golongan !== '-')}
                                        {employee.pangkat} ({employee.golongan})
                                    {:else}
                                        {(employee.pangkat !== '-' ? employee.pangkat : '') || (employee.golongan !== '-' ? employee.golongan : '')}
                                    {/if}
                                </span>
                            {/if}
                            <span class="text-[10px] text-slate-400 font-mono">NIP. {employee.nip || '-'}</span>
                        </div>
                    </label>
                {/each}
            {/if}
        </div>
    </div>
    {:else}
    <!-- Kebijakan Protokol -->
    <div class="bg-white rounded-xl shadow-sm border border-slate-200 overflow-hidden flex flex-col h-[500px]">
        <div class="px-6 py-4 border-b border-slate-100 bg-slate-50/50 flex justify-between items-center">
            <h3 class="font-semibold text-slate-800">Pilih Petugas</h3>
            <span class="text-xs font-medium px-2 py-1 bg-blue-100 text-blue-700 rounded-full">
                {selectedSpjEmployees.length} dipilih
            </span>
        </div>
        <div class="p-3 border-b border-slate-100">
            <div class="relative">
                <div class="absolute inset-y-0 left-0 pl-3 flex items-center pointer-events-none text-slate-400">
                    <svg xmlns="http://www.w3.org/2000/svg" class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z" />
                    </svg>
                </div>
                <Input type="text" placeholder="Cari nama, NIP, jabatan..." bind:value={searchQuerySpj} class="pl-9 bg-slate-50/50 border-slate-200 text-sm h-9" />
            </div>
        </div>
        <div class="p-4 overflow-y-auto flex-1 space-y-2">
            {#if isLoading}
                <div class="space-y-3 animate-pulse">
                    {#each Array(5) as _}
                    <div class="flex items-center space-x-3 p-3 rounded-lg border border-slate-100 bg-slate-50/50">
                        <div class="h-5 w-5 bg-slate-200 rounded shrink-0"></div>
                        <div class="grid gap-2 w-full">
                            <div class="h-4 bg-slate-200 rounded w-1/2"></div>
                            <div class="h-3 bg-slate-200 rounded w-3/4"></div>
                        </div>
                    </div>
                    {/each}
                </div>
            {:else if filteredSpjEmployees.length === 0}
                <div class="text-center py-6 text-sm text-slate-500 italic">Tidak ada petugas yang cocok.</div>
            {:else}
                {#each filteredSpjEmployees as employee (employee.id)}
                    <label class="flex items-center space-x-3 p-3 rounded-lg border border-transparent {readonly ? 'cursor-default opacity-50 bg-slate-50' : 'hover:border-slate-200 hover:bg-slate-50 cursor-pointer'} transition-all has-[:checked]:bg-blue-50/30 has-[:checked]:border-blue-200">
                        <input type="checkbox"
                            checked={selectedSpjEmployees.includes(employee.id)}
                            on:change={() => toggleSpjEmployee(employee.id)}
                            disabled={readonly}
                            class="accent-blue-600 h-5 w-5 rounded border-slate-300 shrink-0 {readonly ? 'opacity-50' : ''}"
                        />
                        <div class="grid gap-0.5">
                            <span class="text-sm font-semibold text-slate-800 flex items-center gap-2">
                                {employee.name}
                            </span>
                            {#if employee.jabatan}
                                <span class="text-xs font-medium text-slate-700">{employee.jabatan}</span>
                            {/if}
                            {#if (employee.pangkat && employee.pangkat !== '-') || (employee.golongan && employee.golongan !== '-')}
                                <span class="text-xs text-slate-500">
                                    {#if (employee.pangkat && employee.pangkat !== '-') && (employee.golongan && employee.golongan !== '-')}
                                        {employee.pangkat} ({employee.golongan})
                                    {:else}
                                        {(employee.pangkat !== '-' ? employee.pangkat : '') || (employee.golongan !== '-' ? employee.golongan : '')}
                                    {/if}
                                </span>
                            {/if}
                            <span class="text-[10px] text-slate-400 font-mono">NIP. {employee.nip || '-'}</span>
                        </div>
                    </label>
                {/each}
            {/if}
        </div>
    </div>
    {/if}

    <!-- Calculation Summary Card matching CostEstimateCard style -->
    <div class="bg-white rounded-2xl shadow-lg shadow-slate-200/40 border border-slate-100 overflow-hidden mt-6 transition-all duration-300 hover:shadow-xl group">
        <!-- Compact Header -->
        <div class="px-6 py-4 bg-slate-900 text-white flex items-center justify-between">
            <div class="flex items-center gap-3">
                <div class="p-2 bg-white/10 rounded-lg backdrop-blur-sm border border-white/10">
                    <svg xmlns="http://www.w3.org/2000/svg" class="h-4 w-4 text-indigo-300" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 7h6m0 10v-3m-3 3h.01M9 17h.01M9 14h.01M12 14h.01M15 11h.01M12 11h.01M9 11h.01M7 21h10a2 2 0 002-2V5a2 2 0 00-2-2H7a2 2 0 00-2 2v14a2 2 0 002 2z" />
                    </svg>
                </div>
                <h3 class="text-sm font-bold tracking-wide uppercase">Estimasi Biaya</h3>
            </div>
            <span class="text-[10px] font-bold text-slate-400 uppercase tracking-widest bg-white/5 px-2 py-1 rounded-md">Dalam Kota</span>
        </div>
        
        <div class="p-5 space-y-4 bg-white">
            {#if dalkotType === 'SPJ RIIL'}
                <!-- Baris Petugas SPJ -->
                <div class="space-y-1">
                    <div class="flex items-center gap-2 mb-2">
                        <div class="w-2 h-2 rounded-full bg-blue-500"></div>
                        <span class="text-xs font-bold text-slate-500 uppercase tracking-widest">Petugas SPJ</span>
                    </div>
                    <div class="flex justify-between items-center px-3 py-2 rounded-lg bg-blue-50/60 border border-blue-100">
                        <div class="text-sm text-slate-600">
                            <span class="font-bold text-blue-700">{selectedSpjEmployees?.length || 0}</span> orang
                            <span class="text-slate-400 mx-1">×</span>
                            <span class="font-mono text-slate-600">Rp {(spjCostPerPerson || 0).toLocaleString('id-ID')}</span>
                        </div>
                        <div class="text-sm font-bold text-blue-700 font-mono">
                            Rp {((selectedSpjEmployees?.length || 0) * (spjCostPerPerson || 0)).toLocaleString('id-ID')}
                        </div>
                    </div>
                </div>

                <!-- Baris Petugas Riil -->
                <div class="space-y-1">
                    <div class="flex items-center gap-2 mb-2">
                        <div class="w-2 h-2 rounded-full bg-amber-500"></div>
                        <span class="text-xs font-bold text-slate-500 uppercase tracking-widest">Petugas Riil</span>
                    </div>
                    <div class="flex justify-between items-center px-3 py-2 rounded-lg bg-amber-50/60 border border-amber-100">
                        <div class="text-sm text-slate-600">
                            <span class="font-bold text-amber-700">{selectedRiilEmployees?.length || 0}</span> orang
                            <span class="text-slate-400 mx-1">×</span>
                            <span class="font-mono text-slate-600">Rp {(actualCostPerPerson || 0).toLocaleString('id-ID')}</span>
                        </div>
                        <div class="text-sm font-bold text-amber-700 font-mono">
                            Rp {((selectedRiilEmployees?.length || 0) * (actualCostPerPerson || 0)).toLocaleString('id-ID')}
                        </div>
                    </div>
                </div>
            {:else}
                <!-- Kebijakan Protokol: hanya SPJ -->
                <div class="space-y-1">
                    <div class="flex items-center gap-2 mb-2">
                        <div class="w-2 h-2 rounded-full bg-blue-500"></div>
                        <span class="text-xs font-bold text-slate-500 uppercase tracking-widest">Petugas</span>
                    </div>
                    <div class="flex justify-between items-center px-3 py-2 rounded-lg bg-blue-50/60 border border-blue-100">
                        <div class="text-sm text-slate-600">
                            <span class="font-bold text-blue-700">{selectedSpjEmployees?.length || 0}</span> orang
                            <span class="text-slate-400 mx-1">×</span>
                            <span class="font-mono text-slate-600">Rp {(spjCostPerPerson || 0).toLocaleString('id-ID')}</span>
                        </div>
                        <div class="text-sm font-bold text-blue-700 font-mono">
                            Rp {((selectedSpjEmployees?.length || 0) * (spjCostPerPerson || 0)).toLocaleString('id-ID')}
                        </div>
                    </div>
                </div>
            {/if}

            <!-- Total Section -->
            <div class="pt-4 border-t border-slate-100">
                <div class="bg-indigo-600 p-4 rounded-xl flex items-center justify-between text-white shadow-lg shadow-indigo-200">
                    <div class="flex flex-col">
                        <span class="text-[10px] font-bold text-indigo-200 uppercase tracking-widest">Total Estimasi</span>
                        <span class="text-xl font-black tracking-tight">
                            {#if dalkotType === 'SPJ RIIL'}
                                {formatToElegantStyle(((selectedSpjEmployees?.length || 0) * spjCostPerPerson) + ((selectedRiilEmployees?.length || 0) * actualCostPerPerson))}
                            {:else}
                                {formatToElegantStyle((selectedSpjEmployees?.length || 0) * spjCostPerPerson)}
                            {/if}
                        </span>
                    </div>
                    <div class="p-2 bg-white/10 rounded-lg">
                        <svg xmlns="http://www.w3.org/2000/svg" class="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13 7l5 5m0 0l-5 5m5-5H6" />
                        </svg>
                    </div>
                </div>
            </div>
        </div>

        {#if !readonly}
        <div class="p-4 border-t border-slate-100 bg-slate-50/50">
            <p class="text-xs text-slate-500 text-center mb-3">Pastikan data sudah benar sebelum menyimpan.</p>
            <Button 
                class="w-full bg-blue-600 hover:bg-blue-700 text-white shadow-lg shadow-blue-500/20" 
                on:click={() => dispatch('submit')}
                disabled={(selectedSpjEmployees?.length || 0) === 0 && (selectedRiilEmployees?.length || 0) === 0}
            >
                Simpan Pengajuan Dalkot
            </Button>
        </div>
        {/if}
    </div>
</div>
