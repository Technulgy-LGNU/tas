<script setup lang="ts">
import { ref, watch } from 'vue'
import axios from 'axios'
import Cookies from 'js-cookie'

const props = defineProps({
  visible: {
    type: Boolean,
    required: true,
  },
  teamId: {
    type: Number,
    required: true,
  },
})

const emit = defineEmits(['close', 'editedTeam'])

interface Team {
  Id: number
  Name: string
  Email: string
  League: string
  Members: TeamMembers[]
}

interface TeamMembers {
  Id: number
  Name: string
}

const team = ref<Team>({} as Team)

const fetchTeam = async () => {
  try {
    await axios
      .get(`/api/getTeam/${props.teamId}`, {
        headers: {
          'Authorization': `Bearer ${Cookies.get('token')}`,
          'Content-Type': 'application/json',
          'Accept': 'application/json',
        },
      })
      .then((response) => {
        team.value = response.data
        console.log(team.value)
      })
  } catch (error) {
    emit('close')
    console.error('Error fetching team:', error)
  }
}

const saveTeam = async () => {
  try {
    await axios
      .put(`/api/updateTeam/${props.teamId}`, {
        "name": team.value.Name,
        "email": team.value.Email,
        "league": team.value.League,
      }, {
        headers: {
          'Authorization': `Bearer ${Cookies.get('token')}`,
          'Content-Type': 'application/json',
          'Accept': 'application/json',
        },
      })
      .then(() => {
        emit('editedTeam')
        emit('close')
      })
  } catch (error) {
    emit('close')
    console.error('Error updating team:', error)
  }
}

watch(
  () => props.visible,
  () => {
    if (props.teamId != -1) {
      fetchTeam()
    }
  },
)
</script>

<template>
  <div v-if="visible" class="fixed inset-0 bg-opacity-50 flex items-center justify-center">
    <div class="bg-white p-6 rounded-2xl shadow-lg w-full max-w-2xl max-h-[90vh] overflow-y-auto">
      <h3 class="text-lg font-semibold mb-4">Edit Team</h3>


      <!-- Form to edit team -->
      <div>
        <form @submit.prevent="saveTeam">
          <div class="grid grid-cols-2 gap-4">
            <!-- Left Column: Team Information -->
            <div class="w-1/2 pr-4">
              <div class="mb-4">
                <label for="name" class="block">Name:</label>
                <input v-model="team.Name" id="name" type="text" class="input-field border-2 border-b-black rounded-sm" required />
              </div>
              <div class="mb-4">
                <label for="email" class="block">Email:</label>
                <input v-model="team.Email" id="email" type="email" class="input-field border-2 border-b-black rounded-sm" required />
              </div>
              <div class="mb-4">
                <label for="league" class="block">League:</label>
                <select v-model="team.League" id="league" class="input-field border-2 border-b-black rounded-sm">
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

            <!-- Right Column: Members -->
            <div class="w-1/2 pl-4">
              <label class="block mb-2">Members: </label>
              <div class="permissions-grid">
                <div v-for="member in team.Members" :key="member.Id" class="flex items-center mb-2">
                  <span>{{ member.Name }}</span>
                </div>
              </div>
            </div>
          </div>

          <div class="p-3 flex space-x-2 mt-4 justify-end">
            <button type="submit" class="bg-gray-800 text-white px-3 py-1 rounded hover:bg-gray-700">Save</button>
            <button type="button" @click="emit('close')" class="bg-green-600 text-white px-3 py-1 rounded hover:bg-green-700" >Cancel</button>
          </div>
        </form>
      </div>
    </div>
  </div>
</template>

<style scoped>
</style>
