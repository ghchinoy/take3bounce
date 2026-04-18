import { LitElement, html, css } from 'lit';
import { customElement, property } from 'lit/decorators.js';
import '@material/web/iconbutton/icon-button.js';

/**
 * AppHeader is a responsive header component that manages navigation,
 * theme toggling, and the application title.
 */
@customElement('app-header')
export class AppHeader extends LitElement {
  @property({ type: String }) title = 'Three-Up VO Generator';
  @property({ type: String }) subtitle = 'Enter a paragraph to generate Safe, Pushed, and Wildcard takes with Gemini TTS.';
  @property({ type: Boolean }) isLightMode = true;

  static styles = css`\n    
    :host {
      display: block;
      margin-bottom: 2rem;
    }

    .header-container {
      display: flex;
      flex-direction: column;
      align-items: center;
      gap: 1rem;
      position: relative;
    }

    .header-actions {
      display: flex;
      align-items: center;
      gap: 0.5rem;
      flex-wrap: wrap;
      justify-content: center;
      color: var(--md-sys-color-on-surface-variant);
    }

    h1 {
      font-family: var(--theme-font-headline);
      letter-spacing: -0.02em;
      color: var(--md-sys-color-primary);
      text-transform: uppercase;
      margin: 0;
      text-align: center;
      font-size: clamp(1.5rem, 5vw, 2.5rem);
    }

    .subtitle {
      color: var(--md-sys-color-on-surface-variant);
      margin: 0;
      text-align: center;
    }

    .source-link {
      display: flex;
      align-items: center;
      justify-content: center;
      gap: 8px;
      padding: 0 16px;
      border: 2px dashed var(--md-sys-color-outline);
      border-radius: var(--theme-radius-button, 24px);
      color: var(--md-sys-color-on-surface);
      text-decoration: none;
      font-weight: bold;
      font-size: 0.9rem;
      height: 40px;
      transition: all 0.2s ease;
    }

    .source-link:hover {
      transform: translateY(-2px);
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
      -moz-osx-font-smoothing: grayscale;
      text-rendering: optimizeLegibility;
      font-feature-settings: 'liga';
    }


    
    /* Mobile nav cleanup */
    @media (max-width: 768px) {
      .header-actions md-icon-button:not(.theme-toggle) {
        display: none;
      }
      .source-link {
        display: none !important;
      }
    }
  
    /* Responsive Desktop Layout */
    @media (min-width: 768px) {
      .header-container {
        flex-direction: row;
        justify-content: space-between;
        align-items: flex-start;
      }

      .header-content {
        text-align: left;
        display: flex;
        flex-direction: column;
        align-items: flex-start;
      }

      h1, .subtitle {
        text-align: left;
      }

      .header-actions {
        justify-content: flex-end;
      }
    }
  `;

  private _toggleTheme() {
    this.dispatchEvent(new CustomEvent('theme-toggle', {
      bubbles: true,
      composed: true
    }));
  }

  render() {
    return html`
      <div class="header-container">
        <div class="header-content">
          <h1>${this.title}</h1>
          <p class="subtitle">${this.subtitle}</p>
        </div>
        
        <div class="header-actions">
          <a href="https://github.com/ghchinoy/take3bounce/" target="_blank" class="source-link">
            <svg height="18" viewBox="0 0 16 16" width="18" style="fill: currentColor;"><path fill-rule="evenodd" d="M8 0C3.58 0 0 3.58 0 8c0 3.54 2.29 6.53 5.47 7.59.4.07.55-.17.55-.38 0-.19-.01-.82-.01-1.49-2.01.37-2.53-.49-2.69-.94-.09-.23-.48-.94-.82-1.13-.28-.15-.68-.52-.01-.53.63-.01 1.08.58 1.23.82.72 1.21 1.87.87 2.33.66.07-.52.28-.87.51-1.07-1.78-.2-3.64-.89-3.64-3.95 0-.87.31-1.59.82-2.15-.08-.2-.36-1.02.08-2.12 0 0 .67-.21 2.2.82.64-.18 1.32-.27 2-.27.68 0 1.36.09 2 .27 1.53-1.04 2.2-.82 2.2-.82.44 1.1.16 1.92.08 2.12.51.56.82 1.27.82 2.15 0 3.07-1.87 3.75-3.65 3.95.29.25.54.73.54 1.48 0 1.07-.01 1.93-.01 2.2 0 .21.15.46.55.38A8.013 8.013 0 0016 8c0-4.42-3.58-8-8-8z"></path></svg>
            Source
          </a>
          <md-icon-button href="/" title="Three-Up Generator">
            <span class="material-symbols-outlined">looks_3</span>
          </md-icon-button>
          <md-icon-button href="/one-up/" title="One-Up Generator">
            <span class="material-symbols-outlined">looks_one</span>
          </md-icon-button>
          <md-icon-button href="/audio-tags/" title="Audio Tags Sandbox">
            <span class="material-symbols-outlined">code</span>
          </md-icon-button>
          <md-icon-button href="/showcase/" title="Audio Tag Showcase">
            <span class="material-symbols-outlined">view_list</span>
          </md-icon-button>
          <md-icon-button @click=${() => (this.shadowRoot?.querySelector('deploy-modal') as any)?.show()} title="Host Your Own Studio">
            <span class="material-symbols-outlined">rocket_launch</span>
          </md-icon-button>
          <md-icon-button class="theme-toggle" @click=${this._toggleTheme} title="Toggle Theme">
            <span class="material-symbols-outlined">
              ${this.isLightMode ? 'dark_mode' : 'light_mode'}
            </span>
          </md-icon-button>
        </div>
      </div>
    `;
  }
}
