<script setup lang="ts">
import { ref, onMounted, onUnmounted, computed } from 'vue'
import { toast } from 'vue-sonner'
import { useConfirm } from '../composables/useConfirm'
import { useWebSocket } from '../composables/useWebSocket'
import { authFetch } from '../composables/useAuth'

const activeTab = ref('users')
const { on, off } = useWebSocket()
const searchQuery = ref('')

const users = ref<any[]>([])
const showUserModal = ref(false)
const isEditingUser = ref(false)
const currentUser = ref<any>(null)
const userForm = ref({
    id: 0,
    username: '',
    password: '',
    role_id: 0
})

const roles = ref<any[]>([])
const showRoleModal = ref(false)
const showRoleDropdown = ref(false)
const isEditingRole = ref(false)
const roleForm = ref({
    id: 0,
    name: '',
    description: '',
    permission_ids: [] as number[]
})

const selectedRoleName = computed(() => {
    const role = roles.value.find(r => r.id === userForm.value.role_id)
    return role?.name || 'Select a role'
})

const selectedRoleDescription = computed(() => {
    const role = roles.value.find(r => r.id === userForm.value.role_id)
    return role?.description || 'Choose a role for this user'
})

const selectRole = (role: any) => {
    userForm.value.role_id = role.id
    showRoleDropdown.value = false
}

const filteredUsers = computed(() => {
    if (!searchQuery.value.trim()) return users.value
    const q = searchQuery.value.toLowerCase()
    return users.value.filter(u => 
        u.username?.toLowerCase().includes(q) ||
        u.role?.name?.toLowerCase().includes(q)
    )
})

const filteredRoles = computed(() => {
    if (!searchQuery.value.trim()) return roles.value
    const q = searchQuery.value.toLowerCase()
    return roles.value.filter(r => 
        r.name?.toLowerCase().includes(q) ||
        r.description?.toLowerCase().includes(q)
    )
})

const permissions = ref<any[]>([])
const groupedPermissions = computed(() => {
    const groups: Record<string, { read?: any, create?: any, update?: any, delete?: any }> = {}
    permissions.value.forEach(p => {
        const [resource, action] = p.code.split(':')
        const resourceName = resource.split('_').map((w: string) => w.charAt(0).toUpperCase() + w.slice(1)).join(' ')
        
        if (!groups[resourceName]) {
            groups[resourceName] = {}
        }
        
        if (action === 'read') groups[resourceName].read = p
        else if (action === 'create') groups[resourceName].create = p
        else if (action === 'update') groups[resourceName].update = p
        else if (action === 'delete') groups[resourceName].delete = p
    })
    return groups
})

const isLoading = ref(false)
const { confirm } = useConfirm()

const fetchUsers = async () => {
    const res = await authFetch('/api/users')
    if (res.ok) {
        users.value = await res.json()
    }
}

const fetchRoles = async () => {
    const res = await authFetch('/api/roles')
    if (res.ok) {
        roles.value = await res.json()
    }
}

const fetchPermissions = async () => {
    const res = await authFetch('/api/permissions')
    if (res.ok) {
        permissions.value = await res.json()
    }
}

const fetchProfile = async () => {
    const res = await authFetch('/api/profile')
    if (res.ok) {
        currentUser.value = await res.json()
    }
}

const openCreateUserModal = () => {
    isEditingUser.value = false
    userForm.value = { id: 0, username: '', password: '', role_id: roles.value[0]?.id || 0 }
    showUserModal.value = true
}

const openEditUserModal = (user: any) => {
    isEditingUser.value = true
    userForm.value = { 
        id: user.id, 
        username: user.username, 
        password: '', 
        role_id: user.role_id || 0 
    }
    showUserModal.value = true
}

const saveUser = async () => {
    if (!userForm.value.username) {
        toast.error('Username is required')
        return
    }
    if (!isEditingUser.value && !userForm.value.password) {
        toast.error('Password is required for new users')
        return
    }

    isLoading.value = true
    const url = isEditingUser.value ? `/api/users/${userForm.value.id}` : '/api/users'
    const method = isEditingUser.value ? 'PUT' : 'POST'

    try {
        const res = await authFetch(url, {
            method,
            headers: { 
                'Content-Type': 'application/json'
            },
            body: JSON.stringify(userForm.value)
        })

        if (res.ok) {
            toast.success(isEditingUser.value ? 'User updated' : 'User created')
            showUserModal.value = false
        } else {
            const data = await res.json()
            toast.error(data.error || 'Failed to save user')
        }
    } catch (e) {
        toast.error('Connection error')
    } finally {
        isLoading.value = false
    }
}

const deleteUser = async (user: any) => {
    if (currentUser.value && user.id === currentUser.value.id) {
        toast.error('You cannot delete yourself')
        return
    }

    const confirmed = await confirm(
        'Delete User',
        `Are you sure you want to delete user "${user.username}"?`,
        { type: 'danger', confirmText: 'Delete' }
    )
    if (!confirmed) return

    const res = await authFetch(`/api/users/${user.id}`, {
        method: 'DELETE'
    })

    if (res.ok) {
        toast.success('User deleted')
        fetchUsers()
    } else {
        toast.error('Failed to delete user')
    }
}

const openCreateRoleModal = () => {
    isEditingRole.value = false
    roleForm.value = { id: 0, name: '', description: '', permission_ids: [] }
    showRoleModal.value = true
}

const openEditRoleModal = (role: any) => {
    isEditingRole.value = true
    roleForm.value = { 
        id: role.id, 
        name: role.name, 
        description: role.description,
        permission_ids: role.permissions ? role.permissions.map((p: any) => p.id) : []
    }
    showRoleModal.value = true
}

const saveRole = async () => {
    if (!roleForm.value.name) {
        toast.error('Role name is required')
        return
    }

    isLoading.value = true
    const url = isEditingRole.value ? `/api/roles/${roleForm.value.id}` : '/api/roles'
    const method = isEditingRole.value ? 'PUT' : 'POST'

    try {
        const res = await authFetch(url, {
            method,
            headers: { 
                'Content-Type': 'application/json'
            },
            body: JSON.stringify(roleForm.value)
        })

        if (res.ok) {
            toast.success(isEditingRole.value ? 'Role updated' : 'Role created')
            showRoleModal.value = false
        } else {
            const data = await res.json()
            toast.error(data.error || 'Failed to save role')
        }
    } catch (e) {
        toast.error('Connection error')
    } finally {
        isLoading.value = false
    }
}

const deleteRole = async (role: any) => {
    const confirmed = await confirm(
        'Delete Role',
        `Are you sure you want to delete role "${role.name}"?`,
        { type: 'danger', confirmText: 'Delete' }
    )
    if (!confirmed) return

    const res = await authFetch(`/api/roles/${role.id}`, {
        method: 'DELETE'
    })

    if (res.ok) {
        toast.success('Role deleted')
    } else {
        const data = await res.json()
        toast.error(data.error || 'Failed to delete role')
    }
}

const handleUserCreated = (user: any) => {
    if (!users.value.find(u => u.id === user.id)) {
        users.value.push(user)
    }
}

const handleUserUpdated = (user: any) => {
    const index = users.value.findIndex(u => u.id === user.id)
    if (index !== -1) {
        users.value[index] = user
    }
}

const handleUserDeleted = (payload: { id: number }) => {
    users.value = users.value.filter(u => u.id !== Number(payload.id))
}

const handleRoleCreated = (role: any) => {
    if (!roles.value.find(r => r.id === role.id)) {
        roles.value.push(role)
    }
}

const handleRoleUpdated = (role: any) => {
    const index = roles.value.findIndex(r => r.id === role.id)
    if (index !== -1) {
        roles.value[index] = role
    }
    fetchUsers()
}

const handleRoleDeleted = (payload: { id: number }) => {
    roles.value = roles.value.filter(r => r.id !== Number(payload.id))
}

onMounted(() => {
    fetchUsers()
    fetchRoles()
    fetchPermissions()
    fetchProfile()
    
    on('user_created', handleUserCreated)
    on('user_updated', handleUserUpdated)
    on('user_deleted', handleUserDeleted)
    on('role_created', handleRoleCreated)
    on('role_updated', handleRoleUpdated)
    on('role_deleted', handleRoleDeleted)
})

onUnmounted(() => {
    off('user_created', handleUserCreated)
    off('user_updated', handleUserUpdated)
    off('user_deleted', handleUserDeleted)
    off('role_created', handleRoleCreated)
    off('role_updated', handleRoleUpdated)
    off('role_deleted', handleRoleDeleted)
})
</script>

<template>
    <div>
        <!-- Header & Tabs -->
        <div class="bg-white rounded-t-xl p-5 border-b border-gray-200 shadow-sm overflow-hidden relative">
            <!-- Gradient accent bar -->
            <div class="absolute top-0 left-0 right-0 h-1 bg-blue-500"></div>
            
            <div class="flex flex-col lg:flex-row justify-between items-start lg:items-center gap-4 mb-5">
                <div>
                    <h2 class="text-2xl font-bold text-gray-800 tracking-tight">User Management</h2>
                    <p class="text-sm text-gray-500 mt-1">Manage users and roles for your system</p>
                </div>
                <div class="flex items-center gap-3 w-full lg:w-auto">
                    <div class="relative flex-1 lg:flex-initial">
                        <input 
                            v-model="searchQuery" 
                            type="text" 
                            :placeholder="activeTab === 'users' ? 'Search users...' : 'Search roles...'" 
                            class="w-full lg:w-64 pl-10 pr-4 py-2.5 border border-gray-200 rounded-xl text-sm focus:ring-2 focus:ring-blue-500/20 focus:border-blue-500 transition-all bg-gray-50/50"
                        />
                        <svg xmlns="http://www.w3.org/2000/svg" class="h-4 w-4 text-gray-400 absolute left-3.5 top-1/2 -translate-y-1/2" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z" />
                        </svg>
                    </div>
                    <button 
                        v-if="activeTab === 'users'"
                        @click="openCreateUserModal"
                        class="bg-blue-500 hover:bg-blue-600 text-white px-5 py-2.5 rounded-xl shadow-lg shadow-blue-500/25 hover:shadow-blue-500/40 transition-all duration-300 text-sm font-semibold flex items-center gap-2 whitespace-nowrap"
                    >
                        <svg xmlns="http://www.w3.org/2000/svg" class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4v16m8-8H4" />
                        </svg>
                        <span class="hidden sm:inline">Add User</span>
                        <span class="sm:hidden">Add</span>
                    </button>
                    <button 
                        v-else
                        @click="openCreateRoleModal"
                        class="bg-purple-500 hover:bg-purple-600 text-white px-5 py-2.5 rounded-xl shadow-lg shadow-purple-500/25 hover:shadow-purple-500/40 transition-all duration-300 text-sm font-semibold flex items-center gap-2 whitespace-nowrap"
                    >
                        <svg xmlns="http://www.w3.org/2000/svg" class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4v16m8-8H4" />
                        </svg>
                        <span class="hidden sm:inline">Add Role</span>
                        <span class="sm:hidden">Add</span>
                    </button>
                </div>
            </div>

            <!-- Enhanced Tabs -->
            <div class="flex gap-1 bg-gray-100/80 p-1 rounded-xl w-fit">
                <button 
                    @click="activeTab = 'users'"
                    class="px-5 py-2 text-sm font-semibold transition-all duration-300 rounded-lg flex items-center gap-2"
                    :class="activeTab === 'users' ? 'bg-white text-blue-600 shadow-sm' : 'text-gray-500 hover:text-gray-700'"
                >
                    <svg xmlns="http://www.w3.org/2000/svg" class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4.354a4 4 0 110 5.292M15 21H3v-1a6 6 0 0112 0v1zm0 0h6v-1a6 6 0 00-9-5.197M13 7a4 4 0 11-8 0 4 4 0 018 0z" />
                    </svg>
                    Users
                    <span v-if="users.length" class="bg-blue-100 text-blue-600 text-xs px-2 py-0.5 rounded-full">{{ users.length }}</span>
                </button>
                <button 
                    @click="activeTab = 'roles'"
                    class="px-5 py-2 text-sm font-semibold transition-all duration-300 rounded-lg flex items-center gap-2"
                    :class="activeTab === 'roles' ? 'bg-white text-purple-600 shadow-sm' : 'text-gray-500 hover:text-gray-700'"
                >
                    <svg xmlns="http://www.w3.org/2000/svg" class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 12l2 2 4-4m5.618-4.016A11.955 11.955 0 0112 2.944a11.955 11.955 0 01-8.618 3.04A12.02 12.02 0 003 9c0 5.591 3.824 10.29 9 11.622 5.176-1.332 9-6.03 9-11.622 0-1.042-.133-2.052-.382-3.016z" />
                    </svg>
                    Roles
                    <span v-if="roles.length" class="bg-purple-100 text-purple-600 text-xs px-2 py-0.5 rounded-full">{{ roles.length }}</span>
                </button>
            </div>
        </div>

        <!-- Desktop Column Headers for Users -->
        <div v-if="activeTab === 'users'" class="bg-gray-50/80 px-5 py-3 border-b text-xs font-bold text-gray-500 uppercase tracking-wider border-x border-gray-200 hidden lg:grid lg:grid-cols-12 gap-4">
            <div class="col-span-1">ID</div>
            <div class="col-span-5">User</div>
            <div class="col-span-3">Role</div>
            <div class="col-span-3 text-right pr-2">Actions</div>
        </div>

        <!-- Users List -->
        <div v-if="activeTab === 'users'" class="bg-white rounded-b-xl shadow-lg border-x border-b border-gray-200">
            <!-- Empty State -->
            <div v-if="users.length === 0" class="p-12 text-center">
                <div class="mb-4">
                    <div class="mx-auto w-20 h-20 rounded-2xl bg-blue-500 flex items-center justify-center shadow-lg shadow-blue-500/30">
                        <svg xmlns="http://www.w3.org/2000/svg" class="h-10 w-10 text-white" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M12 4.354a4 4 0 110 5.292M15 21H3v-1a6 6 0 0112 0v1zm0 0h6v-1a6 6 0 00-9-5.197M13 7a4 4 0 11-8 0 4 4 0 018 0z" />
                        </svg>
                    </div>
                </div>
                <p class="text-lg font-bold text-gray-700 mb-2">No users found</p>
                <p class="text-sm text-gray-400 mb-6">Create users to manage access to your system</p>
                <button @click="openCreateUserModal" class="inline-flex items-center gap-2 bg-blue-500 hover:bg-blue-600 text-white px-5 py-2.5 rounded-xl shadow-lg shadow-blue-500/25 transition-all text-sm font-semibold">
                    <svg xmlns="http://www.w3.org/2000/svg" class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4v16m8-8H4" />
                    </svg>
                    Add User
                </button>
            </div>

            <!-- No Results -->
            <div v-else-if="filteredUsers.length === 0" class="p-12 text-center">
                <div class="mb-4">
                    <div class="mx-auto w-20 h-20 rounded-2xl bg-gray-100 flex items-center justify-center">
                        <svg xmlns="http://www.w3.org/2000/svg" class="h-10 w-10 text-gray-400" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z" />
                        </svg>
                    </div>
                </div>
                <p class="text-lg font-bold text-gray-700 mb-2">No results found</p>
                <p class="text-sm text-gray-400">No users match "{{ searchQuery }}"</p>
            </div>

            <!-- Users List -->
            <template v-else>
                <!-- Desktop View -->
                <div v-for="user in filteredUsers" :key="user.id" class="hidden lg:grid lg:grid-cols-12 gap-4 p-4 border-b border-gray-100 hover:bg-gray-50/50 transition-all duration-200 last:border-b-0 group items-center">
                    <!-- ID -->
                    <div class="col-span-1">
                        <span class="text-sm text-gray-400 font-mono">#{{ user.id }}</span>
                    </div>
                    
                    <!-- User Info -->
                    <div class="col-span-5 flex items-center gap-3">
                        <div class="w-10 h-10 rounded-xl bg-blue-500 text-white flex items-center justify-center text-sm font-bold shadow-lg shadow-blue-500/20">
                            {{ user.username.charAt(0).toUpperCase() }}
                        </div>
                        <div>
                            <p class="text-sm font-bold text-gray-900 flex items-center gap-2">
                                {{ user.username }}
                                <span v-if="currentUser && user.id === currentUser.id" class="px-2 py-0.5 rounded-full text-[10px] font-bold bg-green-500 text-white">YOU</span>
                            </p>
                            <p class="text-xs text-gray-400">User Account</p>
                        </div>
                    </div>
                    
                    <!-- Role -->
                    <div class="col-span-3">
                        <span class="px-3 py-1.5 rounded-lg text-xs font-semibold bg-gray-100 text-gray-700 border border-gray-200">
                            {{ user.role?.name || 'No Role' }}
                        </span>
                    </div>
                    
                    <!-- Actions -->
                    <div class="col-span-3 flex justify-end gap-2">
                        <button @click="openEditUserModal(user)" class="p-2 text-gray-400 hover:text-blue-600 hover:bg-blue-50 rounded-lg transition-all" title="Edit">
                            <svg xmlns="http://www.w3.org/2000/svg" class="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M11 5H6a2 2 0 00-2 2v11a2 2 0 002 2h11a2 2 0 002-2v-5m-1.414-9.414a2 2 0 112.828 2.828L11.828 15H9v-2.828l8.586-8.586z" />
                            </svg>
                        </button>
                        <button 
                            @click="deleteUser(user)" 
                            :disabled="currentUser && user.id === currentUser.id"
                            :class="currentUser && user.id === currentUser.id ? 'text-gray-300 cursor-not-allowed' : 'text-gray-400 hover:text-red-600 hover:bg-red-50'"
                            class="p-2 rounded-lg transition-all"
                            title="Delete"
                        >
                            <svg xmlns="http://www.w3.org/2000/svg" class="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16" />
                            </svg>
                        </button>
                    </div>
                </div>

                <!-- Mobile View -->
                <div v-for="user in filteredUsers" :key="'mobile-' + user.id" class="lg:hidden p-4 border-b border-gray-100 last:border-b-0">
                    <div class="bg-white rounded-xl border border-gray-200 overflow-hidden shadow-sm hover:shadow-md transition-all duration-300">
                        <!-- Status indicator strip -->
                        <div class="h-1 bg-blue-500"></div>
                        
                        <div class="p-4">
                            <!-- Header -->
                            <div class="flex items-center gap-3 mb-4">
                                <div class="w-12 h-12 rounded-xl bg-blue-500 text-white flex items-center justify-center text-lg font-bold shadow-lg shadow-blue-500/20">
                                    {{ user.username.charAt(0).toUpperCase() }}
                                </div>
                                <div class="flex-1 min-w-0">
                                    <p class="text-base font-bold text-gray-900 truncate flex items-center gap-2">
                                        {{ user.username }}
                                        <span v-if="currentUser && user.id === currentUser.id" class="px-2 py-0.5 rounded-full text-[10px] font-bold bg-green-500 text-white">YOU</span>
                                    </p>
                                    <p class="text-xs text-gray-400">User ID: #{{ user.id }}</p>
                                </div>
                            </div>
                            
                            <!-- Role Badge -->
                            <div class="flex items-center gap-2 mb-4">
                                <span class="text-xs text-gray-400">Role:</span>
                                <span class="px-3 py-1 rounded-lg text-xs font-semibold bg-gray-100 text-gray-700 border border-gray-200">
                                    {{ user.role?.name || 'No Role' }}
                                </span>
                            </div>
                            
                            <!-- Actions -->
                            <div class="flex gap-2 pt-3 border-t border-gray-100">
                                <button @click="openEditUserModal(user)" class="flex-1 flex items-center justify-center gap-2 px-4 py-2.5 bg-blue-500 hover:bg-blue-600 text-white rounded-lg text-sm font-semibold transition-all shadow-sm">
                                    <svg xmlns="http://www.w3.org/2000/svg" class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                                        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M11 5H6a2 2 0 00-2 2v11a2 2 0 002 2h11a2 2 0 002-2v-5m-1.414-9.414a2 2 0 112.828 2.828L11.828 15H9v-2.828l8.586-8.586z" />
                                    </svg>
                                    Edit
                                </button>
                                <button 
                                    @click="deleteUser(user)"
                                    :disabled="currentUser && user.id === currentUser.id"
                                    :class="currentUser && user.id === currentUser.id ? 'opacity-40 cursor-not-allowed' : 'hover:bg-red-50 hover:border-red-200'"
                                    class="flex items-center justify-center gap-2 px-4 py-2.5 border border-gray-200 text-red-500 rounded-lg text-sm font-semibold transition-all"
                                >
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

        <!-- Desktop Column Headers for Roles -->
        <div v-if="activeTab === 'roles'" class="bg-gray-50/80 px-5 py-3 border-b text-xs font-bold text-gray-500 uppercase tracking-wider border-x border-gray-200 hidden lg:grid lg:grid-cols-12 gap-4">
            <div class="col-span-1">ID</div>
            <div class="col-span-3">Role Name</div>
            <div class="col-span-5">Description</div>
            <div class="col-span-3 text-right pr-2">Actions</div>
        </div>

        <!-- Roles List -->
        <div v-if="activeTab === 'roles'" class="bg-white rounded-b-xl shadow-lg border-x border-b border-gray-200">
            <!-- Empty State -->
            <div v-if="roles.length === 0" class="p-12 text-center">
                <div class="mb-4">
                    <div class="mx-auto w-20 h-20 rounded-2xl bg-purple-500 flex items-center justify-center shadow-lg shadow-purple-500/30">
                        <svg xmlns="http://www.w3.org/2000/svg" class="h-10 w-10 text-white" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M9 12l2 2 4-4m5.618-4.016A11.955 11.955 0 0112 2.944a11.955 11.955 0 01-8.618 3.04A12.02 12.02 0 003 9c0 5.591 3.824 10.29 9 11.622 5.176-1.332 9-6.03 9-11.622 0-1.042-.133-2.052-.382-3.016z" />
                        </svg>
                    </div>
                </div>
                <p class="text-lg font-bold text-gray-700 mb-2">No roles defined</p>
                <p class="text-sm text-gray-400 mb-6">Create roles to manage permissions and access levels</p>
                <button @click="openCreateRoleModal" class="inline-flex items-center gap-2 bg-purple-500 hover:bg-purple-600 text-white px-5 py-2.5 rounded-xl shadow-lg shadow-purple-500/25 transition-all text-sm font-semibold">
                    <svg xmlns="http://www.w3.org/2000/svg" class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4v16m8-8H4" />
                    </svg>
                    Add Role
                </button>
            </div>

            <!-- No Results -->
            <div v-else-if="filteredRoles.length === 0" class="p-12 text-center">
                <div class="mb-4">
                    <div class="mx-auto w-20 h-20 rounded-2xl bg-gray-100 flex items-center justify-center">
                        <svg xmlns="http://www.w3.org/2000/svg" class="h-10 w-10 text-gray-400" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z" />
                        </svg>
                    </div>
                </div>
                <p class="text-lg font-bold text-gray-700 mb-2">No results found</p>
                <p class="text-sm text-gray-400">No roles match "{{ searchQuery }}"</p>
            </div>

            <!-- Roles List -->
            <template v-else>
                <!-- Desktop View -->
                <div v-for="role in filteredRoles" :key="role.id" class="hidden lg:grid lg:grid-cols-12 gap-4 p-4 border-b border-gray-100 hover:bg-gray-50/50 transition-all duration-200 last:border-b-0 group items-center">
                    <!-- ID -->
                    <div class="col-span-1">
                        <span class="text-sm text-gray-400 font-mono">#{{ role.id }}</span>
                    </div>
                    
                    <!-- Role Name -->
                    <div class="col-span-3 flex items-center gap-3">
                        <div class="w-10 h-10 rounded-xl bg-purple-500 text-white flex items-center justify-center shadow-lg shadow-purple-500/20">
                            <svg xmlns="http://www.w3.org/2000/svg" class="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 12l2 2 4-4m5.618-4.016A11.955 11.955 0 0112 2.944a11.955 11.955 0 01-8.618 3.04A12.02 12.02 0 003 9c0 5.591 3.824 10.29 9 11.622 5.176-1.332 9-6.03 9-11.622 0-1.042-.133-2.052-.382-3.016z" />
                            </svg>
                        </div>
                        <span class="text-sm font-bold text-gray-900">{{ role.name }}</span>
                    </div>
                    
                    <!-- Description -->
                    <div class="col-span-5">
                        <p class="text-sm text-gray-500 truncate">{{ role.description || 'No description' }}</p>
                    </div>
                    
                    <!-- Actions -->
                    <div class="col-span-3 flex justify-end gap-2">
                        <button @click="openEditRoleModal(role)" class="p-2 text-gray-400 hover:text-purple-600 hover:bg-purple-50 rounded-lg transition-all" title="Edit">
                            <svg xmlns="http://www.w3.org/2000/svg" class="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M11 5H6a2 2 0 00-2 2v11a2 2 0 002 2h11a2 2 0 002-2v-5m-1.414-9.414a2 2 0 112.828 2.828L11.828 15H9v-2.828l8.586-8.586z" />
                            </svg>
                        </button>
                        <button @click="deleteRole(role)" class="p-2 text-gray-400 hover:text-red-600 hover:bg-red-50 rounded-lg transition-all" title="Delete">
                            <svg xmlns="http://www.w3.org/2000/svg" class="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16" />
                            </svg>
                        </button>
                    </div>
                </div>

                <!-- Mobile View -->
                <div v-for="role in filteredRoles" :key="'mobile-' + role.id" class="lg:hidden p-4 border-b border-gray-100 last:border-b-0">
                    <div class="bg-white rounded-xl border border-gray-200 overflow-hidden shadow-sm hover:shadow-md transition-all duration-300">
                        <!-- Status indicator strip -->
                        <div class="h-1 bg-purple-500"></div>
                        
                        <div class="p-4">
                            <!-- Header -->
                            <div class="flex items-center gap-3 mb-3">
                                <div class="w-12 h-12 rounded-xl bg-purple-500 text-white flex items-center justify-center shadow-lg shadow-purple-500/20">
                                    <svg xmlns="http://www.w3.org/2000/svg" class="h-6 w-6" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                                        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 12l2 2 4-4m5.618-4.016A11.955 11.955 0 0112 2.944a11.955 11.955 0 01-8.618 3.04A12.02 12.02 0 003 9c0 5.591 3.824 10.29 9 11.622 5.176-1.332 9-6.03 9-11.622 0-1.042-.133-2.052-.382-3.016z" />
                                    </svg>
                                </div>
                                <div class="flex-1 min-w-0">
                                    <p class="text-base font-bold text-gray-900 truncate">{{ role.name }}</p>
                                    <p class="text-xs text-gray-400">Role ID: #{{ role.id }}</p>
                                </div>
                            </div>
                            
                            <!-- Description -->
                            <div class="bg-gray-50 rounded-lg p-3 mb-4">
                                <p class="text-xs text-gray-400 uppercase tracking-wide mb-1">Description</p>
                                <p class="text-sm text-gray-600">{{ role.description || 'No description' }}</p>
                            </div>
                            
                            <!-- Actions -->
                            <div class="flex gap-2 pt-3 border-t border-gray-100">
                                <button @click="openEditRoleModal(role)" class="flex-1 flex items-center justify-center gap-2 px-4 py-2.5 bg-purple-500 hover:bg-purple-600 text-white rounded-lg text-sm font-semibold transition-all shadow-sm">
                                    <svg xmlns="http://www.w3.org/2000/svg" class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                                        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M11 5H6a2 2 0 00-2 2v11a2 2 0 002 2h11a2 2 0 002-2v-5m-1.414-9.414a2 2 0 112.828 2.828L11.828 15H9v-2.828l8.586-8.586z" />
                                    </svg>
                                    Edit
                                </button>
                                <button @click="deleteRole(role)" class="flex items-center justify-center gap-2 px-4 py-2.5 border border-gray-200 text-red-500 hover:bg-red-50 hover:border-red-200 rounded-lg text-sm font-semibold transition-all">
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

        <!-- User Modal -->
        <div v-if="showUserModal" class="fixed inset-0 bg-black/60 flex items-center justify-center z-50 p-4 backdrop-blur-sm">
            <div class="bg-white rounded-2xl shadow-2xl w-full max-w-md transform transition-all scale-100 overflow-hidden">
                <!-- Gradient Header -->
                <div class="bg-blue-500 px-6 py-4 flex justify-between items-center">
                    <h3 class="text-lg font-bold text-white">{{ isEditingUser ? 'Edit User' : 'Add User' }}</h3>
                    <button @click="showUserModal = false" class="text-white/80 hover:text-white transition-colors">
                        <svg xmlns="http://www.w3.org/2000/svg" class="h-6 w-6" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
                        </svg>
                    </button>
                </div>
                
                <div class="p-6 space-y-5">
                    <div>
                        <label for="userUsername" class="block text-sm font-semibold text-gray-700 mb-2">Username <span class="text-red-500">*</span></label>
                        <input 
                            id="userUsername"
                            v-model="userForm.username" 
                            type="text" 
                            class="w-full border border-gray-200 rounded-xl px-4 py-3 focus:ring-2 focus:ring-blue-500/20 focus:border-blue-500 outline-none transition-all bg-gray-50/50"
                            placeholder="Enter username"
                        />
                    </div>
                    <div class="relative">
                        <label for="userRole" class="block text-sm font-semibold text-gray-700 mb-2">Role <span class="text-red-500">*</span></label>
                        
                        <!-- Custom Dropdown Button -->
                        <button 
                            type="button"
                            @click="showRoleDropdown = !showRoleDropdown"
                            class="w-full border border-gray-200 rounded-xl px-4 py-3 focus:ring-2 focus:ring-blue-500/20 focus:border-blue-500 outline-none transition-all bg-gray-50/50 flex items-center justify-between gap-3 text-left hover:border-blue-300"
                        >
                            <div class="flex items-center gap-3 min-w-0">
                                <div class="w-8 h-8 rounded-lg bg-gradient-to-br from-purple-500 to-purple-600 text-white flex items-center justify-center flex-shrink-0 shadow-sm">
                                    <svg xmlns="http://www.w3.org/2000/svg" class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                                        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 12l2 2 4-4m5.618-4.016A11.955 11.955 0 0112 2.944a11.955 11.955 0 01-8.618 3.04A12.02 12.02 0 003 9c0 5.591 3.824 10.29 9 11.622 5.176-1.332 9-6.03 9-11.622 0-1.042-.133-2.052-.382-3.016z" />
                                    </svg>
                                </div>
                                <div class="min-w-0">
                                    <p class="text-sm font-semibold text-gray-800 truncate">{{ selectedRoleName }}</p>
                                    <p class="text-xs text-gray-400 truncate">{{ selectedRoleDescription }}</p>
                                </div>
                            </div>
                            <svg xmlns="http://www.w3.org/2000/svg" class="h-5 w-5 text-gray-400 flex-shrink-0 transition-transform duration-200" :class="{ 'rotate-180': showRoleDropdown }" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 9l-7 7-7-7" />
                            </svg>
                        </button>
                        
                        <!-- Dropdown Options -->
                        <Transition
                            enter-active-class="transition duration-150 ease-out"
                            enter-from-class="transform scale-95 opacity-0"
                            enter-to-class="transform scale-100 opacity-100"
                            leave-active-class="transition duration-100 ease-in"
                            leave-from-class="transform scale-100 opacity-100"
                            leave-to-class="transform scale-95 opacity-0"
                        >
                            <div 
                                v-if="showRoleDropdown" 
                                class="absolute z-50 mt-2 w-full bg-white border border-gray-200 rounded-xl shadow-xl overflow-hidden"
                            >
                                <div class="max-h-64 overflow-y-auto py-1">
                                    <button
                                        v-for="role in roles"
                                        :key="role.id"
                                        type="button"
                                        @click="selectRole(role)"
                                        class="w-full px-4 py-3 flex items-center gap-3 hover:bg-blue-50 transition-colors text-left"
                                        :class="{ 'bg-blue-50 border-l-2 border-blue-500': userForm.role_id === role.id }"
                                    >
                                        <div class="w-8 h-8 rounded-lg flex items-center justify-center flex-shrink-0 shadow-sm" :class="userForm.role_id === role.id ? 'bg-gradient-to-br from-blue-500 to-blue-600 text-white' : 'bg-gray-100 text-gray-500'">
                                            <svg xmlns="http://www.w3.org/2000/svg" class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                                                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 12l2 2 4-4m5.618-4.016A11.955 11.955 0 0112 2.944a11.955 11.955 0 01-8.618 3.04A12.02 12.02 0 003 9c0 5.591 3.824 10.29 9 11.622 5.176-1.332 9-6.03 9-11.622 0-1.042-.133-2.052-.382-3.016z" />
                                            </svg>
                                        </div>
                                        <div class="min-w-0 flex-1">
                                            <p class="text-sm font-semibold truncate" :class="userForm.role_id === role.id ? 'text-blue-600' : 'text-gray-800'">{{ role.name }}</p>
                                            <p class="text-xs text-gray-400 truncate">{{ role.description || 'No description' }}</p>
                                        </div>
                                        <svg v-if="userForm.role_id === role.id" xmlns="http://www.w3.org/2000/svg" class="h-5 w-5 text-blue-500 flex-shrink-0" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                                            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 13l4 4L19 7" />
                                        </svg>
                                    </button>
                                </div>
                            </div>
                        </Transition>
                        
                        <!-- Click outside to close -->
                        <div v-if="showRoleDropdown" @click="showRoleDropdown = false" class="fixed inset-0 z-40"></div>
                    </div>
                    <div>
                        <label for="userPassword" class="block text-sm font-semibold text-gray-700 mb-2">
                            {{ isEditingUser ? 'New Password (leave blank to keep current)' : 'Password' }}
                        </label>
                        <input 
                            id="userPassword"
                            v-model="userForm.password" 
                            type="password" 
                            class="w-full border border-gray-200 rounded-xl px-4 py-3 focus:ring-2 focus:ring-blue-500/20 focus:border-blue-500 outline-none transition-all bg-gray-50/50"
                            placeholder="••••••••"
                        />
                    </div>
                </div>

                <div class="px-6 py-4 bg-gray-50 flex justify-end gap-3 border-t border-gray-100">
                    <button 
                        @click="showUserModal = false" 
                        class="px-5 py-2.5 text-gray-600 hover:text-gray-800 font-semibold transition-colors rounded-xl hover:bg-gray-100"
                    >
                        Cancel
                    </button>
                    <button 
                        @click="saveUser" 
                        :disabled="isLoading"
                        class="bg-blue-500 hover:bg-blue-600 text-white px-6 py-2.5 rounded-xl shadow-lg shadow-blue-500/25 transition-all font-semibold flex items-center gap-2 disabled:opacity-50"
                    >
                        <svg v-if="isLoading" class="animate-spin h-4 w-4 text-white" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24">
                            <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle>
                            <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
                        </svg>
                        {{ isEditingUser ? 'Update' : 'Create' }}
                    </button>
                </div>
            </div>
        </div>

        <!-- Role Modal -->
        <div v-if="showRoleModal" class="fixed inset-0 bg-black/60 flex items-center justify-center z-50 p-4 backdrop-blur-sm">
            <div class="bg-white rounded-2xl shadow-2xl w-full max-w-lg transform transition-all scale-100 overflow-hidden max-h-[90vh] flex flex-col">
                <!-- Gradient Header -->
                <div class="bg-purple-500 px-6 py-4 flex justify-between items-center flex-shrink-0">
                    <h3 class="text-lg font-bold text-white">{{ isEditingRole ? 'Edit Role' : 'Add Role' }}</h3>
                    <button @click="showRoleModal = false" class="text-white/80 hover:text-white transition-colors">
                        <svg xmlns="http://www.w3.org/2000/svg" class="h-6 w-6" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
                        </svg>
                    </button>
                </div>
                
                <div class="p-6 space-y-5 overflow-y-auto flex-1">
                    <div>
                        <label for="roleName" class="block text-sm font-semibold text-gray-700 mb-2">Role Name <span class="text-red-500">*</span></label>
                        <input 
                            id="roleName"
                            v-model="roleForm.name" 
                            type="text" 
                            class="w-full border border-gray-200 rounded-xl px-4 py-3 focus:ring-2 focus:ring-purple-500/20 focus:border-purple-500 outline-none transition-all bg-gray-50/50"
                            placeholder="e.g. Editor"
                        />
                    </div>
                    <div>
                        <label for="roleDesc" class="block text-sm font-semibold text-gray-700 mb-2">Description</label>
                        <textarea 
                            id="roleDesc"
                            v-model="roleForm.description" 
                            class="w-full border border-gray-200 rounded-xl px-4 py-3 focus:ring-2 focus:ring-purple-500/20 focus:border-purple-500 outline-none transition-all bg-gray-50/50"
                            placeholder="Role description..."
                            rows="3"
                        ></textarea>
                    </div>
                    
                    <div>
                        <span class="block text-sm font-semibold text-gray-700 mb-3">Permissions</span>
                        <div class="border border-gray-200 rounded-xl overflow-hidden">
                            <table class="min-w-full divide-y divide-gray-200">
                                <thead class="bg-gray-50">
                                    <tr>
                                        <th class="px-4 py-3 text-left text-xs font-bold text-gray-600 uppercase tracking-wider">Resource</th>
                                        <th class="px-4 py-3 text-center text-xs font-bold text-gray-600 uppercase tracking-wider">Read</th>
                                        <th class="px-4 py-3 text-center text-xs font-bold text-gray-600 uppercase tracking-wider">Create</th>
                                        <th class="px-4 py-3 text-center text-xs font-bold text-gray-600 uppercase tracking-wider">Update</th>
                                        <th class="px-4 py-3 text-center text-xs font-bold text-gray-600 uppercase tracking-wider">Delete</th>
                                    </tr>
                                </thead>
                                <tbody class="bg-white divide-y divide-gray-100">
                                    <tr v-for="(perms, resource) in groupedPermissions" :key="resource" class="hover:bg-gray-50/50 transition-colors">
                                        <td class="px-4 py-3 text-sm font-semibold text-gray-800">{{ resource }}</td>
                                        <td class="px-4 py-3 text-center">
                                            <input v-if="perms.read" type="checkbox" :value="perms.read.id" v-model="roleForm.permission_ids" class="w-4 h-4 text-purple-600 border-gray-300 rounded focus:ring-purple-500" />
                                            <span v-else class="text-gray-300">-</span>
                                        </td>
                                        <td class="px-4 py-3 text-center">
                                            <input v-if="perms.create" type="checkbox" :value="perms.create.id" v-model="roleForm.permission_ids" class="w-4 h-4 text-purple-600 border-gray-300 rounded focus:ring-purple-500" />
                                            <span v-else class="text-gray-300">-</span>
                                        </td>
                                        <td class="px-4 py-3 text-center">
                                            <input v-if="perms.update" type="checkbox" :value="perms.update.id" v-model="roleForm.permission_ids" class="w-4 h-4 text-purple-600 border-gray-300 rounded focus:ring-purple-500" />
                                            <span v-else class="text-gray-300">-</span>
                                        </td>
                                        <td class="px-4 py-3 text-center">
                                            <input v-if="perms.delete" type="checkbox" :value="perms.delete.id" v-model="roleForm.permission_ids" class="w-4 h-4 text-purple-600 border-gray-300 rounded focus:ring-purple-500" />
                                            <span v-else class="text-gray-300">-</span>
                                        </td>
                                    </tr>
                                </tbody>
                            </table>
                        </div>
                    </div>
                </div>

                <div class="px-6 py-4 bg-gray-50 flex justify-end gap-3 border-t border-gray-100 flex-shrink-0">
                    <button 
                        @click="showRoleModal = false" 
                        class="px-5 py-2.5 text-gray-600 hover:text-gray-800 font-semibold transition-colors rounded-xl hover:bg-gray-100"
                    >
                        Cancel
                    </button>
                    <button 
                        @click="saveRole" 
                        :disabled="isLoading"
                        class="bg-purple-500 hover:bg-purple-600 text-white px-6 py-2.5 rounded-xl shadow-lg shadow-purple-500/25 transition-all font-semibold flex items-center gap-2 disabled:opacity-50"
                    >
                        <svg v-if="isLoading" class="animate-spin h-4 w-4 text-white" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24">
                            <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle>
                            <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
                        </svg>
                        {{ isEditingRole ? 'Update' : 'Create' }}
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

/* Table row animations */
tbody tr {
    animation: slideIn 0.25s cubic-bezier(0.22, 1, 0.36, 1) forwards;
    opacity: 0;
}

tbody tr:nth-child(1) { animation-delay: 0ms; }
tbody tr:nth-child(2) { animation-delay: 30ms; }
tbody tr:nth-child(3) { animation-delay: 60ms; }
tbody tr:nth-child(4) { animation-delay: 90ms; }
tbody tr:nth-child(5) { animation-delay: 120ms; }
tbody tr:nth-child(6) { animation-delay: 150ms; }
tbody tr:nth-child(7) { animation-delay: 180ms; }
tbody tr:nth-child(8) { animation-delay: 210ms; }

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

/* Tab transition */
.border-b-2 {
    transition: all 0.2s cubic-bezier(0.22, 1, 0.36, 1);
}
</style>
