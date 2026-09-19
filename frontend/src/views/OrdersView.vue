<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { useRouter } from 'vue-router'
import AppIcon from '@/components/AppIcon.vue'
import {
  canViewOrders,
  manageOrders,
  money,
  ordersRequest,
  jsonOptions,
  type ListSummary,
} from '@/lib/orders'
const router = useRouter()
const lists = ref<ListSummary[]>([]),
  loading = ref(true),
  error = ref(''),
  filter = ref(''),
  status = ref('all'),
  busy = ref(false),
  name = ref('')
const dialog = ref<HTMLDialogElement>()
const visible = computed(() =>
  lists.value.filter(
    (l) =>
      (status.value === 'all' || l.status === status.value) &&
      l.name.toLowerCase().includes(filter.value.toLowerCase()),
  ),
)
async function load() {
  loading.value = true
  error.value = ''
  try {
    lists.value = (await ordersRequest<{ lists: ListSummary[] }>('/lists')).lists
  } catch (e) {
    error.value = e instanceof Error ? e.message : 'Could not load lists.'
  } finally {
    loading.value = false
  }
}
async function create() {
  busy.value = true
  error.value = ''
  try {
    const result = await ordersRequest<{ list: { id: string } }>(
      '/lists',
      jsonOptions('POST', { name: name.value }),
    )
    dialog.value?.close()
    await router.push(`/orders/${result.list.id}`)
  } catch (e) {
    error.value = e instanceof Error ? e.message : 'Could not create list.'
  } finally {
    busy.value = false
  }
}
function openNewList() {
  name.value = ''
  error.value = ''
  dialog.value?.showModal()
}
onMounted(() => {
  if (canViewOrders.value) void load()
  else loading.value = false
})
</script>
<template>
  <section class="orders-page">
    <div class="page-heading">
      <div>
        <p class="eyebrow">PARTS & PURCHASING</p>
        <h1>Order lists</h1>
        <p class="muted">Collect parts, review requests and track orders by shop.</p>
      </div>
      <div v-if="canViewOrders" class="inline-actions">
        <RouterLink class="button secondary" to="/orders/standard-parts">Standard parts</RouterLink
        ><button v-if="manageOrders" class="button primary" @click="openNewList">
          <AppIcon name="plus" :size="16" />New list
        </button>
      </div>
    </div>
    <p v-if="!canViewOrders" class="info-panel">An order role is needed to view this section.</p>
    <template v-else>
      <div class="order-toolbar">
        <label class="search-field"
          ><AppIcon name="search" /><input
            v-model="filter"
            aria-label="Search order lists"
            placeholder="Search lists…" /></label
        ><select v-model="status" class="small-select" aria-label="List status">
          <option value="all">All lists</option>
          <option value="open">Open</option>
          <option value="closed">Closed</option></select
        ><button class="button secondary" :disabled="loading" @click="load">Refresh</button>
      </div>
      <p v-if="error" class="error" role="alert">{{ error }}</p>
      <p v-if="loading" class="hint" role="status">Loading order lists…</p>
      <div v-else class="order-list-grid">
        <RouterLink
          v-for="list in visible"
          :key="list.id"
          :to="`/orders/${list.id}`"
          class="order-list-card"
          ><div class="section-title">
            <span class="badge" :class="{ 'order-open': list.status === 'open' }">{{
              list.status === 'open' ? 'Open for parts' : 'Closed · ordering'
            }}</span
            ><AppIcon name="arrow" :size="18" />
          </div>
          <h2>{{ list.name }}</h2>
          <strong class="order-card-total">{{ money(list.totalCents) }}</strong>
          <p class="muted">
            {{ list.partCount }} part lines<span v-if="list.pendingCount">
              · {{ list.pendingCount }} pending requests</span
            >
          </p>
          <template v-if="list.status === 'closed'"
            ><progress
              :value="list.orderedCount"
              :max="list.partCount || 1"
              :aria-label="`${list.orderedCount} of ${list.partCount} parts ordered`"
            />
            <p class="field-help">
              {{ list.orderedCount }} / {{ list.partCount }} ordered
            </p></template
          ></RouterLink
        >
      </div>
      <p v-if="!loading && !visible.length" class="info-panel">
        {{
          lists.length
            ? 'No matching lists.'
            : 'No order lists yet. An order admin can create the first one.'
        }}
      </p>
    </template>
    <dialog
      ref="dialog"
      class="image-dialog"
      aria-label="New order list"
      @cancel="
        (e) => {
          if (busy) e.preventDefault()
        }
      "
    >
      <form @submit.prevent="create">
        <h2>New order list</h2>
        <label class="form-field"
          >List name<input
            v-model="name"
            aria-label="List name"
            required
            maxlength="200"
            :disabled="busy"
            autofocus
        /></label>
        <p class="hint">
          Prices use EUR. You can add categories for teams, projects or shops after creating the
          list.
        </p>
        <p v-if="error" class="error" role="alert">{{ error }}</p>
        <div class="dialog-actions">
          <button type="button" class="button secondary" :disabled="busy" @click="dialog?.close()">
            Cancel</button
          ><button class="button primary" :disabled="busy">
            {{ busy ? 'Creating…' : 'Create list' }}
          </button>
        </div>
      </form>
    </dialog>
  </section>
</template>
