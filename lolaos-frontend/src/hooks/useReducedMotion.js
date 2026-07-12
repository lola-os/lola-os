import { useEffect, useState } from 'react'

/**
 * Tracks the user's `prefers-reduced-motion` setting so components can
 * skip or soften animation. Defaults to `true` (reduced) during SSR / the
 * first paint to avoid animating before we know the preference.
 */
function useReducedMotion() {
  const [reduced, setReduced] = useState(() => {
    if (typeof window === 'undefined' || !window.matchMedia) return true
    return window.matchMedia('(prefers-reduced-motion: reduce)').matches
  })

  useEffect(() => {
    const media = window.matchMedia('(prefers-reduced-motion: reduce)')
    const update = () => setReduced(media.matches)
    update()
    media.addEventListener('change', update)
    return () => media.removeEventListener('change', update)
  }, [])

  return reduced
}

export default useReducedMotion
