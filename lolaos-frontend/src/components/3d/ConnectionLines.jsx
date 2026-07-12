import { useRef, useMemo } from 'react'
import { useFrame } from '@react-three/fiber'
import * as THREE from 'three'

function ConnectionLines() {
  const linesRef = useRef([])

  const count = 12
  const spacing = 0.28

  const positions = useMemo(() => {
    const arr = []
    for (let i = 0; i < count; i++) {
      const y = (i - count/2) * spacing
      arr.push([-3.6, y, 0, 3.6, y, 0])
    }
    return arr
  }, [])

  useFrame((state) => {
    const t = state.clock.elapsedTime
    linesRef.current.forEach((line, i) => {
      if (!line?.material) return
      const phase = (i * 0.7 + t * 2.2) % (Math.PI * 2)
      line.material.opacity = 0.3 + Math.sin(phase) * 0.4
    })
  })

  return (
    <group>
      {positions.map((pos, i) => (
        <line key={i} ref={el => { linesRef.current[i] = el }}>
          <bufferGeometry>
            <float32BufferAttribute attach="attributes-position" args={[new Float32Array(pos), 3]} />
          </bufferGeometry>
          <lineBasicMaterial
            color="#a3a3a3"
            transparent
            opacity={0.4}
            linewidth={1.5}
          />
        </line>
      ))}
    </group>
  )
}

export default ConnectionLines