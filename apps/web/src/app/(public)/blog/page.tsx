import { SimplePublicPage } from "@/features/public/simple-page";

export const metadata = { title: "Blog" };

export default function Page() {
  return (
    <SimplePublicPage
      title="Blog"
      description="Product updates, market insights, and platform announcements."
    />
  );
}
