"use client";

import React, { useState } from 'react';
import { Search, Loader2 } from 'lucide-react';

interface TrackingSearchProps {
    onSearch: (trackingNumber: string) => Promise<void>;
    isLoading: boolean;
}

export const TrackingSearch: React.FC<TrackingSearchProps> = ({ onSearch, isLoading }) => {
    const [input, setInput] = useState('');

    const handleSubmit = async (e: React.FormEvent) => {
        e.preventDefault();
        if (input.trim()) {
            await onSearch(input.trim());
        }
    };

    return (
        <div className="w-full max-w-xl mx-auto p-6 md:p-8 bg-black/20 backdrop-blur-md rounded-2xl border border-white/10 shadow-xl">
            <h1 className="text-3xl md:text-4xl font-bold mb-3 text-center bg-linear-to-r from-blue-400 to-purple-400 bg-clip-text text-transparent">
                Track Your Shipment
            </h1>
            <p className="text-center text-gray-400 mb-8 text-sm md:text-base">
                Enter your tracking number to see real-time progress
            </p>

            <form onSubmit={handleSubmit} className="relative group">
                <input
                    type="text"
                    value={input}
                    onChange={(e) => setInput(e.target.value)}
                    placeholder="e.g. TRK-..."
                    className="w-full bg-gray-900/50 text-white py-4 pl-6 pr-14 rounded-xl border border-gray-700 outline-none focus:border-blue-500 focus:ring-1 focus:ring-blue-500 transition-all placeholder-gray-600 font-mono text-lg"
                    disabled={isLoading}
                />
                <button
                    type="submit"
                    disabled={isLoading || !input.trim()}
                    className="absolute right-2 top-2 bottom-2 bg-blue-500 hover:bg-blue-400 text-white p-2 md:px-4 rounded-lg transition-all disabled:opacity-50 disabled:hover:bg-blue-500"
                >
                    {isLoading ? <Loader2 className="animate-spin" size={20} /> : <Search size={20} />}
                </button>
            </form>
        </div>
    );
};
