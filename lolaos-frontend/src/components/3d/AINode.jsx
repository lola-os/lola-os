import { useRef } from "react"
import { useFrame } from "@react-three/fiber"
import { Text, Line } from "@react-three/drei"

// Grayscale only — depth comes from lightness and emissive intensity,
// never from hue (branding.md).
const NODE = {
  color: "#8a8a8a",
  emissive: "#c9c9c9",
  core: "#f5f5f5",
  label: "AI Agent",
}

function AINode({ position }) {
  const agentRef = useRef()
  const particlesRef = useRef([])
  const linesRef = useRef([])

  useFrame((state) => {
    const t = state.clock.elapsedTime

    if (agentRef.current) {
      agentRef.current.position.y = position[1] + Math.sin(t * 0.5) * 0.1
      agentRef.current.rotation.y = t * 0.2

      const innerCore = agentRef.current.getObjectByName("innerCore")
      if (innerCore) {
        innerCore.rotation.x = t * 0.5
        innerCore.rotation.z = t * 0.3
      }
    }

    particlesRef.current.forEach((p, i) => {
      if (!p) return
      p.position.x = Math.cos(t * 2 + i) * 1.8
      p.position.y = Math.sin(t * 1.5 + i) * 1.8
      p.position.z = Math.sin(t * 1.2 + i) * 1.8
    })

    linesRef.current.forEach((l, i) => {
      if (!l?.material) return
      l.material.opacity = 0.25 + Math.sin(t * 2 + i) * 0.2
    })
  })

  return (
    <group ref={agentRef} position={position}>
      {/* Main agent body */}
      <mesh name="agentBody" castShadow receiveShadow>
        <dodecahedronGeometry args={[1.2, 0]} />
        <meshPhysicalMaterial
          color={NODE.color}
          emissive={NODE.emissive}
          emissiveIntensity={0.18}
          metalness={0.65}
          roughness={0.28}
          transparent
          opacity={0.92}
          clearcoat={1}
          clearcoatRoughness={0.05}
        />
      </mesh>

      {/* Inner core */}
      <mesh name="innerCore" rotation={[Math.PI / 4, 0, 0]}>
        <octahedronGeometry args={[0.6, 1]} />
        <meshStandardMaterial
          color={NODE.core}
          emissive={NODE.emissive}
          emissiveIntensity={0.35}
          metalness={0.9}
          roughness={0.12}
        />
      </mesh>

      {/* Orbiting particles */}
      <group>
        {Array.from({ length: 10 }).map((_, i) => (
          <mesh key={i} ref={(el) => (particlesRef.current[i] = el)}>
            <sphereGeometry args={[0.03, 6, 6]} />
            <meshBasicMaterial color="#e5e5e5" transparent opacity={0.75} />
          </mesh>
        ))}
      </group>

      {/* Thought lines */}
      <group>
        {Array.from({ length: 5 }).map((_, i) => {
          const angle = (i / 5) * Math.PI * 2
          const end = [Math.cos(angle) * 2.5, Math.sin(angle + i) * 0.3, Math.sin(angle) * 2.5]
          return (
            <Line
              key={i}
              ref={(el) => (linesRef.current[i] = el)}
              points={[[0, 0, 0], end]}
              color="#a3a3a3"
              lineWidth={1.2}
              transparent
              opacity={0.3}
            />
          )
        })}
      </group>

      {/* Label */}
      <Text
        position={[0, -0.85, 0]}
        fontSize={0.16}
        color="#e5e5e5"
        anchorX="center"
        anchorY="middle"
        outlineWidth={0.006}
        outlineColor="#0a0a0a"
      >
        {NODE.label}
      </Text>

      {/* Status dot */}
      <mesh position={[0, 1.5, 0]}>
        <sphereGeometry args={[0.07, 16, 16]} />
        <meshStandardMaterial color="#f5f5f5" emissive="#d4d4d4" emissiveIntensity={0.8} />
      </mesh>
    </group>
  )
}

export default AINode
