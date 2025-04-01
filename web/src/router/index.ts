import { createRouter, createWebHashHistory } from 'vue-router'
import HomeView from '@/views/HomeView.vue'

const router = createRouter({
  history: createWebHashHistory(import.meta.env.BASE_URL),
  routes: [
    {
      path: '/',
      name: 'home',
      component: HomeView,
    },
    {
      path: '/login',
      name: 'login',
      component: () => import('@/views/LoginView.vue'),
    },
    {
      path: '/resetPassword',
      name: 'resetPassword',
      component: () => import('@/views/ResetPasswordView.vue'),
    },
    {
      path: '/members',
      name: 'members',
      component: () => import('@/views/MembersView.vue'),
    },
    {
      path: '/teams',
      name: 'teams',
      component: () => import('@/views/TeamsView.vue'),
    },
    {
      path: '/events',
      name: 'events',
      component: () => import('@/views/EventsView.vue'),
    },
    {
      path: '/newsletter',
      name: 'newsletter',
      component: () => import('@/views/NewsletterView.vue'),
    },
    {
      path: '/forms',
      name: 'forms',
      component: () => import('@/views/FormView.vue'),
    },
    {
      path: '/website',
      name: 'website',
      component: () => import('@/views/WebsiteView.vue'),
    },
    {
      path: '/orders',
      name: 'orders',
      component: () => import('@/views/OrdersView.vue'),
    },
    {
      path: '/sponsors',
      name: 'sponsors',
      component: () => import('@/views/SponsorsView.vue'),
    }
  ],
})

export default router
