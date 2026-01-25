<script setup>
import { reactive, watch, ref } from "vue";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
} from "@/components/ui/dialog";
import { Plus, Trash2, Save, Settings } from "lucide-vue-next";
import { pb } from "@/lib/pocketbase";

const props = defineProps({
  project: Object,
  triggerClass: String,
});

const emit = defineEmits(["saved"]);
const open = ref(false);

const form = reactive({
  repoUrl: "",
  branch: "main",
  port: 3000,
  startCommand: "",
  envVars: [],
});

// Sync form when dialog opens
watch(
  () => props.project,
  (p) => {
    if (p) {
      form.repoUrl = p.repoUrl || "";
      form.port = p.port;
      const settings = p.settings || {};
      form.branch = settings.branch || "main";
      form.startCommand = settings.startCommand || "";
      form.envVars = settings.envVars
        ? JSON.parse(JSON.stringify(settings.envVars))
        : [];
    }
  },
  { immediate: true },
);

const addEnv = () => form.envVars.push({ key: "", value: "" });
const removeEnv = (idx) => form.envVars.splice(idx, 1);

const saveSettings = async () => {
  try {
    const payload = {
      repoUrl: form.repoUrl,
      port: form.port,
      settings: {
        branch: form.branch,
        startCommand: form.startCommand,
        envVars: form.envVars,
      },
    };

    await pb.collection("projects").update(props.project.id, payload);
    emit("saved");
    open.value = false;
    alert("Settings saved. Please Redeploy to apply changes.");
  } catch (err) {
    alert("Failed to save: " + err.message);
  }
};
</script>

<template>
  <Dialog v-model:open="open">
    <DialogTrigger as-child>
      <slot name="trigger">
        <Button variant="outline" size="sm" :class="triggerClass">
          <Settings class="w-4 h-4 mr-2" /> Configure
        </Button>
      </slot>
    </DialogTrigger>
    <DialogContent class="sm:max-w-[600px] max-h-[85vh] overflow-y-auto">
      <DialogHeader>
        <DialogTitle>Project Configuration</DialogTitle>
        <DialogDescription>
          Update deployment settings. Changes require a redeploy to take effect.
        </DialogDescription>
      </DialogHeader>

      <div class="grid gap-4 py-4">
        <div class="grid gap-2">
          <Label>Repository URL</Label>
          <Input v-model="form.repoUrl" placeholder="https://github.com/..." />
        </div>

        <div class="grid grid-cols-2 gap-4">
          <div class="grid gap-2">
            <Label>Branch</Label>
            <Input v-model="form.branch" placeholder="main" />
          </div>
          <div class="grid gap-2">
            <Label>Port</Label>
            <Input v-model.number="form.port" type="number" />
          </div>
        </div>

        <div class="grid gap-2">
          <Label>Start Command</Label>
          <Input
            v-model="form.startCommand"
            placeholder="npm start"
            class="font-mono text-sm"
          />
        </div>

        <!-- Env Vars -->
        <div class="space-y-3 pt-4 border-t">
          <div class="flex justify-between items-center">
            <Label>Environment Variables</Label>
            <Button size="xs" variant="outline" @click="addEnv"
              ><Plus class="w-3 h-3 mr-1" /> Add</Button
            >
          </div>
          <div v-for="(env, idx) in form.envVars" :key="idx" class="flex gap-2">
            <Input
              v-model="env.key"
              placeholder="KEY"
              class="uppercase font-mono text-xs"
            />
            <Input
              v-model="env.value"
              placeholder="VALUE"
              class="font-mono text-xs"
            />
            <Button
              size="icon"
              variant="ghost"
              class="text-red-500 hover:bg-red-50"
              @click="removeEnv(idx)"
              ><Trash2 class="w-4 h-4"
            /></Button>
          </div>
        </div>
      </div>

      <DialogFooter>
        <Button variant="outline" @click="open = false">Cancel</Button>
        <Button @click="saveSettings">Save Changes</Button>
      </DialogFooter>
    </DialogContent>
  </Dialog>
</template>
