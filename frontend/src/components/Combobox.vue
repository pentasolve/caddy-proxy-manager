<script setup lang="ts">
import { ref, computed, onMounted, onUnmounted, nextTick } from 'vue'

const props = withDefaults(defineProps<{
  modelValue: string
  options: string[]
  placeholder?: string
  disabled?: boolean
  color?: 'pink' | 'green' | 'purple' | 'blue' | 'orange'
}>(), {
  color: 'pink'
})

const emit = defineEmits(['update:modelValue', 'blur'])

const isOpen = ref(false)
const containerRef = ref<HTMLElement | null>(null)
const inputRef = ref<HTMLInputElement | null>(null)
const dropdownStyle = ref<Record<string, string>>({})

const filteredOptions = computed(() => {
  const query = (props.modelValue || '').toLowerCase()
  return props.options.filter(opt => opt.toLowerCase().includes(query))
})

const colorClasses = computed(() => {
  const map = {
    pink: {
      focus: 'focus:ring-pink-500/50 focus:border-pink-500',
      hover: 'hover:bg-pink-50'
    },
    green: {
      focus: 'focus:ring-green-500/50 focus:border-green-500',
      hover: 'hover:bg-green-50'
    },
    purple: {
      focus: 'focus:ring-purple-500/50 focus:border-purple-500',
      hover: 'hover:bg-purple-50'
    },
    blue: {
      focus: 'focus:ring-blue-500/50 focus:border-blue-500',
      hover: 'hover:bg-blue-50'
    },
    orange: {
      focus: 'focus:ring-orange-500/50 focus:border-orange-500',
      hover: 'hover:bg-orange-50'
    }
  }
  return map[props.color] || map.pink
})

const updatePosition = () => {
  if (containerRef.value) {
    const rect = containerRef.value.getBoundingClientRect()
    dropdownStyle.value = {
      position: 'absolute',
      top: `${rect.bottom + window.scrollY}px`,
      left: `${rect.left + window.scrollX}px`,
      width: `${rect.width}px`,
      zIndex: '9999'
    }
  }
}

const onInput = async (e: Event) => {
  const value = (e.target as HTMLInputElement).value
  emit('update:modelValue', value)
  isOpen.value = true
  await nextTick()
  updatePosition()
}

const select = (value: string) => {
  emit('update:modelValue', value)
  isOpen.value = false
}

const close = (e: MouseEvent) => {
  if (containerRef.value && !containerRef.value.contains(e.target as Node)) {
    const dropdownEl = document.getElementById('combobox-dropdown-' + uniqueId)
    if (dropdownEl && dropdownEl.contains(e.target as Node)) {
        return
    }
    isOpen.value = false
  }
}

const onFocus = async () => {
    if (!props.disabled) {
        isOpen.value = true
        await nextTick()
        updatePosition()
    }
}

const onBlur = () => {
    emit('blur')
}

const uniqueId = Math.random().toString(36).substr(2, 9)

const handleResize = () => {
    if (isOpen.value) updatePosition()
}

const handleScroll = (e: Event) => {
    if (isOpen.value) {
      const dropdownEl = document.getElementById('combobox-dropdown-' + uniqueId)
      if (dropdownEl && dropdownEl.contains(e.target as Node)) {
        return
      }
      isOpen.value = false
    }
}

const getScrollableParents = (element: HTMLElement | null): HTMLElement[] => {
  const parents: HTMLElement[] = []
  let current = element?.parentElement
  while (current) {
    const style = getComputedStyle(current)
    if (style.overflow === 'auto' || style.overflow === 'scroll' || 
        style.overflowY === 'auto' || style.overflowY === 'scroll') {
      parents.push(current)
    }
    current = current.parentElement
  }
  return parents
}

let scrollableParents: HTMLElement[] = []

onMounted(() => {
  document.addEventListener('click', close)
  window.addEventListener('resize', handleResize)
  window.addEventListener('scroll', handleScroll, true)
  scrollableParents = getScrollableParents(containerRef.value)
  scrollableParents.forEach(parent => {
    parent.addEventListener('scroll', handleScroll)
  })
})

onUnmounted(() => {
  document.removeEventListener('click', close)
  window.removeEventListener('resize', handleResize)
  window.removeEventListener('scroll', handleScroll, true)
  scrollableParents.forEach(parent => {
    parent.removeEventListener('scroll', handleScroll)
  })
})
</script>

<template>
  <div class="relative" ref="containerRef">
    <input
      ref="inputRef"
      type="text"
      :value="modelValue"
      @input="onInput"
      @focus="onFocus"
      @blur="onBlur"
      :placeholder="placeholder"
      :disabled="disabled"
      :class="[
        'w-full border rounded-lg px-4 py-2.5 text-gray-700 bg-white focus:outline-none focus:ring-2 transition-all duration-200 shadow-sm hover:border-gray-400',
        colorClasses.focus,
        disabled ? 'bg-gray-50 text-gray-400 cursor-not-allowed' : ''
      ]"
    />

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
            v-if="isOpen && filteredOptions.length > 0" 
            :id="'combobox-dropdown-' + uniqueId"
            :style="dropdownStyle"
            class="fixed mt-1 bg-white shadow-lg max-h-60 rounded-md py-1 text-base ring-1 ring-black ring-opacity-5 overflow-auto focus:outline-none sm:text-sm"
        >
            <ul class="divide-y divide-gray-100">
            <li 
                v-for="option in filteredOptions" 
                :key="option"
                @click="select(option)"
                class="cursor-pointer select-none relative py-2.5 pl-4 pr-4 transition-colors duration-150 text-gray-900"
                :class="colorClasses.hover"
            >
                <span class="block truncate">
                {{ option }}
                </span>
            </li>
            </ul>
        </div>
        </transition>
    </Teleport>
  </div>
</template>
