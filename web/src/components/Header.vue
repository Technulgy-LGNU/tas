<script setup lang="ts">
import { ref } from "vue";
import { useRouter } from "vue-router";
import axios from 'axios'
import Cookies from 'js-cookie'

const router = useRouter();
const mobileMenuOpen = ref(false);

const toggleMobileMenu = () => {
  mobileMenuOpen.value = !mobileMenuOpen.value;
};

const props = defineProps({
  page: {
    type: String,
  }
})

const logout = async () => {
  try {
    await axios
      .delete('/api/logout', {
        data: {
          token: Cookies.get("token"),
          deviceId: Cookies.get("deviceId"),
        }
      })
      .then(() => {
        Cookies.remove("token")
        Cookies.remove("admin")
        Cookies.remove("members")
        Cookies.remove("teams")
        Cookies.remove("events")
        Cookies.remove("newsletter")
        Cookies.remove("form")
        Cookies.remove("website")
        Cookies.remove("orders")
        Cookies.remove("inventory")
        Cookies.remove("sponsors")
        router.push({ name: "login" })
      });
  } catch (error) {
    console.error("Logout failed:", error);
  }
};
</script>

<template>
  <header class="w-full bg-gray-800 text-white shadow-md">
    <div class="max-w-7xl mx-auto px-6 flex justify-between items-center h-16">
      <!-- Logo -->
      <router-link to="/" class="text-xl font-semibold">
        T.A.S. {{ props.page }}
      </router-link>

      <!-- Navigation (Desktop) -->
      <nav class="hidden md:flex space-x-6">
        <router-link to="/" class="hover:text-gray-300" :class="{ 'text-gray-300': props.page === 'Dashboard'}">Dashboard</router-link>
        <router-link v-if="Cookies.get('admin') === 'true' || Number(Cookies.get('members')) >= 1" to="/members" class="hover:text-gray-300" :class="{ 'text-gray-300': props.page === 'Members'}">Members</router-link>
        <router-link v-if="Cookies.get('admin') === 'true' || Number(Cookies.get('teams')) >= 1" to="/teams" class="hover:text-gray-300" :class="{ 'text-gray-300': props.page === 'Teams'}">Teams</router-link>
        <router-link v-if="Cookies.get('admin') === 'true' || Number(Cookies.get('forms')) >= 1" to="/forms" class="hover:text-gray-300" :class="{ 'text-gray-300': props.page === 'Forms'}">Forms</router-link>
        <router-link v-if="Cookies.get('admin') === 'true' || Number(Cookies.get('inventory')) >= 1" to="/inventory" lass="hover:text-gray-300" :class="{ 'text-gray-300': props.page === 'Inventory'}">Inventory</router-link>
      </nav>

      <!-- Logout Button -->
      <button @click="logout" class="bg-red-500 px-4 py-2 rounded-lg hover:bg-red-600 transition">
        Logout
      </button>

      <!-- Mobile Menu Button -->
      <button @click="toggleMobileMenu" class="md:hidden text-white text-2xl">
        ☰
      </button>
    </div>

    <!-- Mobile Navigation -->
    <div v-if="mobileMenuOpen" class="md:hidden bg-gray-700 py-3">
      <router-link :v-if="Cookies.get('admin') === 'true' || Number(Cookies.get('members')) >= 1" to="/members" class="hover:text-gray-300" :class="{ 'text-gray-300': props.page === 'Members'}">Members</router-link>
      <router-link :v-if="Cookies.get('admin') === 'true' || Number(Cookies.get('teams')) >= 1" to="/teams" class="hover:text-gray-300" :class="{ 'text-gray-300': props.page === 'Teams'}">Teams</router-link>
      <router-link :v-if="Cookies.get('admin') === 'true' || Number(Cookies.get('events')) >= 1" to="/events" class="hover:text-gray-300" :class="{ 'text-gray-300': props.page === 'Events'}">Events</router-link>
      <router-link :v-if="Cookies.get('admin') === 'true' || Number(Cookies.get('newsletter')) >= 1" to="/newsletter" class="hover:text-gray-300" :class="{ 'text-gray-300': props.page === 'Newsletter'}">Newsletter</router-link>
      <router-link :v-if="Cookies.get('admin') === 'true' || Number(Cookies.get('forms')) >= 1" to="/forms" class="hover:text-gray-300" :class="{ 'text-gray-300': props.page === 'Forms'}">Forms</router-link>
      <router-link :v-if="Cookies.get('admin') === 'true' || Number(Cookies.get('website')) >= 1" to="/website" class="hover:text-gray-300" :class="{ 'text-gray-300': props.page === 'Website'}">Website</router-link>
      <router-link :v-if="Cookies.get('admin') === 'true' || Number(Cookies.get('orders')) >= 1" to="/orders" class="hover:text-gray-300" :class="{ 'text-gray-300': props.page === 'Orders'}">Orders</router-link>
      <router-link :v-if="Cookies.get('admin') === 'true' || Number(Cookies.get('sponsors')) >= 1" to="/sponsors" class="hover:text-gray-300" :class="{ 'text-gray-300': props.page === 'Sponsors'}">Sponsors</router-link>
    </div>
  </header>
</template>

<style scoped>

</style>
