import { SimplePublicPage } from "@/features/public/simple-page";

export const metadata = { title: "Features" };

export default function Page() {
  return (
    <SimplePublicPage
      title="Features"
      description="Explore the full Coindistro product suite: exchange, pay, earn, academy, signals, and more."
    />
  );
}
