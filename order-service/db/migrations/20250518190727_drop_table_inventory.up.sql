-- Since we're migrating to use an external warehouse service,
-- we no longer need the local inventory table
DROP TABLE IF EXISTS inventory;