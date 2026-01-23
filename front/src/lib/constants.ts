// ============================================
// BRANDING CONSTANTS (Plug and Play)
// ============================================
// Change APP_NAME to rebrand the entire application
export const APP_NAME = process.env.NEXT_PUBLIC_COMPANY_NAME || "Airway Bill";

/**
 * Auto-generate tracking prefix from company name
 * Examples: "Airway Bill" -> "AWB", "Test Express" -> "TEX", "MyCompany" -> "MYC"
 */
function generateTrackingPrefix(name: string): string {
    const trimmed = name.trim();
    if (!trimmed) return "AWB";

    const words = trimmed.split(/\s+/);

    // Multi-word: take first letter of each word
    if (words.length > 1) {
        return words
            .map(word => word[0])
            .join('')
            .toUpperCase();
    }

    // Single word: take first 3 letters
    return trimmed.substring(0, 3).toUpperCase();
}

export const TRACKING_PREFIX = generateTrackingPrefix(APP_NAME);
export const APP_DESCRIPTION = "Real-time shipment tracking with advanced privacy protection.";

export const CONTACT_EMAIL = process.env.NEXT_PUBLIC_CONTACT_EMAIL || "support@airwaybill.com";
export const CONTACT_PHONE = process.env.NEXT_PUBLIC_CONTACT_PHONE || "+1 (555) 123-4567";
export const CONTACT_HQ = process.env.NEXT_PUBLIC_CONTACT_HQ || "San Francisco, CA";

export const COUNTRY_COORDS: Record<string, [number, number]> = {
    'USA': [37.0902, -95.7129],
    'United States': [37.0902, -95.7129],
    'US': [37.0902, -95.7129],
    'China': [35.8617, 104.1954],
    'Germany': [51.1657, 10.4515],
    'UK': [55.3781, -3.4360],
    'United Kingdom': [55.3781, -3.4360],
    'France': [46.2276, 2.2137],
    'Japan': [36.2048, 138.2529],
    'Australia': [-25.2744, 133.7751],
    'Canada': [56.1304, -106.3468],
    'Brazil': [-14.2350, -51.9253],
    'India': [20.5937, 78.9629],
    'Mexico': [23.6345, -102.5528],
    'Spain': [40.4637, -3.7492],
    'Italy': [41.8719, 12.5674],
    'Nigeria': [9.0820, 8.6753],
    'Portugal': [39.3999, -8.2245],
    'Argentina': [-38.4161, -63.6167],
    'Colombia': [4.5709, -74.2973],
    'Peru': [-9.1900, -75.0152],
    'Chile': [-35.6751, -71.5430],
    'Ecuador': [-1.8312, -78.1834],
    'Venezuela': [6.4238, -66.5897],
    'Bolivia': [-16.2902, -63.5887],
    'Dominican Republic': [18.7357, -70.1627],
    'Guatemala': [15.7835, -90.2308],
    'Cuba': [21.5218, -77.7812],
    'Honduras': [15.2000, -86.2419],
    'Paraguay': [-23.4425, -58.4438],
    'El Salvador': [13.7942, -88.8965],
    'Nicaragua': [12.8654, -85.2072],
    'Costa Rica': [9.7489, -83.7534],
    'Panama': [8.5380, -80.7821],
    'Uruguay': [-32.5228, -55.7658],
    'Puerto Rico': [18.2208, -66.5901],
    'South Africa': [-30.5595, 22.9375],
    'Kenya': [-0.0236, 37.9062],
    'Egypt': [26.8206, 30.8025],
    'Ghana': [7.9465, -1.0232],
    'Morocco': [31.7917, -7.0926],
    'Ethiopia': [9.1450, 40.4897],
    'Tanzania': [-6.3690, 34.8888],
    'South Korea': [35.9078, 127.7669],
    'Thailand': [15.8700, 100.9925],
    'Vietnam': [14.0583, 108.2772],
    'Philippines': [12.8797, 121.7740],
    'Indonesia': [-0.7893, 113.9213],
    'Malaysia': [4.2105, 101.9758],
    'Singapore': [1.3521, 103.8198],
    'Turkey': [38.9637, 35.2433],
    'Russia': [61.5240, 105.3188],
    'Poland': [51.9194, 19.1451],
    'Netherlands': [52.1326, 5.2913],
    'Belgium': [50.5039, 4.4699],
    'Sweden': [60.1282, 18.6435],
    'Norway': [60.4720, 8.4689],
    'Denmark': [56.2639, 9.5018],
    'Finland': [61.9241, 25.7482],
    'Switzerland': [46.8182, 8.2275],
    'Austria': [47.5162, 14.5501],
    'Greece': [39.0742, 21.8243],
    'Ireland': [53.4129, -8.2439],
    'New Zealand': [-40.9006, 174.8860],
    'UAE': [23.4241, 53.8478],
    'Saudi Arabia': [23.8859, 45.0792],
    'Israel': [31.0461, 34.8516],
}

export const ADMIN_TIMEZONE = process.env.ADMIN_TIMEZONE || "Africa/Lagos";

/**
 * Validates tracking number format.
 * Matches backend: PREFIX-123456789 (Prefix + '-' + 9 digits)
 */
export function isValidTrackingNumber(id: string): boolean {
    if (!id) return false;

    // More lenient regex to handle various potential prefix lengths
    // and ensuring it matches the backend pattern of Prefix-XXXXXXXXX (9 digits)
    const regex = /^[A-Z0-9]{2,6}-[0-9]{5,12}$/i;

    return regex.test(id);
}
