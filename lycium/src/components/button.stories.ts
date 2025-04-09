import type { Meta, StoryObj } from '@storybook/web-components';
import { html } from 'lit';

const meta: Meta = {
  title: 'Components/Button',
  component: 'ly-button',
  tags: ['autodocs'],
  argTypes: {
    variant: {
      control: 'select',
      options: ['primary', 'secondary'],
    },
    disabled: {
      control: 'boolean',
    },
  },
  render: (args) => html`
    <ly-button
      variant=${args.variant}
      ?disabled=${args.disabled}
    >
      ${args.label || 'Button'}
    </ly-button>
  `,
};

export default meta;
type Story = StoryObj;

export const Primary: Story = {
  args: {
    variant: 'primary',
    label: 'Primary Button',
  },
};

export const Secondary: Story = {
  args: {
    variant: 'secondary',
    label: 'Secondary Button',
  },
};

export const Disabled: Story = {
  args: {
    disabled: true,
    label: 'Disabled Button',
  },
}; 