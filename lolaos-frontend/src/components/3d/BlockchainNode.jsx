import { useRef } from "react";
import { useFrame } from "@react-three/fiber";
import { Text, Float } from "@react-three/drei";

function BlockchainNode({ position = [0, 0, 0] }) {
  const groupRef = useRef();
  const blocksRef = useRef([]);
  
  // Grayish blockchain colors
  const colors = {
    block: "#737373",
    blockAlt: "#a3a3a3",
    chain: "#525252",
    ledger: "#404040",
    accent: "#d4d4d4"
  };

  useFrame((state) => {
    const t = state.clock.elapsedTime;

    if (groupRef.current) {
      groupRef.current.rotation.y = t * 0.05;
    }

    // Subtle block floating animation
    blocksRef.current.forEach((block, i) => {
      if (!block) return;
      const timeOffset = i * 0.3;
      block.position.y = Math.sin(t * 0.5 + timeOffset) * 0.05;
    });
  });

  return (
    <group ref={groupRef} position={position}>
      {/* Central Ledger - Now much smaller and cleaner */}
      <mesh castShadow>
        <boxGeometry args={[0.7, 0.7, 0.7]} />
        <meshPhysicalMaterial
          color={colors.ledger}
          emissive={colors.chain}
          emissiveIntensity={0.1}
          metalness={0.7}
          roughness={0.25}
          clearcoat={0.5}
        />
      </mesh>

      {/* Ledger inner detail */}
      <mesh position={[0, 0, 0.21]}>
        <boxGeometry args={[0.7, 0.7, 0.02]} />
        <meshStandardMaterial
          color={colors.accent}
          emissive={colors.accent}
          emissiveIntensity={0.3}
          metalness={0.9}
          roughness={0.1}
        />
      </mesh>

      {/* Blockchain - Series of connected blocks in a circle */}
      <group>
        {Array.from({ length: 8 }).map((_, i) => {
          const angle = (i / 8) * Math.PI * 2;
          const radius = 1.2;
          const x = Math.cos(angle) * radius;
          const z = Math.sin(angle) * radius;
          const height = 0.1;

          // Different sizes for blocks to represent blockchain structure
          const blockSize = 0.2 + (i % 3) * 0.05;
          
          return (
            <group key={i}>
              <mesh
                ref={(el) => (blocksRef.current[i] = el)}
                position={[x, height, z]}
                rotation={[0, angle, 0]}
                castShadow
              >
                <boxGeometry args={[blockSize, 0.1, 0.1]} />
                <meshStandardMaterial
                  color={i % 2 === 0 ? colors.block : colors.blockAlt}
                  metalness={0.6}
                  roughness={0.3}
                />
                
                {/* Block number/data */}
                <mesh position={[0, 0.06, 0.06]}>
                  <planeGeometry args={[0.08, 0.04]} />
                  <meshBasicMaterial color={colors.accent} />
                </mesh>
              </mesh>
              
              {/* Connect blocks in chain */}
              {i < 7 && (
                <mesh>
                  <cylinderGeometry args={[0.015, 0.015, radius * 0.5, 4]} />
                  <meshStandardMaterial
                    color={colors.chain}
                    emissive={colors.chain}
                    emissiveIntensity={0.1}
                  />
                </mesh>
              )}
            </group>
          );
        })}
      </group>

      {/* Chain links connecting to center */}
      {Array.from({ length: 4 }).map((_, i) => {
        const angle = (i / 4) * Math.PI * 2;
        const startRadius = 0.25;
        const endRadius = 1.0;
        
        return (
          <mesh key={`link-${i}`} position={[
            Math.cos(angle) * (startRadius + endRadius) / 2,
            0,
            Math.sin(angle) * (startRadius + endRadius) / 2
          ]}>
            <cylinderGeometry args={[0.01, 0.01, endRadius - startRadius, 4]} />
            <meshBasicMaterial
              color={colors.chain}
              transparent
              opacity={0.4}
            />
          </mesh>
        );
      })}

      {/* Data flow indicators */}
      <group>
        {Array.from({ length: 3 }).map((_, i) => (
            <mesh key={`data-${i}`} position={[0, -0.1, 0]}>
              <ringGeometry args={[0.8, 0.85, 32]} />
              <meshBasicMaterial
                color={colors.accent}
                transparent
                opacity={0.1}
                side={2}
              />
            </mesh>
          ))}
      </group>

      {/* Node label */}
      <Float floatIntensity={0.2} speed={1}>
        <Text
          position={[0, -0.8, 0]}
          fontSize={0.2}
          color="#e5e5e5"
          anchorX="center"
        >
          Blockchain
        </Text>
        <Text
          position={[0, -1.1, 0]}
          fontSize={0.12}
          color="#a3a3a3"
          anchorX="center"
        >
          Immutable Ledger
        </Text>
      </Float>
    </group>
  );
}

export default BlockchainNode;