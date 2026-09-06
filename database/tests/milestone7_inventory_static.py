from pathlib import Path
import importlib.util
import re
import unittest

ROOT = Path(__file__).resolve().parents[1]

spec = importlib.util.spec_from_file_location(
    "migrate",
    ROOT / "migrate.py",
)
migrate = importlib.util.module_from_spec(spec)
spec.loader.exec_module(migrate)


class Milestone7InventoryDatabaseTests(unittest.TestCase):

    def test_six_contiguous_migrations(self):
        migrations = migrate.load_migrations(ROOT / "migrations")

        self.assertEqual(
            [item[0] for item in migrations],
            [1, 2, 3, 4, 5, 6],
        )

        self.assertEqual(
            migrations[-1][1],
            "000006_hardware_inventory_m7.up.sql",
        )

    def test_migration_six_is_non_destructive_index_only(self):
        path = ROOT / "migrations" / "000006_hardware_inventory_m7.up.sql"
        text = path.read_text()

        self.assertIn(
            "CREATE INDEX hardware_inventory_node_observed_idx",
            text,
        )

        self.assertIn(
            "ON hardware_inventory(node_id, observed_at DESC)",
            text,
        )

        forbidden = [
            "DROP TABLE",
            "DROP COLUMN",
            "DELETE FROM",
            "TRUNCATE",
            "UPDATE hardware_inventory",
            "ALTER TABLE hardware_inventory DROP",
        ]

        upper = text.upper()

        for value in forbidden:
            self.assertNotIn(value.upper(), upper)

    def test_runtime_role_has_select_insert_only(self):
        text = (ROOT / "runtime-grants.sql").read_text()

        self.assertRegex(
            text,
            r"GRANT SELECT \(id,node_id,payload,observed_at\) "
            r"ON hardware_inventory TO meshalot;",
        )

        self.assertRegex(
            text,
            r"GRANT INSERT \(id,node_id,payload,observed_at\) "
            r"ON hardware_inventory TO meshalot;",
        )

        self.assertNotRegex(
            text,
            r"GRANT UPDATE[^\n]*hardware_inventory",
        )

        self.assertNotRegex(
            text,
            r"GRANT DELETE[^\n]*hardware_inventory",
        )

        self.assertNotRegex(
            text,
            r"GRANT ALL[^\n]*hardware_inventory",
        )

    def test_runtime_revoke_covers_hardware_inventory(self):
        text = (ROOT / "runtime-grants.sql").read_text()

        revoke = re.search(
            r"REVOKE ALL PRIVILEGES ON(.*?)FROM meshalot;",
            text,
            re.S,
        )

        self.assertIsNotNone(revoke)
        self.assertIn(
            "hardware_inventory",
            revoke.group(1),
        )

    def test_apply_plan_contains_migration_six(self):
        migrations = migrate.load_migrations(ROOT / "migrations")

        sql = migrate.build_sql(
            migrations,
            "meshalot_test",
            "apply",
        )

        self.assertIn(
            "Applying 000006_hardware_inventory_m7.up.sql",
            sql,
        )

        self.assertIn(
            "hardware_inventory_node_observed_idx",
            sql,
        )


if __name__ == "__main__":
    unittest.main()
