<script>
    import { recordsStore } from '$lib/stores/records';
    import { userStore } from '$lib/stores/auth';
    import Button from '$lib/components/ui/button/Button.svelte';

    // Granular Components
    import WelcomeBanner from '$lib/components/dashboard/welcome/WelcomeBanner.svelte';
    import WelcomeContent from '$lib/components/dashboard/welcome/WelcomeContent.svelte';
    import WelcomeActions from '$lib/components/dashboard/welcome/WelcomeActions.svelte';
    
    import StatsGrid from '$lib/components/dashboard/stats/StatsGrid.svelte';
    import StatCard from '$lib/components/dashboard/stats/StatCard.svelte';
    
    import RecentActivityCard from '$lib/components/dashboard/recent/RecentActivityCard.svelte';
    import ActivityItem from '$lib/components/dashboard/recent/ActivityItem.svelte';
    import EmptyActivity from '$lib/components/dashboard/recent/EmptyActivity.svelte';
    
    import PieChart from '$lib/components/ui/charts/PieChart.svelte';
    import TimelineCalendar from '$lib/components/dashboard/timeline/TimelineCalendar.svelte';
    
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
    $: totalCost = statsSource.reduce((acc, r) => acc + (r.totalCost || 0), 0);    $: activeTrips = statsSource.filter(r => r.reportStatus !== 'Completed').length;
    $: pendingReports = myUniqueTrips.filter(r => (r.status === 'Approved' || r.status === 'Submitted') && r.reportStatus !== 'Completed').length;
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

    <!-- 2. Stats Section (Organism) -->
    <StatsGrid>
            <StatCard 
                title="Total Perjalanan" 
                value={totalTrips} 
                description="Kegiatan tercatat tahun ini" 
                iconColor="blue"
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
            >
                <div slot="icon">
                    <svg xmlns="http://www.w3.org/2000/svg" class="h-6 w-6" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 8c-1.657 0-3 .895-3 2s1.343 2 3 2 3 .895 3 2-1.343 2-3 2m0-8c1.11 0 2.08.402 2.599 1M12 8V7m0 1v8m0 0v1m0-1c-1.11 0-2.08-.402-2.599-1M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
                    </svg>
                </div>
            </StatCard>
            {/if}

            <StatCard 
                title="Sedang Berjalan" 
                value={activeTrips} 
                description="Tim aktif di lapangan" 
                iconColor="amber"
            >
                <div slot="icon">
                    <svg xmlns="http://www.w3.org/2000/svg" class="h-6 w-6" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 8v4l3 3m6-3a9 9 0 11-18 0 9 9 0 0118 0z" />
                    </svg>
                </div>
            </StatCard>

            {#if $userStore.role !== 'super_admin' && $userStore.role !== 'keuangan' && $userStore.role !== 'kasubag'}
            <StatCard 
                title="Penugasan Baru" 
                value={newAssignments} 
                description="Menunggu konfirmasi Anda" 
                iconColor="rose"
            >
                <div slot="icon">
                    <svg xmlns="http://www.w3.org/2000/svg" class="h-6 w-6" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 17h5l-1.405-1.405A2.032 2.032 0 0118 14.158V11a6.002 6.002 0 00-4-5.659V5a2 2 0 10-4 0v.341C7.67 6.165 6 8.388 6 11v3.159c0 .538-.214 1.055-.595 1.436L4 17h5m6 0v1a3 3 0 11-6 0v-1m6 0H9" />
                    </svg>
                </div>
                <div slot="action">
                    {#if newAssignments > 0}
                    <a href="/dashboard/pengajuan" class="inline-flex items-center justify-center px-3 py-1.5 text-[10px] font-bold tracking-wide text-white bg-rose-500 hover:bg-rose-600 rounded-lg shadow-sm transition-colors shadow-rose-500/20 hover:shadow-rose-500/40">
                        LIHAT <span class="sr-only">Penugasan Baru</span>
                    </a>
                    {/if}
                </div>
            </StatCard>
            {/if}

            <StatCard 
                title="Laporan Pending" 
                value={pendingReports} 
                description="Menunggu kelengkapan dokumen" 
                iconColor="purple"
            >
                <div slot="icon">
                    <svg xmlns="http://www.w3.org/2000/svg" class="h-6 w-6" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z" />
                    </svg>
                </div>
            </StatCard>
        </StatsGrid>
        
        <!-- 2.5 Charts Section (New) -->
        {#if $userStore.role === 'super_admin' || $userStore.role === 'keuangan' || $userStore.role === 'kasubag'}
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
