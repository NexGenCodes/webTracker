# Competitive Strategy & Roadmap

Based on deep analysis of 12+ competitor platforms (Deprixa Plus, Portis, PickPack, GoFreight, FourKites, Shippeo, Logeestico, Pango, Logitude World, Shipeezi, FrateZone, LOG-NET) and full codebase audit.

## Competitive Landscape Summary

| Area | CargoHive (You) | Deprixa Plus ($79) | Portis ($499-999/mo) | PickPack (Enterprise) | GoFreight (Volume) |
|---|---|---|---|---|---|
| Multi-tenant | ✅ RLS+JWT | ✅ SaaS | ✅ Unlimited clients | ✅ "By design" | ✅ |
| WhatsApp Bot | ✅ **Native (whatsmeow)** | ✅ UltraMsg | ❌ | ❌ | ❌ |
| Manifest Parsing | ✅ **Regex+Gemini** | ❌ | ❌ | ❌ | ❌ |
| AI Parser | ✅ Gemini | ❌ | ❌ | ❌ | ❌ |
| Paystack | ✅ | ✅ | ❌ | ❌ | ❌ |
| White-label | ✅ Branding per tenant | ✅ | ✅ Portal | ✅ End-customer | ✅ Portal |
| i18n | **8 languages** | 2 (EN/ES) | ❌ | ❌ | ❌ |
| Driver Dispatch | ❌ Not built | ✅ Leaflet+Mapbox | ❌ | ✅ | ❌ |
| Customer Portal | ❌ Public track only | ✅ Full self-service | ✅ Docs+quotes | ❌ | ✅ |
| REST API | ❌ Internal only | ✅ Sanctum | ❌ | ✅ | ✅ |
| Barcode Scan | ❌ | ✅ | ❌ | ✅ | ❌ |
| COD Pipeline | ❌ | ✅ | ❌ | ❌ | ❌ |
| POD (Proof of Delivery) | ❌ | ❌ | ❌ | ✅ | ✅ |
| Multi-currency | ❌ NGN only | ✅ | ❌ | ✅ | ✅ |
| Rate Calculator | ❌ | ✅ Public /rate | ❌ | ❌ | ❌ |
| Predictive ETA | ❌ Static | ❌ | ❌ | ❌ | ❌ |
| Public API/SDKs | ❌ | ✅ REST API | ❌ | ✅ EDI+API | ✅ API |
| Mobile App (Driver) | ❌ | ❌ | ❌ | ❌ | ❌ |

---

## 1. Billing & Multi-Currency

### Current State
- 3 plans: Starter (₦12K), Pro (₦30K), Scale (₦85K) — all NGN-only
- Paystack integration with HMAC webhook verification — solid
- Subscription expiry, rate limiting by tier — implemented
- Trial: 7 days for Starter plan — live

### Competitor Benchmark
- **Deprixa Plus**: 5 plans (Free/Starter/Growth/Pro/Enterprise) with wallet-based billing, auto-renewal, grace periods, full lifecycle management
- **Portis**: $499-999/mo, transparent tiered pricing with feature gating
- **GoFreight**: Volume-based pricing (tied to shipment count)

### Recommended Upgrades
1. **Multi-currency payouts** — Add USD and GBP pricing via Paystack's multi-currency support. Enables international logistics companies to subscribe.
   - Backend: Add `Currency` field to `Plan` struct (already partially there at `Plan.Currency`)
   - Frontend: Currency toggle on pricing page (NGN/USD/GBP)
   - Paystack: Use `currency` parameter in payment initialization

2. **Annual billing discount** — 15-20% discount for annual commitment (Portis does "save 15% annually")
   - Backend: Add `Interval` field to plans (monthly/annually) with adjusted pricing

3. **Usage-based overage billing** — Beyond shipment caps, charge per extra shipment (₦500/shipment for Starter, ₦200 for Pro)
   - Backend: Already has `MaxShipments` in Plan, add overage tracking in billing use case

4. **Invoice generation** — Auto-generate PDF invoices for each payment cycle
   - Frontend: Invoice download button in billing history tab

5. **Grace period + dunning** — 3-day grace before downgrade, email reminders on day 1, 3, and expiry
   - Backend: Cron job for subscription expiry checks (similar to existing status transition cron)

---

## 2. RBAC (Role-Based Access Control)

### Current State
- 2 roles: `admin` (company owner) and `super_admin` (platform-level)
- No staff/dispatch rider role
- JWT has `role` claim — Easy to extend
- Frontend middleware already checks role for route protection

### Competitor Benchmark
- **Deprixa Plus**: 102 granular permissions across 5 role types (Super Admin, Admin, Employee (38), Driver (9), Customer (6)) via Spatie Laravel Permission
- **GoFreight**: Role-based access per client in multi-tenant setup

### Recommended Upgrades
1. **Staff role** — Dashboard-only access (no billing, no settings). For office staff to create/manage shipments.
   ```
   RoleStaff = "staff"
   Permissions: shipments.read, shipments.write, tracking.view
   Denied: billing.*, settings.*, whatsapp.*
   ```
   - Frontend: Hide billing/settings tabs from staff users in DashboardClient

2. **Dispatch Rider role** — Mobile-only access via WhatsApp. Can update status to "Delivered" and send location.
   ```
   RoleRider = "rider"
   Permissions: shipments.read_assigned, shipments.update_status
   Auth: Phone-based PIN (not JWT), tied to `company_id`
   ```
   - Backend: Simple PIN auth endpoint, lightweight JWT with limited expiry
   - WhatsApp: `!status TRACK123 delivered` command already exists — extend with rider auth

3. **Permission system** — Move from hardcoded role checks to a permission-bitmask or Spatie-style system
   - Backend: Create `permissions` table with `(company_id, role, resource, action)` rows
   - Middleware: Permission-checking middleware runs after JWT auth

4. **Team management UI** — Invite staff via email, assign role, revoke access
   - Frontend: "Team" tab in Settings page

---

## 3. Authentication Improvements

### Current State
- Email/password registration with OTP verification — works
- ES256 JWT with company_id, role, plan_type claims — solid
- Password reset via email OTP — implemented
- HttpOnly cookie JWT storage — secure
- Supabase RLS with custom JWT — correctly configured

### Competitor Benchmark
- **Deprixa Plus**: Google OAuth social login, Laravel Sanctum token auth for API
- **Portis**: None visible (managed SaaS, no self-service auth)

### Recommended Upgrades
1. **Social login (Google OAuth)** — Deprixa Plus has this. Reduces registration friction.
   - Backend: Google OAuth callback endpoint, creates company on first login
   - Frontend: "Sign up with Google" button on auth page

2. **Magic link login** — Email-based passwordless login for non-technical logistics owners
   - Backend: Generate short-lived token, email via Brevo, validate and issue JWT

3. **Session management UI** — View active sessions (device, last active), revoke sessions
   - Frontend: "Sessions" section in Settings page
   - Backend: Redis-backed session store keyed by JWT ID (`jti`)

4. **Rate-limiting on auth endpoints** — Already have Redis rate limiter in `internal/cache/limiter.go` — apply it to login/register/OTP endpoints

---

## 4. Live Map Tracking

### Current State
- Leaflet map with origin/destination markers and dashed flight path — ✅ Good foundation
- Quadratic Bezier curve with animated plane icon — ✅ Impressive UX
- Progress-based interpolation with rotation — ✅ Smart
- Dark mode tile inversion — ✅ Done
- **Map fixed at h-[600px] on mobile** — ❌ Too tall for phones (85% of viewport)

### Competitor Benchmark
- **Deprixa Plus**: Leaflet + Mapbox GL with live driver tracking on dispatch map
- **PickPack**: Route optimization with real-time GPS tracking
- **Logeestico**: Real-time driver tracking with fleet management
- **FourKites**: Real-time GPS tracking with predictive ETA

### Recommended Upgrades
1. **Mobile-responsive map height**
   ```
   h-[300px] sm:h-[400px] md:h-[600px] lg:h-[750px]
   ```
   or use `max-h-[60vh]` for dynamic viewport-aware sizing

2. **Live GPS tracking** — For Scale tier, integrate with driver's mobile GPS
   - Option A: WhatsApp `LocationMessage` — drivers share live location, bot stores to DB
   - Option B: Lightweight PWA — simple "Share Location" page that uses Geolocation API and pings backend every 30s
   - Frontend: Update map marker in real-time via Supabase Realtime or polling

3. **Geofencing** — Auto-trigger status transitions when driver enters/exits geofence zones
   - Backend: Store geofence points per shipment (origin warehouse, destination), check via cron
   - Use case: Driver crosses destination geofence → auto-set "Delivered" status

4. **Multi-stop route visualization** — Show all stops on map with numbered markers
   - Currently: Only origin→destination single leg

5. **Map clustering** — For dashboard overview showing all active shipments on one map

---

## 5. Frontend Mobile Responsiveness

### Audit Score: GOOD (with specific gaps)

**What's done well:**
- MobileNavOverlay with body scroll lock
- Hamburger menu on both marketing and dashboard headers
- SuperAdminClient has best-practice mobile card list (hidden desktop table + mobile card view)
- All modals use `max-h-[90vh]` with overflow scroll
- `input-premium` sets `font-size: 16px` preventing iOS zoom
- `active:scale-95` touch feedback on buttons
- Landing sections consistently use `grid-cols-1 md:grid-cols-N`
- `background-attachment: scroll` on mobile prevents scroll jank

**Critical issues to fix:**

| Issue | File | Fix |
|-------|------|-----|
| Shipments table has no mobile card view | `ShipmentsTab.tsx:212-330` | Follow `SuperAdminClient` pattern: `hidden md:table` + `md:hidden` card list |
| Map too tall on mobile (600px) | `MapComponent.tsx:92` | `h-[300px] sm:h-[400px] md:h-[600px] lg:h-[750px]` |
| ShipmentMapBar 3-col grid too tight | `ShipmentMapBar.tsx:22` | Collapse to 1 or 2 columns on `max-[400px]` |
| Billing transaction history no mobile view | `BillingClient.tsx:358-418` | Card list pattern like SuperAdminClient |
| Tracking ID truncated at 120px | `ShipmentStatusHeader.tsx:33` | `max-w-[180px]` or `break-all` |
| ShipmentDetails text too small (7px) | `ShipmentDetails.tsx:28` | Bump to `text-[10px] md:text-sm` |
| No skip-to-content link | `layout.tsx` | Add `<a href="#main-content">` for accessibility |
| No theme-color meta tag | `layout.tsx` | Add `<meta name="theme-color">` for PWA-like chrome |
| Analytics legend may overflow | `AnalyticsTab.tsx:189` | `flex-wrap` on legend container |

**This is NOT a standalone skill — it's standard UI/UX implementation work.** Every fix above is a specific Tailwind CSS change or component pattern refactor. No new library required. The patterns already exist in SuperAdminClient. Time estimate: 4-6 hours for all fixes.

---

## Can CargoHive Beat These Competitors?

### Areas Where You Already Win

| Area | Why You Win |
|------|-------------|
| **WhatsApp manifest parsing** | No competitor does this. Regex-first (95%) + Gemini fallback. This is your moat. |
| **Per-tenant WhatsApp bot instances** | Each company gets its own connected bot. Deprixa Plus uses shared UltraMsg channel. |
| **Affordable pricing** | ₦12K-85K/mo (~$8-55 USD). Portis starts at $499/mo. Deprixa Plus is $79 one-time (no support). |
| **8-language i18n** | Only Deprixa Plus has EN/ES. You support Arabic, Chinese, French for global logistics. |
| **Super admin portal** | Tenant management, audit logs, forced disconnect. No $79 CodeCanyon product has this. |
| **Edge-core hybrid architecture** | 50ms tracking lookups via Vercel edge + persistent WhatsApp on VPS. Unique split. |

### Areas Where You're Behind

| Area | Gap | Effort to Close |
|------|-----|-----------------|
| **Customer portal** | Portis charges $499/mo for this alone. Yours + WhatsApp would beat them. | **2-3 weeks** |
| **Driver dispatch + map** | Deprixa Plus has this built-in for $79 total. | **3-4 weeks** |
| **Public REST API** | Required for Scale tier. Deprixa Plus has Sanctum API. | **1 week** |
| **Multi-currency** | NGN-only limits addressable market. | **3-5 days** |
| **Barcode scanning** | Deprixa Plus has it. Needed for warehouse workflow. | **1 week** |
| **POD with signature/photo** | Table stakes for B2B logistics. | **2 weeks** |
| **Mobile app** | Shipeezi, Logeestico have iOS/Android apps. | **Long-term** |

### Verdict

**Yes, you can beat them — but not by out-featuring them.** Your strategy should be:

1. **Lead with WhatsApp automation** — This is your wedge. No competitor can match it. Every logistics company in Nigeria/Ghana/Kenya already uses WhatsApp for operations. You automate what they do manually.

2. **Copy the gaps that matter** — Customer portal (biggest quick win), public API, driver dispatch. Portis makes $499/mo on just the portal alone. You can offer portal + WhatsApp for ₦30K/mo.

3. **Ignore enterprise features** — Don't chase FourKites/Shippeo on predictive ETAs or 1000+ carrier integrations. Target the underserved SME logistics market in Africa/South Asia where WhatsApp is the operating system.

4. **Pricing advantage is real** — Your Pro (₦30K/~$20) is cheaper than any Western competitor's entry tier. But Deprixa Plus at $79 one-time undercuts you on upfront cost. Your moat is ongoing WhatsApp automation they can't replicate.

### Recommended Priority Order

1. **Customer portal** — Biggest ROI. Portis charges $499/mo for this alone. You already have the tracking data.
2. **Public REST API** — Unlocks Scale tier. Required for e-commerce integrations.
3. **Multi-currency** — Opens USD/GBP markets. Paystack already supports this.
4. **Driver dispatch + live map** — Closes biggest feature gap vs Deprixa Plus.
5. **Barcode scanning** — Operational necessity.
6. **POD with photo/signature** — Enterprise table stakes.
7. **Mobile app** — When you have driver network effects.

---

## Appendix: Mobile Responsiveness Fix Plan

| Priority | File | Change | Est. Time |
|----------|------|--------|-----------|
| P0 | `MapComponent.tsx:92` | Add sm/md breakpoints to height | 5 min |
| P0 | `ShipmentsTab.tsx` | Mobile card view (copy SuperAdminClient pattern) | 2 hr |
| P1 | `BillingClient.tsx` | Mobile card view for transaction history | 1 hr |
| P1 | `ShipmentMapBar.tsx:22` | Collapse grid on small screens | 15 min |
| P2 | `ShipmentStatusHeader.tsx:33` | Increase truncation width | 2 min |
| P2 | `ShipmentDetails.tsx:28` | Bump minimum font size | 5 min |
| P2 | `layout.tsx` | Add skip-to-content + theme-color | 10 min |
| P3 | `AnalyticsTab.tsx:189` | Wrap legend | 5 min |
