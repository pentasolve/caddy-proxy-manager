<script setup lang="ts">
import { ref, onMounted, computed } from 'vue'
import CustomSelect from '../components/CustomSelect.vue'
import Combobox from '../components/Combobox.vue'
import { toast } from 'vue-sonner'
import { useConfirm } from '../composables/useConfirm'
import { authFetch, getValidToken } from '../composables/useAuth'

const { confirm } = useConfirm()

interface Certificate {
  id: number
  domain: string
  created_at: string
  expires_at?: string
  status?: string
  error?: string
  auto_renew?: boolean
  provider?: string
}

const certs = ref<Certificate[]>([])
const hostOptions = ref<string[]>([])
const searchQuery = ref('')

const filteredCerts = computed(() => {
    if (!searchQuery.value.trim()) return certs.value
    const q = searchQuery.value.toLowerCase()
    return certs.value.filter(c => 
        c.domain?.toLowerCase().includes(q) ||
        c.status?.toLowerCase().includes(q)
    )
})
const domain = ref('')
const generateDomain = ref('')
const generateProvider = ref('selfsigned')
const generateEmail = ref('')
const generateUseDNS = ref(false)
const generateDNSProvider = ref('cloudflare')
const generateDNSToken = ref('')
const isGenerating = ref(false)

const providerOptions = [
    { label: 'Self Signed', value: 'selfsigned' },
    { label: "Let's Encrypt", value: 'letsencrypt' },
    { label: 'ZeroSSL', value: 'zerossl' }
]

const dnsProviderOptions = [
    { label: 'Cloudflare', value: 'cloudflare' }
]

const certFile = ref<File | null>(null)
const keyFile = ref<File | null>(null)
const showModal = ref(false)
const showGenerateModal = ref(false)

const fetchCerts = async () => {
  const res = await authFetch('/api/certificates')
  if (res.ok) {
    certs.value = await res.json()
  }
}

const fetchHosts = async () => {
    const res = await authFetch('/api/hosts')
    if (res.ok) {
        const hosts = await res.json()
        hostOptions.value = hosts.map((h: any) => h.domain)
    }
}

const handleCertUpload = (event: Event) => {
    const target = event.target as HTMLInputElement
    if (target.files) certFile.value = target.files[0]
}

const handleKeyUpload = (event: Event) => {
    const target = event.target as HTMLInputElement
    if (target.files) keyFile.value = target.files[0]
}

const uploadCert = async () => {
    if (!certFile.value || !keyFile.value || !domain.value) return

    const formData = new FormData()
    formData.append('domain', domain.value)
    formData.append('cert_file', certFile.value)
    formData.append('key_file', keyFile.value)

    const token = await getValidToken()
    const res = await fetch('/api/certificates', {
        method: 'POST',
        headers: { 'Authorization': `Bearer ${token}` },
        body: formData
    })

    if (res.ok) {
        showModal.value = false
        domain.value = ''
        certFile.value = null
        keyFile.value = null
        fetchCerts()
        toast.success('Certificate uploaded successfully')
    } else {
        toast.error('Failed to upload certificate')
    }
}

const generateCert = async () => {
    if (!generateDomain.value) return
    
    isGenerating.value = true
    const formData = new FormData()
    formData.append('domain', generateDomain.value)
    formData.append('type', generateProvider.value)
    if (generateProvider.value !== 'selfsigned') {
        formData.append('email', generateEmail.value)
        formData.append('use_dns_challenge', generateUseDNS.value.toString())
        if (generateUseDNS.value) {
            formData.append('dns_provider', generateDNSProvider.value)
            formData.append('dns_token', generateDNSToken.value)
        }
    }

    const token = await getValidToken()
    const res = await fetch('/api/certificates', {
        method: 'POST',
        headers: { 'Authorization': `Bearer ${token}` },
        body: formData
    })

    isGenerating.value = false

    if (res.ok) {
        showGenerateModal.value = false
        generateDomain.value = ''
        generateProvider.value = 'selfsigned'
        generateUseDNS.value = false
        generateDNSToken.value = ''
        fetchCerts()
        toast.success('Certificate generated successfully')
    } else {
        const err = await res.json()
        toast.error("Failed to generate certificate: " + (err.error || "Unknown error"))
    }
}

const deleteCert = async (id: number) => {
    const confirmed = await confirm(
        'Delete Certificate', 
        'Are you sure you want to delete this certificate? This action cannot be undone.',
        { type: 'danger', confirmText: 'Delete' }
    )
    if(!confirmed) return;

    const res = await authFetch(`/api/certificates/${id}`, {
        method: 'DELETE'
    })
    
    if (res.ok) {
        toast.success('Certificate deleted successfully')
        fetchCerts()
    } else {
        const err = await res.json()
        toast.error(err.error || 'Failed to delete certificate')
    }
}

const toggleAutoRenew = async (cert: Certificate) => {
    const newValue = !cert.auto_renew
    const res = await authFetch(`/api/certificates/${cert.id}/auto-renew`, {
        method: 'PUT',
        headers: { 
            'Content-Type': 'application/json'
        },
        body: JSON.stringify({ auto_renew: newValue })
    })
    
    if (res.ok) {
        toast.success(`Auto-renew ${newValue ? 'enabled' : 'disabled'} for ${cert.domain}`)
    } else {
        const err = await res.json()
        toast.error(err.error || 'Failed to update auto-renew setting')
        fetchCerts()
    }
}

const downloadFile = async (id: number, type: 'cert' | 'key', domain: string) => {
    const token = await getValidToken()
    const res = await fetch(`/api/certificates/${id}/download?type=${type}`, {
        headers: { 'Authorization': `Bearer ${token}` }
    })
    
    if (res.ok) {
        const blob = await res.blob()
        const url = window.URL.createObjectURL(blob)
        const a = document.createElement('a')
        a.href = url
        a.download = `${domain}.${type === 'cert' ? 'crt' : 'key'}`
        document.body.appendChild(a)
        a.click()
        window.URL.revokeObjectURL(url)
        document.body.removeChild(a)
    } else {
        toast.error("Failed to download file")
    }
}

const formatDate = (dateStr: string) => {
    if (!dateStr || dateStr.startsWith('0001')) return 'Unknown'
    return new Date(dateStr).toLocaleDateString()
}

const getExpiryStatus = (dateStr?: string) => {
    if (!dateStr || dateStr.startsWith('0001')) return 'unknown'
    const expiry = new Date(dateStr)
    const now = new Date()
    const daysUntil = Math.ceil((expiry.getTime() - now.getTime()) / (1000 * 60 * 60 * 24))
    
    if (daysUntil < 0) return 'expired'
    if (daysUntil < 30) return 'warning'
    return 'valid'
}

const getExpiryColor = (status: string) => {
    switch (status) {
        case 'expired': return 'text-red-600 bg-red-50 border-red-200'
        case 'warning': return 'text-yellow-600 bg-yellow-50 border-yellow-200'
        case 'valid': return 'text-green-600 bg-green-50 border-green-200'
        default: return 'text-gray-600 bg-gray-50 border-gray-200'
    }
}


const formatProvider = (provider?: string) => {
    if (!provider) return 'Unknown'
    const p = provider.toLowerCase()
    if (p === 'letsencrypt') return "Let's Encrypt"
    if (p === 'zerossl') return "ZeroSSL"
    if (p === 'selfsigned') return "Self Signed"
    return provider.charAt(0).toUpperCase() + provider.slice(1)
}

import { useWebSocket } from '../composables/useWebSocket'

const { connect, on, off } = useWebSocket()

const handleCertUpdated = (updatedCert: Certificate) => {
    const index = certs.value.findIndex(c => c.id === updatedCert.id)
    if (index !== -1) {
        const newCerts = [...certs.value]
        newCerts[index] = { ...newCerts[index], ...updatedCert }
        certs.value = newCerts
    } else {
        certs.value.unshift(updatedCert)
    }
}

const handleCertCreated = (newCert: Certificate) => {
    certs.value.unshift(newCert)
}

const handleCertDeleted = (payload: any) => {
    certs.value = certs.value.filter(c => c.id != payload.id)
}

onMounted(() => {
    fetchCerts()
    fetchHosts()
    
    on('cert_updated', handleCertUpdated)
    on('cert_created', handleCertCreated)
    on('cert_deleted', handleCertDeleted)
})

import { onUnmounted } from 'vue'
onUnmounted(() => {
    off('cert_updated', handleCertUpdated)
    off('cert_created', handleCertCreated)
    off('cert_deleted', handleCertDeleted)
})
</script>

<template>
  <div>

    <!-- Header -->
    <div class="bg-white rounded-t-xl p-5 border-b border-gray-200 shadow-sm overflow-hidden relative">
        <!-- Gradient accent bar -->
        <div class="absolute top-0 left-0 right-0 h-1 bg-orange-500"></div>
        
        <div class="flex flex-col lg:flex-row justify-between items-start lg:items-center gap-4">
            <div>
                <h2 class="text-2xl font-bold text-gray-800 tracking-tight">SSL Certificates</h2>
                <p class="text-sm text-gray-500 mt-1">Manage SSL/TLS certificates for your domains</p>
            </div>
            <div class="flex flex-col sm:flex-row items-stretch sm:items-center gap-3 w-full lg:w-auto">
                <div class="relative flex-1 lg:flex-initial">
                    <input 
                        v-model="searchQuery" 
                        type="text" 
                        placeholder="Search certificates..." 
                        class="w-full lg:w-64 pl-10 pr-4 py-2.5 border border-gray-200 rounded-xl text-sm focus:ring-2 focus:ring-orange-500/20 focus:border-orange-500 transition-all bg-gray-50/50"
                    />
                    <svg xmlns="http://www.w3.org/2000/svg" class="h-4 w-4 text-gray-400 absolute left-3.5 top-1/2 -translate-y-1/2" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z" />
                    </svg>
                </div>
                <div class="flex gap-2">
                    <button @click="showGenerateModal = true" class="flex-1 sm:flex-initial bg-orange-500 hover:bg-orange-600 text-white px-5 py-2.5 rounded-xl shadow-lg shadow-orange-500/25 hover:shadow-orange-500/40 transition-all duration-300 font-semibold text-sm flex items-center justify-center gap-2 whitespace-nowrap">
                        <svg xmlns="http://www.w3.org/2000/svg" class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 9v3m0 0v3m0-3h3m-3 0H9m12 0a9 9 0 11-18 0 9 9 0 0118 0z" />
                        </svg>
                        <span>Generate</span>
                    </button>
                    <button @click="showModal = true" class="flex-1 sm:flex-initial bg-white border-2 border-orange-200 hover:border-orange-300 text-orange-600 hover:bg-orange-50 px-5 py-2.5 rounded-xl transition-all duration-300 font-semibold text-sm flex items-center justify-center gap-2 whitespace-nowrap">
                        <svg xmlns="http://www.w3.org/2000/svg" class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 16v1a3 3 0 003 3h10a3 3 0 003-3v-1m-4-8l-4-4m0 0L8 8m4-4v12" />
                        </svg>
                        <span>Upload</span>
                    </button>
                </div>
            </div>
        </div>
    </div>

    <!-- Column Headers - Desktop -->
    <div class="bg-gray-50/80 px-5 py-3 border-b text-xs font-bold text-gray-500 uppercase tracking-wider border-x border-gray-200 hidden lg:grid lg:grid-cols-12 gap-4">
        <div class="col-span-3 pl-14">Domain</div>
        <div class="col-span-2">Provider</div>
        <div class="col-span-2">Expires</div>
        <div class="col-span-2">Auto Renew</div>
        <div class="col-span-3 text-right pr-2">Actions</div>
    </div>

    <!-- List -->
    <div class="bg-white rounded-b-xl shadow-lg border-x border-b border-gray-200">
        <!-- Empty State -->
        <div v-if="certs.length === 0" class="p-12 text-center">
            <div class="mb-4">
                <div class="mx-auto w-20 h-20 rounded-2xl bg-orange-500 flex items-center justify-center shadow-lg shadow-orange-500/30">
                    <svg xmlns="http://www.w3.org/2000/svg" class="h-10 w-10 text-white" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M9 12l2 2 4-4m5.618-4.016A11.955 11.955 0 0112 2.944a11.955 11.955 0 01-8.618 3.04A12.02 12.02 0 003 9c0 5.591 3.824 10.29 9 11.622 5.176-1.332 9-6.03 9-11.622 0-1.042-.133-2.052-.382-3.016z" />
                    </svg>
                </div>
            </div>
            <p class="text-lg font-bold text-gray-700 mb-2">No certificates found</p>
            <p class="text-sm text-gray-400 mb-6">Generate or upload SSL certificates for your domains</p>
            <div class="flex items-center justify-center gap-3">
                <button @click="showGenerateModal = true" class="inline-flex items-center gap-2 bg-orange-500 hover:bg-orange-600 text-white px-5 py-2.5 rounded-xl shadow-lg shadow-orange-500/25 transition-all text-sm font-semibold">
                    <svg xmlns="http://www.w3.org/2000/svg" class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 9v3m0 0v3m0-3h3m-3 0H9m12 0a9 9 0 11-18 0 9 9 0 0118 0z" />
                    </svg>
                    Generate
                </button>
                <button @click="showModal = true" class="inline-flex items-center gap-2 border-2 border-orange-200 text-orange-600 hover:bg-orange-50 px-5 py-2.5 rounded-xl transition-all text-sm font-semibold">
                    <svg xmlns="http://www.w3.org/2000/svg" class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 16v1a3 3 0 003 3h10a3 3 0 003-3v-1m-4-8l-4-4m0 0L8 8m4-4v12" />
                    </svg>
                    Upload
                </button>
            </div>
        </div>

        <!-- No Search Results -->
        <div v-else-if="filteredCerts.length === 0" class="p-12 text-center">
            <div class="mb-4">
                <div class="mx-auto w-20 h-20 rounded-2xl bg-gray-100 flex items-center justify-center">
                    <svg xmlns="http://www.w3.org/2000/svg" class="h-10 w-10 text-gray-400" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z" />
                    </svg>
                </div>
            </div>
            <p class="text-lg font-bold text-gray-700 mb-2">No results found</p>
            <p class="text-sm text-gray-400">No certificates match "{{ searchQuery }}"</p>
        </div>

        <template v-else>
            <!-- Desktop View -->
            <div v-for="cert in filteredCerts" :key="cert.id" class="hidden lg:grid lg:grid-cols-12 gap-4 p-4 border-b border-gray-100 hover:bg-gray-50/50 transition-all duration-200 last:border-b-0 group items-center">
                <!-- Icon + Domain -->
                <div class="col-span-3 flex items-center gap-3">
                    <div class="w-10 h-10 rounded-xl bg-orange-500 flex items-center justify-center shadow-lg shadow-orange-500/20 group-hover:shadow-orange-500/30 transition-all">
                        <svg xmlns="http://www.w3.org/2000/svg" class="h-5 w-5 text-white" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 12l2 2 4-4m5.618-4.016A11.955 11.955 0 0112 2.944a11.955 11.955 0 01-8.618 3.04A12.02 12.02 0 003 9c0 5.591 3.824 10.29 9 11.622 5.176-1.332 9-6.03 9-11.622 0-1.042-.133-2.052-.382-3.016z" />
                        </svg>
                    </div>
                    <div class="min-w-0">
                        <p class="text-sm font-bold text-gray-900 truncate">{{ cert.domain }}</p>
                        <p class="text-xs text-gray-400">Added {{ formatDate(cert.created_at) }}</p>
                    </div>
                </div>

                <!-- Provider -->
                <div class="col-span-2">
                    <div class="flex items-center gap-2">
                        <span class="text-sm font-medium text-gray-700">{{ formatProvider(cert.provider) }}</span>
                        <span v-if="cert.provider === 'zerossl'" class="px-1.5 py-0.5 rounded text-[10px] font-bold bg-blue-50 text-blue-600 border border-blue-100">EAB</span>
                    </div>
                </div>
                
                <!-- Status -->
                <div class="col-span-2">
                    <div v-if="cert.status === 'generating'" class="inline-flex items-center gap-1.5 px-3 py-1.5 rounded-lg bg-yellow-50 border border-yellow-200 text-yellow-700">
                        <svg class="animate-spin h-3.5 w-3.5" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24">
                            <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle>
                            <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
                        </svg>
                        <span class="text-xs font-bold">Generating</span>
                    </div>
                    <div v-else-if="cert.status === 'failed'" class="group/tooltip relative inline-flex items-center gap-1.5 px-3 py-1.5 rounded-lg bg-red-50 border border-red-200 text-red-600 cursor-help">
                        <svg xmlns="http://www.w3.org/2000/svg" class="h-3.5 w-3.5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 8v4m0 4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
                        </svg>
                        <span class="text-xs font-bold">Failed</span>
                        <div class="absolute bottom-full left-1/2 transform -translate-x-1/2 mb-2 w-64 max-w-xs bg-gray-900 text-white text-xs rounded-lg p-3 opacity-0 group-hover/tooltip:opacity-100 transition-opacity z-50 pointer-events-none shadow-xl">
                            {{ cert.error || 'Unknown error' }}
                            <div class="absolute top-full left-1/2 transform -translate-x-1/2 border-4 border-transparent border-t-gray-900"></div>
                        </div>
                    </div>
                    <div v-else :class="[
                        'inline-flex items-center gap-1.5 px-3 py-1.5 rounded-lg text-xs font-bold border',
                        getExpiryStatus(cert.expires_at) === 'valid' ? 'bg-green-50 border-green-200 text-green-700' :
                        getExpiryStatus(cert.expires_at) === 'warning' ? 'bg-yellow-50 border-yellow-200 text-yellow-700' :
                        getExpiryStatus(cert.expires_at) === 'expired' ? 'bg-red-50 border-red-200 text-red-600' :
                        'bg-gray-50 border-gray-200 text-gray-600'
                    ]">
                        <svg v-if="getExpiryStatus(cert.expires_at) === 'valid'" xmlns="http://www.w3.org/2000/svg" class="h-3.5 w-3.5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z" />
                        </svg>
                        <svg v-else-if="getExpiryStatus(cert.expires_at) === 'warning'" xmlns="http://www.w3.org/2000/svg" class="h-3.5 w-3.5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-3L13.732 4c-.77-1.333-2.694-1.333-3.464 0L3.34 16c-.77 1.333.192 3 1.732 3z" />
                        </svg>
                        {{ formatDate(cert.expires_at || '-') }}
                    </div>
                </div>
                
                <!-- Auto Renew -->
                <div class="col-span-2">
                    <label class="relative inline-flex items-center cursor-pointer">
                        <input 
                            type="checkbox" 
                            :checked="cert.auto_renew !== false" 
                            @change="toggleAutoRenew(cert)"
                            class="sr-only peer"
                        >
                        <div class="w-11 h-6 bg-gray-200 peer-focus:outline-none peer-focus:ring-2 peer-focus:ring-orange-300 rounded-full peer peer-checked:after:translate-x-full peer-checked:after:border-white after:content-[''] after:absolute after:top-[2px] after:left-[2px] after:bg-white after:border-gray-300 after:border after:rounded-full after:h-5 after:w-5 after:transition-all peer-checked:bg-orange-500"></div>
                        <span class="ml-2 text-xs font-semibold text-gray-600">{{ cert.auto_renew !== false ? 'On' : 'Off' }}</span>
                    </label>
                </div>
                
                <!-- Actions -->
                <div class="col-span-3 flex justify-end gap-1">
                    <button @click="downloadFile(cert.id, 'cert', cert.domain)" class="p-2 text-gray-400 hover:text-blue-600 hover:bg-blue-50 rounded-lg transition-all" title="Download Certificate">
                        <svg xmlns="http://www.w3.org/2000/svg" class="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 16v1a3 3 0 003 3h10a3 3 0 003-3v-1m-4-4l-4 4m0 0l-4-4m4 4V4" />
                        </svg>
                    </button>
                    <button @click="downloadFile(cert.id, 'key', cert.domain)" class="p-2 text-gray-400 hover:text-yellow-600 hover:bg-yellow-50 rounded-lg transition-all" title="Download Private Key">
                        <svg xmlns="http://www.w3.org/2000/svg" class="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 7a2 2 0 012 2m4 0a6 6 0 01-7.743 5.743L11 17H9v2H7v2H4a1 1 0 01-1-1v-2.586a1 1 0 01.293-.707l5.964-5.964A6 6 0 1121 9z" />
                        </svg>
                    </button>
                    <button @click="deleteCert(cert.id)" class="p-2 text-gray-400 hover:text-red-600 hover:bg-red-50 rounded-lg transition-all" title="Delete">
                        <svg xmlns="http://www.w3.org/2000/svg" class="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16" />
                        </svg>
                    </button>
                </div>
            </div>

            <!-- Mobile View -->
            <div v-for="cert in filteredCerts" :key="'mobile-' + cert.id" class="lg:hidden p-4 border-b border-gray-100 last:border-b-0">
                <div class="bg-white rounded-xl border border-gray-200 overflow-hidden shadow-sm hover:shadow-md transition-all duration-300">
                    <!-- Status indicator strip -->
                    <div :class="[
                        'h-1',
                        cert.status === 'generating' ? 'bg-yellow-400 animate-pulse' :
                        cert.status === 'failed' ? 'bg-red-500' :
                        getExpiryStatus(cert.expires_at) === 'valid' ? 'bg-green-500' :
                        getExpiryStatus(cert.expires_at) === 'warning' ? 'bg-yellow-500' :
                        'bg-red-500'
                    ]"></div>
                    
                    <div class="p-4">
                        <!-- Header -->
                        <div class="flex items-center gap-3 mb-4">
                            <div class="w-12 h-12 rounded-xl bg-orange-500 flex items-center justify-center shadow-lg shadow-orange-500/20">
                                <svg xmlns="http://www.w3.org/2000/svg" class="h-6 w-6 text-white" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 12l2 2 4-4m5.618-4.016A11.955 11.955 0 0112 2.944a11.955 11.955 0 01-8.618 3.04A12.02 12.02 0 003 9c0 5.591 3.824 10.29 9 11.622 5.176-1.332 9-6.03 9-11.622 0-1.042-.133-2.052-.382-3.016z" />
                                </svg>
                            </div>
                            <div class="flex-1 min-w-0">
                                <p class="text-base font-bold text-gray-900 truncate">{{ cert.domain }}</p>
                                <p class="text-xs text-gray-400">Added {{ formatDate(cert.created_at) }}</p>
                            </div>
                        </div>
                        
                        <!-- Info Pills -->
                        <div class="flex flex-wrap gap-2 mb-4">
                            <!-- Provider -->
                            <div class="inline-flex items-center gap-2 px-3 py-1.5 rounded-lg bg-gray-50 border border-gray-200">
                                <span class="text-xs text-gray-500">Provider:</span>
                                <span class="text-xs font-semibold text-gray-700">{{ formatProvider(cert.provider) }}</span>
                            </div>
                            <!-- Status -->
                            <div v-if="cert.status === 'generating'" class="inline-flex items-center gap-1.5 px-3 py-1.5 rounded-lg bg-yellow-50 border border-yellow-200 text-yellow-700">
                                <svg class="animate-spin h-3 w-3" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24">
                                    <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle>
                                    <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
                                </svg>
                                <span class="text-xs font-bold">Generating</span>
                            </div>
                            <div v-else-if="cert.status === 'failed'" class="inline-flex items-center gap-1.5 px-3 py-1.5 rounded-lg bg-red-50 border border-red-200 text-red-600">
                                <svg xmlns="http://www.w3.org/2000/svg" class="h-3 w-3" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 8v4m0 4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
                                </svg>
                                <span class="text-xs font-bold">Failed</span>
                            </div>
                            <div v-else :class="[
                                'inline-flex items-center gap-1.5 px-3 py-1.5 rounded-lg text-xs font-bold border',
                                getExpiryStatus(cert.expires_at) === 'valid' ? 'bg-green-50 border-green-200 text-green-700' :
                                getExpiryStatus(cert.expires_at) === 'warning' ? 'bg-yellow-50 border-yellow-200 text-yellow-700' :
                                'bg-red-50 border-red-200 text-red-600'
                            ]">
                                Expires {{ formatDate(cert.expires_at || '-') }}
                            </div>
                            
                            <!-- Auto Renew -->
                            <div class="inline-flex items-center gap-2 px-3 py-1.5 rounded-lg bg-gray-50 border border-gray-200">
                                <span class="text-xs text-gray-500">Auto Renew:</span>
                                <button @click="toggleAutoRenew(cert)" :class="cert.auto_renew !== false ? 'bg-orange-500' : 'bg-gray-300'" class="relative inline-flex h-5 w-9 items-center rounded-full transition-colors">
                                    <span :class="cert.auto_renew !== false ? 'translate-x-4' : 'translate-x-1'" class="inline-block h-3.5 w-3.5 transform rounded-full bg-white transition-transform" />
                                </button>
                            </div>
                        </div>
                        
                        <!-- Actions -->
                        <div class="flex gap-2 pt-3 border-t border-gray-100">
                            <button @click="downloadFile(cert.id, 'cert', cert.domain)" class="flex-1 flex items-center justify-center gap-2 px-4 py-2.5 bg-blue-500 hover:bg-blue-600 text-white rounded-lg text-sm font-semibold transition-all shadow-sm">
                                <svg xmlns="http://www.w3.org/2000/svg" class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 16v1a3 3 0 003 3h10a3 3 0 003-3v-1m-4-4l-4 4m0 0l-4-4m4 4V4" />
                                </svg>
                                Cert
                            </button>
                            <button @click="downloadFile(cert.id, 'key', cert.domain)" class="flex-1 flex items-center justify-center gap-2 px-4 py-2.5 bg-yellow-500 hover:bg-yellow-600 text-white rounded-lg text-sm font-semibold transition-all shadow-sm">
                                <svg xmlns="http://www.w3.org/2000/svg" class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 7a2 2 0 012 2m4 0a6 6 0 01-7.743 5.743L11 17H9v2H7v2H4a1 1 0 01-1-1v-2.586a1 1 0 01.293-.707l5.964-5.964A6 6 0 1121 9z" />
                                </svg>
                                Key
                            </button>
                            <button @click="deleteCert(cert.id)" class="flex items-center justify-center gap-2 px-4 py-2.5 border border-gray-200 text-red-500 hover:bg-red-50 hover:border-red-200 rounded-lg text-sm font-semibold transition-all">
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
    <!-- Upload Modal -->
    <div v-if="showModal" class="fixed inset-0 z-50 flex items-center justify-center bg-black/60 backdrop-blur-sm">
      <div class="bg-white rounded-2xl shadow-2xl w-full max-w-md transform transition-all scale-100 overflow-hidden">
        <!-- Gradient Header -->
        <div class="bg-orange-500 px-6 py-4 flex justify-between items-center">
            <h3 class="text-lg font-bold text-white">Add Custom Certificate</h3>
            <button @click="showModal = false" class="text-white/80 hover:text-white transition-colors text-2xl">&times;</button>
        </div>
        
        <div class="p-6 space-y-5">
            <div>
                <span class="block text-sm font-semibold text-gray-700 mb-2">Domain Name</span>
                <Combobox 
                    v-model="domain" 
                    :options="hostOptions" 
                    placeholder="example.com" 
                    color="orange"
                />
            </div>

            <div>
                <span class="block text-sm font-semibold text-gray-700 mb-2">Certificate (.crt)</span>
                <input type="file" @change="handleCertUpload" class="w-full border border-gray-200 rounded-xl px-4 py-3 text-gray-700 bg-gray-50/50 focus:outline-none focus:ring-2 focus:ring-orange-500/20 focus:border-orange-500 transition-all file:mr-4 file:py-2 file:px-4 file:rounded-lg file:border-0 file:text-sm file:font-semibold file:bg-orange-100 file:text-orange-700 hover:file:bg-orange-200">
            </div>

            <div>
                <span class="block text-sm font-semibold text-gray-700 mb-2">Private Key (.key)</span>
                <input type="file" @change="handleKeyUpload" class="w-full border border-gray-200 rounded-xl px-4 py-3 text-gray-700 bg-gray-50/50 focus:outline-none focus:ring-2 focus:ring-orange-500/20 focus:border-orange-500 transition-all file:mr-4 file:py-2 file:px-4 file:rounded-lg file:border-0 file:text-sm file:font-semibold file:bg-orange-100 file:text-orange-700 hover:file:bg-orange-200">
            </div>
        </div>

        <div class="px-6 py-4 bg-gray-50 border-t border-gray-100 flex justify-end gap-3">
          <button @click="showModal = false" class="px-5 py-2.5 text-gray-600 hover:text-gray-800 font-semibold transition-colors rounded-xl hover:bg-gray-100">Cancel</button>
          <button @click="uploadCert" class="bg-orange-500 hover:bg-orange-600 text-white px-6 py-2.5 rounded-xl shadow-lg shadow-orange-500/25 transition-all font-semibold">Upload</button>
        </div>
      </div>
    </div>

    <!-- Generate Modal -->
    <div v-if="showGenerateModal" class="fixed inset-0 z-50 flex items-center justify-center bg-black/60 backdrop-blur-sm">
      <div class="bg-white rounded-2xl shadow-2xl w-full max-w-md transform transition-all scale-100 overflow-hidden max-h-[90vh] flex flex-col">
        <!-- Gradient Header -->
        <div class="bg-orange-500 px-6 py-4 flex justify-between items-center flex-shrink-0">
            <h3 class="text-lg font-bold text-white">Generate Certificate</h3>
            <button @click="showGenerateModal = false" class="text-white/80 hover:text-white transition-colors text-2xl">&times;</button>
        </div>
        
        <div class="p-6 space-y-5 overflow-y-auto flex-1">
            <div>
                <span class="block text-sm font-semibold text-gray-700 mb-2">Domain Name</span>
                <Combobox 
                    v-model="generateDomain" 
                    :options="hostOptions" 
                    placeholder="example.com" 
                    color="orange"
                />
            </div>

            <div>
                <span class="block text-sm font-semibold text-gray-700 mb-2">Provider</span>
                <CustomSelect 
                    v-model="generateProvider" 
                    :options="providerOptions" 
                    color="orange"
                />
            </div>

            <div v-if="generateProvider !== 'selfsigned'">
                <label for="genEmail" class="block text-sm font-semibold text-gray-700 mb-2">Email Address <span class="text-red-500">*</span></label>
                <input id="genEmail" v-model="generateEmail" type="email" class="w-full border border-gray-200 rounded-xl px-4 py-3 text-gray-700 bg-gray-50/50 focus:outline-none focus:ring-2 focus:ring-orange-500/20 focus:border-orange-500 transition-all" placeholder="admin@example.com">
            </div>

            <div v-if="generateProvider !== 'selfsigned'" class="flex items-center gap-3 p-4 bg-gray-50 rounded-xl">
                <input type="checkbox" v-model="generateUseDNS" id="genUseDNS" class="w-4 h-4 rounded text-orange-600 focus:ring-orange-500">
                <label for="genUseDNS" class="text-sm font-semibold text-gray-700">Use DNS Challenge</label>
            </div>

            <div v-if="generateProvider !== 'selfsigned' && generateUseDNS" class="pl-4 border-l-4 border-orange-200 space-y-4 bg-orange-50/50 rounded-r-xl p-4">
                <div>
                    <span class="block text-xs font-semibold text-gray-600 mb-2">DNS Provider</span>
                    <CustomSelect 
                        v-model="generateDNSProvider" 
                        :options="dnsProviderOptions" 
                        color="orange"
                    />
                </div>
                <div>
                    <label for="dnsToken" class="block text-xs font-semibold text-gray-600 mb-2">API Token <span class="text-red-500">*</span></label>
                    <input id="dnsToken" v-model="generateDNSToken" type="password" class="w-full border border-gray-200 rounded-xl px-4 py-2.5 text-sm text-gray-700 bg-white focus:outline-none focus:ring-2 focus:ring-orange-500/20 focus:border-orange-500 transition-all">
                </div>
            </div>
        </div>

        <div class="px-6 py-4 bg-gray-50 border-t border-gray-100 flex justify-end gap-3 flex-shrink-0">
          <button @click="showGenerateModal = false" class="px-5 py-2.5 text-gray-600 hover:text-gray-800 font-semibold transition-colors rounded-xl hover:bg-gray-100">Cancel</button>
          <button @click="generateCert" :disabled="isGenerating" class="bg-orange-500 hover:bg-orange-600 text-white px-6 py-2.5 rounded-xl shadow-lg shadow-orange-500/25 transition-all font-semibold disabled:opacity-50 flex items-center gap-2">
              <svg v-if="isGenerating" class="animate-spin h-4 w-4" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24">
                  <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle>
                  <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
              </svg>
              <span>{{ isGenerating ? 'Generating...' : 'Generate' }}</span>
          </button>
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
