"use client";

import * as React from "react";
import { motion } from "framer-motion";

/** Hero fade-in-up: runs once on mount, no scroll trigger needed since
 * the hero is always in view on load. */
export function HeroFade({ children }: { children: React.ReactNode }) {
  return (
    <motion.div
      initial={{ opacity: 0, y: 16 }}
      animate={{ opacity: 1, y: 0 }}
      transition={{ duration: 0.5, ease: "easeOut" }}
    >
      {children}
    </motion.div>
  );
}

/** Scroll-triggered fade-in for cards and sections further down the page,
 * per branding.md's "quiet guides, not distractions" motion philosophy. */
export function ScrollFade({
  children,
  delay = 0,
}: {
  children: React.ReactNode;
  delay?: number;
}) {
  return (
    <motion.div
      initial={{ opacity: 0, y: 12 }}
      whileInView={{ opacity: 1, y: 0 }}
      viewport={{ once: true, margin: "-80px" }}
      transition={{ duration: 0.5, ease: "easeOut", delay }}
    >
      {children}
    </motion.div>
  );
}
