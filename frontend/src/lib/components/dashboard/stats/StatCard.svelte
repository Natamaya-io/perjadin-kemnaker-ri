<script>
    import { cn } from '$lib/utils';
    export let title = '';
    export let value = '';
    export let description = '';
    export let iconColor = 'blue'; // blue, emerald, amber, purple, rose

    const colorClasses = {
        blue: 'bg-blue-50 text-blue-600 group-hover:bg-blue-600 group-hover:text-white',
        emerald: 'bg-emerald-50 text-emerald-600 group-hover:bg-emerald-600 group-hover:text-white',
        amber: 'bg-amber-50 text-amber-600 group-hover:bg-amber-600 group-hover:text-white',
        purple: 'bg-purple-50 text-purple-600 group-hover:bg-purple-600 group-hover:text-white',
        rose: 'bg-rose-50 text-rose-600 group-hover:bg-rose-600 group-hover:text-white'
    };
</script>

<div class="group relative bg-white rounded-2xl p-6 border border-slate-100 shadow-sm hover:shadow-lg transition-all duration-300 ease-out hover:-translate-y-1 overflow-hidden">
    <!-- Decorative background blob -->
    <div class="absolute -right-6 -top-6 h-24 w-24 rounded-full bg-slate-50 transition-all group-hover:scale-150 group-hover:bg-slate-100/50"></div>

    <div class="relative z-10 flex flex-col h-full justify-between gap-4">
        <div class="flex justify-between items-start">
            <div class={cn("p-3.5 rounded-xl transition-colors duration-300", colorClasses[iconColor] || colorClasses.blue)}>
                <slot name="icon" />
            </div>

            <!-- Optional Action Slot (e.g. for buttons) -->
            <div class="ml-auto">
                <slot name="action" />
            </div>
        </div>

        <div class="space-y-1 overflow-hidden">
            <div class="text-2xl lg:text-3xl font-bold text-slate-900 tracking-tight font-feature-settings-tnum truncate" title={value}>
                {value}
            </div>
            <div class="flex flex-col">
                <span class="text-xs font-semibold text-slate-400 uppercase tracking-wider truncate">{title}</span>
                <span class="text-[11px] text-slate-500 font-medium mt-0.5 truncate">{description}</span>
            </div>
        </div>
    </div>
</div>
<style>
    .font-feature-settings-tnum {
        font-feature-settings: "tnum";
        font-variant-numeric: tabular-nums;
    }
</style>
