import { type ClassValue, clsx } from 'clsx';
import { twMerge } from 'tailwind-merge';

/**
 * Utility to merge tailwind classes safely.
 */
export function cn(...inputs: ClassValue[]) {
    return twMerge(clsx(inputs));
}

/**
 * Splits a word into its phonetic syllables using a basic heuristic.
 * Replicates the logic in backend/internal/utils/syllable.go
 */
export function splitIntoSyllables(word: string): string[] {
    word = word.trim();
    if (!word) return [];

    const cleanWord = word.replace(/[^a-zA-Z]/g, '');
    if (cleanWord.length <= 3) return [cleanWord];

    const vowels = "aeiouyAEIOUY";
    const isVowel = (c: string) => vowels.includes(c);

    const syllables: string[] = [];
    let start = 0;
    const chars = Array.from(cleanWord);

    for (let i = 0; i < chars.length; i++) {
        if (isVowel(chars[i])) {
            if (i + 1 < chars.length && !isVowel(chars[i + 1])) {
                let hasMoreVowels = false;
                for (let j = i + 1; j < chars.length; j++) {
                    if (isVowel(chars[j])) {
                        hasMoreVowels = true;
                        break;
                    }
                }

                if (hasMoreVowels) {
                    // Basic split rule: VCV -> V-CV, VCCV -> VC-CV
                    if (i + 2 < chars.length && !isVowel(chars[i + 2])) {
                        // VCCV
                        syllables.push(chars.slice(start, i + 2).join(''));
                        start = i + 2;
                        i = i + 1;
                    } else {
                        // VCV
                        syllables.push(chars.slice(start, i + 1).join(''));
                        start = i + 1;
                    }
                }
            }
        }
    }

    if (start < chars.length) {
        syllables.push(chars.slice(start).join(''));
    }

    // Post-process: Merge very short leftovers if they aren't the only ones
    if (syllables.length > 1 && syllables[syllables.length - 1].length < 2) {
        const last = syllables.pop()!;
        syllables[syllables.length - 1] += last;
    }

    return syllables;
}

/**
 * Generates a 3-character abbreviation for a company name.
 * Replicates the logic in backend/internal/config/config.go
 */
export function generateAbbreviation(name: string): string {
    name = name.trim().replace(/\s+/g, '');
    if (!name) return "AWB";

    const clean = name.replace(/[^a-zA-Z]/g, '');
    if (!clean) return "AWB";

    const syllables = splitIntoSyllables(clean);
    const count = syllables.length;

    let abbr = "";
    if (count === 1) {
        const s = syllables[0];
        if (s.length <= 3) {
            abbr = s;
        } else {
            abbr = s[0] + s[Math.floor(s.length / 2)] + s[s.length - 1];
        }
    } else if (count === 2) {
        const s1 = syllables[0];
        const s2 = syllables[1];
        const p1 = s1.length > 2 ? s1.substring(0, 2) : s1;
        const p2 = s2[0];
        abbr = p1 + p2;
    } else {
        for (let i = 0; i < 3 && i < count; i++) {
            abbr += syllables[i][0];
        }
    }

    abbr = abbr.toUpperCase();
    if (abbr.length > 3) {
        abbr = abbr.substring(0, 3);
    }
    while (abbr.length < 3) {
        abbr += "X";
    }

    return abbr;
}

/**
 * Validates tracking number format.
 * Matches backend: PREFIX-123456789 (Prefix + '-' + 9 digits)
 */
export function isValidTrackingNumber(id: string): boolean {
    if (!id) return false;
    const regex = /^[A-Z0-9]{2,6}-[0-9]{5,12}$/i;
    return regex.test(id);
}

/**
 * Resolves the backend API URL dynamically.
 * Uses localhost:8080 during local development, and the .env URL in production.
 */
export function getApiUrl(): string {
    return process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080';
}
