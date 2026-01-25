<script setup>
import { ref } from 'vue';
import DashboardLayout from '@/layouts/DashboardLayout.vue';
import { Button } from '@/components/ui/button';
import { 
    Plus, 
    MoreHorizontal, 
    Shield, 
    Mail,
    UserPlus,
    Clock
} from 'lucide-vue-next';

// Dummy Team Data
const members = ref([
    { id: 1, name: 'Melvin', role: 'Owner', email: 'meldev@senvanda.local', avatar: 'https://api.dicebear.com/7.x/avataaars/svg?seed=Melvin', status: 'active', lastActive: 'Now' },
    { id: 2, name: 'Sarah Connor', role: 'DevOps', email: 'sarah@senvanda.local', avatar: 'https://api.dicebear.com/7.x/avataaars/svg?seed=Sarah', status: 'active', lastActive: '2h ago' },
    { id: 3, name: 'John Doe', role: 'Frontend', email: 'john@senvanda.local', avatar: 'https://api.dicebear.com/7.x/avataaars/svg?seed=John', status: 'offline', lastActive: '2d ago' },
]);

// Dummy Invites
const invites = ref([
    { id: 1, email: 'alex@senvanda.local', role: 'Developer', sentAt: '1d ago' }
]);
</script>

<template>
  <DashboardLayout>
    <div class="grid grid-cols-1 lg:grid-cols-3 gap-8">
        <!-- Main Team Section (2 Cols) -->
        <div class="lg:col-span-2 space-y-6">
            <div class="flex items-center justify-between">
                <div>
                   <h2 class="text-2xl font-bold text-gray-900">Your Team</h2>
                   <p class="text-gray-500 text-sm">Manage access and roles for your organization.</p>
                </div>
                <Button class="rounded-full bg-black text-white hover:bg-gray-800">
                    <UserPlus class="w-4 h-4 mr-2" />
                    Invite Member
                </Button>
            </div>

            <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
                <!-- Member Card -->
                <div v-for="member in members" :key="member.id" class="bg-white rounded-[2rem] p-6 border border-gray-100 shadow-sm flex flex-col justify-between h-[200px] group hover:shadow-md transition-all">
                    <div class="flex justify-between items-start">
                        <div class="flex gap-4">
                            <div class="w-12 h-12 rounded-full bg-gray-50 overflow-hidden border-2 border-white shadow-sm">
                                <img :src="member.avatar" class="w-full h-full object-cover" />
                            </div>
                            <div>
                                <h3 class="font-bold text-gray-900">{{ member.name }}</h3>
                                <p class="text-xs text-gray-400 font-mono">{{ member.email }}</p>
                            </div>
                        </div>
                        <button class="w-8 h-8 rounded-full hover:bg-gray-50 flex items-center justify-center">
                            <MoreHorizontal class="w-4 h-4 text-gray-400" />
                        </button>
                    </div>
                    
                    <div class="mt-auto">
                        <div class="flex items-center gap-2 mb-4">
                            <span 
                                class="px-2.5 py-1 rounded-full text-[10px] font-bold uppercase tracking-wider border"
                                :class="member.role === 'Owner' 
                                    ? 'bg-indigo-50 text-indigo-700 border-indigo-100' 
                                    : 'bg-blue-50 text-blue-700 border-blue-100'"
                            >
                                {{ member.role }}
                            </span>
                            <span class="flex items-center text-xs text-gray-400 gap-1">
                                <span class="w-2 h-2 rounded-full block" :class="member.status === 'active' ? 'bg-green-400' : 'bg-gray-300'"></span>
                                {{ member.lastActive }}
                            </span>
                        </div>
                    </div>
                </div>

                <!-- Add New Placehoder -->
                 <div class="border-2 border-dashed border-gray-200 rounded-[2rem] p-6 flex flex-col items-center justify-center h-[200px] text-center cursor-pointer hover:bg-gray-50 transition-colors group">
                     <div class="w-12 h-12 rounded-full bg-gray-100 flex items-center justify-center mb-3 group-hover:scale-110 transition-transform">
                         <Plus class="w-6 h-6 text-gray-400 group-hover:text-black" />
                     </div>
                     <p class="font-bold text-gray-500 group-hover:text-black">Add Seat</p>
                     <p class="text-xs text-gray-400">Collaborate with more devs</p>
                 </div>
            </div>
        </div>

        <!-- Sidebar / Pending Invites (1 Col) -->
        <div class="lg:col-span-1 space-y-6">
            <!-- Pending Invites Card -->
            <div class="bg-[#F3F2EE] rounded-[2rem] p-6">
                <div class="flex items-center gap-2 mb-6">
                    <Mail class="w-5 h-5 text-gray-700" />
                    <h3 class="font-bold text-gray-900">Pending Invites</h3>
                </div>
                
                <div class="space-y-3">
                    <div v-for="invite in invites" :key="invite.id" class="bg-white p-4 rounded-xl shadow-sm flex items-center justify-between">
                         <div class="flex items-center gap-3">
                             <div class="w-8 h-8 rounded-full bg-orange-100 text-orange-600 flex items-center justify-center font-bold text-xs">
                                 {{ invite.email[0].toUpperCase() }}
                             </div>
                             <div>
                                 <p class="text-sm font-bold text-gray-800 leading-tight">Alex</p>
                                 <p class="text-[10px] text-gray-400">{{ invite.email }}</p>
                             </div>
                         </div>
                         <button class="text-[10px] font-bold text-red-400 hover:text-red-500">Revoke</button>
                    </div>
                </div>
                
                <div class="mt-6 pt-6 border-t border-gray-200/50">
                    <h4 class="font-bold text-sm text-gray-900 mb-2">Role Permissions</h4>
                    <div class="space-y-2 text-xs text-gray-500">
                        <div class="flex justify-between">
                            <span>Owner</span>
                            <span class="text-gray-900 font-medium">Full Access</span>
                        </div>
                        <div class="flex justify-between">
                            <span>Admin</span>
                            <span class="text-gray-900 font-medium">Deploy & Manage</span>
                        </div>
                        <div class="flex justify-between">
                            <span>Developer</span>
                            <span class="text-gray-900 font-medium">Push & Logs</span>
                        </div>
                    </div>
                </div>
            </div>
            
            <!-- Security Audit Preview -->
             <div class="bg-gray-900 text-white rounded-[2rem] p-6 relative overflow-hidden">
                 <div class="relative z-10">
                     <div class="w-10 h-10 rounded-xl bg-white/10 flex items-center justify-center mb-4 backdrop-blur-sm">
                         <Shield class="w-5 h-5" />
                     </div>
                     <h3 class="font-bold text-lg">Audit Log</h3>
                     <p class="text-xs text-gray-400 mt-1 mb-4">Monitor team actions for security compliance.</p>
                     
                     <div class="space-y-2 font-mono text-[10px] text-emerald-400">
                         <p>> Sarah deployed `api-gateway`</p>
                         <p>> Melvin updated env vars</p>
                         <p class="text-gray-500">> John failed login (3x)</p>
                     </div>
                 </div>
             </div>
        </div>
    </div>
  </DashboardLayout>
</template>
