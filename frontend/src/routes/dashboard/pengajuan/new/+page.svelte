<script>
    import { onMount } from 'svelte';
    import { api } from '$lib/api';
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
    import CostEstimateCard from '$lib/components/dashboard/pengajuan/CostEstimateCard.svelte';
    
    import { ConfirmationModal } from '$lib/components/ui/confirmation-modal';

    $: selectedType = $page.url.searchParams.get('type');

    // Filter users for employee selection (Protokol role only)
    $: protokolOfficers = $usersStore.filter(u => u.role === 'protokol');

    // Form State
    /** 
     * @type {{
     *   startDate: string,
     *   endDate: string,
     *   suratTugas: File | null,
     *   suratTugasNumber: string,
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
        suratTugas: null,
        suratTugasNumber: '',
        location: '',
        province: '',
        purpose: 'persiapan', // Default
        stakeholder: '',
        agenda: '',
        selectedEmployees: []
    };

    let isReadOnly = false;

    let isConfirmOpen = false;

    // Cost Calculations
    $: selectedProvinceData = $provincesStore.find(p => p.name === formData.province);
    $: rate = selectedProvinceData ? selectedProvinceData.luarKota : 0;
    
    $: days = (() => {
        if (!formData.startDate || !formData.endDate) return 0;
        const start = new Date(formData.startDate);
        const end = new Date(formData.endDate);
        const diffTime = end.getTime() - start.getTime();
        const diffDays = Math.ceil(diffTime / (1000 * 60 * 60 * 24)) + 1; 
        return diffDays > 0 ? diffDays : 0;
    })();

        $: totalCost = rate * days * formData.selectedEmployees.length;

    $: disabledEmployeeIds = $recordsStore
        .filter(r => {
             if (r.status === 'Rejected') return false;
             if (!formData.startDate || !formData.endDate) return false;
             const start = new Date(formData.startDate);
             const end = new Date(formData.endDate);
             const rStart = new Date(r.startDate);
             const rEnd = new Date(r.endDate);
             return (rStart <= end && rEnd >= start);
        })
        .map(r => r.employee?.id).filter(Boolean);

    $: {
        if (formData.startDate && formData.endDate) {
            const conflicts = formData.selectedEmployees.filter(id => disabledEmployeeIds.includes(id));
            if (conflicts.length > 0) {
                 formData.selectedEmployees = formData.selectedEmployees.filter(id => !disabledEmployeeIds.includes(id));
                 toast.warning('Beberapa petugas yang dipilih telah dihapus karena jadwal bentrok.');
            }
        }
    }

    /** @param {CustomEvent} event */
    function toggleEmployee(event) {
        if (isReadOnly) return;
        const employeeId = event.detail;
        if (disabledEmployeeIds.includes(employeeId)) {
             toast.error('Petugas ini sedang bertugas pada tanggal tersebut.');
             return;
        }
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

        if (days <= 0) {
            toast.error('Tanggal selesai tidak boleh mendahului tanggal mulai.');
            return;
        }

        isConfirmOpen = true;
    }

    async function processSubmit() {
        // Map selected IDs back to full user objects
        const selectedUsers = protokolOfficers.filter(u => formData.selectedEmployees.includes(u.id));

        let uploadedSuratTugasPath = null;
        if (formData.suratTugas) {
            try {
                toast.info('Mengunggah Surat Tugas...');
                const res = await api.uploadFile(formData.suratTugas);
                uploadedSuratTugasPath = res.path;
            } catch (e) {
                toast.error('Gagal mengunggah Surat Tugas.');
                isConfirmOpen = false;
                return;
            }
        }

        const tripData = {
            spd: generateId(),
            email: $userStore.email,
            startDate: formData.startDate,
            endDate: formData.endDate,
            suratTugasPath: uploadedSuratTugasPath,
            suratTugasNumber: formData.suratTugasNumber,
            location: formData.location,
            province: formData.province,
            purpose: formData.purpose === 'persiapan' ? 'Persiapan dan Pendampingan Kunjungan Kerja' : 'Koordinasi dan Konsultasi Kunjungan Kerja',
            stakeholder: formData.stakeholder,
            agenda: formData.agenda,
            employees: selectedUsers,
            type: selectedType, // Save the type as well
            totalCost: totalCost // Set calculated cost
        };
        
        try {
            await addRecord(tripData);
            toast.success('Pengajuan Berhasil Disimpan!');
            
            // WA Notification is now handled automatically by the backend via Fonnte API.
            
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
    onMount(async () => {
        if ($userStore.role !== 'super_admin' && $userStore.role !== 'kasubag') {
            goto('/dashboard');
            return;
        }
        if ($recordsStore.length === 0) {
            await loadRecords();
        }
    });
</script>

<div class="max-w-6xl mx-auto space-y-8 pb-20">
    {#if !selectedType}
        <!-- Back Button -->
        <div class="w-full mb-4">
            <a href="/dashboard/pengajuan" class="inline-flex items-center px-4 py-2 text-sm font-medium text-slate-700 bg-white border border-slate-300 rounded-xl hover:bg-slate-50 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-blue-500 transition-all shadow-sm">
                <svg xmlns="http://www.w3.org/2000/svg" class="h-4 w-4 mr-2" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M10 19l-7-7m0 0l7-7m-7 7h18" />
                </svg>
                Kembali
            </a>
        </div>

        <!-- Selection Screen -->
        <div class="flex flex-col items-center justify-center min-h-[50vh] md:min-h-[60vh] animate-in fade-in zoom-in duration-300 py-6 md:py-0">
            <div class="text-center mb-8 md:mb-10 space-y-2 px-4">
                <h2 class="text-2xl md:text-3xl font-serif font-bold text-slate-800 tracking-tight">Jenis Perjalanan Dinas</h2>
                <p class="text-sm md:text-base text-slate-500 max-w-md mx-auto">Silakan pilih jenis perjalanan dinas yang akan diajukan.</p>
            </div>
            
            <div class="grid grid-cols-1 md:grid-cols-3 gap-4 md:gap-6 w-full px-0 md:px-4">
                <!-- Card Dalam Kota (Under Construction) -->
                <button type="button" on:click={() => toast.info('Halaman ini sedang dalam tahap pengembangan dan akan segera tersedia.')} class="group relative flex flex-col items-center p-6 md:p-8 bg-slate-50/50 rounded-2xl shadow-sm border border-slate-200 cursor-not-allowed opacity-80">
                    <div class="h-16 w-16 md:h-24 md:w-24 bg-slate-100 rounded-full flex items-center justify-center mb-4 md:mb-6">
                        <svg xmlns="http://www.w3.org/2000/svg" class="h-8 w-8 md:h-10 md:w-10 text-slate-400" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 21V5a2 2 0 00-2-2H7a2 2 0 00-2 2v16m14 0h2m-2 0h-5m-9 0H3m2 0h5M9 7h1m-1 4h1m4-4h1m-1 4h1m-5 10v-5a1 1 0 011-1h2a1 1 0 011 1v5m-4 0h4" />
                        </svg>
                    </div>
                    <h3 class="text-lg md:text-xl font-bold text-slate-700 mb-2">Dalam Kota</h3>
                    <p class="text-xs md:text-sm text-slate-500 text-center leading-relaxed">
                        Perjalanan dinas ke instansi atau lokasi di dalam wilayah kota/kabupaten yang sama atau jarak dekat.
                    </p>
                    <div class="absolute inset-x-0 bottom-0 top-0 flex items-center justify-center bg-white/60 backdrop-blur-[1px] rounded-2xl opacity-0 group-hover:opacity-100 transition-opacity duration-300">
                        <span class="bg-amber-100 text-amber-800 text-xs font-bold px-3 py-1.5 rounded-full border border-amber-200 shadow-sm text-center mx-4">
                            Segera Hadir
                        </span>
                    </div>
                </button>

                <!-- Card Luar Kota -->
                <a href="?type=luar_kota" class="group relative flex flex-col items-center p-6 md:p-8 bg-white rounded-2xl shadow-sm border border-slate-200 hover:border-indigo-500 hover:shadow-lg hover:shadow-indigo-500/10 transition-all duration-300 ring-2 ring-transparent active:ring-indigo-200">
                    <div class="h-16 w-16 md:h-24 md:w-24 bg-indigo-50 rounded-full flex items-center justify-center mb-4 md:mb-6 group-hover:bg-indigo-100 transition-colors">
                        <svg xmlns="http://www.w3.org/2000/svg" class="h-8 w-8 md:h-10 md:w-10 text-indigo-600" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M3.055 11H5a2 2 0 012 2v1a2 2 0 002 2 2 2 0 012 2v2.945M8 3.935V5.5A2.5 2.5 0 0010.5 8h.5a2 2 0 012 2 2 2 0 104 0 2 2 0 012-2h1.064M15 20.488V18a2 2 0 012-2h3.064M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
                        </svg>
                    </div>
                    <h3 class="text-lg md:text-xl font-bold text-slate-800 mb-2 group-hover:text-indigo-700">Luar Kota</h3>
                    <p class="text-xs md:text-sm text-slate-500 text-center leading-relaxed">
                        Perjalanan dinas ke luar kota/kabupaten atau lintas provinsi yang memerlukan akomodasi menginap.
                    </p>
                    <div class="absolute bottom-6 opacity-0 translate-y-2 group-hover:opacity-100 group-hover:translate-y-0 transition-all duration-300 hidden md:block">
                        <span class="text-indigo-600 font-medium text-sm flex items-center">
                            Pilih
                            <svg xmlns="http://www.w3.org/2000/svg" class="h-4 w-4 ml-1" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M17 8l4 4m0 0l-4 4m4-4H3" />
                            </svg>
                        </span>
                    </div>
                     <div class="h-0 md:h-8"></div>
                </a>

                <!-- Card Luar Negeri (Under Construction) -->
                <button type="button" on:click={() => toast.info('Halaman ini sedang dalam tahap pengembangan dan akan segera tersedia.')} class="group relative flex flex-col items-center p-6 md:p-8 bg-slate-50/50 rounded-2xl shadow-sm border border-slate-200 cursor-not-allowed opacity-80">
                    <div class="h-16 w-16 md:h-24 md:w-24 bg-slate-100 rounded-full flex items-center justify-center mb-4 md:mb-6">
                        <svg xmlns="http://www.w3.org/2000/svg" class="h-8 w-8 md:h-10 md:w-10 text-slate-400" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M3.055 11H5a2 2 0 012 2v1a2 2 0 002 2 2 2 0 012 2v2.945M8 3.935V5.5A2.5 2.5 0 0010.5 8h.5a2 2 0 012 2 2 2 0 104 0 2 2 0 012-2h1.064M15 20.488V18a2 2 0 012-2h3.064M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
                        </svg>
                    </div>
                    <h3 class="text-lg md:text-xl font-bold text-slate-700 mb-2">Luar Negeri</h3>
                    <p class="text-xs md:text-sm text-slate-500 text-center leading-relaxed">
                        Perjalanan dinas ke luar negeri untuk keperluan tugas negara, konferensi internasional, atau studi banding.
                    </p>
                    <div class="absolute inset-x-0 bottom-0 top-0 flex items-center justify-center bg-white/60 backdrop-blur-[1px] rounded-2xl opacity-0 group-hover:opacity-100 transition-opacity duration-300">
                        <span class="bg-amber-100 text-amber-800 text-xs font-bold px-3 py-1.5 rounded-full border border-amber-200 shadow-sm text-center mx-4">
                            Segera Hadir
                        </span>
                    </div>
                </button>
            </div>
        </div>
    {:else}
        <!-- Back Button & Form -->
        <div class="flex items-center mb-6 gap-3">
            <a href="/dashboard/pengajuan/new" class="inline-flex items-center px-4 py-2 text-sm font-medium text-slate-700 bg-white border border-slate-300 rounded-xl hover:bg-slate-50 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-blue-500 transition-all shadow-sm">
                <svg xmlns="http://www.w3.org/2000/svg" class="h-4 w-4 mr-2" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M10 19l-7-7m0 0l7-7m-7 7h18" />
                </svg>
                Kembali ke Pilihan
            </a>
            <div class="h-8 w-px bg-slate-200"></div>
            <span class="text-sm font-semibold text-indigo-700 bg-indigo-50 px-3 py-1.5 rounded-lg border border-indigo-100">
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
                    bind:suratTugas={formData.suratTugas}
                    bind:suratTugasNumber={formData.suratTugasNumber}
                    readonly={isReadOnly}
                />
                
                <LocationCard 
                    bind:location={formData.location}
                    bind:province={formData.province}
                    bind:purpose={formData.purpose}
                    bind:agenda={formData.agenda}
                    provinces={$provincesStore}
                    readonly={isReadOnly}
                />
            </ProposalForm>

            <ProposalSidebar>
                <StakeholderCard 
                    stakeholders={$stakeholdersStore}
                    bind:selectedStakeholder={formData.stakeholder}
                    readonly={isReadOnly}
                />
                
                <EmployeeSelectorCard disabledIds={disabledEmployeeIds} 
                    employees={protokolOfficers}
                    selectedEmployees={formData.selectedEmployees}
                    readonly={isReadOnly}
                    on:toggle={toggleEmployee}
                    on:submit={handleSubmit}
                />

                {#if formData.province && days > 0 && formData.selectedEmployees.length > 0}
                    <CostEstimateCard
                        days={days}
                        rate={rate}
                        employeeCount={formData.selectedEmployees.length}
                        totalCost={totalCost}
                        readonly={isReadOnly}
                    />
                {/if}
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



