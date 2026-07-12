"use client";

import * as React from "react";
import Link from "next/link";
import { usePathname } from "next/navigation";
import { Menu } from "lucide-react";
import { cn } from "@/lib/utils";
import { Button } from "@/components/ui/button";
import { Sheet, SheetTrigger, SheetContent } from "@/components/ui/sheet";
import { ThemeToggle } from "@/components/theme-toggle";
import { LolaMark } from "@/components/lola-mark";

const NAV_LINKS = [
  { href: "/docs", label: "Docs" },
  { href: "/docs/getting-started", label: "Getting started" },
  { href: "/docs/examples/python-5min", label: "Examples" },
  { href: "https://github.com/lola-os", label: "GitHub" },
];

export function SiteHeader() {
  const [open, setOpen] = React.useState(false);
  const pathname = usePathname();

  return (
    <header className="sticky top-0 z-40 border-b border-gray-200/70 bg-gray-50/80 backdrop-blur-md dark:border-gray-800/70 dark:bg-gray-950/80">
      <div className="container flex h-[4.5rem] items-center justify-between">
        <Link href="/" className="flex items-center gap-2.5" aria-label="LOLA OS home">
          <LolaMark className="h-[1.4rem] w-[1.4rem] text-gray-900 dark:text-gray-50" />
          <span className="text-[0.95rem] font-bold tracking-tight text-gray-900 dark:text-gray-50">
            LOLA <span className="font-medium text-gray-400 dark:text-gray-500">OS</span>
          </span>
        </Link>

        {/* Desktop nav */}
        <nav className="hidden items-center gap-8 md:flex">
          {NAV_LINKS.map((link) => {
            const active = pathname === link.href;
            return (
              <Link
                key={link.href}
                href={link.href}
                className={cn(
                  "text-[0.875rem] font-medium transition-colors",
                  active
                    ? "text-gray-900 dark:text-gray-50"
                    : "text-gray-500 hover:text-gray-900 dark:text-gray-400 dark:hover:text-gray-50"
                )}
              >
                {link.label}
              </Link>
            );
          })}
        </nav>

        <div className="hidden items-center gap-3 md:flex">
          <ThemeToggle />
          <Button asChild size="sm">
            <Link href="/docs/getting-started">Get started</Link>
          </Button>
        </div>

        {/* Mobile nav */}
        <div className="flex items-center gap-1 md:hidden">
          <ThemeToggle />
          <Sheet open={open} onOpenChange={setOpen}>
            <SheetTrigger asChild>
              <Button variant="ghost" size="icon" aria-label="Open menu">
                <Menu className="h-5 w-5" />
              </Button>
            </SheetTrigger>
            <SheetContent>
              <nav className="mt-12 flex flex-col gap-1">
                {NAV_LINKS.map((link) => (
                  <Link
                    key={link.href}
                    href={link.href}
                    onClick={() => setOpen(false)}
                    className="rounded-lg px-3 py-3 text-[0.95rem] font-medium text-gray-700 hover:bg-gray-100 dark:text-gray-200 dark:hover:bg-gray-900"
                  >
                    {link.label}
                  </Link>
                ))}
                <Button asChild className="mt-4" onClick={() => setOpen(false)}>
                  <Link href="/docs/getting-started">Get started</Link>
                </Button>
              </nav>
            </SheetContent>
          </Sheet>
        </div>
      </div>
    </header>
  );
}
