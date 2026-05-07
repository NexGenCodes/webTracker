'use client';

import React, { useState } from 'react';
import {
    Package, Search, Filter, Plus,
    Eye, Edit, Trash2,
    ChevronLeft, ChevronRight,
    Loader2, CheckCircle2, Clock, Truck, AlertCircle
} from 'lucide-react';
import { motion, AnimatePresence } from 'framer-motion';
import { useQuery } from '@tanstack/react-query';
import { createClient } from '@/lib/supabase/client';
import { deleteShipmentAction, updateShipmentAction } from '@/app/actions/shipment';
import { ShipmentModal, ShipmentFormValues } from './ShipmentModal';
import toast from 'react-hot-toast';

interface ShipmentsTabProps {
    companyId: string;
    jwt?: string;
}

const STATUS_CONFIG: Record<string, { label: string; color: string; icon: React.ElementType }> = {
    'PENDING': { label: 'Pending', color: 'bg-warning/10 text-warning border-warning/20', icon: Clock },
    'IN_TRANSIT': { label: 'In Transit', color: 'bg-accent/10 text-accent border-accent/20', icon: Truck },
    'OUT_FOR_DELIVERY': { label: 'Out for Delivery', color: 'bg-primary/10 text-primary border-primary/20', icon: Truck },
    'DELIVERED': { label: 'Delivered', color: 'bg-success/10 text-success border-success/20', icon: CheckCircle2 },
    'CANCELED': { label: 'Canceled', color: 'bg-error/10 text-error border-error/20', icon: AlertCircle },
};

export function ShipmentsTab({ companyId, jwt }: ShipmentsTabProps) {
    const [search, setSearch] = useState('');
    const [statusFilter, setStatusFilter] = useState<string>('ALL');
    const [page, setPage] = useState(1);
    const [isModalOpen, setIsModalOpen] = useState(false);
    const [editingShipment, setEditingShipment] = useState<Record<string, unknown> | null>(null);
    const [selectedIds, setSelectedIds] = useState<string[]>([]);
    const pageSize = 10;

    const supabase = createClient(jwt);

    const { data, isLoading, refetch } = useQuery({
        queryKey: ['shipments-list', companyId, page, statusFilter, search],
        queryFn: async () => {
            let query = supabase
                .from('shipment')
                .select('*', { count: 'exact' })
                .eq('company_id', companyId)
                .order('created_at', { ascending: false })
                .range((page - 1) * pageSize, page * pageSize - 1);

            if (statusFilter !== 'ALL') {
                query = query.eq('status', statusFilter.toLowerCase().replace('_', ''));
            }

            if (search) {
                query = query.or(`tracking_id.ilike.%${search}%,recipient_name.ilike.%${search}%,recipient_phone.ilike.%${search}%`);
            }

            const { data, error, count } = await query;

            if (error) throw error;

            // Normalize field names if they come from DB as snake_case but components expect snake_case (which we used in Modal)
            // Actually the Modal uses snake_case because of how DB typically stores them.
            return { shipments: data, total: count || 0 };
        },
        enabled: !!companyId,
    });

    const totalPages = Math.ceil((data?.total || 0) / pageSize);

    const handleDelete = async (id: string) => {
        if (!confirm('Are you sure you want to delete this shipment?')) return;

        const result = await deleteShipmentAction(id, companyId);
        if (result.success) {
            toast.success('Shipment deleted.');
            refetch();
        } else {
            toast.error(result.error || 'Failed to delete.');
        }
    };

    const handleEdit = (shipment: Record<string, unknown>) => {
        setEditingShipment(shipment);
        setIsModalOpen(true);
    };

    const handleCreate = () => {
        setEditingShipment(null);
        setIsModalOpen(true);
    };

    const toggleSelectAll = () => {
        if (selectedIds.length === (data?.shipments?.length || 0)) {
            setSelectedIds([]);
        } else {
            setSelectedIds(data?.shipments?.map((s: Record<string, unknown>) => s.id as string) || []);
        }
    };

    const toggleSelect = (id: string) => {
        setSelectedIds(prev =>
            prev.includes(id) ? prev.filter(i => i !== id) : [...prev, id]
        );
    };

    const handleBulkDelete = async () => {
        if (!confirm(`Are you sure you want to delete ${selectedIds.length} shipments?`)) return;

        const results = await Promise.all(selectedIds.map(id => deleteShipmentAction(id, companyId)));
        const successCount = results.filter(r => r.success).length;

        toast.success(`Successfully deleted ${successCount} shipments.`);
        setSelectedIds([]);
        refetch();
    };

    const handleBulkStatusUpdate = async (status: string) => {
        const results = await Promise.all(selectedIds.map(id => updateShipmentAction(id, companyId, { status })));
        const successCount = results.filter(r => r.success).length;

        toast.success(`Successfully updated ${successCount} shipments to ${status}.`);
        setSelectedIds([]);
        refetch();
    };

    return (
        <div className="space-y-6">
            {/* Header Actions */}
            <div className="flex flex-col md:flex-row md:items-center justify-between gap-4">
                <div className="relative flex-1 max-w-md group">
                    <Search className="absolute left-4 top-1/2 -translate-y-1/2 text-text-muted group-focus-within:text-accent transition-colors" size={18} />
                    <input
                        type="text"
                        placeholder="Search tracking #, name, or phone..."
                        value={search}
                        onChange={(e) => setSearch(e.target.value)}
                        className="w-full bg-surface border border-border/50 rounded-2xl py-3.5 pl-12 pr-4 text-sm font-medium focus:outline-none focus:border-accent/50 focus:ring-4 focus:ring-accent/5 transition-all"
                    />
                </div>

                <div className="flex items-center gap-3">
                    <div className="relative">
                        <select
                            value={statusFilter}
                            onChange={(e) => setStatusFilter(e.target.value)}
                            className="appearance-none bg-surface border border-border/50 rounded-xl py-3 px-6 pr-10 text-xs font-black uppercase tracking-widest focus:outline-none focus:border-accent/50 transition-all cursor-pointer"
                        >
                            <option value="ALL">All Status</option>
                            <option value="PENDING">Pending</option>
                            <option value="IN_TRANSIT">In Transit</option>
                            <option value="OUT_FOR_DELIVERY">Out for Delivery</option>
                            <option value="DELIVERED">Delivered</option>
                            <option value="CANCELED">Canceled</option>
                        </select>
                        <Filter className="absolute right-4 top-1/2 -translate-y-1/2 text-text-muted pointer-events-none" size={14} />
                    </div>

                    <button
                        onClick={handleCreate}
                        className="btn-primary py-3 px-6 text-xs flex items-center gap-2 active:scale-95"
                    >
                        <Plus size={16} /> New Shipment
                    </button>
                </div>
            </div>

            {/* Bulk Actions Bar */}
            <AnimatePresence>
                {selectedIds.length > 0 && (
                    <motion.div
                        initial={{ opacity: 0, y: -20 }}
                        animate={{ opacity: 1, y: 0 }}
                        exit={{ opacity: 0, y: -20 }}
                        className="flex items-center justify-between bg-accent/5 border border-accent/20 rounded-2xl p-4"
                    >
                        <div className="flex items-center gap-3">
                            <span className="text-xs font-black text-accent uppercase tracking-widest px-3 py-1 bg-accent/10 rounded-full">
                                {selectedIds.length} Selected
                            </span>
                        </div>
                        <div className="flex items-center gap-3">
                            <div className="flex items-center gap-2 border-r border-accent/20 pr-3">
                                {['pending', 'intransit', 'outfordelivery', 'delivered'].map(status => (
                                    <button
                                        key={status}
                                        onClick={() => handleBulkStatusUpdate(status)}
                                        className="text-[10px] font-black uppercase tracking-widest text-accent hover:bg-accent/10 px-3 py-1.5 rounded-lg transition-colors"
                                    >
                                        Mark {status}
                                    </button>
                                ))}
                            </div>
                            <button
                                onClick={handleBulkDelete}
                                className="flex items-center gap-2 text-[10px] font-black uppercase tracking-widest text-error hover:bg-error/10 px-4 py-2 rounded-xl transition-colors"
                            >
                                <Trash2 size={14} /> Bulk Delete
                            </button>
                        </div>
                    </motion.div>
                )}
            </AnimatePresence>

            {/* Table Container */}
            <div className="glass-panel overflow-hidden border-border/50">
                <div className="overflow-x-auto">
                    <table className="w-full text-left border-collapse">
                        <thead>
                            <tr className="border-b border-border/50 bg-surface/30">
                                <th className="px-6 py-4 w-10">
                                    <input
                                        type="checkbox"
                                        checked={selectedIds.length > 0 && selectedIds.length === (data?.shipments?.length || 0)}
                                        onChange={toggleSelectAll}
                                        className="w-4 h-4 rounded border-border/50 bg-surface accent-accent cursor-pointer"
                                    />
                                </th>
                                <th className="px-6 py-4 text-[10px] font-black uppercase tracking-[0.2em] text-text-muted">Tracking ID</th>
                                <th className="px-6 py-4 text-[10px] font-black uppercase tracking-[0.2em] text-text-muted">Receiver</th>
                                <th className="px-6 py-4 text-[10px] font-black uppercase tracking-[0.2em] text-text-muted">Destination</th>
                                <th className="px-6 py-4 text-[10px] font-black uppercase tracking-[0.2em] text-text-muted">Status</th>
                                <th className="px-6 py-4 text-[10px] font-black uppercase tracking-[0.2em] text-text-muted">Date</th>
                                <th className="px-6 py-4 text-[10px] font-black uppercase tracking-[0.2em] text-text-muted text-right">Actions</th>
                            </tr>
                        </thead>
                        <tbody className="divide-y divide-border/30">
                            <AnimatePresence mode="popLayout">
                                {isLoading ? (
                                    <tr>
                                        <td colSpan={6} className="px-6 py-20 text-center">
                                            <Loader2 className="animate-spin mx-auto text-accent mb-2" size={32} />
                                            <p className="text-xs font-bold text-text-muted uppercase tracking-widest">Loading Shipments...</p>
                                        </td>
                                    </tr>
                                ) : data?.shipments?.length === 0 ? (
                                    <tr>
                                        <td colSpan={6} className="px-6 py-20 text-center">
                                            <Package className="mx-auto text-text-muted/20 mb-4" size={48} />
                                            <p className="text-sm font-bold text-text-muted">No shipments found</p>
                                        </td>
                                    </tr>
                                ) : (
                                    data?.shipments?.map((shipment: Record<string, unknown>) => {
                                        const status = (shipment.status as string).toUpperCase();
                                        const config = STATUS_CONFIG[status] || STATUS_CONFIG['PENDING'];
                                        const StatusIcon = config.icon;

                                        return (
                                            <motion.tr
                                                key={shipment.id as string}
                                                initial={{ opacity: 0 }}
                                                animate={{ opacity: 1 }}
                                                exit={{ opacity: 0 }}
                                                className={`group hover:bg-surface/50 transition-colors ${selectedIds.includes(shipment.id as string) ? 'bg-accent/5' : ''}`}
                                            >
                                                <td className="px-6 py-4">
                                                    <input
                                                        type="checkbox"
                                                        checked={selectedIds.includes(shipment.id as string)}
                                                        onChange={() => toggleSelect(shipment.id as string)}
                                                        className="w-4 h-4 rounded border-border/50 bg-surface accent-accent cursor-pointer"
                                                    />
                                                </td>
                                                <td className="px-6 py-4">
                                                    <span className="text-sm font-black text-text-main tracking-tight group-hover:text-accent transition-colors">
                                                        {shipment.tracking_id as string}
                                                    </span>
                                                </td>
                                                <td className="px-6 py-4">
                                                    <div className="flex flex-col">
                                                        <span className="text-sm font-bold text-text-main">{shipment.recipient_name as string}</span>
                                                        <span className="text-[10px] font-medium text-text-muted">{shipment.recipient_phone as string}</span>
                                                    </div>
                                                </td>
                                                <td className="px-6 py-4">
                                                    <span className="text-xs font-bold text-text-muted uppercase tracking-widest">
                                                        {(shipment.destination as string) || 'N/A'}
                                                    </span>
                                                </td>
                                                <td className="px-6 py-4">
                                                    <div className={`inline-flex items-center gap-1.5 px-3 py-1 rounded-full border text-[10px] font-black uppercase tracking-widest ${config.color}`}>
                                                        <StatusIcon size={10} />
                                                        {config.label}
                                                    </div>
                                                </td>
                                                <td className="px-6 py-4">
                                                    <span className="text-xs font-medium text-text-muted">
                                                        {new Date(shipment.created_at as string).toLocaleDateString()}
                                                    </span>
                                                </td>
                                                <td className="px-6 py-4 text-right">
                                                    <div className="flex items-center justify-end gap-2">
                                                        <button
                                                            onClick={() => window.open(`/track/${shipment.tracking_id as string}`, '_blank')}
                                                            className="p-2 hover:bg-surface rounded-lg text-text-muted hover:text-accent transition-all"
                                                            title="View Public Tracking"
                                                        >
                                                            <Eye size={16} />
                                                        </button>
                                                        <button
                                                            onClick={() => handleEdit(shipment)}
                                                            className="p-2 hover:bg-surface rounded-lg text-text-muted hover:text-primary transition-all"
                                                            title="Edit Shipment"
                                                        >
                                                            <Edit size={16} />
                                                        </button>
                                                        <button
                                                            onClick={() => handleDelete(shipment.id as string)}
                                                            className="p-2 hover:bg-error/10 rounded-lg text-text-muted hover:text-error transition-all"
                                                            title="Delete"
                                                        >
                                                            <Trash2 size={16} />
                                                        </button>
                                                    </div>
                                                </td>
                                            </motion.tr>
                                        );
                                    })
                                )}
                            </AnimatePresence>
                        </tbody>
                    </table>
                </div>

                {/* Pagination */}
                {totalPages > 1 && (
                    <div className="px-6 py-4 border-t border-border/50 bg-surface/30 flex items-center justify-between">
                        <p className="text-xs font-bold text-text-muted">
                            Showing <span className="text-text-main">{(page - 1) * pageSize + 1}</span> to <span className="text-text-main">{Math.min(page * pageSize, data?.total || 0)}</span> of <span className="text-text-main">{data?.total}</span> shipments
                        </p>
                        <div className="flex items-center gap-2">
                            <button
                                onClick={() => setPage(p => Math.max(1, p - 1))}
                                disabled={page === 1}
                                className="p-2 border border-border rounded-lg disabled:opacity-30 hover:bg-surface transition-colors"
                            >
                                <ChevronLeft size={16} />
                            </button>
                            <div className="flex items-center gap-1">
                                {[...Array(totalPages)].map((_, i) => (
                                    <button
                                        key={i}
                                        onClick={() => setPage(i + 1)}
                                        className={`w-8 h-8 rounded-lg text-xs font-black transition-all ${page === i + 1 ? 'bg-accent text-white shadow-lg shadow-accent/20' : 'hover:bg-surface text-text-muted'}`}
                                    >
                                        {i + 1}
                                    </button>
                                ))}
                            </div>
                            <button
                                onClick={() => setPage(p => Math.min(totalPages, p + 1))}
                                disabled={page === totalPages}
                                className="p-2 border border-border rounded-lg disabled:opacity-30 hover:bg-surface transition-colors"
                            >
                                <ChevronRight size={16} />
                            </button>
                        </div>
                    </div>
                )}
            </div>

            <ShipmentModal
                isOpen={isModalOpen}
                onClose={() => setIsModalOpen(false)}
                companyId={companyId}
                shipment={editingShipment as unknown as (ShipmentFormValues & { id: string; tracking_id?: string }) | null}
                onSuccess={refetch}
            />
        </div>
    );
}
