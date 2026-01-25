<script setup>
import { computed } from "vue";

const props = defineProps({
  currentStep: {
    type: Number,
    default: 1,
  },
});

const steps = [
  { label: "Source" },
  { label: "Analyzing" },
  { label: "Configure" },
  { label: "Deploy" },
];
</script>

<template>
  <div
    class="flex items-center justify-between mb-12 relative max-w-lg mx-auto"
  >
    <div class="absolute top-1/2 left-0 w-full h-1 bg-gray-100 -z-10"></div>

    <div
      v-for="(step, i) in steps"
      :key="i"
      class="flex flex-col items-center gap-2 bg-white px-2"
    >
      <div
        class="w-8 h-8 rounded-full flex items-center justify-center font-bold text-sm transition-all border-2"
        :class="
          i + 1 <= currentStep
            ? 'bg-black text-white border-black'
            : 'bg-white text-gray-300 border-gray-200'
        "
      >
        {{ i + 1 }}
      </div>
      <span
        class="text-[10px] font-bold uppercase tracking-wider transition-colors"
        :class="i + 1 <= currentStep ? 'text-black' : 'text-gray-300'"
        >{{ step.label }}</span
      >
    </div>
  </div>
</template>
