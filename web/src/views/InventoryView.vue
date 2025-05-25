<script setup lang="ts">
import Header from '@/components/Header.vue'
import PopUp from '@/components/PopUp.vue'
import { ref, onMounted } from 'vue'
import axios from 'axios'
import Cookies from 'js-cookie'
import InventoryCategoryComponent from '@/components/InventoryCategoryComponent.vue'
import router from '@/router'


interface Entry {
  id: number
  name: string
  quantity: number
  link: string
  location: string
}

interface Inventory {
  id: number
  name: string
  entries: Entry[]
}

const inventories = ref<Inventory[]>([])

async function fetchInventories() {
  try {
    await axios
      .get('/api/inventories', {
        headers: {
          'Content-Type': 'application/json',
          'Accept': 'application/json',
          'Authorization': `Bearer ${Cookies.get('token')}`
        }
      })
    .then((res) => {
      inventories.value = res.data
    })
  } catch (error: any) {
    if (error.response.status === 401) {
      popUp.value?.show('Unauthorized')
      Cookies.remove('token')
      await router.push({ name: 'login' })
    }
    popUp.value?.show("Error fetching inventories.")
    console.log(error)
  }
}

const newCategory = ref<string>('')

const createCategory = async () => {
  try {
    await axios
      .post(`/api/inventory/category/create`, { name: newCategory.value }, {
        headers: {
          'Content-Type': 'application/json',
          'Accept': 'application/json',
          'Authorization': `Bearer ${Cookies.get('token')}`
        }
      })
    .then((res) => {
      inventories.value = res.data
      newCategory.value = ''
    })
  } catch (error: any) {
    if (error.response.status === 401) {
      popUp.value?.show('Unauthorized')
      Cookies.remove('token')
      await router.push({ name: 'login' })
    }
    popUp.value?.show("Error creating category.")
    console.log(error)
  }
}

const showError = (msg: string) => {
  popUp.value?.show(msg)
}

const popUp = ref<InstanceType<typeof PopUp> | null>(null)

onMounted(async () => {
  await fetchInventories()
})
</script>

<template>
  <div class="flex flex-col min-h-screen">
    <Header page="Inventory" />
    <div class="max-w-4xl mx-auto p-6 space-y-2">
      <div class="max-w-7xl mx-auto">
        <h2 class="text-2xl font-semibold mb-4">Inventory Management</h2>
        <input
          v-if="Cookies.get('admin') === 'true' || Number(Cookies.get('inventory')) >= 3"
          v-model="newCategory"
          class="border p-1 rounded"
          placeholder="New Category"
        />
        <button
          v-if="Cookies.get('admin') === 'true' || Number(Cookies.get('inventory')) >= 3"
          @click="createCategory"
          class="bg-gray-800 hover:bg-gray-700 text-white px-4 py-2 rounded">
          Create
        </button>
      </div>
    </div>
    <div v-for="inv in inventories" :key="inv.id" >
      <InventoryCategoryComponent
        :inv="inv"
        @error="showError"
        @update="fetchInventories"
      />
    </div>

    <PopUp ref="popUp" />
  </div>
</template>

<style scoped>
</style>
