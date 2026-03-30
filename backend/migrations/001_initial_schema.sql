-- ============================================================================
-- ArtShop E-Commerce Platform - Initial Database Schema
-- Database: Supabase (PostgreSQL)
-- ============================================================================
-- This migration creates all tables needed for a production art e-commerce
-- platform with buyer/seller/admin roles, AI features, and full order flow.
-- ============================================================================

-- Enable required extensions
CREATE EXTENSION IF NOT EXISTS "pgcrypto";      -- For gen_random_uuid()
CREATE EXTENSION IF NOT EXISTS "pg_trgm";       -- For fuzzy text search

-- ============================================================================
-- 1. USERS - Core user table supporting buyer, seller, and admin roles
-- ============================================================================
CREATE TABLE users (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email           VARCHAR(255) UNIQUE NOT NULL,
    password_hash   VARCHAR(255) NOT NULL,
    full_name       VARCHAR(255) NOT NULL,
    role            VARCHAR(20) DEFAULT 'buyer' CHECK (role IN ('buyer', 'seller', 'admin')),
    avatar_url      TEXT,
    phone           VARCHAR(20),
    bio             TEXT,
    address_line1   VARCHAR(255),
    address_line2   VARCHAR(255),
    city            VARCHAR(100),
    state           VARCHAR(100),
    country         VARCHAR(100),
    zip_code        VARCHAR(20),
    is_verified     BOOLEAN DEFAULT FALSE,
    is_active       BOOLEAN DEFAULT TRUE,
    last_login_at   TIMESTAMPTZ,
    created_at      TIMESTAMPTZ DEFAULT NOW(),
    updated_at      TIMESTAMPTZ DEFAULT NOW()
);

-- ============================================================================
-- 2. SELLER PROFILES - Extended info for users with role='seller'
-- ============================================================================
CREATE TABLE seller_profiles (
    id                UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id           UUID UNIQUE NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    store_name        VARCHAR(255) NOT NULL,
    store_description TEXT,
    logo_url          TEXT,
    banner_url        TEXT,
    is_verified       BOOLEAN DEFAULT FALSE,
    total_sales       INTEGER DEFAULT 0,
    total_revenue     DECIMAL(12,2) DEFAULT 0,
    rating            DECIMAL(3,2) DEFAULT 0,
    commission_rate   DECIMAL(4,2) DEFAULT 10.00,  -- platform commission %
    created_at        TIMESTAMPTZ DEFAULT NOW(),
    updated_at        TIMESTAMPTZ DEFAULT NOW()
);

-- ============================================================================
-- 3. CATEGORIES - Hierarchical product categories (supports parent/child)
-- ============================================================================
CREATE TABLE categories (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name        VARCHAR(100) NOT NULL,
    slug        VARCHAR(100) UNIQUE NOT NULL,
    description TEXT,
    image_url   TEXT,
    parent_id   UUID REFERENCES categories(id) ON DELETE SET NULL,
    sort_order  INTEGER DEFAULT 0,
    is_active   BOOLEAN DEFAULT TRUE,
    created_at  TIMESTAMPTZ DEFAULT NOW()
);

-- ============================================================================
-- 4. PRODUCTS - Art products listed by sellers
-- ============================================================================
CREATE TABLE products (
    id                UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    seller_id         UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    title             VARCHAR(255) NOT NULL,
    description       TEXT,
    price             DECIMAL(10,2) NOT NULL CHECK (price >= 0),
    compare_at_price  DECIMAL(10,2),                     -- original price for sales
    category_id       UUID REFERENCES categories(id) ON DELETE SET NULL,
    images            TEXT[] DEFAULT '{}',                -- array of image URLs
    thumbnail         TEXT,
    stock             INTEGER DEFAULT 0 CHECK (stock >= 0),
    sku               VARCHAR(100),
    tags              TEXT[] DEFAULT '{}',
    medium            VARCHAR(100),                       -- oil, watercolor, digital, etc.
    dimensions        VARCHAR(100),                       -- e.g., "24x36 inches"
    weight            DECIMAL(8,2),                       -- in kg, for shipping calc
    is_published      BOOLEAN DEFAULT TRUE,
    is_featured       BOOLEAN DEFAULT FALSE,
    avg_rating        DECIMAL(3,2) DEFAULT 0,
    total_reviews     INTEGER DEFAULT 0,
    total_sales       INTEGER DEFAULT 0,
    view_count        INTEGER DEFAULT 0,
    ai_description    TEXT,                               -- AI-generated description
    ai_tags           TEXT[] DEFAULT '{}',                -- AI-generated tags
    created_at        TIMESTAMPTZ DEFAULT NOW(),
    updated_at        TIMESTAMPTZ DEFAULT NOW()
);

-- ============================================================================
-- 5. CART ITEMS - Server-side cart persistence per user
-- ============================================================================
CREATE TABLE cart_items (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id     UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    product_id  UUID NOT NULL REFERENCES products(id) ON DELETE CASCADE,
    quantity    INTEGER NOT NULL DEFAULT 1 CHECK (quantity > 0),
    created_at  TIMESTAMPTZ DEFAULT NOW(),
    UNIQUE(user_id, product_id)
);

-- ============================================================================
-- 6. WISHLISTS - Users can save products they love
-- ============================================================================
CREATE TABLE wishlists (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id     UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    product_id  UUID NOT NULL REFERENCES products(id) ON DELETE CASCADE,
    created_at  TIMESTAMPTZ DEFAULT NOW(),
    UNIQUE(user_id, product_id)
);

-- ============================================================================
-- 7. ORDERS - Purchase orders placed by buyers
-- ============================================================================
CREATE TABLE orders (
    id                      UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    buyer_id                UUID NOT NULL REFERENCES users(id),
    order_number            VARCHAR(20) UNIQUE NOT NULL,
    status                  VARCHAR(30) DEFAULT 'pending'
        CHECK (status IN ('pending','confirmed','processing','shipped','delivered','cancelled','refunded')),
    subtotal                DECIMAL(10,2) NOT NULL,
    shipping_cost           DECIMAL(10,2) DEFAULT 0,
    tax                     DECIMAL(10,2) DEFAULT 0,
    discount                DECIMAL(10,2) DEFAULT 0,
    total                   DECIMAL(10,2) NOT NULL,
    shipping_name           VARCHAR(255),
    shipping_address_line1  VARCHAR(255),
    shipping_address_line2  VARCHAR(255),
    shipping_city           VARCHAR(100),
    shipping_state          VARCHAR(100),
    shipping_country        VARCHAR(100),
    shipping_zip            VARCHAR(20),
    shipping_phone          VARCHAR(20),
    payment_method          VARCHAR(50),
    payment_status          VARCHAR(30) DEFAULT 'pending'
        CHECK (payment_status IN ('pending','paid','failed','refunded')),
    payment_intent_id       VARCHAR(255),          -- for Stripe/payment gateway
    tracking_number         VARCHAR(100),
    notes                   TEXT,
    created_at              TIMESTAMPTZ DEFAULT NOW(),
    updated_at              TIMESTAMPTZ DEFAULT NOW()
);

-- ============================================================================
-- 8. ORDER ITEMS - Individual items within an order
-- ============================================================================
CREATE TABLE order_items (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    order_id    UUID NOT NULL REFERENCES orders(id) ON DELETE CASCADE,
    product_id  UUID REFERENCES products(id) ON DELETE SET NULL,
    seller_id   UUID REFERENCES users(id) ON DELETE SET NULL,
    title       VARCHAR(255) NOT NULL,
    price       DECIMAL(10,2) NOT NULL,
    quantity    INTEGER NOT NULL CHECK (quantity > 0),
    thumbnail   TEXT,
    status      VARCHAR(30) DEFAULT 'pending'
        CHECK (status IN ('pending','confirmed','shipped','delivered','cancelled')),
    created_at  TIMESTAMPTZ DEFAULT NOW()
);

-- ============================================================================
-- 9. REVIEWS - Product reviews and ratings from verified buyers
-- ============================================================================
CREATE TABLE reviews (
    id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    product_id          UUID NOT NULL REFERENCES products(id) ON DELETE CASCADE,
    user_id             UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    order_id            UUID REFERENCES orders(id) ON DELETE SET NULL,
    rating              INTEGER NOT NULL CHECK (rating >= 1 AND rating <= 5),
    title               VARCHAR(255),
    comment             TEXT,
    images              TEXT[] DEFAULT '{}',
    is_verified_purchase BOOLEAN DEFAULT FALSE,
    helpful_count       INTEGER DEFAULT 0,
    created_at          TIMESTAMPTZ DEFAULT NOW(),
    updated_at          TIMESTAMPTZ DEFAULT NOW(),
    UNIQUE(product_id, user_id)
);

-- ============================================================================
-- 10. AI RECOMMENDATIONS - Personalized product suggestions
-- ============================================================================
CREATE TABLE ai_recommendations (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id     UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    product_id  UUID NOT NULL REFERENCES products(id) ON DELETE CASCADE,
    score       DECIMAL(5,4) NOT NULL,
    reason      TEXT,                              -- human-readable explanation
    algorithm   VARCHAR(50) DEFAULT 'collaborative',
    expires_at  TIMESTAMPTZ,
    created_at  TIMESTAMPTZ DEFAULT NOW()
);

-- ============================================================================
-- 11. BROWSING HISTORY - Tracks views for AI recommendation engine
-- ============================================================================
CREATE TABLE browsing_history (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id     UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    product_id  UUID NOT NULL REFERENCES products(id) ON DELETE CASCADE,
    viewed_at   TIMESTAMPTZ DEFAULT NOW()
);

-- ============================================================================
-- 12. NOTIFICATIONS - In-app notifications for all user types
-- ============================================================================
CREATE TABLE notifications (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id     UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    type        VARCHAR(50) NOT NULL,              -- order_update, review, promotion, system
    title       VARCHAR(255) NOT NULL,
    message     TEXT,
    is_read     BOOLEAN DEFAULT FALSE,
    action_url  TEXT,                              -- deep link for click action
    data        JSONB DEFAULT '{}',                -- flexible metadata
    created_at  TIMESTAMPTZ DEFAULT NOW()
);

-- ============================================================================
-- INDEXES - For query performance at scale
-- ============================================================================

-- User lookups
CREATE INDEX idx_users_email ON users(email);
CREATE INDEX idx_users_role ON users(role);

-- Product queries (the most frequent queries in e-commerce)
CREATE INDEX idx_products_seller ON products(seller_id);
CREATE INDEX idx_products_category ON products(category_id);
CREATE INDEX idx_products_published ON products(is_published) WHERE is_published = TRUE;
CREATE INDEX idx_products_featured ON products(is_featured) WHERE is_featured = TRUE;
CREATE INDEX idx_products_price ON products(price);
CREATE INDEX idx_products_created ON products(created_at DESC);

-- Full-text search on product title + description
CREATE INDEX idx_products_search ON products
    USING GIN(to_tsvector('english', title || ' ' || COALESCE(description, '')));

-- Trigram index for fuzzy/partial matching (e.g., "watercol" matches "watercolor")
CREATE INDEX idx_products_title_trgm ON products USING GIN(title gin_trgm_ops);

-- Order queries
CREATE INDEX idx_orders_buyer ON orders(buyer_id);
CREATE INDEX idx_orders_status ON orders(status);
CREATE INDEX idx_orders_created ON orders(created_at DESC);
CREATE INDEX idx_order_items_order ON order_items(order_id);
CREATE INDEX idx_order_items_seller ON order_items(seller_id);

-- Review queries
CREATE INDEX idx_reviews_product ON reviews(product_id);
CREATE INDEX idx_reviews_user ON reviews(user_id);

-- Cart & wishlist
CREATE INDEX idx_cart_user ON cart_items(user_id);
CREATE INDEX idx_wishlist_user ON wishlists(user_id);

-- AI & browsing
CREATE INDEX idx_recommendations_user ON ai_recommendations(user_id);
CREATE INDEX idx_browsing_user_time ON browsing_history(user_id, viewed_at DESC);

-- Notifications
CREATE INDEX idx_notifications_user_unread ON notifications(user_id, is_read) WHERE is_read = FALSE;

-- ============================================================================
-- TRIGGERS - Auto-update timestamps
-- ============================================================================
CREATE OR REPLACE FUNCTION update_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trg_users_updated_at
    BEFORE UPDATE ON users FOR EACH ROW EXECUTE FUNCTION update_updated_at();

CREATE TRIGGER trg_products_updated_at
    BEFORE UPDATE ON products FOR EACH ROW EXECUTE FUNCTION update_updated_at();

CREATE TRIGGER trg_orders_updated_at
    BEFORE UPDATE ON orders FOR EACH ROW EXECUTE FUNCTION update_updated_at();

CREATE TRIGGER trg_seller_profiles_updated_at
    BEFORE UPDATE ON seller_profiles FOR EACH ROW EXECUTE FUNCTION update_updated_at();

CREATE TRIGGER trg_reviews_updated_at
    BEFORE UPDATE ON reviews FOR EACH ROW EXECUTE FUNCTION update_updated_at();

-- ============================================================================
-- TRIGGER - Auto-update product rating when reviews change
-- ============================================================================
CREATE OR REPLACE FUNCTION update_product_rating()
RETURNS TRIGGER AS $$
BEGIN
    UPDATE products SET
        avg_rating = COALESCE((SELECT AVG(rating)::DECIMAL(3,2) FROM reviews WHERE product_id = COALESCE(NEW.product_id, OLD.product_id)), 0),
        total_reviews = (SELECT COUNT(*) FROM reviews WHERE product_id = COALESCE(NEW.product_id, OLD.product_id))
    WHERE id = COALESCE(NEW.product_id, OLD.product_id);
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trg_reviews_update_rating
    AFTER INSERT OR UPDATE OR DELETE ON reviews
    FOR EACH ROW EXECUTE FUNCTION update_product_rating();

-- ============================================================================
-- SEED DATA - Default categories and admin user
-- ============================================================================
INSERT INTO categories (name, slug, description, sort_order) VALUES
    ('Paintings',    'paintings',    'Original paintings in oil, acrylic, watercolor, and mixed media', 1),
    ('Prints',       'prints',       'High-quality art prints, giclees, and limited editions', 2),
    ('Sculptures',   'sculptures',   'Three-dimensional artworks in stone, metal, wood, and clay', 3),
    ('Digital Art',  'digital-art',  'Digital illustrations, NFT-ready art, and computer-generated pieces', 4),
    ('Photography',  'photography',  'Fine art photography prints and collections', 5),
    ('Handcraft',    'handcraft',    'Handmade artistic crafts, pottery, and decorative pieces', 6),
    ('Mixed Media',  'mixed-media',  'Artworks combining multiple materials and techniques', 7),
    ('Textile Art',  'textile-art',  'Woven, embroidered, and fabric-based artworks', 8);

-- Default admin user (password: Admin@123456 - change immediately in production!)
-- Password hash is bcrypt of 'Admin@123456'
INSERT INTO users (email, password_hash, full_name, role, is_verified) VALUES
    ('admin@artshop.com', '$2a$12$mmzq1dzNnsR9wA.pQKBUJOUYKaWKY9fQgkE/Sx1tY7PMqISrsKVd2', 'ArtShop Admin', 'admin', TRUE);
