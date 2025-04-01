<script setup lang="ts">
defineProps({
  visible: {
    type: Boolean,
    required: true,
  },
  content: {
    type: String,
    default: "",
  },
  loading: {
    type: Boolean,
    default: false,
  },
});

const emit = defineEmits(["close"]);

const close = () => {
  emit("close");
};
</script>

<template>
  <div v-if="visible" class="fixed inset-0 flex items-center justify-center bg-opacity-50 backdrop-blur-sm">
    <div class="bg-white p-6 rounded-lg shadow-lg w-96 relative">
      <!-- Close Button -->
      <button @click="close" class="absolute top-3 right-3 text-gray-500 hover:text-gray-800">
        &times;
      </button>

      <h2 class="text-xl font-semibold text-gray-800 mb-4">Form Content</h2>

      <p v-if="loading" class="text-gray-500">Loading content...</p>
      <p v-else class="text-gray-700 whitespace-pre-wrap">{{ content }}</p>
    </div>
  </div>
</template>

<style scoped>

</style>
