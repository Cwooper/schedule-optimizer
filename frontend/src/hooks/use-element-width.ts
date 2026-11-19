import { useCallback, useState } from "react"

/** A callback ref and the element's content width, kept current as it resizes. */
export function useElementWidth<T extends HTMLElement>(): [
  (el: T | null) => void,
  number,
] {
  const [width, setWidth] = useState(0)
  const ref = useCallback((el: T | null) => {
    if (!el) return
    const observer = new ResizeObserver(([entry]) =>
      setWidth(Math.floor(entry.contentRect.width))
    )
    observer.observe(el)
    return () => observer.disconnect()
  }, [])
  return [ref, width]
}
