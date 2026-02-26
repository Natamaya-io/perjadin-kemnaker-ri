<script>
    import { onMount } from 'svelte';
    import { goto } from '$app/navigation';
    import { userStore } from '$lib/stores/auth';
    import { browser } from '$app/environment';

    onMount(() => {
        if (browser) {
            if ($userStore.loggedIn) {
                goto('/dashboard');
            } else {
                goto('/login');
            }
        }
    });
</script>
