<script>
    import { login } from '$lib/stores/auth';
    import { goto } from '$app/navigation';
    import { fly } from 'svelte/transition';
    
    // Login Components
    import LoginBackground from '$lib/components/login/LoginBackground.svelte';
    import LoginCard from '$lib/components/login/LoginCard.svelte';
    import LoginHeader from '$lib/components/login/LoginHeader.svelte';
    import LoginBody from '$lib/components/login/LoginBody.svelte';
    import DemoBanner from '$lib/components/login/DemoBanner.svelte';
    import LoginForm from '$lib/components/login/LoginForm.svelte';
    import LoginFooter from '$lib/components/login/LoginFooter.svelte';

    let email = '';
    let password = '';
    let isLoading = false;
    let errorMessage = '';

    async function handleLogin() {
        errorMessage = '';
        if (!email || !password) {
            errorMessage = 'Email dan Password tidak boleh kosong';
            return;
        }

        isLoading = true;
        // Simulate network delay if needed, or rely on api delay
        // await new Promise(r => setTimeout(r, 800));
        
        const result = await login(email, password);
        
        if (result.success) {
            goto('/dashboard');
        } else {
            errorMessage = 'Email atau password salah.';
            isLoading = false;
        }
    }
</script>

<div class="relative min-h-[100dvh] w-full flex flex-col items-center justify-center p-4 sm:p-6 overflow-x-hidden">
    <LoginBackground />

    <div class="relative z-10 w-full max-w-[400px] flex flex-col my-auto" in:fly={{ y: 20, duration: 600, delay: 100 }}>
        <LoginCard>
            <LoginHeader />
            <LoginBody>
                <DemoBanner />
                {#if errorMessage}
                    <div class="text-xs font-medium text-red-600 bg-red-50 p-3 rounded-lg border border-red-100 text-center animate-shake">
                        {errorMessage}
                    </div>
                {/if}
                <LoginForm bind:email bind:password {isLoading} on:submit={handleLogin} />
            </LoginBody>
            <LoginFooter />
        </LoginCard>
        
        <div class="mt-8 text-center">
            <a href="/" class="text-xs text-white/40 hover:text-white/80 transition-colors font-medium">Hubungi admin untuk bantuan teknis</a>
        </div>
    </div>
</div>

<style>
    @keyframes shake {
        0%, 100% { transform: translateX(0); }
        25% { transform: translateX(-4px); }
        75% { transform: translateX(4px); }
    }
    .animate-shake {
        animation: shake 0.3s ease-in-out;
    }
</style>
