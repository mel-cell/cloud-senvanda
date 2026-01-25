<script setup>
import { onMounted, ref, reactive, computed } from "vue";
import { useRouter } from "vue-router";
import DashboardLayout from "@/layouts/DashboardLayout.vue";
import ProjectCard from "@/components/ProjectCard.vue";
import LegacyContainerCard from "@/components/LegacyContainerCard.vue";
import { Button } from "@/components/ui/button";
import {
  Plus,
  Filter,
  Link,
  Paperclip,
  Heart,
  Share2,
  Calendar,
  ArrowUpRight,
  WifiOff,
  Trash2,
} from "lucide-vue-next";
import { pb } from "@/lib/pocketbase";

const router = useRouter();
const projects = ref([]);
const legacyApps = ref([]); // NEW: Discovery
const filter = ref("all");
const loading = ref(true);
const error = ref(null);

const loadProjects = async () => {
  loading.value = true;
  error.value = null;
  try {
    const [projectRes, legacyRes] = await Promise.all([
      pb.send("/api/senvanda/deploy/projects", { method: "GET" }),
      pb.send("/api/senvanda/deploy/legacy", { method: "GET" }),
    ]);

    projects.value = projectRes || [];
    legacyApps.value = legacyRes || [];
  } catch (err) {
    console.error("Failed to load apps", err);
    error.value = "Engine connection refused. Is the backend running?";
  } finally {
    loading.value = false;
  }
};

const handleProjectAction = async (id, action) => {
  // Special Client-Side Actions
  if (action === "resume") {
    // Logic to resume draft would go here.
    // For now, simpler to just start clean or send ID to wizard
    // router.push({ name: 'create-project', query: { draftId: id }})
    alert(
      "Resume Draft functionality coming next! For now, please create new.",
    );
    return;
  }

  try {
    await pb.send(`/api/senvanda/deploy/${id}/action`, {
      method: "POST",
      body: { action },
    });
    await loadProjects();
  } catch (err) {
    alert("Action failed: " + err.message);
  }
};

const filteredProjects = computed(() => {
  if (filter.value === "all") return projects.value;
  return projects.value.filter((p) => p.status === filter.value);
});

// Real Stats
const stats = computed(() => {
  return {
    running: projects.value.filter((p) => p.status === "running").length,
    draft: projects.value.filter((p) => p.status === "draft").length,
    stopped: projects.value.filter(
      (p) => p.status !== "running" && p.status !== "draft",
    ).length,
    total: projects.value.length,
    prunable: projects.value.filter(
      (p) => p.status !== "running" && p.status !== "draft",
    ).length,
  };
});

const isPruning = ref(false);
const handlePrune = async () => {
  if (!confirm("Hapus semua record project yang tidak ada kontainernya? (Termasuk record 'Untitled')")) return;
  
  isPruning.value = true;
  try {
    const res = await pb.send("/api/senvanda/deploy/prune", { method: "POST" });
    alert(`DevOps Hygiene: Berhasil membersihkan ${res.pruned_count} project hantu/untitled!`);
    await loadProjects();
  } catch (err) {
    alert("Gagal pruning: " + err.message);
  } finally {
    isPruning.value = false;
  }
};

const navigateToAdopt = async (container) => {
  try {
    await pb.send("/api/senvanda/deploy/adopt", {
      method: "POST",
      body: { containerID: container.ID },
    });
    alert(`Successfully adopted ${container.name}!`);
    await loadProjects();
  } catch (err) {
    alert("Adoption failed: " + err.message);
  }
};

onMounted(() => {
  loadProjects();
});
</script>

<template>
  <DashboardLayout>
    <div class="space-y-8">
      <!-- CONNECTION ERROR STATE -->
      <div
        v-if="error"
        class="bg-red-50 text-red-900 px-6 py-4 rounded-2xl flex items-center justify-between border border-red-100"
      >
        <div class="flex items-center gap-3">
          <WifiOff class="w-6 h-6 text-red-500" />
          <div>
            <h3 class="font-bold">System Offline</h3>
            <p class="text-sm opacity-80">{{ error }}</p>
          </div>
        </div>
        <Button
          size="sm"
          variant="outline"
          class="bg-white hover:bg-red-100 border-red-200"
          @click="loadProjects"
          >Retry Connection</Button
        >
      </div>

      <!-- Top Toolbar -->
      <div class="flex flex-wrap items-center justify-between gap-4">
        <!-- Left Filters (Pills) -->
        <div
          class="flex items-center gap-2 overflow-x-auto pb-2 scrollbar-hide"
        >
          <button
            class="px-5 py-2.5 rounded-full text-sm font-semibold transition-all border shrink-0"
            :class="
              filter === 'all'
                ? 'bg-black text-white hover:bg-gray-800'
                : 'bg-white text-gray-500 hover:bg-gray-50 border-gray-200'
            "
            @click="filter = 'all'"
          >
            All Items
            <span
              class="ml-1 opacity-60 text-xs bg-white/20 px-1.5 py-0.5 rounded-full"
              >{{ stats.total }}</span
            >
          </button>
          <button
            class="px-5 py-2.5 rounded-full text-sm font-semibold transition-all border shrink-0"
            :class="
              filter === 'running'
                ? 'bg-[#F2E8FF] text-purple-900 border-purple-100'
                : 'bg-white text-gray-500 hover:bg-gray-50 border-gray-200'
            "
            @click="filter = 'running'"
          >
            Running
            <span
              class="ml-1 opacity-60 text-xs bg-purple-900/10 px-1.5 py-0.5 rounded-full"
              >{{ stats.running }}</span
            >
          </button>
          <button
            class="px-5 py-2.5 rounded-full text-sm font-semibold transition-all border shrink-0"
            :class="
              filter === 'draft'
                ? 'bg-gray-100 text-gray-900 border-gray-300'
                : 'bg-white text-gray-500 hover:bg-gray-50 border-gray-200'
            "
            @click="filter = 'draft'"
          >
            Drafts
            <span
              class="ml-1 opacity-60 text-xs bg-gray-300 px-1.5 py-0.5 rounded-full"
              >{{ stats.draft }}</span
            >
          </button>

          <!-- More filters icon -->
          <button
            class="w-10 h-10 rounded-full bg-black text-white flex items-center justify-center ml-2 shrink-0"
          >
            <Filter class="w-4 h-4" />
          </button>

          <!-- PRUNE BUTTON -->
          <button
            @click="handlePrune"
            :disabled="isPruning"
            class="flex items-center gap-2 px-4 py-2 rounded-full border border-red-200 text-red-600 font-bold text-xs hover:bg-red-50 transition-colors ml-4 disabled:opacity-50 disabled:cursor-not-allowed"
            title="Clean up Missing/Untitled records"
          >
            <Trash2 class="w-3 h-3" v-if="!isPruning" />
            <span class="animate-spin h-3 w-3 border-2 border-red-600 border-t-transparent rounded-full" v-else></span>
            {{ isPruning ? 'Cleaning...' : 'Prune Missing' }}
          </button>
        </div>

        <!-- Right Stats (Integrated) -->
        <div
          class="flex items-center gap-8 px-6 text-center hidden md:flex animate-in fade-in duration-700"
        >
          <div>
            <p
              class="text-[10px] text-gray-400 font-bold uppercase tracking-wider mb-1"
            >
              Total Cargo
            </p>
            <p class="text-xl font-medium text-gray-800">{{ stats.total }}</p>
          </div>
          <div>
            <p
              class="text-[10px] text-gray-400 font-bold uppercase tracking-wider mb-1"
            >
              Active
            </p>
            <p class="text-xl font-medium text-green-600">
              {{ stats.running }}
            </p>
          </div>
          <div class="flex gap-2">
            <button
              class="w-10 h-10 rounded-full border border-gray-200 flex items-center justify-center hover:bg-gray-50 hover:border-gray-300 text-gray-500 transition-colors"
            >
              <Link class="w-4 h-4" />
            </button>
          </div>
        </div>
      </div>

      <!-- Bento Grid Layout -->
      <div
        class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4 gap-4 pb-20"
      >
        <!-- Create New Card -->
        <div
          @click="$router.push('/projects/new')"
          class="relative rounded-[2rem] p-6 flex flex-col items-center justify-center bg-[#F3F2EE] h-[220px] border-2 border-transparent hover:border-gray-300 transition-all cursor-pointer group hover:bg-white dashed-border text-center"
        >
          <div
            class="w-12 h-12 rounded-full bg-white flex items-center justify-center shadow-sm mb-4 group-hover:scale-110 transition-transform"
          >
            <Plus class="w-6 h-6 text-gray-400 group-hover:text-black" />
          </div>
          <h3 class="font-bold text-gray-500 group-hover:text-black">
            Create New<br />Program
          </h3>
        </div>

        <!-- Dynamic Project Cards -->
        <template v-if="!loading">
          <ProjectCard
            v-for="p in filteredProjects"
            :key="p.id"
            :project="p"
            @action="handleProjectAction"
          />
        </template>

        <!-- Skeleton Loader -->
        <template v-else>
          <div
            v-for="i in 3"
            :key="i"
            class="rounded-[2rem] bg-gray-100 h-[220px] animate-pulse"
          ></div>
        </template>
      </div>

      <!-- DISCOVERY SECTION (LEGACY CONTAINERS) -->
      <div v-if="legacyApps.length > 0" class="mt-12 animate-in slide-in-from-bottom duration-1000">
        <div class="flex items-center gap-3 mb-6">
          <div class="w-1.5 h-6 bg-blue-500 rounded-full"></div>
          <h2 class="text-xl font-bold text-gray-800">Discovered on Server</h2>
          <span class="px-2 py-0.5 rounded-full bg-blue-50 text-blue-600 text-[10px] font-bold uppercase tracking-wider">Unmanaged</span>
        </div>

        <div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4 gap-4 pb-20">
          <LegacyContainerCard 
            v-for="c in legacyApps" 
            :key="c.id" 
            :container="c" 
            @adopt="navigateToAdopt"
          />
        </div>
      </div>
    </div>
  </DashboardLayout>
</template>

<style scoped>
.dashed-border {
  background-image: url("data:image/svg+xml,%3csvg width='100%25' height='100%25' xmlns='http://www.w3.org/2000/svg'%3e%3crect width='100%25' height='100%25' fill='none' rx='32' ry='32' stroke='%23D1D5DBFF' stroke-width='2' stroke-dasharray='12%2c 12' stroke-dashoffset='0' stroke-linecap='square'/%3e%3c/svg%3e");
}
</style>
<style scoped>
.dashed-border {
  background-image: url("data:image/svg+xml,%3csvg width='100%25' height='100%25' xmlns='http://www.w3.org/2000/svg'%3e%3crect width='100%25' height='100%25' fill='none' rx='32' ry='32' stroke='%23D1D5DBFF' stroke-width='2' stroke-dasharray='12%2c 12' stroke-dashoffset='0' stroke-linecap='square'/%3e%3c/svg%3e");
}
</style>
