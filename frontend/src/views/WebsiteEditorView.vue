<script setup lang="ts">
import { computed, onBeforeUnmount, ref, watch } from 'vue'
import { onBeforeRouteLeave, onBeforeRouteUpdate } from 'vue-router'
import { currentUser } from '@/lib/auth'
import { bilingual, newEntry, normalizeEntry, websiteRequest, type Entry } from '@/lib/website'
import LocalizedField from '@/components/website/LocalizedField.vue'
import ContentImages from '@/components/website/ContentImages.vue'
import ArticleBlocks from '@/components/website/ArticleBlocks.vue'
import AppIcon from '@/components/AppIcon.vue'
const props = defineProps<{ kind: string }>()
const labels: Record<string, string> = {
  home: 'Home',
  teams: 'Teams',
  events: 'Participation History',
  sponsors: 'Sponsors',
  publications: 'Publication categories',
  blog: 'Blog articles',
}
const canEdit = computed(
  () => currentUser.value?.roles?.some((role) => ['editor', 'admin'].includes(role)) ?? false,
)
const isAdmin = computed(() => currentUser.value?.roles?.includes('admin') ?? false)
const entries = ref<Entry[]>([]),
  events = ref<Entry[]>([])
const selected = ref<Entry | null>(null)
const saved = ref('')
const dirty = computed(
  () => selected.value !== null && JSON.stringify(selected.value) !== saved.value,
)
const loading = ref(false),
  busy = ref(false),
  error = ref(''),
  notice = ref('')
const filter = ref(''),
  previewLanguage = ref('en'),
  contactEnabled = ref(false)
const deleting = ref<HTMLDialogElement>()
const visibleEntries = computed(() =>
  entries.value.filter((e) =>
    `${e.content.name.en} ${e.content.name.de} ${e.slug}`
      .toLowerCase()
      .includes(filter.value.toLowerCase()),
  ),
)
let loadGeneration = 0
function select(entry: Entry) {
  if (dirty.value && !window.confirm('Discard your unsaved changes?')) return
  selected.value = normalizeEntry(entry)
  saved.value = JSON.stringify(selected.value)
  error.value = ''
  notice.value = ''
}
function create() {
  select(newEntry(props.kind))
}
async function load() {
  const generation = ++loadGeneration
  loading.value = true
  error.value = ''
  selected.value = null
  entries.value = []
  filter.value = ''
  try {
    const [data, eventData, settings] = await Promise.all([
      websiteRequest<{ entries: Entry[] }>(`/${props.kind}`),
      props.kind === 'teams'
        ? websiteRequest<{ entries: Entry[] }>('/events')
        : Promise.resolve({ entries: [] }),
      props.kind === 'home'
        ? websiteRequest<{ contactEnabled: boolean }>('/settings')
        : Promise.resolve({ contactEnabled: false }),
    ])
    if (generation !== loadGeneration) return
    entries.value = data.entries
    events.value = eventData.entries
    contactEnabled.value = settings.contactEnabled
    if (props.kind === 'home') select(entries.value[0] || newEntry('home'))
    else if (entries.value[0]) select(entries.value[0])
  } catch (err) {
    if (generation === loadGeneration)
      error.value = err instanceof Error ? err.message : 'Could not load content.'
  } finally {
    if (generation === loadGeneration) loading.value = false
  }
}
async function save() {
  if (!selected.value || !canEdit.value || busy.value) return
  busy.value = true
  error.value = ''
  notice.value = ''
  try {
    const entry = selected.value
    const response = await websiteRequest<{ entry: Entry }>(
      `/${props.kind}${entry.id ? `/${entry.id}` : ''}`,
      {
        method: entry.id ? 'PUT' : 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(entry),
      },
    )
    selected.value = normalizeEntry(response.entry)
    saved.value = JSON.stringify(selected.value)
    const index = entries.value.findIndex((e) => e.id === response.entry.id)
    if (index >= 0) entries.value[index] = response.entry
    else entries.value.unshift(response.entry)
    notice.value = selected.value.published
      ? props.kind === 'blog' &&
        selected.value.publishAt &&
        new Date(selected.value.publishAt) > new Date()
        ? 'Article saved and scheduled.'
        : 'Changes saved. Published content is now available to the website.'
      : 'Draft saved. It is not visible on the public website.'
  } catch (err) {
    error.value = err instanceof Error ? err.message : 'Could not save content.'
  } finally {
    busy.value = false
  }
}
async function remove() {
  if (!selected.value?.id || busy.value) return
  busy.value = true
  error.value = ''
  try {
    await websiteRequest(`/${props.kind}/${selected.value.id}`, { method: 'DELETE' })
    entries.value = entries.value.filter((e) => e.id !== selected.value?.id)
    selected.value = null
    saved.value = ''
    deleting.value?.close()
    if (props.kind === 'home') select(newEntry('home'))
    notice.value = 'Content deleted.'
  } catch (err) {
    error.value = err instanceof Error ? err.message : 'Could not delete content.'
    deleting.value?.close()
  } finally {
    busy.value = false
  }
}
const publicationTime = computed({
  get() {
    if (!selected.value?.publishAt) return ''
    const date = new Date(selected.value.publishAt)
    return new Date(date.getTime() - date.getTimezoneOffset() * 60000).toISOString().slice(0, 16)
  },
  set(value: string) {
    if (selected.value) selected.value.publishAt = value ? new Date(value).toISOString() : null
  },
})
function moveVideo(index: number, direction: number) {
  const list = selected.value?.content.videos
  if (!list || !list[index] || !list[index + direction]) return
  const current = list[index]
  list[index] = list[index + direction]!
  list[index + direction] = current
}
function beforeUnload(event: BeforeUnloadEvent) {
  if (dirty.value) {
    event.preventDefault()
    event.returnValue = ''
  }
}
window.addEventListener('beforeunload', beforeUnload)
onBeforeUnmount(() => {
  loadGeneration++
  window.removeEventListener('beforeunload', beforeUnload)
})
onBeforeRouteLeave(() => !dirty.value || window.confirm('Discard your unsaved changes?'))
onBeforeRouteUpdate(() => !dirty.value || window.confirm('Discard your unsaved changes?'))
watch(
  () => props.kind,
  () => {
    void load()
  },
  { immediate: true },
)
</script>
<template>
  <div class="website-editor-page">
    <div class="section-title">
      <h2>{{ labels[kind] }}</h2>
      <div class="inline-actions">
        <label class="sr-only" for="public-language">Public preview language</label
        ><select id="public-language" v-model="previewLanguage" class="small-select">
          <option value="en">EN</option>
          <option value="de">DE</option></select
        ><a
          class="button secondary"
          :href="`/website/${kind === 'events' ? 'participation-history' : kind}?lang=${previewLanguage}`"
          target="_blank"
          rel="noopener"
          >Public JSON ↗</a
        ><button
          v-if="canEdit && kind !== 'home'"
          class="button primary"
          :disabled="busy"
          @click="create"
        >
          <AppIcon name="plus" :size="16" />Add new
        </button>
      </div>
    </div>
    <p v-if="error" class="form-error" role="alert">
      {{ error }} <button v-if="!selected" class="search-submit" @click="load">Retry</button>
    </p>
    <p v-if="notice" class="success-message" role="status">{{ notice }}</p>
    <p v-if="loading" class="hint" role="status">Loading website content…</p>
    <div v-else class="content-editor-layout" :class="{ 'single-editor': kind === 'home' }">
      <aside v-if="kind !== 'home'" class="entry-list">
        <label class="sr-only" for="filter-entries">Filter entries</label
        ><input
          id="filter-entries"
          v-model="filter"
          class="entry-filter"
          placeholder="Find content…"
        />
        <p v-if="!visibleEntries.length" class="hint">
          {{ entries.length ? 'No matching entries.' : 'No content yet. Add your first entry.' }}
        </p>
        <button
          v-for="entry in visibleEntries"
          :key="entry.id"
          class="entry-row"
          :class="{ selected: selected?.id === entry.id }"
          :disabled="busy"
          @click="select(entry)"
        >
          <strong>{{ entry.content.name.en || entry.content.name.de || entry.slug }}</strong
          ><small
            >{{
              entry.published
                ? entry.publishAt && new Date(entry.publishAt) > new Date()
                  ? 'Scheduled'
                  : 'Published'
                : 'Draft'
            }}<span v-if="kind === 'events'"> · {{ entry.content.date }}</span></small
          >
        </button>
      </aside>
      <form v-if="selected" class="content-editor" @submit.prevent="save">
        <div class="editor-actions">
          <span class="badge"
            >{{ selected.published ? 'Published / scheduled' : 'Draft'
            }}{{ dirty ? ' · Unsaved changes' : '' }}</span
          ><button
            v-if="isAdmin && selected.id"
            type="button"
            class="icon-button danger-icon"
            aria-label="Delete content"
            :disabled="busy"
            @click="deleting?.showModal()"
          >
            <AppIcon name="trash" /></button
          ><button v-if="canEdit" class="button primary" :disabled="busy">
            {{ busy ? 'Saving…' : 'Save changes' }}
          </button>
        </div>
        <fieldset :disabled="!canEdit || busy" class="editor-fields">
          <div class="publishing-options">
            <label><input v-model="selected.published" type="checkbox" />Publish on website</label
            ><label v-if="kind !== 'home'"
              >Display order<input
                v-model.number="selected.sortOrder"
                type="number"
                min="0"
                max="100000"
            /></label>
          </div>
          <template v-if="kind === 'home'">
            <ContentImages
              v-model="selected.content.images"
              :disabled="!canEdit || busy"
              label="Slideshow images"
            />
            <LocalizedField
              v-model="selected.content.about"
              label="About us"
              multiline
              :maxlength="20000"
            />
            <div class="info-panel">
              <h3>Latest blog articles</h3>
              <p>
                The home API automatically includes the latest three published articles with their
                names, descriptions, front images and links. Drafts and future-dated articles stay
                hidden.
              </p>
              <RouterLink to="/website/blog">Manage blog articles →</RouterLink>
            </div>
            <div class="section-title">
              <h3>Published videos</h3>
              <button
                type="button"
                class="button secondary"
                @click="selected.content.videos.push({ title: bilingual(), url: '' })"
              >
                + Add video
              </button>
            </div>
            <div
              v-for="(video, index) in selected.content.videos"
              :key="index"
              class="article-block"
            >
              <div class="section-title">
                <h4>Video {{ index + 1 }}</h4>
                <div class="inline-actions">
                  <button
                    type="button"
                    class="icon-button"
                    :disabled="index === 0"
                    aria-label="Move video up"
                    @click="moveVideo(index, -1)"
                  >
                    ↑</button
                  ><button
                    type="button"
                    class="icon-button"
                    :disabled="index === selected.content.videos.length - 1"
                    aria-label="Move video down"
                    @click="moveVideo(index, 1)"
                  >
                    ↓</button
                  ><button
                    type="button"
                    class="icon-button danger-icon"
                    aria-label="Remove video"
                    @click="selected.content.videos.splice(index, 1)"
                  >
                    <AppIcon name="trash" :size="16" />
                  </button>
                </div>
              </div>
              <LocalizedField v-model="video.title" label="Video title" :maxlength="200" /><label
                class="form-field"
                >YouTube URL<input
                  v-model="video.url"
                  type="url"
                  maxlength="2048"
                  placeholder="https://www.youtube.com/watch?v=…"
              /></label>
            </div>
            <div class="info-panel">
              <h3>Contact form</h3>
              <p>
                {{
                  contactEnabled
                    ? 'Email delivery is configured. Messages from the website are sent to your configured recipient.'
                    : 'Email delivery is not configured yet. Complete [website.contact] in config.toml and restart TAS.'
                }}
              </p>
              <p>
                The website submits the visitor's name, email, subject and message to the contact
                endpoint.
              </p>
            </div>
          </template>
          <template v-else>
            <LocalizedField
              v-model="selected.content.name"
              :label="kind === 'publications' ? 'Category name' : 'Name'"
              :maxlength="200"
            />
            <label class="form-field"
              >URL slug<input
                v-model="selected.slug"
                maxlength="120"
                pattern="[a-z0-9]+(-[a-z0-9]+)*"
                required
                placeholder="e.g. german-open-2026"
              /><span class="field-help"
                >Keep this stable after publishing so existing links continue to work.</span
              ></label
            >
            <LocalizedField
              v-model="selected.content.description"
              label="Short description"
              multiline
              :maxlength="5000"
            />
            <ContentImages
              v-model="selected.content.images"
              v-model:cover="selected.content.coverImageId"
              :choose-cover="kind === 'teams' || kind === 'blog'"
              :disabled="!canEdit || busy"
              :label="kind === 'blog' ? 'Front image' : 'Images'"
              :max="kind === 'blog' ? 1 : 50"
            />
            <template v-if="kind === 'teams'"
              ><label class="form-field"
                >Team status<select v-model="selected.content.teamStatus">
                  <option value="active">Active</option>
                  <option value="retired">Retired</option>
                </select></label
              >
              <div class="section-title">
                <h3>Participation &amp; results</h3>
                <button
                  type="button"
                  class="button secondary"
                  @click="
                    selected.content.awards.push({
                      eventId: '',
                      league: bilingual(),
                      result: bilingual(),
                    })
                  "
                >
                  + Add result
                </button>
              </div>
              <p class="field-help">
                Create dated events in Participation History first. The same results appear with the
                team and the event.
              </p>
              <div
                v-for="(award, index) in selected.content.awards"
                :key="index"
                class="article-block"
              >
                <div class="section-title">
                  <h4>Result {{ index + 1 }}</h4>
                  <button
                    type="button"
                    class="icon-button danger-icon"
                    aria-label="Remove result"
                    @click="selected.content.awards.splice(index, 1)"
                  >
                    <AppIcon name="trash" :size="16" />
                  </button>
                </div>
                <label class="form-field"
                  >Event<select v-model="award.eventId" aria-label="Event" required>
                    <option value="" disabled>Select an event</option>
                    <option v-for="event in events" :key="event.id" :value="event.id">
                      {{ event.content.name.en || event.content.name.de }} · {{ event.content.date
                      }}{{ !event.published ? ' (draft)' : '' }}
                    </option>
                  </select></label
                ><LocalizedField
                  v-model="award.league"
                  label="League"
                  :maxlength="200"
                /><LocalizedField
                  v-model="award.result"
                  label="Result / award"
                  :maxlength="500"
                /></div
            ></template>
            <label v-if="kind === 'events'" class="form-field"
              >Event date<input v-model="selected.content.date" type="date" required /><span
                class="field-help"
                >The year comes from this date. Participation history is returned newest
                first.</span
              ></label
            >
            <label v-if="kind === 'sponsors' || kind === 'publications'" class="form-field"
              >External link (optional)<input
                v-model="selected.content.url"
                type="url"
                maxlength="2048"
                placeholder="https://…"
            /></label>
            <template v-if="kind === 'blog'"
              ><label class="form-field"
                >Publication date &amp; time<input
                  v-model="publicationTime"
                  type="datetime-local"
                /><span class="field-help"
                  >Your local time. Choose a future date to schedule a reviewed article; leave
                  Publish unchecked to keep a draft.</span
                ></label
              ><ArticleBlocks v-model="selected.content.blocks" :disabled="!canEdit || busy"
            /></template>
          </template>
        </fieldset>
      </form>
      <div v-else class="empty-state">
        <AppIcon name="edit" :size="32" />
        <h3>Select content to edit</h3>
        <p>Choose an entry or create a new one. New content starts as a draft.</p>
      </div>
    </div>
    <dialog
      ref="deleting"
      class="image-dialog"
      aria-labelledby="delete-content-title"
      @cancel="
        (event) => {
          if (busy) event.preventDefault()
        }
      "
    >
      <div class="picker-inner">
        <h2 id="delete-content-title">Delete this content?</h2>
        <p class="delete-description">
          Delete <strong>{{ selected?.content.name.en || labels[kind] }}</strong
          >? This removes it from the public website and cannot be undone. Library images are kept.
        </p>
        <div class="dialog-actions">
          <button class="button secondary" :disabled="busy" @click="deleting?.close()">
            Cancel</button
          ><button class="button danger" :disabled="busy" @click="remove">Delete content</button>
        </div>
      </div>
    </dialog>
  </div>
</template>
