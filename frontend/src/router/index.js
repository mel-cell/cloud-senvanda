import { createRouter, createWebHistory } from "vue-router";
import { useAuthStore } from "../stores/auth";

const router = createRouter({
  history: createWebHistory(import.meta.env.BASE_URL),
  routes: [
    {
      path: "/login",
      name: "login",
      component: () => import("../views/LoginView.vue"),
      meta: { guest: true },
    },
    {
      path: "/",
      name: "dashboard",
      component: () => import("../views/DashboardHome.vue"),
      meta: { requiresAuth: true },
    },
    {
      path: "/projects",
      name: "projects",
      // Placeholder for now, redirect to dashboard or create a view
      component: () => import("../views/DashboardHome.vue"),
      meta: { requiresAuth: true },
    },
    {
      path: "/projects/new",
      name: "create-project",
      component: () => import("../views/CreateProjectView.vue"),
      meta: { requiresAuth: true },
    },
    {
      path: "/projects/:id",
      name: "project-detail",
      component: () => import("../views/ProjectDetailView.vue"),
      meta: { requiresAuth: true },
    },
    {
      path: "/analytics",
      name: "analytics",
      component: () => import("../views/AnalyticsView.vue"),
      meta: { requiresAuth: true },
    },
    {
      path: "/team",
      name: "team",
      component: () => import("../views/TeamView.vue"),
      meta: { requiresAuth: true },
    },
    {
      path: "/settings",
      name: "settings",
      component: () => import("../views/SettingsView.vue"),
      meta: { requiresAuth: true },
    },
  ],
});

router.beforeEach((to, from, next) => {
  const auth = useAuthStore();

  if (to.meta.requiresAuth && !auth.user) {
    next("/login");
  } else if (to.meta.guest && auth.user) {
    next("/");
  } else {
    next();
  }
});

export default router;
