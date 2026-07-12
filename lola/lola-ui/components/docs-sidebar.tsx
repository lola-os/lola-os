"use client";

import Link from "next/link";
import { usePathname } from "next/navigation";
import { cn } from "@/lib/utils";
import { DOCS_NAV } from "@/lib/docs-nav";

export function DocsSidebar() {
  const pathname = usePathname();

  return (
    <nav aria-label="Documentation" className="space-y-9 text-sm">
      {DOCS_NAV.map((section) => (
        <div key={section.title}>
          <h4 className="mb-3 text-[0.7rem] font-semibold uppercase tracking-wider text-gray-400 dark:text-gray-600">
            {section.title}
          </h4>
          <ul className="space-y-0.5 border-l border-gray-200 dark:border-gray-800">
            {section.items.map((item) => {
              const active = pathname === item.href;
              return (
                <li key={item.href} className="relative">
                  {active && (
                    <span className="absolute -left-px top-0 h-full w-px bg-gray-900 dark:bg-gray-50" />
                  )}
                  <Link
                    href={item.href}
                    className={cn(
                      "block py-1.5 pl-4 transition-colors",
                      active
                        ? "font-medium text-gray-900 dark:text-gray-50"
                        : "text-gray-500 hover:text-gray-900 dark:text-gray-500 dark:hover:text-gray-200"
                    )}
                  >
                    {item.title}
                  </Link>
                </li>
              );
            })}
          </ul>
        </div>
      ))}
    </nav>
  );
}
