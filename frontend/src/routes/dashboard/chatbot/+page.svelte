<svelte:head>
    <title>Chatbot AI - Perjadin Kemnaker RI</title>
</svelte:head>

<div class="space-y-6 pb-20 w-full">
    <!-- Header Halaman -->
    <div class="flex flex-col md:flex-row md:items-center justify-between gap-4 bg-white p-6 rounded-2xl shadow-sm border border-slate-100">
        <div>
            <div class="flex items-center gap-3 mb-1">
                <div class="p-2 rounded-xl bg-gradient-to-br from-violet-50 to-indigo-50 border border-violet-100">
                    <svg xmlns="http://www.w3.org/2000/svg" class="h-5 w-5 text-violet-600" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.75" d="M9.813 15.904L9 18.75l-.813-2.846a4.5 4.5 0 00-3.09-3.09L2.25 12l2.846-.813a4.5 4.5 0 003.09-3.09L9 5.25l.813 2.846a4.5 4.5 0 003.09 3.09L15.75 12l-2.846.813a4.5 4.5 0 00-3.09 3.09zM18.259 8.715L18 9.75l-.259-1.035a3.375 3.375 0 00-2.455-2.456L14.25 6l1.036-.259a3.375 3.375 0 002.455-2.456L18 2.25l.259 1.035a3.375 3.375 0 002.456 2.456L21.75 6l-1.035.259a3.375 3.375 0 00-2.456 2.456z" />
                    </svg>
                </div>
                <h1 class="text-2xl font-bold text-slate-800 tracking-tight">Chatbot AI</h1>
            </div>
            <p class="text-sm text-slate-500 mt-1">Asisten cerdas berbasis AI untuk membantu pengelolaan perjalanan dinas.</p>
        </div>
        <span class="inline-flex items-center gap-1.5 px-3 py-1.5 rounded-full text-xs font-semibold bg-amber-50 text-amber-700 border border-amber-200 whitespace-nowrap self-start md:self-auto">
            <span class="w-1.5 h-1.5 rounded-full bg-amber-500 animate-pulse"></span>
            Dalam Pengembangan
        </span>
    </div>

    <!-- Konten Placeholder -->
    <div class="rounded-xl border border-slate-200 shadow-sm bg-white overflow-hidden">
        <!-- Ilustrasi Utama -->
        <div class="flex flex-col items-center justify-center py-20 px-8 text-center">
            <!-- Ikon Animasi -->
            <div class="relative mb-8">
                <!-- Lingkaran latar -->
                <div class="w-28 h-28 rounded-full bg-gradient-to-br from-violet-100 to-indigo-100 flex items-center justify-center">
                    <div class="w-20 h-20 rounded-full bg-gradient-to-br from-violet-200 to-indigo-200 flex items-center justify-center">
                        <svg xmlns="http://www.w3.org/2000/svg" class="h-10 w-10 text-violet-600" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M9.813 15.904L9 18.75l-.813-2.846a4.5 4.5 0 00-3.09-3.09L2.25 12l2.846-.813a4.5 4.5 0 003.09-3.09L9 5.25l.813 2.846a4.5 4.5 0 003.09 3.09L15.75 12l-2.846.813a4.5 4.5 0 00-3.09 3.09zM18.259 8.715L18 9.75l-.259-1.035a3.375 3.375 0 00-2.455-2.456L14.25 6l1.036-.259a3.375 3.375 0 002.455-2.456L18 2.25l.259 1.035a3.375 3.375 0 002.456 2.456L21.75 6l-1.035.259a3.375 3.375 0 00-2.456 2.456z" />
                        </svg>
                    </div>
                </div>
                <!-- Dekorasi partikel -->
                <span class="absolute top-1 right-1 w-3 h-3 rounded-full bg-violet-400 opacity-70 animate-bounce" style="animation-delay: 0.1s;"></span>
                <span class="absolute bottom-2 left-0 w-2 h-2 rounded-full bg-indigo-400 opacity-60 animate-bounce" style="animation-delay: 0.3s;"></span>
                <span class="absolute top-4 -left-1 w-2 h-2 rounded-full bg-violet-300 opacity-50 animate-bounce" style="animation-delay: 0.5s;"></span>
            </div>

            <h2 class="text-xl font-bold text-slate-800 mb-2">Chatbot AI Segera Hadir</h2>
            <p class="text-sm text-slate-500 max-w-md leading-relaxed mb-8">
                Fitur ini sedang dalam tahap pengembangan aktif. Chatbot AI akan membantu Anda menjawab pertanyaan, menyusun laporan, dan menganalisis data perjalanan dinas secara otomatis.
            </p>

            <!-- Fitur yang akan datang -->
            <div class="grid grid-cols-1 sm:grid-cols-3 gap-4 w-full max-w-2xl">
                {#each [
                    { icon: 'M8.625 12a.375.375 0 11-.75 0 .375.375 0 01.75 0zm0 0H8.25m4.125 0a.375.375 0 11-.75 0 .375.375 0 01.75 0zm0 0H12m4.125 0a.375.375 0 11-.75 0 .375.375 0 01.75 0zm0 0h-.375M21 12c0 4.556-4.03 8.25-9 8.25a9.764 9.764 0 01-2.555-.337A5.972 5.972 0 015.41 20.97a5.969 5.969 0 01-.474-.065 4.48 4.48 0 00.978-2.025c.09-.457-.133-.901-.467-1.226C3.93 16.178 3 14.189 3 12c0-4.556 4.03-8.25 9-8.25s9 3.694 9 8.25z', label: 'Tanya Jawab', desc: 'Ajukan pertanyaan seputar regulasi dan prosedur perjadin' },
                    { icon: 'M19.5 14.25v-2.625a3.375 3.375 0 00-3.375-3.375h-1.5A1.125 1.125 0 0113.5 7.125v-1.5a3.375 3.375 0 00-3.375-3.375H8.25m0 12.75h7.5m-7.5 3H12M10.5 2.25H5.625c-.621 0-1.125.504-1.125 1.125v17.25c0 .621.504 1.125 1.125 1.125h12.75c.621 0 1.125-.504 1.125-1.125V11.25a9 9 0 00-9-9z', label: 'Ringkasan Laporan', desc: 'Buat ringkasan laporan perjalanan dinas secara otomatis' },
                    { icon: 'M3 13.125C3 12.504 3.504 12 4.125 12h2.25c.621 0 1.125.504 1.125 1.125v6.75C7.5 20.496 6.996 21 6.375 21h-2.25A1.125 1.125 0 013 19.875v-6.75zM9.75 8.625c0-.621.504-1.125 1.125-1.125h2.25c.621 0 1.125.504 1.125 1.125v11.25c0 .621-.504 1.125-1.125 1.125h-2.25a1.125 1.125 0 01-1.125-1.125V8.625zM16.5 4.125c0-.621.504-1.125 1.125-1.125h2.25C20.496 3 21 3.504 21 4.125v15.75c0 .621-.504 1.125-1.125 1.125h-2.25a1.125 1.125 0 01-1.125-1.125V4.125z', label: 'Analisis Data', desc: 'Analisis tren dan rekapitulasi anggaran secara cerdas' }
                ] as feature}
                    <div class="flex flex-col items-center gap-3 p-5 rounded-xl border border-slate-100 bg-slate-50/50 hover:border-violet-200 hover:bg-violet-50/30 transition-colors">
                        <div class="p-2.5 rounded-xl bg-white border border-slate-200 shadow-sm">
                            <svg xmlns="http://www.w3.org/2000/svg" class="h-5 w-5 text-violet-500" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.75" d={feature.icon} />
                            </svg>
                        </div>
                        <div class="text-center">
                            <p class="text-sm font-semibold text-slate-700 mb-1">{feature.label}</p>
                            <p class="text-xs text-slate-500 leading-relaxed">{feature.desc}</p>
                        </div>
                    </div>
                {/each}
            </div>
        </div>
    </div>
</div>
