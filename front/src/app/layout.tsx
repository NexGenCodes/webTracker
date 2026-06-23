import type { Metadata } from "next";
import { Geist, Geist_Mono } from "next/font/google";
import "./globals.css";
import { PLATFORM_NAME, APP_DESCRIPTION } from "@/constants";
import Providers from '@/components/providers/Providers';
import { getServerSession } from '@/lib/auth';

const geistSans = Geist({
  variable: "--font-geist-sans",
  subsets: ["latin"],
});

const geistMono = Geist_Mono({
  variable: "--font-geist-mono",
  subsets: ["latin"],
});

export const viewport = {
  width: "device-width",
  initialScale: 1,
  userScalable: true,
  themeColor: [
    { media: "(prefers-color-scheme: light)", color: "#fafafa" },
    { media: "(prefers-color-scheme: dark)", color: "#0a0a0a" },
  ],
};

export const metadata: Metadata = {
  title: `${PLATFORM_NAME} - Premium Logistics`,
  description: APP_DESCRIPTION,
};

import { ClientTransitionProvider } from '@/components/providers/ClientTransitionProvider';
import { Toaster } from "react-hot-toast";
import { SpeedInsights } from "@vercel/speed-insights/next";

export default async function RootLayout({
  children,
}: Readonly<{
  children: React.ReactNode;
}>) {
  const { user } = await getServerSession();

  return (
    <html lang="en" suppressHydrationWarning>
      <body
        className={`${geistSans.variable} ${geistMono.variable} antialiased`}
      >
        <a href="#main-content" className="sr-only focus:not-sr-only focus:fixed focus:top-4 focus:left-4 focus:z-[9999] focus:px-4 focus:py-2 focus:bg-accent focus:text-white focus:rounded-xl focus:text-sm focus:font-bold">
          Skip to content
        </a>
        <Providers initialUser={user} initialCompanyId={user?.company_id}>
          <Toaster position="top-left" toastOptions={{
            className: 'font-bold uppercase text-[11px] tracking-[0.1em] border shadow-2xl rounded-2xl p-4 min-w-[320px]',
            duration: 5000,
            style: {
              background: 'var(--glass-bg)',
              backdropFilter: 'blur(20px)',
              WebkitBackdropFilter: 'blur(20px)',
              color: 'var(--color-text-main)',
              border: '1px solid var(--color-border)',
              boxShadow: 'var(--glass-shadow)',
            },
            success: {
              iconTheme: {
                primary: 'var(--color-success)',
                secondary: 'white',
              },
              style: {
                borderLeft: '4px solid var(--color-success)',
              }
            },
            error: {
              iconTheme: {
                primary: 'var(--color-error)',
                secondary: 'white',
              },
              style: {
                borderLeft: '4px solid var(--color-error)',
                color: 'var(--color-error)',
              }
            }
          }} />
          <ClientTransitionProvider>
            <main id="main-content">
              {children}
            </main>
            <SpeedInsights />
          </ClientTransitionProvider>
        </Providers>
      </body>
    </html>
  );
}
