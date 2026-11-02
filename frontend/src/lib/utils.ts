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
 * Decodes HTML entities via the browser's parser. Mirrors the backend's
 * cleanText: repeats for double-escaped text and turns &nbsp; into a space.
 */
let _htmlDecoder: HTMLTextAreaElement | null = null
export function decodeHtmlEntities(html: string): string {
  _htmlDecoder ??= document.createElement("textarea")
  for (let i = 0; i < 3; i++) {
    _htmlDecoder.innerHTML = html
    const decoded = _htmlDecoder.value
    if (decoded === html) break
    html = decoded
  }
  return html.replaceAll("\u00a0", " ")
}

/** Format a 6-digit term code to a readable name (e.g., "202620" → "Spring 2026") */
export function formatTermCode(code: string): string {
  if (code.length !== 6) return code
  const year = code.slice(0, 4)
  const quarter = code.slice(4)
  const quarterNames: Record<string, string> = {
    "10": "Winter",
    "20": "Spring",
    "30": "Summer",
    "40": "Fall",
  }
  const name = quarterNames[quarter]
  return name ? `${name} '${year.slice(2)}` : code
}
