# NexGen WebTracker - Frontend Interface 🌐

Modern React-based tracking interface for the NexGen WebTracker logistics platform.

## 🚀 Features

- **Real-time Tracking Interface**: Beautiful search and status visualization for customers.
- **Interactive Map**: Live shipment journey visualization using Leaflet.js.
- **Admin Dashboard**: Comprehensive management of shipments, users, and system configuration.
- **AI-Powered Parser UI**: Dedicated interface for parsing manifest text using Gemini AI via the Go backend.
- **Multi-language Support**: Fully localized in English and Portuguese.
- **Dark Mode**: Premium cosmic-themed UI with system preference detection.

## 🏗️ Architecture

- **Framework**: Next.js 15+ (App Router)
- **Styling**: Tailwind CSS 4
- **Database**: PostgreSQL (via Prisma) for Admin UI and Auth
- **Backend Communication**: Directly interfaces with the Go-based WhatsApp Bot & Tracking API
- **Maps**: Leaflet.js

## 📋 Prerequisites

- Node.js 20+ and npm
- PostgreSQL database (Local or Cloud)
- Go Backend (Running separately)

## 🛠️ Installation

### 1. Install Dependencies

```bash
cd front
npm install
```

### 2. Configure Environment Variables

Create `.env` file in the `front/` directory:

```env
# Database (Prisma)
DATABASE_URL="postgresql://user:pass@localhost:5432/webtracker"

# Authentication
NEXTAUTH_URL="http://localhost:3000"
NEXTAUTH_SECRET="your-generated-secret"

# Admin Credentials (Initial setup)
ADMIN_USERNAME="admin"
ADMIN_PASSWORD="your-strong-password"
ADMIN_EMAIL="admin@yourdomain.com"

# External API Integration (Go Backend)
NEXT_PUBLIC_API_URL="http://localhost:8080"
API_AUTH_TOKEN="your-shared-secret-token"
```

### 3. Database Sync

```bash
npx prisma generate
npx prisma db push
```

### 4. Run Development Server

```bash
npm run dev
```

## 🎨 Branding Customization

Edit `src/lib/constants.ts` to rebrand the entire platform:

```typescript
export const COMPANY_NAME = "NexGen Logistics";
export const SUPPORT_EMAIL = "support@nexgenlogistics.com";
```

## 🔐 Security

- **NextAuth**: Handles admin session management.
- **API Token**: All requests to the Go backend are secured via a shared `API_AUTH_TOKEN`.
- **Protected Routes**: `/admin` is strictly restricted to authenticated users.

## 📝 License

Proprietary software. All rights reserved.
