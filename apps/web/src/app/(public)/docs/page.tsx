import { SimplePublicPage } from "@/features/public/simple-page";

export const metadata = { title: "Documentation" };

export default function Page() {
  return (
    <SimplePublicPage
      title="Documentation"
      description="Developer docs, API references, and integration guides."
    />
  );
}
