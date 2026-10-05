export const money = (cents: number) =>
  new Intl.NumberFormat('de-DE', { style: 'currency', currency: 'EUR' }).format(cents / 100)

export const unitPrice = (cents: number) =>
  new Intl.NumberFormat('de-DE', {
    style: 'currency',
    currency: 'EUR',
    minimumFractionDigits: 2,
    maximumFractionDigits: 5,
  }).format(cents / 100)

export const priceInput = (cents: number) => (cents / 100).toFixed(5)

export function priceCents(value: string): number {
  if (!/^\d{1,7}(?:[.,]\d{1,5})?$/.test(value))
    throw new Error('Enter a unit price with at most five decimal places.')
  const [whole = '0', fraction = ''] = value.replace(',', '.').split('.')
  const units = Number(whole) * 100000 + Number(fraction.padEnd(5, '0'))
  if (units > 100000000000) throw new Error('Unit price cannot exceed 1,000,000 EUR.')
  return units / 1000
}

// Sum exact thousandths of cents before rounding the displayed total to cents.
// BigInt also keeps maximum-size lists within exact integer arithmetic.
export function totalCents(parts: { amount: number; unitPriceCents: number }[]): number {
  const units = parts.reduce(
    (sum, p) => sum + BigInt(p.amount) * BigInt(Math.round(p.unitPriceCents * 1000)),
    0n,
  )
  return Number((units + 500n) / 1000n)
}
