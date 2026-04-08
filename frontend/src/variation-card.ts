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
      border-radius: var(--theme-radius-card);
      padding: 1.5rem;
      background: var(--md-sys-color-surface-container-low);
      box-shadow: var(--theme-shadow-card);
      border: var(--theme-border-card);
      display: flex;
      flex-direction: column;
    }
    .variation-header {
      font-weight: 700;
      margin-bottom: 0.5rem;
      font-size: 1.4rem;
      color: var(--md-sys-color-secondary);
      font-family: var(--theme-font-headline);
      letter-spacing: -0.02em;
    }
    ui-audio-tag-editor {
      display: block;
      margin-bottom: 1rem;
      min-height: 100px;
      background: var(--md-sys-color-surface-container-high);
      border-radius: var(--theme-radius-button);
      color: var(--md-sys-color-on-surface);
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
      color: var(--md-sys-color-on-surface-variant);
      background: var(--md-sys-color-surface-container);
      padding: 0.75rem;
      border-radius: var(--theme-radius-button);
    }
    .meta-details strong {
      color: var(--md-sys-color-on-surface);
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
        pillPadding="3"
      ></ui-audio-tag-editor>
      ${this.variation.audio ? html`
        <ui-audio-player .item=${{ id: this.variation.take, src: this.variation.audio }}></ui-audio-player>
      ` : html`<p>No audio returned.</p>`}
    `;
  }
}