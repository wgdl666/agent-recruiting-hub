import { createRouter, createWebHistory } from 'vue-router'

const router = createRouter({
  history: createWebHistory(),
  routes: [
    { path: '/', name: 'app', component: () => import('../layouts/TabShell.vue') },
    { path: '/candidate/:id', redirect: (to) => ({ path: '/', query: { tab: `candidate-${to.params.id}` } }) },
    { path: '/upload', redirect: { path: '/', query: { tab: 'upload' } } },
    { path: '/docs', redirect: { path: '/', query: { tab: 'docs' } } },
  ],
})

export default router
