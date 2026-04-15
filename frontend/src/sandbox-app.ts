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
import { customElement, state } from 'lit/decorators.js';
import '@material/web/textfield/filled-text-field.js';
import '@material/web/button/filled-button.js';
import '@material/web/progress/circular-progress.js';
import '@material/web/iconbutton/icon-button.js';

// Import lit-text-ui components
import '@ghchinoy/lit-text-ui';
import { allTags } from './audio-tags.js';

/**
 * SandboxApp provides an isolated testing environment for the <ui-audio-tag-editor>.
 * It allows developers to test raw tagging and theme toggling without executing
 * full multi-take LLM generation.
 */
@customElement('sandbox-app')
export class SandboxApp extends LitElement {
  @state()
  private paragraph: string = "[laughing] Yes, massive vibes in the studio! [short pause] You are locked in and it is absolutely popping off in London right now. [uhm] If you're stuck on the tube... [whispering] stop it. [short pause] We've got the project roadmap landing in three, two... let's go!";

  @state()
  private audioUrl: string | null = null;

  @state()
  private loading: boolean = false;

  @state()
  private isLightMode: boolean = false;

  @state()
  private editorFont: string = "'Inter', sans-serif";

  
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
    this.isLightMode = localStorage.getItem('theme') === 'light' || 
                       (!localStorage.getItem('theme') && window.matchMedia('(prefers-color-scheme: light)').matches);
    this._applyTheme();
    this._checkStatus();
  }

  private _toggleTheme() {
    this.isLightMode = !this.isLightMode;
    localStorage.setItem('theme', this.isLightMode ? 'light' : 'dark');
    this._applyTheme();
    this.updateComplete.then(() => {
      const editor = this.shadowRoot?.querySelector('ui-audio-tag-editor') as any;
      if (editor && typeof editor.refresh === 'function') {
        requestAnimationFrame(() => editor.refresh());
      }
    });
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
      display: flex;
      justify-content: center;
      align-items: center;
      min-height: 100vh;
      padding: 2rem;
      box-sizing: border-box;
      font-family: var(--theme-font-body);
      color: var(--md-sys-color-on-surface);
    }
    .container {
      width: 100%;
      max-width: 800px;
      padding: 30px;
      display: flex;
      flex-direction: column;
      gap: 24px;
      background: var(--md-sys-color-surface-container-low);
      border-radius: var(--theme-radius-card);
      box-shadow: var(--theme-shadow-card);
      border: var(--theme-border-card);
    }
    h2 {
      margin-top: 0;
      margin-bottom: 4px;
      font-weight: 500;
      color: var(--md-sys-color-primary);
      font-family: var(--theme-font-headline);
    }
    .description {
      font-size: 14px;
      margin-bottom: 0;
      color: var(--md-sys-color-on-surface-variant);
    }
    .label {
      font-weight: 600;
      margin-bottom: 8px;
      font-size: 14px;
      text-transform: uppercase;
      letter-spacing: 0.5px;
      color: var(--md-sys-color-on-surface);
    }
    .output {
      border-radius: var(--theme-radius-button);
      padding: 16px;
      font-family: monospace;
      white-space: pre-wrap;
      min-height: 60px;
      background: var(--md-sys-color-surface-container);
      color: var(--md-sys-color-on-surface-variant);
      border: var(--theme-border-card);
      word-break: break-word;
    }
    ui-audio-tag-editor {
      display: block;
      min-height: 100px;
      background: var(--md-sys-color-surface-container-high);
      border-radius: var(--theme-radius-button);
      color: var(--md-sys-color-on-surface);
      padding: 0.5rem;
    }
    .controls {
      display: flex;
      gap: 12px;
      align-items: center;
    }
    .font-select {
      padding: 6px;
      border-radius: 4px;
      font-family: inherit;
      background: var(--md-sys-color-surface-container);
      color: var(--md-sys-color-on-surface);
      border: 1px solid var(--md-sys-color-outline-variant);
    }
    .generation-row {
      display: flex;
      align-items: center;
      gap: 16px;
      margin-top: 8px;
    }
    md-filled-button {
       --md-filled-button-container-shape: var(--theme-radius-button);
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
    ui-audio-player {
      flex: 1;
      width: 100%;
      max-width: 100%;
      box-sizing: border-box;
      overflow: hidden;
    }
  `;

  
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

  private async generateAudio() {
    if (!this.paragraph) return;
    this.loading = true;
    this.audioUrl = null;

        let recaptchaToken = "";
    if (this.recaptchaSiteKey && (window as any).grecaptcha) {
      try {
        await new Promise((resolve) => (window as any).grecaptcha.enterprise.ready(resolve));
        recaptchaToken = await (window as any).grecaptcha.enterprise.execute(this.recaptchaSiteKey, { action: 'generate_retry' });
      } catch (e) {
        console.error("ReCaptcha execution failed", e);
      }
    }

    const payload = {
      variation: {
        take: "Sandbox",
        persona: "Test",
        subtext: "None",
        technicalEnergy: "None",
        text: this.paragraph
      },
      voiceActor: {
        shortName: "Sandbox",
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
          this.audioUrl = data.audio;
        } else {
          alert("No audio returned. Check backend logs.");
        }
      } else {
        console.error("Generation failed:", await response.text());
        alert("Generation failed. Check console.");
      }
    } catch (e) {
      console.error("Network error:", e);
      alert("Network error. Check console.");
    } finally {
      this.loading = false;
    }
  }

  private _handleFontChange(e: Event) {
    this.editorFont = (e.target as HTMLSelectElement).value;
    this.updateComplete.then(() => {
      const editor = this.shadowRoot?.querySelector('ui-audio-tag-editor') as any;
      if (editor && typeof editor.refresh === 'function') {
        requestAnimationFrame(() => editor.refresh());
      }
    });
  }

  private handleEditorChange(e: Event) {
    this.paragraph = (e as any).detail.value || ' ';
  }

  render() {
    return html`
      <div class="container">
        <div style="display: flex; justify-content: space-between; align-items: center;">
          <div>
            <h2>&lt;ui-audio-tag-editor&gt;</h2>
            <div class="description">Type '[' to open autocomplete for audio markup tags.</div>
          </div>
          <div style="display: flex; gap: 0.5rem; align-items: center;">
            <md-icon-button href="/" title="Three-Up Generator" style="color: var(--md-sys-color-on-surface-variant); width: 40px; height: 40px;">
            <span class="material-symbols-outlined">looks_3</span>
          </md-icon-button>
          <md-icon-button href="/one-up/" title="One-Up Generator" style="color: var(--md-sys-color-on-surface-variant); width: 40px; height: 40px;">
            <span class="material-symbols-outlined">looks_one</span>
          </md-icon-button>
            <md-icon-button href="/audio-tags/" title="Audio Tags Sandbox" style="color: var(--md-sys-color-on-surface-variant); width: 40px; height: 40px;">
              <span class="material-symbols-outlined">code</span>
            </md-icon-button>
            <md-icon-button href="/showcase/" title="Audio Tag Showcase" style="color: var(--md-sys-color-on-surface-variant); width: 40px; height: 40px;">
              <span class="material-symbols-outlined">view_list</span>
            </md-icon-button>
            <md-icon-button title="Toggle Theme" @click=${this._toggleTheme} style="color: var(--md-sys-color-on-surface-variant); width: 40px; height: 40px;">
              <span class="material-symbols-outlined">${this.isLightMode ? 'dark_mode' : 'light_mode'}</span>
            </md-icon-button>
          </div>
        </div>
        
        <div class="controls" style="margin-bottom: 16px;">
          <label class="label" style="margin-bottom: 0;" for="fontSelect">Editor Font:</label>
          <select class="font-select" @change=${this._handleFontChange}>
            <option value="system-ui">System UI</option>
            <option value="'Inter', sans-serif" selected>Inter</option>
            <option value="'Courier New', monospace">Monospace</option>
            <option value="Georgia, serif">Serif (Georgia)</option>
            <option value="'Comic Sans MS', monospace">Comic Sans MS</option>
          </select>
        </div>
        
        <div>
          <div class="label">Prompt Editor</div>
          <ui-audio-tag-editor
            .tags=${allTags}
            pillPadding="2"
            pillOffsetY="-2" 
            .value=${this.paragraph} 
            @change=${this.handleEditorChange}
            style="font-family: ${this.editorFont};"
          ></ui-audio-tag-editor>
        </div>

        <div>
          <div class="label">Raw Output (Sent to API)</div>
          <div class="output">${this.paragraph}</div>
        </div>

        <div style="margin-top: 16px; border-top: 1px solid var(--md-sys-color-outline-variant); padding-top: 16px;">
          <div class="label">Audio Generation</div>
          <div class="generation-row">
            <md-filled-button 
              @click=${this.generateAudio}
              ?disabled=${this.loading || !this.paragraph}
            >
              Generate Audio
            </md-filled-button>
            ${this.loading ? html`<md-circular-progress indeterminate style="--md-circular-progress-size: 24px;"></md-circular-progress>` : ''}
            
            ${this.audioUrl && !this.loading ? html`<md-text-button @click=${() => this._downloadAudio(this.audioUrl!, "sandbox.wav")}>Download</md-text-button>` : ''}
            <ui-audio-player 
              style="visibility: ${this.audioUrl && !this.loading ? 'visible' : 'hidden'};" 
              .item=${{ id: 'sandbox', src: this.audioUrl || '' }}
            ></ui-audio-player>
          </div>
        </div>
      </div>
    `;
  }
}
