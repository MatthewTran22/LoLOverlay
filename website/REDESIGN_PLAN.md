# GhostDraft Website Redesign Plan

## Design Direction: Hextech Arcane

A dark, mystical gaming aesthetic inspired by League's Hextech/Arcane visual language.

### Color Palette
- **Void Black**: #050810 (deepest background)
- **Abyss**: #0a0e17 (primary background)
- **Deep Navy**: #0d1829 (card backgrounds)
- **Arcane Blue**: #152238 (elevated surfaces)
- **Hextech Gold**: #c9a227 (primary accent)
- **Bright Gold**: #f0c14b (highlights)
- **Pale Gold**: #ffe4a0 (text highlights)
- **Arcane Cyan**: #00d4ff (secondary accent)

### Typography
- **Display Font**: Cinzel (regal, fantasy headers)
- **Body Font**: Rajdhani (sleek gaming body text)

### Key Visual Elements
1. Hexagonal patterns as subtle backgrounds
2. Glowing gold borders and accents
3. Animated gradient text for headlines
4. Floating particle effects
5. Cards with gold glow on hover
6. Corner accent decorations
7. Clip-path buttons (angled corners)

---

## Files to Update

### 1. globals.css
Replace with Hextech theme:
- CSS variables for colors
- Import Cinzel + Rajdhani fonts
- .hex-card class with glow effects
- .gradient-text animated class
- .text-glow for glowing text
- .reveal animations with delays
- .hover-line animated underlines
- Custom scrollbar styling
- Win rate color classes (.wr-high, .wr-mid, .wr-low)

### 2. layout.tsx
- Update font imports to use Cinzel + Rajdhani
- Add hex-pattern class to body
- Keep Header/Footer structure

### 3. Header.tsx - Redesign
```tsx
- Hextech-styled logo with glow effect
- Navigation with hover-line animations
- Transparent background with blur
- Gold accent border on bottom
```

### 4. Footer.tsx - Redesign
```tsx
- Matching dark theme
- Gold divider line at top
- Subtle hex pattern background
- Glowing social/link icons
```

### 5. page.tsx (Landing) - Full Redesign

**Hero Section:**
- Large Cinzel headline with gradient-text animation
- Floating particle background effect
- Two CTA buttons (hextech style + outline)
- Subtle hex pattern overlay

**Features Section:**
- 6 hex-cards in grid
- Icons with gold glow
- Staggered reveal animations
- Corner accent decorations

**Preview Section:**
- App mockup with glowing border
- Floating animation effect
- Tab UI matching app design

**How It Works:**
- 3 numbered steps
- Connecting line between steps (gold gradient)
- Icon circles with pulse-glow animation

**Privacy Highlights:**
- Asymmetric 2-column layout
- Checkmark list with cyan accents
- Data sources card with runic border

**Download CTA:**
- Large centered section
- Animated background gradient
- Primary hextech button
- Version/platform info below

### 6. stats/page.tsx - Redesign
- Role tabs with hextech styling
- Champion cards showing:
  - Rank badge (1-5)
  - Champion icon with gold border
  - Name in Rajdhani
  - Win rate with color coding
  - Pick rate stat
- Tier indicators (S/A/B with glow colors)
- Loading shimmer effect

### 7. privacy/page.tsx & terms/page.tsx
- Consistent dark theme
- Section headers in Cinzel
- Body text in Rajdhani
- Gold accent dividers between sections
- Highlighted boxes for important info

---

## CSS Classes to Create

```css
/* Core */
.hex-pattern - hexagonal SVG background
.hex-card - card with gold border + hover glow
.gradient-text - animated gold gradient text
.text-glow - gold text shadow
.text-glow-cyan - cyan text shadow

/* Buttons */
.btn-hextech - primary gold button with clip-path
.btn-outline - outlined gold button

/* Animations */
.reveal + .reveal-delay-1 through 6
.pulse-glow - pulsing box shadow
.float - gentle floating motion
.hover-line - underline on hover

/* Stats */
.wr-high - green for >52%
.wr-mid - gold for 50-52%
.wr-low - red for <50%
```

---

## Implementation Order

1. globals.css - Base theme and utilities
2. layout.tsx - Font setup
3. Header.tsx - Navigation redesign
4. Footer.tsx - Footer redesign
5. page.tsx - Landing page (largest)
6. stats/page.tsx - Stats display
7. privacy + terms pages - Content pages

---

## After Session Restart

1. Read each file before editing
2. Apply the Hextech theme systematically
3. Test build after each major change
4. Verify responsive design works
