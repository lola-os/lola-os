import { useRef } from "react";
import { useFrame } from "@react-three/fiber";
import { MeshWobbleMaterial, Text, Float } from "@react-three/drei";

function BridgeModel() {
  const bridgeRef = useRef();
  const leftParticlesRef = useRef([]);
  const rightParticlesRef = useRef([]);
  const dataStreamsRef = useRef([]);

  // Grayish colors
  const aiColor = "#737373";
  const blockchainColor = "#525252";
  const bridgeColor = "#a3a3a3";

  useFrame((state) => {
    const t = state.clock.elapsedTime;

    if (bridgeRef.current) {
      bridgeRef.current.rotation.y = Math.sin(t * 0.1) * 0.03;
    }

    // Animate particles moving from AI to Blockchain (left to right)
    leftParticlesRef.current.forEach((particle, i) => {
      if (!particle) return;
      const speed = 0.3 + Math.sin(i * 0.2) * 0.1;
      const progress = (t * speed + i * 0.2) % 1;
      
      // Move from AI side to Blockchain side
      const startX = -3.0;
      const endX = 3.0;
      const x = startX + progress * (endX - startX);
      
      // Add subtle vertical movement
      const y = 0.1 + Math.sin(t * 2 + i) * 0.05;
      const z = Math.sin(t + i) * 0.2;
      
      particle.position.set(x, y, z);
      
      // Change color gradually as it crosses
      if (particle.material) {
        if (progress < 0.5) {
          particle.material.color.set(aiColor);
        } else {
          particle.material.color.set(blockchainColor);
        }
      }
    });

    // Animate particles moving from Blockchain to AI (right to left)
    rightParticlesRef.current.forEach((particle, i) => {
      if (!particle) return;
      const speed = 0.25 + Math.cos(i * 0.2) * 0.1;
      const progress = (t * speed + i * 0.3) % 1;
      
      // Move from Blockchain side to AI side
      const startX = 3.0;
      const endX = -3.0;
      const x = startX + progress * (endX - startX);
      
      // Different vertical movement pattern
      const y = 0.15 + Math.cos(t * 1.5 + i) * 0.05;
      const z = Math.cos(t * 0.8 + i) * 0.15;
      
      particle.position.set(x, y, z);
      
      // Change color gradually as it crosses
      if (particle.material) {
        if (progress < 0.5) {
          particle.material.color.set(blockchainColor);
        } else {
          particle.material.color.set(aiColor);
        }
      }
    });

    // Animate data streams
    dataStreamsRef.current.forEach((stream, i) => {
      if (!stream || !stream.material) return;
      const pulse = Math.sin(t * 3 + i) * 0.5 + 0.5;
      stream.material.opacity = 0.2 + pulse * 0.2;
    });
  });

  return (
    <group ref={bridgeRef}>
      {/* Main Bridge Deck */}
      <mesh position={[0, 0, 0]} castShadow receiveShadow>
        <boxGeometry args={[6, 0.08, 1.2]} />
        <MeshWobbleMaterial
          color={bridgeColor}
          emissive="#d4d4d4"
          emissiveIntensity={0.1}
          metalness={0.7}
          roughness={0.3}
          speed={0.2}
          factor={0.02}
        />
      </mesh>

      {/* Support Pillars */}
      {[-2.5, 0, 2.5].map((x, i) => (
        <mesh key={i} position={[x, -0.5, 0]} castShadow>
          <cylinderGeometry args={[0.1, 0.15, 1, 6]} />
          <meshStandardMaterial
            color="#737373"
            metalness={0.8}
            roughness={0.2}
          />
        </mesh>
      ))}

      {/* Data Streams - Continuous flow lines */}
      <group>
        {Array.from({ length: 3 }).map((_, i) => {
          const height = 0.05 + i * 0.04;
          return (
            <mesh
              key={`stream-${i}`}
              ref={(el) => (dataStreamsRef.current[i] = el)}
              position={[0, height, 0]}
            >
              <cylinderGeometry args={[0.01, 0.01, 6, 8]} />
              <meshBasicMaterial
                color={bridgeColor}
                transparent
                opacity={0.3}
              />
            </mesh>
          );
        })}
      </group>

      {/* Particles moving from AI to Blockchain (left to right) */}
      <group>
        {Array.from({ length: 8 }).map((_, i) => (
          <mesh
            key={`to-blockchain-${i}`}
            ref={(el) => (leftParticlesRef.current[i] = el)}
          >
            <icosahedronGeometry args={[0.03, 0]} />
            <meshStandardMaterial
              color={aiColor}
              metalness={0.8}
              roughness={0.2}
            />
          </mesh>
        ))}
      </group>

      {/* Particles moving from Blockchain to AI (right to left) */}
      <group>
        {Array.from({ length: 8 }).map((_, i) => (
          <mesh
            key={`to-ai-${i}`}
            ref={(el) => (rightParticlesRef.current[i] = el)}
          >
            <octahedronGeometry args={[0.04, 0]} />
            <meshStandardMaterial
              color={blockchainColor}
              metalness={0.8}
              roughness={0.2}
            />
          </mesh>
        ))}
      </group>

      {/* Connection Points - Minimal indicators */}
      {[-3.0, 3.0].map((x, i) => {
        const isAI = i === 0;
        return (
          <mesh key={i} position={[x, 0.1, 0]}>
            <ringGeometry args={[0.1, 0.12, 6]} />
            <meshBasicMaterial
              color={isAI ? aiColor : blockchainColor}
              transparent
              opacity={0.3}
              side={2}
            />
          </mesh>
        );
      })}

      {/* Bridge Labels */}
      <Float floatIntensity={0.2} speed={1}>
        {/* AI Side Label */}
        <Text
          position={[-2.5, 0.8, 0]}
          fontSize={0.18}
          color={aiColor}
          anchorX="center"
        >
          AI Agents
        </Text>
        {/* Bridge Center Label */}
        <Text
          position={[0, 0.8, 0]}
          fontSize={0.2}
          color="#f5f5f5"
          anchorX="center"
        >
          →
        </Text>
        {/* Blockchain Side Label */}
        <Text
          position={[2.5, 0.8, 0]}
          fontSize={0.18}
          color={blockchainColor}
          anchorX="center"
        >
          Blockchain
        </Text>
      </Float>

      {/* Bridge Energy Field - Subtle glow */}
      <mesh>
        <boxGeometry args={[6.2, 0.3, 1.4]} />
        <meshBasicMaterial
          color={bridgeColor}
          transparent
          opacity={0.05}
          side={2}
        />
      </mesh>

      {/* Ground Connection Lines */}
      {Array.from({ length: 5 }).map((_, i) => {
        const x = -2.4 + i * 1.2;
        return (
          <mesh key={`ground-line-${i}`} position={[x, -0.2, 0]}>
            <cylinderGeometry args={[0.005, 0.005, 0.4, 4]} />
            <meshBasicMaterial
              color={bridgeColor}
              transparent
              opacity={0.2}
            />
          </mesh>
        );
      })}
    </group>
  );
}

export default BridgeModel;