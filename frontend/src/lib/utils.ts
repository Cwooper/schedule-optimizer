import { clsx, type ClassValue } from "clsx"
import { twMerge } from "tailwind-merge"

export function cn(...inputs: ClassValue[]) {
  return twMerge(clsx(inputs))
}

/** Generate a random ID, with fallback for non-secure contexts (HTTP on mobile). */
export function genId(): string {
  if (crypto.randomUUID) return crypto.randomUUID()
  const bytes = crypto.getRandomValues(new Uint8Array(16))
  return Array.from(bytes, (b) => b.toString(16).padStart(2, "0")).join("")
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
