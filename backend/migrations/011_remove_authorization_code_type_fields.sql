-- 社区版不区分部署类型和加密类型，幂等移除历史字段及相关索引。
SET @drop_software_status_index = (
    SELECT IF(COUNT(*) > 0,
        'DROP INDEX idx_authorization_codes_software_status ON authorization_codes',
        'SELECT 1')
    FROM information_schema.statistics
    WHERE table_schema = DATABASE()
      AND table_name = 'authorization_codes'
      AND index_name = 'idx_authorization_codes_software_status'
);
PREPARE statement FROM @drop_software_status_index;
EXECUTE statement;
DEALLOCATE PREPARE statement;

SET @drop_deployment_type_index = (
    SELECT IF(COUNT(*) > 0,
        'DROP INDEX idx_authorization_codes_deployment_type ON authorization_codes',
        'SELECT 1')
    FROM information_schema.statistics
    WHERE table_schema = DATABASE()
      AND table_name = 'authorization_codes'
      AND index_name = 'idx_authorization_codes_deployment_type'
);
PREPARE statement FROM @drop_deployment_type_index;
EXECUTE statement;
DEALLOCATE PREPARE statement;

SET @drop_deployment_type_column = (
    SELECT IF(COUNT(*) > 0,
        'ALTER TABLE authorization_codes DROP COLUMN deployment_type',
        'SELECT 1')
    FROM information_schema.columns
    WHERE table_schema = DATABASE()
      AND table_name = 'authorization_codes'
      AND column_name = 'deployment_type'
);
PREPARE statement FROM @drop_deployment_type_column;
EXECUTE statement;
DEALLOCATE PREPARE statement;

SET @drop_encryption_type_column = (
    SELECT IF(COUNT(*) > 0,
        'ALTER TABLE authorization_codes DROP COLUMN encryption_type',
        'SELECT 1')
    FROM information_schema.columns
    WHERE table_schema = DATABASE()
      AND table_name = 'authorization_codes'
      AND column_name = 'encryption_type'
);
PREPARE statement FROM @drop_encryption_type_column;
EXECUTE statement;
DEALLOCATE PREPARE statement;
