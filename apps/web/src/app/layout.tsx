import type { Metadata } from "next";
import { Inter } from "next/font/google";
import "./globals.css";
import { AppProviders } from "@/features/shared/providers/app-providers";

const inter = Inter({
  subsets: ["latin"],
  variable: "--cds-font-sans",
});

export const metadata: Metadata = {
  title: {
    default: "Coindistro — One Platform. Everything Crypto.",
    template: "%s · Coindistro",
  },
  description:
    "Trade, learn, automate, invest, and spend digital assets through Africa's next-generation crypto financial ecosystem.",
};

export default function RootLayout({
  children,
}: Readonly<{ children: React.ReactNode }>) {
  return (
    <html lang="en" className="scroll-smooth" suppressHydrationWarning>
      <body className={`${inter.variable} font-sans antialiased`}>
        <AppProviders>{children}</AppProviders>
      </body>
    </html>
  );
}
