/**
 * Re-export from the split dictionaries module.
 * The actual locale JSON files live in ./dictionaries/*.json and are
 * loaded lazily via dynamic import() — see ./dictionaries/index.ts.
 *
 * This shim keeps existing `import { Locale, Dictionary } from '@/lib/dictionaries'`
 * working without any consumer changes.
 */
export { loadDictionary, isValidLocale, SUPPORTED_LOCALES } from './dictionaries/index';
export type { Locale, Dictionary } from './dictionaries/index';
