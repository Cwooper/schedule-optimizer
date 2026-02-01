import { clsx, type ClassValue } from "clsx"
import { twMerge } from "tailwind-merge"

export function cn(...inputs: ClassValue[]) {
  return twMerge(clsx(inputs))
}

/**
 * Decodes HTML entities in a string (e.g., "&amp;" -> "&", "&nbsp;" -> " ")
 * Uses a textarea element to leverage the browser's built-in decoding.
 */
export function decodeHtmlEntities(html: string): string {
  const txt = document.createElement("textarea")
  txt.innerHTML = html
  return txt.value
}
