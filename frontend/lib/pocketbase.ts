import PocketBase from 'pocketbase';

// Logic URL:
// 1. Jika di Browser: Gunakan '/' (relative path) agar kena Proxy Next.js -> Bypass CORS
// 2. Jika di Server: Gunakan 'http://127.0.0.1:8090' (Direct Internal Access)
const backendUrl = typeof window === 'undefined' 
  ? 'http://127.0.0.1:8090' 
  : '/'; // Relative path trigers Next.js Rewrite rule

const pb = new PocketBase(backendUrl);

// Matikan auto cancellation agar request tidak batal tiba-tiba saat React strict mode render ganda
pb.autoCancellation(false);

export default pb;
