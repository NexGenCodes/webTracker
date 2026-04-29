import { useState } from 'react';
import { AuthMode, RegisterStep } from '@/lib/validations/auth';

export function useAuthFlow() {
    const [mode, setMode] = useState<AuthMode>('signin');
    const [registerStep, setRegisterStep] = useState<RegisterStep>('credentials');
    const [emailCache, setEmailCache] = useState('');
    const [error, setError] = useState<string | null>(null);
    const [successMessage, setSuccessMessage] = useState<string | null>(null);

    const switchMode = (newMode: AuthMode) => {
        setMode(newMode);
        setError(null);
        setSuccessMessage(null);
    };

    return {
        mode,
        registerStep,
        setRegisterStep,
        emailCache,
        setEmailCache,
        error,
        setError,
        successMessage,
        setSuccessMessage,
        switchMode,
    };
}
