import React, { useState } from 'react';
import { UseFormRegisterReturn } from 'react-hook-form';
import { LucideIcon, Eye, EyeOff } from 'lucide-react';

interface AuthInputProps extends React.InputHTMLAttributes<HTMLInputElement> {
    label: string;
    icon: LucideIcon;
    registration: UseFormRegisterReturn;
    error?: string;
    actionLabel?: string;
    onActionClick?: () => void;
}

export function AuthInput({ 
    label, 
    icon: Icon, 
    registration, 
    error, 
    actionLabel, 
    onActionClick, 
    type = 'text',
    ...props 
}: AuthInputProps) {
    const [showPassword, setShowPassword] = useState(false);
    const isPassword = type === 'password';
    const currentType = isPassword ? (showPassword ? 'text' : 'password') : type;

    return (
        <div className="space-y-2 w-full">
            <div className="flex justify-between items-center ml-1">
                <label className="text-[10px] font-black uppercase tracking-[0.2em] text-accent/80">
                    {label}
                </label>
                {actionLabel && onActionClick && (
                    <button 
                        type="button" 
                        onClick={onActionClick}
                        className="text-[10px] font-bold text-text-muted hover:text-accent uppercase tracking-widest transition-colors"
                    >
                        {actionLabel}
                    </button>
                )}
            </div>
            <div className="relative group w-full">
                <Icon className="absolute left-5 top-1/2 -translate-y-1/2 text-text-muted group-focus-within:text-accent transition-colors" size={20} />
                <input
                    {...registration}
                    {...props}
                    type={currentType}
                    className={`input-premium pl-12! ${isPassword ? 'pr-12!' : ''} w-full ${error ? 'border-error' : ''}`}
                />
                {isPassword && (
                    <button
                        type="button"
                        onClick={() => setShowPassword(!showPassword)}
                        className="absolute right-4 top-1/2 -translate-y-1/2 p-1 text-text-muted hover:text-accent transition-colors rounded-lg hover:bg-accent/10"
                        tabIndex={-1}
                    >
                        {showPassword ? <EyeOff size={18} /> : <Eye size={18} />}
                    </button>
                )}
            </div>
            {error && (
                <p className="text-error text-xs font-bold ml-1 animate-fade-in">{error}</p>
            )}
        </div>
    );
}
