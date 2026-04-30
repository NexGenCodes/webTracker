import React, { memo, useEffect, useState } from "react";
import { useTheme } from "next-themes";
import { Sun, Moon } from "lucide-react";

const ThemeToggleComponent = memo(() => {
    const { theme, resolvedTheme, setTheme } = useTheme();
    const [mounted, setMounted] = useState(false);

    useEffect(() => setMounted(true), []);

    if (!mounted) return <div className="p-2 w-10 h-10" />;

    const currentTheme = resolvedTheme || theme;

    return (
        <button
            onClick={() => setTheme(currentTheme === "dark" ? "light" : "dark")}
            className="p-2 rounded-lg bg-white/5 hover:bg-white/10 border border-white/10 transition-all text-accent group"
            aria-label="Toggle Theme"
        >
            {currentTheme === "dark" ? (
                <Sun className="w-5 h-5 text-yellow-400 animate-in fade-in zoom-in duration-300" />
            ) : (
                <Moon className="w-5 h-5 text-indigo-600 animate-in fade-in zoom-in duration-300" />
            )}
        </button>
    );
});

ThemeToggleComponent.displayName = 'ThemeToggle';
export const ThemeToggle = ThemeToggleComponent;
