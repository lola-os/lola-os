"use client";

import * as React from "react";
import { Menu } from "lucide-react";
import { Button } from "@/components/ui/button";
import { Sheet, SheetTrigger, SheetContent } from "@/components/ui/sheet";
import { DocsSidebar } from "@/components/docs-sidebar";

export default function DocsLayout({ children }: { children: React.ReactNode }) {
  const [open, setOpen] = React.useState(false);

  return (
    <div className="container py-12 sm:py-16">
      <div className="mb-8 md:hidden">
        <Sheet open={open} onOpenChange={setOpen}>
          <SheetTrigger asChild>
            <Button variant="outline" size="sm" className="gap-2">
              <Menu className="h-4 w-4" />
              Docs menu
            </Button>
          </SheetTrigger>
          <SheetContent>
            <div className="mt-12" onClick={() => setOpen(false)}>
              <DocsSidebar />
            </div>
          </SheetContent>
        </Sheet>
      </div>

      <div className="grid gap-16 md:grid-cols-[220px_1fr] lg:grid-cols-[240px_1fr]">
        <aside className="hidden md:block">
          <div className="sticky top-28">
            <DocsSidebar />
          </div>
        </aside>
        <article className="docs-prose min-w-0 animate-fade-in">{children}</article>
      </div>
    </div>
  );
}
