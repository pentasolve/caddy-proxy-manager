<script setup lang="ts">
import { RouterView, useRoute } from 'vue-router'
import Layout from './components/Layout.vue'
import { Toaster } from 'vue-sonner'
import ConfirmModal from './components/ConfirmModal.vue'
import { useConfirm } from './composables/useConfirm'

const route = useRoute()
const { isOpen, title, message, confirmText, cancelText, type, handleConfirm, handleCancel } = useConfirm()
</script>

<template>
  <div class="min-h-screen bg-gray-100">
    <Toaster position="top-center" richColors closeButton />
    <ConfirmModal 
        :isOpen="isOpen"
        :title="title"
        :message="message"
        :confirmText="confirmText"
        :cancelText="cancelText"
        :type="type"
        @confirm="handleConfirm"
        @cancel="handleCancel"
    />
    <Layout v-if="route.name !== 'login' && route.name !== 'setup'">
      <RouterView />
    </Layout>
    <RouterView v-else />
  </div>
</template>
