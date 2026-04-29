import { PLATFORM_NAME } from '@/constants';

// ============================================
// Supported locales — add new locale codes here and
// add a matching <code>.json in this directory.
// ============================================
export const SUPPORTED_LOCALES = [
  'en', 'es', 'de', 'zh', 'fr', 'pt', 'it', 'ar',
] as const;

export type Locale = (typeof SUPPORTED_LOCALES)[number];

/**
 * Lazy loaders – each locale is its own webpack chunk so the full
 * 79 KB dictionary is never serialised as a single cache string.
 */
const loaders: Record<Locale, () => Promise<{ default: Record<string, unknown> }>> = {
  en: () => import('./en.json'),
  es: () => import('./es.json'),
  de: () => import('./de.json'),
  zh: () => import('./zh.json'),
  fr: () => import('./fr.json'),
  pt: () => import('./pt.json'),
  it: () => import('./it.json'),
  ar: () => import('./ar.json'),
};

// ---------- placeholder interpolation ----------

/** Recursively replace `{{PLATFORM_NAME}}` in every string value. */
function interpolate<T>(obj: T): T {
  if (typeof obj === 'string') {
    return obj.replaceAll('{{PLATFORM_NAME}}', PLATFORM_NAME) as T;
  }
  if (Array.isArray(obj)) {
    return obj.map(interpolate) as T;
  }
  if (obj !== null && typeof obj === 'object') {
    const out: Record<string, unknown> = {};
    for (const [k, v] of Object.entries(obj)) {
      out[k] = interpolate(v);
    }
    return out as T;
  }
  return obj;
}

// ---------- public API ----------

/** The shape of every locale dictionary (typed from the English file). */
export type Dictionary = typeof import('./en.json');

/** Load (and cache) a dictionary for `locale`. */
const cache = new Map<Locale, Dictionary>();

export async function loadDictionary(locale: Locale): Promise<Dictionary> {
  const cached = cache.get(locale);
  if (cached) return cached;

  const loader = loaders[locale];
  if (!loader) {
    // Fallback to English for unknown locales
    return loadDictionary('en');
  }

  const mod = await loader();
  const dict = interpolate(mod.default) as Dictionary;
  cache.set(locale, dict);
  return dict;
}

/** Check if a locale code is valid */
export function isValidLocale(code: string): code is Locale {
  return (SUPPORTED_LOCALES as readonly string[]).includes(code);
}
