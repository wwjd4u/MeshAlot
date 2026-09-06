BEGIN;
CREATE INDEX hardware_inventory_node_observed_idx
    ON hardware_inventory(node_id, observed_at DESC);
COMMIT;
