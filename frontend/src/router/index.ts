import { createRouter, createWebHistory } from 'vue-router'
import { useAuthStore } from '@/stores/auth'

const router = createRouter({
  history: createWebHistory(),
  routes: [
    {
      path: '/',
      name: 'splash',
      component: () => import('@/pages/splash/Splash.vue'),
      meta: { hideBottomNav: true }
    },
    {
      path: '/login',
      name: 'login',
      component: () => import('@/pages/auth/Login.vue'),
      meta: { guest: true, hideBottomNav: true }
    },
    {
      path: '/ukey',
      name: 'UkeyManager',
      component: () => import('../pages/profile/UkeyManager.vue'),
      meta: { requiresAuth: true, hideBottomNav: true }
    },
    {
      path: '/recurring',
      name: 'RecurringManager',
      component: () => import('../pages/profile/RecurringManager.vue'),
      meta: { requiresAuth: true, hideBottomNav: true }
    },
    {
      path: '/recurring/add',
      name: 'RecurringAdd',
      component: () => import('../pages/profile/RecurringAdd.vue'),
      meta: { requiresAuth: true, hideBottomNav: true }
    },
    {
      path: '/recurring/edit/:id',
      name: 'RecurringEdit',
      component: () => import('../pages/profile/RecurringAdd.vue'),
      meta: { requiresAuth: true, hideBottomNav: true }
    },
    {
      path: '/:pathMatch(.*)*',
      name: 'register',
      component: () => import('@/pages/auth/Register.vue'),
      meta: { guest: true, hideBottomNav: true }
    },
    {
      // 主页 (原来的是 dashboard)
      path: '/home',
      name: 'home',
      component: () => import('@/pages/home/Home.vue'),
      meta: { auth: true }
    },
    {
      // 报表页
      path: '/analytics',
      name: 'analytics',
      component: () => import('@/pages/analytics/Analytics.vue'),
      meta: { auth: true }
    },
    {
      // 我的主页
      path: '/profile',
      name: 'profile',
      component: () => import('@/pages/profile/Profile.vue'),
      meta: { auth: true }
    },
    {
      // 邮箱绑定页 (子页面)
      path: '/settings',
      name: 'settings',
      component: () => import('@/pages/settings/Settings.vue'),
      meta: { auth: true, hideBottomNav: true }
    },
    {
      // 账单明细流
      path: '/bills',
      name: 'bills',
      component: () => import('@/pages/bills/Bills.vue'),
      meta: { auth: true }
    },
    {
      // 账单独立编辑页
      path: '/bill/edit/:id',
      name: 'BillEdit',
      component: () => import('@/pages/bills/BillEdit.vue'),
      meta: { auth: true, hideBottomNav: true }
    },
    {
      // 手动记账页
      path: '/bill/add',
      name: 'BillAdd',
      component: () => import('@/pages/bills/BillAdd.vue'),
      meta: { auth: true, hideBottomNav: true }
    },
    {
      // 修改密码页 (子页面)
      path: '/password',
      name: 'password',
      component: () => import('@/pages/profile/Password.vue'),
      meta: { auth: true, hideBottomNav: true }
    },
    {
      // 管理员入口页
      path: '/admin',
      name: 'AdminIndex',
      component: () => import('@/pages/admin/AdminIndex.vue'),
      meta: { auth: true, requiresAdmin: true, hideBottomNav: true }
    },
    {
      // 管理员邀请码管理
      path: '/admin/invitation',
      name: 'AdminInvitation',
      component: () => import('@/pages/admin/Invitation.vue'),
      meta: { auth: true, requiresAdmin: true, hideBottomNav: true }
    },
    {
      // 管理员用户管理
      path: '/admin/users',
      name: 'AdminUsers',
      component: () => import('@/pages/admin/UserManagement.vue'),
      meta: { auth: true, requiresAdmin: true, hideBottomNav: true }
    },
    {
      // 管理层系统监控
      path: '/admin/stats',
      name: 'AdminStats',
      component: () => import('@/pages/admin/SystemStats.vue'),
      meta: { auth: true, requiresAdmin: true, hideBottomNav: true }
    }
  ]
})

// 路由守卫
router.beforeEach(async (to, _from, next) => {
  const auth = useAuthStore()
  
  if (to.meta.auth && !auth.isLoggedIn) {
    next('/login')
  } else if (to.meta.guest && auth.isLoggedIn) {
    // 登录页不需要看到，直接去 /
    next('/')
  } else if (to.meta.requiresAdmin) {
    // 管理员权限校验
    if (auth.user?.role !== 'admin') {
      next('/home') // 非管理员跳转到主页
    } else {
      next()
    }
  } else {
    next()
  }
})

export default router
