<script setup lang="ts">
import { computed, onBeforeUnmount, onMounted, ref, watch } from 'vue'
import { useRoute } from 'vue-router'
import AppIcon from './AppIcon.vue'
import SidebarContent from './SidebarContent.vue'

const route = useRoute()
const collapsed = ref(false)
const drawer = ref<HTMLDialogElement>()
const drawerOpen = ref(false)
const title = computed(() =>
  route.path.startsWith('/orders')
    ? 'Orders'
    : route.path.startsWith('/website')
      ? 'Website'
      : route.name === 'images'
        ? 'Images'
        : route.name === 'home'
          ? 'Overview'
          : 'Page not found',
)
function toggle() {
  collapsed.value = !collapsed.value
  try {
    localStorage.setItem('tas-sidebar-collapsed', String(collapsed.value))
  } catch {
    /* Storage may be disabled. */
  }
}
function openDrawer() {
  drawer.value?.showModal()
  drawerOpen.value = true
}
function closeDrawer() {
  drawer.value?.close()
  drawerOpen.value = false
}
function resize() {
  if (window.innerWidth >= 768) closeDrawer()
}
watch(() => route.fullPath, closeDrawer)
onMounted(() => {
  try {
    collapsed.value = localStorage.getItem('tas-sidebar-collapsed') === 'true'
  } catch {
    /* Use expanded default. */
  }
  window.addEventListener('resize', resize)
})
onBeforeUnmount(() => window.removeEventListener('resize', resize))
</script>

<template>
  <div class="app-shell" :class="{ 'sidebar-collapsed': collapsed }">
    <a href="#main-content" class="skip-link" @click.prevent="($refs.main as HTMLElement)?.focus()"
      >Skip to content</a
    >
    <aside class="desktop-sidebar">
      <SidebarContent :collapsed="collapsed" @toggle="toggle" />
    </aside>
    <dialog
      id="mobile-navigation"
      ref="drawer"
      class="sidebar-drawer"
      aria-label="Navigation"
      @close="drawerOpen = false"
      @click="
        (event) => {
          if (event.target === drawer) closeDrawer()
        }
      "
    >
      <SidebarContent mobile @navigate="closeDrawer" @close="closeDrawer" />
    </dialog>
    <div class="workspace">
      <header class="workspace-header">
        <button
          class="icon-button menu-toggle"
          aria-label="Open navigation"
          aria-controls="mobile-navigation"
          :aria-expanded="drawerOpen"
          @click="openDrawer"
        >
          <AppIcon name="menu" />
        </button>
        <span class="breadcrumb"
          >Workspace <span>/</span> <strong>{{ title }}</strong></span
        >
        <span class="workspace-label">Technulgy Admin</span>
      </header>
      <main id="main-content" ref="main" class="workspace-main" tabindex="-1"><slot /></main>
    </div>
  </div>
</template>
