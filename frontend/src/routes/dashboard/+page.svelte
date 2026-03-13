<script>
    import { recordsStore } from '$lib/features/pengajuan/store';
    import { userStore } from '$lib/features/auth/store';
    import Button from '$lib/shared/ui/button/Button.svelte';

    // Granular Components
    import WelcomeBanner from '$lib/features/dashboard/ui/welcome/WelcomeBanner.svelte';
    import WelcomeContent from '$lib/features/dashboard/ui/welcome/WelcomeContent.svelte';
    import WelcomeActions from '$lib/features/dashboard/ui/welcome/WelcomeActions.svelte';

    import StatsGrid from '$lib/features/dashboard/ui/stats/StatsGrid.svelte';
    import StatCard from '$lib/features/dashboard/ui/stats/StatCard.svelte';

    import RecentActivityCard from '$lib/features/dashboard/ui/recent/RecentActivityCard.svelte';
    import ActivityItem from '$lib/features/dashboard/ui/recent/ActivityItem.svelte';
    import EmptyActivity from '$lib/features/dashboard/ui/recent/EmptyActivity.svelte';

    import TimelineCalendar from '$lib/features/dashboard/ui/timeline/TimelineCalendar.svelte';
    import PieChart from '$lib/shared/ui/charts/PieChart.svelte';
    
    // Helper for currency if not in utils
    function formatIDR(amount) {
        return new Intl.NumberFormat('id-ID', { style: 'currency', currency: 'IDR', maximumFractionDigits: 0 }).format(amount);
    }

    $: records = $recordsStore;
    $: myRecords = ($userStore.role === 'super_admin' || $userStore.role === 'keuangan' || $userStore.role === 'kasubag') ? records : records.filter(r => r.email === $userStore.email || (r.employee && r.employee.email === $userStore.email));

    // Stats Logic (Scoped to Role)
    $: statsSource = ($userStore.role === 'super_admin' || $userStore.role === 'keuangan' || $userStore.role === 'kasubag') ? records : myRecords;

    // Group by SPD to avoid counting multiple employees in the same trip as multiple trips
    $: uniqueTrips = Object.values(statsSource.reduce((acc, r) => {
        if (!acc[r.spd]) acc[r.spd] = r;
        return acc;
    }, {}));

    $: myUniqueTrips = Object.values(myRecords.reduce((acc, r) => {
        if (!acc[r.spd]) acc[r.spd] = r;
        return acc;
    }, {}));

    $: totalTrips = uniqueTrips.length;
    $: totalCost = statsSource.reduce((acc, r) => acc + (r.totalCost || 0), 0);
    
    $: activeTrips = uniqueTrips.filter(r => {
        if (!r.startDate || !r.endDate) return false;
        const now = new Date();
        const start = new Date(r.startDate);
        const end = new Date(r.endDate);
        // Reset time for accurate day comparison
        now.setHours(0,0,0,0);
        start.setHours(0,0,0,0);
        end.setHours(0,0,0,0);
        return now >= start && now <= end && (r.status === 'Approved' || r.status === 'Submitted');
    }).length;
    
    $: pendingReports = myUniqueTrips.filter(r => (r.status === 'Approved' || r.status === 'Submitted' || r.status === 'Draft') && r.reportStatus !== 'Completed').length;
    $: newAssignments = myUniqueTrips.filter(r => r.status === 'Draft').length;        
    
    // Recent Logic
    $: recentRecords = [...myRecords].sort((a, b) => new Date(b.startDate) - new Date(a.startDate)).slice(0, 5).map(record => {
        const allRecordsForSpd = records.filter(r => r.spd === record.spd);
        const officerIndex = allRecordsForSpd.findIndex(r => r.id === record.id);
        const nomorSpdPetugas = String(officerIndex + 1).padStart(3, '0');
        return { ...record, nomorSpdPetugas };
    });    
    
    // Chart Data
    $: statusData = [
        { label: 'Completed', value: records.filter(r => r.paymentStatus === 'Paid').length, color: '#10b981' }, // emerald-500
        { label: 'In Progress', value: records.filter(r => (r.status === 'Submitted' || r.status === 'Approved') && r.paymentStatus !== 'Paid').length, color: '#f97316' }, // orange-500
        { label: 'Assigned', value: records.filter(r => r.status === 'Draft').length, color: '#eab308' }, // yellow-500
        { label: 'Ditolak', value: records.filter(r => r.status === 'Rejected').length, color: '#ef4444' }   // red-500
    ].filter(d => d.value > 0);

    $: reportData = [
         { label: 'Laporan Selesai', value: records.filter(r => r.reportStatus === 'Completed').length, color: '#3b82f6' }, // blue-500
         { label: 'Belum Lapor', value: records.filter(r => r.reportStatus !== 'Completed').length, color: '#a855f7' } // purple-500
    ].filter(d => d.value > 0);
</script>

<div class="space-y-8 pb-20">
    <!-- 1. Welcome Section (Organism) -->
    <WelcomeBanner role={$userStore.role}>
        <WelcomeContent role={$userStore.role} />
        
        <WelcomeActions>
            {#if $userStore.role === 'super_admin' || $userStore.role === 'kasubag'}
            <a href="/dashboard/pengajuan/new" class="w-full md:w-auto">
                <Button class="w-full md:w-auto bg-white text-blue-600 hover:bg-blue-50 border-0 shadow-lg font-semibold h-12 px-6 rounded-xl transition-transform active:scale-95 whitespace-nowrap">
                    + Pengajuan Perjalanan
                </Button>
            </a>
            {/if}
            {#if $userStore.role === 'super_admin' || $userStore.role === 'keuangan' || $userStore.role === 'kasubag'}
                 <a href="/dashboard/admin/perdin" class="w-full md:w-auto">
                    <Button variant="outline" class="w-full md:w-auto bg-blue-600/30 border-white/20 text-white hover:bg-blue-600/50 h-12 px-6 rounded-xl backdrop-blur-sm transition-transform active:scale-95 whitespace-nowrap">
                        Kelola Keuangan
                    </Button>
                </a>
            {/if}
        </WelcomeActions>
    </WelcomeBanner>

    <!-- 1.5 Alert Penugasan Baru (Khusus Protokol) -->
    {#if $userStore.role !== 'super_admin' && $userStore.role !== 'keuangan' && $userStore.role !== 'kasubag'}
        {#if newAssignments > 0}
        <div class="bg-gradient-to-r from-rose-500 to-pink-600 rounded-2xl p-6 shadow-lg shadow-rose-500/20 text-white flex flex-col md:flex-row items-center justify-between gap-6 relative overflow-hidden mb-8 animate-in fade-in slide-in-from-bottom-2">
            <div class="absolute -right-6 -top-6 h-32 w-32 rounded-full bg-white/10 blur-2xl pointer-events-none"></div>
            <div class="absolute -left-6 -bottom-6 h-24 w-24 rounded-full bg-white/10 blur-xl pointer-events-none"></div>
            
            <div class="relative z-10 flex flex-col sm:flex-row items-center text-center sm:text-left gap-5 w-full md:w-auto">
                <div class="h-14 w-14 rounded-full bg-white/20 flex items-center justify-center backdrop-blur-sm shrink-0 border border-white/20 shadow-inner">
                    <svg xmlns="http://www.w3.org/2000/svg" class="h-7 w-7 text-white" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 17h5l-1.405-1.405A2.032 2.032 0 0118 14.158V11a6.002 6.002 0 00-4-5.659V5a2 2 0 10-4 0v.341C7.67 6.165 6 8.388 6 11v3.159c0 .538-.214 1.055-.595 1.436L4 17h5m6 0v1a3 3 0 11-6 0v-1m6 0H9" />
                    </svg>
                </div>
                <div>
                    <h3 class="text-xl sm:text-2xl font-bold tracking-tight mb-1">Anda memiliki {newAssignments} Penugasan Baru!</h3>
                    <p class="text-rose-100 text-sm">Ada perjalanan dinas baru yang menunggu konfirmasi dan tindakan dari Anda.</p>
                </div>
            </div>
            
            <div class="relative z-10 w-full md:w-auto shrink-0 mt-2 md:mt-0">
                <a href="/dashboard/laporan?from=notif" class="inline-flex w-full md:w-auto items-center justify-center px-6 py-3 text-sm font-bold tracking-wide text-rose-600 bg-white hover:bg-rose-50 rounded-xl shadow-md shadow-black/10 transition-all hover:scale-105 active:scale-95">
                    Lihat Detail Penugasan
                    <svg xmlns="http://www.w3.org/2000/svg" class="h-4 w-4 ml-2" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2.5" d="M14 5l7 7m0 0l-7 7m7-7H3" />
                    </svg>
                </a>
            </div>
        </div>
        {:else}
        <div class="bg-white border border-slate-200 rounded-2xl p-5 flex flex-col sm:flex-row items-center justify-between gap-4 mb-8 shadow-sm">
            <div class="flex items-center gap-4 text-center sm:text-left">
                <div class="h-12 w-12 rounded-full bg-slate-50 flex items-center justify-center border border-slate-100 shrink-0">
                    <svg xmlns="http://www.w3.org/2000/svg" class="h-6 w-6 text-slate-400" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 13l4 4L19 7" />
                    </svg>
                </div>
                <div>
                    <h3 class="text-base font-bold text-slate-700">Tidak ada penugasan baru</h3>
                    <p class="text-slate-500 text-sm mt-0.5">Anda sudah menyelesaikan semua konfirmasi penugasan saat ini.</p>
                </div>
            </div>
            <div class="w-full sm:w-auto">
                <a href="/dashboard/laporan" class="inline-flex w-full sm:w-auto items-center justify-center px-4 py-2 text-xs font-semibold text-slate-600 bg-slate-50 hover:bg-slate-100 border border-slate-200 rounded-lg transition-colors">
                    Lihat Riwayat Perjalanan
                </a>
            </div>
        </div>
        {/if}
    {/if}

    <!-- 2. Stats Section (Organism) -->
    <StatsGrid>
            <StatCard 
                title="Total Perjalanan" 
                value={totalTrips} 
                description="Kegiatan tercatat tahun ini" 
                iconColor="blue"
                bgClass="bg-gradient-to-br from-blue-500 to-indigo-600 border-transparent shadow-lg shadow-blue-500/20"
                textColorClass="text-white"
                titleColorClass="text-blue-100"
                descColorClass="text-blue-50"
                blobClass="bg-white/10 group-hover:bg-white/20"
                iconContainerClass="bg-white/20 text-white"
            >
                <div slot="icon">
                    <svg xmlns="http://www.w3.org/2000/svg" class="h-6 w-6" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 19l9 2-9-18-9 18 9-2zm0 0v-8" />
                    </svg>
                </div>
            </StatCard>

            {#if $userStore.role === 'super_admin' || $userStore.role === 'keuangan' || $userStore.role === 'kasubag'}
            <StatCard 
                title="Total Anggaran" 
                value={formatIDR(totalCost)} 
                description="Realisasi biaya perjalanan" 
                iconColor="emerald"
                bgClass="bg-gradient-to-br from-emerald-500 to-teal-600 border-transparent shadow-lg shadow-emerald-500/20"
                textColorClass="text-white"
                titleColorClass="text-emerald-100"
                descColorClass="text-emerald-50"
                blobClass="bg-white/10 group-hover:bg-white/20"
                iconContainerClass="bg-white/20 text-white"
                isSecret={true}
            >
                <div slot="icon">
                    <svg xmlns="http://www.w3.org/2000/svg" class="h-6 w-6" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 8c-1.657 0-3 .895-3 2s1.343 2 3 2 3 .895 3 2-1.343 2-3 2m0-8c1.11 0 2.08.402 2.599 1M12 8V7m0 1v8m0 0v1m0-1c-1.11 0-2.08-.402-2.599-1M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
                    </svg>
                </div>
            </StatCard>
            {/if}

            {#if $userStore.role !== 'keuangan'}
            <StatCard 
                title="Sedang Berjalan" 
                value={activeTrips} 
                description="Tim aktif di lapangan" 
                iconColor="amber"
                bgClass="bg-gradient-to-br from-amber-500 to-orange-600 border-transparent shadow-lg shadow-amber-500/20"
                textColorClass="text-white"
                titleColorClass="text-amber-100"
                descColorClass="text-amber-50"
                blobClass="bg-white/10 group-hover:bg-white/20"
                iconContainerClass="bg-white/20 text-white"
            >
                <div slot="icon">
                    <svg xmlns="http://www.w3.org/2000/svg" class="h-6 w-6" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 8v4l3 3m6-3a9 9 0 11-18 0 9 9 0 0118 0z" />
                    </svg>
                </div>
            </StatCard>

            <StatCard
                title="Laporan Pending"
                value={pendingReports}
                description="Menunggu kelengkapan dokumen"
                iconColor="purple"
                bgClass="bg-gradient-to-br from-purple-500 to-fuchsia-600 border-transparent shadow-lg shadow-purple-500/20"
                textColorClass="text-white"
                titleColorClass="text-purple-100"
                descColorClass="text-purple-50"
                blobClass="bg-white/10 group-hover:bg-white/20"
                iconContainerClass="bg-white/20 text-white"
            >
                <div slot="icon">
                    <svg xmlns="http://www.w3.org/2000/svg" class="h-6 w-6" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z" />
                    </svg>
                </div>
            
                <div slot="action">
                    {#if pendingReports > 0 && $userStore.role !== 'super_admin' && $userStore.role !== 'keuangan' && $userStore.role !== 'kasubag'}
                    <a href="/dashboard/laporan?from=notif" class="inline-flex items-center justify-center px-3 py-1.5 text-[10px] font-bold tracking-wide text-white bg-purple-500 hover:bg-purple-600 rounded-lg shadow-sm transition-colors shadow-purple-500/20 hover:shadow-purple-500/40">
                        LIHAT <span class="sr-only">Laporan Pending</span>
                    </a>
                    {/if}
                </div>
            </StatCard>
            {/if}
        </StatsGrid>
        
        <!-- 2.5 Charts Section (New) -->
        {#if $userStore.role === 'super_admin' || $userStore.role === 'kasubag'}
        <div class="grid grid-cols-1 md:grid-cols-2 gap-6 animate-in fade-in slide-in-from-bottom-4 duration-700 delay-150">
            <PieChart title="Status Pengajuan" data={statusData} />
            <PieChart title="Status Laporan" data={reportData} />
        </div>
        {/if}

        <!-- 3. Recent Activity Section (Organism) -->
        <RecentActivityCard 
            title={$userStore.role !== 'super_admin' && $userStore.role !== 'keuangan' && $userStore.role !== 'kasubag' ? "Riwayat Terbaru" : "Aktivitas Terbaru"}
            viewAllLink={$userStore.role === 'super_admin' || $userStore.role === 'keuangan' || $userStore.role === 'kasubag' ? "/dashboard/admin/perdin" : ""}
        >
            {#each recentRecords as record (record.id)}
                <ActivityItem {record} {formatIDR} />
            {/each}
            
            {#if recentRecords.length === 0}
                <EmptyActivity />
            {/if}
        </RecentActivityCard>

        <!-- 4. Timeline Calendar (Protokol User) -->
        {#if $userStore.role !== 'super_admin' && $userStore.role !== 'keuangan' && $userStore.role !== 'kasubag'}
            <div class="pt-4">
                <TimelineCalendar records={myUniqueTrips} />
            </div>
        {/if}
</div>
