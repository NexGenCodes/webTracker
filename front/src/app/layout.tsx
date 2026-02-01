import type { Metadata } from "next";
import { Geist, Geist_Mono } from "next/font/google";
import "./globals.css";
import { APP_NAME, APP_DESCRIPTION } from "@/lib/constants";
import { ThemeProvider } from "@/components/ThemeProvider";
import { I18nProvider } from "@/components/I18nContext";

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
  maximumScale: 1,
};

export const metadata: Metadata = {
  title: `${APP_NAME} - Premium Logistics`,
  description: APP_DESCRIPTION,
};

import { ClientTransitionProvider } from "@/components/ClientTransitionProvider";
import AuthProvider from "@/components/AuthProvider";
import { LayoutHeader } from "@/components/LayoutHeader";
import { Toaster } from "react-hot-toast";
import { SpeedInsights } from "@vercel/speed-insights/next";

export default function RootLayout({
  children,
}: Readonly<{
  children: React.ReactNode;
}>) {
  return (
    <html lang="en" suppressHydrationWarning>
      <body
        className={`${geistSans.variable} ${geistMono.variable} antialiased`}
      >
        <ThemeProvider attribute="class" defaultTheme="dark" enableSystem>
          <I18nProvider>
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
            <LayoutHeader />
            <ClientTransitionProvider>
              <AuthProvider>
                {children}
              </AuthProvider>
            </ClientTransitionProvider>
          </I18nProvider>
        </ThemeProvider>
        <SpeedInsights />
      </body>
    </html>
  );
}
