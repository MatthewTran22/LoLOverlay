"use client";

import Link from "next/link";
import Image from "next/image";
import ChampionSearch from "./ChampionSearch";

export function Header() {
  return (
    <header className="border-b border-[var(--hextech-gold)]/20 bg-[var(--void-black)]/80 backdrop-blur-md sticky top-0 z-50">
      <div className="max-w-6xl mx-auto px-6 py-2 flex items-center justify-between gap-4">
        <Link href="/" className="flex items-center gap-2 group shrink-0">
          <Image
            src="/logo.png"
            alt="GhostDraft"
            width={75}
            height={75}
            className="drop-shadow-[0_0_12px_rgba(201,162,39,0.6)]"
          />
          <span className="text-xl font-semibold text-[var(--pale-gold)] font-display tracking-wide group-hover:text-glow transition-all hidden sm:block">
            GhostDraft
          </span>
        </Link>

        <div className="flex-1 max-w-xs">
          <ChampionSearch compact />
        </div>

        <nav className="flex items-center gap-6 shrink-0">
          <Link
            href="/#features"
            className="text-[var(--text-secondary)] hover:text-[var(--hextech-gold)] transition-colors hover-line font-medium hidden md:block"
          >
            Features
          </Link>
          <Link
            href="/stats"
            className="text-[var(--text-secondary)] hover:text-[var(--hextech-gold)] transition-colors hover-line font-medium"
          >
            Stats
          </Link>
          <Link
            href="/#download"
            className="text-[var(--text-secondary)] hover:text-[var(--hextech-gold)] transition-colors hover-line font-medium hidden md:block"
          >
            Download
          </Link>
          <Link
            href="https://github.com/MatthewTran22/LoLOverlay"
            target="_blank"
            rel="noopener noreferrer"
            className="text-[var(--text-secondary)] hover:text-[var(--arcane-cyan)] transition-colors hover-line font-medium"
          >
            GitHub
          </Link>
        </nav>
      </div>
    </header>
  );
}
