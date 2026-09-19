<script setup lang="ts">
import { ref } from 'vue'
import { currentUser, logout } from '@/lib/auth'

const error = ref('')
const signingOut = ref(false)
async function signOut() {
  signingOut.value = true
  error.value = ''
  try { await logout() }
  catch { error.value = 'Could not sign out. Please try again.'; signingOut.value = false }
}
</script>

<template>
  <section class="home-card">
    <p class="eyebrow">Technulgy Admin Software</p>
    <h1>Welcome to T.A.S.</h1>
    <p>Signed in as {{ currentUser?.email || currentUser?.username || currentUser?.id }}.</p>
    <p v-if="error" class="error" role="alert">{{ error }}</p>
    <button :disabled="signingOut" @click="signOut">{{ signingOut ? 'Signing out…' : 'Sign out' }}</button>
  </section>
</template>
