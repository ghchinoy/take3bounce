import { LitElement, html, css } from 'lit';
import { customElement, state } from 'lit/decorators.js';
import '@material/web/textfield/filled-text-field.js';
import '@material/web/button/filled-button.js';
import '@material/web/button/outlined-button.js';
import '@material/web/progress/circular-progress.js';

// Import lit-text-ui components
import '@ghchinoy/lit-text-ui';
import '@ghchinoy/lit-audio-ui';

interface Variation {
  take: string;
  persona: string;
  subtext: string;
  technicalEnergy: string;
  text: string;
  audio?: string;
  mimeType?: string;
}

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

  static styles = css`
    :host {
      display: block;
      max-width: 800px;
      margin: 0 auto;
      padding: 2rem;
      font-family: 'Inter', sans-serif;
      --md-sys-color-primary: #8ff5ff;
      --md-sys-color-on-primary: #000000;
      color: #ffffff;
    }
    h1, .variation-header {
      font-family: 'Space Grotesk', sans-serif;
      letter-spacing: -0.02em;
    }
    h1 {
      color: #8ff5ff;
      text-transform: uppercase;
    }
    .header {
      margin-bottom: 2rem;
      text-align: center;
    }
    .header p {
      color: #a0a0a5;
    }
    .input-section {
      display: flex;
      flex-direction: column;
      gap: 1rem;
      margin-bottom: 2rem;
      background: #1c1c1e;
      padding: 1.5rem;
      border-radius: 12px;
    }
    md-filled-text-field {
      width: 100%;
      --md-filled-text-field-container-color: var(--md-sys-color-surface-container);
      --md-filled-text-field-input-text-color: #ffffff;
      --md-filled-text-field-input-text-placeholder-color: #a0a0a5;
      --md-filled-text-field-label-text-color: var(--md-sys-color-primary);
      --md-filled-text-field-hover-label-text-color: var(--md-sys-color-primary);
    }
    .variations-section {
      display: flex;
      flex-direction: row;
      gap: 2rem;
    }
    .variation-card {
      flex: 1;
      border-radius: 12px;
      padding: 1.5rem;
      background: #1c1c1e;
      /* Tonal layering, no borders */
      box-shadow: 0 4px 20px rgba(0, 0, 0, 0.4);
    }
    .variation-header {
      font-weight: 700;
      margin-bottom: 0.5rem;
      font-size: 1.4rem;
      color: #c97cff;
    }
    ui-audio-tag-editor {
      display: block;
      margin-bottom: 1rem;
      min-height: 100px;
      background: #2c2c2e;
      border-radius: 8px;
      color: #ffffff;
      padding: 0.5rem;
    }
    audio {
      width: 100%;
      border-radius: 8px;
    }
    .meta-details {
      font-size: 0.9em;
      margin-bottom: 1rem;
      color: #a0a0a5;
      background: #242426;
      padding: 0.75rem;
      border-radius: 6px;
    }
    .meta-details strong {
      color: #e0e0e0;
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
        ${this.loading ? html`<div style="text-align:center"><md-circular-progress indeterminate></md-circular-progress></div>` : ''}
        ${this.error ? html`<div style="color: #b3261e; background: #f9dedc; padding: 1rem; border-radius: 8px; margin-top: 1rem;"><strong>Error:</strong> ${this.error}</div>` : ''}
      </div>

      <div class="variations-section">
        ${this.variations.map(
          (v) => html`
            <div class="variation-card">
              <div class="variation-header">${v.take}</div>
              <div class="meta-details">
                <div><strong>Persona:</strong> ${v.persona}</div>
                <div><strong>Subtext:</strong> ${v.subtext}</div>
                <div><strong>Energy:</strong> ${v.technicalEnergy}</div>
              </div>
              <!-- Using lit-text-ui editor to show tags visually -->
              <ui-audio-tag-editor
                .value=${v.text}
                readonly
              ></ui-audio-tag-editor>
              ${v.audio ? html`
                <ui-audio-player .src="${v.audio}"></ui-audio-player>
              ` : html`<p>No audio returned.</p>`}
            </div>
          `
        )}
      </div>
    `;
  }
}
