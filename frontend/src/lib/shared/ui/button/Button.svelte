<script>
    import { cn } from '$lib/shared/utils/utils';
    
    let className = undefined;
    export { className as class };
    export let variant = 'default';
    export let size = 'default';
    
    /** @type {Record<string, string>} */
    const variants = {
        default: 'clay-default bg-indigo-500 text-white hover:bg-indigo-400 border-2 border-indigo-400 transition-all hover:-translate-y-0.5 hover:shadow-lg hover:shadow-indigo-500/30 active:scale-95',
        destructive: 'clay-destructive bg-rose-500 text-white hover:bg-rose-400 border-2 border-rose-400 transition-all hover:-translate-y-0.5 hover:shadow-lg hover:shadow-rose-500/30 active:scale-95',
        outline: 'border border-slate-300 bg-white text-slate-700 hover:bg-slate-50 hover:text-slate-900 transition-colors',
        secondary: 'clay-secondary bg-slate-100 text-slate-800 hover:bg-slate-200 border-2 border-slate-200 transition-all hover:-translate-y-0.5 hover:shadow-lg hover:shadow-slate-500/10 active:scale-95',
        ghost: 'bg-slate-50/50 hover:bg-slate-100 text-slate-700 transition-colors',
        glass: 'clay-glass bg-white/10 hover:bg-white/20 border border-white/20 text-white transition-all hover:-translate-y-0.5 active:scale-95',
        link: 'text-primary underline-offset-4 hover:underline'
    };
    
    /** @type {Record<string, string>} */
    const sizes = {
        default: 'h-11 px-6 py-2.5 rounded-xl',
        sm: 'h-9 rounded-lg px-4',
        lg: 'h-12 rounded-2xl px-10 text-base',
        icon: 'h-11 w-11 rounded-xl'
    };
    
    $: classes = cn(
        'inline-flex items-center justify-center whitespace-nowrap text-sm font-bold ring-offset-background focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2 disabled:pointer-events-none disabled:opacity-50',
        variants[variant],
        sizes[size],
        className
    );
</script>

<button class={classes} {...$$restProps} on:click>
    <slot />
</button>

<style>
    /* Base Claymorphism Shadows */
    .clay-default {
        box-shadow: inset 2px 2px 5px rgba(255, 255, 255, 0.4), inset -3px -3px 7px rgba(0, 0, 0, 0.15);
    }
    .clay-destructive {
        box-shadow: inset 2px 2px 5px rgba(255, 255, 255, 0.4), inset -3px -3px 7px rgba(0, 0, 0, 0.15);
    }
    .clay-secondary {
        box-shadow: inset 2px 2px 5px rgba(255, 255, 255, 0.6), inset -3px -3px 7px rgba(0, 0, 0, 0.05);
    }
    .clay-glass {
        box-shadow: inset 2px 2px 5px rgba(255, 255, 255, 0.2), inset -3px -3px 7px rgba(0, 0, 0, 0.2);
    }
</style>
