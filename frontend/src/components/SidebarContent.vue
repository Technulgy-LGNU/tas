<script setup lang="ts">
import { computed, ref } from 'vue'
import { currentUser, logout } from '@/lib/auth'
import AppIcon from './AppIcon.vue'
import { canViewOrders } from '@/lib/orders'
withDefaults(defineProps<{ collapsed?: boolean; mobile?: boolean }>(), {
  collapsed: false,
  mobile: false,
})
const emit = defineEmits<{ navigate: []; toggle: []; close: [] }>()
const name = computed(
  () => currentUser.value?.email || currentUser.value?.username || 'Team member',
)
const signingOut = ref(false)
const error = ref('')
async function signOut() {
  signingOut.value = true
  try {
    await logout()
  } catch {
    error.value = 'Could not sign out. Try again.'
    signingOut.value = false
  }
}
</script>

<template>
  <div class="sidebar-content" :class="{ collapsed }">
    <div class="sidebar-brand">
      <RouterLink to="/" class="brand-link" aria-label="TAS home" @click="emit('navigate')">
        <span class="brand-mark">t.</span
        ><span v-if="!collapsed" class="brand-word">T.A.S.<small>TECHNULGY</small></span>
      </RouterLink>
      <button
        v-if="mobile"
        class="icon-button mobile-close"
        aria-label="Close navigation"
        @click="emit('close')"
      >
        <AppIcon name="close" />
      </button>
    </div>
    <p v-if="!collapsed" class="nav-label">WORKSPACE</p>
    <nav aria-label="Main navigation">
      <RouterLink
        to="/"
        class="nav-link"
        :title="collapsed ? 'Overview' : undefined"
        @click="emit('navigate')"
        ><AppIcon name="home" /><span :class="{ 'sr-only': collapsed }">Overview</span></RouterLink
      >
      <RouterLink
        to="/images"
        class="nav-link"
        :title="collapsed ? 'Images' : undefined"
        @click="emit('navigate')"
        ><AppIcon name="image" /><span :class="{ 'sr-only': collapsed }">Images</span></RouterLink
      >
      <RouterLink
        to="/website"
        class="nav-link"
        :title="collapsed ? 'Website' : undefined"
        @click="emit('navigate')"
        ><AppIcon name="website" /><span :class="{ 'sr-only': collapsed }"
          >Website</span
        ></RouterLink
      >
      <RouterLink
        v-if="canViewOrders"
        to="/orders"
        class="nav-link"
        :title="collapsed ? 'Orders' : undefined"
        @click="emit('navigate')"
        ><AppIcon name="orders" /><span :class="{ 'sr-only': collapsed }">Orders</span></RouterLink
      >
    </nav>
    <div class="sidebar-footer">
      <p v-if="error" class="error" role="alert">{{ error }}</p>
      <div class="sidebar-user" :title="name">
        <span class="avatar">{{ name.charAt(0).toUpperCase() }}</span>
        <div v-if="!collapsed" class="user-details">
          <strong>{{ name }}</strong
          ><small>{{ currentUser?.roles?.join(', ') || 'Team member' }}</small>
        </div>
      </div>
      <p v-if="currentUser?.localDevelopment && !collapsed" class="hint">
        Local development · admin
      </p>
      <button
        v-if="!currentUser?.localDevelopment"
        class="nav-link sign-out"
        :disabled="signingOut"
        :title="collapsed ? 'Sign out' : undefined"
        @click="signOut"
      >
        <AppIcon name="logout" /><span :class="{ 'sr-only': collapsed }">{{
          signingOut ? 'Signing out…' : 'Sign out'
        }}</span>
      </button>
      <button
        v-if="!mobile"
        class="nav-link sidebar-toggle"
        :aria-label="collapsed ? 'Expand sidebar' : 'Collapse sidebar'"
        :aria-expanded="!collapsed"
        @click="emit('toggle')"
      >
        <AppIcon :name="collapsed ? 'expand' : 'collapse'" /><span v-if="!collapsed"
          >Collapse sidebar</span
        >
      </button>
    </div>
  </div>
</template>
