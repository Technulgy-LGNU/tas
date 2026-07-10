<script setup lang="ts">
import { computed, onMounted, reactive, ref } from 'vue'
import {
  Box,
  Check,
  ChevronDown,
  ChevronRight,
  ClipboardList,
  Edit3,
  Globe2,
  LogIn,
  LogOut,
  PackagePlus,
  Plus,
  RefreshCw,
  Save,
  Send,
  Upload,
  Users,
  X
} from 'lucide-vue-next'
import {
  clearToken,
  completeLogin,
  fetchAuthConfig,
  getJSON,
  getToken,
  patchJSON,
  postJSON,
  putJSON,
  startLogin,
  uploadFile
} from './api'
import type {
  AuthConfig,
  Competition,
  HomeArticle,
  InventoryCategory,
  InventoryItem,
  InventoryMatch,
  MeResponse,
  Member,
  OrderList,
  OrderListItem,
  OrderRequest,
  Prize,
  Role,
  Sponsor,
  SponsorCategory,
  Team,
  UploadedImage
} from './types'

type TabKey = 'inventory' | 'orders' | 'website' | 'members'
type CategoryNode = InventoryCategory & { children: CategoryNode[]; depth: number }

const authConfig = ref<AuthConfig | null>(null)
const me = ref<MeResponse | null>(null)
const activeTab = ref<TabKey>('inventory')
const loading = ref(true)
const error = ref('')
const notice = ref('')

const roles = ref<Role[]>([])
const members = ref<Member[]>([])
const categories = ref<InventoryCategory[]>([])
const inventory = ref<InventoryItem[]>([])
const requests = ref<OrderRequest[]>([])
const lists = ref<OrderList[]>([])
const selectedListId = ref<number | null>(null)
const selectedList = ref<OrderList | null>(null)
const matches = ref<Record<number, InventoryMatch[]>>({})
const editingInventoryId = ref<number | null>(null)
const editingRequestId = ref<number | null>(null)
const editingOrderItemId = ref<number | null>(null)
const expandedCategoryIds = ref<Set<number>>(new Set())
const selectedCategoryFilter = ref<number | null>(null)
const images = ref<UploadedImage[]>([])
const teams = ref<Team[]>([])
const competitions = ref<Competition[]>([])
const prizes = ref<Prize[]>([])
const sponsorCategories = ref<SponsorCategory[]>([])
const sponsors = ref<Sponsor[]>([])
const homeArticles = ref<HomeArticle[]>([])

const categoryForm = reactive({ name: '', parent_id: null as number | null, sort_order: 0 })
const itemForm = reactive({
  name: '',
  quantity: 0,
  vendor_id: '',
  product_url: '',
  website: '',
  notes: '',
  category_id: null as number | null,
  confirmed: true
})
const requestForm = reactive({ name: '', quantity: 1, unit_price_cents: 0, url: '', notes: '', shop_name: '' })
const listForm = reactive({ name: '' })
const listItemForm = reactive({ name: '', quantity: 1, unit_price_cents: 0, url: '', notes: '', shop_name: '' })
const inventoryEditForm = reactive({
  name: '',
  quantity: 0,
  vendor_id: '',
  product_url: '',
  website: '',
  notes: '',
  category_id: null as number | null,
  confirmed: true
})
const requestEditForm = reactive({ name: '', quantity: 1, unit_price_cents: 0, url: '', notes: '', shop_name: '' })
const orderItemEditForm = reactive({ name: '', quantity: 1, unit_price_cents: 0, url: '', notes: '', shop_name: '' })
const imageFile = ref<File | null>(null)
const teamForm = reactive({ name: '', slug: '', summary: '', image_id: null as number | null, published: false, sort_order: 0 })
const competitionForm = reactive({
  name: '',
  slug: '',
  location: '',
  starts_on: '',
  ends_on: '',
  published: false,
  sort_order: 0
})
const prizeForm = reactive({ title: '', placement: '', notes: '', team_id: null as number | null, competition_id: null as number | null, sort_order: 0 })
const sponsorCategoryForm = reactive({ name: '', sort_order: 0 })
const sponsorForm = reactive({ name: '', category_id: null as number | null, logo_image_id: null as number | null, website_url: '', published: false, sort_order: 0 })
const homeForms = reactive<Record<number, { title: string; body: string; image_id: number | null; published: boolean }>>({
  1: { title: '', body: '', image_id: null, published: false },
  2: { title: '', body: '', image_id: null, published: false },
  3: { title: '', body: '', image_id: null, published: false }
})

const approved = computed(() => me.value?.member.status === 'approved')
const loggedIn = computed(() => Boolean(me.value))
const selectedPermissions = computed(() => new Set(me.value?.permissions ?? []))
const visibleTabs = computed(() =>
  [
    { key: 'inventory' as const, label: 'Inventory', icon: Box, permission: 'inventory:view' },
    { key: 'orders' as const, label: 'Orders', icon: ClipboardList, permission: 'orders:view' },
    { key: 'website' as const, label: 'Website', icon: Globe2, permission: 'website:view' },
    { key: 'members' as const, label: 'Members', icon: Users, permission: 'members:view' }
  ].filter((tab) => has(tab.permission) || authConfig.value?.dev_allow)
)
const pendingRequestsByShop = computed(() => {
  const grouped = new Map<string, OrderRequest[]>()
  for (const request of requests.value.filter((item) => item.status === 'pending')) {
    const key = request.shop_name || request.shop_domain || 'unknown shop'
    grouped.set(key, [...(grouped.get(key) ?? []), request])
  }
  return Array.from(grouped.entries()).map(([shop, items]) => ({ shop, items }))
})
const categoryTree = computed(() => buildCategoryTree(categories.value))
const visibleCategoryNodes = computed(() => flattenVisibleCategories(categoryTree.value, expandedCategoryIds.value))
const categoryOptions = computed(() => flattenAllCategories(categoryTree.value))
const filteredInventory = computed(() => {
  if (!selectedCategoryFilter.value) {
    return inventory.value
  }
  const categoryIDs = new Set<number>([selectedCategoryFilter.value, ...descendantCategoryIds(selectedCategoryFilter.value, categoryTree.value)])
  return inventory.value.filter((item) => item.category_id !== null && item.category_id !== undefined && categoryIDs.has(item.category_id))
})

onMounted(async () => {
  await run(async () => {
    authConfig.value = await fetchAuthConfig()
    if (authConfig.value.issuer && authConfig.value.client_id) {
      await completeLogin(authConfig.value)
    }
    if (getToken() || authConfig.value.dev_allow) {
      await loadMe()
    }
  }, true)
  loading.value = false
})

function has(permission: string) {
  return selectedPermissions.value.has(permission)
}

function canAny(...permissions: string[]) {
  return permissions.some((permission) => has(permission)) || Boolean(authConfig.value?.dev_allow)
}

async function run(task: () => Promise<void>, quiet = false) {
  error.value = ''
  if (!quiet) {
    notice.value = ''
  }
  try {
    await task()
  } catch (err) {
    error.value = err instanceof Error ? err.message : String(err)
  }
}

async function login() {
  if (!authConfig.value) {
    return
  }
  await startLogin(authConfig.value)
}

function logout() {
  clearToken()
  me.value = null
}

async function loadMe() {
  me.value = await getJSON<MeResponse>('/me')
  if (approved.value) {
    await loadDashboard()
  }
}

async function loadDashboard() {
  const jobs: Promise<unknown>[] = []
  if (canAny('inventory:view')) jobs.push(loadInventory())
  if (canAny('orders:view')) jobs.push(loadOrders())
  if (canAny('website:view')) jobs.push(loadWebsite())
  if (canAny('members:view', 'members:manage')) jobs.push(loadMembers())
  await Promise.all(jobs)
  if (visibleTabs.value.length && !visibleTabs.value.some((tab) => tab.key === activeTab.value)) {
    activeTab.value = visibleTabs.value[0].key
  }
}

async function refreshAll() {
  await run(async () => {
    await loadMe()
    notice.value = 'Data refreshed'
  })
}

async function loadMembers() {
  ;[roles.value, members.value] = await Promise.all([getJSON<Role[]>('/roles'), getJSON<Member[]>('/members')])
}

async function saveMember(member: Member) {
  await run(async () => {
    const roleIds = (member.roles ?? []).map((role) => role.id)
    await patchJSON<Member>(`/members/${member.id}`, { status: member.status, role_ids: roleIds })
    await loadMembers()
    notice.value = 'Member saved'
  })
}

function toggleMemberRole(member: Member, role: Role, checked: boolean) {
  const current = member.roles ?? []
  member.roles = checked ? [...current, role] : current.filter((item) => item.id !== role.id)
}

async function loadInventory() {
  const [loadedCategories, loadedInventory] = await Promise.all([
    getJSON<InventoryCategory[]>('/inventory/categories'),
    getJSON<InventoryItem[]>('/inventory/items')
  ])
  categories.value = loadedCategories
  inventory.value = loadedInventory
  if (expandedCategoryIds.value.size === 0) {
    expandedCategoryIds.value = new Set(loadedCategories.filter((category) => !category.parent_id).map((category) => category.id))
  }
}

async function createCategory() {
  await run(async () => {
    await postJSON<InventoryCategory>('/inventory/categories', clean(categoryForm))
    Object.assign(categoryForm, { name: '', parent_id: null, sort_order: 0 })
    await loadInventory()
    notice.value = 'Category created'
  })
}

async function createInventoryItem() {
  await run(async () => {
    await postJSON<InventoryItem>('/inventory/items', clean(itemForm))
    Object.assign(itemForm, { name: '', quantity: 0, vendor_id: '', product_url: '', website: '', notes: '', category_id: null, confirmed: true })
    await loadInventory()
    notice.value = 'Inventory item created'
  })
}

function editInventoryItem(item: InventoryItem) {
  editingInventoryId.value = item.id
  Object.assign(inventoryEditForm, {
    name: item.name,
    quantity: item.quantity,
    vendor_id: item.vendor_id,
    product_url: item.product_url,
    website: item.website,
    notes: item.notes,
    category_id: item.category_id ?? null,
    confirmed: item.confirmed
  })
}

function cancelInventoryEdit() {
  editingInventoryId.value = null
}

async function saveInventoryItem(item: InventoryItem) {
  await run(async () => {
    await patchJSON<InventoryItem>(`/inventory/items/${item.id}`, clean(inventoryEditForm))
    editingInventoryId.value = null
    await loadInventory()
    notice.value = 'Inventory item saved'
  })
}

function toggleCategory(category: InventoryCategory) {
  const next = new Set(expandedCategoryIds.value)
  if (next.has(category.id)) {
    next.delete(category.id)
  } else {
    next.add(category.id)
  }
  expandedCategoryIds.value = next
}

function selectCategoryFilter(categoryID: number | null) {
  selectedCategoryFilter.value = categoryID
}

async function reorder(item: InventoryItem) {
  await run(async () => {
    await postJSON<OrderRequest>(`/inventory/items/${item.id}/reorder`, { quantity: 1, unit_price_cents: 0, notes: `Reorder ${item.name}` })
    await loadOrders()
    activeTab.value = 'orders'
    notice.value = 'Reorder request created'
  })
}

async function loadOrders() {
  ;[requests.value, lists.value] = await Promise.all([getJSON<OrderRequest[]>('/orders/requests'), getJSON<OrderList[]>('/orders/lists')])
  if (!selectedListId.value && lists.value.length) {
    selectedListId.value = lists.value[0].id
  }
  if (selectedListId.value) {
    await loadSelectedList()
  }
}

async function createRequest() {
  await run(async () => {
    await postJSON<OrderRequest>('/orders/requests', clean(requestForm))
    Object.assign(requestForm, { name: '', quantity: 1, unit_price_cents: 0, url: '', notes: '', shop_name: '' })
    await loadOrders()
    notice.value = 'Request created'
  })
}

async function createList() {
  await run(async () => {
    const list = await postJSON<OrderList>('/orders/lists', clean(listForm))
    selectedListId.value = list.id
    listForm.name = ''
    await loadOrders()
    notice.value = 'Order list created'
  })
}

async function loadSelectedList() {
  if (!selectedListId.value) {
    selectedList.value = null
    return
  }
  selectedList.value = await getJSON<OrderList>(`/orders/lists/${selectedListId.value}`)
}

function openList(list: OrderList) {
  selectedListId.value = list.id
  void loadSelectedList()
}

async function approveRequest(request: OrderRequest) {
  await run(async () => {
    await postJSON<OrderListItem>(`/orders/requests/${request.id}/approve`, { order_list_id: selectedListId.value, list_name: listForm.name })
    await loadOrders()
    notice.value = 'Request added to order list'
  })
}

function editRequest(request: OrderRequest) {
  editingRequestId.value = request.id
  Object.assign(requestEditForm, {
    name: request.name,
    quantity: request.quantity,
    unit_price_cents: request.unit_price_cents,
    url: request.url,
    notes: request.notes,
    shop_name: request.shop_name
  })
}

function cancelRequestEdit() {
  editingRequestId.value = null
}

async function saveRequest(request: OrderRequest) {
  await run(async () => {
    await patchJSON<OrderRequest>(`/orders/requests/${request.id}`, clean(requestEditForm))
    editingRequestId.value = null
    await loadOrders()
    notice.value = 'Request saved'
  })
}

async function addListItem() {
  if (!selectedListId.value) return
  await run(async () => {
    await postJSON<OrderListItem>(`/orders/lists/${selectedListId.value}/items`, clean(listItemForm))
    Object.assign(listItemForm, { name: '', quantity: 1, unit_price_cents: 0, url: '', notes: '', shop_name: '' })
    await loadOrders()
    notice.value = 'List item added'
  })
}

function editOrderItem(item: OrderListItem) {
  editingOrderItemId.value = item.id
  Object.assign(orderItemEditForm, {
    name: item.name,
    quantity: item.quantity,
    unit_price_cents: item.unit_price_cents,
    url: item.url,
    notes: item.notes,
    shop_name: item.shop_name
  })
}

function cancelOrderItemEdit() {
  editingOrderItemId.value = null
}

async function saveOrderItem(item: OrderListItem) {
  await run(async () => {
    await patchJSON<OrderListItem>(`/orders/lists/${item.order_list_id}/items/${item.id}`, clean(orderItemEditForm))
    editingOrderItemId.value = null
    await loadSelectedList()
    await loadOrders()
    notice.value = 'Order list item saved'
  })
}

async function publishList() {
  if (!selectedListId.value) return
  await run(async () => {
    await postJSON<OrderList>(`/orders/lists/${selectedListId.value}/publish`, {})
    await loadOrders()
    notice.value = 'Order list published'
  })
}

async function markOrdered(item: OrderListItem) {
  await run(async () => {
    await postJSON<OrderListItem>(`/orders/lists/${item.order_list_id}/items/${item.id}/ordered`, {})
    await loadSelectedList()
  })
}

async function loadMatches(item: OrderListItem) {
  await run(async () => {
    matches.value[item.id] = await getJSON<InventoryMatch[]>(`/orders/lists/${item.order_list_id}/items/${item.id}/matches`)
  })
}

async function receiveItem(item: OrderListItem) {
  const firstMatch = matches.value[item.id]?.[0]
  await run(async () => {
    await postJSON(`/orders/lists/${item.order_list_id}/items/${item.id}/receive`, {
      matched_inventory_item_id: firstMatch?.item.id ?? null,
      category_id: firstMatch?.item.category_id ?? null,
      vendor_id: firstMatch?.item.vendor_id ?? '',
      notes: 'Received from order list'
    })
    await Promise.all([loadSelectedList(), loadInventory()])
    notice.value = 'Item received into unconfirmed inventory'
  })
}

async function loadWebsite() {
  ;[images.value, teams.value, competitions.value, prizes.value, sponsorCategories.value, sponsors.value, homeArticles.value] = await Promise.all([
    getJSON<UploadedImage[]>('/images'),
    getJSON<Team[]>('/website/teams'),
    getJSON<Competition[]>('/website/competitions'),
    getJSON<Prize[]>('/website/prizes'),
    getJSON<SponsorCategory[]>('/website/sponsor-categories'),
    getJSON<Sponsor[]>('/website/sponsors'),
    getJSON<HomeArticle[]>('/website/home')
  ])
  for (const article of homeArticles.value) {
    if (article.slot >= 1 && article.slot <= 3) {
      homeForms[article.slot] = {
        title: article.title,
        body: article.body,
        image_id: article.image_id ?? null,
        published: article.published
      }
    }
  }
}

async function uploadImage() {
  if (!imageFile.value) return
  await run(async () => {
    await uploadFile<UploadedImage>('/images', imageFile.value as File)
    imageFile.value = null
    await loadWebsite()
    notice.value = 'Image uploaded'
  })
}

async function createTeam() {
  await run(async () => {
    await postJSON<Team>('/website/teams', clean(teamForm))
    Object.assign(teamForm, { name: '', slug: '', summary: '', image_id: null, published: false, sort_order: 0 })
    await loadWebsite()
    notice.value = 'Team created'
  })
}

async function createCompetition() {
  await run(async () => {
    await postJSON<Competition>('/website/competitions', clean({ ...competitionForm, starts_on: nullableText(competitionForm.starts_on), ends_on: nullableText(competitionForm.ends_on) }))
    Object.assign(competitionForm, { name: '', slug: '', location: '', starts_on: '', ends_on: '', published: false, sort_order: 0 })
    await loadWebsite()
    notice.value = 'Competition created'
  })
}

async function createPrize() {
  await run(async () => {
    await postJSON<Prize>('/website/prizes', clean(prizeForm))
    Object.assign(prizeForm, { title: '', placement: '', notes: '', team_id: null, competition_id: null, sort_order: 0 })
    await loadWebsite()
    notice.value = 'Prize created'
  })
}

async function createSponsorCategory() {
  await run(async () => {
    await postJSON<SponsorCategory>('/website/sponsor-categories', clean(sponsorCategoryForm))
    Object.assign(sponsorCategoryForm, { name: '', sort_order: 0 })
    await loadWebsite()
    notice.value = 'Sponsor category created'
  })
}

async function createSponsor() {
  await run(async () => {
    await postJSON<Sponsor>('/website/sponsors', clean(sponsorForm))
    Object.assign(sponsorForm, { name: '', category_id: null, logo_image_id: null, website_url: '', published: false, sort_order: 0 })
    await loadWebsite()
    notice.value = 'Sponsor created'
  })
}

async function saveHome(slot: number) {
  await run(async () => {
    await putJSON<HomeArticle>(`/website/home/${slot}`, clean(homeForms[slot]))
    await loadWebsite()
    notice.value = 'Homepage article saved'
  })
}

function onFileChange(event: Event) {
  const input = event.target as HTMLInputElement
  imageFile.value = input.files?.[0] ?? null
}

function formatCents(value: number) {
  return new Intl.NumberFormat('de-DE', { style: 'currency', currency: 'EUR' }).format(value / 100)
}

function clean<T extends Record<string, unknown>>(value: T): T {
  return Object.fromEntries(Object.entries(value).map(([key, entry]) => [key, entry === '' ? null : entry])) as T
}

function nullableText(value: string) {
  const trimmed = value.trim()
  return trimmed === '' ? null : trimmed
}

function buildCategoryTree(source: InventoryCategory[]) {
  const nodes = new Map<number, CategoryNode>()
  for (const category of source) {
    nodes.set(category.id, { ...category, children: [], depth: 0 })
  }
  const roots: CategoryNode[] = []
  for (const node of nodes.values()) {
    if (node.parent_id && nodes.has(node.parent_id)) {
      const parent = nodes.get(node.parent_id)
      if (parent) {
        node.depth = parent.depth + 1
        parent.children.push(node)
      }
    } else {
      roots.push(node)
    }
  }
  const sortNodes = (items: CategoryNode[]) => {
    items.sort((a, b) => a.sort_order - b.sort_order || a.name.localeCompare(b.name))
    for (const item of items) {
      for (const child of item.children) {
        child.depth = item.depth + 1
      }
      sortNodes(item.children)
    }
  }
  sortNodes(roots)
  return roots
}

function flattenVisibleCategories(nodes: CategoryNode[], expanded: Set<number>) {
  const visible: CategoryNode[] = []
  const visit = (node: CategoryNode) => {
    visible.push(node)
    if (expanded.has(node.id)) {
      node.children.forEach(visit)
    }
  }
  nodes.forEach(visit)
  return visible
}

function flattenAllCategories(nodes: CategoryNode[]) {
  const all: CategoryNode[] = []
  const visit = (node: CategoryNode) => {
    all.push(node)
    node.children.forEach(visit)
  }
  nodes.forEach(visit)
  return all
}

function descendantCategoryIds(categoryID: number, nodes: CategoryNode[]) {
  const ids: number[] = []
  const visit = (node: CategoryNode) => {
    if (node.id === categoryID) {
      collect(node)
      return true
    }
    return node.children.some(visit)
  }
  const collect = (node: CategoryNode) => {
    for (const child of node.children) {
      ids.push(child.id)
      collect(child)
    }
  }
  nodes.some(visit)
  return ids
}

function categoryHasChildren(category: CategoryNode) {
  return category.children.length > 0
}

function categoryInventoryCount(category: CategoryNode) {
  const ids = new Set<number>([category.id, ...descendantCategoryIds(category.id, categoryTree.value)])
  return inventory.value.filter((item) => item.category_id !== null && item.category_id !== undefined && ids.has(item.category_id)).length
}

function categoryIndentClass(depth: number) {
  const classes = ['pl-2', 'pl-6', 'pl-10', 'pl-14', 'pl-16', 'pl-20']
  return classes[Math.min(depth, classes.length - 1)]
}
</script>

<template>
  <div class="app-shell">
    <aside class="sidebar">
      <div class="brand">
        <strong>TAS</strong>
        <span>Technulgy Admin Software</span>
      </div>
      <nav class="nav">
        <button
          v-for="tab in visibleTabs"
          :key="tab.key"
          :class="{ active: activeTab === tab.key }"
          type="button"
          @click="activeTab = tab.key"
        >
          <component :is="tab.icon" :size="17" />
          {{ tab.label }}
        </button>
      </nav>
    </aside>

    <main class="main">
      <header class="topbar">
        <div>
          <h1>{{ visibleTabs.find((tab) => tab.key === activeTab)?.label ?? 'Dashboard' }}</h1>
          <p v-if="me">{{ me.member.name || me.member.email }} · {{ me.member.status }}</p>
          <p v-else>Inventory, orders, website content, and member approval</p>
        </div>
        <div class="actions">
          <button v-if="loggedIn" type="button" @click="refreshAll"><RefreshCw :size="16" /> Refresh</button>
          <button v-if="loggedIn" type="button" @click="logout"><LogOut :size="16" /> Logout</button>
          <button v-else class="primary" type="button" :disabled="!authConfig?.issuer && !authConfig?.dev_allow" @click="login">
            <LogIn :size="16" /> Login
          </button>
        </div>
      </header>

      <section class="content">
        <div v-if="loading" class="panel">Loading TAS...</div>
        <div v-if="error" class="error">{{ error }}</div>
        <div v-if="notice" class="success">{{ notice }}</div>

        <div v-if="!loading && !loggedIn" class="panel">
          <div class="panel-header">
            <div>
              <h2>Sign in</h2>
              <p>Use FusionAuth to access the management platform.</p>
            </div>
            <button class="primary" type="button" :disabled="!authConfig?.issuer" @click="login"><LogIn :size="16" /> Login</button>
          </div>
        </div>

        <div v-else-if="loggedIn && !approved" class="panel">
          <div class="panel-header">
            <div>
              <h2>Approval required</h2>
              <p>Your account exists in TAS and needs approval before app features are available.</p>
            </div>
            <span class="badge warn">{{ me?.member.status }}</span>
          </div>
        </div>

        <template v-else-if="approved">
          <section v-if="activeTab === 'inventory'" class="split">
            <div class="panel">
              <div class="panel-header">
                <div>
                  <h2>Inventory</h2>
                  <p>Confirmed and received-unconfirmed stock.</p>
                </div>
              </div>
              <div class="table-wrap">
                <table class="min-w-[920px]">
                  <colgroup>
                    <col class="w-[34%]" />
                    <col class="w-20" />
                    <col class="w-[18%]" />
                    <col class="w-[14%]" />
                    <col class="w-28" />
                    <col class="w-52" />
                  </colgroup>
                  <thead>
                    <tr>
                      <th>Name</th>
                      <th>Qty</th>
                      <th>Category</th>
                      <th>Vendor</th>
                      <th>Status</th>
                      <th></th>
                    </tr>
                  </thead>
                  <tbody>
                    <tr v-for="item in filteredInventory" :key="item.id">
                      <template v-if="editingInventoryId === item.id">
                        <td class="edit-cell">
                          <label>Name<input v-model="inventoryEditForm.name" /></label>
                          <label>Product URL<input v-model="inventoryEditForm.product_url" /></label>
                        </td>
                        <td><label>Qty<input v-model.number="inventoryEditForm.quantity" type="number" min="0" /></label></td>
                        <td>
                          <label>
                            Category
                            <select v-model="inventoryEditForm.category_id">
                              <option :value="null">None</option>
                              <option v-for="category in categoryOptions" :key="category.id" :value="category.id">
                                {{ `${'· '.repeat(category.depth)}${category.name}` }}
                              </option>
                            </select>
                          </label>
                        </td>
                        <td><label>Vendor<input v-model="inventoryEditForm.vendor_id" /></label></td>
                        <td>
                          <label><span><input v-model="inventoryEditForm.confirmed" type="checkbox" /> Confirmed</span></label>
                          <label>Website<input v-model="inventoryEditForm.website" /></label>
                        </td>
                        <td class="cell-actions">
                          <div class="actions">
                            <button class="primary" type="button" @click="saveInventoryItem(item)"><Save :size="16" /> Save</button>
                            <button type="button" @click="cancelInventoryEdit"><X :size="16" /> Cancel</button>
                          </div>
                        </td>
                      </template>
                      <template v-else>
                        <td>
                          <strong class="cell-title">{{ item.name }}</strong>
                          <span class="cell-subtle">{{ item.product_url || item.website }}</span>
                        </td>
                        <td>{{ item.quantity }}</td>
                        <td>{{ item.category?.name ?? '-' }}</td>
                        <td>{{ item.vendor_id || '-' }}</td>
                        <td><span class="badge" :class="{ good: item.confirmed, warn: !item.confirmed }">{{ item.confirmed ? 'confirmed' : 'unconfirmed' }}</span></td>
                        <td class="cell-actions">
                          <div class="actions">
                            <button type="button" :disabled="!canAny('inventory:edit', 'inventory:manage')" @click="editInventoryItem(item)"><Edit3 :size="16" /> Edit</button>
                            <button type="button" :disabled="!canAny('orders:request')" @click="reorder(item)"><PackagePlus :size="16" /> Reorder</button>
                          </div>
                        </td>
                      </template>
                    </tr>
                  </tbody>
                </table>
              </div>
            </div>

            <div class="panel">
              <div class="panel-header"><h3>Add Inventory</h3></div>
              <div class="grid two">
                <label>Name<input v-model="itemForm.name" /></label>
                <label>Quantity<input v-model.number="itemForm.quantity" type="number" min="0" /></label>
                <label>Vendor ID<input v-model="itemForm.vendor_id" /></label>
                <label>Category<select v-model="itemForm.category_id"><option :value="null">None</option><option v-for="category in categoryOptions" :key="category.id" :value="category.id">{{ `${'· '.repeat(category.depth)}${category.name}` }}</option></select></label>
                <label>Product URL<input v-model="itemForm.product_url" /></label>
                <label>Website<input v-model="itemForm.website" /></label>
              </div>
              <label>Notes<textarea v-model="itemForm.notes" /></label>
              <label><span><input v-model="itemForm.confirmed" type="checkbox" /> Confirmed stock</span></label>
              <button class="primary" type="button" :disabled="!canAny('inventory:edit', 'inventory:manage')" @click="createInventoryItem"><Plus :size="16" /> Add item</button>

              <div class="panel-header"><h3>Categories</h3></div>
              <div class="grid two">
                <label>Name<input v-model="categoryForm.name" /></label>
                <label>Parent<select v-model="categoryForm.parent_id"><option :value="null">Top level</option><option v-for="category in categoryOptions" :key="category.id" :value="category.id">{{ `${'· '.repeat(category.depth)}${category.name}` }}</option></select></label>
              </div>
              <button type="button" :disabled="!canAny('inventory:edit', 'inventory:manage')" @click="createCategory"><Plus :size="16" /> Add category</button>
              <div class="tree">
                <button class="tree-row" :class="{ active: selectedCategoryFilter === null }" type="button" @click="selectCategoryFilter(null)">
                  <span>All categories</span>
                  <span class="badge">{{ inventory.length }}</span>
                </button>
                <button
                  v-for="category in visibleCategoryNodes"
                  :key="category.id"
                  class="tree-row"
                  :class="[categoryIndentClass(category.depth), { active: selectedCategoryFilter === category.id }]"
                  type="button"
                  @click="selectCategoryFilter(category.id)"
                >
                  <span class="tree-name">
                    <span class="tree-toggle" @click.stop="categoryHasChildren(category) && toggleCategory(category)">
                      <ChevronDown v-if="categoryHasChildren(category) && expandedCategoryIds.has(category.id)" :size="15" />
                      <ChevronRight v-else-if="categoryHasChildren(category)" :size="15" />
                    </span>
                    {{ category.name }}
                  </span>
                  <span class="badge">{{ categoryInventoryCount(category) }}</span>
                </button>
              </div>
            </div>
          </section>

          <section v-if="activeTab === 'orders'" class="split">
            <div class="panel">
              <div class="panel-header">
                <div>
                  <h2>Requests by Shop</h2>
                  <p>Requests use URL domain grouping with optional manager overrides.</p>
                </div>
              </div>
              <div v-if="!pendingRequestsByShop.length" class="empty">No pending requests.</div>
              <div v-for="group in pendingRequestsByShop" :key="group.shop" class="group">
                <div class="panel-header"><h3>{{ group.shop }}</h3></div>
                <div class="table-wrap">
                  <table class="min-w-[760px]">
                    <colgroup>
                      <col class="w-[46%]" />
                      <col class="w-[20%]" />
                      <col class="w-[14%]" />
                      <col class="w-52" />
                    </colgroup>
                    <tbody>
                      <tr v-for="request in group.items" :key="request.id">
                        <template v-if="editingRequestId === request.id">
                          <td class="edit-cell">
                            <label>Name<input v-model="requestEditForm.name" /></label>
                            <label>URL<input v-model="requestEditForm.url" /></label>
                            <label>Notes<textarea v-model="requestEditForm.notes" /></label>
                          </td>
                          <td>
                            <label>Quantity<input v-model.number="requestEditForm.quantity" type="number" min="1" /></label>
                            <label>Unit price cents<input v-model.number="requestEditForm.unit_price_cents" type="number" min="0" /></label>
                          </td>
                          <td><label>Shop override<input v-model="requestEditForm.shop_name" /></label></td>
                          <td class="cell-actions">
                            <div class="actions">
                              <button class="primary" type="button" @click="saveRequest(request)"><Save :size="16" /> Save</button>
                              <button type="button" @click="cancelRequestEdit"><X :size="16" /> Cancel</button>
                            </div>
                          </td>
                        </template>
                        <template v-else>
                          <td><strong class="cell-title">{{ request.name }}</strong><span class="cell-subtle">{{ request.url }}</span></td>
                          <td>{{ request.quantity }} × {{ formatCents(request.unit_price_cents) }}</td>
                          <td>{{ formatCents(request.total_price_cents) }}</td>
                          <td class="cell-actions">
                            <div class="actions">
                              <button type="button" :disabled="!canAny('orders:edit', 'orders:manage')" @click="editRequest(request)"><Edit3 :size="16" /> Edit</button>
                              <button type="button" :disabled="!canAny('orders:manage')" @click="approveRequest(request)"><Check :size="16" /> Approve</button>
                            </div>
                          </td>
                        </template>
                      </tr>
                    </tbody>
                  </table>
                </div>
              </div>
            </div>

            <div class="panel">
              <div class="panel-header"><h3>Create Request</h3></div>
              <div class="grid two">
                <label>Name<input v-model="requestForm.name" /></label>
                <label>Quantity<input v-model.number="requestForm.quantity" type="number" min="1" /></label>
                <label>Unit price cents<input v-model.number="requestForm.unit_price_cents" type="number" min="0" /></label>
                <label>Shop override<input v-model="requestForm.shop_name" /></label>
              </div>
              <label>URL<input v-model="requestForm.url" /></label>
              <label>Notes<textarea v-model="requestForm.notes" /></label>
              <button class="primary" type="button" :disabled="!canAny('orders:request')" @click="createRequest"><Send :size="16" /> Request</button>

              <div class="panel-header"><h3>Lists</h3></div>
              <div class="grid two">
                <label>New list<input v-model="listForm.name" /></label>
              </div>
              <div class="actions">
                <button type="button" :disabled="!canAny('orders:edit', 'orders:manage')" @click="createList"><Plus :size="16" /> Create list</button>
                <button class="warning" type="button" :disabled="!selectedList || !canAny('orders:manage')" @click="publishList"><Check :size="16" /> Publish</button>
              </div>
              <div class="list-browser">
                <button
                  v-for="list in lists"
                  :key="list.id"
                  class="list-row"
                  :class="{ active: selectedListId === list.id }"
                  type="button"
                  @click="openList(list)"
                >
                  <span>
                    <strong>{{ list.name }}</strong>
                    <small>{{ list.items?.length ?? 0 }} entries</small>
                  </span>
                  <span class="badge" :class="{ good: list.status === 'published' }">{{ list.status }}</span>
                </button>
                <div v-if="!lists.length" class="empty">No order lists yet.</div>
              </div>
            </div>

            <div class="panel col-span-full">
              <div class="panel-header">
                <div>
                  <h2>{{ selectedList?.name ?? 'No list selected' }}</h2>
                  <p>{{ selectedList?.status ?? '' }}</p>
                </div>
              </div>
              <div class="grid">
                <label>Name<input v-model="listItemForm.name" /></label>
                <label>Quantity<input v-model.number="listItemForm.quantity" type="number" min="1" /></label>
                <label>Unit price cents<input v-model.number="listItemForm.unit_price_cents" type="number" min="0" /></label>
                <label>Shop override<input v-model="listItemForm.shop_name" /></label>
              </div>
              <label>URL<input v-model="listItemForm.url" /></label>
              <button type="button" :disabled="!selectedListId || !canAny('orders:edit', 'orders:manage')" @click="addListItem"><Plus :size="16" /> Add direct item</button>
              <div class="table-wrap">
                <table class="min-w-[980px]">
                  <colgroup>
                    <col class="w-[32%]" />
                    <col class="w-[18%]" />
                    <col class="w-28" />
                    <col class="w-24" />
                    <col class="w-24" />
                    <col class="w-64" />
                  </colgroup>
                  <thead><tr><th>Item</th><th>Shop</th><th>Total</th><th>Ordered</th><th>Received</th><th></th></tr></thead>
                  <tbody>
                    <tr v-for="item in selectedList?.items ?? []" :key="item.id">
                      <template v-if="editingOrderItemId === item.id">
                        <td class="edit-cell">
                          <label>Name<input v-model="orderItemEditForm.name" /></label>
                          <label>URL<input v-model="orderItemEditForm.url" /></label>
                          <label>Notes<textarea v-model="orderItemEditForm.notes" /></label>
                        </td>
                        <td><label>Shop override<input v-model="orderItemEditForm.shop_name" /></label></td>
                        <td>
                          <label>Quantity<input v-model.number="orderItemEditForm.quantity" type="number" min="1" /></label>
                          <label>Unit price cents<input v-model.number="orderItemEditForm.unit_price_cents" type="number" min="0" /></label>
                        </td>
                        <td colspan="2"><span class="badge" :class="{ good: item.ordered }">{{ item.ordered ? 'ordered' : 'not ordered' }}</span></td>
                        <td class="cell-actions">
                          <div class="actions">
                            <button class="primary" type="button" @click="saveOrderItem(item)"><Save :size="16" /> Save</button>
                            <button type="button" @click="cancelOrderItemEdit"><X :size="16" /> Cancel</button>
                          </div>
                        </td>
                      </template>
                      <template v-else>
                        <td><strong class="cell-title">{{ item.name }}</strong><span class="cell-subtle">{{ item.quantity }} × {{ formatCents(item.unit_price_cents) }}</span></td>
                        <td><span class="cell-subtle">{{ item.shop_name || item.shop_domain }}</span></td>
                        <td>{{ formatCents(item.total_price_cents) }}</td>
                        <td><span class="badge" :class="{ good: item.ordered }">{{ item.ordered ? 'yes' : 'no' }}</span></td>
                        <td><span class="badge" :class="{ good: item.received }">{{ item.received ? 'yes' : 'no' }}</span></td>
                        <td class="cell-actions">
                          <div class="actions">
                            <button
                              type="button"
                              :disabled="!canAny('orders:edit', 'orders:manage') || (selectedList?.status === 'published' && !canAny('orders:manage'))"
                              @click="editOrderItem(item)"
                            >
                              <Edit3 :size="16" /> Edit
                            </button>
                            <button type="button" :disabled="item.ordered || !canAny('orders:manage')" @click="markOrdered(item)">Ordered</button>
                            <button type="button" :disabled="!canAny('orders:manage')" @click="loadMatches(item)">Matches</button>
                            <button type="button" :disabled="item.received || !canAny('orders:manage')" @click="receiveItem(item)">Receive</button>
                            <span v-if="matches[item.id]?.length" class="cell-subtle">{{ matches[item.id][0].item.name }} · {{ matches[item.id][0].score }}</span>
                          </div>
                        </td>
                      </template>
                    </tr>
                  </tbody>
                </table>
              </div>
            </div>
          </section>

          <section v-if="activeTab === 'website'" class="content !p-0">
            <div class="panel">
              <div class="panel-header"><h2>Images</h2></div>
              <div class="actions">
                <input type="file" accept="image/*" @change="onFileChange" />
                <button class="primary" type="button" :disabled="!imageFile || !canAny('website:edit', 'website:manage')" @click="uploadImage"><Upload :size="16" /> Upload</button>
              </div>
              <div class="actions">
                <span v-for="image in images" :key="image.id" class="badge">{{ image.filename }}</span>
              </div>
            </div>

            <div class="split">
              <div class="panel">
                <div class="panel-header"><h2>Teams</h2></div>
                <div class="grid two">
                  <label>Name<input v-model="teamForm.name" /></label>
                  <label>Slug<input v-model="teamForm.slug" /></label>
                  <label>Image<select v-model="teamForm.image_id"><option :value="null">None</option><option v-for="image in images" :key="image.id" :value="image.id">{{ image.filename }}</option></select></label>
                  <label><span><input v-model="teamForm.published" type="checkbox" /> Published</span></label>
                </div>
                <label>Summary<textarea v-model="teamForm.summary" /></label>
                <button type="button" :disabled="!canAny('website:edit', 'website:manage')" @click="createTeam"><Plus :size="16" /> Add team</button>
                <div class="table-wrap">
                  <table class="min-w-[420px]">
                    <colgroup>
                      <col class="w-[70%]" />
                      <col class="w-[30%]" />
                    </colgroup>
                    <tbody>
                      <tr v-for="team in teams" :key="team.id">
                        <td><span class="cell-title">{{ team.name }}</span></td>
                        <td><span class="badge" :class="{ good: team.published }">{{ team.published ? 'published' : 'draft' }}</span></td>
                      </tr>
                    </tbody>
                  </table>
                </div>
              </div>

              <div class="panel">
                <div class="panel-header"><h2>Participation</h2></div>
                <div class="grid two">
                  <label>Competition<input v-model="competitionForm.name" /></label>
                  <label>Slug<input v-model="competitionForm.slug" /></label>
                  <label>Location<input v-model="competitionForm.location" /></label>
                  <label>Starts<input v-model="competitionForm.starts_on" type="date" /></label>
                  <label>Ends<input v-model="competitionForm.ends_on" type="date" /></label>
                  <label><span><input v-model="competitionForm.published" type="checkbox" /> Published</span></label>
                </div>
                <button type="button" :disabled="!canAny('website:edit', 'website:manage')" @click="createCompetition"><Plus :size="16" /> Add competition</button>
                <div class="grid two">
                  <label>Prize<input v-model="prizeForm.title" /></label>
                  <label>Placement<input v-model="prizeForm.placement" /></label>
                  <label>Team<select v-model="prizeForm.team_id"><option :value="null">Choose</option><option v-for="team in teams" :key="team.id" :value="team.id">{{ team.name }}</option></select></label>
                  <label>Competition<select v-model="prizeForm.competition_id"><option :value="null">Choose</option><option v-for="competition in competitions" :key="competition.id" :value="competition.id">{{ competition.name }}</option></select></label>
                </div>
                <button type="button" :disabled="!canAny('website:edit', 'website:manage')" @click="createPrize"><Plus :size="16" /> Add prize</button>
              </div>
            </div>

            <div class="split">
              <div class="panel">
                <div class="panel-header"><h2>Sponsors</h2></div>
                <div class="grid two">
                  <label>Category<input v-model="sponsorCategoryForm.name" /></label>
                  <label>Order<input v-model.number="sponsorCategoryForm.sort_order" type="number" /></label>
                </div>
                <button type="button" :disabled="!canAny('website:edit', 'website:manage')" @click="createSponsorCategory"><Plus :size="16" /> Add category</button>
                <div class="grid two">
                  <label>Name<input v-model="sponsorForm.name" /></label>
                  <label>Category<select v-model="sponsorForm.category_id"><option :value="null">Choose</option><option v-for="category in sponsorCategories" :key="category.id" :value="category.id">{{ category.name }}</option></select></label>
                  <label>Logo<select v-model="sponsorForm.logo_image_id"><option :value="null">None</option><option v-for="image in images" :key="image.id" :value="image.id">{{ image.filename }}</option></select></label>
                  <label>Website<input v-model="sponsorForm.website_url" /></label>
                  <label><span><input v-model="sponsorForm.published" type="checkbox" /> Published</span></label>
                </div>
                <button type="button" :disabled="!canAny('website:edit', 'website:manage')" @click="createSponsor"><Plus :size="16" /> Add sponsor</button>
              </div>

              <div class="panel">
                <div class="panel-header"><h2>Homepage</h2></div>
                <div v-for="slot in [1, 2, 3]" :key="slot" class="group">
                  <div class="panel-header"><h3>Article {{ slot }}</h3></div>
                  <label>Title<input v-model="homeForms[slot].title" /></label>
                  <label>Body<textarea v-model="homeForms[slot].body" /></label>
                  <label>Image<select v-model="homeForms[slot].image_id"><option :value="null">None</option><option v-for="image in images" :key="image.id" :value="image.id">{{ image.filename }}</option></select></label>
                  <label><span><input v-model="homeForms[slot].published" type="checkbox" /> Published</span></label>
                  <button type="button" :disabled="!canAny('website:edit', 'website:manage')" @click="saveHome(slot)"><Save :size="16" /> Save</button>
                </div>
              </div>
            </div>
          </section>

          <section v-if="activeTab === 'members'" class="panel">
            <div class="panel-header">
              <div>
                <h2>Members</h2>
                <p>Approve registrations and assign TAS roles.</p>
              </div>
            </div>
            <div class="table-wrap">
              <table class="min-w-[820px]">
                <colgroup>
                  <col class="w-[34%]" />
                  <col class="w-36" />
                  <col class="w-[38%]" />
                  <col class="w-32" />
                </colgroup>
                <thead><tr><th>Member</th><th>Status</th><th>Roles</th><th></th></tr></thead>
                <tbody>
                  <tr v-for="member in members" :key="member.id">
                    <td><strong class="cell-title">{{ member.name || member.email }}</strong><span class="cell-subtle">{{ member.email }}</span></td>
                    <td>
                      <select v-model="member.status">
                        <option value="pending">pending</option>
                        <option value="approved">approved</option>
                        <option value="rejected">rejected</option>
                      </select>
                    </td>
                    <td>
                      <div class="flex flex-wrap gap-1.5">
                        <label v-for="role in roles" :key="role.id" class="inline-flex h-7 items-center gap-1.5 rounded-md border border-slate-200 bg-slate-50 px-2 text-xs font-medium text-slate-700">
                          <input
                            type="checkbox"
                            :checked="member.roles?.some((item) => item.id === role.id)"
                            @change="toggleMemberRole(member, role, ($event.target as HTMLInputElement).checked)"
                          />
                          {{ role.name }}
                        </label>
                      </div>
                    </td>
                    <td><button class="primary" type="button" :disabled="!canAny('members:manage')" @click="saveMember(member)"><Save :size="16" /> Save</button></td>
                  </tr>
                </tbody>
              </table>
            </div>
          </section>
        </template>
      </section>
    </main>
  </div>
</template>
