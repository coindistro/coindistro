import { redirect } from "next/navigation";

/** Invitation registration — reuses register with invite query param. */
export default async function InvitePage({
  searchParams,
}: {
  searchParams: Promise<{ code?: string; invite?: string }>;
}) {
  const sp = await searchParams;
  const code = sp.code || sp.invite || "";
  redirect(`/register?invite=${encodeURIComponent(code)}`);
}
