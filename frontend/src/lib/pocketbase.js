import PocketBase from 'pocketbase';

// Use relative URL to leverage Vite Proxy
export const pb = new PocketBase('/');

export const currentUser = pb.authStore.model;
