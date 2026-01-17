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
            <LayoutHeader />
            <ClientTransitionProvider>
              <AuthProvider>
                {children}
              </AuthProvider>
            </ClientTransitionProvider>
          </I18nProvider>
        </ThemeProvider>
      </body>
    </html>
  );
}
