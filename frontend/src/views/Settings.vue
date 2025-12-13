<script setup lang="ts">
import { ref, onMounted, nextTick } from 'vue'
import { toast } from 'vue-sonner'
import { useRoute } from 'vue-router'
import { authFetch } from '../composables/useAuth'

const route = useRoute()

const activeTab = ref('general')
const defaultPageHtml = ref('')
const isLoading = ref(false)

const zerosslEabKid = ref('')
const zerosslEabHmacKey = ref('')
const zerosslEabConfigured = ref(false)
const isLoadingEab = ref(false)

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
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ html: defaultPageHtml.value })
    })
    isLoading.value = false
    if (res.ok) {
        toast.success('Default page updated successfully')
    } else {
        toast.error('Failed to update default page')
    }
}

const fetchZeroSSLEAB = async () => {
    const res = await authFetch('/api/settings/zerossl-eab')
    if (res.ok) {
        const data = await res.json()
        zerosslEabKid.value = data.kid || ''
        zerosslEabHmacKey.value = ''
        zerosslEabConfigured.value = data.configured || false
    }
}

const saveZeroSSLEAB = async () => {
    isLoadingEab.value = true
    const res = await authFetch('/api/settings/zerossl-eab', {
        method: 'PUT',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ kid: zerosslEabKid.value, hmac_key: zerosslEabHmacKey.value })
    })
    isLoadingEab.value = false
    if (res.ok) {
        toast.success('ZeroSSL EAB credentials saved successfully')
        zerosslEabConfigured.value = zerosslEabKid.value !== '' && zerosslEabHmacKey.value !== ''
        zerosslEabHmacKey.value = ''
        await fetchZeroSSLEAB()
    } else {
        toast.error('Failed to save ZeroSSL EAB credentials')
    }
}

onMounted(async () => {
    fetchDefaultPage()
    fetchZeroSSLEAB()

    if (route.query.tab) {
        activeTab.value = route.query.tab as string
    }

    if (route.query.focus) {
        await nextTick()
        setTimeout(() => {
            const el = document.getElementById(route.query.focus as string)
            if (el) {
                el.scrollIntoView({ behavior: 'smooth', block: 'center' })
                el.classList.add('ring-2', 'ring-amber-500', 'ring-offset-2')
                setTimeout(() => el.classList.remove('ring-2', 'ring-amber-500', 'ring-offset-2'), 2000)
            }
        }, 300)
    }
})
</script>

<template>
    <div>
        <!-- Header -->
        <div class="bg-white rounded-t-xl p-5 border-b border-gray-200 dark:border-gray-700 shadow-sm overflow-hidden relative">
            <div class="absolute top-0 left-0 right-0 h-1 bg-gray-500"></div>
            <div class="flex flex-col sm:flex-row justify-between items-start sm:items-center gap-4">
                <div>
                    <h2 class="text-2xl font-bold text-gray-800 dark:text-gray-100 tracking-tight">Settings</h2>
                    <p class="text-sm text-gray-500 dark:text-gray-400 mt-1">Configure your Caddy Proxy Manager instance</p>
                </div>
            </div>
        </div>

        <div class="bg-white dark:bg-gray-800 rounded-b-xl shadow-lg border-x border-b border-gray-200 dark:border-gray-700">
            <!-- Tabs -->
            <div class="border-b border-gray-200 dark:border-gray-700">
                <nav class="flex px-6 -mb-px">
                    <button 
                        @click="activeTab = 'general'"
                        :class="[
                            'py-4 px-4 text-sm font-medium border-b-2 transition-colors',
                            activeTab === 'general' 
                                ? 'border-gray-700 text-gray-700 dark:text-gray-200 dark:text-gray-300' 
                                : 'border-transparent text-gray-500 dark:text-gray-400 hover:text-gray-700 hover:border-gray-300 dark:border-gray-600'
                        ]"
                    >
                        <svg xmlns="http://www.w3.org/2000/svg" class="h-4 w-4 inline mr-2" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M10.325 4.317c.426-1.756 2.924-1.756 3.35 0a1.724 1.724 0 002.573 1.066c1.543-.94 3.31.826 2.37 2.37a1.724 1.724 0 001.065 2.572c1.756.426 1.756 2.924 0 3.35a1.724 1.724 0 00-1.066 2.573c.94 1.543-.826 3.31-2.37 2.37a1.724 1.724 0 00-2.572 1.065c-.426 1.756-2.924 1.756-3.35 0a1.724 1.724 0 00-2.573-1.066c-1.543.94-3.31-.826-2.37-2.37a1.724 1.724 0 00-1.065-2.572c-1.756-.426-1.756-2.924 0-3.35a1.724 1.724 0 001.066-2.573c-.94-1.543.826-3.31 2.37-2.37.996.608 2.296.07 2.572-1.065z" />
                            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 12a3 3 0 11-6 0 3 3 0 016 0z" />
                        </svg>
                        General
                    </button>
                    <button 
                        @click="activeTab = 'dns'"
                        :class="[
                            'py-4 px-4 text-sm font-medium border-b-2 transition-colors flex items-center gap-2',
                            activeTab === 'dns' 
                                ? 'border-amber-600 text-amber-600' 
                                : 'border-transparent text-gray-500 dark:text-gray-400 hover:text-gray-700 hover:border-gray-300 dark:border-gray-600'
                        ]"
                    >
                        <svg xmlns="http://www.w3.org/2000/svg" class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M21 12a9 9 0 01-9 9m9-9a9 9 0 00-9-9m9 9H3m9 9a9 9 0 01-9-9m9 9c1.657 0 3-4.03 3-9s-1.343-9-3-9m0 18c-1.657 0-3-4.03-3-9s1.343-9 3-9m-9 9a9 9 0 019-9" />
                        </svg>
                        DNS
                        <span v-if="!zerosslEabConfigured" class="w-2 h-2 bg-amber-500 rounded-full"></span>
                    </button>
                </nav>
            </div>

            <div class="p-6">
                <!-- General Tab -->
                <div v-show="activeTab === 'general'">
                    <div class="bg-white dark:bg-gray-800 rounded-xl border border-gray-200 dark:border-gray-700 overflow-hidden shadow-sm">
                        <div class="px-6 py-4 bg-gray-50 dark:bg-gray-700 border-b border-gray-200 dark:border-gray-700">
                            <div class="flex items-center gap-3">
                                <div class="w-10 h-10 rounded-xl bg-gray-600 flex items-center justify-center shadow-lg shadow-gray-600/20">
                                    <svg xmlns="http://www.w3.org/2000/svg" class="h-5 w-5 text-white" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                                        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M10 20l4-16m4 4l4 4-4 4M6 16l-4-4 4-4" />
                                    </svg>
                                </div>
                                <div>
                                    <h3 class="text-lg font-bold text-gray-800 dark:text-gray-100 dark:text-gray-200">Default Page Configuration</h3>
                                    <p class="text-sm text-gray-500 dark:text-gray-400 mt-0.5">HTML content served on port 80 when no proxy matches</p>
                                </div>
                            </div>
                        </div>
                        <div class="p-6">
                            <div class="mb-5">
                                <span class="block text-sm font-semibold text-gray-700 dark:text-gray-200 mb-2">HTML Content</span>
                                <textarea 
                                    v-model="defaultPageHtml" 
                                    rows="16" 
                                    class="w-full font-mono text-sm border border-gray-200 dark:border-gray-700 rounded-xl p-4 bg-gray-50/50 dark:bg-gray-700 focus:ring-2 focus:ring-gray-500/20 focus:border-gray-500 focus:bg-white outline-none transition-all resize-y min-h-[300px]"
                                    placeholder="<!DOCTYPE html>..."
                                ></textarea>
                            </div>
                            <div class="flex items-center justify-between pt-4 border-t border-gray-100 dark:border-gray-700">
                                <p class="text-xs text-gray-400 dark:text-gray-500">Changes are applied immediately after saving</p>
                                <button @click="saveDefaultPage" :disabled="isLoading" class="bg-gray-700 hover:bg-gray-800 text-white px-6 py-2.5 rounded-xl shadow-lg transition-all font-semibold disabled:opacity-50 flex items-center gap-2">
                                    <svg v-if="isLoading" class="animate-spin h-4 w-4" fill="none" viewBox="0 0 24 24"><circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle><path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path></svg>
                                    <svg v-else class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 13l4 4L19 7" /></svg>
                                    <span>{{ isLoading ? 'Saving...' : 'Save Changes' }}</span>
                                </button>
                            </div>
                        </div>
                    </div>
                </div>

                <!-- DNS Tab -->
                <div v-show="activeTab === 'dns'">
                    <div class="bg-white dark:bg-gray-800 rounded-xl border border-gray-200 dark:border-gray-700 overflow-hidden shadow-sm">
                        <div class="px-6 py-4 bg-amber-50 dark:bg-amber-900/20 border-b border-amber-200 dark:border-amber-800">
                            <div class="flex items-center gap-3">
                                <div class="w-10 h-10 rounded-xl bg-amber-600 flex items-center justify-center shadow-lg shadow-amber-600/20">
                                    <svg xmlns="http://www.w3.org/2000/svg" class="h-5 w-5 text-white" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                                        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 12l2 2 4-4m5.618-4.016A11.955 11.955 0 0112 2.944a11.955 11.955 0 01-8.618 3.04A12.02 12.02 0 003 9c0 5.591 3.824 10.29 9 11.622 5.176-1.332 9-6.03 9-11.622 0-1.042-.133-2.052-.382-3.016z" />
                                    </svg>
                                </div>
                                <div class="flex-1">
                                    <div class="flex items-center gap-2">
                                        <h3 class="text-lg font-bold text-gray-800 dark:text-gray-100 dark:text-gray-200">ZeroSSL EAB Credentials</h3>
                                        <span v-if="zerosslEabConfigured" class="px-2 py-0.5 text-xs font-medium bg-green-100 text-green-700 rounded-full">Configured</span>
                                        <span v-else class="px-2 py-0.5 text-xs font-medium bg-amber-100 text-amber-700 rounded-full">Not Configured</span>
                                    </div>
                                    <p class="text-sm text-gray-500 dark:text-gray-400 mt-0.5">Required for ZeroSSL DNS challenge support</p>
                                </div>
                            </div>
                        </div>
                        <div class="p-6">
                            <div id="zerossl" class="mb-5 p-4 bg-blue-50 dark:bg-blue-900/20 border border-blue-200 dark:border-blue-800 rounded-xl transition-all duration-300">
                                <div class="flex items-start gap-3">
                                    <svg xmlns="http://www.w3.org/2000/svg" class="h-5 w-5 text-blue-500 dark:text-blue-400 flex-shrink-0 mt-0.5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                                        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13 16h-1v-4h-1m1-4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
                                    </svg>
                                    <div class="text-sm text-blue-700 dark:text-blue-300">
                                        <p class="font-medium">How to get EAB credentials:</p>
                                        <ol class="mt-1 ml-4 list-decimal space-y-1">
                                            <li>Go to <a href="https://app.zerossl.com/developer" target="_blank" class="underline font-medium">ZeroSSL Developer Portal</a></li>
                                            <li>Create a free account or login</li>
                                            <li>Navigate to "EAB Credentials for ACME Clients"</li>
                                            <li>Generate new credentials and paste them below</li>
                                        </ol>
                                    </div>
                                </div>
                            </div>
                            <div class="space-y-4">
                                <div>
                                    <label class="block text-sm font-semibold text-gray-700 dark:text-gray-200 mb-2">EAB Key ID (KID)</label>
                                    <input v-model="zerosslEabKid" type="text" class="w-full font-mono text-sm border border-gray-200 dark:border-gray-700 rounded-xl px-4 py-3 bg-gray-50/50 dark:bg-gray-700 focus:ring-2 focus:ring-amber-500/20 focus:border-amber-500 focus:bg-white outline-none transition-all" placeholder="e.g. abc123def456..." />
                                </div>
                                <div>
                                    <label class="block text-sm font-semibold text-gray-700 dark:text-gray-200 mb-2">EAB HMAC Key</label>
                                    <input v-model="zerosslEabHmacKey" type="password" class="w-full font-mono text-sm border border-gray-200 dark:border-gray-700 rounded-xl px-4 py-3 bg-gray-50/50 dark:bg-gray-700 focus:ring-2 focus:ring-amber-500/20 focus:border-amber-500 focus:bg-white outline-none transition-all" placeholder="Enter new HMAC key to update..." />
                                    <p class="text-xs text-gray-400 mt-1">Leave empty to keep existing key</p>
                                </div>
                            </div>
                            <div class="flex items-center justify-between pt-4 mt-4 border-t border-gray-100 dark:border-gray-600">
                                <button @click="saveZeroSSLEAB" :disabled="isLoadingEab" class="bg-amber-600 hover:bg-amber-700 text-white px-6 py-2.5 rounded-xl shadow-lg transition-all font-semibold disabled:opacity-50 flex items-center gap-2">
                                    <svg v-if="isLoadingEab" class="animate-spin h-4 w-4" fill="none" viewBox="0 0 24 24"><circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle><path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path></svg>
                                    <svg v-else class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 13l4 4L19 7" /></svg>
                                    <span>{{ isLoadingEab ? 'Saving...' : 'Save Credentials' }}</span>
                                </button>
                            </div>
                        </div>
                    </div>
                </div>
            </div>
        </div>
    </div>
</template>

<style scoped>
div { animation: fadeInUp 0.3s cubic-bezier(0.22, 1, 0.36, 1) forwards; }
@keyframes fadeInUp { from { opacity: 0; transform: translateY(10px); } to { opacity: 1; transform: translateY(0); } }
</style>
