-- Grant permissions (run this as superuser first)
-- If you have superuser access, run this before the main migration

-- For Google Cloud SQL, you might need to run this as the 'postgres' or 'cloudsqlsuperuser'
GRANT ALL ON SCHEMA public TO kinetik_automation;
GRANT ALL PRIVILEGES ON DATABASE kinetik_automation TO kinetik_automation;

-- If on PostgreSQL 15+, you might also need:
-- GRANT CREATE ON SCHEMA public TO kinetik_automation;
