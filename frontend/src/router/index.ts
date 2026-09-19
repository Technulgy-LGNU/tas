import { createRouter, createWebHashHistory } from 'vue-router'
import { checkSession } from '@/lib/auth'

const router = createRouter({
  history: createWebHashHistory(import.meta.env.BASE_URL),
  routes: [
    { path: '/', name: 'home', component: () => import('@/views/HomeView.vue') },
    { path: '/login', name: 'login', component: () => import('@/views/LoginView.vue'), meta: { public: true } },
    { path: '/:pathMatch(.*)*', name: '404', component: () => import('@/views/404View.vue') },
  ],
})

router.beforeEach(async (to) => {
  if (to.meta.public) return true
  try {
    if (await checkSession()) return true
    return { name: 'login', query: { returnTo: to.fullPath } }
  } catch {
    return { name: 'login', query: { error: 'unavailable', returnTo: to.fullPath } }
  }
})

export default router
