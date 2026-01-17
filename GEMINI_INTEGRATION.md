# Gemini AI Integration Guide

## Overview

This guide explains how to integrate Google Gemini 2.0 Flash for intelligent shipment data parsing in your WhatsApp logistics bot.

## What's Been Added

### 1. API Route: `/api/parse-shipment`

**File**: `src/app/api/parse-shipment/route.ts`

This endpoint uses Gemini 2.0 Flash to extract shipping information from natural language messages.

**Features**:

- Extracts sender and receiver information from unstructured text
- Validates that all required fields are present
- Returns helpful correction messages if data is incomplete
- Uses JSON mode for reliable structured output

**Request**:

```json
POST /api/parse-shipment
{
  "message": "Send package from John in USA to Mary at +234123456789, 123 Lagos St, Nigeria"
}
```

**Response (SUCCESS)**:

```json
{
  "status": "SUCCESS",
  "sender": {
    "name": "John",
    "country": "USA"
  },
  "receiver": {
    "name": "Mary",
    "phone": "+234123456789",
    "address": "123 Lagos St",
    "country": "Nigeria"
  },
  "correction": null
}
```

**Response (INCOMPLETE)**:

```json
{
  "status": "INCOMPLETE",
  "sender": {
    "name": "Jane",
    "country": "UK"
  },
  "receiver": {
    "name": "Bob",
    "phone": null,
    "address": null,
    "country": "Nigeria"
  },
  "correction": "Missing required information: receiver's phone number and address. Please provide the complete delivery address and contact number."
}
```

### 2. Cron Route: `/api/cron/status-sync`

**File**: `src/app/api/cron/status-sync/route.ts`

Secure endpoint that syncs PENDING shipments to IN_TRANSIT and sends WhatsApp notifications.

**Features**:

- Dual authentication (Vercel + external cron)
- Updates all PENDING shipments to IN_TRANSIT
- Creates event history
- Sends WhatsApp notifications with tracking updates
- Returns detailed sync report

## Setup Instructions

### 1. Install Dependencies

Run in PowerShell as Administrator:

```powershell
Set-ExecutionPolicy -ExecutionPolicy RemoteSigned -Scope CurrentUser
```

Then install the package:

```bash
npm install @google/generative-ai
```

### 2. Get Gemini API Key

1. Go to <https://aistudio.google.com/app/apikey>
2. Click "Create API Key"
3. Copy the key
4. Add to `.env`:

```env
GEMINI_API_KEY=your-api-key-here
```

### 3. Update WhatsApp Webhook (Optional)

If you want to use Gemini parsing in your WhatsApp bot, update the Supabase Edge Function:

```typescript
// In supabase/functions/whatsapp-webhook/index.ts

// Instead of regex parsing, call your Next.js API
const parseResponse = await fetch('https://your-app.vercel.app/api/parse-shipment', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ message: body })
});

const parsedData = await parseResponse.json();

if (parsedData.status === 'INCOMPLETE') {
    // Send correction message to user
    await sendWhatsAppMessage(from, parsedData.correction);
    return;
}

// Continue with SUCCESS data
const { sender, receiver } = parsedData;
```

### 4. Configure Cron Job

Add to cron-job.org:

**Job: Status Sync**

- URL: `https://your-app.vercel.app/api/cron/status-sync`
- Schedule: `0 * * * *` (hourly)
- Header: `Authorization: Bearer YOUR_EXTERNAL_CRON_SECRET`

## Cost Analysis

| Service | Free Tier | Usage |
|---------|-----------|-------|
| Gemini 2.0 Flash | 1,500 requests/day | ~50 messages/day = FREE |
| Vercel | 100GB bandwidth | API calls = FREE |
| Supabase | 500MB database | Small tracking data = FREE |
| WhatsApp | 1,000 conversations/month | 8-member group = FREE |
| cron-job.org | Unlimited | FREE |

**Total Cost**: $0/month

## System Prompt

The Gemini system prompt is designed to:

1. Extract structured data from natural language
2. Validate all required fields are present
3. Provide helpful correction messages
4. Return consistent JSON format

**Key Features**:

- Temperature: 0.1 (low randomness for consistency)
- Response format: JSON mode (guaranteed valid JSON)
- Validation: Checks for all required fields
- Error handling: Provides specific missing field feedback

## Testing

### Test the Parse Endpoint

```bash
curl -X POST http://localhost:3000/api/parse-shipment \
  -H "Content-Type: application/json" \
  -d '{"message": "Ship from John in USA to Mary at +234123456789, 123 Lagos St, Nigeria"}'
```

### Test Status Sync

```bash
curl -H "Authorization: Bearer YOUR_CRON_SECRET" \
  http://localhost:3000/api/cron/status-sync
```

## Integration with Existing System

The Gemini parsing can be integrated into your existing workflow:

1. **WhatsApp Message Received** → Parse with Gemini
2. **If INCOMPLETE** → Send correction message
3. **If SUCCESS** → Create shipment with PENDING status
4. **Hourly Cron** → Sync PENDING → IN_TRANSIT
5. **Send Notification** → WhatsApp update to sender

## Troubleshooting

### Gemini API Errors

- **401 Unauthorized**: Check `GEMINI_API_KEY` is correct
- **429 Rate Limit**: Free tier limit reached (1,500/day)
- **Invalid JSON**: Check system prompt and response format

### Status Sync Issues

- **No shipments updated**: Check database for PENDING shipments
- **Notifications not sent**: Verify `WHATSAPP_TOKEN` and `WHATSAPP_PHONE_NUMBER_ID`
- **Unauthorized**: Check `CRON_SECRET` or `EXTERNAL_CRON_SECRET`

## Next Steps

1. Install `@google/generative-ai` package
2. Get Gemini API key from Google AI Studio
3. Test `/api/parse-shipment` endpoint locally
4. Optionally integrate into WhatsApp webhook
5. Set up `/api/cron/status-sync` in cron-job.org
6. Monitor Gemini usage in Google AI Studio dashboard

## Benefits Over Regex Parsing

| Feature | Regex | Gemini AI |
|---------|-------|-----------|
| Flexibility | Strict format required | Natural language |
| Error messages | Generic | Specific, helpful |
| Field extraction | Manual patterns | Intelligent understanding |
| Maintenance | Update regex for changes | Self-adapting |
| User experience | Frustrating | Friendly |

The AI-powered approach makes your bot much more user-friendly and reduces errors!
