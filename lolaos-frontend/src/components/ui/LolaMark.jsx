/**
 * The LOLA OS mark: two parallel horizontal lines — AI (top) and
 * blockchain (bottom) — joined by a diagonal slash (LOLA, the bridge).
 * Per branding.md the mark is always grayscale, never glowed, gradient,
 * or shadowed. It draws with `currentColor` so it inherits the text
 * color of its container in both light and dark themes.
 */
function LolaMark({ className = '', title = 'LOLA OS' }) {
  return (
    <svg
      viewBox="0 0 48 48"
      fill="none"
      xmlns="http://www.w3.org/2000/svg"
      className={className}
      role="img"
      aria-label={title}
    >
      <line x1="7" y1="15" x2="30" y2="15" stroke="currentColor" strokeWidth="3.2" strokeLinecap="round" />
      <line x1="18" y1="33" x2="41" y2="33" stroke="currentColor" strokeWidth="3.2" strokeLinecap="round" />
      <line x1="15" y1="15" x2="33" y2="33" stroke="currentColor" strokeWidth="3.2" strokeLinecap="round" />
    </svg>
  )
}

export default LolaMark
