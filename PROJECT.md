# ArtShop — Project Description

ArtShop is a full-stack art e-commerce marketplace where buyers discover and purchase artwork, sellers manage their storefronts and inventory, and administrators oversee the entire platform. It is built with a Go backend, React frontend, and Supabase (PostgreSQL) for data persistence and file storage.

## What the project does

ArtShop supports three user roles — buyer, seller, and admin — each with their own dashboard and set of capabilities:

- **Buyers** browse art by category, medium, and price range, add items to a persistent cart, manage wishlists, leave reviews, track orders, and receive AI-powered recommendations.
- **Sellers** create and manage product listings with image uploads, use AI (Claude API) to generate descriptions and tags, fulfill orders, and track revenue.
- **Admins** monitor platform health through a dashboard with revenue charts, manage users and roles, moderate products, and oversee all orders.

The checkout flow includes a multi-step form with shipping address, payment method selection (with card input UI), and order confirmation. Authentication uses JWT with automatic token refresh.

## Why it was built

This project was built as a portfolio piece to demonstrate full-stack development skills across a realistic, non-trivial application. It covers authentication, authorization, database design, REST API design, state management, responsive UI, file uploads, AI integration, and multi-role access control — all within a single cohesive product.

## Architecture decisions

- **Go + Chi** for the backend: chosen for performance, type safety, and clean HTTP routing without framework overhead.
- **GORM** as the ORM: provides migrations, relation preloading, and query building while staying close to SQL.
- **Zustand** for frontend state: lightweight alternative to Redux with minimal boilerplate, used for auth, cart, wishlist, products, and notifications.
- **Supabase** for database and storage: managed PostgreSQL with built-in file storage, eliminating the need to self-host either service.
- **Claude API** for AI features: generates product descriptions, suggests tags, and provides personalized recommendations based on browsing history.
- **Clean architecture** in the backend: handlers (HTTP) -> services (business logic) -> repositories (data access), with models shared across layers.
