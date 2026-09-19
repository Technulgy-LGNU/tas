<script setup lang="ts">
import { computed, onBeforeUnmount, onMounted, ref } from 'vue'
import AppIcon from '@/components/AppIcon.vue'
import { apiFetch, currentUser } from '@/lib/auth'

interface LibraryImage {
  id: string
  cloudflareId: string
  name: string
  altText: string
  filename: string
  contentType: string
  size: number
  status: string
  createdAt: string
  url: string
}
interface ImagePage {
  images: LibraryImage[]
  total: number
  page: number
  pageSize: number
}
const canEdit = computed(
  () => currentUser.value?.roles?.some((r) => r === 'editor' || r === 'admin') ?? false,
)
const canDelete = computed(() => currentUser.value?.roles?.includes('admin') ?? false)
const images = ref<LibraryImage[]>([])
const total = ref(0)
const page = ref(1)
const pageSize = ref(24)
const pages = computed(() => Math.max(1, Math.ceil(total.value / pageSize.value)))
const search = ref('')
const appliedSearch = ref('')
const loading = ref(true)
const loadError = ref('')
const notice = ref('')
const brokenImages = ref<Record<string, boolean>>({})
let controller: AbortController | undefined

async function responseError(response: Response): Promise<Error> {
  if (response.status === 413)
    return new Error('This file is too large. Choose an image under 10 MB.')
  const data = (await response.json().catch(() => null)) as { error?: string } | null
  return new Error(data?.error || 'The request could not be completed. Please try again.')
}
async function loadImages(targetPage = page.value) {
  controller?.abort()
  const requestController = new AbortController()
  controller = requestController
  loading.value = true
  loadError.value = ''
  try {
    const query = new URLSearchParams({ page: String(targetPage), search: appliedSearch.value })
    const response = await apiFetch(`/api/v1/images?${query}`, { signal: requestController.signal })
    if (!response.ok) throw await responseError(response)
    const data: ImagePage = await response.json()
    if (requestController.signal.aborted) return
    images.value = data.images
    total.value = data.total
    page.value = data.page
    pageSize.value = data.pageSize
    brokenImages.value = {}
    if (targetPage > 1 && !data.images.length) {
      await loadImages(Math.max(1, Math.ceil(data.total / data.pageSize)))
      return
    }
  } catch (err) {
    if (!requestController.signal.aborted)
      loadError.value = err instanceof Error ? err.message : 'Could not load images.'
  } finally {
    if (controller === requestController) loading.value = false
  }
}
function applySearch() {
  appliedSearch.value = search.value.trim()
  void loadImages(1)
}
function clearSearch() {
  search.value = ''
  applySearch()
}

const dialog = ref<HTMLDialogElement>()
const mode = ref<'upload' | 'edit' | 'delete'>('upload')
const selected = ref<LibraryImage | null>(null)
const name = ref('')
const altText = ref('')
const file = ref<File | null>(null)
const fileInput = ref<HTMLInputElement>()
const preview = ref('')
const formError = ref('')
const busy = ref(false)
const dragging = ref(false)
const dialogTitle = computed(() =>
  mode.value === 'upload' ? 'Add an image' : mode.value === 'edit' ? 'Edit image' : 'Delete image?',
)
function releasePreview() {
  if (preview.value) URL.revokeObjectURL(preview.value)
  preview.value = ''
}
function openDialog(action: 'upload' | 'edit' | 'delete', image: LibraryImage | null = null) {
  mode.value = action
  selected.value = image
  name.value = image?.name ?? ''
  altText.value = image?.altText ?? ''
  file.value = null
  formError.value = ''
  notice.value = ''
  releasePreview()
  if (fileInput.value) fileInput.value.value = ''
  dialog.value?.showModal()
}
function closeDialog() {
  if (!busy.value) dialog.value?.close()
}
function chooseFile(value?: File) {
  formError.value = ''
  if (!value) return
  if (!['image/jpeg', 'image/png', 'image/gif', 'image/webp'].includes(value.type)) {
    formError.value = 'Choose a JPEG, PNG, GIF or WebP image.'
    return
  }
  if (value.size === 0 || value.size > 10 * 1024 * 1024) {
    formError.value = 'Choose a non-empty image up to 10 MB.'
    return
  }
  releasePreview()
  file.value = value
  preview.value = URL.createObjectURL(value)
  if (!name.value) name.value = value.name.replace(/\.[^.]+$/, '').slice(0, 200)
}
function dropFile(event: DragEvent) {
  dragging.value = false
  if (busy.value) return
  if (event.dataTransfer?.files.length !== 1) {
    formError.value = 'Please add one image at a time.'
    return
  }
  chooseFile(event.dataTransfer.files[0])
}
function formatSize(bytes: number) {
  return bytes >= 1024 * 1024
    ? `${(bytes / (1024 * 1024)).toFixed(1)} MB`
    : `${Math.max(1, Math.round(bytes / 1024))} KB`
}
function formatDate(value: string) {
  return new Intl.DateTimeFormat(undefined, {
    day: 'numeric',
    month: 'short',
    year: 'numeric',
  }).format(new Date(value))
}
async function submit() {
  if (busy.value) return
  if (mode.value === 'upload' && !file.value) {
    formError.value = 'Choose an image to upload.'
    return
  }
  if (mode.value !== 'delete' && !name.value.trim()) {
    formError.value = 'Give this image a name.'
    return
  }
  busy.value = true
  formError.value = ''
  try {
    let response: Response
    if (mode.value === 'upload') {
      const body = new FormData()
      body.append('file', file.value!)
      body.append('name', name.value.trim())
      body.append('altText', altText.value.trim())
      response = await apiFetch('/api/v1/images', { method: 'POST', body })
    } else if (mode.value === 'edit') {
      response = await apiFetch(`/api/v1/images/${selected.value!.id}`, {
        method: 'PATCH',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ name: name.value.trim(), altText: altText.value.trim() }),
      })
    } else {
      response = await apiFetch(`/api/v1/images/${selected.value!.id}`, { method: 'DELETE' })
    }
    if (!response.ok) throw await responseError(response)
    notice.value =
      mode.value === 'upload'
        ? 'Image uploaded.'
        : mode.value === 'edit'
          ? 'Image details saved.'
          : 'Image deleted.'
    dialog.value?.close()
    if (mode.value === 'upload') {
      search.value = ''
      appliedSearch.value = ''
    }
    await loadImages(mode.value === 'upload' ? 1 : page.value)
  } catch (err) {
    formError.value = err instanceof Error ? err.message : 'The request failed. Please try again.'
    // Refresh after failures as uploads/deletions may have partially completed.
    await loadImages()
  } finally {
    busy.value = false
  }
}
onMounted(() => {
  void loadImages()
})
onBeforeUnmount(() => {
  controller?.abort()
  releasePreview()
})
</script>

<template>
  <section aria-labelledby="images-title">
    <div class="page-heading">
      <div>
        <p class="eyebrow">CONTENT</p>
        <h1 id="images-title">Image library</h1>
        <p class="muted">A home for your team's visual content.</p>
      </div>
      <button v-if="canEdit" class="button primary" @click="openDialog('upload')">
        <AppIcon name="plus" :size="18" /> Add image
      </button>
      <span v-else class="badge">View only</span>
    </div>
    <div class="library-toolbar">
      <form class="search-field" role="search" @submit.prevent="applySearch">
        <AppIcon name="search" :size="18" /><input
          v-model="search"
          type="search"
          maxlength="200"
          placeholder="Search images by name…"
          aria-label="Search images by name"
        /><button class="search-submit" type="submit">Search</button>
      </form>
      <span class="library-count" aria-live="polite"
        >{{ total }} {{ total === 1 ? 'image' : 'images' }}{{ appliedSearch ? ' found' : '' }}</span
      >
    </div>
    <p v-if="notice" class="success-message" role="status">
      <AppIcon name="check" :size="18" />{{ notice }}
    </p>
    <div v-if="loadError" class="empty-state" role="alert">
      <AppIcon name="image" :size="36" />
      <h2>Could not load the library</h2>
      <p>{{ loadError }}</p>
      <button class="button secondary" @click="loadImages()">Try again</button>
    </div>
    <div v-else-if="loading" class="image-grid" aria-busy="true" aria-label="Loading images">
      <div v-for="i in 6" :key="i" class="image-card skeleton">
        <div class="image-preview" />
        <div class="skeleton-line" />
      </div>
      <span class="sr-only" role="status">Loading images…</span>
    </div>
    <div v-else-if="!images.length" class="empty-state">
      <span class="empty-icon"><AppIcon name="image" :size="32" /></span>
      <h2>{{ appliedSearch ? 'No matching images' : 'Your library starts here' }}</h2>
      <p>
        {{
          appliedSearch
            ? 'Try another name or clear your search.'
            : canEdit
              ? 'Upload your first image and give it a name so it is easy to find.'
              : 'Images uploaded by your team will appear here.'
        }}
      </p>
      <button v-if="appliedSearch" class="button secondary" @click="clearSearch">
        Clear search</button
      ><button v-else-if="canEdit" class="button primary" @click="openDialog('upload')">
        <AppIcon name="plus" :size="18" /> Add your first image
      </button>
    </div>
    <div v-else class="image-grid">
      <article v-for="image in images" :key="image.id" class="image-card">
        <div class="image-preview">
          <img
            v-if="image.url && !brokenImages[image.id]"
            :src="image.url"
            :alt="image.altText || image.name"
            loading="lazy"
            decoding="async"
            @error="brokenImages[image.id] = true"
          />
          <div v-else class="image-placeholder">
            <AppIcon name="image" :size="30" /><span>{{
              image.status !== 'ready' ? 'Upload incomplete' : 'Preview unavailable'
            }}</span>
          </div>
          <span class="image-format">{{
            image.contentType.replace('image/', '').toUpperCase()
          }}</span>
        </div>
        <div class="image-details">
          <h2 :title="image.name">{{ image.name }}</h2>
          <p class="image-meta">
            {{ formatSize(image.size) }}<span>·</span>{{ formatDate(image.createdAt) }}
          </p>
          <div class="image-card-footer">
            <span class="image-alt-status">{{
              image.status !== 'ready'
                ? 'Refresh or ask an admin to remove'
                : image.altText
                  ? 'Alt text added'
                  : 'No alt text'
            }}</span>
            <div class="image-actions">
              <button
                v-if="canEdit"
                class="icon-button"
                :aria-label="`Edit ${image.name}`"
                title="Edit image"
                @click="openDialog('edit', image)"
              >
                <AppIcon name="edit" :size="16" /></button
              ><button
                v-if="canDelete"
                class="icon-button danger-icon"
                :aria-label="`Delete ${image.name}`"
                title="Delete image"
                @click="openDialog('delete', image)"
              >
                <AppIcon name="trash" :size="16" />
              </button>
            </div>
          </div>
        </div>
      </article>
    </div>
    <div v-if="!loading && !loadError && total > pageSize" class="pagination">
      <button class="button secondary" :disabled="page <= 1" @click="loadImages(page - 1)">
        Previous</button
      ><span>Page {{ page }} of {{ pages }}</span
      ><button class="button secondary" :disabled="page >= pages" @click="loadImages(page + 1)">
        Next
      </button>
    </div>

    <dialog
      ref="dialog"
      class="image-dialog"
      aria-labelledby="image-dialog-title"
      @cancel="
        (event) => {
          if (busy) event.preventDefault()
        }
      "
      @close="releasePreview"
      @click="
        (event) => {
          if (event.target === dialog) closeDialog()
        }
      "
    >
      <form @submit.prevent="submit">
        <div class="dialog-heading">
          <div>
            <p class="eyebrow">IMAGE LIBRARY</p>
            <h2 id="image-dialog-title">{{ dialogTitle }}</h2>
          </div>
          <button
            type="button"
            class="icon-button"
            aria-label="Close dialog"
            :disabled="busy"
            @click="closeDialog"
          >
            <AppIcon name="close" />
          </button>
        </div>
        <template v-if="mode === 'delete'"
          ><p class="delete-description">
            Permanently delete <strong>{{ selected?.name }}</strong
            >? This removes the image from your library and Cloudflare. Any pages using it will lose
            the image.
          </p>
          <p class="hint">This cannot be undone.</p></template
        >
        <template v-else>
          <label
            v-if="mode === 'upload'"
            class="file-drop"
            :class="{ dragging, 'has-preview': preview }"
            @dragover.prevent="dragging = true"
            @dragleave.prevent="dragging = false"
            @drop.prevent="dropFile"
          >
            <input
              ref="fileInput"
              class="sr-only"
              type="file"
              accept="image/jpeg,image/png,image/gif,image/webp"
              :disabled="busy"
              aria-label="Choose image file"
              @change="chooseFile(($event.target as HTMLInputElement).files?.[0])"
            />
            <img v-if="preview" :src="preview" alt="Selected image preview" /><AppIcon
              v-else
              name="upload"
              :size="30"
            />
            <strong>{{ file ? file.name : 'Choose an image or drop it here' }}</strong
            ><span>{{
              file
                ? `${formatSize(file.size)} · Click to change`
                : 'JPEG, PNG, GIF or WebP · Up to 10 MB'
            }}</span>
          </label>
          <div v-else class="edit-preview">
            <img v-if="selected?.url" :src="selected.url" :alt="selected.name" />
            <div>
              <strong>{{ selected?.filename }}</strong>
              <p class="hint">{{ formatSize(selected?.size ?? 0) }}</p>
            </div>
          </div>
          <label class="form-field"
            >Image name
            <input
              v-model="name"
              required
              maxlength="200"
              :disabled="busy"
              placeholder="e.g. Summer campaign hero"
          /></label>
          <label class="form-field"
            >Alt text <span class="optional">Optional</span
            ><textarea
              v-model="altText"
              maxlength="1000"
              rows="3"
              :disabled="busy"
              placeholder="Describe what is in the image…"
            /><span class="field-help"
              >A short description for people who use screen readers.</span
            ></label
          >
        </template>
        <p v-if="formError" class="form-error" role="alert">{{ formError }}</p>
        <p v-if="busy && mode === 'upload'" class="hint" role="status">
          Uploading your image. Please keep this window open.
        </p>
        <div class="dialog-actions">
          <button type="button" class="button secondary" :disabled="busy" @click="closeDialog">
            Cancel</button
          ><button
            type="submit"
            class="button"
            :class="mode === 'delete' ? 'danger' : 'primary'"
            :disabled="busy || (mode === 'upload' && !file)"
          >
            <span v-if="busy" class="spinner" />{{
              busy
                ? mode === 'upload'
                  ? 'Uploading…'
                  : mode === 'delete'
                    ? 'Deleting…'
                    : 'Saving…'
                : mode === 'upload'
                  ? 'Upload image'
                  : mode === 'delete'
                    ? 'Delete image'
                    : 'Save changes'
            }}
          </button>
        </div>
      </form>
    </dialog>
  </section>
</template>
