import { Spinner } from "@coindistro/cds";

export default function RootLoading() {
  return (
    <div className="flex min-h-screen items-center justify-center">
      <Spinner label="Loading Coindistro" />
    </div>
  );
}
