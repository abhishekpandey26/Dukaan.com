// Client-side search helpers.
// We load the products ONCE, then search / filter / paginate in the browser.
// No API call happens while the user types, so results update instantly.

// Keeps products whose name, description or category contains the search text
// (case-insensitive) and that match the selected category ('' = all).
export function filterProducts(products, search, category = '') {
  const query = search.trim().toLowerCase()
  return products.filter((p) => {
    if (category && p.category !== category) return false
    if (!query) return true
    return (
      p.name.toLowerCase().includes(query) ||
      p.description.toLowerCase().includes(query) ||
      p.category.toLowerCase().includes(query)
    )
  })
}

// Unique, sorted category names from a list of products.
export function getCategories(products) {
  return [...new Set(products.map((p) => p.category).filter(Boolean))].sort()
}

// Returns the items for one page (pages start at 1).
export function paginate(items, page, pageSize) {
  return items.slice((page - 1) * pageSize, page * pageSize)
}
