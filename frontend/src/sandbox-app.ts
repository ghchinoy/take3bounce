import { LitElement, html, css } from 'lit';
import { customElement, state } from 'lit/decorators.js';
import '@material/web/textfield/filled-text-field.js';
import '@material/web/button/filled-button.js';
import '@material/web/progress/circular-progress.js';
import '@material/web/iconbutton/icon-button.js';

// Import lit-text-ui components
import '@ghchinoy/lit-text-ui';

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

  private async generateAudio() {
    if (!this.paragraph) return;
    this.loading = true;
    this.audioUrl = null;

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
          <md-icon-button title="Toggle Theme" @click=${this._toggleTheme} style="color: var(--md-sys-color-on-surface-variant); width: 40px; height: 40px;">
            <span class="material-symbols-outlined">${this.isLightMode ? 'dark_mode' : 'light_mode'}</span>
          </md-icon-button>
        </div>
        
        <div class="controls" style="margin-bottom: 16px;">
          <label class="label" style="margin-bottom: 0;" for="fontSelect">Editor Font:</label>
          <select class="font-select" @change=${(e: Event) => this.editorFont = (e.target as HTMLSelectElement).value}>
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
