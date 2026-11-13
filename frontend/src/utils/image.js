// If an image URL is broken (or the image site is down), show our local placeholder.
export const PLACEHOLDER_IMAGE = '/images/placeholder.svg'

export function onImageError(e) {
  if (!e.target.src.endsWith(PLACEHOLDER_IMAGE)) e.target.src = PLACEHOLDER_IMAGE
}
