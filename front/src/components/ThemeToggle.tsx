import React, { memo, useEffect, useState } from "react";
import { useTheme } from "next-themes";
import { Sun, Moon } from "lucide-react";

export const ThemeToggle = memo(() => {
    const { theme, setTheme } = useTheme();
    const [mounted, setMounted] = useState(false);

    useEffect(() => setMounted(true), []);

    if (!mounted) return <div className="p-2 w-10 h-10" />;

    return (
        <button
            onClick={() => setTheme(theme === "dark" ? "light" : "dark")}
            className="p-2 rounded-lg bg-white/5 hover:bg-white/10 border border-white/10 transition-all text-accent group"
            aria-label="Toggle Theme"
        >
            {theme === "dark" ? (
                <Sun className="w-5 h-5 text-yellow-400 animate-in fade-in zoom-in duration-300" />
            ) : (
                <Moon className="w-5 h-5 text-indigo-600 animate-in fade-in zoom-in duration-300" />
            )}
        </button>
    );
});
