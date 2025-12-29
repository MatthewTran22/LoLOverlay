Product Requirements Document (PRD)
Project Name: LoL Draft Gap (Project LDG) Date: December 28, 2025 Version: 1.0 Status: Ready for Development

1. Executive Summary
LoL Draft Gap is a lightweight, standalone desktop utility for League of Legends. It replaces resource-heavy Overwolf applications by running as a native Go executable (<50MB RAM). It communicates directly with the League Client (LCU) to detect Champion Select and fetches meta-data (Runes, Builds, Counters) directly from U.GG's static CDN endpoints, bypassing the need for a web scraper or heavy browser instance.

2. Core Value Proposition
Performance: Zero impact on game FPS. No background Chromium processes.

Privacy: No ads, no user tracking, no "Overwolf" middleware.

Speed: Fetches pre-compiled JSON data from U.GG's CDN rather than scraping HTML.

3. Technical Architecture
3.1 Tech Stack
Backend: Go (Golang) 1.22+

Frontend/GUI: Wails (Uses native Windows WebView2).

LCU Communication: gorilla/websocket + net/http (Standard Library).

3.2 System Architecture Diagram
App Launch: Checks GitHub Gist for "Remote Config" (API Versions).

LCU Hook: Polls for lockfile, extracts Port/Password, connects via WSS.

Draft Detection: Listens for OnJsonApiEvent_lol-champ-select_v1_session.

Data Fetch: Constructs U.GG CDN URLs to fetch build data.

Action: Pushes Runes back to LCU via PUT /lol-perks/v1/pages.

4. Data Strategy (The U.GG Implementation)
The application will not scrape HTML. It will mimic the network requests of the U.GG frontend to fetch static JSON files.

4.1 Step 1: Dynamic Patch Detection
Endpoint: https://static.bigbrain.gg/assets/lol/riot_patch_update/prod/ugg/patches.json

Logic: Fetch this on launch. Read the first string in the data array.

Example Response: ["15_24", "15_23", ...] -> Current Patch: 15_24

4.2 Step 2: Champion Data Retrieval
URL Template:

Plaintext

https://stats2.u.gg/lol/1.5/overview/[PATCH]/ranked_solo_5x5/[CHAMP_ID]/1.5.0.json
Variables:

[PATCH]: Derived from Step 1 (e.g., 15_24).

[CHAMP_ID]: Derived from LCU Champ Select (e.g., 233 for Briar).

Target Data Points (JSON Unmarshal):

rec_runes -> primary_style, active_perks (IDs for Runes).

rec_starting_items -> ids.

rec_core_items -> ids.

counters -> List of Champion IDs with lowest win rates against target.

4.3 Remote Configuration (Kill Switch)
To prevent the app from breaking if U.GG changes their API version (e.g., 1.5 -> 1.6), the app must fetch a config file from a developer-controlled source (GitHub Gist) on launch.

Config Structure:

JSON

{
  "ugg_api_version": "1.5",
  "ugg_base_url": "https://stats2.u.gg/lol/",
  "is_kill_switch_active": false
}
5. Functional Requirements
5.1 Lifecycle Management
FR-01: App must detect LeagueClientUx.exe. If not found, display "Waiting for League...".

FR-02: When the match starts (LCU state: InProgress), the app must completely exit or minimize to tray to ensure 0% CPU usage.

5.2 Lobby & Draft Phase
FR-03 (Teammate Analysis):

On Lobby Join, fetch Summoner IDs.

(V1 Scope): Display basic "Win Rate" for teammates if available via LCU.

FR-04 (Champion Selection):

Detect when the user hovers or locks a champion.

Immediately fetch U.GG JSON for that champion.

FR-05 (Rune Import):

Display a button: "Push Runes".

On click, format the U.GG rune IDs into the LCU format and send PUT request.

Auto-Import Option: Checkbox to do this automatically on champion lock-in.

5.3 UI/UX Requirements
Window: Fixed size (e.g., 400x700), non-resizable.

Theme: Dark mode only (matching League client).

Error Handling: If U.GG data fails, display "Stats Unavailable" (do not crash).

6. Development Milestones
Proof of Concept (PoC):

Go app that finds lockfile and prints "League Connected".

Data Fetcher:

Go function that fetches patches.json then fetches Briar (233) data from U.GG and prints the Rune IDs to console.

Lobby Integration:

Connect PoC to LCU WebSocket and trigger Data Fetcher when a champion is hovered.

UI Implementation:

Build Wails frontend to display the data.

Rune Writer:

Implement the PUT request to write runes to the client.

7. Risks & Mitigations
Risk: U.GG adds Cloudflare protection/Captcha to the stats2 subdomain.

Mitigation: The app mimics browser User-Agent headers. If fully blocked, release update to switch data source to Lolalytics (Source B).

Risk: Riot changes the LCU WebSocket protocol.

Mitigation: Standard open-source libraries (lcu-connector) usually update quickly. The app is open to maintenance.