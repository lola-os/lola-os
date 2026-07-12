import { forwardRef } from 'react'
import { clsx } from 'clsx'
import { twMerge } from 'tailwind-merge'

const Button = forwardRef(
  ({ children, className, variant = 'primary', size = 'md', loading = false, disabled = false, ...props }, ref) => {
    const baseStyles =
      'inline-flex items-center justify-center font-medium rounded-lg transition-all duration-200 focus:outline-none focus-visible:ring-2 focus-visible:ring-offset-2 focus-visible:ring-gray-900 dark:focus-visible:ring-gray-100 dark:focus-visible:ring-offset-gray-950 disabled:opacity-50 disabled:cursor-not-allowed active:scale-[0.98]'

    const variants = {
      primary:
        'bg-gray-900 text-gray-50 hover:bg-gray-800 dark:bg-gray-50 dark:text-gray-950 dark:hover:bg-gray-200',
      secondary:
        'bg-gray-100 text-gray-900 hover:bg-gray-200 dark:bg-gray-800 dark:text-gray-100 dark:hover:bg-gray-700',
      outline:
        'border border-gray-300 text-gray-800 hover:bg-gray-100 dark:border-gray-700 dark:text-gray-200 dark:hover:bg-gray-800',
      ghost: 'text-gray-700 hover:bg-gray-100 dark:text-gray-300 dark:hover:bg-gray-800',
    }

    const sizes = {
      sm: 'px-3 py-1.5 text-sm gap-1.5',
      md: 'px-4 py-2 text-sm gap-2',
      lg: 'px-5 py-2.5 text-base gap-2',
      xl: 'px-7 py-3.5 text-base gap-2.5',
    }

    return (
      <button
        ref={ref}
        className={twMerge(clsx(baseStyles, variants[variant], sizes[size], loading && 'cursor-wait', className))}
        disabled={disabled || loading}
        {...props}
      >
        {loading && (
          <svg className="-ml-1 h-4 w-4 animate-spin" fill="none" viewBox="0 0 24 24" aria-hidden="true">
            <circle className="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" strokeWidth="4" />
            <path
              className="opacity-75"
              fill="currentColor"
              d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"
            />
          </svg>
        )}
        {children}
      </button>
    )
  }
)

Button.displayName = 'Button'

export default Button
