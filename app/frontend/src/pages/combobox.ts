import { useEffect, useId, useRef, useState } from 'react'

/**
 * The open/close mechanics shared by the aircraft picker and the pilot picker.
 *
 * Extracted when the second picker arrived (Task 21) rather than guessed at
 * when the first one was written. What is here is only the part that is
 * genuinely identical and genuinely fiddly — the two lists differ in what they
 * show, what they filter on and what "add a new one" means, and none of that
 * belongs in a shared hook.
 *
 * CLOSING ON AN OUTSIDE CLICK, NOT ON BLUR. Blur fires before the click on an
 * option lands, so a blur-driven close takes the list away from under the
 * finger that was choosing — on a phone, the only symptom is that tapping an
 * option does nothing.
 */
export function useCombobox() {
  const [open, setOpen] = useState(false)
  const wrapRef = useRef<HTMLDivElement>(null)
  const listId = useId()

  useEffect(() => {
    if (!open) return
    function onDown(e: MouseEvent | TouchEvent) {
      if (!wrapRef.current?.contains(e.target as Node)) setOpen(false)
    }
    document.addEventListener('mousedown', onDown)
    document.addEventListener('touchstart', onDown)
    return () => {
      document.removeEventListener('mousedown', onDown)
      document.removeEventListener('touchstart', onDown)
    }
  }, [open])

  return { open, setOpen, wrapRef, listId }
}
