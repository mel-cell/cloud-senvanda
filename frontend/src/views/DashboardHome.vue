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
const activeTab = ref("application"); // Default tab
const filter = ref("all");
const loading = ref(true);
const error = ref(null);

const loadProjects = async () => {
  loading.value = true;
  error.value = null;
  try {
    const [projectRes] = await Promise.all([
      pb.send("/api/senvanda/deploy/projects", { method: "GET" }),
    ]);

    // Trust the Backend Source of Truth
    projects.value = projectRes || [];

  } catch (err) {
    console.error("Failed to load apps", err);
    error.value = "Engine connection refused. Is the backend running?";
  } finally {
    loading.value = false;
  }
};

const getCategoryCount = (cat) => {
    return projects.value.filter(p => (p.category || 'discovered') === cat).length;
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
  let items = projects.value;

  // 1. Filter by Category (Tab)
  items = items.filter(p => (p.category || 'discovered') === activeTab.value);

  // 2. Filter by Status (Pill)
  if (filter.value !== "all") {
    items = items.filter((p) => p.status === filter.value);
  }
  
  return items;
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
  if (
    !confirm(
      "Hapus semua record project yang tidak ada kontainernya? (Termasuk record 'Untitled')",
    )
  )
    return;

  isPruning.value = true;
  try {
    const res = await pb.send("/api/senvanda/deploy/prune", { method: "POST" });
    alert(
      `DevOps Hygiene: Berhasil membersihkan ${res.pruned_count} project hantu/untitled!`,
    );
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

const subscribeProjects = async () => {
  // Subscribe to real-time changes
  pb.collection("projects").subscribe("*", (e) => {
    console.log(
      "🔔 Real-time update:",
      e.action,
      e.record.name,
      e.record.status,
    );

    if (e.action === "update") {
      const index = projects.value.findIndex((p) => p.id === e.record.id);
      if (index !== -1) {
        // Sync the DB status back to our local model
        // We keep the other calculated fields (like State) unless we want to refetch
        projects.value[index] = {
          ...projects.value[index],
          status: e.record.status,
          db_status: e.record.status,
          internal_ip: e.record.internal_ip,
          url: e.record.url,
        };
      }
    } else if (e.action === "create") {
      loadProjects(); // Easier to just reload all on create to get calculated fields
    } else if (e.action === "delete") {
      projects.value = projects.value.filter((p) => p.id !== e.record.id);
    }
  });
};

onMounted(() => {
  loadProjects();
  subscribeProjects();
});

import { onUnmounted } from "vue";
onUnmounted(() => {
  pb.collection("projects").unsubscribe("*");
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
      <div class="flex flex-col gap-6">
        <!-- CATEGORY TABS -->
        <div class="flex items-center gap-1 bg-gray-100/50 p-1 rounded-full w-fit border border-gray-200">
             <button 
                v-for="tab in ['application', 'infrastructure']" 
                :key="tab"
                @click="activeTab = tab"
                class="px-6 py-2 rounded-full text-sm font-bold transition-all capitalize flex items-center gap-2"
                :class="activeTab === tab ? 'bg-white shadow text-black' : 'text-gray-400 hover:text-gray-600'"
             >
                {{ tab }}
                <span v-if="getCategoryCount(tab) > 0" class="bg-gray-200 text-[10px] px-1.5 py-0.5 rounded-full min-w-[20px]">{{ getCategoryCount(tab) }}</span>
             </button>
        </div>

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
            </button>
 
            <!-- PRUNE BUTTON -->
            <button
              @click="handlePrune"
              :disabled="isPruning"
              class="flex items-center gap-2 px-4 py-2 rounded-full border border-red-200 text-red-600 font-bold text-xs hover:bg-red-50 transition-colors ml-4 disabled:opacity-50 disabled:cursor-not-allowed"
              title="Clean up Missing/Untitled records"
            >
              <Trash2 class="w-3 h-3" v-if="!isPruning" />
              <span
                class="animate-spin h-3 w-3 border-2 border-red-600 border-t-transparent rounded-full"
                v-else
              ></span>
              {{ isPruning ? "Cleaning..." : "Prune Missing" }}
            </button>
          </div>
        </div>
      </div>

      <!-- Bento Grid Layout -->
      <div
        class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4 gap-4 pb-20"
      >
        <!-- Create New Card (Only in Application Tab) -->
        <div
          v-if="activeTab === 'application'"
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
        
        <!-- Empty State -->
        <div v-if="!loading && filteredProjects.length === 0 && activeTab !== 'application'" class="col-span-full py-10 text-center text-gray-400 italic">
            No items found in {{ activeTab }}.
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
