import { NextResponse } from "next/server";
import { fetchAllMatchups, fetchCounterMatchups, fetchMatchup } from "@/lib/stats";

export async function GET(
  request: Request,
  { params }: { params: Promise<{ id: string }> }
) {
  try {
    const { id } = await params;
    const championId = parseInt(id, 10);

    if (isNaN(championId)) {
      return NextResponse.json({ error: "Invalid champion ID" }, { status: 400 });
    }

    const { searchParams } = new URL(request.url);
    const role = searchParams.get("role") || "middle";
    const enemyId = searchParams.get("enemy");
    const counters = searchParams.get("counters") === "true";
    const limit = parseInt(searchParams.get("limit") || "10", 10);

    // If specific enemy requested
    if (enemyId) {
      const enemyChampionId = parseInt(enemyId, 10);
      if (isNaN(enemyChampionId)) {
        return NextResponse.json({ error: "Invalid enemy champion ID" }, { status: 400 });
      }

      const matchup = fetchMatchup(championId, enemyChampionId, role);
      if (!matchup) {
        return NextResponse.json({ error: "No matchup data found" }, { status: 404 });
      }

      return NextResponse.json(matchup);
    }

    // If counters requested
    if (counters) {
      const data = fetchCounterMatchups(championId, role, limit);
      return NextResponse.json({ championId, role, counters: data });
    }

    // Return all matchups
    const data = fetchAllMatchups(championId, role);
    return NextResponse.json({ championId, role, matchups: data });
  } catch (error) {
    console.error("Failed to fetch matchup data:", error);
    return NextResponse.json({ error: "Failed to fetch matchup data" }, { status: 500 });
  }
}
