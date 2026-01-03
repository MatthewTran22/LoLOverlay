import { NextResponse } from "next/server";
import { fetchChampionData } from "@/lib/stats";

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
    const name = searchParams.get("name") || `Champion ${championId}`;

    const data = fetchChampionData(championId, name, role);

    if (!data) {
      return NextResponse.json(
        { error: "No data found for this champion/role" },
        { status: 404 }
      );
    }

    return NextResponse.json(data);
  } catch (error) {
    console.error("Failed to fetch champion data:", error);
    return NextResponse.json({ error: "Failed to fetch champion data" }, { status: 500 });
  }
}
