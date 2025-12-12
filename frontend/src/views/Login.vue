<script setup lang="ts">
import { ref } from 'vue'
import { toast } from 'vue-sonner'
import { setTokens, startAutoRefresh } from '../composables/useAuth'
import { useTheme } from '../composables/useTheme'

const { currentTheme, toggleTheme } = useTheme()

const username = ref('')
const password = ref('')
const error = ref('')
const isLoading = ref(false)

const login = async () => {
  error.value = ''
  isLoading.value = true
  try {
    const res = await fetch('/api/auth/login', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        credentials: 'include',
        body: JSON.stringify({ username: username.value, password: password.value })
    })

    const data = await res.json()
    if (res.ok) {
        setTokens(data.access_token, data.expires_in)
        startAutoRefresh()
        toast.success('Login successful')
        window.location.href = '/'
    } else {
        error.value = data.error
        toast.error(data.error || 'Login failed')
    }
  } catch (e) {
    toast.error('Connection error')
  } finally {
    isLoading.value = false
  }
}
</script>

<template>
  <div class="min-h-screen flex items-center justify-center bg-gray-50 dark:bg-gray-900 p-4">
    <!-- Background decoration -->
    <div class="fixed inset-0 overflow-hidden pointer-events-none">
      <div class="absolute -top-40 -right-40 w-80 h-80 bg-green-400/10 dark:bg-green-400/5 rounded-full blur-3xl"></div>
      <div class="absolute -bottom-40 -left-40 w-80 h-80 bg-emerald-400/10 dark:bg-emerald-400/5 rounded-full blur-3xl"></div>
    </div>

    <div class="relative bg-white dark:bg-gray-800 rounded-2xl shadow-xl w-full max-w-md overflow-hidden border border-gray-100 dark:border-gray-700">
      <!-- Theme Toggle Button - Top Right Corner of Login Box -->
      <button 
        @click="toggleTheme" 
        class="absolute top-4 right-4 p-2.5 rounded-lg bg-gray-100 dark:bg-gray-700 hover:bg-gray-200 dark:hover:bg-gray-600 transition-all z-10"
        title="Toggle theme"
      >
        <svg v-if="currentTheme === 'dark'" xmlns="http://www.w3.org/2000/svg" class="h-5 w-5 text-yellow-500" fill="none" viewBox="0 0 24 24" stroke="currentColor">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 3v1m0 16v1m9-9h-1M4 12H3m15.364 6.364l-.707-.707M6.343 6.343l-.707-.707m12.728 0l-.707.707M6.343 17.657l-.707.707M16 12a4 4 0 11-8 0 4 4 0 018 0z" />
        </svg>
        <svg v-else xmlns="http://www.w3.org/2000/svg" class="h-5 w-5 text-gray-700" fill="none" viewBox="0 0 24 24" stroke="currentColor">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M20.354 15.354A9 9 0 018.646 3.646 9.003 9.003 0 0012 21a9.003 9.003 0 008.354-5.646z" />
        </svg>
      </button>
      <!-- Gradient accent bar -->
      <div class="h-1.5 bg-green-500"></div>
      
      <div class="p-8">
        <!-- Header with Logo -->
        <div class="text-center mb-8">
          <!-- Logo -->
          <div class="w-32 h-32 mx-auto mb-5">
            <img src="../assets/logo.png" alt="Caddy Proxy Manager" class="w-full h-full object-contain drop-shadow-lg" />
          </div>
          <h1 class="text-2xl font-bold text-gray-900 dark:text-gray-100">Welcome Back</h1>
          <p class="text-gray-500 dark:text-gray-400 mt-2 text-sm">Sign in to manage your proxy hosts</p>
        </div>
        
        <!-- Login Form -->
        <form @submit.prevent="login" class="space-y-5">
          <!-- Error Alert -->
          <div v-if="error" class="bg-red-50 dark:bg-red-900/20 border border-red-200 dark:border-red-800 text-red-700 dark:text-red-400 px-4 py-3 rounded-xl text-sm flex items-center gap-2">
            <svg xmlns="http://www.w3.org/2000/svg" class="h-5 w-5 flex-shrink-0" fill="none" viewBox="0 0 24 24" stroke="currentColor">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 8v4m0 4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
            </svg>
            <span>{{ error }}</span>
          </div>

          <!-- Username Field -->
          <div>
            <label class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1.5">Username</label>
            <div class="relative">
              <div class="absolute inset-y-0 left-0 pl-3.5 flex items-center pointer-events-none">
                <svg xmlns="http://www.w3.org/2000/svg" class="h-5 w-5 text-gray-400 dark:text-gray-500" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M16 7a4 4 0 11-8 0 4 4 0 018 0zM12 14a7 7 0 00-7 7h14a7 7 0 00-7-7z" />
                </svg>
              </div>
              <input 
                v-model="username" 
                type="text" 
                class="w-full border border-gray-300 dark:border-gray-600 rounded-xl pl-11 pr-4 py-2.5 focus:ring-2 focus:ring-green-500 focus:border-green-500 outline-none transition-all text-sm bg-white dark:bg-gray-700 dark:text-gray-100"
                placeholder="Enter your username"
                required
              />
            </div>
          </div>
          
          <!-- Password Field -->
          <div>
            <label class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1.5">Password</label>
            <div class="relative">
              <div class="absolute inset-y-0 left-0 pl-3.5 flex items-center pointer-events-none">
                <svg xmlns="http://www.w3.org/2000/svg" class="h-5 w-5 text-gray-400 dark:text-gray-500" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 15v2m-6 4h12a2 2 0 002-2v-6a2 2 0 00-2-2H6a2 2 0 00-2 2v6a2 2 0 002 2zm10-10V7a4 4 0 00-8 0v4h8z" />
                </svg>
              </div>
              <input 
                v-model="password" 
                type="password" 
                class="w-full border border-gray-300 dark:border-gray-600 rounded-xl pl-11 pr-4 py-2.5 focus:ring-2 focus:ring-green-500 focus:border-green-500 outline-none transition-all text-sm bg-white dark:bg-gray-700 dark:text-gray-100"
                placeholder="Enter your password"
                required
              />
            </div>
          </div>

          <!-- Submit Button -->
          <button 
            type="submit" 
            :disabled="isLoading"
            class="w-full bg-green-500 hover:bg-green-600 text-white font-semibold py-3 px-4 rounded-xl transition-all flex justify-center items-center gap-2 disabled:opacity-50 disabled:cursor-not-allowed shadow-lg shadow-green-500/25 hover:shadow-xl hover:shadow-green-500/30 mt-6"
          >
            <svg v-if="isLoading" class="animate-spin h-5 w-5 text-white" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24">
              <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle>
              <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
            </svg>
            <svg v-else xmlns="http://www.w3.org/2000/svg" class="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M11 16l-4-4m0 0l4-4m-4 4h14m-5 4v1a3 3 0 01-3 3H6a3 3 0 01-3-3V7a3 3 0 013-3h7a3 3 0 013 3v1" />
            </svg>
            <span>{{ isLoading ? 'Signing in...' : 'Sign In' }}</span>
          </button>
        </form>
        
        <!-- Footer -->
        <div class="mt-8 pt-6 border-t border-gray-100 dark:border-gray-700 text-center text-xs text-gray-400 dark:text-gray-500">
          &copy; 2025 Caddy Proxy Manager
        </div>
      </div>
    </div>
  </div>
</template>
