<script setup lang="ts">
import { computed, ref } from 'vue'
import {
  emptyPart,
  money,
  partFields,
  priceCents,
  type PartFields,
  type StandardPart,
  type Category,
} from '@/lib/orders'
import AppIcon from '@/components/AppIcon.vue'
const props = withDefaults(
  defineProps<{
    title: string
    submitLabel?: string
    categories?: Category[]
    standards?: StandardPart[]
    busy?: boolean
    error?: string
    review?: boolean
  }>(),
  {
    submitLabel: 'Save part',
    categories: () => [],
    standards: () => [],
    busy: false,
    error: '',
    review: false,
  },
)
const emit = defineEmits<{ save: [part: PartFields, note: string] }>()
const dialog = ref<HTMLDialogElement>()
const draft = ref(emptyPart()),
  price = ref('0.00'),
  note = ref(''),
  localError = ref(''),
  template = ref('')
const estimate = computed(() => {
  try {
    return money(priceCents(price.value) * draft.value.amount)
  } catch {
    return '—'
  }
})
function open(part: PartFields = emptyPart()) {
  draft.value = partFields(part)
  price.value = (part.unitPriceCents / 100).toFixed(2)
  note.value = ''
  localError.value = ''
  template.value = ''
  dialog.value?.showModal()
}
function selectTemplate() {
  const part = props.standards.find((p) => p.id === template.value)
  if (part) {
    draft.value = { ...partFields(part), categoryId: draft.value.categoryId }
    price.value = (part.unitPriceCents / 100).toFixed(2)
  }
}
function save() {
  localError.value = ''
  try {
    emit('save', { ...draft.value, unitPriceCents: priceCents(price.value) }, note.value)
  } catch (err) {
    localError.value = err instanceof Error ? err.message : 'Check the price.'
  }
}
defineExpose({ open, close: () => dialog.value?.close() })
</script>
<template>
  <dialog
    ref="dialog"
    class="image-dialog order-part-dialog"
    :aria-label="title"
    @cancel="
      (event) => {
        if (busy) event.preventDefault()
      }
    "
  >
    <form @submit.prevent="save">
      <div class="dialog-heading">
        <h2>{{ title }}</h2>
        <button
          type="button"
          class="icon-button"
          aria-label="Close part editor"
          :disabled="busy"
          @click="dialog?.close()"
        >
          <AppIcon name="close" />
        </button>
      </div>
      <fieldset class="editor-fields" :disabled="busy">
        <label v-if="standards.length && !review" class="form-field"
          >Start from a standard part<select v-model="template" @change="selectTemplate">
            <option value="">Custom part</option>
            <option v-for="p in standards" :key="p.id" :value="p.id">
              {{ p.name }} · {{ p.shop }} · {{ money(p.unitPriceCents) }}
            </option>
          </select></label
        >
        <label class="form-field"
          >Part name<input
            v-model="draft.name"
            aria-label="Part name"
            required
            maxlength="200"
            autofocus
        /></label>
        <div class="order-form-columns">
          <label class="form-field"
            >Amount<input
              v-model.number="draft.amount"
              aria-label="Amount"
              type="number"
              required
              min="1"
              max="10000"
              step="1"
          /></label>
          <label class="form-field"
            >Unit price (EUR)<input
              v-model="price"
              aria-label="Unit price (EUR)"
              inputmode="decimal"
              required
              maxlength="10"
              placeholder="0.00"
          /></label>
        </div>
        <label class="form-field"
          >Shop<input v-model="draft.shop" aria-label="Shop" required maxlength="200"
        /></label>
        <label class="form-field"
          >Product link (optional)<input
            v-model="draft.link"
            aria-label="Product link"
            type="url"
            maxlength="2048"
            placeholder="https://…"
        /></label>
        <label v-if="categories.length" class="form-field"
          >Category<select v-model="draft.categoryId" aria-label="Category">
            <option value="">Uncategorized</option>
            <option v-for="c in categories" :key="c.id" :value="c.id">{{ c.name }}</option>
          </select></label
        >
        <label v-if="review" class="form-field"
          >Review note (optional)<input v-model="note" aria-label="Review note" maxlength="1000"
        /></label>
        <p class="order-estimate">
          Line total <strong>{{ estimate }}</strong>
        </p>
      </fieldset>
      <p v-if="localError || error" class="error" role="alert">{{ localError || error }}</p>
      <div class="dialog-actions">
        <button type="button" class="button secondary" :disabled="busy" @click="dialog?.close()">
          Cancel</button
        ><button class="button primary" :disabled="busy">
          {{ busy ? 'Saving…' : submitLabel }}
        </button>
      </div>
    </form>
  </dialog>
</template>
