"use client";

import { useState, useEffect } from "react";
import Link from "next/link";

interface StatsInfo {
  patch: string;
  hasData: boolean;
  championCount: number;
  matchupCount: number;
}

export default function AdminPage() {
  const [stats, setStats] = useState<StatsInfo | null>(null);
  const [loading, setLoading] = useState(true);

  const fetchStats = async () => {
    setLoading(true);
    try {
      const res = await fetch("/api/stats");
      const data = await res.json();
      setStats(data);
    } catch (err) {
      console.error("Failed to fetch stats:", err);
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    fetchStats();
  }, []);

  return (
    <div className="max-w-2xl mx-auto px-6 py-16">
      <Link
        href="/stats"
        className="text-[var(--text-secondary)] hover:text-[var(--hextech-gold)] transition-colors hover-line mb-8 inline-block"
      >
        &larr; Back to Stats
      </Link>

      <h1 className="text-4xl font-display font-bold text-[var(--pale-gold)] mb-8 text-glow">
        Admin Panel
      </h1>

      {/* Stats Info Card */}
      <div className="hex-card rounded-xl p-6 mb-8">
        <h2 className="text-xl font-display font-semibold text-[var(--pale-gold)] mb-4">
          Database Status
        </h2>

        {loading ? (
          <div className="text-[var(--text-muted)]">Loading...</div>
        ) : stats ? (
          <div className="space-y-3">
            <div className="flex justify-between">
              <span className="text-[var(--text-secondary)]">Current Patch</span>
              <span className="text-[var(--arcane-cyan)] font-semibold text-glow-cyan">
                {stats.patch || "None"}
              </span>
            </div>
            <div className="flex justify-between">
              <span className="text-[var(--text-secondary)]">Has Data</span>
              <span
                className={
                  stats.hasData ? "text-green-400" : "text-red-400"
                }
              >
                {stats.hasData ? "Yes" : "No"}
              </span>
            </div>
            <div className="flex justify-between">
              <span className="text-[var(--text-secondary)]">Champions</span>
              <span className="text-[var(--text-primary)]">{stats.championCount}</span>
            </div>
            <div className="flex justify-between">
              <span className="text-[var(--text-secondary)]">Matchups</span>
              <span className="text-[var(--text-primary)]">
                {stats.matchupCount.toLocaleString()}
              </span>
            </div>
          </div>
        ) : (
          <div className="text-red-400">Failed to load stats</div>
        )}
      </div>

      {/* Actions Card */}
      <div className="hex-card rounded-xl p-6 mb-8">
        <h2 className="text-xl font-display font-semibold text-[var(--pale-gold)] mb-4">
          Actions
        </h2>

        <div className="space-y-4">
          <div>
            <button
              onClick={fetchStats}
              disabled={loading}
              className="btn-hextech w-full disabled:opacity-50 disabled:cursor-not-allowed"
            >
              {loading ? "Loading..." : "Refresh Status"}
            </button>
          </div>
        </div>
      </div>

      {/* Info */}
      <div className="hex-card rounded-xl p-6 bg-[var(--arcane-blue)]/20">
        <h3 className="text-lg font-display font-semibold text-[var(--pale-gold)] mb-3">
          Read-Only Mode
        </h3>
        <p className="text-[var(--text-secondary)] text-sm mb-3">
          This website connects to a Turso database in read-only mode. Data updates are managed
          by the Data Analyzer pipeline.
        </p>
        <p className="text-[var(--text-muted)] text-sm">
          To update data, run the Data Analyzer with Turso credentials configured.
        </p>
      </div>
    </div>
  );
}
