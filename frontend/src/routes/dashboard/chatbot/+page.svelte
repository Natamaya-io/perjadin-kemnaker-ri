<svelte:head>
    <title>Chatbot AI - Perjadin Kemnaker RI</title>
</svelte:head>

<script>
    import { onMount, tick } from 'svelte';
    import { api } from '$lib/shared/api';
    import BaseModal from '$lib/shared/ui/base-modal/BaseModal.svelte';
    import Chart from 'chart.js/auto';
    
    function renderChart(node, chartData) {
        if (!chartData) return;
        const chart = new Chart(node, {
            type: chartData.type || 'bar',
            data: {
                labels: chartData.labels,
                datasets: [{
                    label: chartData.title,
                    data: chartData.values,
                    backgroundColor: '#FEB05D',
                    borderRadius: 8
                }]
            },
            options: {
                responsive: true,
                maintainAspectRatio: false,
                plugins: {
                    legend: { display: false }
                },
                scales: {
                    x: { grid: { display: false } },
                    y: { grid: { color: '#E5E5E5' } }
                }
            }
        });

        return {
            destroy() {
                chart.destroy();
            }
        };
    }
    
    let messages = [
        { role: 'bot', text: 'Sistem siap. Apa yang ingin Anda cari dari GUP atau Dalkot hari ini?' }
    ];
    let inputText = '';
    let isTyping = false;
    let chatContainer;
    let textareaEl;
    let showInfoModal = false;

    async function scrollToBottom() {
        await tick();
        if (chatContainer) {
            chatContainer.scrollTop = chatContainer.scrollHeight;
        }
    }

    function autoResize() {
        if (!textareaEl) return;
        textareaEl.style.height = 'auto';
        textareaEl.style.height = Math.min(textareaEl.scrollHeight, 128) + 'px'; 
    }

    let sessionId = Math.random().toString(36).substring(2, 15);
    
    async function sendMessage() {
        if (!inputText.trim() || isTyping) return;
        
        let currentInput = inputText.trim();
        messages = [...messages, { role: 'user', text: currentInput }];
        inputText = '';
        if (textareaEl) textareaEl.style.height = 'auto'; 
        
        await scrollToBottom();
        
        isTyping = true;
        await scrollToBottom();
        
        try {
            const response = await api.askChatbot(currentInput, sessionId);
            if (response && response.success) {
                messages = [...messages, { role: 'bot', text: response.data.answer, data: response.data }];
            } else {
                messages = [...messages, { role: 'bot', text: "Terjadi kesalahan. Coba lagi." }];
            }
        } catch (error) {
            messages = [...messages, { role: 'bot', text: "Koneksi terputus. Periksa jaringan Anda." }];
        } finally {
            isTyping = false;
            await scrollToBottom();
        }
    }

    function handleKeydown(event) {
        if (event.key === 'Enter' && !event.shiftKey) {
            event.preventDefault();
            sendMessage();
        }
    }

    function resetConversation() {
        messages = [
            { role: 'bot', text: 'Sistem siap. Apa yang ingin Anda cari dari GUP atau Dalkot hari ini?' }
        ];
        inputText = '';
        sessionId = Math.random().toString(36).substring(2, 15);
        if (textareaEl) {
            textareaEl.style.height = 'auto';
        }
    }
</script>

<div class="space-y-6 pb-20 max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 font-sans">
    
    <!-- Chat Header (Zero-Description Header) -->
    <div class="flex flex-col md:flex-row md:items-center justify-between gap-4 bg-[#2B2A2A] border border-[#2B2A2A] p-5 sm:p-6 rounded-[2rem] sm:rounded-[3rem] text-[#F5F2F2]">
        <div class="flex items-center gap-4">
            <h1 class="text-xl sm:text-2xl font-bold tracking-tight flex items-center gap-3">
                Perjadin AI
                <span class="bg-[#5A7ACD] text-[#F5F2F2] text-xs font-bold px-3 py-1 rounded-full uppercase tracking-widest">Beta</span>
                <button 
                    type="button" 
                    title="Panduan" 
                    on:click={() => showInfoModal = true} 
                    class="text-[#F5F2F2] hover:text-[#FEB05D] active:scale-95 transition-all duration-300 ml-2"
                >
                    <svg xmlns="http://www.w3.org/2000/svg" class="h-6 w-6" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2.5" d="M13 16h-1v-4h-1m1-4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
                    </svg>
                </button>
            </h1>
        </div>
        <div>
            <button
                type="button"
                on:click={resetConversation}
                class="flex items-center gap-2 text-sm font-black text-[#2B2A2A] bg-[#FEB05D] hover:bg-[#F5F2F2] hover:text-[#FEB05D] px-6 py-2.5 rounded-full border border-[#FEB05D] active:scale-95 transition-all duration-300 ease-spring"
            >
                <svg xmlns="http://www.w3.org/2000/svg" class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2.5" d="M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15" />
                </svg>
                RESET
            </button>
        </div>
    </div>

    <!-- Chat Container -->
    <div class="h-[calc(100vh-250px)] min-h-[500px] w-full flex flex-col bg-[#F5F2F2] rounded-[2rem] sm:rounded-[3rem] shadow-sm overflow-hidden relative border border-[#2B2A2A]/10">
        
        <!-- Chat Messages Area -->
        <div bind:this={chatContainer} class="flex-1 overflow-y-auto p-4 sm:p-6 space-y-8 custom-scrollbar">
            
            <div class="flex justify-center mb-8 mt-2">
                <span class="text-[10px] font-bold text-[#5A7ACD] uppercase tracking-widest bg-white/50 px-4 py-1.5 rounded-full">
                    Hari ini
                </span>
            </div>

            {#each messages as msg}
                {#if msg.role === 'bot'}
                    <!-- Bot Message -->
                    <div class="flex items-start gap-3 sm:gap-4 max-w-[90%] sm:max-w-[80%] animate-in fade-in slide-in-from-bottom-2 duration-300">
                        <div class="w-10 h-10 rounded-full bg-[#2B2A2A] flex-shrink-0 flex items-center justify-center text-[#FEB05D]">
                            <svg xmlns="http://www.w3.org/2000/svg" class="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2.5" d="M9.75 17L9 20l-1 1h8l-1-1-.75-3M3 13h18M5 17h14a2 2 0 002-2V5a2 2 0 00-2-2H5a2 2 0 00-2 2v10a2 2 0 002 2z" />
                            </svg>
                        </div>
                        <div class="flex flex-col gap-1 w-full">
                            <span class="text-[10px] font-bold uppercase tracking-wider text-[#5A7ACD] ml-1">Sistem</span>
                            <div class="bg-[#2B2A2A] text-[#F5F2F2] px-6 py-4 rounded-[2rem] rounded-tl-sm shadow-md text-sm leading-relaxed whitespace-pre-line w-fit">
                                {msg.text}
                            </div>
                            
                            <!-- Playful Cards (Metrics) -->
                            {#if msg.data && msg.data.metrics && msg.data.metrics.length > 0}
                            <div class="grid grid-cols-2 gap-3 mt-2 w-full max-w-md">
                                {#each msg.data.metrics as metric}
                                <div class="bg-white border border-[#2B2A2A]/10 p-4 rounded-3xl text-center shadow-sm hover:-translate-y-1 hover:shadow-md transition-all duration-300">
                                    <div class="text-[10px] font-bold text-[#5A7ACD] uppercase tracking-wider">{metric.label}</div>
                                    <div class="font-black text-[#2B2A2A] text-lg mt-1">{metric.value}</div>
                                </div>
                                {/each}
                            </div>
                            {/if}

                            <!-- Actions -->
                            {#if msg.data && msg.data.actions && msg.data.actions.length > 0}
                            <div class="flex flex-wrap gap-2 mt-2">
                                {#each msg.data.actions as action}
                                <a href={action.url} target="_blank" rel="noopener noreferrer" class="inline-flex items-center gap-2 px-5 py-2.5 bg-white border border-[#2B2A2A]/20 text-xs font-bold text-[#2B2A2A] rounded-full hover:bg-[#5A7ACD] hover:text-white hover:border-[#5A7ACD] active:scale-95 transition-all duration-300 ease-spring">
                                    <svg xmlns="http://www.w3.org/2000/svg" class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                                        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2.5" d="M17 17h2a2 2 0 002-2v-4a2 2 0 00-2-2H5a2 2 0 00-2 2v4a2 2 0 002 2h2m2 4h6a2 2 0 002-2v-4a2 2 0 00-2-2H9a2 2 0 00-2 2v4a2 2 0 002 2zm8-12V5a2 2 0 00-2-2H9a2 2 0 00-2 2v4h10z" />
                                    </svg>
                                    {action.label}
                                </a>
                                {/each}
                            </div>
                            {/if}
                            
                            <!-- Chart Area -->
                            {#if msg.data && msg.data.chart}
                            <div class="bg-white border border-[#2B2A2A]/10 p-5 rounded-[2rem] shadow-sm w-full max-w-lg mt-2 hover:shadow-md transition-all duration-300">
                                <h4 class="font-black text-sm text-[#2B2A2A] mb-4">{msg.data.chart.title}</h4>
                                <div class="relative w-full h-48">
                                    <canvas use:renderChart={msg.data.chart}></canvas>
                                </div>
                            </div>
                            {/if}
                        </div>
                    </div>
                {:else}
                    <!-- User Message -->
                    <div class="flex items-start justify-end gap-3 sm:gap-4 animate-in fade-in slide-in-from-bottom-2 duration-300">
                        <div class="flex flex-col gap-1 items-end max-w-[90%] sm:max-w-[80%]">
                            <span class="text-[10px] font-bold uppercase tracking-wider text-[#5A7ACD] mr-2">Anda</span>
                            <div class="bg-[#FEB05D] text-[#2B2A2A] font-medium px-6 py-4 rounded-[2rem] rounded-tr-sm shadow-sm text-sm leading-relaxed">
                                {msg.text}
                            </div>
                        </div>
                    </div>
                {/if}
            {/each}

            {#if isTyping}
                <!-- Typing Indicator -->
                <div class="flex items-start gap-4 max-w-[80%] animate-in fade-in duration-200">
                    <div class="w-10 h-10 rounded-full bg-[#2B2A2A] flex-shrink-0 flex items-center justify-center text-[#FEB05D]">
                        <svg xmlns="http://www.w3.org/2000/svg" class="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2.5" d="M5 12h.01M12 12h.01M19 12h.01M6 12a1 1 0 11-2 0 1 1 0 012 0zm7 0a1 1 0 11-2 0 1 1 0 012 0zm7 0a1 1 0 11-2 0 1 1 0 012 0z" />
                        </svg>
                    </div>
                    <div class="bg-white border border-[#2B2A2A]/10 px-5 py-4 rounded-3xl rounded-tl-sm flex items-center gap-2 mt-1">
                        <span class="w-2.5 h-2.5 rounded-full bg-[#FEB05D] animate-bounce" style="animation-delay: 0s;"></span>
                        <span class="w-2.5 h-2.5 rounded-full bg-[#5A7ACD] animate-bounce" style="animation-delay: 0.15s;"></span>
                        <span class="w-2.5 h-2.5 rounded-full bg-[#2B2A2A] animate-bounce" style="animation-delay: 0.3s;"></span>
                    </div>
                </div>
            {/if}
        </div>

        <!-- Chat Input Area (Capsule Search Bar style) -->
        <div class="p-4 sm:p-6 bg-transparent z-10">
            <div class="relative flex items-end gap-3 bg-white border border-[#2B2A2A] rounded-[2rem] p-2 shadow-sm focus-within:ring-4 focus-within:ring-[#FEB05D]/20 transition-all duration-300">
                <textarea
                    bind:this={textareaEl}
                    bind:value={inputText}
                    on:keydown={handleKeydown}
                    on:input={autoResize}
                    placeholder="Ketik kueri Anda di sini..."
                    class="w-full max-h-32 min-h-[44px] bg-transparent border-none focus:ring-0 resize-none py-3 px-5 text-sm font-bold text-[#2B2A2A] placeholder:text-[#2B2A2A]/40 custom-scrollbar"
                    rows="1"
                ></textarea>
                
                <button 
                    on:click={sendMessage}
                    disabled={!inputText.trim() || isTyping}
                    class="h-12 w-12 shrink-0 flex items-center justify-center rounded-full bg-[#FEB05D] text-[#2B2A2A] hover:bg-[#2B2A2A] hover:text-[#FEB05D] disabled:bg-slate-200 disabled:text-slate-400 active:scale-95 transition-all duration-300 ease-spring shadow-sm cursor-pointer"
                >
                    <svg xmlns="http://www.w3.org/2000/svg" class="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2.5" d="M12 19l9 2-9-18-9 18 9-2zm0 0v-8" />
                    </svg>
                </button>
            </div>
        </div>
    </div>
</div>

<BaseModal bind:open={showInfoModal} maxWidth="max-w-md">
    <h2 slot="header" class="text-xl font-black text-[#2B2A2A]">Aturan Main Perjadin AI</h2>
    
    <div slot="body">
        <div class="space-y-6 text-sm text-[#2B2A2A]">
            <p class="font-medium">Sistem ini memproses data GUP dan Dalkot secara presisi menggunakan kueri kunci.</p>
            
            <ul class="space-y-4">
                <li class="bg-[#F5F2F2] p-4 rounded-2xl border border-[#2B2A2A]/10">
                    <strong class="block text-[#5A7ACD] uppercase tracking-wider text-xs mb-1">GUP & Realisasi</strong>
                    Ketik "GUP", "anggaran", atau "realisasi" untuk ringkasan pagu.
                </li>
                <li class="bg-[#F5F2F2] p-4 rounded-2xl border border-[#2B2A2A]/10">
                    <strong class="block text-[#5A7ACD] uppercase tracking-wider text-xs mb-1">Analisis Menu</strong>
                    Ketik nama menu spesifik, contoh: "ATK" atau "12 menu".
                </li>
                <li class="bg-[#F5F2F2] p-4 rounded-2xl border border-[#2B2A2A]/10">
                    <strong class="block text-[#5A7ACD] uppercase tracking-wider text-xs mb-1">Status Dalkot</strong>
                    Ketik "Status Dalkot" atau "Proses" untuk melihat rekapitulasi draft/selesai.
                </li>
                <li class="bg-[#F5F2F2] p-4 rounded-2xl border border-[#2B2A2A]/10">
                    <strong class="block text-[#5A7ACD] uppercase tracking-wider text-xs mb-1">Visualisasi</strong>
                    Tambahkan kata "Grafik" (contoh: "Grafik GUP") untuk memanggil diagram visual.
                </li>
            </ul>
        </div>
    </div>
</BaseModal>
