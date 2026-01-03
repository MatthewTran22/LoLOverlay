import { NextResponse } from "next/server";
import { getStatsInfo } from "@/lib/stats";
import { checkForUpdates, forceUpdate } from "@/lib/db";

export async function GET() {
  try {
    const info = getStatsInfo();
    return NextResponse.json(info);
  } catch (error) {
    console.error("Failed to get stats info:", error);
    return NextResponse.json({ error: "Failed to get stats info" }, { status: 500 });
  }
}

export async function POST(request: Request) {
  try {
    const body = await request.json();
    const { action } = body;

    if (action === "update") {
      await checkForUpdates();
      const info = getStatsInfo();
      return NextResponse.json({ success: true, ...info });
    }

    if (action === "force-update") {
      await forceUpdate();
      const info = getStatsInfo();
      return NextResponse.json({ success: true, ...info });
    }

    return NextResponse.json({ error: "Unknown action" }, { status: 400 });
  } catch (error) {
    console.error("Failed to update stats:", error);
    return NextResponse.json({ error: "Failed to update stats" }, { status: 500 });
  }
}
