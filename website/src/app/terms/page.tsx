import Link from "next/link";
import type { Metadata } from "next";

export const metadata: Metadata = {
  title: "Terms of Service - GhostDraft",
  description: "GhostDraft terms of service. Understand your rights and responsibilities when using our software.",
};

export default function TermsOfService() {
  return (
    <div className="max-w-4xl mx-auto px-6 py-16">
      <h1 className="text-4xl md:text-5xl font-display font-bold text-[var(--pale-gold)] mb-4 text-glow">
        Terms of Service
      </h1>
      <p className="text-[var(--text-muted)] mb-8">Last updated: January 2, 2026</p>

      <div className="space-y-12">
        <section>
          <h2 className="text-2xl font-display font-semibold text-[var(--pale-gold)] mb-4">1. Acceptance of Terms</h2>
          <p className="text-[var(--text-secondary)] leading-relaxed">
            By downloading, installing, or using GhostDraft (&quot;the Software&quot;), you agree to be
            bound by these Terms of Service. If you do not agree to these terms, do not use
            the Software.
          </p>
        </section>

        <section>
          <h2 className="text-2xl font-display font-semibold text-[var(--pale-gold)] mb-4">2. Description of Service</h2>
          <p className="text-[var(--text-secondary)] leading-relaxed">
            GhostDraft is a free, open-source overlay application that provides champion select
            assistance for League of Legends. The Software displays statistical information
            including matchup win rates, item build recommendations, and team composition analysis.
          </p>
        </section>

        <section>
          <h2 className="text-2xl font-display font-semibold text-[var(--pale-gold)] mb-4">3. License</h2>
          <p className="text-[var(--text-secondary)] leading-relaxed">
            GhostDraft is open-source software. You are granted a non-exclusive, worldwide,
            royalty-free license to use, copy, modify, and distribute the Software in accordance
            with the license terms in our{" "}
            <Link
              href="https://github.com/MatthewTran22/LoLOverlay"
              target="_blank"
              rel="noopener noreferrer"
              className="text-[var(--hextech-gold)] hover:text-[var(--bright-gold)] transition-colors hover-line"
            >
              GitHub repository
            </Link>
            .
          </p>
        </section>

        <section>
          <h2 className="text-2xl font-display font-semibold text-[var(--pale-gold)] mb-4">4. Compliance with Riot Games Terms</h2>
          <p className="text-[var(--text-secondary)] mb-4 leading-relaxed">
            GhostDraft is designed to comply with Riot Games&apos; Terms of Service and API policies.
            The Software:
          </p>
          <ul className="list-disc list-inside text-[var(--text-secondary)] space-y-2 ml-4">
            <li>Does not automate any gameplay actions</li>
            <li>Does not provide unfair advantages during active gameplay</li>
            <li>Only accesses publicly available data through official APIs</li>
            <li>Does not modify game files or memory</li>
            <li>Functions only during champion select, not during active gameplay</li>
          </ul>
          <p className="text-[var(--text-secondary)] mt-4 leading-relaxed">
            Users are responsible for ensuring their use of third-party tools complies with
            Riot Games&apos; current Terms of Service.
          </p>
        </section>

        <section>
          <h2 className="text-2xl font-display font-semibold text-[var(--pale-gold)] mb-4">5. Disclaimer of Warranties</h2>
          <p className="text-[var(--text-secondary)] mb-4 leading-relaxed">
            THE SOFTWARE IS PROVIDED &quot;AS IS&quot; WITHOUT WARRANTY OF ANY KIND, EXPRESS OR IMPLIED,
            INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY, FITNESS FOR A
            PARTICULAR PURPOSE, AND NONINFRINGEMENT.
          </p>
          <p className="text-[var(--text-secondary)] mb-4 leading-relaxed">
            We do not warrant that:
          </p>
          <ul className="list-disc list-inside text-[var(--text-secondary)] space-y-2 ml-4">
            <li>The Software will meet your specific requirements</li>
            <li>The Software will be uninterrupted, timely, secure, or error-free</li>
            <li>The statistical data provided will be accurate or complete</li>
            <li>Any errors in the Software will be corrected</li>
          </ul>
        </section>

        <section>
          <h2 className="text-2xl font-display font-semibold text-[var(--pale-gold)] mb-4">6. Limitation of Liability</h2>
          <p className="text-[var(--text-secondary)] mb-4 leading-relaxed">
            IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM,
            DAMAGES, OR OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT, OR OTHERWISE,
            ARISING FROM, OUT OF, OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER
            DEALINGS IN THE SOFTWARE.
          </p>
          <p className="text-[var(--text-secondary)] leading-relaxed">
            This includes, but is not limited to, any damages resulting from:
          </p>
          <ul className="list-disc list-inside text-[var(--text-secondary)] space-y-2 mt-4 ml-4">
            <li>Use or inability to use the Software</li>
            <li>Any actions taken by Riot Games regarding your account</li>
            <li>Inaccurate or incomplete statistical data</li>
            <li>Loss of data or profits</li>
          </ul>
        </section>

        <section>
          <h2 className="text-2xl font-display font-semibold text-[var(--pale-gold)] mb-4">7. Statistical Data</h2>
          <p className="text-[var(--text-secondary)] mb-4 leading-relaxed">
            The statistical information provided by GhostDraft is for informational purposes only.
            Win rates, build recommendations, and other statistics are based on aggregated
            match data and may not reflect current meta changes or individual skill levels.
          </p>
          <p className="text-[var(--text-secondary)] leading-relaxed">
            Users should use their own judgment when making in-game decisions and not rely
            solely on the Software&apos;s recommendations.
          </p>
        </section>

        <section>
          <h2 className="text-2xl font-display font-semibold text-[var(--pale-gold)] mb-4">8. Updates and Changes</h2>
          <p className="text-[var(--text-secondary)] leading-relaxed">
            We reserve the right to modify or discontinue the Software at any time without notice.
            We may also update these Terms of Service from time to time. Continued use of the
            Software after any changes constitutes acceptance of the new terms.
          </p>
        </section>

        <section>
          <h2 className="text-2xl font-display font-semibold text-[var(--pale-gold)] mb-4">9. User Conduct</h2>
          <p className="text-[var(--text-secondary)] mb-4 leading-relaxed">
            You agree not to:
          </p>
          <ul className="list-disc list-inside text-[var(--text-secondary)] space-y-2 ml-4">
            <li>Use the Software for any illegal purpose</li>
            <li>Modify the Software to violate Riot Games&apos; Terms of Service</li>
            <li>Distribute modified versions that could harm users or violate terms</li>
            <li>Use the Software to harass or harm other players</li>
          </ul>
        </section>

        <section>
          <h2 className="text-2xl font-display font-semibold text-[var(--pale-gold)] mb-4">10. Third-Party Services</h2>
          <p className="text-[var(--text-secondary)] leading-relaxed">
            GhostDraft relies on third-party services including the Riot Games API and Data Dragon.
            These services are subject to their own terms and conditions. We are not responsible
            for the availability or accuracy of third-party services.
          </p>
        </section>

        <section>
          <h2 className="text-2xl font-display font-semibold text-[var(--pale-gold)] mb-4">11. Termination</h2>
          <p className="text-[var(--text-secondary)] leading-relaxed">
            You may stop using the Software at any time by uninstalling it from your computer.
            These terms will survive termination with respect to any actions taken while
            using the Software.
          </p>
        </section>

        <section>
          <h2 className="text-2xl font-display font-semibold text-[var(--pale-gold)] mb-4">12. Contact</h2>
          <p className="text-[var(--text-secondary)] leading-relaxed">
            For questions about these Terms of Service, please open an issue on our{" "}
            <Link
              href="https://github.com/MatthewTran22/LoLOverlay/issues"
              target="_blank"
              rel="noopener noreferrer"
              className="text-[var(--hextech-gold)] hover:text-[var(--bright-gold)] transition-colors hover-line"
            >
              GitHub repository
            </Link>
            .
          </p>
        </section>

        <section className="hex-card rounded-xl p-6 mt-12">
          <h2 className="text-xl font-display font-semibold text-[var(--pale-gold)] mb-4">Riot Games Legal</h2>
          <p className="text-[var(--text-muted)] text-sm leading-relaxed">
            GhostDraft is not endorsed by Riot Games and does not reflect the views or opinions
            of Riot Games or anyone officially involved in producing or managing League of Legends.
            League of Legends and Riot Games are trademarks or registered trademarks of Riot Games, Inc.
          </p>
        </section>
      </div>
    </div>
  );
}
