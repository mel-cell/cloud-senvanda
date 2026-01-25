<script setup>
import { computed, onMounted, onUnmounted, ref } from 'vue';

const props = defineProps({
  initialData: {
    type: Array,
    default: () => Array.from({ length: 24 }, () => Math.floor(Math.random() * 50) + 20)
  },
  height: {
    type: Number,
    default: 120
  },
  realtime: {
    type: Boolean,
    default: false
  }
});

const dataPoints = ref([...props.initialData]);
let interval = null;

onMounted(() => {
    if (props.realtime) {
        interval = setInterval(() => {
            // Simulate organic traffic: fluctuate slightly around last point
            const last = dataPoints.value[dataPoints.value.length - 1];
            const change = Math.floor(Math.random() * 20) - 10;
            let next = last + change;
            if (next < 10) next = 10 + Math.random() * 10;
            if (next > 95) next = 95 - Math.random() * 10;

            dataPoints.value.push(next);
            dataPoints.value.shift(); // Remove oldest
        }, 2000); // Update every 2s
    }
});

onUnmounted(() => {
    if (interval) clearInterval(interval);
});

// Generate SVG Path for smooth area chart
const pathData = computed(() => {
  const points = dataPoints.value;
  const width = 100;
  const max = 100; // Fixed max scale for stability
  
  const stepX = width / (points.length - 1);
  
  let path = `M 0,${100 - (points[0] / max) * 80}`;
  
  points.slice(1).forEach((val, i) => {
     const x = (i + 1) * stepX;
     const y = 100 - (val / max) * 80;
     
     // Bezier curve for smoothness could be added here, 
     // but straight lines are fine for activity graph look
     path += ` L ${x},${y}`;
  });
  
  return path;
});

const fillPath = computed(() => {
    return `${pathData.value} L 100,100 L 0,100 Z`;
});
</script>

<template>
  <div class="w-full h-full flex flex-col justify-end relative overflow-hidden rounded-xl bg-gray-50 border border-gray-100 transition-all duration-500">
      
      <!-- Chart Container -->
      <svg viewBox="0 0 100 100" preserveAspectRatio="none" class="w-full h-full absolute bottom-0 transition-transform duration-500 ease-linear">
          <defs>
            <linearGradient id="gradient" x1="0" x2="0" y1="0" y2="1">
              <stop offset="0%" stop-color="#000" stop-opacity="0.1" />
              <stop offset="100%" stop-color="#000" stop-opacity="0" />
            </linearGradient>
          </defs>
          
          <!-- Area Fill -->
          <path :d="fillPath" fill="url(#gradient)" class="transition-all duration-1000 ease-in-out" />
          
          <!-- Stroke Line -->
          <path :d="pathData" fill="none" stroke="black" stroke-width="0.5" stroke-opacity="0.5" vector-effect="non-scaling-stroke" class="transition-all duration-1000 ease-in-out" />
      </svg>

      <!-- Overlay Info -->
      <div class="absolute top-4 left-4 z-10">
          <p class="text-xs font-bold text-gray-400 uppercase tracking-wider flex items-center gap-2">
             Requests 
             <span v-if="realtime" class="w-2 h-2 rounded-full bg-green-500 animate-pulse"></span>
          </p>
          <div class="flex items-end gap-2">
              <span class="text-2xl font-bold text-gray-900">{{ Math.floor(dataPoints[dataPoints.length-1] * 12.5) }} req/m</span>
          </div>
      </div>

  </div>
</template>
