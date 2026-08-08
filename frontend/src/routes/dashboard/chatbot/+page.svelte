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
                    backgroundColor: '#4f46e5',
                    borderRadius: 4
                }]
            },
            options: {
                responsive: true,
                maintainAspectRatio: false,
                plugins: {
                    legend: { display: false }
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
        { role: 'bot', text: 'Halo! Saya asisten AI Perjadin Kemnaker. Ada yang bisa saya bantu terkait pengelolaan GUP atau Perjalanan Dinas hari ini?' }
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
        
        // Simulate bot typing (Wait for API response)
        isTyping = true;
        await scrollToBottom();
        
        try {
            const response = await api.askChatbot(currentInput);
            if (response && response.success) {
                messages = [...messages, { role: 'bot', text: response.data.answer, data: response.data }];
            } else {
                messages = [...messages, { role: 'bot', text: "Maaf, terjadi kesalahan saat menghubungi server." }];
            }
        } catch (error) {
            messages = [...messages, { role: 'bot', text: "Maaf, gagal menghubungi server AI. Pastikan koneksi internet Anda stabil." }];
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
            { role: 'bot', text: 'Halo! Saya asisten AI Perjadin Kemnaker. Ada yang bisa saya bantu terkait pengelolaan GUP atau Perjalanan Dinas hari ini?' }
        ];
        inputText = '';
        if (textareaEl) {
            textareaEl.style.height = 'auto';
        }
    }
</script>

<div class="space-y-6 pb-20 max-w-7xl mx-auto">
    
    <!-- Chat Header / Page Header -->
    <div class="flex flex-col md:flex-row md:items-center justify-between gap-4 bg-white p-6 rounded-2xl shadow-sm border border-slate-100">
        <div class="flex items-center gap-4">
            <div class="relative">
                <div class="w-12 h-12 rounded-full bg-indigo-600 flex items-center justify-center text-white shadow-md border-2 border-white p-1 overflow-hidden">
                    <img src="https://api.dicebear.com/9.x/bottts/svg?seed=PerjadinAI&baseColor=ffffff,e0e7ff" alt="Robot Icon" class="w-full h-full object-contain" />
                </div>
                <span class="absolute bottom-0.5 right-0.5 w-3.5 h-3.5 bg-emerald-400 border-2 border-white rounded-full"></span>
            </div>
            <div>
                <h1 class="text-2xl font-bold text-slate-800 tracking-tight flex items-center gap-2">
                    Perjadin AI
                    <button 
                        type="button" 
                        title="Panduan Penggunaan" 
                        on:click={() => showInfoModal = true} 
                        class="text-indigo-500 hover:text-indigo-700 bg-indigo-50 hover:bg-indigo-100 rounded-full p-1 transition-colors"
                    >
                        <svg xmlns="http://www.w3.org/2000/svg" class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13 16h-1v-4h-1m1-4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
                        </svg>
                    </button>
                </h1>
                <p class="text-sm text-slate-500 mt-1">Asisten virtual cerdas untuk pengelolaan GUP dan Perjalanan Dinas.</p>
            </div>
        </div>
        <div>
            <button
                type="button"
                on:click={resetConversation}
                class="flex items-center gap-2 text-sm font-medium text-slate-600 hover:text-slate-900 bg-slate-50 hover:bg-slate-100 px-4 py-2 rounded-xl border border-slate-200 transition-colors"
                title="Mulai percakapan baru"
            >
                <svg xmlns="http://www.w3.org/2000/svg" class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15" />
                </svg>
                Reset
            </button>
        </div>
    </div>

    <!-- Chat Container -->
    <div class="h-[calc(100vh-250px)] min-h-[500px] w-full flex flex-col bg-white rounded-2xl shadow-sm border border-slate-200 overflow-hidden relative">
        
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
                        <div class="w-8 h-8 rounded-full bg-indigo-600 flex-shrink-0 flex items-center justify-center text-white shadow-sm mt-1 p-0.5 overflow-hidden">
                            <img src="https://api.dicebear.com/9.x/bottts/svg?seed=PerjadinAI&baseColor=ffffff,e0e7ff" alt="Robot Icon" class="w-full h-full object-contain" />
                        </div>
                        <div class="flex flex-col gap-1">
                            <span class="text-[11px] font-bold uppercase tracking-wider text-slate-400 ml-1">Perjadin AI</span>
                            <div class="bg-white border border-slate-200 px-5 py-3.5 rounded-2xl rounded-tl-sm shadow-sm text-sm text-slate-700 leading-relaxed whitespace-pre-line">
                                {msg.text}
                            </div>
                            
                            {#if msg.data && msg.data.metrics && msg.data.metrics.length > 0}
                            <div class="grid grid-cols-2 gap-2 mt-1">
                                {#each msg.data.metrics as metric}
                                <div class="bg-indigo-50 border border-indigo-100 p-2.5 rounded-xl text-center shadow-sm">
                                    <div class="text-[10px] font-bold text-indigo-400 uppercase tracking-wider">{metric.label}</div>
                                    <div class="font-bold text-indigo-700 text-sm mt-0.5">{metric.value}</div>
                                </div>
                                {/each}
                            </div>
                            {/if}

                            {#if msg.data && msg.data.actions && msg.data.actions.length > 0}
                            <div class="flex flex-wrap gap-2 mt-1">
                                {#each msg.data.actions as action}
                                <a href={action.url} target="_blank" rel="noopener noreferrer" class="inline-flex items-center gap-1.5 px-3 py-1.5 bg-white border border-slate-200 text-xs font-semibold text-slate-600 rounded-lg shadow-sm hover:bg-slate-50 hover:text-indigo-600 hover:border-indigo-200 transition-colors">
                                    <svg xmlns="http://www.w3.org/2000/svg" class="h-3.5 w-3.5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                                        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M17 17h2a2 2 0 002-2v-4a2 2 0 00-2-2H5a2 2 0 00-2 2v4a2 2 0 002 2h2m2 4h6a2 2 0 002-2v-4a2 2 0 00-2-2H9a2 2 0 00-2 2v4a2 2 0 002 2zm8-12V5a2 2 0 00-2-2H9a2 2 0 00-2 2v4h10z" />
                                    </svg>
                                    {action.label}
                                </a>
                                {/each}
                            </div>
                            {/if}
                            
                            {#if msg.data && msg.data.chart}
                            <div class="bg-white border border-slate-200 p-4 rounded-2xl shadow-sm w-full">
                                <h4 class="font-bold text-sm text-slate-800 mb-2">{msg.data.chart.title}</h4>
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
                    <div class="w-8 h-8 rounded-full bg-indigo-600 flex-shrink-0 flex items-center justify-center text-white shadow-sm mt-1 p-0.5 overflow-hidden">
                        <img src="https://api.dicebear.com/9.x/bottts/svg?seed=PerjadinAI&baseColor=ffffff,e0e7ff" alt="Robot Icon" class="w-full h-full object-contain" />
                    </div>
                    <div class="bg-white border border-slate-200 px-5 py-4 rounded-2xl rounded-tl-sm shadow-sm flex items-center gap-1.5 mt-1">
                        <span class="w-2 h-2 rounded-full bg-slate-300 animate-bounce" style="animation-delay: 0s;"></span>
                        <span class="w-2 h-2 rounded-full bg-slate-400 animate-bounce" style="animation-delay: 0.15s;"></span>
                        <span class="w-2 h-2 rounded-full bg-slate-500 animate-bounce" style="animation-delay: 0.3s;"></span>
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
</div>

<BaseModal
    bind:open={showInfoModal}
    maxWidth="max-w-md"
>
    <h2 slot="header" class="text-xl font-bold text-slate-800">Panduan Menggunakan Perjadin AI</h2>
    
    <div slot="body">
        <div class="space-y-4 text-sm text-slate-600">
            <p><strong>Perjadin AI</strong> menggunakan <em>Rule-Based Financial Analysis</em> untuk menarik data secara <em>real-time</em> dari <em>database</em> GUP dan Dalkot Anda.</p>
            
            <p>Anda cukup mengetikkan pertanyaan menggunakan <strong>kata kunci spesifik</strong> berikut:</p>
            
            <ul class="list-disc pl-5 space-y-2 text-slate-600">
                <li><strong>"GUP"</strong>, <strong>"anggaran"</strong>, <strong>"realisasi"</strong>: Menampilkan ringkasan total pagu dan serapan GUP.</li>
                <li><strong>"12 menu"</strong> atau <strong>Nama Menu (contoh: "ATK")</strong>: Menganalisis sisa anggaran pada menu GUP tertentu.</li>
                <li><strong>"Dalkot"</strong>: Menampilkan ringkasan biaya SPJ dan biaya riil Dalkot.</li>
                <li><strong>"Status Dalkot"</strong>, <strong>"Proses"</strong>: Menampilkan jumlah Dalkot yang masih <em>draft</em>, diproses, atau selesai.</li>
                <li><strong>"Petugas SPJ / Riil"</strong>: Menampilkan rekapitulasi beban kerja petugas.</li>
                <li><strong>"Grafik"</strong> atau <strong>"Chart"</strong>: Menyisipkan kata ini (contoh: <em>"Tampilkan grafik gup"</em>) akan memunculkan diagram visual interaktif.</li>
            </ul>
            
            <div class="mt-2 bg-amber-50 border border-amber-100 p-3 rounded-lg text-amber-800">
                <strong class="block mb-1">Catatan Penting:</strong>
                Karena bot ini dibangun tanpa LLM generatif, pertanyaannya harus memuat salah satu kata kunci di atas agar dapat memberikan laporan yang akurat.
            </div>
        </div>
    </div>
</BaseModal>
