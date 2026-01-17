import React from 'react';
import { Trash2, XCircle } from 'lucide-react';

interface ShipmentTableProps {
    shipments: any[];
    dataLoading: boolean;
    dict: any;
    onMarkDelivered: (trackingNumber: string) => void;
    onCancel: (trackingNumber: string) => void;
    onDelete: (trackingNumber: string) => void;
}

export const ShipmentTable: React.FC<ShipmentTableProps> = ({
    shipments,
    dataLoading,
    dict,
    onMarkDelivered,
    onCancel,
    onDelete
}) => {
    return (
        <div className="glass-panel overflow-hidden">
            <div className="overflow-x-auto">
                <table className="w-full min-w-[640px]">
                    <thead className="bg-surface-muted">
                        <tr>
                            <th className="text-left p-3 sm:p-4 text-xs font-black text-text-muted uppercase tracking-widest">Tracking ID</th>
                            <th className="text-left p-3 sm:p-4 text-xs font-black text-text-muted uppercase tracking-widest hidden sm:table-cell">{dict.admin.sender}</th>
                            <th className="text-left p-3 sm:p-4 text-xs font-black text-text-muted uppercase tracking-widest">{dict.admin.receiver}</th>
                            <th className="text-left p-3 sm:p-4 text-xs font-black text-text-muted uppercase tracking-widest">Status</th>
                            <th className="text-left p-3 sm:p-4 text-xs font-black text-text-muted uppercase tracking-widest">{dict.admin.actions}</th>
                        </tr>
                    </thead>
                    <tbody>
                        {dataLoading ? (
                            Array(8).fill({}).map((_, i) => (
                                <tr key={i} className="border-t border-border animate-pulse">
                                    <td className="p-3 sm:p-4"><div className="h-4 w-24 bg-border rounded" /></td>
                                    <td className="p-3 sm:p-4 hidden sm:table-cell"><div className="h-4 w-32 bg-border/50 rounded" /></td>
                                    <td className="p-3 sm:p-4"><div className="h-4 w-32 bg-border/50 rounded" /></td>
                                    <td className="p-3 sm:p-4"><div className="h-6 w-16 bg-border/30 rounded-lg" /></td>
                                    <td className="p-3 sm:p-4"><div className="h-8 w-20 bg-border/40 rounded-xl" /></td>
                                </tr>
                            ))
                        ) : shipments.length === 0 ? (
                            <tr>
                                <td colSpan={5} className="text-center p-8 text-text-muted">{dict.admin.noShipments}</td>
                            </tr>
                        ) : (
                            shipments.map((s) => (
                                <tr key={s.id} className="border-t border-border hover:bg-surface-muted/50 transition-colors">
                                    <td className="p-3 sm:p-4 font-mono text-xs sm:text-sm font-bold text-accent">{s.trackingNumber}</td>
                                    <td className="p-3 sm:p-4 text-xs sm:text-sm text-text-muted hidden sm:table-cell">{s.senderName || 'N/A'}</td>
                                    <td className="p-3 sm:p-4 text-xs sm:text-sm text-text-muted">{s.receiverName || 'N/A'}</td>
                                    <td className="p-3 sm:p-4">
                                        <span className={`px-2 sm:px-3 py-1 rounded-lg text-[10px] sm:text-xs font-black uppercase ${s.isArchived ? 'bg-success/10 text-success' :
                                            s.status === 'IN_TRANSIT' ? 'bg-accent/10 text-accent' :
                                                s.status === 'PENDING' ? 'bg-warning/10 text-warning' :
                                                    'bg-surface text-text-muted'
                                            }`}>
                                            {s.isArchived ? dict.admin.delivered : s.status.replace('_', ' ')}
                                        </span>
                                    </td>
                                    <td className="p-3 sm:p-4">
                                        <div className="flex gap-2">
                                            {!s.isArchived && s.status !== 'CANCELED' && (
                                                <>
                                                    <button
                                                        onClick={() => onMarkDelivered(s.trackingNumber)}
                                                        className="px-2 sm:px-3 py-1 bg-success/10 hover:bg-success text-success hover:text-white rounded-lg text-[10px] sm:text-xs font-black uppercase transition-all"
                                                        title="Mark Delivered"
                                                    >
                                                        ✓
                                                    </button>
                                                    <button
                                                        onClick={() => onCancel(s.trackingNumber)}
                                                        className="px-2 sm:px-3 py-1 bg-warning/10 hover:bg-warning text-warning hover:text-white rounded-lg text-[10px] sm:text-xs font-black uppercase transition-all"
                                                        title="Cancel Shipment"
                                                    >
                                                        <XCircle size={14} />
                                                    </button>
                                                </>
                                            )}
                                            <button
                                                onClick={() => onDelete(s.trackingNumber)}
                                                className="px-2 sm:px-3 py-1 bg-error/10 hover:bg-error text-error hover:text-white rounded-lg text-[10px] sm:text-xs font-black uppercase transition-all"
                                                title="Delete"
                                            >
                                                <Trash2 size={14} />
                                            </button>
                                        </div>
                                    </td>
                                </tr>
                            ))
                        )}
                    </tbody>
                </table>
            </div>
        </div>
    );
};
