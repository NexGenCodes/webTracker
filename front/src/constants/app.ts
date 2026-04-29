// ============================================
// SYSTEM CONSTANTS
// PLATFORM_NAME is used for SEO metadata & legal pages (server-rendered, can't use hooks).
// Company-specific branding comes from the DB via useCompanySettings().
// ============================================
export const PLATFORM_NAME = "CargoHive";
export const APP_DESCRIPTION = "Real-time shipment tracking with advanced privacy protection.";

export const CONTACT_EMAIL = process.env.NEXT_PUBLIC_CONTACT_EMAIL || "support@cargohive.com";
export const CONTACT_PHONE = process.env.NEXT_PUBLIC_CONTACT_PHONE || "+1 (555) 123-4567";
export const CONTACT_HQ = process.env.NEXT_PUBLIC_CONTACT_HQ || "San Francisco, CA";

export const ADMIN_TIMEZONE = process.env.ADMIN_TIMEZONE || "Africa/Lagos";
