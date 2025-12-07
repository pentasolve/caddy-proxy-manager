<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { toast } from 'vue-sonner'
import { authFetch } from '../composables/useAuth'

const defaultPageHtml = ref('')
const isLoading = ref(false)

const fetchDefaultPage = async () => {
    const res = await authFetch('/api/settings/default-page')
    if (res.ok) {
        const data = await res.json()
        defaultPageHtml.value = data.html
    }
}

const saveDefaultPage = async () => {
    isLoading.value = true
    const res = await authFetch('/api/settings/default-page', {
        method: 'PUT',
        headers: { 
            'Content-Type': 'application/json'
        },
        body: JSON.stringify({ html: defaultPageHtml.value })
    })

    isLoading.value = false
    if (res.ok) {
        toast.success('Default page updated successfully')
    } else {
        toast.error('Failed to update default page')
    }
}

onMounted(fetchDefaultPage)
</script>

<template>
    <div>
        <!-- Header -->
        <div class="bg-white rounded-t-xl p-5 border-b border-gray-200 shadow-sm overflow-hidden relative">
            <!-- Gradient accent bar -->
            <div class="absolute top-0 left-0 right-0 h-1 bg-gray-500"></div>
            
            <div class="flex flex-col sm:flex-row justify-between items-start sm:items-center gap-4">
                <div>
                    <h2 class="text-2xl font-bold text-gray-800 tracking-tight">Settings</h2>
                    <p class="text-sm text-gray-500 mt-1">Configure your Caddy Proxy Manager instance</p>
                </div>
            </div>
        </div>

        <div class="bg-white rounded-b-xl shadow-lg border-x border-b border-gray-200">
            <div class="p-6">
                <!-- Section Card -->
                <div class="bg-white rounded-xl border border-gray-200 overflow-hidden shadow-sm">
                    <!-- Section Header -->
                    <div class="px-6 py-4 bg-gray-50 border-b border-gray-200">
                        <div class="flex items-center gap-3">
                            <div class="w-10 h-10 rounded-xl bg-gray-600 flex items-center justify-center shadow-lg shadow-gray-600/20">
                                <svg xmlns="http://www.w3.org/2000/svg" class="h-5 w-5 text-white" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M10 20l4-16m4 4l4 4-4 4M6 16l-4-4 4-4" />
                                </svg>
                            </div>
                            <div>
                                <h3 class="text-lg font-bold text-gray-800">Default Page Configuration</h3>
                                <p class="text-sm text-gray-500 mt-0.5">HTML content served on port 80 when no proxy matches</p>
                            </div>
                        </div>
                    </div>
                    
                    <!-- Section Content -->
                    <div class="p-6">
                        <div class="mb-5">
                            <span class="block text-sm font-semibold text-gray-700 mb-2">HTML Content</span>
                            <div class="relative">
                                <textarea 
                                    v-model="defaultPageHtml" 
                                    rows="16" 
                                    class="w-full font-mono text-sm border border-gray-200 rounded-xl p-4 bg-gray-50/50 focus:ring-2 focus:ring-gray-500/20 focus:border-gray-500 focus:bg-white outline-none transition-all resize-y min-h-[300px]"
                                    placeholder="<!DOCTYPE html>
<html lang='en'>
<head>
    <meta charset='UTF-8'>
    <title>Welcome</title>
</head>
<body>
    <h1>Welcome to Caddy Proxy Manager</h1>
</body>
</html>"
                                ></textarea>
                                <div class="absolute bottom-3 right-3 flex items-center gap-2 text-xs text-gray-400">
                                    <svg xmlns="http://www.w3.org/2000/svg" class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                                        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M10 20l4-16m4 4l4 4-4 4M6 16l-4-4 4-4" />
                                    </svg>
                                    HTML
                                </div>
                            </div>
                        </div>

                        <div class="flex items-center justify-between pt-4 border-t border-gray-100">
                            <p class="text-xs text-gray-400">
                                <svg xmlns="http://www.w3.org/2000/svg" class="h-4 w-4 inline mr-1" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13 16h-1v-4h-1m1-4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
                                </svg>
                                Changes are applied immediately after saving
                            </p>
                            <button 
                                @click="saveDefaultPage" 
                                :disabled="isLoading"
                                class="bg-gray-700 hover:bg-gray-800 text-white px-6 py-2.5 rounded-xl shadow-lg shadow-gray-700/25 transition-all font-semibold disabled:opacity-50 disabled:cursor-not-allowed flex items-center gap-2"
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
                    </div>
                </div>
            </div>
        </div>
    </div>
</template>

<style scoped>
/* Page fade in animation */
div {
    animation: fadeInUp 0.3s cubic-bezier(0.22, 1, 0.36, 1) forwards;
}

@keyframes fadeInUp {
    from {
        opacity: 0;
        transform: translateY(10px);
    }
    to {
        opacity: 1;
        transform: translateY(0);
    }
}

/* Section card animation */
.rounded-xl.border {
    animation: scaleIn 0.3s cubic-bezier(0.22, 1, 0.36, 1) forwards;
    animation-delay: 100ms;
}

@keyframes scaleIn {
    from {
        opacity: 0;
        transform: scale(0.98);
    }
    to {
        opacity: 1;
        transform: scale(1);
    }
}

/* Textarea focus animation */
textarea {
    transition: all 0.2s cubic-bezier(0.22, 1, 0.36, 1);
}

textarea:focus {
    transform: scale(1.005);
}
</style>
