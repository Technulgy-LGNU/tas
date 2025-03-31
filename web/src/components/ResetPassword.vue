<script setup lang="ts">
import { ref, defineExpose } from "vue";

const visible = ref(false);
const email = ref("");

const show = () => {
  email.value = "";
  visible.value = true;
};

const close = () => {
  visible.value = false;
};

const submitEmail = () => {
  console.log("Reset email sent to:", email.value);
  close(); // Close after submission
};

// Expose show method to parent
defineExpose({ show });
</script>

<template>
  <transition
    enter-active-class="transition-opacity duration-300"
    enter-from-class="opacity-0"
    enter-to-class="opacity-100"
    leave-active-class="transition-opacity duration-300"
    leave-from-class="opacity-100"
    leave-to-class="opacity-0"
  >
    <div
      v-if="visible"
      class="fixed inset-0 bg-opacity-30 flex items-center justify-center backdrop-blur-sm"
    >
      <div class="bg-white p-6 rounded-lg shadow-lg w-96 relative">
        <!-- Close Button -->
        <button @click="close" class="absolute top-3 right-3 text-gray-500 hover:text-gray-800">
          &times;
        </button>

        <h2 class="text-xl font-semibold text-gray-800 mb-4">Reset Password</h2>

        <!-- Email Input -->
        <input
          v-model="email"
          type="email"
          placeholder="Enter your email"
          class="w-full p-3 border border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500"
          required
        />

        <!-- Buttons -->
        <div class="mt-4 flex justify-end space-x-2">
          <button @click="close" class="bg-gray-300 px-4 py-2 rounded-lg hover:bg-gray-400">
            Cancel
          </button>
          <button @click="submitEmail" class="bg-gray-700 text-white px-4 py-2 rounded-lg hover:bg-gray-900">
            Submit
          </button>
        </div>
      </div>
    </div>
  </transition>
</template>

<style scoped>

</style>
