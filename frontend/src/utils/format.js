// The backend stores money as integer cents. Convert only when displaying.
// Change 'USD' / 'en-US' to 'INR' / 'en-IN' if you want rupees.
const formatter = new Intl.NumberFormat('en-US', { style: 'currency', currency: 'USD' })

export const formatPrice = (cents) => formatter.format(cents / 100)

export const formatDate = (iso) => new Date(iso).toLocaleString()

export const ORDER_STATUSES = ['pending', 'paid', 'shipped', 'delivered', 'cancelled']
