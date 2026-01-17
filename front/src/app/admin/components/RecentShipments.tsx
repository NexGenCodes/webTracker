import React from 'react';

interface RecentShipmentsProps {
    shipments: any[];
    dataLoading: boolean;
    dict: any;
}

export const RecentShipments: React.FC<RecentShipmentsProps> = ({ shipments, dataLoading, dict }) => {
    return (
        <div className="glass-panel p-4 sm:p-6 relative overflow-hidden">
            {dataLoading && (
                <div className="absolute inset-0 bg-surface-muted/10 backdrop-blur-[2px] z-10 flex items-center justify-center">
                    <div className="flex flex-col items-center gap-2">
                        <div className="w-10 h-10 border-4 border-accent border-t-transparent rounded-full animate-spin" />
                        <span className="text-[10px] font-black uppercase tracking-widest text-text-muted animate-pulse">Syncing...</span>
                    </div>
                </div>
            )}

            <h3 className="text-base sm:text-lg font-black text-text-main mb-4 uppercase tracking-tight">Recent Shipments</h3>
            <div className="space-y-2 sm:space-y-3">
                {shipments.length === 0 && !dataLoading ? (
                    <p className="text-center py-8 text-text-muted text-sm">{dict.admin.noShipments}</p>
                ) : (
                    (dataLoading ? Array(5).fill({}) : shipments.slice(0, 5)).map((s, i) => (
                        <div key={s.id || i} className={`flex flex-col sm:flex-row sm:justify-between sm:items-center gap-2 p-3 sm:p-4 bg-surface-muted rounded-xl border border-border transition-all duration-500 ${dataLoading ? 'animate-pulse' : ''}`}>
                            <div className="flex-1 min-w-0">
                                {dataLoading ? (
                                    <>
                                        <div className="h-4 w-32 bg-border rounded-lg mb-2" />
                                        <div className="h-3 w-48 bg-border/50 rounded-lg" />
                                    </>
                                ) : (
                                    <>
                                        <p className="font-mono text-xs sm:text-sm font-black text-accent truncate">{s.trackingNumber}</p>
                                        <p className="text-xs sm:text-sm text-text-muted truncate">{s.receiverName || 'N/A'}</p>
                                    </>
                                )}
                            </div>
                            {!dataLoading && (
                                <span className={`px-2 sm:px-3 py-1 rounded-lg text-[10px] sm:text-xs font-black uppercase self-start sm:self-auto ${s.isArchived ? 'bg-success/10 text-success' :
                                    s.status === 'IN_TRANSIT' ? 'bg-accent/10 text-accent' :
                                        s.status === 'PENDING' ? 'bg-warning/10 text-warning' :
                                            'bg-surface text-text-muted'
                                    }`}>
                                    {s.isArchived ? dict.admin.delivered : s.status?.replace('_', ' ')}
                                </span>
                            )}
                            {dataLoading && <div className="w-20 h-5 bg-border rounded-lg" />}
                        </div>
                    ))
                )}
            </div>
        </div>
    );
};
