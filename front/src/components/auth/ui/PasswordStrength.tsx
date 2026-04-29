import React from 'react';
import { Shield, ShieldAlert, ShieldCheck } from 'lucide-react';

interface PasswordStrengthProps {
    password?: string;
}

export function PasswordStrength({ password = '' }: PasswordStrengthProps) {
    const calculateStrength = (pass: string) => {
        let score = 0;
        if (!pass) return { score: 0, label: 'None', color: 'bg-surface-muted', text: 'text-text-muted' };
        
        if (pass.length >= 8) score += 1;
        if (pass.match(/[A-Z]/)) score += 1;
        if (pass.match(/[0-9]/)) score += 1;
        if (pass.match(/[^A-Za-z0-9]/)) score += 1;
        
        if (pass.length >= 12) score += 1; // Bonus for length

        switch (score) {
            case 0:
            case 1:
            case 2:
                return { score: 25, label: 'Weak', color: 'bg-error', text: 'text-error', icon: ShieldAlert };
            case 3:
            case 4:
                return { score: 60, label: 'Good', color: 'bg-warning', text: 'text-warning', icon: Shield };
            case 5:
                return { score: 100, label: 'Strong', color: 'bg-success', text: 'text-success', icon: ShieldCheck };
            default:
                return { score: 0, label: 'None', color: 'bg-surface-muted', text: 'text-text-muted', icon: Shield };
        }
    };

    const strength = calculateStrength(password);
    const Icon = strength.icon || Shield;

    return (
        <div className="space-y-2 mt-2 w-full animate-fade-in">
            <div className="flex justify-between items-center px-1">
                <span className="text-[10px] font-bold uppercase tracking-widest text-text-muted flex items-center gap-1">
                    Password Strength
                </span>
                {password && (
                    <span className={`text-[10px] font-black uppercase tracking-widest flex items-center gap-1 ${strength.text}`}>
                        <Icon size={12} /> {strength.label}
                    </span>
                )}
            </div>
            <div className="h-1.5 w-full bg-surface-muted rounded-full overflow-hidden flex gap-1">
                <div 
                    className={`h-full ${strength.color} transition-all duration-500 ease-out`} 
                    style={{ width: password ? `${strength.score}%` : '0%' }}
                />
            </div>
        </div>
    );
}
