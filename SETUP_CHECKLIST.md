# System Integration Verification Checklist

## ✅ Database Schema

- [x] `Shipment` model with all required fields
  - [x] WhatsApp fields: `whatsappMessageId`, `whatsappFrom`
  - [x] Idempotency fields: `lastNotifiedAt`, `lastTransitionAt`
- [x] `Event` model for status history
- [x] `NotificationQueue` model for retry queue
- [ ] **ACTION REQUIRED**: Run `npx prisma migrate dev` to apply schema changes
- [ ] **ACTION REQUIRED**: Run `npx prisma generate` to update Prisma client

## ✅ Environment Variables

### Required in `.env` (Local Development)

- [x] `DATABASE_URL` - Supabase connection string
- [x] `DIRECT_URL` - Direct database connection
- [x] `NEXTAUTH_SECRET` - NextAuth secret key
- [x] `NEXTAUTH_URL` - Application URL
- [x] `ADMIN_USERNAME` - Admin login username
- [x] `ADMIN_PASSWORD` - Admin login password
- [x] `ADMIN_EMAIL` - Admin email address
- [x] `SUPABASE_URL` - Supabase project URL
- [ ] **ACTION REQUIRED**: `SUPABASE_SERVICE_ROLE_KEY` - Get from Supabase Dashboard → Settings → API
- [ ] **ACTION REQUIRED**: `WHATSAPP_VERIFY_TOKEN` - Create a custom token for webhook verification
- [ ] **ACTION REQUIRED**: `WHATSAPP_GROUP_ID` - Get from WhatsApp group (optional)
- [ ] **ACTION REQUIRED**: `WHATSAPP_PHONE_NUMBER_ID` - Get from Meta Developer Console
- [ ] **ACTION REQUIRED**: `WHATSAPP_TOKEN` - Get from Meta Developer Console
- [ ] **ACTION REQUIRED**: `RESEND_API_KEY` - Get from <https://resend.com/api-keys>
- [x] `CRON_SECRET` - For Vercel native cron
- [ ] **ACTION REQUIRED**: `EXTERNAL_CRON_SECRET` - Generate with `openssl rand -base64 32`

### Required in Vercel (Production)

- [ ] **ACTION REQUIRED**: Set all above variables in Vercel Dashboard → Settings → Environment Variables

### Required in Supabase (Edge Functions)

- [ ] **ACTION REQUIRED**: Set WhatsApp variables using `supabase secrets set`
- [ ] **ACTION REQUIRED**: Set Supabase variables
- [ ] **ACTION REQUIRED**: Set Resend API key
- [ ] **ACTION REQUIRED**: Set admin email

## ✅ API Routes

- [x] `/api/cron/transition` - Hourly status transitions
  - [x] Dual authentication support (Vercel + external)
  - [x] Calls `autoTransitionShipments()`
- [x] `/api/cron/retry-notifications` - 5-minute notification retries
  - [x] Dual authentication support (Vercel + external)
  - [x] Calls `processNotificationQueue()`
- [x] `/api/cron/prune` - Daily cleanup
  - [x] Calls `pruneOldShipments()`

## ✅ Server Actions

- [x] `sendWhatsAppNotification()` - Sends WhatsApp messages with retry queue
- [x] `autoTransitionShipments()` - Transitions PENDING → IN_TRANSIT after 1 hour
- [x] `processNotificationQueue()` - Retries failed notifications
- [x] `pruneOldShipments()` - Deletes shipments > 7 days old
- [x] `getTracking()` - Self-healing status transitions

## ⚠️ Supabase Edge Function

- [x] `whatsapp-webhook` function created
- [ ] **ACTION REQUIRED**: Deploy with `supabase functions deploy whatsapp-webhook --no-verify-jwt`
- [ ] **ACTION REQUIRED**: Set environment secrets in Supabase
- [ ] **ACTION REQUIRED**: Configure Meta webhook URL

## ⚠️ WhatsApp Configuration

- [ ] **ACTION REQUIRED**: Go to Meta Developer Console → Your App → WhatsApp
- [ ] **ACTION REQUIRED**: Set Webhook URL: `https://your-project.supabase.co/functions/v1/whatsapp-webhook`
- [ ] **ACTION REQUIRED**: Set Verify Token (same as `WHATSAPP_VERIFY_TOKEN`)
- [ ] **ACTION REQUIRED**: Subscribe to `messages` webhook field

## ⚠️ Cron Job Setup (cron-job.org)

- [ ] **ACTION REQUIRED**: Sign up at <https://cron-job.org>
- [ ] **ACTION REQUIRED**: Create Job 1 - Hourly Transitions
  - URL: `https://your-app.vercel.app/api/cron/transition`
  - Schedule: `0 * * * *`
  - Header: `Authorization: Bearer YOUR_EXTERNAL_CRON_SECRET`
- [ ] **ACTION REQUIRED**: Create Job 2 - Notification Retries
  - URL: `https://your-app.vercel.app/api/cron/retry-notifications`
  - Schedule: `*/5 * * * *`
  - Header: `Authorization: Bearer YOUR_EXTERNAL_CRON_SECRET`
- [ ] **ACTION REQUIRED**: Create Job 3 - Daily Cleanup
  - URL: `https://your-app.vercel.app/api/cron/prune`
  - Schedule: `0 0 * * *`
  - Header: `Authorization: Bearer YOUR_EXTERNAL_CRON_SECRET`

## ⚠️ Deployment

- [ ] **ACTION REQUIRED**: Deploy to Vercel with `vercel`
- [ ] **ACTION REQUIRED**: Set all environment variables in Vercel Dashboard
- [ ] **ACTION REQUIRED**: Test all endpoints after deployment

## 🧪 Testing Checklist

### WhatsApp Bot

- [ ] Send test message to WhatsApp group
- [ ] Verify shipment created with `PENDING` status
- [ ] Check database for `whatsappMessageId` and `whatsappFrom`

### Status Transitions

- [ ] Wait 1 hour or manually trigger `/api/cron/transition`
- [ ] Verify shipment moved to `IN_TRANSIT`
- [ ] Verify WhatsApp notification sent
- [ ] Check `lastTransitionAt` timestamp updated

### Notification Retry

- [ ] Simulate API failure (invalid token)
- [ ] Verify notification queued in `NotificationQueue`
- [ ] Wait 5 minutes or trigger `/api/cron/retry-notifications`
- [ ] Verify retry attempt logged

### Idempotency

- [ ] Trigger same cron endpoint twice within 10 minutes
- [ ] Verify second call skips processing (check logs)
- [ ] Verify no duplicate notifications sent

## 📊 Connection Summary

```
WhatsApp → Supabase Edge Function → Database (Shipment)
                                  ↓
                            Email (Resend)
                                  ↓
                         WhatsApp Reply

Cron (hourly) → /api/cron/transition → autoTransitionShipments()
                                     ↓
                              Update Shipment
                                     ↓
                          sendWhatsAppNotification()
                                     ↓
                          NotificationQueue (if failed)

Cron (5-min) → /api/cron/retry-notifications → processNotificationQueue()
                                              ↓
                                       Retry failed notifications

Cron (daily) → /api/cron/prune → pruneOldShipments()
                                ↓
                         Delete old records
```

## 🔧 Quick Commands

```bash
# Database
npx prisma migrate dev
npx prisma generate
npx prisma studio  # View database

# Development
npm run dev

# Deployment
vercel
supabase functions deploy whatsapp-webhook --no-verify-jwt

# Testing
curl -H "Authorization: Bearer YOUR_CRON_SECRET" http://localhost:3000/api/cron/transition
```
