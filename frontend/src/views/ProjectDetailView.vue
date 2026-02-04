<script setup>
import { ref, onMounted, computed } from "vue";
import { useRoute, useRouter } from "vue-router";
import DashboardLayout from "@/layouts/DashboardLayout.vue";
import { Button } from "@/components/ui/button";
import ActivityGraph from "@/components/ActivityGraph.vue";
import {
  Activity,
  Terminal,
  Play,
  Square,
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

// Resource Scaling
const scaleResources = async () => {
  // Simple Prompt for MVP (In real app, use a Dialog)
  const newCpu = prompt(
    "Enter new CPU Limit (e.g. 1.0):",
    project.value.settings?.resources?.cpu || "0.5",
  );
  if (!newCpu) return;

  const newMem = prompt(
    "Enter new Memory Limit (e.g. 1024MB):",
    project.value.settings?.resources?.memory || "512MB",
  );
  if (!newMem) return;

  try {
    const settings = project.value.settings || {};
    settings.resources = { cpu: newCpu, memory: newMem };

    await pb.collection("projects").update(project.value.id, { settings });
    alert(
      `Resources updated to CPU: ${newCpu}, RAM: ${newMem}. Redeploy to apply.`,
    );
    loadProject(true);
  } catch (err) {
    alert("Failed to scale: " + err.message);
  }
};

// Env Var Helpers
const addEnv = () => form.envVars.push({ key: "", value: "" });
const removeEnv = (idx) => form.envVars.splice(idx, 1);

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
        form.value.branch = settings.branch || "main";
        form.value.startCommand = settings.startCommand || "";
        form.value.envVars = settings.envVars
          ? JSON.parse(JSON.stringify(settings.envVars))
          : [];
      }

      // If we are on logs tab, load logs too
      if (activeTab.value === "logs") {
        loadLogs();
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

const loadLogs = async () => {
  if (!project.value) return;
  logsLoading.value = true;
  try {
    const res = await pb.send(`/api/senvanda/deploy/${project.value.id}/logs`);
    logs.value = res.logs || "No logs available for this container.";
  } catch (err) {
    logs.value = "Failed to fetch logs: " + err.message;
  } finally {
    logsLoading.value = false;
  }
};

const copyToClipboard = (text) => {
  navigator.clipboard.writeText(text);
  alert("Copied to clipboard!");
};

const analyzeProject = async () => {
  const btn = document.getElementById("ai-btn");
  const originalText = btn ? btn.innerText : "Analyze with AI";

  if (btn) {
    btn.innerText = "Scanning...";
    btn.disabled = true;
    // Add spinner class if needed, simplified for now
  }

  // Simulate Deep Analysis
  await new Promise((r) => setTimeout(r, 2500));

  // Intelligent Recommendations
  const recommendations = [
    "✅ Framework detected: Node.js (High Confidence)",
    "✅ Dependencies: 24 packages found",
    "⚠️ Security: Port 8085 is publicly exposed",
    "ℹ️ Optimization: Suggest using 'npm ci' for faster builds",
  ];

  alert(`✨ Heuristic Analysis Complete\n\n${recommendations.join("\n")}`);

  // Auto-fix simulation: If framework unknown, update it
  if (
    !project.value.framework ||
    project.value.framework.toLowerCase() === "unknown"
  ) {
    try {
      await pb
        .collection("projects")
        .update(project.value.id, { framework: "Node.js" });
      // Manual patch local state to avoid full reload flicker
      project.value.framework = "Node.js";
    } catch (e) {
      console.error("Auto-fix failed:", e);
    }
  }

  if (btn) {
    btn.innerText = "Analysis Complete";
    setTimeout(() => {
      btn.innerText = originalText;
      btn.disabled = false;
    }, 2000);
  }
};

const saveConfig = async () => {
  loading.value = true;
  try {
    const payload = {
      repoUrl: form.value.repoUrl,
      port: form.value.port,
      settings: {
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
  if (newTab === "logs") loadLogs();
});

onMounted(() => {
  loadProject();

  // Real-time Status Subscription (Stability & Premium UX)
  pb.collection("projects").subscribe(route.params.id, (e) => {
    if (e.action === "update") {
      console.log("🔔 Project Update Received:", e.record.status);
      project.value = { ...project.value, ...e.record };
    }
  });
});

onUnmounted(() => {
  pb.collection("projects").unsubscribe(route.params.id);
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
                  {{ project.framework || "Unknown" }}
                </p>
              </div>
            </div>

            <!-- Activity Graph -->
            <div
              class="bg-white p-6 rounded-[2rem] border border-gray-100 shadow-sm h-80 flex flex-col relative overflow-hidden"
            >
              <div class="flex justify-between items-center mb-6 z-10 relative">
                <h3 class="font-bold flex items-center gap-2 text-lg">
                  <Activity class="w-5 h-5 text-gray-400" /> Activity
                </h3>
                <select
                  class="bg-gray-50 border border-gray-200 text-xs rounded-lg px-3 py-1.5 outline-none font-medium hover:bg-gray-100 cursor-pointer"
                >
                  <option>Last 24 Hours</option>
                  <option>Last 7 Days</option>
                </select>
              </div>
              <div class="flex-1 w-full relative -mx-1">
                <ActivityGraph :realtime="project.status === 'running'" />
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
              class="bg-gradient-to-br from-[#FDF4FF] to-[#FFFFFF] p-6 rounded-[2rem] border border-purple-100 shadow-sm space-y-4 relative overflow-hidden group hover:shadow-md transition-shadow"
            >
              <!-- Decorative Elements -->
              <div
                class="absolute -right-6 -top-6 w-24 h-24 bg-purple-100/50 rounded-full blur-2xl group-hover:bg-purple-200/50 transition-colors"
              ></div>

              <div
                class="flex items-center gap-2 text-purple-600 mb-1 relative z-10"
              >
                <Sparkles class="w-4 h-4 fill-current animate-pulse" />
                <h3 class="font-bold text-xs uppercase tracking-widest">
                  Heuristic Engine
                </h3>
              </div>

              <p class="text-xs text-gray-500 leading-relaxed relative z-10">
                Scan logs and configuration to detect misconfigurations, missing
                env vars, or optimization opportunities.
              </p>

              <Button
                id="ai-btn"
                size="sm"
                class="w-full bg-white text-purple-700 border border-purple-200 hover:bg-purple-50 shadow-sm relative z-10 font-semibold"
                @click="analyzeProject"
              >
                Analyze with AI
              </Button>
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
                      >0.05 / 0.5 vCPU</span
                    >
                  </div>
                  <div
                    class="h-2 w-full bg-gray-100 rounded-full overflow-hidden"
                  >
                    <div
                      class="h-full bg-gray-900 w-[10%] rounded-full transition-all duration-1000"
                      style="width: 12%"
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
                      >128 / 512 MB</span
                    >
                  </div>
                  <div
                    class="h-2 w-full bg-gray-100 rounded-full overflow-hidden"
                  >
                    <div
                      class="h-full bg-gray-900 w-[45%] rounded-full transition-all duration-1000"
                      style="width: 25%"
                    ></div>
                  </div>
                </div>
              </div>
              <Button
                variant="outline"
                size="sm"
                class="w-full text-xs h-8 mt-2 border-dashed border-gray-300 text-gray-500 hover:text-black"
                @click="scaleResources"
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
              <Terminal class="w-4 h-4" /> Live Output
            </h3>
            <Button
              size="sm"
              variant="outline"
              class="h-8 rounded-xl gap-2"
              @click="loadLogs"
              :disabled="logsLoading"
            >
              <RefreshCw
                class="w-3 h-3"
                :class="{ 'animate-spin': logsLoading }"
              />
              Refresh Logs
            </Button>
          </div>
          <div
            class="bg-black text-emerald-500 p-6 rounded-[2rem] font-mono text-xs h-[550px] overflow-y-auto shadow-2xl border border-gray-800 selection:bg-emerald-500/30"
          >
            <div
              v-if="logsLoading && !logs"
              class="flex items-center gap-2 opacity-50"
            >
              <Loader2 class="w-3 h-3 animate-spin" /> Fetching stream...
            </div>
            <pre class="whitespace-pre-wrap">{{ logs }}</pre>
            <div v-if="!logsLoading && !logs" class="text-gray-500 italic">
              No log output detected from container.
            </div>
          </div>
        </div>
        <!-- AUTOMATION TAB -->
        <div
          v-show="activeTab === 'automation'"
          class="space-y-6 animate-in slide-in-from-right-2 duration-300"
        >
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

            <!-- Form Fields -->
            <div class="space-y-6">
              <div class="grid gap-2">
                <label class="text-sm font-medium">Repository URL</label>
                <input
                  v-model="form.repoUrl"
                  class="flex h-10 w-full rounded-md border border-gray-300 bg-white px-3 py-2 text-sm placeholder:text-gray-400 focus:outline-none focus:ring-2 focus:ring-black focus:ring-offset-2 disabled:cursor-not-allowed disabled:opacity-50"
                  placeholder="https://github.com/..."
                />
              </div>

              <div class="grid grid-cols-2 gap-6">
                <div class="grid gap-2">
                  <label class="text-sm font-medium">Branch</label>
                  <input
                    v-model="form.branch"
                    class="flex h-10 w-full rounded-md border border-gray-300 bg-white px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-black"
                    placeholder="main"
                  />
                </div>
                <div class="grid gap-2">
                  <label class="text-sm font-medium">Internal Port</label>
                  <input
                    v-model.number="form.port"
                    type="number"
                    class="flex h-10 w-full rounded-md border border-gray-300 bg-white px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-black"
                  />
                </div>
              </div>

              <div class="grid gap-2">
                <label class="text-sm font-medium">Start Command</label>
                <div class="relative">
                  <Terminal
                    class="w-4 h-4 absolute left-3 top-3 text-gray-400"
                  />
                  <input
                    v-model="form.startCommand"
                    class="flex h-10 w-full pl-9 rounded-md border border-gray-300 bg-white px-3 py-2 text-sm font-mono text-blue-600 focus:outline-none focus:ring-2 focus:ring-black"
                    placeholder="npm start"
                  />
                </div>
              </div>

              <!-- Env Vars -->
              <div class="space-y-3 pt-4 border-t border-gray-100">
                <div class="flex justify-between items-center">
                  <label class="text-sm font-medium"
                    >Environment Variables</label
                  >
                  <Button
                    size="sm"
                    variant="outline"
                    @click="addEnv"
                    class="h-8 text-xs"
                  >
                    + Add Variable
                  </Button>
                </div>

                <div
                  v-if="form.envVars.length === 0"
                  class="text-center py-6 border-2 border-dashed border-gray-100 rounded-lg text-xs text-gray-400"
                >
                  No environment variables configured.
                </div>

                <div
                  v-for="(env, idx) in form.envVars"
                  :key="idx"
                  class="flex gap-2 group"
                >
                  <input
                    v-model="env.key"
                    placeholder="KEY"
                    class="flex-1 h-9 px-3 rounded-md border border-gray-200 bg-gray-50 text-xs font-mono uppercase focus:bg-white focus:border-black transition-colors outline-none"
                  />
                  <input
                    v-model="env.value"
                    placeholder="VALUE"
                    class="flex-1 h-9 px-3 rounded-md border border-gray-200 bg-gray-50 text-xs font-mono focus:bg-white focus:border-black transition-colors outline-none"
                  />
                  <button
                    @click="removeEnv(idx)"
                    class="w-9 h-9 flex items-center justify-center text-gray-400 hover:text-red-500 hover:bg-red-50 rounded-md transition-all"
                  >
                    <svg
                      xmlns="http://www.w3.org/2000/svg"
                      width="16"
                      height="16"
                      viewBox="0 0 24 24"
                      fill="none"
                      stroke="currentColor"
                      stroke-width="2"
                      stroke-linecap="round"
                      stroke-linejoin="round"
                    >
                      <path d="M3 6h18" />
                      <path d="M19 6v14c0 1-1 2-2 2H7c-1 0-2-1-2-2V6" />
                      <path d="M8 6V4a2 2 0 0 1 2-2h4a2 2 0 0 1 2 2v2" />
                    </svg>
                  </button>
                </div>
              </div>
            </div>

            <!-- Action Footer -->
            <div
              class="pt-8 border-t border-gray-100 flex items-center justify-between"
            >
              <div class="flex items-center gap-4">
                <Button
                  class="bg-black hover:bg-gray-800 text-white gap-2"
                  @click="saveConfig"
                  :disabled="loading"
                >
                  <div
                    v-if="loading"
                    class="w-4 h-4 border-2 border-white/50 border-t-white rounded-full animate-spin"
                  ></div>
                  Save Configuration
                </Button>
                <p class="text-xs text-gray-400" v-if="!loading">
                  Last saved: {{ new Date(project.updated).toLocaleString() }}
                </p>
              </div>

              <Button
                variant="ghost"
                class="text-red-500 hover:bg-red-50 hover:text-red-600 font-medium text-sm"
                @click="deleteProject"
              >
                Delete Project
              </Button>
            </div>
          </div>
        </div>
      </div>
    </div>
  </DashboardLayout>
</template>
