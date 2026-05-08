# Billing Architecture & SaaS Pricing Strategy (Nigerian Market)

This document outlines the planned infrastructure for handling recurring billing using Nigerian-focused payment gateways (Paystack / Flutterwave), expected premium operating costs, and a realistic pricing tier structure in Naira (₦).

## 4. Future SaaS Upgrades (Post-Launch)

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

5. **WhatsApp CSV Bulk Uploads**
   - **Why:** Admins need to create 50-100 shipments at once without using the web dashboard.
   - **How:** Intercept `.csv` `DocumentMessage` payloads in WhatsApp, parse natively using `encoding/csv`, and loop through database insertions, bypassing standard AI parsing constraints.

6. **"Snap & Track" Image Parsing (Vision AI)**
   - **Why:** Dispatchers have handwritten box labels or physical waybills they want to log quickly without manual typing.
   - **How:** Download incoming `ImageMessage` bytes and pipe them to Gemini 1.5 Pro/Flash Vision API to extract text and structure it into a shipment manifest instantly.

7. **PDF Waybill Generation & Proof of Delivery Maps**
   - **Why:** B2B companies require official, printable receipts and exact drop-off accountability.
   - **How:** Use `gofpdf` to generate lightweight, professional PDF documents in RAM and dispatch them via WhatsApp as native `MediaDocument` messages. Leverage `LocationMessage` for live map pins on delivery.
