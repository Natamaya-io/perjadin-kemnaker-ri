<script>
    import { createEventDispatcher } from 'svelte';
    import { portal } from '$lib/actions/portal';
    import { fade, fly } from 'svelte/transition';
    
    import Label from '$lib/components/ui/label/Label.svelte';
    import Input from '$lib/components/ui/input/Input.svelte';
    import Select from '$lib/components/ui/select/Select.svelte';
    import Textarea from '$lib/components/ui/textarea/Textarea.svelte';
    import Button from '$lib/components/ui/button/Button.svelte';
    
    import { updateRecord, deleteRecord, addRecord } from '$lib/stores/records';
    import { usersStore, userStore } from '$lib/stores/auth';
    import { toast } from '$lib/stores/toast';
    import { formatCurrency, getStatusBadge } from '$lib/utils';
    
    export let open = false;
    export let records = []; // All records for the selected SPD
    export let provinces = [];
    export let stakeholders = [];
    
    const dispatch = createEventDispatcher();
    
    let isEditing = false;
    let isLoading = false;
    
    $: protokolOfficers = $usersStore.filter(u => u.role === 'protokol');
    let selectedEmployeeIds = [];
    let initialEmployeeIds = [];
    let prevOpen = false;
    
    let searchQuery = '';
    $: filteredOfficers = protokolOfficers.filter(officer => 
        (officer.name || '').toLowerCase().includes(searchQuery.toLowerCase()) ||
        (officer.nip || '').includes(searchQuery) ||
        (officer.jabatan || '').toLowerCase().includes(searchQuery.toLowerCase())
    );
    
    // Form state initialized when records change
    let formData = {
        startDate: '',
        endDate: '',
        location: '',
        province: '',
        purpose: '',
        stakeholder: '',
        agenda: ''
    };
    
    $: if (open && !prevOpen) {
        prevOpen = true;
        if (records.length > 0) {
            const firstRecord = records[0];
            // Format dates to YYYY-MM-DD for input type="date"
            const formatForInput = (isoString) => {
                if (!isoString) return '';
                const d = new Date(isoString);
                return isNaN(d.getTime()) ? '' : d.toISOString().split('T')[0];
            };
            
            formData = {
                startDate: formatForInput(firstRecord.startDate),
                endDate: formatForInput(firstRecord.endDate),
                location: firstRecord.location || '',
                province: firstRecord.province || '',
                purpose: firstRecord.purpose || '',
                stakeholder: firstRecord.stakeholder || '',
                agenda: firstRecord.agenda || ''
            };
            
            selectedEmployeeIds = records.map(r => r.employee?.id).filter(Boolean);
            initialEmployeeIds = [...selectedEmployeeIds];
            isEditing = false;
        }
    } else if (!open && prevOpen) {
        prevOpen = false;
    }
    
    $: baseRecord = records[0] || {};
    $: isEditable = baseRecord.status === 'Draft' || baseRecord.status === 'Submitted';
    
    // Calculate SBM Cost
    $: selectedProvinceData = provinces.find(p => p.name === formData.province);
    $: rate = selectedProvinceData ? selectedProvinceData.luarKota : 0;
    
    $: days = (() => {
        if (!formData.startDate || !formData.endDate) return 0;
        const start = new Date(formData.startDate);
        const end = new Date(formData.endDate);
        const diffTime = end.getTime() - start.getTime();
        const diffDays = Math.ceil(diffTime / (1000 * 60 * 60 * 24)) + 1; 
        return diffDays > 0 ? diffDays : 0;
    })();
    
    $: totalCostPerPerson = rate * days;
    // When editing, cost is based on new selection. Otherwise based on records length.
    $: currentEmployeeCount = isEditing ? selectedEmployeeIds.length : records.length;
    $: totalSBMCost = totalCostPerPerson * currentEmployeeCount;

    async function handleSave() {
        if (!formData.startDate || !formData.endDate || !formData.province) {
            toast.error('Harap lengkapi field wajib.');
            return;
        }
        
        if (days <= 0) {
            toast.error('Tanggal selesai tidak boleh mendahului tanggal mulai.');
            return;
        }

        if (selectedEmployeeIds.length === 0) {
            toast.error('Pilih minimal satu petugas.');
            return;
        }

        isLoading = true;
        try {
            const commonData = {
                ...formData,
                startDate: new Date(formData.startDate).toISOString(),
                endDate: new Date(formData.endDate).toISOString(),
                totalCost: totalCostPerPerson // Store per-person cost
            };

            // 1. Delete removed employees
            const toDeleteIds = initialEmployeeIds.filter(id => !selectedEmployeeIds.includes(id));
            for (const id of toDeleteIds) {
                const recordToDelete = records.find(r => r.employee.id === id);
                if (recordToDelete) {
                    await deleteRecord(recordToDelete.id);
                }
            }

            // 2. Update existing employees
            const toUpdateIds = initialEmployeeIds.filter(id => selectedEmployeeIds.includes(id));
            for (const id of toUpdateIds) {
                const recordToUpdate = records.find(r => r.employee.id === id);
                if (recordToUpdate) {
                    await updateRecord(recordToUpdate.id, commonData);
                }
            }

            // 3. Add new employees
            const toAddIds = selectedEmployeeIds.filter(id => !initialEmployeeIds.includes(id));
            if (toAddIds.length > 0) {
                const newEmployees = protokolOfficers.filter(u => toAddIds.includes(u.id));
                const tripData = {
                    spd: baseRecord.spd,
                    email: $userStore.email,
                    type: baseRecord.type,
                    ...commonData,
                    employees: newEmployees
                };
                await addRecord(tripData);
            }

            toast.success('Pengajuan berhasil diperbarui.');
            isEditing = false;
            dispatch('saved');
            close();
        } catch (e) {
            toast.error('Gagal memperbarui pengajuan.');
        } finally {
            isLoading = false;
        }
    }
    
    function close() {
        open = false;
        isEditing = false;
        dispatch('close');
    }
</script>

{#if open && records.length > 0}
  <div use:portal>
    <!-- Backdrop -->
    <div transition:fade={{duration: 200}} class="fixed inset-0 z-50 bg-slate-900/60 backdrop-blur-sm" role="button" tabindex="0" aria-label="Close modal" on:click={close} on:keydown={(e) => e.key === 'Escape' && close()}></div>
    
    <!-- Modal Dialog -->
    <div transition:fly={{y: 20, duration: 300}} class="fixed left-[50%] top-[50%] z-50 w-full max-w-4xl translate-x-[-50%] translate-y-[-50%] border border-slate-200 bg-white shadow-2xl sm:rounded-2xl overflow-hidden max-h-[90vh] flex flex-col">
        <!-- Header -->
        <div class="px-6 py-5 border-b border-slate-100 flex justify-between items-start bg-white">
            <div class="flex-1 min-w-0 pr-4">
                <div class="flex flex-wrap items-center gap-2 sm:gap-3 mb-1.5">
                    <h2 class="text-lg sm:text-xl font-bold text-slate-800 leading-tight">Detail Pengajuan</h2>
                    <span class="inline-flex items-center rounded-full px-2.5 py-0.5 text-[10px] font-bold uppercase tracking-wide border {getStatusBadge(baseRecord).class}">
                        {getStatusBadge(baseRecord).label}
                    </span>
                </div>
                <div class="flex flex-wrap items-center gap-2 text-sm text-slate-500">
                    <span class="font-mono text-indigo-600 bg-indigo-50 px-2 py-0.5 rounded border border-indigo-100 text-xs sm:text-sm">{baseRecord.spd}</span>
                    <span class="inline-flex items-center rounded-full px-2 py-0.5 text-[10px] font-bold capitalize tracking-wide bg-slate-100 text-slate-600 border border-slate-200">
                        {baseRecord.type ? baseRecord.type.replace(/_/g, ' ') : 'Dalam Kota'}
                    </span>
                </div>
            </div>
            <button class="rounded-full p-2 hover:bg-slate-100 text-slate-400 hover:text-slate-600 transition-colors" on:click={close} aria-label="Close">
                <svg xmlns="http://www.w3.org/2000/svg" width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" class="h-5 w-5"><line x1="18" x2="6" y1="6" y2="18"></line><line x1="6" x2="18" y1="6" y2="18"></line></svg>
            </button>
        </div>
        
        <!-- Body -->
        <div class="flex-1 overflow-y-auto p-6 bg-slate-50/50">
            <div class="grid grid-cols-1 lg:grid-cols-5 gap-8">
                <!-- Left Column: Details -->
                <div class="lg:col-span-3 space-y-6">
                    <div class="bg-white rounded-xl border border-slate-200 p-5 shadow-sm">
                        <h3 class="text-sm font-bold text-slate-800 flex items-center gap-2 mb-5">
                            <svg xmlns="http://www.w3.org/2000/svg" class="h-4 w-4 text-blue-600" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13 16h-1v-4h-1m1-4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
                            </svg>
                            Informasi Dasar
                        </h3>
                        
                        <div class="space-y-4">
                            <div class="grid grid-cols-1 sm:grid-cols-2 gap-4">
                                <div class="space-y-1.5">
                                    <Label class="text-slate-600 text-xs">Tanggal Mulai <span class="text-red-500">*</span></Label>
                                    <Input type="date" bind:value={formData.startDate} disabled={!isEditing} class={!isEditing ? 'bg-slate-50 border-slate-200 text-slate-700 font-medium opacity-100 cursor-default' : 'bg-white border-blue-200 focus:border-blue-500'} />
                                </div>
                                <div class="space-y-1.5">
                                    <Label class="text-slate-600 text-xs">Tanggal Selesai <span class="text-red-500">*</span></Label>
                                    <Input type="date" bind:value={formData.endDate} disabled={!isEditing} class={!isEditing ? 'bg-slate-50 border-slate-200 text-slate-700 font-medium opacity-100 cursor-default' : 'bg-white border-blue-200 focus:border-blue-500'} />
                                </div>
                            </div>
                            
                            <div class="space-y-1.5">
                                <Label class="text-slate-600 text-xs">Provinsi Tujuan <span class="text-red-500">*</span></Label>
                                <Select bind:value={formData.province} disabled={!isEditing} class={!isEditing ? 'bg-slate-50 border-slate-200 text-slate-700 font-medium opacity-100 cursor-default' : 'bg-white border-blue-200 focus:border-blue-500'}>
                                    <option value="" disabled>Pilih Provinsi</option>
                                    {#each provinces as prov}
                                        <option value={prov.name}>{prov.name}</option>
                                    {/each}
                                </Select>
                            </div>
                            
                            <div class="space-y-1.5">
                                <Label class="text-slate-600 text-xs">Lokasi / Kota Tujuan</Label>
                                <Input bind:value={formData.location} disabled={!isEditing} class={!isEditing ? 'bg-slate-50 border-slate-200 text-slate-700 font-medium opacity-100 cursor-default' : 'bg-white border-blue-200 focus:border-blue-500'} />
                            </div>

                            <div class="space-y-1.5 pt-2">
                                <Label class="text-slate-600 text-xs">Surat Tugas</Label>
                                {#if baseRecord.suratTugasPath}
                                    <div class="flex items-center justify-between p-2.5 rounded-lg border border-slate-200 bg-slate-50">
                                        <div class="flex items-center gap-2 overflow-hidden">
                                            <svg xmlns="http://www.w3.org/2000/svg" class="h-5 w-5 text-rose-500 shrink-0" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                                                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M7 21h10a2 2 0 002-2V9.414a1 1 0 00-.293-.707l-5.414-5.414A1 1 0 0012.586 3H7a2 2 0 00-2 2v14a2 2 0 002 2z" />
                                            </svg>
                                            <span class="text-sm font-medium text-slate-700 truncate" title={baseRecord.suratTugasPath}>{baseRecord.suratTugasPath}</span>
                                        </div>
                                        <a href={`/uploads/${baseRecord.suratTugasPath}`} target="_blank" rel="noopener noreferrer" class="ml-3 shrink-0 inline-flex items-center justify-center px-3 py-1.5 text-xs font-medium text-blue-700 bg-blue-100 hover:bg-blue-200 rounded-md transition-colors">
                                            Lihat Dokumen
                                        </a>
                                    </div>
                                {:else}
                                    <div class="text-xs text-slate-500 italic p-2.5 border border-dashed border-slate-200 rounded-lg bg-slate-50 text-center">
                                        Tidak ada Surat Tugas yang dilampirkan.
                                    </div>
                                {/if}
                            </div>
                        </div>
                    </div>
                    
                    <div class="bg-white rounded-xl border border-slate-200 p-5 shadow-sm">
                        <h3 class="text-sm font-bold text-slate-800 flex items-center gap-2 mb-5">
                            <svg xmlns="http://www.w3.org/2000/svg" class="h-4 w-4 text-amber-600" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 20H5a2 2 0 01-2-2V6a2 2 0 012-2h10a2 2 0 012 2v1m2 13a2 2 0 01-2-2V7m2 13a2 2 0 002-2V9a2 2 0 00-2-2h-2m-4-3H9M7 16h6M7 8h6v4H7V8z" />
                            </svg>
                            Tujuan Kegiatan
                        </h3>
                        
                        <div class="space-y-4">
                            <div class="space-y-1.5">
                                <Label class="text-slate-600 text-xs">Tujuan Perjalanan</Label>
                                <Input bind:value={formData.purpose} disabled={!isEditing} class={!isEditing ? 'bg-slate-50 border-slate-200 text-slate-700 font-medium opacity-100 cursor-default' : 'bg-white border-amber-200 focus:border-amber-500'} />
                            </div>
                            
                            <div class="space-y-1.5">
                                <Label class="text-slate-600 text-xs">Stakeholder / Pejabat (Pendampingan)</Label>
                                <Select bind:value={formData.stakeholder} disabled={!isEditing} class={!isEditing ? 'bg-slate-50 border-slate-200 text-slate-700 font-medium opacity-100 cursor-default' : 'bg-white border-amber-200 focus:border-amber-500'}>
                                    <option value="">Tidak ada/Lainnya</option>
                                    {#each stakeholders as st}
                                        <option value={st}>{st}</option>
                                    {/each}
                                </Select>
                            </div>
                            
                            <div class="space-y-1.5">
                                <Label class="text-slate-600 text-xs">Agenda (Opsional)</Label>
                                <Textarea bind:value={formData.agenda} disabled={!isEditing} class="resize-y min-h-[80px] {!isEditing ? 'bg-slate-50 border-slate-200 text-slate-700 font-medium opacity-100 cursor-default' : 'bg-white border-amber-200 focus:border-amber-500'}" />
                            </div>
                        </div>
                    </div>
                </div>
                
                <!-- Right Column: Employees & Cost -->
                <div class="lg:col-span-2 space-y-6">
                    <div class="bg-white rounded-xl border border-slate-200 p-5 shadow-sm flex flex-col h-full max-h-[350px]">
                        <h3 class="text-sm font-bold text-slate-800 flex items-center justify-between mb-4 pb-3 border-b border-slate-100">
                            <span class="flex items-center gap-2">
                                <svg xmlns="http://www.w3.org/2000/svg" class="h-4 w-4 text-indigo-600" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M17 20h5v-2a3 3 0 00-5.356-1.857M17 20H7m10 0v-2c0-.656-.126-1.283-.356-1.857M7 20H2v-2a3 3 0 015.356-1.857M7 20v-2c0-.656.126-1.283.356-1.857m0 0a5.002 5.002 0 019.288 0M15 7a3 3 0 11-6 0 3 3 0 016 0zm6 3a2 2 0 11-4 0 2 2 0 014 0zM7 10a2 2 0 11-4 0 2 2 0 014 0z" />
                                </svg>
                                Daftar Petugas
                            </span>
                            <span class="bg-indigo-100 text-indigo-700 text-xs font-bold px-2 py-0.5 rounded-full">{currentEmployeeCount}</span>
                        </h3>
                        <div class="overflow-y-auto pr-2 space-y-2 flex-1">
                            {#if !isEditing}
                                {#each records as record}
                                    <div class="p-3 bg-slate-50 rounded-lg border border-slate-100 flex items-center gap-3">
                                        <div class="h-9 w-9 rounded-full bg-indigo-100 border border-indigo-200 flex items-center justify-center text-indigo-700 font-bold text-xs shrink-0">
                                            {record.employee?.name ? (record.employee.name.split(' ').filter(Boolean).map(n=>n[0]).join('').substring(0,2).toUpperCase()) : '?'}
                                        </div>
                                        <div class="min-w-0">
                                            <p class="text-sm font-semibold text-slate-800 truncate">{record.employee?.name || '-'}</p>
                                            {#if (record.employee?.jabatan && record.employee.jabatan !== '-') || (record.employee?.golongan && record.employee.golongan !== '-') || (record.employee?.pangkat && record.employee.pangkat !== '-')}
                                                <p class="text-[11px] text-slate-500 truncate">
                                                    {(record.employee?.jabatan && record.employee.jabatan !== '-') ? record.employee.jabatan : 
                                                     ((record.employee?.pangkat && record.employee.pangkat !== '-') && (record.employee?.golongan && record.employee.golongan !== '-') ? `${record.employee.pangkat} (${record.employee.golongan})` : 
                                                     ((record.employee?.pangkat && record.employee.pangkat !== '-') ? record.employee.pangkat : 
                                                     ((record.employee?.golongan && record.employee.golongan !== '-') ? record.employee.golongan : '')))}
                                                </p>
                                            {:else}
                                                <p class="text-[11px] text-slate-500 truncate">Tidak ada jabatan</p>
                                            {/if}
                                        </div>
                                    </div>
                                {/each}
                            {:else}
                                <div class="mb-3 relative">
                                    <div class="absolute inset-y-0 left-0 pl-2.5 flex items-center pointer-events-none text-slate-400">
                                        <svg xmlns="http://www.w3.org/2000/svg" class="h-3.5 w-3.5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                                            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z" />
                                        </svg>
                                    </div>
                                    <Input type="text" placeholder="Cari nama petugas..." bind:value={searchQuery} class="pl-8 bg-slate-50 border-slate-200 text-xs h-8" />
                                </div>
                                {#if filteredOfficers.length === 0}
                                    <div class="text-center py-4 text-xs text-slate-500 italic">Tidak ada petugas yang cocok.</div>
                                {:else}
                                    {#each filteredOfficers as officer}
                                        <label class="flex items-center space-x-3 p-2.5 rounded-lg border {selectedEmployeeIds.includes(officer.id) ? 'bg-indigo-50/50 border-indigo-200' : 'border-transparent hover:bg-slate-50 hover:border-slate-200'} cursor-pointer transition-all">
                                            <input type="checkbox"
                                                checked={selectedEmployeeIds.includes(officer.id)}
                                                on:change={(e) => {
                                                    if (e.target.checked) selectedEmployeeIds = [...selectedEmployeeIds, officer.id];
                                                    else selectedEmployeeIds = selectedEmployeeIds.filter(id => id !== officer.id);
                                                }}
                                                class="accent-indigo-600 h-4 w-4 rounded border-slate-300 shrink-0"
                                            />
                                            <div class="min-w-0">
                                                <p class="text-sm font-semibold text-slate-800 truncate">{officer.name}</p>
                                                {#if (officer.jabatan && officer.jabatan !== '-') || (officer.golongan && officer.golongan !== '-') || (officer.pangkat && officer.pangkat !== '-')}
                                                    <p class="text-[11px] text-slate-500 truncate">
                                                        {(officer.jabatan && officer.jabatan !== '-') ? officer.jabatan : 
                                                         ((officer.pangkat && officer.pangkat !== '-') && (officer.golongan && officer.golongan !== '-') ? `${officer.pangkat} (${officer.golongan})` : 
                                                         ((officer.pangkat && officer.pangkat !== '-') ? officer.pangkat : 
                                                         ((officer.golongan && officer.golongan !== '-') ? officer.golongan : '')))}
                                                    </p>
                                                {:else}
                                                    <p class="text-[11px] text-slate-500 truncate">Tidak ada jabatan</p>
                                                {/if}
                                            </div>
                                        </label>
                                    {/each}
                                {/if}
                            {/if}
                        </div>
                    </div>
                    
                    <div class="bg-gradient-to-br from-emerald-50 to-teal-50 rounded-xl border border-emerald-100 p-5 shadow-sm relative overflow-hidden">
                        <div class="absolute -right-6 -top-6 text-emerald-500/10">
                            <svg xmlns="http://www.w3.org/2000/svg" class="h-32 w-32" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 8c-1.657 0-3 .895-3 2s1.343 2 3 2 3 .895 3 2-1.343 2-3 2m0-8c1.11 0 2.08.402 2.599 1M12 8V7m0 1v8m0 0v1m0-1c-1.11 0-2.08-.402-2.599-1M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
                            </svg>
                        </div>
                        <h3 class="text-sm font-bold text-emerald-900 mb-4 flex items-center gap-2 relative z-10">
                            Estimasi Total SBM
                        </h3>
                        <div class="space-y-2.5 text-sm text-emerald-800 mb-4 relative z-10">
                            <div class="flex justify-between items-center border-b border-emerald-200/50 pb-2">
                                <span class="text-xs">Durasi Perjalanan</span>
                                <span class="font-semibold bg-white/50 px-2 py-0.5 rounded text-xs">{days} Hari</span>
                            </div>
                            <div class="flex justify-between items-center border-b border-emerald-200/50 pb-2">
                                <span class="text-xs">Tarif / Hari</span>
                                <span class="font-semibold bg-white/50 px-2 py-0.5 rounded text-xs">{formatCurrency(rate)}</span>
                            </div>
                            <div class="flex justify-between items-center border-b border-emerald-200/50 pb-2">
                                <span class="text-xs">Subtotal / Orang</span>
                                <span class="font-semibold bg-white/50 px-2 py-0.5 rounded text-xs">{formatCurrency(totalCostPerPerson)}</span>
                            </div>
                        </div>
                        <div class="pt-2 flex justify-between items-end relative z-10">
                            <span class="font-bold text-emerald-900 text-xs uppercase tracking-wider">Total Akhir<br/><span class="text-[10px] font-normal normal-case text-emerald-700">({currentEmployeeCount} Petugas)</span></span>
                            <span class="text-2xl font-black text-emerald-600">{formatCurrency(totalSBMCost)}</span>
                        </div>
                    </div>
                </div>
            </div>
        </div>
        
        <!-- Footer -->
        <div class="px-6 py-4 border-t border-slate-200 bg-white flex items-center justify-end gap-3 shadow-[0_-4px_6px_-1px_rgba(0,0,0,0.05)] z-10">
            {#if !isEditing}
                <Button variant="outline" on:click={close} class="min-w-[100px]">Tutup</Button>
                {#if isEditable}
                    <Button variant="default" on:click={() => isEditing = true} class="bg-indigo-600 hover:bg-indigo-700 text-white min-w-[120px]">
                        <svg xmlns="http://www.w3.org/2000/svg" class="h-4 w-4 mr-2" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M11 5H6a2 2 0 00-2 2v11a2 2 0 002 2h11a2 2 0 002-2v-5m-1.414-9.414a2 2 0 112.828 2.828L11.828 15H9v-2.828l8.586-8.586z" />
                        </svg>
                        Edit Pengajuan
                    </Button>
                {/if}
            {:else}
                <Button variant="ghost" on:click={() => isEditing = false} disabled={isLoading} class="min-w-[100px] hover:bg-slate-100">Batal</Button>
                <Button variant="default" class="bg-emerald-600 hover:bg-emerald-700 text-white min-w-[140px]" on:click={handleSave} disabled={isLoading}>
                    {#if isLoading}
                        <svg class="animate-spin -ml-1 mr-2 h-4 w-4 text-white" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24">
                            <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle>
                            <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
                        </svg>
                        Menyimpan...
                    {:else}
                        <svg xmlns="http://www.w3.org/2000/svg" class="h-4 w-4 mr-2" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 13l4 4L19 7" />
                        </svg>
                        Simpan Perubahan
                    {/if}
                </Button>
            {/if}
        </div>
    </div>
  </div>
{/if}
