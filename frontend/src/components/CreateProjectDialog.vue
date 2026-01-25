<script setup>
import { ref } from 'vue';
import { 
    Dialog,
    DialogContent,
    DialogDescription,
    DialogHeader,
    DialogTitle,
    DialogTrigger,
    DialogFooter
} from '@/components/ui/dialog';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { Rocket, Loader2 } from 'lucide-vue-next';
import { pb } from '@/lib/pocketbase';

const open = ref(false);
const loading = ref(false);
const name = ref('');
const error = ref('');

const emit = defineEmits(['project-created']);

const handleSubmit = async () => {
    if (!name.value) return;
    
    loading.value = true;
    error.value = '';
    
    try {
        // Call Backend API
        const res = await pb.send('/api/senvanda/deploy/create', {
            method: 'POST',
            body: {
                name: name.value
            }
        });
        
        // Reset & Close
        name.value = '';
        open.value = false;
        emit('project-created', res);
        
    } catch (err) {
        console.error(err);
        error.value = err.data?.message || "Failed to launch cargo. Try a simplified name.";
    } finally {
        loading.value = false;
    }
};
</script>

<template>
  <Dialog v-model:open="open">
    <DialogTrigger as-child>
        <slot />
    </DialogTrigger>
    <DialogContent class="sm:max-w-[425px]">
      <DialogHeader>
        <DialogTitle class="flex items-center gap-2">
            <Rocket class="w-5 h-5 text-blue-600" />
            Launch New Cargo
        </DialogTitle>
        <DialogDescription>
          Deploy a new service container to your fleet.
        </DialogDescription>
      </DialogHeader>
      
      <div class="grid gap-4 py-4">
        <div class="grid gap-2">
          <Label for="name" class="text-right">
            Project Name
          </Label>
          <Input
            id="name"
            v-model="name"
            placeholder="e.g., website-landing"
            class="col-span-3"
            :disabled="loading"
            @keyup.enter="handleSubmit"
          />
        </div>
        
        <div  v-if="error" class="text-sm text-red-500 font-medium">
            {{ error }}
        </div>
      </div>
      
      <DialogFooter>
        <Button 
            type="submit" 
            class="w-full bg-black hover:bg-gray-800" 
            :disabled="loading || !name"
            @click="handleSubmit"
        >
            <Loader2 v-if="loading" class="w-4 h-4 mr-2 animate-spin" />
            {{ loading ? 'Initializing Thrusters...' : 'Deploy Container' }}
        </Button>
      </DialogFooter>
    </DialogContent>
  </Dialog>
</template>
