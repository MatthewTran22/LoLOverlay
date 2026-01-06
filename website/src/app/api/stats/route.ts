import { NextResponse } from "next/server";
import { getStatsInfo } from "@/lib/stats";

export async function GET() {
  try {
    const info = await getStatsInfo();
    return NextResponse.json(info);
  } catch (error) {
    console.error("Failed to get stats info:", error);
    return NextResponse.json({ error: "Failed to get stats info" }, { status: 500 });
  }
}

export async function POST() {
  // Website is now read-only - data updates come from the Data Analyzer
  return NextResponse.json(
    { error: "Database updates are disabled. Data is managed by the Data Analyzer." },
    { status: 403 }
  );
}
