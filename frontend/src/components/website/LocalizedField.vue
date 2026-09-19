<script setup lang="ts">
import type { Localized } from '@/lib/website'
withDefaults(
  defineProps<{
    label: string
    modelValue: Localized
    multiline?: boolean
    maxlength?: number
    disabled?: boolean
  }>(),
  { multiline: false, maxlength: 5000, disabled: false },
)
const emit = defineEmits<{ 'update:modelValue': [value: Localized] }>()
</script>
<template>
  <fieldset class="localized-field" :disabled="disabled">
    <legend>{{ label }}</legend>
    <div class="translation-columns">
      <label v-for="language in ['de', 'en'] as const" :key="language"
        ><span>{{ language === 'de' ? 'Deutsch' : 'English' }}</span>
        <textarea
          v-if="multiline"
          :aria-label="`${label} (${language})`"
          :value="modelValue[language]"
          :maxlength="maxlength"
          rows="4"
          @input="
            emit('update:modelValue', {
              ...modelValue,
              [language]: ($event.target as HTMLTextAreaElement).value,
            })
          "
        />
        <input
          v-else
          :aria-label="`${label} (${language})`"
          :value="modelValue[language]"
          :maxlength="maxlength"
          @input="
            emit('update:modelValue', {
              ...modelValue,
              [language]: ($event.target as HTMLInputElement).value,
            })
          "
        />
      </label>
    </div>
  </fieldset>
</template>
