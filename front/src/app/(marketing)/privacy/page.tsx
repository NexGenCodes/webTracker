"use client";

import { useI18n } from '@/components/providers/I18nContext';
import { ArrowLeft, Shield } from "lucide-react";
import Link from "next/link";
import { PLATFORM_NAME } from '@/constants';

export default function PrivacyPage() {
    const { dict } = useI18n();

    return (
        <main className="min-h-screen p-4 flex flex-col items-center pt-32">
            <div className="w-full max-w-4xl flex flex-col flex-1">
                <article className="glass-panel p-8 md:p-16 animate-fade-in mb-20">
                    <Link href="/" className="inline-flex items-center gap-2 text-sm text-text-muted hover:text-accent mb-8 transition-colors">
                        <ArrowLeft size={16} />
                        {dict.common.home}
                    </Link>

                    <div className="flex items-center gap-4 mb-8">
                        <div className="p-3 bg-accent/10 rounded-xl text-accent">
                            <Shield size={24} />
                        </div>
                        <h1 className="text-4xl font-black tracking-tight">{dict.common.privacy}</h1>
                    </div>

                    <div className="space-y-10 text-text-muted leading-relaxed font-medium">
                        <section>
                            <h2 className="text-text-main text-xl font-bold mb-4 uppercase tracking-widest">{dict.privacy_page.introTitle}</h2>
                            <p>
                                {dict.privacy_page.introDesc.replace('{{PLATFORM_NAME}}', PLATFORM_NAME)}
                            </p>
                        </section>

                        <section>
                            <h2 className="text-text-main text-xl font-bold mb-4 uppercase tracking-widest">{dict.privacy_page.dataTitle}</h2>
                            <p>
                                {dict.privacy_page.dataDesc}
                            </p>
                            <ul className="list-disc pl-6 mt-4 space-y-2">
                                <li>{dict.privacy_page.item1}</li>
                                <li>{dict.privacy_page.item2}</li>
                                <li>{dict.privacy_page.item3}</li>
                            </ul>
                            <p className="mt-4">
                                {dict.privacy_page.dataFinal}
                            </p>
                        </section>

                        <section>
                            <h2 className="text-text-main text-xl font-bold mb-4 uppercase tracking-widest">{dict.privacy_page.securityTitle}</h2>
                            <p>
                                {dict.privacy_page.securityDesc}
                            </p>
                        </section>
                    </div>
                </article>
            </div>
        </main>
    );
}
