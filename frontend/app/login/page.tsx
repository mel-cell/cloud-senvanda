'use client';

import { useState } from 'react';
import { useRouter } from 'next/navigation';
import { motion } from 'framer-motion';
import { Lock, Mail, ArrowRight } from 'lucide-react';
import { clsx } from 'clsx';
import pb from '@/lib/pocketbase';

export default function LoginPage() {
  const router = useRouter();
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState('');
  
  const [email, setEmail] = useState('');
  const [password, setPassword] = useState('');

  const handleLogin = async (e: React.FormEvent) => {
    e.preventDefault();
    setLoading(true);
    setError('');

    try {
      // Manual Auth untuk Admin via endpoint lama (Legacy PocketBase check)
      try {
        // Coba pakai SDK dulu (modern path)
        await pb.admins.authWithPassword(email, password);
      } catch (adminErr: any) {
        // Jika 404, berarti server pakai endpoint lama (/api/admins)
        if (adminErr.status === 404) {
             // Manual fetch ke legacy endpoint
             const legacyAuth = await fetch('/api/admins/auth-with-password', {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({ identity: email, password })
             });
             
             if (!legacyAuth.ok) throw new Error('Legacy auth failed');
             
             const authData = await legacyAuth.json();
             // Save token manually
             pb.authStore.save(authData.token, authData.admin); // Note: Admin object might be structure differently
        } else {
             // Jika error bukan 404 (misal 400 password salah), throw lanjut
             throw adminErr; 
        }
      }
      
      // Redirect ke dashboard
      router.push('/');
    } catch (err: any) {
      console.error(err);
      setError('Invalid email or password');
    } finally {
      setLoading(false);
    }
  };

  return (
    <main className="min-h-screen flex items-center justify-center bg-black p-4">
       <div className="absolute inset-0 bg-[radial-gradient(circle_at_center,_var(--tw-gradient-stops))] from-zinc-900/50 via-black to-black -z-10" />
       
       <motion.div 
         initial={{ opacity: 0, scale: 0.95 }}
         animate={{ opacity: 1, scale: 1 }}
         className="w-full max-w-md bg-zinc-900/50 backdrop-blur-xl border border-white/10 p-8 rounded-3xl shadow-2xl"
       >
          <div className="text-center mb-8">
            <h1 className="text-3xl font-bold bg-clip-text text-transparent bg-gradient-to-r from-blue-400 to-cyan-400">
              Welcome Back
            </h1>
            <p className="text-zinc-400 mt-2">Sign in to manage your cloud</p>
          </div>

          <form onSubmit={handleLogin} className="space-y-6">
            <div className="space-y-2">
              <label className="text-sm font-medium text-zinc-300 ml-1">Email</label>
              <div className="relative group">
                <Mail className="absolute left-3 top-1/2 -translate-y-1/2 w-5 h-5 text-zinc-500 group-focus-within:text-blue-400 transition-colors" />
                <input 
                  type="email" 
                  required
                  value={email}
                  onChange={(e) => setEmail(e.target.value)}
                  className="w-full bg-black/50 border border-white/10 rounded-xl py-3 pl-10 pr-4 outline-none focus:border-blue-500/50 focus:ring-1 focus:ring-blue-500/50 transition-all text-white placeholder:text-zinc-600"
                  placeholder="admin@senvanda.local"
                />
              </div>
            </div>

            <div className="space-y-2">
              <label className="text-sm font-medium text-zinc-300 ml-1">Password</label>
              <div className="relative group">
                <Lock className="absolute left-3 top-1/2 -translate-y-1/2 w-5 h-5 text-zinc-500 group-focus-within:text-blue-400 transition-colors" />
                <input 
                  type="password" 
                  required
                  value={password}
                  onChange={(e) => setPassword(e.target.value)}
                  className="w-full bg-black/50 border border-white/10 rounded-xl py-3 pl-10 pr-4 outline-none focus:border-blue-500/50 focus:ring-1 focus:ring-blue-500/50 transition-all text-white placeholder:text-zinc-600"
                  placeholder="••••••••"
                />
              </div>
            </div>

            {error && (
              <div className="text-red-400 text-sm text-center bg-red-500/10 py-2 rounded-lg">
                {error}
              </div>
            )}

            <button
              type="submit"
              disabled={loading}
              className={clsx(
                "w-full bg-gradient-to-r from-blue-600 to-cyan-600 hover:from-blue-500 hover:to-cyan-500 text-white font-bold py-3 rounded-xl transition-all flex items-center justify-center gap-2 group",
                loading && "opacity-70 cursor-not-allowed"
              )}
            >
              {loading ? 'Signing in...' : 'Sign In'}
              {!loading && <ArrowRight className="w-5 h-5 group-hover:translate-x-1 transition-transform" />}
            </button>
          </form>
       </motion.div>
    </main>
  );
}
