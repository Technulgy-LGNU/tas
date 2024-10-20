<script setup>
import axios from 'axios'
import { ref } from 'vue'
import InfoPopUp from '@/components/InfoPopUp.vue'

const props = defineProps({
  'shown': Boolean,
})

const closePopUp = async () => {
  props.shown = false;
}

const infoPopUp = ref(false);
const infoMessage = ref("");

const resetEmail = ref("");

const resetPassword = async () => {
  try {
    await axios
      .post('api/resetPassword', {
        email: resetEmail.value,
      })
      .then(res => {
        if (res.status === 200) {
          infoPopUp.value = true;
          infoMessage.value = "Reset Email send!";
        }
      })
      .catch(err => {
        if (err.response.status === 404) {
          infoPopUp.value = true;
          infoMessage.value = "Email not found!";
        }
      })
  } catch (error) {
    console.log(error);
  }
}
</script>

<template>
  <div>
    <div v-if="shown" class="modal-overlay">
      <div class="modal-content">
        <div class="login-box">
          <h2 class="text-center">Reset Password</h2>
          <form @submit.prevent="resetPassword">
            <div class="form-group mb-3">
              <label for="email">Email</label>
              <input
                type="email"
                v-model="resetEmail"
                id="email"
                class="input-control"
                placeholder="Enter your email"
                required
              />
            </div>
            <button type="submit" class="btn-primary w-100">Register</button>
          </form>
          <div class="footer-links">
            <button @click="closePopUp" class="btn-link">cancel</button>
          </div>
        </div>
      </div>
    </div>
    <InfoPopUp :shown="infoPopUp" :message="infoMessage" />
  </div>
</template>

<style scoped>

</style>
