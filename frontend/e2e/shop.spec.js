import { test, expect } from '@playwright/test'

// Screenshots go to e2e/screenshots so you can look at every screen.
const shot = (page, name, fullPage = true) =>
  page.screenshot({ path: `e2e/screenshots/${name}.png`, fullPage })

const unique = Date.now()
const customer = { name: 'Riya Test', email: `riya${unique}@example.com`, password: 'secret123' }
const admin = { email: 'admin@gocommerce.com', password: 'admin123' }

async function login(page, email, password) {
  await page.goto('/login')
  await page.getByLabel('Email').fill(email)
  await page.getByLabel('Password').fill(password)
  await page.getByRole('button', { name: 'Login' }).click()
}

// Reads the stock number shown on a product's detail page.
async function stockOf(page, productName) {
  await page.goto('/')
  await page.getByPlaceholder('Search products...').fill(productName)
  await page.locator('.product-name', { hasText: productName }).click()
  const text = await page.locator('.detail-info .stock').innerText()
  return Number(text.split(' ')[0])
}

test.describe.configure({ mode: 'serial' })

test('customer journey: browse, cart, checkout, cancel', async ({ page }) => {
  page.on('dialog', (d) => d.accept()) // auto-click "OK" on confirm() popups

  await test.step('home page lists products', async () => {
    await page.goto('/')
    await expect(page.locator('.product-card').first()).toBeVisible()
    expect(await page.locator('.product-card').count()).toBeGreaterThanOrEqual(8)
    await shot(page, '01-home')
  })

  await test.step('search runs in the browser (no API calls while typing)', async () => {
    const productRequests = []
    const track = (req) => { if (req.url().includes('/api/products')) productRequests.push(req.url()) }
    page.on('request', track)

    const box = page.getByPlaceholder('Search products...')
    await box.pressSequentially('lamp') // types one key at a time, like a user
    await expect(page.locator('.product-card')).toHaveCount(1)
    await expect(page.locator('.product-card')).toContainText('Desk Lamp')
    await expect(page.getByText('1 product found')).toBeVisible()

    await box.fill('LAMP') // case-insensitive
    await expect(page.locator('.product-card')).toHaveCount(1)

    await box.fill('battery') // matches the description of the headphones
    await expect(page.locator('.product-card')).toContainText('Wireless Headphones')

    await box.fill('xyz-nothing')
    await expect(page.locator('.product-card')).toHaveCount(0)
    await expect(page.getByText('No products match "xyz-nothing"')).toBeVisible()

    page.off('request', track)
    expect(productRequests).toEqual([])
  })

  await test.step('search and category filter work together', async () => {
    const box = page.getByPlaceholder('Search products...')
    await box.fill('shirt')
    await page.locator('.toolbar select').selectOption('Fashion')
    await expect(page.locator('.product-card')).toHaveCount(1)
    await page.locator('.toolbar select').selectOption('Home')
    await expect(page.locator('.product-card')).toHaveCount(0)
    await page.locator('.toolbar select').selectOption('')
    await page.getByPlaceholder('Search products...').fill('')
    await page.locator('.toolbar select').selectOption('Fashion')
    await expect(page.locator('.product-card')).toHaveCount(2)
    await shot(page, '02-filter-fashion')
  })

  await test.step('out-of-stock product cannot be added', async () => {
    await page.locator('.toolbar select').selectOption('Sports')
    const card = page.locator('.product-card', { hasText: 'Water Bottle' })
    await expect(card).toContainText('Out of stock')
    await expect(card.getByRole('button', { name: 'Add to cart' })).toBeDisabled()
  })

  await test.step('product detail page', async () => {
    await page.goto('/')
    await page.locator('.product-name', { hasText: 'Wireless Headphones' }).click()
    await expect(page.getByRole('heading', { name: 'Wireless Headphones' })).toBeVisible()
    await expect(page.locator('.detail-info .price')).toHaveText('$79.99')
    await shot(page, '03-product-detail')
  })

  await test.step('logged-out user is sent to login when adding to cart', async () => {
    await page.getByRole('button', { name: 'Add to cart' }).click()
    await expect(page).toHaveURL(/\/login/)
  })

  await test.step('cart page requires login', async () => {
    await page.goto('/cart')
    await expect(page).toHaveURL(/\/login/)
  })

  await test.step('register a new customer', async () => {
    await page.goto('/register')
    await page.getByLabel('Name').fill(customer.name)
    await page.getByLabel('Email').fill(customer.email)
    await page.getByLabel('Password').fill(customer.password)
    await shot(page, '04-register')
    await page.getByRole('button', { name: 'Sign up' }).click()
    await expect(page.getByText(`Hi, ${customer.name}`)).toBeVisible()
  })

  await test.step('duplicate email is rejected', async () => {
    await page.getByRole('button', { name: 'Logout' }).click()
    await page.goto('/register')
    await page.getByLabel('Name').fill('Someone')
    await page.getByLabel('Email').fill(customer.email)
    await page.getByLabel('Password').fill('another1')
    await page.getByRole('button', { name: 'Sign up' }).click()
    await expect(page.locator('.message-error')).toHaveText('email is already registered')
  })

  await test.step('wrong password is rejected, right one works', async () => {
    await login(page, customer.email, 'wrong-password')
    await expect(page.locator('.message-error')).toHaveText('invalid email or password')
    await login(page, customer.email, customer.password)
    await expect(page.getByText(`Hi, ${customer.name}`)).toBeVisible()
  })

  let stockBefore
  await test.step('add items to cart, badge updates', async () => {
    stockBefore = await stockOf(page, 'Coffee Mug')
    await page.goto('/')
    for (const name of ['Coffee Mug', 'Cotton T-Shirt']) {
      await page.locator('.product-card', { hasText: name }).getByRole('button', { name: 'Add to cart' }).click()
      await expect(page.locator('.message-success')).toHaveText('Added to cart!')
    }
    await expect(page.locator('.navbar .badge')).toHaveText('2')
  })

  await test.step('cart: change quantity, totals update', async () => {
    await page.getByRole('link', { name: /Cart/ }).click()
    const mugRow = page.locator('tr', { hasText: 'Coffee Mug' })
    await mugRow.getByRole('button', { name: '+' }).click()
    await expect(mugRow.locator('.qty span')).toHaveText('2')
    // 2 x $8.99 + 1 x $12.99 = $30.97
    await expect(page.locator('.cart-footer .price')).toHaveText('Total: $30.97')
    await expect(page.locator('.navbar .badge')).toHaveText('3')
    await shot(page, '05-cart')
  })

  await test.step('remove an item', async () => {
    await page.locator('tr', { hasText: 'Cotton T-Shirt' }).getByRole('button', { name: 'Remove' }).click()
    await expect(page.locator('tr', { hasText: 'Cotton T-Shirt' })).toHaveCount(0)
    await expect(page.locator('.cart-footer .price')).toHaveText('Total: $17.98')
  })

  await test.step('checkout creates order and reduces stock', async () => {
    await page.getByRole('button', { name: 'Place order' }).click()
    await expect(page).toHaveURL(/\/orders/)
    await expect(page.locator('.message-success')).toContainText('placed successfully')
    await expect(page.locator('.order').first()).toContainText('pending')
    await expect(page.locator('.navbar .badge')).toHaveCount(0) // cart emptied
    await shot(page, '06-orders-after-checkout')
    expect(await stockOf(page, 'Coffee Mug')).toBe(stockBefore - 2)
  })

  await test.step('cancel order puts stock back', async () => {
    await page.goto('/orders')
    await page.locator('.order').first().getByRole('button', { name: 'Cancel' }).click()
    await expect(page.locator('.order').first().locator('.status')).toHaveText('cancelled')
    await shot(page, '07-order-cancelled')
    expect(await stockOf(page, 'Coffee Mug')).toBe(stockBefore)
  })

  await test.step('place a second order (left pending for the admin)', async () => {
    await page.goto('/')
    await page.locator('.product-card', { hasText: 'Mechanical Keyboard' }).getByRole('button', { name: 'Add to cart' }).click()
    await expect(page.locator('.navbar .badge')).toHaveText('1')
    await page.goto('/cart')
    await page.getByRole('button', { name: 'Place order' }).click()
    await expect(page).toHaveURL(/\/orders/)
  })

  await test.step('customer cannot open admin pages', async () => {
    await expect(page.getByRole('link', { name: 'Admin' })).toHaveCount(0)
    await page.goto('/admin')
    await expect(page).toHaveURL('/')
  })
})

test('admin: dashboard, product CRUD, orders, users', async ({ page }) => {
  page.on('dialog', (d) => d.accept())
  const productName = `E2E Test Product ${unique}`

  await test.step('admin login goes to dashboard', async () => {
    await login(page, admin.email, admin.password)
    await expect(page).toHaveURL(/\/admin$/)
    await expect(page.getByRole('heading', { name: 'Dashboard' })).toBeVisible()
    await expect(page.locator('.stat').first()).toBeVisible()
    await shot(page, '08-admin-dashboard')
  })

  await test.step('create a product', async () => {
    await page.locator('.admin-sidebar').getByRole('link', { name: 'Products' }).click()
    await shot(page, '09-admin-products')
    await page.getByRole('button', { name: '+ Add product' }).click()
    await page.getByLabel('Name').fill(productName)
    await page.getByLabel('Description').fill('Created by the Playwright test.')
    await page.getByLabel('Price ($)').fill('19.50')
    await page.getByLabel('Stock').fill('7')
    await page.getByLabel('Category').fill('Testing')
    await shot(page, '10-admin-product-form')
    await page.getByRole('button', { name: 'Save' }).click()
    await expect(page.locator('.message-success')).toContainText('created')
    const row = page.locator('tr', { hasText: productName })
    await expect(row).toContainText('$19.50')
    await expect(row).toContainText('7')
  })

  await test.step('new product appears in the shop (cache was cleared)', async () => {
    await page.goto('/')
    await page.getByPlaceholder('Search products...').fill(productName)
    await expect(page.locator('.product-card', { hasText: productName })).toContainText('$19.50')
  })

  await test.step('edit the product', async () => {
    await page.goto('/admin/products')
    await page.locator('tr', { hasText: productName }).getByRole('button', { name: 'Edit' }).click()
    await expect(page.getByLabel('Price ($)')).toHaveValue('19.50')
    await page.getByLabel('Price ($)').fill('24.00')
    await page.getByLabel('Stock').fill('3')
    await page.getByRole('button', { name: 'Save' }).click()
    await expect(page.locator('.message-success')).toContainText('updated')
    await expect(page.locator('tr', { hasText: productName })).toContainText('$24.00')
  })

  await test.step('invalid product is rejected', async () => {
    await page.getByRole('button', { name: '+ Add product' }).click()
    await page.getByLabel('Name').fill('   ')
    await page.getByLabel('Price ($)').fill('5')
    await page.getByLabel('Stock').fill('1')
    await page.getByRole('button', { name: 'Save' }).click()
    await expect(page.locator('.message-error')).toHaveText('name is required')
    await page.getByRole('button', { name: 'Cancel' }).click()
  })

  await test.step('delete the product', async () => {
    await page.locator('tr', { hasText: productName }).getByRole('button', { name: 'Delete' }).click()
    await expect(page.locator('.message-success')).toContainText('deleted')
    await expect(page.locator('tr', { hasText: productName })).toHaveCount(0)
  })

  await test.step('change order status pending -> shipped', async () => {
    await page.locator('.admin-sidebar').getByRole('link', { name: 'Orders' }).click()
    const order = page.locator('.order', { hasText: customer.email }).filter({ hasText: 'Mechanical Keyboard' }).first()
    await order.locator('select').selectOption('shipped')
    await expect(page.locator('.message-success')).toContainText('"shipped"')
    await expect(order.locator('select')).toHaveValue('shipped')
    await shot(page, '11-admin-orders')
  })

  await test.step('cancelled orders are locked', async () => {
    const cancelled = page.locator('.order', { hasText: customer.email }).filter({ hasText: 'Coffee Mug' }).first()
    await expect(cancelled.locator('select')).toBeDisabled()
  })

  await test.step('status filter', async () => {
    await page.locator('.row select').first().selectOption('cancelled')
    // Only cancelled orders have a disabled dropdown, so after filtering
    // there must be at least one order and zero enabled dropdowns.
    await expect(page.locator('.order select:enabled')).toHaveCount(0)
    await expect(page.locator('.order').first()).toBeVisible()
  })

  await test.step('users list', async () => {
    await page.locator('.admin-sidebar').getByRole('link', { name: 'Users' }).click()
    await expect(page.locator('tr', { hasText: customer.email })).toContainText('customer')
    await expect(page.locator('tr', { hasText: admin.email })).toContainText('admin')
    await shot(page, '12-admin-users')
  })
})

test('mobile layout', async ({ browser }) => {
  const context = await browser.newContext({ viewport: { width: 390, height: 844 } })
  const page = await context.newPage()
  await page.goto('/')
  await expect(page.locator('.product-card').first()).toBeVisible()
  // no horizontal scrolling on a phone
  const overflow = await page.evaluate(() => document.documentElement.scrollWidth > window.innerWidth)
  expect(overflow).toBe(false)
  await shot(page, '13-mobile-home', false)
  await context.close()
})
