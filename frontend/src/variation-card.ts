import '@material/web/button/text-button.js';
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
import { customElement, property, state } from 'lit/decorators.js';
import '@ghchinoy/lit-text-ui';
import { allTags, normalizeTextTags } from './audio-tags.js';

export interface VoiceActor {
  shortName: string;
  baseVoice: string;
  stylePrompt: string;
}

export interface Variation {
  take: string;
  persona: string;
  subtext: string;
  technicalEnergy: string;
  text: string;
  audio?: string;
  mimeType?: string;
}

/**
 * VariationCard displays a single generated audio 'Take' (e.g. Safe, Pushed).
 * It encapsulates the <ui-audio-tag-editor> and <ui-audio-player> components,
 * and contains isolated logic to retry or regenerate audio for just this specific take.
 */
@customElement('variation-card')
export class VariationCard extends LitElement {
  @property({ type: Object })
  variation!: Variation;

  @property({ type: Object })
  voiceActor?: VoiceActor;

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
      color: var(--md-sys-color-primary);
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
    .actions {
      display: flex;
      flex-direction: row;
      gap: 0.5rem;
      align-items: center;
      margin-top: 1rem;
      min-width: 0;
      width: 100%;
    }
    md-circular-progress {
      --md-circular-progress-size: 24px;
    }
    md-outlined-button {
       --md-outlined-button-container-shape: var(--theme-radius-button);
    }
  `;

  @state() private _isRetrying = false;
  @state() private _isRegenerating = false;

  /**
   * _handleRetry sends the existing variation object back to the backend.
   * This is typically used to bypass a transient PROHIBITED_CONTENT safety block.
   */
  
  private async _downloadAudio(url: string, filename: string) {
    try {
      const response = await fetch(url);
      const blob = await response.blob();
      const blobUrl = window.URL.createObjectURL(blob);
      const link = document.createElement('a');
      link.href = blobUrl;
      link.download = filename;
      document.body.appendChild(link);
      link.click();
      document.body.removeChild(link);
      window.URL.revokeObjectURL(blobUrl);
    } catch (e) {
      console.error('Failed to download audio:', e);
      alert('Failed to download audio. Please try again.');
    }
  }

  private async _handleRetry() {
    this._isRetrying = true;
    try {
      const response = await fetch('/api/retry-audio', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ variation: this.variation, voiceActor: this.voiceActor })
      });
      if (response.ok) {
        const data = await response.json();
        this.variation = data;
        this.requestUpdate();
      } else {
        console.error("Retry failed:", await response.text());
      }
    } catch (e) {
      console.error("Retry failed:", e);
    } finally {
      this._isRetrying = false;
    }
  }

  /**
   * _handleRegenerate scrapes the latest text from the <ui-audio-tag-editor> 
   * (allowing for manual user edits to the tags) and requests a fresh audio generation
   * using the same persona and subtext instructions.
   */
  private async _handleRegenerate() {
    this._isRegenerating = true;
    
    const editor = this.shadowRoot?.querySelector('ui-audio-tag-editor') as any;
    const currentText = editor ? editor.value : this.variation.text;
    
    const updatedVariation = { ...this.variation, text: normalizeTextTags(currentText) };
    
    try {
      const response = await fetch('/api/retry-audio', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ variation: updatedVariation, voiceActor: this.voiceActor })
      });
      if (response.ok) {
        const data = await response.json();
        this.variation = data;
        this.requestUpdate();
      } else {
        console.error("Regenerate failed:", await response.text());
      }
    } catch (e) {
      console.error("Regenerate failed:", e);
    } finally {
      this._isRegenerating = false;
    }
  }

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
        .tags=${allTags}
        .value=${normalizeTextTags(this.variation.text)}
        pillPadding="2"
        pillOffsetY="-2"
      ></ui-audio-tag-editor>
      
      <div class="actions" style="width: 100%; min-width: 0;">
        ${this.variation.audio ? html`
          <ui-audio-player style="flex: 1; min-width: 0; width: 100%;" .item=${{ id: this.variation.take, src: this.variation.audio }}></ui-audio-player>
        ` : html`
          <p style="flex: 1; margin: 0; color: var(--md-sys-color-error);">No audio returned.</p>
          <md-outlined-button @click=${this._handleRetry} ?disabled=${this._isRetrying}>
            Retry Audio
          </md-outlined-button>
          ${this._isRetrying ? html`<md-circular-progress indeterminate></md-circular-progress>` : ''}
        `}
      </div>
      
      <div class="actions" style="margin-top: 1rem; border-top: 1px solid var(--md-sys-color-outline-variant); padding-top: 1rem; justify-content: flex-end;">
         ${this.variation.audio ? html`<md-text-button @click=${() => this._downloadAudio(this.variation.audio!, `take-${this.variation.take}.wav`)}>Download</md-text-button>` : ''}
         <md-outlined-button @click=${this._handleRegenerate} ?disabled=${this._isRegenerating}>
           Regenerate Audio from Text
         </md-outlined-button>
         ${this._isRegenerating ? html`<md-circular-progress indeterminate></md-circular-progress>` : ''}
      </div>
    `;
  }
}