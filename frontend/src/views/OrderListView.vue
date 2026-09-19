<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { currentUser } from '@/lib/auth'
import AppIcon from '@/components/AppIcon.vue'
import OrderPartDialog from '@/components/orders/OrderPartDialog.vue'
import {
  canViewOrders,
  manageOrders,
  editOrders,
  money,
  ordersRequest,
  jsonOptions,
  partFields,
  type OrderList,
  type OrderPart,
  type PartRequest,
  type StandardPart,
  type PartFields,
  type OrderAction,
} from '@/lib/orders'
const route = useRoute(),
  router = useRouter()
const list = ref<OrderList | null>(null),
  standards = ref<StandardPart[]>([]),
  loading = ref(true),
  busy = ref(false),
  error = ref(''),
  notice = ref(''),
  formError = ref('')
const filter = ref(''),
  categoryFilter = ref('all'),
  groupBy = ref('category'),
  tab = ref('parts'),
  showHistory = ref(false)
const dialog = ref<InstanceType<typeof OrderPartDialog>>(),
  mode = ref('add_part'),
  target = ref('')
const canEdit = computed(
  () => editOrders.value && (list.value?.status === 'open' || manageOrders.value),
)
const canRequest = computed(() => canViewOrders.value && list.value?.status === 'open')
const pending = computed(
  () => list.value?.content.requests.filter((r) => r.status === 'pending') ?? [],
)
const requests = computed(
  () =>
    list.value?.content.requests.filter((r) => showHistory.value || r.status === 'pending') ?? [],
)
const total = computed(
  () => list.value?.content.parts.reduce((sum, p) => sum + p.amount * p.unitPriceCents, 0) ?? 0,
)
const ordered = computed(() => list.value?.content.parts.filter((p) => p.orderedAt).length ?? 0)
const remaining = computed(
  () =>
    list.value?.content.parts.reduce(
      (sum, p) => sum + (p.orderedAt ? 0 : p.amount * p.unitPriceCents),
      0,
    ) ?? 0,
)
const title = computed(
  () =>
    ({
      add_part: 'Add part',
      edit_part: 'Edit part',
      request_part: 'Request a part',
      edit_request: 'Edit request',
      accept_request: 'Review and accept request',
    })[mode.value] || 'Part',
)
function categoryName(id: string) {
  return list.value?.content.categories.find((c) => c.id === id)?.name || 'Uncategorized'
}
const groups = computed(() => {
  const groups = new Map<string, { key: string; name: string; parts: OrderPart[]; total: number }>()
  const shopMode = list.value?.status === 'closed' || groupBy.value === 'shop'
  for (const p of list.value?.content.parts ?? []) {
    if (categoryFilter.value !== 'all' && p.categoryId !== categoryFilter.value) continue
    if (
      !`${p.name} ${p.shop} ${categoryName(p.categoryId)}`
        .toLowerCase()
        .includes(filter.value.toLowerCase())
    )
      continue
    const name = shopMode ? p.shop : categoryName(p.categoryId),
      key = shopMode ? p.shop.toLowerCase() : p.categoryId
    if (!groups.has(key)) groups.set(key, { key, name, parts: [], total: 0 })
    const group = groups.get(key)!
    group.parts.push(p)
    group.total += p.amount * p.unitPriceCents
  }
  return [...groups.values()]
    .sort((a, b) => a.name.localeCompare(b.name))
    .map((g) => ({ ...g, parts: g.parts.sort((a, b) => a.name.localeCompare(b.name)) }))
})
let generation = 0
async function load() {
  const id = String(route.params.id),
    token = ++generation
  loading.value = true
  error.value = ''
  notice.value = ''
  list.value = null
  dialog.value?.close()
  if (!canViewOrders.value) {
    loading.value = false
    return
  }
  try {
    const [data, library] = await Promise.all([
      ordersRequest<{ list: OrderList }>(`/lists/${id}`),
      ordersRequest<{ parts: StandardPart[] }>('/standard-parts'),
    ])
    if (token !== generation) return
    list.value = data.list
    standards.value = library.parts
  } catch (e) {
    if (token === generation) error.value = e instanceof Error ? e.message : 'Could not load list.'
  } finally {
    if (token === generation) loading.value = false
  }
}
async function perform(command: OrderAction): Promise<boolean> {
  if (!list.value || busy.value) return false
  const id = list.value.id
  busy.value = true
  error.value = ''
  notice.value = ''
  try {
    const data = await ordersRequest<{ list: OrderList }>(
      `/lists/${id}/actions`,
      jsonOptions('POST', { ...command, version: list.value.version }),
    )
    if (list.value?.id === id) {
      list.value = data.list
      notice.value = 'Changes saved.'
    }
    return true
  } catch (e) {
    error.value = e instanceof Error ? e.message : 'Could not save changes.'
    return false
  } finally {
    busy.value = false
  }
}
function openPart(action: string, part?: OrderPart | PartRequest) {
  mode.value = action
  target.value = part?.id ?? ''
  formError.value = ''
  dialog.value?.open(part ? partFields(part) : undefined)
}
async function savePart(part: PartFields, note: string) {
  if (await perform({ action: mode.value, targetId: target.value, part, note }))
    dialog.value?.close()
  else formError.value = error.value
}
async function rename() {
  const name = window.prompt('List name', list.value?.name)
  if (name !== null) await perform({ action: 'rename', name })
}
async function category(action: string, id = '', current = '') {
  if (action === 'delete_category') {
    if (
      window.confirm(
        `Delete category “${current}”? Categories still used by parts or pending requests cannot be deleted.`,
      )
    )
      await perform({ action, targetId: id })
    return
  }
  const name = window.prompt(
    action === 'add_category' ? 'New category name (team, project or shop)' : 'Category name',
    current,
  )
  if (name !== null) await perform({ action, targetId: id, name })
}
async function changeStatus() {
  const close = list.value?.status === 'open'
  if (
    window.confirm(
      close
        ? 'Close this list? Only order admins and admins will be able to change it. Parts will be grouped by shop for ordering.'
        : 'Reopen this list for additions and requests? Existing ordered marks are kept.',
    )
  )
    await perform({ action: close ? 'close' : 'reopen' })
}
async function removeList() {
  if (
    !list.value ||
    !window.confirm(
      `Delete “${list.value.name}” and all its parts and requests? This cannot be undone.`,
    )
  )
    return
  busy.value = true
  error.value = ''
  try {
    await ordersRequest(
      `/lists/${list.value.id}`,
      jsonOptions('DELETE', { version: list.value.version }),
    )
    await router.push('/orders')
  } catch (e) {
    error.value = e instanceof Error ? e.message : 'Could not delete list.'
  } finally {
    busy.value = false
  }
}
async function deletePart(p: OrderPart) {
  if (window.confirm(`Remove “${p.name}” from this list?`))
    await perform({ action: 'delete_part', targetId: p.id })
}
async function reject(r: PartRequest) {
  const note = window.prompt(`Reason for rejecting “${r.name}” (optional)`, '')
  if (note !== null) await perform({ action: 'reject_request', targetId: r.id, note })
}
async function withdraw(r: PartRequest) {
  if (window.confirm(`Withdraw your request for “${r.name}”?`))
    await perform({ action: 'withdraw_request', targetId: r.id })
}
function markOrdered(event: Event, p: OrderPart) {
  const input = event.target as HTMLInputElement,
    next = input.checked
  input.checked = !!p.orderedAt
  void perform({ action: 'set_ordered', targetId: p.id, ordered: next })
}
watch(
  () => route.params.id,
  () => {
    tab.value = 'parts'
    filter.value = ''
    categoryFilter.value = 'all'
    void load()
  },
  { immediate: true },
)
</script>
<template>
  <section class="orders-page">
    <RouterLink class="order-back" to="/orders">← Order lists</RouterLink>
    <p v-if="!canViewOrders" class="info-panel">An order role is needed to view this section.</p>
    <p v-if="loading" class="hint" role="status">Loading order list…</p>
    <p v-if="error" class="form-error" role="alert">
      {{ error }}
      <button class="button secondary" :disabled="busy" @click="load">Refresh list</button>
    </p>
    <template v-if="list">
      <div class="page-heading">
        <div>
          <p class="eyebrow">
            {{ list.status === 'open' ? 'COLLECTING PARTS' : 'ORDERING BY SHOP' }}
          </p>
          <h1>{{ list.name }}</h1>
          <p class="muted">
            {{
              list.status === 'open'
                ? 'Add parts and review team requests before closing the list.'
                : 'This list is closed. Only order admins and admins can edit it or mark parts ordered.'
            }}
          </p>
        </div>
        <div class="inline-actions order-actions">
          <button class="button secondary" :disabled="busy" @click="load">Refresh</button
          ><button v-if="manageOrders" class="button secondary" :disabled="busy" @click="rename">
            Rename</button
          ><button
            v-if="manageOrders"
            class="button primary"
            :disabled="busy || (list.status === 'open' && pending.length > 0)"
            @click="changeStatus"
          >
            {{ list.status === 'open' ? 'Close list' : 'Reopen list' }}</button
          ><button
            v-if="manageOrders"
            class="icon-button danger-icon"
            aria-label="Delete order list"
            :disabled="busy"
            @click="removeList"
          >
            <AppIcon name="trash" />
          </button>
        </div>
      </div>
      <p v-if="notice" class="success-message" role="status">{{ notice }}</p>
      <div class="order-stats">
        <div>
          <span>List total</span><strong>{{ money(total) }}</strong
          ><small>{{ list.content.parts.length }} part lines · EUR</small>
        </div>
        <div>
          <span>{{ list.status === 'closed' ? 'Still to order' : 'Pending requests' }}</span
          ><strong>{{ list.status === 'closed' ? money(remaining) : pending.length }}</strong
          ><small>{{
            list.status === 'closed'
              ? `${ordered} / ${list.content.parts.length} lines ordered`
              : editOrders
                ? 'Review before closing'
                : 'Your requests awaiting review'
          }}</small>
        </div>
      </div>
      <div class="order-tabs" role="tablist" aria-label="Order list sections">
        <button
          :class="{ active: tab === 'parts' }"
          role="tab"
          :aria-selected="tab === 'parts'"
          @click="tab = 'parts'"
        >
          Parts ({{ list.content.parts.length }})</button
        ><button
          :class="{ active: tab === 'requests' }"
          role="tab"
          :aria-selected="tab === 'requests'"
          @click="tab = 'requests'"
        >
          {{ editOrders ? 'Requests' : 'My requests' }} ({{ pending.length }})</button
        ><button
          v-if="manageOrders"
          :class="{ active: tab === 'categories' }"
          role="tab"
          :aria-selected="tab === 'categories'"
          @click="tab = 'categories'"
        >
          Categories
        </button>
      </div>
      <template v-if="tab === 'parts'">
        <div class="order-toolbar">
          <label class="search-field"
            ><AppIcon name="search" /><input
              v-model="filter"
              aria-label="Search parts"
              placeholder="Search parts or shops…" /></label
          ><select v-model="categoryFilter" class="small-select" aria-label="Filter category">
            <option value="all">All categories</option>
            <option value="">Uncategorized</option>
            <option v-for="c in list.content.categories" :key="c.id" :value="c.id">
              {{ c.name }}
            </option></select
          ><select
            v-if="list.status === 'open'"
            v-model="groupBy"
            class="small-select"
            aria-label="Group parts"
          >
            <option value="category">Group by category</option>
            <option value="shop">Group by shop</option></select
          ><button
            v-if="canRequest"
            class="button secondary"
            :disabled="busy"
            @click="openPart('request_part')"
          >
            Request part</button
          ><button
            v-if="canEdit"
            class="button primary"
            :disabled="busy"
            @click="openPart('add_part')"
          >
            Add part
          </button>
        </div>
        <p v-if="canEdit && list.status === 'closed'" class="field-help">
          Editing an ordered part clears its checkmark so the changed line can be ordered again.
        </p>
        <section v-for="g in groups" :key="g.key" class="order-shop-group">
          <div class="section-title">
            <h2>{{ g.name }}</h2>
            <strong>{{ money(g.total) }}</strong>
          </div>
          <div class="order-table-wrap">
            <table class="order-table">
              <thead>
                <tr>
                  <th v-if="list.status === 'closed'">Ordered</th>
                  <th>Part</th>
                  <th>Amount</th>
                  <th>Unit price</th>
                  <th>Total</th>
                  <th>
                    {{ list.status === 'closed' || groupBy === 'shop' ? 'Category' : 'Shop' }}
                  </th>
                  <th v-if="canEdit">Actions</th>
                </tr>
              </thead>
              <tbody>
                <tr v-for="p in g.parts" :key="p.id" :class="{ 'part-ordered': p.orderedAt }">
                  <td v-if="list.status === 'closed'">
                    <input
                      v-if="manageOrders"
                      type="checkbox"
                      :aria-label="`Mark ${p.name} ordered`"
                      :checked="!!p.orderedAt"
                      :disabled="busy"
                      @change="markOrdered($event, p)"
                    /><span v-else>{{ p.orderedAt ? '✓' : '—' }}</span
                    ><small v-if="p.orderedAt">{{
                      new Date(p.orderedAt).toLocaleDateString()
                    }}</small>
                  </td>
                  <td>
                    <strong>{{ p.name }}</strong
                    ><a v-if="p.link" :href="p.link" target="_blank" rel="noopener noreferrer"
                      >Product link ↗</a
                    ><small v-if="p.requestId">From an approved request</small>
                  </td>
                  <td>{{ p.amount }}</td>
                  <td>{{ money(p.unitPriceCents) }}</td>
                  <td>
                    <strong>{{ money(p.amount * p.unitPriceCents) }}</strong>
                  </td>
                  <td>
                    {{
                      list.status === 'closed' || groupBy === 'shop'
                        ? categoryName(p.categoryId)
                        : p.shop
                    }}
                  </td>
                  <td v-if="canEdit">
                    <div class="inline-actions">
                      <button
                        class="icon-button"
                        :aria-label="`Edit ${p.name}`"
                        :disabled="busy"
                        @click="openPart('edit_part', p)"
                      >
                        <AppIcon name="edit" :size="16" /></button
                      ><button
                        class="icon-button danger-icon"
                        :aria-label="`Remove ${p.name}`"
                        :disabled="busy"
                        @click="deletePart(p)"
                      >
                        <AppIcon name="trash" :size="16" />
                      </button>
                    </div>
                  </td>
                </tr>
              </tbody>
            </table>
          </div>
        </section>
        <p v-if="!groups.length" class="info-panel">
          {{
            list.content.parts.length
              ? 'No matching parts.'
              : 'No parts yet. Add a part directly or submit a request for review.'
          }}
        </p>
      </template>
      <template v-if="tab === 'requests'">
        <div class="order-toolbar">
          <label class="cover-option"
            ><input v-model="showHistory" type="checkbox" />Include resolved requests</label
          ><button
            v-if="canRequest"
            class="button primary"
            :disabled="busy"
            @click="openPart('request_part')"
          >
            Request part
          </button>
        </div>
        <p v-if="!editOrders" class="field-help">
          You can see and change your own pending requests. Accepted parts appear in the list for
          everyone.
        </p>
        <article v-for="r in requests" :key="r.id" class="order-request-card">
          <div class="section-title">
            <div>
              <span class="badge">{{ r.status }}</span>
              <h3>{{ r.name }}</h3>
            </div>
            <strong>{{ money(r.amount * r.unitPriceCents) }}</strong>
          </div>
          <p>
            {{ r.amount }} × {{ money(r.unitPriceCents) }} · {{ r.shop }} ·
            {{ categoryName(r.categoryId) }}
          </p>
          <a v-if="r.link" :href="r.link" target="_blank" rel="noopener noreferrer"
            >Product link ↗</a
          >
          <p class="field-help">
            Requested by {{ r.createdByName }} · {{ new Date(r.createdAt).toLocaleString() }}
          </p>
          <p v-if="r.reviewNote" class="hint">Review note: {{ r.reviewNote }}</p>
          <div v-if="r.status === 'pending' && list.status === 'open'" class="inline-actions">
            <template v-if="editOrders"
              ><button
                class="button primary"
                :disabled="busy"
                @click="openPart('accept_request', r)"
              >
                Review & accept</button
              ><button class="button secondary" :disabled="busy" @click="reject(r)">
                Reject
              </button></template
            ><template v-if="r.createdBy === currentUser?.id"
              ><button
                class="button secondary"
                :disabled="busy"
                @click="openPart('edit_request', r)"
              >
                Edit request</button
              ><button class="button secondary" :disabled="busy" @click="withdraw(r)">
                Withdraw
              </button></template
            >
          </div>
        </article>
        <p v-if="!requests.length" class="info-panel">
          {{ showHistory ? 'No requests yet.' : 'No pending requests.' }}
        </p>
      </template>
      <template v-if="tab === 'categories' && manageOrders"
        ><div class="section-title">
          <div>
            <h2>List categories</h2>
            <p class="field-help">
              Use categories for teams, projects or shops. Each part can have one category.
            </p>
          </div>
          <button class="button primary" :disabled="busy" @click="category('add_category')">
            Add category
          </button>
        </div>
        <div v-for="c in list.content.categories" :key="c.id" class="order-category-row">
          <strong>{{ c.name }}</strong>
          <div class="inline-actions">
            <button
              class="button secondary"
              :disabled="busy"
              @click="category('rename_category', c.id, c.name)"
            >
              Rename</button
            ><button
              class="icon-button danger-icon"
              :aria-label="`Delete category ${c.name}`"
              :disabled="busy"
              @click="category('delete_category', c.id, c.name)"
            >
              <AppIcon name="trash" :size="16" />
            </button>
          </div>
        </div>
        <p v-if="!list.content.categories.length" class="info-panel">
          No categories yet. Parts can also remain uncategorized.
        </p></template
      >
    </template>
    <OrderPartDialog
      ref="dialog"
      :title="title"
      :submit-label="
        mode === 'accept_request'
          ? 'Accept into list'
          : mode === 'request_part'
            ? 'Submit request'
            : 'Save part'
      "
      :categories="list?.content.categories ?? []"
      :standards="mode === 'add_part' || mode === 'request_part' ? standards : []"
      :busy="busy"
      :error="formError"
      :review="mode === 'accept_request'"
      @save="savePart"
    />
  </section>
</template>
