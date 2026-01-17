import React from 'react';

interface StatsCardsProps {
    stats: {
        total: number;
        pending: number;
        inTransit: number;
        delivered: number;
    };
    dataLoading: boolean;
    dict: any;
}

export const StatsCards: React.FC<StatsCardsProps> = ({ stats, dataLoading, dict }) => {
    const statItems = [
        { label: dict.admin.totalShipments, value: stats.total, color: 'text-text-main' },
        { label: dict.admin.pending, value: stats.pending, color: 'text-warning' },
        { label: dict.admin.inTransit, value: stats.inTransit, color: 'text-accent' },
        { label: dict.admin.delivered, value: stats.delivered, color: 'text-success' }
    ];

    return (
        <div className="grid grid-cols-2 lg:grid-cols-4 gap-3 sm:gap-6">
            {statItems.map((stat, i) => (
                <div key={i} className="glass-panel p-4 sm:p-6 overflow-hidden relative">
                    {dataLoading && <div className="absolute inset-0 bg-surface-muted/50 animate-pulse pointer-events-none" />}
                    <p className="text-xs sm:text-sm font-black text-text-muted uppercase tracking-widest mb-2">{stat.label}</p>
                    <div className={`text-3xl sm:text-4xl font-black ${stat.color} transition-all duration-500 ${dataLoading ? 'opacity-20 blur-sm' : 'opacity-100 blur-0'}`}>
                        {stat.value}
                    </div>
                </div>
            ))}
        </div>
    );
};
