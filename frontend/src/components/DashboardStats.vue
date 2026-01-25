<script setup>
import { AlertCircle, CheckCircle2, Box } from "lucide-vue-next";

defineProps({
  stats: {
    type: Object,
    default: () => ({
      total: 0,
      running: 0,
      stopped: 0,
      cpu: 0,
    }),
  },
});
</script>

<template>
  <div class="grid grid-cols-1 md:grid-cols-4 gap-4">
    <!-- Main Highlight Card -->
    <div
      class="md:col-span-2 bg-gradient-to-br from-blue-600 to-indigo-700 rounded-2xl p-6 text-white shadow-lg relative overflow-hidden group"
    >
      <div
        class="absolute -right-10 -top-10 w-40 h-40 bg-white/10 rounded-full blur-3xl group-hover:bg-white/20 transition-all duration-700"
      ></div>

      <div class="relative z-10 flex flex-col justify-between h-full">
        <div class="flex items-start justify-between">
          <div>
            <p class="text-blue-100 font-medium mb-1">System Status</p>
            <h2 class="text-3xl font-bold tracking-tight">Optimal</h2>
          </div>
          <div class="bg-white/20 backdrop-blur-md p-2 rounded-lg">
            <CheckCircle2 class="w-6 h-6 text-white" />
          </div>
        </div>

        <div class="mt-8 flex items-center gap-6">
          <div>
            <p
              class="text-xs text-blue-200 uppercase tracking-wider font-semibold mb-1"
            >
              CPU Load
            </p>
            <p class="text-xl font-bold font-mono">{{ stats.cpu }}%</p>
          </div>
          <div>
            <p
              class="text-xs text-blue-200 uppercase tracking-wider font-semibold mb-1"
            >
              Memory
            </p>
            <p class="text-xl font-bold font-mono">1.2GB</p>
          </div>
          <div class="h-8 w-px bg-white/20"></div>
          <div>
            <p
              class="text-xs text-blue-200 uppercase tracking-wider font-semibold mb-1"
            >
              Uptime
            </p>
            <p class="text-xl font-bold font-mono">24d 2h</p>
          </div>
        </div>
      </div>
    </div>

    <!-- Secondary Stat Cards -->
    <div
      class="bg-white rounded-2xl p-5 border border-gray-200 shadow-sm flex flex-col justify-between"
    >
      <div class="flex items-center justify-between mb-4">
        <span class="text-gray-500 font-medium text-sm">Active Containers</span>
        <div class="w-2 h-2 rounded-full bg-emerald-500 animate-pulse"></div>
      </div>
      <div class="flex items-end gap-2">
        <span class="text-4xl font-bold text-gray-800 tracking-tighter">{{
          stats.running
        }}</span>
        <span class="text-gray-400 font-medium mb-1">/ {{ stats.total }}</span>
      </div>
      <div class="mt-4 w-full bg-gray-100 h-1.5 rounded-full overflow-hidden">
        <div
          class="h-full bg-emerald-500 rounded-full transition-all duration-500"
          :style="`width: ${(stats.running / (stats.total || 1)) * 100}%`"
        ></div>
      </div>
    </div>

    <div
      class="bg-white rounded-2xl p-5 border border-gray-200 shadow-sm flex flex-col justify-between"
    >
      <div class="flex items-center justify-between mb-4">
        <span class="text-gray-500 font-medium text-sm">Total Projects</span>
        <Box class="w-4 h-4 text-gray-400" />
      </div>
      <div class="flex items-end gap-2">
        <span class="text-4xl font-bold text-gray-800 tracking-tighter">{{
          stats.total
        }}</span>
      </div>
      <p class="mt-2 text-xs text-gray-500">
        You have
        <span class="font-bold text-gray-700">{{ 5 - stats.total }}</span> free
        slots remaining in your plan.
      </p>
    </div>
  </div>
</template>
