<script setup>
import Cookies from 'js-cookie'
import axios from 'axios'
import { ref } from 'vue'
import InfoPopUp from '@/components/InfoPopUp.vue'
import Footer from '@/components/Footer.vue'
import Header from '@/components/Header.vue'
import PasswordReset from '@/components/PasswordReset.vue'
import router from '@/router/index.js'

// Checks if user is logged in
if (Cookies.get('authToken') !== null) {
  try {
    await axios
      .post('http://localhost:8080/api/checklogin', {
        id: Cookies.get('deviceID'),
        key: Cookies.get('authToken'),
      })
      .then(res => {
        if (res.status === 200) {
          Cookies.set('perms', res.data.perms)
          router.push('/')
        }
      })
      .catch(error => {
        if (error.response.status === 403) {
          Cookies.set('authToken', null)
          Cookies.set('perms', null)
        } else {
          console.log(error.response.status)
        }
      })
  } catch (error) {
    console.log(error)
  }
}

const infoPopUp = ref(false);
const infoMessage = ref('');

const loginEmail = ref('');
const loginPassword = ref('');

const login = async () => {
  try {
    await axios
      .post(':8080/api/login', {
        email: loginEmail.value,
        password: loginPassword.value,
        device: Cookies.get('deviceID'),
      })
      .then(res => {
        if (res.status === 200) {
          Cookies.set('authToken', res.data.key);
          Cookies.set('perms', res.data.perms);
          console.log(res.data.perms);
          router.push('/');
        }
      })
      .catch(error => {
        router.push("/login");
        console.log(error);
      })
  } catch (error) {
    console.log(error);
  }
}

const resetPasswordPopUp = ref(false);

const openResetPasswordPopUp = async () => {
  resetPasswordPopUp.value = true;
}

</script>

<template>
  <div class="page-container">
    <Header :show-logout="false" />

    <!-- Login Section -->
    <div class="login-container">
      <div class="login-box">
        <h2 class="text-center">Login</h2>
        <form @submit.prevent="login">
          <div class="form-group mb-3">
            <label for="email">Email</label>
            <input
              type="email"
              v-model="loginEmail"
              id="email"
              class="input-control"
              placeholder="Enter your email"
              required
            />
          </div>
          <div class="form-group mb-3">
            <label for="password">Password</label>
            <input
              type="password"
              v-model="loginPassword"
              id="password"
              class="input-control"
              placeholder="Enter your password"
              required
            />
          </div>
          <button type="submit" class="btn-primary w-100">Login</button>
        </form>
        <div class="footer-links" style="margin-top: 20px">
          <!-- <a href="https://technulgy.com/" target="_blank" class="btn-link">Website</a> -->
          <router-link to="passwordreset">Website</router-link>
          <button @click="openResetPasswordPopUp" class="btn-link">Reset Password</button>
        </div>
      </div>
    </div>

    <PasswordReset :shown="resetPasswordPopUp" />
    <InfoPopUp :shown="infoPopUp" :message="infoMessage" />
    <Footer />
  </div>
</template>

<style scoped>

</style>
