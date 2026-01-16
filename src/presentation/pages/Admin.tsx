import React, { useState } from 'react';
import type { CreateShipmentDto } from '../../domain/Shipment';
import { EmailParserService } from '../../application/EmailParserService';
import { ShipmentService } from '../../application/ShipmentService';
import { CONFIG } from '../../config';

interface AdminDashboardProps {
    service: ShipmentService;
    onLogout: () => void;
}

export const AdminDashboard: React.FC<AdminDashboardProps> = ({ service, onLogout }) => {
    const [emailText, setEmailText] = useState('');
    const [error, setError] = useState<string | null>(null);
    const [success, setSuccess] = useState<string | null>(null);

    const handleParse = async () => {
        setError(null);
        setSuccess(null);
        try {
            const dto: CreateShipmentDto = EmailParserService.parse(emailText);
            const shipment = await service.createShipment(dto);
            setSuccess(`Shipment Created! Tracking ID: ${shipment.trackingNumber}`);
            setEmailText(''); // Clear input on success
        } catch (err: any) {
            setError(err.message || 'Failed to parse email');
        }
    };

    return (
        <div className="min-h-screen p-4 flex flex-col items-center animate-fade-in text-main">
            <div className="w-full max-w-4xl">
                <div className="flex justify-between items-center mb-8">
                    <h1 className="text-3xl font-bold text-heading">Admin Dashboard</h1>
                    <button
                        onClick={onLogout}
                        className="text-muted hover:text-main text-sm"
                    >
                        Logout
                    </button>
                </div>

                <div className="glass-panel p-8 mb-8">
                    <h2 className="text-xl font-semibold mb-4 text-heading border-b border-panel-border pb-4">
                        Create New Shipment
                    </h2>

                    <div className="grid grid-cols-1 md:grid-cols-2 gap-8">
                        <div>
                            <p className="text-sm text-muted mb-4">
                                Paste the email content below. The system will automatically parse:
                                <br />Receiver, Address, Country, Phone, Sender.
                            </p>
                            <textarea
                                className="w-full h-64 bg-input-bg text-main p-4 rounded-xl border border-input-border outline-none focus:border-accent font-mono text-sm resize-none placeholder-muted"
                                placeholder={`Receiver: John Doe\nAddress: 123 Main St\nCountry: USA\n...`}
                                value={emailText}
                                onChange={e => setEmailText(e.target.value)}
                            />
                        </div>

                        <div className="flex flex-col justify-center">
                            <div className="bg-input-bg p-6 rounded-xl border border-input-border mb-6">
                                <p className="text-sm text-gray-400 mb-2">Instructions:</p>
                                <ul className="text-sm text-main space-y-2 list-disc pl-4">
                                    <li>Copy the full email body.</li>
                                    <li>Paste it into the box on the left.</li>
                                    <li>Click <strong>Generate Tracking ID</strong>.</li>
                                    <li>Send the ID back to the customer.</li>
                                </ul>
                            </div>

                            {error && (
                                <div className="mt-4 p-3 bg-red-500/10 border border-red-500/50 rounded-lg text-red-400 text-sm">
                                    Error: {error}
                                </div>
                            )}

                            {success && (
                                <div className="mt-4 p-4 bg-green-500/10 border border-green-500/50 rounded-lg">
                                    <p className="text-green-400 font-bold mb-2">Success!</p>
                                    <p className="text-heading text-lg">Please reply to the Admin with this ID:</p>
                                    <div className="bg-black/50 p-2 rounded mt-1 font-mono text-xl select-all cursor-copy text-white">
                                        {success.split(': ')[1]}
                                    </div>
                                </div>
                            )}

                            <div className="mt-6 flex justify-end">
                                <button
                                    onClick={handleParse}
                                    className="btn-primary"
                                    disabled={!emailText}
                                >
                                    Generate Tracking ID
                                </button>
                            </div>
                        </div>
                    </div>
                </div>

                <div className="text-center text-muted text-xs">
                    {CONFIG.APP_NAME} Admin System v1.0
                </div>
            </div>
        </div>
    );
};
