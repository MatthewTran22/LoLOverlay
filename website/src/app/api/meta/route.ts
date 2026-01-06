import { NextResponse } from "next/server";
import { fetchAllRolesTopChampions, fetchTopChampionsByRole } from "@/lib/stats";

export async function GET(request: Request) {
  try {
    const { searchParams } = new URL(request.url);
    const role = searchParams.get("role");
    const limit = parseInt(searchParams.get("limit") || "5", 10);

    // If specific role requested
    if (role) {
      const data = await fetchTopChampionsByRole(role, limit);
      return NextResponse.json({ role, champions: data });
    }

    // Return all roles
    const data = await fetchAllRolesTopChampions(limit);
    return NextResponse.json(data);
  } catch (error) {
    console.error("Failed to fetch meta data:", error);
    return NextResponse.json({ error: "Failed to fetch meta data" }, { status: 500 });
  }
}
