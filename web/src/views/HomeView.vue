<script setup>
import Cookies from 'js-cookie'
import axios from 'axios'
import router from '@/router/index.js'

const token = Cookies.get('authToken')
if (token === null) {
  router.push('/login')
}
checkLogin()

async function checkLogin() {
  try {
    const response = await axios.get('api/checkLogin', {
      headers: {
        Authorization: 'Bearer ' + token,
      }
    })
    console.log(response.data)
  } catch (error) {
    console.log(error)
  }
}

</script>


<template>
  <h1>Hello, you are logged in!</h1>
</template>

<style scoped>

</style>
