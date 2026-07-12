import { forwardRef } from 'react'
import { clsx } from 'clsx'
import { twMerge } from 'tailwind-merge'

const Card = forwardRef(({ children, className, variant = 'default', hover = false, ...props }, ref) => {
  const baseStyles = 'rounded-2xl transition-all duration-200'

  const variants = {
    default: 'bg-white border border-gray-200 dark:bg-gray-900 dark:border-gray-800',
    elevated: 'bg-white border border-gray-200 shadow-sm dark:bg-gray-900 dark:border-gray-800',
    subtle: 'bg-gray-100/60 border border-gray-200 dark:bg-gray-900/60 dark:border-gray-800',
    dark: 'bg-gray-900 border border-gray-800 text-gray-50',
  }

  const hoverStyles = hover
    ? 'hover:-translate-y-1 hover:border-gray-300 hover:shadow-lg dark:hover:border-gray-700 dark:hover:shadow-black/40'
    : ''

  return (
    <div ref={ref} className={twMerge(clsx(baseStyles, variants[variant], hoverStyles, className))} {...props}>
      {children}
    </div>
  )
})

Card.displayName = 'Card'

export default Card
