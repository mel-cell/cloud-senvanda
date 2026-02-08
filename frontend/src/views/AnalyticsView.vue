<script setup>
import { ref, onMounted, onUnmounted, computed } from 'vue';
import DashboardLayout from '@/layouts/DashboardLayout.vue';
import ResourceGraph from '@/components/ResourceGraph.vue';
import { 
    Activity, 
    Cpu, 
    HardDrive, 
    Server,
    Zap
} from 'lucide-vue-next';
import { pb } from '@/lib/pocketbase';

const stats = ref(null);
const loading = ref(true);
let timer = null;

// Helper: Format Bytes
const formatBytes = (bytes, decimals = 2) => {
    if (!+bytes) return '0 Bytes';
    const k = 1024;
    const dm = decimals < 0 ? 0 : decimals;
    const sizes = ['Bytes', 'KB', 'MB', 'GB', 'TB'];
    const i = Math.floor(Math.log(bytes) / Math.log(k));
    return `${parseFloat((bytes / Math.pow(k, i)).toFixed(dm))} ${sizes[i]}`;
};

const fetchStats = async () => {
    try {
        const res = await pb.send('/api/senvanda/deploy/cluster/stats');
        stats.value = res;
    } catch(e) {
        console.error("Cluster stats error", e);
    } finally {
        loading.value = false;
    }
};

onMounted(() => {
    fetchStats();
    timer = setInterval(fetchStats, 3000);
});

onUnmounted(() => {
    if(timer) clearInterval(timer);
});
</script>

<template>
  <DashboardLayout>
    <div class="space-y-6 animate-in fade-in duration-500">
        <!-- HEADER: CLUSTER HEALTH -->
        <div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-6">
             <!-- CPU Stats -->
             <div class="bg-black text-white rounded-[2rem] p-6 shadow-lg flex flex-col justify-between h-48 relative overflow-hidden group">
                 <div class="absolute right-0 top-0 w-32 h-32 bg-white/10 rounded-full blur-3xl -mr-10 -mt-10 group-hover:scale-150 transition-transform duration-700"></div>
                <div class="flex justify-between items-start z-10">
                    <div class="p-3 bg-white/20 rounded-2xl">
                        <Cpu class="w-6 h-6" />
                    </div>
                    <span v-if="stats" class="text-xs font-mono text-white/60 bg-white/10 px-2 py-1 rounded">
                        {{ stats.host_info.cpu_cores }} Cores
                    </span>
                </div>
                <div class="z-10">
                    <div class="flex items-end gap-2">
                        <h3 class="text-4xl font-bold tracking-tight">
                            {{ stats ? stats.total_cpu.toFixed(1) : 0 }}%
                        </h3>
                    </div>
                    
                    <div class="w-full h-1 bg-white/20 rounded-full mt-4 overflow-hidden">
                        <div class="h-full bg-emerald-400 rounded-full transition-all duration-500" :style="`width: ${Math.min((stats?.total_cpu || 0), 100)}%`"></div>
                    </div>
                     <p class="text-xs font-medium text-white/50 uppercase tracking-wider mt-2">Total CPU Load</p>
                </div>
             </div>
             
             <!-- Memory Stats -->
             <div class="bg-white rounded-[2rem] p-6 shadow-sm border border-gray-100 flex flex-col justify-between h-48">
                <div class="flex justify-between items-start">
                    <div class="p-3 bg-orange-50 text-orange-600 rounded-2xl">
                        <HardDrive class="w-6 h-6" />
                    </div>
                </div>
                <div>
                     <div class="flex items-end gap-2">
                        <h3 class="text-4xl font-bold tracking-tight text-gray-900">
                             {{ stats ? formatBytes(stats.total_memory) : '0 B' }}
                        </h3>
                    </div>
                     <div class="w-full h-1 bg-gray-100 rounded-full mt-4 overflow-hidden">
                         <!-- Simplified Progress: Assume 8GB max for visual visualization if Host mem unknown -->
                        <div class="h-full bg-orange-400 rounded-full transition-all duration-500" :style="`width: ${Math.min((stats?.total_memory / (8*1024*1024*1024)) * 100, 100)}%`"></div>
                    </div>
                    <p class="text-xs font-medium text-gray-400 uppercase tracking-wider mt-2">Cluster RAM Usage</p>
                </div>
             </div>

             <!-- Active Projects -->
             <div class="bg-white rounded-[2rem] p-6 shadow-sm border border-gray-100 flex flex-col justify-between h-48">
                <div class="flex justify-between items-start">
                    <div class="p-3 bg-blue-50 text-blue-600 rounded-2xl">
                        <Activity class="w-6 h-6" />
                    </div>
                    <span class="text-xs font-bold text-green-600 bg-green-50 px-2 py-1 rounded-full flex items-center gap-1">
                        <div class="w-2 h-2 rounded-full bg-green-500 animate-pulse"></div> Online
                    </span>
                </div>
                <div>
                    <h3 class="text-4xl font-bold text-gray-900 tracking-tight">
                        {{ stats ? stats.active_projects : 0 }}
                    </h3>
                    <p class="text-xs font-medium text-gray-400 uppercase tracking-wider mt-2">Active Containers</p>
                </div>
             </div>

             <!-- Host Info -->
             <div class="bg-[#F3F2EE] rounded-[2rem] p-6 shadow-sm border border-gray-200/50 flex flex-col justify-between h-48">
                <div class="flex justify-between items-start">
                    <div class="p-3 bg-white text-gray-700 rounded-2xl shadow-sm">
                        <Server class="w-6 h-6" />
                    </div>
                </div>
                <div>
                    <h3 class="text-xl font-bold text-gray-900 tracking-tight truncate">
                        {{ stats ? stats.host_info.hostname : 'Loading...' }}
                    </h3>
                    <div class="flex gap-2 mt-2">
                         <span class="text-[10px] font-bold px-2 py-1 bg-white rounded border border-gray-200 uppercase">
                            {{ stats ? stats.host_info.os : 'Unknown' }}
                         </span>
                         <span class="text-[10px] font-bold px-2 py-1 bg-white rounded border border-gray-200 uppercase">
                            x86_64
                         </span>
                    </div>
                    <p class="text-xs font-medium text-gray-400 uppercase tracking-wider mt-2">Host Machine</p>
                </div>
             </div>
        </div>

        <!-- ROW 2: TOP CONSUMERS -->
        <div class="grid grid-cols-1 lg:grid-cols-3 gap-6">
            <!-- Top Apps Table -->
            <div class="lg:col-span-2 bg-white rounded-[2rem] p-8 shadow-sm border border-gray-100 min-h-[400px]">
                <div class="flex items-center justify-between mb-8">
                     <div>
                        <h3 class="font-bold text-xl text-gray-900 flex items-center gap-2">
                            <Zap class="w-5 h-5 text-yellow-500" />
                            Top Resource Consumers
                        </h3>
                        <p class="text-sm text-gray-400">Projects with highest CPU load right now</p>
                    </div>
                </div>
                
                <div v-if="loading" class="space-y-4">
                    <div v-for="i in 3" :key="i" class="h-16 bg-gray-50 rounded-xl animate-pulse"></div>
                </div>

                <div v-else-if="stats?.top_consumers?.length" class="space-y-4">
                    <div 
                        v-for="(app, i) in stats.top_consumers" 
                        :key="i"
                        class="p-4 rounded-2xl border border-gray-100 hover:border-black/10 hover:shadow-md transition-all flex items-center justify-between group bg-gray-50/30"
                    >
                        <div class="flex items-center gap-4">
                            <div class="w-10 h-10 rounded-xl bg-gray-900 text-white flex items-center justify-center font-bold text-sm">
                                {{ (app.project_name || '?')[0].toUpperCase() }}
                            </div>
                            <div>
                                <h4 class="font-bold text-gray-900">{{ app.project_name || 'Unknown' }}</h4>
                                <p class="text-xs text-gray-400 font-mono">ID: {{ app.id?.substring(0,8) }}...</p>
                            </div>
                        </div>
                        
                        <div class="flex items-center gap-8">
                             <div class="text-right">
                                <p class="text-[10px] font-bold uppercase text-gray-400">CPU</p>
                                <p class="text-sm font-bold font-mono" :class="app.cpu_percent > 50 ? 'text-red-500' : 'text-gray-700'">
                                    {{ app.cpu_percent.toFixed(2) }}%
                                </p>
                            </div>
                             <div class="text-right w-24">
                                <p class="text-[10px] font-bold uppercase text-gray-400">Memory</p>
                                <p class="text-sm font-bold font-mono text-gray-700">
                                    {{ formatBytes(app.memory_bytes) }}
                                </p>
                            </div>
                        </div>
                    </div>
                </div>
                
                <div v-else class="text-center py-12 text-gray-400">
                    No active projects found.
                </div>
            </div>

            <!-- Side Metric (Dummy for now or reuse) -->
            <div class="bg-[#F3F2EE] rounded-[2rem] p-8 flex flex-col justify-center items-center text-center">
                 <div class="p-6 bg-white rounded-full shadow-sm mb-6">
                    <Activity class="w-12 h-12 text-gray-900" />
                 </div>
                 <h3 class="font-bold text-xl text-gray-900 mb-2">System Healthy</h3>
                 <p class="text-sm text-gray-500 mb-8 max-w-[200px]">All systems are operational. No critical alerts detected in the cluster.</p>
                 <button class="w-full py-4 bg-white text-black font-bold rounded-xl shadow-sm hover:shadow-md transition-all">
                     View System Logs
                 </button>
            </div>
        </div>
    </div>
  </DashboardLayout>
</template>

