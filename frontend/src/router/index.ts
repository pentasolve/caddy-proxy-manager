import { createRouter, createWebHistory } from 'vue-router'
import Login from '../views/Login.vue'
import Dashboard from '../views/Dashboard.vue'
import { getValidToken, initAuth } from '../composables/useAuth'

const router = createRouter({
    history: createWebHistory(import.meta.env.BASE_URL),
    routes: [
        {
            path: '/login',
            name: 'login',
            component: Login
        },
        {
            path: '/',
            name: 'dashboard',
            component: Dashboard,
            meta: { requiresAuth: true }
        },
        {
            path: '/hosts',
            name: 'proxy-hosts',
            component: () => import('../views/ProxyHosts.vue'),
            meta: { requiresAuth: true }
        },
        {
            path: '/streams',
            name: 'streams',
            component: () => import('../views/Streams.vue'),
            meta: { requiresAuth: true }
        },
        {
            path: '/certificates',
            name: 'certificates',
            component: () => import('../views/Certificates.vue'),
            meta: { requiresAuth: true }
        },
        {
            path: '/access-lists',
            name: 'access-lists',
            component: () => import('../views/AccessLists.vue'),
            meta: { requiresAuth: true }
        },
        {
            path: '/settings',
            name: 'settings',
            component: () => import('../views/Settings.vue'),
            meta: { requiresAuth: true }
        },
        {
            path: '/setup',
            name: 'setup',
            component: () => import('../views/Setup.vue')
        },
        {
            path: '/users',
            name: 'users',
            component: () => import('../views/Users.vue'),
            meta: { requiresAuth: true }
        },
        {
            path: '/profile',
            name: 'profile',
            component: () => import('../views/Profile.vue'),
            meta: { requiresAuth: true }
        },
        {
            path: '/audit-logs',
            name: 'audit-logs',
            component: () => import('../views/AuditLogs.vue'),
            meta: { requiresAuth: true }
        }
    ]
})

router.beforeEach(async (to, _from, next) => {
    if (to.name !== 'setup') {
        try {
            const res = await fetch('/api/setup')
            const data = await res.json()
            if (data.setup_required) {
                next('/setup')
                return
            }
        } catch (e) {
        }
    }

    if (to.meta.requiresAuth) {
        let token = await getValidToken()

        if (!token) {
            const hasSession = await initAuth()
            if (hasSession) {
                token = await getValidToken()
            }
        }

        if (!token) {
            next('/login')
            return
        }
    }

    next()
})

export default router
