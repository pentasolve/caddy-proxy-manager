<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { toast } from 'vue-sonner'

const router = useRouter()
const username = ref('')
const password = ref('')
const confirmPassword = ref('')
const isLoading = ref(false)

onMounted(async () => {
    try {
        const res = await fetch('/api/setup')
        const data = await res.json()
        if (!data.setup_required) {
            router.push('/login')
        }
    } catch (e) {
        console.error(e)
    }
})

const handleSetup = async () => {
    if (!username.value || !password.value) {
        toast.error('Please fill in all fields')
        return
    }

    if (password.value !== confirmPassword.value) {
        toast.error('Passwords do not match')
        return
    }

    isLoading.value = true
    try {
        const res = await fetch('/api/setup', {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({
                username: username.value,
                password: password.value
            })
        })

        if (res.ok) {
            toast.success('Setup completed successfully! Please login.')
            router.push('/login')
        } else {
            const data = await res.json()
            toast.error(data.error || 'Setup failed')
        }
    } catch (e) {
        toast.error('Connection error')
    } finally {
        isLoading.value = false
    }
}
</script>

<template>
    <div class="min-h-screen flex items-center justify-center bg-gray-50 p-4">
        <!-- Background decoration -->
        <div class="fixed inset-0 overflow-hidden pointer-events-none">
            <div class="absolute -top-40 -right-40 w-80 h-80 bg-green-400/10 rounded-full blur-3xl"></div>
            <div class="absolute -bottom-40 -left-40 w-80 h-80 bg-emerald-400/10 rounded-full blur-3xl"></div>
        </div>

        <div class="relative bg-white rounded-2xl shadow-xl w-full max-w-md overflow-hidden border border-gray-100">
            <!-- Gradient accent bar -->
            <div class="h-1.5 bg-green-500"></div>
            
            <div class="p-8">
                <!-- Header -->
                <div class="text-center mb-8">
                    <!-- Logo -->
                    <div class="w-48 h-48 mx-auto mb-5">
                        <img src="../assets/logo.png" alt="Caddy Proxy Manager" class="w-full h-full object-contain drop-shadow-lg" />
                    </div>
                    <h1 class="text-2xl font-bold text-gray-900">Welcome to Caddy Proxy Manager</h1>
                    <p class="text-gray-500 mt-2 text-sm">Create your administrator account to get started</p>
                </div>

                <!-- Setup Form -->
                <form @submit.prevent="handleSetup" class="space-y-5">
                    <!-- Username Field -->
                    <div>
                        <label class="block text-sm font-medium text-gray-700 mb-1.5">Username <span class="text-red-500">*</span></label>
                        <div class="relative">
                            <div class="absolute inset-y-0 left-0 pl-3.5 flex items-center pointer-events-none">
                                <svg xmlns="http://www.w3.org/2000/svg" class="h-5 w-5 text-gray-400" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M16 7a4 4 0 11-8 0 4 4 0 018 0zM12 14a7 7 0 00-7 7h14a7 7 0 00-7-7z" />
                                </svg>
                            </div>
                            <input 
                                v-model="username" 
                                type="text" 
                                class="w-full border border-gray-300 rounded-xl pl-11 pr-4 py-2.5 focus:ring-2 focus:ring-green-500 focus:border-green-500 outline-none transition-all text-sm"
                                placeholder="Enter your username"
                                required
                            />
                        </div>
                    </div>

                    <!-- Password Field -->
                    <div>
                        <label class="block text-sm font-medium text-gray-700 mb-1.5">Password <span class="text-red-500">*</span></label>
                        <div class="relative">
                            <div class="absolute inset-y-0 left-0 pl-3.5 flex items-center pointer-events-none">
                                <svg xmlns="http://www.w3.org/2000/svg" class="h-5 w-5 text-gray-400" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 15v2m-6 4h12a2 2 0 002-2v-6a2 2 0 00-2-2H6a2 2 0 00-2 2v6a2 2 0 002 2zm10-10V7a4 4 0 00-8 0v4h8z" />
                                </svg>
                            </div>
                            <input 
                                v-model="password" 
                                type="password" 
                                class="w-full border border-gray-300 rounded-xl pl-11 pr-4 py-2.5 focus:ring-2 focus:ring-green-500 focus:border-green-500 outline-none transition-all text-sm"
                                placeholder="Create a strong password"
                                required
                            />
                        </div>
                    </div>

                    <!-- Confirm Password Field -->
                    <div>
                        <label class="block text-sm font-medium text-gray-700 mb-1.5">Confirm Password <span class="text-red-500">*</span></label>
                        <div class="relative">
                            <div class="absolute inset-y-0 left-0 pl-3.5 flex items-center pointer-events-none">
                                <svg xmlns="http://www.w3.org/2000/svg" class="h-5 w-5 text-gray-400" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 12l2 2 4-4m5.618-4.016A11.955 11.955 0 0112 2.944a11.955 11.955 0 01-8.618 3.04A12.02 12.02 0 003 9c0 5.591 3.824 10.29 9 11.622 5.176-1.332 9-6.03 9-11.622 0-1.042-.133-2.052-.382-3.016z" />
                                </svg>
                            </div>
                            <input 
                                v-model="confirmPassword" 
                                type="password" 
                                class="w-full border border-gray-300 rounded-xl pl-11 pr-4 py-2.5 focus:ring-2 focus:ring-green-500 focus:border-green-500 outline-none transition-all text-sm"
                                placeholder="Confirm your password"
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
                            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13 10V3L4 14h7v7l9-11h-7z" />
                        </svg>
                        <span>{{ isLoading ? 'Creating Account...' : 'Create Administrator' }}</span>
                    </button>
                </form>

                <!-- Security Note -->
                <div class="mt-6 pt-6 border-t border-gray-100">
                    <div class="flex items-start gap-3 text-xs text-gray-500">
                        <div class="flex-shrink-0 w-8 h-8 bg-green-50 rounded-lg flex items-center justify-center">
                            <svg xmlns="http://www.w3.org/2000/svg" class="h-4 w-4 text-green-600" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 12l2 2 4-4m5.618-4.016A11.955 11.955 0 0112 2.944a11.955 11.955 0 01-8.618 3.04A12.02 12.02 0 003 9c0 5.591 3.824 10.29 9 11.622 5.176-1.332 9-6.03 9-11.622 0-1.042-.133-2.052-.382-3.016z" />
                            </svg>
                        </div>
                        <p class="leading-relaxed">
                            <span class="font-medium text-gray-700">Security tip:</span> Use a strong password with at least 8 characters, including uppercase, lowercase, numbers, and symbols.
                        </p>
                    </div>
                </div>
            </div>
        </div>
    </div>
</template>
