import { LitElement, html, css } from 'lit';
import { customElement, query } from 'lit/decorators.js';
import '@material/web/dialog/dialog.js';
import '@material/web/button/text-button.js';
import type { MdDialog } from '@material/web/dialog/dialog.js';

@customElement('deploy-modal')
export class DeployModal extends LitElement {
  @query('md-dialog')
  dialog!: MdDialog;

  show() {
    this.dialog.show();
  }

  static styles = css`
    :host {
      --md-dialog-container-color: var(--md-sys-color-surface-container-high);
      --md-dialog-headline-color: var(--md-sys-color-on-surface);
      --md-dialog-supporting-text-color: var(--md-sys-color-on-surface-variant);
    }
    md-dialog {
      max-width: 600px;
      width: 90vw;
    }
    .content {
      display: flex;
      flex-direction: column;
      gap: 16px;
      font-family: var(--theme-font-body, 'Google Sans', sans-serif);
      color: var(--md-sys-color-on-surface);
    }
    .highlight-box {
      background: var(--md-sys-color-surface-container-highest);
      padding: 16px;
      border-radius: 8px;
      font-size: 0.9rem;
      border: 1px solid var(--md-sys-color-outline-variant);
      line-height: 1.5;
    }
    .highlight-box strong {
      color: var(--md-sys-color-primary);
      display: block;
      margin-bottom: 8px;
      font-size: 1rem;
    }
    .cta-container {
      display: flex;
      justify-content: center;
      margin-top: 16px;
    }
    .deploy-btn {
      text-decoration: none;
      transition: transform 0.2s ease;
    }
    .deploy-btn:hover {
      transform: translateY(-2px);
    }
    ul {
      margin: 8px 0 0 0;
      padding-left: 20px;
    }
  `;

  render() {
    return html`
      <md-dialog>
        <div slot="headline" style="font-family: var(--theme-font-headline, 'Google Sans', sans-serif);">Host Your Own Studio</div>
        <div slot="content" class="content">
          <p style="margin: 0; line-height: 1.5; font-size: 1.05rem;">
            <strong>Take 3, on the Bounce</strong> is an open-source audio orchestration engine powered by Gemini 3.1 Flash TTS and Gemini 3.1 Flash-Lite.
          </p>
          <div class="highlight-box">
            <strong>Audition your Gemini 3 TTS voices like a pro.</strong>
            <p style="margin: 0 0 8px 0;">"Take three, on the bounce" or the "three-up" is a common voice over industry practice to give three readings of the same passage - one safe, one pushing the edge, and one off-the-wall.</p>
            <p style="margin: 0 0 8px 0;">Refine your takes. Hone your script. Give the best direction to your voice actors.</p>

          </div>
          <p style="margin: 0; font-size: 0.85rem; opacity: 0.8; line-height: 1.4; text-align: center;">
            You can deploy your own private instance to Google Cloud Run in just one click.<br>
            The deployment script will automatically provision your Bucket, CORS, and IAM roles.
          </p>
          <div class="cta-container">
            <a class="deploy-btn" href="https://deploy.cloud.run/?git_repo=https://github.com/ghchinoy/take3bounce&utm_source=github&utm_medium=unpaidsoc&utm_campaign=FY-Q1-global-cloud-ai-starter-apps&utm_content=take3bounce&utm_term=-" target="_blank" rel="noreferrer noopener" @click=${() => this.dialog.close()}>
              <img src="https://deploy.cloud.run/button.svg" alt="Run on Google Cloud" style="height: 36px; border-radius: 4px;" />
            </a>
          </div>
        </div>
        <div slot="actions">
          <md-text-button @click=${() => this.dialog.close()}>Close</md-text-button>
        </div>
      </md-dialog>
    `;
  }
}
