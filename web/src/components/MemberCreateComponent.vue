<script setup lang="ts">
import { onMounted, ref, watch } from 'vue'
import axios from 'axios'
import Cookies from 'js-cookie'

const props = defineProps({
  visible: {
    type: Boolean,
    required: true,
  },
});

const emit = defineEmits(['close', 'createdMember']);

interface Member {
  Id: number;
  Name: string;
  Email: string;
  Gender: string;
  Birthday: string;
  TeamId: number;
  Permissions: {
    Login: boolean;
    Admin: boolean;
    Members: number;
    Teams: number;
    Events: number;
    Newsletter: number;
    Form: number;
    Website: number;
    Order: number;
    Sponsors: number;
  };
}

interface Teams {
  ID: number;
  Name: string;
}

const memberData = ref<Member>({
  Id: -1,
  Name: '',
  Email: '',
  Gender: '',
  Birthday: '20-12-2005',
  TeamId: -1,
  Permissions: {
    Login: false,
    Admin: false,
    Members: 0,
    Teams: 0,
    Events: 0,
    Newsletter: 0,
    Form: 0,
    Website: 0,
    Order: 0,
    Sponsors: 0,
  },
})
const teams = ref<Teams[]>([])

const fetchTeams = async () => {
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
      });
  } catch (error) {
    console.error('Error fetching teams:', error)
    emit('close')
  }
}

const saveMember = async () => {
  try {
    await axios
      .post(`/api/createMember`,
        {
          ...memberData.value,
        },
        {
          headers: {
            'Authorization': `Bearer ${Cookies.get('token')}`,
            'Content-Type': 'application/json',
            'Accept': 'application/json',
          },
        })
      .then((response) => {
        emit('createdMember', response.data)
        emit('close')
      })
  } catch (error) {
    console.error('Error saving member data:', error)
    emit('close')
  }
}

const closePopup = () => {
  emit('close');
}

watch(
  () => props.visible,
  async () => {
    await fetchTeams()
  },
)

onMounted(fetchTeams())

</script>

<template>
  <div v-if="visible" class="fixed inset-0 bg-opacity-50 flex items-center justify-center">
    <div class="bg-white p-6 rounded-2xl shadow-lg w-full max-w-2xl max-h-[90vh] overflow-y-auto">
      <h3 class="text-lg font-semibold mb-4">Edit Member</h3>

      <!-- Form to edit member -->
      <div class="mb-4">
        <form @submit.prevent="saveMember">
          <div class="grid grid-cols-2 gap-4">
            <!-- Left Column: Basic Information -->
            <div class="w-1/2 pr-4">
              <div class="mb-4">
                <label for="name" class="block">Name:</label>
                <input v-model="memberData.Name" id="name" type="text" class="input-field" required />
              </div>
              <div class="mb-4">
                <label for="email" class="block">Email:</label>
                <input v-model="memberData.Email" id="email" type="email" class="input-field" required />
              </div>
              <div class="mb-4">
                <label for="birthday" class="block">Birthday:</label>
                <input v-model="memberData.Birthday" id="birthday" type="date" class="input-field" />
              </div>
              <div class="mb-4">
                <label for="gender" class="block">Gender:</label>
                <select v-model="memberData.Gender" id="gender" class="input-field">
                  <option value="male">Male</option>
                  <option value="female">Female</option>
                  <option value="divers">Divers</option>
                </select>
              </div>
              <div class="mb-4">
                <label for="team" class="block">Team:</label>
                <select v-model="memberData.TeamId" id="team" class="input-field">
                  <option v-for="team in teams" :key="team.ID" :value="team.ID">{{ team.Name }}</option>
                </select>
              </div>
            </div>

            <!-- Right Column: Permissions -->
            <div class="w-1/2 pl-4">
              <label class="block mb-2">Permissions: </label>
              <div class="permissions-grid">
                <div>
                  <label for="login">Login: </label>
                  <input v-model="memberData.Permissions.Login" type="checkbox" id="login" />
                </div>
                <div>
                  <label for="members">Members: </label>
                  <input v-model="memberData.Permissions.Members" type="number" id="members" min="0" max="3" />
                </div>
                <div>
                  <label for="teams">Teams: </label>
                  <input v-model="memberData.Permissions.Teams" type="number" id="teams" min="0" max="3" />
                </div>
                <div>
                  <label for="events">Events: </label>
                  <input v-model="memberData.Permissions.Events" type="number" id="events" min="0" max="3" />
                </div>
                <div>
                  <label for="newsletter">Newsletter: </label>
                  <input v-model="memberData.Permissions.Newsletter" type="number" id="newsletter" min="0" max="3" />
                </div>
                <div>
                  <label for="form">Form: </label>
                  <input v-model="memberData.Permissions.Form" type="number" id="form" min="0" max="3" />
                </div>
                <div>
                  <label for="website">Website: </label>
                  <input v-model="memberData.Permissions.Website" type="number" id="website" min="0" max="3" />
                </div>
                <div>
                  <label for="order">Order: </label>
                  <input v-model="memberData.Permissions.Order" type="number" id="order" min="0" max="3" />
                </div>
                <div>
                  <label for="sponsors">Sponsors: </label>
                  <input v-model="memberData.Permissions.Sponsors" type="number" id="sponsors" min="0" max="3" />
                </div>
              </div>
            </div>
          </div>

          <div class="p-3 flex space-x-2 mt-4 justify-end">
            <button type="submit" class="bg-gray-800 text-white px-3 py-1 rounded hover:bg-gray-700">Create</button>
            <button type="button" @click="closePopup" class="bg-green-600 text-white px-3 py-1 rounded hover:bg-green-700" >Cancel</button>
          </div>
        </form>
      </div>
    </div>
  </div>
</template>

<style scoped>
</style>
