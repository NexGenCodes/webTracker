# WebTracker - International Logistics Tracking Platform

<div align="center">

**A modern, full-stack shipment tracking system with WhatsApp integration and AI-powered manifest parsing**

[![Next.js](https://img.shields.io/badge/Next.js-16.1-black?logo=next.js)](https://nextjs.org/)
[![Go](https://img.shields.io/badge/Go-1.25-00ADD8?logo=go)](https://golang.org/)
[![PostgreSQL](https://img.shields.io/badge/PostgreSQL-Supabase-336791?logo=postgresql)](https://www.postgresql.org/)
[![TypeScript](https://img.shields.io/badge/TypeScript-5.0-3178C6?logo=typescript)](https://www.typescriptlang.org/)

[Features](#-features) • [Architecture](#-architecture) • [Getting Started](#-getting-started) • [Deployment](#-deployment) • [API Reference](#-api-reference)

</div>

---

## 📋 Overview

**WebTracker** is a comprehensive international logistics tracking platform designed for modern shipping operations. It combines a beautiful Next.js web interface with a powerful Go-based WhatsApp bot to provide real-time shipment tracking, automated notifications, and AI-powered manifest processing.

### Key Highlights

- 🌍 **Multi-Country Support** - Track shipments across international borders with country-specific delivery time calculations
- 📱 **WhatsApp Integration** - Automated bot for manifest submission and real-time shipment updates
- 🤖 **AI-Powered Parsing** - Google Gemini AI extracts shipment details from natural language messages
- 🗺️ **Interactive Maps** - Real-time shipment visualization with Leaflet.js
- 🔐 **Secure Admin Portal** - NextAuth-powered authentication with comprehensive shipment management
- 🌓 **Dark Mode** - Beautiful cosmic-themed UI with light/dark mode support
- 🌐 **Internationalization** - Full support for English and Portuguese
- ⚡ **Real-time Updates** - Automated status transitions an## 🏗️ Architecture (Edge-Core Hybrid)

```mermaid
graph TB
    subgraph "Edge Layer (Vercel)"
        A[Web Browser]
        C[Next.js App /api]
        F[Tracking & Admin API]
    end

    subgraph "Core Layer (VPS)"
        B[WhatsApp Bot]
        I[Cron Scheduler]
        M[whatsmeow Client]
    end

    subgraph "Shared Data Layer (Supabase)"
        K[(PostgreSQL)]
    end

    subgraph "AI Layer"
        J[Gemini AI]
    end

    A -->|HTTPS| C
    C -->|Unified Logic| F
    F -->|Direct SQL| K
    
    B -->|Socket| M
    M -->|State| K
    
    I -->|Active Jobs| K
    B -->|OCR/Parse| J
```

### Component Breakdown

#### **Edge (Frontend / Public API)**

- **Next.js & Vercel Edge**: Handles all high-traffic web requests (Tracking, Dashboard).
- **Direct Database Access**: The frontend connects directly to PostgreSQL via `postgres.js` to minimize latency.
- **Unified Routing**: APIs are nested within the frontend project to ensure 100% uptime even if the VPS is down.

#### **Core (WhatsApp Bot / Automation)**

- **Pure WhatsApp Service**: The VPS is now dedicated exclusively to the WhatsApp socket and background processing.
- **Internal Cron**: All proactive tasks (SLA recalculations, automated status alerts) are handled here to stay live with the WhatsApp connection.
- **RAM Optimized**: Stripped of its web server, the bot consumes **70% less memory** than before.

---

## ✨ Features

### 🌐 Edge Tracking Interface

- **Vercel Global Delivery** - Lightning-fast tracking lookups from any location.
- **Interactive Map** - Visual representation of shipment journey.
- **Status Timeline** - Detailed event history with timestamps.
- **Redacted Privacy** - PII (Names/Addresses) are automatically redacted in public views.
-language** - English and Portuguese support

### 🔐 Admin Portal

- **Dashboard** - Overview of active shipments and statistics
- **Shipment Management** - Create, edit, delete, and archive shipments
- **Bulk Operations** - Delete delivered shipments, archive old records
- **Search & Filter** - Find shipments by tracking number, receiver, or status
- **Manual Status Updates** - Add events and update shipment status
- **AI Manifest Parser** - Upload or paste manifest text for automatic parsing

### 📱 WhatsApp Bot Features

- **Automated Manifest Processing** - Messages following manifest patterns are parsed using a **Regex-First** hybrid strategy (AI fallback).
- **Duplicate Optimization** - Detects existing records and skips redundant image generation to save CPU/Network.
- **Error Correction (`!edit`)** - Correct mistakes on-the-fly (e.g., `!edit name Jane Doe`) with automatic receipt regeneration.
- **Premium Terminology** - Consistent use of **"Shipment Information"** across all professional communications.
- **Group Filtering** - Restrict bot activity to specific group JIDs.
- **Professional Pairing** - Security-focused HTML email delivery of WhatsApp pairing codes.
- **Status Notifications** - Automated updates for shipments in transit.

### 🤖 AI & Logic Features

- **Regex-First Hybrid Parsing**: Ultra-robust pattern matching handles 95% of manifests, minimizing AI costs. Supports **ID/Passport** extraction.
- **AI Fallback**: Gemini AI handles messy text with a pre-defined JSON schema.
- **Manifest Validation**: Automated validation of required fields (Receiver Name, Phone, Address, etc.).

### ⚙️ Automation

- **Internal Cron Management** - All recurring tasks are handled internally by the Go bot, removing the need for external triggers.

---

## 🚀 Getting Started

### Prerequisites

- **Node.js** 20+ and npm
- **Go** 1.25+
- **PostgreSQL** database (Supabase recommended)
- **WhatsApp Business Account** with Meta API access
- **Google Gemini API** key

### Frontend Setup

1. **Navigate to frontend directory**

   ```bash
   cd front
   ```

2. **Install dependencies**

   ```bash
   npm install
   ```

3. **Configure environment variables**

   ```bash
   cp .env.example .env
   ```

   Edit `.env` with your credentials:

   ```env
   # Database
   DATABASE_URL="postgresql://..."
   DIRECT_URL="postgresql://..."
   
   # Authentication
   NEXTAUTH_SECRET="your-secret-here"
   NEXTAUTH_URL="http://localhost:3000"
   ADMIN_USERNAME="admin"
   ADMIN_PASSWORD="your-password"
   ADMIN_EMAIL="admin@yourdomain.com"
   
   # AI
   GEMINI_API_KEY="your-gemini-api-key"
   ADMIN_TIMEZONE="Africa/Lagos"
   
   # Public Contact Info
   NEXT_PUBLIC_CONTACT_EMAIL="support@yourlogistics.com"
   NEXT_PUBLIC_CONTACT_PHONE="+1 (555) 000-0000"
   NEXT_PUBLIC_CONTACT_HQ="Global Logistics Center"
   ```

4. **Set up database**

   ```bash
   npx prisma generate
   npx prisma db push
   ```

5. **Run development server**

   ```bash
   npm run dev
   ```

   Open [http://localhost:3000](http://localhost:3000)

### Backend (WhatsApp Bot) Setup

1. **Navigate to backend directory**

   ```bash
   cd backend
   ```

2. **Install Go dependencies**

   ```bash
   go mod download
   ```

3. **Configure environment variables**

   ```bash
   cp .env.example .env
   ```

   Edit `.env` with your credentials:

   ```env
   # Database
   DATABASE_URL="postgresql://..."
   
   # WhatsApp
   WHATSAPP_PHONE_NUMBER="+1234567890"
   WHATSAPP_GROUP_ID=""  # Optional: restrict to specific group
   
   # Supabase (for direct HTTP API calls)
   SUPABASE_URL="https://your-project.supabase.co"
   SUPABASE_ANON_KEY="your-anon-key"
   ```

4. **Build the bot**

   ```bash
   go build -o bot.exe ./cmd/bot
   ```

5. **Run the bot**

   ```bash
   ./bot.exe
   ```

   The bot will generate an 8-character **Pairing Code** and print it in the console.
   If SMTP is configured, this code will also be sent to your `NOTIFY_EMAIL` in a professional HTML template.
   Enter this code in WhatsApp (Linked Devices > Link with Phone Number) to authenticate.

### Cron Jobs

Tracking status updates and notification retries are handled automatically by the Go bot's internal scheduler. No external configuration is required.

---

## 📦 Deployment

### Frontend (Vercel)

The frontend is configured for zero-config deployment on Vercel.

1. **Push to GitHub**
2. **Import to Vercel** (Ensure Root Directory is `front`)
3. **Deploy!**

### Backend (AWS EC2)

The backend is automatically deployed via GitHub Actions when you push to `main`.

**Prerequisites:**
Add these secrets to your GitHub Repository:

- `EC2_HOST`
- `EC2_USER`
- `EC2_SSH_KEY`

**What happens automatically:**

- Builds the Go binary.
- Copies the binary and `webtracker-bot.service` to your server.
- Sets permissions and restarts the systemd service.

### Database (Supabase)

1. Create a new Supabase project
2. Copy connection strings (pooled and direct)
3. Run Prisma migrations:

   ```bash
   npx prisma db push
   ```

---

## 🔌 API Reference

### Public Endpoints

#### `GET /api/tracking/:trackingNumber`

Get shipment details by tracking number.

**Response:**

```json
{
  "trackingNumber": "TRK-ABC123",
  "status": "In Transit",
  "senderCountry": "Nigeria",
  "receiverCountry": "Portugal",
  "receiverName": "John Doe",
  "events": [
    {
      "status": "Picked Up",
      "location": "Lagos, Nigeria",
      "timestamp": "2026-01-15T10:00:00Z"
    }
  ]
}
```

### Admin Endpoints (Protected)

#### `POST /api/admin/shipments`

Create a new shipment.

**Request:**

```json
{
  "receiverName": "John Doe",
  "receiverPhone": "+351912345678",
  "receiverCountry": "Portugal",
  "senderCountry": "Nigeria"
}
```

#### `DELETE /api/admin/shipments/:id`

Delete a shipment by ID.

### Server Actions

#### `createShipment(formData)`

Create shipment from admin form.

#### `deleteShipment(shipmentId)`

Delete a shipment.

#### `bulkDeleteDelivered()`

Delete all delivered shipments.

#### `getTracking(trackingNumber)`

Get shipment by tracking number.

---

## 🗂️ Project Structure

```
webTracker/
├── front/                      # Next.js Frontend
│   ├── src/
│   │   ├── app/               # App Router pages
│   │   │   ├── actions/       # Server actions
│   │   │   ├── admin/         # Admin portal
│   │   │   ├── api/           # API routes
│   │   │   ├── auth/          # Authentication
│   │   │   └── page.tsx       # Home/tracking page
│   │   ├── components/        # React components
│   │   ├── lib/               # Utilities and constants
│   │   ├── services/          # Business logic
│   │   └── types/             # TypeScript types
│   ├── prisma/
│   │   └── schema.prisma      # Database schema
│   └── public/                # Static assets
│
└── backend/                    # Go WhatsApp Bot
    ├── cmd/
    │   └── bot/               # Main entry point
    ├── internal/
    │   ├── commands/          # Bot command handlers
    │   ├── config/            # Configuration
    │   ├── logger/            # Zerolog + Log Rotation
    │   ├── models/            # Data models
    │   ├── notif/             # Professional Email notifications
    │   ├── parser/            # Regex-First Manifest parser
    │   ├── scheduler/         # Cron jobs
    │   ├── supabase/          # Database client
    │   ├── whatsapp/          # WhatsApp client
    │   └── worker/            # Background workers
    └── go.mod                 # Go dependencies
```

---

## 🛠️ Tech Stack

### Frontend

- **Framework**: Next.js 16 (React 19)
- **Language**: TypeScript 5
- **Styling**: Tailwind CSS 4
- **Database ORM**: Prisma 5
- **Authentication**: NextAuth 4
- **Maps**: Leaflet.js + React-Leaflet
- **Animations**: Framer Motion
- **Icons**: Lucide React
- **Validation**: Zod
- **Testing**: Vitest + React Testing Library

### Backend

- **Language**: Go 1.25
- **WhatsApp**: whatsmeow
- **Database**: pgx/v5 (PostgreSQL driver)
- **Logging**: zerolog + lumberjack
- **Scheduling**: robfig/cron
- **UUID**: google/uuid

### Infrastructure

- **Database**: PostgreSQL (Supabase)
- **Hosting**: Vercel (Frontend), VPS (Backend)

- **AI**: Google Gemini API
- **Version Control**: Git + GitHub

---

## 📝 Environment Variables

### Frontend (.env)

| Variable | Description | Required |
|----------|-------------|----------|
| `DATABASE_URL` | PostgreSQL connection string (pooled) | ✅ |
| `DIRECT_URL` | PostgreSQL direct connection | ✅ |
| `NEXTAUTH_SECRET` | NextAuth encryption secret | ✅ |
| `NEXTAUTH_URL` | Application URL | ✅ |
| `ADMIN_USERNAME` | Admin login username | ✅ |
| `ADMIN_PASSWORD` | Admin login password | ✅ |
| `ADMIN_EMAIL` | Admin email address | ✅ |
| `GEMINI_API_KEY` | Google Gemini API key | ✅ |
| `ADMIN_TIMEZONE` | Admin timezone (e.g., Africa/Lagos) | ✅ |
| `NEXT_PUBLIC_CONTACT_EMAIL` | Public contact email | ❌ |
| `NEXT_PUBLIC_CONTACT_PHONE` | Public contact phone | ❌ |
| `NEXT_PUBLIC_CONTACT_HQ` | Company headquarters location | ❌ |

### Backend (.env)

| Variable | Description | Required |
|----------|-------------|----------|
| `DATABASE_URL` | PostgreSQL connection string | ✅ |
| `WHATSAPP_PHONE_NUMBER` | Bot's WhatsApp number | ✅ |
| `WHATSAPP_GROUP_ID` | Restrict to specific group (optional) | ❌ |
| `SUPABASE_URL` | Supabase project URL | ✅ |
| `SUPABASE_ANON_KEY` | Supabase anonymous key | ✅ |

---

## 🤝 Contributing

Contributions are welcome! Please follow these steps:

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

---

## 📄 License

This project is proprietary software. All rights reserved.

---

## 🙏 Acknowledgments

- **Next.js Team** - For the amazing React framework
- **Vercel** - For seamless deployment
- **Supabase** - For PostgreSQL hosting
- **Meta** - For WhatsApp Business API
- **Google** - For Gemini AI API
- **Open Source Community** - For the incredible libraries and tools

---

## 📞 Support

For support, email `emmanuelforchinagorom@gmail.com` or join our WhatsApp group.

---

<div align="center">

**Built with ❤️ by NexGenCodes**

[Website](https://web-tracker-git-main-holy-guys-projects.vercel.app) • [Documentation](https://docs.yourlogistics.com) • [Support](mailto:emmanuelforchinagorom@gmail.com)

</div>
