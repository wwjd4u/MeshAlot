\set ON_ERROR_STOP on

DO $$
BEGIN
    IF current_database() !~ '^meshalot(_[a-z0-9]+)*$' THEN
        RAISE EXCEPTION 'Refusing M6 privilege test outside a MeshAlot database';
    END IF;
    IF has_table_privilege('meshalot','enrollment_tokens','DELETE')
       OR has_table_privilege('meshalot','enrollment_tokens','TRUNCATE')
       OR has_table_privilege('meshalot','nodes','DELETE')
       OR has_table_privilege('meshalot','nodes','TRUNCATE') THEN
        RAISE EXCEPTION 'runtime role has destructive table privileges';
    END IF;
    IF NOT has_column_privilege('meshalot','nodes','identity_public_key','SELECT')
       OR NOT has_column_privilege('meshalot','nodes','identity_public_key','INSERT')
       OR NOT has_column_privilege('meshalot','nodes','identity_public_key','UPDATE') THEN
        RAISE EXCEPTION 'runtime role lacks required node identity privileges';
    END IF;
    IF NOT has_column_privilege('meshalot','enrollment_tokens','token_hash','SELECT')
       OR NOT has_column_privilege('meshalot','enrollment_tokens','token_hash','INSERT')
       OR has_column_privilege('meshalot','enrollment_tokens','token_hash','UPDATE') THEN
        RAISE EXCEPTION 'runtime token hash privileges are incorrect';
    END IF;
    IF NOT has_column_privilege('meshalot','enrollment_tokens','consumed_at','UPDATE')
       OR NOT has_column_privilege('meshalot','enrollment_tokens','consumed_node_id','UPDATE')
       OR has_column_privilege('meshalot','enrollment_tokens','user_id','UPDATE')
       OR has_column_privilege('meshalot','enrollment_tokens','expires_at','UPDATE') THEN
        RAISE EXCEPTION 'runtime token consumption privileges are incorrect';
    END IF;
END
$$;

SELECT 'MILESTONE 6 RUNTIME PRIVILEGES PASSED' AS result;
