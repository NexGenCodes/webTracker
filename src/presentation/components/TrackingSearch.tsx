import React, { useState } from 'react';
import { Search } from 'lucide-react';

interface TrackingSearchProps {
    onSearch: (trackingNumber: string) => void;
    isLoading: boolean;
}

export const TrackingSearch: React.FC<TrackingSearchProps> = ({ onSearch, isLoading }) => {
    const [input, setInput] = useState('');

    const handleSubmit = (e: React.FormEvent) => {
        e.preventDefault();
        if (input.trim()) {
            onSearch(input.trim());
        }
    };

    return (
        <div className="w-full max-w-xl mx-auto p-8 glass-panel animate-fade-in">
            <h1 className="text-4xl font-bold mb-2 text-center text-gradient">Track Your Shipment</h1>
            <p className="text-center text-muted mb-8">Enter your tracking number to see real-time progress</p>

            <form onSubmit={handleSubmit} className="relative">
                <input
                    type="text"
                    value={input}
                    onChange={(e) => setInput(e.target.value)}
                    placeholder="e.g. AWB-X7Y8Z9"
                    className="w-full bg-input-bg text-main py-4 px-6 rounded-xl border border-input-border outline-none focus:border-accent focus:ring-1 focus:ring-accent transition-all placeholder-muted"
                />
                <button
                    type="submit"
                    disabled={isLoading}
                    className="absolute right-2 top-2 bottom-2 bg-accent text-primary font-bold px-6 rounded-lg hover:bg-sky-300 transition-colors disabled:opacity-50"
                >
                    {isLoading ? '...' : <Search size={20} />}
                </button>
            </form>
        </div>
    );
};
