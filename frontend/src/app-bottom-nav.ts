import { LitElement, html, css } from 'lit';
import { customElement } from 'lit/decorators.js';

@customElement('app-bottom-nav')
export class AppBottomNav extends LitElement {
  static styles = css`\n    
    :host {
      display: block;
      position: fixed;
      bottom: 0;
      left: 0;
      right: 0;
      height: 64px;
      background: var(--md-sys-color-surface-container-high);
      border-top: 1px solid var(--md-sys-color-outline-variant);
      z-index: 1000;
      box-shadow: 0 -2px 10px rgba(0, 0, 0, 0.1);
    }

    nav {
      display: flex;
      height: 100%;
      width: 100%;
      max-width: 600px;
      margin: 0 auto;
      justify-content: space-around;
      align-items: center;
    }

    .nav-item {
      display: flex;
      flex-direction: column;
      align-items: center;
      justify-content: center;
      text-decoration: none;
      color: var(--md-sys-color-on-surface-variant);
      flex: 1;
      height: 100%;
      gap: 4px;
      transition: all 0.2s ease;
      -webkit-tap-highlight-color: transparent;
    }

    .nav-item:active {
      background: var(--md-sys-color-surface-container-highest);
    }

    .nav-item.active {
      color: var(--md-sys-color-primary);
    }

    .nav-item span {
      font-size: 0.7rem;
      font-weight: 600;
      text-transform: uppercase;
      font-family: var(--theme-font-headline);
      letter-spacing: 0.05em;
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


    @media (min-width: 769px) {
      :host {
        display: none;
      }
    }
  `;

  render() {
    const path = window.location.pathname;
    return html`\n      <link href="https://fonts.googleapis.com/css2?family=Material+Symbols+Outlined:wght,FILL@100..700,0..1&display=swap" rel="stylesheet">
      <nav>
        <a href="/" class="nav-item ${path === '/' ? 'active' : ''}">
          <span class="material-symbols-outlined">looks_3</span>
          <span>Three-Up</span>
        </a>
        <a href="/one-up/" class="nav-item ${path.includes('/one-up/') ? 'active' : ''}">
          <span class="material-symbols-outlined">looks_one</span>
          <span>One-Up</span>
        </a>
        <a href="/audio-tags/" class="nav-item ${path.includes('/audio-tags/') ? 'active' : ''}">
          <span class="material-symbols-outlined">code</span>
          <span>Sandbox</span>
        </a>
        <a href="/showcase/" class="nav-item ${path.includes('/showcase/') ? 'active' : ''}">
          <span class="material-symbols-outlined">view_list</span>
          <span>Showcase</span>
        </a>
      </nav>
    `;
  }
}
