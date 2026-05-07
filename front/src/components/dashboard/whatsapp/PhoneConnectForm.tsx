import React from 'react';
import { z } from 'zod';
import { useForm } from 'react-hook-form';
import { zodResolver } from '@hookform/resolvers/zod';
import { parsePhoneNumberFromString } from 'libphonenumber-js';
import { Loader2, ChevronRight } from 'lucide-react';

export const COUNTRY_CODES = [
    { value: '+234', label: '🇳🇬 +234' },
    { value: '+27', label: '🇿🇦 +27' },
    { value: '+254', label: '🇰🇪 +254' },
    { value: '+233', label: '🇬🇭 +233' },
    { value: '+1', label: '🇺🇸 +1' },
    { value: '+44', label: '🇬🇧 +44' },
];

const formatLocalPhone = (countryCode: string, phone: string) => {
    let localPhone = phone.replace(/[\s\-()]/g, '');
    if (localPhone.startsWith('0')) {
        localPhone = localPhone.substring(1);
    }
    return localPhone.startsWith('+') ? localPhone : `${countryCode}${localPhone}`;
};

export const phoneSchema = z.object({
    countryCode: z.string(),
    phone: z.string().min(1, "Phone number is required")
}).superRefine((data, ctx) => {
    const fullNumber = formatLocalPhone(data.countryCode, data.phone);
    const phoneNumber = parsePhoneNumberFromString(fullNumber);

    if (!phoneNumber || !phoneNumber.isValid()) {
        ctx.addIssue({
            code: z.ZodIssueCode.custom,
            message: "Please enter a valid phone number",
            path: ["phone"]
        });
    }
});

export type PhoneFormValues = z.infer<typeof phoneSchema>;

interface PhoneConnectFormProps {
    isPending: boolean;
    onSubmit: (data: PhoneFormValues) => void;
}

export function PhoneConnectForm({ isPending, onSubmit }: PhoneConnectFormProps) {
    const { register, handleSubmit, watch, formState: { errors } } = useForm<PhoneFormValues>({
        resolver: zodResolver(phoneSchema),
        defaultValues: {
            countryCode: '+234',
            phone: '',
        }
    });

    const watchPhone = watch('phone');

    return (
        <form onSubmit={handleSubmit(onSubmit)} className="space-y-6">
            <div className="space-y-2">
                <label className="text-xs font-black uppercase tracking-widest text-text-muted">Business Number</label>
                <div className="flex items-stretch gap-0 rounded-2xl border-2 border-transparent focus-within:border-accent/20 focus-within:ring-4 focus-within:ring-accent-soft overflow-hidden transition-all bg-surface-muted">
                    <select
                        {...register('countryCode')}
                        className="shrink-0 appearance-none border-r border-border px-4 py-4 text-sm font-black cursor-pointer focus:outline-none bg-surface text-text-main w-[110px]"
                    >
                        {COUNTRY_CODES.map(c => (
                            <option key={c.value} value={c.value}>{c.label}</option>
                        ))}
                    </select>
                    <input
                        type="tel"
                        {...register('phone')}
                        placeholder="803 000 0000"
                        className="flex-1 min-w-0 px-4 py-4 text-base font-medium outline-none bg-transparent text-text-main caret-text-main"
                    />
                </div>
                {errors.phone && (
                    <p className="text-xs font-bold text-error mt-1">{errors.phone.message}</p>
                )}
                <p className="text-[10px] font-medium text-text-muted pl-1">Exclude the leading zero if applicable (e.g. 803 instead of 0803)</p>
            </div>

            <button
                type="submit"
                disabled={isPending || !watchPhone}
                className="btn-primary w-full py-4 text-sm flex items-center justify-center gap-2 group disabled:opacity-50 disabled:cursor-not-allowed"
            >
                {isPending ? (
                    <>
                        <Loader2 size={18} className="animate-spin" />
                        <span>Processing...</span>
                    </>
                ) : (
                    <>
                        <span>Generate Code</span>
                        <ChevronRight size={18} className="group-hover:translate-x-1 transition-transform" />
                    </>
                )}
            </button>
        </form>
    );
}
