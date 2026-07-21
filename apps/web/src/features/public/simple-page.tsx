import { Typography } from "@coindistro/cds";

export function SimplePublicPage({
  title,
  description,
}: {
  title: string;
  description: string;
}) {
  return (
    <main className="cds-container py-24 pt-32">
      <Typography variant="h1" className="mb-4">
        {title}
      </Typography>
      <Typography variant="bodyLarge" className="max-w-2xl text-muted-foreground">
        {description}
      </Typography>
    </main>
  );
}
