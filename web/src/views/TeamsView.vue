<script setup lang="ts">
import Header from '@/components/Header.vue'
import PopUp from '@/components/PopUp.vue'
import { onMounted, ref } from 'vue'
import Cookies from 'js-cookie'
import axios from 'axios'
import TeamEditComponent from '@/components/TeamEditComponent.vue'
import TeamCreateComponent from '@/components/TeamCreateComponent.vue'

const popUp = ref<InstanceType<typeof PopUp> | null>(null);
const showCreateTeamPopup = ref<boolean>(false);
const showEditTeamPopup = ref<boolean>(false);
const loading = ref<boolean>(false);
const currentTeamId = ref<number>(-1)

interface Team {
  ID: number;
  Name: string;
  League: string;
}

const teams = ref<Team[]>([])

async function fetchTeams() {
  loading.value = true;
  try {
    await axios
      .get('/api/getTeams', {
        headers: {
          'Authorization': `Bearer ${Cookies.get('token')}`,
          'Content-Type': 'application/json',
          'Accept': 'application/json',
        },
      })
      .then((response) => {
        teams.value = response.data
        loading.value = false
      })
  } catch (error) {
    popUp.value?.show('Error fetching teams')
    console.error('Error fetching teams:', error)
    loading.value = false
  }

}
const editTeam = (id: number) => {
  if (id === 0) {
    popUp.value?.show('Cannot edit this team')
    return
  }
  currentTeamId.value = id
  showEditTeamPopup.value = true

}
const deleteTeam = async (id: number) => {
  if (id === 0) {
    popUp.value?.show('Cannot delete this team')
    return
  }
  try {
    await axios
      .delete(`/api/deleteTeam/${id}`, {
        headers: {
          'Authorization': `Bearer ${Cookies.get('token')}`,
          'Content-Type': 'application/json',
          'Accept': 'application/json',
        },
      })
      .then(() => {
        popUp.value?.show('Team deleted successfully')
        fetchTeams()
      })
  } catch (error) {
    popUp.value?.show('Error deleting team')
    console.error('Error deleting team:', error)
  }
}

onMounted(fetchTeams())

</script>

<template>
  <div class="flex flex-col min-h-screen">
    <Header page="Teams" />
    <div class="max-w-4xl mx-auto p-6 space-y-2">
      <div class="max-w-7xl mx-auto">
        <h2 class="text-2xl font-semibold mb-4">Teams</h2>
        <button
          @click="showCreateTeamPopup = true" class="bg-gray-800 text-white px-4 py-2 rounded-lg hover:bg-gray-700">
          Create Team
        </button>
      </div>

      <!-- Loading State -->
      <p v-if="loading" class="text-gray-500">Loading...</p>

      <!-- Teams Table -->
      <div v-else class="bg-white shadow rounded-lg">
        <table class="w-full border-collapse">
          <thead>
          <tr class="bg-gray-200">
            <th class="p-3 text-left">Name</th>
            <th class="p-3 text-left">League</th>
            <th class="p-3">Actions</th>
          </tr>
          </thead>
          <tbody>
          <tr v-for="team in teams" :key="team.ID" class="border-t">
            <td class="p-3">{{ team.Name }}</td>
            <td class="p-3">{{ team.League }}</td>
            <td class="p-3 flex space-x-2">
              <button
                v-if="Cookies.get('admin') === 'true' || Number(Cookies.get('teams')) >= 2"
                @click="editTeam(team.ID)"
                class="bg-gray-800 text-white px-3 py-1 rounded hover:bg-gray-700"
              >
                Edit
              </button>
              <button
                v-if="Cookies.get('admin') === 'true' || Number(Cookies.get('teams')) >= 3"
                @click="deleteTeam(team.ID)"
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

    <TeamEditComponent
      :team-id="currentTeamId"
      :visible="showEditTeamPopup"
      @close="showEditTeamPopup = false"
      @editedTeam="fetchTeams(); showEditTeamPopup = false"
    />

    <TeamCreateComponent
      :visible="showCreateTeamPopup"
      @close="showCreateTeamPopup = false"
      @createdTeam="fetchTeams(); showCreateTeamPopup = false"
    />

    <PopUp ref="popUp" />
  </div>
</template>

<style scoped>

</style>
