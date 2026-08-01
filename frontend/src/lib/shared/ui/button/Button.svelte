<script>
    import { cn } from '$lib/shared/utils/utils';
    
    let className = undefined;
    export { className as class };
    export let variant = 'default';
    export let size = 'default';
    
    /** @type {Record<string, string>} */
    const variants = {
        default: 'clay-default bg-primary text-primary-foreground hover:bg-primary/90',
        destructive: 'clay-destructive bg-destructive text-destructive-foreground hover:bg-destructive/90',
        outline: 'clay-outline border border-slate-200 bg-slate-50 text-slate-700 hover:bg-slate-100 hover:text-slate-900',
        secondary: 'clay-secondary bg-slate-100 text-slate-800 hover:bg-slate-200',
        ghost: 'clay-ghost bg-slate-50/50 hover:bg-slate-100 text-slate-700',
        link: 'text-primary underline-offset-4 hover:underline'
    };
    
    /** @type {Record<string, string>} */
    const sizes = {
        default: 'h-11 px-6 py-2.5 rounded-2xl',
        sm: 'h-9 rounded-xl px-4',
        lg: 'h-12 rounded-[1.25rem] px-10 text-base',
        icon: 'h-11 w-11 rounded-2xl'
    };
    
    $: classes = cn(
        'inline-flex items-center justify-center whitespace-nowrap text-sm font-bold ring-offset-background focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2 disabled:pointer-events-none disabled:opacity-50 transition-colors',
        variants[variant],
        sizes[size],
        className
    );
</script>

<button class={classes} {...$$restProps} on:click>
    <slot />
</button>

<style>
    /* Base Claymorphism Shadows (Active instantly, no hover required) */
    .clay-default {
        box-shadow: inset -4px -4px 10px rgba(0, 0, 0, 0.25), inset 4px 4px 10px rgba(255, 255, 255, 0.45), inset 1px 1px 2px rgba(255, 255, 255, 0.6);
    }
    .clay-default:active {
        box-shadow: inset -2px -2px 5px rgba(0, 0, 0, 0.35), inset 2px 2px 5px rgba(255, 255, 255, 0.35);
    }

    .clay-destructive {
        box-shadow: inset -4px -4px 10px rgba(0, 0, 0, 0.3), inset 4px 4px 10px rgba(255, 255, 255, 0.4), inset 1px 1px 2px rgba(255, 255, 255, 0.5);
    }
    .clay-destructive:active {
        box-shadow: inset -2px -2px 5px rgba(0, 0, 0, 0.4), inset 2px 2px 5px rgba(255, 255, 255, 0.3);
    }

    .clay-outline {
        box-shadow: inset -4px -4px 10px rgba(0, 0, 0, 0.08), inset 4px 4px 10px rgba(255, 255, 255, 1), inset 1px 1px 2px rgba(255, 255, 255, 1);
    }
    .clay-outline:active {
        box-shadow: inset -2px -2px 5px rgba(0, 0, 0, 0.12), inset 2px 2px 5px rgba(255, 255, 255, 0.9);
    }

    .clay-secondary {
        box-shadow: inset -4px -4px 10px rgba(0, 0, 0, 0.1), inset 4px 4px 10px rgba(255, 255, 255, 1), inset 1px 1px 2px rgba(255, 255, 255, 1);
    }
    .clay-secondary:active {
        box-shadow: inset -2px -2px 5px rgba(0, 0, 0, 0.15), inset 2px 2px 5px rgba(255, 255, 255, 0.8);
    }

    .clay-ghost {
        box-shadow: inset -3px -3px 8px rgba(0, 0, 0, 0.05), inset 3px 3px 8px rgba(255, 255, 255, 0.8);
    }
    .clay-ghost:active {
        box-shadow: inset -2px -2px 5px rgba(0, 0, 0, 0.08), inset 2px 2px 5px rgba(255, 255, 255, 0.6);
    }
</style>
