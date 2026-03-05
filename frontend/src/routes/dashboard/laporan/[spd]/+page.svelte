<script>
    import { page } from '$app/stores';
    import { recordsStore, updateRecord, loadRecords } from '$lib/stores/records';
    import { userStore } from '$lib/stores/auth';
    import { provincesStore } from '$lib/stores/master-data';
    import { toast } from '$lib/stores/toast';
    import { onMount } from 'svelte';
    import { goto } from '$app/navigation';
    import { fade } from 'svelte/transition';
    import { cn } from '$lib/utils';
    
    // UI
    import Button from '$lib/components/ui/button/Button.svelte';
    import Label from '$lib/components/ui/label/Label.svelte';
    import Textarea from '$lib/components/ui/textarea/Textarea.svelte';
    import Input from '$lib/components/ui/input/Input.svelte';
    import Select from '$lib/components/ui/select/Select.svelte';
    import DocumentViewer from '$lib/components/ui/document-viewer/DocumentViewer.svelte';
    import Dialog from '$lib/components/ui/dialog/Dialog.svelte';

    let spd = $page.params.spd || ''; 
    
    $: allRecordsForSpd = spd ? $recordsStore.filter(r => r.spd === decodeURIComponent(spd)) : [];
    $: record = allRecordsForSpd.find(r => r.email === $userStore.email || (r.employee && r.employee.email === $userStore.email)) || allRecordsForSpd[0];
    $: recordsList = record ? [record] : [];
    $: officerIndex = record ? allRecordsForSpd.findIndex(r => r.id === record.id) : 0;
    $: nomorSpdPetugas = String(officerIndex + 1).padStart(3, '0');

    let reportText = '';
    /** @type {Array<{name: string, type: string, data: string, timestamp: string}>} */
    let uploadedFiles = []; 
    let isDragging = false;
    
    /** @type {{name: string, type: string, data: string, timestamp: string} | null} */
    let sppdFile = null;
    /** @type {{name: string, type: string, data: string, timestamp: string} | null} */
    let suratTugasFile = null;

    // Preview Modal State
    let isPreviewOpen = false;
    /** @type {{name: string, type: string, data: string} | null} */
    let previewFile = null;

    let localCosts = {};

    $: if (recordsList.length > 0 && Object.keys(localCosts).length === 0) {
        recordsList.forEach(r => {
            localCosts[r.id] = { 
                costs: r.costs ? { ...r.costs, additionalCosts: r.costs.additionalCosts || [] } : { additionalCosts: [] }, 
                totalCost: r.totalCost || 0 
            };
        });
        localCosts = { ...localCosts }; // Trigger reactivity
    }

    $: {
        if (Object.keys(localCosts).length > 0) {
            let updated = false;
            for (let r of recordsList) {
                const empId = r.id;
                if (localCosts[empId]) {
                    const costs = localCosts[empId].costs;
                    const days = getDays(r);
                    const sbmRate = getSbmRate(r);
                    const totalDailyAllowance = days * sbmRate;
                    
                    const totalHotel = (costs.hotelDays || 0) * (costs.hotelRate || 0);
                    const totalTicket = Number(costs.ticketGo || 0) + Number(costs.ticketBack || 0);
                    const totalTransportAmount = Number(costs.transportAmount || 0);
                    const totalAdditional = (costs.additionalCosts || []).reduce((sum, cost) => sum + (Number(cost.amount) || 0), 0);
                    
                    const newTotal = totalTicket + totalDailyAllowance + totalHotel + totalTransportAmount + totalAdditional;
                    
                    if (localCosts[empId].totalCost !== newTotal) {
                        localCosts[empId].totalCost = newTotal;
                        updated = true;
                    }
                }
            }
            if (updated) {
                localCosts = { ...localCosts };
            }
        }
    }

    function getDays(empRecord) {
        if (!empRecord?.startDate || !empRecord?.endDate) return 0;
        const start = new Date(empRecord.startDate);
        const end = new Date(empRecord.endDate);
        start.setHours(0,0,0,0);
        end.setHours(0,0,0,0);
        const diffTime = end.getTime() - start.getTime();
        const diffDays = Math.ceil(diffTime / (1000 * 60 * 60 * 24)) + 1; 
        return diffDays > 0 ? diffDays : 0;
    }

    function getSbmRate(empRecord) {
        return $provincesStore.find(p => p.name === empRecord?.province)?.luarKota || 0;
    }

    function formatCurrency(amount) {
        return new Intl.NumberFormat('id-ID', { style: 'currency', currency: 'IDR' }).format(amount || 0);
    }

    function formatInputNumber(value) {
        if (value === undefined || value === null || value === '') return '';
        const numStr = value.toString().replace(/[^0-9]/g, '');
        if (!numStr) return '';
        return parseInt(numStr, 10).toLocaleString('id-ID');
    }

    function updateCost(empId, field, event) {
        const raw = event.target.value.replace(/[^0-9]/g, '');
        const num = parseInt(raw, 10);
        localCosts[empId].costs[field] = isNaN(num) ? undefined : num;
        localCosts = { ...localCosts };
    }

    function updateAdditionalCostAmount(empId, index, event) {
        const raw = event.target.value.replace(/[^0-9]/g, '');
        const num = parseInt(raw, 10);
        localCosts[empId].costs.additionalCosts[index].amount = isNaN(num) ? undefined : num;
        localCosts = { ...localCosts };
    }

    function addAdditionalCost(empId) {
        if (!localCosts[empId].costs.additionalCosts) localCosts[empId].costs.additionalCosts = [];
        localCosts[empId].costs.additionalCosts = [...localCosts[empId].costs.additionalCosts, { name: '', amount: undefined, file: null }];
        localCosts = { ...localCosts };
    }

    function removeAdditionalCost(empId, index) {
        localCosts[empId].costs.additionalCosts = localCosts[empId].costs.additionalCosts.filter((_, i) => i !== index);
        localCosts = { ...localCosts };
    }

    function handleSpecificFileSelect(empId, e, field) {
        const file = e.target.files[0];
        if (!file) return;
        if (file.size > 5 * 1024 * 1024) {
            toast.error(`Ukuran file melebihi 5MB.`);
            e.target.value = '';
            return;
        }
        const reader = new FileReader();
        reader.onload = (ev) => {
            localCosts[empId].costs[field] = {
                name: file.name,
                size: file.size,
                type: file.type,
                data: ev.target.result
            };
            localCosts = { ...localCosts };
        };
        reader.readAsDataURL(file);
        e.target.value = '';
    }

    function removeSpecificFile(empId, field) {
        localCosts[empId].costs[field] = null;
        localCosts = { ...localCosts };
    }

    function handleAdditionalFileSelect(empId, e, index) {
        const file = e.target.files[0];
        if (!file) return;
        if (file.size > 5 * 1024 * 1024) {
            toast.error(`Ukuran file melebihi 5MB.`);
            e.target.value = '';
            return;
        }
        const reader = new FileReader();
        reader.onload = (ev) => {
            localCosts[empId].costs.additionalCosts[index].file = {
                name: file.name,
                size: file.size,
                type: file.type,
                data: ev.target.result
            };
            localCosts = { ...localCosts };
        };
        reader.readAsDataURL(file);
        e.target.value = '';
    }

    function removeAdditionalFile(empId, index) {
        localCosts[empId].costs.additionalCosts[index].file = null;
        localCosts = { ...localCosts };
    }

    function viewFilePreview(file) {
        if (file && file.data) {
            const win = window.open();
            if (win) {
                win.document.write(`<iframe src="${file.data}" frameborder="0" style="border:0; top:0px; left:0px; bottom:0px; right:0px; width:100%; height:100%;" allowfullscreen></iframe>`);
            }
        }
    }

    /** @param {{name: string, type: string, data: string}} file */
    function openPreview(file) {
        previewFile = file;
        isPreviewOpen = true;
    }

    let isDataLoaded = false;
    $: if (record && record.reportData && !isDataLoaded) {
        reportText = record.reportData.text || '';
        uploadedFiles = record.reportData.files || [];
        sppdFile = record.reportData.sppdFile || null;
        suratTugasFile = record.reportData.suratTugasFile || null;
        isDataLoaded = true;
    }

    onMount(async () => {
        if ($recordsStore.length === 0) {
            await loadRecords();
        }
    });

    /** @param {File[]} selectedFiles */
    function processFiles(selectedFiles) {
        if (uploadedFiles.length + selectedFiles.length > 6) {
            toast.error('Maksimal 6 file yang diperbolehkan.');
            return;
        }

        selectedFiles.forEach(file => {
            if (file.size > 10 * 1024 * 1024) {
                toast.error(`File ${file.name} terlalu besar. Maksimal 10MB.`);
                return;
            }

            if (!file.type.startsWith('image/')) {
                toast.error(`File ${file.name} tidak didukung. Hanya file gambar yang diperbolehkan untuk dokumentasi.`);
                return;
            }

            const reader = new FileReader();
            reader.onload = (e) => {
                uploadedFiles = [...uploadedFiles, {
                    name: file.name,
                    type: file.type,
                    data: e.target.result,
                    timestamp: new Date(file.lastModified).toISOString()
                }];
            };
            reader.readAsDataURL(file);
        });
    }
    /** @param {Event} event */
    function handleFileChange(event) {
        // @ts-ignore
        const selectedFiles = Array.from(event.target.files);
        processFiles(selectedFiles);
        event.target.value = ''; 
    }

    /** @param {Event} event */
    function handleSppdFileChange(event) {
        // @ts-ignore
        const file = event.target.files[0];
        if (!file) return;
        if (file.size > 10 * 1024 * 1024) {
            toast.error('File SPPD terlalu besar. Maksimal 10MB.');
            event.target.value = '';
            return;
        }
        const reader = new FileReader();
        reader.onload = (e) => {
            sppdFile = { 
                name: file.name, 
                type: file.type, 
                data: e.target.result, 
                timestamp: new Date(file.lastModified).toISOString()
            };
        };
        reader.readAsDataURL(file);
        event.target.value = '';
    }

    /** @param {Event} event */
    function handleSuratTugasFileChange(event) {
        // @ts-ignore
        const file = event.target.files[0];
        if (!file) return;
        if (file.size > 10 * 1024 * 1024) {
            toast.error('File Surat Tugas terlalu besar. Maksimal 10MB.');
            event.target.value = '';
            return;
        }
        const reader = new FileReader();
        reader.onload = (e) => {
            suratTugasFile = { 
                name: file.name, 
                type: file.type, 
                data: e.target.result, 
                timestamp: new Date(file.lastModified).toISOString()
            };
        };
        reader.readAsDataURL(file);
        event.target.value = '';
    }

    function handleDragOver(e) {
        e.preventDefault();
        isDragging = true;
    }

    /** @param {DragEvent} e */
    function handleDragLeave(e) {
        e.preventDefault();
        isDragging = false;
    }

    /** @param {DragEvent} e */
    function handleDrop(e) {
        e.preventDefault();
        isDragging = false;
        if (e.dataTransfer.files && e.dataTransfer.files.length > 0) {
            const droppedFiles = Array.from(e.dataTransfer.files).filter(f => 
                f.type.startsWith('image/') || f.type === 'application/pdf' || f.type === 'application/vnd.openxmlformats-officedocument.wordprocessingml.document'
            );
            if (droppedFiles.length !== e.dataTransfer.files.length) {
                toast.warning('Beberapa file diabaikan. Hanya gambar, PDF, dan Word yang diperbolehkan.');
            }
            processFiles(droppedFiles);
        }
    }

    /** @param {number} index */
    function removeFile(index) {
        uploadedFiles = uploadedFiles.filter((_, i) => i !== index);
    }

    async function handleSubmit() {
        if (!reportText) {
             toast.error('Isi laporan tidak boleh kosong.');
             return;
        }

        const words = reportText.trim().split(/\s+/).length;
        if (words > 200) {
            toast.error(`Laporan terlalu panjang (${words} kata). Maksimal 2000 kata.`);
            return;
        }

        if (uploadedFiles.length === 0) {
            toast.error('Harap unggah minimal 1 dokumentasi.');
            return;
        }

        if (!recordsList.length) return;

        try {
            const updatePromises = recordsList.map(r => {
                const local = localCosts[r.id];
                const costsToSave = { ...(local?.costs || r.costs) };
                
                // Inject daily allowance details automatically
                costsToSave.dailyAllowanceDays = getDays(r);
                costsToSave.dailyAllowanceRate = getSbmRate(r);

                return updateRecord(r.id, {
                    reportStatus: 'Completed',
                    reportData: {
                        text: reportText,
                        files: uploadedFiles,
                        sppdFile: sppdFile,
                        suratTugasFile: suratTugasFile,
                        submittedAt: new Date().toISOString()
                    },
                    costs: costsToSave,
                    totalCost: local?.totalCost || r.totalCost
                });
            });
            await Promise.all(updatePromises);

            toast.success('Laporan & Rincian Biaya berhasil disimpan!');
            goto('/dashboard/laporan');
        } catch (error) {
            toast.error('Gagal menyimpan laporan. Silakan coba lagi.');
            console.error('Submit report error:', error);
        }
    }
</script>

<div class="max-w-5xl mx-auto pb-20 space-y-8">
    <div class="flex flex-col sm:flex-row sm:items-center justify-between gap-4">
        <div>
            <h2 class="text-2xl sm:text-3xl font-bold tracking-tight text-slate-900">Form Laporan & Rincian Biaya</h2>
            <p class="text-sm sm:text-base text-slate-500">Isi detail laporan, unggah dokumentasi, dan lengkapi rincian biaya secara komprehensif.</p>
        </div>
        <Button variant="outline" class="w-full sm:w-auto border-slate-200 text-slate-600 hover:text-slate-900" on:click={() => goto('/dashboard/laporan')}>
            <svg xmlns="http://www.w3.org/2000/svg" class="h-4 w-4 mr-2" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M10 19l-7-7m0 0l7-7m-7 7h18" />
            </svg>
            Kembali
        </Button>
    </div>

    {#if !record}
        <div class="text-red-500 p-4 border border-red-200 rounded bg-red-50">Data perjalanan dinas tidak ditemukan.</div>
    {:else}
        <div class="bg-white rounded-xl shadow-sm border border-slate-200 overflow-hidden">
            <!-- Header Info -->
            <div class="bg-slate-50/50 border-b border-slate-100 p-4 sm:p-6 grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4 sm:gap-6 text-sm">
                <div>
                    <span class="block text-xs font-semibold uppercase text-slate-400 tracking-wider mb-1">ID SPD</span>
                    <span class="font-mono text-slate-700 font-medium bg-white px-2 py-1 rounded border border-slate-200 inline-block">{record.spd}</span>
                </div>
                <div>
                    <span class="block text-xs font-semibold uppercase text-slate-400 tracking-wider mb-1">NO. SURAT TUGAS</span>
                    <span class="font-mono text-emerald-700 font-bold bg-emerald-50 px-2 py-1 rounded border border-emerald-200 inline-block">{record.suratTugasNumber || '-'}</span>
                </div>
                <div>
                    <span class="block text-xs font-semibold uppercase text-slate-400 tracking-wider mb-1">NO. SPD</span>
                    <span class="font-mono text-blue-700 font-bold bg-blue-50 px-2 py-1 rounded border border-blue-200 inline-block">{nomorSpdPetugas}</span>
                </div>
                <div>
                    <span class="block text-xs font-semibold uppercase text-slate-400 tracking-wider mb-1">Nama Pegawai</span>
                    <span class="font-medium text-slate-800">{record.employee?.name || '-'}</span>
                </div>
                <div>
                    <span class="block text-xs font-semibold uppercase text-slate-400 tracking-wider mb-1">Lokasi & Tujuan</span>
                    <span class="font-medium text-slate-800">{record.location}, {record.province}</span>
                </div>
                <div class="md:col-span-2 lg:col-span-1">
                    <span class="block text-xs font-semibold uppercase text-slate-400 tracking-wider mb-1">Periode Perjalanan</span>
                    <span class="font-medium text-slate-800">{new Date(record.startDate).toLocaleDateString('id-ID', { day: 'numeric', month: 'long', year: 'numeric'})} &mdash; {new Date(record.endDate).toLocaleDateString('id-ID', { day: 'numeric', month: 'long', year: 'numeric'})}</span>
                </div>
            </div>

            <div class="p-4 sm:p-6 md:p-8 space-y-8 sm:space-y-12">
                <!-- Specific Documents Upload (SPPD & Surat Tugas) -->
                {#if $userStore.role !== 'kasubag'}
                <div class="grid grid-cols-1 md:grid-cols-2 gap-6">
                    <div class="space-y-3">
                        <Label class="text-sm font-bold text-slate-800">Upload Lembar SPPD</Label>
                        <div class="p-4 border border-slate-200 rounded-xl bg-slate-50 relative group flex items-center justify-between">
                            {#if sppdFile}
                                <div class="flex items-center gap-3 w-full">
                                    <div class="w-10 h-10 rounded-lg bg-blue-100 flex items-center justify-center shrink-0">
                                        <svg xmlns="http://www.w3.org/2000/svg" class="h-5 w-5 text-blue-600" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z" /></svg>
                                    </div>
                                    <div class="flex-1 min-w-0">
                                        <p class="text-sm font-medium text-slate-700 truncate">{sppdFile.name}</p>
                                        <p class="text-[10px] text-slate-500">Berhasil diunggah</p>
                                    </div>
                                    <div class="flex gap-2">
                                        <button class="p-1.5 text-blue-600 hover:bg-blue-100 rounded-md transition-colors" on:click={() => openPreview(sppdFile)}><svg xmlns="http://www.w3.org/2000/svg" class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 12a3 3 0 11-6 0 3 3 0 016 0z" /><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M2.458 12C3.732 7.943 7.523 5 12 5c4.478 0 8.268 2.943 9.542 7-1.274 4.057-5.064 7-9.542 7-4.477 0-8.268-2.943-9.542-7z" /></svg></button>
                                        <button class="p-1.5 text-red-500 hover:bg-red-100 rounded-md transition-colors" on:click={() => sppdFile = null}><svg xmlns="http://www.w3.org/2000/svg" class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16" /></svg></button>
                                    </div>
                                </div>
                            {:else}
                                <label class="flex-1 text-center cursor-pointer py-2 text-sm text-slate-500 hover:text-blue-600 transition-colors w-full">
                                    <span class="inline-flex items-center gap-2">
                                        <svg xmlns="http://www.w3.org/2000/svg" class="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 16v1a3 3 0 003 3h10a3 3 0 003-3v-1m-4-8l-4-4m0 0L8 8m4-4v12" /></svg>
                                        Klik untuk unggah SPPD (PDF/Gambar)
                                    </span>
                                    <input type="file" accept="image/*,application/pdf" class="hidden" on:change={handleSppdFileChange} />
                                </label>
                            {/if}
                        </div>
                    </div>

                    <div class="space-y-3">
                        <Label class="text-sm font-bold text-slate-800">Upload Surat Tugas</Label>
                        <div class="p-4 border border-slate-200 rounded-xl bg-slate-50 relative group flex items-center justify-between">
                            {#if suratTugasFile}
                                <div class="flex items-center gap-3 w-full">
                                    <div class="w-10 h-10 rounded-lg bg-emerald-100 flex items-center justify-center shrink-0">
                                        <svg xmlns="http://www.w3.org/2000/svg" class="h-5 w-5 text-emerald-600" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z" /></svg>
                                    </div>
                                    <div class="flex-1 min-w-0">
                                        <p class="text-sm font-medium text-slate-700 truncate">{suratTugasFile.name}</p>
                                        <p class="text-[10px] text-slate-500">Berhasil diunggah</p>
                                    </div>
                                    <div class="flex gap-2">
                                        <button class="p-1.5 text-blue-600 hover:bg-blue-100 rounded-md transition-colors" on:click={() => openPreview(suratTugasFile)}><svg xmlns="http://www.w3.org/2000/svg" class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 12a3 3 0 11-6 0 3 3 0 016 0z" /><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M2.458 12C3.732 7.943 7.523 5 12 5c4.478 0 8.268 2.943 9.542 7-1.274 4.057-5.064 7-9.542 7-4.477 0-8.268-2.943-9.542-7z" /></svg></button>
                                        <button class="p-1.5 text-red-500 hover:bg-red-100 rounded-md transition-colors" on:click={() => suratTugasFile = null}><svg xmlns="http://www.w3.org/2000/svg" class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16" /></svg></button>
                                    </div>
                                </div>
                            {:else}
                                <label class="flex-1 text-center cursor-pointer py-2 text-sm text-slate-500 hover:text-emerald-600 transition-colors w-full">
                                    <span class="inline-flex items-center gap-2">
                                        <svg xmlns="http://www.w3.org/2000/svg" class="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 16v1a3 3 0 003 3h10a3 3 0 003-3v-1m-4-8l-4-4m0 0L8 8m4-4v12" /></svg>
                                        Klik untuk unggah Surat Tugas
                                    </span>
                                    <input type="file" accept="image/*,application/pdf" class="hidden" on:change={handleSuratTugasFileChange} />
                                </label>
                            {/if}
                        </div>
                    </div>
                </div>
                {/if}

                <!-- Report Text Section -->
                <div class="space-y-4">
                    <div class="flex justify-between items-center border-b border-slate-100 pb-3">
                        <Label class="text-lg font-bold text-slate-800">Isi Laporan Kegiatan</Label>
                        <span class="text-xs text-slate-400 font-medium px-2 py-1 bg-slate-50 rounded-full border border-slate-200">
                            {reportText.trim().split(/\s+/).filter(w => w.length > 0).length} / 2000 Kata
                        </span>
                    </div>
                    <Textarea 
                        rows="10" 
                        class="resize-y min-h-[150px] text-base leading-relaxed p-4 border-slate-200 focus:border-blue-300 focus:ring-blue-100 placeholder:text-slate-300 shadow-sm"
                        placeholder="Deskripsikan hasil kegiatan, kendala yang dihadapi, dan tindak lanjut yang diperlukan..." 
                        bind:value={reportText} 
                        disabled={$userStore.role === 'kasubag'}
                    />
                    <p class="text-xs text-slate-400 italic">Maksimal 2000 kata. Gunakan bahasa yang baku dan jelas.</p>
                </div>

                <!-- File Upload Section -->
                <div class="space-y-4 pt-4 border-t border-slate-100">
                    <div class="flex justify-between items-center border-b border-slate-100 pb-3">
                         <Label class="text-lg font-bold text-slate-800">Dokumentasi Kegiatan</Label>
                         {#if $userStore.role !== 'kasubag'}
                             <span class="text-xs text-slate-400">Max 6 File (JPG/PDF), Max 10MB</span>
                         {/if}
                    </div>
                    
                    {#if $userStore.role !== 'kasubag'}
                    <label 
                        class="flex flex-col items-center justify-center w-full h-32 border-2 border-dashed rounded-xl cursor-pointer transition-all group relative
                        {isDragging ? 'border-blue-500 bg-blue-50' : 'border-slate-300 bg-slate-50 hover:bg-slate-100 hover:border-blue-400'}"
                        on:dragover={handleDragOver}
                        on:dragleave={handleDragLeave}
                        on:drop={handleDrop}
                    >
                        <div class="flex flex-col items-center justify-center pt-5 pb-6 pointer-events-none">
                            <svg xmlns="http://www.w3.org/2000/svg" class="w-8 h-8 mb-3 text-slate-400 group-hover:text-blue-500 transition-colors" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M7 16a4 4 0 01-.88-7.903A5 5 0 1115.9 6L16 6a5 5 0 011 9.9M15 13l-3-3m0 0l-3 3m3-3v12" />
                            </svg>
                            <p class="mb-1 text-sm text-slate-500"><span class="font-semibold text-blue-600">Klik untuk unggah</span> atau seret file ke sini</p>
                            <p class="text-xs text-slate-400">PNG, JPG (Maks. 10MB)</p>
                        </div>
                        {#if isDragging}
                            <div class="absolute inset-0 flex items-center justify-center bg-blue-50/90 rounded-xl pointer-events-none">
                                <p class="text-blue-600 font-bold text-lg animate-pulse">Lepaskan file di sini</p>
                            </div>
                        {/if}
                        <input type="file" multiple accept="image/*" class="hidden" on:change={handleFileChange} />
                    </label>
                    {/if}
                    
                    {#if uploadedFiles.length > 0}
                        <div class="grid grid-cols-2 sm:grid-cols-3 md:grid-cols-4 gap-4 mt-6">
                            {#each uploadedFiles as file, i}
                                <div transition:fade class="group relative aspect-square bg-slate-100 rounded-lg overflow-hidden border border-slate-200 shadow-sm cursor-pointer" on:click={() => openPreview(file)}>
                                    {#if file.type.startsWith('image/')}
                                        <img src={file.data} alt="Preview" class="object-cover w-full h-full transition-transform duration-500 group-hover:scale-110" />
                                    {:else if file.type === 'application/pdf'}
                                        <div class="flex flex-col items-center justify-center h-full text-red-500 bg-red-50 p-4 text-center group-hover:bg-red-100 transition-colors">
                                            <svg xmlns="http://www.w3.org/2000/svg" class="h-10 w-10 mb-2" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                                                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M7 21h10a2 2 0 002-2V9.414a1 1 0 00-.293-.707l-5.414-5.414A1 1 0 0012.586 3H7a2 2 0 00-2 2v14a2 2 0 002 2z" />
                                            </svg>
                                            <span class="text-[10px] font-bold uppercase tracking-wider text-red-700">PDF Document</span>
                                            <span class="text-[9px] truncate w-full mt-1 text-red-600/70">{file.name}</span>
                                        </div>
                                    {:else if file.type === 'application/vnd.openxmlformats-officedocument.wordprocessingml.document'}
                                        <div class="flex flex-col items-center justify-center h-full text-blue-600 bg-blue-50 p-4 text-center group-hover:bg-blue-100 transition-colors">
                                            <svg xmlns="http://www.w3.org/2000/svg" class="h-10 w-10 mb-2" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                                                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z" />
                                            </svg>
                                            <span class="text-[10px] font-bold uppercase tracking-wider text-blue-700">Word Document</span>
                                            <span class="text-[9px] truncate w-full mt-1 text-blue-600/70">{file.name}</span>
                                        </div>
                                    {:else}
                                        <div class="flex flex-col items-center justify-center h-full text-slate-400 p-4 text-center">
                                            <svg xmlns="http://www.w3.org/2000/svg" class="h-8 w-8 mb-2" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                                                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z" />
                                            </svg>
                                            <span class="text-[10px] truncate w-full">{file.name}</span>
                                        </div>
                                    {/if}
                                    
                                    <div class="absolute inset-0 bg-black/40 opacity-0 group-hover:opacity-100 transition-opacity flex items-center justify-center z-10" on:click|stopPropagation>
                                        <div class="flex gap-2">
                                             <button 
                                                class="bg-white/20 backdrop-blur-md text-white p-2 rounded-full hover:bg-white/40 transform hover:scale-110 transition-all shadow-lg"
                                                on:click|stopPropagation={() => openPreview(file)}
                                                title="Lihat"
                                            >
                                                <svg xmlns="http://www.w3.org/2000/svg" class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                                                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 12a3 3 0 11-6 0 3 3 0 016 0z" />
                                                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M2.458 12C3.732 7.943 7.523 5 12 5c4.478 0 8.268 2.943 9.542 7-1.274 4.057-5.064 7-9.542 7-4.477 0-8.268-2.943-9.542-7z" />
                                                </svg>
                                            </button>
                                            {#if $userStore.role !== 'kasubag'}
                                            <button 
                                                class="bg-red-500 text-white p-2 rounded-full hover:bg-red-600 transform hover:scale-110 transition-all shadow-lg"
                                                on:click|stopPropagation={() => removeFile(i)}
                                                title="Hapus file"
                                            >
                                                <svg xmlns="http://www.w3.org/2000/svg" class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                                                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16" />
                                                </svg>
                                            </button>
                                            {/if}
                                        </div>
                                    </div>

                                    <div class="absolute bottom-2 right-2 bg-black/60 backdrop-blur-[2px] text-white text-[9px] px-1.5 py-0.5 rounded pointer-events-none z-0">
                                        {new Date(file.timestamp).toLocaleDateString()}
                                    </div>
                                </div>
                            {/each}
                        </div>
                    {/if}
                </div>

                <!-- Form Input Rincian Biaya (Integrated from CostModal) -->
                {#if $userStore.role !== 'kasubag'}
                <div class="space-y-6 pt-8 mt-8">
                    <div class="flex flex-col sm:flex-row sm:justify-between sm:items-center gap-4 bg-gradient-to-r from-blue-600 to-indigo-700 p-6 rounded-2xl shadow-lg text-white">
                        <div class="flex items-center gap-4">
                            <div class="p-3 bg-white/20 rounded-xl backdrop-blur-sm">
                                <svg xmlns="http://www.w3.org/2000/svg" class="h-8 w-8 text-white" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 8c-1.657 0-3 .895-3 2s1.343 2 3 2 3 .895 3 2-1.343 2-3 2m0-8c1.11 0 2.08.402 2.599 1M12 8V7m0 1v8m0 0v1m0-1c-1.11 0-2.08-.402-2.599-1M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
                                </svg>
                            </div>
                            <div>
                                <h3 class="text-xl font-bold tracking-tight">Form Input Rincian Biaya Pegawai</h3>
                                <p class="text-blue-100 text-sm mt-1">Lengkapi rincian detail komponen biaya untuk petugas terkait</p>
                            </div>
                        </div>
                    </div>

                    <div class="space-y-8">
                        {#each recordsList as empRecord, index}
                            {@const empId = empRecord.id}
                            {#if localCosts[empId]}
                                <div id="form-rincian-{empId}" class="bg-white rounded-xl border border-slate-200 shadow-sm overflow-hidden scroll-mt-24">
                                    <div class="p-4 md:p-6 bg-slate-50/30 grid gap-4 md:gap-6">
                                        <!-- Mode Transportasi -->
                                        <div class="p-3 md:p-4 bg-white rounded-xl border border-slate-200 shadow-sm w-full space-y-1.5">
                                            <Label class="text-[10px] md:text-xs font-semibold uppercase text-slate-500 tracking-wider">Mode Transportasi</Label>
                                            <Select bind:value={localCosts[empId].costs.transportMode} class="bg-slate-50 border-slate-200 h-9 md:h-10 text-sm">
                                                <option value="Pesawat">Pesawat Udara</option>
                                                <option value="Kendaraan Umum">Kendaraan Umum / Kereta</option>
                                                <option value="Kendaraan Dinas">Kendaraan Dinas</option>
                                            </Select>
                                        </div>

                                        <!-- Uang Harian SBM -->
                                        <div class="p-3 md:p-4 bg-blue-50/50 rounded-xl border border-blue-100 space-y-2.5 md:space-y-3 w-full">
                                            <div class="flex flex-wrap justify-between items-center gap-2">
                                                <h4 class="text-xs md:text-sm font-semibold text-blue-800 flex items-center gap-1.5">
                                                    <svg xmlns="http://www.w3.org/2000/svg" class="h-3.5 w-3.5 md:h-4 md:w-4 text-blue-500" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                                                        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 8v4l3 3m6-3a9 9 0 11-18 0 9 9 0 0118 0z" />
                                                    </svg>
                                                    Uang Harian (SBM)
                                                </h4>
                                                <span class="text-base md:text-lg font-bold text-blue-700">{formatCurrency(getDays(empRecord) * getSbmRate(empRecord))}</span>
                                            </div>
                                            <div class="flex flex-wrap md:flex-nowrap items-center gap-2 md:gap-3 text-xs md:text-sm text-slate-600 bg-white p-2 md:p-3 rounded-lg border border-blue-50/50 shadow-sm w-full">
                                                <div class="flex-none pr-2 border-r border-slate-100">
                                                    <span class="block text-[9px] md:text-[10px] uppercase tracking-wider text-slate-400 font-semibold mb-0.5">Durasi</span>
                                                    <span class="font-medium whitespace-nowrap">{getDays(empRecord)} Hari</span>
                                                </div>
                                                <div class="flex-1 min-w-0 overflow-hidden">
                                                    <span class="block text-[9px] md:text-[10px] uppercase tracking-wider text-slate-400 font-semibold mb-0.5">Rate Estimasi</span>
                                                    <span class="font-medium truncate block w-full">{formatCurrency(getSbmRate(empRecord))} <span class="text-[10px] md:text-xs text-slate-400 font-normal">/ hari</span></span>
                                                </div>
                                            </div>
                                        </div>

                                        <!-- Tiket & Boarding Pass -->
                                        <div class="p-3 md:p-4 bg-white rounded-xl border border-slate-200 shadow-sm space-y-4 w-full">
                                            <h4 class="text-xs md:text-sm font-semibold text-slate-700 flex items-center gap-1.5">
                                                <svg xmlns="http://www.w3.org/2000/svg" class="h-3.5 w-3.5 md:h-4 md:w-4 text-blue-500" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                                                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M3.055 11H5a2 2 0 012 2v1a2 2 0 002 2 2 2 0 012 2v2.945M8 3.935V5.5A2.5 2.5 0 0010.5 8h.5a2 2 0 012 2 2 2 0 104 0 2 2 0 012-2h1.064M15 20.488V18a2 2 0 012-2h3.064M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
                                                </svg>
                                                Tiket & Boarding Pass
                                            </h4>
                                            
                                            <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
                                                <!-- Tiket Berangkat -->
                                                <div class="space-y-2 p-3 border border-slate-100 bg-slate-50 rounded-lg">
                                                    <Label class="text-[10px] md:text-xs font-semibold uppercase text-slate-500 tracking-wider block">Tiket Berangkat</Label>
                                                    <div class="relative">
                                                        <span class="absolute left-2.5 top-2 md:top-2.5 text-slate-400 text-xs md:text-sm">Rp</span>
                                                        <Input type="text" value={formatInputNumber(localCosts[empId].costs.ticketGo)} on:input={(e) => updateCost(empId, 'ticketGo', e)} class="pl-8 md:pl-9 h-9 md:h-10 text-sm bg-white border-slate-200 focus:bg-white" />
                                                    </div>
                                                    {#if !localCosts[empId].costs.ticketGoFile}
                                                        <label class="flex flex-col items-center justify-center w-full h-16 border border-dashed border-slate-300 rounded-lg cursor-pointer bg-white hover:bg-slate-50 hover:border-blue-400 transition-all group mt-1">
                                                            <div class="flex flex-col items-center justify-center pt-2 pb-2 pointer-events-none">
                                                                <svg xmlns="http://www.w3.org/2000/svg" class="w-4 h-4 mb-1 text-slate-400 group-hover:text-blue-500 transition-colors" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                                                                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M7 16a4 4 0 01-.88-7.903A5 5 0 1115.9 6L16 6a5 5 0 011 9.9M15 13l-3-3m0 0l-3 3m3-3v12" />
                                                                </svg>
                                                                <p class="text-[9px] text-slate-500 text-center px-2"><span class="font-semibold text-blue-600">Klik untuk unggah</span> atau seret file</p>
                                                            </div>
                                                            <input type="file" class="hidden" accept=".pdf,.jpg,.jpeg,.png" on:change={(e) => handleSpecificFileSelect(empId, e, 'ticketGoFile')} />
                                                        </label>
                                                    {/if}
                                                    {#if localCosts[empId].costs.ticketGoFile}
                                                        <div class="flex items-center justify-between p-2 mt-2 bg-white border border-slate-200 rounded-md">
                                                            <span class="text-[10px] md:text-xs text-slate-700 truncate mr-2 flex-1">{localCosts[empId].costs.ticketGoFile.name}</span>
                                                            <div class="flex gap-2 shrink-0 text-[10px]">
                                                                <button type="button" class="inline-flex items-center justify-center px-2 py-1 md:px-2.5 md:py-1 bg-blue-50 hover:bg-blue-100 text-blue-700 font-medium rounded-md border border-blue-200 transition-colors" on:click={() => openPreview(localCosts[empId].costs.ticketGoFile)}>Lihat</button>
                                                                <button type="button" class="inline-flex items-center justify-center px-2 py-1 md:px-2.5 md:py-1 bg-red-50 hover:bg-red-100 text-red-600 font-medium rounded-md border border-red-200 transition-colors" on:click={() => removeSpecificFile(empId, 'ticketGoFile')}>Hapus</button>
                                                            </div>
                                                        </div>
                                                    {/if}
                                                </div>

                                                <!-- Tiket Pulang -->
                                                <div class="space-y-2 p-3 border border-slate-100 bg-slate-50 rounded-lg">
                                                    <Label class="text-[10px] md:text-xs font-semibold uppercase text-slate-500 tracking-wider block">Tiket Pulang</Label>
                                                    <div class="relative">
                                                        <span class="absolute left-2.5 top-2 md:top-2.5 text-slate-400 text-xs md:text-sm">Rp</span>
                                                        <Input type="text" value={formatInputNumber(localCosts[empId].costs.ticketBack)} on:input={(e) => updateCost(empId, 'ticketBack', e)} class="pl-8 md:pl-9 h-9 md:h-10 text-sm bg-white border-slate-200 focus:bg-white" />
                                                    </div>
                                                    {#if !localCosts[empId].costs.ticketBackFile}
                                                        <label class="flex flex-col items-center justify-center w-full h-16 border border-dashed border-slate-300 rounded-lg cursor-pointer bg-white hover:bg-slate-50 hover:border-blue-400 transition-all group mt-1">
                                                            <div class="flex flex-col items-center justify-center pt-2 pb-2 pointer-events-none">
                                                                <svg xmlns="http://www.w3.org/2000/svg" class="w-4 h-4 mb-1 text-slate-400 group-hover:text-blue-500 transition-colors" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                                                                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M7 16a4 4 0 01-.88-7.903A5 5 0 1115.9 6L16 6a5 5 0 011 9.9M15 13l-3-3m0 0l-3 3m3-3v12" />
                                                                </svg>
                                                                <p class="text-[9px] text-slate-500 text-center px-2"><span class="font-semibold text-blue-600">Klik untuk unggah</span> atau seret file</p>
                                                            </div>
                                                            <input type="file" class="hidden" accept=".pdf,.jpg,.jpeg,.png" on:change={(e) => handleSpecificFileSelect(empId, e, 'ticketBackFile')} />
                                                        </label>
                                                    {/if}
                                                    {#if localCosts[empId].costs.ticketBackFile}
                                                        <div class="flex items-center justify-between p-2 mt-2 bg-white border border-slate-200 rounded-md">
                                                            <span class="text-[10px] md:text-xs text-slate-700 truncate mr-2 flex-1">{localCosts[empId].costs.ticketBackFile.name}</span>
                                                            <div class="flex gap-2 shrink-0 text-[10px]">
                                                                <button type="button" class="inline-flex items-center justify-center px-2 py-1 md:px-2.5 md:py-1 bg-blue-50 hover:bg-blue-100 text-blue-700 font-medium rounded-md border border-blue-200 transition-colors" on:click={() => openPreview(localCosts[empId].costs.ticketBackFile)}>Lihat</button>
                                                                <button type="button" class="inline-flex items-center justify-center px-2 py-1 md:px-2.5 md:py-1 bg-red-50 hover:bg-red-100 text-red-600 font-medium rounded-md border border-red-200 transition-colors" on:click={() => removeSpecificFile(empId, 'ticketBackFile')}>Hapus</button>
                                                            </div>
                                                        </div>
                                                    {/if}
                                                </div>

                                                <!-- Boarding Pass -->
                                                <div class="space-y-2 p-3 border border-slate-100 bg-slate-50 rounded-lg md:col-span-2 flex flex-col justify-center">
                                                    <Label class="text-[10px] md:text-xs font-semibold uppercase text-slate-500 tracking-wider block">Boarding Pass</Label>
                                                    {#if !localCosts[empId].costs.boardingPassFile}
                                                        <label class="flex flex-col items-center justify-center w-full h-16 border border-dashed border-slate-300 rounded-lg cursor-pointer bg-white hover:bg-slate-50 hover:border-blue-400 transition-all group mt-1">
                                                            <div class="flex flex-col items-center justify-center pt-2 pb-2 pointer-events-none">
                                                                <svg xmlns="http://www.w3.org/2000/svg" class="w-4 h-4 mb-1 text-slate-400 group-hover:text-blue-500 transition-colors" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                                                                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M7 16a4 4 0 01-.88-7.903A5 5 0 1115.9 6L16 6a5 5 0 011 9.9M15 13l-3-3m0 0l-3 3m3-3v12" />
                                                                </svg>
                                                                <p class="text-[9px] text-slate-500 text-center px-2"><span class="font-semibold text-blue-600">Klik untuk unggah</span> atau seret file</p>
                                                            </div>
                                                            <input type="file" class="hidden" accept=".pdf,.jpg,.jpeg,.png" on:change={(e) => handleSpecificFileSelect(empId, e, 'boardingPassFile')} />
                                                        </label>
                                                    {/if}
                                                    {#if localCosts[empId].costs.boardingPassFile}
                                                        <div class="flex items-center justify-between p-2 mt-2 bg-white border border-slate-200 rounded-md w-full">
                                                            <span class="text-[10px] md:text-xs text-slate-700 truncate mr-2 flex-1">{localCosts[empId].costs.boardingPassFile.name}</span>
                                                            <div class="flex gap-2 shrink-0 text-[10px]">
                                                                <button type="button" class="inline-flex items-center justify-center px-2 py-1 md:px-2.5 md:py-1 bg-blue-50 hover:bg-blue-100 text-blue-700 font-medium rounded-md border border-blue-200 transition-colors" on:click={() => openPreview(localCosts[empId].costs.boardingPassFile)}>Lihat</button>
                                                                <button type="button" class="inline-flex items-center justify-center px-2 py-1 md:px-2.5 md:py-1 bg-red-50 hover:bg-red-100 text-red-600 font-medium rounded-md border border-red-200 transition-colors" on:click={() => removeSpecificFile(empId, 'boardingPassFile')}>Hapus</button>
                                                            </div>
                                                        </div>
                                                    {/if}
                                                </div>
                                            </div>
                                        </div>

                                        <!-- Hotel -->
                                        <div class="p-3 md:p-4 bg-white rounded-xl border border-slate-200 shadow-sm space-y-3 w-full">
                                            <div class="flex justify-between items-center mb-1">
                                                <h4 class="text-xs md:text-sm font-semibold text-slate-700 flex items-center gap-1.5 md:gap-2">
                                                    <svg xmlns="http://www.w3.org/2000/svg" class="h-3.5 w-3.5 md:h-4 md:w-4 text-blue-500" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                                                        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 21V5a2 2 0 00-2-2H7a2 2 0 00-2 2v16m14 0h2m-2 0h-5m-9 0H3m2 0h5M9 7h1m-1 4h1m4-4h1m-1 4h1m-5 10v-5a1 1 0 011-1h2a1 1 0 011 1v5m-4 0h4" />
                                                    </svg>
                                                    Penginapan (Hotel)
                                                </h4>
                                            </div>

                                            <div class="grid grid-cols-3 gap-3 md:gap-4 w-full">
                                                <div class="space-y-1.5 col-span-1">
                                                    <Label class="text-[10px] md:text-xs font-semibold uppercase text-slate-500 tracking-wider">Malam</Label>
                                                    <Input type="number" bind:value={localCosts[empId].costs.hotelDays} class="h-9 md:h-10 text-sm bg-slate-50 border-slate-200 px-2" />
                                                </div>
                                                <div class="space-y-1.5 col-span-2">
                                                    <Label class="text-[10px] md:text-xs font-semibold uppercase text-slate-500 tracking-wider">Rate per Malam</Label>
                                                    <div class="relative">
                                                        <span class="absolute left-2.5 top-2 md:top-2.5 text-slate-400 text-xs md:text-sm">Rp</span>
                                                        <Input type="text" value={formatInputNumber(localCosts[empId].costs.hotelRate)} on:input={(e) => updateCost(empId, 'hotelRate', e)} class="pl-8 md:pl-9 h-9 md:h-10 text-sm bg-slate-50 border-slate-200" />
                                                    </div>
                                                </div>
                                            </div>

                                            {#if !localCosts[empId].costs.hotelFile}
                                                <label class="flex flex-col items-center justify-center w-full h-16 border border-dashed border-slate-300 rounded-lg cursor-pointer bg-slate-50 hover:bg-slate-100 hover:border-blue-400 transition-all group">
                                                    <div class="flex flex-col items-center justify-center pt-2 pb-2 pointer-events-none">
                                                        <svg xmlns="http://www.w3.org/2000/svg" class="w-4 h-4 mb-1 text-slate-400 group-hover:text-blue-500 transition-colors" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                                                            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M7 16a4 4 0 01-.88-7.903A5 5 0 1115.9 6L16 6a5 5 0 011 9.9M15 13l-3-3m0 0l-3 3m3-3v12" />
                                                        </svg>
                                                        <p class="text-[9px] text-slate-500 text-center px-2"><span class="font-semibold text-blue-600">Klik untuk unggah kwitansi hotel</span> atau seret file</p>
                                                    </div>
                                                    <input type="file" class="hidden" accept=".pdf,.jpg,.jpeg,.png" on:change={(e) => handleSpecificFileSelect(empId, e, 'hotelFile')} />
                                                </label>
                                            {/if}
                                            
                                            {#if localCosts[empId].costs.hotelFile}
                                                <div class="flex items-center justify-between p-2 mt-2 bg-slate-50 border border-slate-200 rounded-md w-full">
                                                    <span class="text-[10px] md:text-xs text-slate-700 truncate mr-2 flex-1">{localCosts[empId].costs.hotelFile.name}</span>
                                                    <div class="flex gap-2 shrink-0 text-[10px]">
                                                        <button type="button" class="inline-flex items-center justify-center px-2 py-1 md:px-2.5 md:py-1 bg-blue-50 hover:bg-blue-100 text-blue-700 font-medium rounded-md border border-blue-200 transition-colors" on:click={() => openPreview(localCosts[empId].costs.hotelFile)}>Lihat</button>
                                                        <button type="button" class="inline-flex items-center justify-center px-2 py-1 md:px-2.5 md:py-1 bg-red-50 hover:bg-red-100 text-red-600 font-medium rounded-md border border-red-200 transition-colors" on:click={() => removeSpecificFile(empId, 'hotelFile')}>Hapus</button>
                                                    </div>
                                                </div>
                                            {/if}
                                        </div>

                                        <!-- Bukti Transportasi atau Rental -->
                                        <div class="p-3 md:p-4 bg-white rounded-xl border border-slate-200 shadow-sm space-y-3 w-full">
                                            <div class="flex justify-between items-center mb-1">
                                                <h4 class="text-xs md:text-sm font-semibold text-slate-700 flex items-center gap-1.5">
                                                    <svg xmlns="http://www.w3.org/2000/svg" class="h-3.5 w-3.5 md:h-4 md:w-4 text-blue-500" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                                                        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M8 7h12m0 0l-4-4m4 4l-4 4m0 6H4m0 0l4 4m-4-4l4-4" />
                                                    </svg>
                                                    Transport Darat / Rental
                                                </h4>
                                            </div>

                                            <div class="space-y-1.5 w-full">
                                                <Label class="text-[10px] md:text-xs font-semibold uppercase text-slate-500 tracking-wider">Total Biaya (Opsional)</Label>
                                                <div class="relative">
                                                    <span class="absolute left-2.5 top-2 md:top-2.5 text-slate-400 text-xs md:text-sm">Rp</span>
                                                    <Input type="text" value={formatInputNumber(localCosts[empId].costs.transportAmount)} on:input={(e) => updateCost(empId, 'transportAmount', e)} class="pl-8 md:pl-9 h-9 md:h-10 text-sm bg-slate-50 border-slate-200" />
                                                </div>
                                            </div>

                                            {#if !localCosts[empId].costs.transportFile}
                                                <label class="flex flex-col items-center justify-center w-full h-16 border border-dashed border-slate-300 rounded-lg cursor-pointer bg-slate-50 hover:bg-slate-100 hover:border-blue-400 transition-all group">
                                                    <div class="flex flex-col items-center justify-center pt-2 pb-2 pointer-events-none">
                                                        <svg xmlns="http://www.w3.org/2000/svg" class="w-4 h-4 mb-1 text-slate-400 group-hover:text-blue-500 transition-colors" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                                                            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M7 16a4 4 0 01-.88-7.903A5 5 0 1115.9 6L16 6a5 5 0 011 9.9M15 13l-3-3m0 0l-3 3m3-3v12" />
                                                        </svg>
                                                        <p class="text-[9px] text-slate-500 text-center px-2"><span class="font-semibold text-blue-600">Klik untuk unggah bukti transport</span> atau seret file</p>
                                                    </div>
                                                    <input type="file" class="hidden" accept=".pdf,.jpg,.jpeg,.png" on:change={(e) => handleSpecificFileSelect(empId, e, 'transportFile')} />
                                                </label>
                                            {/if}

                                            {#if localCosts[empId].costs.transportFile}
                                                <div class="flex items-center justify-between p-2 mt-2 bg-slate-50 border border-slate-200 rounded-md w-full">
                                                    <span class="text-[10px] md:text-xs text-slate-700 truncate mr-2 flex-1">{localCosts[empId].costs.transportFile.name}</span>
                                                    <div class="flex gap-2 shrink-0 text-[10px]">
                                                        <button type="button" class="inline-flex items-center justify-center px-2 py-1 md:px-2.5 md:py-1 bg-blue-50 hover:bg-blue-100 text-blue-700 font-medium rounded-md border border-blue-200 transition-colors" on:click={() => openPreview(localCosts[empId].costs.transportFile)}>Lihat</button>
                                                        <button type="button" class="inline-flex items-center justify-center px-2 py-1 md:px-2.5 md:py-1 bg-red-50 hover:bg-red-100 text-red-600 font-medium rounded-md border border-red-200 transition-colors" on:click={() => removeSpecificFile(empId, 'transportFile')}>Hapus</button>
                                                    </div>
                                                </div>
                                            {/if}
                                        </div>

                                        <!-- Biaya Tambahan Lainnya -->
                                        <div class="p-3 md:p-4 bg-slate-50 rounded-xl border border-slate-200 shadow-sm space-y-4 w-full">
                                            <div class="flex justify-between items-center border-b border-slate-200 pb-2">
                                                <h4 class="text-xs md:text-sm font-semibold text-slate-700 flex items-center gap-1.5">
                                                    <svg xmlns="http://www.w3.org/2000/svg" class="h-3.5 w-3.5 md:h-4 md:w-4 text-indigo-500" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                                                        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 6v6m0 0v6m0-6h6m-6 0H6" />
                                                    </svg>
                                                    Biaya Tambahan Lainnya
                                                </h4>
                                                <Button size="sm" class="h-8 px-4 text-xs font-bold tracking-wider bg-indigo-600 hover:bg-indigo-700 text-white shadow-md shadow-indigo-500/20 transition-all rounded-lg flex items-center gap-1.5 hover:scale-[1.02]" on:click={() => addAdditionalCost(empId)}>
                                                    <svg xmlns="http://www.w3.org/2000/svg" class="h-4 w-4" viewBox="0 0 20 20" fill="currentColor"><path fill-rule="evenodd" d="M10 3a1 1 0 011 1v5h5a1 1 0 110 2h-5v5a1 1 0 11-2 0v-5H4a1 1 0 110-2h5V4a1 1 0 011-1z" clip-rule="evenodd" /></svg>
                                                    ADD COST
                                                </Button>
                                            </div>

                                            <div class="space-y-3">
                                                {#if localCosts[empId].costs.additionalCosts && localCosts[empId].costs.additionalCosts.length > 0}
                                                    {#each localCosts[empId].costs.additionalCosts as cost, index}
                                                        <div class="bg-white p-3 border border-slate-200 rounded-lg relative group">
                                                            <button type="button" class="absolute -top-2 -right-2 bg-red-100 text-red-600 rounded-full p-1 border border-red-200 hover:bg-red-200 transition-colors shadow-sm" on:click={() => removeAdditionalCost(empId, index)}>
                                                                <svg xmlns="http://www.w3.org/2000/svg" class="h-3 w-3" viewBox="0 0 20 20" fill="currentColor">
                                                                    <path fill-rule="evenodd" d="M4.293 4.293a1 1 0 011.414 0L10 8.586l4.293-4.293a1 1 0 111.414 1.414L11.414 10l4.293 4.293a1 1 0 01-1.414 1.414L10 11.414l-4.293 4.293a1 1 0 01-1.414-1.414L8.586 10 4.293 5.707a1 1 0 010-1.414z" clip-rule="evenodd" />
                                                                </svg>
                                                            </button>
                                                            <div class="grid grid-cols-1 md:grid-cols-2 gap-3">
                                                                <div class="space-y-1.5">
                                                                    <Label class="text-[10px] font-semibold uppercase text-slate-500 tracking-wider">Nama Biaya</Label>
                                                                    <Input type="text" placeholder="Cth: Taksi Bandara" bind:value={localCosts[empId].costs.additionalCosts[index].name} class="h-8 text-xs bg-slate-50 border-slate-200" />
                                                                </div>
                                                                <div class="space-y-1.5">
                                                                    <Label class="text-[10px] font-semibold uppercase text-slate-500 tracking-wider">Nominal</Label>
                                                                    <div class="relative">
                                                                        <span class="absolute left-2 top-1.5 text-slate-400 text-xs">Rp</span>
                                                                        <Input type="text" value={formatInputNumber(cost.amount)} on:input={(e) => updateAdditionalCostAmount(empId, index, e)} class="pl-7 h-8 text-xs bg-slate-50 border-slate-200" />
                                                                    </div>
                                                                </div>
                                                                <div class="md:col-span-2 mt-1 pt-3 border-t border-slate-100">
                                                                    <span class="block text-[10px] font-semibold uppercase text-slate-500 tracking-wider mb-2">Kwitansi / Bukti (PDF/Gambar)</span>
                                                                    {#if !cost.file}
                                                                        <label class="flex flex-col items-center justify-center w-full h-14 border border-dashed border-slate-300 rounded-lg cursor-pointer bg-slate-50 hover:bg-slate-100 hover:border-blue-400 transition-all group">
                                                                            <div class="flex flex-col items-center justify-center pt-1 pb-1 pointer-events-none">
                                                                                <svg xmlns="http://www.w3.org/2000/svg" class="w-3.5 h-3.5 mb-1 text-slate-400 group-hover:text-blue-500 transition-colors" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                                                                                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M7 16a4 4 0 01-.88-7.903A5 5 0 1115.9 6L16 6a5 5 0 011 9.9M15 13l-3-3m0 0l-3 3m3-3v12" />
                                                                                </svg>
                                                                                <p class="text-[9px] text-slate-500 text-center px-2"><span class="font-semibold text-blue-600">Klik unggah</span> atau seret file</p>
                                                                            </div>
                                                                            <input type="file" class="hidden" accept=".pdf,.jpg,.jpeg,.png" on:change={(e) => handleAdditionalFileSelect(empId, e, index)} />
                                                                        </label>
                                                                    {:else}
                                                                        <div class="flex items-center gap-2 bg-slate-50 p-2 rounded border border-slate-200 w-full">
                                                                            <span class="text-[10px] text-slate-700 truncate max-w-[150px] flex-1">{cost.file.name}</span>
                                                                            <div class="flex gap-2 shrink-0 text-[10px]">
                                                                                <button type="button" class="inline-flex items-center justify-center px-2 py-1 md:px-2.5 md:py-1 bg-blue-50 hover:bg-blue-100 text-blue-700 font-medium rounded-md border border-blue-200 transition-colors" on:click={() => openPreview(cost.file)}>Lihat</button>
                                                                                <button type="button" class="inline-flex items-center justify-center px-2 py-1 md:px-2.5 md:py-1 bg-red-50 hover:bg-red-100 text-red-600 font-medium rounded-md border border-red-200 transition-colors" on:click={() => removeAdditionalFile(empId, index)}>Hapus</button>
                                                                            </div>
                                                                        </div>
                                                                    {/if}
                                                                </div>
                                                            </div>
                                                        </div>
                                                    {/each}
                                                {:else}
                                                    <div class="text-center p-4 border border-dashed border-slate-300 rounded-lg text-xs text-slate-500">
                                                        Belum ada biaya tambahan diinputkan.
                                                    </div>
                                                {/if}
                                            </div>
                                        </div>

                                        <!-- Total Keseluruhan Biaya -->
                                        <div class="mt-6 pt-4 border-t border-slate-200">
                                            <div class="flex flex-col sm:flex-row justify-between items-start sm:items-center gap-2 bg-gradient-to-br from-indigo-50 to-blue-50 p-4 rounded-xl border border-indigo-100 shadow-sm">
                                                <div class="flex items-center gap-3">
                                                    <div class="p-2.5 bg-indigo-100 text-indigo-600 rounded-lg">
                                                        <svg xmlns="http://www.w3.org/2000/svg" class="h-6 w-6" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                                                            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M17 9V7a2 2 0 00-2-2H5a2 2 0 00-2 2v6a2 2 0 002 2h2m2 4h10a2 2 0 002-2v-6a2 2 0 00-2-2H9a2 2 0 00-2 2v6a2 2 0 002 2zm7-5a2 2 0 11-4 0 2 2 0 014 0z" />
                                                        </svg>
                                                    </div>
                                                    <div>
                                                        <h4 class="text-sm font-bold text-slate-800">Total Biaya Perjalanan Dinas</h4>
                                                        <p class="text-[10px] text-slate-500 font-medium">Akumulasi seluruh komponen biaya di atas</p>
                                                    </div>
                                                </div>
                                                <div class="mt-2 sm:mt-0 text-left sm:text-right w-full sm:w-auto">
                                                    <span class="block text-2xl md:text-3xl font-black text-transparent bg-clip-text bg-gradient-to-r from-blue-700 to-indigo-600 tracking-tight">
                                                        {formatCurrency(localCosts[empId].totalCost)}
                                                    </span>
                                                </div>
                                            </div>
                                        </div>

                                    </div>
                                </div>
                            {/if}
                        {/each}
                    </div>
                </div>
                {/if}
            </div>

            {#if $userStore.role !== 'kasubag'}
            <div class="bg-slate-50 border-t border-slate-100 p-6 flex justify-end">
                <Button size="lg" class="bg-blue-600 hover:bg-blue-700 text-white shadow-lg shadow-blue-500/20 px-8" on:click={handleSubmit}>Simpan Laporan & Biaya Akhir</Button>
            </div>
            {/if}
        </div>
    {/if}

    <!-- Document Preview Modal -->
    <Dialog open={isPreviewOpen} hideCloseButton={true} on:close={() => isPreviewOpen = false} class="!w-[95vw] md:!w-[90vw] !max-w-6xl !h-[90vh] md:!h-[85vh] !p-0 overflow-hidden rounded-xl shadow-2xl z-[60]">
        <div class="h-full flex flex-col">
            <div class="flex items-center justify-between px-4 md:px-6 py-3 md:py-4 border-b border-slate-100 bg-slate-50/50 flex-none">
                <div class="flex flex-col min-w-0 pr-4">
                    <h3 class="text-base md:text-lg font-bold text-slate-800 tracking-tight truncate">Pratinjau Dokumen</h3>
                    {#if previewFile}
                        <p class="text-[10px] md:text-xs text-slate-500 truncate max-w-[200px] sm:max-w-xs md:max-w-md">{previewFile.name}</p>
                    {/if}
                </div>
                <button type="button" class="p-2 -mr-2 text-slate-400 hover:text-red-500 hover:bg-slate-100 rounded-full transition-colors flex-shrink-0" on:click={() => isPreviewOpen = false}>
                    <svg xmlns="http://www.w3.org/2000/svg" class="h-5 w-5 md:h-6 md:w-6" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
                    </svg>
                </button>
            </div>
            
            <div class="flex-1 overflow-hidden bg-slate-100 relative p-0">
                {#if previewFile}
                    <DocumentViewer 
                        url={previewFile.data} 
                        type={previewFile.type.startsWith('image/') ? 'image' : (previewFile.type === 'application/pdf' ? 'pdf' : 'docx')} 
                        filename={previewFile.name} 
                    />
                {/if}
            </div>
        </div>
    </Dialog>
</div>