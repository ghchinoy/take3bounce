# Design System Strategy: The Sonic Lab

## 1. Overview & Creative North Star: "Precision Brutalism"
This design system is built to bridge the gap between high-end digital audio workstations (DAWs) and cutting-edge AI interfaces. Our Creative North Star is **"Precision Brutalism."** We move away from the "friendly SaaS" look toward a sophisticated, technical aesthetic that prioritizes high information density and professional authority.

The system breaks the "template" look by using **intentional asymmetry**—aligning technical data to rigid grids while allowing script content and waveforms to breathe. We replace standard borders with tonal layering and neon-infused accents to create a "High-Tech Studio" environment that feels like a physical piece of premium hardware.

---

## 2. Colors & Surface Philosophy
The palette is rooted in deep obsidian tones, punctuated by high-frequency accents that represent different "Takes" or audio states.

### The "No-Line" Rule
**Explicit Instruction:** Do not use 1px solid borders to section off the interface. 
Boundaries must be defined solely through background color shifts. For example, a `surface-container-low` panel sitting on a `surface` background creates a natural edge. This forces a cleaner, more integrated look that mimics high-end studio gear.

### Surface Hierarchy & Nesting
Treat the UI as a series of physical layers using the Surface Tiers:
*   **Base:** `surface` (#0e0e0f) – The foundation.
*   **In-set Panels:** `surface-container-lowest` (#000000) – Used for "wells" like the waveform editor or script input.
*   **Raised Modules:** `surface-container-high` (#201f21) – For floating toolbars or active sidebars.

### The "Glass & Gradient" Rule
To add "soul" to the technical grid, use **Glassmorphism** for floating overlays (e.g., transport controls). Apply a `surface-container` color at 70% opacity with a `20px` backdrop-blur. 
*   **Signature Textures:** Main CTAs or active "Takes" should utilize a linear gradient from `primary` (#8ff5ff) to `primary-container` (#00eefc) at a 135-degree angle.

---

## 3. Typography: Technical Clarity
We pair the architectural `Space Grotesk` with the functional `Inter` to create a hierarchy that feels like a blueprint.

*   **Display & Headlines (Space Grotesk):** Used for global navigation and take IDs. Its geometric nature feels engineered. Use `display-sm` for "Take" numbers to give them a monumental presence.
*   **Script & Data (Inter):** While the request suggests monospaced for scripts, we use `Inter` with increased letter-spacing or a monospace font-variant for technical readouts to maintain legibility at `body-md` (0.875rem) while feeling "code-like."
*   **Labels (Space Grotesk):** All `label-sm` elements should be **all-caps** with a 0.05em tracking to evoke the look of etched labels on a mixing console.

---

## 4. Elevation & Depth
We eschew traditional shadows in favor of **Tonal Layering**.

*   **The Layering Principle:** Depth is achieved by "stacking." Place a `surface-container-high` card on a `surface-container-low` section. The contrast in value creates the lift.
*   **Ambient Shadows:** If an element must float (e.g., a context menu), use an extra-diffused shadow: `0px 24px 48px rgba(0, 0, 0, 0.5)`. Never use pure black shadows; tint them with a hint of the `primary` hue for a "glow" effect.
*   **The "Ghost Border":** If a separator is functionally required for accessibility, use the `outline-variant` (#484849) at **15% opacity**. It should be felt, not seen.

---

## 5. Components

### Buttons & Controls
*   **Primary Action:** A solid `primary` (#8ff5ff) fill with `on-primary` (#005d63) text. Use `sm` (0.125rem) roundedness for a sharper, more technical corner.
*   **Secondary "Take" Buttons:** Use `secondary` (#c97cff) and `tertiary` (#8eff71) to differentiate audio versions. These should be high-contrast and neon-bright against the dark surface.

### Technical Input Fields
*   **The "Well" Input:** Instead of a boxed field, use `surface-container-lowest` (#000000) as the background. Use a `primary` 2px bottom-bar when the field is focused, rather than a full outline.

### Cards & Lists (The Waveform Row)
*   **Constraint:** Forbid the use of divider lines between list items.
*   **The Solution:** Use a 4px vertical gap between list items, allowing the `background` color to act as a natural separator. Each list item should be a `surface-container-low` module.

### Audio Specific Components
*   **The "Take" Indicator:** A small, high-saturation pill using `secondary_fixed` (#e9c4ff) to mark active AI-generated takes.
*   **The Metering Bar:** Use a gradient from `tertiary` (#8eff71) to `error` (#ff716c) for peak indicators, stripped of all containers—just raw data against the black surface.

---

## 6. Do’s and Don’ts

### Do:
*   **Use Mono-spacing for numbers:** Ensure all timing codes and DB levels use monospaced numerals for alignment.
*   **Embrace the "Darkness":** Allow large areas of `#0e0e0f` to exist. Negative space in a dark theme conveys premium quality.
*   **Use Precise Accents:** Use `primary` (#8ff5ff) only for the "Playhead" or critical interactive states.

### Don’t:
*   **No Rounded Corners > 8px:** Avoid `xl` or `full` roundedness except for specific status pills. Large radii feel too "bubbly" for a professional studio tool.
*   **No Standard Greys:** Never use a neutral grey. Every "grey" in this system is slightly blue-tinted or purple-tinted to maintain the "High-Tech" atmosphere.
*   **No Over-shadowing:** If the surface hierarchy is correct, you should almost never need a drop shadow.