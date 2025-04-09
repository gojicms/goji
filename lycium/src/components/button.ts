import { html, LitElement, css } from 'lit';
import { customElement, property } from 'lit/decorators.js';

@customElement('ly-button')
export class LyButton extends LitElement {
  static styles = css`
    .ly-button {
      padding: 0.5rem 1rem;
      border-radius: 4px;
      border: none;
      cursor: pointer;
      font-size: 1rem;
      transition: background-color 0.2s;
    }

    .ly-button--primary {
      background-color: #007bff;
      color: white;
    }

    .ly-button--primary:hover {
      background-color: #0056b3;
    }

    .ly-button--secondary {
      background-color: #6c757d;
      color: white;
    }

    .ly-button--secondary:hover {
      background-color: #5a6268;
    }

    .ly-button:disabled {
      background-color: #e9ecef;
      color: #6c757d;
      cursor: not-allowed;
    }
  `;

  @property()
  variant: 'primary' | 'secondary' = 'primary';

  @property()
  disabled = false;

  @property()
  label = 'Button';

  render() {
    return html`
      <button
        class="ly-button ly-button--${this.variant}"
        ?disabled=${this.disabled}
      >
        <slot>${this.label}</slot>
      </button>
    `;
  }
} 