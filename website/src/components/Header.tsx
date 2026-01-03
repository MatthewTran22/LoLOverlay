import Link from "next/link";

export function Header() {
  return (
    <header className="border-b border-[var(--hextech-gold)]/20 bg-[var(--void-black)]/80 backdrop-blur-md sticky top-0 z-50">
      <div className="max-w-6xl mx-auto px-6 py-4 flex items-center justify-between">
        <Link href="/" className="flex items-center gap-3 group">
          <div className="w-10 h-10 bg-gradient-to-br from-[var(--hextech-gold)] to-[var(--bright-gold)] rounded-lg flex items-center justify-center pulse-glow">
            <span className="text-[var(--void-black)] font-bold text-sm font-display">GD</span>
          </div>
          <span className="text-xl font-semibold text-[var(--pale-gold)] font-display tracking-wide group-hover:text-glow transition-all">
            GhostDraft
          </span>
        </Link>

        <nav className="flex items-center gap-8">
          <Link
            href="/#features"
            className="text-[var(--text-secondary)] hover:text-[var(--hextech-gold)] transition-colors hover-line font-medium"
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
            className="text-[var(--text-secondary)] hover:text-[var(--hextech-gold)] transition-colors hover-line font-medium"
          >
            Download
          </Link>
          <Link
            href="/privacy"
            className="text-[var(--text-secondary)] hover:text-[var(--hextech-gold)] transition-colors hover-line font-medium"
          >
            Privacy
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
