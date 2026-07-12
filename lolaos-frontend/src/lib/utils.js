/**
 * Format a number with commas
 */
export function formatNumber(num) {
  if (typeof num !== 'number') return '0'
  return num.toLocaleString('en-US')
}

/**
 * Debounce function
 */
export function debounce(func, wait) {
  let timeout
  return function executedFunction(...args) {
    const later = () => {
      clearTimeout(timeout)
      func(...args)
    }
    clearTimeout(timeout)
    timeout = setTimeout(later, wait)
  }
}

/**
 * Throttle function
 */
export function throttle(func, limit) {
  let inThrottle
  return function() {
    const args = arguments
    const context = this
    if (!inThrottle) {
      func.apply(context, args)
      inThrottle = true
      setTimeout(() => inThrottle = false, limit)
    }
  }
}

/**
 * Generate unique ID
 */
export function generateId() {
  return Math.random().toString(36).substr(2, 9)
}

/**
 * Validate email address
 */
export function validateEmail(email) {
  const re = /^[^\s@]+@[^\s@]+\.[^\s@]+$/
  return re.test(email)
}

/**
 * Sleep/pause execution
 */
export function sleep(ms) {
  return new Promise(resolve => setTimeout(resolve, ms))
}

/**
 * Copy text to clipboard
 */
export async function copyToClipboard(text) {
  try {
    await navigator.clipboard.writeText(text)
    return true
  } catch (err) {
    console.error('Failed to copy text:', err)
    return false
  }
}

/**
 * Format file size
 */
export function formatFileSize(bytes) {
  if (bytes === 0) return '0 Bytes'
  
  const k = 1024
  const sizes = ['Bytes', 'KB', 'MB', 'GB', 'TB']
  const i = Math.floor(Math.log(bytes) / Math.log(k))
  
  return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i]
}

/**
 * Get current year
 */
export function getCurrentYear() {
  return new Date().getFullYear()
}

/**
 * Check if running on mobile
 */
export function isMobile() {
  return /Android|webOS|iPhone|iPad|iPod|BlackBerry|IEMobile|Opera Mini/i.test(navigator.userAgent)
}

/**
 * Check if running on touch device
 */
export function isTouchDevice() {
  return 'ontouchstart' in window || navigator.maxTouchPoints > 0
}

/**
 * Parse query parameters
 */
export function getQueryParams() {
  const params = {}
  const queryString = window.location.search.substring(1)
  const pairs = queryString.split('&')
  
  pairs.forEach(pair => {
    const [key, value] = pair.split('=')
    if (key) {
      params[decodeURIComponent(key)] = decodeURIComponent(value || '')
    }
  })
  
  return params
}

/**
 * Smooth scroll to element
 */
export function smoothScrollTo(elementId, offset = 0) {
  const element = document.getElementById(elementId)
  if (element) {
    const elementPosition = element.getBoundingClientRect().top
    const offsetPosition = elementPosition + window.pageYOffset - offset
    
    window.scrollTo({
      top: offsetPosition,
      behavior: 'smooth'
    })
  }
}

/**
 * Detect prefers reduced motion
 */
export function prefersReducedMotion() {
  return window.matchMedia('(prefers-reduced-motion: reduce)').matches
}

/**
 * Truncate text with ellipsis
 */
export function truncateText(text, maxLength) {
  if (text.length <= maxLength) return text
  return text.substring(0, maxLength) + '...'
}

/**
 * Capitalize first letter
 */
export function capitalize(text) {
  if (!text) return ''
  return text.charAt(0).toUpperCase() + text.slice(1)
}

/**
 * Remove HTML tags from string
 */
export function stripHtml(html) {
  const tmp = document.createElement('div')
  tmp.innerHTML = html
  return tmp.textContent || tmp.innerText || ''
}

/**
 * Format date
 */
export function formatDate(date, format = 'medium') {
  const d = new Date(date)
  
  if (format === 'short') {
    return d.toLocaleDateString('en-US', {
      month: 'short',
      day: 'numeric',
      year: 'numeric'
    })
  }
  
  if (format === 'long') {
    return d.toLocaleDateString('en-US', {
      weekday: 'long',
      year: 'numeric',
      month: 'long',
      day: 'numeric'
    })
  }
  
  // medium (default)
  return d.toLocaleDateString('en-US', {
    year: 'numeric',
    month: 'short',
    day: 'numeric'
  })
}

/**
 * Calculate percentage
 */
export function calculatePercentage(part, total) {
  if (total === 0) return 0
  return Math.round((part / total) * 100)
}

/**
 * Check if element is in viewport
 */
export function isInViewport(element) {
  if (!element) return false
  
  const rect = element.getBoundingClientRect()
  return (
    rect.top >= 0 &&
    rect.left >= 0 &&
    rect.bottom <= (window.innerHeight || document.documentElement.clientHeight) &&
    rect.right <= (window.innerWidth || document.documentElement.clientWidth)
  )
}

/**
 * Download file
 */
export function downloadFile(content, filename, type = 'text/plain') {
  const blob = new Blob([content], { type })
  const url = URL.createObjectURL(blob)
  const a = document.createElement('a')
  
  a.href = url
  a.download = filename
  document.body.appendChild(a)
  a.click()
  document.body.removeChild(a)
  URL.revokeObjectURL(url)
}