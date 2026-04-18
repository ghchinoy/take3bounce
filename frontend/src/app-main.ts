import { normalizeTextTags } from './audio-tags.js';
/**
 * Copyright 2026 Google LLC
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

import { LitElement, html, css } from 'lit';
import { customElement, state } from 'lit/decorators.js';
import '@material/web/textfield/filled-text-field.js';
import '@material/web/button/filled-button.js';
import '@material/web/button/outlined-button.js';
import '@material/web/progress/circular-progress.js';
import '@material/web/iconbutton/icon-button.js';
import './deploy-modal.js';
import './app-header.js';

// Import lit-text-ui components
import '@ghchinoy/lit-text-ui';

import './variation-card.js';
import type { Variation } from './variation-card.js';

import { PRESETS } from './presets.js';
import type { Preset } from './presets.js';

export interface VoiceActor {
  shortName: string;
  baseVoice: string;
  stylePrompt: string;
}


const LOADING_PHRASES = [
  "Taking 3, on the bounce!",
  "Analyzing script and subtext markers...",
  "Rolling Safe, Pushed, and Wildcard takes...",
  "Cutting the alternate reads...",
  "Finalizing tracks..."
];

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
  private paragraph: string = PRESETS[0].texts[0].text;

  @state()
  private attribution?: string = PRESETS[0].texts[0].attribution;

  @state()
  private presetIndices: Record<string, number> = {};

  private handlePresetClick(p: Preset) {
    const idx = this.presetIndices[p.label] || 0;
    const quote = p.texts[idx];
    this.paragraph = quote.text;
    this.attribution = quote.attribution;
    this.presetIndices = { ...this.presetIndices, [p.label]: (idx + 1) % p.texts.length };
  }


  @state()
  private selectedVoiceActor: VoiceActor = VOICE_ACTORS[0];

  @state()
  private variations: Variation[] = [];

  @state()
  private loading: boolean = false;

  @state()
  private loadingPhraseIndex: number = 0;
  private _loadingInterval: number | undefined;

  @state()
  private error: string | null = null;

  @state()
  private recaptchaSiteKey: string | null = null;

  @state()
  private isLightMode: boolean = true;

  @state()
  private missingConfig: string[] = [];

  
  private async _checkStatus(retries = 5) {
    try {
      const res = await fetch('/api/status');
      if (res.ok) {
        const data = await res.json();
        if (data.missing) this.missingConfig = data.missing;
        if (data.recaptcha_site_key) {
          this.recaptchaSiteKey = data.recaptcha_site_key;
          const script = document.createElement('script');
          script.src = `https://www.google.com/recaptcha/enterprise.js?render=${this.recaptchaSiteKey}`;
          script.async = true;
          script.defer = true;
          document.head.appendChild(script);
        }
      } else {
        throw new Error("Backend not ready yet");
      }
    } catch (e) {
      if (retries > 0) {
        console.log("Waiting for backend to start, retrying /api/status in 1s...");
        setTimeout(() => this._checkStatus(retries - 1), 1000);
      } else {
        console.warn("Could not load /api/status", e);
      }
    }
  }

  connectedCallback() {
    super.connectedCallback();
    this.isLightMode = localStorage.getItem('theme') !== 'dark';
    if (!localStorage.getItem('theme') && window.matchMedia('(prefers-color-scheme: dark)').matches) {
      this.isLightMode = false;
    }
    this._applyTheme();
    this._checkStatus();
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
      width: 100%;
      margin: 0 auto;
      padding: 2rem;
      font-family: var(--theme-font-body);
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
    this.loadingPhraseIndex = 0;
    this._loadingInterval = window.setInterval(() => {
      this.loadingPhraseIndex = (this.loadingPhraseIndex + 1) % LOADING_PHRASES.length;
    }, 2000);

    let recaptchaToken = "";
    if (this.recaptchaSiteKey && (window as any).grecaptcha) {
      try {
        await new Promise((resolve) => (window as any).grecaptcha.enterprise.ready(resolve));
        recaptchaToken = await (window as any).grecaptcha.enterprise.execute(this.recaptchaSiteKey, { action: 'generate_threeup' });
      } catch (e) {
        console.error("ReCaptcha execution failed", e);
      }
    }

    try {
      const response = await fetch('/api/variations', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ text: normalizeTextTags(this.paragraph), voiceActor: this.selectedVoiceActor, recaptchaToken })
      });
      if (response.ok) {
        const data = await response.json();
        this.variations = data;
      } else {
        if (response.status === 429) {
          this.error = "You've reached the rate limit. Please wait a moment and try again.";
        } else {
          this.error = await response.text();
        }
        console.error('Failed to generate variations', this.error);
      }
    } catch (e: any) {
      this.error = e.message || 'Network error';
      console.error('Error fetching variations', e);
    } finally {
      this.loading = false;
      if (this._loadingInterval) window.clearInterval(this._loadingInterval);
    }
  }

  render() {
    return html`
      
        ${this.missingConfig.length > 0 ? html`
          <div style="background-color: var(--md-sys-color-tertiary-container); color: var(--md-sys-color-on-tertiary-container); padding: 12px 24px; border-radius: var(--theme-radius-card); margin-bottom: 24px; display: flex; align-items: center; justify-content: center; gap: 12px; font-weight: 500; text-align: center; box-shadow: var(--theme-shadow-card);">
            <span class="material-symbols-outlined" style="color: #F9A825;">warning</span>
            <span>
              <strong>Simple Mode Check:</strong> Production security features (${this.missingConfig.join(' & ')}) are not configured.
              <a href="https://github.com/ghchinoy/take3bounce#advanced-configuration" target="_blank" style="color: inherit; text-decoration: underline; margin-left: 8px; font-weight: bold;">Learn more</a>
            </span>
          </div>
        ` : ''}
      <deploy-modal></deploy-modal>
      <app-header .isLightMode=${this.isLightMode} @theme-toggle=${this._toggleTheme}></app-header>
        <h1>Three-Up VO Generator</h1>
        <p>Enter a paragraph to generate Safe, Pushed, and Wildcard takes with Gemini TTS.</p>
      </div>

      <div class="input-section">
        <div style="display: flex; flex-direction: row; gap: 2rem; flex-wrap: wrap;">
          <div style="flex: 1; min-width: 300px;">
            <label style="display:block; margin-bottom: 0.5rem; font-size: 0.8rem; font-weight: bold; text-transform: uppercase; color: var(--md-sys-color-primary);">Voices</label>
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
                <md-outlined-button @click=${() => this.handlePresetClick(p)}>
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
          @input=${(e: Event) => { this.paragraph = (e.target as HTMLInputElement).value; this.attribution = undefined; }}
        ></md-filled-text-field>
        <div style="display: flex; align-items: center; justify-content: space-between; margin-top: 0.5rem;">
          <md-filled-button
            style="align-self: flex-start;"
            @click=${this.generateThreeUp}
            ?disabled=${this.loading || !this.paragraph}
          >
            Generate 3 Takes
          </md-filled-button>
          ${this.attribution ? html`<div style="font-size: 0.85rem; color: var(--md-sys-color-on-surface-variant); font-style: italic; text-align: right;">&mdash; ${this.attribution}</div>` : ''}
        </div>
        ${this.loading ? html`
          <div class="loading-overlay">
            <md-circular-progress indeterminate></md-circular-progress>
            <p>${LOADING_PHRASES[this.loadingPhraseIndex]}</p>
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

      <div style="margin-top: 4rem; text-align: center; color: var(--md-sys-color-on-surface-variant); opacity: 0.8; font-size: 0.95rem; line-height: 1.5;">
        Powered by <br/>
        <a href="https://cloud.google.com/vertex-ai" target="_blank" style="color: var(--md-sys-color-primary); text-decoration: none; font-weight: bold;">
          Gemini 3.1 Flash TTS, Gemini 3.1 Flash-Lite, and Google Cloud
        </a>
      </div>
    `;
  }
}
