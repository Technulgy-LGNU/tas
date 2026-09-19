<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { checkSession, login, safeReturnTo } from '@/lib/auth'

const route = useRoute()
const router = useRouter()
const checking = ref(true)
const connectionError = ref('')
const errors: Record<string, string> = {
  invalid_state: 'This sign-in attempt expired or was already used. Please try again.',
  login_failed: 'Sign-in could not be completed. Please try again.',
  access_denied: 'Your account does not have access to this application. Contact your administrator.',
  unavailable: 'The sign-in service is unavailable. Please try again.',
}
const error = computed(() => connectionError.value || errors[String(route.query.error)] || '')

onMounted(async () => {
  try {
    if (await checkSession()) await router.replace(safeReturnTo(route.query.returnTo))
  } catch {
    connectionError.value = errors.unavailable ?? ''
  } finally {
    checking.value = false
  }
})
</script>

<template>
  <section class="login-card" aria-labelledby="login-title">
    <p class="eyebrow">Technulgy Admin Software</p>
    <h1 id="login-title">Sign in to T.A.S.</h1>
    <p>Use your authorized team account to continue.</p>
    <p v-if="error" class="error" role="alert">{{ error }}</p>
    <button :disabled="checking" @click="login(route.query.returnTo)">
      {{ checking ? 'Checking your session…' : 'Sign in with FusionAuth' }}
    </button>
    <p class="hint">Access is limited to registered team members.</p>
  </section>
</template>
