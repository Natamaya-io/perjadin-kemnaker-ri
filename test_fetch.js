fetch("https://staging.gatsu51.com/api/v1/gup/laporan?year=2026")
  .then(res => res.json())
  .then(data => {
    console.log(JSON.stringify(data.rows.map(r => r.jenisPengadaan)));
  })
  .catch(console.error);
