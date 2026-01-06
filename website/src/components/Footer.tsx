import Link from "next/link";
import Image from "next/image";

export function Footer() {
  return (
    <footer className="border-t border-[var(--hextech-gold)]/20 bg-[var(--abyss)] py-12">
      <div className="max-w-6xl mx-auto px-6">
        <div className="grid grid-cols-1 md:grid-cols-4 gap-8">
          <div className="col-span-1 md:col-span-2">
            <div className="flex items-center gap-2 mb-4">
              <Image
                src="/logo.png"
                alt="GhostDraft"
                width={75}
                height={75}
                className="drop-shadow-[0_0_12px_rgba(201,162,39,0.6)]"
              />
              <span className="text-xl font-semibold text-[var(--pale-gold)] font-display tracking-wide">
                GhostDraft
              </span>
            </div>
            <p className="text-[var(--text-secondary)] text-sm max-w-md leading-relaxed">
              A free, open-source champion select assistant for League of Legends.
              Get real-time matchup data, build recommendations, and team composition analysis.
            </p>
          </div>

          <div>
            <h3 className="text-[var(--pale-gold)] font-display font-semibold mb-4 tracking-wide">Links</h3>
            <ul className="space-y-3">
              <li>
                <Link
                  href="/#features"
                  className="text-[var(--text-secondary)] hover:text-[var(--hextech-gold)] transition-colors text-sm hover-line"
                >
                  Features
                </Link>
              </li>
              <li>
                <Link
                  href="/#download"
                  className="text-[var(--text-secondary)] hover:text-[var(--hextech-gold)] transition-colors text-sm hover-line"
                >
                  Download
                </Link>
              </li>
              <li>
                <Link
                  href="https://github.com/MatthewTran22/LoLOverlay"
                  target="_blank"
                  rel="noopener noreferrer"
                  className="text-[var(--text-secondary)] hover:text-[var(--arcane-cyan)] transition-colors text-sm hover-line"
                >
                  GitHub
                </Link>
              </li>
            </ul>
          </div>

          <div>
            <h3 className="text-[var(--pale-gold)] font-display font-semibold mb-4 tracking-wide">Legal</h3>
            <ul className="space-y-3">
              <li>
                <Link
                  href="/privacy"
                  className="text-[var(--text-secondary)] hover:text-[var(--hextech-gold)] transition-colors text-sm hover-line"
                >
                  Privacy Policy
                </Link>
              </li>
              <li>
                <Link
                  href="/terms"
                  className="text-[var(--text-secondary)] hover:text-[var(--hextech-gold)] transition-colors text-sm hover-line"
                >
                  Terms of Service
                </Link>
              </li>
            </ul>
          </div>
        </div>

        <div className="border-t border-[var(--hextech-gold)]/10 mt-8 pt-8 flex flex-col md:flex-row justify-between items-center gap-4">
          <p className="text-[var(--text-muted)] text-sm">
            &copy; {new Date().getFullYear()} GhostDraft. All rights reserved.
          </p>
          <p className="text-[var(--text-muted)] text-xs max-w-xl text-center md:text-right">
            GhostDraft is not endorsed by Riot Games and does not reflect the views or opinions of Riot Games
            or anyone officially involved in producing or managing League of Legends. League of Legends and
            Riot Games are trademarks or registered trademarks of Riot Games, Inc.
          </p>
        </div>
      </div>
    </footer>
  );
}
