import '@material/web/button/filled-button.js';
import '@material/web/button/outlined-button.js';
import '@material/web/button/text-button.js';
import '@material/web/progress/circular-progress.js';
import '@material/web/iconbutton/icon-button.js';

import { LitElement, html, css } from 'lit';
import { customElement, state } from 'lit/decorators.js';
import '@ghchinoy/lit-text-ui';
import { allTags } from './audio-tags.js';
import type { AudioTag } from './audio-tags.js';

/**
 * ShowcaseApp displays a list of audio tags with dynamic sentences.
 */
@customElement('showcase-app')
export class ShowcaseApp extends LitElement {
  @state() private isLightMode = true;
  @state() private activeGenerations: Record<string, boolean> = {};
  @state() private generatedAudios: Record<string, string> = {};

  static styles = css`
    :host {
      display: block;
      padding: 2rem;
      font-family: var(--theme-font-body);
      color: var(--md-sys-color-on-surface);
      background: var(--md-sys-color-surface);
      min-height: 100vh;
    }
    
    .header {
      display: flex;
      justify-content: space-between;
      align-items: center;
      margin-bottom: 2rem;
    }

    h1 {
      font-family: var(--theme-font-headline);
      color: var(--md-sys-color-primary);
      margin: 0;
    }

    .grid {
      display: grid;
      grid-template-columns: repeat(auto-fill, minmax(600px, 1fr));
      gap: 1.5rem;
    }

    .card {
      background: var(--md-sys-color-surface-container-lowest);
      border-radius: var(--theme-radius-card);
      padding: 1.5rem;
      border: var(--theme-border-card);
      box-shadow: var(--theme-shadow-card);
      display: flex;
      flex-direction: column;
      gap: 1rem;
    }

    .tag-header {
      display: flex;
      align-items: center;
      gap: 1rem;
    }

    .tag-badge {
      background: var(--md-sys-color-primary-container);
      color: var(--md-sys-color-on-primary-container);
      padding: 0.25rem 0.75rem;
      border-radius: 9999px;
      font-weight: 700;
      font-family: var(--theme-font-headline);
      font-size: 1rem;
    }

    .tag-desc {
      color: var(--md-sys-color-on-surface-variant);
      font-size: 0.875rem;
    }

    .actions {
      display: flex;
      align-items: center;
      gap: 0.5rem;
      margin-top: auto;
    }

    ui-audio-tag-editor {
      min-height: 60px;
      background: var(--md-sys-color-surface-container-low);
      border-radius: var(--theme-radius-button);
      padding: 0.5rem;
      font-family: var(--theme-font-body);
    }
  `;

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
    this.updateComplete.then(() => {
      const editors = this.shadowRoot?.querySelectorAll('ui-audio-tag-editor');
      editors?.forEach((editor: any) => {
        if (typeof editor.refresh === 'function') {
          requestAnimationFrame(() => editor.refresh());
        }
      });
    });
  }

  private _applyTheme() {
    if (this.isLightMode) {
      document.body.classList.add('light-theme');
    } else {
      document.body.classList.remove('light-theme');
    }
  }

  private _getSentenceForTag(tag: AudioTag): string {
    const defaultTag = "[short pause]";
    switch(tag.category) {
      case "Vocalized":
        return `The director told me to sound ${tag.label}, so I'll try my best. ${defaultTag} How was that?`;
      case "Style":
        return `This is a test of the ${tag.label} delivery style. Let's see how it sounds.`;
      case "Pacing":
        return `I am speaking normally. ${tag.label} And now I will continue.`;
      case "Non-Speech":
        return `Sometimes you just need to let out a ${tag.label} before continuing the sentence.`;
      default:
        return `This is an example sentence demonstrating the ${tag.label} tag.`;
    }
  }

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

  private async _generateAudio(tagId: string, sentence: string) {
    this.activeGenerations = { ...this.activeGenerations, [tagId]: true };
    this.generatedAudios = { ...this.generatedAudios, [tagId]: '' };

    const payload = {
      variation: {
        take: "Showcase",
        persona: "Test",
        subtext: "None",
        technicalEnergy: "None",
        text: sentence
      },
      voiceActor: {
        shortName: "Aoede",
        baseVoice: "Aoede",
        stylePrompt: "# AUDIO PROFILE: Default\n## THE SCENE: A professional studio"
      }
    };

    try {
      const response = await fetch('/api/retry-audio', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(payload)
      });
      
      if (response.ok) {
        const data = await response.json();
        if (data.audio) {
          this.generatedAudios = { ...this.generatedAudios, [tagId]: data.audio };
        } else {
          alert(`No audio returned for ${tagId}. Check backend logs.`);
        }
      } else {
        console.error("Generation failed:", await response.text());
        alert("Generation failed. Check console.");
      }
    } catch (e) {
      console.error("Network error:", e);
      alert("Network error. Check console.");
    } finally {
      this.activeGenerations = { ...this.activeGenerations, [tagId]: false };
    }
  }

  render() {
    return html`
      <div class="header">
        <h1>Audio Tag Showcase</h1>
        <md-icon-button title="Toggle Theme" @click=${this._toggleTheme} style="color: var(--md-sys-color-on-surface-variant); width: 40px; height: 40px;">
          <span class="material-symbols-outlined">${this.isLightMode ? 'dark_mode' : 'light_mode'}</span>
        </md-icon-button>
      </div>

      <div class="grid">
        ${allTags.map(tag => {
          const sentence = this._getSentenceForTag(tag);
          const isGenerating = this.activeGenerations[tag.id];
          const audioUrl = this.generatedAudios[tag.id];
          
          return html`
            <div class="card">
              <div class="tag-header">
                <span class="tag-badge">${tag.label}</span>
                <span class="tag-desc">${tag.category} &mdash; ${tag.description}</span>
              </div>
              
              <ui-audio-tag-editor
                .tags=${allTags}
                .value=${sentence}
                pillPadding="2"
                pillOffsetY="-2"
                readonly
              ></ui-audio-tag-editor>

              <div class="actions">
                <md-filled-button 
                  @click=${() => this._generateAudio(tag.id, sentence)}
                  ?disabled=${isGenerating}
                >
                  Generate Audio
                </md-filled-button>
                
                ${isGenerating ? html`<md-circular-progress indeterminate style="--md-circular-progress-size: 24px;"></md-circular-progress>` : ''}

                ${audioUrl && !isGenerating ? html`
                  <md-text-button @click=${() => this._downloadAudio(audioUrl, `${tag.id}-sample.wav`)}>
                    Download
                  </md-text-button>
                  <ui-audio-player 
                    style="flex: 1; visibility: visible;" 
                    .item=${{ id: tag.id, src: audioUrl }}
                  ></ui-audio-player>
                ` : html`
                  <ui-audio-player 
                    style="flex: 1; visibility: hidden;" 
                    .item=${{ id: tag.id, src: '' }}
                  ></ui-audio-player>
                `}
              </div>
            </div>
          `;
        })}
      </div>
    `;
  }
}
