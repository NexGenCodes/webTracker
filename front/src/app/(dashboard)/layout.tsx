import { ReactNode } from "react";
import { LayoutHeader } from "@/components/layout/LayoutHeader";

export default function DashboardLayout({ children }: { children: ReactNode }) {
    return (
        <>
            <LayoutHeader />
            {children}
        </>
    );
}
