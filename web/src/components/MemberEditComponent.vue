<script setup lang="ts">
import { onMounted, ref, watch } from 'vue'
import axios from 'axios'
import Cookies from 'js-cookie'

const props = defineProps({
  visible: {
    type: Boolean,
    required: true,
  },
  memberId: {
    type: Number,
    required: true,
  },
});

const emit = defineEmits(['close', 'updateMember']);

interface Member {
  Id: number;
  Name: string;
  Email: string;
  Gender: string;
  Birthday: Date;
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
  Id: number;
  Name: string;
}

const loading = ref<boolean>(true)
const memberData = ref<Member>({} as Member)
const teams = ref<Teams[]>([])

const fetchMemberData = async () => {
  try {
    await axios
      .get(`/api/getMember/${props.memberId}`, {
        headers: {
          'Authorization': `Bearer ${Cookies.get('token')}`,
          'Content-Type': 'application/json',
          'Accept': 'application/json',
        },
      })
      .then((response) => {
        memberData.value = response.data;
        loading.value = false;
      });
  } catch (error) {
    console.error('Error fetching member data:', error);
    emit('close');
  }
};

const fetchTeams = async () => {

}

const saveMember = async () => {
  try {
    await axios
      .patch(`/api/updateMember/${ props.memberId }`,
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
        emit('updateMember', response.data);
        emit('close');
      })
  } catch (error) {
    console.error('Error saving member data:', error);
    emit('close');
  }
}

const closePopup = () => {
  emit('close');
}

watch(
  () => props.memberId,
  () => {
    if (props.memberId !== -1) {
      fetchMemberData();
    }
  },
)

onMounted(
  () => {
    if (props.memberId !== -1) {
      fetchMemberData();
      fetchTeams();
    }
  },
)

</script>

<template>
  <div v-if="visible" class="popup-overlay">
    <div class="popup-container">
      <h3 class="text-lg font-semibold mb-4">Edit Member</h3>

      <!-- Loading state -->
      <div v-if="loading" class="text-gray-500">Loading...</div>

      <!-- Form to edit member -->
      <div v-else class="popup-content">
        <form @submit.prevent="saveMember">
          <div class="flex">
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
                  <option v-for="team in teams" :key="team.Id" :value="team.Id">{{ team.Name }}</option>
                </select>
              </div>
            </div>

            <!-- Right Column: Permissions -->
            <div class="w-1/2 pl-4">
              <label class="block mb-2">Permissions:</label>
              <div class="permissions-grid">
                <div>
                  <label for="login">Login:</label>
                  <input v-model="memberData.Permissions.Login" type="checkbox" id="login" />
                </div>
                <div>
                  <label for="admin">Admin:</label>
                  <input v-model="memberData.Permissions.Admin" type="checkbox" id="admin" disabled />
                </div>
                <div>
                  <label for="members">Members:</label>
                  <input v-model="memberData.Permissions.Members" type="number" id="members" min="0" max="3" />
                </div>
                <div>
                  <label for="teams">Teams:</label>
                  <input v-model="memberData.Permissions.Teams" type="number" id="teams" min="0" max="3" />
                </div>
                <div>
                  <label for="events">Events:</label>
                  <input v-model="memberData.Permissions.Events" type="number" id="events" min="0" max="3" />
                </div>
                <div>
                  <label for="newsletter">Newsletter:</label>
                  <input v-model="memberData.Permissions.Newsletter" type="number" id="newsletter" min="0" max="3" />
                </div>
                <div>
                  <label for="form">Form:</label>
                  <input v-model="memberData.Permissions.Form" type="number" id="form" min="0" max="3" />
                </div>
                <div>
                  <label for="website">Website:</label>
                  <input v-model="memberData.Permissions.Website" type="number" id="website" min="0" max="3" />
                </div>
                <div>
                  <label for="order">Order:</label>
                  <input v-model="memberData.Permissions.Order" type="number" id="order" min="0" max="3" />
                </div>
                <div>
                  <label for="sponsors">Sponsors:</label>
                  <input v-model="memberData.Permissions.Sponsors" type="number" id="sponsors" min="0" max="3" />
                </div>
              </div>
            </div>
          </div>

          <div class="flex justify-end mt-4">
            <button type="submit" class="btn-save">Save</button>
            <button type="button" @click="closePopup" class="btn-cancel">Cancel</button>
          </div>
        </form>

        <button @click="closePopup" class="absolute top-2 right-2 text-gray-500">X</button>
      </div>
    </div>
  </div>
</template>

<style scoped>
.popup-overlay {
  position: fixed;
  top: 0;
  left: 0;
  width: 100%;
  height: 100%;
  background-color: rgba(0, 0, 0, 0.5);
  display: flex;
  justify-content: center;
  align-items: center;
}

.popup-container {
  background-color: white;
  padding: 20px;
  border-radius: 8px;
  position: relative;
  width: 600px;
  max-height: 90vh;
  overflow-y: auto;
}

/* Scrollable content */
.popup-content {
  max-height: 70vh;
  overflow-y: auto;
}

/* Input fields */
.input-field {
  width: 100%;
  padding: 8px;
  border: 1px solid #ddd;
  border-radius: 4px;
  margin-top: 4px;
}

/* Buttons */
.btn-save {
  background-color: #4CAF50;
  color: white;
  padding: 8px 16px;
  border-radius: 4px;
  margin-left: 8px;
}

.btn-cancel {
  background-color: #f44336;
  color: white;
  padding: 8px 16px;
  border-radius: 4px;
  margin-left: 8px;
}

/* Permissions layout */
.permissions-grid {
  display: grid;
  grid-template-columns: repeat(2, 1fr);
  gap: 8px;
}
</style>
