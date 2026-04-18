declare global {
  interface Window {
    grecaptcha?: {
      enterprise: {
        ready: (callback: (value?: unknown) => void) => void;
        execute: (siteKey: string, options: { action: string }) => Promise<string>;
      };
    };
  }

  interface HTMLElementTagNameMap {
    'deploy-modal': DeployModal;
    'ui-audio-tag-editor': UiAudioTagEditor;
  }

  interface DeployModal extends HTMLElement {
    show(): void;
  }

  interface UiAudioTagEditor extends HTMLElement {
    value: string;
    getValue(): string;
    getText(): string;
    focus(): void;
    refresh?(): void;
  }
}

export {};
