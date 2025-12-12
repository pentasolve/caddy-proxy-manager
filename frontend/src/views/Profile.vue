<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { toast } from 'vue-sonner'
import { authFetch } from '../composables/useAuth'

const username = ref('')
const currentPassword = ref('')
const password = ref('')
const confirmPassword = ref('')
const isLoading = ref(false)

const fetchProfile = async () => {
    const res = await authFetch('/api/profile')
    if (res.ok) {
        const data = await res.json()
        username.value = data.username
    }
}

const saveProfile = async () => {
    if (password.value && password.value !== confirmPassword.value) {
        toast.error('Passwords do not match')
        return
    }

    if (password.value && !currentPassword.value) {
        toast.error('Current password is required to change password')
        return
    }

    isLoading.value = true
    try {
        const res = await authFetch('/api/profile', {
            method: 'PUT',
            headers: { 
                'Content-Type': 'application/json'
            },
            body: JSON.stringify({
                username: username.value,
                current_password: currentPassword.value,
                password: password.value
            })
        })

        if (res.ok) {
            toast.success('Profile updated successfully')
            currentPassword.value = ''
            password.value = ''
            confirmPassword.value = ''
        } else {
            const data = await res.json()
            toast.error(data.error || 'Failed to update profile')
        }
    } catch (e) {
        toast.error('Connection error')
    } finally {
        isLoading.value = false
    }
}

onMounted(fetchProfile)
</script>

<template>
    <div class="max-w-2xl mx-auto">
        <!-- Header -->
        <div class="bg-white dark:bg-gray-800 rounded-t-xl shadow-sm border border-gray-200 dark:border-gray-700 overflow-hidden">
            <div class="h-1 bg-blue-900"></div>
            <div class="p-5">
                <div class="flex items-center gap-4">
                    <div class="w-14 h-14 rounded-xl bg-blue-900 flex items-center justify-center shadow-lg shadow-blue-900/25">
                        <svg xmlns="http://www.w3.org/2000/svg" class="h-7 w-7 text-white" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M16 7a4 4 0 11-8 0 4 4 0 018 0zM12 14a7 7 0 00-7 7h14a7 7 0 00-7-7z" />
                        </svg>
                    </div>
                    <div>
                        <h2 class="text-xl font-bold text-gray-800 dark:text-gray-100 tracking-tight">My Profile</h2>
                        <p class="text-sm text-gray-500 dark:text-gray-400 mt-0.5">Manage your account settings</p>
                    </div>
                </div>
            </div>
        </div>

        <div class="bg-white dark:bg-gray-800 rounded-b-xl shadow-lg border-x border-b border-gray-200 dark:border-gray-700">
            <form @submit.prevent="saveProfile" class="p-6 space-y-6">
                <!-- Username Section -->
                <div class="bg-white dark:bg-gray-700 rounded-xl border border-gray-200 dark:border-gray-600 p-5">
                    <div class="flex items-center gap-3 mb-4">
                        <div class="w-8 h-8 rounded-lg bg-blue-100 flex items-center justify-center">
                            <svg xmlns="http://www.w3.org/2000/svg" class="h-4 w-4 text-blue-900" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M16 7a4 4 0 11-8 0 4 4 0 018 0zM12 14a7 7 0 00-7 7h14a7 7 0 00-7-7z" />
                            </svg>
                        </div>
                        <span class="font-semibold text-gray-800 dark:text-gray-200">Account Information</span>
                    </div>
                    
                    <div>
                        <span class="block text-sm font-semibold text-gray-700 mb-2">Username</span>
                        <input 
                            v-model="username" 
                            type="text" 
                            class="w-full border border-gray-200 dark:border-gray-600 rounded-xl px-4 py-3 bg-gray-50/50 dark:bg-gray-700 text-gray-900 dark:text-gray-100 focus:ring-2 focus:ring-blue-900/20 focus:border-blue-900 focus:bg-white dark:focus:bg-gray-600 outline-none transition-all"
                        />
                    </div>
                </div>

                <!-- Password Section -->
                <div class="border-l-4 border-blue-900 dark:border-blue-600 bg-blue-50/50 dark:bg-blue-900/20 rounded-r-xl overflow-hidden">
                    <div class="px-5 py-4">
                        <div class="flex items-center gap-3 mb-4">
                            <div class="w-8 h-8 rounded-lg bg-blue-100 dark:bg-blue-900/30 flex items-center justify-center">
                                <svg xmlns="http://www.w3.org/2000/svg" class="h-4 w-4 text-blue-900 dark:text-blue-400" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 7a2 2 0 012 2m4 0a6 6 0 01-7.743 5.743L11 17H9v2H7v2H4a1 1 0 01-1-1v-2.586a1 1 0 01.293-.707l5.964-5.964A6 6 0 1121 9z" />
                                </svg>
                            </div>
                            <div>
                                <span class="font-semibold text-gray-800 dark:text-gray-200">Change Password</span>
                                <p class="text-xs text-gray-500 dark:text-gray-400 mt-0.5">Leave blank to keep current password</p>
                            </div>
                        </div>

                        <div class="space-y-4">
                            <div>
                                <span class="block text-sm font-semibold text-gray-700 dark:text-gray-300 mb-2">
                                    Current Password 
                                    <span class="text-red-500 dark:text-red-400" v-if="password">*</span>
                                </span>
                                <input 
                                    v-model="currentPassword" 
                                    type="password" 
                                    class="w-full border border-gray-200 dark:border-gray-600 rounded-xl px-4 py-3 bg-gray-50/50 dark:bg-gray-700 text-gray-900 dark:text-gray-100 focus:ring-2 focus:ring-blue-900/20 focus:border-blue-900 focus:bg-white dark:focus:bg-gray-600 outline-none transition-all"
                                    :placeholder="password ? 'Required to change password' : 'Only required if changing password'"
                                    :required="!!password"
                                />
                            </div>

                            <div class="grid grid-cols-1 sm:grid-cols-2 gap-4">
                                <div>
                                    <span class="block text-sm font-semibold text-gray-700 dark:text-gray-300 mb-2">New Password</span>
                                    <input 
                                        v-model="password" 
                                        type="password" 
                                        class="w-full border border-gray-200 dark:border-gray-600 rounded-xl px-4 py-3 bg-gray-50/50 dark:bg-gray-700 text-gray-900 dark:text-gray-100 focus:ring-2 focus:ring-blue-900/20 focus:border-blue-900 focus:bg-white dark:focus:bg-gray-600 outline-none transition-all"
                                        placeholder="Enter new password"
                                    />
                                </div>

                                <div>
                                    <span class="block text-sm font-semibold text-gray-700 dark:text-gray-300 mb-2">Confirm Password</span>
                                    <input 
                                        v-model="confirmPassword" 
                                        type="password" 
                                        class="w-full border border-gray-200 dark:border-gray-600 rounded-xl px-4 py-3 bg-gray-50/50 dark:bg-gray-700 text-gray-900 dark:text-gray-100 focus:ring-2 focus:ring-blue-900/20 focus:border-blue-900 focus:bg-white dark:focus:bg-gray-600 outline-none transition-all"
                                        placeholder="Confirm new password"
                                    />
                                </div>
                            </div>
                        </div>
                    </div>
                </div>

                <div class="flex items-center justify-end pt-4 border-t border-gray-100 dark:border-gray-700">
                    <button 
                        type="submit" 
                        :disabled="isLoading"
                        class="bg-blue-900 hover:bg-blue-800 text-white px-6 py-2.5 rounded-xl shadow-lg shadow-blue-900/25 transition-all font-semibold flex items-center gap-2 disabled:opacity-50"
                    >
                        <svg v-if="isLoading" class="animate-spin h-4 w-4 text-white" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24">
                            <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle>
                            <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
                        </svg>
                        <svg v-else xmlns="http://www.w3.org/2000/svg" class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 13l4 4L19 7" />
                        </svg>
                        <span>{{ isLoading ? 'Saving...' : 'Save Changes' }}</span>
                    </button>
                </div>
            </form>
        </div>
    </div>
</template>
