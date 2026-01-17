"use client";

import { motion, AnimatePresence } from "framer-motion";
import { usePathname } from "next/navigation";
import { ReactNode } from "react";

export function ClientTransitionProvider({ children }: { children: ReactNode }) {
    const pathname = usePathname();

    return (
        <AnimatePresence mode="wait">
            <motion.div
                key={pathname}
                initial={{ opacity: 0, y: 10, filter: "blur(10px)" }}
                animate={{ opacity: 1, y: 0, filter: "blur(0px)" }}
                exit={{ opacity: 0, y: -10, filter: "blur(10px)" }}
                transition={{
                    duration: 0.6,
                    ease: [0.16, 1, 0.3, 1] // Custom quintic ease-out
                }}
                className="flex-1 flex flex-col w-full"
            >
                {children}
            </motion.div>
        </AnimatePresence>
    );
}
