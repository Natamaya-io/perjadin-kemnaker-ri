<script>
    import { onMount } from 'svelte';
    import { browser } from '$app/environment';

    /** @type {string} */
    export let size = '200px';
    /** @type {string} */
    export let className = '';

    onMount(() => {
        if (browser) {
            // @ts-ignore
            if (!window.lottiePlayerScriptLoaded) {
                const script = document.createElement('script');
                script.src = 'https://unpkg.com/@lottiefiles/lottie-player@latest/dist/lottie-player.js';
                script.async = true;
                document.head.appendChild(script);
                // @ts-ignore
                window.lottiePlayerScriptLoaded = true;
            }
        }
    });
</script>

<div class="relative flex items-center justify-center {className}">
    <lottie-player 
        src="https://lottie.host/079ca848-38ad-4820-b849-fd547898dcd0/qDroRPjLEg.json" 
        background="transparent" 
        speed="1" 
        style="width: {size}; height: {size};" 
        loop 
        autoplay>
    </lottie-player>
</div>
