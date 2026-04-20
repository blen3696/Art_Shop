<p align="center">
  <img src="frontend/src/assets/logo.png" alt="ArtShop" width="200" />
</p>

<h1 align="center">ArtShop</h1>

<p align="center">
  A full-stack art marketplace built with <strong>Go</strong>, <strong>React</strong>, and <strong>Supabase</strong>.<br/>
  Buyers discover art. Sellers manage storefronts. Admins oversee the platform.
</p>

<p align="center">
  <img src="https://img.shields.io/badge/Go-1.22-00ADD8?logo=go&logoColor=white" alt="Go" />
  <img src="https://img.shields.io/badge/React-19-61DAFB?logo=react&logoColor=black" alt="React" />
  <img src="https://img.shields.io/badge/TypeScript-5.7-3178C6?logo=typescript&logoColor=white" alt="TypeScript" />
  <img src="https://img.shields.io/badge/Tailwind-4.1-06B6D4?logo=tailwindcss&logoColor=white" alt="Tailwind" />
  <img src="https://img.shields.io/badge/Supabase-PostgreSQL-3FCF8E?logo=supabase&logoColor=white" alt="Supabase" />
  <img src="https://img.shields.io/badge/License-MIT-green" alt="License" />
</p>

---

## Screenshots



<div align="center">
  <table>
    <tr>
      <td align="center">
        <img src="frontend/src/assets/screenshots/01-home.png" alt="Home" width="800" /><br/>
        <strong>Home</strong>
      </td>
      <td align="center">
        <img src="frontend/src/assets/screenshots/02-products.png" alt="Products" width="800" /><br/>
        <strong>Products</strong>
      </td>
      <!-- <td align="center">
        <img src="frontend/src/assets/screenshots/03-product-detail.png" alt="Product Detail" width="800" /><br/>
        <strong>Product Detail</strong>
      </td>
      <td align="center">
        <img src="frontend/src/assets/screenshots/04-cart.png" alt="Cart" width="800" /><br/>
        <strong>Cart</strong>
      </td>
      <td align="center">
        <img src="frontend/src/assets/screenshots/05-checkout.png" alt="Checkout" width="800" /><br/>
        <strong>Checkout</strong>
      </td>
      <td align="center">
        <img src="frontend/src/assets/screenshots/06-seller-dashboard.png" alt="Seller Dashboard" width="800" /><br/>
        <strong>Seller Dashboard</strong>
      </td>
      <td align="center">
        <img src="frontend/src/assets/screenshots/07-admin-dashboard.png" alt="Admin Dashboard" width="800" /><br/>
        <strong>Admin Dashboard</strong>
      </td> -->
    </tr>
  </table>
</div>

## Features

<table>
<tr>
<td width="33%" valign="top">

### Buyers
- Browse & search with filters
- Product reviews & ratings
- Wishlist & persistent cart
- Checkout with card input
- Order tracking
- AI-powered recommendations

</td>
<td width="33%" valign="top">

### Sellers
- Seller dashboard & analytics
- Product CRUD with image upload
- AI-generated descriptions & tags
- Order management & fulfillment
- Revenue tracking

</td>
<td width="33%" valign="top">

### Admins
- Platform analytics dashboard
- User management & roles
- Product moderation
- Order oversight
- Monthly revenue charts

</td>
</tr>
</table>

---

## Tech Stack

| Layer | Technology |
|:------|:-----------|
| **Frontend** | React 19, TypeScript, Vite, Tailwind CSS v4 |
| **State** | Zustand |
| **Routing** | React Router v7 |
| **Backend** | Go 1.22, Chi Router v5 |
| **ORM** | GORM |
| **Database** | PostgreSQL (Supabase) |
| **Auth** | JWT with refresh tokens |
| **AI** | Anthropic Claude API |
| **Storage** | Supabase Storage |
| **Deploy** | Docker (multi-stage, distroless) |

---

## Architecture

```
artshop/
├── backend/                          # Go REST API
│   ├── cmd/api/main.go               # Entry point & route wiring
│   ├── internal/
│   │   ├── config/                   # Environment configuration
│   │   ├── database/                 # Database connection
│   │   ├── models/                   # GORM models & DTOs
│   │   ├── repository/              # Data access layer
│   │   ├── services/                # Business logic
│   │   ├── handlers/                # HTTP handlers
│   │   └── middleware/              # Auth, rate limit, logging
│   ├── pkg/
│   │   ├── response/                # Standardized API responses
│   │   └── utils/                   # JWT utilities
│   ├── migrations/                  # SQL schema & seed data
│   ├── Dockerfile                   # Multi-stage build
│   └── Makefile
│
├── frontend/                         # React SPA
│   └── src/
│       ├── lib/                     # API client, types, constants
│       ├── store/                   # Zustand state management
│       ├── components/              # Reusable UI components
│       └── pages/                   # Route pages
│           ├── admin/               # Admin dashboard
│           ├── seller/              # Seller dashboard
│           └── ...                  # Buyer-facing pages
│
└── README.md
```

The backend follows clean architecture: **Handlers** (HTTP) → **Services** (business logic) → **Repositories** (data access). Each layer only depends on the one below it.

---

## Getting Started

### Prerequisites

- [Node.js](https://nodejs.org/) v18+
- [Go](https://golang.org/) v1.22+
- [Supabase](https://supabase.com/) account (free tier works)
- [Anthropic API key](https://console.anthropic.com/) (optional, for AI features)

### 1. Database Setup

1. Create a project at [supabase.com](https://supabase.com)
2. Open **SQL Editor** and run `backend/migrations/001_initial_schema.sql`
3. Optionally run `backend/migrations/002_seed_data.sql` for demo data
4. Copy your connection string from **Settings > Database**

### 2. Backend

```bash
cd backend
cp .env.example .env
# Fill in your Supabase credentials, JWT secret, and optionally Anthropic API key
go mod tidy
make run
```

API starts at `http://localhost:8080`

### 3. Frontend

```bash
cd frontend
cp .env.example .env
npm install
npm run dev
```

App starts at `http://localhost:5173`

### 4. Demo Credentials

| Role | Email | Password |
|:-----|:------|:---------|
| Admin | `admin@artshop.com` | `Admin@123456` |

> Demo sellers and buyers are included in the seed data.

---

## API Reference

<details>
<summary><strong>Authentication</strong></summary>

| Method | Endpoint | Auth | Description |
|:-------|:---------|:-----|:------------|
| `POST` | `/api/auth/register` | - | Register new user |
| `POST` | `/api/auth/login` | - | Login |
| `POST` | `/api/auth/refresh` | - | Refresh token pair |
| `GET` | `/api/auth/me` | Bearer | Current user profile |
| `PUT` | `/api/auth/profile` | Bearer | Update profile |
| `POST` | `/api/auth/change-password` | Bearer | Change password |
| `POST` | `/api/auth/register-seller` | Bearer | Upgrade to seller |

</details>

<details>
<summary><strong>Products</strong></summary>

| Method | Endpoint | Auth | Description |
|:-------|:---------|:-----|:------------|
| `GET` | `/api/products` | - | List with filters & pagination |
| `GET` | `/api/products/featured` | - | Featured products |
| `GET` | `/api/products/:id` | - | Product detail |
| `GET` | `/api/categories` | - | All categories |
| `POST` | `/api/products` | Seller | Create product |
| `PUT` | `/api/products/:id` | Seller | Update product |
| `DELETE` | `/api/products/:id` | Seller | Delete product |

</details>

<details>
<summary><strong>Cart & Wishlist</strong></summary>

| Method | Endpoint | Auth | Description |
|:-------|:---------|:-----|:------------|
| `GET` | `/api/cart` | Bearer | Get cart |
| `POST` | `/api/cart` | Bearer | Add item |
| `PUT` | `/api/cart/:productId` | Bearer | Update quantity |
| `DELETE` | `/api/cart/:productId` | Bearer | Remove item |
| `DELETE` | `/api/cart` | Bearer | Clear cart |
| `GET` | `/api/wishlist` | Bearer | Get wishlist |
| `POST` | `/api/wishlist` | Bearer | Add to wishlist |
| `DELETE` | `/api/wishlist/:productId` | Bearer | Remove from wishlist |

</details>

<details>
<summary><strong>Orders</strong></summary>

| Method | Endpoint | Auth | Description |
|:-------|:---------|:-----|:------------|
| `POST` | `/api/orders` | Bearer | Create order from cart |
| `GET` | `/api/orders` | Bearer | User's orders |
| `GET` | `/api/orders/:id` | Bearer | Order detail |
| `PUT` | `/api/orders/:id/status` | Bearer | Update status |
| `GET` | `/api/seller/orders` | Seller | Seller's orders |

</details>

<details>
<summary><strong>Reviews</strong></summary>

| Method | Endpoint | Auth | Description |
|:-------|:---------|:-----|:------------|
| `GET` | `/api/products/:id/reviews` | - | Product reviews |
| `POST` | `/api/products/:id/reviews` | Bearer | Create review |
| `DELETE` | `/api/reviews/:id` | Bearer | Delete review |

</details>

<details>
<summary><strong>Admin</strong></summary>

| Method | Endpoint | Auth | Description |
|:-------|:---------|:-----|:------------|
| `GET` | `/api/admin/dashboard` | Admin | Dashboard stats |
| `GET` | `/api/admin/users` | Admin | List users |
| `PUT` | `/api/admin/users/:id/role` | Admin | Change role |
| `PUT` | `/api/admin/users/:id/toggle-active` | Admin | Toggle active |
| `GET` | `/api/admin/orders` | Admin | All orders |
| `PUT` | `/api/admin/products/:id/toggle-featured` | Admin | Toggle featured |
| `GET` | `/api/admin/revenue` | Admin | Revenue data |

</details>

<details>
<summary><strong>AI</strong></summary>

| Method | Endpoint | Auth | Description |
|:-------|:---------|:-----|:------------|
| `POST` | `/api/ai/generate-description` | Seller | AI product description |
| `POST` | `/api/ai/generate-tags` | Seller | AI tag suggestions |
| `GET` | `/api/ai/recommendations` | Bearer | Personalized recommendations |

</details>

<details>
<summary><strong>Other</strong></summary>

| Method | Endpoint | Auth | Description |
|:-------|:---------|:-----|:------------|
| `POST` | `/api/upload` | Bearer | Upload image |
| `GET` | `/api/notifications` | Bearer | List notifications |
| `GET` | `/api/notifications/unread-count` | Bearer | Unread count |
| `PUT` | `/api/notifications/:id/read` | Bearer | Mark read |
| `PUT` | `/api/notifications/read-all` | Bearer | Mark all read |

</details>

---

## Deployment

### Docker (Backend)

```bash
cd backend
docker build -t artshop-api .
docker run --env-file .env -p 8080:8080 artshop-api
```

### Frontend

```bash
cd frontend
VITE_API_URL=https://your-api-url.com/api npm run build
# Deploy dist/ to Vercel, Netlify, or any static host
```

---

## License

This project is licensed under the [MIT License](LICENSE).
