<script setup>
import { computed } from "vue";
import {
  BadgeCheck,
  AlertCircle,
  Clock,
  MoreHorizontal,
  Play,
  Square,
  ExternalLink,
} from "lucide-vue-next";
import { Button } from "@/components/ui/button";

const props = defineProps({
  projects: {
    type: Array,
    default: () => [],
  },
});

const emit = defineEmits(["action"]);
</script>

<template>
  <div
    class="bg-white rounded-2xl border border-gray-200 shadow-sm overflow-hidden"
  >
    <!-- Table Header Toolbar -->
    <div
      class="p-4 border-b border-gray-100 flex flex-wrap items-center justify-between gap-4"
    >
      <div class="flex items-center gap-2">
        <h3 class="font-bold text-gray-800">Deployed Projects</h3>
        <span
          class="bg-gray-100 text-gray-600 text-xs px-2 py-0.5 rounded-full font-medium"
          >{{ projects.length }}</span
        >
      </div>

      <div class="flex items-center gap-2">
        <div
          class="flex items-center space-x-1 bg-gray-50 p-1 rounded-lg border border-gray-200"
        >
          <button
            class="px-3 py-1 text-xs font-medium bg-white text-gray-800 shadow-sm rounded-md transition-all"
          >
            All
          </button>
          <button
            class="px-3 py-1 text-xs font-medium text-gray-500 hover:text-gray-800 rounded-md transition-all"
          >
            Running
          </button>
          <button
            class="px-3 py-1 text-xs font-medium text-gray-500 hover:text-gray-800 rounded-md transition-all"
          >
            Stopped
          </button>
        </div>

        <Button variant="outline" size="sm" class="text-xs h-8"> Brief </Button>
      </div>
    </div>

    <!-- Table Content -->
    <div class="overflow-x-auto">
      <table class="w-full text-left text-sm">
        <thead
          class="bg-gray-50/50 text-gray-500 font-medium border-b border-gray-100"
        >
          <tr>
            <th class="px-6 py-3 w-12">
              <input
                type="checkbox"
                class="rounded border-gray-300 text-blue-600 focus:ring-blue-500 my-1"
              />
            </th>
            <th class="px-6 py-3">Project Name</th>
            <th class="px-6 py-3">Endpoint</th>
            <th class="px-6 py-3">Port</th>
            <th class="px-6 py-3 text-center">Status</th>
            <th class="px-6 py-3 text-right">Actions</th>
          </tr>
        </thead>
        <tbody class="divide-y divide-gray-50">
          <tr
            v-for="p in projects"
            :key="p.id"
            class="group hover:bg-gray-50/80 transition-colors"
          >
            <td class="px-6 py-4">
              <input
                type="checkbox"
                class="rounded border-gray-300 text-blue-600 focus:ring-blue-500"
              />
            </td>
            <td class="px-6 py-4">
              <div class="flex flex-col">
                <router-link
                  :to="`/projects/${p.id}`"
                  class="font-bold text-gray-900 hover:text-blue-600 hover:underline transition-colors block cursor-pointer"
                >
                  {{ p.name }}
                </router-link>
                <span class="text-xs text-gray-400 font-mono"
                  >ID: {{ p.id.slice(0, 8) }}...</span
                >
              </div>
            </td>
            <td class="px-6 py-4">
              <a
                :href="p.url"
                target="_blank"
                class="flex items-center gap-1 text-blue-600 hover:text-blue-800 transition-colors font-medium"
              >
                {{ p.url }}
                <ExternalLink class="w-3 h-3" />
              </a>
            </td>
            <td class="px-6 py-4 font-mono text-gray-600">{{ p.port }}</td>
            <td class="px-6 py-4 text-center">
              <div
                class="inline-flex items-center gap-1.5 px-2.5 py-1 rounded-full text-xs font-semibold border"
                :class="
                  p.status === 'running'
                    ? 'bg-emerald-50 text-emerald-700 border-emerald-100'
                    : 'bg-rose-50 text-rose-700 border-rose-100'
                "
              >
                <span
                  class="w-1.5 h-1.5 rounded-full"
                  :class="
                    p.status === 'running' ? 'bg-emerald-500' : 'bg-rose-500'
                  "
                ></span>
                {{ p.status === "running" ? "Active" : "Stopped" }}
              </div>
            </td>
            <td class="px-6 py-4 text-right">
              <div
                class="flex items-center justify-end gap-2 opacity-0 group-hover:opacity-100 transition-opacity"
              >
                <Button
                  v-if="p.status !== 'running'"
                  size="sm"
                  variant="outline"
                  class="h-8 w-8 p-0 rounded-full border-emerald-200 text-emerald-600 hover:bg-emerald-50 hover:border-emerald-300"
                  @click="emit('action', p.id, 'start')"
                >
                  <Play class="w-3.5 h-3.5 fill-current" />
                </Button>

                <Button
                  v-if="p.status === 'running'"
                  size="sm"
                  variant="outline"
                  class="h-8 w-8 p-0 rounded-full border-rose-200 text-rose-600 hover:bg-rose-50 hover:border-rose-300"
                  @click="emit('action', p.id, 'stop')"
                >
                  <Square class="w-3.5 h-3.5 fill-current" />
                </Button>

                <Button
                  variant="ghost"
                  size="sm"
                  class="h-8 w-8 p-0 rounded-full text-gray-400 hover:text-gray-800"
                >
                  <MoreHorizontal class="w-4 h-4" />
                </Button>
              </div>
            </td>
          </tr>
        </tbody>
      </table>

      <!-- Empty State -->
      <div v-if="projects.length === 0" class="p-12 text-center text-gray-400">
        <p>No projects found. Deploy your first service!</p>
      </div>
    </div>
  </div>
</template>
