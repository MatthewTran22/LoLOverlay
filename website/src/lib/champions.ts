// Champion ID to name mapping
export const championNames: Record<number, string> = {
  266: "Aatrox",
  103: "Ahri",
  84: "Akali",
  166: "Akshan",
  12: "Alistar",
  32: "Amumu",
  34: "Anivia",
  1: "Annie",
  523: "Aphelios",
  22: "Ashe",
  136: "Aurelion Sol",
  268: "Azir",
  432: "Bard",
  200: "Bel'Veth",
  53: "Blitzcrank",
  63: "Brand",
  201: "Braum",
  233: "Briar",
  51: "Caitlyn",
  164: "Camille",
  69: "Cassiopeia",
  31: "Cho'Gath",
  42: "Corki",
  122: "Darius",
  131: "Diana",
  119: "Draven",
  36: "Dr. Mundo",
  245: "Ekko",
  60: "Elise",
  28: "Evelynn",
  81: "Ezreal",
  9: "Fiddlesticks",
  114: "Fiora",
  105: "Fizz",
  3: "Galio",
  41: "Gangplank",
  86: "Garen",
  150: "Gnar",
  79: "Gragas",
  104: "Graves",
  887: "Gwen",
  120: "Hecarim",
  74: "Heimerdinger",
  910: "Hwei",
  420: "Illaoi",
  39: "Irelia",
  427: "Ivern",
  40: "Janna",
  59: "Jarvan IV",
  24: "Jax",
  126: "Jayce",
  202: "Jhin",
  222: "Jinx",
  145: "Kai'Sa",
  429: "Kalista",
  43: "Karma",
  30: "Karthus",
  38: "Kassadin",
  55: "Katarina",
  10: "Kayle",
  141: "Kayn",
  85: "Kennen",
  121: "Kha'Zix",
  203: "Kindred",
  240: "Kled",
  96: "Kog'Maw",
  897: "K'Sante",
  7: "LeBlanc",
  64: "Lee Sin",
  89: "Leona",
  876: "Lillia",
  127: "Lissandra",
  236: "Lucian",
  117: "Lulu",
  99: "Lux",
  54: "Malphite",
  90: "Malzahar",
  57: "Maokai",
  11: "Master Yi",
  902: "Milio",
  21: "Miss Fortune",
  62: "Wukong",
  82: "Mordekaiser",
  25: "Morgana",
  950: "Naafiri",
  267: "Nami",
  75: "Nasus",
  111: "Nautilus",
  518: "Neeko",
  76: "Nidalee",
  895: "Nilah",
  56: "Nocturne",
  20: "Nunu & Willump",
  2: "Olaf",
  61: "Orianna",
  516: "Ornn",
  80: "Pantheon",
  78: "Poppy",
  555: "Pyke",
  246: "Qiyana",
  133: "Quinn",
  497: "Rakan",
  33: "Rammus",
  421: "Rek'Sai",
  526: "Rell",
  888: "Renata Glasc",
  58: "Renekton",
  107: "Rengar",
  92: "Riven",
  68: "Rumble",
  13: "Ryze",
  360: "Samira",
  113: "Sejuani",
  235: "Senna",
  147: "Seraphine",
  875: "Sett",
  35: "Shaco",
  98: "Shen",
  102: "Shyvana",
  27: "Singed",
  14: "Sion",
  15: "Sivir",
  72: "Skarner",
  901: "Smolder",
  37: "Sona",
  16: "Soraka",
  50: "Swain",
  517: "Sylas",
  134: "Syndra",
  223: "Tahm Kench",
  163: "Taliyah",
  91: "Talon",
  44: "Taric",
  17: "Teemo",
  412: "Thresh",
  18: "Tristana",
  48: "Trundle",
  23: "Tryndamere",
  4: "Twisted Fate",
  29: "Twitch",
  77: "Udyr",
  6: "Urgot",
  110: "Varus",
  67: "Vayne",
  45: "Veigar",
  161: "Vel'Koz",
  711: "Vex",
  254: "Vi",
  234: "Viego",
  112: "Viktor",
  8: "Vladimir",
  106: "Volibear",
  19: "Warwick",
  498: "Xayah",
  101: "Xerath",
  5: "Xin Zhao",
  157: "Yasuo",
  777: "Yone",
  83: "Yorick",
  350: "Yuumi",
  154: "Zac",
  238: "Zed",
  221: "Zeri",
  115: "Ziggs",
  26: "Zilean",
  142: "Zoe",
  143: "Zyra",
};

// Special cases where champion key differs from name for Data Dragon
const specialCases: Record<number, string> = {
  62: "MonkeyKing",
  20: "Nunu",
  31: "Chogath",
  9: "FiddleSticks",
  121: "Khazix",
  96: "KogMaw",
  7: "Leblanc",
  64: "LeeSin",
  21: "MissFortune",
  421: "RekSai",
  223: "TahmKench",
  4: "TwistedFate",
  161: "Velkoz",
  5: "XinZhao",
  200: "Belveth",
  897: "KSante",
  136: "AurelionSol",
  59: "JarvanIV",
  240: "Kled",
  518: "Neeko",
  888: "Renata",
  360: "Samira",
  235: "Senna",
  147: "Seraphine",
  517: "Sylas",
  163: "Taliyah",
  350: "Yuumi",
  234: "Viego",
};

export function getChampionName(id: number): string {
  return championNames[id] || `Champion ${id}`;
}

export function getChampionKey(id: number): string {
  if (specialCases[id]) {
    return specialCases[id];
  }

  const name = championNames[id];
  if (!name) return "Unknown";

  // Remove special characters and spaces
  return name.replace(/[^a-zA-Z]/g, "");
}

export function getChampionIcon(id: number): string {
  return `https://ddragon.leagueoflegends.com/cdn/14.24.1/img/champion/${getChampionKey(id)}.png`;
}

export function getChampionIdByName(name: string): number | null {
  const normalizedName = name.toLowerCase().replace(/[^a-z]/g, "");
  for (const [id, champName] of Object.entries(championNames)) {
    if (champName.toLowerCase().replace(/[^a-z]/g, "") === normalizedName) {
      return parseInt(id);
    }
  }
  return null;
}

export function getAllChampionIds(): number[] {
  return Object.keys(championNames).map((id) => parseInt(id));
}

export const roleDisplayNames: Record<string, string> = {
  top: "Top",
  jungle: "Jungle",
  middle: "Mid",
  bottom: "ADC",
  utility: "Support",
};

export const roleIcons: Record<string, string> = {
  top: "T",
  jungle: "J",
  middle: "M",
  bottom: "B",
  utility: "S",
};

export function getWinRateClass(winRate: number): string {
  if (winRate >= 52) return "wr-high";
  if (winRate >= 50) return "wr-mid";
  return "wr-low";
}

export function getTier(winRate: number, pickRate: number): string {
  if (winRate >= 53 && pickRate >= 3) return "S+";
  if (winRate >= 52 && pickRate >= 2) return "S";
  if (winRate >= 51 && pickRate >= 1) return "A";
  if (winRate >= 50) return "B";
  if (winRate >= 48) return "C";
  return "D";
}

export function getTierColor(tier: string): string {
  switch (tier) {
    case "S+":
      return "text-[#ff6b6b]";
    case "S":
      return "text-[var(--hextech-gold)]";
    case "A":
      return "text-[#4ade80]";
    case "B":
      return "text-[var(--arcane-cyan)]";
    case "C":
      return "text-[var(--text-secondary)]";
    default:
      return "text-[var(--text-muted)]";
  }
}
