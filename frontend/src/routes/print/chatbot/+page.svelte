<script lang="ts">
    import { page } from '$app/stores';
    import { api } from '$lib/shared/api';
    import { onMount } from 'svelte';
    import LottieLoader from '$lib/shared/ui/loader/LottieLoader.svelte';

    let snapshot: any = null;
    let type: string = '';
    let menuTitle: string = '';
    let loading = true;
    let error = '';
    let reportTitle = 'Laporan Keuangan';

    const titles: Record<string, string> = {
        'ringkasan': 'Laporan Ringkasan Keuangan GUP',
        'gup_menu': 'Laporan GUP',
        'dalkot': 'Laporan Analisis Keuangan Dalkot',
        'dalkot_status': 'Laporan Status Dalkot - Proses dan Selesai',
        'dalkot_spj': 'Laporan Petugas SPJ Dalkot',
        'dalkot_riil': 'Laporan Petugas RIIL Dalkot',
    };

    onMount(async () => {
        type = $page.url.searchParams.get('type') || 'ringkasan';
        menuTitle = $page.url.searchParams.get('menu') || '';
        
        reportTitle = titles[type] || 'Laporan Keuangan';
        if (type === 'gup_menu' && menuTitle) {
            reportTitle += ' - ' + menuTitle;
        }

        try {
            const res = await api.getChatbotSnapshot();
            if (res && res.data) {
                snapshot = res.data;
            } else {
                error = 'Data tidak ditemukan.';
            }
        } catch (e: any) {
            error = e.message || 'Terjadi kesalahan saat memuat data.';
        } finally {
            loading = false;
        }
    });

    function reportMoney(value: number): string {
        return (value < 0 ? '-' : '') + 'Rp' + Math.abs(value).toLocaleString('id-ID');
    }

    function getDalkotOfficers(typeParam: string) {
        const isSpj = typeParam === 'dalkot_spj';
        return isSpj ? snapshot.dalkot.petugas_spj : snapshot.dalkot.petugas_riil;
    }

    function getOfficerType(typeParam: string) {
        return typeParam === 'dalkot_spj' ? 'SPJ' : 'RIIL';
    }
</script>

<svelte:head>
    <title>{reportTitle}</title>
</svelte:head>

{#if loading}
    <div class="min-h-screen flex items-center justify-center bg-slate-50">
        <LottieLoader />
    </div>
{:else if error}
    <div class="min-h-screen flex items-center justify-center bg-slate-50">
        <div class="text-center p-8 bg-white rounded-xl shadow-sm border border-slate-200">
            <h1 class="text-red-500 font-bold text-lg mb-2">Gagal Memuat Laporan</h1>
            <p class="text-slate-600">{error}</p>
        </div>
    </div>
{:else if snapshot}
    <div class="print-container">
        <div class="toolbar">
            <strong>{reportTitle}</strong>
            <button on:click={() => window.print()}>Cetak / Simpan PDF</button>
        </div>
        
        <article class="sheet">
            <header class="head">
                <h1>{reportTitle}</h1>
                <p>GUP-MVC · Database Perjadin · Dibuat {new Date().toLocaleString('id-ID')}</p>
            </header>

            {#if type === 'ringkasan'}
                <div class="cards">
                    <div class="card"><span>Pagu</span><strong>{reportMoney(snapshot.gup.total_anggaran)}</strong></div>
                    <div class="card"><span>Realisasi</span><strong>{reportMoney(snapshot.gup.total_realisasi)}</strong></div>
                    <div class="card"><span>Sisa</span><strong>{reportMoney(snapshot.gup.sisa_anggaran)}</strong></div>
                    <div class="card"><span>Serapan</span><strong>{snapshot.gup.serapan_persen.toFixed(2).replace('.', ',')}%</strong></div>
                </div>
                <h2>Analisis 12 Menu GUP</h2>
                <div class="table-wrap">
                    <table>
                        <thead>
                            <tr>
                                <th>No</th><th>Menu</th><th>Kode Akun</th><th>Transaksi</th>
                                <th>Pagu</th><th>Realisasi</th><th>Sisa</th><th>Serapan</th><th>Status</th>
                            </tr>
                        </thead>
                        <tbody>
                            {#each snapshot.gup_menus as m, i}
                            <tr>
                                <td class="center">{i+1}</td>
                                <td>{m.name}</td>
                                <td>{m.kode_akun}</td>
                                <td class="center">{m.jumlah_transaksi}</td>
                                <td class="num">{reportMoney(m.anggaran)}</td>
                                <td class="num">{reportMoney(m.realisasi)}</td>
                                <td class="num">{reportMoney(m.sisa_anggaran)}</td>
                                <td class="num">{m.serapan.toFixed(2).replace('.', ',')}%</td>
                                <td><span class="badge">{m.status}</span></td>
                            </tr>
                            {/each}
                        </tbody>
                    </table>
                </div>

            {#if snapshot.ls && snapshot.ls.total > 0}
                <h2>Rekap LS Bulanan</h2>
                <div class="table-wrap">
                    <table>
                        <thead>
                            <tr>
                                <th>Bulan</th><th>Total LS</th>
                            </tr>
                        </thead>
                        <tbody>
                            {#each Object.entries(snapshot.ls.bulanan) as [bulan, total]}
                            <tr>
                                <td>{bulan}</td>
                                <td class="num">{reportMoney(total)}</td>
                            </tr>
                            {/each}
                        </tbody>
                    </table>
                </div>
            {/if}

            {:else if type === 'gup_menu' && menuTitle}
                {@const m = snapshot.gup_menus.find((x: any) => x.name === menuTitle)}
                {#if m}
                    <div class="cards">
                        <div class="card"><span>Pagu</span><strong>{reportMoney(m.anggaran)}</strong></div>
                        <div class="card"><span>Realisasi</span><strong>{reportMoney(m.realisasi)}</strong></div>
                        <div class="card"><span>Sisa</span><strong>{reportMoney(m.sisa_anggaran)}</strong></div>
                        <div class="card"><span>Serapan</span><strong>{m.serapan.toFixed(2).replace('.', ',')}%</strong></div>
                    </div>
                    <h2>Ikhtisar Menu</h2>
                    <table>
                        <tbody>
                            <tr><th>Jenis Pengadaan</th><td>{m.name}</td><th>Kode Akun</th><td>{m.kode_akun}</td></tr>
                            <tr><th>Nilai Pengajuan</th><td>{reportMoney(m.nilai_pengajuan)}</td><th>Total Pajak</th><td>{reportMoney(m.pajak)}</td></tr>
                            <tr><th>Status Analisis</th><td colspan="3">{m.status}</td></tr>
                        </tbody>
                    </table>
                    <h2>Transaksi {m.name}</h2>
                    <div class="table-wrap">
                        <table>
                            <thead>
                                <tr>
                                    <th>No</th><th>Tanggal</th><th>Pembayaran</th><th>Penerima</th>
                                    <th>Sumber Dana</th><th>Nilai</th><th>Pajak</th><th>Dibayar</th>
                                </tr>
                            </thead>
                            <tbody>
                                {#each m.transactions as r, i}
                                <tr>
                                    <td class="center">{i+1}</td>
                                    <td>{r.tanggal_kwitansi || '-'}</td>
                                    <td>{r.pembayaran || '-'}</td>
                                    <td>{r.penerima || '-'}</td>
                                    <td>{r.sumber_dana || '-'}</td>
                                    <td class="num">{reportMoney(r.nilai || 0)}</td>
                                    <td class="num">{reportMoney(r.pajak || 0)}</td>
                                    <td class="num">{reportMoney(r.jumlah_dibayarkan || 0)}</td>
                                </tr>
                                {/each}
                                {#if m.transactions.length === 0}
                                <tr><td colspan="8" class="center">Belum ada transaksi pada menu ini.</td></tr>
                                {/if}
                            </tbody>
                        </table>
                    </div>
                {:else}
                    <div class="note text-red-600">Menu tidak ditemukan.</div>
                {/if}

            {:else if type === 'dalkot'}
                <div class="cards">
                    <div class="card"><span>Pengajuan</span><strong>{snapshot.dalkot.jumlah_pengajuan}</strong></div>
                    <div class="card"><span>Total SPJ</span><strong>{reportMoney(snapshot.dalkot.total_spj)}</strong></div>
                    <div class="card"><span>Biaya Riil</span><strong>{reportMoney(snapshot.dalkot.total_riil)}</strong></div>
                    <div class="card"><span>Total Biaya</span><strong>{reportMoney(snapshot.dalkot.total_biaya)}</strong></div>
                </div>
                <h2>Daftar Pengajuan Dalkot</h2>
                <div class="table-wrap">
                    <table>
                        <thead>
                            <tr>
                                <th>No</th><th>Tanggal</th><th>Kegiatan</th><th>Lokasi</th>
                                <th>Jenis</th><th>Status</th><th>SPJ</th><th>Riil</th><th>Total</th>
                            </tr>
                        </thead>
                        <tbody>
                            {#each snapshot.dalkot.rows as r, i}
                            <tr>
                                <td class="center">{i+1}</td>
                                <td>{r.tanggal_pelaksanaan}</td>
                                <td>{r.nama_kegiatan}</td>
                                <td>{r.lokasi}</td>
                                <td>{r.jenis_dalkot}</td>
                                <td><span class="badge">{r.status}</span></td>
                                <td class="num">{reportMoney(r.total_biaya_spj)}</td>
                                <td class="num">{reportMoney(r.total_biaya_riil)}</td>
                                <td class="num">{reportMoney(r.total_biaya_spj + r.total_biaya_riil)}</td>
                            </tr>
                            {/each}
                        </tbody>
                    </table>
                </div>

            {:else if type === 'dalkot_status'}
                <div class="cards">
                    <div class="card"><span>Draft</span><strong>{snapshot.dalkot.status['Draft'] || 0}</strong></div>
                    <div class="card"><span>Pending</span><strong>{snapshot.dalkot.status['Pending'] || 0}</strong></div>
                    <div class="card"><span>Dalam Proses</span><strong>{snapshot.dalkot.proses}</strong></div>
                    <div class="card"><span>Selesai</span><strong>{snapshot.dalkot.selesai}</strong></div>
                </div>
                <h2>Detail Status Pengajuan</h2>
                <table>
                    <thead>
                        <tr>
                            <th>No</th><th>ID</th><th>Tanggal</th><th>Kegiatan</th>
                            <th>Pejabat</th><th>Status</th><th>Total Biaya</th>
                        </tr>
                    </thead>
                    <tbody>
                        {#each snapshot.dalkot.rows as r, i}
                        <tr>
                            <td class="center">{i+1}</td>
                            <td>{r.id || '-'}</td>
                            <td>{r.tanggal_pelaksanaan}</td>
                            <td>{r.nama_kegiatan}</td>
                            <td>{r.pejabat}</td>
                            <td><span class="badge">{r.status}</span></td>
                            <td class="num">{reportMoney(r.total_biaya_spj + r.total_biaya_riil)}</td>
                        </tr>
                        {/each}
                    </tbody>
                </table>

            {:else if type === 'dalkot_spj' || type === 'dalkot_riil'}
                {@const officerType = getOfficerType(type)}
                {@const rows = getDalkotOfficers(type)}
                {@const totalTugas = rows.reduce((acc, r) => acc + r.jumlah_tugas, 0)}
                {@const totalBiaya = rows.reduce((acc, r) => acc + r.total_biaya, 0)}
                
                <div class="cards">
                    <div class="card"><span>Jenis Penugasan</span><strong>{officerType}</strong></div>
                    <div class="card"><span>Jumlah Petugas</span><strong>{rows.length}</strong></div>
                    <div class="card"><span>Total Penugasan</span><strong>{totalTugas}</strong></div>
                    <div class="card"><span>Total Biaya</span><strong>{reportMoney(totalBiaya)}</strong></div>
                </div>
                <h2>Ringkasan Petugas {officerType}</h2>
                <table>
                    <thead>
                        <tr>
                            <th>No</th><th>Petugas</th><th>Jumlah Tugas</th>
                            <th>Dalam Proses</th><th>Selesai</th><th>Total Biaya</th>
                        </tr>
                    </thead>
                    <tbody>
                        {#each rows as r, i}
                        <tr>
                            <td class="center">{i+1}</td>
                            <td>{r.nama}</td>
                            <td class="center">{r.jumlah_tugas}</td>
                            <td class="center">{r.pending}</td>
                            <td class="center">{r.sukses}</td>
                            <td class="num">{reportMoney(r.total_biaya)}</td>
                        </tr>
                        {/each}
                    </tbody>
                </table>
                <h2>Detail Penugasan {officerType}</h2>
                <table>
                    <thead>
                        <tr>
                            <th>Petugas</th><th>Tanggal</th><th>Kegiatan</th>
                            <th>Lokasi</th><th>Status</th>
                        </tr>
                    </thead>
                    <tbody>
                        {#each rows as r}
                            {#if r.kegiatan && r.kegiatan.length > 0}
                                {#each r.kegiatan as k}
                                <tr>
                                    <td>{r.nama}</td>
                                    <td>{k.tanggal}</td>
                                    <td>{k.kegiatan}</td>
                                    <td>{k.lokasi}</td>
                                    <td>{k.status}</td>
                                </tr>
                                {/each}
                            {/if}
                        {/each}
                    </tbody>
                </table>
            {/if}

            <div class="note">
                <strong>Catatan:</strong> Laporan dibuat otomatis dari database aplikasi pada saat halaman dibuka. 
                Untuk dokumen pertanggungjawaban resmi, lakukan verifikasi terhadap dokumen sumber dan otorisasi pejabat yang berwenang.
            </div>
            <div class="signature">
                Jakarta, {new Date().toLocaleDateString('id-ID')}<br>
                Pengelola Keuangan
                <div class="space"></div>
                <strong>________________________</strong>
            </div>
        </article>
    </div>
{/if}

<style>
    .print-container {
        font-family: Arial, sans-serif;
        color: #172033;
        margin: 0;
        background: #eef2f7;
        min-height: 100vh;
    }
    .toolbar {
        position: sticky;
        top: 0;
        background: #0f172a;
        color: #fff;
        padding: 12px 24px;
        display: flex;
        justify-content: space-between;
        align-items: center;
        z-index: 10;
    }
    .toolbar button {
        border: 0;
        border-radius: 8px;
        padding: 9px 14px;
        font-weight: 700;
        cursor: pointer;
        background: #fff;
        color: #0f172a;
    }
    .toolbar button:hover {
        background: #f1f5f9;
    }
    .sheet {
        max-width: 1100px;
        margin: 24px auto;
        background: #fff;
        padding: 38px;
        box-shadow: 0 8px 28px #0001;
    }
    .head {
        border-bottom: 3px solid #1d4ed8;
        padding-bottom: 16px;
        margin-bottom: 22px;
    }
    .head h1 {
        margin: 0;
        font-size: 24px;
    }
    .head p {
        color: #64748b;
        margin: 7px 0 0;
    }
    .cards {
        display: grid;
        grid-template-columns: repeat(4, 1fr);
        gap: 12px;
        margin: 18px 0;
    }
    .card {
        border: 1px solid #dbe3ee;
        border-radius: 12px;
        padding: 14px;
    }
    .card span {
        font-size: 11px;
        color: #64748b;
        text-transform: uppercase;
    }
    .card strong {
        display: block;
        margin-top: 6px;
        font-size: 17px;
    }
    h2 {
        font-size: 17px;
        margin-top: 28px;
        color: #0f2f68;
    }
    table {
        width: 100%;
        border-collapse: collapse;
        margin-top: 10px;
        font-size: 12px;
    }
    th, td {
        border: 1px solid #d8dee8;
        padding: 8px;
        vertical-align: top;
    }
    th {
        background: #edf3fb;
        text-align: left;
    }
    .num {
        text-align: right;
        white-space: nowrap;
    }
    .center {
        text-align: center;
    }
    .badge {
        display: inline-block;
        border-radius: 999px;
        padding: 3px 8px;
        background: #eef2ff;
        font-size: 10px;
        font-weight: 700;
    }
    .note {
        margin-top: 22px;
        padding: 12px;
        border-left: 4px solid #1d4ed8;
        background: #eff6ff;
        font-size: 12px;
        line-height: 1.6;
    }
    .signature {
        margin-top: 48px;
        text-align: right;
    }
    .space {
        height: 55px;
    }
    @media (max-width: 700px) {
        .sheet {
            margin: 0;
            padding: 18px;
        }
        .cards {
            grid-template-columns: 1fr 1fr;
        }
        .table-wrap {
            overflow: auto;
        }
    }
    @media print {
        .print-container {
            background: #fff;
        }
        .toolbar {
            display: none;
        }
        .sheet {
            box-shadow: none;
            margin: 0;
            max-width: none;
            padding: 16px;
        }
        @page {
            size: A4 landscape;
            margin: 12mm;
        }
    }
</style>
