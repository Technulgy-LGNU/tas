<script setup lang="ts">
import { EnvelopeIcon } from '@heroicons/vue/24/solid'
import { GitHubIcon} from 'vue3-simple-icons'
</script>

<template>
  <footer class="w-full h-20 bg-gray-800 text-white flex items-center justify-between px-6">
    <p class="text-sm">© 2025 Technulgy | TAS</p>
    <div class="flex space-x-4">
      <a href="mailto:contact@technulgy.com" aria-label="Email">
        <EnvelopeIcon class="w-6 h-6" />
      </a>
      <a href="https://github.com/technulgy-lgnu/tas/" target="_blank" aria-label="GitHub">
        <GitHubIcon class="w-6 h-6" />
      </a>
    </div>
  </footer>
</template>

<style scoped>
</style>
