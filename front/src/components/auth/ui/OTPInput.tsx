import React, { useRef } from 'react';

interface OTPInputProps {
    otp: string[];
    setOtp: (otp: string[]) => void;
}

export function OTPInput({ otp, setOtp }: OTPInputProps) {
    const otpRefs = useRef<(HTMLInputElement | null)[]>([]);

    const handleOTPChange = (index: number, value: string) => {
        if (!/^\d*$/.test(value)) return;
        const next = [...otp];
        next[index] = value.slice(-1);
        setOtp(next);

        if (value && index < 5) {
            otpRefs.current[index + 1]?.focus();
        }
    };

    const handleOTPKeyDown = (index: number, e: React.KeyboardEvent) => {
        if (e.key === 'Backspace' && !otp[index] && index > 0) {
            otpRefs.current[index - 1]?.focus();
        }
    };

    const handleOTPPaste = (e: React.ClipboardEvent) => {
        e.preventDefault();
        const pasted = e.clipboardData.getData('text').replace(/\D/g, '').slice(0, 6);
        const next = [...otp];
        for (let i = 0; i < pasted.length; i++) {
            next[i] = pasted[i];
        }
        setOtp(next);
        const focusIdx = Math.min(pasted.length, 5);
        otpRefs.current[focusIdx]?.focus();
    };

    return (
        <div className="flex justify-center gap-3 w-full" onPaste={handleOTPPaste}>
            {otp.map((digit, i) => (
                <input
                    key={i}
                    ref={(el) => { otpRefs.current[i] = el; }}
                    type="text"
                    inputMode="numeric"
                    maxLength={1}
                    value={digit}
                    onChange={(e) => handleOTPChange(i, e.target.value)}
                    onKeyDown={(e) => handleOTPKeyDown(i, e)}
                    className="w-12 h-14 text-center text-2xl font-black bg-surface border-2 border-border rounded-xl focus:border-accent focus:ring-2 focus:ring-accent/20 outline-none transition-all duration-200 text-text-main"
                    autoFocus={i === 0}
                />
            ))}
        </div>
    );
}
