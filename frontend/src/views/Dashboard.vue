<script setup lang="ts">
import { ref, onMounted, onUnmounted } from 'vue'
import { useRouter } from 'vue-router'
import { useWebSocket } from '../composables/useWebSocket'
import { authFetch } from '../composables/useAuth'

const router = useRouter()
const { on, off } = useWebSocket()

const stats = ref({
    totalHosts: 0,
    totalRedirects: 0,
    totalStreams: 0,
    totalAccessLists: 0,
    totalCertificates: 0
})

const fetchStats = async () => {
    const [hostsRes, listsRes, certsRes, streamsRes] = await Promise.all([
        authFetch('/api/hosts'),
        authFetch('/api/access-lists'),
        authFetch('/api/certificates'),
        authFetch('/api/streams')
    ])

    if (hostsRes.ok && listsRes.ok && certsRes.ok) {
        const hosts = await hostsRes.json()
        const lists = await listsRes.json()
        const certs = await certsRes.json()

        stats.value.totalHosts = hosts.filter((h: any) => h.type === 'proxy').length
        stats.value.totalRedirects = hosts.filter((h: any) => h.type === 'redirect').length
        stats.value.totalAccessLists = lists.length
        stats.value.totalCertificates = certs.length
    }

    if (streamsRes.ok) {
        const streams = await streamsRes.json()
        stats.value.totalStreams = streams.length
    }
}

const handleHostCreated = () => {
    fetchStats()
}

const handleHostDeleted = () => {
    fetchStats()
}

const handleStreamCreated = () => {
    stats.value.totalStreams++
}

const handleStreamDeleted = () => {
    stats.value.totalStreams--
}

const handleAccessListCreated = () => {
    stats.value.totalAccessLists++
}

const handleAccessListDeleted = () => {
    stats.value.totalAccessLists--
}

const handleCertCreated = () => {
    stats.value.totalCertificates++
}

const handleCertDeleted = () => {
    stats.value.totalCertificates--
}

onMounted(() => {
    fetchStats()
    
    on('host_created', handleHostCreated)
    on('host_deleted', handleHostDeleted)
    on('stream_created', handleStreamCreated)
    on('stream_deleted', handleStreamDeleted)
    on('access_list_created', handleAccessListCreated)
    on('access_list_deleted', handleAccessListDeleted)
    on('cert_created', handleCertCreated)
    on('cert_deleted', handleCertDeleted)
})

onUnmounted(() => {
    off('host_created', handleHostCreated)
    off('host_deleted', handleHostDeleted)
    off('stream_created', handleStreamCreated)
    off('stream_deleted', handleStreamDeleted)
    off('access_list_created', handleAccessListCreated)
    off('access_list_deleted', handleAccessListDeleted)
    off('cert_created', handleCertCreated)
    off('cert_deleted', handleCertDeleted)
})
</script>

<template>
    <div class="p-4 md:p-6 max-w-7xl mx-auto">
        <!-- Header -->
        <div class="bg-white rounded-xl shadow-sm border border-gray-200 mb-6 overflow-hidden">
            <div class="h-1 bg-green-500"></div>
            <div class="p-4 md:p-6">
                <div class="flex flex-col md:flex-row md:items-center md:justify-between gap-4">
                    <div>
                        <h1 class="text-xl font-bold text-gray-900">Dashboard</h1>
                        <p class="text-sm text-gray-500 mt-0.5">Overview of your proxy management</p>
                    </div>
                    <div class="flex items-center gap-2">
                        <span class="inline-flex items-center gap-1.5 px-3 py-1.5 bg-green-50 text-green-700 text-xs font-medium rounded-full">
                            <span class="w-2 h-2 bg-green-500 rounded-full animate-pulse"></span>
                            System Online
                        </span>
                    </div>
                </div>
            </div>
        </div>
        
        <!-- Stats Grid -->
        <div class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 xl:grid-cols-5 gap-4 md:gap-6">
            <!-- Proxy Hosts Card -->
            <div 
                @click="router.push('/hosts')"
                class="group bg-white rounded-xl shadow-sm border border-gray-200 overflow-hidden cursor-pointer transition-all duration-300 hover:shadow-lg hover:shadow-green-500/10 hover:-translate-y-1"
            >
                <div class="h-1 bg-green-500"></div>
                <div class="p-5">
                    <div class="flex items-center justify-between mb-4">
                        <div class="w-12 h-12 bg-green-500 rounded-xl flex items-center justify-center shadow-lg shadow-green-500/25 group-hover:scale-110 transition-transform">
                            <svg xmlns="http://www.w3.org/2000/svg" class="h-6 w-6 text-white" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 12h14M5 12a2 2 0 01-2-2V6a2 2 0 012-2h14a2 2 0 012 2v4a2 2 0 01-2 2M5 12a2 2 0 00-2 2v4a2 2 0 002 2h14a2 2 0 002-2v-4a2 2 0 00-2-2m-2-4h.01M17 16h.01" />
                            </svg>
                        </div>
                        <svg xmlns="http://www.w3.org/2000/svg" class="h-5 w-5 text-gray-300 group-hover:text-green-500 group-hover:translate-x-1 transition-all" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 5l7 7-7 7" />
                        </svg>
                    </div>
                    <div class="text-3xl font-bold text-gray-900 mb-1">{{ stats.totalHosts }}</div>
                    <h3 class="text-gray-500 font-medium text-sm">Proxy Hosts</h3>
                </div>
            </div>

            <!-- Redirections Card -->
            <div 
                @click="router.push('/hosts')"
                class="group bg-white rounded-xl shadow-sm border border-gray-200 overflow-hidden cursor-pointer transition-all duration-300 hover:shadow-lg hover:shadow-yellow-500/10 hover:-translate-y-1"
            >
                <div class="h-1 bg-yellow-500"></div>
                <div class="p-5">
                    <div class="flex items-center justify-between mb-4">
                        <div class="w-12 h-12 bg-yellow-500 rounded-xl flex items-center justify-center shadow-lg shadow-yellow-500/25 group-hover:scale-110 transition-transform">
                            <svg xmlns="http://www.w3.org/2000/svg" class="h-6 w-6 text-white" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13 9l3 3m0 0l-3 3m3-3H8m13 0a9 9 0 11-18 0 9 9 0 0118 0z" />
                            </svg>
                        </div>
                        <svg xmlns="http://www.w3.org/2000/svg" class="h-5 w-5 text-gray-300 group-hover:text-yellow-500 group-hover:translate-x-1 transition-all" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 5l7 7-7 7" />
                        </svg>
                    </div>
                    <div class="text-3xl font-bold text-gray-900 mb-1">{{ stats.totalRedirects }}</div>
                    <h3 class="text-gray-500 font-medium text-sm">Redirections</h3>
                </div>
            </div>

            <!-- Streams Card -->
            <div 
                @click="router.push('/streams')"
                class="group bg-white rounded-xl shadow-sm border border-gray-200 overflow-hidden cursor-pointer transition-all duration-300 hover:shadow-lg hover:shadow-yellow-500/10 hover:-translate-y-1"
            >
                <div class="h-1 bg-yellow-500"></div>
                <div class="p-5">
                    <div class="flex items-center justify-between mb-4">
                        <div class="w-12 h-12 bg-yellow-500 rounded-xl flex items-center justify-center shadow-lg shadow-yellow-500/25 group-hover:scale-110 transition-transform">
                            <svg xmlns="http://www.w3.org/2000/svg" class="h-6 w-6 text-white" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M8 9l3 3-3 3m5 0h3M5 20h14a2 2 0 002-2V6a2 2 0 00-2-2H5a2 2 0 00-2 2v12a2 2 0 002 2z" />
                            </svg>
                        </div>
                        <svg xmlns="http://www.w3.org/2000/svg" class="h-5 w-5 text-gray-300 group-hover:text-yellow-500 group-hover:translate-x-1 transition-all" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 5l7 7-7 7" />
                        </svg>
                    </div>
                    <div class="text-3xl font-bold text-gray-900 mb-1">{{ stats.totalStreams }}</div>
                    <h3 class="text-gray-500 font-medium text-sm">Streams</h3>
                </div>
            </div>

            <!-- Certificates Card -->
            <div 
                @click="router.push('/certificates')"
                class="group bg-white rounded-xl shadow-sm border border-gray-200 overflow-hidden cursor-pointer transition-all duration-300 hover:shadow-lg hover:shadow-orange-500/10 hover:-translate-y-1"
            >
                <div class="h-1 bg-orange-500"></div>
                <div class="p-5">
                    <div class="flex items-center justify-between mb-4">
                        <div class="w-12 h-12 bg-orange-500 rounded-xl flex items-center justify-center shadow-lg shadow-orange-500/25 group-hover:scale-110 transition-transform">
                            <svg xmlns="http://www.w3.org/2000/svg" class="h-6 w-6 text-white" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 12l2 2 4-4m5.618-4.016A11.955 11.955 0 0112 2.944a11.955 11.955 0 01-8.618 3.04A12.02 12.02 0 003 9c0 5.591 3.824 10.29 9 11.622 5.176-1.332 9-6.03 9-11.622 0-1.042-.133-2.052-.382-3.016z" />
                            </svg>
                        </div>
                        <svg xmlns="http://www.w3.org/2000/svg" class="h-5 w-5 text-gray-300 group-hover:text-orange-500 group-hover:translate-x-1 transition-all" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 5l7 7-7 7" />
                        </svg>
                    </div>
                    <div class="text-3xl font-bold text-gray-900 mb-1">{{ stats.totalCertificates }}</div>
                    <h3 class="text-gray-500 font-medium text-sm">SSL Certificates</h3>
                </div>
            </div>

            <!-- Access Lists Card -->
            <div 
                @click="router.push('/access-lists')"
                class="group bg-white rounded-xl shadow-sm border border-gray-200 overflow-hidden cursor-pointer transition-all duration-300 hover:shadow-lg hover:shadow-red-500/10 hover:-translate-y-1"
            >
                <div class="h-1 bg-red-500"></div>
                <div class="p-5">
                    <div class="flex items-center justify-between mb-4">
                        <div class="w-12 h-12 bg-red-500 rounded-xl flex items-center justify-center shadow-lg shadow-red-500/25 group-hover:scale-110 transition-transform">
                            <svg xmlns="http://www.w3.org/2000/svg" class="h-6 w-6 text-white" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4.354a4 4 0 110 5.292M15 21H3v-1a6 6 0 0112 0v1zm0 0h6v-1a6 6 0 00-9-5.197M13 7a4 4 0 11-8 0 4 4 0 018 0z" />
                            </svg>
                        </div>
                        <svg xmlns="http://www.w3.org/2000/svg" class="h-5 w-5 text-gray-300 group-hover:text-red-500 group-hover:translate-x-1 transition-all" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 5l7 7-7 7" />
                        </svg>
                    </div>
                    <div class="text-3xl font-bold text-gray-900 mb-1">{{ stats.totalAccessLists }}</div>
                    <h3 class="text-gray-500 font-medium text-sm">Access Lists</h3>
                </div>
            </div>
        </div>
    </div>
</template>

<style scoped>
/* Page fade in animation */
.p-4, .p-6 {
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

/* Card stagger animations */
.group:nth-child(1) { animation-delay: 0ms; }
.group:nth-child(2) { animation-delay: 50ms; }
.group:nth-child(3) { animation-delay: 100ms; }
.group:nth-child(4) { animation-delay: 150ms; }
.group:nth-child(5) { animation-delay: 200ms; }
</style>
