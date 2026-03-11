import { createRouter, createWebHistory } from 'vue-router'
import { useAuthStore } from '../stores/auth.js'
import LoginView from '../views/LoginView.vue'
import DashboardView from '../views/DashboardView.vue'
import ComparisonView from '../views/ComparisonView.vue'
import ShareView from '../views/ShareView.vue'

const routes = [
  { path: '/login',          component: LoginView,      meta: { public: true } },
  { path: '/dashboard',      component: DashboardView,  meta: { requiresAuth: true } },
  { path: '/comparison/:id', component: ComparisonView, meta: { requiresAuth: true } },
  { path: '/share/:token',   component: ShareView,      meta: { public: true } },
  { path: '/',               redirect: '/dashboard' },
]

const router = createRouter({
  history: createWebHistory(),
  routes,
})

let authInitialized = false

router.beforeEach(async (to) => {
  const auth = useAuthStore()

  if (!authInitialized) {
    await auth.fetchMe()
    authInitialized = true
  }

  if (to.meta.requiresAuth && !auth.user) {
    return '/login'
  }
  if (to.path === '/login' && auth.user) {
    return '/dashboard'
  }
})

export default router
