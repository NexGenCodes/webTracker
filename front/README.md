# WebTracker - Shipment Tracking System

A real-time shipment tracking application with WhatsApp bot integration, automated status transitions, and intelligent notification retry system.

## 🚀 Features

- **WhatsApp Bot Integration**: Automatically create shipments from WhatsApp messages
- **Real-time Tracking**: Track shipments with live status updates and map visualization
- **Smart Notifications**: Automatic WhatsApp alerts with retry queue for failed deliveries
- **Self-Healing Status**: Automatic transition from PENDING to IN_TRANSIT after 1 hour
- **Admin Dashboard**: Manage shipments, view analytics, and control the system
- **Dual Cron System**: Vercel native + cron-job.org backup for reliability

## 📋 Prerequisites

- Node.js 18+ and npm
- Supabase account (free tier)
- Vercel account (free tier)
- Meta/WhatsApp Business API access
- Resend account for email notifications
- cron-job.org account (free)

## 🛠️ Installation

### 1. Clone and Install Dependencies

```bash
git clone <your-repo-url>
cd webTracker
npm install
```

### 2. Database Setup

#### Initialize Prisma

```bash
npx prisma generate
npx prisma migrate dev --name init
```

This creates the following tables:

- `Shipment`: Main shipment data with WhatsApp integration fields
- `Event`: Shipment status history
- `NotificationQueue`: Failed notification retry queue

### 3. Environment Variables

Create `.env` file in the root directory:

```env
# Database
DATABASE_URL="postgresql://..."

# NextAuth
NEXTAUTH_URL="http://localhost:3000"
NEXTAUTH_SECRET="generate-with-openssl-rand-base64-32"

# Admin Credentials
ADMIN_USERNAME="your-admin-username"
ADMIN_PASSWORD="your-secure-password"
ADMIN_EMAIL="admin@yourdomain.com"

# Supabase
SUPABASE_URL="https://your-project.supabase.co"
SUPABASE_SERVICE_ROLE_KEY="your-service-role-key"

# WhatsApp Business API
WHATSAPP_VERIFY_TOKEN="your-custom-verify-token"
WHATSAPP_GROUP_ID="your-whatsapp-group-id" # Optional: restrict to specific group
WHATSAPP_PHONE_NUMBER_ID="your-phone-number-id"
WHATSAPP_TOKEN="your-whatsapp-access-token"

# Email (Resend)
RESEND_API_KEY="re_..."

# Cron Authentication
CRON_SECRET="your-vercel-cron-secret"
EXTERNAL_CRON_SECRET="your-external-cron-secret"
```

### 4. Supabase Edge Function Setup

#### Deploy WhatsApp Webhook

```bash
cd supabase/functions
supabase functions deploy whatsapp-webhook --no-verify-jwt
```

#### Set Environment Variables in Supabase

```bash
supabase secrets set WHATSAPP_VERIFY_TOKEN=your-token
supabase secrets set WHATSAPP_GROUP_ID=your-group-id
supabase secrets set WHATSAPP_PHONE_NUMBER_ID=your-phone-id
supabase secrets set WHATSAPP_TOKEN=your-token
supabase secrets set SUPABASE_URL=your-url
supabase secrets set SUPABASE_SERVICE_ROLE_KEY=your-key
supabase secrets set RESEND_API_KEY=your-key
supabase secrets set ADMIN_EMAIL=your-email
```

### 5. WhatsApp Webhook Configuration

1. Go to Meta Developer Console → Your App → WhatsApp → Configuration
2. Set Webhook URL: `https://your-project.supabase.co/functions/v1/whatsapp-webhook`
3. Set Verify Token: (same as `WHATSAPP_VERIFY_TOKEN`)
4. Subscribe to `messages` webhook field

### 6. Backend Automation

**All automated tasks (status transitions, notifications, pruning) are handled by the backend bot's native scheduler.**

The backend runs the following automated tasks:

- **Status Transitions**: Every 10 minutes, PENDING shipments older than 1 hour are automatically transitioned to IN_TRANSIT
- **Notifications**: Every 2 minutes, WhatsApp notifications are sent for status changes
- **Daily Pruning**: At midnight, shipments older than 7 days are automatically deleted
- **Health Checks**: Every 5 minutes, the bot pings a health check endpoint (if configured)

No additional cron setup is required for the frontend. The frontend includes a client-side "self-healing" fallback that will transition PENDING shipments when they are viewed, but this is secondary to the backend's primary scheduler.

### 7. Deploy to Vercel

```bash
vercel
```

Set environment variables in Vercel dashboard (same as `.env` file).

## 🎨 Branding Customization

To rebrand the entire application, simply edit **one constant** in `src/lib/constants.ts`:

```typescript
// Change this value to rebrand the entire application
export const APP_NAME = "Your Company Name";
```

The tracking prefix is **automatically generated** from your company name:

- Multi-word names: First letter of each word (e.g., "Test Express" → "TEX")
- Single word names: First 3 letters (e.g., "MyCompany" → "MYC")

Changing `APP_NAME` will automatically update:

- All page titles and meta tags
- Logo and footer text
- About Us, Privacy Policy, and Terms of Service pages
- All localized content across 8 languages
- **Tracking ID format** (e.g., TEX-abc123, MYC-xyz789)

No environment variables or additional configuration needed!

## 📱 WhatsApp Message Format

Send messages to your WhatsApp bot in this format:

```
!INFO
Receiver's Name: John Doe
Receiver's Address: 123 Main St, Lagos
Receiver's Phone: +234123456789
Receiver's Country: Nigeria
Sender's Name: Jane Smith
Sender's Country: USA
```

The bot will:

1. Validate all required fields
2. Check for duplicates
3. Create a shipment with status `PENDING`
4. Reply with tracking ID
5. Auto-transition to `IN_TRANSIT` after 1 hour
6. Send WhatsApp notification on status change

## 🔄 System Workflow

### Shipment Lifecycle

1. **Creation** (via WhatsApp or Admin Dashboard)
   - Status: `PENDING`
   - Stored in `Shipment` table with `whatsappMessageId` and `whatsappFrom`

2. **Auto-Transition** (after 1 hour)
   - Cron job or self-healing in `getTracking`
   - Status: `PENDING` → `IN_TRANSIT`
   - WhatsApp notification sent

3. **Notification Retry** (if failed)
   - Failed notifications queued in `NotificationQueue`
   - Retry every 5 minutes (up to 3 attempts)
   - Successful delivery removes from queue

4. **Cleanup** (after 7 days)
   - Daily cron deletes all shipments > 7 days old
   - Includes all statuses (PENDING, IN_TRANSIT, DELIVERED)

### Idempotency Protection

The system prevents duplicate processing when multiple cron jobs fire:

- **Status Transitions**: Only process if `lastTransitionAt` > 10 minutes ago
- **Notifications**: Only send if `lastNotifiedAt` > 10 minutes ago
- **Retry Queue**: Only retry if `lastAttempt` > 5 minutes ago

## 🧪 Testing

### Run Development Server

```bash
npm run dev
```

### Test WhatsApp Webhook Locally

Use ngrok to expose local server:

```bash
ngrok http 3000
```

Update Meta webhook URL to ngrok URL.

### Simulate Cron Jobs

```bash
# Test status transition
curl -H "Authorization: Bearer YOUR_CRON_SECRET" \
  http://localhost:3000/api/cron/transition

# Test notification retry
curl -H "Authorization: Bearer YOUR_CRON_SECRET" \
  http://localhost:3000/api/cron/retry-notifications

# Test cleanup
curl -H "Authorization: Bearer YOUR_CRON_SECRET" \
  http://localhost:3000/api/cron/prune
```

## 🔧 Troubleshooting

### WhatsApp Notifications Not Sending

1. Check `WHATSAPP_TOKEN` is valid
2. Verify `WHATSAPP_PHONE_NUMBER_ID` is correct
3. Check Supabase Edge Function logs
4. Look for queued notifications in `NotificationQueue` table

### Cron Jobs Not Running

1. Verify cron-job.org jobs are enabled
2. Check authorization headers match environment variables
3. Review Vercel function logs
4. Ensure `CRON_SECRET` and `EXTERNAL_CRON_SECRET` are set

### Duplicate Notifications

1. Check `lastNotifiedAt` timestamps in database
2. Verify idempotency logic is working
3. Ensure both cron services aren't using the same secret

### Prisma Client Errors

After schema changes, always run:

```bash
npx prisma generate
npx prisma migrate dev
```

## 📊 Database Schema

### Shipment

- Core tracking data
- WhatsApp integration fields (`whatsappMessageId`, `whatsappFrom`)
- Idempotency fields (`lastNotifiedAt`, `lastTransitionAt`)

### Event

- Status change history
- Linked to shipment via `shipmentId`

### NotificationQueue

- Failed notification retry queue
- Tracks retry attempts and timing

## 🔐 Security Notes

- Never commit `.env` file
- Rotate `CRON_SECRET` and `EXTERNAL_CRON_SECRET` regularly
- Use different secrets for Vercel and external cron
- Keep `SUPABASE_SERVICE_ROLE_KEY` secure (server-side only)
- Validate all WhatsApp webhook requests

## 📝 License

MIT

## 🤝 Support

For issues or questions, please open a GitHub issue.
