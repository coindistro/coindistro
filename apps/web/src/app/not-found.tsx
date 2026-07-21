import Link from "next/link";
import { Button, EmptyState } from "@coindistro/cds";
import { FileQuestion } from "lucide-react";

export default function NotFound() {
  return (
    <div className="flex min-h-screen items-center justify-center p-6">
      <EmptyState
        icon={<FileQuestion className="h-10 w-10" />}
        title="Page not found"
        description="The page you are looking for does not exist or has moved."
        actionLabel="Go home"
        onAction={undefined}
      />
      <div className="fixed bottom-10">
        <Button asChild>
          <Link href="/">Return home</Link>
        </Button>
      </div>
    </div>
  );
}
