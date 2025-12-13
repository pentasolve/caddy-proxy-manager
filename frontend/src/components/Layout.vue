<script setup lang="ts">
import { ref, onMounted, onUnmounted } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { useWebSocket } from '../composables/useWebSocket'
import { getValidToken, logout as authLogout, startAutoRefresh, stopAutoRefresh, initAuth } from '../composables/useAuth'
import { useTheme } from '../composables/useTheme'

const router = useRouter()
const route = useRoute()
const isDropdownOpen = ref(false)
const isMobileMenuOpen = ref(false)
const userProfile = ref<any>(null)
const { connect, isConnected } = useWebSocket()
const { currentTheme, toggleTheme, initTheme } = useTheme()

onMounted(async () => {
    initTheme()
    
    const hasSession = await initAuth()
    if (!hasSession) {
        router.push('/login')
        return
    }
    
    startAutoRefresh()
    
    connect()
    
    const token = await getValidToken()
    if (token) {
        try {
            const res = await fetch('/api/profile', {
                headers: { 'Authorization': `Bearer ${token}` }
            })
            if (res.ok) {
                userProfile.value = await res.json()
            }
        } catch (e) {
            console.error(e)
        }
    }
})

onUnmounted(() => {
    stopAutoRefresh()
})

const menuItems = [
  { name: 'Dashboard', path: '/', icon: 'M4 6a2 2 0 012-2h2a2 2 0 012 2v2a2 2 0 01-2 2H6a2 2 0 01-2-2V6zM14 6a2 2 0 012-2h2a2 2 0 012 2v2a2 2 0 01-2 2h-2a2 2 0 01-2-2V6zM4 16a2 2 0 012-2h2a2 2 0 012 2v2a2 2 0 01-2 2H6a2 2 0 01-2-2v-2zM14 16a2 2 0 012-2h2a2 2 0 012 2v2a2 2 0 01-2 2h-2a2 2 0 01-2-2v-2z' },
  { name: 'Hosts', path: '/hosts', icon: 'M5 12h14M5 12a2 2 0 01-2-2V6a2 2 0 012-2h14a2 2 0 012 2v4a2 2 0 01-2 2M5 12a2 2 0 00-2 2v4a2 2 0 002 2h14a2 2 0 002-2v-4a2 2 0 00-2-2m-2-4h.01M17 16h.01' },
  { name: 'Streams', path: '/streams', icon: 'M7.5 21L3 16.5m0 0L7.5 12M3 16.5h13.5m0-13.5L21 7.5m0 0L16.5 12M21 7.5H7.5' },
  { name: 'Certificates', path: '/certificates', icon: 'M9 12l2 2 4-4m5.618-4.016A11.955 11.955 0 0112 2.944a11.955 11.955 0 01-8.618 3.04A12.02 12.02 0 003 9c0 5.591 3.824 10.29 9 11.622 5.176-1.332 9-6.03 9-11.622 0-1.042-.133-2.052-.382-3.016z' },
  { name: 'Access Lists', path: '/access-lists', icon: 'M12 15v2m-6 4h12a2 2 0 002-2v-6a2 2 0 00-2-2H6a2 2 0 00-2 2v6a2 2 0 002 2zm10-10V7a4 4 0 00-8 0v4h8z' },
  { name: 'Users', path: '/users', icon: 'M12 4.354a4 4 0 110 5.292M15 21H3v-1a6 6 0 0112 0v1zm0 0h6v-1a6 6 0 00-9-5.197M13 7a4 4 0 11-8 0 4 4 0 018 0z' },
  { name: 'Audit Logs', path: '/audit-logs', icon: 'M9 5H7a2 2 0 00-2 2v12a2 2 0 002 2h10a2 2 0 002-2V7a2 2 0 00-2-2h-2M9 5a2 2 0 002 2h2a2 2 0 002-2M9 5a2 2 0 012-2h2a2 2 0 012 2m-3 7h3m-3 4h3m-6-4h.01M9 16h.01' },
  { name: 'Settings', path: '/settings', icon: 'M10.325 4.317c.426-1.756 2.924-1.756 3.35 0a1.724 1.724 0 002.573 1.066c1.543-.94 3.31.826 2.37 2.37a1.724 1.724 0 001.065 2.572c1.756.426 1.756 2.924 0 3.35a1.724 1.724 0 00-1.066 2.573c.94 1.543-.826 3.31-2.37 2.37a1.724 1.724 0 00-2.572 1.065c-.426 1.756-2.924 1.756-3.35 0a1.724 1.724 0 00-2.573-1.066c-1.543.94-3.31-.826-2.37-2.37a1.724 1.724 0 00-1.065-2.572c-1.756-.426-1.756-2.924 0-3.35a1.724 1.724 0 001.066-2.573c-.94-1.543.826-3.31 2.37-2.37.996.608 2.296.07 2.572-1.065z M15 12a3 3 0 11-6 0 3 3 0 016 0z' },
]

const logout = async () => {
    await authLogout()
}

const isActive = (path: string) => {
    if (path === '/' && route.path === '/') return true
    if (path !== '/' && route.path.startsWith(path)) return true
    return false
}
</script>

<template>
  <div class="min-h-screen bg-gray-100 font-sans flex flex-col">
    <!-- Header -->
    <header class="bg-slate-950 text-white shadow-xl border-b border-green-500/20 relative z-50">
        <!-- Subtle background glow -->
        <div class="absolute top-0 left-0 w-full h-full bg-gradient-to-r from-green-500/5 via-transparent to-blue-500/5 pointer-events-none overflow-hidden"></div>
        <div class="container mx-auto px-4 h-16 flex items-center justify-between">
            <!-- Logo & Nav -->
            <div class="flex items-center gap-4 xl:gap-8">
                <div class="flex items-center gap-2 sm:gap-3">
                    <img src="../assets/logo.png" alt="Logo" class="w-10 h-10 sm:w-14 sm:h-14 object-contain filter drop-shadow-sm" />
                    <div class="hidden sm:flex flex-col">
                        <h1 class="text-base font-bold leading-none text-white">Caddy</h1>
                        <span class="text-xs font-medium text-green-400 tracking-wide uppercase">Proxy Manager</span>
                    </div>
                </div>

                <nav class="hidden xl:flex gap-1">
                    <router-link 
                        v-for="item in menuItems" 
                        :key="item.name"
                        :to="item.path"
                        class="px-3 py-2 rounded-md text-sm font-medium transition-all duration-200 flex items-center gap-2"
                        :class="isActive(item.path) ? 'bg-gray-800 text-white shadow-sm' : 'text-gray-400 hover:text-white hover:bg-gray-800'"
                    >
                        <svg xmlns="http://www.w3.org/2000/svg" class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" :d="item.icon" />
                        </svg>
                        {{ item.name }}
                    </router-link>
                </nav>
            </div>

            <!-- User Actions -->
            <!-- User Actions & Mobile Toggle -->
            <div class="flex items-center gap-4">
                <!-- Theme Toggle Button -->
                <button 
                    @click="toggleTheme"
                    class="p-2 rounded-lg hover:bg-gray-800 transition-colors text-gray-400 hover:text-white focus:outline-none"
                    :title="currentTheme === 'dark' ? 'Switch to light mode' : 'Switch to dark mode'"
                >
                    <!-- Sun icon for light mode (shown when dark) -->
                    <svg v-if="currentTheme === 'dark'" xmlns="http://www.w3.org/2000/svg" class="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 3v1m0 16v1m9-9h-1M4 12H3m15.364 6.364l-.707-.707M6.343 6.343l-.707-.707m12.728 0l-.707.707M6.343 17.657l-.707.707M16 12a4 4 0 11-8 0 4 4 0 018 0z" />
                    </svg>
                    <!-- Moon icon for dark mode (shown when light) -->
                    <svg v-else xmlns="http://www.w3.org/2000/svg" class="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M20.354 15.354A9 9 0 018.646 3.646 9.003 9.003 0 0012 21a9.003 9.003 0 008.354-5.646z" />
                    </svg>
                </button>
                
                <div class="relative">
                    <button 
                        @click.stop="isDropdownOpen = !isDropdownOpen"
                        class="flex items-center gap-3 hover:bg-gray-800 px-3 py-1.5 rounded-lg transition-colors group focus:outline-none relative z-50"
                    >
                        <div class="w-10 h-10 rounded-full bg-gray-700 flex items-center justify-center text-sm font-bold text-gray-300 group-hover:bg-gray-600 group-hover:text-white transition-colors border-2 border-gray-600">
                            {{ userProfile?.username?.charAt(0).toUpperCase() || 'A' }}
                        </div>
                        <div class="text-left hidden lg:block">
                            <div class="text-sm font-bold text-gray-200 group-hover:text-white leading-tight">
                                {{ userProfile?.username || 'Loading...' }}
                            </div>
                            <div class="text-xs text-gray-400 group-hover:text-gray-300 font-medium">
                                {{ userProfile?.role?.name || 'User' }}
                            </div>
                        </div>
                    </button>

                    <!-- Backdrop to close dropdown -->
                    <div v-if="isDropdownOpen" @click.stop="isDropdownOpen = false" class="fixed inset-0 z-[60] cursor-default"></div>
                    
                    <!-- Dropdown Menu -->
                    <div 
                        v-if="isDropdownOpen"
                        class="absolute right-0 mt-2 w-48 bg-white rounded-md shadow-lg py-1 z-[70] ring-1 ring-black ring-opacity-5 origin-top-right"
                    >
                        <router-link 
                            to="/profile" 
                            class="block px-4 py-2 text-sm text-gray-700 hover:bg-gray-100 flex items-center gap-2"
                            @click="isDropdownOpen = false"
                        >
                            <svg xmlns="http://www.w3.org/2000/svg" class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M16 7a4 4 0 11-8 0 4 4 0 018 0zM12 14a7 7 0 00-7 7h14a7 7 0 00-7-7z" />
                            </svg>
                            Profile
                        </router-link>
                        <button 
                            @click="logout" 
                            class="block w-full text-left px-4 py-2 text-sm text-red-600 hover:bg-red-50 flex items-center gap-2"
                        >
                            <svg xmlns="http://www.w3.org/2000/svg" class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M17 16l4-4m0 0l-4-4m4 4H7m6 4v1a3 3 0 01-3 3H6a3 3 0 01-3-3V7a3 3 0 013-3h4a3 3 0 013 3v1" />
                            </svg>
                            Logout
                        </button>
                    </div>
                </div>

                <!-- Mobile Menu Button -->
                <button 
                    @click="isMobileMenuOpen = !isMobileMenuOpen"
                    class="xl:hidden p-2 rounded-md text-gray-400 hover:text-white hover:bg-gray-800 focus:outline-none"
                >
                    <svg xmlns="http://www.w3.org/2000/svg" class="h-6 w-6" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                        <path v-if="!isMobileMenuOpen" stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 6h16M4 12h16M4 18h16" />
                        <path v-else stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
                    </svg>
                </button>
            </div>
        </div>

        <!-- Mobile Menu -->
        <div v-show="isMobileMenuOpen" class="xl:hidden border-t border-gray-800">
            <div class="px-2 pt-2 pb-3 space-y-1">
                <router-link 
                    v-for="item in menuItems" 
                    :key="item.name"
                    :to="item.path"
                    class="block px-3 py-2 rounded-md text-base font-medium flex items-center gap-3 transition-colors"
                    :class="isActive(item.path) ? 'bg-gray-800 text-white' : 'text-gray-300 hover:text-white hover:bg-gray-700'"
                    @click="isMobileMenuOpen = false"
                >
                    <svg xmlns="http://www.w3.org/2000/svg" class="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" :d="item.icon" />
                    </svg>
                    {{ item.name }}
                </router-link>
            </div>
        </div>
    </header>

    <!-- Main Content -->
    <main class="flex-1 container mx-auto px-4 py-8">
      <slot></slot>
    </main>
  </div>
</template>
