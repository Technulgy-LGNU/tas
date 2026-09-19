<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import OrderPartDialog from '@/components/orders/OrderPartDialog.vue'
import {
  canViewOrders,
  manageOrders,
  money,
  ordersRequest,
  jsonOptions,
  type StandardPart,
  type PartFields,
} from '@/lib/orders'
const parts = ref<StandardPart[]>([]),
  filter = ref(''),
  loading = ref(true),
  busy = ref(false),
  error = ref(''),
  formError = ref('')
const edit = ref<StandardPart | null>(null),
  dialog = ref<InstanceType<typeof OrderPartDialog>>()
const visible = computed(() =>
  parts.value.filter((p) =>
    `${p.name} ${p.shop}`.toLowerCase().includes(filter.value.toLowerCase()),
  ),
)
async function load() {
  loading.value = true
  error.value = ''
  try {
    parts.value = (await ordersRequest<{ parts: StandardPart[] }>('/standard-parts')).parts
  } catch (e) {
    error.value = e instanceof Error ? e.message : 'Could not load standard parts.'
  } finally {
    loading.value = false
  }
}
function open(p: StandardPart | null) {
  edit.value = p
  formError.value = ''
  dialog.value?.open(p ? { ...p, categoryId: '' } : undefined)
}
async function save(part: PartFields) {
  busy.value = true
  formError.value = ''
  try {
    await ordersRequest(
      `/standard-parts${edit.value ? '/' + edit.value.id : ''}`,
      jsonOptions(edit.value ? 'PUT' : 'POST', { ...part, version: edit.value?.version }),
    )
    dialog.value?.close()
    await load()
  } catch (e) {
    formError.value = e instanceof Error ? e.message : 'Could not save part.'
  } finally {
    busy.value = false
  }
}
async function remove(p: StandardPart) {
  if (!window.confirm(`Delete standard part “${p.name}”? Existing order lists keep their copies.`))
    return
  busy.value = true
  error.value = ''
  try {
    await ordersRequest(`/standard-parts/${p.id}`, jsonOptions('DELETE', { version: p.version }))
    await load()
  } catch (e) {
    error.value = e instanceof Error ? e.message : 'Could not delete part.'
  } finally {
    busy.value = false
  }
}
onMounted(() => {
  if (canViewOrders.value) void load()
  else loading.value = false
})
</script>
<template>
  <section class="orders-page">
    <RouterLink class="order-back" to="/orders">← Order lists</RouterLink>
    <div class="page-heading">
      <div>
        <p class="eyebrow">PARTS LIBRARY</p>
        <h1>Standard parts</h1>
        <p class="muted">
          Reusable part details for additions and requests. Prices are copied when used.
        </p>
      </div>
      <button v-if="manageOrders" class="button primary" :disabled="busy" @click="open(null)">
        New standard part
      </button>
    </div>
    <p v-if="!canViewOrders" class="info-panel">An order role is needed to view this section.</p>
    <template v-else
      ><div class="order-toolbar">
        <input
          v-model="filter"
          class="entry-filter"
          aria-label="Search standard parts"
          placeholder="Find a part or shop…"
        /><button class="button secondary" :disabled="busy || loading" @click="load">
          Refresh
        </button>
      </div>
      <p v-if="error" class="error" role="alert">{{ error }}</p>
      <p v-if="loading" role="status">Loading standard parts…</p>
      <div v-else class="order-table-wrap">
        <table class="order-table">
          <thead>
            <tr>
              <th>Part</th>
              <th>Shop</th>
              <th>Default amount</th>
              <th>Unit price</th>
              <th v-if="manageOrders">Actions</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="p in visible" :key="p.id">
              <td>
                <strong>{{ p.name }}</strong
                ><a v-if="p.link" :href="p.link" target="_blank" rel="noopener noreferrer"
                  >Product link ↗</a
                >
              </td>
              <td>{{ p.shop }}</td>
              <td>{{ p.amount }}</td>
              <td>{{ money(p.unitPriceCents) }}</td>
              <td v-if="manageOrders">
                <div class="inline-actions">
                  <button class="button secondary" :disabled="busy" @click="open(p)">Edit</button
                  ><button class="button secondary" :disabled="busy" @click="remove(p)">
                    Delete
                  </button>
                </div>
              </td>
            </tr>
          </tbody>
        </table>
      </div>
      <p v-if="!loading && !visible.length" class="info-panel">
        No standard parts found.
      </p></template
    >
    <OrderPartDialog
      ref="dialog"
      :title="edit ? 'Edit standard part' : 'New standard part'"
      :busy="busy"
      :error="formError"
      @save="save"
    />
  </section>
</template>
