"use client";

/**
 * PageBackground — Global ambient background layer.
 * 
 * Renders the dot-grid, star field, topography texture, shooting stars,
 * and gradient blurs that give the app its signature premium feel.
 * 
 * Use `variant="marketing"` for the full, high-opacity effect (landing pages).
 * Use `variant="dashboard"` for a subtler, less distracting version.
 */

interface PageBackgroundProps {
    variant?: "marketing" | "dashboard";
}

export function PageBackground({ variant = "marketing" }: PageBackgroundProps) {
    const isMarketing = variant === "marketing";

    return (
        <div className="fixed inset-0 z-0 pointer-events-none overflow-hidden" aria-hidden="true">
            {/* Dot grid */}
            <div className={`absolute inset-0 bg-dot-grid ${isMarketing ? "opacity-[0.1]" : "opacity-[0.06]"}`} />

            {/* Star field */}
            <div className={`bg-stars-layer ${isMarketing ? "opacity-[0.4]" : "opacity-[0.2]"}`} />

            {/* Topography texture */}
            <div className={`absolute inset-0 bg-topography ${isMarketing ? "opacity-[0.2]" : "opacity-[0.1]"}`} />

            {/* Shooting stars */}
            <div className="shooting-star" style={{ top: "10%", left: "80%", animationDelay: "2s" }} />
            <div className="shooting-star" style={{ top: "30%", left: "40%", animationDelay: "7s" }} />
            {isMarketing && (
                <div className="shooting-star" style={{ top: "50%", left: "90%", animationDelay: "15s" }} />
            )}

            {/* Ambient gradient blurs */}
            <div className={`absolute top-0 right-0 rounded-full blur-[120px] ${isMarketing ? "w-[600px] h-[600px] bg-accent/5" : "w-[400px] h-[400px] bg-accent/3"}`} />
            <div className={`absolute bottom-0 left-0 rounded-full blur-[100px] ${isMarketing ? "w-[400px] h-[400px] bg-primary/5" : "w-[300px] h-[300px] bg-primary/3"}`} />
        </div>
    );
}
