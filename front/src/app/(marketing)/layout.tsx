import React from "react";
import { LayoutHeader } from "@/components/layout/LayoutHeader";
import { PageBackground } from "@/components/layout/PageBackground";
import { Footer } from "@/components/layout/Footer";

export default function MarketingLayout({
    children,
}: {
    children: React.ReactNode;
}) {
    return (
        <>
            <PageBackground variant="marketing" />
            <LayoutHeader />
            {children}
            <div className="max-w-7xl mx-auto px-4 md:px-6 w-full mt-auto">
                <Footer />
            </div>
        </>
    );
}
