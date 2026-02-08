<script setup>
import { computed } from 'vue';

const props = defineProps({
  data: {
    type: Array,
    required: true // Array of percentages (0-100)
  },
  label: {
    type: String,
    default: 'Usage'
  },
  color: {
    type: String,
    default: 'black' 
  },
  height: {
      type: Number,
      default: 100
  }
});

// Generate SVG Path
const pathData = computed(() => {
  const points = props.data;
  if (!points || points.length < 2) return `M 0,100 L 100,100`;

  const width = 100;
  const max = 100; 
  
  const stepX = width / (points.length - 1);
  
  let d = `M 0,${100 - (points[0] / max) * 100}`;
  
  points.slice(1).forEach((val, i) => {
     const x = (i + 1) * stepX;
     const y = 100 - (Math.min(val, 100) / max) * 100;
     d += ` L ${x},${y}`;
  });
  
  return d;
});

const fillPath = computed(() => {
    return `${pathData.value} L 100,100 L 0,100 Z`;
});

const strokeColor = computed(() => {
    if (props.color === 'purple') return '#a855f7';
    if (props.color === 'green') return '#22c55e';
    if (props.color === 'blue') return '#3b82f6';
    return '#000000';
});

const fillColor = computed(() => {
     if (props.color === 'purple') return 'rgba(168, 85, 247, 0.1)';
    if (props.color === 'green') return 'rgba(34, 197, 94, 0.1)';
    if (props.color === 'blue') return 'rgba(59, 130, 246, 0.1)';
    return 'rgba(0,0,0,0.05)';
});
</script>

<template>
  <div class="w-full h-full relative overflow-hidden rounded-xl transition-all duration-300">
      <svg viewBox="0 0 100 100" preserveAspectRatio="none" class="w-full h-full absolute bottom-0">
          <path :d="fillPath" :fill="fillColor" class="transition-all duration-300 ease-out" />
          <path :d="pathData" fill="none" :stroke="strokeColor" stroke-width="2" vector-effect="non-scaling-stroke" stroke-linecap="round" stroke-linejoin="round" class="transition-all duration-300 ease-out" />
      </svg>
      
      <!-- Current Value Indication -->
      <div v-if="props.data.length > 0" class="absolute top-2 right-2 text-xs font-mono font-bold" :style="{ color: strokeColor }">
          {{ props.data[props.data.length-1].toFixed(1) }}%
      </div>
       <div class="absolute top-2 left-2 text-[10px] font-bold uppercase tracking-wider text-gray-400">
          {{ label }}
      </div>
  </div>
</template>
