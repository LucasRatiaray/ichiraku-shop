INSERT INTO products (business_id, name, description, category, price_ht) VALUES
  ('PDT-RAM001', 'Ramen Tonkotsu', 'Bouillon de porc 12h', 'ramen', 1250),
  ('PDT-RAM002', 'Ramen Miso', 'Bouillon miso, maïs, beurre', 'ramen', 1150),
  ('PDT-RAM003', 'Ramen Shoyu', 'Bouillon de poulet à la sauce soja', 'ramen', 1100),
  ('PDT-ENT001', 'Gyoza', 'Raviolis grillés au porc', 'entrée', 650),
  ('PDT-ENT002', 'Edamame', 'Fèves de soja au sel', 'entrée', 450),
  ('PDT-BOI001', 'Thé vert', 'Sencha chaud', 'boisson', 300)
ON CONFLICT (business_id) DO NOTHING;