<script setup>
import { ref } from 'vue';
import DashboardLayout from '@/layouts/DashboardLayout.vue';
import { 
    Database, 
    Monitor, 
    ShieldCheck, 
    Zap, 
    Search,
    Rocket,
    Clock,
    Plus
} from 'lucide-vue-next';

const categories = ['All', 'Databases', 'Frameworks', 'VMs', 'Security'];
const activeCat = ref('All');
const search = ref('');

const items = [
    {
        id: 'postgres',
        name: 'PostgreSQL',
        desc: 'Advanced open source relational database.',
        icon: Database,
        color: 'text-blue-600 bg-blue-50',
        tag: 'Databases',
        version: '15.0',
        category: 'other'
    },
    {
        id: 'redis',
        name: 'Redis',
        desc: 'In-memory data structure store, used as a database, cache, and broker.',
        icon: Zap,
        color: 'text-red-600 bg-red-50',
        tag: 'Databases',
        version: '7.0',
        category: 'other'
    },
    {
        id: 'mysql',
        name: 'MySQL',
        desc: 'Reliable, scalable, and fast relational database.',
        icon: Database,
        color: 'text-orange-600 bg-orange-50',
        tag: 'Databases',
        version: '8.0',
        category: 'other'
    },
    {
        id: 'ubuntu-vps',
        name: 'Ubuntu VM',
        desc: 'High performance Ubuntu 22.04 LTS instance.',
        icon: Monitor,
        color: 'text-purple-600 bg-purple-50',
        tag: 'VMs',
        version: '22.04',
        category: 'vm'
    },
    {
        id: 'nginx',
        name: 'Nginx Proxy',
        desc: 'High-performance load balancer and web server.',
        icon: ShieldCheck,
        color: 'text-emerald-600 bg-emerald-50',
        tag: 'Security',
        version: 'Stable',
        category: 'infrastructure'
    }
];

import { pb } from '@/lib/pocketbase';
import { useRouter } from 'vue-router';
const router = useRouter();
const isDeploying = ref(null);
const showDeployModal = ref(false);
const selectedItem = ref(null);
const newProjectName = ref('');

const openDeployModal = (item) => {
    selectedItem.value = item;
    newProjectName.value = `${item.id}-${Math.floor(Math.random()*1000)}`;
    showDeployModal.value = true;
};

const handleConfirmDeploy = async () => {
    if (!newProjectName.value || !selectedItem.value) return;

    const item = selectedItem.value;
    isDeploying.value = item.id;
    showDeployModal.value = false;
    
    try {
        const res = await pb.send('/api/senvanda/deploy/marketplace', {
            method: 'POST',
            body: {
                itemId: item.id,
                name: newProjectName.value,
                category: item.category
            }
        });
        router.push(`/projects/${res.id}`);
    } catch (e) {
        alert('Deployment failed: ' + e.message);
    } finally {
        isDeploying.value = null;
    }
};
</script>

<template>
  <DashboardLayout>
    <div class="space-y-8 animate-in fade-in duration-500 relative">
        <!-- MODAL OVERLAY -->
        <Transition
            enter-active-class="transition duration-300 ease-out"
            enter-from-class="opacity-0 scale-95"
            enter-to-class="opacity-100 scale-100"
            leave-active-class="transition duration-200 ease-in"
            leave-from-class="opacity-100 scale-100"
            leave-to-class="opacity-0 scale-95"
        >
            <div v-if="showDeployModal" class="fixed inset-0 z-50 flex items-center justify-center p-4 bg-black/40 backdrop-blur-sm">
                <div class="bg-white rounded-[2.5rem] w-full max-w-md p-10 shadow-2xl border border-gray-100 animate-in zoom-in-95 duration-300">
                    <div class="flex justify-between items-start mb-8">
                        <div class="p-4 rounded-2xl" :class="selectedItem.color">
                             <component :is="selectedItem.icon" class="w-8 h-8" />
                        </div>
                        <button @click="showDeployModal = false" class="p-2 hover:bg-gray-100 rounded-full text-gray-400 transition-colors">✕</button>
                    </div>

                    <h3 class="text-2xl font-bold text-gray-900">Deploy {{ selectedItem.name }}</h3>
                    <p class="text-gray-500 text-sm mt-2">Initialize a new production-ready instance of {{ selectedItem.name }}.</p>

                    <div class="mt-8 space-y-4">
                        <div class="space-y-2">
                             <label class="text-xs font-bold text-gray-400 uppercase tracking-widest pl-1">Instance Name</label>
                             <input 
                                v-model="newProjectName"
                                type="text" 
                                autofocus
                                class="w-full px-6 py-4 bg-gray-50 border-2 border-transparent focus:border-black focus:bg-white rounded-2xl outline-none transition-all font-medium text-gray-900"
                                placeholder="e.g. production-db"
                             />
                        </div>
                    </div>

                    <div class="mt-10 flex gap-3">
                         <button @click="showDeployModal = false" class="flex-1 py-4 text-gray-500 font-bold hover:bg-gray-50 rounded-2xl transition-all">Cancel</button>
                         <button @click="handleConfirmDeploy" class="flex-1 py-4 bg-black text-white font-bold rounded-2xl shadow-lg shadow-black/20 hover:scale-[1.02] active:scale-95 transition-all flex items-center justify-center gap-2">
                             <Rocket class="w-5 h-5" /> Deploy
                         </button>
                    </div>
                </div>
            </div>
        </Transition>

        <!-- HEADER -->
        <div class="flex flex-col md:flex-row md:items-end justify-between gap-4">
             <div class="max-w-xl">
                <h2 class="text-3xl font-bold tracking-tight text-gray-900">Cloud Marketplace</h2>
                <p class="text-gray-500 mt-2">Deploy production-ready databases and services in seconds. Managed and optimized for Senvanda infrastructure.</p>
             </div>
             
             <!-- SEARCH -->
             <div class="relative w-full md:w-80">
                <Search class="absolute left-4 top-1/2 -translate-y-1/2 w-4 h-4 text-gray-400" />
                <input 
                    v-model="search"
                    type="text" 
                    placeholder="Search resources..." 
                    class="w-full pl-12 pr-4 py-3 bg-[#F3F2EE] border-none rounded-2xl outline-none focus:ring-2 focus:ring-black transition-all"
                />
             </div>
        </div>

        <!-- CATEGORY PILLS -->
        <div class="flex items-center gap-2 overflow-x-auto pb-2 scrollbar-hide">
            <button 
                v-for="cat in categories" 
                :key="cat"
                @click="activeCat = cat"
                class="px-5 py-2.5 rounded-full text-sm font-bold transition-all border shrink-0"
                :class="activeCat === cat ? 'bg-black text-white border-black' : 'bg-white text-gray-500 border-gray-100 hover:bg-gray-50'"
            >
                {{ cat }}
            </button>
        </div>

        <!-- MARKETPLACE GRID -->
        <div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4 gap-6">
            <div 
                v-for="item in items" 
                :key="item.id"
                class="bg-white rounded-[2.5rem] p-8 border border-gray-100 shadow-sm hover:shadow-xl hover:shadow-gray-200/50 transition-all cursor-pointer group flex flex-col justify-between min-h-[340px] relative overflow-hidden"
            >
                <!-- Background Pattern -->
                <div class="absolute -right-4 -top-4 w-32 h-32 opacity-[0.03] text-black group-hover:scale-110 transition-transform">
                    <component :is="item.icon" class="w-full h-full" />
                </div>

                <div>
                    <div class="flex justify-between items-start mb-6">
                        <div class="p-4 rounded-2xl shadow-sm" :class="item.color">
                            <component :is="item.icon" class="w-7 h-7" />
                        </div>
                        <span class="text-[10px] font-bold px-2 py-1 bg-gray-50 text-gray-400 rounded-lg uppercase tracking-wider border border-gray-100 italic">
                            v{{ item.version }}
                        </span>
                    </div>

                    <h3 class="text-xl font-bold text-gray-900 group-hover:text-blue-600 transition-colors">{{ item.name }}</h3>
                    <p class="text-sm text-gray-400 mt-2 line-clamp-3 leading-relaxed">{{ item.desc }}</p>
                </div>

                <div class="mt-8 flex flex-col gap-3">
                     <div class="flex items-center gap-2">
                        <span class="text-[10px] font-bold bg-gray-900 text-white px-2 py-0.5 rounded-md uppercase">{{ item.tag }}</span>
                        <div class="h-1 w-1 rounded-full bg-gray-300"></div>
                        <span class="text-[10px] font-bold text-gray-400 uppercase flex items-center gap-1">
                             <Clock class="w-3 h-3" /> Auto Update
                        </span>
                     </div>
                     
                     <button 
                        @click="openDeployModal(item)"
                        :disabled="isDeploying === item.id"
                        class="w-full bg-[#F3F2EE] hover:bg-black hover:text-white text-gray-900 py-4 rounded-2xl font-bold transition-all flex items-center justify-center gap-2 group-hover:bg-gray-100 group-hover:hover:bg-black disabled:opacity-50 disabled:cursor-not-allowed"
                     >
                        <component :is="isDeploying === item.id ? Clock : Plus" class="w-5 h-5" :class="{'animate-spin': isDeploying === item.id}" />
                        {{ isDeploying === item.id ? 'Deploying...' : 'Deploy Instance' }}
                     </button>
                </div>
            </div>
            
            <!-- COMING SOON CARD -->
            <div class="bg-gray-50/50 rounded-[2.5rem] p-8 border border-dashed border-gray-200 flex flex-col items-center justify-center text-center opacity-60">
                 <div class="w-16 h-16 rounded-full bg-white flex items-center justify-center shadow-sm mb-4">
                    <Rocket class="w-8 h-8 text-gray-300" />
                 </div>
                 <h4 class="font-bold text-gray-400">Request App</h4>
                 <p class="text-xs text-gray-400 mt-1">Want another service?<br/>Suggest it!</p>
            </div>
        </div>
    </div>
  </DashboardLayout>
</template>

<style scoped>
.scrollbar-hide::-webkit-scrollbar {
    display: none;
}
.scrollbar-hide {
    -ms-overflow-style: none;
    scrollbar-width: none;
}
</style>
