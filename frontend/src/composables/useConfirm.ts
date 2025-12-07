import { ref } from 'vue'

const isOpen = ref(false)
const title = ref('')
const message = ref('')
const confirmText = ref('Confirm')
const cancelText = ref('Cancel')
const type = ref<'danger' | 'warning' | 'info'>('info')

let resolvePromise: (value: boolean) => void

const confirm = (
    titleText: string,
    messageText: string,
    options: {
        confirmText?: string,
        cancelText?: string,
        type?: 'danger' | 'warning' | 'info'
    } = {}
) => {
    title.value = titleText
    message.value = messageText
    confirmText.value = options.confirmText || 'Confirm'
    cancelText.value = options.cancelText || 'Cancel'
    type.value = options.type || 'info'
    isOpen.value = true

    return new Promise<boolean>((resolve) => {
        resolvePromise = resolve
    })
}

const handleConfirm = () => {
    isOpen.value = false
    if (resolvePromise) resolvePromise(true)
}

const handleCancel = () => {
    isOpen.value = false
    if (resolvePromise) resolvePromise(false)
}

export function useConfirm() {
    return {
        isOpen,
        title,
        message,
        confirmText,
        cancelText,
        type,
        confirm,
        handleConfirm,
        handleCancel
    }
}
