<script setup lang="ts">
import { currentUser } from '@/lib/auth'
import AppIcon from '@/components/AppIcon.vue'
</script>

<template>
  <section>
    <div class="page-heading">
      <div>
        <p class="eyebrow">YOUR WORKSPACE</p>
        <h1>Welcome to T.A.S.</h1>
        <p class="muted">Manage your team's content in one place.</p>
      </div>
    </div>
    <RouterLink to="/images" class="feature-card"
      ><span class="feature-icon"><AppIcon name="image" :size="28" /></span>
      <div>
        <h2>Image library</h2>
        <p class="muted">Browse, upload and organize your images.</p>
      </div>
      <AppIcon name="arrow"
    /></RouterLink>
    <p class="hint">
      Signed in as {{ currentUser?.email || currentUser?.username || currentUser?.id }}.
    </p>
  </section>
</template>
