<script setup lang="ts">
import Header from '@/components/Header.vue'
import PopUp from '@/components/PopUp.vue'
import { ref } from 'vue'

const popUp = ref<InstanceType<typeof PopUp> | null>(null);

</script>

<template>
  <div class="flex flex-col min-h-screen">
    <Header page="Newsletter" />

    <PopUp ref="popUp" />
  </div>
</template>

<style scoped>

</style>
