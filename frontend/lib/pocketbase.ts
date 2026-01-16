import PocketBase from 'pocketbase';

// Gunakan direct port 8090 untuk dev agar bypass redirect Caddy
const pb = new PocketBase(process.env.NEXT_PUBLIC_POCKETBASE_URL || 'http://localhost:8090');

export default pb;
