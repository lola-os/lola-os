import { forwardRef } from 'react'
import { clsx } from 'clsx'
import { twMerge } from 'tailwind-merge'

const Section = forwardRef(
  ({ children, className, padding = 'default', background = 'default', fullWidth = false, id, ...props }, ref) => {
    const paddingStyles = {
      none: '',
      sm: 'py-8',
      default: 'py-16 md:py-24',
      lg: 'py-20 md:py-32',
    }

    const backgroundStyles = {
      default: 'bg-gray-50 dark:bg-gray-950',
      white: 'bg-white dark:bg-gray-900',
      dark: 'bg-gray-900 text-gray-50 dark:bg-gray-900',
      gradient: 'bg-gradient-to-b from-gray-50 to-white dark:from-gray-950 dark:to-gray-900',
    }

    const widthStyles = fullWidth ? '' : 'section-shell'

    return (
      <section
        ref={ref}
        id={id}
        className={twMerge(clsx(paddingStyles[padding], backgroundStyles[background], widthStyles, className))}
        {...props}
      >
        <div className={fullWidth ? '' : 'mx-auto max-w-7xl'}>{children}</div>
      </section>
    )
  }
)

Section.displayName = 'Section'

export default Section
