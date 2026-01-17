"use client";

import { useI18n } from "@/components/I18nContext";
import { ThemeToggle } from "@/components/ThemeToggle";
import { LanguageToggle } from "@/components/LanguageToggle";
import { Package, ArrowLeft, FileText } from "lucide-react";
import Link from "next/link";
import { APP_NAME } from "@/lib/constants";

export default function TermsPage() {
    const { dict } = useI18n();

    return (
        <main className="min-h-screen p-4 flex flex-col items-center">
            <div className="w-full max-w-4xl flex flex-col flex-1">

                <header className="py-8 flex justify-between items-center mb-16">
                    <Link href="/" className="flex items-center gap-3 font-extrabold text-2xl tracking-tighter">
                        <div className="bg-accent p-2 rounded-xl">
                            <Package className="text-white" size={24} />
                        </div>
                        <span className="text-gradient uppercase">{APP_NAME}</span>
                    </Link>
                    <div className="flex items-center gap-4">
                        <LanguageToggle />
                        <ThemeToggle />
                    </div>
                </header>

                <article className="glass-panel p-8 md:p-16 animate-fade-in mb-20">
                    <Link href="/" className="inline-flex items-center gap-2 text-sm text-text-muted hover:text-accent mb-8 transition-colors">
                        <ArrowLeft size={16} />
                        {dict.common.home}
                    </Link>

                    <div className="flex items-center gap-4 mb-8">
                        <div className="p-3 bg-accent/10 rounded-xl text-accent">
                            <FileText size={24} />
                        </div>
                        <h1 className="text-4xl font-black tracking-tight">{dict.common.terms}</h1>
                    </div>

                    <div className="space-y-10 text-text-muted leading-relaxed font-medium">
                        <section>
                            <h2 className="text-text-main text-xl font-bold mb-4 uppercase tracking-widest">{dict.terms_page.agreeTitle}</h2>
                            <p>
                                {dict.terms_page.agreeDesc}
                            </p>
                        </section>

                        <section>
                            <h2 className="text-text-main text-xl font-bold mb-4 uppercase tracking-widest">{dict.terms_page.useTitle}</h2>
                            <p>
                                {dict.terms_page.useDesc}
                            </p>
                        </section>

                        <section>
                            <h2 className="text-text-main text-xl font-bold mb-4 uppercase tracking-widest">{dict.terms_page.limitTitle}</h2>
                            <p>
                                {dict.terms_page.limitDesc}
                            </p>
                        </section>
                    </div>
                </article>

                <footer className="mt-auto py-12 text-center text-text-muted text-[10px] font-black uppercase tracking-widest">
                    <p>&copy; {new Date().getFullYear()} {APP_NAME} {dict.common.corp}</p>
                </footer>
            </div>
        </main>
    );
}
