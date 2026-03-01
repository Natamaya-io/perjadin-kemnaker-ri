<script>
    import { recordsStore } from '$lib/stores/records';
    import { userStore } from '$lib/stores/auth';
    
    // Components
    import ReportHeader from '$lib/components/dashboard/laporan/ReportHeader.svelte';
    import ReportList from '$lib/components/dashboard/laporan/ReportList.svelte';
    import ReportItem from '$lib/components/dashboard/laporan/ReportItem.svelte';
    import EmptyState from '$lib/components/dashboard/laporan/EmptyState.svelte';

    $: myRecords = $recordsStore.filter(r => r.email === $userStore.email || (r.employee && r.employee.email === $userStore.email) || $userStore.role === 'super_admin' || $userStore.role === 'keuangan' || $userStore.role === 'kasubag');

    // Filter & Sort State
    let searchQuery = '';
    let statusFilter = 'all'; // 'all', 'Completed', 'Pending'
    let sortOption = 'date-desc'; // 'date-desc', 'date-asc'

    $: filteredRecords = myRecords
        .filter(r => {
            const query = searchQuery.toLowerCase();
            const matchSearch = 
                (r.purpose?.toLowerCase() || '').includes(query) ||
                (r.location?.toLowerCase() || '').includes(query) ||
                (r.spd?.toLowerCase() || '').includes(query) ||
                (r.employee?.name?.toLowerCase() || '').includes(query);
            
            const matchStatus = statusFilter === 'all' || r.reportStatus === statusFilter;

            return matchSearch && matchStatus;
        })
        .sort((a, b) => {
            if (sortOption === 'date-desc') return new Date(b.startDate).getTime() - new Date(a.startDate).getTime();
            if (sortOption === 'date-asc') return new Date(a.startDate).getTime() - new Date(b.startDate).getTime();
            return 0;
        });

    $: groupedRecords = filteredRecords.reduce((acc, record) => {
        if (!acc[record.spd]) {
            acc[record.spd] = { ...record, employeesList: [record] };
        } else {
            acc[record.spd].employeesList.push(record);
        }
        return acc;
    }, {});

    $: uniqueRecords = Object.values(groupedRecords);
</script>

<div class="space-y-8 pb-20">
    <ReportHeader 
        bind:searchQuery 
        bind:statusFilter 
        bind:sortOption 
    />

    {#if uniqueRecords.length === 0}
        <EmptyState />
    {:else}
        <ReportList>
            {#each uniqueRecords as record (record.spd)}
                <ReportItem {record} />
            {/each}
        </ReportList>
    {/if}
</div>
