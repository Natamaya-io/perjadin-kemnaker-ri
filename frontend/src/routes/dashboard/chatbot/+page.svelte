<svelte:head>
    <title>Chatbot AI - Perjadin Kemnaker RI</title>
</svelte:head>

<script>
    import { onMount, tick } from 'svelte';
    
    let messages = [
        { role: 'bot', text: 'Halo! Saya asisten AI Perjadin Kemnaker. Ada yang bisa saya bantu terkait pengelolaan GUP atau Perjalanan Dinas hari ini?' }
    ];
    let inputText = '';
    let isTyping = false;
    let chatContainer;
    let textareaEl;

    async function scrollToBottom() {
        await tick();
        if (chatContainer) {
            chatContainer.scrollTop = chatContainer.scrollHeight;
        }
    }

    function autoResize() {
        if (!textareaEl) return;
        textareaEl.style.height = 'auto';
        textareaEl.style.height = Math.min(textareaEl.scrollHeight, 128) + 'px'; // Max 32 (128px)
    }

    async function sendMessage() {
        if (!inputText.trim() || isTyping) return;
        
        // Add user message
        let currentInput = inputText.trim();
        messages = [...messages, { role: 'user', text: currentInput }];
        inputText = '';
        if (textareaEl) textareaEl.style.height = 'auto'; // Reset height
        
        await scrollToBottom();
        
        // Simulate bot typing
        isTyping = true;
        await scrollToBottom();
        
        setTimeout(async () => {
            isTyping = false;
            let reply = "Maaf, saat ini saya berjalan dalam mode simulasi (dummy). Integrasi dengan model AI asli sedang dalam tahap pengembangan!";
            
            const lowerInput = currentInput.toLowerCase();
            if (lowerInput.includes("laporan") || lowerInput.includes("gup")) {
                reply = "Untuk membuat laporan GUP, pastikan Anda sudah mencatat semua transaksi pada menu Pengajuan, lalu buka menu Laporan untuk mencetak ringkasannya secara otomatis.";
            } else if (lowerInput.includes("pagu") || lowerInput.includes("anggaran")) {
                reply = "Pagu anggaran dapat dilihat secara rinci di halaman Dashboard. Pastikan Anda sudah menginput nilai LS (Uang Persediaan) bulan ini agar sisa anggaran dapat dihitung akurat.";
            } else if (lowerInput.includes("halo") || lowerInput.includes("hai")) {
                reply = "Halo! Selamat datang. Silakan tanyakan hal-hal terkait prosedur Perjalanan Dinas, GUP, atau navigasi sistem ini.";
            }

            messages = [...messages, { role: 'bot', text: reply }];
            await scrollToBottom();
        }, 1500);
    }

    function handleKeydown(event) {
        if (event.key === 'Enter' && !event.shiftKey) {
            event.preventDefault();
            sendMessage();
        }
    }
</script>

<div class="h-[calc(100vh-120px)] w-full flex flex-col bg-white rounded-2xl shadow-sm border border-slate-200 overflow-hidden relative">
    
    <!-- Chat Header -->
    <div class="px-6 py-4 border-b border-slate-100 flex items-center justify-between bg-white z-10 shadow-[0_4px_20px_-10px_rgba(0,0,0,0.05)]">
        <div class="flex items-center gap-3">
            <div class="relative">
                <div class="w-11 h-11 rounded-full bg-gradient-to-br from-violet-500 to-indigo-600 flex items-center justify-center text-white shadow-md border-2 border-white">
                    <svg xmlns="http://www.w3.org/2000/svg" class="h-6 w-6" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9.813 15.904L9 18.75l-.813-2.846a4.5 4.5 0 00-3.09-3.09L2.25 12l2.846-.813a4.5 4.5 0 003.09-3.09L9 5.25l.813 2.846a4.5 4.5 0 003.09 3.09L15.75 12l-2.846.813a4.5 4.5 0 00-3.09 3.09z" />
                    </svg>
                </div>
                <span class="absolute bottom-0.5 right-0.5 w-3 h-3 bg-emerald-400 border-2 border-white rounded-full"></span>
            </div>
            <div>
                <h2 class="text-lg font-bold text-slate-800 leading-tight">Perjadin AI</h2>
                <p class="text-xs text-slate-500 font-medium">Asisten Virtual Cerdas</p>
            </div>
        </div>
        <div class="flex gap-2">
            <span class="inline-flex items-center gap-1.5 px-3 py-1.5 rounded-full text-[10px] font-bold uppercase tracking-wider bg-violet-50 text-violet-700 border border-violet-200">
                <span class="w-1.5 h-1.5 rounded-full bg-violet-500 animate-pulse"></span>
                Mode Simulasi
            </span>
        </div>
    </div>

    <!-- Chat Messages Area -->
    <div bind:this={chatContainer} class="flex-1 overflow-y-auto p-4 sm:p-6 bg-slate-50/50 space-y-6 custom-scrollbar">
        
        <div class="flex justify-center mb-8 mt-2">
            <span class="text-[11px] font-semibold text-slate-400 uppercase tracking-widest bg-white px-4 py-1.5 rounded-full shadow-sm border border-slate-100">
                Hari ini
            </span>
        </div>

        {#each messages as msg}
            {#if msg.role === 'bot'}
                <!-- Bot Message -->
                <div class="flex items-start gap-3 sm:gap-4 max-w-[90%] sm:max-w-[80%] animate-in fade-in slide-in-from-bottom-2 duration-300">
                    <div class="w-8 h-8 rounded-full bg-gradient-to-br from-violet-500 to-indigo-600 flex-shrink-0 flex items-center justify-center text-white shadow-sm mt-1">
                        <svg xmlns="http://www.w3.org/2000/svg" class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13 10V3L4 14h7v7l9-11h-7z" />
                        </svg>
                    </div>
                    <div class="flex flex-col gap-1">
                        <span class="text-[11px] font-bold uppercase tracking-wider text-slate-400 ml-1">Perjadin AI</span>
                        <div class="bg-white border border-slate-200 px-5 py-3.5 rounded-2xl rounded-tl-sm shadow-sm text-sm text-slate-700 leading-relaxed">
                            {msg.text}
                        </div>
                    </div>
                </div>
            {:else}
                <!-- User Message -->
                <div class="flex items-start justify-end gap-3 sm:gap-4 animate-in fade-in slide-in-from-bottom-2 duration-300">
                    <div class="flex flex-col gap-1 items-end max-w-[90%] sm:max-w-[80%]">
                        <span class="text-[11px] font-bold uppercase tracking-wider text-slate-400 mr-1">Anda</span>
                        <div class="bg-indigo-600 px-5 py-3.5 rounded-2xl rounded-tr-sm shadow-md text-sm text-white leading-relaxed">
                            {msg.text}
                        </div>
                    </div>
                </div>
            {/if}
        {/each}

        {#if isTyping}
            <!-- Typing Indicator -->
            <div class="flex items-start gap-4 max-w-[80%] animate-in fade-in duration-200">
                <div class="w-8 h-8 rounded-full bg-gradient-to-br from-violet-500 to-indigo-600 flex-shrink-0 flex items-center justify-center text-white shadow-sm mt-1">
                    <svg xmlns="http://www.w3.org/2000/svg" class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 12h.01M12 12h.01M19 12h.01M6 12a1 1 0 11-2 0 1 1 0 012 0zm7 0a1 1 0 11-2 0 1 1 0 012 0zm7 0a1 1 0 11-2 0 1 1 0 012 0z" />
                    </svg>
                </div>
                <div class="bg-white border border-slate-200 px-5 py-4 rounded-2xl rounded-tl-sm shadow-sm flex items-center gap-1.5 mt-1">
                    <span class="w-2 h-2 rounded-full bg-violet-400 animate-bounce" style="animation-delay: 0s;"></span>
                    <span class="w-2 h-2 rounded-full bg-indigo-400 animate-bounce" style="animation-delay: 0.15s;"></span>
                    <span class="w-2 h-2 rounded-full bg-sky-400 animate-bounce" style="animation-delay: 0.3s;"></span>
                </div>
            </div>
        {/if}
    </div>

    <!-- Chat Input Area -->
    <div class="p-4 bg-white border-t border-slate-100 z-10 shadow-[0_-4px_20px_-10px_rgba(0,0,0,0.02)]">
        <div class="relative flex items-end gap-2 bg-slate-50 border border-slate-200 rounded-2xl p-1.5 shadow-inner focus-within:ring-2 focus-within:ring-indigo-500/20 focus-within:border-indigo-300 transition-all">
            <textarea
                bind:this={textareaEl}
                bind:value={inputText}
                on:keydown={handleKeydown}
                on:input={autoResize}
                placeholder="Ketik pertanyaan Anda di sini..."
                class="w-full max-h-32 min-h-[44px] bg-transparent border-none focus:ring-0 resize-none py-3 px-4 text-sm text-slate-700 placeholder:text-slate-400 custom-scrollbar"
                rows="1"
            ></textarea>
            
            <button 
                on:click={sendMessage}
                disabled={!inputText.trim() || isTyping}
                class="h-11 w-11 shrink-0 flex items-center justify-center rounded-xl m-0.5 bg-indigo-600 text-white hover:bg-indigo-700 disabled:bg-slate-200 disabled:text-slate-400 transition-colors shadow-sm"
            >
                <svg xmlns="http://www.w3.org/2000/svg" class="h-5 w-5 rotate-90" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 19l9 2-9-18-9 18 9-2zm0 0v-8" />
                </svg>
            </button>
        </div>
        <p class="text-[10px] text-center text-slate-400 mt-3 font-medium">
            Perjadin AI dapat membuat kesalahan. Harap verifikasi informasi penting dengan peraturan resmi.
        </p>
    </div>
</div>
