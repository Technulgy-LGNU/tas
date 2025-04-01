<script setup lang="ts">
import Header from '@/components/Header.vue'
import PopUp from '@/components/PopUp.vue'
import { onMounted, ref } from 'vue'
import axios from 'axios'
import Cookies from 'js-cookie'
import MemberEditComponent from '@/components/MemberEditComponent.vue'
import MemberCreateComponent from '@/components/MemberCreateComponent.vue'

const popUp = ref<InstanceType<typeof PopUp> | null>(null);

interface Member {
  Id: number;
  Name: string;
  Email: string;
  Gender: string;
  Birth: Date;
  TeamId: number;
}

interface Teams {
  ID: number;
  Name: string;
}

const teams = ref<Teams[]>([])

const members = ref<Member[]>([])
const loading = ref<boolean>(true)
const popupVisible = ref(false);
const currentUserId = ref<number>(-1)

const fetchMembers = async () => {
  try {
    await axios
      .get('/api/getMembers', {
        headers: {
          'Authorization': `Bearer ${Cookies.get('token')}`,
          'Content-Type': 'application/json',
          'Accept': 'application/json',
        },
      })
      .then((response) => {
        members.value = response.data
        loading.value = false
      })
  } catch (error) {
    popUp.value?.show('Error fetching members')
    console.error('Error fetching members:', error)
    loading.value = false
  }
}

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
      })
  } catch (error) {
    console.error('Error fetching teams:', error)
  }
}

const editMember = (id: number) => {
  popupVisible.value = true
  currentUserId.value = id
}

const closePopup = () => {
  popupVisible.value = false;
  currentUserId.value = -1;
}

const updateMember = (updatedMember: any) => {
  const index = members.value.findIndex((m) => m.Id === updatedMember.ID);
  if (index !== -1) {
    members.value[index] = updatedMember;
  }
  fetchMembers()
  popUp.value?.show('Member updated successfully');
}

const deleteMember = async (id: number) => {
  try {
    await axios
      .delete(`/api/deleteMember/${id}`, {
        headers: {
          'Authorization': `Bearer ${Cookies.get('token')}`,
          'Content-Type': 'application/json',
          'Accept': 'application/json',
        },
      })
      .then(() => {
        popUp.value?.show('Member deleted successfully')
        fetchMembers()
      })
  } catch (error) {
    popUp.value?.show('Error deleting member')
    console.error('Error deleting member:', error)
  }
}

const showCreateMemberPopup = ref<boolean>(false);

const handleMemberCreated = () => {
  showCreateMemberPopup.value = false;
  popUp.value?.show('Member created successfully');
  fetchMembers();
};

onMounted(async () => {
  await fetchMembers()
  await fetchTeams()
})
</script>

<template>
  <div class="flex flex-col min-h-screen">
    <Header page="Members" />
    <div class="max-w-4xl mx-auto p-6 space-y-2">
      <div class="max-w-7xl mx-auto">
        <h2 class="text-2xl font-semibold mb-4">Members</h2>
        <button
          @click="showCreateMemberPopup = true" class="bg-gray-800 text-white px-4 py-2 rounded-lg hover:bg-gray-700">
          Create Member
        </button>
      </div>

      <!-- Loading State -->
      <p v-if="loading" class="text-gray-500">Loading...</p>

      <!-- Members Table -->
      <div v-else class="bg-white shadow rounded-lg">
        <table class="w-full border-collapse">
          <thead>
          <tr class="bg-gray-200">
            <th class="p-3 text-left">Name</th>
            <th class="p-3 text-left">Email</th>
            <th class="p-3 text-left">Team</th>
            <th class="p-3">Actions</th>
          </tr>
          </thead>
          <tbody>
          <tr v-for="member in members" :key="member.Id" class="border-t">
            <td class="p-3">{{ member.Name}}</td>
            <td class="p-3">{{ member.Email }}</td>
            <td class="p-3">
              <div>
                {{ teams.find(team => team.ID === member.TeamId)?.Name || 'No Team' }}
              </div>
            </td>
            <td class="p-3 flex space-x-2">
              <button
                v-if="Cookies.get('admin') === 'true' || Number(Cookies.get('members')) >= 2"
                @click="editMember(member.Id)"
                class="bg-gray-800 text-white px-3 py-1 rounded hover:bg-gray-700"
              >
                Edit
              </button>
              <button
                v-if="Cookies.get('admin') === 'true' || Number(Cookies.get('members')) >= 3"
                @click="deleteMember(member.Id)"
                class="bg-red-500 text-white px-3 py-1 rounded hover:bg-red-600"
              >
                Delete
              </button>
            </td>
          </tr>
          </tbody>
        </table>
      </div>

      <!-- MembersEditComponent Popup -->
      <MemberEditComponent
        :visible="popupVisible"
        :member-id="currentUserId"
        @close="closePopup"
        @updateMember="updateMember"
      />

      <MemberCreateComponent
        :visible="showCreateMemberPopup"
        @close="showCreateMemberPopup = false"
        @created-member="handleMemberCreated"
      />
    </div>

    <PopUp ref="popUp" />
  </div>
</template>

<style scoped>
</style>
