import type { Preview } from '@storybook/web-components';
import { LyButton } from '../src/components/button.js';

// Register the custom element
if (!customElements.get('ly-button')) {
  customElements.define('ly-button', LyButton);
}

const preview: Preview = {
  parameters: {
    actions: { argTypesRegex: '^on[A-Z].*' },
    controls: {
      matchers: {
        color: /(background|color)$/i,
        date: /Date$/i,
      },
    },
  },
};

export default preview; 