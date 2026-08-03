<script>
	import Input from '$lib/shared/ui/input/Input.svelte';
	import Select from '$lib/shared/ui/select/Select.svelte';
	import Datepicker from '$lib/shared/ui/datepicker/Datepicker.svelte';

	export let searchQuery = '';
	export let statusFilter = 'all';
	export let sortOption = 'date-desc';
	export let startDate = '';
	export let endDate = '';
	export let statusOptions = [
		{ value: 'all', label: 'Semua Status' },
		{ value: 'In Progress', label: 'In Progress' },
		{ value: 'Completed', label: 'Completed' }
	];
</script>

<div class="grid grid-cols-1 md:grid-cols-[1fr_auto_auto_auto_auto] gap-3 w-full items-center">

	<!-- Search Bar — mengisi sisa ruang -->
	<div class="relative w-full min-w-0">
		<div class="absolute inset-y-0 left-0 pl-3 flex items-center pointer-events-none text-slate-400">
			<svg xmlns="http://www.w3.org/2000/svg" class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
				<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z" />
			</svg>
		</div>
		<Input
			type="text"
			placeholder="Cari No. SPJ atau Tujuan..."
			bind:value={searchQuery}
			class="pl-9 bg-white border-slate-200 w-full"
		/>
	</div>

	<!-- Filter Status -->
	<div class="w-full md:w-36">
		<Select
			bind:value={statusFilter}
			class="bg-white border-slate-200 w-full"
			options={statusOptions}
		/>
	</div>

	<!-- Sort -->
	<div class="w-full md:w-44">
		<Select
			bind:value={sortOption}
			class="bg-white border-slate-200 w-full"
			options={[
				{ value: 'spj-desc', label: 'ID SPJ Terbaru' },
				{ value: 'spj-asc', label: 'ID SPJ Terlama' },
				{ value: 'date-desc', label: 'Tanggal Terbaru' },
				{ value: 'date-asc', label: 'Tanggal Terlama' },
				{ value: 'cost-desc', label: 'Biaya Tertinggi' },
				{ value: 'cost-asc', label: 'Biaya Terendah' }
			]}
		/>
	</div>

	<!-- Datepicker Dari -->
	<div class="w-full md:w-44">
		<Datepicker bind:value={startDate} placeholder="Dari tanggal" class="bg-white border-slate-200" />
	</div>

	<!-- Datepicker Sampai + Reset -->
	<div class="flex items-center gap-2 w-full md:w-auto">
		<div class="flex-1 md:w-44">
			<Datepicker bind:value={endDate} min={startDate} placeholder="Sampai tanggal" class="bg-white border-slate-200" />
		</div>
		{#if startDate || endDate}
			<button
				type="button"
				class="shrink-0 text-xs font-semibold text-red-500 hover:text-red-700 hover:bg-red-50 px-2.5 py-1.5 rounded-lg border border-red-100 transition-all whitespace-nowrap"
				on:click={() => { startDate = ''; endDate = ''; }}
			>
				Reset
			</button>
		{/if}
	</div>
</div>
