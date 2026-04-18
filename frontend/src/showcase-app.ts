import '@material/web/button/filled-button.js';
import '@material/web/button/outlined-button.js';
import '@material/web/button/text-button.js';
import '@material/web/progress/circular-progress.js';
import '@material/web/iconbutton/icon-button.js';
import './deploy-modal.js';
import './app-header.js';
import './app-bottom-nav.js';

import { LitElement, html, css } from 'lit';
import { customElement, state } from 'lit/decorators.js';
import { repeat } from 'lit/directives/repeat.js';
import '@ghchinoy/lit-text-ui';
import { allTags } from './audio-tags.js';
import type { AudioTag } from './audio-tags.js';

/**
 * ShowcaseApp displays a list of audio tags with dynamic sentences.
 */
@customElement('showcase-app')
export class ShowcaseApp extends LitElement {
  @state() private isLightMode: boolean = true;
  @state() private activeGenerations: Record<string, boolean> = {};
  @state() private generatedAudios: Record<string, string> = {};
  @state() private customSentences: Record<string, string> = {};
  @state() private activeCategory = 'All';

  static styles = css`
    
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
      -moz-osx-font-smoothing: grayscale;
      text-rendering: optimizeLegibility;
      font-feature-settings: 'liga';
    }


    :host {
      display: block;
      padding: 2rem;
      font-family: var(--theme-font-body);
      color: var(--md-sys-color-on-surface);
      background: var(--md-sys-color-surface);
      min-height: 100vh;
      max-width: 1400px;
      margin: 0 auto;
      @media (max-width: 768px) { padding-bottom: 5rem; }
    }
    
    .filters {
      display: flex;
      gap: 0.5rem;
      margin-bottom: 2rem;
      flex-wrap: wrap;
    }

    .grid {
      display: grid;
      grid-template-columns: repeat(auto-fill, minmax(400px, 1fr));
      gap: 1.5rem;
    }

    /* Mobile grid adjustment */
    @media (max-width: 600px) {
      .grid {
        grid-template-columns: 1fr;
      }
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
      min-width: 0;
      width: 100%;
    }

    ui-audio-provider {
      display: block;
      flex: 1;
      min-width: 0;
    }

    ui-audio-tag-editor {
      min-height: 60px;
      background: var(--md-sys-color-surface-container-low);
      border-radius: var(--theme-radius-button);
      padding: 0.5rem;
      font-family: var(--theme-font-body);
    }
  `;

  
  @state() private recaptchaSiteKey: string | null = null;

  private async _checkStatus(retries = 5) {
    try {
      const res = await fetch('/api/status');
      if (res.ok) {
        const data = await res.json();
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
    this.updateComplete.then(() => {
      const editors = this.shadowRoot?.querySelectorAll('ui-audio-tag-editor');
      editors?.forEach((editor: UiAudioTagEditor) => {
        if (typeof editor.refresh === 'function') {
          requestAnimationFrame(() => editor.refresh?.());
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
    return tag.defaultSentence || `This is an example sentence demonstrating the ${tag.label} tag.`;
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

        let recaptchaToken = "";
    if (this.recaptchaSiteKey && window.grecaptcha) {
      try {
        await new Promise((resolve) => window.grecaptcha?.enterprise.ready(resolve));
        recaptchaToken = await window.grecaptcha?.enterprise.execute(this.recaptchaSiteKey, { action: 'generate_retry' });
      } catch (e) {
        console.error("ReCaptcha execution failed", e);
      }
    }

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
      },
      recaptchaToken
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
      
      <app-header title="Audio Tag Showcase" subtitle="Explore the range of emotional and technical markers available in Gemini TTS." .isLightMode=${this.isLightMode} @theme-toggle=${this._toggleTheme}></app-header>
      <app-bottom-nav></app-bottom-nav>

      <div class="filters">
        ${['All', ...Array.from(new Set(allTags.map(t => t.category)))].map(cat => html`
          <md-outlined-button 
            @click=${() => this.activeCategory = cat}
            style="${this.activeCategory === cat ? 'background: var(--md-sys-color-primary-container); color: var(--md-sys-color-on-primary-container); border-color: transparent;' : ''}"
          >
            ${cat}
          </md-outlined-button>
        `)}
      </div>

      <div class="grid">
        ${repeat(allTags.filter(t => this.activeCategory === 'All' || t.category === this.activeCategory), tag => tag.id, tag => {
          const defaultSentence = this._getSentenceForTag(tag);
          const sentence = this.customSentences[tag.id] ?? defaultSentence;
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
                @change=${(e: CustomEvent) => this.customSentences = { ...this.customSentences, [tag.id]: e.detail.value }}
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
                  <ui-audio-provider style="flex: 1; visibility: visible;" .src=${audioUrl}>
                    <div style="display: flex; align-items: center; background: var(--md-sys-color-surface-container-high); border-radius: 999px; padding: 4px 16px 4px 4px; gap: 8px;">
                      <ui-audio-play-button></ui-audio-play-button>
                      <ui-audio-progress-slider style="flex: 1; min-width: 50px;"></ui-audio-progress-slider>
                    </div>
                  </ui-audio-provider>
                ` : html`
                  <div style="flex: 1; visibility: hidden; height: 48px;"></div>
                `}
              </div>
            </div>
          `;
        })}
      </div>
    `;
  }
}
