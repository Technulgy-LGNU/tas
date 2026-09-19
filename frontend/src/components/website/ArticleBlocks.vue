<script setup lang="ts">
import { ref, watch } from 'vue'
import { bilingual, renderText, libraryImages, loadLibraryImage, type Block } from '@/lib/website'
import LocalizedField from './LocalizedField.vue'
import ContentImages from './ContentImages.vue'
import AppIcon from '@/components/AppIcon.vue'
const props = withDefaults(defineProps<{ modelValue: Block[]; disabled?: boolean }>(), {
  disabled: false,
})
const emit = defineEmits<{ 'update:modelValue': [value: Block[]] }>()
const preview = ref(false),
  language = ref<'de' | 'en'>('en')
watch(
  () => props.modelValue.flatMap((block) => block.images.map((image) => image.id)),
  (ids) => {
    for (const id of ids) void loadLibraryImage(id)
  },
  { immediate: true },
)
const types: Block['type'][] = ['text', 'heading', 'image', 'gallery', 'video']
function update(index: number, fields: Partial<Block>) {
  emit(
    'update:modelValue',
    props.modelValue.map((block, i) => (i === index ? { ...block, ...fields } : block)),
  )
}
function add(type: Block['type']) {
  emit('update:modelValue', [
    ...props.modelValue,
    { id: crypto.randomUUID(), type, text: bilingual(), level: 2, images: [], url: '' },
  ])
}
function move(index: number, direction: number) {
  const list = [...props.modelValue],
    first = list[index],
    second = list[index + direction]
  if (!first || !second) return
  list[index] = second
  list[index + direction] = first
  emit('update:modelValue', list)
}
</script>
<template>
  <section class="article-blocks">
    <div class="section-title">
      <h3>Article content</h3>
      <button type="button" class="button secondary" @click="preview = !preview">
        {{ preview ? 'Edit blocks' : 'Preview text' }}
      </button>
    </div>
    <p class="field-help">
      Arrange blocks with the arrows. Text supports Markdown: **bold**, *italic*, lists and
      [links](https://…). Add images through image or gallery blocks.
    </p>
    <div v-if="preview" class="article-preview">
      <label class="form-field"
        >Preview language<select v-model="language">
          <option value="de">Deutsch</option>
          <option value="en">English</option>
        </select></label
      >
      <section v-for="block in modelValue" :key="block.id">
        <component v-if="block.type === 'heading'" :is="block.level === 3 ? 'h3' : 'h2'">{{
          block.text[language]
        }}</component>
        <div
          v-else-if="block.type === 'text'"
          class="prose"
          v-html="renderText(block.text[language])"
        />
        <p v-else-if="block.type === 'video'">
          <a :href="block.url" target="_blank" rel="noopener">{{
            block.text[language] || 'YouTube video'
          }}</a>
        </p>
        <figure v-else>
          <div class="block-preview-images">
            <img
              v-for="image in block.images"
              :key="image.id"
              :src="libraryImages[image.id]?.url"
              :alt="image.alt[language]"
            />
          </div>
          <figcaption>{{ block.text[language] }}</figcaption>
        </figure>
      </section>
    </div>
    <template v-else
      ><div v-for="(block, index) in modelValue" :key="block.id" class="article-block">
        <div class="section-title">
          <h4>{{ index + 1 }}. {{ block.type }}</h4>
          <div v-if="!disabled" class="inline-actions">
            <button
              type="button"
              class="icon-button"
              :disabled="index === 0"
              aria-label="Move block up"
              @click="move(index, -1)"
            >
              ↑</button
            ><button
              type="button"
              class="icon-button"
              :disabled="index === modelValue.length - 1"
              aria-label="Move block down"
              @click="move(index, 1)"
            >
              ↓</button
            ><button
              type="button"
              class="icon-button danger-icon"
              aria-label="Remove block"
              @click="
                emit(
                  'update:modelValue',
                  modelValue.filter((_, i) => i !== index),
                )
              "
            >
              <AppIcon name="trash" :size="16" />
            </button>
          </div>
        </div>
        <label v-if="block.type === 'heading'" class="form-field"
          >Heading size<select
            :value="block.level"
            :disabled="disabled"
            @change="update(index, { level: Number(($event.target as HTMLSelectElement).value) })"
          >
            <option :value="2">Heading 2</option>
            <option :value="3">Heading 3</option>
          </select></label
        >
        <LocalizedField
          :label="
            block.type === 'text'
              ? 'Text'
              : block.type === 'heading'
                ? 'Heading'
                : block.type === 'video'
                  ? 'Video title'
                  : 'Caption (optional)'
          "
          :model-value="block.text"
          :multiline="block.type === 'text'"
          :maxlength="20000"
          :disabled="disabled"
          @update:model-value="update(index, { text: $event })"
        />
        <ContentImages
          v-if="block.type === 'image' || block.type === 'gallery'"
          :model-value="block.images"
          :max="block.type === 'image' ? 1 : 20"
          :disabled="disabled"
          @update:model-value="update(index, { images: $event })"
        />
        <label v-if="block.type === 'video'" class="form-field"
          >YouTube URL<input
            :value="block.url"
            type="url"
            :disabled="disabled"
            placeholder="https://www.youtube.com/watch?v=…"
            @input="update(index, { url: ($event.target as HTMLInputElement).value })"
        /></label>
      </div>
      <div v-if="!disabled" class="add-block-bar">
        <span>Add a block</span
        ><button
          v-for="type in types"
          :key="type"
          type="button"
          class="button secondary"
          :disabled="modelValue.length >= 200"
          @click="add(type)"
        >
          + {{ type }}
        </button>
      </div>
      <p v-if="!modelValue.length" class="hint">
        Start with a heading or a text block, then add images, galleries or videos.
      </p></template
    >
  </section>
</template>
