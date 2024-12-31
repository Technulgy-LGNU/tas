<script setup>
import { RouterView } from 'vue-router'

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
</script>

<template>
    <RouterView />
</template>
