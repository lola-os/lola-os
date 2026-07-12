import { useRef, useMemo } from 'react'
import { useFrame } from '@react-three/fiber'
import { Instances, Instance } from '@react-three/drei'
import * as THREE from 'three'

function NetworkNodes() {
  const nodesRef = useRef([])
  const linesRef = useRef([])
  const groupRef = useRef()
  
  // Memoized node positions
  const nodeData = useMemo(() => {
    return Array.from({ length: 12 }).map((_, i) => {
      const angle = (i / 12) * Math.PI * 2
      const radius = 3
      return {
        x: Math.cos(angle) * radius,
        y: Math.sin(angle) * 1.5,
        z: Math.sin(angle) * radius,
        angle: angle,
        radius: radius
      }
    })
  }, [])
  
  // Create line geometries once
  const lineGeometries = useMemo(() => {
    return Array.from({ length: 6 }).map((_, i) => {
      const startAngle = (i / 6) * Math.PI * 2
      const endAngle = ((i + 3) / 6) * Math.PI * 2
      
      const startX = Math.cos(startAngle) * 3
      const startY = Math.sin(startAngle) * 1.5
      const startZ = Math.sin(startAngle) * 3
      
      const endX = Math.cos(endAngle) * 3
      const endY = Math.sin(endAngle) * 1.5
      const endZ = Math.sin(endAngle) * 3
      
      const geom = new THREE.BufferGeometry()
      geom.setAttribute(
        'position',
        new THREE.Float32BufferAttribute([startX, startY, startZ, endX, endY, endZ], 3)
      )
      return geom
    })
  }, [])
  
  useFrame((state) => {
    const time = state.clock.elapsedTime
    
    // Animate nodes
    nodesRef.current.forEach((node, i) => {
      if (node) {
        const data = nodeData[i]
        const radius = data.radius + Math.sin(time * 0.5 + i) * 0.3
        
        node.position.x = Math.cos(time * 0.2 + data.angle) * radius
        node.position.y = Math.sin(time * 0.3 + data.angle) * 1.5
        node.position.z = Math.sin(time * 0.2 + data.angle) * radius
        
        node.rotation.x = time * 0.1
        node.rotation.y = time * 0.15
        
        // Scale animation
        const scale = 0.8 + Math.sin(time * 2 + i) * 0.2
        node.scale.setScalar(scale)
      }
    })
    
    // Animate lines
    linesRef.current.forEach((line, i) => {
      if (line) {
        const startIndex = i
        const endIndex = (i + 3) % 12
        const startNode = nodesRef.current[startIndex]
        const endNode = nodesRef.current[endIndex]
        
        if (startNode && endNode) {
          const positions = line.geometry.attributes.position.array
          positions[0] = startNode.position.x
          positions[1] = startNode.position.y
          positions[2] = startNode.position.z
          positions[3] = endNode.position.x
          positions[4] = endNode.position.y
          positions[5] = endNode.position.z
          line.geometry.attributes.position.needsUpdate = true
          
          // Animate line opacity
          const opacity = 0.3 + Math.sin(time * 1.5 + i) * 0.2
          line.material.opacity = opacity
        }
      }
    })
    
    // Rotate entire group
    if (groupRef.current) {
      groupRef.current.rotation.y = time * 0.05
    }
  })
  
  return (
    <group ref={groupRef}>
      {/* Network Nodes */}
      {nodeData.map((data, i) => (
        <mesh
          key={i}
          ref={el => nodesRef.current[i] = el}
          position={[data.x, data.y, data.z]}
          castShadow
        >
          <octahedronGeometry args={[0.2, 0]} />
          <meshStandardMaterial
            color="#a3a3a3"
            emissive="#737373"
            emissiveIntensity={0.3}
            metalness={0.8}
            roughness={0.2}
          />
        </mesh>
      ))}
      
      {/* Connection Lines */}
      {lineGeometries.map((geom, i) => (
        <line 
          key={`line-${i}`} 
          ref={el => linesRef.current[i] = el}
        >
          <primitive object={geom} />
          <lineBasicMaterial
            color="#d4d4d4"
            linewidth={1}
            transparent
            opacity={0.5}
          />
        </line>
      ))}
    </group>
  )
}

export default NetworkNodes