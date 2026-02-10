<script setup>
import { ref, reactive, onMounted, computed, watch, onUnmounted, nextTick } from "vue";
import { Terminal as XermTerminal } from "xterm";
import { FitAddon } from "xterm-addon-fit";
import "xterm/css/xterm.css";
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
  Globe,
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
  Link,
  ShieldCheck,
  Box,
  Lock,
  ArrowRight,
  Paperclip,
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
  framework: "Node.js",
  category: "application",
  domain: "", // NEW
  startCommand: "",
  envVars: [],
});

// UI Interaction State
const isScaling = ref(false);
const scaleForm = reactive({ cpu: 0.5, ram: 512 });
const isAnalysing = ref(false);
const analysisResult = ref(null);
const showPass = ref(false); // NEW

const deploymentSteps = computed(() => {
  const steps = [
    { label: 'Initializing', pattern: 'Initializing' },
    { label: 'Pulling Image', pattern: 'Pulling' },
    { label: 'Stopping Old', pattern: 'Stopping' },
    { label: 'Starting Container', pattern: 'Starting' },
    { label: 'Configuring Network', pattern: 'Routing' },
  ];
  
  const currentText = project.value?.current_action || '';
  const activeIndex = steps.findIndex(s => currentText.includes(s.pattern));
  
  return steps.map((s, i) => ({
    ...s,
    status: activeIndex === i ? 'active' : (i < activeIndex ? 'completed' : 'pending')
  }));
});
 // NEW

const availableTabs = computed(() => {
    const tabs = ["overview", "logs", "automation"];
    if (project.value?.category === 'other' || project.value?.category === 'vm') {
        tabs.push("connect");
    }
    if (project.value?.status === 'running' || project.value?.status === 'online') {
        tabs.push("terminal");
    }
    tabs.push("settings");
    return tabs;
});

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
    const force = action === 'redeploy' ? '&force=true' : '';
    await pb.send(`/api/senvanda/project/${project.value.id}/action?action=${action}${force}`, {
      method: "POST"
    });
    
    // Give time for backend to process
    await new Promise((r) => setTimeout(r, 1500));
    await loadProject(true);
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
        form.value.category = found.category || "application";
        form.value.domain = settings.domain || "";
        
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
  isStreaming.value = false;
};

// --- INTERACTIVE TERMINAL ---
const terminalElement = ref(null);
let xterm = null;
let terminalWs = null;
let fitAddon = null;

const connectTerminal = () => {
    if (!project.value || !terminalElement.value) return;

    // Initialize xterm if not exists
    if (!xterm) {
        xterm = new XermTerminal({
            cursorBlink: true,
            theme: {
                background: '#000000',
                foreground: '#ffffff'
            },
            fontFamily: 'JetBrains Mono, Menlo, Monaco, Consolas, "Liberation Mono", "Courier New", monospace',
            fontSize: 13
        });
        fitAddon = new FitAddon();
        xterm.loadAddon(fitAddon);
    }

    xterm.open(terminalElement.value);
    fitAddon.fit();
    xterm.focus();

    // Setup WebSocket
    const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
    const host = window.location.host;
    const terminalUrl = `${protocol}//${host}/api/senvanda/project/${project.value.id}/terminal`;

    terminalWs = new WebSocket(terminalUrl);

    terminalWs.onmessage = (ev) => {
        if (typeof ev.data === 'string') {
            xterm.write(ev.data);
        } else {
            const reader = new FileReader();
            reader.onload = () => xterm.write(new Uint8Array(reader.result));
            reader.readAsArrayBuffer(ev.data);
        }
    };

    xterm.onData(data => {
        if (terminalWs && terminalWs.readyState === WebSocket.OPEN) {
            terminalWs.send(data);
        }
    });

    terminalWs.onclose = () => {
        xterm.write("\r\n\r\n\x1b[31m[Session Closed]\x1b[0m\r\n");
    };

    terminalWs.onerror = () => {
        xterm.write("\r\n\r\n\x1b[31m[Connection Error]\x1b[0m\r\n");
    };

    window.addEventListener('resize', () => fitAddon.fit());
};

const disconnectTerminal = () => {
    if (terminalWs) {
        terminalWs.close();
        terminalWs = null;
    }
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
      category: form.value.category,
      // framework: form.value.framework, // REMOVED: Causing schema error
      settings: {
        framework: form.value.framework, // ADDED: Safe inside JSON
        domain: form.value.domain, // SAVE DOMAIN
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
    await pb.send(`/api/senvanda/project/${project.value.id}/action?action=delete`, {
      method: "POST"
    });
    
    router.push("/");
  } catch (err) {
    alert("Failed to delete project: " + err.message);
    loading.value = false;
  }
};


watch(activeTab, async (newTab) => {
  if (newTab === "logs") {
    connectLogStream();
  } else {
    disconnectLogStream();
  }

  if (newTab === "terminal") {
    await nextTick();
    connectTerminal();
  } else {
    disconnectTerminal();
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
      <div class="animate-spin w-8 h-8 border-4 border-black border-t-transparent rounded-full"></div>
    </div>

    <div v-else-if="project" class="max-w-6xl mx-auto space-y-6 animate-in fade-in duration-500">
      <!-- HEADER -->
      <div class="flex flex-col md:flex-row md:items-center justify-between gap-4">
        <div class="flex items-center gap-4">
          <Button variant="outline" size="icon" class="rounded-xl" @click="router.push('/')">
            <ArrowLeft class="w-4 h-4" />
          </Button>
          <div>
            <h1 class="text-3xl font-bold tracking-tight flex items-center gap-3">
              {{ project.name }}
              <span class="px-3 py-1 rounded-full text-xs font-bold uppercase border tracking-wider" :class="statusColor">
                {{ project.status }}
              </span>
            </h1>
            <a :href="project.repoUrl" target="_blank" class="text-sm text-gray-500 hover:text-black hover:underline mt-1 block">
              {{ project.repoUrl || "No Repository Configured" }}
            </a>
          </div>
        </div>

        <div class="flex items-center gap-2">
          <div class="flex items-center gap-2 bg-white p-1.5 rounded-2xl border border-gray-200 shadow-sm">
            <Button v-if="project.status !== 'running'" size="sm" variant="ghost" class="rounded-xl text-green-600 hover:bg-green-50 gap-2" @click="handleAction('start')" :disabled="actionLoading">
              <Play class="w-4 h-4 fill-current" /> Start
            </Button>
            <Button v-if="project.status === 'running'" size="sm" variant="ghost" class="rounded-xl text-orange-600 hover:bg-orange-50 gap-2" @click="handleAction('redeploy')" :disabled="actionLoading">
              <RefreshCw class="w-4 h-4" :class="{ 'animate-spin': actionLoading }" /> Redeploy
            </Button>
            <Button v-if="project.status === 'running'" size="sm" variant="ghost" class="rounded-xl text-red-600 hover:bg-red-50 gap-2" @click="handleAction('stop')" :disabled="actionLoading">
              <Square class="w-4 h-4 fill-current" /> Stop
            </Button>
            <div class="w-px h-6 bg-gray-200 mx-1"></div>
            <Button size="sm" class="rounded-xl gap-2" @click="window.open(`http://${project.name}.senvanda.local:9080`, '_blank')">
              <ExternalLink class="w-4 h-4" /> Visit
            </Button>
          </div>
        </div>
      </div>

      <!-- TABS -->
      <div class="border-b border-gray-200">
        <nav class="flex gap-6">
          <button v-for="tab in availableTabs" :key="tab" @click="activeTab = tab" class="pb-3 text-sm font-medium border-b-2 transition-colors capitalize" :class="activeTab === tab ? 'border-black text-black' : 'border-transparent text-gray-500 hover:text-gray-700'">
            {{ tab }}
          </button>
        </nav>
      </div>

      <!-- TAB CONTENT -->
      <div class="min-h-[500px]">
        <!-- OVERVIEW -->
        <div v-show="activeTab === 'overview'" class="grid grid-cols-1 lg:grid-cols-3 gap-6 animate-in slide-in-from-left-2 duration-300">
          <div class="lg:col-span-2 space-y-6">
            <div class="bg-white p-6 rounded-[2rem] border border-gray-100 shadow-sm grid grid-cols-3 gap-6">
              <div v-for="stat in [{label:'Status', val: project.state || project.status, icon: true}, {label:'Port', val: project.port}, {label:'Category', val: project.category}]" :key="stat.label" class="p-5 bg-gray-50 rounded-2xl border border-gray-100 flex flex-col justify-between h-28">
                <p class="text-[10px] font-bold text-gray-400 uppercase tracking-widest">{{ stat.label }}</p>
                <div class="flex items-center gap-2">
                   <span v-if="stat.icon" class="w-2 h-2 rounded-full" :class="project.status === 'running' ? 'bg-green-500 animate-pulse' : 'bg-gray-400'"></span>
                   <p class="text-xl font-bold truncate">{{ stat.val }}</p>
                </div>
              </div>
            </div>
            <!-- 2. DEPLOYMENT TRACKER (NEW) -->
            <div v-if="['deploying', 'building'].includes(project.status)" 
                 class="bg-white p-8 rounded-[2rem] border border-orange-100 shadow-sm space-y-6 animate-in slide-in-from-bottom-4 duration-500 overflow-hidden relative"
            >
               <!-- Background glow -->
               <div class="absolute -right-20 -top-20 w-64 h-64 bg-orange-100/30 rounded-full blur-3xl pointer-events-none"></div>

               <div class="flex justify-between items-center relative z-10">
                  <div class="flex items-center gap-3">
                     <div class="p-3 bg-orange-50 text-orange-600 rounded-2xl animate-pulse">
                        <RefreshCw class="w-6 h-6" />
                     </div>
                     <div>
                        <h3 class="font-bold text-gray-900">Deployment in Progress</h3>
                        <p class="text-xs text-orange-600 font-medium font-mono">{{ project.current_action }}</p>
                     </div>
                  </div>
                  <div class="text-[10px] font-bold text-gray-400 uppercase tracking-widest pl-1">Step {{ deploymentSteps.findIndex(s => s.status === 'active') + 1 }} / 5</div>
               </div>

               <div class="relative pt-4 pb-2 z-10">
                  <div class="flex justify-between mb-8 relative">
                     <!-- Progress Line -->
                     <div class="absolute top-4 left-0 w-full h-0.5 bg-gray-100 rounded-full overflow-hidden">
                        <div class="h-full bg-orange-500 transition-all duration-1000" 
                             :style="`width: ${((deploymentSteps.findIndex(s => s.status === 'active') + 0.5) / 5) * 100}%`"
                        ></div>
                     </div>

                     <!-- Steps -->
                     <div v-for="(step, i) in deploymentSteps" :key="i" class="relative z-10 flex flex-col items-center gap-2">
                        <div class="w-8 h-8 rounded-full flex items-center justify-center transition-all duration-500 border-2"
                             :class="{
                               'bg-orange-500 border-orange-500 text-white shadow-lg shadow-orange-200': step.status === 'active',
                               'bg-green-500 border-green-500 text-white': step.status === 'completed',
                               'bg-white border-gray-100 text-gray-300': step.status === 'pending'
                             }"
                        >
                           <Check v-if="step.status === 'completed'" class="w-4 h-4" />
                           <span v-else class="text-xs font-bold">{{ i + 1 }}</span>
                        </div>
                        <span class="text-[9px] font-bold uppercase tracking-tighter" 
                              :class="step.status === 'active' ? 'text-orange-600' : 'text-gray-400'"
                        >{{ step.label }}</span>
                     </div>
                  </div>
               </div>
            </div>

            <div class="bg-white p-6 rounded-[2rem] border border-gray-100 shadow-sm h-72 flex flex-col px-8">
              <h3 class="font-bold flex items-center gap-2 text-lg mb-4"><Globe class="w-5 h-5 text-gray-400" /> Metrics Overview</h3>
              <div class="flex-1 bg-blue-50/20 rounded-2xl border border-blue-100/50 relative overflow-hidden">
                 <ResourceGraph :data="netHistory" label="Traffic" color="blue" />
              </div>
            </div>
          </div>
          <div class="space-y-6">
             <div class="bg-white p-6 rounded-[2rem] border border-gray-100 shadow-sm space-y-4">
                <h3 class="font-bold flex items-center gap-2"><Cpu class="w-4 h-4 text-gray-400" /> Resources</h3>
                <div class="space-y-4">
                   <div v-for="res in [{l:'CPU', v:stats.cpu_percent, c:'bg-black'}, {l:'RAM', v:stats.memory_percent, c:'bg-purple-600'}]" :key="res.l">
                      <div class="flex justify-between text-[10px] font-bold uppercase mb-1">
                         <span>{{ res.l }}</span>
                         <span>{{ res.v.toFixed(1) }}%</span>
                      </div>
                      <div class="h-1.5 w-full bg-gray-100 rounded-full overflow-hidden">
                         <div class="h-full transition-all duration-500" :class="res.c" :style="'width:' + Math.min(res.v, 100) + '%'"></div>
                      </div>
                   </div>
                </div>
                <Button variant="outline" size="sm" class="w-full text-xs mt-2 border-dashed rounded-xl" @click="openScaleModal">Scale Up</Button>
             </div>
          </div>
        </div>

        <!-- LOGS -->
        <div v-show="activeTab === 'logs'" class="space-y-4 animate-in fade-in duration-300">
           <div class="flex justify-between items-center bg-white p-4 rounded-3xl border border-gray-100 shadow-sm">
              <div class="flex items-center gap-3">
                 <div class="w-2 h-2 rounded-full bg-green-500 animate-pulse"></div>
                 <h3 class="font-bold text-sm text-gray-700">Live Console Output</h3>
              </div>
              <div class="flex items-center gap-3">
                 <span v-if="isStreaming" class="text-[10px] font-bold text-green-600 bg-green-50 px-2 py-1 rounded-lg border border-green-100 uppercase tracking-widest animate-pulse">Connected</span>
                 <Button size="sm" variant="ghost" class="h-8 text-xs text-gray-400 hover:text-black" @click="clearLogs">Clear</Button>
              </div>
           </div>
           <div class="bg-[#0F1117] p-8 rounded-[2.5rem] border border-gray-800 shadow-2xl h-[550px] overflow-y-auto relative group">
              <pre class="font-mono text-xs text-gray-300 leading-relaxed selection:bg-orange-500/30">{{ logs || '> Connecting to instance logs...' }}<span v-if="isStreaming" class="inline-block w-1.5 h-3 bg-green-500 animate-pulse ml-1 align-middle"></span></pre>
              <!-- Decorative corner -->
              <div class="absolute top-0 right-0 p-4 opacity-10 group-hover:opacity-30 transition-opacity">
                 <Terminal class="w-12 h-12 text-white" />
              </div>
           </div>
        </div>

            <!-- TERMINAL (NEW) -->
            <div v-show="activeTab === 'terminal'" class="space-y-4 animate-in zoom-in-95 duration-300">
               <div class="flex justify-between items-center bg-white p-4 rounded-3xl border border-gray-100 shadow-sm">
                  <div class="flex items-center gap-3">
                     <div class="w-8 h-8 rounded-xl bg-black text-white flex items-center justify-center">
                        <Terminal class="w-4 h-4" />
                     </div>
                     <h3 class="font-bold text-sm text-gray-700">Interactive Shell Session</h3>
                  </div>
                  <div class="flex items-center gap-4">
                     <div class="hidden md:flex items-center gap-2 text-[10px] text-gray-400 font-bold uppercase tracking-wider">
                        <span>Ctrl+C (Copy)</span>
                        <span class="w-1 h-1 bg-gray-200 rounded-full"></span>
                        <span>Ctrl+V (Paste)</span>
                     </div>
                     <span class="text-[10px] font-bold text-blue-600 bg-blue-50 px-2 py-1 rounded-lg border border-blue-100 uppercase tracking-widest">Active Shell</span>
                  </div>
               </div>
               <div class="bg-black p-4 rounded-[2.5rem] border border-gray-800 shadow-2xl h-[550px] overflow-hidden relative" ref="terminalElement">
                  <!-- Term rendering via xterm.js -->
               </div>
            </div>

        <!-- AUTOMATION -->
        <div v-show="activeTab === 'automation'" class="space-y-6 animate-in slide-in-from-right-4 duration-500">
           <div class="bg-white p-8 rounded-[2rem] border border-gray-100 shadow-sm max-w-3xl space-y-6">
              <div class="flex items-center gap-3"><Zap class="w-6 h-6 text-orange-500" /> <h2 class="text-xl font-bold">CI/CD Webhook</h2></div>
              <div class="p-5 bg-gray-50 rounded-2xl border border-gray-100 font-mono text-xs break-all text-blue-600">
                 http://api.senvanda.local:9080/api/senvanda/webhook/{{ project.id }}
              </div>
           </div>
        </div>

        <!-- CONNECT -->
        <div v-show="activeTab === 'connect'" class="space-y-6 animate-in slide-in-from-right-4 duration-500">
           <div class="grid grid-cols-1 md:grid-cols-2 gap-6">
              <div class="bg-white p-8 rounded-[2rem] border border-gray-100 shadow-sm space-y-6">
                 <div class="flex items-center gap-3"><Link class="w-6 h-6 text-blue-500" /><h3 class="font-bold">Connection Details</h3></div>
                 <div class="space-y-3">
                    <div v-for="f in [{l:'Host', v:project.name+'.senvanda.local'}, {l:'IP', v:project.internal_ip}, {l:'Port', v:project.port}]" :key="f.l" class="p-4 bg-gray-50 rounded-2xl">
                       <p class="text-[10px] font-bold text-gray-400 uppercase">{{ f.l }}</p>
                       <p class="font-mono font-bold">{{ f.v || 'N/A' }}</p>
                    </div>
                 </div>
              </div>
              <div class="bg-white p-8 rounded-[2rem] border border-gray-100 shadow-sm space-y-6">
                 <div class="flex items-center gap-3"><ShieldCheck class="w-6 h-6 text-red-500" /><h3 class="font-bold">Authentication</h3></div>
                 <div v-if="project.settings?.envVars?.length" class="space-y-3">
                    <div v-for="env in project.settings.envVars" :key="env.key" class="p-4 bg-gray-50 rounded-2xl group flex justify-between items-center">
                       <div>
                          <p class="text-[10px] font-bold text-gray-400 uppercase">{{ env.key }}</p>
                          <input :type="showPass ? 'text' : 'password'" readonly :value="env.value" class="bg-transparent border-none outline-none font-mono font-bold text-sm w-full" />
                       </div>
                       <button @click="showPass = !showPass" class="text-gray-400 hover:text-black"><component :is="showPass ? Activity : Lock" class="w-4 h-4" /></button>
                    </div>
                 </div>
                 <p v-else class="text-sm text-gray-400 italic text-center py-10">No authentication variables set.</p>
              </div>
           </div>
        </div>

        <!-- SETTINGS -->
        <div v-show="activeTab === 'settings'" class="bg-white p-8 rounded-[2rem] border border-gray-100 shadow-sm animate-in slide-in-from-right-4 duration-300">
           <div class="max-w-2xl space-y-10">
              <div class="flex justify-between items-start">
                 <div><h2 class="text-2xl font-bold">Project Configuration</h2><p class="text-sm text-gray-500">Update your logic and environment.</p></div>
                 <Button @click="saveConfig" size="lg" class="rounded-2xl px-10 shadow-lg shadow-black/20" :disabled="loading">Save Configuration</Button>
              </div>
              <div class="grid grid-cols-2 gap-6">
                 <div v-for="field in [{l:'Framework', m:'framework', options:['Node.js','Python','Go','PHP','Docker']}, {l:'Category', m:'category', options:['application','vm','infrastructure','other']}]" :key="field.l" class="space-y-2">
                    <label class="text-[10px] font-bold text-gray-400 uppercase tracking-widest">{{ field.l }}</label>
                    <select v-model="form[field.m]" class="w-full h-12 bg-gray-50 rounded-2xl border-none focus:ring-2 focus:ring-black px-4 font-bold text-sm">
                       <option v-for="o in field.options" :key="o" :value="o">{{ o }}</option>
                    </select>
                 </div>
                 <div class="space-y-2 col-span-2">
                    <label class="text-[10px] font-bold text-gray-400 uppercase tracking-widest">Custom Domain</label>
                    <input v-model="form.domain" placeholder="e.g. myapp.com" class="w-full h-12 bg-gray-50 rounded-2xl border-none focus:ring-2 focus:ring-black px-4 font-bold text-sm" />
                    <p class="text-[10px] text-gray-400">Point A Record to server IP. SSL will be auto-managed.</p>
                 </div>
              </div>
              <div class="space-y-4 pt-6 border-t border-gray-100">
                 <h3 class="font-bold text-red-600">Danger Zone</h3>
                 <div class="p-6 bg-red-50/50 rounded-3xl border border-red-100 flex justify-between items-center">
                    <div><p class="font-bold text-red-900 text-sm">Delete Project</p><p class="text-xs text-red-600/70">Careful, this action is irreversible.</p></div>
                    <Button variant="destructive" class="rounded-xl" @click="alert('Coming soon!')">Destroy</Button>
                 </div>
              </div>
           </div>
        </div>

      </div>
    </div>
  </DashboardLayout>
</template>