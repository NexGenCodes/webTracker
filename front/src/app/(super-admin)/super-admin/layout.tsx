import { ReactNode } from "react";
import { LayoutHeader } from "@/components/layout/LayoutHeader";
import { PageBackground } from "@/components/layout/PageBackground";

export default function SuperAdminLayout({ children }: { children: ReactNode }) {
    return (
        <>
            <PageBackground variant="dashboard" />
            <LayoutHeader />
            {children}
        </>
    );
}
