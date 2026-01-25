import { defineStore } from 'pinia';
import { pb } from '../lib/pocketbase';
import { ref } from 'vue';

export const useAuthStore = defineStore('auth', () => {
  const user = ref(pb.authStore.model);

  const login = async (email, password) => {
    // Try Admin Auth first
    console.log("Attempting Admin Auth...");
    try {
        if (!pb.admins) {
            console.error("SDK Mismatch: pb.admins is undefined! You are likely using PocketBase SDK v0.23+ but the backend is older.");
        }
        await pb.admins.authWithPassword(email, password);
        console.log("Admin Auth Success");
    } catch (err) {
        console.error("Admin Auth Failed:", err);
        // If fail, try User Auth
        console.warn("Trying User auth...");
        try {
             await pb.collection('users').authWithPassword(email, password);
        } catch (userErr) {
            console.error("User Auth Failed:", userErr);
            throw userErr;
        }
    }
    user.value = pb.authStore.model;
  };

  const logout = () => {
    pb.authStore.clear();
    user.value = null;
  };

  return { user, login, logout };
});
