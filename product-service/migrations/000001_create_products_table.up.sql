-- product-service/migrations/001_create_products_table.up.sql
CREATE TABLE IF NOT EXISTS products (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    category VARCHAR(50) NOT NULL CHECK (category IN ('smartphone', 'laptop', 'tablet', 'tv', 'camera')),
    brand VARCHAR(100) NOT NULL,
    model VARCHAR(100) NOT NULL,
    price DECIMAL(10,2) NOT NULL CHECK (price >= 0),
    stock INTEGER NOT NULL DEFAULT 0 CHECK (stock >= 0),
    image_url VARCHAR(500),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Индексы для быстрого поиска
CREATE INDEX IF NOT EXISTS idx_products_category ON products(category);
CREATE INDEX IF NOT EXISTS idx_products_brand ON products(brand);
CREATE INDEX IF NOT EXISTS idx_products_price ON products(price);
CREATE INDEX IF NOT EXISTS idx_products_created_at ON products(created_at);

-- Триггер для автоматического обновления updated_at
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ language 'plpgsql';

CREATE TRIGGER update_products_updated_at 
    BEFORE UPDATE ON products 
    FOR EACH ROW 
    EXECUTE FUNCTION update_updated_at_column();

-- Начальные данные для тестирования
INSERT INTO products (name, description, category, brand, model, price, stock, image_url) VALUES
('iPhone 15 Pro', 'Latest Apple smartphone with A17 Pro chip', 'smartphone', 'Apple', 'iPhone 15 Pro', 999.99, 50, 'https://example.com/iphone15pro.jpg'),
('MacBook Air M2', 'Lightweight laptop with Apple M2 chip', 'laptop', 'Apple', 'MacBook Air M2', 1199.99, 30, 'https://example.com/macbook-air-m2.jpg'),
('Galaxy S23', 'Samsung flagship smartphone', 'smartphone', 'Samsung', 'Galaxy S23', 799.99, 75, 'https://example.com/galaxy-s23.jpg'),
('ThinkPad X1', 'Business laptop from Lenovo', 'laptop', 'Lenovo', 'ThinkPad X1', 1299.99, 25, 'https://example.com/thinkpad-x1.jpg'),
('iPad Air', 'Versatile tablet from Apple', 'tablet', 'Apple', 'iPad Air', 599.99, 40, 'https://example.com/ipad-air.jpg')
ON CONFLICT DO NOTHING;