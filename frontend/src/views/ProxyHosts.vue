<script setup lang="ts">
import { ref, onMounted, computed, watch, onUnmounted } from 'vue'
import Combobox from '../components/Combobox.vue'
import CustomSelect from '../components/CustomSelect.vue'
import { toast } from 'vue-sonner'
import { useConfirm } from '../composables/useConfirm'
import { useWebSocket } from '../composables/useWebSocket'
import { authFetch } from '../composables/useAuth'

const { confirm } = useConfirm()
const { on, off, send, isConnected } = useWebSocket()

const isValidDomain = (domain: string): boolean => {
    if (!domain || domain.trim() === '') return false
    
    const domainPattern = /^(\*\.)?([a-zA-Z0-9]([a-zA-Z0-9\-]{0,61}[a-zA-Z0-9])?\.)*[a-zA-Z0-9]([a-zA-Z0-9\-]{0,61}[a-zA-Z0-9])?$/
    
    if (domain === 'localhost') return true
    
    if (!domainPattern.test(domain)) return false
    
    if (domain.length > 253) return false
    
    const labels = domain.replace(/^\*\./, '').split('.')
    for (const label of labels) {
        if (label.length > 63) return false
    }
    
    return true
}

const cleanDomain = (domain: string): string => {
    if (!domain) return domain
    let cleaned = domain.trim()
    cleaned = cleaned.replace(/^https?:\/\//i, '')
    cleaned = cleaned.replace(/:\d+/, '')
    cleaned = cleaned.replace(/\/.*$/, '')
    return cleaned
}

const cleanTarget = (target: string): string => {
    if (!target) return target
    let cleaned = target.trim()
    cleaned = cleaned.replace(/^https?:\/\//i, '')
    cleaned = cleaned.replace(/(:\d+)\/.*$/, '$1')
    cleaned = cleaned.replace(/\/.*$/, '')
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

interface Upstream {
    target: string
    weight: number
    max_fails: number
    fail_timeout: string
}

interface Location {
    path: string
    target: string
    upstreams?: Upstream[]
    load_balancing?: string
    lb_try_duration?: string
    lb_try_interval?: string
    health_check?: boolean
    health_check_path?: string
    health_check_interval?: string
}

interface Host {
  id: number
  domain: string
  target: string
  type: string
  ssl: boolean
  ssl_provider: string
  ssl_actual_provider?: string
  hsts_enabled: boolean
  block_exploits: boolean
  access_list_id?: number
  is_active: boolean
  cache_assets: boolean
  websockets: boolean
  forwarding_code: number
  incoming_port: number
  certificate_id?: number
  ssl_status?: string
  ssl_error?: string
  locations?: Location[]
  upstreams?: Upstream[]
  load_balancing?: string
  lb_try_duration?: string
  lb_try_interval?: string
  health_check?: boolean
  health_check_path?: string
  health_check_interval?: string
}

interface AccessList {
    id: number
    name: string
}

interface Certificate {
    id: number
    domain: string
}

const hosts = ref<Host[]>([])
const accessLists = ref<AccessList[]>([])
const certificates = ref<Certificate[]>([])
const expandedHosts = ref<number[]>([])
const sslErrorPinned = ref<number | null>(null)
const searchQuery = ref('')

const filteredHosts = computed(() => {
    if (!searchQuery.value.trim()) return hosts.value
    const q = searchQuery.value.toLowerCase()
    return hosts.value.filter(h => 
        h.domain.toLowerCase().includes(q) ||
        h.target?.toLowerCase().includes(q) ||
        h.type?.toLowerCase().includes(q) ||
        h.upstreams?.some(u => u.target.toLowerCase().includes(q)) ||
        h.locations?.some(l => l.target?.toLowerCase().includes(q) || l.path?.toLowerCase().includes(q))
    )
})

const copyToClipboard = async (text: string) => {
    try {
        await navigator.clipboard.writeText(text)
        toast.success('Copied to clipboard!')
    } catch (e) {
        toast.error('Failed to copy')
    }
}

const handleClickOutside = (event: MouseEvent) => {
    if (sslErrorPinned.value !== null) {
        const target = event.target as HTMLElement
        if (!target.closest('.ssl-error-tooltip')) {
            sslErrorPinned.value = null
        }
    }
}

const toggleExpandHost = (hostId: number) => {
    const idx = expandedHosts.value.indexOf(hostId)
    if (idx > -1) {
        expandedHosts.value.splice(idx, 1)
    } else {
        expandedHosts.value.push(hostId)
    }
}

const newHost = ref<{
    domain: string
    target: string
    type: string
    ssl: boolean
    ssl_provider: string
    hsts_enabled: boolean
    block_exploits: boolean
    cache_assets: boolean
    websockets: boolean
    forwarding_code: number
    incoming_port: number
    access_list_id: number | null
    is_active: boolean
    locations: Location[]
    upstreams: Upstream[]
    load_balancing: string
    lb_try_duration: string
    lb_try_interval: string
    health_check: boolean
    health_check_path: string
    health_check_interval: string
    certificate_id?: number
    use_dns_challenge?: boolean
    dns_provider?: string
    dns_token?: string
}>({ 
    domain: '', 
    target: '', 
    type: 'proxy', 
    ssl: true, 
    ssl_provider: 'auto', 
    hsts_enabled: false, 
    block_exploits: false, 
    cache_assets: false,
    websockets: true,
    forwarding_code: 301,
    incoming_port: 0,
    access_list_id: null, 
    is_active: true,
    locations: [],
    upstreams: [],
    load_balancing: '',
    lb_try_duration: '',
    lb_try_interval: '',
    health_check: false,
    health_check_path: '',
    health_check_interval: ''
})
const showModal = ref(false)
const isEditing = ref(false)
const editingId = ref<number | null>(null)
const activeTab = ref<'details' | 'ssl' | 'locations' | 'loadbalancer' | 'advanced'>('details')

watch(() => newHost.value.type, (newType, oldType) => {
    if (newType === 'loadbalancer' && oldType !== 'loadbalancer') {
        if (newHost.value.upstreams.length === 0) {
            newHost.value.upstreams.push({ target: '', weight: 1, max_fails: 0, fail_timeout: '' })
        }
        if (!newHost.value.load_balancing) {
            newHost.value.load_balancing = 'round_robin'
        }
        newHost.value.target = ''
    }
    if (newType === 'proxy' && oldType === 'loadbalancer') {
        newHost.value.upstreams = []
        newHost.value.load_balancing = ''
        newHost.value.health_check = false
        newHost.value.health_check_path = ''
        newHost.value.health_check_interval = ''
        newHost.value.lb_try_duration = ''
        newHost.value.lb_try_interval = ''
    }
    if (newType !== 'loadbalancer' && activeTab.value === 'loadbalancer') {
        activeTab.value = 'details'
    }
})

const fetchAccessLists = async () => {
    const res = await authFetch('/api/access-lists')
    if (res.ok) {
        accessLists.value = await res.json()
    }
}

const fetchCertificates = async () => {
    const res = await authFetch('/api/certificates')
    if (res.ok) {
        certificates.value = await res.json()
    }
}

const openAddModal = () => {
    isEditing.value = false
    editingId.value = null
    newHost.value = { 
        domain: '', 
        target: '', 
        type: 'proxy', 
        ssl: true, 
        ssl_provider: 'auto', 
        hsts_enabled: false, 
        block_exploits: false, 
        cache_assets: false,
        websockets: true,
        forwarding_code: 301,
        incoming_port: 0,
        access_list_id: null, 
        is_active: true,
        locations: [],
        upstreams: [],
        load_balancing: '',
        lb_try_duration: '',
        lb_try_interval: '',
        health_check: false,
        health_check_path: '',
        health_check_interval: ''
    }
    customCertFile.value = null
    customKeyFile.value = null
    activeTab.value = 'details'
    showModal.value = true
}

const openEditModal = (host: Host) => {
    isEditing.value = true
    editingId.value = host.id
    
    let uiType = host.type
    if (host.type === 'proxy' && host.upstreams && host.upstreams.length > 0) {
        uiType = 'loadbalancer'
    }
    
    newHost.value = { 
        ...host, 
        type: uiType,
        access_list_id: host.access_list_id || null,
        locations: host.locations ? host.locations.map(l => ({
            ...l,
            upstreams: l.upstreams ? l.upstreams.map(u => ({...u})) : [],
            load_balancing: l.load_balancing || '',
            lb_try_duration: l.lb_try_duration || '',
            lb_try_interval: l.lb_try_interval || '',
            health_check: l.health_check || false,
            health_check_path: l.health_check_path || '',
            health_check_interval: l.health_check_interval || ''
        })) : [],
        upstreams: host.upstreams ? host.upstreams.map(u => ({...u})) : [],
        load_balancing: host.load_balancing || '',
        lb_try_duration: host.lb_try_duration || '',
        lb_try_interval: host.lb_try_interval || '',
        health_check: host.health_check || false,
        health_check_path: host.health_check_path || '',
        health_check_interval: host.health_check_interval || ''
    }
    activeTab.value = 'details'
    showModal.value = true
}

const addLocation = () => {
    newHost.value.locations.push({ 
        path: '/', 
        target: '',
        upstreams: [],
        load_balancing: '',
        lb_try_duration: '',
        lb_try_interval: '',
        health_check: false,
        health_check_path: '',
        health_check_interval: ''
    })
}

const removeLocation = (index: number) => {
    newHost.value.locations.splice(index, 1)
}

const setLocationMode = (locIndex: number, mode: 'single' | 'loadbalancer') => {
    const loc = newHost.value.locations[locIndex]
    if (mode === 'single') {
        loc.upstreams = []
        loc.load_balancing = ''
    } else {
        if (!loc.upstreams || loc.upstreams.length === 0) {
            loc.upstreams = [{ target: '', weight: 1, max_fails: 0, fail_timeout: '' }]
        }
        if (!loc.load_balancing) {
            loc.load_balancing = 'round_robin'
        }
    }
}

const addLocationUpstream = (locIndex: number) => {
    const loc = newHost.value.locations[locIndex]
    if (!loc.upstreams) {
        loc.upstreams = []
    }
    loc.upstreams.push({ target: '', weight: 1, max_fails: 0, fail_timeout: '' })
}

const removeLocationUpstream = (locIndex: number, upstreamIndex: number) => {
    const loc = newHost.value.locations[locIndex]
    if (loc.upstreams) {
        loc.upstreams.splice(upstreamIndex, 1)
        if (loc.upstreams.length === 0) {
            loc.load_balancing = ''
        }
    }
}

const addUpstream = () => {
    newHost.value.upstreams.push({ target: '', weight: 1, max_fails: 0, fail_timeout: '' })
}

const removeUpstream = (index: number) => {
    newHost.value.upstreams.splice(index, 1)
}

const loadBalancingOptions = [
    { label: 'Round Robin', value: 'round_robin' },
    { label: 'Least Connections', value: 'least_conn' },
    { label: 'IP Hash (Sticky)', value: 'ip_hash' },
    { label: 'Random', value: 'random' },
    { label: 'Random Choose 2', value: 'random_choose' },
    { label: 'First Available', value: 'first' },
    { label: 'URI Hash', value: 'uri_hash' },
    { label: 'Cookie (Sticky)', value: 'cookie' }
]

const saveHost = async () => {
  if (!newHost.value.domain) {
      toast.error('Domain Name is required')
      return
  }
  if (!isValidDomain(newHost.value.domain)) {
      toast.error('Invalid Domain Name format. Use a valid domain (e.g., example.com, sub.example.com, *.example.com)')
      return
  }
  if (newHost.value.type === 'proxy' && !newHost.value.target) {
      toast.error('Forward Hostname / IP is required')
      return
  }
  if (newHost.value.type === 'proxy' && newHost.value.target && !isValidTarget(newHost.value.target)) {
      toast.error('Invalid Forward Hostname / IP format. Use format: hostname:port or ip:port (e.g., example.com:8080 or 192.168.1.1:80)')
      return
  }
  if (newHost.value.type === 'loadbalancer' && newHost.value.upstreams.length === 0) {
      toast.error('At least one upstream server is required for Load Balancer')
      return
  }
  if (newHost.value.type === 'loadbalancer') {
      const invalidUpstream = newHost.value.upstreams.find(u => u.target && !isValidTarget(u.target))
      if (invalidUpstream) {
          toast.error(`Invalid upstream target: ${invalidUpstream.target}. Use format: hostname:port or ip:port`)
          return
      }
  }
  for (const loc of newHost.value.locations) {
      if (loc.target && !isValidTarget(loc.target)) {
          toast.error(`Invalid location target for path ${loc.path}: ${loc.target}. Use format: hostname:port or ip:port`)
          return
      }
      if (loc.upstreams && loc.upstreams.length > 0) {
          const invalidLocUpstream = loc.upstreams.find(u => u.target && !isValidTarget(u.target))
          if (invalidLocUpstream) {
              toast.error(`Invalid upstream target in location ${loc.path}: ${invalidLocUpstream.target}. Use format: hostname:port or ip:port`)
              return
          }
      }
  }
  if (newHost.value.type === 'redirect' && !newHost.value.forwarding_code) {
      toast.error('Forwarding Code is required')
      return
  }
  
  const payload = { ...newHost.value }
  payload.ssl = payload.ssl_provider !== 'none'
  
  if (payload.type === 'loadbalancer') {
      payload.type = 'proxy'
      payload.target = ''
  }
  
  payload.locations = payload.locations.filter((l: Location) => l.path && (l.target || (l.upstreams && l.upstreams.length > 0)))
    .map((l: Location) => {
      const isLoadBalanced = l.upstreams && l.upstreams.length > 0
      return {
        ...l,
        target: isLoadBalanced ? '' : l.target,
        upstreams: isLoadBalanced ? l.upstreams!.filter((u: Upstream) => u.target) : []
      }
    })
  
  payload.upstreams = payload.upstreams.filter((u: Upstream) => u.target)

  const url = isEditing.value ? `/api/hosts/${editingId.value}` : '/api/hosts'
  const method = isEditing.value ? 'PUT' : 'POST'

  try {
      const res = await authFetch(url, {
        method: method,
        headers: { 
          'Content-Type': 'application/json'
        },
        body: JSON.stringify(payload)
      })

        if (res.ok) {
        showModal.value = false
        toast.success(isEditing.value ? 'Host updated successfully' : 'Host created successfully')
        newHost.value = { 
            domain: '', 
            target: '', 
            type: 'proxy', 
            ssl: true, 
            ssl_provider: 'auto', 
            hsts_enabled: false, 
            block_exploits: false, 
            cache_assets: false,
            websockets: true,
            forwarding_code: 301,
            incoming_port: 0,
            access_list_id: null, 
            is_active: true,
            locations: [],
            upstreams: [],
            load_balancing: '',
            lb_try_duration: '',
            lb_try_interval: '',
            health_check: false,
            health_check_path: '',
            health_check_interval: ''
        }
      } else {
        const data = await res.json()
        toast.error(data.error || 'Failed to save host')
      }
  } catch (e) {
      toast.error('Connection error')
  }
}

const deleteHost = async (id: number) => {
    const confirmed = await confirm(
        'Delete Host', 
        'Are you sure you want to delete this host? This action cannot be undone.',
        { type: 'danger', confirmText: 'Delete' }
    )
    if(!confirmed) return;

    const res = await authFetch(`/api/hosts/${id}`, {
        method: 'DELETE'
    })
    
    if (res.ok) {
        toast.success('Host deleted successfully')
    } else {
        toast.error('Failed to delete host')
    }
}

const getAccessListName = (id: number) => {
    const list = accessLists.value.find((l: AccessList) => l.id === id)
    return list ? list.name : 'Protected'
}

const getSSLProviderName = (provider: string, actualProvider?: string) => {
    const providerNames: Record<string, string> = {
        'letsencrypt': "Let's Encrypt",
        'zerossl': 'ZeroSSL',
        'selfsigned': 'Self Signed',
        'custom': 'Custom',
        'none': 'None'
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

const toggleStatus = async (host: Host) => {
    const newStatus = !host.is_active
    
    host.is_active = newStatus

    const res = await authFetch(`/api/hosts/${host.id}`, {
        method: 'PUT',
        headers: { 
            'Content-Type': 'application/json'
        },
        body: JSON.stringify(host)
    })

    if (!res.ok) {
        host.is_active = !newStatus
        toast.error("Failed to update status")
    } else {
        toast.success(`Host ${newStatus ? 'enabled' : 'disabled'}`)
    }
}

const customCertFile = ref<File | null>(null)
const customKeyFile = ref<File | null>(null)
const useExistingCert = ref(false)
const certificateOptions = computed(() => {
    return certificates.value.map(c => ({ label: c.domain, value: c.id }))
})

const handleCertUpload = (event: Event) => {
    const target = event.target as HTMLInputElement
    if (target.files && target.files.length > 0) {
        customCertFile.value = target.files[0]
    }
}

const handleKeyUpload = (event: Event) => {
    const target = event.target as HTMLInputElement
    if (target.files && target.files.length > 0) {
        customKeyFile.value = target.files[0]
    }
}

const schemeOptions = [
    { label: 'http/https', value: 'http' }
]

const typeOptions = [
    { label: 'Reverse Proxy', value: 'proxy' },
    { label: 'Load Balancer', value: 'loadbalancer' },
    { label: 'Redirect', value: 'redirect' }
]

const forwardingCodeOptions = [
    { label: '301 (Permanent)', value: 301 },
    { label: '302 (Found)', value: 302 },
    { label: '300 (Multiple Choices)', value: 300 },
    { label: '303 (See Other)', value: 303 },
    { label: '307 (Temporary)', value: 307 },
    { label: '308 (Permanent)', value: 308 }
]

const dnsProviderOptions = [
    { label: 'Cloudflare', value: 'cloudflare' }
]

const sslProviderOptions = [
    { label: 'Auto', value: 'auto' },
    { label: 'Let\'s Encrypt', value: 'letsencrypt' },
    { label: 'ZeroSSL', value: 'zerossl' },
    { label: 'Self Signed (Internal)', value: 'selfsigned' },
    { label: 'Custom (Uploaded Certificate)', value: 'custom' },
    { label: 'None (HTTP Only)', value: 'none' }
]

const hostOptions = computed(() => hosts.value.map(h => h.domain))

const accessListOptions = computed(() => {
    const opts: Array<{ label: string; value: number | null }> = [{ label: 'None (Public)', value: null }]
    return opts.concat(accessLists.value.map(l => ({ label: l.name, value: l.id })))
})

const handleHostUpdated = (updatedHost: Host) => {
    const index = hosts.value.findIndex(h => h.id === updatedHost.id)
    if (index !== -1) {
        const newHosts = [...hosts.value]
        newHosts[index] = { ...newHosts[index], ...updatedHost }
        hosts.value = newHosts
    } else {
        hosts.value.unshift(updatedHost)
    }
}

const handleHostCreated = (newHost: Host) => {
    if (!hosts.value.find(h => h.id === newHost.id)) {
        hosts.value.unshift(newHost)
    }
}

const handleHostDeleted = (payload: any) => {
    const idToDelete = Number(payload.id)
    hosts.value = hosts.value.filter(h => h.id !== idToDelete)
}

const fetchHosts = async () => {
    try {
        const res = await authFetch('/api/hosts')
        if (res.ok) {
            hosts.value = await res.json()
        }
    } catch (e) {
        console.error('Failed to fetch hosts:', e)
    }
}

let checkConnInterval: ReturnType<typeof setInterval> | null = null

watch(isConnected, (connected: boolean, wasConnected: boolean) => {
    if (connected && !wasConnected) {
        fetchHosts()
    }
})

onMounted(() => {
    fetchAccessLists()
    fetchCertificates()
    fetchHosts()
    
    document.addEventListener('click', handleClickOutside)
    
    on('hosts_list', (data: Host[]) => {
        hosts.value = data
    })

    if (isConnected.value) {
        send('get_hosts')
    } else {
        checkConnInterval = setInterval(() => {
            if (isConnected.value) {
                send('get_hosts')
                if (checkConnInterval) {
                    clearInterval(checkConnInterval)
                    checkConnInterval = null
                }
            }
        }, 500)
        
        setTimeout(() => {
            if (checkConnInterval) {
                clearInterval(checkConnInterval)
                checkConnInterval = null
            }
        }, 5000)
    }

    on('host_updated', handleHostUpdated)
    on('host_created', handleHostCreated)
    on('host_deleted', handleHostDeleted)
})

onUnmounted(() => {
    if (checkConnInterval) {
        clearInterval(checkConnInterval)
        checkConnInterval = null
    }
    off('hosts_list')
    off('host_updated', handleHostUpdated)
    off('host_created', handleHostCreated)
    off('host_deleted', handleHostDeleted)
    document.removeEventListener('click', handleClickOutside)
})
</script>

<template>
  <div>

    <!-- Header -->
    <div class="bg-white rounded-t-xl p-5 border-b border-gray-200 shadow-sm overflow-hidden relative">
        <!-- Gradient accent bar -->
        <div class="absolute top-0 left-0 right-0 h-1 bg-green-500"></div>
        
        <div class="flex flex-col lg:flex-row justify-between items-start lg:items-center gap-4">
            <div class="flex items-center gap-3">
                <div>
                    <h2 class="text-2xl font-bold text-gray-800 tracking-tight">Proxy Hosts</h2>
                    <p class="text-sm text-gray-500 mt-1">Route traffic to your backend services</p>
                </div>
            </div>
            <div class="flex items-center gap-3 w-full lg:w-auto">
                <div class="relative flex-1 lg:flex-initial">
                    <input 
                        v-model="searchQuery" 
                        type="text" 
                        placeholder="Search hosts..." 
                        class="w-full lg:w-64 pl-10 pr-4 py-2.5 border border-gray-200 rounded-xl text-sm focus:ring-2 focus:ring-green-500/20 focus:border-green-500 transition-all bg-gray-50/50"
                    />
                    <svg xmlns="http://www.w3.org/2000/svg" class="h-4 w-4 text-gray-400 absolute left-3.5 top-1/2 -translate-y-1/2" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z" />
                    </svg>
                </div>
                <button @click="openAddModal" class="bg-green-500 hover:bg-green-600 text-white px-5 py-2.5 rounded-xl shadow-lg shadow-green-500/25 hover:shadow-green-500/40 transition-all duration-300 font-semibold text-sm flex items-center gap-2 whitespace-nowrap">
                    <svg xmlns="http://www.w3.org/2000/svg" class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4v16m8-8H4" />
                    </svg>
                    <span class="hidden sm:inline">Add Proxy Host</span>
                    <span class="sm:hidden">Add</span>
                </button>
            </div>
        </div>
    </div>

    <!-- Column Headers -->
    <div class="bg-gray-50/80 px-5 py-3 border-b text-xs font-bold text-gray-500 uppercase tracking-wider border-x border-gray-200 hidden lg:grid lg:grid-cols-12 gap-4">
        <div class="col-span-3 pl-14">Source</div>
        <div class="col-span-3">Destination</div>
        <div class="col-span-2">SSL</div>
        <div class="col-span-2">Access</div>
        <div class="col-span-2 text-right pr-2">Status</div>
    </div>

    <!-- List -->
    <div class="bg-white rounded-b-xl shadow-lg border-x border-b border-gray-200">
        <div v-if="hosts.length === 0" class="p-12 text-center">
            <div class="mb-4">
                <div class="mx-auto w-20 h-20 rounded-2xl bg-green-500 flex items-center justify-center shadow-lg shadow-green-500/30">
                    <svg xmlns="http://www.w3.org/2000/svg" class="h-10 w-10 text-white" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M21 12a9 9 0 01-9 9m9-9a9 9 0 00-9-9m9 9H3m9 9a9 9 0 01-9-9m9 9c1.657 0 3-4.03 3-9s-1.343-9-3-9m0 18c-1.657 0-3-4.03-3-9s1.343-9 3-9m-9 9a9 9 0 019-9" />
                    </svg>
                </div>
            </div>
            <p class="text-lg font-bold text-gray-700 mb-2">No proxy hosts configured</p>
            <p class="text-sm text-gray-400 mb-6">Create a proxy host to route traffic to your backend services</p>
            <button @click="openAddModal" class="inline-flex items-center gap-2 bg-green-500 hover:bg-green-600 text-white px-5 py-2.5 rounded-xl shadow-lg shadow-green-500/25 transition-all text-sm font-semibold">
                <svg xmlns="http://www.w3.org/2000/svg" class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4v16m8-8H4" />
                </svg>
                Add Proxy Host
            </button>
        </div>
        <div v-else-if="filteredHosts.length === 0" class="p-12 text-center">
            <div class="mb-4">
                <div class="mx-auto w-20 h-20 rounded-2xl bg-gray-100 flex items-center justify-center">
                    <svg xmlns="http://www.w3.org/2000/svg" class="h-10 w-10 text-gray-400" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z" />
                    </svg>
                </div>
            </div>
            <p class="text-lg font-bold text-gray-700 mb-2">No results found</p>
            <p class="text-sm text-gray-400">No hosts match "{{ searchQuery }}"</p>
        </div>

        <div v-for="host in filteredHosts" :key="host.id" class="group p-4 border-b border-gray-100 hover:bg-gray-50/50 transition-all duration-200 last:border-b-0">
            <!-- Desktop Layout -->
            <div class="hidden lg:grid lg:grid-cols-12 gap-4 items-center">
                <!-- Source -->
                <div class="col-span-3 flex items-center gap-3 overflow-hidden">
                    <div class="flex-shrink-0 w-10 h-10 rounded-xl bg-green-500 flex items-center justify-center shadow-lg shadow-green-500/20 group-hover:shadow-green-500/30 transition-all">
                        <svg xmlns="http://www.w3.org/2000/svg" class="h-5 w-5 text-white" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M21 12a9 9 0 01-9 9m9-9a9 9 0 00-9-9m9 9H3m9 9a9 9 0 01-9-9m9 9c1.657 0 3-4.03 3-9s-1.343-9-3-9m0 18c-1.657 0-3-4.03-3-9s1.343-9 3-9m-9 9a9 9 0 019-9" />
                        </svg>
                    </div>
                    <div class="min-w-0">
                        <p class="text-sm font-bold text-gray-900 truncate">{{ host.domain }}</p>
                        <a :href="(host.ssl ? 'https://' : 'http://') + host.domain" target="_blank" class="text-xs text-blue-500 hover:underline flex items-center gap-1">
                            Visit site
                            <svg xmlns="http://www.w3.org/2000/svg" class="h-3 w-3" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M10 6H6a2 2 0 00-2 2v10a2 2 0 002 2h10a2 2 0 002-2v-4M14 4h6m0 0v6m0-6L10 14" />
                            </svg>
                        </a>
                    </div>
                </div>

                <!-- Destination -->
                <div class="col-span-3">
                    <!-- Redirect -->
                    <template v-if="host.type === 'redirect'">
                        <div class="inline-flex items-center gap-2 bg-orange-50 border border-orange-200 rounded-lg px-3 py-1.5">
                            <svg xmlns="http://www.w3.org/2000/svg" class="h-4 w-4 text-orange-500" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13 7l5 5m0 0l-5 5m5-5H6" />
                            </svg>
                            <span class="font-mono text-xs text-gray-700 break-all">{{ host.target || '-' }}</span>
                            <span class="text-[10px] font-semibold text-orange-600 bg-orange-100 px-1.5 py-0.5 rounded">{{ host.forwarding_code || 301 }}</span>
                        </div>
                    </template>
                    
                    <!-- Single Target (Reverse Proxy) -->
                    <template v-else-if="!host.upstreams || host.upstreams.length === 0">
                        <div class="inline-flex items-center gap-2 bg-blue-50 border border-blue-200 rounded-lg px-3 py-1.5">
                            <svg xmlns="http://www.w3.org/2000/svg" class="h-4 w-4 text-blue-500" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M7 16V4m0 0L3 8m4-4l4 4m6 0v12m0 0l4-4m-4 4l-4-4" />
                            </svg>
                            <span class="font-mono text-xs text-gray-700 break-all">{{ host.target || '-' }}</span>
                        </div>
                    </template>
                    
                    <!-- Load Balancer -->
                    <template v-else>
                        <div class="bg-purple-50 border border-purple-200 rounded-lg p-2.5 space-y-2">
                            <div class="flex items-center justify-between">
                                <div class="flex items-center gap-2">
                                    <div class="flex items-center justify-center w-6 h-6 bg-purple-500 rounded-md">
                                        <svg xmlns="http://www.w3.org/2000/svg" class="h-3.5 w-3.5 text-white" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                                            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15" />
                                        </svg>
                                    </div>
                                    <span class="text-xs font-bold text-purple-700">Load Balancer</span>
                                </div>
                                <span class="text-[10px] font-semibold text-purple-600 bg-purple-100 px-2 py-0.5 rounded-full">
                                    {{ host.upstreams.length }} {{ host.upstreams.length === 1 ? 'server' : 'servers' }}
                                </span>
                            </div>
                            <div class="space-y-1">
                                <div v-for="(upstream, idx) in host.upstreams.slice(0, expandedHosts.includes(host.id) ? host.upstreams.length : 2)" 
                                     :key="idx" 
                                     class="flex items-center gap-2 bg-white/70 backdrop-blur-sm px-2 py-1 rounded-md border border-purple-100">
                                    <div class="flex-shrink-0 w-1.5 h-1.5 rounded-full bg-green-400 animate-pulse"></div>
                                    <span class="font-mono text-[11px] text-gray-700 break-all flex-1">{{ upstream.target }}</span>
                                    <span v-if="upstream.weight > 1" class="text-[9px] text-gray-400 bg-gray-100 px-1.5 py-0.5 rounded">w:{{ upstream.weight }}</span>
                                </div>
                            </div>
                            <button v-if="host.upstreams.length > 2" 
                                    @click.stop="toggleExpandHost(host.id)" 
                                    class="w-full text-[10px] text-purple-600 hover:text-purple-800 font-medium py-1 bg-purple-100/50 hover:bg-purple-100 rounded-md transition-colors flex items-center justify-center gap-1">
                                <svg v-if="!expandedHosts.includes(host.id)" xmlns="http://www.w3.org/2000/svg" class="h-3 w-3" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 9l-7 7-7-7" />
                                </svg>
                                <svg v-else xmlns="http://www.w3.org/2000/svg" class="h-3 w-3" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 15l7-7 7 7" />
                                </svg>
                                {{ expandedHosts.includes(host.id) ? 'Show less' : `Show ${host.upstreams.length - 2} more` }}
                            </button>
                        </div>
                    </template>
                </div>

                <!-- SSL -->
                <div class="col-span-2">
                    <div v-if="host.ssl_status === 'generating'" 
                         class="inline-flex items-center gap-1.5 bg-yellow-50 border border-yellow-200 text-yellow-700 px-2.5 py-1 rounded-full text-[10px] font-semibold animate-pulse">
                        <svg class="animate-spin h-3 w-3" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24">
                            <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle>
                            <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
                        </svg>
                        Generating
                    </div>
                    <div v-else-if="host.ssl_status === 'failed'" 
                         class="ssl-error-tooltip group/ssl relative inline-flex items-center gap-1.5 bg-red-50 border border-red-200 text-red-700 px-2.5 py-1 rounded-full text-[10px] font-semibold cursor-pointer" 
                         @click.stop="sslErrorPinned = sslErrorPinned === host.id ? null : host.id">
                        <svg xmlns="http://www.w3.org/2000/svg" class="h-3 w-3" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 8v4m0 4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
                        </svg>
                        Failed
                        <div :class="['ssl-error-tooltip absolute bottom-full left-1/2 transform -translate-x-1/2 mb-2 w-64 max-w-xs bg-gray-800 text-white text-xs rounded p-2 transition-opacity z-50 text-center shadow-lg break-words select-text', sslErrorPinned === host.id ? 'opacity-100' : 'opacity-0 group-hover/ssl:opacity-100 pointer-events-none']">
                            <div class="flex items-center justify-between gap-2 mb-1" v-if="sslErrorPinned === host.id">
                                <span class="text-gray-400 text-[10px]">Click to copy</span>
                                <button @click.stop="sslErrorPinned = null" class="text-gray-400 hover:text-white">
                                    <svg xmlns="http://www.w3.org/2000/svg" class="h-3 w-3" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" /></svg>
                                </button>
                            </div>
                            <div @click.stop="copyToClipboard(host.ssl_error || 'Unknown error')" class="cursor-pointer hover:bg-gray-700 rounded p-1 -m-1">
                                {{ host.ssl_error || 'Unknown error' }}
                            </div>
                            <div class="absolute top-full left-1/2 transform -translate-x-1/2 border-4 border-transparent border-t-gray-800"></div>
                        </div>
                    </div>
                    <!-- Fallback SSL: ACME failed but using Self-Signed -->
                    <div v-else-if="host.ssl && host.ssl_provider === 'auto' && host.ssl_actual_provider === 'selfsigned'" 
                         class="ssl-error-tooltip group/ssl relative inline-flex items-center gap-1.5 bg-orange-50 border border-orange-200 text-orange-700 px-2.5 py-1 rounded-full text-[10px] font-semibold cursor-pointer" 
                         @click.stop="sslErrorPinned = sslErrorPinned === host.id ? null : host.id">
                        <svg xmlns="http://www.w3.org/2000/svg" class="h-3 w-3" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-3L13.732 4c-.77-1.333-2.694-1.333-3.464 0L3.34 16c-.77 1.333.192 3 1.732 3z" />
                        </svg>
                        Auto (Self Signed)
                        <div :class="['ssl-error-tooltip absolute bottom-full left-1/2 transform -translate-x-1/2 mb-2 w-64 max-w-xs bg-gray-800 text-white text-xs rounded p-2 transition-opacity z-50 text-center shadow-lg break-words select-text', sslErrorPinned === host.id ? 'opacity-100' : 'opacity-0 group-hover/ssl:opacity-100 pointer-events-none']">
                            <div class="flex items-center justify-between gap-2 mb-1" v-if="sslErrorPinned === host.id">
                                <span class="text-gray-400 text-[10px]">Click to copy</span>
                                <button @click.stop="sslErrorPinned = null" class="text-gray-400 hover:text-white">
                                    <svg xmlns="http://www.w3.org/2000/svg" class="h-3 w-3" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" /></svg>
                                </button>
                            </div>
                            <div @click.stop="copyToClipboard(host.ssl_error || 'Public SSL failed, using self-signed certificate')" class="cursor-pointer hover:bg-gray-700 rounded p-1 -m-1">
                                ⚠️ {{ host.ssl_error || 'Public SSL (Let\'s Encrypt/ZeroSSL) failed. Using self-signed certificate as fallback.' }}
                            </div>
                            <div class="absolute top-full left-1/2 transform -translate-x-1/2 border-4 border-transparent border-t-gray-800"></div>
                        </div>
                    </div>
                    <div v-else-if="host.ssl" 
                         class="inline-flex items-center gap-1.5 bg-green-50 border border-green-200 text-green-700 px-2.5 py-1 rounded-full text-[10px] font-semibold">
                        <svg xmlns="http://www.w3.org/2000/svg" class="h-3 w-3" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 12l2 2 4-4m5.618-4.016A11.955 11.955 0 0112 2.944a11.955 11.955 0 01-8.618 3.04A12.02 12.02 0 003 9c0 5.591 3.824 10.29 9 11.622 5.176-1.332 9-6.03 9-11.622 0-1.042-.133-2.052-.382-3.016z" />
                        </svg>
                        {{ getSSLProviderName(host.ssl_provider, host.ssl_actual_provider) }}
                    </div>
                    <div v-else 
                         class="inline-flex items-center gap-1.5 bg-gray-100 border border-gray-200 text-gray-600 px-2.5 py-1 rounded-full text-[10px] font-semibold">
                        <svg xmlns="http://www.w3.org/2000/svg" class="h-3 w-3" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M8 11V7a4 4 0 118 0m-4 8v2m-6 4h12a2 2 0 002-2v-6a2 2 0 00-2-2H6a2 2 0 00-2 2v6a2 2 0 002 2z" />
                        </svg>
                        No SSL
                    </div>
                </div>

                <!-- Access -->
                <div class="col-span-2">
                    <div v-if="host.access_list_id" 
                         class="inline-flex items-center gap-1.5 bg-blue-50 border border-blue-200 text-blue-700 px-2.5 py-1 rounded-full text-[10px] font-semibold">
                        <svg xmlns="http://www.w3.org/2000/svg" class="h-3 w-3" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 15v2m-6 4h12a2 2 0 002-2v-6a2 2 0 00-2-2H6a2 2 0 00-2 2v6a2 2 0 002 2zm10-10V7a4 4 0 00-8 0v4h8z" />
                        </svg>
                        {{ getAccessListName(host.access_list_id) }}
                    </div>
                    <div v-else 
                         class="inline-flex items-center gap-1.5 bg-gray-100 border border-gray-200 text-gray-600 px-2.5 py-1 rounded-full text-[10px] font-semibold">
                        <svg xmlns="http://www.w3.org/2000/svg" class="h-3 w-3" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M3.055 11H5a2 2 0 012 2v1a2 2 0 002 2 2 2 0 012 2v2.945M8 3.935V5.5A2.5 2.5 0 0010.5 8h.5a2 2 0 012 2 2 2 0 104 0 2 2 0 012-2h1.064M15 20.488V18a2 2 0 012-2h3.064M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
                        </svg>
                        Public
                    </div>
                </div>

                <!-- Status & Actions -->
                <div class="col-span-2 flex items-center justify-end gap-3">
                    <button @click="toggleStatus(host)" :class="host.is_active ? 'bg-green-500' : 'bg-gray-300'" class="relative inline-flex h-6 w-11 items-center rounded-full transition-all focus:outline-none focus:ring-2 focus:ring-green-500 focus:ring-offset-2">
                        <span :class="host.is_active ? 'translate-x-6' : 'translate-x-1'" class="inline-block h-4 w-4 transform rounded-full bg-white transition-transform shadow-sm" />
                    </button>
                    <div class="flex items-center gap-1">
                        <button @click="openEditModal(host)" class="p-2 text-gray-400 hover:text-green-600 hover:bg-green-50 rounded-lg transition-all" title="Edit">
                            <svg xmlns="http://www.w3.org/2000/svg" class="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M11 5H6a2 2 0 00-2 2v11a2 2 0 002 2h11a2 2 0 002-2v-5m-1.414-9.414a2 2 0 112.828 2.828L11.828 15H9v-2.828l8.586-8.586z" />
                            </svg>
                        </button>
                        <button @click="deleteHost(host.id)" class="p-2 text-gray-400 hover:text-red-600 hover:bg-red-50 rounded-lg transition-all" title="Delete">
                            <svg xmlns="http://www.w3.org/2000/svg" class="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16" />
                            </svg>
                        </button>
                    </div>
                </div>
            </div>

            <!-- Mobile Layout - Premium Card Style -->
            <div class="lg:hidden">
                <!-- Card Container with gradient border effect -->
                <div class="relative rounded-xl overflow-hidden">
                    <!-- Status indicator strip -->
                    <div :class="host.is_active ? 'bg-green-500' : 'bg-gray-300'" class="absolute top-0 left-0 right-0 h-1"></div>
                    
                    <!-- Main Content -->
                    <div class="pt-3 space-y-4">
                        <!-- Header: Domain + Status Badge -->
                        <div class="flex items-start justify-between gap-3">
                            <div class="flex items-center gap-3 min-w-0 flex-1">
                                <!-- Animated Icon Container -->
                                <div class="relative flex-shrink-0">
                                    <div :class="host.is_active 
                                        ? 'bg-green-500 shadow-lg shadow-green-200' 
                                        : 'bg-gray-300'" 
                                        class="w-12 h-12 rounded-xl flex items-center justify-center transition-all duration-300">
                                        <svg xmlns="http://www.w3.org/2000/svg" class="h-6 w-6 text-white" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                                            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M21 12a9 9 0 01-9 9m9-9a9 9 0 00-9-9m9 9H3m9 9a9 9 0 01-9-9m9 9c1.657 0 3-4.03 3-9s-1.343-9-3-9m0 18c-1.657 0-3-4.03-3-9s1.343-9 3-9m-9 9a9 9 0 019-9" />
                                        </svg>
                                    </div>
                                    <!-- Online pulse indicator -->
                                    <span v-if="host.is_active" class="absolute -top-1 -right-1 flex h-3 w-3">
                                        <span class="animate-ping absolute inline-flex h-full w-full rounded-full bg-green-400 opacity-75"></span>
                                        <span class="relative inline-flex rounded-full h-3 w-3 bg-green-500 border-2 border-white"></span>
                                    </span>
                                </div>
                                
                                <div class="min-w-0 flex-1">
                                    <div class="flex items-center gap-2">
                                        <p class="text-base font-bold text-gray-900 truncate">{{ host.domain }}</p>
                                    </div>
                                    <div class="flex items-center gap-2 mt-0.5">
                                        <span :class="host.is_active ? 'text-green-600 bg-green-50' : 'text-gray-500 bg-gray-100'" 
                                              class="text-[10px] font-bold uppercase tracking-wider px-2 py-0.5 rounded-full">
                                            {{ host.is_active ? '● Online' : '○ Offline' }}
                                        </span>
                                        <a :href="(host.ssl ? 'https://' : 'http://') + host.domain" target="_blank" 
                                           class="text-[10px] text-blue-500 hover:text-blue-700 flex items-center gap-0.5 font-medium">
                                            Visit
                                            <svg xmlns="http://www.w3.org/2000/svg" class="h-3 w-3" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                                                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M10 6H6a2 2 0 00-2 2v10a2 2 0 002 2h10a2 2 0 002-2v-4M14 4h6m0 0v6m0-6L10 14" />
                                            </svg>
                                        </a>
                                    </div>
                                </div>
                            </div>
                            
                            <!-- Toggle Switch -->
                            <button @click="toggleStatus(host)" 
                                    :class="host.is_active ? 'bg-green-500' : 'bg-gray-300'" 
                                    class="relative inline-flex h-7 w-12 flex-shrink-0 items-center rounded-full transition-colors duration-200 focus:outline-none focus:ring-2 focus:ring-green-500 focus:ring-offset-2">
                                <span :class="host.is_active ? 'translate-x-6' : 'translate-x-1'" 
                                      class="inline-block h-5 w-5 transform rounded-full bg-white transition-transform duration-200 shadow-md" />
                            </button>
                        </div>

                        <!-- Destination Card -->
                        <div class="bg-gray-50 rounded-xl p-3 border border-gray-200/50">
                            <div class="flex items-center gap-2 mb-2">
                                <svg xmlns="http://www.w3.org/2000/svg" class="h-4 w-4 text-gray-400" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13 10V3L4 14h7v7l9-11h-7z" />
                                </svg>
                                <span class="text-xs font-semibold text-gray-500 uppercase tracking-wider">Destination</span>
                            </div>
                            
                            <template v-if="host.type === 'redirect'">
                                <div class="flex items-center gap-2 bg-white rounded-lg px-3 py-2 shadow-sm border border-orange-100">
                                    <div class="flex-shrink-0 w-8 h-8 rounded-lg bg-orange-500 flex items-center justify-center">
                                        <svg xmlns="http://www.w3.org/2000/svg" class="h-4 w-4 text-white" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                                            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13 7l5 5m0 0l-5 5m5-5H6" />
                                        </svg>
                                    </div>
                                    <div class="min-w-0 flex-1">
                                        <p class="text-[10px] text-orange-600 font-semibold uppercase tracking-wide">Redirect {{ host.forwarding_code || 301 }}</p>
                                        <p class="font-mono text-sm text-gray-800 truncate">{{ host.target || '-' }}</p>
                                    </div>
                                </div>
                            </template>
                            
                            <template v-else-if="!host.upstreams || host.upstreams.length === 0">
                                <div class="flex items-center gap-2 bg-white rounded-lg px-3 py-2 shadow-sm border border-blue-100">
                                    <div class="flex-shrink-0 w-8 h-8 rounded-lg bg-blue-500 flex items-center justify-center">
                                        <svg xmlns="http://www.w3.org/2000/svg" class="h-4 w-4 text-white" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                                            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 12h14M12 5l7 7-7 7" />
                                        </svg>
                                    </div>
                                    <div class="min-w-0 flex-1">
                                        <p class="text-[10px] text-blue-600 font-semibold uppercase tracking-wide">Reverse Proxy</p>
                                        <p class="font-mono text-sm text-gray-800 truncate">{{ host.target || '-' }}</p>
                                    </div>
                                </div>
                            </template>
                            
                            <template v-else>
                                <div class="bg-white rounded-lg px-3 py-2 shadow-sm border border-purple-100">
                                    <div class="flex items-center gap-2 mb-2">
                                        <div class="flex-shrink-0 w-8 h-8 rounded-lg bg-purple-500 flex items-center justify-center">
                                            <svg xmlns="http://www.w3.org/2000/svg" class="h-4 w-4 text-white" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                                                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15" />
                                            </svg>
                                        </div>
                                        <div class="flex-1">
                                            <p class="text-[10px] text-purple-600 font-semibold uppercase tracking-wide">Load Balancer</p>
                                            <p class="text-xs text-gray-500">{{ host.upstreams.length }} upstream servers</p>
                                        </div>
                                    </div>
                                    <div class="space-y-1 pl-10">
                                        <div v-for="(upstream, idx) in host.upstreams.slice(0, 2)" :key="idx" 
                                             class="flex items-center gap-2 text-xs">
                                            <span class="w-1.5 h-1.5 rounded-full bg-green-400"></span>
                                            <span class="font-mono text-gray-600 truncate">{{ upstream.target }}</span>
                                        </div>
                                        <p v-if="host.upstreams.length > 2" class="text-[10px] text-purple-500 font-medium pl-3.5">
                                            +{{ host.upstreams.length - 2 }} more servers
                                        </p>
                                    </div>
                                </div>
                            </template>
                        </div>

                        <!-- Info Pills Row -->
                        <div class="flex flex-wrap gap-2">
                            <!-- SSL Pill -->
                            <div v-if="host.ssl_status === 'generating'" 
                                 class="inline-flex items-center gap-1.5 bg-yellow-50 border border-yellow-200 text-yellow-700 px-3 py-1.5 rounded-full text-xs font-semibold animate-pulse">
                                <svg class="animate-spin h-3.5 w-3.5" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24">
                                    <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle>
                                    <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4z"></path>
                                </svg>
                                SSL Generating
                            </div>
                            <div v-else-if="host.ssl_status === 'failed'" 
                                 class="inline-flex items-center gap-1.5 bg-red-50 border border-red-200 text-red-700 px-3 py-1.5 rounded-full text-xs font-semibold">
                                <svg xmlns="http://www.w3.org/2000/svg" class="h-3.5 w-3.5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 8v4m0 4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
                                </svg>
                                SSL Failed
                            </div>
                            <!-- Fallback SSL: ACME failed but using Self-Signed -->
                            <div v-else-if="host.ssl && host.ssl_provider === 'auto' && host.ssl_actual_provider === 'selfsigned'" 
                                 class="ssl-error-tooltip group/ssl relative inline-flex items-center gap-1.5 bg-orange-50 border border-orange-200 text-orange-700 px-3 py-1.5 rounded-full text-xs font-semibold cursor-pointer"
                                 @click.stop="sslErrorPinned = sslErrorPinned === host.id ? null : host.id">
                                <svg xmlns="http://www.w3.org/2000/svg" class="h-3.5 w-3.5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-3L13.732 4c-.77-1.333-2.694-1.333-3.464 0L3.34 16c-.77 1.333.192 3 1.732 3z" />
                                </svg>
                                Auto (Self Signed)
                                <div :class="['ssl-error-tooltip absolute bottom-full left-1/2 transform -translate-x-1/2 mb-2 w-64 max-w-xs bg-gray-800 text-white text-xs rounded p-2 transition-opacity z-50 text-center shadow-lg break-words select-text', sslErrorPinned === host.id ? 'opacity-100' : 'opacity-0 group-hover/ssl:opacity-100 pointer-events-none']">
                                    <div @click.stop="copyToClipboard(host.ssl_error || 'Public SSL failed')" class="cursor-pointer hover:bg-gray-700 rounded p-1 -m-1">
                                        ⚠️ {{ host.ssl_error || 'Public SSL (Let\'s Encrypt/ZeroSSL) failed. Using self-signed certificate.' }}
                                    </div>
                                    <div class="absolute top-full left-1/2 transform -translate-x-1/2 border-4 border-transparent border-t-gray-800"></div>
                                </div>
                            </div>
                            <div v-else-if="host.ssl" 
                                 class="inline-flex items-center gap-1.5 bg-green-50 border border-green-200 text-green-700 px-3 py-1.5 rounded-full text-xs font-semibold">
                                <svg xmlns="http://www.w3.org/2000/svg" class="h-3.5 w-3.5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 12l2 2 4-4m5.618-4.016A11.955 11.955 0 0112 2.944a11.955 11.955 0 01-8.618 3.04A12.02 12.02 0 003 9c0 5.591 3.824 10.29 9 11.622 5.176-1.332 9-6.03 9-11.622 0-1.042-.133-2.052-.382-3.016z" />
                                </svg>
                                {{ getSSLProviderName(host.ssl_provider, host.ssl_actual_provider) }}
                            </div>
                            <div v-else 
                                 class="inline-flex items-center gap-1.5 bg-gray-100 border border-gray-200 text-gray-600 px-3 py-1.5 rounded-full text-xs font-semibold">
                                <svg xmlns="http://www.w3.org/2000/svg" class="h-3.5 w-3.5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M8 11V7a4 4 0 118 0m-4 8v2m-6 4h12a2 2 0 002-2v-6a2 2 0 00-2-2H6a2 2 0 00-2 2v6a2 2 0 002 2z" />
                                </svg>
                                No SSL
                            </div>

                            <!-- Access Pill -->
                            <div v-if="host.access_list_id" 
                                 class="inline-flex items-center gap-1.5 bg-blue-50 border border-blue-200 text-blue-700 px-3 py-1.5 rounded-full text-xs font-semibold">
                                <svg xmlns="http://www.w3.org/2000/svg" class="h-3.5 w-3.5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 15v2m-6 4h12a2 2 0 002-2v-6a2 2 0 00-2-2H6a2 2 0 00-2 2v6a2 2 0 002 2zm10-10V7a4 4 0 00-8 0v4h8z" />
                                </svg>
                                {{ getAccessListName(host.access_list_id) }}
                            </div>
                            <div v-else 
                                 class="inline-flex items-center gap-1.5 bg-gray-100 border border-gray-200 text-gray-600 px-3 py-1.5 rounded-full text-xs font-semibold">
                                <svg xmlns="http://www.w3.org/2000/svg" class="h-3.5 w-3.5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M3.055 11H5a2 2 0 012 2v1a2 2 0 002 2 2 2 0 012 2v2.945M8 3.935V5.5A2.5 2.5 0 0010.5 8h.5a2 2 0 012 2 2 2 0 104 0 2 2 0 012-2h1.064M15 20.488V18a2 2 0 012-2h3.064M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
                                </svg>
                                Public
                            </div>
                        </div>

                        <!-- Action Buttons -->
                        <div class="flex items-center gap-2 pt-3 border-t border-gray-100">
                            <button @click="openEditModal(host)" 
                                    class="flex-1 flex items-center justify-center gap-2 bg-green-500 hover:bg-green-600 text-white font-semibold py-2.5 px-4 rounded-xl transition-all duration-200 shadow-md hover:shadow-lg active:scale-[0.98]">
                                <svg xmlns="http://www.w3.org/2000/svg" class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M11 5H6a2 2 0 00-2 2v11a2 2 0 002 2h11a2 2 0 002-2v-5m-1.414-9.414a2 2 0 112.828 2.828L11.828 15H9v-2.828l8.586-8.586z" />
                                </svg>
                                Edit
                            </button>
                            <button @click="deleteHost(host.id)" 
                                    class="flex items-center justify-center gap-2 bg-white hover:bg-red-50 text-red-600 font-semibold py-2.5 px-4 rounded-xl border border-gray-200 hover:border-red-200 transition-all duration-200 active:scale-[0.98]">
                                <svg xmlns="http://www.w3.org/2000/svg" class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16" />
                                </svg>
                            </button>
                        </div>
                    </div>
                </div>
            </div>
        </div>
    </div>

    <!-- Modal -->
    <div v-if="showModal" class="fixed inset-0 z-50 flex items-center justify-center bg-black/60 backdrop-blur-sm p-4">
      <div class="bg-white rounded-2xl shadow-2xl w-full max-w-lg transform transition-all overflow-hidden max-h-[90vh] flex flex-col">
        <!-- Modal Header -->
        <div class="bg-green-500 px-6 py-4 flex justify-between items-center flex-shrink-0">
            <div class="flex items-center gap-3">
                <div class="w-10 h-10 rounded-xl bg-white/20 flex items-center justify-center">
                    <svg xmlns="http://www.w3.org/2000/svg" class="h-5 w-5 text-white" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M21 12a9 9 0 01-9 9m9-9a9 9 0 00-9-9m9 9H3m9 9a9 9 0 01-9-9m9 9c1.657 0 3-4.03 3-9s-1.343-9-3-9m0 18c-1.657 0-3-4.03-3-9s1.343-9 3-9m-9 9a9 9 0 019-9" />
                    </svg>
                </div>
                <h3 class="text-lg font-bold text-white">{{ isEditing ? 'Edit Proxy Host' : 'Add Proxy Host' }}</h3>
            </div>
            <button @click="showModal = false" class="text-white/80 hover:text-white hover:bg-white/20 rounded-lg p-1.5 transition-all">
                <svg xmlns="http://www.w3.org/2000/svg" class="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
                </svg>
            </button>
        </div>
        
        <div class="p-6 space-y-4 overflow-y-auto flex-1">
            
            <!-- Tabs -->
            <div class="bg-gray-100 rounded-xl p-1 flex gap-1">
                <button @click="activeTab = 'details'" :class="activeTab === 'details' ? 'bg-white shadow-sm text-green-600' : 'text-gray-600 hover:text-gray-800'" class="flex-1 py-2 px-3 rounded-lg font-medium text-sm transition-all">
                    Details
                </button>
                <button @click="activeTab = 'ssl'" :class="activeTab === 'ssl' ? 'bg-white shadow-sm text-green-600' : 'text-gray-600 hover:text-gray-800'" class="flex-1 py-2 px-3 rounded-lg font-medium text-sm transition-all">
                    SSL
                </button>
                <button v-if="newHost.type === 'loadbalancer'" @click="activeTab = 'loadbalancer'" :class="activeTab === 'loadbalancer' ? 'bg-white shadow-sm text-green-600' : 'text-gray-600 hover:text-gray-800'" class="flex-1 py-2 px-3 rounded-lg font-medium text-sm transition-all">
                    Upstreams
                </button>
                <button @click="activeTab = 'locations'" :class="activeTab === 'locations' ? 'bg-white shadow-sm text-green-600' : 'text-gray-600 hover:text-gray-800'" class="flex-1 py-2 px-3 rounded-lg font-medium text-sm transition-all">
                    Locations
                </button>
                <button @click="activeTab = 'advanced'" :class="activeTab === 'advanced' ? 'bg-white shadow-sm text-green-600' : 'text-gray-600 hover:text-gray-800'" class="flex-1 py-2 px-3 rounded-lg font-medium text-sm transition-all">
                    Advanced
                </button>
            </div>

            <!-- Details Tab -->
            <div v-if="activeTab === 'details'" class="space-y-4">
                <div>
                    <label class="block text-sm font-medium text-gray-700 mb-1">Domain Names <span class="text-red-500">*</span></label>
                    <Combobox 
                        v-model="newHost.domain" 
                        :options="hostOptions" 
                        placeholder="example.com" 
                        color="green"
                        @blur="newHost.domain = cleanDomain(newHost.domain)"
                    />
                    <p class="text-xs text-gray-500 mt-1">e.g., example.com, sub.example.com, *.example.com</p>
                </div>

                <div class="grid grid-cols-2 gap-4">
                    <div>
                        <label class="block text-sm font-medium text-gray-700 mb-1">Scheme</label>
                        <CustomSelect 
                            :modelValue="'http'" 
                            :options="schemeOptions" 
                            :disabled="true"
                            color="green"
                        />
                    </div>
                    <div>
                        <label class="block text-sm font-medium text-gray-700 mb-1">Type</label>
                        <CustomSelect 
                            v-model="newHost.type" 
                            :options="typeOptions" 
                            color="green"
                        />
                    </div>
                </div>

                <div v-if="newHost.type === 'proxy'">
                    <label class="block text-sm font-medium text-gray-700 mb-1">Forward Hostname / IP <span class="text-red-500">*</span></label>
                    <input v-model="newHost.target" @blur="newHost.target = cleanTarget(newHost.target)" class="w-full border border-gray-300 rounded-lg px-4 py-2.5 text-gray-700 bg-white focus:outline-none focus:ring-2 focus:ring-green-500/50 focus:border-green-500 transition-all duration-200 shadow-sm hover:border-gray-400" placeholder="127.0.0.1:8080">
                    <p class="text-xs text-gray-500 mt-1">Format: hostname:port or ip:port (e.g., example.com:8080, 192.168.1.1:80)</p>
                </div>

                <div v-if="newHost.type === 'redirect'">
                    <label class="block text-sm font-medium text-gray-700 mb-1">Redirect To <span class="text-red-500">*</span></label>
                    <input v-model="newHost.target" class="w-full border border-gray-300 rounded-lg px-4 py-2.5 text-gray-700 bg-white focus:outline-none focus:ring-2 focus:ring-green-500/50 focus:border-green-500 transition-all duration-200 shadow-sm hover:border-gray-400" placeholder="https://example.com">
                </div>

                <div v-if="newHost.type === 'redirect'">
                    <label class="block text-sm font-medium text-gray-700 mb-1">Forwarding Code</label>
                    <CustomSelect 
                        v-model="newHost.forwarding_code" 
                        :options="forwardingCodeOptions" 
                        color="green"
                    />
                </div>

                <div v-if="newHost.type !== 'redirect'">
                    <label class="block text-sm font-medium text-gray-700 mb-1">Access List</label>
                    <CustomSelect 
                        v-model="newHost.access_list_id" 
                        :options="accessListOptions" 
                        color="green"
                    />
                </div>
            </div>

            <!-- SSL Tab -->
            <div v-if="activeTab === 'ssl'" class="space-y-4">
                <div>
                    <label class="block text-sm font-medium text-gray-700 mb-1">SSL Provider</label>
                    <CustomSelect 
                        v-model="newHost.ssl_provider" 
                        :options="sslProviderOptions" 
                        color="green"
                    />
                </div>

                <div v-if="newHost.ssl_provider !== 'none'" class="space-y-2">
                    <!-- Custom Certificate Upload -->
                    <div v-if="newHost.ssl_provider === 'custom'" class="pl-4 border-l-2 border-green-100 space-y-3 mb-3">
                        <div class="flex gap-4 mb-2">
                            <label class="flex items-center cursor-pointer">
                                <input type="radio" name="cert_mode" v-model="useExistingCert" :value="false" class="mr-2 text-green-600 focus:ring-green-500">
                                <span class="text-sm text-gray-700">Upload New</span>
                            </label>
                            <label class="flex items-center cursor-pointer">
                                <input type="radio" name="cert_mode" v-model="useExistingCert" :value="true" class="mr-2 text-green-600 focus:ring-green-500">
                                <span class="text-sm text-gray-700">Use Existing</span>
                            </label>
                        </div>

                        <div v-if="!useExistingCert">
                            <div class="mb-2">
                                <label class="block text-xs font-medium text-gray-600 mb-1">Certificate File (.pem/.crt)</label>
                                <input type="file" @change="handleCertUpload" accept=".pem,.crt,.cer" class="block w-full text-sm text-gray-500 file:mr-4 file:py-2 file:px-4 file:rounded-full file:border-0 file:text-xs file:font-semibold file:bg-green-50 file:text-green-700 hover:file:bg-green-100 transition-colors">
                            </div>
                            <div>
                                <label class="block text-xs font-medium text-gray-600 mb-1">Private Key File (.key)</label>
                                <input type="file" @change="handleKeyUpload" accept=".key" class="block w-full text-sm text-gray-500 file:mr-4 file:py-2 file:px-4 file:rounded-full file:border-0 file:text-xs file:font-semibold file:bg-green-50 file:text-green-700 hover:file:bg-green-100 transition-colors">
                            </div>
                        </div>

                        <div v-else>
                            <label class="block text-xs font-medium text-gray-600 mb-1">Select Certificate</label>
                            <CustomSelect 
                                v-model="newHost.certificate_id" 
                                :options="certificateOptions" 
                                color="green"
                            />
                        </div>
                    </div>

                    <div v-if="['auto', 'letsencrypt', 'zerossl'].includes(newHost.ssl_provider)">
                        <label class="flex items-center cursor-pointer group">
                            <input type="checkbox" v-model="newHost.use_dns_challenge" class="mr-3 w-4 h-4 rounded text-green-600 focus:ring-green-500 border-gray-300 transition-colors">
                            <span class="text-sm text-gray-700 group-hover:text-gray-900 transition-colors">Use DNS Challenge</span>
                        </label>

                        <div v-if="newHost.use_dns_challenge" class="pl-7 space-y-3 border-l-2 border-green-100 mt-2">
                            <div>
                                <label class="block text-xs font-medium text-gray-600 mb-1">DNS Provider</label>
                                <CustomSelect 
                                    v-model="newHost.dns_provider" 
                                    :options="dnsProviderOptions" 
                                    color="green"
                                />
                            </div>
                            <div>
                                <label class="block text-xs font-medium text-gray-600 mb-1">API Token / Credentials</label>
                                <input v-model="newHost.dns_token" type="password" class="w-full border border-gray-300 rounded-lg px-4 py-2 text-sm text-gray-700 bg-white focus:outline-none focus:ring-2 focus:ring-green-500/50 focus:border-green-500 transition-all duration-200 shadow-sm hover:border-gray-400" placeholder="Enter API Token">
                            </div>
                        </div>
                    </div>

                    <label class="flex items-center cursor-pointer group">
                        <input type="checkbox" v-model="newHost.hsts_enabled" class="mr-3 w-4 h-4 rounded text-green-600 focus:ring-green-500 border-gray-300 transition-colors">
                        <span class="text-sm text-gray-700 group-hover:text-gray-900 transition-colors">Force SSL (HSTS)</span>
                    </label>
                </div>
            </div>

            <!-- Load Balancer Tab -->
            <div v-if="activeTab === 'loadbalancer'" class="space-y-4">
                <div class="bg-blue-50 border border-blue-200 rounded-lg p-3 text-sm text-blue-800">
                    <div class="flex items-start gap-2">
                        <svg xmlns="http://www.w3.org/2000/svg" class="h-5 w-5 flex-shrink-0 mt-0.5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13 16h-1v-4h-1m1-4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
                        </svg>
                        <span>Configure multiple backend servers for load balancing. When upstreams are defined, they will be used instead of the single target.</span>
                    </div>
                </div>

                <!-- Load Balancing Policy -->
                <div>
                    <label class="block text-sm font-medium text-gray-700 mb-1">Load Balancing Policy</label>
                    <CustomSelect
                        v-model="newHost.load_balancing"
                        :options="loadBalancingOptions"
                        color="green"
                    />
                </div>

                <!-- Upstreams List -->
                <div class="space-y-2">
                    <div class="flex justify-between items-center">
                        <label class="block text-sm font-medium text-gray-700">Upstream Servers</label>
                        <span class="text-xs text-gray-500">{{ newHost.upstreams.length }} server(s)</span>
                    </div>
                    
                    <div v-if="newHost.upstreams.length === 0" class="text-center text-gray-500 py-6 text-sm border-2 border-dashed border-gray-300 rounded-lg">
                        <svg xmlns="http://www.w3.org/2000/svg" class="h-8 w-8 mx-auto text-gray-400 mb-2" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 12h14M5 12a2 2 0 01-2-2V6a2 2 0 012-2h14a2 2 0 012 2v4a2 2 0 01-2 2M5 12a2 2 0 00-2 2v4a2 2 0 002 2h14a2 2 0 002-2v-4a2 2 0 00-2-2m-2-4h.01M17 16h.01" />
                        </svg>
                        <p>No upstream servers defined</p>
                        <p class="text-xs text-gray-400 mt-1">Add servers below for load balancing</p>
                    </div>
                    
                    <div v-for="(upstream, index) in newHost.upstreams" :key="index" class="bg-gray-50 p-3 rounded-lg border border-gray-200 space-y-2">
                        <div class="flex justify-between items-center">
                            <span class="text-xs font-medium text-gray-500">Server #{{ index + 1 }}</span>
                            <button @click="removeUpstream(index)" class="text-red-500 hover:text-red-700 p-1">
                                <svg xmlns="http://www.w3.org/2000/svg" class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16" />
                                </svg>
                            </button>
                        </div>
                        <div class="grid grid-cols-2 gap-2">
                            <div class="col-span-2">
                                <label class="block text-xs font-medium text-gray-500 mb-1">Target Address</label>
                                <input v-model="upstream.target" @blur="upstream.target = cleanTarget(upstream.target)" class="w-full border border-gray-300 rounded px-3 py-1.5 text-sm focus:outline-none focus:border-green-500" placeholder="192.168.1.100:8080">
                                <p class="text-xs text-gray-400 mt-0.5">hostname:port or ip:port</p>
                            </div>
                            <div>
                                <label class="block text-xs font-medium text-gray-500 mb-1">Weight</label>
                                <input v-model.number="upstream.weight" type="number" min="1" class="w-full border border-gray-300 rounded px-3 py-1.5 text-sm focus:outline-none focus:border-green-500" placeholder="1">
                            </div>
                            <div>
                                <label class="block text-xs font-medium text-gray-500 mb-1">Max Fails</label>
                                <input v-model.number="upstream.max_fails" type="number" min="0" class="w-full border border-gray-300 rounded px-3 py-1.5 text-sm focus:outline-none focus:border-green-500" placeholder="0">
                            </div>
                        </div>
                    </div>
                    
                    <button @click="addUpstream" class="w-full text-sm text-green-600 hover:text-green-700 font-medium flex items-center justify-center gap-1 py-2 border-2 border-dashed border-green-300 rounded-lg hover:border-green-400 hover:bg-green-50 transition-colors">
                        <svg xmlns="http://www.w3.org/2000/svg" class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4v16m8-8H4" />
                        </svg>
                        Add Upstream Server
                    </button>
                </div>

                <!-- Health Check Settings -->
                <div class="border border-gray-200 rounded-lg overflow-hidden">
                    <div class="bg-gray-50 px-4 py-3 flex items-center justify-between">
                        <div class="flex items-center gap-2">
                            <svg xmlns="http://www.w3.org/2000/svg" class="h-5 w-5 text-green-600" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z" />
                            </svg>
                            <span class="font-medium text-gray-700">Health Checks</span>
                        </div>
                        <button @click="newHost.health_check = !newHost.health_check" :class="newHost.health_check ? 'bg-green-500' : 'bg-gray-300'" class="relative inline-flex h-5 w-9 items-center rounded-full transition-colors focus:outline-none focus:ring-2 focus:ring-green-500 focus:ring-offset-2">
                            <span :class="newHost.health_check ? 'translate-x-4' : 'translate-x-1'" class="inline-block h-3.5 w-3.5 transform rounded-full bg-white transition-transform" />
                        </button>
                    </div>
                    <div v-if="newHost.health_check" class="p-4 space-y-3 border-t border-gray-200">
                        <div>
                            <label class="block text-xs font-medium text-gray-600 mb-1">Health Check Path</label>
                            <input v-model="newHost.health_check_path" class="w-full border border-gray-300 rounded px-3 py-1.5 text-sm focus:outline-none focus:border-green-500" placeholder="/health">
                        </div>
                        <div>
                            <label class="block text-xs font-medium text-gray-600 mb-1">Check Interval</label>
                            <input v-model="newHost.health_check_interval" class="w-full border border-gray-300 rounded px-3 py-1.5 text-sm focus:outline-none focus:border-green-500" placeholder="10s">
                        </div>
                    </div>
                </div>

                <!-- Retry Settings -->
                <div class="grid grid-cols-2 gap-4">
                    <div>
                        <label class="block text-xs font-medium text-gray-600 mb-1">Try Duration</label>
                        <input v-model="newHost.lb_try_duration" class="w-full border border-gray-300 rounded px-3 py-1.5 text-sm focus:outline-none focus:border-green-500" placeholder="5s">
                        <p class="text-xs text-gray-400 mt-1">How long to try backends</p>
                    </div>
                    <div>
                        <label class="block text-xs font-medium text-gray-600 mb-1">Try Interval</label>
                        <input v-model="newHost.lb_try_interval" class="w-full border border-gray-300 rounded px-3 py-1.5 text-sm focus:outline-none focus:border-green-500" placeholder="250ms">
                        <p class="text-xs text-gray-400 mt-1">Interval between retries</p>
                    </div>
                </div>
            </div>

            <!-- Locations Tab -->
            <div v-if="activeTab === 'locations'" class="space-y-4">
                <div v-if="newHost.locations.length === 0" class="text-center text-gray-500 py-4 text-sm">
                    No custom locations defined.
                </div>
                <div v-for="(loc, index) in newHost.locations" :key="index" class="bg-gray-50 p-3 rounded-lg border border-gray-200 space-y-3">
                    <!-- Location Header -->
                    <div class="flex items-center justify-between">
                        <span class="text-sm font-medium text-gray-700">Location {{ index + 1 }}</span>
                        <button @click="removeLocation(index)" class="text-red-500 hover:text-red-700 p-1">
                            <svg xmlns="http://www.w3.org/2000/svg" class="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16" />
                            </svg>
                        </button>
                    </div>
                    
                    <!-- Path -->
                    <div>
                        <label class="block text-xs font-medium text-gray-500 mb-1">Path</label>
                        <input v-model="loc.path" class="w-full border border-gray-300 rounded px-2 py-1.5 text-sm focus:outline-none focus:border-green-500" placeholder="/api">
                    </div>
                    
                    <!-- Mode Toggle: Single Target vs Load Balancer -->
                    <div class="flex items-center gap-4">
                        <label class="flex items-center cursor-pointer">
                            <input type="radio" :name="'loc-mode-' + index" :checked="!loc.upstreams || loc.upstreams.length === 0" @change="setLocationMode(index, 'single')" class="mr-2">
                            <span class="text-sm text-gray-700">Single Target</span>
                        </label>
                        <label class="flex items-center cursor-pointer">
                            <input type="radio" :name="'loc-mode-' + index" :checked="loc.upstreams && loc.upstreams.length > 0" @change="setLocationMode(index, 'loadbalancer')" class="mr-2">
                            <span class="text-sm text-gray-700">Load Balancer</span>
                        </label>
                    </div>
                    
                    <!-- Single Target -->
                    <div v-if="!loc.upstreams || loc.upstreams.length === 0">
                        <label class="block text-xs font-medium text-gray-500 mb-1">Target</label>
                        <input v-model="loc.target" @blur="loc.target = cleanTarget(loc.target)" class="w-full border border-gray-300 rounded px-2 py-1.5 text-sm focus:outline-none focus:border-green-500" placeholder="localhost:3000">
                        <p class="text-xs text-gray-400 mt-0.5">Format: hostname:port or ip:port</p>
                    </div>
                    
                    <!-- Load Balancer Settings -->
                    <div v-else class="space-y-3 border-t border-gray-200 pt-3">
                        <div>
                            <label class="block text-xs font-medium text-gray-500 mb-1">LB Policy</label>
                            <CustomSelect
                                v-model="loc.load_balancing"
                                :options="[
                                    { label: 'Round Robin', value: 'round_robin' },
                                    { label: 'Least Connections', value: 'least_conn' },
                                    { label: 'Random', value: 'random' },
                                    { label: 'First Available', value: 'first' },
                                    { label: 'IP Hash', value: 'ip_hash' },
                                    { label: 'URI Hash', value: 'uri_hash' },
                                    { label: 'Cookie', value: 'cookie' }
                                ]"
                                placeholder="Select policy..."
                            />
                        </div>
                        
                        <!-- Upstream Servers -->
                        <div class="space-y-2">
                            <label class="block text-xs font-medium text-gray-500">Upstream Servers</label>
                            <div v-for="(up, upIdx) in loc.upstreams" :key="upIdx" class="flex gap-2 items-center bg-white p-2 rounded border border-gray-200">
                                <input v-model="up.target" @blur="up.target = cleanTarget(up.target)" class="flex-1 border border-gray-300 rounded px-2 py-1 text-sm focus:outline-none focus:border-green-500" placeholder="10.0.0.1:8080">
                                <input v-model.number="up.weight" type="number" min="1" class="w-16 border border-gray-300 rounded px-2 py-1 text-sm focus:outline-none focus:border-green-500" placeholder="Weight" title="Weight">
                                <button @click="removeLocationUpstream(index, upIdx)" class="text-red-500 hover:text-red-700 p-1">
                                    <svg xmlns="http://www.w3.org/2000/svg" class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                                        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
                                    </svg>
                                </button>
                            </div>
                            <button @click="addLocationUpstream(index)" class="text-xs text-green-600 hover:text-green-700 font-medium flex items-center gap-1">
                                <svg xmlns="http://www.w3.org/2000/svg" class="h-3 w-3" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4v16m8-8H4" />
                                </svg>
                                Add Upstream
                            </button>
                        </div>
                        
                        <!-- Health Check Toggle -->
                        <label class="flex items-center cursor-pointer">
                            <input type="checkbox" v-model="loc.health_check" class="mr-2 w-4 h-4 rounded text-green-600 focus:ring-green-500 border-gray-300">
                            <span class="text-xs text-gray-700">Enable Health Check</span>
                        </label>
                        
                        <!-- Health Check Settings -->
                        <div v-if="loc.health_check" class="grid grid-cols-2 gap-2">
                            <div>
                                <label class="block text-xs font-medium text-gray-500 mb-1">Path</label>
                                <input v-model="loc.health_check_path" class="w-full border border-gray-300 rounded px-2 py-1 text-sm focus:outline-none focus:border-green-500" placeholder="/health">
                            </div>
                            <div>
                                <label class="block text-xs font-medium text-gray-500 mb-1">Interval</label>
                                <input v-model="loc.health_check_interval" class="w-full border border-gray-300 rounded px-2 py-1 text-sm focus:outline-none focus:border-green-500" placeholder="10s">
                            </div>
                        </div>
                    </div>
                </div>
                <button @click="addLocation" class="text-sm text-green-600 hover:text-green-700 font-medium flex items-center gap-1">
                    <svg xmlns="http://www.w3.org/2000/svg" class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4v16m8-8H4" />
                    </svg>
                    Add Location
                </button>
            </div>

            <!-- Advanced Tab -->
            <div v-if="activeTab === 'advanced'" class="space-y-4">
                <div class="bg-white p-4 rounded-xl border border-gray-200 space-y-3">
                    <label class="flex items-center cursor-pointer group p-2 rounded-lg hover:bg-gray-100 transition-colors">
                        <input type="checkbox" v-model="newHost.block_exploits" class="mr-3 w-4 h-4 rounded text-green-600 focus:ring-green-500 border-gray-300 transition-colors">
                        <span class="text-sm text-gray-700 group-hover:text-gray-900 transition-colors">Block Common Exploits</span>
                    </label>
                    <label class="flex items-center cursor-pointer group p-2 rounded-lg hover:bg-gray-100 transition-colors">
                        <input type="checkbox" v-model="newHost.cache_assets" class="mr-3 w-4 h-4 rounded text-green-600 focus:ring-green-500 border-gray-300 transition-colors">
                        <span class="text-sm text-gray-700 group-hover:text-gray-900 transition-colors">Cache Assets</span>
                    </label>
                    <label class="flex items-center cursor-pointer group p-2 rounded-lg hover:bg-gray-100 transition-colors">
                        <input type="checkbox" v-model="newHost.websockets" class="mr-3 w-4 h-4 rounded text-green-600 focus:ring-green-500 border-gray-300 transition-colors">
                        <span class="text-sm text-gray-700 group-hover:text-gray-900 transition-colors">Websockets Support</span>
                    </label>
                </div>
            </div>

        </div>

        <div class="bg-gray-50 px-6 py-4 border-t border-gray-200 flex justify-end gap-3 flex-shrink-0">
          <button @click="showModal = false" class="px-5 py-2.5 border border-gray-200 rounded-xl text-gray-700 hover:bg-gray-100 transition-all font-medium">Cancel</button>
          <button @click="saveHost" class="px-6 py-2.5 bg-green-500 hover:bg-green-600 text-white rounded-xl shadow-lg shadow-green-500/25 transition-all font-semibold">Save Host</button>
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
.divide-y > div:nth-child(9) { animation-delay: 240ms; }
.divide-y > div:nth-child(10) { animation-delay: 270ms; }

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

/* Hover effects */
.hover\:shadow-lg {
    transition: all 0.2s cubic-bezier(0.22, 1, 0.36, 1);
}
</style>
