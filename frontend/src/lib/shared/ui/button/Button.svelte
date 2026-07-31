<script>
    import { cn } from '$lib/shared/utils/utils';
    
    let className = undefined;
    export { className as class };
    export let variant = 'default';
    export let size = 'default';
    
    /** @type {Record<string, string>} */
    const variants = {
        default: 'bg-primary text-primary-foreground shadow-[inset_-4px_-4px_8px_rgba(0,0,0,0.2),inset_4px_4px_8px_rgba(255,255,255,0.4),4px_8px_16px_rgba(0,0,0,0.15)] hover:bg-primary/90 hover:-translate-y-0.5 hover:shadow-[inset_-4px_-4px_8px_rgba(0,0,0,0.2),inset_4px_4px_8px_rgba(255,255,255,0.4),4px_12px_20px_rgba(0,0,0,0.2)] active:translate-y-0 active:scale-95 active:shadow-[inset_-2px_-2px_4px_rgba(0,0,0,0.3),inset_2px_2px_4px_rgba(255,255,255,0.3),2px_4px_8px_rgba(0,0,0,0.15)]',
        destructive: 'bg-destructive text-destructive-foreground shadow-[inset_-4px_-4px_8px_rgba(0,0,0,0.25),inset_4px_4px_8px_rgba(255,255,255,0.35),4px_8px_16px_rgba(239,68,68,0.25)] hover:bg-destructive/90 hover:-translate-y-0.5 hover:shadow-[inset_-4px_-4px_8px_rgba(0,0,0,0.25),inset_4px_4px_8px_rgba(255,255,255,0.35),4px_12px_20px_rgba(239,68,68,0.35)] active:translate-y-0 active:scale-95 active:shadow-[inset_-2px_-2px_4px_rgba(0,0,0,0.35),inset_2px_2px_4px_rgba(255,255,255,0.25),2px_4px_8px_rgba(239,68,68,0.25)]',
        outline: 'border-2 border-primary/10 bg-background text-primary shadow-[inset_-4px_-4px_8px_rgba(0,0,0,0.05),inset_4px_4px_8px_rgba(255,255,255,1),4px_8px_16px_rgba(0,0,0,0.05)] hover:bg-accent hover:text-accent-foreground hover:-translate-y-0.5 hover:shadow-[inset_-4px_-4px_8px_rgba(0,0,0,0.05),inset_4px_4px_8px_rgba(255,255,255,1),4px_12px_20px_rgba(0,0,0,0.1)] active:translate-y-0 active:scale-95 active:shadow-[inset_-2px_-2px_4px_rgba(0,0,0,0.1),inset_2px_2px_4px_rgba(255,255,255,0.8),2px_4px_8px_rgba(0,0,0,0.05)]',
        secondary: 'bg-secondary text-secondary-foreground shadow-[inset_-4px_-4px_8px_rgba(0,0,0,0.1),inset_4px_4px_8px_rgba(255,255,255,0.8),4px_8px_16px_rgba(0,0,0,0.08)] hover:bg-secondary/80 hover:-translate-y-0.5 hover:shadow-[inset_-4px_-4px_8px_rgba(0,0,0,0.1),inset_4px_4px_8px_rgba(255,255,255,0.8),4px_12px_20px_rgba(0,0,0,0.12)] active:translate-y-0 active:scale-95 active:shadow-[inset_-2px_-2px_4px_rgba(0,0,0,0.15),inset_2px_2px_4px_rgba(255,255,255,0.6),2px_4px_8px_rgba(0,0,0,0.08)]',
        ghost: 'hover:bg-accent/50 hover:text-accent-foreground hover:shadow-[inset_-2px_-2px_4px_rgba(0,0,0,0.05),inset_2px_2px_4px_rgba(255,255,255,0.8)] active:scale-95 transition-all',
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
        'inline-flex items-center justify-center whitespace-nowrap text-sm font-bold ring-offset-background transition-all duration-300 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2 disabled:pointer-events-none disabled:opacity-50 transform-gpu',
        variants[variant],
        sizes[size],
        className
    );
</script>

<button class={classes} {...$$restProps} on:click>
    <slot />
</button>
