# Billing Architecture & SaaS Pricing Strategy (Nigerian Market)

This document outlines the planned infrastructure for handling recurring billing using Nigerian-focused payment gateways (Paystack / Flutterwave), expected premium operating costs, and a realistic pricing tier structure in Naira (₦).

## 1. Gateway Selection

We will integrate **Paystack** (or Flutterwave) to process payments. 
* **Why Paystack / Flutterwave?** They natively process Naira (NGN), have incredibly high success rates for local Nigerian bank cards (including **Verve**, Mastercard, and Visa), and provide robust APIs for recurring subscription billing.
* **Webhook Architecture:** The Go backend will expose a `/api/webhooks/paystack` endpoint. When a customer pays, Paystack fires an event. The Go backend verifies the webhook signature and updates the `companies.subscription_status` column securely.

## 2. Expected Server Costs (Premium & Reliable)

Since the WhatsApp bot requires persistent memory (to keep whatsmeow sessions alive) and low-latency database queries, we must use premium hosting for maximum reliability.

| Resource | Recommended Service | Expected Monthly Cost |
| :--- | :--- | :--- |
| **Database** | Supabase Pro OR AWS RDS | ~$25 - $40 (₦37,500 - ₦60,000) for high-availability pooled connections |
| **App Hosting (VPS)** | AWS EC2 (t3.medium) or equivalent | ~$40 (₦60,000) for 4GB RAM + reliable CPU credits |
| **AI Processing** | Gemini API | ~$10 - $25 (₦15,000 - ₦37,500) depending on token volume |
| **Email Delivery** | Brevo / Resend (Paid Tier) | ~$15 - $20 (₦22,500 - ₦30,000) for reliable inbox delivery |
| **Total Base Cost** | | **~$100 - $125 (₦150,000 - ₦187,500) / month** |

*Note: This premium setup can comfortably handle ~50 logistics companies before needing significant scaling. That puts the raw server cost per company at around ₦3,000 to ₦4,000 per month.*

## 3. Realistic Pricing Tiers (₦)

To be honest and realistic, the pricing must cover the premium infrastructure costs while remaining highly affordable for local businesses compared to hiring a human customer support agent.

### Tier 1: Starter (₦10,000 / month)
* **Target:** Small Instagram vendors, independent dispatch riders.
* **Profit Margin:** ~60%
* **Features:** 
    * 1 WhatsApp Bot Instance
    * Up to 300 automated tracking updates/receipts per month
    * Standard AI responses
    * Email Notifications

### Tier 2: Professional (₦25,000 / month) - *Recommended Default*
* **Target:** Standard logistics companies, interstate couriers.
* **Profit Margin:** ~80%
* **Features:** 
    * 1 WhatsApp Bot Instance (Priority Queueing)
    * Up to 2,000 automated tracking updates/receipts per month
    * Advanced Gemini AI logic (complex parsing, custom routing rules)
    * Multi-admin group chat support

### Tier 3: Enterprise (₦75,000+ / month)
* **Target:** Large-scale cargo companies, international freight forwarders.
* **Profit Margin:** ~90%+
* **Features:**
    * Unlimited tracking updates
    * High-volume AI processing
    * API Access for their internal software
    * Dedicated technical support

## 4. Implementation Steps (When Ready)
1. **Database Update:** Add `paystack_customer_id`, `paystack_subscription_id`, and `plan_id` to the `companies` table.
2. **Go API:** Create `POST /api/billing/subscribe` to generate a Paystack Checkout URL.
3. **Go Webhook:** Implement `POST /api/webhooks/paystack` to listen for subscription renewals/cancellations.
4. **Next.js UI:** Create a `/dashboard/billing` page with pricing cards displaying the NGN prices, redirecting to the Paystack modal.

---
*Note: This phase is currently paused while the core multi-tenant engine, receipt formatting, and email delivery pipelines are thoroughly verified.*

## 5. Future SaaS Upgrades (Post-Launch)
Based on the current backend architecture, these are the next highly-valuable features to implement to increase conversion and retention:

1. **Team Members / Dispatch Rider Roles (RBAC)**
   - **Why:** Business owners do not want to share admin credentials with dispatch riders.
   - **How:** Add a "Staff" role tied to a `company_id`. Staff can log in on mobile and update shipment statuses (e.g., mark as "Delivered") without having access to billing or settings.

2. **API Keys & Webhooks**
   - **Why:** Required for "Scale" tier customers to integrate directly with their Shopify/WooCommerce stores.
   - **How:** Add an API Settings page to generate tokens. The backend already handles `/api/admin/shipments` via JWT, so introducing an API Key auth middleware alongside it is straightforward. Implement a webhook dispatch service to send out events.

3. **Custom Domains / CNAME Mapping**
   - **Why:** The ultimate "white-label" feature for Pro users. Instead of `webtracker.com/track`, they want `track.their-company.com`.
   - **How:** Use Vercel's Domains API to allow users to map their own subdomains to the tracking portal.

4. **Analytics Dashboard**
   - **Why:** Logistics owners need to visualize their volume and transit times.
   - **How:** Upgrade the basic `GetStats` counts into time-series data to render charts on the frontend.
