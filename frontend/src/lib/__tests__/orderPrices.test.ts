import assert from 'node:assert/strict'
import { test } from 'node:test'
import { money, priceCents, priceInput, totalCents, unitPrice } from '../orderPrices.ts'

test('prices accept five EUR decimals and preserve them when reopened', () => {
  for (const [input, cents] of [
    ['0.00001', 0.001],
    ['0,00123', 0.123],
    ['1.99', 199],
    ['999999.99999', 99999999.999],
    ['1000000.00000', 100000000],
    ['0', 0],
  ] as const) {
    assert.equal(priceCents(input), cents)
    assert.equal(priceCents(priceInput(cents)), cents)
  }
  for (const invalid of ['0.000001', '-1', '1000000.00001', '1e-5', '', 'NaN']) {
    assert.throws(() => priceCents(invalid))
  }
  assert.equal(unitPrice(0.001), '0,00001 €')
  assert.equal(unitPrice(199), '1,99 €')
})

test('totals multiply and sum before rounding to cents', () => {
  assert.equal(money(totalCents([{ amount: 10000, unitPriceCents: 0.001 }])), '0,10 €')
  assert.equal(
    totalCents([
      { amount: 1, unitPriceCents: 0.249 },
      { amount: 1, unitPriceCents: 0.251 },
    ]),
    1,
  )
  assert.equal(totalCents([{ amount: 1000, unitPriceCents: 0.123 }]), 123)
  assert.equal(
    totalCents(
      Array.from({ length: 1000 }, () => ({
        amount: 10000,
        unitPriceCents: 99999999.999,
      })),
    ),
    999999999990000,
  )
})
