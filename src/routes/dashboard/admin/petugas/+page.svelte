<script>
    import { employeesStore, addEmployee, updateEmployee, removeEmployee } from '$lib/stores/records';
    import { userStore } from '$lib/stores/auth';
    import { cn } from '$lib/utils';
    import { fly } from 'svelte/transition';

    // UI Components
    import Button from '$lib/components/ui/button/Button.svelte';
    import Input from '$lib/components/ui/input/Input.svelte';
    import Label from '$lib/components/ui/label/Label.svelte';
    import Dialog from '$lib/components/ui/dialog/Dialog.svelte';
    import DialogHeader from '$lib/components/ui/dialog/DialogHeader.svelte';
    import DialogTitle from '$lib/components/ui/dialog/DialogTitle.svelte';
    import DialogFooter from '$lib/components/ui/dialog/DialogFooter.svelte';
    import Table from '$lib/components/ui/table/Table.svelte';
    import TableHeader from '$lib/components/ui/table/TableHeader.svelte';
    import TableRow from '$lib/components/ui/table/TableRow.svelte';
    import TableHead from '$lib/components/ui/table/TableHead.svelte';
    import TableBody from '$lib/components/ui/table/TableBody.svelte';
    import TableCell from '$lib/components/ui/table/TableCell.svelte';
    import { ConfirmationModal } from '$lib/components/ui/confirmation-modal';

    let searchQuery = '';
    let isModalOpen = false;
    let isConfirmOpen = false;
    let editingId = null; // null for add mode, number for edit mode

    // Form State
    let formData = {
        name: '',
        nip: '',
        rank: '',
        golongan: ''
    };

    $: filteredEmployees = $employeesStore.filter(emp => 
        (emp.name?.toLowerCase() || '').includes(searchQuery.toLowerCase()) || 
        (emp.nip || '').includes(searchQuery)
    );

    function openAddModal() {
        editingId = null;
        formData = { name: '', nip: '', rank: '', golongan: '' };
        isModalOpen = true;
    }

    function openEditModal(employee) {
        editingId = employee.id;
        formData = { ...employee };
        isModalOpen = true;
    }

    function confirmSubmit() {
        if (!formData.name || !formData.nip) {
            alert('Nama dan NIP wajib diisi');
            return;
        }
        isConfirmOpen = true;
    }

    function processSubmit() {
        if (editingId) {
            updateEmployee(editingId, formData);
        } else {
            addEmployee(formData);
        }
        isModalOpen = false;
    }

    function handleDelete(id) {
        if (confirm('Apakah Anda yakin ingin menghapus petugas ini?')) {
            removeEmployee(id);
        }
    }
</script>

<div class="space-y-6 pb-20">
    <div class="flex flex-col md:flex-row md:items-center justify-between gap-4">
        <div>
            <h2 class="text-2xl font-bold tracking-tight text-slate-900">Manajemen Petugas</h2>
            <p class="text-slate-500">Kelola data pegawai yang dapat ditugaskan dalam perjalanan dinas.</p>
        </div>
        <div class="flex items-center space-x-2 w-full md:w-auto">
            <Button on:click={openAddModal} class="bg-blue-600 hover:bg-blue-700 text-white shadow-lg shadow-blue-500/20">
                <svg xmlns="http://www.w3.org/2000/svg" class="h-4 w-4 mr-2" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4v16m8-8H4" />
                </svg>
                Tambah Petugas
            </Button>
        </div>
    </div>

    {#if $userStore.role !== 'super_admin'}
        <div class="flex flex-col items-center justify-center p-12 text-center border-2 border-dashed border-slate-200 rounded-xl bg-slate-50">
            <svg xmlns="http://www.w3.org/2000/svg" class="h-12 w-12 text-slate-300 mb-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 15v2m-6 4h12a2 2 0 002-2v-6a2 2 0 00-2-2H6a2 2 0 00-2 2v6a2 2 0 002 2zm10-10V7a4 4 0 00-8 0v4h8z" />
            </svg>
            <h3 class="text-lg font-medium text-slate-900">Akses Dibatasi</h3>
            <p class="text-slate-500 max-w-sm mt-1">Halaman ini khusus untuk Super Admin. Silakan login dengan akun yang sesuai.</p>
        </div>
    {:else}
        <!-- Search Bar -->
        <div class="relative w-full md:w-96">
            <div class="absolute inset-y-0 left-0 pl-3 flex items-center pointer-events-none text-slate-400">
                <svg xmlns="http://www.w3.org/2000/svg" class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z" />
                </svg>
            </div>
            <Input type="text" placeholder="Cari Nama atau NIP..." bind:value={searchQuery} class="pl-9 bg-white border-slate-200" />
        </div>

        <!-- Desktop Table View -->
        <div class="hidden md:block rounded-xl border border-slate-200 shadow-sm bg-white overflow-hidden">
            <div class="overflow-x-auto w-full">
                <Table class="w-full text-sm text-left">
                    <TableHeader class="bg-slate-50 border-b border-slate-200">
                        <TableRow class="hover:bg-slate-50/50">
                            <TableHead class="min-w-[200px] font-semibold text-slate-700 pl-4 py-3">Nama Pegawai</TableHead>
                            <TableHead class="min-w-[150px] font-semibold text-slate-700 py-3">NIP</TableHead>
                            <TableHead class="min-w-[150px] font-semibold text-slate-700 py-3">Pangkat</TableHead>
                            <TableHead class="min-w-[100px] font-semibold text-slate-700 py-3">Golongan</TableHead>
                            <TableHead class="w-[100px] min-w-[100px] font-semibold text-slate-700 text-right pr-4 py-3">Aksi</TableHead>
                        </TableRow>
                    </TableHeader>
                    <TableBody>
                        {#if filteredEmployees.length === 0}
                            <TableRow>
                                <TableCell colspan="5" class="text-center py-12 text-slate-500 italic bg-slate-50/20">
                                    <div class="flex flex-col items-center justify-center gap-2">
                                        <svg xmlns="http://www.w3.org/2000/svg" class="h-8 w-8 text-slate-300" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                                            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4.354a4 4 0 110 5.292M15 21H3v-1a6 6 0 0112 0v1zm0 0h6v-1a6 6 0 00-9-5.197M13 7a4 4 0 11-8 0 4 4 0 018 0z" />
                                        </svg>
                                        <span>Tidak ada data pegawai ditemukan.</span>
                                    </div>
                                </TableCell>
                            </TableRow>
                        {:else}
                            {#each filteredEmployees as emp (emp.id)}
                                <TableRow class="hover:bg-slate-50 transition-colors border-b border-slate-100 last:border-0">
                                    <TableCell class="pl-4 py-3 font-medium text-slate-900">{emp.name}</TableCell>
                                    <TableCell class="py-3 font-mono text-xs text-slate-600 bg-slate-50/50 rounded-sm px-2 w-fit mx-2">{emp.nip}</TableCell>
                                    <TableCell class="py-3 text-slate-700">{emp.rank}</TableCell>
                                    <TableCell class="py-3 text-slate-700">
                                        <span class="inline-flex items-center px-2 py-0.5 rounded text-xs font-medium bg-blue-50 text-blue-700 border border-blue-100">
                                            {emp.golongan}
                                        </span>
                                    </TableCell>
                                    <TableCell class="text-right pr-4 py-3">
                                        <div class="flex justify-end items-center gap-1">
                                            <button class="p-1.5 rounded-md text-slate-400 hover:text-blue-600 hover:bg-blue-50 transition-colors" title="Edit" on:click={() => openEditModal(emp)}>
                                                <svg xmlns="http://www.w3.org/2000/svg" class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                                                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M11 5H6a2 2 0 00-2 2v11a2 2 0 002 2h11a2 2 0 002-2v-5m-1.414-9.414a2 2 0 112.828 2.828L11.828 15H9v-2.828l8.586-8.586z" />
                                                </svg>
                                            </button>
                                            <button class="p-1.5 rounded-md text-slate-400 hover:text-red-600 hover:bg-red-50 transition-colors" title="Hapus" on:click={() => handleDelete(emp.id)}>
                                                <svg xmlns="http://www.w3.org/2000/svg" class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                                                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16" />
                                                </svg>
                                            </button>
                                        </div>
                                    </TableCell>
                                </TableRow>
                            {/each}
                        {/if}
                    </TableBody>
                </Table>
            </div>
        </div>

        <!-- Mobile Card View -->
        <div class="grid grid-cols-1 gap-4 md:hidden">
            {#if filteredEmployees.length === 0}
                <div class="text-center py-8 text-slate-500 italic bg-white rounded-xl border border-slate-200">Tidak ada data pegawai ditemukan.</div>
            {:else}
                {#each filteredEmployees as emp (emp.id)}
                    <div class="bg-white p-5 rounded-xl border border-slate-200 shadow-sm space-y-3 relative">
                        <div class="absolute top-4 right-4 flex gap-2">
                             <button class="text-slate-400 hover:text-blue-600" on:click={() => openEditModal(emp)}>
                                <svg xmlns="http://www.w3.org/2000/svg" class="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M11 5H6a2 2 0 00-2 2v11a2 2 0 002 2h11a2 2 0 002-2v-5m-1.414-9.414a2 2 0 112.828 2.828L11.828 15H9v-2.828l8.586-8.586z" />
                                </svg>
                             </button>
                             <button class="text-slate-400 hover:text-red-600" on:click={() => handleDelete(emp.id)}>
                                <svg xmlns="http://www.w3.org/2000/svg" class="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16" />
                                </svg>
                             </button>
                        </div>
                        <div>
                            <h3 class="font-bold text-slate-800">{emp.name}</h3>
                            <p class="text-xs font-mono text-slate-500 mt-1">NIP. {emp.nip}</p>
                        </div>
                        <div class="flex gap-4 text-sm text-slate-600 border-t border-slate-50 pt-2">
                            <div>
                                <span class="block text-[10px] text-slate-400 uppercase">Pangkat</span>
                                {emp.rank}
                            </div>
                            <div>
                                <span class="block text-[10px] text-slate-400 uppercase">Golongan</span>
                                {emp.golongan}
                            </div>
                        </div>
                    </div>
                {/each}
            {/if}
        </div>
    {/if}

    <Dialog open={isModalOpen} on:close={() => isModalOpen = false}>
        <DialogHeader class="border-b border-slate-100 pb-4">
            <DialogTitle class="text-xl">{editingId ? 'Edit' : 'Tambah'} Petugas</DialogTitle>
            <p class="text-sm text-slate-500">Lengkapi data diri pegawai.</p>
        </DialogHeader>

        <div class="space-y-4 py-4">
            <div class="space-y-2">
                <Label>Nama Lengkap</Label>
                <Input type="text" placeholder="Contoh: Budi Santoso" bind:value={formData.name} />
            </div>
            <div class="space-y-2">
                <Label>NIP</Label>
                <Input type="text" placeholder="Contoh: 198501012010011001" bind:value={formData.nip} />
            </div>
            <div class="grid grid-cols-2 gap-4">
                <div class="space-y-2">
                    <Label>Pangkat</Label>
                    <Input type="text" placeholder="Contoh: Penata" bind:value={formData.rank} />
                </div>
                <div class="space-y-2">
                    <Label>Golongan</Label>
                    <Input type="text" placeholder="Contoh: III/c" bind:value={formData.golongan} />
                </div>
            </div>
        </div>

        <DialogFooter class="border-t border-slate-100 pt-4 flex justify-end gap-2">
            <Button variant="outline" on:click={() => isModalOpen = false}>Batal</Button>
            <Button class="bg-blue-600 hover:bg-blue-700 text-white" on:click={confirmSubmit}>Simpan</Button>
        </DialogFooter>
    </Dialog>

    <ConfirmationModal
        bind:open={isConfirmOpen}
        title={editingId ? 'Simpan Data Petugas' : 'Tambah Petugas Baru'}
        description="Apakah Anda yakin data petugas ini sudah benar?"
        confirmText="Ya, Simpan"
        onConfirm={processSubmit}
    />
</div>
