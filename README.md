# ArtShop - Production Art E-Commerce Platform

A full-stack, production-grade e-commerce platform for buying and selling art. Built with a React TypeScript frontend and a Go backend, powered by Supabase (PostgreSQL).

## Architecture

```
ArtShop/
├── ArtShop_Frontend/     # React 19 + TypeScript + Vite + Tailwind CSS
├── backend/              # Go + Chi Router + GORM + Supabase
└── README.md
```

## Features

### For Buyers
- Browse and search art products with filters (category, price range, medium, sort)
- Full-text search with fuzzy matching
- Product reviews and ratings
- Wishlist to save favorite pieces
- Shopping cart with server-side persistence
- Full checkout flow with multiple payment methods
- Order tracking with status updates
- AI-powered personalized recommendations

### For Sellers
- Register as a seller from any buyer account
- Seller dashboard with sales analytics
- Product CRUD with image upload
- AI-generated product descriptions and tag suggestions (Claude API)
- Order management with shipping tracking
- Revenue tracking

### For Admins
- Admin dashboard with platform analytics
- User management (roles, activation)
- Product moderation (feature/unfeature)
- Order oversight across all sellers
- Monthly revenue charts

### Technical Features
- JWT authentication with refresh tokens
- Role-based access control (buyer/seller/admin)
- Server-side cart that syncs across devices
- Rate limiting per IP
- Structured logging
- Full-text search with PostgreSQL GIN indexes
- Optimized database queries with proper indexing
- Multi-stage Docker build (distroless runtime)
- CORS configuration for production

## Tech Stack

| Layer | Technology |
|-------|-----------|
| Frontend | React 19, TypeScript, Vite, Tailwind CSS v4 |
| State | Zustand (client), API sync (server) |
| Routing | React Router v7 |
| Backend | Go 1.22, Chi Router v5 |
| ORM | GORM v1.25 |
| Database | Supabase (PostgreSQL) |
| Auth | JWT (access + refresh tokens) |
| AI | Anthropic Claude API |
| Storage | Supabase Storage |
| Deploy | Docker, any container platform |

## Prerequisites

- [Node.js](https://nodejs.org/) v18+
- [Go](https://golang.org/) v1.22+
- [Supabase](https://supabase.com/) account (free tier works)
- [Anthropic API key](https://console.anthropic.com/) (for AI features)

## Setup

### 1. Supabase Database

1. Create a new project at [supabase.com](https://supabase.com)
2. Go to **SQL Editor** and run the migration:
   ```
   backend/migrations/001_initial_schema.sql
   ```
3. Go to **Settings > Database** and copy the connection string
4. Go to **Settings > API** and copy the URL and keys

### 2. Backend

```bash
cd backend

# Copy env file and fill in your values
cp .env.example .env

# Edit .env with your Supabase credentials:
# DATABASE_URL=postgresql://postgres:YOUR_PASSWORD@db.YOUR_REF.supabase.co:5432/postgres
# JWT_SECRET=your-secret-key
# SUPABASE_URL=https://YOUR_REF.supabase.co
# SUPABASE_SERVICE_KEY=your-service-key
# ANTHROPIC_API_KEY=your-key (optional, for AI features)

# Install dependencies
go mod tidy

# Run the server
make run
# or: go run cmd/api/main.go
```

The API will start at `http://localhost:8080`.

### 3. Frontend

```bash
cd ArtShop_Frontend

# Copy env file
cp .env.example .env
# VITE_API_URL=http://localhost:8080/api  (default)

# Install dependencies
npm install

# Run dev server
npm run dev
```

The frontend will start at `http://localhost:5173`.

### 4. Default Admin Login

```
Email: admin@artshop.com
Password: Admin@123456
```

Change this immediately in production.

## API Endpoints

### Auth
| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | /api/auth/register | Register new user |
| POST | /api/auth/login | Login |
| POST | /api/auth/refresh | Refresh tokens |
| POST | /api/auth/register-seller | Upgrade to seller |
| GET | /api/auth/me | Get current user |

### Products
| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | /api/products | List products (with filters) |
| GET | /api/products/:id | Get product detail |
| GET | /api/products/featured | Featured products |
| GET | /api/products/search | Search products |
| GET | /api/categories | List categories |
| POST | /api/products | Create product (seller) |
| PUT | /api/products/:id | Update product (seller) |
| DELETE | /api/products/:id | Delete product (seller) |

### Cart & Wishlist
| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | /api/cart | Get cart items |
| POST | /api/cart | Add to cart |
| PUT | /api/cart/:productId | Update quantity |
| DELETE | /api/cart/:productId | Remove item |
| GET | /api/wishlist | Get wishlist |
| POST | /api/wishlist | Add to wishlist |
| DELETE | /api/wishlist/:productId | Remove from wishlist |

### Orders
| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | /api/orders | Create order from cart |
| GET | /api/orders | List user's orders |
| GET | /api/orders/:id | Get order detail |
| PUT | /api/orders/:id/status | Update status |

### Reviews
| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | /api/products/:id/reviews | Get product reviews |
| POST | /api/products/:id/reviews | Create review |
| DELETE | /api/reviews/:id | Delete review |

### Admin
| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | /api/admin/dashboard | Dashboard stats |
| GET | /api/admin/users | List users |
| PUT | /api/admin/users/:id/role | Change user role |
| GET | /api/admin/orders | All orders |
| GET | /api/admin/revenue | Revenue data |

### AI
| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | /api/ai/generate-description | AI product description |
| POST | /api/ai/generate-tags | AI tag suggestions |
| GET | /api/ai/recommendations | Personalized recs |

## Production Deployment

### Docker

```bash
cd backend
docker build -t artshop-api .
docker run --env-file .env -p 8080:8080 artshop-api
```

### Frontend Build

```bash
cd ArtShop_Frontend
npm run build
# Deploy the dist/ folder to Vercel, Netlify, or any static host
```

### Environment Variables

Set `VITE_API_URL` to your production API URL before building the frontend.

## Project Structure

### Backend (Clean Architecture)

```
backend/
├── cmd/api/main.go              # Entry point, wiring, server
├── internal/
│   ├── config/                  # Environment config loader
���   ├── database/                # GORM database connection
│   ├── models/                  # GORM models & DTOs
│   ├── middleware/               # Auth, CORS, rate limit, logger
│   ├── repository/              # Data access layer
│   ├── services/                # Business logic layer
│   └── handlers/                # HTTP handlers (controllers)
├── pkg/
│   ├── response/                # Standardized API responses
│   └── utils/                   # JWT utilities
├── migrations/                  # SQL migration files
├── Dockerfile                   # Multi-stage Docker build
├── Makefile                     # Dev commands
└── go.mod                       # Go module definition
```

### Frontend

```
ArtShop_Frontend/src/
├── lib/                         # API client, types, constants
├── store/                       # Zustand stores (auth, cart, products, wishlist, notifications)
├── hooks/                       # Custom React hooks
├── components/
│   ├── ui/                      # Reusable UI components (Button, Input, Modal, etc.)
│   ├── header/                  # Navigation header with search
│   ├── footer/                  # Site footer
│   ├── layout/                  # Page layout wrapper
��   ├── search/                  # Search bar with live results
│   ├���─ reviews/                 # Review card & form
│   ├── ai/                      # AI recommendations component
│   └── ...                      # Home page sections
├── pages/
│   ├── admin/                   # Admin dashboard (layout, users, products, orders)
│   ├── seller/                  # Seller dashboard (layout, products, orders, add product)
│   ├── search/                  # Search results page
│   ├── wishlist/                # Saved products
│   ├── settings/                # User/store settings
│   └── ...                      # Home, Product, Cart, Checkout, Auth, Profile, Orders
└── App.tsx                      # Route definitions
```

