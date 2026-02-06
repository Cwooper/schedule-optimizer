import { useState, useEffect, useRef, useCallback } from "react"

/**
 * Local state that syncs to an external setter after a debounce delay.
 * Returns [localValue, setLocal, flush] — where flush() writes immediately.
 */
export function useDebouncedState<T>(
  storeValue: T,
  onSync: (value: T) => void,
  delay = 200
): [T, (value: T) => void, () => void] {
  const [local, setLocal] = useState(storeValue)
  const timerRef = useRef<ReturnType<typeof setTimeout> | null>(null)
  const latestRef = useRef(local)
  const dirtyRef = useRef(false)
  // Track what we last wrote to the store so we can detect external changes
  const lastSyncedRef = useRef(storeValue)

  // Must stay in sync during render so flush() reads the latest value
  // eslint-disable-next-line react-hooks/refs
  latestRef.current = local

  // Sync from store → local when store changes externally (e.g. clear filters).
  // If dirty, we compare against what we last synced — if storeValue differs
  // from our last write, another source changed it and we must accept it.
  useEffect(() => {
    if (!dirtyRef.current) {
      setLocal(storeValue)
    } else if (storeValue !== lastSyncedRef.current) {
      // External change while dirty — cancel pending debounce and accept store value
      if (timerRef.current) {
        clearTimeout(timerRef.current)
        timerRef.current = null
      }
      dirtyRef.current = false
      setLocal(storeValue)
    }
    lastSyncedRef.current = storeValue
  }, [storeValue])

  const setLocalAndSchedule = useCallback(
    (value: T) => {
      setLocal(value)
      dirtyRef.current = true
      latestRef.current = value
      if (timerRef.current) clearTimeout(timerRef.current)
      timerRef.current = setTimeout(() => {
        dirtyRef.current = false
        lastSyncedRef.current = value as T
        onSync(value)
      }, delay)
    },
    [onSync, delay]
  )

  const flush = useCallback(() => {
    if (timerRef.current) {
      clearTimeout(timerRef.current)
      timerRef.current = null
    }
    dirtyRef.current = false
    lastSyncedRef.current = latestRef.current
    onSync(latestRef.current)
  }, [onSync])

  // Cleanup on unmount
  useEffect(() => {
    return () => {
      if (timerRef.current) clearTimeout(timerRef.current)
      dirtyRef.current = false
    }
  }, [])

  return [local, setLocalAndSchedule, flush]
}
