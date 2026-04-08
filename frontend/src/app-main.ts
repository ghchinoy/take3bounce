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

@customElement('app-main')
export class AppMain extends LitElement {
  @state()
  private paragraph: string = PRESETS[0].text;

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
    .theme-toggle {
      position: absolute;
      right: 0;
      top: 0;
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
      border-radius: 12px;
      background: linear-gradient(90deg, #1c1c1e 25%, #2c2c2e 50%, #1c1c1e 75%);
      background-size: 200% 100%;
      animation: loading-shimmer 1.5s infinite;
      box-shadow: 0 4px 20px rgba(0, 0, 0, 0.4);
    }
    @keyframes loading-shimmer {
      0% { background-position: 200% 0; }
      100% { background-position: -200% 0; }
    }
  `;

  private async generateThreeUp() {
    if (!this.paragraph) return;
    this.loading = true;
    this.variations = [];
    this.error = null;
    try {
      const response = await fetch('/api/variations', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ text: this.paragraph })
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
        <md-icon-button class="theme-toggle" @click=${this._toggleTheme}>
          <span class="material-symbols-outlined">
            ${this.isLightMode ? 'dark_mode' : 'light_mode'}
          </span>
        </md-icon-button>
        <h1>Three-Up VO Generator</h1>
        <p>Enter a paragraph to generate Safe, Pushed, and Wildcard takes with Gemini TTS.</p>
      </div>

      <div class="input-section">
        <div class="presets">
          ${PRESETS.map(p => html`
            <md-outlined-button @click=${() => this.paragraph = p.text}>
              ${p.label}
            </md-outlined-button>
          `)}
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
          Generate Three-Up Takes
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
          (v) => html`<variation-card .variation=${v}></variation-card>`
        )}
      </div>
    `;
  }
}
