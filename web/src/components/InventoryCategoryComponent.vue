<script setup lang="ts">

import { type PropType, ref } from 'vue'
import axios from 'axios'
import Cookies from 'js-cookie'
import InventoryEntryCreateEditComponent from '@/components/InventoryEntryCreateEditComponent.vue'
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

const props = defineProps({
  inv: {
    type: Object as PropType<Inventory>,
    required: true
  }
})

const emits = defineEmits(['update', 'error'])

const isOpen = ref(true);

function toggleOpen() {
  isOpen.value = !isOpen.value;
}

const name = ref<string>(props.inv?.name || '')
const editCategory = async (id: number, name: string) => {
  try {
    await axios
      .post(`/api/inventory/category/update/${id}`, {name: name}, {
        headers: {
          'Content-Type': 'application/json',
          'Accept': 'application/json',
          'Authorization': `Bearer ${Cookies.get('token')}`,
        }
      })
      .then(() => {
        emits('update')
        emits('error', 'Successfully updated category.')
      })
  } catch (error: any) {
    if (error.response.status === 401) {
      emits('error', 'Unauthorized')
      Cookies.remove('token')
      await router.push({ name: 'login' })
    }
    console.error(error)
    emits('error', 'Error updating category')
  }
}

const deleteCategory = async (id: number) => {
  try {
    await axios
      .delete(`/api/inventory/category/delete/${id}`, {
        headers: {
          'Content-Type': 'application/json',
          'Accept': 'application/json',
          'Authorization': `Bearer ${Cookies.get('token')}`
        }
      })
    .then(() => {
      emits('update')
      emits('error', 'Successfully deleted category')
    })
  } catch (error: any) {
    if (error.response.status === 401) {
      emits('error', 'Unauthorized')
      Cookies.remove('token')
      await router.push({ name: 'login' })
    }
    emits('error', 'Error deleting category')
    console.error(error)
  }
}

const entryToEdit = ref<Entry>({} as Entry)
const editCreate = ref<boolean>(false)
const mode = ref<string>('edit')
const showMessage = (msg: string) => {
  emits('error', msg)
}

const deleteEntry = async (id: number) => {
  try {
    await axios
      .delete(`/api/inventory/entry/delete/${id}`, {
        headers: {
          'Content-Type': 'application/json',
          'Accept': 'application/json',
          'Authorization': `Bearer ${Cookies.get('token')}`
        }
      })
      .then(() => {
        emits('error', 'Successfully deleted entry')
        emits('update')
      })
  } catch (error: any) {
    if (error.response.status === 401) {
      emits('error', 'Unauthorized')
      Cookies.remove('token')
      await router.push({ name: 'login' })
    }
    emits('error', 'Error deleting entry')
    console.error(error)
  }
}

const updateQuantity = async (id: number, quantity: number) => {
  try {
    await axios
      .post(`/api/inventory/entry/updateAmount/${id}/${quantity}`, {}, {
        headers: {
          'Content-Type': 'application/json',
          'Accept': 'application/json',
          'Authorization': `Bearer ${Cookies.get('token')}`
        }
      })
    .then(() => {
      emits('update')
      emits('error', 'Successfully updated entry')
    })
  } catch (error: any) {
    if (error.response.status === 401) {
      emits('error', 'Unauthorized')
      Cookies.remove('token')
      await router.push({ name: 'login' })
    }
    emits('error', 'Error deleting entry')
    console.error(error)
  }
}

</script>

<template>
  <div class="w-full max-w-7xl mx-auto p-6 mb-6 bg-white rounded-lg shadow-md">
    <div class="flex justify-between items-center mb-3">
      <div class="flex items-center space-x-3">
        <!-- Editable name input -->
        <input
          v-if="Cookies.get('admin') === 'true' || Number(Cookies.get('inventory')) >= 3"
          v-model="name"
          type="text"
          class="border border-gray-300 rounded-md px-3 py-1 text-2xl font-bold text-gray-900 focus:outline-none focus:ring-2 focus:ring-blue-500"
        />
        <input
          v-else
          v-model="name"
          type="text"
          class="border border-gray-300 rounded-md px-3 py-1 text-2xl font-bold text-gray-900 focus:outline-none focus:ring-2 focus:ring-blue-500"
          readonly
        />
        <!-- Entries count badge -->
        <span
          class="inline-block bg-blue-600 text-white text-sm font-semibold px-3 py-1 rounded-full select-none"
        >{{ props.inv?.entries.length }} entries</span
        >
      </div>

      <div class="space-x-2">
        <button
          @click="toggleOpen"
          class="px-3 py-1 text-sm bg-blue-600 text-white rounded hover:bg-blue-700 transition"
        >
          {{ isOpen ? 'Collapse' : 'Expand' }}
        </button>
        <button
          v-if="Cookies.get('admin') === 'true' || Number(Cookies.get('inventory')) >= 3"
          @click="editCategory(props.inv?.id, name)"
          class="px-3 py-1 text-sm bg-gray-800 text-white rounded hover:bg-gray-700 transition"
        >
          Update
        </button>
        <button
          v-if="Cookies.get('admin') === 'true' || Number(Cookies.get('inventory')) >= 2"
          @click="editCreate = true; mode = 'create'"
          class="px-3 py-1 text-sm bg-gray-800 text-white rounded hover:bg-gray-700 transition"
        >
          Create Entry
        </button>
        <button
          v-if="Cookies.get('admin') === 'true' || Number(Cookies.get('inventory')) >= 3"
          @click="deleteCategory(props.inv?.id)"
          class="px-3 py-1 text-sm bg-red-600 text-white rounded hover:bg-red-700 transition"
        >
          Delete
        </button>
      </div>
    </div>

    <transition name="fade" appear>
      <div
        v-show="isOpen"
        class="border border-gray-300 rounded-md bg-gray-50"
      >
        <div class="bg-gray-200 px-4 py-2 font-semibold text-gray-700 border-b border-gray-300">
          Entries
        </div>
        <ul>
          <li
            v-for="entry in props.inv?.entries"
            :key="entry.id"
            class="flex flex-col p-3 border-b border-gray-200 hover:bg-gray-100"
          >
            <div class="flex justify-between items-center mb-1">
              <a
                :href="entry.link"
                target="_blank"
                rel="noopener noreferrer"
                class="text-blue-600 hover:underline font-medium"
              >{{ entry.name }}</a
              >
              <input
                v-if="Cookies.get('admin') === 'true' || Number(Cookies.get('inventory')) >= 1"
                v-model="entry.quantity"
                type="number"
                class="border border-gray-300 rounded-md px-3 py-1 text-2xl font-bold text-gray-900 focus:outline-none focus:ring-2 focus:ring-blue-500"
              />
            </div>
            <div class="flex justify-between items-center text-sm text-gray-500 mb-2">
              <span>Location: {{ entry.location }}</span>
              <div class="space-x-2">
                <button
                  v-if="Cookies.get('admin') === 'true' || Number(Cookies.get('inventory')) >= 1"
                  @click="updateQuantity(entry.id, entry.quantity)"
                  class="px-2 py-0.5 bg-green-500 text-white rounded hover:bg-green-600 transition text-xs"
                >
                  Update
                </button>
                <button
                  v-if="Cookies.get('admin') === 'true' || Number(Cookies.get('inventory')) >= 2"
                  @click="editCreate = true; entryToEdit = entry; mode = 'edit'"
                  class="px-2 py-0.5 bg-green-500 text-white rounded hover:bg-green-600 transition text-xs"
                >
                  Edit
                </button>
                <button
                  v-if="Cookies.get('admin') === 'true' || Number(Cookies.get('inventory')) >= 3"
                  @click="deleteEntry(entry.id)"
                  class="px-2 py-0.5 bg-red-600 text-white rounded hover:bg-red-700 transition text-xs"
                >
                  Delete
                </button>
              </div>
            </div>
          </li>
        </ul>
      </div>
    </transition>

    <InventoryEntryCreateEditComponent
      v-if="editCreate"
      :mode="mode"
      :entry="entryToEdit"
      :category-i-d="props.inv?.id"
      @close="editCreate = false; entryToEdit = {} as Entry"
      @finished="emits('update')"
      @msg="showMessage"
    />
  </div>
</template>

<style scoped>
</style>
