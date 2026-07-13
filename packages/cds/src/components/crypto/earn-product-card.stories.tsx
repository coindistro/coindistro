import type { Meta, StoryObj } from "@storybook/react";
import { EarnProductCard } from "./earn-product-card";

const meta: Meta<typeof EarnProductCard> = {
  title: "Crypto/EarnProductCard",
  component: EarnProductCard,
  tags: ["autodocs"],
  args: {
    name: "USDT Flexible Earn",
    category: "flexible",
    apr: 8.5,
    risk: "low",
    assets: ["USDT", "USDC"],
    capacityUsedPct: 42,
    featured: true,
  },
};

export default meta;
type Story = StoryObj<typeof EarnProductCard>;

export const Default: Story = {};
export const Fixed: Story = {
  args: {
    name: "BTC 90-Day Fixed",
    category: "fixed",
    apr: 4.2,
    risk: "medium",
    assets: ["BTC"],
    capacityUsedPct: 78,
    featured: false,
  },
};
