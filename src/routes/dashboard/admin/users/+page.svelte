<script>
    import { usersStore, addUser, updateUser, removeUser } from '$lib/stores/auth';
    import { userStore } from '$lib/stores/auth';
    import { cn } from '$lib/utils';
    
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
    import Select from '$lib/components/ui/select/Select.svelte';
    import { ConfirmationModal } from '$lib/components/ui/confirmation-modal';
    import AlertModal from '$lib/components/ui/alert-modal/AlertModal.svelte';

    let searchQuery = '';
    let isModalOpen = false;
    let isConfirmOpen = false;
    let editingId = null; // null for add mode, number for edit mode

    // Alert & Delete State
    let isAlertOpen = false;
    let alertTitle = '';
    let alertDescription = '';
    let isDeleteConfirmOpen = false;
    let deleteId = null;

    import { toast } from '$lib/stores/toast';

    // Form State
    let formData = {
        name: '',
        email: '',
        password: '',
        role: 'user'
    };

    $: filteredUsers = $usersStore.filter(u => 
        (u.name?.toLowerCase() || '').includes(searchQuery.toLowerCase()) || 
        (u.email?.toLowerCase() || '').includes(searchQuery.toLowerCase())
    );

    function openAddModal() {
        editingId = null;
        formData = { name: '', email: '', password: '', role: 'user' };
        isModalOpen = true;
    }

    function openEditModal(user) {
        editingId = user.id;
        formData = { ...user };
        isModalOpen = true;
    }

    function confirmSubmit() {
        if (!formData.name || !formData.email || !formData.password) {
            toast.error('Semua field wajib diisi');
            return;
        }
        isConfirmOpen = true;
    }

    function processSubmit() {
        if (editingId) {
            updateUser(editingId, formData);
        } else {
            addUser(formData);
        }
        isModalOpen = false;
    }

    function handleDelete(id) {
        deleteId = id;
        isDeleteConfirmOpen = true;
    }

    function processDelete() {
        if (deleteId) {
            removeUser(deleteId);
            deleteId = null;
        }
        isDeleteConfirmOpen = false;
    }
</script>

<div class="space-y-6 pb-20">
    <div class="flex flex-col md:flex-row md:items-center justify-between gap-4">
        <div>
            <h2 class="text-2xl font-bold tracking-tight text-slate-900">Manajemen User</h2>
            <p class="text-slate-500">Kelola akun pengguna dan hak akses aplikasi.</p>
        </div>
        <div class="flex items-center space-x-2 w-full md:w-auto">
            <Button on:click={openAddModal} class="bg-blue-600 hover:bg-blue-700 text-white shadow-lg shadow-blue-500/20">
                <svg xmlns="http://www.w3.org/2000/svg" class="h-4 w-4 mr-2" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4v16m8-8H4" />
                </svg>
                Tambah User
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
            <Input type="text" placeholder="Cari Nama atau Email..." bind:value={searchQuery} class="pl-9 bg-white border-slate-200" />
        </div>

        <!-- Desktop Table View -->
        <div class="hidden md:block rounded-xl border border-slate-200 shadow-sm bg-white overflow-hidden">
            <div class="overflow-x-auto w-full">
                <Table class="w-full text-sm text-left">
                    <TableHeader class="bg-slate-50 border-b border-slate-200">
                        <TableRow class="hover:bg-slate-50/50">
                            <TableHead class="min-w-[200px] font-semibold text-slate-700 pl-4 py-3">Nama User</TableHead>
                            <TableHead class="min-w-[200px] font-semibold text-slate-700 py-3">Email</TableHead>
                            <TableHead class="min-w-[150px] font-semibold text-slate-700 py-3">Role</TableHead>
                            <TableHead class="w-[100px] min-w-[100px] font-semibold text-slate-700 text-right pr-4 py-3">Aksi</TableHead>
                        </TableRow>
                    </TableHeader>
                    <TableBody>
                        {#if filteredUsers.length === 0}
                            <TableRow>
                                <TableCell colspan="4" class="text-center py-12 text-slate-500 italic bg-slate-50/20">
                                    <div class="flex flex-col items-center justify-center gap-2">
                                        <svg xmlns="http://www.w3.org/2000/svg" class="h-8 w-8 text-slate-300" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                                            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4.354a4 4 0 110 5.292M15 21H3v-1a6 6 0 0112 0v1zm0 0h6v-1a6 6 0 00-9-5.197M13 7a4 4 0 11-8 0 4 4 0 018 0z" />
                                        </svg>
                                        <span>Tidak ada user ditemukan.</span>
                                    </div>
                                </TableCell>
                            </TableRow>
                        {:else}
                            {#each filteredUsers as user (user.id)}
                                <TableRow class="hover:bg-slate-50 transition-colors border-b border-slate-100 last:border-0">
                                    <TableCell class="pl-4 py-3 font-medium text-slate-900">{user.name}</TableCell>
                                    <TableCell class="py-3 text-slate-600">{user.email}</TableCell>
                                    <TableCell class="py-3">
                                        <span class={cn("inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-bold uppercase tracking-wide border",
                                            user.role === 'super_admin' ? "bg-purple-50 text-purple-700 border-purple-200" :
                                            user.role === 'keuangan' ? "bg-indigo-50 text-indigo-700 border-indigo-200" :
                                            user.role === 'ppk' ? "bg-amber-50 text-amber-700 border-amber-200" :
                                            "bg-blue-50 text-blue-700 border-blue-200")}>
                                            {user.role}
                                        </span>
                                    </TableCell>
                                    <TableCell class="text-right pr-4 py-3">
                                        <div class="flex justify-end items-center gap-1">
                                            <button class="p-1.5 rounded-md text-slate-400 hover:text-blue-600 hover:bg-blue-50 transition-colors" title="Edit" on:click={() => openEditModal(user)}>
                                                <svg xmlns="http://www.w3.org/2000/svg" class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                                                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M11 5H6a2 2 0 00-2 2v11a2 2 0 002 2h11a2 2 0 002-2v-5m-1.414-9.414a2 2 0 112.828 2.828L11.828 15H9v-2.828l8.586-8.586z" />
                                                </svg>
                                            </button>
                                            <button class="p-1.5 rounded-md text-slate-400 hover:text-red-600 hover:bg-red-50 transition-colors" title="Hapus" on:click={() => handleDelete(user.id)}>
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
            {#if filteredUsers.length === 0}
                <div class="text-center py-8 text-slate-500 italic bg-white rounded-xl border border-slate-200">Tidak ada user ditemukan.</div>
            {:else}
                {#each filteredUsers as user (user.id)}
                    <div class="bg-white p-5 rounded-xl border border-slate-200 shadow-sm space-y-3 relative">
                        <div class="absolute top-4 right-4 flex gap-2">
                             <button class="text-slate-400 hover:text-blue-600" on:click={() => openEditModal(user)}>
                                <svg xmlns="http://www.w3.org/2000/svg" class="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M11 5H6a2 2 0 00-2 2v11a2 2 0 002 2h11a2 2 0 002-2v-5m-1.414-9.414a2 2 0 112.828 2.828L11.828 15H9v-2.828l8.586-8.586z" />
                                </svg>
                             </button>
                             <button class="text-slate-400 hover:text-red-600" on:click={() => handleDelete(user.id)}>
                                <svg xmlns="http://www.w3.org/2000/svg" class="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16" />
                                </svg>
                             </button>
                        </div>
                        <div>
                            <h3 class="font-bold text-slate-800">{user.name}</h3>
                            <p class="text-xs font-mono text-slate-500 mt-1">{user.email}</p>
                        </div>
                        <div class="flex gap-4 text-sm text-slate-600 border-t border-slate-50 pt-2">
                             <span class={cn("px-2 py-0.5 rounded-full text-[10px] font-bold uppercase tracking-wide",
                                user.role === 'super_admin' ? "bg-purple-100 text-purple-700" :
                                user.role === 'keuangan' ? "bg-indigo-100 text-indigo-700" :
                                user.role === 'ppk' ? "bg-amber-100 text-amber-700" :
                                "bg-blue-100 text-blue-700")}>
                                {user.role}
                            </span>
                        </div>
                    </div>
                {/each}
            {/if}
        </div>
    {/if}

    <Dialog open={isModalOpen} on:close={() => isModalOpen = false}>
        <DialogHeader class="border-b border-slate-100 pb-4">
            <DialogTitle class="text-xl">{editingId ? 'Edit' : 'Tambah'} User</DialogTitle>
            <p class="text-sm text-slate-500">Lengkapi data akun pengguna.</p>
        </DialogHeader>

        <div class="space-y-4 py-4">
            <div class="space-y-2">
                <Label>Nama Lengkap</Label>
                <Input type="text" placeholder="Contoh: Staf Pengaju" bind:value={formData.name} />
            </div>
            <div class="space-y-2">
                <Label>Email Kedinasan</Label>
                <Input type="email" placeholder="nama@kemnaker.go.id" bind:value={formData.email} />
            </div>
            <div class="space-y-2">
                <Label>Password</Label>
                <Input type="text" placeholder="Minimal 6 karakter" bind:value={formData.password} />
            </div>
            <div class="space-y-2">
                <Label>Role (Hak Akses)</Label>
                <div class="relative w-full">
                    <Select bind:value={formData.role} class="bg-white border-slate-200">
                        <option value="user">User (Staf Pengaju)</option>
                        <option value="super_admin">Super Admin</option>
                        <option value="keuangan">Admin Keuangan</option>
                        <option value="ppk">Pejabat Pembuat Komitmen (PPK)</option>
                    </Select>
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
        title={editingId ? 'Simpan Perubahan User' : 'Tambah User Baru'}
        description="Apakah Anda yakin data user ini sudah benar?"
        confirmText="Ya, Simpan"
        onConfirm={processSubmit}
    />
    
    <ConfirmationModal
        bind:open={isDeleteConfirmOpen}
        title="Hapus User"
        description="Apakah Anda yakin ingin menghapus user ini? Tindakan ini tidak dapat dibatalkan."
        confirmText="Ya, Hapus"
        onConfirm={processDelete}
        variant="destructive"
    />

    <AlertModal
        bind:open={isAlertOpen}
        title={alertTitle}
        description={alertDescription}
    />
</div>
