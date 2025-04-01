<script setup lang="ts">
import Header from '@/components/Header.vue'
import PopUp from '@/components/PopUp.vue'
import { onMounted, ref } from 'vue'
import axios from 'axios'
import Cookies from 'js-cookie'
import FormContentComponent from '@/components/FormContentComponent.vue'

const popUp = ref<InstanceType<typeof PopUp> | null>(null);

interface Form {
  ID: number;
  CreatedAt: Date;
  DeletedAt: Date | null;
  UpdatedAt: Date | null;
  Name: string;
  Email: string;
  Message: string;
}

const forms = ref<Form[]>([])
const loading = ref<boolean>(true)

const popupVisible = ref(false);
const popupContent = ref("");
const loadingContent = ref(false);


const fetchForms = async () => {
  try {
    await axios
      .get('/api/getForms', {
        headers: {
          'Authorization': `Bearer ${Cookies.get('token')}`,
          'Content-Type': 'application/json',
          'Accept': 'application/json',
        },
      })
      .then((response) => {
        forms.value = response.data
        loading.value = false
      })
  } catch (error) {
    popUp.value?.show('Error fetching forms')
    console.error('Error fetching forms:', error)
    loading.value = false
  }
}

const formatDate = (date: any): string => {
  // Check if it's already a Date object, otherwise convert it
  const dateObj = new Date(date);

  // Check if the date is valid
  if (isNaN(dateObj.getTime())) {
    return ''; // Return empty if the date is invalid
  }

  const day = String(dateObj.getDate()).padStart(2, '0');
  const month = String(dateObj.getMonth() + 1).padStart(2, '0'); // Months are 0-based
  const year = dateObj.getFullYear();
  const hours = String(dateObj.getHours()).padStart(2, '0');
  const minutes = String(dateObj.getMinutes()).padStart(2, '0');

  return `${day}.${month}.${year} ${hours}:${minutes}`;
};

const openPopup = async (id: number) => {
  popupVisible.value = true;
  loadingContent.value = true;

  try {
    await axios
      .get(`/api/getForm/${id}`, {
        headers: {
          'Authorization': `Bearer ${Cookies.get('token')}`,
          'Content-Type': 'application/json',
          'Accept': 'application/json',
        },
      })
      .then((response) => {
        popupContent.value = response.data.Message;
        loadingContent.value = false
      })
  } catch (error) {
    loadingContent.value = false
    popupVisible.value = false;
    popUp.value?.show('Error fetching form content')
    console.error("Error fetching form content:", error);
  }
};

const closePopup = () => {
  popupVisible.value = false;
};

const deleteForm = async (id: number) => {
  try {
    await axios
      .delete(`/api/deleteForm/${id}`, {
        headers: {
          'Authorization': `Bearer ${Cookies.get('token')}`,
          'Content-Type': 'application/json',
          'Accept': 'application/json',
        },
      })
      .then(() => {
        forms.value = forms.value.filter((form) => form.ID !== id)
        popUp.value?.show('Form deleted successfully')
      })
  } catch (error) {
    popUp.value?.show('Error deleting form')
    console.error('Error deleting form:', error)
  }
}

onMounted(fetchForms)

</script>

<template>
  <div class="flex flex-col min-h-screen">
    <Header page="Forms" />
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
            <th class="p-3 text-left">Email</th>
            <th class="p-3 text-left">Created At</th>
            <th class="p-3">Actions</th>
          </tr>
          </thead>
          <tbody>
          <tr v-for="form in forms" :key="form.ID" class="border-t">
            <td class="p-3">{{ form.Name }}</td>
            <td class="p-3"><a href="mailto:{{ form.Email }}" class="text-blue-600 hover:underline">{{ form.Email }}</a></td>
            <td class="p-3">{{ formatDate(form.CreatedAt) }}</td>
            <td class="p-3 flex space-x-2">
              <button
                @click="openPopup(form.ID)"
                class="bg-green-800 text-white px-3 py-1 rounded hover:bg-green-700"
              >
                View Content
              </button>
              <button
                v-if="Cookies.get('admin') === 'true' || Number(Cookies.get('form')) >= 3"
                @click="deleteForm(form.ID)"
                class="bg-red-500 text-white px-3 py-1 rounded hover:bg-red-600"
              >
                Delete
              </button>
            </td>
          </tr>
          </tbody>
        </table>
      </div>
      <!-- Popup Component -->
      <FormContentComponent
        :visible="popupVisible"
        :content="popupContent"
        :loading="loadingContent"
        @close="closePopup"
      />
    </div>
    <PopUp ref="popUp" />
  </div>
</template>

<style scoped>

</style>
