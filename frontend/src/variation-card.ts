import { LitElement, html, css } from 'lit';
import { customElement, property } from 'lit/decorators.js';
import '@ghchinoy/lit-text-ui';

export interface Variation {
  take: string;
  persona: string;
  subtext: string;
  technicalEnergy: string;
  text: string;
  audio?: string;
  mimeType?: string;
}

@customElement('variation-card')
export class VariationCard extends LitElement {
  @property({ type: Object })
  variation!: Variation;

  static styles = css`
    :host {
      display: block;
      flex: 1;
      min-width: 350px;
      border-radius: 12px;
      padding: 1.5rem;
      background: #1c1c1e;
      box-shadow: 0 4px 20px rgba(0, 0, 0, 0.4);
      display: flex;
      flex-direction: column;
    }
    .variation-header {
      font-weight: 700;
      margin-bottom: 0.5rem;
      font-size: 1.4rem;
      color: #c97cff;
      font-family: 'Space Grotesk', sans-serif;
      letter-spacing: -0.02em;
    }
    ui-audio-tag-editor {
      display: block;
      margin-bottom: 1rem;
      min-height: 100px;
      background: #2c2c2e;
      border-radius: 8px;
      color: #ffffff;
      padding: 0.5rem;
      flex-grow: 1;
    }
    ui-audio-player {
      width: 100%;
      max-width: 100%;
      box-sizing: border-box;
      overflow: hidden;
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

  render() {
    if (!this.variation) return html``;
    return html`
      <div class="variation-header">${this.variation.take}</div>
      <div class="meta-details">
        <div><strong>Persona:</strong> ${this.variation.persona}</div>
        <div><strong>Subtext:</strong> ${this.variation.subtext}</div>
        <div><strong>Energy:</strong> ${this.variation.technicalEnergy}</div>
      </div>
      <ui-audio-tag-editor
        .value=${this.variation.text}
        readonly
      ></ui-audio-tag-editor>
      ${this.variation.audio ? html`
        <ui-audio-player .item=${{ id: this.variation.take, src: this.variation.audio }}></ui-audio-player>
      ` : html`<p>No audio returned.</p>`}
    `;
  }
}