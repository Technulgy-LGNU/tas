<script setup lang="ts">
import axios from 'axios'
import { type PropType, ref } from 'vue'
import Cookies from 'js-cookie'
import router from '@/router'

interface Entry {
  id: number
  name: string
  quantity: number
  link: string
  location: string
}

const props = defineProps({
  mode: {
    type: String,
    default: 'edit',
    required: true,
  },
  categoryID: {
    type: Number,
    default: 0,
  },
  entry: {
    type: Object as PropType<Entry>,
    required: true,
  }
})

const entryToEdit = ref<Entry>(props.entry)

const emit = defineEmits(['finished', 'msg', 'close'])

const editOrCreate = async () => {
  const url = ref<string>('')
  if (props.mode === 'edit' && props.entry !== null) {
    url.value = `/api/inventory/entry/update/${props.entry?.id}`
  } else if (props.mode === 'create' && props.categoryID !== 0) {
    url.value = `/api/inventory/entry/create/${props.categoryID}`
  } else {
    emit('msg', 'Invalid category')
  }
  try {
    await axios
      .post(url.value, {
        ...entryToEdit.value,
      }, {
        headers: {
          'Content-Type': 'application/json',
          'Accept': 'application/json',
          'Authorization': `Bearer ${Cookies.get('token')}`
        }
      })
      .then(() => {
        emit('finished')
        emit('close')
        emit('msg', `Successfully ${props.mode} entry`)
      })
  } catch (error: any) {
    if (error.response.status === 401) {
      emit('msg', 'Unauthorized')
      Cookies.remove('token')
      await router.push({ name: 'login' })
    }
    console.log(error)
    emit('msg', 'Error creating entry')
  }
}
</script>

<template>
  <div class="fixed inset-0 bg-opacity-50 flex items-center justify-center">
    <div class="bg-white p-6 rounded-2xl shadow-lg w-full max-w-2xl max-h-[90vh] overflow-y-auto">
      <h3 class="text-lg font-semibold mb-4">Create User</h3>

      <!-- Form to create user -->
      <div class="mb-4">
        <form @submit.prevent="editOrCreate">
            <div class="w-1/2 pr-4">
              <div class="mb-4">
                <label for="name" class="block">Name:</label>
                <input v-model="entryToEdit.name" id="name" type="text" class="input-field border-2 border-b-black rounded-sm" required />
              </div>
              <div class="mb-4">
                <label for="link" class="block">Link:</label>
                <input v-model="entryToEdit.link" id="link" type="text" class="input-field border-2 border-b-black rounded-sm" />
              </div>
              <div class="mb-4">
                <label for="location" class="block">Location:</label>
                <input v-model="entryToEdit.location" id="location" type="text" class="input-field border-2 border-b-black rounded-sm" />
              </div>
              <div class="mb-4">
                <label for="quantity" class="block">Quantity:</label>
                <input v-model="entryToEdit.quantity" id="quantity" type="number" class="input-field border-2 border-b-black rounded-sm" required />
              </div>
            </div>

          <div class="p-3 flex space-x-2 mt-4 justify-end">
            <button v-if="props.mode === 'create'" type="submit" class="bg-gray-800 text-white px-3 py-1 rounded hover:bg-gray-700">Create</button>
            <button v-else type="submit" class="bg-gray-800 text-white px-3 py-1 rounded hover:bg-gray-700">Update</button>
            <button type="button" @click="emit('close')" class="bg-green-600 text-white px-3 py-1 rounded hover:bg-green-700" >Cancel</button>
          </div>
        </form>
      </div>
    </div>
  </div>
</template>

<style scoped>
</style>
