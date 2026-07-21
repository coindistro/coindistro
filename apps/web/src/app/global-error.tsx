"use client";

import { useEffect } from "react";
import { Alert, AlertDescription, AlertTitle, Button } from "@coindistro/cds";

export default function GlobalError({
  error,
  reset,
}: {
  error: Error & { digest?: string };
  reset: () => void;
}) {
  useEffect(() => {
    console.error(error);
  }, [error]);

  return (
    <html>
      <body>
        <div className="flex min-h-screen items-center justify-center bg-background p-6">
          <div className="w-full max-w-md space-y-4 text-center">
            <div className="text-6xl font-bold text-destructive">500</div>
            <Alert variant="danger">
              <AlertTitle>Server error</AlertTitle>
              <AlertDescription>
                {error.message || "An unexpected error occurred on the server."}
              </AlertDescription>
            </Alert>
            <Button onClick={reset} className="w-full">
              Try again
            </Button>
          </div>
        </div>
      </body>
    </html>
  );
}