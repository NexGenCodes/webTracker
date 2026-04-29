import React from "react";
import { LayoutHeader } from "@/components/layout/LayoutHeader";

export default function MarketingLayout({
    children,
}: {
    children: React.ReactNode;
}) {
    return (
        <>
            <LayoutHeader />
            {children}
        </>
    );
}
