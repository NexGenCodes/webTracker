"use client";

import { useState, useEffect, useCallback } from 'react';
import { createShipment, getAdminDashboardData, deleteShipment, bulkDeleteDelivered, markAsDelivered, cancelShipment, getAdminShipments, resendReceipt } from '../actions/shipment';
import { CreateShipmentDto, ShipmentData, DashboardStats, Pagination } from '@/types/shipment';
import { LayoutDashboard, List, Package, PlusCircle, Search, Filter, ChevronLeft, ChevronRight } from 'lucide-react';
import { toast } from 'react-hot-toast';

// Components
import { useI18n } from '@/components/I18nContext';
import { signIn, useSession } from "next-auth/react";
import { StatsCards } from './components/StatsCards';
import { RecentShipments } from './components/RecentShipments';
import { ShipmentTable } from './components/ShipmentTable';
import { SuccessDisplay } from './components/SuccessDisplay';
import { ShipmentForm } from './components/ShipmentForm';

type Tab = 'dashboard' | 'manage' | 'create';

export default function AdminPage() {
    const { dict } = useI18n();
    const { status } = useSession();
    const [activeTab, setActiveTab] = useState<Tab>('dashboard');

    const [trackingId, setTrackingId] = useState<string | null>(null);
    const [createdShipment, setCreatedShipment] = useState<CreateShipmentDto | null>(null);
    const [error, setError] = useState<string | null>(null);
    const [loading, setLoading] = useState(false);
    const [copied, setCopied] = useState(false);
    
    // Dashboard Data
    const [recentShipments, setRecentShipments] = useState<ShipmentData[]>([]);
    const [stats, setStats] = useState<DashboardStats>({ total: 0, inTransit: 0, outForDelivery: 0, delivered: 0, pending: 0, canceled: 0 });
    const [dataLoading, setDataLoading] = useState(false);

    // Manage (Paginated) Data
    const [paginatedShipments, setPaginatedShipments] = useState<ShipmentData[]>([]);
    const [pagination, setPagination] = useState<Pagination | null>(null);
    const [currentPage, setCurrentPage] = useState(1);
    const [searchQuery, setSearchQuery] = useState('');
    const [statusFilter, setStatusFilter] = useState('');
    const [tableLoading, setTableLoading] = useState(false);

    const loadDashboardData = useCallback(async () => {
        setDataLoading(true);
        try {
            const result = await getAdminDashboardData();
            if (result.success && result.data) {
                setRecentShipments(result.data.shipments);
                setStats(result.data.stats);
            }
        } finally {
            setDataLoading(false);
        }
    }, []);

    const loadPaginatedShipments = useCallback(async () => {
        setTableLoading(true);
        try {
            const result = await getAdminShipments({
                page: currentPage,
                limit: 20,
                search: searchQuery,
                status: statusFilter
            });
            if (result.success && result.data) {
                setPaginatedShipments(result.data.data);
                setPagination(result.data.pagination);
            }
        } finally {
            setTableLoading(false);
        }
    }, [currentPage, searchQuery, statusFilter]);

    useEffect(() => {
        if (status === 'authenticated') {
            if (activeTab === 'dashboard') loadDashboardData();
            if (activeTab === 'manage') loadPaginatedShipments();
        }
    }, [status, activeTab, loadDashboardData, loadPaginatedShipments]);

    // Debounced search for manage tab
    useEffect(() => {
        if (activeTab === 'manage' && status === 'authenticated') {
            const timer = setTimeout(() => {
                setCurrentPage(1);
                loadPaginatedShipments();
            }, 500);
            return () => clearTimeout(timer);
        }
    }, [searchQuery, activeTab, status, loadPaginatedShipments]);

    const refreshData = () => {
        if (activeTab === 'dashboard') loadDashboardData();
        if (activeTab === 'manage') loadPaginatedShipments();
    };

    const handleCreateShipment = async (data: CreateShipmentDto) => {
        setError(null);
        setLoading(true);
        try {
            const result = await createShipment(data);
            if (result.success) {
                setTrackingId(result.data?.trackingNumber ?? null);
                setCreatedShipment(data);
                refreshData();
            } else {
                setError(result.error ?? dict.admin.failedCreate);
            }
        } catch (err: unknown) {
            setError(err instanceof Error ? err.message : dict.admin.failedCreate);
        } finally {
            setLoading(false);
        }
    };

    const handleCopy = async () => {
        if (trackingId) {
            await navigator.clipboard.writeText(trackingId);
            setCopied(true);
            setTimeout(() => setCopied(false), 2000);
        }
    };

    const handleBack = () => {
        setTrackingId(null);
        setCreatedShipment(null);
        setCopied(false);
    };

    const handleDelete = async (trackingNumber: string) => {
        toast((t) => (
            <div className="flex flex-col gap-4">
                <p className="font-bold text-sm">Delete shipment {trackingNumber}?</p>
                <div className="flex gap-2">
                    <button onClick={() => toast.dismiss(t.id)} className="px-3 py-1.5 bg-surface-muted rounded-lg text-xs font-bold">Cancel</button>
                    <button onClick={async () => {
                        toast.dismiss(t.id);
                        const result = await deleteShipment(trackingNumber);
                        if (result.success) {
                            toast.success('Shipment deleted');
                            refreshData();
                        }
                    }} className="px-3 py-1.5 bg-error text-white rounded-lg text-xs font-bold">Delete</button>
                </div>
            </div>
        ), { duration: 5000 });
    };

    const handleBulkDelete = async () => {
        toast((t) => (
            <div className="flex flex-col gap-4">
                <p className="font-bold text-sm">Bulk delete all delivered shipments?</p>
                <div className="flex gap-2">
                    <button onClick={() => toast.dismiss(t.id)} className="px-3 py-1.5 bg-surface-muted rounded-lg text-xs font-bold">Cancel</button>
                    <button onClick={async () => {
                        toast.dismiss(t.id);
                        const result = await bulkDeleteDelivered();
                        if (result.success) {
                            toast.success('Cleaned up delivered shipments');
                            refreshData();
                        }
                    }} className="px-3 py-1.5 bg-error text-white rounded-lg text-xs font-bold">Confirm</button>
                </div>
            </div>
        ), { duration: 5000 });
    };

    const handleMarkDelivered = async (trackingNumber: string) => {
        const result = await markAsDelivered(trackingNumber);
        if (result.success) {
            toast.success('Marked as delivered');
            refreshData();
        }
    };

    const handleCancel = async (trackingNumber: string) => {
        toast((t) => (
            <div className="flex flex-col gap-4">
                <p className="font-bold text-sm">Cancel shipment {trackingNumber}?</p>
                <div className="flex gap-2">
                    <button onClick={() => toast.dismiss(t.id)} className="px-3 py-1.5 bg-surface-muted rounded-lg text-xs font-bold">No, Keep</button>
                    <button onClick={async () => {
                        toast.dismiss(t.id);
                        const result = await cancelShipment(trackingNumber);
                        if (result.success) {
                            toast.success('Shipment canceled');
                            refreshData();
                        } else {
                            toast.error(result.error || 'Failed to cancel');
                        }
                    }} className="px-3 py-1.5 bg-warning text-white rounded-lg text-xs font-bold">Yes, Cancel</button>
                </div>
            </div>
        ), { duration: 5000 });
    };

    const handleResendReceipt = async (trackingNumber: string) => {
        toast((t) => (
            <div className="flex flex-col gap-4">
                <p className="font-bold text-sm">Resend receipt for {trackingNumber} to WhatsApp?</p>
                <div className="flex gap-2">
                    <button onClick={() => toast.dismiss(t.id)} className="px-3 py-1.5 bg-surface-muted rounded-lg text-xs font-bold">Cancel</button>
                    <button onClick={async () => {
                        toast.dismiss(t.id);
                        const toastId = toast.loading('Resending receipt...');
                        const result = await resendReceipt(trackingNumber);
                        if (result.success) {
                            toast.success('Receipt successfully sent to WhatsApp!', { id: toastId });
                        } else {
                            toast.error(result.error || 'Failed to resend receipt', { id: toastId });
                        }
                    }} className="px-3 py-1.5 bg-primary text-surface-bg rounded-lg text-xs font-bold">Yes, Resend</button>
                </div>
            </div>
        ), { duration: 5000 });
    };

    useEffect(() => {
        if (status === "unauthenticated") {
            signIn();
        }
    }, [status]);

    if (status === "loading" || status === "unauthenticated") {
        return (
            <div className="min-h-screen flex items-center justify-center p-4">
                <div className="flex flex-col items-center gap-4 animate-pulse">
                    <div className="bg-accent p-3 rounded-2xl">
                        <Package className="text-white" size={32} />
                    </div>
                    <p className="text-text-muted font-black uppercase tracking-widest text-sm">
                        {status === "loading" ? "Initializing..." : "Redirecting to login..."}
                    </p>
                </div>
            </div>
        );
    }

    if (trackingId && createdShipment) {
        return (
            <SuccessDisplay
                trackingId={trackingId}
                shipmentData={createdShipment}
                copied={copied}
                onCopy={handleCopy}
                onBack={handleBack}
                onResendReceipt={handleResendReceipt}
                dict={dict}
            />
        );
    }

    return (
        <div className="min-h-screen pt-24 md:pt-32 px-6">
            <div className="max-w-7xl mx-auto mb-6 sm:mb-8">
                <div className="flex gap-2 sm:gap-4 border-b border-border overflow-x-auto">
                    {[
                        { id: 'dashboard', icon: LayoutDashboard, label: dict.admin.dashboard },
                        { id: 'manage', icon: List, label: dict.admin.manageShipments },
                        { id: 'create', icon: PlusCircle, label: dict.admin.createNew }
                    ].map((tab) => (
                        <button
                            key={tab.id}
                            onClick={() => setActiveTab(tab.id as Tab)}
                            className={`flex items-center gap-2 px-4 sm:px-6 py-3 sm:py-4 font-black text-xs sm:text-sm uppercase tracking-widest transition-all whitespace-nowrap ${activeTab === tab.id
                                ? 'text-accent border-b-2 border-accent'
                                : 'text-text-muted hover:text-text-main'
                                }`}
                        >
                            <tab.icon size={16} />
                            {tab.label}
                        </button>
                    ))}
                </div>
            </div>

            <div className="max-w-7xl mx-auto">
                {activeTab === 'dashboard' && (
                    <div className="space-y-6 sm:space-y-8 animate-fade-in">
                        <StatsCards stats={stats} dataLoading={dataLoading} dict={dict} />
                        <RecentShipments shipments={recentShipments} dataLoading={dataLoading} dict={dict} />
                    </div>
                )}

                {activeTab === 'manage' && (
                    <div className="space-y-4 sm:space-y-6 animate-fade-in">
                        <div className="grid grid-cols-1 sm:grid-cols-12 gap-3 sm:gap-4">
                            <div className="sm:col-span-6 relative">
                                <Search className="absolute left-3 sm:left-4 top-1/2 -translate-y-1/2 text-text-muted" size={18} />
                                <input
                                    type="text"
                                    placeholder={dict.admin.search}
                                    className="w-full pl-10 sm:pl-12 pr-4 py-3 bg-surface-muted border border-border rounded-xl text-text-main focus:border-accent outline-none transition-all text-sm sm:text-base"
                                    value={searchQuery}
                                    onChange={(e) => setSearchQuery(e.target.value)}
                                />
                            </div>
                            <div className="sm:col-span-3 relative">
                                <Filter className="absolute left-3 top-1/2 -translate-y-1/2 text-text-muted" size={16} />
                                <select 
                                    className="w-full pl-10 pr-4 py-3 bg-surface-muted border border-border rounded-xl text-text-main focus:border-accent outline-none transition-all text-xs uppercase font-bold appearance-none"
                                    value={statusFilter}
                                    onChange={(e) => {
                                        setStatusFilter(e.target.value);
                                        setCurrentPage(1);
                                    }}
                                >
                                    <option value="">All Statuses</option>
                                    <option value="pending">Pending</option>
                                    <option value="in_transit">In Transit</option>
                                    <option value="out_for_delivery">Out For Delivery</option>
                                    <option value="delivered">Delivered</option>
                                    <option value="canceled">Canceled</option>
                                </select>
                            </div>
                            <button
                                onClick={handleBulkDelete}
                                className="sm:col-span-3 px-4 py-3 bg-error/10 hover:bg-error text-error hover:text-white rounded-xl font-black text-xs uppercase tracking-widest transition-all whitespace-nowrap"
                            >
                                {dict.admin.bulkDelete}
                            </button>
                        </div>

                        <ShipmentTable
                            shipments={paginatedShipments}
                            dataLoading={tableLoading}
                            dict={dict}
                            onMarkDelivered={handleMarkDelivered}
                            onCancel={handleCancel}
                            onDelete={handleDelete}
                            onResendReceipt={handleResendReceipt}
                        />

                        {pagination && pagination.totalPages > 1 && (
                            <div className="flex items-center justify-between mt-6 bg-surface-muted p-4 rounded-2xl border border-border">
                                <p className="text-[10px] font-bold uppercase tracking-widest text-text-muted">
                                    Showing {(currentPage - 1) * pagination.limit + 1} - {Math.min(currentPage * pagination.limit, pagination.total)} of {pagination.total}
                                </p>
                                <div className="flex gap-2">
                                    <button 
                                        disabled={currentPage === 1}
                                        onClick={() => setCurrentPage(prev => prev - 1)}
                                        className="p-2 bg-bg border border-border rounded-lg disabled:opacity-30 hover:border-accent transition-all"
                                    >
                                        <ChevronLeft size={16} />
                                    </button>
                                    <div className="flex items-center px-4 bg-bg border border-border rounded-lg text-xs font-black">
                                        {currentPage} / {pagination.totalPages}
                                    </div>
                                    <button 
                                        disabled={currentPage === pagination.totalPages}
                                        onClick={() => setCurrentPage(prev => prev + 1)}
                                        className="p-2 bg-bg border border-border rounded-lg disabled:opacity-30 hover:border-accent transition-all"
                                    >
                                        <ChevronRight size={16} />
                                    </button>
                                </div>
                            </div>
                        )}
                    </div>
                )}

                {activeTab === 'create' && (
                    <ShipmentForm
                        onSubmit={handleCreateShipment}
                        loading={loading}
                        error={error}
                    />
                )}
            </div>
        </div>
    );
}
