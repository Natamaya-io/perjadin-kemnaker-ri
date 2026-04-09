const fetch = require('node-fetch');

async function test() {
    try {
        const loginRes = await fetch('http://localhost:3000/api/v1/auth/login', {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ email: 'kasubag', password: '12345678' })
        });
        const loginData = await loginRes.json();
        console.log('Login status:', loginRes.status);
        if (!loginData.token) {
            console.log('Login failed:', loginData);
            return;
        }
        const token = loginData.token;

        const recordsRes = await fetch('http://localhost:3000/api/v1/records', {
            headers: { 'Authorization': `Bearer ${token}` }
        });
        const records = await recordsRes.json();
        console.log('Records length:', records.length);
        if (records.length > 0) {
            console.log('Record startDate:', records[0].startDate);
        } else {
            console.log('Records array is empty');
        }
    } catch (e) {
        console.error(e);
    }
}
test();
