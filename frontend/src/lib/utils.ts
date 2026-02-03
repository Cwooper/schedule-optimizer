import { clsx, type ClassValue } from "clsx"
import { twMerge } from "tailwind-merge"

export function cn(...inputs: ClassValue[]) {
  return twMerge(clsx(inputs))
}

/**
 * Decodes HTML entities in a string (e.g., "&amp;" -> "&", "&nbsp;" -> " ")
 * Uses a textarea element to leverage the browser's built-in decoding.
 * Lazily initialized and reused for performance.
 */
let _htmlDecoder: HTMLTextAreaElement | null = null
export function decodeHtmlEntities(html: string): string {
  if (!_htmlDecoder) {
    _htmlDecoder = document.createElement("textarea")
  }
  _htmlDecoder.innerHTML = html
  return _htmlDecoder.value
}
