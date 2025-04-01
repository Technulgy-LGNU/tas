<script setup lang="ts">
import { ref } from 'vue'
import axios from 'axios'
import Cookies from 'js-cookie'

defineProps({
  visible: {
    type: Boolean,
    required: true,
  },
})

const emit = defineEmits(['close', 'createdTeam'])

const teamName = ref<string>('')
const teamLeague = ref<string>('')

const createTeam = async () => {
  try {
    await axios
      .post('/api/createTeam', {
        "name": teamName.value,
        "league": teamLeague.value,
      }, {
        headers: {
          'Authorization': `Bearer ${Cookies.get('token')}`,
          'Content-Type': 'application/json',
          'Accept': 'application/json',
        },
      })
      .then(() => {
        emit('createdTeam')
        teamName.value = ''
        teamLeague.value = ''
      })
  } catch (error) {
    emit('close')
    console.error('Error creating team:', error)
  }
}
</script>

<template>
  <div v-if="visible" class="fixed inset-0 bg-opacity-50 flex items-center justify-center">
    <div class="bg-white p-6 rounded-2xl shadow-lg w-full max-w-2xl max-h-[90vh] overflow-y-auto">
      <h3 class="text-lg font-semibold mb-4">Edit Team</h3>


      <!-- Form to create team -->
      <div>
        <form @submit.prevent="createTeam">
          <div class="grid grid-cols-2 gap-4">
            <div class="w-1/2 pr-4">
              <div class="mb-4">
                <label for="name" class="block">Name:</label>
                <input v-model="teamName" id="name" type="text" class="input-field" required />
              </div>
              <div class="mb-4">
                <label for="league" class="block">League:</label>
                <select v-model="teamLeague" id="league" class="input-field">
                  <option value="Soccer Entry">Soccer Entry</option>
                  <option value="Soccer LightWeight Entry">Soccer LightWeight Entry</option>
                  <option value="Soccer LightWeight int.">Soccer LightWeight int.</option>
                  <option value="Soccer Open int.">Soccer Open int.</option>
                  <option value="Rescue Line Entry">Rescue Line Entry</option>
                  <option value="Rescue Line int.">Rescue Line int.</option>
                  <option value="Rescue Maze Entry">Rescue Maze Entry</option>
                  <option value="Rescue Maze int.">Rescue Maze int.</option>
                  <option value="Onstage Entry.">Onstage Entry</option>
                  <option value="Onstage int.">Onstage int.</option>
                </select>
              </div>
            </div>
          </div>

          <div class="p-3 flex space-x-2 mt-4 justify-end">
            <button type="submit" class="bg-gray-800 text-white px-3 py-1 rounded hover:bg-gray-700">Create</button>
            <button type="button" @click="emit('close')" class="bg-green-600 text-white px-3 py-1 rounded hover:bg-green-700" >Cancel</button>
          </div>
        </form>
      </div>
    </div>
  </div>
</template>

<style scoped>

</style>
