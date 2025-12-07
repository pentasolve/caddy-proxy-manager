import { ref } from 'vue'

const isConnected = ref(false)
let socket: WebSocket | null = null
const listeners: Record<string, Set<Function>> = {}
let reconnectAttempts = 0
const maxReconnectAttempts = 10
let reconnectTimeout: ReturnType<typeof setTimeout> | null = null
let isConnecting = false
let intentionalClose = false

export function useWebSocket() {
    const connect = () => {
        if (isConnecting) {
            return
        }

        if (socket) {
            if (socket.readyState === WebSocket.OPEN) {
                return
            }
            if (socket.readyState === WebSocket.CONNECTING) {
                return
            }
            if (socket.readyState === WebSocket.CLOSING || socket.readyState === WebSocket.CLOSED) {
                socket = null
            }
        }

        if (reconnectTimeout) {
            clearTimeout(reconnectTimeout)
            reconnectTimeout = null
        }

        isConnecting = true
        intentionalClose = false
        const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:'
        const host = window.location.host

        const wsUrl = `${protocol}//${host}/api/ws`

        socket = new WebSocket(wsUrl)

        socket.onopen = () => {
            isConnected.value = true
            isConnecting = false
            reconnectAttempts = 0
        }

        socket.onclose = () => {
            isConnected.value = false
            isConnecting = false
            socket = null

            if (!intentionalClose && reconnectAttempts < maxReconnectAttempts) {
                const delay = Math.min(1000 * Math.pow(2, reconnectAttempts), 30000)
                reconnectTimeout = setTimeout(() => {
                    reconnectAttempts++
                    connect()
                }, delay)
            } else if (intentionalClose) {
                intentionalClose = false
            }
        }

        socket.onerror = () => {
            isConnecting = false
        }

        socket.onmessage = (event) => {
            try {
                const data = JSON.parse(event.data)
                const type = data.type
                const payload = data.payload

                if (listeners[type]) {
                    listeners[type].forEach(cb => cb(payload))
                }
            } catch (e) {
                console.error('WebSocket: Failed to parse message:', e)
            }
        }
    }

    const disconnect = () => {
        if (reconnectTimeout) {
            clearTimeout(reconnectTimeout)
            reconnectTimeout = null
        }
        intentionalClose = true
        reconnectAttempts = 0
        if (socket) {
            socket.close(1000, 'Client disconnect')
            socket = null
        }
        isConnected.value = false
        isConnecting = false
    }

    const on = (type: string, callback: Function) => {
        if (!listeners[type]) {
            listeners[type] = new Set()
        }
        listeners[type].add(callback)
    }

    const off = (type: string, callback?: Function) => {
        if (!listeners[type]) return
        if (callback) {
            listeners[type].delete(callback)
        } else {
            delete listeners[type]
        }
    }

    const send = (type: string, payload: any = {}) => {
        if (socket && socket.readyState === WebSocket.OPEN) {
            socket.send(JSON.stringify({ type, payload }))
        }
    }

    return {
        isConnected,
        connect,
        disconnect,
        on,
        off,
        send
    }
}
