import { PlaceholderPage } from "@/features/shared/components/placeholder-page";

export const metadata = { title: "Users" };

export default function Page() {
  return (
    <PlaceholderPage
      title="Users"
      description="Identity users, roles, and status management."
      module="admin-users"
    />
  );
}
