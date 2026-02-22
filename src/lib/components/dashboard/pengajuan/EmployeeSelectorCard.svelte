<script>
    import { createEventDispatcher } from 'svelte';
    import Button from '$lib/components/ui/button/Button.svelte';

    export let employees = [];
    export let selectedEmployees = [];

    const dispatch = createEventDispatcher();

    function toggleEmployee(id) {
        dispatch('toggle', id);
    }
</script>

<div class="bg-white rounded-xl shadow-sm border border-slate-200 overflow-hidden flex flex-col h-[500px]">
    <div class="px-6 py-4 border-b border-slate-100 bg-slate-50/50 flex justify-between items-center">
        <h3 class="font-semibold text-slate-800">Pilih Petugas</h3>
        <span class="text-xs font-medium px-2 py-1 bg-blue-100 text-blue-700 rounded-full">
            {selectedEmployees.length}/6
        </span>
    </div>
    <div class="p-4 overflow-y-auto flex-1 space-y-2">
        {#each employees as employee}
            <label class="flex items-center space-x-3 p-3 rounded-lg border border-transparent hover:border-slate-200 hover:bg-slate-50 transition-all cursor-pointer has-[:checked]:bg-blue-50/30 has-[:checked]:border-blue-200">
                <input type="checkbox" 
                    checked={selectedEmployees.includes(employee.id)}
                    on:change={() => toggleEmployee(employee.id)}
                    class="accent-blue-600 h-5 w-5 rounded border-slate-300 shrink-0"
                />
                <div class="grid gap-0.5">
                    <span class="text-sm font-semibold text-slate-800">{employee.name}</span>
                    <span class="text-xs text-slate-500">{employee.rank} - {employee.golongan}</span>
                </div>
            </label>
        {/each}
    </div>
    <div class="p-4 border-t border-slate-100 bg-slate-50/50">
        <p class="text-xs text-slate-500 text-center mb-3">Pastikan data sudah benar sebelum menyimpan.</p>
        <Button class="w-full bg-blue-600 hover:bg-blue-700 text-white shadow-lg shadow-blue-500/20" on:click={() => dispatch('submit')}>
            Simpan Pengajuan
        </Button>
    </div>
</div>
