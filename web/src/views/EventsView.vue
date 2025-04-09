<script setup lang="ts">
import Header from '@/components/Header.vue'
import PopUp from '@/components/PopUp.vue'
import { onMounted, ref } from 'vue'
import axios from 'axios'
import Cookies from 'js-cookie'

const popUp = ref<InstanceType<typeof PopUp> | null>(null);
const loading = ref<boolean>(true);
const selectedEvent = ref<number>(-1);

interface Event {
  ID: number,
  Name: string,
  Location: string,
  StartDate: string,
  EndDate: string,
}

const events = ref<Event[]>([])

const fetchEvents = async () => {
  loading.value = true
  try {
    await axios
      .get('/api/getEvents', {
        headers: {
          'Authorization': `Bearer ${Cookies.get('token')}`,
          'Content-Type': 'application/json',
          'Accept': 'application/json',
        },
      })
      .then((response) => {
        events.value = response.data
        loading.value = false
      })
  } catch (error) {
    popUp.value?.show('Error fetching events')
    console.error('Error fetching events:', error)
  }
}

const deleteEvent = async (id: number) => {
  try {
    await axios
      .delete(`/api/deleteEvent/${id}`, {
        headers: {
          'Authorization': `Bearer ${Cookies.get('token')}`,
          'Content-Type': 'application/json',
          'Accept': 'application/json',
        },
      })
      .then(() => {
        popUp.value?.show('Event deleted successfully')
        fetchEvents()
      })
  } catch (error) {
    popUp.value?.show('Error deleting event')
    console.error('Error deleting event:', error)
  }
}

const openEditEvent = async (id: number) => {
  selectedEvent.value = id
  console.log(id)
}


onMounted(fetchEvents())
</script>

<template>
  <div class="flex flex-col min-h-screen">
    <Header page="Events" />
    <div class="max-w-4xl mx-auto p-6">
      <h2 class="text-2xl font-semibold mb-4">Forms</h2>

      <!-- Loading State -->
      <p v-if="loading" class="text-gray-500">Loading...</p>

      <!-- Forms Table -->
      <div v-else class="bg-white shadow rounded-lg">
        <table class="w-full border-collapse">
          <thead>
          <tr class="bg-gray-200">
            <th class="p-3 text-left">Name</th>
            <th class="p-3 text-left">Location</th>
            <th class="p-3 text-left">Start Date</th>
            <th class="p-3 text-left">End Date</th>
            <th class="p-3">Actions</th>
          </tr>
          </thead>
          <tbody>
          <tr v-for="event in events" :key="event.ID" class="border-t">
            <td class="p-3">{{ event.Name }}</td>
            <td class="p-3">{{ event.Location }}</td>
            <td class="p-3">{{ event.StartDate }}</td>
            <td class="p-3">{{ event.EndDate }}</td>
            <td class="p-3 flex space-x-2">
              <button
                v-if="Cookies.get('admin') === 'true' || Number(Cookies.get('events')) >= 2"
                @click="openEditEvent(event.ID)"
                class="bg-gray-800 text-white px-3 py-1 rounded hover:bg-gray-700"
              >
                Edit
              </button>
              <button
                v-if="Cookies.get('admin') === 'true' || Number(Cookies.get('events')) >= 3"
                @click="deleteEvent(event.ID)"
                class="bg-red-500 text-white px-3 py-1 rounded hover:bg-red-600"
              >
                Delete
              </button>
            </td>
          </tr>
          </tbody>
        </table>
      </div>
    </div>
    <PopUp ref="popUp" />
  </div>
</template>

<style scoped>

</style>
