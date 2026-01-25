<script setup>
import { Box, Play, Trash2, Anchor } from "lucide-vue-next";

const props = defineProps({
  container: {
    type: Object,
    required: true,
  },
});

const emit = defineEmits(["adopt"]);
</script>

<template>
  <div
    class="relative rounded-[2rem] p-6 flex flex-col justify-between transition-all duration-300 hover:shadow-md bg-white border-2 border-dashed border-gray-200 h-[220px] group"
  >
    <div class="flex justify-between items-start mb-2">
      <div class="flex items-center gap-3">
        <div class="w-10 h-10 rounded-full bg-blue-50 flex items-center justify-center text-blue-600">
          <Box class="w-5 h-5" />
        </div>
        <div>
          <h3 class="font-bold text-gray-900 truncate w-32" :title="container.name">
            {{ container.name.replace("/", "") }}
          </h3>
          <p class="text-[10px] font-bold text-blue-500 uppercase">Legacy Container</p>
        </div>
      </div>
      <div class="px-2 py-1 rounded-full bg-gray-100 text-[10px] font-bold text-gray-500">
        {{ container.state }}
      </div>
    </div>

    <div class="flex-1 flex flex-col justify-center">
      <div class="flex items-center gap-2 opacity-60 mb-1">
        <span class="text-xs font-medium truncate">Image: {{ container.image }}</span>
      </div>
      <div v-if="container.ports && container.ports.length" class="flex flex-wrap gap-1">
        <span v-for="p in container.ports" :key="p" class="px-1.5 py-0.5 bg-gray-50 rounded text-[9px] font-mono border">
          :{{ p }}
        </span>
      </div>
    </div>

    <div class="flex items-center justify-between mt-auto">
      <div class="flex -space-x-2">
         <div v-for="i in 3" :key="i" class="w-6 h-6 rounded-full border-2 border-white bg-gray-200"></div>
      </div>
      
      <button
        @click="$emit('adopt', container)"
        class="flex items-center gap-2 px-4 py-2 rounded-full bg-blue-600 text-white text-xs font-bold hover:bg-blue-700 transition-all shadow-sm group-hover:scale-105"
      >
        <Anchor class="w-3 h-3" />
        Adopt
      </button>
    </div>
  </div>
</template>
