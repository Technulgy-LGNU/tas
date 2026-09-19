<script setup lang="ts">
const tabs = [
  ['Home', 'home'],
  ['Teams', 'teams'],
  ['Participation History', 'participation-history'],
  ['Sponsors', 'sponsors'],
  ['Publications', 'publications'],
  ['Blog', 'blog'],
  ['SSL', 'ssl'],
]
</script>
<template>
  <section class="website-workspace">
    <div class="page-heading">
      <div>
        <p class="eyebrow">CONTENT MANAGEMENT</p>
        <h1>Website</h1>
        <p class="muted">Build and maintain your German and English website content.</p>
      </div>
    </div>
    <nav class="website-tabs" aria-label="Website sections">
      <RouterLink v-for="[label, path] in tabs" :key="path" :to="`/website/${path}`">{{
        label
      }}</RouterLink>
    </nav>
    <RouterView />
  </section>
</template>
