"use client";

import { useState, useEffect } from 'react';
import { createShipment, getAdminDashboardData, deleteShipment, bulkDeleteDelivered, markAsDelivered, cancelShipment } from '../actions/shipment';
import { ShipmentForm } from './components/ShipmentForm';
import { CreateShipmentDto } from '@/types/shipment';
import { ChevronLeft, LayoutDashboard, List, Package, PlusCircle, Search } from 'lucide-react';

// Components
import { useI18n } from '@/components/I18nContext';
import { ThemeToggle } from '@/components/ThemeToggle';
import { LanguageToggle } from '@/components/LanguageToggle';

import { signIn, signOut, useSession } from "next-auth/react";
import { Logo } from '@/components/Logo';
import { StatsCards } from './components/StatsCards';
import { RecentShipments } from './components/RecentShipments';
import { ShipmentTable } from './components/ShipmentTable';
import { SuccessDisplay } from './components/SuccessDisplay';


type Tab = 'dashboard' | 'manage' | 'create';

export default function AdminPage() {
    const { dict } = useI18n();
    const { status } = useSession();
    const [activeTab, setActiveTab] = useState<Tab>('dashboard');

    const [trackingId, setTrackingId] = useState<string | null>(null);
    const [error, setError] = useState<string | null>(null);
    const [loading, setLoading] = useState(false);
    const [copied, setCopied] = useState(false);
    const [shipments, setShipments] = useState<any[]>([]);
    const [searchQuery, setSearchQuery] = useState('');
    const [stats, setStats] = useState({ total: 0, inTransit: 0, delivered: 0, pending: 0, canceled: 0 });
    const [dataLoading, setDataLoading] = useState(false);

    const loadShipments = async () => {
        setDataLoading(true);
        try {
            const result = await getAdminDashboardData();
            if (result.success && result.data) {
                setShipments(result.data.shipments);
                setStats(result.data.stats);
            }
        } finally {
            setDataLoading(false);
        }
    };

    useEffect(() => {
        if (status === 'authenticated') {
            loadShipments();
        }
    }, [status]);

    const handleCreateShipment = async (data: CreateShipmentDto) => {
        setError(null);
        setLoading(true);
        try {
            const result = await createShipment(data);
            if (result.success) {
                setTrackingId(result.data?.trackingNumber ?? null);
                loadShipments();
            } else {
                setError(result.error ?? dict.admin.failedCreate);
            }
        } catch (err: any) {
            setError(err.message || dict.admin.failedCreate);
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

        setCopied(false);
    };

    const handleDelete = async (trackingNumber: string) => {
        if (confirm(dict.admin.confirmDelete)) {
            const result = await deleteShipment(trackingNumber);
            if (result.success) {
                loadShipments();
            }
        }
    };

    const handleBulkDelete = async () => {
        if (confirm(dict.admin.confirmBulkDelete)) {
            const result = await bulkDeleteDelivered();
            if (result.success) {
                loadShipments();
            }
        }
    };

    const handleMarkDelivered = async (trackingNumber: string) => {
        const result = await markAsDelivered(trackingNumber);
        if (result.success) {
            loadShipments();
        }
    };

    const handleCancel = async (trackingNumber: string) => {
        if (confirm('Are you sure you want to cancel this shipment?')) {
            const result = await cancelShipment(trackingNumber);
            if (result.success) {
                loadShipments();
            } else {
                alert(result.error);
            }
        }
    };

    const filteredShipments = shipments.filter(s =>
        s.trackingNumber.toLowerCase().includes(searchQuery.toLowerCase()) ||
        s.receiverName?.toLowerCase().includes(searchQuery.toLowerCase()) ||
        s.senderName?.toLowerCase().includes(searchQuery.toLowerCase())
    );

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

    if (trackingId) {
        return (
            <SuccessDisplay
                trackingId={trackingId}
                copied={copied}
                onCopy={handleCopy}
                onBack={handleBack}
                dict={dict}
            />
        );
    }

    return (
        <div className="min-h-screen pt-24 md:pt-32 px-6">

            {/* Tabs */}
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
                {/* Dashboard Tab */}
                {activeTab === 'dashboard' && (
                    <div className="space-y-6 sm:space-y-8 animate-fade-in">
                        <StatsCards stats={stats} dataLoading={dataLoading} dict={dict} />
                        <RecentShipments shipments={shipments} dataLoading={dataLoading} dict={dict} />
                    </div>
                )}

                {/* Manage Tab */}
                {activeTab === 'manage' && (
                    <div className="space-y-4 sm:space-y-6 animate-fade-in">
                        <div className="flex flex-col sm:flex-row gap-3 sm:gap-4">
                            <div className="relative flex-1">
                                <Search className="absolute left-3 sm:left-4 top-1/2 -translate-y-1/2 text-text-muted" size={18} />
                                <input
                                    type="text"
                                    placeholder={dict.admin.search}
                                    className="w-full pl-10 sm:pl-12 pr-4 py-3 bg-surface-muted border border-border rounded-xl text-text-main focus:border-accent outline-none transition-all text-sm sm:text-base"
                                    value={searchQuery}
                                    onChange={(e) => setSearchQuery(e.target.value)}
                                />
                            </div>
                            <button
                                onClick={handleBulkDelete}
                                className="px-4 sm:px-6 py-3 bg-error/10 hover:bg-error text-error hover:text-white rounded-xl font-black text-xs sm:text-sm uppercase tracking-widest transition-all whitespace-nowrap"
                            >
                                {dict.admin.bulkDelete}
                            </button>
                        </div>

                        <ShipmentTable
                            shipments={filteredShipments}
                            dataLoading={dataLoading}
                            dict={dict}
                            onMarkDelivered={handleMarkDelivered}
                            onCancel={handleCancel}
                            onDelete={handleDelete}
                        />
                    </div>
                )}

                {/* Create Tab */}
                {activeTab === 'create' && (
                    <ShipmentForm
                        onSubmit={handleCreateShipment}
                        loading={loading}
                        error={error}
                        marketingDict={dict}
                    />
                )}
            </div>
        </div>
    );
}
