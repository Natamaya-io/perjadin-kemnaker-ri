<script>
    import { userStore, usersStore } from '$lib/stores/auth';
    import { provincesStore, stakeholdersStore } from '$lib/stores/master-data';
    import { recordsStore, addRecord } from '$lib/stores/records';
    import { toast } from '$lib/stores/toast';
    import { goto } from '$app/navigation';
    import { page } from '$app/stores';
    
    // Components
    import ProposalHeader from '$lib/components/dashboard/pengajuan/ProposalHeader.svelte';
    import ProposalForm from '$lib/components/dashboard/pengajuan/ProposalForm.svelte';
    import ProposalSidebar from '$lib/components/dashboard/pengajuan/ProposalSidebar.svelte';
    
    import BasicInfoCard from '$lib/components/dashboard/pengajuan/BasicInfoCard.svelte';
    import LocationCard from '$lib/components/dashboard/pengajuan/LocationCard.svelte';
    import StakeholderCard from '$lib/components/dashboard/pengajuan/StakeholderCard.svelte';
    import EmployeeSelectorCard from '$lib/components/dashboard/pengajuan/EmployeeSelectorCard.svelte';
    
    import { ConfirmationModal } from '$lib/components/ui/confirmation-modal';

    $: selectedType = $page.url.searchParams.get('type');

    // Filter users for employee selection (Protokol role only)
    $: protokolOfficers = $usersStore.filter(u => u.role === 'protokol');

    // Form State
    /** 
     * @type {{
     *   startDate: string,
     *   endDate: string,
     *   location: string,
     *   province: string,
     *   purpose: string,
     *   stakeholder: string,
     *   agenda: string,
     *   selectedEmployees: any[]
     * }} 
     */
    let formData = {
        startDate: '',
        endDate: '',
        location: '',
        province: '',
        purpose: 'persiapan', // Default
        stakeholder: '',
        agenda: '',
        selectedEmployees: []
    };

    let isConfirmOpen = false;

    /** @param {CustomEvent} event */
    function toggleEmployee(event) {
        const employeeId = event.detail;
        if (formData.selectedEmployees.includes(employeeId)) {
            formData.selectedEmployees = formData.selectedEmployees.filter(id => id !== employeeId);
        } else {
            formData.selectedEmployees = [...formData.selectedEmployees, employeeId];
        }
    }

    function generateId() {
        const date = new Date().toISOString().slice(0, 10).replace(/-/g, '');
        const random = Math.floor(Math.random() * 1000).toString().padStart(3, '0');
        return `PERDIN-${date}-${random}`;
    }
    
    function handleSubmit() {
        if (formData.selectedEmployees.length > 6) {
            toast.warning('Maksimal 6 Petugas Protokol yang diperbolehkan dalam satu pengajuan.');
            return;
        }

        if (!formData.startDate || !formData.endDate || !formData.province || formData.selectedEmployees.length === 0) {
            toast.error('Harap lengkapi semua field wajib dan pilih minimal satu pegawai.');
            return;
        }

        isConfirmOpen = true;
    }

    async function processSubmit() {
        // Map selected IDs back to full user objects
        const selectedUsers = protokolOfficers.filter(u => formData.selectedEmployees.includes(u.id));

        const tripData = {
            spd: generateId(),
            email: $userStore.email,
            startDate: formData.startDate,
            endDate: formData.endDate,
            location: formData.location,
            province: formData.province,
            purpose: formData.purpose === 'persiapan' ? 'Persiapan dan Pendampingan Kunjungan Kerja' : 'Koordinasi dan Konsultasi Kunjungan Kerja',
            stakeholder: formData.stakeholder,
            agenda: formData.agenda,
            employees: selectedUsers,
            type: selectedType // Save the type as well
        };
        
        try {
            await addRecord(tripData);
            toast.success('Pengajuan Berhasil Disimpan!');
            
            if ($userStore.role === 'super_admin' || $userStore.role === 'keuangan') {
                goto('/dashboard/admin/perdin');
            } else {
                goto('/dashboard');
            }
        } catch (e) {
            toast.error('Gagal menyimpan pengajuan.');
        }
    }

    /** @param {string} type */
    function getLabel(type) {
        if (type === 'dalam_kota') return 'Dalam Kota';
        if (type === 'luar_kota') return 'Luar Kota';
        if (type === 'luar_negeri') return 'Luar Negeri';
        return '';
    }
</script>

<div class="max-w-6xl mx-auto space-y-8 pb-20">
    {#if !selectedType}
        <!-- Selection Screen -->
        <div class="flex flex-col items-center justify-center min-h-[60vh] animate-in fade-in zoom-in duration-300">
            <div class="text-center mb-10 space-y-2">
                <h2 class="text-3xl font-serif font-bold text-slate-800 tracking-tight">Jenis Perjalanan Dinas</h2>
                <p class="text-slate-500 max-w-md mx-auto">Silakan pilih jenis perjalanan dinas yang akan diajukan.</p>
            </div>
            
            <div class="grid grid-cols-1 md:grid-cols-3 gap-6 w-full px-4">
                <!-- Card Dalam Kota (Under Construction) -->
                <button type="button" on:click={() => toast.info('Halaman ini sedang dalam tahap pengembangan dan akan segera tersedia.')} class="group relative flex flex-col items-center p-8 bg-slate-50/50 rounded-2xl shadow-sm border border-slate-200 cursor-not-allowed opacity-80">
                    <div class="h-24 w-24 bg-slate-100 rounded-full flex items-center justify-center mb-6">
                        <svg xmlns="http://www.w3.org/2000/svg" class="h-10 w-10 text-slate-400" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 21V5a2 2 0 00-2-2H7a2 2 0 00-2 2v16m14 0h2m-2 0h-5m-9 0H3m2 0h5M9 7h1m-1 4h1m4-4h1m-1 4h1m-5 10v-5a1 1 0 011-1h2a1 1 0 011 1v5m-4 0h4" />
                        </svg>
                    </div>
                    <h3 class="text-xl font-bold text-slate-700 mb-2">Dalam Kota</h3>
                    <p class="text-sm text-slate-500 text-center leading-relaxed">
                        Perjalanan dinas ke instansi atau lokasi di dalam wilayah kota/kabupaten yang sama atau jarak dekat.
                    </p>
                    <div class="absolute inset-x-0 bottom-0 top-0 flex items-center justify-center bg-white/60 backdrop-blur-[1px] rounded-2xl opacity-0 group-hover:opacity-100 transition-opacity duration-300">
                        <span class="bg-amber-100 text-amber-800 text-xs font-bold px-3 py-1.5 rounded-full border border-amber-200 shadow-sm text-center mx-4">
                            Halaman ini sedang dalam tahap pengembangan dan akan segera tersedia.
                        </span>
                    </div>
                </button>

                <!-- Card Luar Kota -->
                <a href="?type=luar_kota" class="group relative flex flex-col items-center p-8 bg-white rounded-2xl shadow-sm border border-slate-200 hover:border-indigo-500 hover:shadow-lg hover:shadow-indigo-500/10 transition-all duration-300">
                    <div class="h-24 w-24 bg-indigo-50 rounded-full flex items-center justify-center mb-6 group-hover:bg-indigo-100 transition-colors">
                        <svg xmlns="http://www.w3.org/2000/svg" class="h-10 w-10 text-indigo-600" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M3.055 11H5a2 2 0 012 2v1a2 2 0 002 2 2 2 0 012 2v2.945M8 3.935V5.5A2.5 2.5 0 0010.5 8h.5a2 2 0 012 2 2 2 0 104 0 2 2 0 012-2h1.064M15 20.488V18a2 2 0 012-2h3.064M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
                        </svg>
                    </div>
                    <h3 class="text-xl font-bold text-slate-800 mb-2 group-hover:text-indigo-700">Luar Kota</h3>
                    <p class="text-sm text-slate-500 text-center leading-relaxed">
                        Perjalanan dinas ke luar kota/kabupaten atau lintas provinsi yang memerlukan akomodasi menginap.
                    </p>
                    <div class="absolute bottom-6 opacity-0 translate-y-2 group-hover:opacity-100 group-hover:translate-y-0 transition-all duration-300">
                        <span class="text-indigo-600 font-medium text-sm flex items-center">
                            Pilih
                            <svg xmlns="http://www.w3.org/2000/svg" class="h-4 w-4 ml-1" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M17 8l4 4m0 0l-4 4m4-4H3" />
                            </svg>
                        </span>
                    </div>
                     <div class="h-8"></div>
                </a>

                <!-- Card Luar Negeri (Under Construction) -->
                <button type="button" on:click={() => toast.info('Halaman ini sedang dalam tahap pengembangan dan akan segera tersedia.')} class="group relative flex flex-col items-center p-8 bg-slate-50/50 rounded-2xl shadow-sm border border-slate-200 cursor-not-allowed opacity-80">
                    <div class="h-24 w-24 bg-slate-100 rounded-full flex items-center justify-center mb-6">
                        <svg xmlns="http://www.w3.org/2000/svg" class="h-10 w-10 text-slate-400" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M3.055 11H5a2 2 0 012 2v1a2 2 0 002 2 2 2 0 012 2v2.945M8 3.935V5.5A2.5 2.5 0 0010.5 8h.5a2 2 0 012 2 2 2 0 104 0 2 2 0 012-2h1.064M15 20.488V18a2 2 0 012-2h3.064M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
                        </svg>
                    </div>
                    <h3 class="text-xl font-bold text-slate-700 mb-2">Luar Negeri</h3>
                    <p class="text-sm text-slate-500 text-center leading-relaxed">
                        Perjalanan dinas ke luar negeri untuk keperluan tugas negara, konferensi internasional, atau studi banding.
                    </p>
                    <div class="absolute inset-x-0 bottom-0 top-0 flex items-center justify-center bg-white/60 backdrop-blur-[1px] rounded-2xl opacity-0 group-hover:opacity-100 transition-opacity duration-300">
                        <span class="bg-amber-100 text-amber-800 text-xs font-bold px-3 py-1.5 rounded-full border border-amber-200 shadow-sm text-center mx-4">
                            Halaman ini sedang dalam tahap pengembangan dan akan segera tersedia.
                        </span>
                    </div>
                </button>
            </div>
        </div>
    {:else}
        <!-- Back Button & Form -->
        <div class="flex items-center mb-4">
            <a href="/dashboard/pengajuan/new" class="inline-flex items-center text-sm text-slate-500 hover:text-slate-800 transition-colors">
                <svg xmlns="http://www.w3.org/2000/svg" class="h-4 w-4 mr-1" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M10 19l-7-7m0 0l7-7m-7 7h18" />
                </svg>
                Kembali ke Pilihan
            </a>
            <span class="mx-2 text-slate-300">|</span>
            <span class="text-sm font-semibold text-slate-700 bg-slate-100 px-2 py-0.5 rounded">
                {getLabel(selectedType)}
            </span>
        </div>

        <ProposalHeader />
        
        <div class="grid grid-cols-1 lg:grid-cols-3 gap-8">
            <ProposalForm>
                <BasicInfoCard 
                    email={$userStore.email} 
                    bind:startDate={formData.startDate}
                    bind:endDate={formData.endDate}
                />
                
                <LocationCard 
                    bind:location={formData.location}
                    bind:province={formData.province}
                    bind:purpose={formData.purpose}
                    bind:agenda={formData.agenda}
                    provinces={$provincesStore}
                />
            </ProposalForm>

            <ProposalSidebar>
                <StakeholderCard 
                    stakeholders={$stakeholdersStore}
                    bind:selectedStakeholder={formData.stakeholder}
                />
                
                <EmployeeSelectorCard 
                    employees={protokolOfficers}
                    selectedEmployees={formData.selectedEmployees}
                    on:toggle={toggleEmployee}
                    on:submit={handleSubmit}
                />
            </ProposalSidebar>
        </div>

        <ConfirmationModal 
            bind:open={isConfirmOpen}
            title="Simpan Pengajuan"
            description="Apakah Anda yakin data pengajuan ini sudah benar? Setelah disimpan, data akan masuk ke sistem."
            confirmText="Ya, Simpan"
            onConfirm={processSubmit}
        />
    {/if}
</div>
