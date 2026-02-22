<script>
    import { userStore } from '$lib/stores/auth';
    import { goto } from '$app/navigation';
    import { onMount } from 'svelte';
    import { browser } from '$app/environment';

    onMount(() => {
        if (browser && !$userStore.loggedIn) {
            goto('/login');
        }
    });

    userStore.subscribe(user => {
        if (browser && !user.loggedIn) {
            goto('/login');
        }
    });
</script>

<slot />
