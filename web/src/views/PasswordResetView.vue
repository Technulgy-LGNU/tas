<script setup>
import Header from '@/components/Header.vue'
import Footer from '@/components/Footer.vue'
import { ref } from 'vue'
import axios from 'axios'

const code = ref("");
const newPassword = ref("");
const rNewPassword = ref("");

const resetPassword = async () => {
  try {
    await axios
      .post('api/resetPassword/code', {
        code: code.value,
        newpassword: newPassword.value,
        rnewpassword: rNewPassword.value,
      })
  } catch (error) {
    console.log(error)
  }
}
</script>

<template>
  <div class="page-container">
    <Header :show-logout="false" />

    <!-- Login Section -->
    <div class="login-container">
      <div class="login-box">
        <h2 class="text-center">Reset Password</h2>
        <form @submit.prevent="resetPassword">
          <div class="form-group mb-3">
            <label for="email">Code</label>
            <input
              type="text"
              v-model="code"
              id="email"
              class="input-control"
              placeholder="Enter the code ..."
              required
            />
          </div>
          <div class="form-group mb-3">
            <label for="password">New Password</label>
            <input
              type="password"
              v-model="newPassword"
              id="newPassword"
              class="input-control"
              placeholder="Enter your password"
              required
            />
          </div>
          <div class="form-group mb-3">
            <label for="password">Repeat Password</label>
            <input
              type="password"
              v-model="rNewPassword"
              id="rNewPassword"
              class="input-control"
              placeholder="Enter your password"
              required
            />
          </div>
          <button type="submit" class="btn-primary w-100">Login</button>
        </form>
        <div class="footer-links" style="margin-top: 20px">
          <a href="https://technulgy.com/" target="_blank" class="btn-link">Website</a>
          <router-link to="/login" class="btn-link">Login</router-link>
        </div>
      </div>
    </div>
    <Footer />
  </div>
</template>

<style scoped>

</style>
