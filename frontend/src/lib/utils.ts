import { clsx, type ClassValue } from "clsx"
import { twMerge } from "tailwind-merge"

export function cn(...inputs: ClassValue[]) {
  return twMerge(clsx(inputs))
}

/**
 * Decodes HTML entities in a string (e.g., "&amp;" -> "&", "&nbsp;" -> " ")
 * Uses a textarea element to leverage the browser's built-in decoding.
 * Reuses a single element for performance.
 */
const _htmlDecoder = document.createElement("textarea")
export function decodeHtmlEntities(html: string): string {
  _htmlDecoder.innerHTML = html
  return _htmlDecoder.value
}
