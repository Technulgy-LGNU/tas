import { computed } from 'vue'
import { apiFetch, currentUser } from './auth'
export const orderRoles = ['admin', 'order_admin', 'editor', 'order_request']
export const canViewOrders = computed(
  () => currentUser.value?.roles?.some((r) => orderRoles.includes(r)) ?? false,
)
export const manageOrders = computed(
  () => currentUser.value?.roles?.some((r) => ['admin', 'order_admin'].includes(r)) ?? false,
)
export const editOrders = computed(
  () => manageOrders.value || (currentUser.value?.roles?.includes('editor') ?? false),
)
export interface PartFields {
  name: string
  amount: number
  unitPriceCents: number
  shop: string
  link: string
  categoryId: string
}
export interface OrderPart extends PartFields {
  id: string
  createdBy: string
  createdByName: string
  createdAt: string
  requestId?: string
  orderedAt: string | null
  orderedBy: string
}
export interface PartRequest extends PartFields {
  id: string
  status: 'pending' | 'accepted' | 'rejected' | 'withdrawn'
  createdBy: string
  createdByName: string
  createdAt: string
  reviewedBy: string
  reviewedAt: string | null
  reviewNote: string
}
export interface Category {
  id: string
  name: string
}
export interface OrderList {
  id: string
  name: string
  status: 'open' | 'closed'
  currency: string
  version: number
  content: { categories: Category[]; parts: OrderPart[]; requests: PartRequest[] }
  createdAt: string
  updatedAt: string
  closedAt: string | null
}
export interface ListSummary {
  id: string
  name: string
  status: 'open' | 'closed'
  currency: string
  version: number
  partCount: number
  orderedCount: number
  pendingCount: number
  totalCents: number
  orderedTotalCents: number
  updatedAt: string
}
export interface StandardPart extends Omit<PartFields, 'categoryId'> {
  id: string
  version: number
}
export interface OrderStats {
  openLists: number
  closedLists: number
  pendingRequests: number
  remainingParts: number
  openTotalCents: number
  remainingTotalCents: number
  currency: string
}
export interface OrderAction {
  action: string
  targetId?: string
  name?: string
  part?: PartFields
  note?: string
  ordered?: boolean
}
export const emptyPart = (): PartFields => ({
  name: '',
  amount: 1,
  unitPriceCents: 0,
  shop: '',
  link: '',
  categoryId: '',
})
export function partFields(p: PartFields | StandardPart): PartFields {
  return {
    name: p.name,
    amount: p.amount,
    unitPriceCents: p.unitPriceCents,
    shop: p.shop,
    link: p.link,
    categoryId: 'categoryId' in p ? p.categoryId : '',
  }
}
export const money = (cents: number) =>
  new Intl.NumberFormat('de-DE', { style: 'currency', currency: 'EUR' }).format(cents / 100)
export function priceCents(value: string): number {
  if (!/^\d{1,7}(?:[.,]\d{1,2})?$/.test(value))
    throw new Error('Enter a unit price with at most two decimal places.')
  const [whole = '0', fraction = ''] = value.replace(',', '.').split('.')
  const cents = Number(whole) * 100 + Number(fraction.padEnd(2, '0'))
  if (cents > 100000000) throw new Error('Unit price cannot exceed 1,000,000 EUR.')
  return cents
}
export async function ordersRequest<T>(path: string, options: RequestInit = {}): Promise<T> {
  const response = await apiFetch(`/api/v1/orders${path}`, options)
  if (!response.ok) {
    const body = (await response.json().catch(() => null)) as { error?: string } | null
    throw new Error(body?.error || `Order request failed (${response.status}).`)
  }
  return response.status === 204 ? (undefined as T) : (response.json() as Promise<T>)
}
export const jsonOptions = (method: string, body: unknown): RequestInit => ({
  method,
  headers: { 'Content-Type': 'application/json' },
  body: JSON.stringify(body),
})
