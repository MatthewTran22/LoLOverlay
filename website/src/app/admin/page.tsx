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
  const [updating, setUpdating] = useState(false);
  const [message, setMessage] = useState<{ type: "success" | "error"; text: string } | null>(null);

  const fetchStats = async () => {
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

  const handleUpdate = async (force: boolean) => {
    setUpdating(true);
    setMessage(null);

    try {
      const res = await fetch("/api/stats", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ action: force ? "force-update" : "update" }),
      });

      const data = await res.json();

      if (data.success) {
        setStats(data);
        setMessage({
          type: "success",
          text: force
            ? `Force updated to patch ${data.patch}`
            : data.patch
              ? `Updated to patch ${data.patch}`
              : "Already up to date",
        });
      } else {
        setMessage({ type: "error", text: data.error || "Update failed" });
      }
    } catch (err) {
      setMessage({ type: "error", text: "Failed to connect to server" });
      console.error(err);
    } finally {
      setUpdating(false);
    }
  };

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
              onClick={() => handleUpdate(false)}
              disabled={updating}
              className="btn-hextech w-full disabled:opacity-50 disabled:cursor-not-allowed"
            >
              {updating ? "Updating..." : "Check for Updates"}
            </button>
            <p className="text-[var(--text-muted)] text-sm mt-2">
              Compares remote manifest version with local. Downloads only if newer.
            </p>
          </div>

          <div>
            <button
              onClick={() => handleUpdate(true)}
              disabled={updating}
              className="btn-outline w-full disabled:opacity-50 disabled:cursor-not-allowed"
            >
              {updating ? "Updating..." : "Force Update"}
            </button>
            <p className="text-[var(--text-muted)] text-sm mt-2">
              Clears local data and re-downloads everything from remote.
            </p>
          </div>

          <div>
            <button
              onClick={fetchStats}
              disabled={loading}
              className="w-full px-4 py-2 bg-[var(--arcane-blue)] text-[var(--text-secondary)] rounded-lg hover:bg-[var(--mystic-slate)] transition-colors disabled:opacity-50"
            >
              Refresh Status
            </button>
          </div>
        </div>
      </div>

      {/* Message */}
      {message && (
        <div
          className={`rounded-xl p-4 ${
            message.type === "success"
              ? "bg-green-500/10 border border-green-500/30 text-green-400"
              : "bg-red-500/10 border border-red-500/30 text-red-400"
          }`}
        >
          {message.text}
        </div>
      )}

      {/* Info */}
      <div className="mt-8 text-[var(--text-muted)] text-sm">
        <p>
          Data is fetched from:{" "}
          <code className="text-[var(--arcane-cyan)]">
            github.com/MatthewTran22/LoLOverlay-Data
          </code>
        </p>
      </div>
    </div>
  );
}
