<script setup>
import { ref } from 'vue';
import DashboardLayout from '@/layouts/DashboardLayout.vue';
import { Button } from '@/components/ui/button';
import { 
    User, 
    Lock, 
    Bell, 
    Palette, 
    Key, 
    ShieldCheck,
    Terminal
} from 'lucide-vue-next';
import { useAuthStore } from '@/stores/auth';

const auth = useAuthStore();
const activeTab = ref('general');

const menuItems = [
    { id: 'general', label: 'General', icon: User },
    { id: 'security', label: 'Security', icon: Lock },
    { id: 'ssh', label: 'SSH Keys', icon: Terminal },
    { id: 'billing', label: 'Billing', icon: ShieldCheck },
];
</script>

<template>
  <DashboardLayout>
    <div class="grid grid-cols-1 lg:grid-cols-4 gap-8 h-full">
        <!-- Settings Sidebar (Bento Style) -->
        <div class="lg:col-span-1 space-y-4">
            <div class="bg-gray-50 rounded-[2rem] p-6 h-full min-h-[400px]">
                <h3 class="font-bold text-xl mb-6 px-2">Settings</h3>
                <nav class="space-y-2">
                    <button 
                        v-for="item in menuItems" 
                        :key="item.id"
                        class="w-full flex items-center gap-3 px-4 py-3 rounded-xl transition-all font-medium text-sm"
                        :class="activeTab === item.id 
                            ? 'bg-black text-white shadow-lg' 
                            : 'text-gray-500 hover:bg-white hover:text-black hover:shadow-sm'"
                        @click="activeTab = item.id"
                    >
                        <component :is="item.icon" class="w-4 h-4" />
                        {{ item.label }}
                    </button>
                </nav>

                <div class="mt-8 bg-amber-50 rounded-xl p-4 border border-amber-100">
                    <div class="flex items-center gap-2 mb-2 text-amber-800 font-bold text-sm">
                        <Key class="w-4 h-4" />
                        Pro Tip
                    </div>
                    <p class="text-xs text-amber-900/70 leading-relaxed">
                        Add your public SSH key to enable git push deployments without password prompts.
                    </p>
                </div>
            </div>
        </div>

        <!-- Main Settings Content (Bento Grid) -->
        <div class="lg:col-span-3">
             <!-- GENERAL TAB -->
             <div v-if="activeTab === 'general'" class="grid grid-cols-1 md:grid-cols-2 gap-4">
                <!-- Profile Card -->
                <div class="md:col-span-2 bg-[#F3F2EE] rounded-[2rem] p-8 flex flex-col md:flex-row items-center gap-6 relative group overflow-hidden">
                    <div class="w-24 h-24 rounded-full bg-white flex items-center justify-center text-3xl font-bold shadow-sm z-10">
                        {{ auth.user?.email?.[0]?.toUpperCase() }}
                    </div>
                    <div class="z-10 text-center md:text-left">
                        <h2 class="text-2xl font-bold text-gray-900">Melvin</h2>
                        <p class="text-gray-500">{{ auth.user?.email }}</p>
                         <div class="flex gap-2 mt-4 justify-center md:justify-start">
                             <Button size="sm" class="rounded-full bg-black text-white px-6">Upload Avatar</Button>
                             <Button size="sm" variant="outline" class="rounded-full bg-white border-0 hover:bg-gray-100">Delete</Button>
                         </div>
                    </div>
                    <!-- Decorative Circle -->
                    <div class="absolute -right-10 -top-10 w-48 h-48 bg-white/50 rounded-full blur-3xl group-hover:bg-white/80 transition-all"></div>
                </div>

                <!-- Theme Preference -->
                <div class="bg-white border rounded-[2rem] p-6 flex flex-col justify-between h-[200px] hover:shadow-md transition-shadow">
                    <div class="flex justify-between items-start">
                        <div class="w-10 h-10 rounded-full bg-blue-50 flex items-center justify-center">
                            <Palette class="w-5 h-5 text-blue-600" />
                        </div>
                        <span class="bg-green-100 text-green-700 text-[10px] font-bold px-2 py-1 rounded-full uppercase">Auto</span>
                    </div>
                    <div>
                        <h3 class="font-bold text-lg mb-1">Interface Theme</h3>
                        <p class="text-xs text-gray-400 mb-4">Select your preferred dashboard appearance.</p>
                        <div class="flex gap-2">
                             <div class="w-6 h-6 rounded-full bg-gray-900 cursor-pointer ring-2 ring-offset-2 ring-gray-900"></div>
                             <div class="w-6 h-6 rounded-full bg-white border border-gray-200 cursor-pointer hover:border-gray-400"></div>
                        </div>
                    </div>
                </div>

                <!-- Email Notifications -->
                 <div class="bg-purple-50 rounded-[2rem] p-6 flex flex-col justify-between h-[200px] hover:shadow-md transition-shadow">
                    <div class="flex justify-between items-start">
                        <div class="w-10 h-10 rounded-full bg-white flex items-center justify-center shadow-sm">
                            <Bell class="w-5 h-5 text-purple-600" />
                        </div>
                         <div class="w-10 h-6 bg-purple-200 rounded-full p-1 cursor-pointer flex justify-end">
                             <div class="w-4 h-4 bg-white rounded-full shadow-sm"></div>
                         </div>
                    </div>
                    <div>
                        <h3 class="font-bold text-lg mb-1 text-purple-900">Notifications</h3>
                        <p class="text-xs text-purple-900/60 ">Get emails about build failures and deployment success.</p>
                    </div>
                </div>
             </div>

             <!-- SSH TAB (Placeholder structure) -->
             <div v-else-if="activeTab === 'ssh'" class="bg-gray-900 text-gray-200 rounded-[2rem] p-8 min-h-[400px] flex flex-col relative overflow-hidden">
                <div class="absolute right-0 top-0 opacity-10">
                    <Terminal class="w-64 h-64 text-white" />
                </div>
                
                <h2 class="text-2xl font-bold text-white mb-2 z-10">SSH Keys</h2>
                <p class="text-gray-400 text-sm max-w-md z-10 mb-8">Manage SSH keys to access your repositories and servers securely.</p>
                
                <div class="bg-black/50 backdrop-blur-md rounded-xl p-4 border border-white/10 mb-4 z-10">
                    <div class="flex items-center justify-between mb-2">
                        <div class="flex items-center gap-3">
                             <Key class="w-4 h-4 text-emerald-400" />
                             <span class="font-mono text-sm font-bold text-emerald-400">melvin_laptop_key</span>
                        </div>
                        <button class="text-xs text-red-400 hover:text-red-300 hover:underline">Revoke</button>
                    </div>
                    <code class="text-[10px] text-gray-500 font-mono break-all">ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAI...</code>
                </div>

                 <Button class="bg-white text-black hover:bg-gray-200 mt-auto w-max z-10">
                    <PlusCircle class="w-4 h-4 mr-2" />
                    Add New Key
                </Button>
             </div>
        </div>
    </div>
  </DashboardLayout>
</template>
