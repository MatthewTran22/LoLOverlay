import { getChampionData, getDDragonVersion } from "@/lib/ddragon";
import { NextResponse } from "next/server";

export const revalidate = 3600; // Cache for 1 hour

export async function GET() {
  const [champions, version] = await Promise.all([
    getChampionData(),
    getDDragonVersion(),
  ]);

  const championList = Array.from(champions.entries()).map(([id, data]) => ({
    id,
    name: data.name,
    icon: `https://ddragon.leagueoflegends.com/cdn/${version}/img/champion/${data.key}.png`,
  }));

  // Sort alphabetically by name
  championList.sort((a, b) => a.name.localeCompare(b.name));

  return NextResponse.json(championList);
}
