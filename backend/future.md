# 🚀 Technical Roadmap & The Honest Truth

This document dives deep into the current architecture of `webTracker`, identifies the inevitable bottlenecks of a growing logistics company, and maps out the transition to a global, enterprise-grade engine.

---

## 1. The Bottlenecks (Honest Analysis)

The current system is a **High-Performance MVP**. It is designed for speed and low cost, but it has inherent limits that will break at scale.

### 🆘 SQLite Concurrency (The Writing Wall)

* **Current State**: We use SQLite. It's fast, but it only allows **one writer at a time**.
* **The Breakpoint**: When you have 50+ users typing manifests simultaneously, the database will experience "Locked" errors.
* **Honest Solution**: Migrate to **PostgreSQL**. However, for a 1GB RAM VPS, SQLite is actually the *correct* choice until you hit high volume.

### 🆘 Synchronous Image Rendering (The IO Lag)

* **Current State**: When a user sends a manifest, the worker *stops* everything to draw the PNG receipt.
* **The Breakpoint**: Drawing a receipt takes ~200-500ms CPU time. If 10 people send a manifest at once, the 10th person waits several seconds for a response.
* **Honest Solution**: Move to an **Internal Go Channel Queue** (see section 2).

### 🆘 The "Predictive" Blind Spot (Reality Gap)

* **Current State**: The system *assumes* the package is moving because the clock says it's 8:00 AM.
* **The Breakpoint**: If a truck breaks down or a flight is canceled, your system is technically **lying** to the customer until you manually edit it.
* **Honest Solution**: Transition to an **Event-Driven Architecture** where physical scans trigger status changes.

---

## 2. Optimization for Low-Resource Environments (AWS Free Tier)

Since the system is currently deployed on an **AWS Free Tier (1GB RAM)**, we must prioritize **Memory-First Engineering**.

### 🛠️ The "Lean Queue" Strategy

Instead of installing Redis or RabbitMQ (which would eat 20% of your RAM), we use **Standard Go Channels**.

* **Benefit**: This allows us to set a `CONCURRENCY_LIMIT = 1`.
* **Why?**: This ensures that no matter how many people message the bot, it only ever renders **one** image at a time. This prevents RAM "spikes" that would otherwise crash your 1GB VPS.

### 🛠️ SQLite: The Hidden Powerhouse

While PostgreSQL is standard for enterprise, it is a memory hog.

* **The Choice**: Stay on **SQLite**. It uses 0 RAM when the bot is idle. For a single-VPS deployment, it is the most stable choice for low-memory hardware.

---

## 3. Phase 2: Intelligence & Visibility

### 🧠 Vision Layer (OCR)

Don't make users type. Let them snap a photo of a handwritten manifest.

* **Tech Stack**: Integrate Google Vision API.
* **Benefit**: Reduces user friction and typing errors to near zero.

### 🎛️ The Command Center (Admin Dashboard)

You are currently flying "blind" inside terminal logs.

* **Requirement**: A React/Next.js dashboard.
* **Features**:
  * **Live Map**: See where all "In Transit" packages are supposed to be.
  * **Revenue Analytics**: Track costs and profit margins.

---

## 4. The "Honest Truth" Summary

| Feature | Current MVP | Enterprise Standard |
| :--- | :--- | :--- |
| **Trust Model** | "Trust the System's Schedule" | "Trust the Physical Scan" |
| **Database** | SQLite (Single-user lean) | PostgreSQL (Multi-user scale) |
| **Parsing** | Regex + AI Fallback | OCR + LLM Extraction |
| **Status Update** | Automated (Predictive) | Manual (Event-Triggered) |

### 💡 Final Thought

The current system is perfect for **Operation Bootstrap**. It gives you the professional look of a big player without the $50k/month engineering bill. However, the moment your team grows beyond 1-2 dispatchers, the **Internal Queue** and **Admin Dashboard** move from "nice to have" to "mission critical."
