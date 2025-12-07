<script setup lang="ts">
import { ref, watch } from 'vue'

const props = defineProps<{
    isOpen: boolean
    title: string
    message: string
    confirmText?: string
    cancelText?: string
    type?: 'danger' | 'warning' | 'info'
}>()

const emit = defineEmits(['confirm', 'cancel'])

const confirm = () => {
    emit('confirm')
}

const cancel = () => {
    emit('cancel')
}
</script>

<template>
    <div v-if="isOpen" class="fixed inset-0 z-[60] flex items-center justify-center bg-black bg-opacity-60 backdrop-blur-sm transition-opacity">
        <div class="bg-white rounded-lg shadow-2xl w-full max-w-md transform transition-all scale-100 overflow-hidden border border-gray-100">
            <div class="p-6">
                <div class="flex items-center gap-4 mb-4">
                    <div v-if="type === 'danger'" class="flex-shrink-0 w-10 h-10 rounded-full bg-red-100 flex items-center justify-center">
                        <svg xmlns="http://www.w3.org/2000/svg" class="h-6 w-6 text-red-600" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-3L13.732 4c-.77-1.333-2.694-1.333-3.464 0L3.34 16c-.77 1.333.192 3 1.732 3z" />
                        </svg>
                    </div>
                    <div v-else-if="type === 'warning'" class="flex-shrink-0 w-10 h-10 rounded-full bg-yellow-100 flex items-center justify-center">
                         <svg xmlns="http://www.w3.org/2000/svg" class="h-6 w-6 text-yellow-600" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-3L13.732 4c-.77-1.333-2.694-1.333-3.464 0L3.34 16c-.77 1.333.192 3 1.732 3z" />
                        </svg>
                    </div>
                    <div v-else class="flex-shrink-0 w-10 h-10 rounded-full bg-blue-100 flex items-center justify-center">
                        <svg xmlns="http://www.w3.org/2000/svg" class="h-6 w-6 text-blue-600" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13 16h-1v-4h-1m1-4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
                        </svg>
                    </div>
                    
                    <div>
                        <h3 class="text-lg font-bold text-gray-900">{{ title }}</h3>
                        <p class="text-sm text-gray-500 mt-1">{{ message }}</p>
                    </div>
                </div>
                
                <div class="flex justify-end gap-3 mt-6">
                    <button @click="cancel" class="px-4 py-2 bg-white border border-gray-300 rounded-md text-gray-700 hover:bg-gray-50 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-gray-500 transition-colors font-medium text-sm">
                        {{ cancelText || 'Cancel' }}
                    </button>
                    <button @click="confirm" :class="type === 'danger' ? 'bg-red-600 hover:bg-red-700 focus:ring-red-500' : 'bg-blue-600 hover:bg-blue-700 focus:ring-blue-500'" class="px-4 py-2 text-white rounded-md shadow-sm focus:outline-none focus:ring-2 focus:ring-offset-2 transition-colors font-medium text-sm">
                        {{ confirmText || 'Confirm' }}
                    </button>
                </div>
            </div>
        </div>
    </div>
</template>
