<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { currentUser } from '@/lib/auth'
import { canViewOrders, editOrders, money, ordersRequest, type OrderStats } from '@/lib/orders'
import AppIcon from '@/components/AppIcon.vue'
const stats = ref<OrderStats | null>(null),
  error = ref(''),
  loading = ref(false)
async function load() {
  loading.value = true
  error.value = ''
  try {
    stats.value = await ordersRequest<OrderStats>('/stats')
  } catch (e) {
    error.value = e instanceof Error ? e.message : 'Could not load order stats.'
  } finally {
    loading.value = false
  }
}
onMounted(() => {
  if (canViewOrders.value) void load()
})
</script>
<template>
  <section>
    <div class="page-heading">
      <div>
        <p class="eyebrow">YOUR WORKSPACE</p>
        <h1>Welcome to T.A.S.</h1>
        <p class="muted">Manage your team's website, images and orders in one place.</p>
      </div>
    </div>
    <div class="overview-links">
      <RouterLink to="/website/home" class="feature-card"
        ><span class="feature-icon"><AppIcon name="website" :size="28" /></span>
        <div>
          <h2>Website management</h2>
          <p class="muted">Edit pages, publish articles and manage team content.</p>
        </div>
        <AppIcon name="arrow" /></RouterLink
      ><RouterLink to="/images" class="feature-card"
        ><span class="feature-icon"><AppIcon name="image" :size="28" /></span>
        <div>
          <h2>Image library</h2>
          <p class="muted">Browse, upload and organize your images.</p>
        </div>
        <AppIcon name="arrow" /></RouterLink
      ><RouterLink v-if="canViewOrders" to="/orders" class="feature-card"
        ><span class="feature-icon"><AppIcon name="orders" :size="28" /></span>
        <div>
          <h2>Orders</h2>
          <p class="muted">Collect parts, review requests and track purchasing.</p>
        </div>
        <AppIcon name="arrow"
      /></RouterLink>
    </div>
    <template v-if="canViewOrders"
      ><div class="section-title">
        <h2>Orders at a glance</h2>
        <RouterLink class="button secondary" to="/orders">View order lists →</RouterLink>
      </div>
      <p v-if="loading" class="hint" role="status">Loading order stats…</p>
      <p v-if="error" class="error" role="alert">
        {{ error }} <button class="button secondary" @click="load">Retry</button>
      </p>
      <div v-if="stats" class="order-stats overview-stats">
        <div>
          <span>Open lists</span><strong>{{ stats.openLists }}</strong
          ><small>{{ money(stats.openTotalCents) }} in planned parts</small>
        </div>
        <div>
          <span>{{ editOrders ? 'Pending requests' : 'My pending requests' }}</span
          ><strong>{{ stats.pendingRequests }}</strong
          ><small>Awaiting review</small>
        </div>
        <div>
          <span>Still to order</span><strong>{{ stats.remainingParts }}</strong
          ><small>{{ money(stats.remainingTotalCents) }} across closed lists</small>
        </div>
        <div>
          <span>Closed lists</span><strong>{{ stats.closedLists }}</strong
          ><small>Grouped by shop for ordering</small>
        </div>
      </div></template
    >
    <p class="hint">
      Signed in as {{ currentUser?.email || currentUser?.username || currentUser?.id }}.
    </p>
  </section>
</template>
