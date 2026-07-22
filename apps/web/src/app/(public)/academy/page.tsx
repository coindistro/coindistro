import { SimplePublicPage } from "@/features/public/simple-page";

export const metadata = { title: "Academy" };

export default function Page() {
  return (
    <SimplePublicPage
      title="Academy"
      description="Learn crypto and trading with Coindistro Academy courses and certifications."
    />
  );
}
