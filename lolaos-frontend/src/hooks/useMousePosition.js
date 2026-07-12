import { useState, useEffect } from 'react'

function useMousePosition() {
  const [position, setPosition] = useState({ x: 0, y: 0 })
  const [normalizedPosition, setNormalizedPosition] = useState({ x: 0, y: 0 })
  const [isMoving, setIsMoving] = useState(false)

  useEffect(() => {
    let movementTimer = null

    const handleMouseMove = (event) => {
      const x = event.clientX
      const y = event.clientY
      
      setPosition({ x, y })
      
      // Normalize to -1 to 1 range for 3D effects
      const normalizedX = (x / window.innerWidth) * 2 - 1
      const normalizedY = (y / window.innerHeight) * 2 - 1
      setNormalizedPosition({ x: normalizedX, y: normalizedY })
      
      // Set moving state
      setIsMoving(true)
      
      // Clear existing timer
      if (movementTimer) clearTimeout(movementTimer)
      
      // Set timer to reset moving state
      movementTimer = setTimeout(() => {
        setIsMoving(false)
      }, 100)
    }

    window.addEventListener('mousemove', handleMouseMove)
    
    return () => {
      window.removeEventListener('mousemove', handleMouseMove)
      if (movementTimer) clearTimeout(movementTimer)
    }
  }, [])

  return {
    ...position,
    normalizedX: normalizedPosition.x,
    normalizedY: normalizedPosition.y,
    isMoving,
  }
}

export default useMousePosition