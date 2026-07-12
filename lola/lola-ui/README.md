# lola-ui

The LOLA OS landing page and developer documentation — Next.js 14 (App
Router), Tailwind CSS, shadcn/ui-pattern components, Framer Motion, MDX
docs. Follows `branding.md` exactly: grayscale only, Inter + JetBrains
Mono, dark mode, calm animations.

## Design principles

The interface leads with hierarchy and spacing rather than color. A real
display type scale (`tailwind.config.js`'s `display-sm/md/lg`), a
consistent `container` max-width, generous section rhythm
(`py-24`/`py-32`), and icon-anchored feature blocks keep the layout
calm. The landing page uses a code block with real window chrome, and the
docs render in a narrow (`max-w-[42rem]`) reading column with a refined
type system in `globals.css`'s `.docs-prose`. "How it works" is a
quiet, icon-based step flow. The grayscale palette is used throughout.

## Setup

```bash
cd lola-ui
npm install
npm run dev
```

Then open `http://localhost:3000`.

```bash
npm run build   # production build
npm run lint    # eslint
```

## Structure

```
app/
├── page.tsx                  Landing page
├── layout.tsx                 Root layout: fonts, theme provider, header/footer
├── globals.css                 Design tokens, dark mode, docs-prose styles
└── docs/
    ├── layout.tsx               Sidebar + mobile sheet nav, wraps every docs page
    ├── page.mdx                 Docs home
    ├── getting-started/page.mdx
    ├── configuration/page.mdx
    ├── security/page.mdx
    ├── api-reference/page.mdx
    ├── faq/page.mdx
    ├── operational/
    │   ├── replay/page.mdx       Full plan.json schema breakdown
    │   ├── doctor/page.mdx
    │   ├── registry/page.mdx
    │   └── metrics/page.mdx
    └── examples/
        ├── python-5min/page.mdx
        ├── replay-plan/page.mdx
        ├── crewai/page.mdx
        ├── langchain/page.mdx
        └── custom-bot/page.mdx
components/
├── ui/                          shadcn/ui-pattern primitives (button, card, sheet)
├── site-header.tsx, site-footer.tsx, docs-sidebar.tsx
├── code-block.tsx                Syntax-highlighted code with copy button
├── motion.tsx                    Framer Motion fade-in-up / scroll-triggered fade
├── lola-mark.tsx                  The bridge logo, as SVG
└── theme-provider.tsx, theme-toggle.tsx
lib/
├── utils.ts                      cn() className merge helper
└── docs-nav.ts                    Sidebar navigation structure
mdx-components.tsx                 Routes fenced code blocks through CodeBlock
```

## UI primitives

`components/ui/button.tsx`, `card.tsx`, and `sheet.tsx` are
self-contained shadcn/ui-pattern components. They use the same
conventions as the shadcn registry — `cva` variant structure, Radix
primitives, and the `cn()` utility — so they are drop-in compatible with
shadcn CLI output. Adding more components from the CLI later slots in
alongside these three with zero conflicts.

## Design adherence

- **Colors**: every class in every component traces back to the
  `gray-50`-`gray-950` ramp from `branding.md` section 2. There is no
  other hue anywhere in this codebase — check `tailwind.config.js`.
- **Typography**: Inter (sans) and JetBrains Mono (mono), loaded via
  `next/font/google` in `app/layout.tsx`, exposed as CSS variables
  Tailwind's `font-sans`/`font-mono` resolve to.
- **Motion**: only `fade-in`, `slide-up`, `pulse-slow`, and `float` exist
  as animations (see `tailwind.config.js` keyframes), matching
  branding.md section 8 exactly. `prefers-reduced-motion` is respected
  globally in `globals.css`.
- **Voice**: copy throughout (landing page, docs) avoids hype words —
  "simple", "works with any chain", "operationally mature" — per
  branding.md section 5.

## Notes

- **Syntax highlighting**: `code-block.tsx` defines a grayscale Prism
  theme against `react-syntax-highlighter`'s style-object API.
- **MDX + JSX**: a few `.mdx` files use raw JSX (e.g. a wrapping `<div
  className="overflow-x-auto">` around a wide table), supported by
  `@next/mdx` with `mdxRs: true`.
