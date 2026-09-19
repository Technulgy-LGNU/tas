<script setup lang="ts">
import { computed, onBeforeUnmount, ref, watch } from 'vue'
import { apiFetch } from '@/lib/auth'
import {
  bilingual,
  libraryImages,
  loadLibraryImage,
  type ContentImage,
  type LibraryImage,
} from '@/lib/website'
import LocalizedField from './LocalizedField.vue'
import AppIcon from '@/components/AppIcon.vue'
const props = withDefaults(
  defineProps<{
    modelValue: ContentImage[]
    cover?: string
    chooseCover?: boolean
    max?: number
    disabled?: boolean
    label?: string
  }>(),
  { cover: '', chooseCover: false, max: 50, disabled: false, label: 'Images' },
)
const emit = defineEmits<{
  'update:modelValue': [value: ContentImage[]]
  'update:cover': [value: string]
}>()
const dialog = ref<HTMLDialogElement>()
const library = ref<LibraryImage[]>([])
const known = libraryImages
const page = ref(1),
  total = ref(0),
  search = ref(''),
  loading = ref(false),
  error = ref('')
let controller: AbortController | undefined
const selected = computed(() => new Set(props.modelValue.map((image) => image.id)))
async function load(target = 1) {
  controller?.abort()
  const abort = new AbortController()
  controller = abort
  loading.value = true
  error.value = ''
  try {
    const response = await apiFetch(
      `/api/v1/images?${new URLSearchParams({ page: String(target), search: search.value })}`,
      { signal: abort.signal },
    )
    if (!response.ok) throw new Error('Could not load the image library.')
    const data: { images: LibraryImage[]; total: number } = await response.json()
    if (abort.signal.aborted) return
    library.value = data.images.filter((image) => image.status === 'ready')
    total.value = data.total
    page.value = target
    for (const image of data.images) known[image.id] = image
  } catch (err) {
    if (!abort.signal.aborted)
      error.value = err instanceof Error ? err.message : 'Could not load images.'
  } finally {
    if (controller === abort) loading.value = false
  }
}
function open() {
  dialog.value?.showModal()
  void load()
}
function add(image: LibraryImage) {
  if (selected.value.has(image.id) || props.modelValue.length >= props.max) return
  emit('update:modelValue', [...props.modelValue, { id: image.id, alt: bilingual() }])
  if (props.chooseCover && !props.cover) emit('update:cover', image.id)
  if (props.max === 1) dialog.value?.close()
}
function remove(index: number) {
  const next = props.modelValue.filter((_, i) => i !== index)
  if (props.modelValue[index]?.id === props.cover) emit('update:cover', next[0]?.id ?? '')
  emit('update:modelValue', next)
}
function move(index: number, direction: number) {
  const next = [...props.modelValue],
    other = next[index + direction],
    current = next[index]
  if (!other || !current) return
  next[index] = other
  next[index + direction] = current
  emit('update:modelValue', next)
}
watch(
  () => props.modelValue.map((image) => image.id),
  (ids) => {
    for (const id of ids) void loadLibraryImage(id)
  },
  { immediate: true },
)
onBeforeUnmount(() => controller?.abort())
</script>
<template>
  <section class="content-images">
    <div class="section-title">
      <h3>
        {{ label }} <span class="muted">({{ modelValue.length }})</span>
      </h3>
      <button
        v-if="!disabled && modelValue.length < max"
        class="button secondary"
        type="button"
        @click="open"
      >
        <AppIcon name="plus" :size="16" />Choose images
      </button>
    </div>
    <p v-if="!modelValue.length" class="field-help">
      Select images from the library. Their order here is their order on the website.
    </p>
    <div v-for="(image, index) in modelValue" :key="image.id" class="selected-image">
      <div class="selected-image-bar">
        <img v-if="known[image.id]?.url" :src="known[image.id]?.url" alt="" /><AppIcon
          v-else
          name="image"
        />
        <span>{{ known[image.id]?.name || `Image ${index + 1}` }}</span>
        <label v-if="chooseCover" class="cover-option"
          ><input
            type="radio"
            :name="`cover-${label}`"
            :checked="cover === image.id"
            :disabled="disabled"
            @change="emit('update:cover', image.id)"
          />Start / cover</label
        >
        <div v-if="!disabled" class="inline-actions">
          <button
            type="button"
            class="icon-button"
            :disabled="index === 0"
            aria-label="Move image up"
            @click="move(index, -1)"
          >
            ↑</button
          ><button
            type="button"
            class="icon-button"
            :disabled="index === modelValue.length - 1"
            aria-label="Move image down"
            @click="move(index, 1)"
          >
            ↓</button
          ><button
            type="button"
            class="icon-button danger-icon"
            aria-label="Remove image from content"
            @click="remove(index)"
          >
            <AppIcon name="close" :size="16" />
          </button>
        </div>
      </div>
      <LocalizedField
        label="Image alt text"
        :model-value="image.alt"
        :maxlength="1000"
        :disabled="disabled"
        @update:model-value="
          emit(
            'update:modelValue',
            modelValue.map((item, i) => (i === index ? { ...item, alt: $event } : item)),
          )
        "
      />
    </div>
    <dialog
      ref="dialog"
      class="image-dialog image-picker-dialog"
      aria-label="Choose library images"
    >
      <div class="picker-inner">
        <div class="dialog-heading">
          <h2>Choose images</h2>
          <button
            type="button"
            class="icon-button"
            aria-label="Close image picker"
            @click="dialog?.close()"
          >
            <AppIcon name="close" />
          </button>
        </div>
        <form class="search-field" @submit.stop.prevent="load()">
          <input
            v-model="search"
            aria-label="Search library"
            placeholder="Search library…"
          /><button class="search-submit">Search</button>
        </form>
        <p v-if="error" class="error" role="alert">{{ error }}</p>
        <p v-if="loading" role="status">Loading images…</p>
        <div v-else class="picker-grid">
          <button
            v-for="image in library"
            :key="image.id"
            type="button"
            class="picker-image"
            :class="{ chosen: selected.has(image.id) }"
            :disabled="selected.has(image.id) || modelValue.length >= max"
            @click="add(image)"
          >
            <img :src="image.url" alt="" loading="lazy" /><span>{{ image.name }}</span
            ><small v-if="selected.has(image.id)">Selected</small>
          </button>
        </div>
        <p v-if="!loading && !library.length" class="hint">
          No ready images found. Upload images in the image library first.
        </p>
        <div class="pagination">
          <button
            type="button"
            class="button secondary"
            :disabled="page <= 1 || loading"
            @click="load(page - 1)"
          >
            Previous</button
          ><span>{{ page }}</span
          ><button
            type="button"
            class="button secondary"
            :disabled="page * 24 >= total || loading"
            @click="load(page + 1)"
          >
            Next
          </button>
        </div>
        <div class="dialog-actions">
          <a href="/#/images" target="_blank" rel="noopener" class="button secondary"
            >Open image library</a
          ><button type="button" class="button primary" @click="dialog?.close()">Done</button>
        </div>
      </div>
    </dialog>
  </section>
</template>
