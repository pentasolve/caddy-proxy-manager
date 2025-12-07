<script setup lang="ts">
import { ref, computed, onMounted, onUnmounted, nextTick } from 'vue'

const props = withDefaults(defineProps<{
  modelValue: any
  options: Array<{ label: string, value: any } | string>
  placeholder?: string
  disabled?: boolean
  color?: 'pink' | 'green' | 'purple' | 'blue' | 'red' | 'orange'
}>(), {
  color: 'green'
})

const emit = defineEmits(['update:modelValue'])

const isOpen = ref(false)
const containerRef = ref<HTMLElement | null>(null)
const dropdownRef = ref<HTMLElement | null>(null)
const dropdownStyle = ref<Record<string, string>>({})

const normalizedOptions = computed(() => {
  return props.options.map(opt => {
    if (typeof opt === 'string') {
      return { label: opt, value: opt }
    }
    return opt
  })
})

const colorClasses = computed(() => {
  const map = {
    pink: {
      focus: 'focus:ring-pink-500/50 focus:!border-pink-500',
      hover: 'hover:bg-pink-50',
      text: 'text-pink-900',
      bg: 'bg-pink-50',
      icon: 'text-pink-600'
    },
    green: {
      focus: 'focus:ring-green-500/50 focus:!border-green-500',
      hover: 'hover:bg-green-50',
      text: 'text-green-900',
      bg: 'bg-green-50',
      icon: 'text-green-600'
    },
    purple: {
      focus: 'focus:ring-purple-500/50 focus:!border-purple-500',
      hover: 'hover:bg-purple-50',
      text: 'text-purple-900',
      bg: 'bg-purple-50',
      icon: 'text-purple-600'
    },
    blue: {
      focus: 'focus:ring-blue-500/50 focus:!border-blue-500',
      hover: 'hover:bg-blue-50',
      text: 'text-blue-900',
      bg: 'bg-blue-50',
      icon: 'text-blue-600'
    },
    red: {
      focus: 'focus:ring-red-500/50 focus:!border-red-500',
      hover: 'hover:bg-red-50',
      text: 'text-red-900',
      bg: 'bg-red-50',
      icon: 'text-red-600'
    },
    orange: {
      focus: 'focus:ring-orange-500/50 focus:!border-orange-500',
      hover: 'hover:bg-orange-50',
      text: 'text-orange-900',
      bg: 'bg-orange-50',
      icon: 'text-orange-600'
    }
  }
  return map[props.color] || map.green
})

const selectedLabel = computed(() => {
  const found = normalizedOptions.value.find(opt => opt.value === props.modelValue)
  return found ? found.label : props.placeholder || 'Select option'
})

const updatePosition = () => {
  if (containerRef.value) {
    const rect = containerRef.value.getBoundingClientRect()
    dropdownStyle.value = {
      position: 'fixed',
      top: `${rect.bottom}px`,
      left: `${rect.left}px`,
      width: `${rect.width}px`,
      zIndex: '9999'
    }
  }
}

const toggle = async () => {
  if (!props.disabled) {
    if (!isOpen.value) {
      isOpen.value = true
      await nextTick()
      updatePosition()
    } else {
      isOpen.value = false
    }
  }
}

const select = (value: any) => {
  emit('update:modelValue', value)
  isOpen.value = false
}

const closeDropdown = (e: MouseEvent) => {
  if (!containerRef.value?.contains(e.target as Node) && 
      !dropdownRef.value?.contains(e.target as Node)) {
    isOpen.value = false
  }
}

const handleScroll = (e: Event) => {
  if (isOpen.value && dropdownRef.value && !dropdownRef.value.contains(e.target as Node)) {
    isOpen.value = false
  }
}

const handleResize = () => {
  if (isOpen.value) updatePosition()
}

onMounted(() => {
  document.addEventListener('mousedown', closeDropdown)
  document.addEventListener('scroll', handleScroll, true)
  window.addEventListener('resize', handleResize)
})

onUnmounted(() => {
  document.removeEventListener('mousedown', closeDropdown)
  document.removeEventListener('scroll', handleScroll, true)
  window.removeEventListener('resize', handleResize)
})
</script>

<template>
  <div class="relative" ref="containerRef">
    <!-- Trigger Button -->
    <button 
      type="button"
      @click="toggle"
      :class="[
        'w-full text-left border rounded-lg px-4 py-2.5 flex items-center justify-between transition-all duration-200 shadow-sm outline-none',
        colorClasses.focus,
        disabled 
          ? 'bg-gray-50 border-gray-200 text-gray-400 cursor-not-allowed' 
          : 'bg-white border-gray-300 text-gray-700 hover:border-gray-400 focus:ring-2'
      ]"
      :disabled="disabled"
    >
      <span class="block truncate">{{ selectedLabel }}</span>
      <svg 
        xmlns="http://www.w3.org/2000/svg" 
        class="h-5 w-5 text-gray-400 transition-transform duration-200"
        :class="{ 'rotate-180': isOpen }"
        fill="none" 
        viewBox="0 0 24 24" 
        stroke="currentColor"
      >
        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 9l-7 7-7-7" />
      </svg>
    </button>

    <!-- Dropdown Menu -->
    <Teleport to="body">
        <transition
        enter-active-class="transition ease-out duration-100"
        enter-from-class="transform opacity-0 scale-95"
        enter-to-class="transform opacity-100 scale-100"
        leave-active-class="transition ease-in duration-75"
        leave-from-class="transform opacity-100 scale-100"
        leave-to-class="transform opacity-0 scale-95"
        >
        <div 
            v-if="isOpen" 
            ref="dropdownRef"
            :style="dropdownStyle"
            class="fixed mt-1 bg-white shadow-lg max-h-60 rounded-md py-1 text-base ring-1 ring-black ring-opacity-5 overflow-auto focus:outline-none sm:text-sm"
        >
            <ul class="divide-y divide-gray-100">
            <li 
                v-for="option in normalizedOptions" 
                :key="option.value"
                @click="select(option.value)"
                class="cursor-pointer select-none relative py-2.5 pl-4 pr-9 transition-colors duration-150"
                :class="[
                    colorClasses.hover,
                    modelValue === option.value ? [colorClasses.text, colorClasses.bg, 'font-medium'] : 'text-gray-900'
                ]"
            >
                <span class="block truncate">
                {{ option.label }}
                </span>

                <span 
                v-if="modelValue === option.value"
                class="absolute inset-y-0 right-0 flex items-center pr-4"
                :class="colorClasses.icon"
                >
                <svg xmlns="http://www.w3.org/2000/svg" class="h-5 w-5" viewBox="0 0 20 20" fill="currentColor">
                    <path fill-rule="evenodd" d="M16.707 5.293a1 1 0 010 1.414l-8 8a1 1 0 01-1.414 0l-4-4a1 1 0 011.414-1.414L8 12.586l7.293-7.293a1 1 0 011.414 0z" clip-rule="evenodd" />
                </svg>
                </span>
            </li>
            </ul>
        </div>
        </transition>
    </Teleport>
  </div>
</template>