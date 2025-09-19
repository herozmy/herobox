import { createRouter, createWebHistory } from 'vue-router'

const routes = [
  {
    path: '/',
    name: 'Layout',
    component: () => import('../views/Layout.vue'),
    redirect: '/dashboard',
    children: [
      {
        path: '/dashboard',
        name: 'Dashboard',
        component: () => import('../views/Dashboard.vue'),
        meta: { title: '仪表板' }
      },
      {
        path: '/sing-box',
        name: 'SingBoxManage',
        component: () => import('../views/SingBoxManage.vue'),
        meta: { title: 'Sing-Box管理' }
      },
      {
        path: '/sing-box/inbound',
        name: 'SingBoxInbound',
        component: () => import('../views/SingBoxInbound.vue'),
        meta: { title: '入站设置' }
      },
      {
        path: '/sing-box/proxy',
        name: 'SingBoxProxy',
        component: () => import('../views/SingBoxProxy.vue'),
        meta: { title: '代理/出站' }
      },
      {
        path: '/sing-box/rules',
        name: 'SingBoxRules',
        component: () => import('../views/SingBoxRules.vue'),
        meta: { title: '规则设置' }
      },
      {
        path: '/mosdns',
        name: 'MosdnsManage',
        component: () => import('../views/MosdnsManage.vue'),
        meta: { title: 'MosDNS管理' }
      },
      {
        path: '/logs',
        name: 'Logs',
        component: () => import('../views/Logs.vue'),
        meta: { title: '日志查看' }
      }
    ]
  }
]

const router = createRouter({
  history: createWebHistory(),
  routes
})

export default router
