<script setup>
import { ref } from "vue";
import { useAuthStore } from "../stores/auth";
import { useRouter } from "vue-router";
import { Button } from "@/components/ui/button";
import { pb } from "@/lib/pocketbase";

const email = ref("");
const password = ref("");
const loading = ref(false);
const error = ref("");
const auth = useAuthStore();
const router = useRouter();

const handleLogin = async () => {
  loading.value = true;
  error.value = "";
  try {
    await auth.login(email.value, password.value);
    router.replace("/");
  } catch (err) {
    error.value = "Invalid email or password";
  } finally {
    loading.value = false;
  }
};
</script>

<template>
  <div class="min-h-screen grid lg:grid-cols-2">
    <!-- Left Side: Form -->
    <div class="flex items-center justify-center p-8 bg-background">
      <div
        class="mx-auto flex w-full flex-col justify-center space-y-6 sm:w-[350px]"
      >
        <div class="flex flex-col space-y-2 text-center">
          <h1 class="text-2xl font-semibold tracking-tight">
            Login to Senvanda
          </h1>
          <p class="text-sm text-muted-foreground">
            Enter your credentials to access your console
          </p>
        </div>

        <form @submit.prevent="handleLogin" class="space-y-4">
          <div class="space-y-2">
            <label
              class="text-sm font-medium leading-none peer-disabled:cursor-not-allowed peer-disabled:opacity-70"
              for="email"
              >Email</label
            >
            <input
              v-model="email"
              class="flex h-10 w-full rounded-md border border-input bg-background px-3 py-2 text-sm ring-offset-background file:border-0 file:bg-transparent file:text-sm file:font-medium placeholder:text-muted-foreground focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2 disabled:cursor-not-allowed disabled:opacity-50"
              id="email"
              placeholder="m@example.com"
              type="email"
              required
            />
          </div>
          <div class="space-y-2">
            <label
              class="text-sm font-medium leading-none peer-disabled:cursor-not-allowed peer-disabled:opacity-70"
              for="password"
              >Password</label
            >
            <input
              v-model="password"
              class="flex h-10 w-full rounded-md border border-input bg-background px-3 py-2 text-sm ring-offset-background file:border-0 file:bg-transparent file:text-sm file:font-medium placeholder:text-muted-foreground focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2 disabled:cursor-not-allowed disabled:opacity-50"
              id="password"
              type="password"
              required
            />
          </div>

          <div v-if="error" class="text-sm font-medium text-destructive">
            {{ error }}
          </div>

          <Button class="w-full" type="submit" :disabled="loading">
            <span v-if="loading">Signing in...</span>
            <span v-else>Sign In with Email</span>
          </Button>
        </form>
      </div>
    </div>

    <!-- Right Side: Decorative -->
    <div class="hidden bg-muted lg:block relative overflow-hidden">
      <div class="absolute inset-0 bg-primary/90 mix-blend-multiply" />
      <!-- Generate an image placeholder or gradient -->
      <div class="absolute bottom-10 left-10 text-white z-20">
        <blockquote class="space-y-2">
          <p class="text-lg">
            "Deploying apps on your own infrastructure has never been this
            beautiful."
          </p>
          <footer class="text-sm">Cloud Senvanda Team</footer>
        </blockquote>
      </div>
    </div>
  </div>
</template>
