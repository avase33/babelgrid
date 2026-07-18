/** @type {import('next').NextConfig} */
const nextConfig = {
  reactStrictMode: true,
  env: {
    NEXT_PUBLIC_SFU_URL: process.env.NEXT_PUBLIC_SFU_URL || "ws://localhost:8080",
  },
};
module.exports = nextConfig;
