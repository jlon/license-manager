-- 支持创建不关联客户的授权码，以及后续匿名授权激活
ALTER TABLE authorization_codes
    MODIFY customer_id VARCHAR(36) NULL COMMENT '关联客户ID，NULL表示无客户';

ALTER TABLE licenses
    MODIFY customer_id VARCHAR(36) NULL COMMENT '客户ID，冗余字段便于查询，NULL表示无客户';
