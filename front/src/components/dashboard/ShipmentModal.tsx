'use client';

import React, { useEffect } from 'react';
import { X, Save, Loader2, Package, User, Box } from 'lucide-react';
import { motion } from 'framer-motion';
import { useForm, SubmitHandler, Resolver } from 'react-hook-form';
import { zodResolver } from '@hookform/resolvers/zod';
import * as z from 'zod';
import { PremiumInput, PremiumTextarea } from '@/components/shared/PremiumInput';
import { createShipmentAction, updateShipmentAction } from '@/app/actions/shipment';
import toast from 'react-hot-toast';

const shipmentSchema = z.object({
    recipient_name: z.string().min(2, 'Name is required'),
    recipient_phone: z.string().min(5, 'Valid phone is required'),
    recipient_address: z.string().min(5, 'Address is required'),
    destination: z.string().min(2, 'Destination country is required'),
    recipient_email: z.string().email().optional().or(z.literal('')),
    sender_name: z.string().min(2, 'Sender name is required'),
    origin: z.string().min(2, 'Origin country is required'),
    weight: z.preprocess((val) => Number(val), z.number().min(0.1, 'Weight must be at least 0.1')),
    cargo_type: z.string().optional(),
    status: z.string().optional(),
});

export type ShipmentFormValues = z.infer<typeof shipmentSchema>;

interface ShipmentModalProps {
    isOpen: boolean;
    onClose: () => void;
    companyId: string;
    shipment?: (ShipmentFormValues & { id: string; tracking_id?: string }) | null; // If present, we are in EDIT mode
    onSuccess: () => void;
}

export function ShipmentModal({ isOpen, onClose, companyId, shipment, onSuccess }: ShipmentModalProps) {
    const isEdit = !!shipment;

    const { register, handleSubmit, reset, formState: { errors, isSubmitting } } = useForm<ShipmentFormValues>({
        resolver: zodResolver(shipmentSchema) as unknown as Resolver<ShipmentFormValues>,
        defaultValues: {
            recipient_name: '',
            recipient_phone: '',
            recipient_address: '',
            destination: '',
            recipient_email: '',
            sender_name: '',
            origin: '',
            weight: 0.5,
            cargo_type: 'General Cargo',
            status: 'pending',
        }
    });

    useEffect(() => {
        if (shipment) {
            reset({
                recipient_name: shipment.recipient_name,
                recipient_phone: shipment.recipient_phone,
                recipient_address: shipment.recipient_address,
                destination: shipment.destination,
                recipient_email: shipment.recipient_email || '',
                sender_name: shipment.sender_name,
                origin: shipment.origin,
                weight: shipment.weight,
                cargo_type: shipment.cargo_type || 'General Cargo',
                status: shipment.status,
            });
        } else {
            reset({
                recipient_name: '',
                recipient_phone: '',
                recipient_address: '',
                destination: '',
                recipient_email: '',
                sender_name: '',
                origin: '',
                weight: 0.5,
                cargo_type: 'General Cargo',
                status: 'pending',
            });
        }
    }, [shipment, reset, isOpen]);

    const onSubmit: SubmitHandler<ShipmentFormValues> = async (values) => {
        try {
            let result;
            if (isEdit && shipment) {
                result = await updateShipmentAction(shipment.id, companyId, values as Record<string, unknown>);
            } else {
                result = await createShipmentAction(companyId, values as Record<string, unknown>);
            }

            if (result.success) {
                toast.success(isEdit ? 'Shipment updated successfully!' : 'Shipment created successfully!');
                onSuccess();
                onClose();
            } else {
                toast.error(result.error || 'Something went wrong.');
            }
        } catch {
            toast.error('A network error occurred.');
        }
    };

    const onFormSubmit = handleSubmit(onSubmit);

    if (!isOpen) return null;

    return (
        <div className="fixed inset-0 z-[100] flex items-center justify-center p-4">
            <motion.div
                initial={{ opacity: 0 }}
                animate={{ opacity: 1 }}
                exit={{ opacity: 0 }}
                onClick={onClose}
                className="absolute inset-0 bg-black/60 backdrop-blur-sm"
            />

            <motion.div
                initial={{ opacity: 0, scale: 0.9, y: 20 }}
                animate={{ opacity: 1, scale: 1, y: 0 }}
                exit={{ opacity: 0, scale: 0.9, y: 20 }}
                className="relative w-full max-w-2xl bg-background border border-border/50 rounded-[2.5rem] shadow-2xl overflow-hidden flex flex-col max-h-[90vh]"
            >
                {/* Header */}
                <div className="px-8 py-6 border-b border-border/50 flex items-center justify-between bg-surface/30">
                    <div className="flex items-center gap-4">
                        <div className="w-12 h-12 rounded-2xl bg-accent/10 flex items-center justify-center text-accent">
                            <Package size={24} />
                        </div>
                        <div>
                            <h2 className="text-xl font-black text-text-main uppercase tracking-tighter">
                                {isEdit ? 'Edit Shipment' : 'New Shipment'}
                            </h2>
                            <p className="text-xs font-bold text-text-muted uppercase tracking-widest">
                                {isEdit ? `Tracking: ${shipment.tracking_id}` : 'Fill in the details below'}
                            </p>
                        </div>
                    </div>
                    <button
                        onClick={onClose}
                        className="p-3 hover:bg-surface rounded-2xl text-text-muted hover:text-text-main transition-all"
                    >
                        <X size={20} />
                    </button>
                </div>

                {/* Form */}
                <form onSubmit={handleSubmit(onSubmit)} className="overflow-y-auto p-8 space-y-8 custom-scrollbar">
                    {/* Receiver Information */}
                    <section className="space-y-6">
                        <div className="flex items-center gap-2 text-accent">
                            <User size={16} />
                            <h3 className="text-xs font-black uppercase tracking-[0.2em]">Receiver Information</h3>
                        </div>
                        <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
                            <div className="space-y-1">
                                <PremiumInput
                                    label="Receiver Name"
                                    placeholder="e.g. John Doe"
                                    {...register('recipient_name')}
                                />
                                {errors.recipient_name && <p className="text-[10px] font-bold text-error uppercase ml-2">{errors.recipient_name.message}</p>}
                            </div>
                            <div className="space-y-1">
                                <PremiumInput
                                    label="Phone Number"
                                    placeholder="+351 ..."
                                    {...register('recipient_phone')}
                                />
                                {errors.recipient_phone && <p className="text-[10px] font-bold text-error uppercase ml-2">{errors.recipient_phone.message}</p>}
                            </div>
                        </div>
                        <div className="space-y-1">
                            <PremiumTextarea
                                label="Delivery Address"
                                placeholder="Full street address, city, postal code..."
                                rows={2}
                                {...register('recipient_address')}
                            />
                            {errors.recipient_address && <p className="text-[10px] font-bold text-error uppercase ml-2">{errors.recipient_address.message}</p>}
                        </div>
                        <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
                            <div className="space-y-1">
                                <PremiumInput
                                    label="Destination Country"
                                    placeholder="e.g. Portugal"
                                    {...register('destination')}
                                />
                                {errors.destination && <p className="text-[10px] font-bold text-error uppercase ml-2">{errors.destination.message}</p>}
                            </div>
                            <PremiumInput
                                label="Email (Optional)"
                                placeholder="john@example.com"
                                type="email"
                                {...register('recipient_email')}
                            />
                        </div>
                    </section>

                    <div className="h-px bg-border/30 w-full" />

                    {/* Shipment Details */}
                    <section className="space-y-6">
                        <div className="flex items-center gap-2 text-primary">
                            <Box size={16} />
                            <h3 className="text-xs font-black uppercase tracking-[0.2em]">Shipment Details</h3>
                        </div>
                        <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
                            <div className="space-y-1">
                                <PremiumInput
                                    label="Sender Name"
                                    placeholder="Origin company or person"
                                    {...register('sender_name')}
                                />
                                {errors.sender_name && <p className="text-[10px] font-bold text-error uppercase ml-2">{errors.sender_name.message}</p>}
                            </div>
                            <div className="space-y-1">
                                <PremiumInput
                                    label="Origin Country"
                                    placeholder="e.g. Nigeria"
                                    {...register('origin')}
                                />
                                {errors.origin && <p className="text-[10px] font-bold text-error uppercase ml-2">{errors.origin.message}</p>}
                            </div>
                        </div>
                        <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
                            <div className="space-y-1">
                                <PremiumInput
                                    label="Weight (kg)"
                                    type="number"
                                    step="0.1"
                                    {...register('weight')}
                                />
                                {errors.weight && <p className="text-[10px] font-bold text-error uppercase ml-2">{errors.weight.message}</p>}
                            </div>
                            <PremiumInput
                                label="Cargo Type"
                                placeholder="e.g. Electronics, Clothing"
                                {...register('cargo_type')}
                            />
                        </div>

                        {isEdit && (
                            <div className="space-y-3">
                                <label className="text-xs font-black uppercase tracking-[0.2em] text-text-muted ml-1 opacity-60">
                                    Current Status
                                </label>
                                <select
                                    {...register('status')}
                                    className="w-full bg-surface-muted text-text-main py-4 px-6 rounded-2xl border-2 border-transparent outline-none focus:border-accent/30 focus:bg-surface focus:ring-4 focus:ring-accent/10 transition-all font-medium shadow-inner"
                                >
                                    <option value="pending">Pending</option>
                                    <option value="intransit">In Transit</option>
                                    <option value="outfordelivery">Out for Delivery</option>
                                    <option value="delivered">Delivered</option>
                                    <option value="canceled">Canceled</option>
                                </select>
                            </div>
                        )}
                    </section>
                </form>

                {/* Footer */}
                <div className="px-8 py-6 border-t border-border/50 bg-surface/30 flex items-center justify-end gap-4">
                    <button
                        type="button"
                        onClick={onClose}
                        className="px-6 py-4 rounded-2xl font-black text-xs uppercase tracking-widest text-text-muted hover:text-text-main transition-colors"
                    >
                        Cancel
                    </button>
                    <button
                        onClick={onFormSubmit}
                        disabled={isSubmitting}
                        className="btn-primary px-10 py-4 text-xs flex items-center gap-2 active:scale-95 shadow-xl shadow-accent/20"
                    >
                        {isSubmitting ? (
                            <Loader2 size={16} className="animate-spin" />
                        ) : (
                            <Save size={16} />
                        )}
                        {isEdit ? 'Update Shipment' : 'Create Shipment'}
                    </button>
                </div>
            </motion.div>
        </div>
    );
}
