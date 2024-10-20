<script setup>
import { RouterView } from 'vue-router'
import Cookies from 'js-cookie'
import router from '@/router/index.js'
import axios from 'axios'

// Checks if a device id exists
if (localStorage.getItem('id') === null) {
  localStorage.setItem("id", makeID(32))
}

// Creates a new device id
function makeID(length) {
  let result = '';
  const characters = 'ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789';
  const charactersLength = characters.length;
  let counter = 0;
  while (counter < length) {
    result += characters.charAt(Math.floor(Math.random() * charactersLength));
    counter += 1;
  }
  return result;
}

// Reads stored data
const authKey = Cookies.get('authToken')
const deviceId = localStorage.getItem('id')

if (authKey === null) {
  router.push('/login')
} else {
  axios
    .post('api/check-login', {
      id: deviceId,
      key: authKey
    })
    .then(res => {
      if (res.status === 200) {
        router.push('/');
      }
    })
    .catch(err => {
      if (err.response.status === 403) {
        Cookies.clear()
        router.push('/login');
      } else {
        Cookies.clear();
        router.push('/login');
        console.log(err);
      }
    })
}

</script>

<template>
    <RouterView />
</template>
