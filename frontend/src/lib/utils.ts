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
 * Applies a single-pass entity decoder until the text is stable, mirroring the
 * backend's cleanText: Banner has sent double-escaped text, and &nbsp; becomes
 * a plain space.
 */
export function decodeRepeated(text: string, decodeOnce: (s: string) => string): string {
  for (let i = 0; i < 3; i++) {
    const decoded = decodeOnce(text)
    if (decoded === text) break
    text = decoded
  }
  return text.replaceAll("\u00a0", " ")
}

let _htmlDecoder: HTMLTextAreaElement | null = null
export function decodeHtmlEntities(html: string): string {
  return decodeRepeated(html, (s) => {
    _htmlDecoder ??= document.createElement("textarea")
    _htmlDecoder.innerHTML = s
    return _htmlDecoder.value
  })
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
