<script setup>
import { ref, onMounted } from "vue";
import { Cpu } from "lucide-vue-next";
import { pb } from "@/lib/pocketbase";

const props = defineProps(["url"]);
const emit = defineEmits(["complete"]);
const logs = ref([]);
const scanMeta = ref(null);

const log = async (msg) => {
  logs.value.push(msg);
  // Ensure the UI scrolls to bottom or stays reactive
};

onMounted(async () => {
  try {
    await log("Initializing Heuristic Engine cluster...");
    await log("Accessing secure buffer for " + props.url);

    const startTime = Date.now();

    // CALL REAL BACKEND
    const result = await pb.send("/api/senvanda/deploy/scan", {
      method: "POST",
      body: { url: props.url },
    });

    if (result.tracingLogs) {
      for (const msg of result.tracingLogs) {
        await new Promise((r) => setTimeout(r, 600)); 
        await log(msg);
      }
    }

    scanMeta.value = result;
    const duration = ((Date.now() - startTime) / 1000).toFixed(2);
    
    await log(`Heuristic Deep Discovery successful in ${duration}s.`);
    await log("Architectural blueprint generated. Transitioning to Review...");
    
    await new Promise((r) => setTimeout(r, 1200));
    emit("complete", result);
  } catch (err) {
    console.error(err);
    await log("CRITICAL HEURISTIC FAILURE: " + (err.message || "Deep Scan failed."));
    await log("Manual intervention required. Check repository visibility.");
  }
});
</script>

<template>
  <div class="animate-in fade-in zoom-in-95 duration-500 max-w-2xl mx-auto">
    <div
      class="bg-[#1e1e2e] text-green-400 font-mono p-6 rounded-[2rem] shadow-2xl min-h-[400px] flex flex-col relative overflow-hidden"
    >
      <!-- Terminal Controls -->
      <div class="flex gap-2 mb-6 opacity-30">
        <div class="w-3 h-3 rounded-full bg-red-500"></div>
        <div class="w-3 h-3 rounded-full bg-yellow-500"></div>
        <div class="w-3 h-3 rounded-full bg-green-500"></div>
      </div>

      <!-- Logs -->
      <div class="flex-1 space-y-2 text-sm z-10 font-mono tracking-tight">
        <p
          v-for="(msg, i) in logs"
          :key="i"
          class="flex items-start gap-2 animate-in slide-in-from-left-2 fade-in duration-300"
        >
          <span class="text-blue-500 shrink-0">➜</span>
          <span
            :class="msg.includes('ERROR') ? 'text-red-400 font-bold' : ''"
            >{{ msg }}</span
          >
        </p>
        <div class="flex items-center gap-2 mt-4 animate-pulse">
          <span class="w-2 h-4 bg-green-400"></span>
        </div>
      </div>

      <!-- Background Graphic -->
      <Cpu
        class="absolute -right-10 -bottom-10 w-64 h-64 text-white opacity-5 animate-spin-slow pointer-events-none"
      />
    </div>
    <div class="flex justify-center mt-6 gap-4">
      <p class="text-gray-400 text-sm font-medium self-center">
        Heuristic AI is analyzing your codebase...
      </p>
      <button
        @click="$emit('cancel')"
        class="text-red-500 hover:text-red-700 text-sm font-bold bg-red-50 hover:bg-red-100 px-4 py-2 rounded-lg transition-colors"
      >
        Abort & Save Draft
      </button>
    </div>
  </div>
</template>

<style scoped>
.animate-spin-slow {
  animation: spin 8s linear infinite;
}
@keyframes spin {
  from {
    transform: rotate(0deg);
  }
  to {
    transform: rotate(360deg);
  }
}
</style>
