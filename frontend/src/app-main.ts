import { LitElement, html, css } from 'lit';
import { customElement, state } from 'lit/decorators.js';
import '@material/web/textfield/filled-text-field.js';
import '@material/web/button/filled-button.js';
import '@material/web/button/outlined-button.js';
import '@material/web/progress/circular-progress.js';
import '@material/web/iconbutton/icon-button.js';

// Import lit-text-ui components
import '@ghchinoy/lit-text-ui';

import './variation-card.js';
import type { Variation } from './variation-card.js';

const PRESETS = [
  {
    label: "Kittens",
    text: "Our kittens are raised in a cage-free environment with 24/7 medical supervision."
  },
  {
    label: "Hamlet",
    text: "To be, or not to be, that is the question: Whether 'tis nobler in the mind to suffer the slings and arrows of outrageous fortune, or to take arms against a sea of troubles and by opposing end them."
  },
  {
    label: "Crazy Ones",
    text: "Here's to the crazy ones. The misfits. The rebels. The troublemakers. The round pegs in the square holes. The ones who see things differently."
  }
];

export interface VoiceActor {
  shortName: string;
  baseVoice: string;
  stylePrompt: string;
}

const VOICE_ACTORS: VoiceActor[] = [
  {
    shortName: "Cosmic Curator",
    baseVoice: "Orus",
    stylePrompt: "# AUDIO PROFILE: Cosmic Curator, Celestial Sage\n## THE SCENE: A darkened observatory overlooking a nebula of swirling starlight. The air is cool and still.\n### DIRECTOR'S NOTES\n* Style: Reverent, slightly breathless with awe\n* Pace: Deliberate and rhythmic\n* Accent: Standard American",
  },
  {
    shortName: "Auntie Mae",
    baseVoice: "Leda",
    stylePrompt: "# AUDIO PROFILE: Auntie Mae, The Matriarch\n## THE SCENE: A sun-drenched kitchen smelling of fresh eucalyptus and baking bread. A ceiling fan whirs softly overhead.\n### DIRECTOR'S NOTES\n* Style: Kind, welcoming, with a wide vocal smile\n* Pace: Relaxed and conversational\n* Accent: Rural Australian",
  },
  {
    shortName: "Glitch",
    baseVoice: "Fenrir",
    stylePrompt: "# AUDIO PROFILE: Glitch, Urban Rogue\n## THE SCENE: A neon-lit alleyway drenched in rain with the constant hum of hover-traffic. The air is thick with ozone.\n### DIRECTOR'S NOTES\n* Style: Paranoiac, urgent, and sharp\n* Pace: Rapid-fire and staccato\n* Accent: Estuary English",
  },
  {
    shortName: "Jaz R.",
    baseVoice: "Puck",
    stylePrompt: "# AUDIO PROFILE: Jaz R.\n## \"The Morning Hype\"\n## THE SCENE: The London Studio\nIt is 10:00 PM in a glass-walled studio overlooking the moonlit London skyline, but inside, it is blindingly bright.\n### DIRECTOR'S NOTES\n* Style: The \"Vocal Smile\": You must hear the grin in the audio.\n* Pace: High-speed delivery with fluid transitions.\n* Accent: Jaz is from Brixton, London.",
  }
];

/**
 * AppMain is the primary layout and orchestrator component for the Three-Up VO Generator.
 * It manages the global theme state, holds the script presets and Voice Actor list,
 * and handles the primary fetch request to the /api/variations backend to generate takes.
 */
@customElement('app-main')
export class AppMain extends LitElement {
  @state()
  private paragraph: string = PRESETS[0].text;

  @state()
  private selectedVoiceActor: VoiceActor = VOICE_ACTORS[0];

  @state()
  private variations: Variation[] = [];

  @state()
  private loading: boolean = false;

  @state()
  private error: string | null = null;

  @state()
  private isLightMode: boolean = false;

  connectedCallback() {
    super.connectedCallback();
    this.isLightMode = localStorage.getItem('theme') === 'light' || 
                       (!localStorage.getItem('theme') && window.matchMedia('(prefers-color-scheme: light)').matches);
    this._applyTheme();
  }

  private _toggleTheme() {
    this.isLightMode = !this.isLightMode;
    localStorage.setItem('theme', this.isLightMode ? 'light' : 'dark');
    this._applyTheme();
  }

  private _applyTheme() {
    if (this.isLightMode) {
      document.body.classList.add('light-theme');
    } else {
      document.body.classList.remove('light-theme');
    }
  }
  static styles = css`
    :host {
      display: block;
      max-width: 1400px;
      margin: 0 auto;
      padding: 2rem;
      font-family: var(--theme-font-body);
    }
    .material-symbols-outlined {
      font-family: 'Material Symbols Outlined';
      font-weight: normal;
      font-style: normal;
      font-size: 24px;
      line-height: 1;
      letter-spacing: normal;
      text-transform: none;
      display: inline-block;
      white-space: nowrap;
      word-wrap: normal;
      direction: ltr;
      -webkit-font-smoothing: antialiased;
    }
    h1 {
      font-family: var(--theme-font-headline);
      letter-spacing: -0.02em;
      color: var(--md-sys-color-primary);
      text-transform: uppercase;
    }
    .header {
      display: flex;
      flex-direction: column;
      align-items: center;
      position: relative;
      margin-bottom: 2rem;
    }
    .header-actions {
      position: absolute;
      right: 0;
      top: 0;
      display: flex;
      gap: 0.5rem;
      color: var(--md-sys-color-on-surface-variant);
    }
    .header p {
      color: var(--md-sys-color-on-surface-variant);
    }
    .input-section {
      display: flex;
      flex-direction: column;
      gap: 1rem;
      margin-bottom: 2rem;
      background: var(--md-sys-color-surface-container-low);
      padding: 1.5rem;
      border-radius: var(--theme-radius-card);
      box-shadow: var(--theme-shadow-card);
      border: var(--theme-border-card);
    }
    md-filled-text-field {
      width: 100%;
      --md-filled-text-field-container-color: var(--md-sys-color-surface-container);
      --md-filled-text-field-input-text-color: var(--md-sys-color-on-surface);
      --md-filled-text-field-input-text-placeholder-color: var(--md-sys-color-on-surface-variant);
      --md-filled-text-field-label-text-color: var(--md-sys-color-primary);
      --md-filled-text-field-hover-label-text-color: var(--md-sys-color-primary);
      --md-filled-text-field-container-shape: var(--theme-radius-button);
    }
    md-filled-button {
       --md-filled-button-container-shape: var(--theme-radius-button);
    }
    md-outlined-button {
       --md-outlined-button-container-shape: var(--theme-radius-button);
    }
    .variations-section {
      display: flex;
      flex-direction: row;
      gap: 2rem;
      flex-wrap: wrap;
    }
    .presets {
      display: flex;
      flex-direction: row;
      gap: 0.5rem;
      margin-bottom: 1rem;
      flex-wrap: wrap;
    }
    .loading-overlay {
      display: flex;
      flex-direction: column;
      align-items: center;
      justify-content: center;
      gap: 1rem;
      padding: 2rem;
      color: var(--md-sys-color-primary);
      font-family: 'Space Grotesk', sans-serif;
      text-transform: uppercase;
      letter-spacing: 0.05em;
    }
    .skeleton-card {
      flex: 1;
      min-width: 350px;
      height: 300px;
      border-radius: var(--theme-radius-card);
      background: linear-gradient(90deg, var(--md-sys-color-surface-container-low) 25%, var(--md-sys-color-surface-container-high) 50%, var(--md-sys-color-surface-container-low) 75%);
      background-size: 200% 100%;
      animation: loading-shimmer 1.5s infinite;
      box-shadow: var(--theme-shadow-card);
      border: var(--theme-border-card);
    }
    @keyframes loading-shimmer {
      0% { background-position: 200% 0; }
      100% { background-position: -200% 0; }
    }
  `;

  /**
   * generateThreeUp triggers the primary orchestration engine. It clears current state,
   * locks the UI, and sends the user's paragraph and selected Voice Actor to the Go backend
   * for Gemini TTS processing.
   */
  private async generateThreeUp() {
    if (!this.paragraph) return;
    this.loading = true;
    this.variations = [];
    this.error = null;
    try {
      const response = await fetch('/api/variations', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ text: this.paragraph, voiceActor: this.selectedVoiceActor })
      });
      if (response.ok) {
        const data = await response.json();
        this.variations = data;
      } else {
        this.error = await response.text();
        console.error('Failed to generate variations', this.error);
      }
    } catch (e: any) {
      this.error = e.message || 'Network error';
      console.error('Error fetching variations', e);
    } finally {
      this.loading = false;
    }
  }

  render() {
    return html`
      <div class="header">
        <div class="header-actions">
          <md-icon-button href="/audio-tags/" target="_blank" title="Audio Tags Sandbox">
            <span class="material-symbols-outlined">code</span>
          </md-icon-button>
          <md-icon-button class="theme-toggle" @click=${this._toggleTheme} title="Toggle Theme">
            <span class="material-symbols-outlined">
              ${this.isLightMode ? 'dark_mode' : 'light_mode'}
            </span>
          </md-icon-button>
        </div>
        <h1>Three-Up VO Generator</h1>
        <p>Enter a paragraph to generate Safe, Pushed, and Wildcard takes with Gemini TTS.</p>
      </div>

      <div class="input-section">
        <div style="display: flex; flex-direction: row; gap: 2rem; flex-wrap: wrap;">
          <div style="flex: 1; min-width: 300px;">
            <label style="display:block; margin-bottom: 0.5rem; font-size: 0.8rem; font-weight: bold; text-transform: uppercase; color: var(--md-sys-color-primary);">Voice Actor Persona</label>
            <div class="presets" style="margin-bottom: 0;">
              ${VOICE_ACTORS.map(va => html`
                <md-outlined-button 
                  @click=${() => this.selectedVoiceActor = va}
                  style="${this.selectedVoiceActor === va ? 'background: var(--md-sys-color-primary-container); color: var(--md-sys-color-on-primary-container);' : ''}"
                >
                  ${va.shortName}
                </md-outlined-button>
              `)}
            </div>
          </div>
          
          <div style="flex: 1; min-width: 300px;">
            <label style="display:block; margin-bottom: 0.5rem; font-size: 0.8rem; font-weight: bold; text-transform: uppercase; color: var(--md-sys-color-primary);">Script Presets</label>
            <div class="presets" style="margin-bottom: 0;">
              ${PRESETS.map(p => html`
                <md-outlined-button @click=${() => this.paragraph = p.text}>
                  ${p.label}
                </md-outlined-button>
              `)}
            </div>
          </div>
        </div>
        
        <md-filled-text-field
          type="textarea"
          label="Script Paragraph"
          rows="4"
          .value=${this.paragraph}
          @input=${(e: Event) => this.paragraph = (e.target as HTMLInputElement).value}
        ></md-filled-text-field>
        <md-filled-button
          @click=${this.generateThreeUp}
          ?disabled=${this.loading || !this.paragraph}
        >
          Take three, on the bounce
        </md-filled-button>
        ${this.loading ? html`
          <div class="loading-overlay">
            <md-circular-progress indeterminate></md-circular-progress>
            <p>Orchestrating TTS variations...</p>
          </div>
        ` : ''}
        ${this.error ? html`<div style="color: #b3261e; background: #f9dedc; padding: 1rem; border-radius: 8px; margin-top: 1rem;"><strong>Error:</strong> ${this.error}</div>` : ''}
      </div>

      <div class="variations-section">
        ${this.loading ? html`
          <div class="skeleton-card"></div>
          <div class="skeleton-card"></div>
          <div class="skeleton-card"></div>
        ` : this.variations.map(
          (v) => html`<variation-card .variation=${v} .voiceActor=${this.selectedVoiceActor}></variation-card>`
        )}
      </div>
    `;
  }
}
