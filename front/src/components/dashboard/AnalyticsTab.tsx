'use client';

import React from 'react';
import { 
    XAxis, YAxis, CartesianGrid, Tooltip, ResponsiveContainer, 
    PieChart, Pie, Cell, AreaChart, Area
} from 'recharts';
import { TrendingUp, Package, Calendar, ArrowUpRight } from 'lucide-react';
import { motion } from 'framer-motion';

interface AnalyticsTabProps {
    shipmentStats: { total: number; active: number; delivered: number } | undefined;
}

const COLORS = ['#6366f1', '#10b981', '#f59e0b', '#ef4444'];

const mockMonthlyData = [
    { name: 'Jan', count: 40 },
    { name: 'Feb', count: 30 },
    { name: 'Mar', count: 65 },
    { name: 'Apr', count: 45 },
    { name: 'May', count: 90 },
    { name: 'Jun', count: 75 },
];

const mockStatusData = [
    { name: 'Delivered', value: 400 },
    { name: 'In Transit', value: 300 },
    { name: 'Pending', value: 100 },
    { name: 'Canceled', value: 50 },
];

export function AnalyticsTab({ shipmentStats }: AnalyticsTabProps) {
    const realStatusData = [
        { name: 'Delivered', value: shipmentStats?.delivered || 0 },
        { name: 'Active', value: shipmentStats?.active || 0 },
        { name: 'Total', value: shipmentStats?.total || 0 },
    ].filter(d => d.value > 0);

    // Fallback if no data yet
    const displayData = realStatusData.length > 0 ? realStatusData : mockStatusData;

    const stats = [
        { 
            label: 'Total Shipments', 
            value: shipmentStats?.total.toString() || '0', 
            icon: Package, 
            color: 'text-accent', 
            trend: 'up',
            sub: 'All time volume'
        },
        { 
            label: 'Success Rate', 
            value: shipmentStats?.total ? `${Math.round((shipmentStats.delivered / shipmentStats.total) * 100)}%` : '0%', 
            icon: TrendingUp, 
            color: 'text-success', 
            trend: 'up',
            sub: 'Delivery completion'
        },
        { 
            label: 'Avg. Transit', 
            value: '4.2 Days', 
            icon: Calendar, 
            color: 'text-primary', 
            trend: 'neutral',
            sub: 'Estimated performance'
        },
    ];

    return (
        <div className="space-y-8 pb-10">
            {/* Quick Stats Grid */}
            <div className="grid grid-cols-1 sm:grid-cols-3 gap-6">
                {stats.map((stat, i) => (
                    <motion.div 
                        key={stat.label}
                        initial={{ opacity: 0, y: 20 }}
                        animate={{ opacity: 1, y: 0 }}
                        transition={{ delay: i * 0.1 }}
                        className="glass-panel p-6 border-border/40 hover:border-accent/30 transition-all group relative overflow-hidden"
                    >
                        <div className="relative z-10">
                            <div className="flex items-center justify-between mb-4">
                                <div className="w-10 h-10 rounded-xl bg-surface border border-border flex items-center justify-center text-text-muted group-hover:text-accent transition-colors shadow-inner">
                                    <stat.icon size={18} />
                                </div>
                                <div className="px-2 py-1 rounded-lg bg-surface/50 border border-border/50 flex items-center gap-1">
                                    {stat.trend === 'up' && <ArrowUpRight size={10} className="text-success" />}
                                    <span className="text-[9px] font-black text-text-muted uppercase tracking-tighter">{stat.sub}</span>
                                </div>
                            </div>
                            <p className="text-[10px] font-black uppercase tracking-[0.2em] text-text-muted mb-1">{stat.label}</p>
                            <p className={`text-3xl font-black tracking-tighter ${stat.color}`}>{stat.value}</p>
                        </div>
                        {/* Subtle background glow */}
                        <div className="absolute -bottom-4 -right-4 w-24 h-24 bg-accent/5 rounded-full blur-3xl group-hover:bg-accent/10 transition-colors" />
                    </motion.div>
                ))}
            </div>

            {/* Charts Grid */}
            <div className="grid grid-cols-1 lg:grid-cols-2 gap-8">
                {/* Volume Trend */}
                <div className="glass-panel p-8 border-border/40">
                    <div className="flex items-center justify-between mb-8">
                        <div>
                            <h3 className="text-sm font-black text-text-main uppercase tracking-widest">Shipment Volume</h3>
                            <p className="text-xs font-bold text-text-muted mt-1">Monthly performance overview</p>
                        </div>
                        <div className="px-3 py-1 rounded-lg bg-surface border border-border text-[10px] font-black uppercase tracking-widest text-text-muted">
                            Last 6 Months
                        </div>
                    </div>
                    <div className="h-[300px] w-full">
                        <ResponsiveContainer width="100%" height="100%">
                            <AreaChart data={mockMonthlyData}>
                                <defs>
                                    <linearGradient id="colorCount" x1="0" y1="0" x2="0" y2="1">
                                        <stop offset="5%" stopColor="#6366f1" stopOpacity={0.3}/>
                                        <stop offset="95%" stopColor="#6366f1" stopOpacity={0}/>
                                    </linearGradient>
                                </defs>
                                <CartesianGrid strokeDasharray="3 3" vertical={false} stroke="rgba(255,255,255,0.05)" />
                                <XAxis 
                                    dataKey="name" 
                                    axisLine={false} 
                                    tickLine={false} 
                                    tick={{ fontSize: 10, fontWeight: 700, fill: 'rgba(255,255,255,0.3)' }}
                                    dy={10}
                                />
                                <YAxis 
                                    axisLine={false} 
                                    tickLine={false} 
                                    tick={{ fontSize: 10, fontWeight: 700, fill: 'rgba(255,255,255,0.3)' }}
                                />
                                <Tooltip 
                                    contentStyle={{ 
                                        backgroundColor: '#111', 
                                        border: '1px solid rgba(255,255,255,0.1)',
                                        borderRadius: '12px',
                                        fontSize: '12px',
                                        boxShadow: '0 20px 25px -5px rgba(0, 0, 0, 0.5)'
                                    }}
                                    itemStyle={{ color: '#fff', fontWeight: 800 }}
                                />
                                <Area type="monotone" dataKey="count" stroke="#6366f1" strokeWidth={4} fillOpacity={1} fill="url(#colorCount)" />
                            </AreaChart>
                        </ResponsiveContainer>
                    </div>
                </div>

                {/* Status Distribution */}
                <div className="glass-panel p-8 border-border/40">
                    <div className="flex items-center justify-between mb-8">
                        <div>
                            <h3 className="text-sm font-black text-text-main uppercase tracking-widest">Status Distribution</h3>
                            <p className="text-xs font-bold text-text-muted mt-1">Current shipment breakdown</p>
                        </div>
                    </div>
                    <div className="h-[300px] w-full flex flex-col sm:flex-row items-center justify-center">
                        <div className="flex-1 w-full h-full">
                            <ResponsiveContainer width="100%" height="100%">
                                <PieChart>
                                    <Pie
                                        data={displayData}
                                        cx="50%"
                                        cy="50%"
                                        innerRadius={65}
                                        outerRadius={95}
                                        paddingAngle={10}
                                        dataKey="value"
                                        stroke="none"
                                    >
                                        {displayData.map((entry, index) => (
                                            <Cell key={`cell-${index}`} fill={COLORS[index % COLORS.length]} />
                                        ))}
                                    </Pie>
                                    <Tooltip 
                                        contentStyle={{ 
                                            backgroundColor: '#111', 
                                            border: '1px solid rgba(255,255,255,0.1)',
                                            borderRadius: '12px',
                                            boxShadow: '0 20px 25px -5px rgba(0, 0, 0, 0.5)'
                                        }}
                                    />
                                </PieChart>
                            </ResponsiveContainer>
                        </div>
                        <div className="flex flex-col gap-4 min-w-[140px] px-6 py-4 bg-surface/30 rounded-2xl border border-border/30">
                            {displayData.map((item, i) => (
                                <div key={item.name} className="flex items-center justify-between gap-3">
                                    <div className="flex items-center gap-2">
                                        <div className="w-2 h-2 rounded-full shadow-[0_0_8px_rgba(var(--color-accent),0.5)]" style={{ backgroundColor: COLORS[i % COLORS.length] }} />
                                        <span className="text-[10px] font-black uppercase tracking-widest text-text-muted">{item.name}</span>
                                    </div>
                                    <span className="text-xs font-black text-text-main">{item.value}</span>
                                </div>
                            ))}
                        </div>
                    </div>
                </div>
            </div>
        </div>
    );
}
