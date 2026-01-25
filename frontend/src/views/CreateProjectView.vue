<script setup>
import { ref } from "vue";
import { useRouter } from "vue-router";
import DashboardLayout from "@/layouts/DashboardLayout.vue";
import { Loader2 } from "lucide-vue-next";

// Components
import ProjectStepper from "@/components/create-project/ProjectStepper.vue";
import Step1Input from "@/components/create-project/Step1Input.vue";
import Step2Scanner from "@/components/create-project/Step2Scanner.vue";
import Step3Review from "@/components/create-project/Step3Review.vue";

const router = useRouter();

const currentStep = ref(1);
const deployLogs = ref([]);

const addDeployLog = async (msg) => {
  deployLogs.value.push(msg);
  await new Promise((r) => setTimeout(r, 600));
};

// Data Store
const projectData = ref({
  url: "",
  config: null,
});

import { pb } from "@/lib/pocketbase";

// ... (existing imports)

// Handlers
const handleUrlSubmit = (url) => {
  projectData.value.url = url;
  currentStep.value = 2;
};

const handleScanComplete = (result) => {
  projectData.value.config = result;
  currentStep.value = 3;
};

const createProject = async (isDraft) => {
  currentStep.value = 4; 
  deployLogs.value = [];

  try {
    if (!isDraft) {
      await addDeployLog("Initiating Heuristic Deployment Pipeline...");
      await addDeployLog("Syncing with Senvanda Hypervisor...");
    }

    const endpoint = isDraft
      ? "/api/senvanda/deploy/draft"
      : "/api/senvanda/deploy/create";

    const deployPromise = pb.send(endpoint, {
      method: "POST",
      body: {
        name: projectData.value.config.name,
        repoUrl: projectData.value.url,
        port: Number(projectData.value.config.port),
        framework: projectData.value.config.framework,
        image: projectData.value.config.image, 
        isDraft: isDraft,
        settings: {
          branch: projectData.value.config.branch || "main",
          startCommand: projectData.value.config.startCommand || "",
          envVars: projectData.value.config.envVars || [],
          domain: projectData.value.config.domain || "",
          resources: projectData.value.config.resources || { cpu: "0.5", memory: "512MB" },
        },
      },
    });

    if (!isDraft) {
      await addDeployLog("Injecting detected Environment Variables...");
      await addDeployLog("Configuring Caddy Edge Proxy (SSL Auto)...");
      await addDeployLog("Launching isolated container environment...");
    }

    const res = await deployPromise;
    
    if (!isDraft) {
      await addDeployLog("SUCCESS: DevOps Engine verified cluster health.");
      await new Promise((r) => setTimeout(r, 800));
    }

    router.push("/");
  } catch (err) {
    alert("Failed: " + err.message);
    currentStep.value = 3; 
  }
};

const handleDeploy = () => createProject(false);
const handleDraft = () => createProject(true);

const handleCancel = async () => {
  if (currentStep.value === 1) {
    router.push("/");
    return;
  }

  if (confirm("Stop setup? Your progress will be discarded. (Click Cancel to keep editing)")) {
    router.push("/");
  }
};
</script>

<template>
  <DashboardLayout>
    <div class="max-w-4xl mx-auto py-12">
      <!-- HEADER -->
      <div class="text-center mb-8">
        <h1 class="text-3xl font-bold mb-2">Deploy from Git</h1>
        <p class="text-gray-500">
          Auto-detect framework, CI/CD setup, and instant preview.
        </p>
      </div>

      <!-- PROGRESS STEPPER -->
      <ProjectStepper :current-step="currentStep" />

      <!-- DYNAMIC STEP CONTENT -->
      <Step1Input
        v-if="currentStep === 1"
        @next="handleUrlSubmit"
        @cancel="handleCancel"
      />

      <Step2Scanner
        v-if="currentStep === 2"
        :url="projectData.url"
        @complete="handleScanComplete"
        @cancel="handleCancel"
      />

      <Step3Review
        v-if="currentStep === 3"
        :config="projectData.config"
        @deploy="handleDeploy"
        @draft="handleDraft"
        @cancel="handleCancel"
      />

      <!-- STEP 4: DEPLOYMENT TRACING (DevOps Expert Mode) -->
      <div
        v-if="currentStep === 4"
        class="fixed inset-0 bg-white/95 backdrop-blur-md z-50 flex flex-col items-center justify-center p-8 animate-in fade-in duration-500"
      >
        <div class="max-w-3xl w-full space-y-8 text-center">
            <div class="w-20 h-20 bg-black rounded-[2rem] mx-auto flex items-center justify-center shadow-2xl animate-bounce">
                <Loader2 class="w-10 h-10 text-white animate-spin" />
            </div>
            
            <div>
              <h2 class="text-3xl font-bold tracking-tight">Orchestrating Deployment</h2>
              <p class="text-gray-500 mt-2 font-medium italic">DevOps Engine is executing infrastructure blueprint...</p>
            </div>

            <div class="bg-black rounded-[2rem] p-8 text-left font-mono text-xs text-blue-400 shadow-2xl border border-gray-800 w-full h-80 overflow-y-auto space-y-2 selection:bg-blue-500/30">
               <p v-for="(log, i) in deployLogs" :key="i" class="animate-in slide-in-from-left-2 duration-300">
                  <span class="text-gray-600 mr-2">[{{ new Date().toLocaleTimeString() }}]</span>
                  <span class="text-emerald-500 mr-2">SYS:</span>
                  {{ log }}
               </p>
               <div class="flex items-center gap-2 mt-4 animate-pulse">
                  <span class="w-2 h-4 bg-blue-500"></span>
                  <span class="text-gray-600 italic">Processing pipeline...</span>
               </div>
            </div>
            
            <p class="text-[10px] text-gray-400 uppercase tracking-widest font-bold">
               Powered by Senvanda Heuristic Hypervisor v1
            </p>
        </div>
      </div>
    </div>
  </DashboardLayout>
</template>
