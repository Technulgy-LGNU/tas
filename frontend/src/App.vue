<script setup lang="ts">
import { onMounted, onUnmounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { checkSession } from '@/lib/auth'
import AppShell from '@/components/AppShell.vue'

const route = useRoute()
const router = useRouter()
let timer: ReturnType<typeof setInterval> | undefined
let checking = false
async function recheckSession() {
  if (route.meta.public || document.hidden || checking) return
  checking = true
  try {
    if (!await checkSession()) await router.replace({ name: 'login', query: { returnTo: route.fullPath } })
  }
  catch { await router.replace({ name: 'login', query: { error: 'unavailable', returnTo: route.fullPath } }) }
  finally { checking = false }
}
onMounted(() => {
  timer = setInterval(() => { void recheckSession() }, 60_000)
  window.addEventListener('focus', recheckSession)
})
onUnmounted(() => {
  clearInterval(timer)
  window.removeEventListener('focus', recheckSession)
})
</script>

<template>
  <main v-if="route.meta.public" class="auth-main"><RouterView /></main>
  <AppShell v-else><RouterView /></AppShell>
</template>
