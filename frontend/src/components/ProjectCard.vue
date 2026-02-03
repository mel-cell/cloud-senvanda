<script setup>
import { computed } from "vue";
import { useRouter } from "vue-router";
import { Play, Square, MoreHorizontal, Globe, Activity } from "lucide-vue-next";

const props = defineProps({
  project: {
    type: Object,
    required: true,
  },
});

const emit = defineEmits(["action"]);
const router = useRouter(); // Use Router

const navigateToDetail = () => {
  router.push(`/projects/${props.project.id}`);
};

// Pastel Color Map mimicking the reference image
// 'running': Soft Yellow/Green (Active)
// 'stopped': Soft Gray/Purple (Inactive)
// 'building': Soft Blue/Pink
const cardTheme = computed(() => {
  switch (props.project.status) {
    case "building":
    case "deploying":
      return {
        bg: "bg-[#FFF7D1]", // Soft Yellow
        text: "text-yellow-900",
        accent: "bg-yellow-200",
        icon: "text-yellow-600",
      };
    case "online":
    case "running":
      return {
        bg: "bg-[#F2E8FF]", // Light Purple
        text: "text-purple-900",
        accent: "bg-purple-200",
        icon: "text-purple-600",
      };
    case "stopped":
      return {
        bg: "bg-[#F3F2EE]", // Soft Gray
        text: "text-gray-900",
        accent: "bg-gray-200",
        icon: "text-gray-500",
      };
    case "draft":
      return {
        bg: "bg-gray-50", // Very light gray (draft)
        text: "text-gray-500",
        accent: "bg-gray-200",
        icon: "text-gray-400",
      };
    case "failed":
      return {
        bg: "bg-red-50",
        text: "text-red-900",
        accent: "bg-red-200",
        icon: "text-red-600",
      };
    default:
      return {
        bg: "bg-[#F3F2EE]",
        text: "text-gray-900",
        accent: "bg-gray-200",
        icon: "text-gray-600",
      };
  }
});
</script>

<template>
  <div
    class="relative rounded-[2rem] p-6 flex flex-col justify-between transition-all duration-300 hover:shadow-md group h-[220px] border-2 cursor-pointer"
    :class="[
      cardTheme.bg,
      project.status === 'draft'
        ? 'border-dashed border-gray-300 opacity-80 hover:opacity-100'
        : 'border-transparent',
    ]"
    @click="navigateToDetail"
  >
    <!-- Top Row: Title & Action -->
    <div class="flex justify-between items-start mb-4">
      <div>
        <div class="flex items-center gap-2 mb-1">
          <span
            class="w-2 h-2 rounded-full"
            :class="
              ['running', 'online'].includes(project.status)
                ? 'bg-green-500 animate-pulse'
                : project.status === 'draft'
                  ? 'bg-gray-400'
                  : ['building', 'deploying'].includes(project.status)
                    ? 'bg-yellow-400 animate-pulse'
                    : 'bg-red-400'
            "
          ></span>
          <p class="text-[10px] font-bold uppercase tracking-widest opacity-60">
            {{ project.status }}
          </p>
        </div>

        <h3
          class="font-bold text-lg leading-tight truncate w-32 hover:opacity-75"
          :title="project.name"
        >
          {{ project.name || "Untitled Project" }}
        </h3>
      </div>

      <button
        class="w-8 h-8 rounded-full bg-white/50 hover:bg-white flex items-center justify-center transition-colors"
        @click.stop
      >
        <MoreHorizontal class="w-4 h-4 opacity-50" />
      </button>
    </div>

    <!-- Middle: Metric / Info -->
    <div class="flex-1 flex flex-col justify-center">
      <!-- Progress Bar Visual -->
      <div class="w-full h-1 bg-black/5 rounded-full mb-3 overflow-hidden">
        <div class="h-full bg-black/20 rounded-full" style="width: 65%"></div>
      </div>

      <div class="flex items-center gap-2 opacity-60">
        <div
          class="w-6 h-6 rounded-full bg-white/40 flex items-center justify-center"
        >
          <Globe class="w-3 h-3" />
        </div>
        <span class="text-sm font-medium truncate">{{
          project.repoUrl || "No Source"
        }}</span>
      </div>
    </div>

    <!-- Bottom: Owner & Controls -->
    <div class="flex items-center justify-between mt-auto">
      <!-- Owner Avatar -->
      <div class="flex items-center gap-3">
        <div
          class="w-10 h-10 rounded-full bg-white shadow-sm flex items-center justify-center overflow-hidden"
        >
          <img
            v-if="project.avatarUrl"
            :src="project.avatarUrl"
            class="w-full h-full object-cover"
          />
          <span v-else class="text-sm font-bold opacity-50">M</span>
        </div>
        <div class="flex flex-col">
          <span class="text-xs font-bold opacity-80">Melvin</span>
          <span class="text-[10px] opacity-50">21 Nov 23</span>
        </div>
      </div>

      <!-- FAB Action Button -->
      <!-- DRAFT: RESUME -->
      <button
        v-if="project.status === 'draft'"
        class="w-10 h-10 rounded-full bg-white text-gray-600 border border-gray-200 flex items-center justify-center shadow-sm hover:bg-gray-50 hover:scale-105 transition-all"
        @click.stop="$emit('action', project.id, 'resume')"
        title="Resume Setup"
      >
        <Activity class="w-4 h-4" />
      </button>

      <!-- STOPPED: START -->
      <button
        v-else-if="project.status !== 'running'"
        class="w-10 h-10 rounded-full bg-black text-white flex items-center justify-center shadow-lg hover:bg-gray-800 hover:scale-105 transition-all"
        @click.stop="$emit('action', project.id, 'start')"
      >
        <Play class="w-4 h-4 fill-current ml-0.5" />
      </button>

      <!-- RUNNING: STOP -->
      <button
        v-else
        class="w-10 h-10 rounded-full bg-white text-black flex items-center justify-center shadow-lg hover:bg-gray-100 hover:scale-105 transition-all"
        @click.stop="$emit('action', project.id, 'stop')"
      >
        <Square class="w-4 h-4 fill-current" />
      </button>
    </div>
  </div>
</template>
