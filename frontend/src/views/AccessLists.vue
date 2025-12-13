<script setup lang="ts">
import { ref, onMounted, onUnmounted, computed } from 'vue'
import CustomSelect from '../components/CustomSelect.vue'
import { toast } from 'vue-sonner'
import { useConfirm } from '../composables/useConfirm'
import { useWebSocket } from '../composables/useWebSocket'
import { authFetch } from '../composables/useAuth'

const { confirm } = useConfirm()
const { on, off } = useWebSocket()

interface AccessListClient {
    username: string
    password?: string
}

interface AccessListRule {
    ip: string
    action: string
}

interface AccessList {
    id: number
    name: string
    clients: AccessListClient[]
    rules: AccessListRule[]
    created_at: string
}

const lists = ref<AccessList[]>([])
const showModal = ref(false)
const isEditing = ref(false)
const editingId = ref<number | null>(null)
const activeTab = ref<'users' | 'rules'>('users')
const searchQuery = ref('')

const filteredLists = computed(() => {
    if (!searchQuery.value.trim()) return lists.value
    const q = searchQuery.value.toLowerCase()
    return lists.value.filter(l => 
        l.name?.toLowerCase().includes(q) ||
        l.clients?.some(c => c.username?.toLowerCase().includes(q)) ||
        l.rules?.some(r => r.ip?.toLowerCase().includes(q))
    )
})

const newList = ref({
    name: '',
    clients: [{ username: '', password: '' }],
    rules: [{ ip: '', action: 'allow' }]
})

const actionOptions = [
    { label: 'Allow', value: 'allow' },
    { label: 'Deny', value: 'deny' }
]

const fetchLists = async () => {
    const res = await authFetch('/api/access-lists')
    if (res.ok) {
        lists.value = await res.json()
    }
}

const openAddModal = () => {
    isEditing.value = false
    editingId.value = null
    newList.value = {
        name: '',
        clients: [{ username: '', password: '' }],
        rules: [{ ip: '', action: 'allow' }]
    }
    activeTab.value = 'users'
    showModal.value = true
}

const openEditModal = (list: AccessList) => {
    isEditing.value = true
    editingId.value = list.id
    newList.value = {
        name: list.name,
        clients: list.clients.map(c => ({ username: c.username, password: '' })),
        rules: list.rules.length > 0 ? list.rules.map(r => ({ ip: r.ip, action: r.action })) : [{ ip: '', action: 'allow' }]
    }
    if (newList.value.clients.length === 0) newList.value.clients.push({ username: '', password: '' })
    
    activeTab.value = 'users'
    showModal.value = true
}

const addClientRow = () => {
    newList.value.clients.push({ username: '', password: '' })
}

const removeClientRow = (index: number) => {
    newList.value.clients.splice(index, 1)
}

const addRuleRow = () => {
    newList.value.rules.push({ ip: '', action: 'allow' })
}

const removeRuleRow = (index: number) => {
    newList.value.rules.splice(index, 1)
}

const saveList = async () => {
    if (!newList.value.name.trim()) {
        toast.error('Please enter a name for the access list')
        return
    }
    
    for (const client of newList.value.clients) {
        const hasUsername = client.username.trim()
        const hasPassword = client.password
        
        if (hasUsername && !hasPassword && !isEditing.value) {
            toast.error('Password is required for all users')
            return
        }
        
        if (!hasUsername && hasPassword) {
            toast.error('Username is required for all users with password')
            return
        }
    }
    
    const usernames = newList.value.clients
        .map(c => c.username.trim().toLowerCase())
        .filter(u => u !== '')
    const uniqueUsernames = new Set(usernames)
    if (usernames.length !== uniqueUsernames.size) {
        toast.error('Duplicate usernames are not allowed')
        return
    }
    
    const ipRegex = /^(\d{1,3}\.\d{1,3}\.\d{1,3}\.\d{1,3})(\/\d{1,2})?$/
    for (const rule of newList.value.rules) {
        if (rule.ip.trim() && !ipRegex.test(rule.ip.trim())) {
            toast.error(`Invalid IP format: ${rule.ip}. Use format: 192.168.1.1 or 192.168.1.0/24`)
            return
        }
    }
    
    const validClients = newList.value.clients.filter(c => {
        if (!c.username.trim()) return false
        if (!isEditing.value && !c.password) return false
        return true
    })
    
    const validRules = newList.value.rules.filter(r => r.ip.trim())

    const payload = {
        name: newList.value.name.trim(),
        clients: validClients,
        rules: validRules
    }

    const url = isEditing.value ? `/api/access-lists/${editingId.value}` : '/api/access-lists'
    const method = isEditing.value ? 'PUT' : 'POST'

    const res = await authFetch(url, {
        method: method,
        headers: { 
            'Content-Type': 'application/json'
        },
        body: JSON.stringify(payload)
    })

    if (res.ok) {
        showModal.value = false
        fetchLists()
        toast.success(isEditing.value ? 'Access list updated successfully' : 'Access list created successfully')
    } else {
        toast.error('Failed to save access list')
    }
}

const deleteList = async (id: number) => {
    const confirmed = await confirm(
        'Delete Access List', 
        'Are you sure you want to delete this access list? This action cannot be undone.',
        { type: 'danger', confirmText: 'Delete' }
    )
    if (!confirmed) return

    const res = await authFetch(`/api/access-lists/${id}`, {
        method: 'DELETE'
    })
    
    if (res.ok) {
        toast.success('Access list deleted successfully')
    } else {
        toast.error('Failed to delete access list')
    }
}

const handleAccessListCreated = (newAccessList: AccessList) => {
    if (!lists.value.find(l => l.id === newAccessList.id)) {
        lists.value.unshift(newAccessList)
    }
}

const handleAccessListUpdated = (updatedList: AccessList) => {
    const index = lists.value.findIndex(l => l.id === updatedList.id)
    if (index !== -1) {
        lists.value[index] = updatedList
    }
}

const handleAccessListDeleted = (payload: { id: number }) => {
    lists.value = lists.value.filter(l => l.id !== Number(payload.id))
}

onMounted(() => {
    fetchLists()
    
    on('access_list_created', handleAccessListCreated)
    on('access_list_updated', handleAccessListUpdated)
    on('access_list_deleted', handleAccessListDeleted)
})

onUnmounted(() => {
    off('access_list_created', handleAccessListCreated)
    off('access_list_updated', handleAccessListUpdated)
    off('access_list_deleted', handleAccessListDeleted)
})
</script>

<template>
  <div>

    <!-- Header -->
    <div class="bg-white dark:bg-gray-800 rounded-t-xl shadow-sm border border-gray-200 dark:border-gray-700 overflow-hidden">
        <div class="h-1 bg-red-500"></div>
        <div class="p-4 flex flex-col lg:flex-row justify-between items-start lg:items-center gap-3">
            <div>
                <h2 class="text-xl font-bold text-gray-800 dark:text-gray-100 tracking-wide">Access Lists</h2>
                <p class="text-sm text-gray-500 dark:text-gray-400 mt-0.5">Manage authentication and IP-based access control</p>
            </div>
        <div class="flex items-center gap-3 w-full lg:w-auto">
            <div class="relative flex-1 lg:flex-initial">
                <input 
                    v-model="searchQuery" 
                    type="text" 
                    placeholder="Search access lists..." 
                    class="w-full lg:w-64 pl-9 pr-4 py-2 border border-gray-300 dark:border-gray-600 rounded-lg text-sm focus:ring-2 focus:ring-red-500 focus:border-red-500 transition-all bg-white dark:bg-gray-800 dark:text-gray-100"
                />
                <svg xmlns="http://www.w3.org/2000/svg" class="h-4 w-4 text-gray-400 dark:text-gray-500 absolute left-3 top-1/2 -translate-y-1/2" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z" />
                </svg>
            </div>
            <button @click="openAddModal" class="bg-red-500 hover:bg-red-600 text-white px-4 py-2 rounded-lg shadow-md hover:shadow-lg transition-all duration-200 font-medium text-sm flex items-center gap-2 whitespace-nowrap">
                <svg xmlns="http://www.w3.org/2000/svg" class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4v16m8-8H4" />
                </svg>
                <span>Add Access List</span>
            </button>
        </div>
        </div>
    </div>

    <!-- Column Headers - Desktop -->
    <div class="bg-gray-50/80 dark:bg-gray-700/80 px-5 py-3 border-b text-xs font-bold text-gray-500 dark:text-gray-400 uppercase tracking-wider border-x border-gray-200 dark:border-gray-700 hidden lg:grid lg:grid-cols-12 gap-4 items-center">
        <div class="col-span-4 pl-14">Name</div>
        <div class="col-span-3">Users</div>
        <div class="col-span-3">IP Rules</div>
        <div class="col-span-2 text-right pr-4">Actions</div>
    </div>

    <!-- List -->
    <div class="bg-white dark:bg-gray-800 rounded-b-lg shadow-md overflow-hidden border-x border-b border-gray-200 dark:border-gray-700 dark:border-gray-700">
        <!-- Empty State -->
        <div v-if="lists.length === 0" class="p-16 text-center">
            <div class="mb-6">
                <div class="mx-auto w-20 h-20 rounded-2xl bg-red-100 flex items-center justify-center shadow-lg">
                    <svg xmlns="http://www.w3.org/2000/svg" class="h-10 w-10 text-red-500" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M12 15v2m-6 4h12a2 2 0 002-2v-6a2 2 0 00-2-2H6a2 2 0 00-2 2v6a2 2 0 002 2zm10-10V7a4 4 0 00-8 0v4h8z" />
                    </svg>
                </div>
            </div>
            <p class="text-xl font-bold text-gray-800 dark:text-gray-100 mb-2">No access lists created</p>
            <p class="text-sm text-gray-500 dark:text-gray-400 mb-6 max-w-md mx-auto">Create access lists to protect your proxy hosts with authentication and IP-based access control</p>
            <button @click="openAddModal" class="inline-flex items-center gap-2 bg-red-500 hover:bg-red-600 text-white px-6 py-3 rounded-xl shadow-lg hover:shadow-xl transition-all text-sm font-semibold">
                <svg xmlns="http://www.w3.org/2000/svg" class="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4v16m8-8H4" />
                </svg>
                Create Your First Access List
            </button>
        </div>

        <!-- No Results State -->
        <div v-else-if="filteredLists.length === 0" class="p-16 text-center">
            <div class="mb-6">
                <div class="mx-auto w-20 h-20 rounded-2xl bg-gray-100 flex items-center justify-center">
                    <svg xmlns="http://www.w3.org/2000/svg" class="h-10 w-10 text-gray-400 dark:text-gray-500" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z" />
                    </svg>
                </div>
            </div>
            <p class="text-xl font-bold text-gray-800 dark:text-gray-100 mb-2">No results found</p>
            <p class="text-sm text-gray-500 dark:text-gray-400 dark:text-gray-400">No access lists match "<span class="font-medium">{{ searchQuery }}</span>"</p>
        </div>

        <!-- List Items -->
        <div v-for="list in filteredLists" :key="list.id" class="group p-4 border-b border-gray-200 dark:border-gray-700 hover:bg-gray-50 dark:hover:bg-gray-700 transition-all duration-200 last:border-b-0">
            <!-- Desktop Layout -->
            <div class="hidden lg:grid lg:grid-cols-12 gap-4 items-center">
                <!-- Icon + Name -->
                <div class="col-span-4 flex items-center gap-3 overflow-hidden">
                    <div class="flex-shrink-0 w-10 h-10 rounded-xl bg-red-100 flex items-center justify-center group-hover:shadow-md transition-all">
                        <svg xmlns="http://www.w3.org/2000/svg" class="h-5 w-5 text-red-600" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 15v2m-6 4h12a2 2 0 002-2v-6a2 2 0 00-2-2H6a2 2 0 00-2 2v6a2 2 0 002 2zm10-10V7a4 4 0 00-8 0v4h8z" />
                        </svg>
                    </div>
                    <div class="min-w-0">
                        <p class="text-sm font-bold text-gray-900 dark:text-gray-100 truncate">{{ list.name }}</p>
                        <p class="text-xs text-gray-400 dark:text-gray-500">{{ list.clients.length }} user{{ list.clients.length !== 1 ? 's' : '' }} • {{ list.rules?.length || 0 }} rule{{ (list.rules?.length || 0) !== 1 ? 's' : '' }}</p>
                    </div>
                </div>

                <!-- Users Pills -->
                <div class="col-span-3">
                    <div class="flex flex-wrap gap-1.5">
                        <span v-for="(client, idx) in list.clients.slice(0, 3)" :key="'c'+idx" 
                              class="inline-flex items-center gap-1 px-2.5 py-1 rounded-full text-xs font-semibold bg-gray-50 dark:bg-gray-700 border border-gray-200 dark:border-gray-700 text-gray-700 dark:text-gray-200 dark:text-gray-300">
                            <svg xmlns="http://www.w3.org/2000/svg" class="h-3 w-3 text-gray-400 dark:text-gray-500" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M16 7a4 4 0 11-8 0 4 4 0 018 0zM12 14a7 7 0 00-7 7h14a7 7 0 00-7-7z" />
                            </svg>
                            {{ client.username }}
                        </span>
                        <span v-if="list.clients.length > 3" class="text-xs text-gray-400 font-medium py-1">+{{ list.clients.length - 3 }} more</span>
                        <span v-if="list.clients.length === 0" class="text-xs text-gray-400 italic">No users</span>
                    </div>
                </div>

                <!-- Rules Pills -->
                <div class="col-span-3">
                    <div class="flex flex-wrap gap-1.5">
                        <span v-for="(rule, idx) in (list.rules || []).slice(0, 2)" :key="'r'+idx" 
                              :class="rule.action === 'allow' 
                                ? 'bg-green-50 border-green-200 text-green-700' 
                                : 'bg-red-50 border-red-200 text-red-700'" 
                              class="inline-flex items-center gap-1 px-2.5 py-1 rounded-full text-xs font-semibold border">
                            <svg v-if="rule.action === 'allow'" xmlns="http://www.w3.org/2000/svg" class="h-3 w-3" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 13l4 4L19 7" />
                            </svg>
                            <svg v-else xmlns="http://www.w3.org/2000/svg" class="h-3 w-3" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
                            </svg>
                            {{ rule.ip }}
                        </span>
                        <span v-if="(list.rules || []).length > 2" class="text-xs text-gray-400 font-medium py-1">+{{ list.rules.length - 2 }} more</span>
                        <span v-if="!list.rules || list.rules.length === 0" class="text-xs text-gray-400 italic">No rules</span>
                    </div>
                </div>

                <!-- Actions -->
                <div class="col-span-2 flex items-center justify-end gap-1">
                    <button @click="openEditModal(list)" class="text-gray-400 dark:text-gray-500 hover:text-blue-600 dark:hover:text-blue-400 transition-colors p-2 rounded-lg hover:bg-blue-50 dark:hover:bg-blue-900/20" title="Edit">
                        <svg xmlns="http://www.w3.org/2000/svg" class="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M11 5H6a2 2 0 00-2 2v11a2 2 0 002 2h11a2 2 0 002-2v-5m-1.414-9.414a2 2 0 112.828 2.828L11.828 15H9v-2.828l8.586-8.586z" />
                        </svg>
                    </button>
                    <button @click="deleteList(list.id)" class="text-gray-400 dark:text-gray-500 hover:text-red-600 dark:hover:text-red-400 transition-colors p-2 rounded-lg hover:bg-red-50 dark:hover:bg-red-900/20" title="Delete">
                        <svg xmlns="http://www.w3.org/2000/svg" class="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16" />
                        </svg>
                    </button>
                </div>
            </div>

            <!-- Mobile Layout -->
            <div class="lg:hidden">
                <div class="relative rounded-xl overflow-hidden">
                    <!-- Main Content -->
                    <div class="space-y-4">
                        <!-- Header: Name + Actions -->
                        <div class="flex items-start justify-between gap-3">
                            <div class="flex items-center gap-3 min-w-0 flex-1">
                                <div class="flex-shrink-0 w-12 h-12 rounded-xl bg-red-500 flex items-center justify-center shadow-lg shadow-red-200">
                                    <svg xmlns="http://www.w3.org/2000/svg" class="h-6 w-6 text-white" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                                        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 15v2m-6 4h12a2 2 0 002-2v-6a2 2 0 00-2-2H6a2 2 0 00-2 2v6a2 2 0 002 2zm10-10V7a4 4 0 00-8 0v4h8z" />
                                    </svg>
                                </div>
                                <div class="min-w-0 flex-1">
                                    <p class="text-base font-bold text-gray-900 dark:text-gray-100 truncate">{{ list.name }}</p>
                                    <div class="flex items-center gap-2 mt-0.5">
                                        <span class="text-xs text-gray-500 dark:text-gray-400 bg-gray-100 px-2 py-0.5 rounded-full">
                                            {{ list.clients.length }} user{{ list.clients.length !== 1 ? 's' : '' }}
                                        </span>
                                        <span class="text-xs text-gray-500 dark:text-gray-400 bg-gray-100 px-2 py-0.5 rounded-full">
                                            {{ list.rules?.length || 0 }} rule{{ (list.rules?.length || 0) !== 1 ? 's' : '' }}
                                        </span>
                                    </div>
                                </div>
                            </div>
                        </div>

                        <!-- Users Section -->
                        <div v-if="list.clients.length > 0" class="bg-gray-50 dark:bg-gray-700/50 rounded-xl p-3 border border-gray-200 dark:border-gray-700 dark:border-gray-600/50">
                            <div class="flex items-center gap-2 mb-2">
                                <svg xmlns="http://www.w3.org/2000/svg" class="h-4 w-4 text-gray-400 dark:text-gray-500" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4.354a4 4 0 110 5.292M15 21H3v-1a6 6 0 0112 0v1zm0 0h6v-1a6 6 0 00-9-5.197M13 7a4 4 0 11-8 0 4 4 0 018 0z" />
                                </svg>
                                <span class="text-xs font-semibold text-gray-500 dark:text-gray-400 uppercase tracking-wider">Authorized Users</span>
                            </div>
                            <div class="flex flex-wrap gap-1.5">
                                <span v-for="(client, idx) in list.clients.slice(0, 4)" :key="'mc'+idx" 
                                      class="inline-flex items-center gap-1 px-2.5 py-1 rounded-lg text-xs font-medium bg-white dark:bg-gray-800 border border-gray-200 dark:border-gray-700 text-gray-700 dark:text-gray-200 shadow-sm">
                                    <svg xmlns="http://www.w3.org/2000/svg" class="h-3 w-3 text-gray-400 dark:text-gray-500" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                                        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M16 7a4 4 0 11-8 0 4 4 0 018 0zM12 14a7 7 0 00-7 7h14a7 7 0 00-7-7z" />
                                    </svg>
                                    {{ client.username }}
                                </span>
                                <span v-if="list.clients.length > 4" class="text-xs text-gray-400 font-medium py-1 px-1">+{{ list.clients.length - 4 }} more</span>
                            </div>
                        </div>

                        <!-- Rules Section -->
                        <div v-if="list.rules && list.rules.length > 0" class="bg-gray-50 rounded-xl p-3 border border-gray-200/50">
                            <div class="flex items-center gap-2 mb-2">
                                <svg xmlns="http://www.w3.org/2000/svg" class="h-4 w-4 text-gray-400 dark:text-gray-500" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 12l2 2 4-4m5.618-4.016A11.955 11.955 0 0112 2.944a11.955 11.955 0 01-8.618 3.04A12.02 12.02 0 003 9c0 5.591 3.824 10.29 9 11.622 5.176-1.332 9-6.03 9-11.622 0-1.042-.133-2.052-.382-3.016z" />
                                </svg>
                                <span class="text-xs font-semibold text-gray-500 dark:text-gray-400 uppercase tracking-wider">IP Rules</span>
                            </div>
                            <div class="flex flex-wrap gap-1.5">
                                <span v-for="(rule, idx) in list.rules.slice(0, 3)" :key="'mr'+idx" 
                                      :class="rule.action === 'allow' 
                                        ? 'bg-white border-green-200 text-green-700' 
                                        : 'bg-white border-red-200 text-red-700'" 
                                      class="inline-flex items-center gap-1 px-2.5 py-1 rounded-lg text-xs font-medium border shadow-sm">
                                    <svg v-if="rule.action === 'allow'" xmlns="http://www.w3.org/2000/svg" class="h-3 w-3" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                                        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 13l4 4L19 7" />
                                    </svg>
                                    <svg v-else xmlns="http://www.w3.org/2000/svg" class="h-3 w-3" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                                        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
                                    </svg>
                                    {{ rule.ip }}
                                </span>
                                <span v-if="list.rules.length > 3" class="text-xs text-gray-400 font-medium py-1 px-1">+{{ list.rules.length - 3 }} more</span>
                            </div>
                        </div>

                        <!-- Action Buttons -->
                        <div class="flex items-center gap-2 pt-3 border-t border-gray-100 dark:border-gray-700">
                            <button @click="openEditModal(list)" 
                                    class="flex-1 flex items-center justify-center gap-2 bg-blue-500 hover:bg-blue-600 text-white font-semibold py-2.5 px-4 rounded-xl transition-all duration-200 shadow-md hover:shadow-lg active:scale-[0.98]">
                                <svg xmlns="http://www.w3.org/2000/svg" class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M11 5H6a2 2 0 00-2 2v11a2 2 0 002 2h11a2 2 0 002-2v-5m-1.414-9.414a2 2 0 112.828 2.828L11.828 15H9v-2.828l8.586-8.586z" />
                                </svg>
                                Edit
                            </button>
                            <button @click="deleteList(list.id)" 
                                    class="flex items-center justify-center gap-2 bg-white dark:bg-gray-800 hover:bg-red-50 text-red-600 font-semibold py-2.5 px-4 rounded-xl border-2 border-red-200 hover:border-red-300 transition-all duration-200 active:scale-[0.98]">
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
    <div v-if="showModal" class="fixed inset-0 z-50 flex items-center justify-center bg-black bg-opacity-60 backdrop-blur-sm">
      <div class="bg-white dark:bg-gray-800 rounded-2xl shadow-2xl w-full max-w-lg overflow-hidden transform transition-all scale-100 mx-4">
        <div class="bg-red-500 px-6 py-4 flex justify-between items-center">
            <h3 class="text-lg font-bold text-white">{{ isEditing ? 'Edit Access List' : 'Add Access List' }}</h3>
            <button @click="showModal = false" class="text-white/80 hover:text-white transition-colors">
                <svg xmlns="http://www.w3.org/2000/svg" class="h-6 w-6" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
                </svg>
            </button>
        </div>
        
        <div class="p-6 space-y-4 max-h-[70vh] overflow-y-auto">
            <div>
                <label class="block text-sm font-semibold text-gray-700 dark:text-gray-200 mb-2">Name <span class="text-red-500">*</span></label>
                <input v-model="newList.name" class="w-full border border-gray-300 dark:border-gray-600 rounded-xl px-4 py-3 text-gray-700 dark:text-gray-200 bg-white dark:bg-gray-800 focus:outline-none focus:ring-2 focus:ring-red-500/50 focus:border-red-500 transition-all duration-200 shadow-sm hover:border-gray-400" placeholder="My Access List">
            </div>

            <!-- Tabs -->
            <div class="border-b border-gray-200 dark:border-gray-700 dark:border-gray-700">
                <nav class="-mb-px flex gap-1 bg-gray-100 dark:bg-gray-700 rounded-t-lg p-1">
                    <button @click="activeTab = 'users'" 
                            :class="activeTab === 'users' 
                                ? 'bg-white dark:bg-gray-800 text-red-600 shadow-sm' 
                                : 'text-gray-500 dark:text-gray-400 hover:text-gray-700 dark:hover:text-gray-300'" 
                            class="flex-1 whitespace-nowrap py-2.5 px-4 rounded-lg font-medium text-sm transition-all duration-200 flex items-center justify-center gap-2">
                        <svg xmlns="http://www.w3.org/2000/svg" class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4.354a4 4 0 110 5.292M15 21H3v-1a6 6 0 0112 0v1zm0 0h6v-1a6 6 0 00-9-5.197M13 7a4 4 0 11-8 0 4 4 0 018 0z" />
                        </svg>
                        Users
                    </button>
                    <button @click="activeTab = 'rules'" 
                            :class="activeTab === 'rules' 
                                ? 'bg-white dark:bg-gray-800 text-red-600 shadow-sm' 
                                : 'text-gray-500 dark:text-gray-400 hover:text-gray-700 dark:hover:text-gray-300'" 
                            class="flex-1 whitespace-nowrap py-2.5 px-4 rounded-lg font-medium text-sm transition-all duration-200 flex items-center justify-center gap-2">
                        <svg xmlns="http://www.w3.org/2000/svg" class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 12l2 2 4-4m5.618-4.016A11.955 11.955 0 0112 2.944a11.955 11.955 0 01-8.618 3.04A12.02 12.02 0 003 9c0 5.591 3.824 10.29 9 11.622 5.176-1.332 9-6.03 9-11.622 0-1.042-.133-2.052-.382-3.016z" />
                        </svg>
                        IP Rules
                    </button>
                </nav>
            </div>

            <!-- Users Tab -->
            <div v-if="activeTab === 'users'" class="space-y-3">
                <div v-for="(client, index) in newList.clients" :key="index" class="flex gap-2 items-start p-3 bg-gray-50 dark:bg-gray-700 rounded-xl">
                    <div class="flex-1 space-y-2">
                        <input v-model="client.username" class="w-full border border-gray-300 dark:border-gray-600 rounded-lg px-3 py-2 text-sm text-gray-700 dark:text-gray-200 bg-white dark:bg-gray-800 focus:outline-none focus:ring-2 focus:ring-red-500/50 focus:border-red-500 transition-all" placeholder="Username">
                        <input v-model="client.password" type="password" class="w-full border border-gray-300 dark:border-gray-600 rounded-lg px-3 py-2 text-sm text-gray-700 dark:text-gray-200 bg-white dark:bg-gray-800 focus:outline-none focus:ring-2 focus:ring-red-500/50 focus:border-red-500 transition-all" :placeholder="isEditing ? 'New Password (Optional)' : 'Password'">
                    </div>
                    <button @click="removeClientRow(index)" class="text-red-400 hover:text-red-600 p-2 hover:bg-red-50 rounded-lg transition-colors" v-if="newList.clients.length > 0">
                        <svg xmlns="http://www.w3.org/2000/svg" class="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16" />
                        </svg>
                    </button>
                </div>
                <button @click="addClientRow" class="w-full py-2.5 border-2 border-dashed border-gray-300 dark:border-gray-600 rounded-xl text-sm text-gray-500 dark:text-gray-400 hover:text-red-600 hover:border-red-300 transition-colors flex items-center justify-center gap-2">
                    <svg xmlns="http://www.w3.org/2000/svg" class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4v16m8-8H4" />
                    </svg>
                    Add User
                </button>
            </div>

            <!-- Rules Tab -->
            <div v-if="activeTab === 'rules'" class="space-y-3">
                <!-- Help text for IP rules -->
                <div class="bg-blue-50 border border-blue-200 rounded-xl p-3 text-sm">
                    <div class="flex items-start gap-2">
                        <svg xmlns="http://www.w3.org/2000/svg" class="h-5 w-5 text-blue-500 flex-shrink-0 mt-0.5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13 16h-1v-4h-1m1-4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
                        </svg>
                        <div class="text-blue-700">
                            <p class="font-medium">IP Format Examples:</p>
                            <ul class="mt-1 space-y-0.5 text-blue-600">
                                <li>• Single IP: <code class="bg-blue-100 px-1 rounded">192.168.1.100</code></li>
                                <li>• IP Range (CIDR): <code class="bg-blue-100 px-1 rounded">192.168.1.0/24</code> (192.168.1.0-255)</li>
                                <li>• Subnet: <code class="bg-blue-100 px-1 rounded">10.0.0.0/8</code> (10.x.x.x)</li>
                            </ul>
                        </div>
                    </div>
                </div>
                <div v-for="(rule, index) in newList.rules" :key="index" class="flex gap-2 items-center p-3 bg-gray-50 dark:bg-gray-700 rounded-xl">
                    <div class="flex-1">
                        <input v-model="rule.ip" class="w-full border border-gray-300 dark:border-gray-600 rounded-lg px-3 py-2 text-sm text-gray-700 dark:text-gray-200 bg-white dark:bg-gray-800 focus:outline-none focus:ring-2 focus:ring-red-500/50 focus:border-red-500 transition-all" placeholder="192.168.1.0/24 or 10.0.0.1">
                    </div>
                    <div class="w-28">
                        <CustomSelect 
                            v-model="rule.action" 
                            :options="actionOptions" 
                            color="red"
                        />
                    </div>
                    <button @click="removeRuleRow(index)" class="text-red-400 hover:text-red-600 p-2 hover:bg-red-50 rounded-lg transition-colors" v-if="newList.rules.length > 0">
                        <svg xmlns="http://www.w3.org/2000/svg" class="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16" />
                        </svg>
                    </button>
                </div>
                <button @click="addRuleRow" class="w-full py-2.5 border-2 border-dashed border-gray-300 dark:border-gray-600 rounded-xl text-sm text-gray-500 dark:text-gray-400 hover:text-red-600 hover:border-red-300 transition-colors flex items-center justify-center gap-2">
                    <svg xmlns="http://www.w3.org/2000/svg" class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4v16m8-8H4" />
                    </svg>
                    Add Rule
                </button>
            </div>
        </div>

        <div class="bg-gray-50 dark:bg-gray-800 px-6 py-4 border-t border-gray-200 dark:border-gray-700 flex justify-end gap-3">
          <button @click="showModal = false" class="px-5 py-2.5 border border-gray-300 dark:border-gray-600 rounded-xl text-gray-700 dark:text-gray-200 hover:bg-gray-100 dark:hover:bg-gray-700 transition-colors font-medium">Cancel</button>
          <button @click="saveList" class="px-6 py-2.5 bg-red-500 hover:bg-red-600 text-white rounded-xl shadow-md hover:shadow-lg transition-all font-semibold">Save</button>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
/* Page fade in animation */
.max-w-4xl {
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

/* Grid item animations */
.grid > div {
    animation: scaleIn 0.25s cubic-bezier(0.22, 1, 0.36, 1) forwards;
    opacity: 0;
}

.grid > div:nth-child(1) { animation-delay: 0ms; }
.grid > div:nth-child(2) { animation-delay: 40ms; }
.grid > div:nth-child(3) { animation-delay: 80ms; }
.grid > div:nth-child(4) { animation-delay: 120ms; }
.grid > div:nth-child(5) { animation-delay: 160ms; }
.grid > div:nth-child(6) { animation-delay: 200ms; }

@keyframes scaleIn {
    from {
        opacity: 0;
        transform: scale(0.96);
    }
    to {
        opacity: 1;
        transform: scale(1);
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
