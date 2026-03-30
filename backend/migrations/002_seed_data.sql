-- ============================================================================
-- ArtShop E-Commerce Platform - Seed Data
-- Run this AFTER 001_initial_schema.sql
-- This populates the database with realistic demo data so every UI component
-- has data to display.
-- ============================================================================

-- ============================================================================
-- USERS (1 admin from migration + 2 sellers + 5 buyers)
-- All passwords are 'Password123!' hashed with bcrypt
-- ============================================================================
-- Password hash for 'Password123!' (bcrypt cost 12)
-- $2a$12$LQv3c1yqBo9SkvXS7QTJPe0h5KjH9t0YlKGqr5W8FZdYqGpGkXHiy

-- Sellers
INSERT INTO users (id, email, password_hash, full_name, role, phone, bio, avatar_url, address_line1, city, state, country, zip_code, is_verified, is_active) VALUES
('a1111111-1111-1111-1111-111111111111', 'elena@artshop.com', '$2a$12$mmzq1dzNnsR9wA.pQKBUJOUYKaWKY9fQgkE/Sx1tY7PMqISrsKVd2', 'Elena Rodriguez', 'seller', '+1-555-0101', 'Contemporary artist specializing in vibrant oil paintings and mixed media works. Based in Barcelona, inspired by Mediterranean light.', 'https://images.unsplash.com/photo-1494790108755-2616b612d5a0?w=150', '45 Passeig de Gracia', 'Barcelona', 'Catalonia', 'Spain', '08007', true, true),
('a2222222-2222-2222-2222-222222222222', 'james@artshop.com', '$2a$12$mmzq1dzNnsR9wA.pQKBUJOUYKaWKY9fQgkE/Sx1tY7PMqISrsKVd2', 'James Chen', 'seller', '+1-555-0102', 'Digital artist and photographer exploring the intersection of technology and nature. Creating art from my studio in San Francisco.', 'https://images.unsplash.com/photo-1507003211169-0a1dd7228f2d?w=150', '88 Mission Street', 'San Francisco', 'CA', 'USA', '94105', true, true);

-- Buyers
INSERT INTO users (id, email, password_hash, full_name, role, phone, bio, avatar_url, address_line1, city, state, country, zip_code, is_verified, is_active) VALUES
('b1111111-1111-1111-1111-111111111111', 'sarah@example.com', '$2a$12$mmzq1dzNnsR9wA.pQKBUJOUYKaWKY9fQgkE/Sx1tY7PMqISrsKVd2', 'Sarah Johnson', 'buyer', '+1-555-0201', 'Art collector and interior designer. Love discovering emerging artists.', 'https://images.unsplash.com/photo-1438761681033-6461ffad8d80?w=150', '200 Park Avenue', 'New York', 'NY', 'USA', '10017', true, true),
('b2222222-2222-2222-2222-222222222222', 'michael@example.com', '$2a$12$mmzq1dzNnsR9wA.pQKBUJOUYKaWKY9fQgkE/Sx1tY7PMqISrsKVd2', 'Michael Thompson', 'buyer', '+1-555-0202', NULL, 'https://images.unsplash.com/photo-1500648767791-00dcc994a43e?w=150', '15 Oxford Street', 'London', NULL, 'UK', 'W1D 1BS', true, true),
('b3333333-3333-3333-3333-333333333333', 'aisha@example.com', '$2a$12$mmzq1dzNnsR9wA.pQKBUJOUYKaWKY9fQgkE/Sx1tY7PMqISrsKVd2', 'Aisha Patel', 'buyer', '+1-555-0203', 'Home decor enthusiast looking for unique pieces.', NULL, '42 MG Road', 'Mumbai', 'Maharashtra', 'India', '400001', true, true),
('b4444444-4444-4444-4444-444444444444', 'lucas@example.com', '$2a$12$mmzq1dzNnsR9wA.pQKBUJOUYKaWKY9fQgkE/Sx1tY7PMqISrsKVd2', 'Lucas Weber', 'buyer', '+49-555-0204', NULL, 'https://images.unsplash.com/photo-1472099645785-5658abf4ff4e?w=150', '10 Alexanderplatz', 'Berlin', NULL, 'Germany', '10178', true, true),
('b5555555-5555-5555-5555-555555555555', 'maria@example.com', '$2a$12$mmzq1dzNnsR9wA.pQKBUJOUYKaWKY9fQgkE/Sx1tY7PMqISrsKVd2', 'Maria Santos', 'buyer', '+55-555-0205', 'Passionate about supporting local and emerging artists.', 'https://images.unsplash.com/photo-1534528741775-53994a69daeb?w=150', '100 Copacabana', 'Rio de Janeiro', 'RJ', 'Brazil', '22070-000', true, true);

-- ============================================================================
-- SELLER PROFILES
-- ============================================================================
INSERT INTO seller_profiles (user_id, store_name, store_description, logo_url, banner_url, is_verified, total_sales, total_revenue, rating, commission_rate) VALUES
('a1111111-1111-1111-1111-111111111111', 'Elena Art Studio', 'Contemporary art inspired by Mediterranean warmth. Each piece tells a story of light, color, and emotion. Original paintings and limited edition prints available.', 'https://images.unsplash.com/photo-1513364776144-60967b0f800f?w=200', 'https://images.unsplash.com/photo-1541961017774-22349e4a1262?w=1200', true, 47, 15680.00, 4.8, 8.00),
('a2222222-2222-2222-2222-222222222222', 'Chen Digital Gallery', 'Where technology meets art. Digital illustrations, AI-enhanced photography, and limited NFT-ready prints. Pushing boundaries of modern art.', 'https://images.unsplash.com/photo-1561070791-2526d30994b5?w=200', 'https://images.unsplash.com/photo-1558618666-fcd25c85f82e?w=1200', true, 32, 8940.00, 4.6, 10.00);

-- ============================================================================
-- PRODUCTS (20 products across all categories)
-- Using Unsplash images for realistic art product photos
-- ============================================================================

-- Elena's Products (10 products)
INSERT INTO products (id, seller_id, title, description, price, compare_at_price, category_id, images, thumbnail, stock, sku, tags, medium, dimensions, weight, is_published, is_featured, avg_rating, total_reviews, total_sales, view_count, ai_description, ai_tags) VALUES

-- Paintings (4)
('c1111111-1111-1111-1111-111111111111', 'a1111111-1111-1111-1111-111111111111',
 'Sunset Over Barcelona',
 'A vivid oil painting capturing the golden hour over Barcelona''s Gothic Quarter. Layers of warm oranges, deep purples, and shimmering golds create a sense of peaceful majesty. Painted en plein air from the terrace of Park Guell.',
 850.00, 1100.00,
 (SELECT id FROM categories WHERE slug = 'paintings'),
 ARRAY['https://images.unsplash.com/photo-1579783902614-a3fb3927b6a5?w=800', 'https://images.unsplash.com/photo-1578301978693-85fa9c0320b9?w=800', 'https://images.unsplash.com/photo-1549887534-1541e9326642?w=800'],
 'https://images.unsplash.com/photo-1579783902614-a3fb3927b6a5?w=500',
 3, 'EL-PAINT-001',
 ARRAY['sunset', 'barcelona', 'cityscape', 'golden hour', 'oil painting', 'contemporary'],
 'Oil Paint', '36x24 inches', 2.5, true, true, 4.8, 12, 8, 342,
 'This breathtaking oil painting transports viewers to Barcelona''s enchanting Gothic Quarter during golden hour. The artist masterfully layers warm oranges and deep purples to create an atmospheric depth that captures the Mediterranean light.',
 ARRAY['impressionist', 'cityscape', 'warm tones', 'european art']),

('c1111111-1111-1111-1111-222222222222', 'a1111111-1111-1111-1111-111111111111',
 'Ocean Whispers',
 'Abstract seascape painted with palette knife technique. Deep blues and turquoise waves crash against an unseen shore, evoking the raw power and tranquility of the ocean.',
 620.00, NULL,
 (SELECT id FROM categories WHERE slug = 'paintings'),
 ARRAY['https://images.unsplash.com/photo-1549490349-8643362247b5?w=800', 'https://images.unsplash.com/photo-1500462918059-b1a0cb512f1d?w=800'],
 'https://images.unsplash.com/photo-1549490349-8643362247b5?w=500',
 5, 'EL-PAINT-002',
 ARRAY['ocean', 'abstract', 'seascape', 'blue', 'palette knife'],
 'Oil Paint', '30x40 inches', 3.0, true, true, 4.6, 8, 5, 218,
 'A powerful abstract seascape that channels the raw energy of the ocean through bold palette knife strokes. The interplay of deep blues and turquoise creates mesmerizing depth.',
 ARRAY['abstract', 'seascape', 'palette knife', 'blue tones']),

('c1111111-1111-1111-1111-333333333333', 'a1111111-1111-1111-1111-111111111111',
 'Meadow Dreams',
 'A soft, impressionist painting of a wildflower meadow in the French countryside. Delicate brushstrokes of lavender, poppy red, and sunflower yellow dance across the canvas.',
 480.00, 550.00,
 (SELECT id FROM categories WHERE slug = 'paintings'),
 ARRAY['https://images.unsplash.com/photo-1460661419201-fd4cecdf8a8b?w=800', 'https://images.unsplash.com/photo-1578301978018-3005759f48f7?w=800'],
 'https://images.unsplash.com/photo-1460661419201-fd4cecdf8a8b?w=500',
 7, 'EL-PAINT-003',
 ARRAY['meadow', 'flowers', 'impressionist', 'countryside', 'nature'],
 'Acrylic', '24x18 inches', 1.8, true, false, 4.9, 6, 4, 156,
 NULL, ARRAY[]::TEXT[]),

('c1111111-1111-1111-1111-444444444444', 'a1111111-1111-1111-1111-111111111111',
 'Portrait of Solitude',
 'A contemplative figurative painting depicting a woman gazing through a rain-streaked window. Muted earth tones and soft focus create an intimate, reflective atmosphere.',
 1200.00, NULL,
 (SELECT id FROM categories WHERE slug = 'paintings'),
 ARRAY['https://images.unsplash.com/photo-1578662996442-48f60103fc96?w=800'],
 'https://images.unsplash.com/photo-1578662996442-48f60103fc96?w=500',
 1, 'EL-PAINT-004',
 ARRAY['portrait', 'figurative', 'contemplative', 'rainy day', 'woman'],
 'Oil Paint', '40x30 inches', 4.0, true, true, 5.0, 3, 2, 89,
 'An intimate figurative masterpiece that captures the beauty of solitary reflection. The rain-streaked window serves as both a literal and metaphorical barrier between the subject and the world beyond.',
 ARRAY['figurative', 'portrait', 'emotional', 'earth tones']),

-- Prints (2)
('c1111111-1111-1111-1111-555555555555', 'a1111111-1111-1111-1111-111111111111',
 'Barcelona Skyline - Limited Edition Print',
 'Giclée print on archival cotton rag paper. Limited edition of 50, signed and numbered. Reproduces the rich colors and textures of the original oil painting.',
 120.00, 180.00,
 (SELECT id FROM categories WHERE slug = 'prints'),
 ARRAY['https://images.unsplash.com/photo-1541961017774-22349e4a1262?w=800', 'https://images.unsplash.com/photo-1523554888454-84137e72ca08?w=800'],
 'https://images.unsplash.com/photo-1541961017774-22349e4a1262?w=500',
 38, 'EL-PRINT-001',
 ARRAY['print', 'limited edition', 'barcelona', 'cityscape', 'giclee'],
 NULL, '18x24 inches', 0.5, true, false, 4.7, 15, 12, 520,
 NULL, ARRAY[]::TEXT[]),

('c1111111-1111-1111-1111-666666666666', 'a1111111-1111-1111-1111-111111111111',
 'Floral Abstractions Series - Set of 3',
 'Three complementary abstract floral prints that work beautifully as a triptych. Each print captures a different season through color and form.',
 195.00, 250.00,
 (SELECT id FROM categories WHERE slug = 'prints'),
 ARRAY['https://images.unsplash.com/photo-1579783901586-d88db74b4fe4?w=800', 'https://images.unsplash.com/photo-1578301978162-7aae4d755744?w=800'],
 'https://images.unsplash.com/photo-1579783901586-d88db74b4fe4?w=500',
 20, 'EL-PRINT-002',
 ARRAY['abstract', 'floral', 'triptych', 'set', 'home decor'],
 NULL, '12x16 inches (each)', 1.2, true, true, 4.5, 9, 8, 310,
 NULL, ARRAY[]::TEXT[]),

-- Sculptures (2)
('c1111111-1111-1111-1111-777777777777', 'a1111111-1111-1111-1111-111111111111',
 'Eternal Harmony',
 'Hand-carved white marble sculpture representing the balance between strength and serenity. Flowing curves evoke the harmony of nature and the human spirit.',
 2400.00, NULL,
 (SELECT id FROM categories WHERE slug = 'sculptures'),
 ARRAY['https://images.unsplash.com/photo-1544413660-299165566b1d?w=800', 'https://images.unsplash.com/photo-1561839561-b13bcfe95249?w=800'],
 'https://images.unsplash.com/photo-1544413660-299165566b1d?w=500',
 1, 'EL-SCULPT-001',
 ARRAY['marble', 'sculpture', 'abstract', 'harmony', 'white'],
 'Marble', '18x12x8 inches', 15.0, true, true, 5.0, 2, 1, 67,
 'A stunning hand-carved marble sculpture that embodies the eternal dance between strength and grace. The flowing abstract forms create a sense of movement frozen in stone.',
 ARRAY['sculpture', 'marble', 'abstract', 'minimalist']),

('c1111111-1111-1111-1111-888888888888', 'a1111111-1111-1111-1111-111111111111',
 'Dancing Flames - Bronze',
 'Cast bronze sculpture of intertwined abstract forms reminiscent of flickering flames. Patinated finish gives each piece a unique character.',
 1650.00, 1900.00,
 (SELECT id FROM categories WHERE slug = 'sculptures'),
 ARRAY['https://images.unsplash.com/photo-1549887534-1541e9326642?w=800'],
 'https://images.unsplash.com/photo-1549887534-1541e9326642?w=500',
 2, 'EL-SCULPT-002',
 ARRAY['bronze', 'sculpture', 'abstract', 'fire', 'modern'],
 'Bronze', '24x10x10 inches', 8.5, true, false, 4.5, 4, 2, 112,
 NULL, ARRAY[]::TEXT[]),

-- Handcraft (2)
('c1111111-1111-1111-1111-999999999999', 'a1111111-1111-1111-1111-111111111111',
 'Hand-Painted Ceramic Vase - Mediterranean Blue',
 'A one-of-a-kind ceramic vase hand-painted with intricate Mediterranean-inspired patterns. Glazed finish in shades of cobalt blue and white.',
 280.00, NULL,
 (SELECT id FROM categories WHERE slug = 'handcraft'),
 ARRAY['https://images.unsplash.com/photo-1565193566173-7a0ee3dbe261?w=800', 'https://images.unsplash.com/photo-1578749556568-bc2c40e68b61?w=800'],
 'https://images.unsplash.com/photo-1565193566173-7a0ee3dbe261?w=500',
 4, 'EL-CRAFT-001',
 ARRAY['ceramic', 'vase', 'hand-painted', 'mediterranean', 'blue', 'decor'],
 'Ceramic', '14x8 inches', 3.0, true, false, 4.3, 5, 3, 198,
 NULL, ARRAY[]::TEXT[]),

('c1111111-1111-1111-1111-aaaaaaaaaaaa', 'a1111111-1111-1111-1111-111111111111',
 'Woven Wall Tapestry - Earth & Sky',
 'Handwoven macramé wall hanging using natural cotton and dyed fibers. Abstract landscape inspired by the meeting of earth and sky at sunset.',
 340.00, 420.00,
 (SELECT id FROM categories WHERE slug = 'textile-art'),
 ARRAY['https://images.unsplash.com/photo-1596462502278-27bfdc403348?w=800'],
 'https://images.unsplash.com/photo-1596462502278-27bfdc403348?w=500',
 3, 'EL-TEXT-001',
 ARRAY['macrame', 'tapestry', 'wall hanging', 'handwoven', 'boho'],
 'Cotton & Mixed Fibers', '36x24 inches', 1.5, true, false, 4.7, 3, 2, 87,
 NULL, ARRAY[]::TEXT[]);

-- James Chen's Products (10 products)
INSERT INTO products (id, seller_id, title, description, price, compare_at_price, category_id, images, thumbnail, stock, sku, tags, medium, dimensions, weight, is_published, is_featured, avg_rating, total_reviews, total_sales, view_count, ai_description, ai_tags) VALUES

-- Digital Art (4)
('d2222222-2222-2222-2222-111111111111', 'a2222222-2222-2222-2222-222222222222',
 'Neon Tokyo Nights',
 'A vibrant digital illustration of Tokyo''s neon-lit streets at night. Cyberpunk aesthetics meet traditional Japanese architecture in this eye-catching piece.',
 250.00, NULL,
 (SELECT id FROM categories WHERE slug = 'digital-art'),
 ARRAY['https://images.unsplash.com/photo-1545569341-9eb8b30979d9?w=800', 'https://images.unsplash.com/photo-1536098561742-ca998e48cbcc?w=800'],
 'https://images.unsplash.com/photo-1545569341-9eb8b30979d9?w=500',
 999, 'JC-DIG-001',
 ARRAY['digital', 'cyberpunk', 'tokyo', 'neon', 'night', 'japan'],
 'Digital', '4000x3000 px (print up to 40x30")', 0.0, true, true, 4.9, 18, 15, 876,
 'Step into the electric streets of future Tokyo with this stunning digital illustration. Neon signs reflect off rain-slicked streets while traditional torii gates peek between futuristic buildings.',
 ARRAY['cyberpunk', 'japanese', 'neon', 'cityscape', 'night']),

('d2222222-2222-2222-2222-aaaaaaaaaaaa', 'a2222222-2222-2222-2222-222222222222',
 'Cosmic Garden',
 'Surreal digital artwork blending botanical elements with cosmic imagery. Flowers bloom among stars and nebulae in this dreamlike composition.',
 180.00, 220.00,
 (SELECT id FROM categories WHERE slug = 'digital-art'),
 ARRAY['https://images.unsplash.com/photo-1534447677768-be436bb09401?w=800'],
 'https://images.unsplash.com/photo-1534447677768-be436bb09401?w=500',
 999, 'JC-DIG-002',
 ARRAY['surreal', 'cosmic', 'botanical', 'fantasy', 'space'],
 'Digital', '5000x3500 px', 0.0, true, false, 4.4, 7, 5, 345,
 NULL, ARRAY[]::TEXT[]),

('d2222222-2222-2222-2222-333333333333', 'a2222222-2222-2222-2222-222222222222',
 'AI Dreams #42 - Emergence',
 'Part of the AI Dreams collection. Created using custom neural style transfer algorithms, this piece explores the boundary between human creativity and machine intelligence.',
 350.00, NULL,
 (SELECT id FROM categories WHERE slug = 'digital-art'),
 ARRAY['https://images.unsplash.com/photo-1547891654-e66ed7ebb968?w=800', 'https://images.unsplash.com/photo-1558591710-4b4a1ae0f04d?w=800'],
 'https://images.unsplash.com/photo-1547891654-e66ed7ebb968?w=500',
 999, 'JC-DIG-003',
 ARRAY['AI art', 'neural network', 'abstract', 'generative', 'technology'],
 'Digital / AI-Enhanced', '6000x4000 px', 0.0, true, true, 4.7, 11, 8, 543,
 'A groundbreaking piece from the AI Dreams collection that explores the frontier of human-machine creative collaboration. Neural networks were guided by the artist to produce emergent patterns of haunting beauty.',
 ARRAY['AI-generated', 'generative art', 'abstract', 'technology']),

('d2222222-2222-2222-2222-444444444444', 'a2222222-2222-2222-2222-222222222222',
 'Minimalist Geometry - Series I',
 'Clean geometric digital artwork featuring overlapping shapes and a restrained color palette. Perfect for modern office or living spaces.',
 95.00, NULL,
 (SELECT id FROM categories WHERE slug = 'digital-art'),
 ARRAY['https://images.unsplash.com/photo-1550859492-d5da9d8e45f3?w=800'],
 'https://images.unsplash.com/photo-1550859492-d5da9d8e45f3?w=500',
 999, 'JC-DIG-004',
 ARRAY['minimalist', 'geometric', 'modern', 'clean', 'office decor'],
 'Digital', '4000x4000 px', 0.0, true, false, 4.2, 6, 10, 412,
 NULL, ARRAY[]::TEXT[]),

-- Photography (3)
('d2222222-2222-2222-2222-555555555555', 'a2222222-2222-2222-2222-222222222222',
 'Misty Morning - Yosemite',
 'Fine art landscape photograph captured at dawn in Yosemite Valley. Morning mist weaves through ancient sequoias as the first light illuminates El Capitan.',
 450.00, 550.00,
 (SELECT id FROM categories WHERE slug = 'photography'),
 ARRAY['https://images.unsplash.com/photo-1506905925346-21bda4d32df4?w=800', 'https://images.unsplash.com/photo-1464822759023-fed622ff2c3b?w=800'],
 'https://images.unsplash.com/photo-1506905925346-21bda4d32df4?w=500',
 15, 'JC-PHOTO-001',
 ARRAY['photography', 'landscape', 'yosemite', 'nature', 'morning', 'mist'],
 'Photography', '30x20 inches (printed on metallic paper)', 0.8, true, true, 4.8, 10, 7, 623,
 'A masterful landscape photograph that captures the ephemeral beauty of dawn in Yosemite Valley. The interplay of mist, ancient trees, and golden light creates a scene of profound natural beauty.',
 ARRAY['landscape', 'nature', 'fine art photography', 'national park']),

('d2222222-2222-2222-2222-666666666666', 'a2222222-2222-2222-2222-222222222222',
 'Urban Reflections - NYC',
 'Street photography capturing Manhattan reflections in rain puddles. The inverted skyline creates a dreamlike parallel universe beneath pedestrians'' feet.',
 320.00, NULL,
 (SELECT id FROM categories WHERE slug = 'photography'),
 ARRAY['https://images.unsplash.com/photo-1534430480872-3498386e7856?w=800'],
 'https://images.unsplash.com/photo-1534430480872-3498386e7856?w=500',
 20, 'JC-PHOTO-002',
 ARRAY['photography', 'street', 'urban', 'nyc', 'reflections', 'rain'],
 'Photography', '24x36 inches', 0.6, true, false, 4.5, 5, 3, 287,
 NULL, ARRAY[]::TEXT[]),

('d2222222-2222-2222-2222-777777777777', 'a2222222-2222-2222-2222-222222222222',
 'Abstract Macro: Autumn Leaves',
 'Extreme macro photography of autumn leaves revealing hidden patterns invisible to the naked eye. Color-enhanced to emphasize the fractal-like vein structures.',
 200.00, 260.00,
 (SELECT id FROM categories WHERE slug = 'photography'),
 ARRAY['https://images.unsplash.com/photo-1507003211169-0a1dd7228f2d?w=800'],
 'https://images.unsplash.com/photo-1507003211169-0a1dd7228f2d?w=500',
 25, 'JC-PHOTO-003',
 ARRAY['macro', 'autumn', 'leaves', 'abstract', 'nature', 'close-up'],
 'Photography', '20x20 inches', 0.4, true, false, 4.6, 4, 3, 156,
 NULL, ARRAY[]::TEXT[]),

-- Mixed Media (2)
('d2222222-2222-2222-2222-888888888888', 'a2222222-2222-2222-2222-222222222222',
 'Analog Meets Digital: Glitch Portrait',
 'A unique piece combining traditional charcoal portraiture with digital glitch effects. The human face emerges from and dissolves into streams of data.',
 580.00, NULL,
 (SELECT id FROM categories WHERE slug = 'mixed-media'),
 ARRAY['https://images.unsplash.com/photo-1561214115-f2f134cc4912?w=800', 'https://images.unsplash.com/photo-1549490349-8643362247b5?w=800'],
 'https://images.unsplash.com/photo-1561214115-f2f134cc4912?w=500',
 3, 'JC-MIX-001',
 ARRAY['mixed media', 'glitch art', 'portrait', 'digital', 'charcoal'],
 'Charcoal + Digital', '24x30 inches', 1.2, true, true, 4.8, 6, 3, 234,
 'A thought-provoking fusion of analog and digital art forms. The traditional charcoal portrait gradually dissolves into digital glitch patterns, questioning the nature of identity in the digital age.',
 ARRAY['mixed media', 'portrait', 'glitch', 'contemporary']),

('d2222222-2222-2222-2222-999999999999', 'a2222222-2222-2222-2222-222222222222',
 'Circuit Board Mandala',
 'Recycled electronics components arranged in a meditative mandala pattern. Resistors, capacitors, and circuit boards find new life as art.',
 420.00, NULL,
 (SELECT id FROM categories WHERE slug = 'mixed-media'),
 ARRAY['https://images.unsplash.com/photo-1558618666-fcd25c85f82e?w=800'],
 'https://images.unsplash.com/photo-1558618666-fcd25c85f82e?w=500',
 2, 'JC-MIX-002',
 ARRAY['recycled', 'electronics', 'mandala', 'assemblage', 'eco art'],
 'Recycled Electronics', '30x30 inches', 5.0, true, false, 4.4, 3, 1, 98,
 NULL, ARRAY[]::TEXT[]);

-- ============================================================================
-- CART ITEMS (Sarah has items in cart for demo)
-- ============================================================================
INSERT INTO cart_items (user_id, product_id, quantity) VALUES
('b1111111-1111-1111-1111-111111111111', 'c1111111-1111-1111-1111-222222222222', 1),
('b1111111-1111-1111-1111-111111111111', 'd2222222-2222-2222-2222-111111111111', 2),
('b1111111-1111-1111-1111-111111111111', 'c1111111-1111-1111-1111-555555555555', 1);

-- ============================================================================
-- WISHLIST ITEMS
-- ============================================================================
INSERT INTO wishlists (user_id, product_id) VALUES
('b1111111-1111-1111-1111-111111111111', 'c1111111-1111-1111-1111-111111111111'),
('b1111111-1111-1111-1111-111111111111', 'c1111111-1111-1111-1111-777777777777'),
('b1111111-1111-1111-1111-111111111111', 'd2222222-2222-2222-2222-333333333333'),
('b2222222-2222-2222-2222-222222222222', 'c1111111-1111-1111-1111-444444444444'),
('b2222222-2222-2222-2222-222222222222', 'd2222222-2222-2222-2222-555555555555'),
('b3333333-3333-3333-3333-333333333333', 'c1111111-1111-1111-1111-111111111111'),
('b5555555-5555-5555-5555-555555555555', 'd2222222-2222-2222-2222-888888888888');

-- ============================================================================
-- ORDERS (6 orders in various states)
-- ============================================================================
INSERT INTO orders (id, buyer_id, order_number, status, subtotal, shipping_cost, tax, discount, total, shipping_name, shipping_address_line1, shipping_city, shipping_state, shipping_country, shipping_zip, shipping_phone, payment_method, payment_status, tracking_number, created_at) VALUES

('e1111111-1111-1111-1111-111111111111', 'b1111111-1111-1111-1111-111111111111',
 'ART-100001', 'delivered', 850.00, 0.00, 68.00, 0.00, 918.00,
 'Sarah Johnson', '200 Park Avenue', 'New York', 'NY', 'USA', '10017', '+1-555-0201',
 'card', 'paid', 'TRACK-NYC-001', NOW() - INTERVAL '30 days'),

('e2222222-2222-2222-2222-222222222222', 'b1111111-1111-1111-1111-111111111111',
 'ART-100002', 'shipped', 500.00, 9.99, 40.80, 0.00, 550.79,
 'Sarah Johnson', '200 Park Avenue', 'New York', 'NY', 'USA', '10017', '+1-555-0201',
 'card', 'paid', 'TRACK-NYC-002', NOW() - INTERVAL '7 days'),

('e3333333-3333-3333-3333-333333333333', 'b2222222-2222-2222-2222-222222222222',
 'ART-100003', 'processing', 1200.00, 25.00, 0.00, 0.00, 1225.00,
 'Michael Thompson', '15 Oxford Street', 'London', NULL, 'UK', 'W1D 1BS', '+44-555-0202',
 'paypal', 'paid', NULL, NOW() - INTERVAL '3 days'),

('e4444444-4444-4444-4444-444444444444', 'b3333333-3333-3333-3333-333333333333',
 'ART-100004', 'confirmed', 375.00, 15.00, 30.00, 0.00, 420.00,
 'Aisha Patel', '42 MG Road', 'Mumbai', 'Maharashtra', 'India', '400001', '+91-555-0203',
 'card', 'paid', NULL, NOW() - INTERVAL '1 day'),

('e5555555-5555-5555-5555-555555555555', 'b4444444-4444-4444-4444-444444444444',
 'ART-100005', 'pending', 250.00, 12.00, 20.00, 0.00, 282.00,
 'Lucas Weber', '10 Alexanderplatz', 'Berlin', NULL, 'Germany', '10178', '+49-555-0204',
 'bank_transfer', 'pending', NULL, NOW() - INTERVAL '6 hours'),

('e6666666-6666-6666-6666-666666666666', 'b5555555-5555-5555-5555-555555555555',
 'ART-100006', 'cancelled', 620.00, 0.00, 49.60, 0.00, 669.60,
 'Maria Santos', '100 Copacabana', 'Rio de Janeiro', 'RJ', 'Brazil', '22070-000', '+55-555-0205',
 'card', 'refunded', NULL, NOW() - INTERVAL '14 days');

-- ============================================================================
-- ORDER ITEMS
-- ============================================================================
INSERT INTO order_items (order_id, product_id, seller_id, title, price, quantity, thumbnail, status) VALUES
-- Order 1 (delivered)
('e1111111-1111-1111-1111-111111111111', 'c1111111-1111-1111-1111-111111111111', 'a1111111-1111-1111-1111-111111111111',
 'Sunset Over Barcelona', 850.00, 1, 'https://images.unsplash.com/photo-1579783902614-a3fb3927b6a5?w=200', 'delivered'),

-- Order 2 (shipped) - 2 items
('e2222222-2222-2222-2222-222222222222', 'd2222222-2222-2222-2222-111111111111', 'a2222222-2222-2222-2222-222222222222',
 'Neon Tokyo Nights', 250.00, 1, 'https://images.unsplash.com/photo-1545569341-9eb8b30979d9?w=200', 'shipped'),
('e2222222-2222-2222-2222-222222222222', 'd2222222-2222-2222-2222-aaaaaaaaaaaa', 'a2222222-2222-2222-2222-222222222222',
 'Cosmic Garden', 180.00, 1, 'https://images.unsplash.com/photo-1534447677768-be436bb09401?w=200', 'shipped'),
('e2222222-2222-2222-2222-222222222222', 'c1111111-1111-1111-1111-333333333333', 'a1111111-1111-1111-1111-111111111111',
 'Meadow Dreams (Print)', 70.00, 1, 'https://images.unsplash.com/photo-1460661419201-fd4cecdf8a8b?w=200', 'shipped'),

-- Order 3 (processing)
('e3333333-3333-3333-3333-333333333333', 'c1111111-1111-1111-1111-444444444444', 'a1111111-1111-1111-1111-111111111111',
 'Portrait of Solitude', 1200.00, 1, 'https://images.unsplash.com/photo-1578662996442-48f60103fc96?w=200', 'confirmed'),

-- Order 4 (confirmed)
('e4444444-4444-4444-4444-444444444444', 'd2222222-2222-2222-2222-333333333333', 'a2222222-2222-2222-2222-222222222222',
 'AI Dreams #42 - Emergence', 350.00, 1, 'https://images.unsplash.com/photo-1547891654-e66ed7ebb968?w=200', 'pending'),
('e4444444-4444-4444-4444-444444444444', 'c1111111-1111-1111-1111-666666666666', 'a1111111-1111-1111-1111-111111111111',
 'Floral Abstractions Series', 195.00, 1, 'https://images.unsplash.com/photo-1579783901586-d88db74b4fe4?w=200', 'pending'),

-- Order 5 (pending)
('e5555555-5555-5555-5555-555555555555', 'd2222222-2222-2222-2222-111111111111', 'a2222222-2222-2222-2222-222222222222',
 'Neon Tokyo Nights', 250.00, 1, 'https://images.unsplash.com/photo-1545569341-9eb8b30979d9?w=200', 'pending'),

-- Order 6 (cancelled)
('e6666666-6666-6666-6666-666666666666', 'c1111111-1111-1111-1111-222222222222', 'a1111111-1111-1111-1111-111111111111',
 'Ocean Whispers', 620.00, 1, 'https://images.unsplash.com/photo-1549490349-8643362247b5?w=200', 'cancelled');

-- ============================================================================
-- REVIEWS (Realistic reviews for multiple products)
-- ============================================================================
INSERT INTO reviews (product_id, user_id, order_id, rating, title, comment, is_verified_purchase, helpful_count, created_at) VALUES

-- Reviews for Sunset Over Barcelona
('c1111111-1111-1111-1111-111111111111', 'b1111111-1111-1111-1111-111111111111', 'e1111111-1111-1111-1111-111111111111',
 5, 'Absolutely stunning!', 'The colors are even more vibrant in person. This painting now hangs in my living room and every guest comments on it. Elena is incredibly talented.', true, 8, NOW() - INTERVAL '25 days'),
('c1111111-1111-1111-1111-111111111111', 'b2222222-2222-2222-2222-222222222222', NULL,
 5, 'Museum quality', 'I''ve collected art for 20 years and this piece rivals anything I''ve seen in galleries. The texture and depth of the oil work is remarkable.', false, 5, NOW() - INTERVAL '20 days'),
('c1111111-1111-1111-1111-111111111111', 'b3333333-3333-3333-3333-333333333333', NULL,
 4, 'Beautiful but shipping took long', 'The painting itself is gorgeous - truly captures the Barcelona sunset. Shipping took 3 weeks though, which was longer than expected.', false, 2, NOW() - INTERVAL '15 days'),

-- Reviews for Neon Tokyo Nights
('d2222222-2222-2222-2222-111111111111', 'b1111111-1111-1111-1111-111111111111', 'e2222222-2222-2222-2222-222222222222',
 5, 'Perfect for my gaming room', 'The neon colors pop so beautifully when printed. I got it on metallic paper and it looks incredible under my LED lights.', true, 12, NOW() - INTERVAL '5 days'),
('d2222222-2222-2222-2222-111111111111', 'b4444444-4444-4444-4444-444444444444', NULL,
 5, 'Incredible detail', 'Zooming into the digital file reveals amazing hidden details. Every neon sign is readable, every reflection is intentional. True craftsmanship.', false, 6, NOW() - INTERVAL '10 days'),
('d2222222-2222-2222-2222-111111111111', 'b5555555-5555-5555-5555-555555555555', NULL,
 4, 'Love the aesthetic', 'Great cyberpunk vibe. Would love to see a series of these featuring other Asian cities.', false, 3, NOW() - INTERVAL '8 days'),

-- Reviews for Ocean Whispers
('c1111111-1111-1111-1111-222222222222', 'b2222222-2222-2222-2222-222222222222', NULL,
 5, 'Can almost hear the waves', 'The palette knife texture makes this painting come alive. You can feel the ocean spray when you look at it. Magnificent.', false, 4, NOW() - INTERVAL '18 days'),
('c1111111-1111-1111-1111-222222222222', 'b4444444-4444-4444-4444-444444444444', NULL,
 4, 'Rich blues', 'The blues in this painting are incredible. Works perfectly in my blue and white themed bedroom.', false, 1, NOW() - INTERVAL '12 days'),

-- Reviews for AI Dreams #42
('d2222222-2222-2222-2222-333333333333', 'b3333333-3333-3333-3333-333333333333', 'e4444444-4444-4444-4444-444444444444',
 5, 'The future of art', 'This piece blew my mind. The way human creativity guides AI to produce something neither could create alone is revolutionary.', true, 9, NOW() - INTERVAL '1 day'),
('d2222222-2222-2222-2222-333333333333', 'b1111111-1111-1111-1111-111111111111', NULL,
 4, 'Thought-provoking', 'Makes you question what art really means in the age of AI. Beautiful and intellectually stimulating.', false, 3, NOW() - INTERVAL '9 days'),

-- Reviews for Misty Morning
('d2222222-2222-2222-2222-555555555555', 'b5555555-5555-5555-5555-555555555555', NULL,
 5, 'Transported to Yosemite', 'I''ve been to Yosemite dozens of times and this photo captures something special that most photographers miss. The mist gives it a magical quality.', false, 7, NOW() - INTERVAL '22 days'),
('d2222222-2222-2222-2222-555555555555', 'b2222222-2222-2222-2222-222222222222', NULL,
 5, 'Print quality is exceptional', 'The metallic paper makes this look absolutely stunning. Colors are rich and the detail is sharp even at 30x20.', false, 4, NOW() - INTERVAL '16 days'),

-- Reviews for Eternal Harmony (sculpture)
('c1111111-1111-1111-1111-777777777777', 'b4444444-4444-4444-4444-444444444444', NULL,
 5, 'A masterpiece in marble', 'Words cannot describe how beautiful this sculpture is. The craftsmanship is extraordinary. It''s the centerpiece of our home.', false, 6, NOW() - INTERVAL '28 days'),

-- Reviews for Glitch Portrait
('d2222222-2222-2222-2222-888888888888', 'b1111111-1111-1111-1111-111111111111', NULL,
 5, 'Unique concept, flawless execution', 'The combination of traditional charcoal and digital glitch is unlike anything I''ve seen. It makes you do a double-take every time.', false, 5, NOW() - INTERVAL '11 days'),
('d2222222-2222-2222-2222-888888888888', 'b3333333-3333-3333-3333-333333333333', NULL,
 4, 'Conversation starter', 'Everyone who visits asks about this piece. The blend of old and new art techniques is fascinating.', false, 2, NOW() - INTERVAL '6 days'),

-- Reviews for Barcelona Print
('c1111111-1111-1111-1111-555555555555', 'b5555555-5555-5555-5555-555555555555', NULL,
 5, 'Great value for limited edition', 'Amazing quality giclee print at a fraction of the original''s price. Colors are true to the artist''s palette.', false, 3, NOW() - INTERVAL '19 days'),
('c1111111-1111-1111-1111-555555555555', 'b3333333-3333-3333-3333-333333333333', NULL,
 4, 'Nice print, quick shipping', 'Good print quality. Arrived well-packaged within a week. The paper quality feels premium.', false, 1, NOW() - INTERVAL '14 days');

-- ============================================================================
-- BROWSING HISTORY (for AI recommendations)
-- ============================================================================
INSERT INTO browsing_history (user_id, product_id, viewed_at) VALUES
('b1111111-1111-1111-1111-111111111111', 'c1111111-1111-1111-1111-111111111111', NOW() - INTERVAL '2 hours'),
('b1111111-1111-1111-1111-111111111111', 'c1111111-1111-1111-1111-222222222222', NOW() - INTERVAL '1 hour'),
('b1111111-1111-1111-1111-111111111111', 'd2222222-2222-2222-2222-111111111111', NOW() - INTERVAL '30 minutes'),
('b1111111-1111-1111-1111-111111111111', 'c1111111-1111-1111-1111-444444444444', NOW() - INTERVAL '3 hours'),
('b1111111-1111-1111-1111-111111111111', 'd2222222-2222-2222-2222-555555555555', NOW() - INTERVAL '1 day'),
('b2222222-2222-2222-2222-222222222222', 'c1111111-1111-1111-1111-777777777777', NOW() - INTERVAL '4 hours'),
('b2222222-2222-2222-2222-222222222222', 'c1111111-1111-1111-1111-888888888888', NOW() - INTERVAL '5 hours'),
('b2222222-2222-2222-2222-222222222222', 'd2222222-2222-2222-2222-888888888888', NOW() - INTERVAL '6 hours'),
('b3333333-3333-3333-3333-333333333333', 'd2222222-2222-2222-2222-333333333333', NOW() - INTERVAL '1 day'),
('b3333333-3333-3333-3333-333333333333', 'd2222222-2222-2222-2222-111111111111', NOW() - INTERVAL '2 days');

-- ============================================================================
-- AI RECOMMENDATIONS
-- ============================================================================
INSERT INTO ai_recommendations (user_id, product_id, score, reason, algorithm) VALUES
('b1111111-1111-1111-1111-111111111111', 'c1111111-1111-1111-1111-333333333333', 0.92, 'Based on your interest in oil paintings by Elena Rodriguez', 'content_based'),
('b1111111-1111-1111-1111-111111111111', 'd2222222-2222-2222-2222-333333333333', 0.87, 'Popular with buyers who purchased Neon Tokyo Nights', 'collaborative'),
('b1111111-1111-1111-1111-111111111111', 'd2222222-2222-2222-2222-888888888888', 0.85, 'Trending mixed media artwork matching your taste', 'trending'),
('b1111111-1111-1111-1111-111111111111', 'c1111111-1111-1111-1111-777777777777', 0.83, 'Top-rated sculpture by an artist you follow', 'content_based'),
('b2222222-2222-2222-2222-222222222222', 'c1111111-1111-1111-1111-111111111111', 0.91, 'Highly rated painting similar to your recent views', 'content_based'),
('b2222222-2222-2222-2222-222222222222', 'd2222222-2222-2222-2222-555555555555', 0.88, 'Popular photography matching your interests', 'collaborative');

-- ============================================================================
-- NOTIFICATIONS
-- ============================================================================
INSERT INTO notifications (user_id, type, title, message, is_read, action_url, created_at) VALUES
('b1111111-1111-1111-1111-111111111111', 'order_update', 'Order Delivered!', 'Your order ART-100001 has been delivered. Enjoy your new artwork!', true, '/order', NOW() - INTERVAL '25 days'),
('b1111111-1111-1111-1111-111111111111', 'order_update', 'Order Shipped', 'Your order ART-100002 has been shipped! Tracking: TRACK-NYC-002', false, '/order', NOW() - INTERVAL '5 days'),
('b1111111-1111-1111-1111-111111111111', 'promotion', 'Weekend Sale!', '20% off all prints this weekend. Use code ARTLOVER20 at checkout.', false, '/products?category=prints', NOW() - INTERVAL '1 day'),
('b1111111-1111-1111-1111-111111111111', 'system', 'Welcome to ArtShop!', 'Thank you for joining ArtShop. Discover amazing artwork from talented artists around the world.', true, '/', NOW() - INTERVAL '60 days'),
('b2222222-2222-2222-2222-222222222222', 'order_update', 'Order Processing', 'Your order ART-100003 is being prepared by the artist.', false, '/order', NOW() - INTERVAL '2 days'),
('a1111111-1111-1111-1111-111111111111', 'order_update', 'New Order!', 'You have a new order for "Portrait of Solitude". Check your seller dashboard.', false, '/seller/orders', NOW() - INTERVAL '3 days'),
('a1111111-1111-1111-1111-111111111111', 'review', 'New 5-Star Review', 'Sarah Johnson left a 5-star review on "Sunset Over Barcelona". Great job!', false, '/products/c1111111-1111-1111-1111-111111111111', NOW() - INTERVAL '25 days'),
('a2222222-2222-2222-2222-222222222222', 'order_update', 'New Order!', 'You have a new order for "Neon Tokyo Nights". Ship within 3 days.', false, '/seller/orders', NOW() - INTERVAL '6 hours');
