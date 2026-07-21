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
};

export default nextConfig;