import { motion, AnimatePresence } from 'framer-motion';
import { X } from 'lucide-react';
import { useEffect } from 'react';

interface MobileNavOverlayProps {
    isOpen: boolean;
    onClose: () => void;
    children: React.ReactNode;
    zIndex?: string;
}

export function MobileNavOverlay({ isOpen, onClose, children, zIndex = "z-[999]" }: MobileNavOverlayProps) {
    useEffect(() => {
        if (isOpen) {
            document.body.style.overflow = 'hidden';
        } else {
            document.body.style.overflow = '';
        }
        return () => {
            document.body.style.overflow = '';
        };
    }, [isOpen]);

    return (
        <AnimatePresence mode="wait">
            {isOpen && (
                <motion.div
                    initial={{ opacity: 0 }}
                    animate={{ opacity: 1 }}
                    exit={{ opacity: 0 }}
                    transition={{ duration: 0.2 }}
                    className={`fixed inset-0 ${zIndex} bg-surface/95 backdrop-blur-xl flex flex-col pt-24 px-6 pb-10 overflow-y-auto`}
                >
                    <button
                        onClick={onClose}
                        className="absolute top-6 right-6 p-4 text-text-muted hover:text-accent transition-colors bg-surface-muted rounded-2xl"
                        aria-label="Close menu"
                    >
                        <X size={32} strokeWidth={3} />
                    </button>
                    {children}
                </motion.div>
            )}
        </AnimatePresence>
    );
}
