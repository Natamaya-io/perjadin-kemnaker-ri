<script>
    function getDisplayName(name) {
        if (!name) return '';
        const parts = name.split('/');
        const cleanName = parts[parts.length - 1];
        if (cleanName.length > 36 && cleanName.includes('_')) {
            const splitName = cleanName.split('_');
            if (splitName[0].length === 36 || splitName[0].length === 32) {
                return 'Surat_Tugas_' + splitName.slice(1).join('_');
            }
        }
        return cleanName;
    }
    import { page } from '$app/stores';
    import { recordsStore, updateRecord, loadRecords } from '$lib/features/pengajuan/store';
    import { userStore } from '$lib/features/auth/store';
    import { provincesStore } from '$lib/shared/stores/master-data';
    import { toast } from '$lib/shared/stores/toast';
    import { onMount } from 'svelte';
    import { goto } from '$app/navigation';
    import { getInitials, toTitleCase } from '$lib/shared/utils/utils';
    import { fade } from 'svelte/transition';
    import { cn } from '$lib/shared/utils/utils';
    
    // UI
    import Button from '$lib/shared/ui/button/Button.svelte';
    import Label from '$lib/shared/ui/label/Label.svelte';
    import Textarea from '$lib/shared/ui/textarea/Textarea.svelte';
    import Input from '$lib/shared/ui/input/Input.svelte';
    import Select from '$lib/shared/ui/select/Select.svelte';
    import DocumentViewer from '$lib/shared/ui/document-viewer/DocumentViewer.svelte';
    import Dialog from '$lib/shared/ui/dialog/Dialog.svelte';

    let spd = $page.params.spd || ''; 
    $: currentTab = $page.url.searchParams.get('tab') || 'laporan';
    
    $: allRecordsForSpd = spd ? $recordsStore.filter(r => r.spd === decodeURIComponent(spd)) : [];
    $: record = allRecordsForSpd.find(r => r.email === $userStore.email || (r.employee && r.employee.email === $userStore.email)) || allRecordsForSpd[0];
    
    // Check if the current user is part of this SPD group
    $: isUserInSpdGroup = allRecordsForSpd.some(r => r.email === $userStore.email || (r.employee && r.employee.email === $userStore.email));
    
    // Allow access to all records if user is in the group (representative) or is admin/kasubag
    $: recordsList = ($userStore.role === 'super_admin' || $userStore.role === 'kasubag' || $userStore.role === 'protokol' || isUserInSpdGroup) ? allRecordsForSpd : [];
    
    $: allRecordsSorted = [...$recordsStore].sort((a, b) => new Date(a.createdAt).getTime() - new Date(b.createdAt).getTime());
    $: recordToIndexMap = new Map(allRecordsSorted.map((r, i) => [r.id, i + 1]));

    $: officerIndex = record ? allRecordsForSpd.findIndex(r => r.id === record.id) : 0;
    $: nomorSpdPetugas = record ? String(recordToIndexMap.get(record.id) || 0).padStart(3, '0') : '000';

    let reportText = '';
    let manualSuratTugasNumber = '';
    let manualSuratTugasDate = '';
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
            const locations = getLocations(r);
            let details = r.costs?.details || [];
            
            // Migration for legacy data if details is empty but costs exist
            if (details.length === 0 && r.costs && (r.costs.hotelDays || r.costs.transportAmount || r.costs.ticketGo)) {
                 details = [{ ...r.costs, details: undefined }];
            }
            
            // Ensure details array matches locations length
            if (details.length < locations.length) {
                for (let i = details.length; i < locations.length; i++) {
                    details.push({
                        transportMode: 'Pesawat',
                        ticketGo: 0,
                        ticketBack: 0,
                        hotelDays: 0,
                        hotelRate: 0,
                        transportAmount: 0,
                        additionalCosts: [],
                        boardingPassFiles: []
                    });
                }
            }

            localCosts[r.id] = { 
                costs: { 
                    ...r.costs, 
                    details: details 
                }, 
                totalCost: r.totalCost || 0,
                selectedLocationIndex: 0
            };
        });
        localCosts = { ...localCosts }; // Trigger reactivity
    }

    $: {
        if (Object.keys(localCosts).length > 0) {
            let updated = false;
            for (let r of recordsList) {
                const empId = r.id;
                if (localCosts[empId] && localCosts[empId].costs.details) {
                    const locations = getLocations(r);
                    let grandTotal = 0;

                    // Calculate total across all locations
                    localCosts[empId].costs.details.forEach((detail, idx) => {
                        const loc = locations[idx];
                        if (!loc) return;
                        
                        const days = getDaysForLocation(loc);
                        const rate = getRateForLocation(loc); // SBM Rate
                        const sbmTotal = days * rate;

                        const totalHotel = (detail.hotelDays || 0) * (detail.hotelRate || 0);
                        const totalTicket = Number(detail.ticketGo || 0) + Number(detail.ticketBack || 0);
                        const totalTransportAmount = Number(detail.transportAmount || 0);
                        const totalAdditional = (detail.additionalCosts || []).reduce((sum, cost) => sum + (Number(cost.amount) || 0), 0);
                        
                        grandTotal += sbmTotal + totalHotel + totalTicket + totalTransportAmount + totalAdditional;
                    });
                    
                    if (localCosts[empId].totalCost !== grandTotal) {
                        localCosts[empId].totalCost = grandTotal;
                        updated = true;
                    }
                }
            }
            if (updated) {
                localCosts = { ...localCosts };
            }
        }
    }

	function getTotalDays(empRecord) {
    	const locs = empRecord?.locations && empRecord.locations.length > 0 
       	 ? empRecord.locations 
        	: [{ startDate: empRecord?.startDate, endDate: empRecord?.endDate, province: empRecord?.province }];
    
    	return locs.reduce((total, loc) => {
        	const start = new Date(loc.startDate);
        	const end = new Date(loc.endDate);
        	let locDays = 0;
        	if (!isNaN(start.getTime()) && !isNaN(end.getTime())) {
            	const diffTime = end.getTime() - start.getTime();
            	const diffDays = Math.ceil(diffTime / (1000 * 60 * 60 * 24)) + 1;
            	locDays = diffDays > 0 ? diffDays : 0;
        	}
        	return total + locDays;
    	}, 0);
	}

    function getSbmRate(empRecord) {
        const provData = $provincesStore.find(p => p.name === empRecord.province);
        return provData ? provData.luarKota : 0;
    }

	function getTotalDailyAllowance(empRecord) {
		const locs = empRecord?.locations && empRecord.locations.length > 0 
	   	 ? empRecord.locations 
	    	: [{ startDate: empRecord?.startDate, endDate: empRecord?.endDate, province: empRecord?.province }];

		return locs.reduce((total, loc) => {
	    	const provData = $provincesStore.find(p => p.name === loc.province);
	    	const rate = provData ? provData.luarKota : 0;
	    	const start = new Date(loc.startDate);
	    	const end = new Date(loc.endDate);
	    	let locDays = 0;
	    	if (!isNaN(start.getTime()) && !isNaN(end.getTime())) {
	        	const diffTime = end.getTime() - start.getTime();
	        	const diffDays = Math.ceil(diffTime / (1000 * 60 * 60 * 24)) + 1;
	        	locDays = diffDays > 0 ? diffDays : 0;
	    	}
	    	return total + (rate * locDays);
		}, 0);
	}

	function getLocations(empRecord) {
	     return empRecord?.locations && empRecord.locations.length > 0
	        ? empRecord.locations
	        : [{ startDate: empRecord?.startDate, endDate: empRecord?.endDate, province: empRecord?.province, location: empRecord?.location }];
	}
	function getDaysForLocation(loc) {
	    if (!loc?.startDate || !loc?.endDate) return 0;
	    const start = new Date(loc.startDate);
	    const end = new Date(loc.endDate);
	    let locDays = 0;
	    if (!isNaN(start.getTime()) && !isNaN(end.getTime())) {
	        const diffTime = end.getTime() - start.getTime();
	        const diffDays = Math.ceil(diffTime / (1000 * 60 * 60 * 24)) + 1;
	        locDays = diffDays > 0 ? diffDays : 0;
	    }
	    return locDays;
	}

	function getRateForLocation(loc) {
	     const provData = $provincesStore.find(p => p.name === loc.province);
	     return provData ? provData.luarKota : 0;
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

    function updateCost(empId, field, event, locationIndex) {
        const raw = event.target.value.replace(/[^0-9]/g, '');
        const num = parseInt(raw, 10);
        localCosts[empId].costs.details[locationIndex][field] = isNaN(num) ? undefined : num;
        localCosts = { ...localCosts };
    }

    function updateAdditionalCostAmount(empId, index, event, locationIndex) {
        const raw = event.target.value.replace(/[^0-9]/g, '');
        const num = parseInt(raw, 10);
        localCosts[empId].costs.details[locationIndex].additionalCosts[index].amount = isNaN(num) ? undefined : num;
        localCosts = { ...localCosts };
    }

    function addAdditionalCost(empId, locationIndex) {
        if (!localCosts[empId].costs.details[locationIndex].additionalCosts) localCosts[empId].costs.details[locationIndex].additionalCosts = [];
        localCosts[empId].costs.details[locationIndex].additionalCosts = [...localCosts[empId].costs.details[locationIndex].additionalCosts, { name: '', amount: undefined, file: null }];
        localCosts = { ...localCosts };
    }

    function removeAdditionalCost(empId, index, locationIndex) {
        localCosts[empId].costs.details[locationIndex].additionalCosts = localCosts[empId].costs.details[locationIndex].additionalCosts.filter((_, i) => i !== index);
        localCosts = { ...localCosts };
    }

    function handleSpecificFileSelect(empId, e, field, locationIndex) {
        const file = e.target.files[0];
        if (!file) return;
        if (file.size > 5 * 1024 * 1024) {
            toast.error(`Ukuran file melebihi 5MB.`);
            e.target.value = '';
            return;
        }
        const reader = new FileReader();
        reader.onload = (ev) => {
            localCosts[empId].costs.details[locationIndex][field] = {
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

    function handleBoardingPassFileSelect(empId, e, locationIndex) {
        const file = e.target.files[0];
        if (!file) return;
        if (file.size > 5 * 1024 * 1024) {
            toast.error(`Ukuran file melebihi 5MB.`);
            e.target.value = '';
            return;
        }
        
        const detail = localCosts[empId].costs.details[locationIndex];
        if (!detail.boardingPassFiles) {
            detail.boardingPassFiles = [];
            if (detail.boardingPassFile) {
                detail.boardingPassFiles.push(detail.boardingPassFile);
                delete detail.boardingPassFile;
            }
        }

        const reader = new FileReader();
        reader.onload = (ev) => {
            localCosts[empId].costs.details[locationIndex].boardingPassFiles.push({
                name: file.name,
                size: file.size,
                type: file.type,
                data: ev.target.result
            });
            localCosts = { ...localCosts };
        };
        reader.readAsDataURL(file);
        e.target.value = '';
    }

    function removeBoardingPassFile(empId, index, locationIndex) {
        if (localCosts[empId].costs.details[locationIndex].boardingPassFiles) {
            localCosts[empId].costs.details[locationIndex].boardingPassFiles = localCosts[empId].costs.details[locationIndex].boardingPassFiles.filter((_, i) => i !== index);
        }
        localCosts = { ...localCosts };
    }

    function removeSpecificFile(empId, field, locationIndex) {
        localCosts[empId].costs.details[locationIndex][field] = null;
        localCosts = { ...localCosts };
    }

    function handleAdditionalFileSelect(empId, e, index, locationIndex) {
        const file = e.target.files[0];
        if (!file) return;
        if (file.size > 5 * 1024 * 1024) {
            toast.error(`Ukuran file melebihi 5MB.`);
            e.target.value = '';
            return;
        }
        const reader = new FileReader();
        reader.onload = (ev) => {
            localCosts[empId].costs.details[locationIndex].additionalCosts[index].file = {
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

    function removeAdditionalFile(empId, index, locationIndex) {
        localCosts[empId].costs.details[locationIndex].additionalCosts[index].file = null;
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
    $: if (record && !manualSuratTugasNumber && record.suratTugasNumber) { manualSuratTugasNumber = record.suratTugasNumber; }
    $: if (record && !manualSuratTugasDate && record.suratTugasDate && record.suratTugasDate !== '0001-01-01T00:00:00Z') { 
        manualSuratTugasDate = new Date(record.suratTugasDate).toISOString().split('T')[0]; 
    }
    $: if (record && record.reportData && !isDataLoaded) {
        reportText = record.reportData.text || '';
        uploadedFiles = record.reportData.files || [];
        sppdFile = record.reportData.sppdFile || null;
        suratTugasFile = record.reportData.suratTugasFile || null;
        manualSuratTugasNumber = record.suratTugasNumber || '';
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

    async function saveSuratTugasInfoOnly() {
        if (!manualSuratTugasNumber) {
            toast.error('Nomor Surat Tugas tidak boleh kosong.');
            return;
        }

        try {
            const updatePromises = recordsList.map(r => {
                return updateRecord(r.id, {
                    suratTugasNumber: manualSuratTugasNumber,
                    suratTugasDate: manualSuratTugasDate ? new Date(manualSuratTugasDate).toISOString() : undefined
                });
            });
            await Promise.all(updatePromises);

            toast.success('Informasi Surat Tugas berhasil diperbarui.');
            // Refresh record logic if needed, but the store usually handles it
        } catch (error) {
            toast.error('Gagal memperbarui informasi Surat Tugas.');
            console.error('Save ST info error:', error);
        }
    }

    async function saveDraftLaporan() {
        if (!recordsList.length) return;
        
        const words = reportText.trim().split(/\s+/).length;
        if (reportText && words > 2000) {
            toast.error(`Laporan terlalu panjang (${words} kata). Maksimal 2000 kata.`);
            return;
        }

        try {
            const updatePromises = recordsList.map(r => {
                return updateRecord(r.id, {
                    reportStatus: 'Draft',
                    suratTugasNumber: manualSuratTugasNumber,
                    suratTugasDate: manualSuratTugasDate ? new Date(manualSuratTugasDate).toISOString() : undefined,
                    reportData: {
                        text: reportText,
                        files: uploadedFiles,
                        sppdFile: sppdFile,
                        suratTugasFile: suratTugasFile,
                        lastDraftSavedAt: new Date().toISOString()
                    }
                });
            });
            await Promise.all(updatePromises);

            toast.success('Draf Laporan Kegiatan berhasil disimpan!');
            // Stay on the same page when saving as draft
        } catch (error) {
            toast.error('Gagal menyimpan draf laporan. Silakan coba lagi.');
            console.error('Save draft report error:', error);
        }
    }

    async function submitLaporan() {
        if (!reportText) {
             toast.error('Isi laporan tidak boleh kosong.');
             return;
        }

        const words = reportText.trim().split(/\s+/).length;
        if (words > 2000) {
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
                return updateRecord(r.id, {
                    reportStatus: 'Completed',
                    suratTugasNumber: manualSuratTugasNumber,
                    suratTugasDate: manualSuratTugasDate ? new Date(manualSuratTugasDate).toISOString() : undefined,
                    reportData: {
                        text: reportText,
                        files: uploadedFiles,
                        sppdFile: sppdFile,
                        suratTugasFile: suratTugasFile,
                        submittedAt: new Date().toISOString()
                    }
                });
            });
            await Promise.all(updatePromises);

            toast.success('Laporan Kegiatan berhasil disubmit!');
            goto('/dashboard/laporan');
        } catch (error) {
            toast.error('Gagal mensubmit laporan. Silakan coba lagi.');
            console.error('Submit report error:', error);
        }
    }

    async function saveDraftRincian() {
        if (!recordsList.length) return;

        try {
            const updatePromises = recordsList.map(r => {
                const local = localCosts[r.id];
                const costsToSave = { ...local.costs };
                
                // Note: We are saving the full structure including 'details'
                costsToSave.lastDraftSavedAt = new Date().toISOString();

                return updateRecord(r.id, {
                    costs: costsToSave,
                    totalCost: local.totalCost
                });
            });
            await Promise.all(updatePromises);

            toast.success('Draf Rincian Biaya berhasil disimpan!');
            // Stay on the same page when saving as draft
        } catch (error) {
            toast.error('Gagal menyimpan draf rincian biaya. Silakan coba lagi.');
            console.error('Save draft costs error:', error);
        }
    }

    async function submitRincian() {
        if (!recordsList.length) return;

        try {
            const updatePromises = recordsList.map(r => {
                const local = localCosts[r.id];
                const costsToSave = { ...local.costs };
                
                // Inject daily allowance details automatically (already done via getLocations iteration but redundant safe-check)
                costsToSave.dailyAllowanceDays = getTotalDays(r);
                costsToSave.dailyAllowanceRate = getSbmRate(r);

                return updateRecord(r.id, {
                    costs: costsToSave,
                    totalCost: local.totalCost,
                    status: 'Submitted'
                });
            });
            await Promise.all(updatePromises);

            toast.success('Rincian Biaya berhasil disubmit!');
        } catch (error) {
            toast.error('Gagal menyimpan rincian biaya. Silakan coba lagi.');
            console.error('Submit costs error:', error);
        }
    }
</script>

<div class="max-w-5xl mx-auto pb-20 space-y-8">
    <div class="flex flex-col sm:flex-row sm:items-center justify-between gap-4">
        <div>
            <h2 class="text-2xl sm:text-3xl font-bold tracking-tight text-slate-900">Form {currentTab === 'laporan' ? 'Laporan Kegiatan' : 'Rincian Biaya'}</h2>
            <p class="text-sm sm:text-base text-slate-500">
                {currentTab === 'laporan' ? 'Isi detail laporan dan unggah dokumentasi kegiatan perjalanan dinas.' : 'Lengkapi rincian komponen biaya untuk petugas terkait secara komprehensif.'}
            </p>
        </div>
        <Button variant="outline" class="w-full sm:w-auto border-slate-200 text-slate-600 hover:text-slate-900" on:click={() => goto('/dashboard/laporan')}>
            <svg xmlns="http://www.w3.org/2000/svg" class="h-4 w-4 mr-2" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M10 19l-7-7m0 0l7-7m-7 7h18" />
            </svg>
            Kembali
        </Button>
    </div>

    <!-- Tab Navigation & Actions -->
    <div class="flex flex-col sm:flex-row items-start sm:items-center justify-between gap-4">
        <div class="flex items-center gap-2 p-1.5 bg-slate-100/80 rounded-xl border border-slate-200 w-full sm:w-max">
            <a href="?tab=laporan" class="flex-1 sm:flex-none">
                <button class={cn("w-full px-6 py-2.5 rounded-lg text-sm font-semibold transition-all flex items-center justify-center gap-2", currentTab === 'laporan' ? "bg-white text-blue-600 shadow-sm border border-slate-200" : "text-slate-500 hover:text-slate-700 hover:bg-slate-200/50")}>
                    <svg xmlns="http://www.w3.org/2000/svg" class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z" /></svg>
                    Laporan Kegiatan
                </button>
            </a>
            <a href="?tab=rincian" class="flex-1 sm:flex-none">
                <button class={cn("w-full px-6 py-2.5 rounded-lg text-sm font-semibold transition-all flex items-center justify-center gap-2", currentTab === 'rincian' ? "bg-white text-indigo-600 shadow-sm border border-slate-200" : "text-slate-500 hover:text-slate-700 hover:bg-slate-200/50")}>
                    <svg xmlns="http://www.w3.org/2000/svg" class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 8c-1.657 0-3 .895-3 2s1.343 2 3 2 3 .895 3 2-1.343 2-3 2m0-8c1.11 0 2.08.402 2.599 1M12 8V7m0 1v8m0 0v1m0-1c-1.11 0-2.08-.402-2.599-1M21 12a9 9 0 11-18 0 9 9 0 0118 0z" /></svg>
                    Rincian Biaya
                </button>
            </a>
        </div>

        <!-- Print Button -->
        {#if record}
            <a href={`/print?type=gabungan&id=${record.id}&spd=${encodeURIComponent(record.spd)}`} target="_blank" class="w-full sm:w-auto">
                <button class="w-full sm:w-auto px-4 py-2.5 bg-white border border-slate-200 text-slate-700 font-semibold rounded-xl hover:bg-slate-50 hover:text-blue-600 transition-colors shadow-sm flex items-center justify-center gap-2">
                    <svg xmlns="http://www.w3.org/2000/svg" class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M17 17h2a2 2 0 002-2v-4a2 2 0 00-2-2H5a2 2 0 00-2 2v4a2 2 0 002 2h2m2 4h6a2 2 0 002-2v-4a2 2 0 00-2-2H9a2 2 0 00-2 2v4a2 2 0 002 2zm8-12V5a2 2 0 00-2-2H9a2 2 0 00-2 2v4h10z" />
                    </svg>
                    Cetak Dokumen
                </button>
            </a>
        {/if}
    </div>

    {#if !record}
        <div class="text-red-500 p-4 border border-red-200 rounded bg-red-50">Data perjalanan dinas tidak ditemukan.</div>
    {:else}
        <!-- Header Info -->
        {#if currentTab === 'laporan'}
            <div class="bg-white rounded-xl shadow-sm border border-slate-200 overflow-hidden mb-8">
                <!-- ... existing header code ... -->
                <div class="p-4 sm:p-6 border-b border-slate-100">
                    <div class="grid grid-cols-1 lg:grid-cols-2 gap-8 mb-8">
                         <!-- Left: SPD Details -->
                         <div class="space-y-6">
                             <div class="flex items-start gap-4">
                                 <div class="p-3 bg-blue-50 text-blue-600 rounded-xl shrink-0 border border-blue-100">
                                     <svg xmlns="http://www.w3.org/2000/svg" class="h-6 w-6" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                                         <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z" />
                                     </svg>
                                 </div>
                                 <div>
                                     <h3 class="text-xs font-bold text-slate-400 uppercase tracking-widest mb-1">ID SPD</h3>
                                     <div class="text-xl font-bold text-slate-900 tracking-tight">{record.spd}</div>
                                 </div>
                             </div>

                             <!-- Grid for Surat Tugas & Date -->
                             <div class="grid grid-cols-1 md:grid-cols-2 gap-4 pt-2">
                                 <div>
                                     <div class="flex justify-between items-center mb-1.5">
                                         <span class="block text-xs font-bold text-slate-400 uppercase tracking-widest">Nomor Surat Tugas</span>
                                         {#if ($userStore.role === 'super_admin' || $userStore.role === 'protokol') && (record.reportStatus !== 'Completed') && (manualSuratTugasNumber !== (record.suratTugasNumber || ''))}
                                             <button 
                                                 on:click={saveSuratTugasInfoOnly}
                                                 class="text-[10px] text-blue-600 hover:text-blue-700 font-bold uppercase tracking-tight flex items-center gap-1 bg-blue-50 px-2 py-0.5 rounded-md border border-blue-100"
                                             >
                                                 Simpan
                                             </button>
                                         {/if}
                                     </div>
                                     {#if ($userStore.role === 'super_admin' || $userStore.role === 'protokol') && (record.reportStatus !== 'Completed')}
                                         <input 
                                             type="text" 
                                             bind:value={manualSuratTugasNumber}
                                             placeholder="Input No. Surat Tugas..."
                                             class="font-semibold text-slate-800 bg-white px-3 py-1.5 rounded-lg border border-slate-200 focus:ring-2 focus:ring-blue-100 focus:border-blue-400 outline-none w-full text-sm transition-all shadow-sm"
                                         />
                                     {:else}
                                         <div class="font-medium text-slate-800 bg-slate-50 px-3 py-1.5 rounded-lg border border-slate-100 inline-block text-sm">
                                             {manualSuratTugasNumber || record.suratTugasNumber || '-'}
                                         </div>
                                     {/if}
                                 </div>
                                 <div>
                                     <div class="flex justify-between items-center mb-1.5">
                                         <span class="block text-xs font-bold text-slate-400 uppercase tracking-widest">Tanggal Surat Tugas</span>
                                         {#if ($userStore.role === 'super_admin' || $userStore.role === 'protokol') && (record.reportStatus !== 'Completed') && (manualSuratTugasDate !== (record.suratTugasDate && record.suratTugasDate !== '0001-01-01T00:00:00Z' ? new Date(record.suratTugasDate).toISOString().split('T')[0] : ''))}
                                             <button 
                                                 on:click={saveSuratTugasInfoOnly}
                                                 class="text-[10px] text-blue-600 hover:text-blue-700 font-bold uppercase tracking-tight flex items-center gap-1 bg-blue-50 px-2 py-0.5 rounded-md border border-blue-100"
                                             >
                                                 Simpan
                                             </button>
                                         {/if}
                                     </div>
                                     {#if ($userStore.role === 'super_admin' || $userStore.role === 'protokol') && (record.reportStatus !== 'Completed')}
                                         <input 
                                             type="date" 
                                             bind:value={manualSuratTugasDate}
                                             class="font-semibold text-slate-800 bg-white px-3 py-1.5 rounded-lg border border-slate-200 focus:ring-2 focus:ring-blue-100 focus:border-blue-400 outline-none w-full text-sm transition-all shadow-sm"
                                         />
                                     {:else}
                                         <div class="font-medium text-slate-800 bg-slate-50 px-3 py-1.5 rounded-lg border border-slate-100 inline-block text-sm">
                                             {manualSuratTugasDate ? new Date(manualSuratTugasDate).toLocaleDateString('id-ID', { day: '2-digit', month: 'long', year: 'numeric' }) : '-'}
                                         </div>
                                     {/if}
                                 </div>
                             </div>
                         </div>

                         <!-- Right: Location & Date -->
                         <div class="bg-slate-50/50 rounded-2xl p-6 border border-slate-200/60 flex flex-col justify-center space-y-6 relative overflow-hidden">
                             <div class="absolute top-0 right-0 w-32 h-32 bg-blue-50 rounded-full blur-3xl -mr-16 -mt-16 opacity-50"></div>
                             
                             <div class="flex items-start gap-4 relative z-10">
                                 <div class="p-2 bg-white rounded-lg shadow-sm text-red-500 border border-slate-100 shrink-0">
                                    <svg xmlns="http://www.w3.org/2000/svg" class="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                                        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M17.657 16.657L13.414 20.9a1.998 1.998 0 01-2.827 0l-4.244-4.243a8 8 0 1111.314 0z" />
                                        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 11a3 3 0 11-6 0 3 3 0 016 0z" />
                                    </svg>
                                 </div>
                                 <div>
                                     <span class="block text-xs font-bold text-slate-400 uppercase tracking-widest mb-1">Lokasi Tujuan</span>
                                     {#each getLocations(record) as loc}
                                         <div class="font-bold text-lg text-slate-800 leading-snug mb-1 last:mb-0">
                                             {loc.location ? toTitleCase(loc.location) + ', ' : ''}{toTitleCase(loc.province)}
                                         </div>
                                     {/each}
                                 </div>
                             </div>
                             
                             <div class="flex items-start gap-4 pt-4 border-t border-slate-200 relative z-10">
                                 <div class="p-2 bg-white rounded-lg shadow-sm text-indigo-500 border border-slate-100 shrink-0">
                                    <svg xmlns="http://www.w3.org/2000/svg" class="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                                        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M8 7V3m8 4V3m-9 8h10M5 21h14a2 2 0 002-2V7a2 2 0 00-2-2H5a2 2 0 00-2 2v12a2 2 0 002 2z" />
                                    </svg>
                                 </div>
                                 <div>
                                     <span class="block text-xs font-bold text-slate-400 uppercase tracking-widest mb-1">Periode Perjalanan</span>
                                     <div class="font-medium text-slate-700 text-sm">
                                         <span class="font-semibold text-slate-900">{new Date(record.startDate).toLocaleDateString('id-ID', { day: '2-digit', month: 'long', year: 'numeric' })}</span> 
                                         <span class="text-slate-400 mx-1.5">s/d</span> 
                                         <span class="font-semibold text-slate-900">{new Date(record.endDate).toLocaleDateString('id-ID', { day: '2-digit', month: 'long', year: 'numeric' })}</span>
                                     </div>
                                 </div>
                             </div>
                         </div>
                    </div>

                    <!-- Divider -->
                    <div class="border-t border-slate-100 mb-8"></div>

                    <!-- Bottom Section: Employee List -->
                    <div class="mb-8">
                        <h4 class="text-sm font-bold text-slate-700 uppercase tracking-wide mb-5 flex items-center gap-2">
                            <div class="w-1 h-5 bg-blue-500 rounded-full"></div>
                            Daftar Pegawai ({allRecordsForSpd.length})
                        </h4>
                        
                        <div class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-4">
                            {#each allRecordsForSpd as emp, i}
                                <div class="flex items-center gap-3 p-3.5 rounded-xl border border-slate-200 hover:border-blue-300 hover:bg-blue-50/30 transition-all group bg-white shadow-sm hover:shadow-md">
                                     <!-- Initial Avatar -->
                                     <div class="w-10 h-10 rounded-full bg-slate-100 text-slate-600 font-bold flex items-center justify-center border border-slate-200 group-hover:bg-white group-hover:text-blue-600 group-hover:border-blue-200 transition-colors shadow-sm shrink-0">
                                         {getInitials(emp.employee?.name || '-')}
                                     </div>
                                     <div class="flex-1 min-w-0">
                                         <div class="text-sm font-bold text-slate-800 truncate leading-tight mb-1">{emp.employee?.name || '-'}</div>
                                         <div class="text-xs text-slate-500 flex items-center gap-1.5">
                                             <span class="bg-slate-50 px-2 py-0.5 rounded text-[10px] font-mono font-medium text-slate-500 border border-slate-200 group-hover:border-blue-200 group-hover:text-blue-600 transition-colors">SPD {String(recordToIndexMap.get(emp.id) || 0).padStart(3, '0')}</span>
                                         </div>
                                     </div>
                                </div>
                            {/each}
                        </div>
                    </div>

                    <div class="grid grid-cols-1 md:grid-cols-2 gap-4 mt-8">
                        <!-- Upload SPPD & Surat Tugas -->
                        <!-- ... (same as before) ... -->
                        <div class="relative">
                        {#if sppdFile}
                            <div class="flex items-center justify-between p-3 border border-slate-200 rounded-xl bg-slate-50">
                                <div class="flex items-center gap-3 w-full">
                                    <div class="w-8 h-8 rounded bg-blue-100 flex items-center justify-center shrink-0">
                                        <svg xmlns="http://www.w3.org/2000/svg" class="h-4 w-4 text-blue-600" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z" /></svg>
                                    </div>
                                    <div class="flex-1 min-w-0">
                                        <p class="text-sm font-medium text-slate-700 truncate">{sppdFile.name}</p>
                                    </div>
                                    <div class="flex gap-2">
                                        <button type="button" class="p-1.5 text-blue-600 hover:bg-blue-100 rounded-md transition-colors" on:click={() => openPreview(sppdFile)}><svg xmlns="http://www.w3.org/2000/svg" class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 12a3 3 0 11-6 0 3 3 0 016 0z" /><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M2.458 12C3.732 7.943 7.523 5 12 5c4.478 0 8.268 2.943 9.542 7-1.274 4.057-5.064 7-9.542 7-4.477 0-8.268-2.943-9.542-7z" /></svg></button>
                                        <button type="button" class="p-1.5 text-red-500 hover:bg-red-100 rounded-md transition-colors" on:click={() => sppdFile = null}><svg xmlns="http://www.w3.org/2000/svg" class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16" /></svg></button>
                                    </div>
                                </div>
                            </div>
                        {:else}
                            <label class="flex items-center justify-center p-3.5 border border-slate-200 rounded-xl cursor-pointer hover:bg-slate-50 transition-colors w-full group">
                                <span class="inline-flex items-center gap-2 text-slate-600 group-hover:text-blue-600 transition-colors text-sm font-semibold">
                                    <span class="flex items-center justify-center w-6 h-6 bg-blue-50 text-blue-600 rounded-md group-hover:bg-blue-600 group-hover:text-white transition-colors shadow-sm">
                                        <svg xmlns="http://www.w3.org/2000/svg" class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 16v1a3 3 0 003 3h10a3 3 0 003-3v-1m-4-8l-4-4m0 0L8 8m4-4v12" /></svg>
                                    </span>
                                    Klik untuk unggah SPPD (PDF/Gambar)
                                </span>
                                <input type="file" accept="image/*,application/pdf" class="hidden" on:change={handleSppdFileChange} />
                            </label>
                        {/if}
                    </div>

                    <!-- Upload Surat Tugas -->
                    <div class="relative">
                        {#if suratTugasFile}
                            <!-- ... -->
                             <div class="flex items-center justify-between p-3 border border-slate-200 rounded-xl bg-slate-50">
                                <div class="flex items-center gap-3 w-full">
                                    <div class="w-8 h-8 rounded bg-emerald-100 flex items-center justify-center shrink-0">
                                        <svg xmlns="http://www.w3.org/2000/svg" class="h-4 w-4 text-emerald-600" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z" /></svg>
                                    </div>
                                    <div class="flex-1 min-w-0">
                                        <p class="text-sm font-medium text-slate-700 truncate" title={getDisplayName(suratTugasFile.name)}>{getDisplayName(suratTugasFile.name)}</p>
                                    </div>
                                    <div class="flex gap-2">
                                        <button type="button" class="p-1.5 text-blue-600 hover:bg-blue-100 rounded-md transition-colors" on:click={() => openPreview(suratTugasFile)}><svg xmlns="http://www.w3.org/2000/svg" class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 12a3 3 0 11-6 0 3 3 0 016 0z" /><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M2.458 12C3.732 7.943 7.523 5 12 5c4.478 0 8.268 2.943 9.542 7-1.274 4.057-5.064 7-9.542 7-4.477 0-8.268-2.943-9.542-7z" /></svg></button>
                                        <button type="button" class="p-1.5 text-red-500 hover:bg-red-100 rounded-md transition-colors" on:click={() => suratTugasFile = null}><svg xmlns="http://www.w3.org/2000/svg" class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16" /></svg></button>
                                    </div>
                                </div>
                            </div>
                        {:else}
                            <label class="flex items-center justify-center p-3.5 border border-slate-200 rounded-xl cursor-pointer hover:bg-slate-50 transition-colors w-full group">
                                <span class="inline-flex items-center gap-2 text-slate-600 group-hover:text-blue-600 transition-colors text-sm font-semibold">
                                    <span class="flex items-center justify-center w-6 h-6 bg-blue-50 text-blue-600 rounded-md group-hover:bg-blue-600 group-hover:text-white transition-colors shadow-sm">
                                        <svg xmlns="http://www.w3.org/2000/svg" class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 16v1a3 3 0 003 3h10a3 3 0 003-3v-1m-4-8l-4-4m0 0L8 8m4-4v12" /></svg>
                                    </span>
                                    Klik untuk unggah Surat Tugas
                                </span>
                                <input type="file" accept="image/*,application/pdf" class="hidden" on:change={handleSuratTugasFileChange} />
                            </label>
                        {/if}
                    </div>
                </div>
                        <div class="p-4 sm:p-6 md:p-8 space-y-8">
                            <!-- Report Text Section -->
                            <div class="space-y-4" transition:fade={{ duration: 200 }}>
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
                            <!-- ... (same as before) ... -->
                            <div class="space-y-4 pt-4 border-t border-slate-100" transition:fade={{ duration: 200 }}>
                                <div class="flex justify-between items-center border-b border-slate-100 pb-3">
                                    <Label class="text-lg font-bold text-slate-800">Dokumentasi Kegiatan</Label>
                                    
                                        <span class="text-xs text-slate-400">Max 6 File (JPG/PDF), Max 10MB</span>
                                </div>
                                
                                
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
                                                        
                                                        <button 
                                                            class="bg-red-500 text-white p-2 rounded-full hover:bg-red-600 transform hover:scale-110 transition-all shadow-lg"
                                                            on:click|stopPropagation={() => removeFile(i)}
                                                            title="Hapus file"
                                                        >
                                                            <svg xmlns="http://www.w3.org/2000/svg" class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                                                                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16" /></svg>
                                                        </button>
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
                        </div>

                        {#if $userStore.role !== 'kasubag'}
                            <div class="bg-slate-50 border-t border-slate-100 p-4 sm:p-6 -mx-4 sm:-mx-6 md:-mx-8 -mb-4 sm:-mb-6 md:-mb-8 flex flex-col sm:flex-row justify-end gap-3 rounded-b-xl mt-8">
                                <Button variant="outline" size="lg" class="border-blue-200 text-blue-700 hover:bg-blue-50 px-6 w-full sm:w-auto" on:click={saveDraftLaporan}>Simpan Draft</Button>
                                <Button size="lg" class="bg-blue-600 hover:bg-blue-700 text-white shadow-lg shadow-blue-500/20 px-8 w-full sm:w-auto" on:click={submitLaporan}>Submit Laporan Kegiatan</Button>
                            </div>
                        {/if}
                    </div>
                </div>
                {/if}

                {#if currentTab === 'rincian'}
                    <!-- Form Input Rincian Biaya (Integrated from CostModal) -->
                    <div class="bg-white rounded-xl shadow-sm border border-slate-200 overflow-hidden mt-8" transition:fade={{ duration: 200 }}>
                        <div class="p-4 sm:p-6 md:p-8 space-y-4">
                            {#each recordsList as empRecord, index}
                                {@const empId = empRecord.id}
                                {#if localCosts[empId]}
                                    <div id="form-rincian-{empId}" class="bg-white rounded-xl shadow-sm overflow-hidden scroll-mt-24 border border-slate-200">
                                        <button 
                                            class="w-full flex items-center justify-between p-4 bg-indigo-600 hover:bg-indigo-700 transition-colors text-white font-bold text-left rounded-t-xl {localCosts[empId]._expanded ? 'rounded-b-none' : 'rounded-b-xl'}"
                                            on:click={() => {
                                                localCosts[empId]._expanded = !localCosts[empId]._expanded;
                                                localCosts = { ...localCosts };
                                            }}
                                        >
                                            <span class="text-sm">{empRecord.employee?.name || 'Petugas'}</span>
                                            <svg xmlns="http://www.w3.org/2000/svg" class="h-5 w-5 transform transition-transform duration-200 {localCosts[empId]._expanded ? 'rotate-180' : ''}" viewBox="0 0 20 20" fill="currentColor">
                                                <path fill-rule="evenodd" d="M5.293 7.293a1 1 0 011.414 0L10 10.586l3.293-3.293a1 1 0 111.414 1.414l-4 4a1 1 0 01-1.414 0l-4-4a1 1 0 010-1.414z" clip-rule="evenodd" />
                                            </svg>
                                        </button>

                                        {#if localCosts[empId]._expanded}
                                            <div class="p-4 md:p-6 bg-slate-50/30 grid gap-4 md:gap-6 border-t border-slate-200" transition:fade={{ duration: 150 }}>
                                            
                                            <!-- Location Tabs -->
                                            <div class="col-span-1 md:col-span-2 -mb-2">
                                                <Label class="text-[10px] md:text-xs font-semibold uppercase text-slate-500 tracking-wider mb-2 block">Pilih Lokasi</Label>
                                                <div class="flex gap-2 overflow-x-auto pb-2">
                                                    {#each getLocations(empRecord) as loc, idx}
                                                        <button 
                                                            class={cn(
                                                                "px-4 py-2 rounded-lg text-sm font-semibold whitespace-nowrap transition-colors border",
                                                                localCosts[empId].selectedLocationIndex === idx 
                                                                    ? "bg-indigo-600 text-white border-indigo-600 shadow-sm" 
                                                                    : "bg-white text-slate-600 border-slate-200 hover:bg-slate-50 hover:border-slate-300"
                                                            )}
                                                            on:click={() => { localCosts[empId].selectedLocationIndex = idx; localCosts = {...localCosts}; }}
                                                        >
                                                            {loc.province ? toTitleCase(loc.province) : toTitleCase(empRecord.province)}
                                                        </button>
                                                    {/each}
                                                </div>
                                            </div>

                                            {#each getLocations(empRecord) as loc, idx}
                                                {#if localCosts[empId].selectedLocationIndex === idx && localCosts[empId].costs.details && localCosts[empId].costs.details[idx]}
                                                {@const detail = localCosts[empId].costs.details[idx]}
                                                {@const sbmDays = getDaysForLocation(loc)}
                                                {@const sbmRate = getRateForLocation(loc)}
                                                {@const sbmTotal = sbmDays * sbmRate}

                                                <!-- Mode Transportasi -->
                                                <div class="p-3 md:p-4 bg-white rounded-xl border border-slate-200 shadow-sm w-full space-y-1.5" transition:fade={{ duration: 150 }}>
                                                    <Label class="text-[10px] md:text-xs font-semibold uppercase text-slate-500 tracking-wider">Mode Transportasi ({toTitleCase(loc.province || empRecord.province)})</Label>
                                                    <Select bind:value={detail.transportMode} class="bg-slate-50 border-slate-200 h-9 md:h-10 text-sm">
                                                        <option value="Pesawat">Pesawat Udara</option>
                                                        <option value="Kendaraan Umum">Kendaraan Umum / Kereta</option>
                                                        <option value="Kendaraan Dinas">Kendaraan Dinas</option>
                                                    </Select>
                                                </div>

                                                <!-- Uang Harian SBM (Specific to Location) -->
                                                <div class="p-3 md:p-4 bg-blue-50/50 rounded-xl border border-blue-100 space-y-2.5 md:space-y-3 w-full" transition:fade={{ duration: 150 }}>
                                                    <div class="flex flex-wrap justify-between items-center gap-2 border-b border-blue-200 pb-2 mb-2">
                                                        <h4 class="text-xs md:text-sm font-semibold text-blue-800 flex items-center gap-1.5">
                                                            <svg xmlns="http://www.w3.org/2000/svg" class="h-3.5 w-3.5 md:h-4 md:w-4 text-blue-500" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                                                                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 8v4l3 3m6-3a9 9 0 11-18 0 9 9 0 0118 0z" />
                                                            </svg>
                                                            Uang Harian (SBM) - {toTitleCase(loc.province || empRecord.province)}
                                                        </h4>
                                                        <span class="text-base md:text-lg font-bold text-blue-700">{formatCurrency(sbmTotal)}</span>
                                                    </div>
                                                    
                                                    <div class="bg-white p-3 rounded-lg border border-blue-100 shadow-sm space-y-2">
                                                        <div class="flex flex-wrap md:flex-nowrap items-center gap-2 md:gap-3 text-xs text-slate-600">
                                                            <div class="flex-none pr-2 border-r border-slate-100">
                                                                <span class="block text-[9px] uppercase tracking-wider text-slate-400 font-semibold mb-0.5">Durasi</span>
                                                                <span class="font-medium whitespace-nowrap">{sbmDays} Hari</span>
                                                            </div>
                                                            <div class="flex-1 min-w-0 overflow-hidden">
                                                                <span class="block text-[9px] uppercase tracking-wider text-slate-400 font-semibold mb-0.5">Rate Estimasi</span>
                                                                <span class="font-medium truncate block w-full">{formatCurrency(sbmRate)} <span class="text-[9px] text-slate-400 font-normal">/ hari</span></span>
                                                            </div>
                                                        </div>
                                                    </div>
                                                </div>

                                                <!-- Tiket & Boarding Pass -->
                                                <div class="p-3 md:p-4 bg-white rounded-xl border border-slate-200 shadow-sm space-y-4 w-full" transition:fade={{ duration: 150 }}>
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
                                                                <Input type="text" disabled={$userStore.role === 'kasubag'} value={formatInputNumber(detail.ticketGo)} on:input={(e) => updateCost(empId, 'ticketGo', e, idx)} class="pl-8 md:pl-9 h-9 md:h-10 text-sm bg-white border-slate-200 focus:bg-white" />
                                                            </div>
                                                            {#if !detail.ticketGoFile}
                                                                <label class="flex flex-col items-center justify-center w-full h-16 border border-dashed border-slate-300 rounded-lg cursor-pointer bg-white hover:bg-slate-50 hover:border-blue-400 transition-all group mt-1">
                                                                    <div class="flex flex-col items-center justify-center pt-2 pb-2 pointer-events-none">
                                                                        <svg xmlns="http://www.w3.org/2000/svg" class="w-4 h-4 mb-1 text-slate-400 group-hover:text-blue-500 transition-colors" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                                                                            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M7 16a4 4 0 01-.88-7.903A5 5 0 1115.9 6L16 6a5 5 0 011 9.9M15 13l-3-3m0 0l-3 3m3-3v12" />
                                                                        </svg>
                                                                        <p class="text-[9px] text-slate-500 text-center px-2"><span class="font-semibold text-blue-600">Klik untuk unggah</span> atau seret file</p>
                                                                    </div>
                                                                    <input type="file" class="hidden" accept=".pdf,.jpg,.jpeg,.png" on:change={(e) => handleSpecificFileSelect(empId, e, 'ticketGoFile', idx)} />
                                                                </label>
                                                            {/if}
                                                            {#if detail.ticketGoFile}
                                                                <div class="flex items-center justify-between p-2 mt-2 bg-white border border-slate-200 rounded-md">
                                                                    <span class="text-[10px] md:text-xs text-slate-700 truncate mr-2 flex-1">{detail.ticketGoFile.name}</span>
                                                                    <div class="flex gap-2 shrink-0 text-[10px]">
                                                                        <button type="button" class="inline-flex items-center justify-center px-2 py-1 md:px-2.5 md:py-1 bg-blue-50 hover:bg-blue-100 text-blue-700 font-medium rounded-md border border-blue-200 transition-colors" on:click={() => openPreview(detail.ticketGoFile)}>Lihat</button>
                                                                        <button type="button" class="inline-flex items-center justify-center px-2 py-1 md:px-2.5 md:py-1 bg-red-50 hover:bg-red-100 text-red-600 font-medium rounded-md border border-red-200 transition-colors" on:click={() => removeSpecificFile(empId, 'ticketGoFile', idx)}>Hapus</button>
                                                                    </div>
                                                                </div>
                                                            {/if}
                                                        </div>

                                                        <!-- Tiket Pulang -->
                                                        <div class="space-y-2 p-3 border border-slate-100 bg-slate-50 rounded-lg">
                                                            <Label class="text-[10px] md:text-xs font-semibold uppercase text-slate-500 tracking-wider block">Tiket Pulang</Label>
                                                            <div class="relative">
                                                                <span class="absolute left-2.5 top-2 md:top-2.5 text-slate-400 text-xs md:text-sm">Rp</span>
                                                                <Input type="text" disabled={$userStore.role === 'kasubag'} value={formatInputNumber(detail.ticketBack)} on:input={(e) => updateCost(empId, 'ticketBack', e, idx)} class="pl-8 md:pl-9 h-9 md:h-10 text-sm bg-white border-slate-200 focus:bg-white" />
                                                            </div>
                                                            {#if !detail.ticketBackFile}
                                                                <label class="flex flex-col items-center justify-center w-full h-16 border border-dashed border-slate-300 rounded-lg cursor-pointer bg-white hover:bg-slate-50 hover:border-blue-400 transition-all group mt-1">
                                                                    <div class="flex flex-col items-center justify-center pt-2 pb-2 pointer-events-none">
                                                                        <svg xmlns="http://www.w3.org/2000/svg" class="w-4 h-4 mb-1 text-slate-400 group-hover:text-blue-500 transition-colors" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                                                                            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M7 16a4 4 0 01-.88-7.903A5 5 0 1115.9 6L16 6a5 5 0 011 9.9M15 13l-3-3m0 0l-3 3m3-3v12" />
                                                                        </svg>
                                                                        <p class="text-[9px] text-slate-500 text-center px-2"><span class="font-semibold text-blue-600">Klik untuk unggah</span> atau seret file</p>
                                                                    </div>
                                                                    <input type="file" class="hidden" accept=".pdf,.jpg,.jpeg,.png" on:change={(e) => handleSpecificFileSelect(empId, e, 'ticketBackFile', idx)} />
                                                                </label>
                                                            {/if}
                                                            {#if detail.ticketBackFile}
                                                                <div class="flex items-center justify-between p-2 mt-2 bg-white border border-slate-200 rounded-md">
                                                                    <span class="text-[10px] md:text-xs text-slate-700 truncate mr-2 flex-1">{detail.ticketBackFile.name}</span>
                                                                    <div class="flex gap-2 shrink-0 text-[10px]">
                                                                        <button type="button" class="inline-flex items-center justify-center px-2 py-1 md:px-2.5 md:py-1 bg-blue-50 hover:bg-blue-100 text-blue-700 font-medium rounded-md border border-blue-200 transition-colors" on:click={() => openPreview(detail.ticketBackFile)}>Lihat</button>
                                                                        <button type="button" class="inline-flex items-center justify-center px-2 py-1 md:px-2.5 md:py-1 bg-red-50 hover:bg-red-100 text-red-600 font-medium rounded-md border border-red-200 transition-colors" on:click={() => removeSpecificFile(empId, 'ticketBackFile', idx)}>Hapus</button>
                                                                    </div>
                                                                </div>
                                                            {/if}
                                                        </div>

                                                        <!-- Boarding Pass -->
                                                        <div class="space-y-2 p-3 border border-slate-100 bg-slate-50 rounded-lg md:col-span-2 flex flex-col justify-center">
                                                            <Label class="text-[10px] md:text-xs font-semibold uppercase text-slate-500 tracking-wider block">Boarding Pass</Label>
                                                            
                                                            {#if detail.boardingPassFiles && detail.boardingPassFiles.length > 0}
                                                                <div class="grid grid-cols-1 md:grid-cols-2 gap-2 mt-2">
                                                                    {#each detail.boardingPassFiles as bpFile, bpIdx}
                                                                        <div class="flex items-center justify-between p-2 bg-white border border-slate-200 rounded-md w-full">
                                                                            <span class="text-[10px] md:text-xs text-slate-700 truncate mr-2 flex-1">{bpFile.name}</span>
                                                                            <div class="flex gap-2 shrink-0 text-[10px]">
                                                                                <button type="button" class="inline-flex items-center justify-center px-2 py-1 md:px-2.5 md:py-1 bg-blue-50 hover:bg-blue-100 text-blue-700 font-medium rounded-md border border-blue-200 transition-colors" on:click={() => openPreview(bpFile)}>Lihat</button>
                                                                                <button type="button" class="inline-flex items-center justify-center px-2 py-1 md:px-2.5 md:py-1 bg-red-50 hover:bg-red-100 text-red-600 font-medium rounded-md border border-red-200 transition-colors" on:click={() => removeBoardingPassFile(empId, bpIdx, idx)}>Hapus</button>
                                                                            </div>
                                                                        </div>
                                                                    {/each}
                                                                </div>
                                                            {/if}
            

                                                            <label class="flex flex-col items-center justify-center w-full h-12 border border-dashed border-slate-300 rounded-lg cursor-pointer bg-white hover:bg-slate-50 hover:border-blue-400 transition-all group mt-2">
                                                                <div class="flex items-center justify-center pointer-events-none gap-2">
                                                                    <svg xmlns="http://www.w3.org/2000/svg" class="w-4 h-4 text-slate-400 group-hover:text-blue-500 transition-colors" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                                                                        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4v16m8-8H4" />
                                                                    </svg>
                                                                    <p class="text-[10px] md:text-xs font-semibold text-blue-600">Tambah Boarding Pass</p>
                                                                </div>
                                                                <input type="file" class="hidden" accept=".pdf,.jpg,.jpeg,.png" on:change={(e) => handleBoardingPassFileSelect(empId, e, idx)} />
                                                            </label>
                                                        </div>
                                                    </div>
                                                </div>

                                                <!-- Hotel -->
                                                <div class="p-3 md:p-4 bg-white rounded-xl border border-slate-200 shadow-sm space-y-3 w-full" transition:fade={{ duration: 150 }}>
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
                                                            <Input type="number" disabled={$userStore.role === 'kasubag'} bind:value={detail.hotelDays} class="h-9 md:h-10 text-sm bg-slate-50 border-slate-200 px-2" />
                                                        </div>
                                                        <div class="space-y-1.5 col-span-2">
                                                            <Label class="text-[10px] md:text-xs font-semibold uppercase text-slate-500 tracking-wider">Rate per Malam</Label>
                                                            <div class="relative">
                                                                <span class="absolute left-2.5 top-2 md:top-2.5 text-slate-400 text-xs md:text-sm">Rp</span>
                                                                <Input type="text" disabled={$userStore.role === 'kasubag'} value={formatInputNumber(detail.hotelRate)} on:input={(e) => updateCost(empId, 'hotelRate', e, idx)} class="pl-8 md:pl-9 h-9 md:h-10 text-sm bg-slate-50 border-slate-200" />
                                                            </div>
                                                        </div>
                                                    </div>

                                                    {#if !detail.hotelFile}
                                                        <label class="flex flex-col items-center justify-center w-full h-16 border border-dashed border-slate-300 rounded-lg cursor-pointer bg-slate-50 hover:bg-slate-100 hover:border-blue-400 transition-all group">
                                                            <div class="flex flex-col items-center justify-center pt-2 pb-2 pointer-events-none">
                                                                <svg xmlns="http://www.w3.org/2000/svg" class="w-4 h-4 mb-1 text-slate-400 group-hover:text-blue-500 transition-colors" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                                                                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M7 16a4 4 0 01-.88-7.903A5 5 0 1115.9 6L16 6a5 5 0 011 9.9M15 13l-3-3m0 0l-3 3m3-3v12" />
                                                                </svg>
                                                                <p class="text-[9px] text-slate-500 text-center px-2"><span class="font-semibold text-blue-600">Klik untuk unggah kwitansi hotel</span> atau seret file</p>
                                                            </div>
                                                            <input type="file" class="hidden" accept=".pdf,.jpg,.jpeg,.png" on:change={(e) => handleSpecificFileSelect(empId, e, 'hotelFile', idx)} />
                                                        </label>
                                                    {/if}
                                                    
                                                    {#if detail.hotelFile}
                                                        <div class="flex items-center justify-between p-2 mt-2 bg-slate-50 border border-slate-200 rounded-md w-full">
                                                            <span class="text-[10px] md:text-xs text-slate-700 truncate mr-2 flex-1">{detail.hotelFile.name}</span>
                                                            <div class="flex gap-2 shrink-0 text-[10px]">
                                                                <button type="button" class="inline-flex items-center justify-center px-2 py-1 md:px-2.5 md:py-1 bg-blue-50 hover:bg-blue-100 text-blue-700 font-medium rounded-md border border-blue-200 transition-colors" on:click={() => openPreview(detail.hotelFile)}>Lihat</button>
                                                                <button type="button" class="inline-flex items-center justify-center px-2 py-1 md:px-2.5 md:py-1 bg-red-50 hover:bg-red-100 text-red-600 font-medium rounded-md border border-red-200 transition-colors" on:click={() => removeSpecificFile(empId, 'hotelFile', idx)}>Hapus</button>
                                                            </div>
                                                        </div>
                                                    {/if}
                                                </div>

                                                <!-- Bukti Transportasi atau Rental -->
                                                <div class="p-3 md:p-4 bg-white rounded-xl border border-slate-200 shadow-sm space-y-3 w-full" transition:fade={{ duration: 150 }}>
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
                                                            <Input type="text" disabled={$userStore.role === 'kasubag'} value={formatInputNumber(detail.transportAmount)} on:input={(e) => updateCost(empId, 'transportAmount', e, idx)} class="pl-8 md:pl-9 h-9 md:h-10 text-sm bg-slate-50 border-slate-200" />
                                                        </div>
                                                    </div>

                                                    {#if !detail.transportFile}
                                                        <label class="flex flex-col items-center justify-center w-full h-16 border border-dashed border-slate-300 rounded-lg cursor-pointer bg-slate-50 hover:bg-slate-100 hover:border-blue-400 transition-all group">
                                                            <div class="flex flex-col items-center justify-center pt-2 pb-2 pointer-events-none">
                                                                <svg xmlns="http://www.w3.org/2000/svg" class="w-4 h-4 mb-1 text-slate-400 group-hover:text-blue-500 transition-colors" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                                                                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M7 16a4 4 0 01-.88-7.903A5 5 0 1115.9 6L16 6a5 5 0 011 9.9M15 13l-3-3m0 0l-3 3m3-3v12" />
                                                                </svg>
                                                                <p class="text-[9px] text-slate-500 text-center px-2"><span class="font-semibold text-blue-600">Klik untuk unggah bukti transport</span> atau seret file</p>
                                                            </div>
                                                            <input type="file" class="hidden" accept=".pdf,.jpg,.jpeg,.png" on:change={(e) => handleSpecificFileSelect(empId, e, 'transportFile', idx)} />
                                                        </label>
                                                    {/if}

                                                    {#if detail.transportFile}
                                                        <div class="flex items-center justify-between p-2 mt-2 bg-slate-50 border border-slate-200 rounded-md w-full">
                                                            <span class="text-[10px] md:text-xs text-slate-700 truncate mr-2 flex-1">{detail.transportFile.name}</span>
                                                            <div class="flex gap-2 shrink-0 text-[10px]">
                                                                <button type="button" class="inline-flex items-center justify-center px-2 py-1 md:px-2.5 md:py-1 bg-blue-50 hover:bg-blue-100 text-blue-700 font-medium rounded-md border border-blue-200 transition-colors" on:click={() => openPreview(detail.transportFile)}>Lihat</button>
                                                                <button type="button" class="inline-flex items-center justify-center px-2 py-1 md:px-2.5 md:py-1 bg-red-50 hover:bg-red-100 text-red-600 font-medium rounded-md border border-red-200 transition-colors" on:click={() => removeSpecificFile(empId, 'transportFile', idx)}>Hapus</button>
                                                            </div>
                                                        </div>
                                                    {/if}
                                                </div>

                                                <!-- Biaya Tambahan Lainnya -->
                                                <div class="p-3 md:p-4 bg-slate-50 rounded-xl border border-slate-200 shadow-sm space-y-4 w-full" transition:fade={{ duration: 150 }}>
                                                    <div class="flex justify-between items-center border-b border-slate-200 pb-2">
                                                        <h4 class="text-xs md:text-sm font-semibold text-slate-700 flex items-center gap-1.5">
                                                            <svg xmlns="http://www.w3.org/2000/svg" class="h-3.5 w-3.5 md:h-4 md:w-4 text-indigo-500" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                                                                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 6v6m0 0v6m0-6h6m-6 0H6" />
                                                            </svg>
                                                            Biaya Tambahan Lainnya
                                                        </h4>
                                                        <Button size="sm" class="h-8 px-4 text-xs font-bold tracking-wider bg-indigo-600 hover:bg-indigo-700 text-white shadow-md shadow-indigo-500/20 transition-all rounded-lg flex items-center gap-1.5 hover:scale-[1.02]" on:click={() => addAdditionalCost(empId, idx)}>
                                                            <svg xmlns="http://www.w3.org/2000/svg" class="h-4 w-4" viewBox="0 0 20 20" fill="currentColor"><path fill-rule="evenodd" d="M10 3a1 1 0 011 1v5h5a1 1 0 110 2h-5v5a1 1 0 11-2 0v-5H4a1 1 0 110-2h5V4a1 1 0 011-1z" clip-rule="evenodd" /></svg>
                                                            ADD COST</Button>
                                                    </div>

                                                    <div class="space-y-3">
                                                        {#if detail.additionalCosts && detail.additionalCosts.length > 0}
                                                            {#each detail.additionalCosts as cost, costIdx}
                                                                <div class="bg-white p-3 border border-slate-200 rounded-lg relative group">
                                                                    <button type="button" class="absolute -top-2 -right-2 bg-red-100 text-red-600 rounded-full p-1 border border-red-200 hover:bg-red-200 transition-colors shadow-sm" on:click={() => removeAdditionalCost(empId, costIdx, idx)}>
                                                                        <svg xmlns="http://www.w3.org/2000/svg" class="h-3 w-3" viewBox="0 0 20 20" fill="currentColor">
                                                                            <path fill-rule="evenodd" d="M4.293 4.293a1 1 0 011.414 0L10 8.586l4.293-4.293a1 1 0 111.414 1.414L11.414 10l4.293 4.293a1 1 0 01-1.414 1.414L10 11.414l-4.293 4.293a1 1 0 01-1.414-1.414L8.586 10 4.293 5.707a1 1 0 010-1.414z" clip-rule="evenodd" />
                                                                        </svg>
                                                                    </button>
                                                                    <div class="grid grid-cols-1 md:grid-cols-2 gap-3">
                                                                        <div class="space-y-1.5">
                                                                            <Label class="text-[10px] font-semibold uppercase text-slate-500 tracking-wider">Nama Biaya</Label>
                                                                            <Input type="text" disabled={$userStore.role === 'kasubag'} placeholder="Cth: Taksi Bandara" bind:value={detail.additionalCosts[costIdx].name} class="h-8 text-xs bg-slate-50 border-slate-200" />
                                                                        </div>
                                                                        <div class="space-y-1.5">
                                                                            <Label class="text-[10px] font-semibold uppercase text-slate-500 tracking-wider">Nominal</Label>
                                                                            <div class="relative">
                                                                                <span class="absolute left-2 top-1.5 text-slate-400 text-xs">Rp</span>
                                                                                <Input type="text" disabled={$userStore.role === 'kasubag'} value={formatInputNumber(cost.amount)} on:input={(e) => updateAdditionalCostAmount(empId, costIdx, e, idx)} class="pl-7 h-8 text-xs bg-slate-50 border-slate-200" />
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
                                                                                    <input type="file" class="hidden" accept=".pdf,.jpg,.jpeg,.png" on:change={(e) => handleAdditionalFileSelect(empId, e, costIdx, idx)} />
                                                                                </label>
                                                                            {:else}
                                                                                <div class="flex items-center gap-2 bg-slate-50 p-2 rounded border border-slate-200 w-full">
                                                                                    <span class="text-[10px] text-slate-700 truncate max-w-[150px] flex-1">{cost.file.name}</span>
                                                                                    <div class="flex gap-2 shrink-0 text-[10px]">
                                                                                        <button type="button" class="inline-flex items-center justify-center px-2 py-1 md:px-2.5 md:py-1 bg-blue-50 hover:bg-blue-100 text-blue-700 font-medium rounded-md border border-blue-200 transition-colors" on:click={() => openPreview(cost.file)}>Lihat</button>
                                                                                        <button type="button" class="inline-flex items-center justify-center px-2 py-1 md:px-2.5 md:py-1 bg-red-50 hover:bg-red-100 text-red-600 font-medium rounded-md border border-red-200 transition-colors" on:click={() => removeAdditionalFile(empId, costIdx, idx)}>Hapus</button>
                                                                                    </div>
                                                                                </div>
                                                                            {/if}
                                                                        </div>
                                                                    </div>
                                                                </div>
                                                            {/each}
                                                        {:else}
                                                            <div class="text-center p-4 border border-dashed border-slate-300 rounded-lg text-xs text-slate-500">
                                                                Belum ada biaya tambahan diinputkan untuk lokasi ini.
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
                                                                <p class="text-[10px] text-slate-500 font-medium">Akumulasi seluruh komponen biaya semua lokasi</p>
                                                            </div>
                                                        </div>
                                                        <div class="mt-2 sm:mt-0 text-left sm:text-right w-full sm:w-auto">
                                                            <span class="block text-2xl md:text-3xl font-black text-transparent bg-clip-text bg-gradient-to-r from-blue-700 to-indigo-600 tracking-tight">
                                                                {formatCurrency(localCosts[empId].totalCost)}
                                                            </span>
                                                        </div>
                                                    </div>
                                                </div>
                                                
                                                {/if}
                                            {/each}
                                        </div>
                                        {/if}
                                    </div>
                                {/if}
                            {/each}
                            
                            {#if $userStore.role !== 'kasubag'}
                                <div class="bg-slate-50 border-t border-slate-100 p-4 sm:p-6 mt-8 -mx-4 sm:-mx-6 md:-mx-8 -mb-4 sm:-mb-6 md:-mb-8 flex flex-col sm:flex-row justify-end gap-3 rounded-b-xl">
                                    <Button variant="outline" size="lg" class="border-indigo-200 text-indigo-700 hover:bg-indigo-50 px-6 w-full sm:w-auto" on:click={saveDraftRincian}>Simpan Draft</Button>
                                    <Button size="lg" class="bg-indigo-600 hover:bg-indigo-700 text-white shadow-lg shadow-indigo-500/20 px-8 w-full sm:w-auto" on:click={submitRincian}>Simpan Rincian Biaya</Button>
                                </div>
                            {/if}
                        </div>
                    </div>
                {/if}
    {/if}
</div>

    <!-- Document Preview Modal -->
    <Dialog open={isPreviewOpen} hideCloseButton={true} on:close={() => isPreviewOpen = false} class="!w-[95vw] md:!w-[90vw] !max-w-6xl !h-[90vh] md:!h-[85vh] !p-0 overflow-hidden rounded-xl shadow-2xl z-[60]">
        <div class="h-full flex flex-col">
            <div class="flex items-center justify-between px-4 md:px-6 py-3 md:py-4 border-b border-slate-100 bg-slate-50/50 flex-none">
                <div class="flex flex-col min-w-0 pr-4">
                    <h3 class="text-base md:text-lg font-bold text-slate-800 tracking-tight truncate">Pratinjau Dokumen</h3>
                    {#if previewFile}
                        <p class="text-[10px] md:text-xs text-slate-500 truncate max-w-[200px] sm:max-w-xs md:max-w-md" title={getDisplayName(previewFile.name)}>{getDisplayName(previewFile.name)}</p>
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