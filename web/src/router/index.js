import { createRouter, createWebHistory } from 'vue-router'
import LoginView from '@/views/LoginView.vue'
import HomeView from '@/views/HomeView.vue'
import PasswordReset from '@/components/PasswordReset.vue'
import PasswordResetView from '@/views/PasswordResetView.vue'

const routes = [
  {
    path: '/',
    name: 'Home',
    component: HomeView
  },
  {
    path: '/login',
    name: 'Login',
    component: LoginView
  },
  {
    path: '/mon',
    name: 'Monitor',
    redirect: '/monitor'
  },
  {
    path: '/passwordreset',
    name: 'PasswordReset',
    component: PasswordResetView
  },
]

const router = createRouter({
  history: createWebHistory(import.meta.env.BASE_URL),
  routes,

  scrollBehavior(to ) {
    if (to.hash) {
      return { selector: to.hash };
    } else {
      return { x: 0, y: 0 };
    }
  },
});

export default router
