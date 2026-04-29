"use client";

import { useEffect, useRef, useState } from 'react';
import { Package, Globe, Shield, Clock } from 'lucide-react';

import { useI18n } from '@/components/providers/I18nContext';

function getStats(dict: any) {
  return [
    { icon: Package, value: "100+", label: dict.marketing?.trust?.shipments || "Shipments Tracked", target: 100 },
    { icon: Globe, value: "3+", label: dict.marketing?.trust?.countries || "Countries", target: 3 },
    { icon: Shield, value: "99.9%", label: dict.marketing?.trust?.uptime || "Uptime", target: 99.9 },
    { icon: Clock, value: "24/7", label: dict.marketing?.trust?.live || "Live Monitoring", target: 0 },
  ];
}

function AnimatedCounter({ target, suffix = "" }: { target: number; suffix?: string }) {
  const [count, setCount] = useState(0);
  const ref = useRef<HTMLSpanElement>(null);
  const hasAnimated = useRef(false);

  useEffect(() => {
    const observer = new IntersectionObserver(
      ([entry]) => {
        if (entry.isIntersecting && !hasAnimated.current) {
          hasAnimated.current = true;
          if (target === 0) return; // Skip "24/7" type values

          const duration = 2000;
          const startTime = performance.now();

          const animate = (currentTime: number) => {
            const elapsed = currentTime - startTime;
            const progress = Math.min(elapsed / duration, 1);
            const eased = 1 - Math.pow(1 - progress, 3); // easeOutCubic
            setCount(Math.floor(eased * target));

            if (progress < 1) requestAnimationFrame(animate);
          };

          requestAnimationFrame(animate);
        }
      },
      { threshold: 0.3 }
    );

    if (ref.current) observer.observe(ref.current);
    return () => observer.disconnect();
  }, [target]);

  if (target === 0) return null; // "24/7" rendered separately

  return (
    <span ref={ref}>
      {target >= 1000
        ? `${(count / 1000).toFixed(count >= target ? 0 : 0).replace(/\.0$/, '')}${count >= 1000 ? ',000' : ''}`
        : target >= 100
          ? count.toFixed(target % 1 !== 0 ? 1 : 0)
          : count
      }
      {suffix}
    </span>
  );
}

export function TrustMetricsSection() {
  const { dict } = useI18n();
  const stats = getStats(dict);

  return (
    <section className="py-16 relative z-10 w-full">
      <div className="max-w-7xl mx-auto px-4 md:px-6">
        <div className="glass-panel p-8 md:p-12 rounded-[2rem]">
          <div className="grid grid-cols-2 md:grid-cols-4 gap-8 md:gap-4">
            {stats.map((stat, i) => {
              const Icon = stat.icon;
              return (
                <div
                  key={i}
                  className="flex flex-col items-center text-center group"
                >
                  <div className="w-12 h-12 bg-accent/10 rounded-xl flex items-center justify-center text-accent mb-4 group-hover:bg-accent group-hover:text-white transition-all duration-500">
                    <Icon size={22} />
                  </div>
                  <span className="text-3xl md:text-4xl font-black tracking-tighter mb-1">
                    {stat.target === 0 ? (
                      stat.value
                    ) : stat.target >= 100 ? (
                      <>
                        <AnimatedCounter target={stat.target} />
                        {stat.value.includes('+') ? '+' : stat.value.includes('%') ? '%' : ''}
                      </>
                    ) : (
                      <>
                        <AnimatedCounter target={stat.target} />
                        {stat.value.includes('+') ? '+' : ''}
                      </>
                    )}
                  </span>
                  <span className="text-[10px] font-black uppercase tracking-[0.2em] text-text-muted opacity-60">
                    {stat.label}
                  </span>
                </div>
              );
            })}
          </div>
        </div>
      </div>
    </section>
  );
}
