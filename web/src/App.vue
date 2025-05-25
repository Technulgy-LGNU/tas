<script setup lang="ts">
import { RouterView } from 'vue-router'
import Cookies from 'js-cookie'
import router from '@/router'
import axios from 'axios'
import Footer from '@/components/Footer.vue'

// Check if the deviceId cookie exists and create it if not (length 16 characters)
const deviceId = Cookies.get('deviceId')
if (deviceId === undefined || deviceId.length !== 16) {
  Cookies.set('deviceId', makeID(16), { expires: 90 })
}
function makeID(length: number): string {
  let result = '';
  const characters = 'ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789+-*/=?!"§$%&/(){[]}#,.:;<>|@';
  const charactersLength = characters.length;
  let counter = 0;
  while (counter < length) {
    result += characters.charAt(Math.floor(Math.random() * charactersLength));
    counter += 1;
  }
  return result;
}

// Check if user is on route to the login page or the reset password page
const isLoginPage = router.currentRoute.value.path === '/login'
const isResetPasswordPage = router.currentRoute.value.path === '/resetPassword'
if (!isLoginPage || !isResetPasswordPage) {
  // Check if the token cookies exists and if not send user to login page
  const token = Cookies.get('token')
  if (token === undefined || token.length === 0) {
    router.push({ name: 'login' })
  } else {
    // Check if the token is still valid against backend
    axios
      .post('/api/checkLogin',
        {
          token: Cookies.get('token'),
          deviceId: Cookies.get('deviceId')
        },
        {
          headers: {
            'Content-Type': 'application/json',
            'Accept': 'application/json'
          }
        }
      )
      .then(res => {
        if (res.status === 200) {
          Cookies.set('token', res.data.token)
          Cookies.set('admin', res.data.perms.admin)
          Cookies.set('members', res.data.perms.members)
          Cookies.set('teams', res.data.perms.teams)
          Cookies.set('events', res.data.perms.events)
          Cookies.set('newsletter', res.data.perms.newsletter)
          Cookies.set('form', res.data.perms.form)
          Cookies.set('website', res.data.perms.website)
          Cookies.set('orders', res.data.perms.orders)
          Cookies.set('inventory', res.data.perms.inventory)
          Cookies.set('sponsors', res.data.perms.sponsors)
        } else {
          // Token is not valid, send user to login page
          Cookies.remove('token')
          router.push({ name: 'login' })
        }
      })
      .catch(() => {
        // Token is not valid, send user to login page
        Cookies.remove('token')
        router.push({ name: 'login' })
      })
  }
}

</script>

<template>
  <div class="flex flex-col min-h-screen">
    <main class="flex-grow">
      <RouterView />
    </main>
    <Footer />
  </div>
</template>

<style scoped>
</style>
