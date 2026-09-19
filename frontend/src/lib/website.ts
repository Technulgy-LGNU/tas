import { reactive } from 'vue'
import { apiFetch } from './auth'
import MarkdownIt from 'markdown-it'
export interface Localized {
  de: string
  en: string
}
export interface ContentImage {
  id: string
  alt: Localized
}
export interface Video {
  title: Localized
  url: string
}
export interface Award {
  eventId: string
  league: Localized
  result: Localized
}
export interface Block {
  id: string
  type: 'text' | 'heading' | 'image' | 'gallery' | 'video'
  text: Localized
  level: number
  images: ContentImage[]
  url: string
}
export interface WebsiteContent {
  name: Localized
  description: Localized
  images: ContentImage[]
  coverImageId: string
  about: Localized
  videos: Video[]
  teamStatus: string
  awards: Award[]
  date: string
  url: string
  blocks: Block[]
}
export interface Entry {
  id: string
  kind: string
  slug: string
  published: boolean
  publishAt: string | null
  sortOrder: number
  version: number
  content: WebsiteContent
}
export interface LibraryImage {
  id: string
  name: string
  url: string
  status: string
}
export const bilingual = (): Localized => ({ de: '', en: '' })
export function newEntry(kind: string): Entry {
  return {
    id: '',
    kind,
    slug: kind === 'home' ? 'home' : '',
    published: false,
    publishAt: null,
    sortOrder: 0,
    version: 0,
    content: {
      name: bilingual(),
      description: bilingual(),
      images: [],
      coverImageId: '',
      about: bilingual(),
      videos: [],
      teamStatus: 'active',
      awards: [],
      date: '',
      url: '',
      blocks: [],
    },
  }
}
export function normalizeEntry(value: Entry): Entry {
  const entry = JSON.parse(JSON.stringify(value)) as Entry
  for (const key of ['images', 'videos', 'awards', 'blocks'] as const) entry.content[key] ??= []
  for (const block of entry.content.blocks) block.images ??= []
  return entry
}
export async function websiteRequest<T>(path: string, options: RequestInit = {}): Promise<T> {
  const response = await apiFetch(`/api/v1/website${path}`, options)
  if (!response.ok) {
    const data = (await response.json().catch(() => null)) as { error?: string } | null
    throw new Error(data?.error || 'Could not complete the request.')
  }
  if (response.status === 204) return undefined as T
  return response.json() as Promise<T>
}
const markdown = new MarkdownIt({ html: false, linkify: false, breaks: true })
// Images are managed by image/gallery blocks, so text blocks cannot embed remote images.
markdown.disable('image')
export const renderText = (text: string) => markdown.render(text)

export const libraryImages = reactive<Record<string, LibraryImage>>({})
const imageRequests = new Map<string, Promise<void>>()
export async function loadLibraryImage(id: string): Promise<void> {
  if (libraryImages[id]) return
  if (imageRequests.has(id)) return imageRequests.get(id)
  const request = (async () => {
    try {
      const response = await apiFetch(`/api/v1/images/${encodeURIComponent(id)}`)
      if (response.ok) {
        const data: { image: LibraryImage } = await response.json()
        libraryImages[id] = data.image
      }
    } catch {
      /* The picker will still show the reference if a preview is unavailable. */
    } finally {
      imageRequests.delete(id)
    }
  })()
  imageRequests.set(id, request)
  return request
}
