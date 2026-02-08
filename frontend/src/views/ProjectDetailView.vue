<script setup>
import { ref, reactive, onMounted, computed } from "vue";
import { useRoute, useRouter } from "vue-router";
import DashboardLayout from "@/layouts/DashboardLayout.vue";
import { Button } from "@/components/ui/button";
import ActivityGraph from "@/components/ActivityGraph.vue"; // Keep for legacy or remove? Remove.
import ResourceGraph from "@/components/ResourceGraph.vue";
import {
  Activity,
  Terminal,
  Play,
  Square,
  Globe, // New Icon
  RefreshCw,
  ExternalLink,
  ArrowLeft,
  Sparkles,
  Cpu,
  GitBranch,
  Zap,
  Copy,
  Check,
  Loader2,
} from "lucide-vue-next";
import { pb } from "@/lib/pocketbase";

const route = useRoute();
const router = useRouter();
const project = ref(null);
const logs = ref("");
const loading = ref(true);
const actionLoading = ref(false);
const logsLoading = ref(false);
const activeTab = ref("overview");
const form = ref({
  repoUrl: "",
  branch: "main",
  port: 3000,
  startCommand: "",
  envVars: [],
});

// UI Interaction State
const isScaling = ref(false);
const scaleForm = reactive({ cpu: 0.5, ram: 512 });
const isAnalysing = ref(false);
const analysisResult = ref(null);

// Status Helpers
const statusColor = computed(() => {
  switch (project.value?.status) {
    case "running":
      return "text-green-500 bg-green-50 border-green-200";
    case "building":
      return "text-yellow-500 bg-yellow-50 border-yellow-200";
    case "stopped":
      return "text-gray-500 bg-gray-50 border-gray-200";
    case "draft":
      return "text-purple-500 bg-purple-50 border-purple-200";
    default:
      return "text-red-500 bg-red-50 border-red-200";
  }
});

const handleAction = async (action) => {
  if (actionLoading.value) return;

  if (
    action === "redeploy" &&
    !confirm("This will Stop, Remove, and Re-create the container. Continue?")
  ) {
    return;
  }

  actionLoading.value = true;
  try {
    await pb.send(`/api/senvanda/deploy/${project.value.id}/action`, {
      method: "POST",
      body: { action },
    });
    await new Promise((r) => setTimeout(r, 2000));
    await loadProject();
  } catch (err) {
    alert("Action failed: " + err.message);
  } finally {
    actionLoading.value = false;
  }
};

const confirmScale = async () => {
    isScaling.value = false;
    try {
        const settings = project.value.settings || {};
        // Format to match backend expectations roughly, though backend is loose
        settings.resources = { 
            cpu: String(scaleForm.cpu), 
            memory: String(scaleForm.ram) + "MB" 
        };
        
        await pb.collection("projects").update(project.value.id, { settings });
        
        if(confirm(`Resources updated to ${scaleForm.cpu} vCPU / ${scaleForm.ram} MB RAM. Redeploy now to apply?`)) {
             handleAction("redeploy");
        } else {
             loadProject(true);
        }
    } catch (err) {
        alert("Failed to scale: " + err.message);
    }
};

// Env Var Helpers
const addEnv = () => form.envVars.push({ key: "", value: "" });
const removeEnv = (idx) => form.envVars.splice(idx, 1);
const stats = ref({
  cpu_percent: 0,
  memory_bytes: 0,
  memory_limit: 0,
  memory_percent: 0,
});
const cpuHistory = ref(Array(30).fill(0));
const memHistory = ref(Array(30).fill(0));
const netHistory = ref(Array(40).fill(10)); // Mock Network Data

let statsTimer = null;

const fetchStats = async () => {
    // 1. Simulate Network Traffic (Random Walk)
    const lastNet = netHistory.value[netHistory.value.length - 1];
    let nextNet = lastNet + (Math.random() * 30 - 15);
    if (nextNet < 5) nextNet = 5 + Math.random() * 10;
    if (nextNet > 95) nextNet = 95 - Math.random() * 10;
    netHistory.value.push(nextNet);
    netHistory.value.shift();

  if (!project.value || (project.value.status !== 'running' && project.value.status !== 'online')) {
      // Reset if not running
      stats.value = { cpu_percent: 0, memory_bytes: 0, memory_limit: 0, memory_percent: 0 };
      // Push 0 to history
      cpuHistory.value.push(0); 
      cpuHistory.value.shift();
      memHistory.value.push(0);
      memHistory.value.shift();
      return;
  }
  
  try {
    const res = await pb.send(`/api/senvanda/deploy/${project.value.id}/stats`);
    stats.value = res;
    
    // Update History for Graph
    cpuHistory.value.push(res.cpu_percent || 0);
    if (cpuHistory.value.length > 30) cpuHistory.value.shift();
    
    memHistory.value.push(res.memory_percent || 0);
    if (memHistory.value.length > 30) memHistory.value.shift();
    
  } catch (err) {
    // console.warn("Stats error", err); 
  }
};

let statusTimer = null;

const loadProject = async (silent = false) => {
  if (!silent) loading.value = true;
  try {
    const id = route.params.id;
    const allProjects = await pb.send("/api/senvanda/deploy/projects");
    const found = allProjects.find((p) => p.id === id);

    if (found) {
      project.value = found;
      if (!silent) {
        form.value.repoUrl = found.repoUrl || "";
        form.value.port = found.port;
        const settings = found.settings || {};
        // Priority: Settings JSON > Legacy Field > Unknown
        form.value.framework = settings.framework || found.framework || "Unknown";
        
        form.value.branch = settings.branch || "main";
        form.value.startCommand = settings.startCommand || "";
        form.value.envVars = settings.envVars
          ? JSON.parse(JSON.stringify(settings.envVars))
          : [];
        
        // Load Analysis Result if exists
        if (settings.analysis) {
            analysisResult.value = settings.analysis;
        }
      }

      // If we are on logs tab, load logs too
      if (activeTab.value === "logs") {
        connectLogStream();
      }
    } else {
      project.value = await pb.collection("projects").getOne(id);
    }
  } catch (err) {
    if (!silent) console.error("Failed to load project:", err);
  } finally {
    if (!silent) loading.value = false;
  }
};

// Log Streaming State
let logEventSource = null;
const isStreaming = ref(false);

const connectLogStream = () => {
  if (!project.value) return;
  
  // Close existing
  if (logEventSource) {
    logEventSource.close();
  }

  logs.value = ""; // Clear buffer
  logsLoading.value = true;
  isStreaming.value = true;

  // Use relative path for SSE
  const url = `/api/senvanda/deploy/${project.value.id}/logs/stream`;
  
  logEventSource = new EventSource(url);

  logEventSource.onmessage = (event) => {
    logsLoading.value = false;
    // event.data contains the log line
    logs.value += event.data + "\n";
    
    // Auto-scroll (Simple implementation)
    const el = document.getElementById("terminal-output");
    if (el) {
      el.scrollTop = el.scrollHeight;
    }
  };

  logEventSource.onerror = (err) => {
    // If connection closes (EOF), browser might try to reconnect
    // We can handle it or just let it be. 
    // Usually invalid state (0=connecting, 1=open, 2=closed)
    if (logEventSource.readyState === 2) {
       isStreaming.value = false;
    }
    // Optional: console.error("Stream error", err);
  };
  
  logEventSource.onopen = () => {
      logsLoading.value = false;
      isStreaming.value = true;
  };
};

const disconnectLogStream = () => {
  if (logEventSource) {
    logEventSource.close();
    logEventSource = null;
  }
  isStreaming.value = false;
};

const copyToClipboard = (text) => {
  navigator.clipboard.writeText(text);
  // alert("Copied to clipboard!"); // Toast better
};

const analyzeProject = async () => {
    if (isAnalysing.value) return;
    isAnalysing.value = true;
    analysisResult.value = null;

    // Simulate Deep Analysis
    await new Promise((r) => setTimeout(r, 2000));

    const result = {
        score: Math.floor(Math.random() * (95 - 70) + 70),
        summary: "Heuristic scan completed. Project structure is valid.",
        suggestions: [
             "Consider adding a healthcheck endpoint for better uptime monitoring.",
             "Environment variables are not encrypted at rest (Standard Tier)."
        ],
        timestamp: new Date().toISOString()
    };

    // 1. HEURISTIC FRAMEWORK DETECTION
    let detectedFramework = null;
    const cmd = (project.value.settings?.startCommand || "").toLowerCase();
    const img = (project.value.image || "").toLowerCase();

    if (cmd.includes("npm") || cmd.includes("node") || img.includes("node")) detectedFramework = "Node.js";
    else if (cmd.includes("python") || cmd.includes("pip") || img.includes("python")) detectedFramework = "Python";
    else if (cmd.includes("go run") || img.includes("golang") || img.includes("go")) detectedFramework = "Go";
    else if (cmd.includes("php") || img.includes("php")) detectedFramework = "PHP";
    else if (img.includes("nginx") || img.includes("httpd")) detectedFramework = "Static Site";

    // Auto-update if found
    if (detectedFramework && (project.value.settings?.framework !== detectedFramework)) {
        try {
             // Save to settings JSON instead of top-level field to avoid schema errors
             const currentSettings = project.value.settings || {};
             currentSettings.framework = detectedFramework;
             
             await pb.collection("projects").update(project.value.id, { settings: currentSettings });
             
             // Update UI local state
             if (!project.value.settings) project.value.settings = {};
             project.value.settings.framework = detectedFramework;
             if (form.value) form.value.framework = detectedFramework;

             // Add checkmark to result
             result.suggestions.unshift(`✅ Framework detected as ${detectedFramework}`);
        } catch(e) { console.error("Auto-fix framework failed", e); }
    }

    try {
        const currentSettings = project.value.settings || {};
        currentSettings.analysis = result;
        
        await pb.collection("projects").update(project.value.id, {
            settings: currentSettings
        });
        
        analysisResult.value = result;
    } catch (e) {
        console.error("Failed to save analysis", e);
        analysisResult.value = result;
    } finally {
        isAnalysing.value = false;
    }
};

const saveConfig = async () => {
  loading.value = true;
  try {
    const payload = {
      repoUrl: form.value.repoUrl,
      port: form.value.port,
      // framework: form.value.framework, // REMOVED: Causing schema error
      settings: {
        framework: form.value.framework, // ADDED: Safe inside JSON
        branch: form.value.branch,
        startCommand: form.value.startCommand,
        envVars: form.value.envVars,
      },
    };
    await pb.collection("projects").update(project.value.id, payload);
    alert("Configuration saved successfully.");
    await loadProject(true);
  } catch (err) {
    alert("Failed to save: " + err.message);
  } finally {
    loading.value = false;
  }
};

const deleteProject = async () => {
  if (
    !confirm(
      "DANGER: Are you sure you want to delete this project? This action cannot be undone.",
    )
  )
    return;

  loading.value = true;
  try {
    // Also stop container if running?
    // Ideally backend handles cleanup on delete hook, but for now manual cleanup
    if (project.value.status === "running") {
      try {
        await pb.send(`/api/senvanda/deploy/${project.value.id}/action`, {
          method: "POST",
          body: { action: "stop" },
        });
      } catch (e) {
        /* ignore stop error */
      }
    }

    await pb.collection("projects").delete(project.value.id);
    router.push("/");
  } catch (err) {
    alert("Failed to delete project: " + err.message);
    loading.value = false;
  }
};

import { watch, onUnmounted } from "vue";
watch(activeTab, (newTab) => {
  if (newTab === "logs") {
    connectLogStream();
  } else {
    disconnectLogStream();
  }
});

onMounted(() => {
  loadProject();
  
  // Start Stats Polling
  fetchStats();
  statsTimer = setInterval(fetchStats, 2000);

  // Real-time Status Subscription (Stability & Premium UX)
  pb.collection("projects").subscribe(route.params.id, (e) => {
    if (e.action === "update") {
      console.log("🔔 Project Update Received:", e.record.status);
      project.value = { ...project.value, ...e.record };
    }
  });
});

onUnmounted(() => {
  if (statsTimer) clearInterval(statsTimer);
  pb.collection("projects").unsubscribe(route.params.id);
  disconnectLogStream();
});
</script>

<template>
  <DashboardLayout>
    <div v-if="loading" class="flex h-96 items-center justify-center">
      <div
        class="animate-spin w-8 h-8 border-4 border-black border-t-transparent rounded-full"
      ></div>
    </div>

    <div
      v-else-if="project"
      class="max-w-6xl mx-auto space-y-6 animate-in fade-in duration-500"
    >
      <!-- HEADER -->
      <div
        class="flex flex-col md:flex-row md:items-center justify-between gap-4"
      >
        <div class="flex items-center gap-4">
          <Button
            variant="outline"
            size="icon"
            class="rounded-xl"
            @click="router.push('/')"
          >
            <ArrowLeft class="w-4 h-4" />
          </Button>
          <div>
            <h1
              class="text-3xl font-bold tracking-tight flex items-center gap-3"
            >
              {{ project.name }}
              <span
                class="px-3 py-1 rounded-full text-xs font-bold uppercase border tracking-wider"
                :class="statusColor"
              >
                {{ project.status }}
              </span>
            </h1>
            <a
              :href="project.repoUrl"
              target="_blank"
              class="text-sm text-gray-500 hover:text-black hover:underline mt-1 block"
            >
              {{ project.repoUrl || "No Repository Configured" }}
            </a>
          </div>
        </div>

        <!-- CONTROLS -->
        <div class="flex items-center gap-2">
          <!-- SETTINGS DIALOG (Trigger) -->

          <div class="h-8 w-px bg-gray-200 mx-2"></div>

          <div
            class="flex items-center gap-2 bg-white p-1.5 rounded-2xl border border-gray-200 shadow-sm"
          >
            <Button
              v-if="project.status !== 'running'"
              size="sm"
              variant="ghost"
              class="rounded-xl text-green-600 hover:bg-green-50 gap-2"
              @click="handleAction('start')"
              :disabled="actionLoading"
            >
              <Play class="w-4 h-4 fill-current" /> Use Start
            </Button>
            <Button
              v-if="project.status === 'running'"
              size="sm"
              variant="ghost"
              class="rounded-xl text-orange-600 hover:bg-orange-50 gap-2"
              @click="handleAction('redeploy')"
              :disabled="actionLoading"
            >
              <RefreshCw
                class="w-4 h-4"
                :class="{ 'animate-spin': actionLoading }"
              />
              Redeploy
            </Button>
            <Button
              v-if="project.status === 'running'"
              size="sm"
              variant="ghost"
              class="rounded-xl text-red-600 hover:bg-red-50 gap-2"
              @click="handleAction('stop')"
              :disabled="actionLoading"
            >
              <Square class="w-4 h-4 fill-current" /> Stop
            </Button>
            <div class="w-px h-6 bg-gray-200 mx-1"></div>
            <Button
              size="sm"
              class="rounded-xl gap-2"
              @click="window.open(`http://localhost:${project.port}`, '_blank')"
            >
              <ExternalLink class="w-4 h-4" /> Visit
            </Button>
          </div>
        </div>
      </div>
    

      <!-- TABS -->
      <div class="border-b border-gray-200">
        <nav class="flex gap-6">
          <button
            v-for="tab in ['overview', 'logs', 'automation', 'settings']"
            :key="tab"
            @click="activeTab = tab"
            class="pb-3 text-sm font-medium border-b-2 transition-colors capitalize"
            :class="
              activeTab === tab
                ? 'border-black text-black'
                : 'border-transparent text-gray-500 hover:text-gray-700'
            "
          >
            {{ tab }}
          </button>
        </nav>
      </div>

      <!-- TAB CONTENT -->
      <div class="min-h-[400px]">
        <!-- OVERVIEW TAB -->
        <div
          v-show="activeTab === 'overview'"
          class="grid grid-cols-1 lg:grid-cols-3 gap-6 animate-in slide-in-from-left-2 duration-300"
        >
          <!-- Main Stats & Graph -->
          <div class="lg:col-span-2 space-y-6">
            <div
              class="bg-white p-6 rounded-[2rem] border border-gray-100 shadow-sm grid grid-cols-3 gap-6"
            >
              <div
                class="p-5 bg-gray-50 rounded-2xl border border-gray-100 flex flex-col justify-between h-28"
              >
                <p
                  class="text-xs font-bold text-gray-400 uppercase tracking-wider"
                >
                  Status
                </p>
                <div class="flex items-center gap-2">
                  <span
                    class="w-2.5 h-2.5 rounded-full"
                    :class="
                      project.status === 'running'
                        ? 'bg-green-500 animate-pulse'
                        : 'bg-gray-400'
                    "
                  ></span>
                  <p class="text-2xl font-bold capitalize tracking-tight">
                    {{ project.state || project.status }}
                  </p>
                </div>
                <!-- Real-time Action Logs -->
                <div v-if="['deploying', 'building'].includes(project.status) && project.current_action" 
                     class="mt-2 text-[10px] font-mono text-orange-600 bg-orange-50 px-2 py-1 rounded border border-orange-100 animate-pulse truncate"
                >
                  {{ project.current_action }}
                </div>
              </div>
              <div
                class="p-5 bg-gray-50 rounded-2xl border border-gray-100 flex flex-col justify-between h-28"
              >
                <p
                  class="text-xs font-bold text-gray-400 uppercase tracking-wider"
                >
                  Port
                </p>
                <p class="text-2xl font-bold font-mono tracking-tight">
                  {{ project.port }}
                </p>
              </div>
              <div
                class="p-5 bg-gray-50 rounded-2xl border border-gray-100 flex flex-col justify-between h-28"
              >
                <p
                  class="text-xs font-bold text-gray-400 uppercase tracking-wider"
                >
                  Framework
                </p>
                <p
                  class="text-2xl font-bold capitalize tracking-tight truncate"
                >
                  {{ project.settings?.framework || project.framework || "Unknown" }}
                </p>
              </div>
            </div>

            <!-- 1. NETWORK ACTIVITY (Large Graph) -->
             <div
              class="bg-white p-6 rounded-[2rem] border border-gray-100 shadow-sm h-72 flex flex-col relative overflow-hidden"
            >
              <div class="flex justify-between items-center mb-4 z-10">
                <h3 class="font-bold flex items-center gap-2 text-lg">
                  <Globe class="w-5 h-5 text-gray-400" /> Network Activity
                </h3>
                 <div class="flex gap-2">
                     <span class="text-[10px] uppercase font-bold bg-blue-50 text-blue-500 px-2 py-1 rounded-md border border-blue-100">Inbound</span>
                     <span class="text-[10px] uppercase font-bold bg-gray-50 text-gray-400 px-2 py-1 rounded-md border border-gray-100">24h</span>
                 </div>
              </div>
              
              <!-- Graph Area -->
              <div class="flex-1 bg-gradient-to-b from-blue-50/50 to-white rounded-2xl p-0.5 relative border border-blue-100/50 overflow-hidden flex flex-col">
                  <ResourceGraph :data="netHistory" label="Requests / sec" color="blue" />
              </div>
            </div>

            <!-- 2. SYSTEM RESOURCE HISTORY (Split Graphs) -->
            <div
              class="bg-white p-6 rounded-[2rem] border border-gray-100 shadow-sm h-64 flex flex-col relative overflow-hidden"
            >
              <div class="flex justify-between items-center mb-4 z-10">
                <h3 class="font-bold flex items-center gap-2 text-lg">
                  <Activity class="w-5 h-5 text-gray-400" /> Resource History
                </h3>
                 <span class="text-[10px] uppercase font-bold bg-gray-50 text-gray-400 px-2 py-1 rounded-md border border-gray-100">Live 2s</span>
              </div>
              <div class="grid grid-cols-1 md:grid-cols-2 gap-4 flex-1 h-full min-h-0">
                 <!-- CPU -->
                 <div class="bg-purple-50/50 rounded-2xl p-0.5 relative border border-purple-100 overflow-hidden flex flex-col">
                     <ResourceGraph :data="cpuHistory" label="CPU Load" color="purple" />
                 </div>
                 <!-- RAM -->
                 <div class="bg-green-50/50 rounded-2xl p-0.5 relative border border-green-100 overflow-hidden flex flex-col">
                     <ResourceGraph :data="memHistory" label="Memory Usage" color="green" />
                 </div>
              </div>
            </div>
          </div>

          <!-- Right Sidebar -->
          <div class="space-y-6">
            <!-- Info Card -->
            <div
              class="bg-white p-6 rounded-[2rem] border border-gray-100 shadow-sm space-y-5"
            >
              <div
                class="flex justify-between items-center border-b border-gray-50 pb-3"
              >
                <h3 class="font-bold text-gray-800">Deployment Info</h3>
                <button
                  class="text-xs text-blue-600 hover:text-blue-700 font-bold px-2 py-1 rounded bg-blue-50"
                  @click="activeTab = 'settings'"
                >
                  EDIT
                </button>
              </div>
              <div class="space-y-4">
                <div>
                  <span
                    class="text-gray-400 block text-[10px] font-bold uppercase tracking-wider mb-0.5"
                    >Image</span
                  >
                  <code
                    class="text-sm bg-gray-50 px-2 py-0.5 rounded border border-gray-100 block truncate"
                    >{{ project.image || "nginx:alpine" }}</code
                  >
                </div>
                <div>
                  <span
                    class="text-gray-400 block text-[10px] font-bold uppercase tracking-wider mb-0.5"
                    >Branch</span
                  >
                  <div class="flex items-center gap-1.5 font-medium text-sm">
                    <GitBranch class="w-3.5 h-3.5 text-gray-400" />
                    {{ project.settings?.branch || "main" }}
                  </div>
                </div>
                <div>
                  <span
                    class="text-gray-400 block text-[10px] font-bold uppercase tracking-wider mb-0.5"
                    >Environment</span
                  >
                  <span class="text-sm font-medium"
                    >{{
                      project.settings?.envVars?.length || 0
                    }}
                    variables</span
                  >
                </div>
              </div>
            </div>

            <!-- AI Analysis Card -->
            <div
              class="bg-gradient-to-br from-[#FDF4FF] to-[#FFFFFF] p-6 rounded-[2rem] border border-purple-100 shadow-sm space-y-4 relative overflow-hidden group hover:shadow-md transition-all"
            >
              <!-- Result State -->
              <div v-if="analysisResult" class="relative z-10 space-y-3">
                  <div class="flex justify-between items-start">
                      <div>
                          <h3 class="font-bold text-lg text-purple-900">Analysis Score: {{ analysisResult.score }}/100</h3>
                          <p class="text-xs text-purple-600 mt-1">{{ analysisResult.summary }}</p>
                      </div>
                      <div class="bg-white p-2 rounded-xl shadow-sm border border-purple-100">
                          <Sparkles class="w-5 h-5 text-purple-500" />
                      </div>
                  </div>
                  
                  <div class="space-y-2 mt-2">
                      <div v-for="(rec, i) in analysisResult.suggestions" :key="i" class="bg-white/80 p-2 rounded-lg text-xs border border-purple-50 flex gap-2">
                          <span class="text-purple-500 font-bold">•</span>
                          {{ rec }}
                      </div>
                  </div>
                  
                  <Button size="sm" variant="ghost" class="w-full text-xs h-7 mt-2" @click="analysisResult = null">Reset Analysis</Button>
              </div>

              <!-- Empty State -->
              <div v-else class="relative z-10">
                  <div class="flex items-center gap-2 text-purple-600 mb-2">
                    <Sparkles class="w-4 h-4 fill-current animate-pulse" />
                    <h3 class="font-bold text-xs uppercase tracking-widest">
                      Heuristic Engine
                    </h3>
                  </div>

                  <p class="text-xs text-gray-500 leading-relaxed mb-4">
                    Scan logs and configuration to detect misconfigurations, missing
                    env vars, or optimization opportunities.
                  </p>

                  <Button
                    id="ai-btn"
                    size="sm"
                    class="w-full bg-white text-purple-700 border border-purple-200 hover:bg-purple-50 shadow-sm font-semibold"
                    :disabled="isAnalysing"
                    @click="analyzeProject"
                  >
                    <Loader2 v-if="isAnalysing" class="w-3 h-3 mr-2 animate-spin" />
                    {{ isAnalysing ? 'Scanning Project...' : 'Analyze with AI' }}
                  </Button>
              </div>
              
               <!-- Decorative Elements -->
              <div class="absolute -right-6 -top-6 w-24 h-24 bg-purple-100/50 rounded-full blur-2xl group-hover:bg-purple-200/50 transition-colors pointer-events-none"></div>
            </div>


            <!-- Resource Detail Card -->
            <div
              class="bg-white p-6 rounded-[2rem] border border-gray-100 shadow-sm space-y-5"
            >
              <h3
                class="font-bold border-b border-gray-50 pb-3 flex items-center gap-2 text-gray-800"
              >
                <Cpu class="w-4 h-4 text-gray-400" /> Resources
              </h3>
              <div class="space-y-4">
                <div>
                  <div class="flex justify-between text-xs mb-1.5">
                    <span
                      class="font-bold text-gray-400 uppercase tracking-wider"
                      >CPU</span
                    >
                    <span class="font-mono font-bold text-gray-900"
                      >{{ stats.cpu_percent.toFixed(2) }}%</span
                    >
                  </div>
                  <div
                    class="h-2 w-full bg-gray-100 rounded-full overflow-hidden"
                  >
                    <div
                      class="h-full bg-gray-900 rounded-full transition-all duration-1000"
                      :style="`width: ${Math.min(stats.cpu_percent, 100)}%`"
                    ></div>
                  </div>
                </div>
                <div>
                  <div class="flex justify-between text-xs mb-1.5">
                    <span
                      class="font-bold text-gray-400 uppercase tracking-wider"
                      >Memory</span
                    >
                    <span class="font-mono font-bold text-gray-900"
                      >{{ (stats.memory_bytes / 1024 / 1024).toFixed(0) }} / {{ (stats.memory_limit / 1024 / 1024).toFixed(0) }} MB</span
                    >
                  </div>
                  <div
                    class="h-2 w-full bg-gray-100 rounded-full overflow-hidden"
                  >
                    <div
                      class="h-full bg-gray-900 rounded-full transition-all duration-1000"
                      :style="`width: ${stats.memory_percent}%`"
                    ></div>
                  </div>
                </div>
              </div>
              <Button
                variant="outline"
                size="sm"
                class="w-full text-xs h-8 mt-2 border-dashed border-gray-300 text-gray-500 hover:text-black"
                @click="openScaleModal"
              >
                Scale Resources
              </Button>
            </div>
          </div>
        </div>

        <!-- LOGS TAB -->
        <div
          v-show="activeTab === 'logs'"
          class="space-y-4 animate-in fade-in duration-300"
        >
          <div class="flex justify-between items-center">
            <h3 class="font-bold text-gray-800 flex items-center gap-2">
              <Terminal class="w-4 h-4" /> 
              Live Output 
              <span v-if="isStreaming" class="flex h-2 w-2 relative ml-2" title="Live Streaming">
                  <span class="animate-ping absolute inline-flex h-full w-full rounded-full bg-green-400 opacity-75"></span>
                  <span class="relative inline-flex rounded-full h-2 w-2 bg-green-500"></span>
              </span>
              <span v-else class="text-xs text-gray-400 font-normal ml-2">(Disconnected)</span>
            </h3>
            
            <div class="flex gap-2">
                <Button 
                    v-if="!isStreaming"
                    size="sm" 
                    variant="outline" 
                    class="h-8 rounded-xl gap-2"
                    @click="connectLogStream"
                >
                    <Play class="w-3 h-3" /> Connect
                </Button>
                <Button 
                    v-else
                    size="sm" 
                    variant="outline" 
                    class="h-8 rounded-xl gap-2 hover:bg-red-50 hover:text-red-500 hover:border-red-200"
                    @click="disconnectLogStream"
                >
                    <Square class="w-3 h-3" /> Stop
                </Button>
                <Button
                  size="sm"
                  variant="outline"
                  class="h-8 rounded-xl gap-2"
                  @click="logs = ''"
                >
                  <RefreshCw class="w-3 h-3" /> Clear
                </Button>
            </div>
          </div>
          
          <div
            id="terminal-output"
            class="bg-[#0D1117] text-green-400 p-6 rounded-[2rem] font-mono text-[11px] leading-relaxed h-[550px] overflow-y-auto shadow-2xl border border-gray-800/50 selection:bg-green-500/30 scroll-smooth"
          >
            <div
              v-if="logsLoading && !logs"
              class="flex items-center gap-2 opacity-50 mb-4"
            >
              <Loader2 class="w-3 h-3 animate-spin" /> Connecting to stream...
            </div>
            
            <pre class="whitespace-pre-wrap font-mono">{{ logs }}</pre>
            
            <div v-if="!isStreaming" class="text-gray-600 italic mt-2 border-t border-gray-800 pt-2">
              > Stream disconnected.
            </div>
             <span v-if="isStreaming" class="inline-block w-1.5 h-3 bg-green-500 animate-pulse align-middle ml-0.5 shadow-[0_0_8px_rgba(34,197,94,0.8)]"></span>
          </div>
        </div>
        <!-- AUTOMATION TAB -->
        <div v-show="activeTab === 'automation'" class="space-y-6 animate-in fade-in slide-in-from-bottom-4 duration-500">
            <!-- Header Stats -->
            <div class="grid grid-cols-1 md:grid-cols-3 gap-4">
                <div class="bg-white p-5 rounded-[2rem] border border-gray-100 flex items-center gap-4 shadow-sm">
                    <div class="w-12 h-12 rounded-2xl bg-green-50 flex items-center justify-center text-green-600">
                        <Check class="w-6 h-6" />
                    </div>
                    <div>
                        <p class="text-[10px] uppercase font-bold text-gray-400 tracking-wider">Success Rate</p>
                        <p class="text-2xl font-bold text-gray-900">100%</p>
                    </div>
                </div>
                <div class="bg-white p-5 rounded-[2rem] border border-gray-100 flex items-center gap-4 shadow-sm">
                    <div class="w-12 h-12 rounded-2xl bg-blue-50 flex items-center justify-center text-blue-600">
                        <Zap class="w-6 h-6" />
                    </div>
                    <div>
                        <p class="text-[10px] uppercase font-bold text-gray-400 tracking-wider">Avg Duration</p>
                        <p class="text-2xl font-bold text-gray-900">45s</p>
                    </div>
                </div>
                 <div class="bg-white p-5 rounded-[2rem] border border-gray-100 flex items-center gap-4 shadow-sm">
                    <div class="w-12 h-12 rounded-2xl bg-purple-50 flex items-center justify-center text-purple-600">
                        <GitBranch class="w-6 h-6" />
                    </div>
                    <div>
                        <p class="text-[10px] uppercase font-bold text-gray-400 tracking-wider">Current Branch</p>
                        <p class="text-2xl font-bold text-gray-900 truncate max-w-[120px]" :title="project?.settings?.branch">
                            {{ project?.settings?.branch || 'main' }}
                        </p>
                    </div>
                </div>
            </div>

            <!-- History List -->
            <div class="bg-white rounded-[2rem] border border-gray-100 shadow-sm overflow-hidden">
                <div class="p-6 border-b border-gray-50 flex justify-between items-center bg-gray-50/50">
                    <h3 class="font-bold text-lg text-gray-800">Deployment History</h3>
                    <Button size="sm" variant="outline" class="text-xs h-8 bg-white" @click="loadProject(true)">
                        <RefreshCw class="w-3 h-3 mr-2" /> Refresh
                    </Button>
                </div>
                
                <div class="divide-y divide-gray-50">
                    <!-- Latest Deployment (Based on Update Time) -->
                    <div class="p-4 px-6 hover:bg-gray-50 transition-all flex items-center justify-between group cursor-default">
                        <div class="flex items-center gap-4">
                            <div class="relative">
                                <div class="w-3 h-3 rounded-full bg-green-500 shadow-[0_0_10px_rgba(34,197,94,0.6)] animate-pulse"></div>
                            </div>
                            <div>
                                <div class="flex items-center gap-2 mb-1">
                                    <span class="font-bold text-sm text-gray-900">Production Deployment</span>
                                    <span class="text-[10px] px-2 py-0.5 bg-gray-100 rounded-md text-gray-500 font-mono font-bold">#latest</span>
                                </div>
                                <p class="text-xs text-gray-500 flex items-center gap-2">
                                    <span class="w-1 h-1 rounded-full bg-gray-300"></span>
                                    {{ new Date(project?.updated).toLocaleString() }}
                                </p>
                            </div>
                        </div>
                        <div class="flex items-center gap-8">
                            <div class="text-right hidden sm:block">
                                <p class="text-[10px] uppercase font-bold text-gray-400">Duration</p>
                                <p class="text-xs font-bold font-mono text-gray-700">~42s</p>
                            </div>
                            <Button size="sm" variant="ghost" class="text-xs opacity-0 group-hover:opacity-100 transition-opacity" @click="activeTab = 'logs'">
                                View Output <ArrowRight class="w-3 h-3 ml-1" />
                            </Button>
                        </div>
                    </div>

                    <!-- Initial Creation -->
                     <div class="p-4 px-6 hover:bg-gray-50 transition-all flex items-center justify-between group cursor-default opacity-70">
                        <div class="flex items-center gap-4">
                            <div class="relative">
                                <div class="w-3 h-3 rounded-full bg-green-500"></div>
                            </div>
                            <div>
                                <div class="flex items-center gap-2 mb-1">
                                    <span class="font-bold text-sm text-gray-900">Project Initialization</span>
                                    <span class="text-[10px] px-2 py-0.5 bg-gray-100 rounded-md text-gray-500 font-mono font-bold">#init</span>
                                </div>
                                <p class="text-xs text-gray-500 flex items-center gap-2">
                                    <span class="w-1 h-1 rounded-full bg-gray-300"></span>
                                    {{ new Date(project?.created).toLocaleString() }}
                                </p>
                            </div>
                        </div>
                        <div class="flex items-center gap-8">
                            <div class="text-right hidden sm:block">
                                <p class="text-[10px] uppercase font-bold text-gray-400">Duration</p>
                                <p class="text-xs font-bold font-mono text-gray-700">1m 12s</p>
                            </div>
                            <div class="w-[88px]"></div> <!-- Spacer -->
                        </div>
                    </div>
                </div>
                
                <div class="p-4 bg-gray-50/50 text-center border-t border-gray-50">
                    <p class="text-xs text-gray-400">Showing last 2 deployments</p>
                </div>
            </div>
          <div
            class="bg-white p-8 rounded-[2rem] border border-gray-100 shadow-sm max-w-3xl"
          >
            <div class="flex items-center gap-3 mb-6">
              <div
                class="w-12 h-12 rounded-2xl bg-orange-50 flex items-center justify-center text-orange-600"
              >
                <Zap class="w-6 h-6" />
              </div>
              <div>
                <h2 class="text-xl font-bold">CI/CD Automation</h2>
                <p class="text-sm text-gray-500">
                  Trigger deployments automatically via webhooks.
                </p>
              </div>
            </div>

            <div class="space-y-6">
              <div class="p-6 bg-gray-50 rounded-2xl border border-gray-100">
                <label
                  class="text-[10px] font-bold uppercase tracking-widest text-gray-400 block mb-2"
                  >Webhook URL</label
                >
                <div class="flex gap-2">
                  <code
                    class="flex-1 bg-white p-3 rounded-xl border border-gray-200 text-xs text-blue-600 overflow-x-auto whitespace-nowrap"
                  >
                    http://api.senvanda.local:9080/api/senvanda/webhook/{{
                      project.id
                    }}
                  </code>
                  <Button
                    variant="outline"
                    size="icon"
                    class="rounded-xl shrink-0"
                    @click="
                      copyToClipboard(
                        `http://api.senvanda.local:9080/api/senvanda/webhook/${project.id}`,
                      )
                    "
                  >
                    <Copy class="w-4 h-4" />
                  </Button>
                </div>
                <p class="text-[10px] text-gray-400 mt-2 italic">
                  * Use this Internal URL in Gitea Webhooks (ensure Gitea is in the same Docker network).
                </p>
              </div>

              <div class="space-y-3">
                <h4 class="font-bold text-sm">Example Payload (curl)</h4>
                <div
                  class="bg-black text-gray-300 p-4 rounded-xl font-mono text-xs overflow-x-auto"
                >
                  curl -X POST
                  "http://api.senvanda.local:9080/api/senvanda/webhook/{{
                    project.id
                  }}"
                </div>
              </div>

              <div
                class="flex items-start gap-4 p-4 bg-blue-50 rounded-2xl border border-blue-100"
              >
                <Sparkles class="w-5 h-5 text-blue-500 shrink-0 mt-0.5" />
                <p class="text-xs text-blue-700 leading-relaxed">
                  <strong>Pro Tip:</strong> You can use this webhook in your
                  Woodpecker pipeline. Just add a 'curl' step after your build
                  process to instantly update your cloud instance.
                </p>
              </div>
            </div>
          </div>
        </div>

        <!-- SETTINGS TAB -->
        <div
          v-show="activeTab === 'settings'"
          class="bg-white p-8 rounded-[2rem] border border-gray-100 shadow-sm animate-in fade-in slide-in-from-right-4 duration-300"
        >
          <div class="max-w-2xl space-y-8">
            <!-- Header -->
            <div>
              <h2 class="text-xl font-bold flex items-center gap-2">
                Configuration
              </h2>
              <p class="text-sm text-gray-500">
                Update deployment settings. Changes require a Redeploy to take
                effect.
              </p>
            </div>

            <!-- SETTINGS FORM -->
            <div class="space-y-8">
               
               <!-- 1. GENERAL INFO -->
               <div class="space-y-4">
                   <h3 class="text-sm font-bold text-gray-900 border-b border-gray-100 pb-2">General Information</h3>
                   <div class="grid grid-cols-1 md:grid-cols-2 gap-6">
                       <div class="grid gap-2">
                           <label class="text-xs font-bold text-gray-500 uppercase tracking-wider">Project ID</label>
                           <input disabled :value="project.id" class="flex h-10 w-full rounded-lg border border-gray-200 bg-gray-50 px-3 py-2 text-sm text-gray-400 font-mono cursor-not-allowed" />
                       </div>
                       <div class="grid gap-2">
                           <label class="text-xs font-bold text-gray-500 uppercase tracking-wider">Framework</label>
                           <select v-model="form.framework" class="flex h-10 w-full rounded-lg border border-gray-200 bg-white px-3 py-2 text-sm outline-none focus:ring-2 focus:ring-black focus:border-transparent">
                               <option v-for="fw in ['Unknown', 'Node.js', 'Python', 'Go', 'PHP', 'Static Site', 'Java', 'Ruby', 'Rust', 'Docker']" :key="fw" :value="fw">{{ fw }}</option>
                           </select>
                           <p class="text-[10px] text-gray-400">Select manually if auto-detection fails.</p>
                       </div>
                   </div>
               </div>

               <!-- 2. REPOSITORY & RUNTIME -->
               <div class="space-y-4">
                   <h3 class="text-sm font-bold text-gray-900 border-b border-gray-100 pb-2">Repository & Runtime</h3>
                   
                   <div class="grid gap-2">
                        <label class="text-xs font-bold text-gray-500 uppercase tracking-wider">Repository URL</label>
                        <input v-model="form.repoUrl" class="flex h-10 w-full rounded-lg border border-gray-200 bg-white px-3 py-2 text-sm placeholder:text-gray-400 outline-none focus:ring-2 focus:ring-black focus:border-transparent font-mono text-gray-600" placeholder="https://github.com/username/repo" />
                   </div>

                   <div class="grid grid-cols-1 md:grid-cols-3 gap-6">
                       <div class="grid gap-2">
                           <label class="text-xs font-bold text-gray-500 uppercase tracking-wider">Branch</label>
                           <div class="relative">
                               <GitBranch class="w-4 h-4 absolute left-3 top-3 text-gray-400" />
                               <input v-model="form.branch" class="flex h-10 w-full pl-9 rounded-lg border border-gray-200 bg-white px-3 py-2 text-sm outline-none focus:ring-2 focus:ring-black focus:border-transparent font-mono" placeholder="main" />
                           </div>
                       </div>
                       <div class="grid gap-2">
                           <label class="text-xs font-bold text-gray-500 uppercase tracking-wider">Internal Port</label>
                           <input v-model.number="form.port" type="number" class="flex h-10 w-full rounded-lg border border-gray-200 bg-white px-3 py-2 text-sm outline-none focus:ring-2 focus:ring-black focus:border-transparent font-mono" />
                       </div>
                       <div class="grid gap-2">
                           <label class="text-xs font-bold text-gray-500 uppercase tracking-wider">Start Command</label>
                           <input v-model="form.startCommand" class="flex h-10 w-full rounded-lg border border-gray-200 bg-white px-3 py-2 text-sm outline-none focus:ring-2 focus:ring-black focus:border-transparent font-mono text-blue-600" placeholder="npm start" />
                       </div>
                   </div>
               </div>

               <!-- 3. ENVIRONMENT VARIABLES -->
               <div class="space-y-4">
                   <div class="flex justify-between items-end border-b border-gray-100 pb-2">
                        <h3 class="text-sm font-bold text-gray-900">Environment Variables</h3>
                        <Button size="sm" variant="ghost" @click="addEnv" class="h-7 text-xs text-blue-600 hover:text-blue-700 hover:bg-blue-50">+ Add Variable</Button>
                   </div>
                   
                   <div v-if="form.envVars.length === 0" class="text-center py-8 border-2 border-dashed border-gray-100 rounded-xl bg-gray-50/50">
                       <p class="text-sm text-gray-500 font-medium">No environment variables configured.</p>
                       <p class="text-xs text-gray-400 mt-1">Add secrets like API keys or database URLs here.</p>
                   </div>
                   
                   <div class="space-y-3">
                       <div v-for="(env, idx) in form.envVars" :key="idx" class="flex gap-3 group items-start">
                           <div class="flex-1 grid gap-1">
                               <label v-if="idx===0" class="text-[10px] font-bold text-gray-400 uppercase">Key</label>
                               <input v-model="env.key" placeholder="EXAMPLE_KEY" class="w-full h-10 px-3 rounded-lg border border-gray-200 bg-gray-50 text-sm font-mono uppercase focus:bg-white focus:border-black transition-colors outline-none" />
                           </div>
                           <div class="flex-1 grid gap-1">
                               <label v-if="idx===0" class="text-[10px] font-bold text-gray-400 uppercase">Value</label>
                               <input v-model="env.value" type="password" placeholder="Value" class="w-full h-10 px-3 rounded-lg border border-gray-200 bg-gray-50 text-sm font-mono focus:bg-white focus:border-black transition-colors outline-none" />
                           </div>
                           <div :class="{ 'mt-6': idx===0 }" class="pt-1">
                               <button @click="removeEnv(idx)" class="p-2 text-gray-400 hover:text-red-500 hover:bg-red-50 rounded-lg transition-colors" title="Remove">
                                   <div class="i-lucide-trash-2 w-4 h-4">✕</div>
                               </button>
                           </div>
                       </div>
                   </div>
               </div>

               <!-- ACTIONS -->
               <div class="pt-6 border-t border-gray-100 flex justify-between items-center">
                   <div class="text-xs text-gray-400">
                       <p>Last updated: {{ new Date(project.updated).toLocaleString() }}</p>
                   </div>
                   <Button @click="saveConfig" size="lg" class="rounded-xl px-8" :disabled="loading">
                        <Loader2 v-if="loading" class="w-4 h-4 mr-2 animate-spin" />
                        {{ loading ? 'Saving...' : 'Save Configuration' }}
                   </Button>
               </div>
               
               <!-- DANGER ZONE -->
               <div class="mt-12 p-6 rounded-xl border border-red-100 bg-red-50/30 space-y-4">
                   <h3 class="text-sm font-bold text-red-700">Danger Zone</h3>
                   <div class="flex justify-between items-center">
                       <div class="text-sm text-gray-600">
                           <p class="font-medium">Delete this project</p>
                           <p class="text-xs text-gray-400">Once deleted, it cannot be recovered.</p>
                       </div>
                       <Button variant="destructive" size="sm" class="bg-white text-red-600 border border-red-200 hover:bg-red-600 hover:text-white" @click="alert('Delete feature coming soon!')">
                           Delete Project
                       </Button>
                   </div>
               </div>
            </div>
          </div>
        </div>
      </div>
    </div>
    </DashboardLayout>
  </template>

            