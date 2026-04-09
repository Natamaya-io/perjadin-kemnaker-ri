const records = [
  { id: '1', spd: 'ID-SPJ-001', employeeId: 'other-id', employee: { email: 'chandra' }, status: 'Draft', reportStatus: 'Pending', startDate: '2026-04-08T00:00:00Z', endDate: '2026-04-08T00:00:00Z' },
  { id: '2', spd: 'ID-SPJ-001', employeeId: 'other-id-2', employee: { email: 'doni' }, status: 'Draft', reportStatus: 'Pending', startDate: '2026-04-08T00:00:00Z', endDate: '2026-04-08T00:00:00Z' },
  { id: '3', spd: 'ID-SPJ-001', employeeId: 'dhika-id', employee: { email: 'dhikanurkhaliffa' }, status: 'Draft', reportStatus: 'Pending', startDate: '2026-04-08T00:00:00Z', endDate: '2026-04-08T00:00:00Z' }
];

const userStore = { id: 'dhika-id', email: 'dhikanurkhaliffa', role: 'protokol', name: 'Dhika Nur Khaliffa' };

const myRecords = records.filter(r => {
    if (!r) return false;
    const myEmail = userStore.email || '';
    const myName = userStore.name || '';
    const recEmail = r.email || '';
    const empEmail = r.employee?.email || '';
    const empName = r.employee?.name || '';
    
    return recEmail === myEmail || 
           empEmail === myEmail || 
           empName === myName ||
           r.creatorId === userStore.id ||
           r.employeeId === userStore.id;
});

console.log("myRecords length:", myRecords.length);
console.log("myRecords spd:", myRecords.map(r=>r.spd));
