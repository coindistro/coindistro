import { redirect } from "next/navigation";

/** Referral registration deep link. */
export default async function ReferralPage({
  searchParams,
}: {
  searchParams: Promise<{ code?: string; ref?: string }>;
}) {
  const sp = await searchParams;
  const code = sp.ref || sp.code || "";
  redirect(`/register?ref=${encodeURIComponent(code)}`);
}
