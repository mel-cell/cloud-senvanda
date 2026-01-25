<script setup>
import { ref, watch, reactive } from "vue";
import {
  Rocket,
  Save,
  Settings,
  Globe,
  Terminal,
  Server,
  Cpu,
  Database,
  Plus,
  Trash2,
  Box,
  GitBranch,
  Loader2,
} from "lucide-vue-next";
import { Button } from "@/components/ui/button";
import { Label } from "@/components/ui/label";

const props = defineProps({
  config: {
    type: Object,
    required: true,
  },
  isDeploying: Boolean,
});

const emit = defineEmits(["deploy", "draft", "cancel"]);

// Active Tab State
const activeTab = ref("general"); // general, build, env, resources, devops

// Form State (Local copy)
const form = reactive({
  name: props.config.name || "",
  framework: props.config.framework || "Node.js",
  port: props.config.port || 80,
  startCommand: props.config.startCommand || "",
  branch: props.config.branch || "main",
  domain: props.config.domain || `${props.config.name}.senvanda.local`,
  envVars: props.config.envVars && props.config.envVars.length ? [...props.config.envVars] : [{ key: "", value: "" }],
  resources: {
    cpu: "0.5",
    memory: "512MB"
  }
});

// Watch changes and sync back to parent prop
watch(
  form,
  (newVal) => {
    props.config.name = newVal.name;
    props.config.framework = newVal.framework;
    props.config.startCommand = newVal.startCommand;
    props.config.port = Number(newVal.port);
    props.config.branch = newVal.branch;
    props.config.domain = newVal.domain;
    props.config.envVars = newVal.envVars.filter((e) => e.key);
    props.config.resources = newVal.resources;
  },
  { deep: true },
);

// Env Var Logic
const addEnv = () => form.envVars.push({ key: "", value: "" });
const removeEnv = (index) => form.envVars.splice(index, 1);
</script>

<template>
  <div class="animate-in fade-in slide-in-from-bottom-4 duration-500 max-w-3xl mx-auto">
    <!-- DETECTION BANNER -->
    <div class="mb-6 bg-emerald-50 border border-emerald-100 p-4 rounded-2xl flex items-start gap-4 shadow-sm animate-in zoom-in-95 duration-700 delay-150">
      <div class="w-10 h-10 bg-emerald-100 rounded-full flex items-center justify-center shrink-0">
        <Rocket class="w-5 h-5 text-emerald-600" />
      </div>
      <div>
        <h3 class="font-bold text-emerald-900 text-sm">AI Analysis Complete</h3>
        <p class="text-emerald-700 text-xs mt-1 leading-relaxed">
          We detected a <strong>{{ props.config.framework || "Generic" }}</strong> project structure. 
          We've auto-configured the build settings for you. Please review below.
        </p>
      </div>
    </div>

    <div class="bg-white rounded-[2rem] border border-gray-100 shadow-xl overflow-hidden flex flex-col md:flex-row">
      <!-- LEFT SIDEBAR (Tabs) -->
      <div class="w-full md:w-64 bg-gray-50 border-r border-gray-100 p-6 flex flex-col gap-2">
        <div class="mb-6">
          <h2 class="text-xl font-bold flex items-center gap-2 font-[Outfit]">
            <Settings class="w-5 h-5" /> Config
          </h2>
          <p class="text-xs text-gray-500 mt-1">Fine-tune your deployment.</p>
        </div>

        <button
          v-for="tab in ['general', 'build', 'env', 'resources', 'devops']"
          :key="tab"
          @click="activeTab = tab"
          class="px-4 py-3 rounded-xl text-sm font-medium text-left transition-all flex items-center gap-3"
          :class="activeTab === tab ? 'bg-white shadow-sm text-black border border-gray-200' : 'text-gray-500 hover:bg-gray-100/50 hover:text-gray-900'"
        >
          <Globe v-if="tab === 'general'" class="w-4 h-4" />
          <Box v-if="tab === 'build'" class="w-4 h-4" />
          <Database v-if="tab === 'env'" class="w-4 h-4" />
          <Cpu v-if="tab === 'resources'" class="w-4 h-4" />
          <Rocket v-if="tab === 'devops'" class="w-4 h-4" />
          <span class="capitalize">{{ tab === 'devops' ? 'DevOps Analysis' : tab }}</span>
        </button>
      </div>

      <!-- RIGHT CONTENT (Forms) -->
      <div class="flex-1 p-8 min-h-[500px] flex flex-col">
        <!-- TAB: GENERAL -->
        <div v-show="activeTab === 'general'" class="space-y-6 animate-in fade-in slide-in-from-right-4 duration-300">
          <div class="grid gap-2">
            <Label>Project Name</Label>
            <div class="relative">
              <div class="absolute left-3 top-3.5 text-gray-400 font-bold text-xs">senvanda.local/</div>
              <input v-model="form.name" class="w-full h-12 pl-28 pr-4 rounded-xl border border-gray-200 bg-white focus:border-black transition-all outline-none font-bold" placeholder="my-project" />
            </div>
          </div>

          <div class="grid gap-2">
            <Label>Git Branch</Label>
            <div class="relative">
              <GitBranch class="w-4 h-4 absolute left-3 top-3.5 text-gray-400" />
              <input v-model="form.branch" class="w-full h-12 pl-10 pr-4 rounded-xl border border-gray-200 bg-white focus:border-black transition-all outline-none" placeholder="main" />
            </div>
          </div>

          <div class="grid gap-2">
            <Label>Custom Domain (DNS)</Label>
            <div class="relative">
              <Globe class="w-4 h-4 absolute left-3 top-3.5 text-gray-400" />
              <input v-model="form.domain" class="w-full h-12 pl-10 pr-4 rounded-xl border border-gray-200 bg-white focus:border-black transition-all outline-none" placeholder="app.example.com" />
            </div>
            <p class="text-xs text-gray-400 font-medium">Automatic HTTPS via Caddy will be provisioned.</p>
          </div>
        </div>

        <!-- TAB: BUILD -->
        <div v-show="activeTab === 'build'" class="space-y-6 animate-in fade-in slide-in-from-right-4 duration-300">
          <div class="grid grid-cols-2 gap-6">
            <div class="grid gap-2">
              <Label>Framework / Runtime</Label>
              <div class="relative">
                <Box class="w-4 h-4 absolute left-3 top-3.5 text-gray-400" />
                <select v-model="form.framework" class="w-full h-12 pl-10 pr-4 rounded-xl border border-gray-200 bg-white focus:border-black outline-none appearance-none">
                  <option value="Node.js">Node.js</option>
                  <option value="Next.js">Next.js</option>
                  <option value="Go">Go</option>
                  <option value="PHP">PHP</option>
                  <option value="Python">Python</option>
                  <option value="Static/Unknown">Static / Dockerfile</option>
                </select>
              </div>
            </div>
            <div class="grid gap-2">
              <Label>Internal Port</Label>
              <div class="relative">
                <Server class="w-4 h-4 absolute left-3 top-3.5 text-gray-400" />
                <input v-model.number="form.port" type="number" class="w-full h-12 pl-10 pr-4 rounded-xl border border-gray-200 bg-white focus:border-black outline-none" placeholder="e.g. 3000" />
              </div>
            </div>
          </div>

          <div class="grid gap-2">
            <Label>Start Command</Label>
            <div class="relative">
              <Terminal class="w-4 h-4 absolute left-3 top-3.5 text-gray-400" />
              <input v-model="form.startCommand" class="w-full h-12 pl-10 pr-4 rounded-xl border border-gray-200 bg-white focus:border-black outline-none font-mono text-sm text-blue-600" placeholder="npm start" />
            </div>
          </div>
        </div>

        <!-- TAB: ENV -->
        <div v-show="activeTab === 'env'" class="space-y-4 animate-in fade-in slide-in-from-right-4 duration-300">
          <div class="flex justify-between items-center">
            <Label>Environment Variables</Label>
            <Button size="xs" variant="outline" @click="addEnv"><Plus class="w-3 h-3 mr-1" /> Add</Button>
          </div>
          <div class="space-y-3 max-h-[300px] overflow-y-auto pr-2 custom-scrollbar">
            <div v-for="(env, idx) in form.envVars" :key="idx" class="flex gap-2 group">
              <input v-model="env.key" placeholder="KEY" class="flex-1 h-10 px-3 rounded-lg border border-gray-200 bg-white focus:border-black outline-none font-mono text-xs uppercase" />
              <span class="text-gray-300 self-center">=</span>
              <input v-model="env.value" placeholder="VALUE" class="flex-1 h-10 px-3 rounded-lg border border-gray-200 bg-white focus:border-black outline-none font-mono text-xs" />
              <button @click="removeEnv(idx)" class="w-10 h-10 flex items-center justify-center text-gray-300 hover:text-red-500 transition-colors">
                <Trash2 class="w-4 h-4" />
              </button>
            </div>
          </div>
        </div>

        <!-- TAB: RESOURCES -->
        <div v-show="activeTab === 'resources'" class="space-y-6 animate-in fade-in slide-in-from-right-4 duration-300">
          <div class="p-4 bg-emerald-50 rounded-2xl border border-emerald-100 flex gap-3">
             <Server class="w-5 h-5 text-emerald-600 shrink-0" />
             <p class="text-[10px] font-bold text-emerald-800 uppercase tracking-widest mt-0.5">Performance Profiling Enabled</p>
          </div>
          <div class="grid grid-cols-2 gap-6">
            <div class="grid gap-2">
              <Label>CPU Limit (vCPU)</Label>
              <select v-model="form.resources.cpu" class="w-full h-12 px-4 rounded-xl border border-gray-200 bg-white focus:border-black outline-none">
                 <option value="0.25">0.25 (Starter)</option>
                 <option value="0.5">0.5 (Standard)</option>
                 <option value="1.0">1.0 (Performance)</option>
                 <option value="2.0">2.0 (High Power)</option>
              </select>
            </div>
            <div class="grid gap-2">
              <Label>Memory Limit (MB)</Label>
              <select v-model="form.resources.memory" class="w-full h-12 px-4 rounded-xl border border-gray-200 bg-white focus:border-black outline-none">
                 <option value="256MB">256 MB</option>
                 <option value="512MB">512 MB</option>
                 <option value="1024MB">1 GB</option>
                 <option value="2048MB">2 GB</option>
              </select>
            </div>
          </div>
        </div>

        <!-- TAB: DEVOPS -->
        <div v-show="activeTab === 'devops'" class="space-y-6 animate-in fade-in slide-in-from-right-4 duration-300">
           <div v-if="props.config.securityHints && props.config.securityHints.length" class="p-4 bg-red-50 border border-red-100 rounded-2xl space-y-2">
              <h4 class="text-xs font-bold text-red-600 flex items-center gap-2">
                 <Trash2 class="w-3 h-3" /> SECURITY VULNERABILITIES DETECTED
              </h4>
              <ul class="text-[10px] text-red-700 list-disc list-inside">
                 <li v-for="hint in props.config.securityHints" :key="hint">{{ hint }}</li>
              </ul>
           </div>
           <div class="space-y-4">
              <h4 class="text-xs font-bold text-gray-400 uppercase tracking-widest">Heuristic Discovery Profile</h4>
              <div class="grid grid-cols-2 gap-3">
                 <div v-for="(val, key) in props.config.devOpsProfile" :key="key" class="p-4 bg-gray-50 rounded-xl border border-gray-100">
                    <p class="text-[9px] font-bold text-gray-400 uppercase">{{ key }}</p>
                    <p class="text-sm font-bold text-gray-900 mt-1">{{ val }}</p>
                 </div>
              </div>
           </div>
           <div class="p-4 bg-blue-50 rounded-2xl border border-blue-100">
              <p class="text-[10px] text-blue-700 leading-relaxed font-medium italic">
                 * This profile was generated by Senvanda Heuristic Engine v1.0.
              </p>
           </div>
        </div>

        <!-- BOTTOM ACTIONS -->
        <div class="mt-auto pt-8 border-t border-gray-100 flex items-center gap-3">
          <Button variant="ghost" class="h-12 px-6 rounded-xl text-gray-500 hover:text-red-500 hover:bg-red-50" @click="$emit('cancel')">Cancel</Button>
          <div class="flex-1"></div>
          <Button variant="outline" class="h-12 px-6 rounded-xl border-gray-200 hover:bg-gray-50 text-gray-700 font-medium" @click="$emit('draft')">
            <Save class="w-4 h-4 mr-2" /> Save Draft
          </Button>
          <Button size="lg" class="h-12 px-8 rounded-xl text-md gap-2 bg-black hover:bg-gray-800 shadow-lg" @click="$emit('deploy')" :disabled="isDeploying">
            <Rocket class="w-4 h-4" v-if="!isDeploying" />
            <Loader2 class="w-4 h-4 animate-spin" v-else />
            {{ isDeploying ? "Deploying..." : "Deploy" }}
          </Button>
        </div>
      </div>
    </div>
  </div>
</template>
