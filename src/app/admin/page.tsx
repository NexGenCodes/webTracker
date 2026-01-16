"use client";

import { useState } from 'react';
import { Copy, ChevronLeft, Check } from 'lucide-react';
import { createShipment } from '../actions/shipment'; // We will create this server action later
import { parseEmail } from '@/lib/email-parser'; // We will port the parser or keep it inline

export default function AdminPage() {
    const [emailText, setEmailText] = useState('');
    const [trackingId, setTrackingId] = useState<string | null>(null);
    const [error, setError] = useState<string | null>(null);
    const [loading, setLoading] = useState(false);
    const [copied, setCopied] = useState(false);

    const handleGenerate = async () => {
        setError(null);
        setLoading(true);
        try {
            // Logic to parse email - for now assuming simple client-side parse or passing text to server
            // Let's pass text to server action to keep client light? 
            // Or parse here. The user said "paste email". 
            // I'll assume we keep logic consistent: Parse -> Send DTO to server.
            const dto = parseEmail(emailText);
            const result = await createShipment(dto);
            if (result.success) {
                setTrackingId(result.trackingNumber ?? null);
            } else {
                setError(result.error ?? 'Failed to create shipment');
            }
        } catch (err: any) {
            setError(err.message || 'Invalid email format');
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
        setEmailText('');
        setCopied(false);
    };

    if (trackingId) {
        // Success View
        return (
            <div className="flex flex-col items-center justify-center min-h-[60vh] p-4 text-center animate-fade-in space-y-8">
                <div className="space-y-2">
                    <h2 className="text-3xl font-bold text-green-500">Shipment Created!</h2>
                    <p className="text-gray-400">Please send this Tracking ID to the customer.</p>
                </div>

                <div className="w-full max-w-md bg-gray-900 border border-gray-700 rounded-xl p-6 flex flex-col items-center gap-4">
                    <span className="text-sm text-gray-500 uppercase tracking-wider">Tracking ID</span>
                    <span className="text-4xl font-mono text-white font-bold tracking-widest break-all">
                        {trackingId}
                    </span>
                </div>

                <div className="flex flex-col w-full max-w-sm gap-4">
                    <button
                        onClick={handleCopy}
                        className="btn-primary flex items-center justify-center gap-2 py-4 text-lg w-full"
                    >
                        {copied ? <Check /> : <Copy />}
                        {copied ? 'Copied!' : 'Copy to Clipboard'}
                    </button>
                    <button
                        onClick={handleBack}
                        className="flex items-center justify-center gap-2 text-gray-400 hover:text-white py-2"
                    >
                        <ChevronLeft size={20} />
                        Create Another
                    </button>
                </div>
            </div>
        );
    }

    // Input View
    return (
        <div className="max-w-xl mx-auto p-4 flex flex-col gap-6">
            <h1 className="text-2xl font-bold text-white mb-2">Create New Shipment</h1>

            <div className="space-y-4">
                <div className="bg-gray-800/50 p-4 rounded-lg border border-gray-700 text-sm text-gray-300">
                    <p className="font-semibold mb-2 text-white">Instructions:</p>
                    <ul className="list-disc pl-4 space-y-1">
                        <li>Copy the full email body.</li>
                        <li>Paste it below.</li>
                        <li>Click Generate.</li>
                    </ul>
                </div>

                <textarea
                    className="w-full h-64 bg-gray-900 text-white p-4 rounded-xl border border-gray-700 focus:border-blue-500 outline-none resize-none font-mono text-sm"
                    placeholder="Paste email content here..."
                    value={emailText}
                    onChange={(e) => setEmailText(e.target.value)}
                />

                {error && (
                    <div className="p-3 bg-red-500/20 border border-red-500/50 rounded-lg text-red-300 text-sm">
                        {error}
                    </div>
                )}

                <button
                    onClick={handleGenerate}
                    disabled={!emailText || loading}
                    className="btn-primary w-full py-4 text-lg font-bold disabled:opacity-50 disabled:cursor-not-allowed"
                >
                    {loading ? 'Processing...' : 'Generate Tracking ID'}
                </button>
            </div>
        </div>
    );
}
