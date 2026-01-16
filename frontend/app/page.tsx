'use client';

import { useState, useEffect } from 'react';
import { useRouter } from 'next/navigation';
import { motion } from 'framer-motion';
import { Server, Container, Activity, RefreshCw, Box } from 'lucide-react';
import { clsx } from 'clsx';
import pb from '@/lib/pocketbase';

interface DockerInfo {
  containers: number;
  running: number;
  server_version: string;
  message: string;
}

export default function Home() {
  const router = useRouter(); // Need to import useRouter
  const [info, setInfo] = useState<DockerInfo | null>(null);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState('');

  const fetchDockerInfo = async () => {
    setLoading(true);
    setError('');
    try {
      // Panggil Custom Endpoint (Path Baru dengan Auth)
      const res = await pb.send('/api/senvanda/deploy/info', { method: 'POST' });
      setInfo(res as DockerInfo);
    } catch (err: any) {
      setError(err.message || 'Failed to fetch docker info');
      // Jika 401/403 (Token Expired), kick ke login
      if (err.status === 401 || err.status === 403) {
         pb.authStore.clear();
         router.push('/login');
      }
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    // Auth Guard Sederhana
    if (!pb.authStore.isValid) {
        router.push('/login');
        return;
    }
    fetchDockerInfo();
  }, []);

  return (
    <main className="min-h-screen bg-[radial-gradient(ellipse_at_top,_var(--tw-gradient-stops))] from-zinc-900 via-black to-black p-8">
      <div className="max-w-4xl mx-auto space-y-12">
        {/* Header */}
        <motion.div 
          initial={{ opacity: 0, y: -20 }}
          animate={{ opacity: 1, y: 0 }}
          transition={{ duration: 0.5 }}
          className="text-center space-y-4"
        >
          <div className="inline-flex items-center justify-center p-3 rounded-2xl bg-primary/10 mb-4">
             <Server className="w-8 h-8 text-primary" />
          </div>
          <h1 className="text-5xl font-bold bg-clip-text text-transparent bg-gradient-to-r from-blue-400 to-cyan-400">
            Cloud Senvanda
          </h1>
          <p className="text-muted-foreground text-lg">
            Self-Hosted Cloud Infrastructure Management
          </p>
        </motion.div>

        {/* Dashboard Grid */}
        <div className="grid grid-cols-1 md:grid-cols-3 gap-6">
          {/* Status Card 1: Server Version */}
          <StatusCard 
            icon={<Box className="w-6 h-6 text-accent" />}
            label="Docker Version"
            value={info?.server_version || '-'}
            loading={loading}
            delay={0.1}
          />
          
          {/* Status Card 2: Total Containers */}
          <StatusCard 
            icon={<Container className="w-6 h-6 text-purple-400" />}
            label="Total Containers"
            value={info?.containers?.toString() || '0'}
            loading={loading}
            delay={0.2}
          />

          {/* Status Card 3: Running */}
          <StatusCard 
            icon={<Activity className="w-6 h-6 text-green-400" />}
            label="Running Containers"
            value={info?.running?.toString() || '0'}
            loading={loading}
            delay={0.3}
          />
        </div>

        {/* Action Area */}
        <motion.div 
          initial={{ opacity: 0 }}
          animate={{ opacity: 1 }}
          transition={{ delay: 0.4 }}
          className="flex justify-center"
        >
          <button
            onClick={fetchDockerInfo}
            disabled={loading}
            className={clsx(
              "flex items-center gap-2 px-6 py-3 rounded-full font-medium transition-all",
              "bg-primary hover:bg-blue-600 active:scale-95",
              "disabled:opacity-50 disabled:cursor-not-allowed"
            )}
          >
            <RefreshCw className={clsx("w-5 h-5", loading && "animate-spin")} />
            {loading ? 'Refreshing...' : 'Refresh Status'}
          </button>
        </motion.div>

        {/* Error Message */}
        {error && (
          <motion.div 
            initial={{ opacity: 0 }}
            animate={{ opacity: 1 }}
            className="p-4 rounded-xl bg-red-500/10 border border-red-500/20 text-red-400 text-center"
          >
            {error}
          </motion.div>
        )}
      </div>
    </main>
  );
}

function StatusCard({ icon, label, value, loading, delay }: any) {
  return (
    <motion.div
      initial={{ opacity: 0, scale: 0.9 }}
      animate={{ opacity: 1, scale: 1 }}
      transition={{ delay }}
      className="glass p-6 rounded-2xl space-y-4 hover:border-primary/50 transition-colors group"
    >
      <div className="flex items-center justify-between">
        <div className="p-2 rounded-lg bg-zinc-800 group-hover:bg-zinc-700 transition-colors">
          {icon}
        </div>
        {loading && <div className="w-2 h-2 rounded-full bg-primary animate-ping" />}
      </div>
      <div>
        <h3 className="text-muted-foreground text-sm font-medium">{label}</h3>
        <p className="text-3xl font-bold mt-1">
            {loading ? <span className="animate-pulse bg-zinc-800 rounded h-8 w-16 block"></span> : value}
        </p>
      </div>
    </motion.div>
  )
}
