import { createRouter, createWebHashHistory } from 'vue-router'
import { checkSession } from '@/lib/auth'

const router = createRouter({
  history: createWebHashHistory(import.meta.env.BASE_URL),
  routes: [
    { path: '/', name: 'home', component: () => import('@/views/HomeView.vue') },
    {
      path: '/website',
      component: () => import('@/views/WebsiteLayout.vue'),
      redirect: '/website/home',
      children: [
        ...[
          ['home', 'home'],
          ['teams', 'teams'],
          ['participation-history', 'events'],
          ['sponsors', 'sponsors'],
          ['publications', 'publications'],
          ['blog', 'blog'],
        ].map(([path, kind]) => ({
          path: path!,
          name: `website-${kind}`,
          component: () => import('@/views/WebsiteEditorView.vue'),
          props: { kind },
        })),
        { path: 'ssl', name: 'website-ssl', component: () => import('@/views/WebsiteSSLView.vue') },
      ],
    },
    { path: '/images', name: 'images', component: () => import('@/views/ImagesView.vue') },
    {
      path: '/login',
      name: 'login',
      component: () => import('@/views/LoginView.vue'),
      meta: { public: true },
    },
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
