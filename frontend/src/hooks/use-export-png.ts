import { useCallback, useState, type RefObject } from "react"

export function useExportPng(
  gridRef: RefObject<HTMLDivElement | null>,
  termName: string | undefined
) {
  const [isExporting, setIsExporting] = useState(false)

  const handleDownloadPng = useCallback(async () => {
    const node = gridRef.current
    if (!node) return

    setIsExporting(true)
    try {
      const EXPORT_WIDTH = 1200
      const bgColor = getComputedStyle(document.body).backgroundColor

      // Temporarily resize the node so the browser re-layouts at export width
      const savedCssText = node.style.cssText
      Object.assign(node.style, {
        width: `${EXPORT_WIDTH}px`,
        height: "auto",
        overflow: "visible",
        flex: "none",
      })

      const { toPng } = await import("html-to-image")
      const dataUrl = await toPng(node, {
        pixelRatio: 2,
        backgroundColor: bgColor,
      })

      // Restore original styles
      node.style.cssText = savedCssText

      const blob = await (await fetch(dataUrl)).blob()
      const fileName = termName
        ? `${termName.replace(/\s+/g, "")}-Schedule.png`
        : "Schedule.png"

      // Try Web Share API first (requires secure context / HTTPS)
      if (navigator.share) {
        try {
          const file = new File([blob], fileName, { type: "image/png" })
          await navigator.share({ files: [file], title: "My Schedule" })
          return
        } catch (err) {
          if (err instanceof Error && err.name === "AbortError") return
          // Other errors (NotAllowedError, TypeError, etc.): fall through to download
        }
      }

      // Fallback: standard file download
      const link = document.createElement("a")
      link.download = fileName
      link.href = URL.createObjectURL(blob)
      link.click()
      URL.revokeObjectURL(link.href)
    } catch (err) {
      if (err instanceof Error && err.name !== "AbortError") {
        console.error("PNG export failed:", err)
      }
    } finally {
      setIsExporting(false)
    }
  }, [gridRef, termName])

  return { isExporting, handleDownloadPng }
}
