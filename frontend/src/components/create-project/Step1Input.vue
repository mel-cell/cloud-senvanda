<script setup>
import { GitBranch, ArrowRight, ShieldCheck } from "lucide-vue-next";
import { Button } from "@/components/ui/button";
import { Label } from "@/components/ui/label";
import { ref } from "vue";

const url = ref("");
const emit = defineEmits(["next", "cancel"]);

const handleSubmit = () => {
  if (url.value) emit("next", url.value);
};
</script>

<template>
  <div
    class="animate-in fade-in slide-in-from-bottom-4 duration-500 max-w-2xl mx-auto"
  >
    <div
      class="bg-white p-8 rounded-[2rem] border border-gray-100 shadow-sm space-y-6"
    >
      <div class="flex flex-col gap-4">
        <Label class="text-lg font-medium">Repository URL</Label>
        <div class="relative group">
          <div
            class="absolute -inset-0.5 bg-gradient-to-r from-pink-600 to-purple-600 rounded-xl opacity-20 group-focus-within:opacity-100 transition duration-1000 group-hover:duration-200 blur"
          ></div>
          <div
            class="relative flex items-center bg-white rounded-xl overflow-hidden border border-gray-200 focus-within:border-transparent"
          >
            <div class="pl-4 text-gray-400">
              <GitBranch class="w-5 h-5" />
            </div>
            <input
              v-model="url"
              type="text"
              class="w-full h-14 pl-3 pr-4 outline-none text-lg text-gray-800 placeholder-gray-300"
              placeholder="https://github.com/username/repo"
              @keyup.enter="handleSubmit"
            />
          </div>
        </div>
        <p class="text-sm text-gray-400 flex items-center gap-1">
          <ShieldCheck class="w-4 h-4 text-green-500" />
          Only Public Repositories supported for now.
        </p>
      </div>

      <div class="flex flex-col gap-3">
        <Button
          size="lg"
          class="w-full h-14 rounded-xl text-lg gap-2 bg-black hover:bg-gray-800 transition-all shadow-lg hover:shadow-xl"
          :disabled="!url"
          @click="handleSubmit"
        >
          Analyze Project <ArrowRight class="w-5 h-5" />
        </Button>

        <Button
          variant="ghost"
          class="w-full text-gray-500 hover:bg-gray-50 hover:text-red-500 h-10 rounded-lg"
          @click="$emit('cancel')"
        >
          Cancel
        </Button>
      </div>
    </div>

    <!-- Quick Tips -->
    <div class="mt-8 grid grid-cols-2 gap-4 text-center">
      <div class="p-4 bg-gray-50 rounded-2xl border border-gray-100">
        <span class="text-2xl mb-2 block">⚡️</span>
        <p class="font-bold text-sm text-gray-800">Instant Deploy</p>
        <p class="text-xs text-gray-400">Push to git, we handle the rest.</p>
      </div>
      <div class="p-4 bg-gray-50 rounded-2xl border border-gray-100">
        <span class="text-2xl mb-2 block">🔒</span>
        <p class="font-bold text-sm text-gray-800">Auto SSL</p>
        <p class="text-xs text-gray-400">Free HTTPS for every project.</p>
      </div>
    </div>
  </div>
</template>
