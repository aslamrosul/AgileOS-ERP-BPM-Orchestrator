import createNextIntlPlugin from 'next-intl/plugin';

const withNextIntl = createNextIntlPlugin('./i18n/request.ts');

/** @type {import('next').NextConfig} */
const nextConfig = {
  reactStrictMode: true,
  
  // Environment variables for production
  env: {
    NEXT_PUBLIC_API_URL: process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080',
  },

  // Performance Optimizations
  compiler: {
    // Remove console logs in production
    removeConsole: process.env.NODE_ENV === 'production',
  },

  // Image optimization
  images: {
    formats: ['image/avif', 'image/webp'],
    deviceSizes: [640, 750, 828, 1080, 1200, 1920, 2048, 3840],
    imageSizes: [16, 32, 48, 64, 96, 128, 256, 384],
  },

  // Webpack optimizations - simplified to avoid vendor chunk issues
  webpack: (config) => {
    return config;
  },

  // Experimental features for better performance
  experimental: {
    optimizePackageImports: ['recharts', 'reactflow', '@reactflow/core'],
  },

  // Compression
  compress: true,

  // Power by header removal for security
  poweredByHeader: false,
};

export default withNextIntl(nextConfig);
