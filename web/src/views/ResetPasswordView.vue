<script setup lang="ts">
import PopUp from '@/components/PopUp.vue'
import { ref } from 'vue'

const code = ref('')
const password1 = ref('')
const password2 = ref('')

const handleReset = async () => {

}

const popUp = ref<InstanceType<typeof PopUp> | null>(null);

const showMessage = () => {
  popUp.value?.show("This is a test message!");
};
</script>

<template>
  <div class="flex min-h-screen items-center justify-center bg-gray-100">
    <div class="w-full max-w-md bg-white p-8 rounded-2xl shadow-lg">
      <h2 class="text-2xl font-semibold text-gray-800 text-center mb-6">Login</h2>
      <form @submit.prevent="handleReset" class="space-y-4">
        <div>
          <label class="block text-gray-600 text-sm mb-1">Reset Code</label>
          <input
            v-model="code"
            type="text"
            class="w-full p-3 border border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500"
            required
          />
        </div>
        <div>
          <label class="block text-gray-600 text-sm mb-1">Password</label>
          <input
            v-model="password1"
            type="password"
            class="w-full p-3 border border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500"
            required
          />
        </div>
        <div>
          <label class="block text-gray-600 text-sm mb-1">Repeat Password</label>
          <input
            v-model="password2"
            type="password"
            class="w-full p-3 border border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500"
            required
          />
        </div>
        <button
          type="submit"
          class="w-full bg-gray-700 text-white p-3 rounded-lg hover:bg-gray-900 transition">
          Reset Password
        </button>
      </form>
      <p class="text-center text-gray-500 text-sm mt-4">
        Remembered your password? <a href="#/login" class="hover:underline text-blue-500 hover:text-blue-700">Login</a>
      </p>
    </div>
    <PopUp ref='popUp' />
  </div>
</template>

<style scoped>

</style>
