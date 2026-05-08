export default function DashboardLoading() {
    return (
        <div className="pb-32 md:pb-24 relative bg-background overflow-x-hidden">
            <div className="max-w-6xl mx-auto z-10 relative pt-24 md:pt-32 px-4 sm:px-8 animate-pulse">
                {/* Header skeleton */}
                <div className="flex flex-col md:flex-row md:items-end justify-between gap-6 mb-12">
                    <div className="flex items-center gap-6">
                        <div className="w-16 h-16 rounded-2xl bg-surface-muted" />
                        <div className="space-y-3">
                            <div className="h-10 w-56 bg-surface-muted rounded-lg" />
                            <div className="h-4 w-40 bg-surface-muted rounded-md" />
                        </div>
                    </div>
                </div>

                {/* Tab bar skeleton */}
                <div className="flex gap-4 border-b border-border/50 mb-10 pb-2">
                    <div className="h-10 w-28 bg-surface-muted rounded-xl" />
                    <div className="h-10 w-28 bg-surface-muted rounded-xl" />
                </div>

                {/* Stats cards skeleton */}
                <div className="grid grid-cols-1 sm:grid-cols-3 gap-6 mb-10">
                    {[1, 2, 3].map((i) => (
                        <div key={i} className="h-32 bg-surface-muted rounded-2xl border border-border/30" />
                    ))}
                </div>

                {/* Content skeleton */}
                <div className="space-y-4">
                    <div className="h-48 bg-surface-muted rounded-2xl border border-border/30" />
                    <div className="h-32 bg-surface-muted rounded-2xl border border-border/30" />
                </div>
            </div>
        </div>
    );
}
