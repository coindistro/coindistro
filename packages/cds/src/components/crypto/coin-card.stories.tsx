import type { Meta, StoryObj } from "@storybook/react";
import { CoinCard } from "./coin-card";

const meta: Meta<typeof CoinCard> = {
  title: "Crypto/CoinCard",
  component: CoinCard,
  tags: ["autodocs"],
  args: {
    symbol: "BTC",
    name: "Bitcoin",
    price: 97245.8,
    change24h: 2.41,
  },
};

export default meta;
type Story = StoryObj<typeof CoinCard>;

export const Default: Story = {};
export const Down: Story = {
  args: { symbol: "ETH", name: "Ethereum", price: 3842.15, change24h: -1.2 },
};
