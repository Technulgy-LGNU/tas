<script setup lang="ts">
  // Check if user is logged in
</script>

<template>
  <div class="flex felx-col min-h-screen">
    <main ref="mainContent" tabindex="1" class="grow outline-none focus:outline-none">
      <RouterView />
    </main>
  </div>
</template>

<style scoped></style>
