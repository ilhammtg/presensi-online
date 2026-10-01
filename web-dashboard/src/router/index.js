import { createRouter, createWebHistory } from 'vue-router'
import { useAuthStore } from '../stores/auth'
import LoginView from '../views/LoginView.vue'
import DosenSessionView from '../views/DosenSessionView.vue'
import AdminProdiView from '../views/AdminProdiView.vue'
import AdminSystemView from '../views/AdminSystemView.vue'

const routes = [
  {
    path: '/login',
    name: 'Login',
    component: LoginView,
    meta: { public: true }
  },
  {
    path: '/',
    name: 'Home',
    redirect: () => {
      const auth = useAuthStore()
      if (!auth.isAuthenticated) return '/login'
      if (auth.isDosen) return '/dosen'
      if (auth.isSuperAdmin) return '/admin/system'
      return '/admin/prodi'
    }
  },
  {
    path: '/dosen',
    name: 'Dosen',
    component: DosenSessionView,
    meta: { requiresAuth: true, role: 'dosen' }
  },
  {
    path: '/dosen/sesi/:sessionId',
    name: 'DosenSession',
    component: DosenSessionView,
    meta: { requiresAuth: true, role: 'dosen' }
  },
  {
    path: '/admin',
    name: 'Admin',
    redirect: () => {
      const auth = useAuthStore()
      if (auth.isSuperAdmin) return '/admin/system'
      return '/admin/prodi'
    }
  },
  {
    path: '/admin/prodi',
    name: 'AdminProdi',
    component: AdminProdiView,
    meta: { requiresAuth: true, roles: ['admin_prodi', 'superadmin'] }
  },
  {
    path: '/admin/system',
    name: 'AdminSystem',
    component: AdminSystemView,
    meta: { requiresAuth: true, roles: ['superadmin'] }
  }
]

const router = createRouter({
  history: createWebHistory(),
  routes
})

router.beforeEach((to, from, next) => {
  const auth = useAuthStore()

  if (to.meta.requiresAuth && !auth.isAuthenticated) {
    return next('/login')
  }

  if (to.path === '/login' && auth.isAuthenticated) {
    if (auth.isDosen) return next('/dosen')
    if (auth.isSuperAdmin) return next('/admin/system')
    return next('/admin/prodi')
  }

  // Dosen cannot access /admin
  if (to.path.startsWith('/admin') && auth.isDosen) {
    return next('/dosen')
  }

  // Admin Prodi CANNOT access /admin/system
  if (to.path.startsWith('/admin/system') && !auth.isSuperAdmin) {
    return next('/admin/prodi')
  }

  next()
})

export default router
