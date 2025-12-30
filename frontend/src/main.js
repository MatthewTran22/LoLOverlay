import './style.css';
import { GetConnectionStatus } from '../wailsjs/go/main/App';
import { EventsOn } from '../wailsjs/runtime/runtime';

// Initial HTML structure
document.querySelector('#app').innerHTML = `
    <div class="overlay-box" id="overlay-box">
        <div class="header drag-region">
            <h1>GhostDraft</h1>
        </div>

        <div class="status-card" id="status-card">
            <div class="status-indicator">
                <div class="status-dot waiting" id="status-dot"></div>
                <span class="status-message" id="status-message">Initializing...</span>
            </div>
        </div>

        <div class="teamcomp-card hidden" id="teamcomp-card">
            <div class="teamcomp-warning" id="teamcomp-warning"></div>
        </div>

        <div class="bans-card hidden" id="bans-card">
            <div class="bans-header">Recommended Bans</div>
            <div class="bans-subheader" id="bans-subheader"></div>
            <div class="bans-list" id="bans-list"></div>
        </div>

        <div class="build-card hidden" id="build-card">
            <div class="build-role" id="build-role"></div>
            <div class="build-matchup">
                <span class="winrate-label" id="winrate-label">Win Rate</span>
                <span class="build-winrate" id="build-winrate"></span>
            </div>
        </div>
    </div>
`;

// DOM elements
const statusDot = document.getElementById('status-dot');
const statusMessage = document.getElementById('status-message');
const statusCard = document.getElementById('status-card');
const teamcompCard = document.getElementById('teamcomp-card');
const teamcompWarning = document.getElementById('teamcomp-warning');
const bansCard = document.getElementById('bans-card');
const bansSubheader = document.getElementById('bans-subheader');
const bansList = document.getElementById('bans-list');
const buildCard = document.getElementById('build-card');
const buildRole = document.getElementById('build-role');
const buildWinrate = document.getElementById('build-winrate');
const winrateLabel = document.getElementById('winrate-label');

// Format role name for display
function formatRole(role) {
    const roleMap = {
        'utility': 'Support',
        'middle': 'Mid',
        'bottom': 'ADC',
        'jungle': 'Jungle',
        'top': 'Top'
    };
    return roleMap[role] || role || '';
}

// Update connection status
function updateStatus(status) {
    statusMessage.textContent = status.message;
    statusDot.className = status.connected ? 'status-dot connected' : 'status-dot waiting';
}

// Update champ select state
function updateChampSelect(data) {
    if (!data.inChampSelect) {
        buildCard.classList.add('hidden');
        bansCard.classList.add('hidden');
        teamcompCard.classList.add('hidden');
        statusCard.classList.remove('hidden');
        return;
    }
}

// Update team comp warning
function updateTeamComp(data) {
    if (!data || !data.show) {
        teamcompCard.classList.add('hidden');
        return;
    }

    teamcompCard.classList.remove('hidden');
    teamcompCard.className = `teamcomp-card ${data.severity}`;
    teamcompWarning.textContent = data.recommendation;
}

// Update recommended bans
function updateBans(data) {
    if (!data || !data.hasBans || !data.bans || data.bans.length === 0) {
        bansCard.classList.add('hidden');
        return;
    }

    // Show bans card
    bansCard.classList.remove('hidden');
    statusCard.classList.add('hidden');

    // Update header
    bansSubheader.textContent = `Counters for ${data.championName} (${formatRole(data.role)})`;

    // Build ban list HTML
    let html = '';
    for (let i = 0; i < data.bans.length; i++) {
        const ban = data.bans[i];
        const wr = typeof ban.winRate === 'number' ? ban.winRate.toFixed(1) : ban.winRate;
        const dmgClass = ban.damageType === 'AP' ? 'ap' : ban.damageType === 'AD' ? 'ad' : 'mixed';
        html += `
            <div class="ban-row">
                <span class="ban-rank">${i + 1}</span>
                <img class="ban-icon" src="${ban.iconURL}" alt="${ban.championName}" />
                <span class="ban-name">${ban.championName}</span>
                <span class="ban-dmg ${dmgClass}">${ban.damageType}</span>
                <span class="ban-wr losing">${wr}%</span>
            </div>
        `;
    }
    bansList.innerHTML = html;
}

// Update build/matchup data
function updateBuild(data) {
    if (!data.hasBuild) {
        buildCard.classList.add('hidden');
        return;
    }

    buildCard.classList.remove('hidden');
    statusCard.classList.add('hidden');

    buildRole.textContent = formatRole(data.role);
    winrateLabel.textContent = data.winRateLabel || 'Win Rate';
    buildWinrate.textContent = data.winRate;

    buildWinrate.classList.remove('winning', 'losing', 'even');
    if (data.matchupStatus) {
        buildWinrate.classList.add(data.matchupStatus);
    }
}

// Event listeners
EventsOn('lcu:status', updateStatus);
EventsOn('champselect:update', updateChampSelect);
EventsOn('build:update', updateBuild);
EventsOn('bans:update', updateBans);
EventsOn('teamcomp:update', updateTeamComp);

// Get initial status
GetConnectionStatus()
    .then(updateStatus)
    .catch(() => {
        updateStatus({ connected: false, message: 'Waiting for League...' });
    });
