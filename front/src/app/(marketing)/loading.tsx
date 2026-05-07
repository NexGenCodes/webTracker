export default function MarketingLoading() {
    return (
        <main className="min-h-screen flex flex-col items-center overflow-x-hidden relative animate-pulse">
            <div className="w-full max-w-7xl mx-auto px-4 md:px-6 pt-28 md:pt-36">
                {/* Hero skeleton */}
                <div className="text-center space-y-6 mb-16">
                    <div className="h-6 w-40 bg-surface-muted rounded-full mx-auto" />
                    <div className="h-14 w-3/4 bg-surface-muted rounded-xl mx-auto" />
                    <div className="h-6 w-2/3 bg-surface-muted rounded-lg mx-auto" />
                    <div className="h-14 w-full max-w-lg bg-surface-muted rounded-2xl mx-auto mt-8" />
                </div>

                {/* Feature cards skeleton */}
                <div className="grid grid-cols-1 md:grid-cols-3 gap-8 mt-24">
                    {[1, 2, 3].map((i) => (
                        <div key={i} className="h-56 bg-surface-muted rounded-2xl border border-border/30" />
                    ))}
                </div>
            </div>
        </main>
    );
}
