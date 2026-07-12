import { Canvas } from "@react-three/fiber";
import { Environment, Html, OrbitControls, PerspectiveCamera, useProgress } from "@react-three/drei";
import { Suspense } from "react";
import { useTheme } from "../theme/themeContext";
import useReducedMotion from "../../hooks/useReducedMotion";

import BridgeModel from "./BridgeModel";
import NetworkNodes from "./NetworkNodes";
import ConnectionLines from "./ConnectionLines";
import AINode from "./AINode";
import BlockchainNode from "./BlockchainNode";

function Loader() {
  const { progress } = useProgress();
  return (
    <Html center>
      <div className="font-mono text-xs text-gray-400">{Math.round(progress)}%</div>
    </Html>
  );
}

export default function HeroScene() {
  const reducedMotion = useReducedMotion();
  const { theme } = useTheme();
  const bg = theme === "dark" ? "#0a0a0a" : "#fafafa";

  return (
    <div className="absolute inset-0">
      <Canvas
        shadows
        dpr={[1, 1.75]}
        frameloop={reducedMotion ? "demand" : "always"}
        gl={{ antialias: true, powerPreference: "high-performance" }}
        camera={{ position: [0, 5, 10], fov: 50 }}
      >
        <color attach="background" args={[bg]} />
        <fog attach="fog" args={[bg, 11, 26]} />

        <PerspectiveCamera makeDefault position={[0, 5, 10]} fov={50} />

        <OrbitControls
          enableZoom={false}
          enablePan={false}
          autoRotate={!reducedMotion}
          autoRotateSpeed={0.3}
          maxPolarAngle={Math.PI / 2.2}
          minPolarAngle={Math.PI / 4}
        />

        <ambientLight intensity={theme === "dark" ? 0.4 : 0.65} />
        <directionalLight
          position={[10, 10, 5]}
          intensity={theme === "dark" ? 0.9 : 1.1}
          castShadow
          shadow-mapSize-width={1024}
          shadow-mapSize-height={1024}
        />

        <Suspense fallback={<Loader />}>
          <AINode position={[-4, 0, 0]} />
          <BridgeModel />
          <BlockchainNode position={[4, 0, 0]} />
          <ConnectionLines />
          <NetworkNodes />
          <Environment files="/hdri/studio.hdr" background={false} />
        </Suspense>
      </Canvas>
    </div>
  );
}
