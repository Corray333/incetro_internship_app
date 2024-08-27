import { createRouter, createWebHistory } from 'vue-router'
declare const Telegram: any
import { renewTokens } from '@/utils/helpers'

const router = createRouter({
  history: createWebHistory(import.meta.env.BASE_URL),
  routes: [
    {
      path: '/',
      name: 'home',
      // route level code-splitting
      // this generates a separate chunk (About.[hash].js) for this route
      // which is lazy-loaded when the route is visited.
      component: () => import('../views/HomeView.vue')
    },
    {
      path: '/tasks/:task_id',
      name: 'task',
      // route level code-splitting
      // this generates a separate chunk (About.[hash].js) for this route
      // which is lazy-loaded when the route is visited.
      component: () => import('../views/TaskView.vue')
    }
  ]
})

router.beforeEach(async (to, from, next) => {
  const tg = Telegram.WebApp
	var BackButton = Telegram.WebApp.BackButton
	BackButton.show() 

  next()
})

export default router
