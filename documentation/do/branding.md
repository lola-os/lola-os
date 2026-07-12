# LOLA OS Branding Guide

**Version:** 1.0  
**Date:** June 10, 2026  
**Philosophy:** Communicate to the user’s heart, not their ego. No jargon. Just clarity, trust, and quiet power.

---

## 1. Brand Essence (The Soul of LOLA)

**LOLA OS is The Bridge.**  
We don’t replace. We don’t lock in. We connect AI agents to blockchains, oracles, and APIs – so developers can stay in their world while we handle the messy parts.

**Core Attributes:**
- **Neutral** – All chains, all frameworks, all languages are equal.
- **Precise** – Every pixel, every word serves understanding.
- **Accessible** – A junior developer can grok it in minutes.
- **Calm** – No hype, no “revolution”. Just infrastructure that works.

**Brand Personality:**  
Confident but not arrogant. Technical but not cold. Minimal but not sterile. Like a senior engineer who explains complex things with a warm smile.

---

## 2. Color Palette (Monochromatic – No Favourites)

We use **pure grayscale** because we don’t pick sides. Depth comes from contrast, not colour.

| Token | Hex | Usage |
|-------|-----|-------|
| gray-50 | `#fafafa` | Light mode page backgrounds |
| gray-100 | `#f5f5f5` | Card backgrounds, hover states |
| gray-200 | `#e5e5e5` | Borders, dividers |
| gray-300 | `#d4d4d4` | Disabled text, placeholders |
| gray-400 | `#a3a3a3` | Secondary text, icons (dark mode) |
| gray-500 | `#737373` | Body text (light mode) |
| gray-600 | `#525252` | Subtitles, medium emphasis |
| gray-700 | `#404040` | High emphasis text (light mode) |
| gray-800 | `#262626` | Primary text (light mode), dark UI surfaces |
| gray-900 | `#171717` | Dark mode backgrounds |
| gray-950 | `#0a0a0a` | Deepest black – used sparingly |

**Accessibility:**  
Always meet WCAG AA. Test `gray-700` on `gray-50` (≥ 7:1) and `gray-100` on `gray-900` (≥ 12:1).

---

## 3. Typography (Clean, Readable, Developer‑Friendly)

- **Sans‑serif (UI & Marketing):** Inter – modern, highly legible on screens.
- **Monospace (Code, Logs, Terminal):** JetBrains Mono – excellent ligatures, developer‑proof.

| Element | Font | Size / Weight |
|---------|------|----------------|
| H1 (Hero) | Inter Bold | 3rem (48px) |
| H2 | Inter SemiBold | 2.25rem (36px) |
| H3 | Inter Medium | 1.5rem (24px) |
| Body | Inter Regular | 1rem (16px) |
| Small text | Inter Regular | 0.875rem (14px) |
| Inline code | JetBrains Mono | 0.9em |
| Code blocks | JetBrains Mono | 0.875rem |

**Line height:** 1.5 for body, 1.2 for headings.

---

## 4. Logo & Symbol (Quiet, Geometric, Bridge)

**The Mark:**  
A minimalist bridge – two parallel horizontal lines connected by a diagonal slash.  
- Represents AI (top line) and blockchain (bottom line) connected by LOLA (the slash).  
- Always grayscale (`gray-900` on light, `gray-100` on dark).  
- Never coloured, never distorted.

**Wordmark:**  
`LOLA` in Inter Bold, `OS` in Inter Regular, all uppercase, letterspaced.  
Sits to the right of the mark or can be used alone.

**Clear Space:**  
At least the height of the “L” on all sides of the combined mark+wordmark.

**Minimum Size:**  
Mark alone: 48px wide. Combined: 120px wide.

**Don’ts:**  
- ❌ No colour other than grayscale.  
- ❌ No drop shadows, glows, or gradients.  
- ❌ No rotation or squashing.  
- ❌ No placing on busy backgrounds without contrast.

---

## 5. Voice & Tone (Speak to the Heart, Not the Ego)

### Principles

1. **Clear over clever** – Use plain English. Explain blockchain concepts as if to a smart non‑expert.
2. **Confident but humble** – State facts without exaggeration. “Connects to any EVM chain” not “Unleashes the infinite power of blockchain”.
3. **Helpful** – Anticipate the next question. Every statement should lead to action or a code snippet.
4. **Warm** – A human wrote this. Smile in the text.

### Examples

| Instead of… | Say… |
|-------------|------|
| “Leverage our cutting‑edge multi‑chain architecture” | “Works with Ethereum, Polygon, and Solana – all at once.” |
| “Seamless integration with leading AI frameworks” | “Add `@lola_tool` to any CrewAI or LangChain agent.” |
| “Robust security with military‑grade encryption” | “Your keys never leave your machine. We use AES‑256, the same as your bank.” |
| “Our proprietary HITL system” | “We’ll ask for approval when you want it – you control the UI.” |

### Voice in Practice

- **Documentation:** “Here’s how to install LOLA. It takes two commands.”  
- **Error messages:** “Can’t reach Ethereum RPC. Try setting `ETH_RPC_URL` or run `lola doctor`.”  
- **Marketing:** “Five minutes. One decorator. Any blockchain.”

---

## 6. Logs & Terminal UI (Beautiful by Default)

LOLA’s logs are **not** boring monochrome lines. They are rich, colour‑coded, and icon‑driven.

- **Colour hints:** 🟢 green = success, 🔴 red = error, 🟡 yellow = warning, 🔵 blue = info.
- **Icons:** Fast visual scanning (✅, ⚠️, ❌, 🔄, ⏳).
- **Tables:** Alignment and borders for configuration status.
- **Progressive disclosure:** `LOLA_LOG_LEVEL=debug` for deep detail.

**We do not prescribe a specific UI for dashboards or landing pages.**  
Developers are free to build their own – we only provide the data (JSON logs) and the terminal experience. Let the developer’s creativity lead.

---

## 7. Iconography & Illustrations (Minimal, Geometric)

- **Icons:** Stroke‑based, 2px weight, from a neutral set (e.g., Lucide). Use `gray-600` (light) / `gray-300` (dark).
- **Illustrations:** Abstract geometric compositions – lines, dots, subtle gradients of gray. Evoke connectivity and data flow. No cartoon characters.
- **Data visualisation:** Monochromatic sequential scale (`gray-200` to `gray-800`). The darkest bar = primary metric.

---

## 8. Motion (Subtle, Purposeful, Respectful)

Animations are **quiet guides**, not distractions.

- **`fade-in`** (0.5s ease-out) – for mounting content, modals, toasts.
- **`slide-up`** (0.5s ease-out) – for cards and panels entering from bottom.
- **`pulse-slow`** (3s ease-in-out infinite) – for subtle loading indicators or live connections.
- **`float`** (6s ease-in-out infinite) – for hero graphics, used sparingly.

**Always respect `prefers-reduced-motion`.** Disable non‑essential animations when set.

---

## 9. What We *Don’t* Specify (And Why)

We intentionally **do not** design your landing page, documentation layout, dashboard UI, or mobile app.  

- Your users are developers – they value function over flashy templates.
- Every team has different needs (CLI‑first, web‑first, desktop app).
- We provide the **brand elements** (logo, colours, fonts, voice) and the **log data**. You build the experience.

**Hint for landing page:** Lead with “Add `@lola_tool` to any function” – a code snippet speaks louder than a thousand words. Use the grayscale palette. Let the code shine.

**Hint for documentation:** Start with “5 minutes to first blockchain call”. Use real examples. No jargon. Break long pages into short, scannable sections.

---

## 10. Deliverables for Implementation

- [ ] Logo pack – SVG + PNG (mark, wordmark, horizontal, vertical, mark‑only)
- [ ] Tailwind preset (or CSS variables) with the grayscale ramp
- [ ] Font files (Inter + JetBrains Mono) or CDN links
- [ ] This `branding.md` as the single source of truth
- [ ] Example terminal log styles (ANSI colour codes) for `rich` format

---

*This guide exists to give LOLA OS a consistent, quiet, trustworthy face – not to restrict creativity. Build what feels right. Just keep it grayscale, keep it clean, and speak human.*