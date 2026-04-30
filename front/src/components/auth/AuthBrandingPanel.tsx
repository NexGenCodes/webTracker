"use client";

import React from 'react';
import { Package, Truck, Globe, ShieldCheck, Zap, MessageSquare } from 'lucide-react';
import { motion } from 'framer-motion';
import { PLATFORM_NAME } from '@/constants';
import { useI18n } from '@/components/providers/I18nContext';

export function AuthBrandingPanel() {
    const { dict } = useI18n();

    const features = [
        { icon: Truck, label: dict.auth?.featTracking || 'Real-Time Tracking', desc: dict.auth?.featTrackingDesc || 'Live GPS shipment visibility' },
        { icon: MessageSquare, label: dict.auth?.featWhatsApp || 'WhatsApp Alerts', desc: dict.auth?.featWhatsAppDesc || 'Automated customer notifications' },
        { icon: ShieldCheck, label: dict.auth?.featSecurity || 'Enterprise Security', desc: dict.auth?.featSecurityDesc || 'End-to-end encrypted data' },
        { icon: Zap, label: dict.auth?.featAI || 'AI-Powered Parsing', desc: dict.auth?.featAIDesc || 'Instant manifest extraction' },
    ];

    const stats = [
        { value: "2,400+", label: dict.auth?.statCompanies || 'Companies' },
        { value: "1.2M", label: dict.auth?.statShipments || 'Shipments Tracked' },
        { value: "99.9%", label: dict.auth?.statUptime || 'Uptime' },
    ];
    return (
        <div className="hidden lg:flex lg:w-1/2 relative overflow-hidden bg-primary">
            {/* Gradient mesh background */}
            <div className="absolute inset-0">
                <div className="absolute inset-0 bg-gradient-to-br from-accent/20 via-transparent to-accent-deep/30" />
                <div className="absolute top-0 right-0 w-[500px] h-[500px] bg-accent/10 blur-[120px] rounded-full" />
                <div className="absolute bottom-0 left-0 w-[400px] h-[400px] bg-accent-deep/15 blur-[100px] rounded-full" />
                <div className="absolute inset-0 bg-dot-grid opacity-[0.06]" />
                {/* Animated floating orbs */}
                <motion.div
                    className="absolute top-[20%] right-[15%] w-3 h-3 rounded-full bg-accent/40"
                    animate={{ y: [-10, 10, -10], opacity: [0.3, 0.7, 0.3] }}
                    transition={{ duration: 4, repeat: Infinity, ease: "easeInOut" }}
                />
                <motion.div
                    className="absolute top-[60%] left-[20%] w-2 h-2 rounded-full bg-accent/30"
                    animate={{ y: [10, -15, 10], opacity: [0.2, 0.6, 0.2] }}
                    transition={{ duration: 5, repeat: Infinity, ease: "easeInOut", delay: 1 }}
                />
                <motion.div
                    className="absolute bottom-[25%] right-[30%] w-4 h-4 rounded-full bg-accent-deep/20"
                    animate={{ y: [-8, 12, -8], x: [-5, 5, -5], opacity: [0.2, 0.5, 0.2] }}
                    transition={{ duration: 6, repeat: Infinity, ease: "easeInOut", delay: 2 }}
                />
            </div>

            {/* Content */}
            <div className="relative z-10 flex flex-col justify-between p-12 xl:p-16 w-full">
                {/* Top — Logo */}
                <motion.div
                    initial={{ opacity: 0, y: -20 }}
                    animate={{ opacity: 1, y: 0 }}
                    transition={{ duration: 0.6 }}
                    className="flex items-center gap-4"
                >
                    <div className="bg-white/10 backdrop-blur-sm p-3 rounded-2xl border border-white/10">
                        <Package className="text-white" size={24} strokeWidth={2.5} />
                    </div>
                    <div>
                        <h2 className="text-white font-black text-2xl tracking-tighter uppercase">{PLATFORM_NAME}</h2>
                        <p className="text-white/40 text-[9px] font-black uppercase tracking-[0.4em]">{dict.auth?.premiumLogistics || 'Premium Logistics'}</p>
                    </div>
                </motion.div>

                {/* Middle — Value proposition + Features */}
                <div className="flex-1 flex flex-col justify-center py-12">
                    <motion.div
                        initial={{ opacity: 0, y: 20 }}
                        animate={{ opacity: 1, y: 0 }}
                        transition={{ duration: 0.6, delay: 0.2 }}
                    >
                        <h1 className="text-white font-black text-4xl xl:text-5xl tracking-tighter uppercase leading-[0.95] mb-4">
                            {dict.auth?.brandHeadline || 'Ship Smarter.'}<br />
                            <span className="text-accent">{dict.auth?.brandHighlight || 'Track Everything.'}</span>
                        </h1>
                        <p className="text-white/50 text-sm font-medium max-w-md leading-relaxed mb-12">
                            {dict.auth?.brandDesc || 'The all-in-one logistics platform trusted by thousands of businesses to automate shipment tracking, customer notifications, and delivery management.'}
                        </p>
                    </motion.div>

                    {/* Feature list */}
                    <div className="grid grid-cols-2 gap-4">
                        {features.map((feat, i) => (
                            <motion.div
                                key={feat.label}
                                initial={{ opacity: 0, x: -20 }}
                                animate={{ opacity: 1, x: 0 }}
                                transition={{ duration: 0.4, delay: 0.4 + i * 0.1 }}
                                className="flex items-start gap-3 p-4 rounded-2xl bg-white/[0.04] border border-white/[0.06] backdrop-blur-sm hover:bg-white/[0.08] transition-all duration-300"
                            >
                                <div className="bg-accent/20 p-2 rounded-xl shrink-0">
                                    <feat.icon size={16} className="text-accent" />
                                </div>
                                <div>
                                    <p className="text-white font-bold text-xs uppercase tracking-wide">{feat.label}</p>
                                    <p className="text-white/40 text-[11px] mt-0.5">{feat.desc}</p>
                                </div>
                            </motion.div>
                        ))}
                    </div>
                </div>

                {/* Bottom — Social proof stats */}
                <motion.div
                    initial={{ opacity: 0, y: 20 }}
                    animate={{ opacity: 1, y: 0 }}
                    transition={{ duration: 0.6, delay: 0.8 }}
                    className="flex items-center gap-8 pt-8 border-t border-white/[0.08]"
                >
                    {stats.map((stat, i) => (
                        <React.Fragment key={stat.label}>
                            <div className="text-center">
                                <p className="text-white font-black text-2xl tracking-tight">{stat.value}</p>
                                <p className="text-white/40 text-[10px] font-bold uppercase tracking-widest mt-1">{stat.label}</p>
                            </div>
                            {i < stats.length - 1 && (
                                <div className="w-px h-10 bg-white/10" />
                            )}
                        </React.Fragment>
                    ))}
                </motion.div>
            </div>
        </div>
    );
}
