<script setup lang="ts">
import { ref } from 'vue'
import PopUp from '@/components/PopUp.vue'
import ResetPassword from '@/components/ResetPassword.vue'
import axios from 'axios'
import Cookies from 'js-cookie'
import router from '@/router'

const email = ref('')
const password = ref('')
const handleLogin = async () => {
  try {
    await axios
      .post('/api/login', {
        email: email.value,
        password: password.value,
        deviceId: Cookies.get('deviceId'),
      }, {
        headers: {
          'Content-Type': 'application/json',
        },
      })
      .then((res) => {
        if (res.status === 200) {
          Cookies.set('token', res.data.token, { expires: 90, sameSite: 'strict' })
          Cookies.set('admin', res.data.perms.admin, { expires: 90 })
          Cookies.set('members', res.data.perms.members, { expires: 90 })
          Cookies.set('teams', res.data.perms.teams, { expires: 90 })
          Cookies.set('events', res.data.perms.events, { expires: 90 })
          Cookies.set('newsletter', res.data.perms.newsletter, { expires: 90 })
          Cookies.set('form', res.data.perms.form, { expires: 90 })
          Cookies.set('website', res.data.perms.website, { expires: 90 })
          Cookies.set('orders', res.data.perms.orders, { expires: 90 })
          Cookies.set('inventory', res.data.perms.inventory, { expires: 90 })
          Cookies.set('sponsors', res.data.perms.sponsors, { expires: 90 })
          router.push({ name: 'home' })
        } else {
          popUp.value?.show("Login failed")
          console.error(res)
        }
      })
  } catch (error) {
    console.error('Login error:', error)
    popUp.value?.show("Login failed | Wrong email or password")
    password.value = ''
  }
}

const resetPopupRef = ref<InstanceType<typeof ResetPassword> | null>(null);

const openResetPopup = () => {
  resetPopupRef.value?.show();
};

const popUp = ref<InstanceType<typeof PopUp> | null>(null);
</script>

<template>
  <div class="flex min-h-screen items-center justify-center bg-gray-100">
    <div class="w-full max-w-md bg-white p-8 rounded-2xl shadow-lg">
      <h2 class="text-2xl font-semibold text-gray-800 text-center mb-6">Login</h2>
      <form @submit.prevent="handleLogin" class="space-y-4">
        <div>
          <label class="block text-gray-600 text-sm mb-1">Email</label>
          <input
            v-model="email"
            type="email"
            class="w-full p-3 border border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500"
            required
          />
        </div>
        <div>
          <label class="block text-gray-600 text-sm mb-1">Password</label>
          <input
            v-model="password"
            type="password"
            class="w-full p-3 border border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500"
            required
          />
        </div>
        <button
          type="submit"
          class="w-full bg-gray-700 text-white p-3 rounded-lg hover:bg-gray-900 transition">
          Login
        </button>
      </form>
      <p class="text-center text-gray-500 text-sm mt-4">
        Forgot your password?
        <button
          class="text-blue-500 hover:text-blue-700 hover:underline transition"
          @click="openResetPopup">
          Reset Password</button>
      </p>
    </div>
    <ResetPassword ref="resetPopupRef" />
    <PopUp ref='popUp' />
  </div>
</template>

<style scoped>
</style>
