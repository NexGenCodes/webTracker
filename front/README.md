# CargoHive WebTracker — Frontend 🌐

Modern Next.js tracking interface for the CargoHive logistics platform.

## 🚀 Features

- **Real-time Tracking**: Public search with live status visualization and interactive Leaflet maps.
- **Admin Dashboard**: Shipment management with Supabase Realtime sync for live updates.
- **WhatsApp Integration**: QR/phone-code pairing modal for multi-device WhatsApp bot linking.
- **AI-Powered Parsing**: Dedicated interface for parsing shipping manifests via the Go backend.
- **Billing & Subscriptions**: Paystack-integrated plan management with payment history.
- **Multi-language**: Fully localized in English and Portuguese.
- **Dark Mode**: Premium glassmorphism UI with system preference detection.

## 🏗️ Architecture

| Layer | Technology |
|-------|-----------|
| **Framework** | Next.js 15 (App Router) |
| **UI** | React 19 + Tailwind CSS 4 + Framer Motion |
| **Auth** | Custom JWT (ES256) via Go backend, HttpOnly cookies, jose verification |
| **Database** | Supabase (PostgreSQL + Realtime) |
| **Data Fetching** | React Query (TanStack) + Server Actions |
| **Forms** | React Hook Form + Zod |
| **Maps** | Leaflet.js / react-leaflet |
| **Charts** | Recharts |
| **Payments** | Paystack |
| **Monitoring** | Sentry + Vercel Analytics + Speed Insights |
| **Testing** | Vitest + React Testing Library |

## 📋 Prerequisites

- Node.js 20+ and npm
- Go Backend running (provides auth, shipment APIs, WhatsApp bot)
- Supabase project (for Realtime and direct DB access)

## 🛠️ Setup

### 1. Install Dependencies

```bash
cd front
npm install
```

### 2. Configure Environment

Copy `.env.example` to `.env` and fill in values:

```env
# Supabase
NEXT_PUBLIC_SUPABASE_URL="https://[PROJECT-ID].supabase.co"
NEXT_PUBLIC_SUPABASE_ANON_KEY="your-anon-key"

# JWT (server-only — no NEXT_PUBLIC_ prefix)
JWT_PUBLIC_KEY="-----BEGIN PUBLIC KEY-----\n...\n-----END PUBLIC KEY-----"

# Backend
NEXT_PUBLIC_API_URL="http://localhost:8080"

# Sentry
NEXT_PUBLIC_SENTRY_DSN="https://..."
```

### 3. Run Development Server

```bash
npm run dev
```

## 🔐 Security

- **JWT Auth**: Go backend issues ES256 JWTs. Next.js middleware + `jose` verifies them server-side.
- **HttpOnly Cookies**: JWT is stored in an HttpOnly cookie — never accessible to client JavaScript.
- **Server Actions**: All mutations (login, register, billing) go through Next.js Server Actions, keeping tokens server-side.
- **RLS**: Supabase Row Level Security enforces tenant isolation at the database level.

## 📁 Project Structure

```
src/
├── app/
│   ├── (auth)/          # Login, register, OTP, password reset
│   ├── (dashboard)/     # Dashboard, billing, settings
│   ├── (marketing)/     # Landing, pricing, about, tracking, legal
│   └── actions/         # Server Actions (auth, billing, setup, etc.)
├── components/
│   ├── auth/            # Auth form views
│   ├── billing/         # Plan cards, payment history
│   ├── dashboard/       # Dashboard client, overview, WhatsApp
│   ├── landing/         # Marketing page sections
│   ├── layout/          # Header, footer, backgrounds
│   ├── map/             # Leaflet map components
│   ├── providers/       # React Query, theme, i18n, multi-tenant
│   ├── shared/          # Reusable inputs, toggles
│   └── tracking/        # Tracking search, timeline, details
├── hooks/               # Custom hooks (auth, company settings)
├── lib/                 # Utilities, Supabase clients, auth, logger
├── services/            # Server-side data fetching services
├── types/               # TypeScript type definitions
└── constants/           # App constants, geo data
```

## 📝 License

Proprietary software. All rights reserved.
