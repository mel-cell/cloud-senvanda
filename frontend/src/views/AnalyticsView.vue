<script setup>
import DashboardLayout from '@/layouts/DashboardLayout.vue';
import { 
    ArrowUpRight, 
    ArrowDownRight, 
    Activity, 
    Cpu, 
    HardDrive, 
    Globe,    
    Clock
} from 'lucide-vue-next';

// Dummy Data Generators for visuals
const generateBars = (count) => Array.from({ length: count }, () => Math.floor(Math.random() * 80) + 10);
const trafficData = generateBars(24);
</script>

<template>
  <DashboardLayout>
    <div class="grid grid-cols-1 lg:grid-cols-4 gap-6">
        <!-- HEADER STATS (4 Col) -->
        <div class="lg:col-span-4 grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4">
             <div class="bg-white rounded-[2rem] p-6 shadow-sm border border-gray-100 flex flex-col justify-between h-40">
                <div class="flex justify-between items-start">
                    <div class="p-3 bg-blue-50 text-blue-600 rounded-2xl">
                        <Activity class="w-6 h-6" />
                    </div>
                    <span class="flex items-center text-xs font-bold text-green-600 bg-green-50 px-2 py-1 rounded-full">
                        +12.5% <ArrowUpRight class="w-3 h-3 ml-1" />
                    </span>
                </div>
                <div>
                    <h3 class="text-3xl font-bold text-gray-900 tracking-tight">142.3k</h3>
                    <p class="text-xs font-medium text-gray-400 uppercase tracking-wider mt-1">Total Requests</p>
                </div>
             </div>

             <div class="bg-white rounded-[2rem] p-6 shadow-sm border border-gray-100 flex flex-col justify-between h-40">
                <div class="flex justify-between items-start">
                    <div class="p-3 bg-purple-50 text-purple-600 rounded-2xl">
                        <Clock class="w-6 h-6" />
                    </div>
                    <span class="flex items-center text-xs font-bold text-red-500 bg-red-50 px-2 py-1 rounded-full">
                        -2.1% <ArrowDownRight class="w-3 h-3 ml-1" />
                    </span>
                </div>
                <div>
                    <h3 class="text-3xl font-bold text-gray-900 tracking-tight">42<span class="text-lg text-gray-400 font-normal ml-1">ms</span></h3>
                    <p class="text-xs font-medium text-gray-400 uppercase tracking-wider mt-1">Avg Latency</p>
                </div>
             </div>

             <div class="bg-black text-white rounded-[2rem] p-6 shadow-lg flex flex-col justify-between h-40 relative overflow-hidden group">
                 <div class="absolute right-0 top-0 w-32 h-32 bg-white/10 rounded-full blur-3xl -mr-10 -mt-10 group-hover:scale-150 transition-transform duration-700"></div>
                <div class="flex justify-between items-start z-10">
                    <div class="p-3 bg-white/20 rounded-2xl">
                        <Cpu class="w-6 h-6" />
                    </div>
                </div>
                <div class="z-10">
                    <div class="flex items-end gap-2">
                        <h3 class="text-3xl font-bold tracking-tight">24%</h3>
                        <span class="text-xs text-white/50 mb-1.5 font-mono">/ 4 CORES</span>
                    </div>
                    
                    <div class="w-full h-1 bg-white/20 rounded-full mt-3 overflow-hidden">
                        <div class="h-full bg-emerald-400 w-1/4 rounded-full"></div>
                    </div>
                     <p class="text-xs font-medium text-white/50 uppercase tracking-wider mt-2">CPU Load</p>
                </div>
             </div>
             
             <div class="bg-white rounded-[2rem] p-6 shadow-sm border border-gray-100 flex flex-col justify-between h-40">
                <div class="flex justify-between items-start">
                    <div class="p-3 bg-orange-50 text-orange-600 rounded-2xl">
                        <HardDrive class="w-6 h-6" />
                    </div>
                </div>
                <div>
                     <div class="flex items-end gap-2">
                        <h3 class="text-3xl font-bold tracking-tight text-gray-900">1.2<span class="text-lg text-gray-400 font-normal">GB</span></h3>
                        <span class="text-xs text-gray-400 mb-1.5 font-mono">/ 4 GB</span>
                    </div>
                     <div class="w-full h-1 bg-gray-100 rounded-full mt-3 overflow-hidden">
                        <div class="h-full bg-orange-400 w-1/3 rounded-full"></div>
                    </div>
                    <p class="text-xs font-medium text-gray-400 uppercase tracking-wider mt-2">Memory Usage</p>
                </div>
             </div>
        </div>

        <!-- MAIN CHART (Col Span 3) -->
        <div class="lg:col-span-3 bg-white rounded-[2rem] p-8 shadow-sm border border-gray-100 min-h-[400px]">
            <div class="flex items-center justify-between mb-8">
                 <div>
                    <h3 class="font-bold text-xl text-gray-900">Traffic Overview</h3>
                    <p class="text-sm text-gray-400">Incoming requests over the last 24 hours</p>
                </div>
                
                 <div class="flex gap-2">
                     <button class="px-4 py-2 bg-black text-white text-xs font-bold rounded-xl">24H</button>
                      <button class="px-4 py-2 bg-gray-50 text-gray-500 hover:text-black hover:bg-gray-100 text-xs font-bold rounded-xl transition-colors">7D</button>
                       <button class="px-4 py-2 bg-gray-50 text-gray-500 hover:text-black hover:bg-gray-100 text-xs font-bold rounded-xl transition-colors">30D</button>
                 </div>
            </div>
            
            <!-- Custom CSS Bar Chart for aesthetics -->
            <div class="flex items-end justify-between h-64 w-full gap-2">
                <div 
                    v-for="(h, i) in trafficData" 
                    :key="i"
                    class="bg-gray-100 hover:bg-blue-500 hover:shadow-lg hover:shadow-blue-200 transition-all duration-300 w-full rounded-t-lg relative group cursor-pointer"
                    :style="`height: ${h}%`"
                >
                    <!-- Tooltip -->
                    <div class="absolute bottom-full mb-2 left-1/2 -translate-x-1/2 bg-black text-white text-[10px] py-1 px-2 rounded opacity-0 group-hover:opacity-100 transition-opacity pointer-events-none whitespace-nowrap z-10">
                        {{ Math.floor(h * 12) }} reqs
                    </div>
                </div>
            </div>
             <div class="flex justify-between mt-4 text-xs text-gray-300 font-mono">
                 <span>00:00</span>
                 <span>06:00</span>
                 <span>12:00</span>
                 <span>18:00</span>
                 <span>23:59</span>
             </div>
        </div>

        <!-- REGIONAL (Col Span 1) -->
        <div class="lg:col-span-1 bg-[#F3F2EE] rounded-[2rem] p-6 flex flex-col">
            <h3 class="font-bold text-xl text-gray-900 mb-6">Top Regions</h3>
            
            <div class="space-y-4 flex-1">
                <div class="bg-white p-4 rounded-2xl flex items-center justify-between shadow-sm border border-gray-200/50">
                    <div class="flex items-center gap-3">
                         <span class="text-xl">🇮🇩</span>
                         <span class="font-bold text-sm text-gray-700">Indonesia</span>
                    </div>
                    <span class="font-mono text-xs font-bold text-gray-400">82%</span>
                </div>
                 <div class="bg-white p-4 rounded-2xl flex items-center justify-between shadow-sm border border-gray-200/50">
                    <div class="flex items-center gap-3">
                         <span class="text-xl">🇸🇬</span>
                         <span class="font-bold text-sm text-gray-700">Singapore</span>
                    </div>
                    <span class="font-mono text-xs font-bold text-gray-400">12%</span>
                </div>
                 <div class="bg-white p-4 rounded-2xl flex items-center justify-between shadow-sm border border-gray-200/50">
                    <div class="flex items-center gap-3">
                         <span class="text-xl">🇺🇸</span>
                         <span class="font-bold text-sm text-gray-700">USA</span>
                    </div>
                    <span class="font-mono text-xs font-bold text-gray-400">4%</span>
                </div>
                 <div class="bg-white/50 p-4 rounded-2xl flex items-center justify-between shadow-sm border border-gray-200/50 opacity-60">
                    <div class="flex items-center gap-3">
                         <span class="text-xl">🇯🇵</span>
                         <span class="font-bold text-sm text-gray-700">Japan</span>
                    </div>
                    <span class="font-mono text-xs font-bold text-gray-400">1%</span>
                </div>
            </div>

            <button class="mt-6 w-full py-4 rounded-xl border border-gray-300 text-gray-500 text-xs font-bold hover:bg-white hover:text-black transition-colors uppercase tracking-wider flex items-center justify-center gap-2">
                <Globe class="w-4 h-4" />
                View Global Map
            </button>
        </div>
    </div>
  </DashboardLayout>
</template>
