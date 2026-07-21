import { Spinner } from "@coindistro/cds";

export default function Loading() {
  return (
    <div className="flex justify-center py-12">
      <Spinner />
    </div>
  );
}
