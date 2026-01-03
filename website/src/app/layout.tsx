import type { Metadata } from "next";
import "./globals.css";
import { Header } from "@/components/Header";
import { Footer } from "@/components/Footer";

export const metadata: Metadata = {
  title: "GhostDraft - League of Legends Champion Select Assistant",
  description: "Real-time champion select overlay providing matchup data, build recommendations, and team composition analysis for League of Legends.",
  keywords: ["League of Legends", "LoL", "champion select", "overlay", "matchup", "build", "statistics"],
};

export default function RootLayout({
  children,
}: Readonly<{
  children: React.ReactNode;
}>) {
  return (
    <html lang="en" className="dark">
      <body className="font-body antialiased min-h-screen flex flex-col hex-pattern">
        <Header />
        <main className="flex-1">
          {children}
        </main>
        <Footer />
      </body>
    </html>
  );
}
