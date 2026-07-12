import { motion } from 'framer-motion'

/**
 * Shared section header: small eyebrow label, a title, and an optional
 * lede. Fades/rises into view on scroll (and respects reduced motion via
 * the global CSS override).
 */
function SectionHeading({ eyebrow, title, description, align = 'center', className = '' }) {
  const alignment = align === 'left' ? 'text-left items-start' : 'text-center items-center mx-auto'

  return (
    <div className={`flex max-w-2xl flex-col gap-4 ${alignment} ${className}`}>
      {eyebrow && (
        <motion.span
          initial={{ opacity: 0, y: 12 }}
          whileInView={{ opacity: 1, y: 0 }}
          viewport={{ once: true, amount: 0.6 }}
          transition={{ duration: 0.5 }}
          className="eyebrow"
        >
          {eyebrow}
        </motion.span>
      )}
      <motion.h2
        initial={{ opacity: 0, y: 12 }}
        whileInView={{ opacity: 1, y: 0 }}
        viewport={{ once: true, amount: 0.6 }}
        transition={{ duration: 0.5, delay: 0.05 }}
        className="text-balance text-3xl font-bold tracking-tight text-gray-900 dark:text-gray-50 sm:text-4xl"
      >
        {title}
      </motion.h2>
      {description && (
        <motion.p
          initial={{ opacity: 0, y: 12 }}
          whileInView={{ opacity: 1, y: 0 }}
          viewport={{ once: true, amount: 0.6 }}
          transition={{ duration: 0.5, delay: 0.1 }}
          className="text-balance text-lg leading-relaxed text-gray-600 dark:text-gray-400"
        >
          {description}
        </motion.p>
      )}
    </div>
  )
}

export default SectionHeading
