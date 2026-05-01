"use client";

import React, { useState } from "react";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { ThemeProvider } from "./ThemeProvider";
import { I18nProvider } from "./I18nContext";
import { SpeedInsights } from "@vercel/speed-insights/next";
import { Analytics } from "@vercel/analytics/react";
import MultiTenantProvider from "./MultiTenantProvider";

export default function Providers({ 
  children,
  initialUser = null,
  initialCompanyId = null,
}: { 
  children: React.ReactNode;
  initialUser?: { email: string; company_name: string; plan_type: string } | null;
  initialCompanyId?: string | null;
}) {
  const [queryClient] = useState(
    () =>
      new QueryClient({
        defaultOptions: {
          queries: {
            staleTime: 60 * 1000,
            retry: 1,
          },
        },
      })
  );

  return (
    <MultiTenantProvider initialUser={initialUser} initialCompanyId={initialCompanyId}>
      <QueryClientProvider client={queryClient}>
        <ThemeProvider attribute="class" defaultTheme="system" enableSystem disableTransitionOnChange>
          <I18nProvider>
            {children}
            <SpeedInsights />
            <Analytics />
          </I18nProvider>
        </ThemeProvider>
      </QueryClientProvider>
    </MultiTenantProvider>
  );
}
