export interface Permission {
  id: number
  key: string
  description: string
}

export interface Role {
  id: number
  name: string
  description: string
  permissions?: Permission[]
}

export interface Member {
  id: number
  fusion_auth_user_id: string
  email: string
  name: string
  status: 'pending' | 'approved' | 'rejected'
  roles?: Role[]
}

export interface MeResponse {
  member: Member
  permissions: string[]
}

export interface InventoryCategory {
  id: number
  name: string
  parent_id?: number | null
  sort_order: number
}

export interface InventoryItem {
  id: number
  name: string
  quantity: number
  vendor_id: string
  product_url: string
  website: string
  notes: string
  category_id?: number | null
  category?: InventoryCategory
  confirmed: boolean
}

export interface OrderRequest {
  id: number
  name: string
  quantity: number
  unit_price_cents: number
  total_price_cents: number
  url: string
  notes: string
  shop_domain: string
  shop_name: string
  status: 'pending' | 'approved' | 'rejected'
  requester?: Member
}

export interface OrderListItem {
  id: number
  order_list_id: number
  name: string
  quantity: number
  unit_price_cents: number
  total_price_cents: number
  url: string
  notes: string
  shop_domain: string
  shop_name: string
  ordered: boolean
  received: boolean
}

export interface OrderList {
  id: number
  name: string
  status: 'draft' | 'published' | 'archived'
  published_at?: string | null
  items?: OrderListItem[]
}

export interface UploadedImage {
  id: number
  cloudflare_image_id: string
  filename: string
  content_type: string
  delivery_url: string
}

export interface Team {
  id: number
  name: string
  slug: string
  summary: string
  image_id?: number | null
  image?: UploadedImage
  published: boolean
  sort_order: number
}

export interface Competition {
  id: number
  name: string
  slug: string
  location: string
  starts_on?: string | null
  ends_on?: string | null
  published: boolean
  sort_order: number
}

export interface Prize {
  id: number
  title: string
  placement: string
  notes: string
  team_id: number
  competition_id: number
  team?: Team
  competition?: Competition
  sort_order: number
}

export interface SponsorCategory {
  id: number
  name: string
  sort_order: number
}

export interface Sponsor {
  id: number
  name: string
  category_id: number
  category?: SponsorCategory
  logo_image_id?: number | null
  logo_image?: UploadedImage
  website_url: string
  published: boolean
  sort_order: number
}

export interface HomeArticle {
  id: number
  slot: number
  title: string
  body: string
  image_id?: number | null
  image?: UploadedImage
  published: boolean
}

export interface AuthConfig {
  issuer: string
  client_id: string
  required_role: string
  dev_allow: boolean
}

export interface InventoryMatch {
  item: InventoryItem
  score: number
}
