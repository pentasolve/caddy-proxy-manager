import { createApp } from 'vue'
import './style.css'
import App from './App.vue'
import router from './router'
import { toast } from 'vue-sonner'

const app = createApp(App)

const originalFetch = window.fetch
window.fetch = async (...args) => {
    try {
        const response = await originalFetch(...args)
        if (response.status === 401) {
            const currentPath = router.currentRoute.value.path
            const isAuthPage = currentPath === '/login' || currentPath === '/setup'
            const url = args[0] instanceof Request ? args[0].url : String(args[0])
            const isAuthEndpoint = url.includes('/api/setup') || url.includes('/api/auth/')

            if (!isAuthPage && !isAuthEndpoint) {
                localStorage.removeItem('token')
                toast.error('Session expired. Please login again.')
                router.push('/login')
            }
        }
        return response
    } catch (error) {
        throw error
    }
}

app.use(router).mount('#app')
