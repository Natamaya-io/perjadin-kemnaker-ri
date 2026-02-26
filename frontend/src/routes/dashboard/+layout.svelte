<script>
    import { userStore } from '$lib/stores/auth';
    import { goto } from '$app/navigation';
    import { onMount, onDestroy } from 'svelte';
    import { browser } from '$app/environment';

    let unsubscribe;

    onMount(() => {
        if (browser && !$userStore.loggedIn) {
            goto('/login');
        }

        unsubscribe = userStore.subscribe(user => {
            if (browser && !user.loggedIn) {
                goto('/login');
            }
        });
    });

    onDestroy(() => {
        if (unsubscribe) unsubscribe();
    });
</script>

<slot />
