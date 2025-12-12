<script setup lang="ts">
import { ref, onMounted, onUnmounted, watch, computed } from 'vue'
import CustomSelect from '../components/CustomSelect.vue'
import { toast } from 'vue-sonner'
import { useConfirm } from '../composables/useConfirm'
import { useWebSocket } from '../composables/useWebSocket'
import { authFetch } from '../composables/useAuth'

const { confirm } = useConfirm()
const { on, off, isConnected } = useWebSocket()

const cleanTarget = (target: string): string => {
    if (!target) return target
    let cleaned = target.trim()
    cleaned = cleaned.replace(/^https?:\/\//i, '')
    cleaned = cleaned.replace(/(:\d+)\/.*$/, '$1')
    cleaned = cleaned.replace(/\/.*$/, '')
    return cleaned
}

const isValidTarget = (target: string): boolean => {
    if (!target || target.trim() === '') return false
    
    const targetPattern = /^([a-zA-Z0-9]([a-zA-Z0-9\-]{0,61}[a-zA-Z0-9])?\.)*[a-zA-Z0-9]([a-zA-Z0-9\-]{0,61}[a-zA-Z0-9])?:\d{1,5}$|^(\d{1,3}\.){3}\d{1,3}:\d{1,5}$/
    
    if (!targetPattern.test(target)) return false
    
    const portMatch = target.match(/:(\d+)$/)
    if (portMatch) {
        const port = parseInt(portMatch[1], 10)
        if (port < 1 || port > 65535) return false
    }
    
    const ipMatch = target.match(/^(\d{1,3})\.(\d{1,3})\.(\d{1,3})\.(\d{1,3}):/)
    if (ipMatch) {
        for (let i = 1; i <= 4; i++) {
            const octet = parseInt(ipMatch[i], 10)
            if (octet < 0 || octet > 255) return false
        }
    }
    
    return true
}

interface Certificate {
  id: number
  domain: string
}

interface Stream {
  id: number
  name: string
  listen_port: number
  tcp_enabled: boolean
  udp_enabled: boolean
  target: string
  ssl: boolean
  ssl_provider: string
  ssl_actual_provider?: string
  ssl_status?: string
  ssl_error?: string
  certificate_id: number | null
  is_active: boolean
}

const streams = ref<Stream[]>([])
const certificates = ref<Certificate[]>([])
const showModal = ref(false)
const isEditing = ref(false)
const editingId = ref<number | null>(null)
const searchQuery = ref('')
const activeTab = ref<'details' | 'ssl'>('details')

const filteredStreams = computed(() => {
    if (!searchQuery.value.trim()) return streams.value
    const q = searchQuery.value.toLowerCase()
    return streams.value.filter(s => 
        s.name?.toLowerCase().includes(q) ||
        s.target?.toLowerCase().includes(q) ||
        String(s.listen_port).includes(q)
    )
})

const newStream = ref({
    name: '',
    listen_port: 0,
    tcp_enabled: true,
    udp_enabled: false,
    target: '',
    ssl: false,
    ssl_provider: 'auto',
    certificate_id: null as number | null,
    is_active: true
})

const sslProviderOptions = [
    { label: 'Auto', value: 'auto' },
    { label: 'Let\'s Encrypt', value: 'letsencrypt' },
    { label: 'ZeroSSL', value: 'zerossl' },
    { label: 'Self-Signed', value: 'selfsigned' },
    { label: 'Custom Certificate', value: 'custom' }
]

const getSSLProviderName = (provider: string, actualProvider?: string) => {
    const providerNames: Record<string, string> = {
        'letsencrypt': "Let's Encrypt",
        'zerossl': 'ZeroSSL',
        'selfsigned': 'Self Signed',
        'custom': 'Custom'
    }

    if (provider === 'auto' && actualProvider) {
        const actualName = providerNames[actualProvider] || actualProvider
        return `Auto (${actualName})`
    }

    if (provider === 'letsencrypt' || provider === 'zerossl') {
        return providerNames[provider] || provider
    }

    if (provider === 'auto') {
        return 'Auto'
    }

    return providerNames[provider] || provider
}

const certificateOptions = computed(() => {
    return certificates.value.map((cert: Certificate) => ({
        label: cert.domain,
        value: cert.id
    }))
})

const fetchStreams = async () => {
    try {
        const res = await authFetch('/api/streams')
        if (res.ok) {
            streams.value = await res.json()
        }
    } catch (e) {
        console.error('Failed to fetch streams:', e)
    }
}

const fetchCertificates = async () => {
    try {
        const res = await authFetch('/api/certificates')
        if (res.ok) {
            certificates.value = await res.json()
        }
    } catch (e) {
        console.error('Failed to fetch certificates:', e)
    }
}

const openAddModal = () => {
    isEditing.value = false
    editingId.value = null
    newStream.value = {
        name: '',
        listen_port: 0,
        tcp_enabled: true,
        udp_enabled: false,
        target: '',
        ssl: false,
        ssl_provider: 'auto',
        certificate_id: null,
        is_active: true
    }
    activeTab.value = 'details'
    showModal.value = true
}

const openEditModal = (stream: Stream) => {
    isEditing.value = true
    editingId.value = stream.id
    newStream.value = { 
        name: stream.name,
        listen_port: stream.listen_port,
        tcp_enabled: stream.tcp_enabled,
        udp_enabled: stream.udp_enabled,
        target: stream.target,
        ssl: stream.ssl,
        ssl_provider: stream.ssl_provider || 'auto',
        certificate_id: stream.certificate_id,
        is_active: stream.is_active
    }
    activeTab.value = 'details'
    showModal.value = true
}

const saveStream = async () => {
    if (!newStream.value.listen_port || newStream.value.listen_port <= 0) {
        toast.error('Listen Port is required and must be greater than 0')
        return
    }
    if (!newStream.value.target) {
        toast.error('Target Address is required')
        return
    }
    if (!isValidTarget(newStream.value.target)) {
        toast.error('Invalid Target Address format. Use format: hostname:port or ip:port (e.g., example.com:3306 or 192.168.1.1:5432)')
        return
    }
    if (!newStream.value.tcp_enabled && !newStream.value.udp_enabled) {
        toast.error('At least one protocol (TCP or UDP) must be enabled')
        return
    }
    if (newStream.value.ssl && newStream.value.ssl_provider === 'custom' && !newStream.value.certificate_id) {
        toast.error('Please select a certificate for custom SSL')
        return
    }

    const url = isEditing.value ? `/api/streams/${editingId.value}` : '/api/streams'
    const method = isEditing.value ? 'PUT' : 'POST'

    try {
        const res = await authFetch(url, {
            method: method,
            headers: {
                'Content-Type': 'application/json'
            },
            body: JSON.stringify(newStream.value)
        })

        if (res.ok) {
            showModal.value = false
            toast.success(isEditing.value ? 'Stream updated successfully' : 'Stream created successfully')
            fetchStreams()
        } else {
            const data = await res.json()
            toast.error(data.error || 'Failed to save stream')
        }
    } catch (e) {
        toast.error('Connection error')
    }
}

const deleteStream = async (id: number) => {
    const confirmed = await confirm(
        'Delete Stream',
        'Are you sure you want to delete this stream? This action cannot be undone.',
        { type: 'danger', confirmText: 'Delete' }
    )
    if (!confirmed) return

    try {
        const res = await authFetch(`/api/streams/${id}`, {
            method: 'DELETE'
        })

        if (res.ok) {
            toast.success('Stream deleted successfully')
            fetchStreams()
        } else {
            toast.error('Failed to delete stream')
        }
    } catch (e) {
        toast.error('Connection error')
    }
}

const toggleStatus = async (stream: Stream) => {
    const newStatus = !stream.is_active

    stream.is_active = newStatus

    try {
        const res = await authFetch(`/api/streams/${stream.id}`, {
            method: 'PUT',
            headers: {
                'Content-Type': 'application/json'
            },
            body: JSON.stringify({ ...stream, is_active: newStatus })
        })

        if (!res.ok) {
            stream.is_active = !newStatus
            toast.error('Failed to update status')
        } else {
            toast.success(`Stream ${newStatus ? 'enabled' : 'disabled'}`)
        }
    } catch (e) {
        stream.is_active = !newStatus
        toast.error('Connection error')
    }
}

const handleStreamUpdated = (updatedStream: Stream) => {
    const index = streams.value.findIndex((s: Stream) => s.id === updatedStream.id)
    if (index !== -1) {
        streams.value[index] = { ...streams.value[index], ...updatedStream }
    }
}

const handleStreamCreated = (stream: Stream) => {
    if (!streams.value.find((s: Stream) => s.id === stream.id)) {
        streams.value.unshift(stream)
    }
}

const handleStreamDeleted = (payload: { id: number | string }) => {
    const idToDelete = Number(payload.id)
    streams.value = streams.value.filter((s: Stream) => s.id !== idToDelete)
}

watch(isConnected, (connected: boolean, wasConnected: boolean) => {
    if (connected && !wasConnected) {
        fetchStreams()
    }
})

onMounted(() => {
    fetchStreams()
    fetchCertificates()
    on('stream_updated', handleStreamUpdated)
    on('stream_created', handleStreamCreated)
    on('stream_deleted', handleStreamDeleted)
})

onUnmounted(() => {
    off('stream_updated', handleStreamUpdated)
    off('stream_created', handleStreamCreated)
    off('stream_deleted', handleStreamDeleted)
})
</script>

<template>
  <div>
    <!-- Header -->
    <div class="bg-white rounded-t-xl p-5 border-b border-gray-200 dark:border-gray-700 shadow-sm overflow-hidden relative">
        <!-- Accent bar -->
        <div class="absolute top-0 left-0 right-0 h-1 bg-yellow-500"></div>
        
        <div class="flex flex-col lg:flex-row justify-between items-start lg:items-center gap-4">
            <div class="flex items-center gap-3">
                <div>
                    <h2 class="text-2xl font-bold text-gray-800 dark:text-gray-100 tracking-tight">Streams</h2>
                    <p class="text-sm text-gray-500 dark:text-gray-400 mt-1">TCP/UDP proxy for backend services</p>
                </div>
            </div>
            <div class="flex items-center gap-3 w-full lg:w-auto">
                <div class="relative flex-1 lg:flex-initial">
                    <input 
                        v-model="searchQuery" 
                        type="text" 
                        placeholder="Search streams..." 
                        class="w-full lg:w-64 pl-10 pr-4 py-2.5 border border-gray-200 dark:border-gray-700 rounded-xl text-sm focus:ring-2 focus:ring-yellow-500/20 focus:border-yellow-500 transition-all bg-gray-50/50"
                    />
                    <svg xmlns="http://www.w3.org/2000/svg" class="h-4 w-4 text-gray-400 absolute left-3.5 top-1/2 -translate-y-1/2" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z" />
                    </svg>
                </div>
                <button @click="openAddModal" class="bg-yellow-500 hover:bg-yellow-600 text-white px-5 py-2.5 rounded-xl shadow-lg shadow-yellow-500/25 hover:shadow-yellow-500/40 transition-all duration-300 font-semibold text-sm flex items-center gap-2 whitespace-nowrap">
                    <svg xmlns="http://www.w3.org/2000/svg" class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4v16m8-8H4" />
                    </svg>
                    <span class="hidden sm:inline">Add Stream</span>
                    <span class="sm:hidden">Add</span>
                </button>
            </div>
        </div>
    </div>

    <!-- Column Headers - Desktop -->
    <div class="bg-gray-50/80 dark:bg-gray-700/80 px-5 py-3 border-b text-xs font-bold text-gray-500 dark:text-gray-400 uppercase tracking-wider border-x border-gray-200 dark:border-gray-700 hidden lg:grid lg:grid-cols-12 gap-4">
        <div class="col-span-3 pl-14">Name</div>
        <div class="col-span-2">Listen Port</div>
        <div class="col-span-2">Protocol</div>
        <div class="col-span-3">Target</div>
        <div class="col-span-2 text-right pr-2">Status</div>
    </div>

    <!-- List -->
    <div class="bg-white rounded-b-xl shadow-lg border-x border-b border-gray-200 dark:border-gray-700 dark:border-gray-700">
        <!-- Empty State -->
        <div v-if="streams.length === 0" class="p-12 text-center">
            <div class="mb-4">
                <div class="mx-auto w-20 h-20 rounded-2xl bg-yellow-500 flex items-center justify-center shadow-lg shadow-yellow-500/30">
                    <svg xmlns="http://www.w3.org/2000/svg" class="h-10 w-10 text-white" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M7.5 21L3 16.5m0 0L7.5 12M3 16.5h13.5m0-13.5L21 7.5m0 0L16.5 12M21 7.5H7.5" />
                    </svg>
                </div>
            </div>
            <p class="text-lg font-bold text-gray-700 dark:text-gray-200 mb-2">No streams configured</p>
            <p class="text-sm text-gray-400 mb-6">Create a stream to proxy TCP/UDP traffic to backend services</p>
            <button @click="openAddModal" class="inline-flex items-center gap-2 bg-yellow-500 hover:bg-yellow-600 text-white px-5 py-2.5 rounded-xl shadow-lg shadow-yellow-500/25 transition-all text-sm font-semibold">
                <svg xmlns="http://www.w3.org/2000/svg" class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4v16m8-8H4" />
                </svg>
                Add Stream
            </button>
        </div>

        <!-- No Search Results -->
        <div v-else-if="filteredStreams.length === 0" class="p-12 text-center">
            <div class="mb-4">
                <div class="mx-auto w-20 h-20 rounded-2xl bg-gray-100 flex items-center justify-center">
                    <svg xmlns="http://www.w3.org/2000/svg" class="h-10 w-10 text-gray-400 dark:text-gray-500" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z" />
                    </svg>
                </div>
            </div>
            <p class="text-lg font-bold text-gray-700 dark:text-gray-200 mb-2">No results found</p>
            <p class="text-sm text-gray-400 dark:text-gray-500">No streams match "{{ searchQuery }}"</p>
        </div>

        <template v-else>
            <!-- Desktop View -->
            <div v-for="stream in filteredStreams" :key="stream.id" class="hidden lg:grid lg:grid-cols-12 gap-4 p-4 border-b border-gray-100 hover:bg-gray-50/50 transition-all duration-200 last:border-b-0 group items-center">
                <!-- Icon + Name -->
                <div class="col-span-3 flex items-center gap-3">
                    <div class="w-10 h-10 rounded-xl bg-yellow-500 flex items-center justify-center shadow-lg shadow-yellow-500/20 group-hover:shadow-yellow-500/30 transition-all">
                        <svg xmlns="http://www.w3.org/2000/svg" class="h-5 w-5 text-white" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M7.5 21L3 16.5m0 0L7.5 12M3 16.5h13.5m0-13.5L21 7.5m0 0L16.5 12M21 7.5H7.5" />
                        </svg>
                    </div>
                    <span class="text-sm font-bold text-gray-900 dark:text-gray-100 dark:text-gray-100">{{ stream.name || 'Unnamed Stream' }}</span>
                </div>
                
                <!-- Listen Port -->
                <div class="col-span-2">
                    <span class="font-mono bg-gray-100 px-3 py-1.5 rounded-lg text-gray-700 dark:text-gray-200 text-sm font-bold border border-gray-200 dark:border-gray-700 dark:border-gray-700">:{{ stream.listen_port }}</span>
                </div>
                
                <!-- Protocol -->
                <div class="col-span-2 flex flex-wrap gap-1.5">
                    <span v-if="stream.tcp_enabled" class="bg-purple-100 text-purple-700 px-2.5 py-1 rounded-lg text-[10px] font-bold uppercase tracking-wide border border-purple-200">TCP</span>
                    <span v-if="stream.udp_enabled" class="bg-orange-100 text-orange-700 px-2.5 py-1 rounded-lg text-[10px] font-bold uppercase tracking-wide border border-orange-200">UDP</span>
                    <span v-if="stream.ssl" class="bg-green-100 text-green-700 px-2.5 py-1 rounded-lg text-[10px] font-bold uppercase tracking-wide border border-green-200 flex items-center gap-1">
                        <svg xmlns="http://www.w3.org/2000/svg" class="h-3 w-3" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 15v2m-6 4h12a2 2 0 002-2v-6a2 2 0 00-2-2H6a2 2 0 00-2 2v6a2 2 0 002 2zm10-10V7a4 4 0 00-8 0v4h8z" />
                        </svg>
                        {{ getSSLProviderName(stream.ssl_provider, stream.ssl_actual_provider) }}
                    </span>
                </div>
                
                <!-- Target -->
                <div class="col-span-3">
                    <span class="font-mono bg-gray-50 dark:bg-gray-700 px-2 py-1 rounded-lg text-gray-600 dark:text-gray-300 text-xs border border-gray-200 dark:border-gray-700 break-all">{{ stream.target }}</span>
                </div>
                
                <!-- Status & Actions -->
                <div class="col-span-2 flex items-center justify-end gap-3">
                    <button @click="toggleStatus(stream)" :class="stream.is_active ? 'bg-green-500' : 'bg-gray-300'" class="relative inline-flex h-6 w-11 items-center rounded-full transition-all focus:outline-none focus:ring-2 focus:ring-yellow-500 focus:ring-offset-2">
                        <span :class="stream.is_active ? 'translate-x-6' : 'translate-x-1'" class="inline-block h-4 w-4 transform rounded-full bg-white dark:bg-gray-800 transition-transform shadow-sm" />
                    </button>
                    <div class="flex items-center gap-1">
                        <button @click="openEditModal(stream)" class="p-2 text-gray-400 hover:text-yellow-600 hover:bg-yellow-50 rounded-lg transition-all" title="Edit">
                            <svg xmlns="http://www.w3.org/2000/svg" class="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M11 5H6a2 2 0 00-2 2v11a2 2 0 002 2h11a2 2 0 002-2v-5m-1.414-9.414a2 2 0 112.828 2.828L11.828 15H9v-2.828l8.586-8.586z" />
                            </svg>
                        </button>
                        <button @click="deleteStream(stream.id)" class="p-2 text-gray-400 hover:text-red-600 hover:bg-red-50 rounded-lg transition-all" title="Delete">
                            <svg xmlns="http://www.w3.org/2000/svg" class="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16" />
                            </svg>
                        </button>
                    </div>
                </div>
            </div>

            <!-- Mobile View -->
            <div v-for="stream in filteredStreams" :key="'mobile-' + stream.id" class="lg:hidden p-4 border-b border-gray-100 last:border-b-0">
                <div class="bg-white rounded-xl border border-gray-200 dark:border-gray-700 overflow-hidden shadow-sm hover:shadow-md transition-all duration-300">
                    <!-- Status indicator strip -->
                    <div :class="stream.is_active ? 'bg-green-500' : 'bg-gray-300'" class="h-1"></div>
                    
                    <div class="p-4">
                        <!-- Header -->
                        <div class="flex items-center gap-3 mb-4">
                            <div class="w-12 h-12 rounded-xl bg-yellow-500 flex items-center justify-center shadow-lg shadow-yellow-500/20">
                                <svg xmlns="http://www.w3.org/2000/svg" class="h-6 w-6 text-white" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M7.5 21L3 16.5m0 0L7.5 12M3 16.5h13.5m0-13.5L21 7.5m0 0L16.5 12M21 7.5H7.5" />
                                </svg>
                            </div>
                            <div class="flex-1 min-w-0">
                                <p class="text-base font-bold text-gray-900 dark:text-gray-100 truncate">{{ stream.name || 'Unnamed Stream' }}</p>
                                <div class="flex items-center gap-2 mt-1">
                                    <span class="font-mono bg-gray-100 px-2 py-0.5 rounded text-gray-700 dark:text-gray-200 text-xs font-bold">:{{ stream.listen_port }}</span>
                                    <span class="text-gray-300">></span>
                                    <span class="font-mono text-gray-500 dark:text-gray-400 text-xs truncate">{{ stream.target }}</span>
                                </div>
                            </div>
                        </div>
                        
                        <!-- Info Pills -->
                        <div class="flex flex-wrap gap-2 mb-4">
                            <!-- Protocols -->
                            <span v-if="stream.tcp_enabled" class="bg-purple-50 text-purple-700 px-2.5 py-1 rounded-lg text-xs font-bold border border-purple-200">TCP</span>
                            <span v-if="stream.udp_enabled" class="bg-orange-50 text-orange-700 px-2.5 py-1 rounded-lg text-xs font-bold border border-orange-200">UDP</span>
                            <span v-if="stream.ssl" class="bg-green-50 text-green-700 px-2.5 py-1 rounded-lg text-xs font-bold border border-green-200 flex items-center gap-1">
                                <svg xmlns="http://www.w3.org/2000/svg" class="h-3 w-3" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 15v2m-6 4h12a2 2 0 002-2v-6a2 2 0 00-2-2H6a2 2 0 00-2 2v6a2 2 0 002 2zm10-10V7a4 4 0 00-8 0v4h8z" />
                                </svg>
                                {{ getSSLProviderName(stream.ssl_provider, stream.ssl_actual_provider) }}
                            </span>
                            
                            <!-- Status Toggle -->
                            <div class="flex items-center gap-2 ml-auto">
                                <span :class="stream.is_active ? 'text-green-600' : 'text-gray-400 dark:text-gray-500 dark:text-gray-400'" class="text-xs font-bold">{{ stream.is_active ? 'Active' : 'Inactive' }}</span>
                                <button @click="toggleStatus(stream)" :class="stream.is_active ? 'bg-green-500' : 'bg-gray-300'" class="relative inline-flex h-6 w-11 items-center rounded-full transition-all">
                                    <span :class="stream.is_active ? 'translate-x-6' : 'translate-x-1'" class="inline-block h-4 w-4 transform rounded-full bg-white dark:bg-gray-800 transition-transform shadow-sm" />
                                </button>
                            </div>
                        </div>
                        
                        <!-- Actions -->
                        <div class="flex gap-2 pt-3 border-t border-gray-100 dark:border-gray-700">
                            <button @click="openEditModal(stream)" class="flex-1 flex items-center justify-center gap-2 px-4 py-2.5 bg-yellow-500 hover:bg-yellow-600 text-white rounded-lg text-sm font-semibold transition-all shadow-sm">
                                <svg xmlns="http://www.w3.org/2000/svg" class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M11 5H6a2 2 0 00-2 2v11a2 2 0 002 2h11a2 2 0 002-2v-5m-1.414-9.414a2 2 0 112.828 2.828L11.828 15H9v-2.828l8.586-8.586z" />
                                </svg>
                                Edit
                            </button>
                            <button @click="deleteStream(stream.id)" class="flex items-center justify-center gap-2 px-4 py-2.5 border border-gray-200 dark:border-gray-700 text-red-500 hover:bg-red-50 hover:border-red-200 rounded-lg text-sm font-semibold transition-all">
                                <svg xmlns="http://www.w3.org/2000/svg" class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16" />
                                </svg>
                            </button>
                        </div>
                    </div>
                </div>
            </div>
        </template>
    </div>

    <!-- Modal -->
    <div v-if="showModal" class="fixed inset-0 z-50 flex items-center justify-center bg-black/60 backdrop-blur-sm p-4">
      <div class="bg-white rounded-2xl shadow-2xl w-full max-w-lg transform transition-all overflow-hidden max-h-[90vh] flex flex-col">
        <!-- Modal Header -->
        <div class="bg-yellow-500 px-6 py-4 flex justify-between items-center flex-shrink-0">
            <div class="flex items-center gap-3">
                <div class="w-10 h-10 rounded-xl bg-white/20 flex items-center justify-center">
                    <svg xmlns="http://www.w3.org/2000/svg" class="h-5 w-5 text-white" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M7.5 21L3 16.5m0 0L7.5 12M3 16.5h13.5m0-13.5L21 7.5m0 0L16.5 12M21 7.5H7.5" />
                    </svg>
                </div>
                <h3 class="text-lg font-bold text-white">{{ isEditing ? 'Edit Stream' : 'Add Stream' }}</h3>
            </div>
            <button @click="showModal = false" class="text-white/80 hover:text-white hover:bg-white/20 rounded-lg p-1.5 transition-all">
                <svg xmlns="http://www.w3.org/2000/svg" class="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
                </svg>
            </button>
        </div>

        <!-- Tabs -->
        <div class="px-6 pt-4 flex-shrink-0">
            <div class="flex gap-1 p-1 bg-gray-100 rounded-xl">
                <button @click="activeTab = 'details'" :class="activeTab === 'details' ? 'bg-white dark:bg-gray-800 shadow-sm text-yellow-600' : 'text-gray-600 hover:text-gray-800 dark:text-gray-200'" class="flex-1 py-2 px-3 rounded-lg font-medium text-sm transition-all">
                    Details
                </button>
                <button @click="activeTab = 'ssl'" :class="activeTab === 'ssl' ? 'bg-white dark:bg-gray-800 shadow-sm text-yellow-600' : 'text-gray-600 hover:text-gray-800 dark:text-gray-200'" class="flex-1 py-2 px-3 rounded-lg font-medium text-sm transition-all">
                    SSL/TLS
                </button>
            </div>
        </div>

        <div class="p-6 space-y-5 overflow-y-auto flex-1">
            <!-- Details Tab -->
            <div v-if="activeTab === 'details'" class="space-y-5">
                <div>
                    <span class="block text-sm font-semibold text-gray-700 dark:text-gray-200 mb-2">Name <span class="text-red-500">*</span></span>
                    <input v-model="newStream.name" class="w-full border border-gray-200 dark:border-gray-700 rounded-xl px-4 py-3 text-gray-700 dark:text-gray-200 bg-gray-50/50 dark:bg-gray-700 focus:outline-none focus:ring-2 focus:ring-yellow-500/20 focus:border-yellow-500 focus:bg-white transition-all" placeholder="MySQL Database">
                </div>

                <div>
                    <span class="block text-sm font-semibold text-gray-700 dark:text-gray-200 mb-2">Listen Port <span class="text-red-500">*</span></span>
                    <input v-model.number="newStream.listen_port" type="number" min="1" max="65535" class="w-full border border-gray-200 dark:border-gray-700 rounded-xl px-4 py-3 text-gray-700 dark:text-gray-200 bg-gray-50/50 dark:bg-gray-700 focus:outline-none focus:ring-2 focus:ring-yellow-500/20 focus:border-yellow-500 focus:bg-white transition-all" placeholder="3306">
                </div>

                <!-- Protocols Section -->
                <div class="border border-gray-200 dark:border-gray-700 rounded-xl overflow-hidden">
                    <div class="bg-gray-50 dark:bg-gray-700 px-4 py-3 border-b border-gray-200 dark:border-gray-700">
                        <span class="font-semibold text-gray-700 dark:text-gray-200">Protocols</span>
                    </div>
                    <div class="divide-y divide-gray-100 dark:divide-gray-700 bg-white dark:bg-gray-800">
                        <div class="flex items-center justify-between px-4 py-3 hover:bg-gray-50/50 dark:hover:bg-gray-700/50 transition-colors">
                            <div class="flex items-center gap-2">
                                <span class="bg-purple-100 dark:bg-purple-900/30 text-purple-700 dark:text-purple-400 px-2 py-0.5 rounded text-xs font-bold">TCP</span>
                                <span class="text-sm text-gray-600 dark:text-gray-400">Transmission Control Protocol</span>
                            </div>
                            <button type="button" @click="newStream.tcp_enabled = !newStream.tcp_enabled" :class="newStream.tcp_enabled ? 'bg-yellow-500' : 'bg-gray-300 dark:bg-gray-600'" class="relative inline-flex h-6 w-11 items-center rounded-full transition-all">
                                <span :class="newStream.tcp_enabled ? 'translate-x-6' : 'translate-x-1'" class="inline-block h-4 w-4 transform rounded-full bg-white transition-transform shadow-sm" />
                            </button>
                        </div>
                        <div class="flex items-center justify-between px-4 py-3 hover:bg-gray-50/50 dark:hover:bg-gray-700/50 transition-colors">
                            <div class="flex items-center gap-2">
                                <span class="bg-orange-100 dark:bg-orange-900/30 text-orange-700 dark:text-orange-400 px-2 py-0.5 rounded text-xs font-bold">UDP</span>
                                <span class="text-sm text-gray-600 dark:text-gray-400">User Datagram Protocol</span>
                            </div>
                            <button type="button" @click="newStream.udp_enabled = !newStream.udp_enabled" :class="newStream.udp_enabled ? 'bg-yellow-500' : 'bg-gray-300 dark:bg-gray-600'" class="relative inline-flex h-6 w-11 items-center rounded-full transition-all">
                                <span :class="newStream.udp_enabled ? 'translate-x-6' : 'translate-x-1'" class="inline-block h-4 w-4 transform rounded-full bg-white transition-transform shadow-sm" />
                            </button>
                        </div>
                    </div>
                </div>

                <div>
                    <span class="block text-sm font-semibold text-gray-700 dark:text-gray-200 mb-2">Target Address <span class="text-red-500">*</span></span>
                    <input v-model="newStream.target" @blur="newStream.target = cleanTarget(newStream.target)" class="w-full border border-gray-200 dark:border-gray-700 rounded-xl px-4 py-3 text-gray-700 dark:text-gray-200 bg-gray-50/50 dark:bg-gray-700 focus:outline-none focus:ring-2 focus:ring-yellow-500/20 focus:border-yellow-500 focus:bg-white transition-all font-mono" placeholder="192.168.1.100:3306">
                    <p class="text-xs text-gray-500 dark:text-gray-400 mt-2">Format: hostname:port or ip:port</p>
                </div>
            </div>

            <!-- SSL/TLS Tab -->
            <div v-if="activeTab === 'ssl'" class="space-y-5">
                <!-- TLS Toggle -->
                <div class="border border-gray-200 dark:border-gray-700 rounded-xl overflow-hidden">
                    <div class="px-4 py-4 flex items-center justify-between bg-gray-50 dark:bg-gray-700 dark:bg-gray-800">
                        <div class="flex items-center gap-3">
                            <div class="w-10 h-10 rounded-xl bg-green-100 flex items-center justify-center">
                                <svg xmlns="http://www.w3.org/2000/svg" class="h-5 w-5 text-green-600" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 15v2m-6 4h12a2 2 0 002-2v-6a2 2 0 00-2-2H6a2 2 0 00-2 2v6a2 2 0 002 2zm10-10V7a4 4 0 00-8 0v4h8z" />
                                </svg>
                            </div>
                            <div>
                                <span class="font-semibold text-gray-700 dark:text-gray-200 dark:text-gray-300">Enable TLS Termination</span>
                                <p class="text-xs text-gray-500 dark:text-gray-400 dark:text-gray-400 dark:text-gray-500">Decrypt incoming TLS traffic</p>
                            </div>
                        </div>
                        <button @click="newStream.ssl = !newStream.ssl" :class="newStream.ssl ? 'bg-green-500' : 'bg-gray-300'" class="relative inline-flex h-6 w-11 items-center rounded-full transition-all">
                            <span :class="newStream.ssl ? 'translate-x-6' : 'translate-x-1'" class="inline-block h-4 w-4 transform rounded-full bg-white dark:bg-gray-800 transition-transform shadow-sm" />
                        </button>
                    </div>
                </div>

                <!-- SSL Options (shown when enabled) -->
                <div v-if="newStream.ssl" class="space-y-4">
                    <div>
                        <span class="block text-sm font-semibold text-gray-700 dark:text-gray-200 mb-2">SSL Provider</span>
                        <CustomSelect v-model="newStream.ssl_provider" :options="sslProviderOptions" placeholder="Select SSL Provider" />
                    </div>
                    <div v-if="newStream.ssl_provider === 'custom'">
                        <span class="block text-sm font-semibold text-gray-700 dark:text-gray-200 mb-2">Certificate <span class="text-red-500">*</span></span>
                        <CustomSelect v-model="newStream.certificate_id" :options="certificateOptions" placeholder="Select a certificate" />
                    </div>
                </div>

                <!-- Info when SSL is disabled -->
                <div v-else class="bg-gray-50 rounded-xl p-6 text-center">
                    <div class="w-16 h-16 rounded-2xl bg-gray-100 flex items-center justify-center mx-auto mb-4">
                        <svg xmlns="http://www.w3.org/2000/svg" class="h-8 w-8 text-gray-400 dark:text-gray-500" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M8 11V7a4 4 0 118 0m-4 8v2m-6 4h12a2 2 0 002-2v-6a2 2 0 00-2-2H6a2 2 0 00-2 2v6a2 2 0 002 2z" />
                        </svg>
                    </div>
                    <p class="text-gray-600 font-medium">TLS is disabled</p>
                    <p class="text-sm text-gray-400 mt-1">Enable TLS to configure SSL certificates</p>
                </div>
            </div>
        </div>

        <div class="bg-gray-50 px-6 py-4 border-t border-gray-200 dark:border-gray-700 flex justify-end gap-3 flex-shrink-0">
          <button @click="showModal = false" class="px-5 py-2.5 border border-gray-200 dark:border-gray-700 rounded-xl text-gray-700 dark:text-gray-200 hover:bg-gray-100 transition-all font-medium">Cancel</button>
          <button @click="saveStream" class="px-6 py-2.5 bg-yellow-500 hover:bg-yellow-600 text-white rounded-xl shadow-lg shadow-yellow-500/25 transition-all font-semibold">Save Stream</button>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
/* Page fade in animation */
.max-w-6xl {
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

/* List item animations */
.divide-y > div {
    animation: slideIn 0.25s cubic-bezier(0.22, 1, 0.36, 1) forwards;
    opacity: 0;
}

.divide-y > div:nth-child(1) { animation-delay: 0ms; }
.divide-y > div:nth-child(2) { animation-delay: 30ms; }
.divide-y > div:nth-child(3) { animation-delay: 60ms; }
.divide-y > div:nth-child(4) { animation-delay: 90ms; }
.divide-y > div:nth-child(5) { animation-delay: 120ms; }
.divide-y > div:nth-child(6) { animation-delay: 150ms; }
.divide-y > div:nth-child(7) { animation-delay: 180ms; }
.divide-y > div:nth-child(8) { animation-delay: 210ms; }

@keyframes slideIn {
    from {
        opacity: 0;
        transform: translateX(-8px);
    }
    to {
        opacity: 1;
        transform: translateX(0);
    }
}

/* Modal animation */
.fixed.inset-0 > div {
    animation: modalIn 0.25s cubic-bezier(0.22, 1, 0.36, 1) forwards;
}

@keyframes modalIn {
    from {
        opacity: 0;
        transform: scale(0.96) translateY(10px);
    }
    to {
        opacity: 1;
        transform: scale(1) translateY(0);
    }
}
</style>
