<script setup>
import { useAuthStore } from '../stores/auth';
import { useRouter, useRoute } from 'vue-router';
import { Button } from '@/components/ui/button';
import { 
    LayoutDashboard, 
    Box, 
    Settings, 
    LogOut, 
    Search,
    Bell,
    Users,
    PieChart,
    PlusCircle
} from 'lucide-vue-next';

const auth = useAuthStore();
const router = useRouter();
const route = useRoute();

const logout = () => {
    auth.logout();
    router.push('/login');
};

const navItems = [
    { name: 'Dashboard', path: '/', icon: LayoutDashboard },
    { name: 'Projects', path: '/projects', icon: Box },
    { name: 'Analytics', path: '/analytics', icon: PieChart },
    { name: 'Team', path: '/team', icon: Users },
    { name: 'Settings', path: '/settings', icon: Settings },
];
</script>

<template>
  <div class="flex h-screen bg-[#F3F2EE] font-[Inter] overflow-hidden p-3 gap-3">
    <!-- Sidebar (Floating Bento Style) -->
    <aside class="w-20 bg-white rounded-[2rem] flex flex-col items-center py-8 shadow-sm h-full shrink-0">
        <!-- Logo -->
        <div class="mb-8">
            <div class="w-10 h-10 bg-black text-white rounded-xl flex items-center justify-center font-bold text-xl">
                S
            </div>
        </div>

        <!-- Nav Items -->
        <nav class="flex-1 w-full flex flex-col items-center gap-4 px-2">
            <router-link 
                v-for="item in navItems" 
                :key="item.path" 
                :to="item.path"
                class="w-10 h-10 flex items-center justify-center rounded-2xl transition-all duration-300 group relative"
                :class="route.path === item.path 
                    ? 'bg-black text-white shadow-lg shadow-black/20' 
                    : 'text-gray-400 hover:bg-gray-100 hover:text-gray-600'"
            >
                <component :is="item.icon" class="w-5 h-5" />
                
                <!-- Tooltip -->
                <span class="absolute left-14 bg-black text-white text-xs px-2 py-1 rounded-md opacity-0 group-hover:opacity-100 transition-opacity whitespace-nowrap z-20 pointer-events-none">
                    {{ item.name }}
                </span>
            </router-link>
            
            <div class="w-8 h-[1px] bg-gray-200 mt-2"></div>
             
             <!-- Special Action -->
             <router-link to="/projects/new" class="w-10 h-10 flex items-center justify-center rounded-2xl text-gray-400 hover:bg-blue-50 hover:text-blue-600 transition-colors">
                 <PlusCircle class="w-5 h-5" />
             </router-link>
        </nav>

        <!-- User Profile (Bottom) -->
        <div class="mt-auto flex flex-col items-center gap-4">
             <div class="w-8 h-8 rounded-full bg-indigo-100 flex items-center justify-center text-xs font-bold text-indigo-600 cursor-pointer hover:ring-2 hover:ring-offset-2 hover:ring-indigo-500 transition-all">
                 {{ auth.user?.email?.[0]?.toUpperCase() || 'U' }}
             </div>
             <button @click="logout" class="text-gray-300 hover:text-red-500 transition-colors">
               <LogOut class="w-5 h-5" />
             </button>
        </div>
    </aside>

    <!-- Main Content Wrapper -->
    <div class="flex-1 bg-white rounded-[2rem] shadow-sm flex flex-col overflow-hidden relative">
        <!-- Top Header (Integrated) -->
        <header class="h-20 flex items-center justify-between px-8 shrink-0">
             <!-- Breadcrumb / Page Title -->
            <div class="flex flex-col">
                <span class="text-xs font-semibold text-gray-400 uppercase tracking-widest mb-1">Overview</span>
                <h1 class="text-2xl font-bold text-gray-900 leading-none">
                    {{ route.name ? route.name.charAt(0).toUpperCase() + route.name.slice(1) : 'Operations' }}
                </h1>
            </div>

            <!-- Right Actions -->
            <div class="flex items-center gap-2 bg-[#F3F2EE] p-1.5 rounded-full">
                <!-- Search -->
                <div class="relative px-2">
                    <Search class="w-4 h-4 text-gray-400" />
                </div>
                
                <div class="h-4 w-[1px] bg-gray-300 mx-1"></div>

                <Button variant="ghost" size="sm" class="rounded-full h-8 px-4 text-xs font-medium hover:bg-white text-gray-600">
                    Feedback
                </Button>
                <Button variant="ghost" size="sm" class="rounded-full h-8 px-4 text-xs font-medium bg-white text-gray-900 shadow-sm">
                    Reports
                </Button>
                <Button size="icon" class="rounded-full h-8 w-8 bg-black hover:bg-gray-800 text-white shadow-md">
                    <Bell class="w-4 h-4" />
                </Button>
            </div>
        </header>

        <!-- Main Scrollable Area -->
        <main class="flex-1 overflow-auto p-8 pt-0 scroll-smooth">
            <div class="max-w-7xl mx-auto h-full">
                <slot />
            </div>
        </main>
    </div>
  </div>
</template>

<style scoped>
/* Precise warm background matching the refined aesthetic */
</style>
