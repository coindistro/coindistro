/** @type {import('next').NextConfig} */
const nextConfig = {
  reactStrictMode: true,
  transpilePackages: ["@coindistro/cds"],
  experimental: {
    optimizePackageImports: ["lucide-react", "@coindistro/cds"],
  },
  images: {
    remotePatterns: [],
  },
  /**
   * Short public paths map onto the authenticated app shell under /app/*.
   * Auth guards still protect the destination routes.
   */
  async redirects() {
    return [
      { source: "/dashboard", destination: "/app/dashboard", permanent: false },
      { source: "/wallets", destination: "/app/wallet", permanent: false },
      { source: "/wallet", destination: "/app/wallet", permanent: false },
      { source: "/earn", destination: "/app/earn", permanent: false },
      { source: "/markets", destination: "/app/markets", permanent: false },
      { source: "/p2p", destination: "/app/p2p", permanent: false },
      { source: "/signals", destination: "/app/signals", permanent: false },
      { source: "/referrals", destination: "/app/referrals", permanent: false },
      { source: "/academy", destination: "/app/academy", permanent: false },
      {
        source: "/notifications",
        destination: "/app/notifications",
        permanent: false,
      },
      { source: "/settings", destination: "/app/settings", permanent: false },
      { source: "/profile", destination: "/app/profile", permanent: false },
    ];
  },
};

export default nextConfig;
