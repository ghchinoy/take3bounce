# Design System Strategy: Sunrise Studio (Light Theme)

## 1. Overview & Creative North Star: "Approachable Clarity"
This design system provides a happier, lighter, and more approachable alternative to the core "Precision Brutalism" dark theme. It trades the high-tech, intense DAW studio aesthetic for a clean, accessible, and friendly web application look.

---

## 2. Colors & Surface Philosophy
The palette is rooted in soft, clean whites and warm greys, punctuated by vibrant, optimistic accents.

### Surface Hierarchy & Nesting
*   **Base:** `surface` (#FAFAFA - Override Neutral) – The foundational light background.
*   **In-set Panels / Cards:** `surface-container` (Clean whites or very light greys) to provide soft contrast.
*   **Shadows:** Unlike the dark theme's tonal layering, this theme utilizes soft, diffused drop shadows (e.g., `0 4px 12px rgba(0, 0, 0, 0.05)`) to create depth and lift.

### The Palette
*   **Primary:** #4285F4 (Bright Blue) - Used for the main playhead, active states, and primary CTAs.
*   **Secondary:** #4CAF50 (Friendly Green) - Used for Take A or success states.
*   **Tertiary:** #03A9F4 (Bright Blue) - Used for Take B or informational states.
*   **Background / Neutral:** #FAFAFA.

---

## 3. Typography: Friendly & Readable
*   **Display & Headlines (Sora):** A geometric sans-serif that is crisp but friendly. Used for the main titles, headers, and "Take" names.
*   **Body & Script (Nunito Sans):** A highly readable, well-balanced sans-serif that is softer than Inter, making the dense technical scripts feel less intimidating.

---

## 4. Components
*   **Roundness:** `ROUND_FULL` - Moving away from the sharp, technical corners of the dark theme, the Sunrise Studio theme embraces fully rounded pills for buttons, tags, and meters, creating a softer, more organic feel.
*   **Borders:** Soft, 1px solid borders (e.g., `#E0E0E0`) are acceptable in this theme to clearly delineate cards and input fields, replacing the strict "No-Line" rule of the dark theme.