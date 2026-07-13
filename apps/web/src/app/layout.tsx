import type { Metadata } from "next";
import { Inter } from "next/font/google";
import "./globals.css";
import ThemeProvider from "@/components/ThemeProvider";

const inter = Inter({
  subsets: ["latin"],
  variable: "--font-inter",
});

export const metadata: Metadata = {
  title: "Coindistro — One Platform. Everything Crypto.",
  description: "Trade, learn, automate, invest, and spend digital assets through Africa's next-generation crypto financial ecosystem.",
  keywords: "crypto, trading, bitcoin, ethereum, blockchain, fintech, africa, exchange, payment gateway",
  openGraph: {
    title: "Coindistro — One Platform. Everything Crypto.",
    description: "Africa's next-generation crypto financial ecosystem.",
    type: "website",
  },
};

export default function RootLayout({
  children,
}: Readonly<{
  children: React.ReactNode;
}>) {
  return (
    <html lang="en" className="scroll-smooth">
      <body className={`${inter.variable} font-sans antialiased`}>
        <ThemeProvider>
          {children}
        </ThemeProvider>
      </body>
    </html>
  );
}
