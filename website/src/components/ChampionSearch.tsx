"use client";

import { useState, useEffect, useRef } from "react";
import { useRouter } from "next/navigation";

interface Champion {
  id: number;
  name: string;
  icon: string;
}

interface ChampionSearchProps {
  compact?: boolean;
}

export default function ChampionSearch({ compact = false }: ChampionSearchProps) {
  const [query, setQuery] = useState("");
  const [champions, setChampions] = useState<Champion[]>([]);
  const [filtered, setFiltered] = useState<Champion[]>([]);
  const [isOpen, setIsOpen] = useState(false);
  const [selectedIndex, setSelectedIndex] = useState(0);
  const inputRef = useRef<HTMLInputElement>(null);
  const dropdownRef = useRef<HTMLDivElement>(null);
  const router = useRouter();

  // Fetch champions on mount
  useEffect(() => {
    fetch("/api/champions")
      .then((res) => res.json())
      .then((data) => setChampions(data))
      .catch(console.error);
  }, []);

  // Filter champions when query changes
  useEffect(() => {
    if (!query.trim()) {
      setFiltered([]);
      setIsOpen(false);
      return;
    }

    const q = query.toLowerCase();
    const matches = champions.filter((c) =>
      c.name.toLowerCase().includes(q)
    ).slice(0, 8);

    setFiltered(matches);
    setIsOpen(matches.length > 0);
    setSelectedIndex(0);
  }, [query, champions]);

  // Close dropdown when clicking outside
  useEffect(() => {
    const handleClickOutside = (e: MouseEvent) => {
      if (
        dropdownRef.current &&
        !dropdownRef.current.contains(e.target as Node) &&
        inputRef.current &&
        !inputRef.current.contains(e.target as Node)
      ) {
        setIsOpen(false);
      }
    };

    document.addEventListener("mousedown", handleClickOutside);
    return () => document.removeEventListener("mousedown", handleClickOutside);
  }, []);

  const handleSelect = (champion: Champion) => {
    setQuery("");
    setIsOpen(false);
    router.push(`/stats/${champion.id}`);
  };

  const handleKeyDown = (e: React.KeyboardEvent) => {
    if (!isOpen) return;

    switch (e.key) {
      case "ArrowDown":
        e.preventDefault();
        setSelectedIndex((i) => (i + 1) % filtered.length);
        break;
      case "ArrowUp":
        e.preventDefault();
        setSelectedIndex((i) => (i - 1 + filtered.length) % filtered.length);
        break;
      case "Enter":
        e.preventDefault();
        if (filtered[selectedIndex]) {
          handleSelect(filtered[selectedIndex]);
        }
        break;
      case "Escape":
        setIsOpen(false);
        break;
    }
  };

  return (
    <div className={`relative w-full ${compact ? "" : "max-w-md mx-auto"}`}>
      <div className="relative">
        <input
          ref={inputRef}
          type="text"
          value={query}
          onChange={(e) => setQuery(e.target.value)}
          onKeyDown={handleKeyDown}
          onFocus={() => query && filtered.length > 0 && setIsOpen(true)}
          placeholder="Search champions..."
          className={`w-full bg-[var(--abyss)] border border-[var(--hextech-gold)]/30 rounded-lg text-[var(--text-primary)] placeholder-[var(--text-muted)] focus:outline-none focus:border-[var(--hextech-gold)] transition-colors ${
            compact ? "px-3 py-2 pl-9 text-sm" : "px-4 py-3 pl-12"
          }`}
        />
        <svg
          className={`absolute top-1/2 -translate-y-1/2 text-[var(--text-muted)] ${
            compact ? "left-3 w-4 h-4" : "left-4 w-5 h-5"
          }`}
          fill="none"
          stroke="currentColor"
          viewBox="0 0 24 24"
        >
          <path
            strokeLinecap="round"
            strokeLinejoin="round"
            strokeWidth={2}
            d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z"
          />
        </svg>
      </div>

      {isOpen && (
        <div
          ref={dropdownRef}
          className="absolute top-full left-0 right-0 mt-2 bg-[var(--abyss)] border border-[var(--hextech-gold)]/30 rounded-lg overflow-hidden shadow-xl z-50"
        >
          {filtered.map((champion, index) => (
            <button
              key={champion.id}
              onClick={() => handleSelect(champion)}
              onMouseEnter={() => setSelectedIndex(index)}
              className={`w-full flex items-center gap-3 text-left transition-colors ${
                compact ? "px-3 py-2" : "px-4 py-3"
              } ${
                index === selectedIndex
                  ? "bg-[var(--hextech-gold)]/20 text-[var(--pale-gold)]"
                  : "text-[var(--text-primary)] hover:bg-[var(--hextech-gold)]/10"
              }`}
            >
              <img
                src={champion.icon}
                alt={champion.name}
                className={`rounded-full border border-[var(--hextech-gold)]/30 ${
                  compact ? "w-6 h-6" : "w-8 h-8"
                }`}
              />
              <span className={compact ? "text-sm" : "font-medium"}>{champion.name}</span>
            </button>
          ))}
        </div>
      )}
    </div>
  );
}
