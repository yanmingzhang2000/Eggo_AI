-- ═══════════════════════════════════════════════════════════════
-- 迁移 008：将 virtual_* 表的 user_id 列从 BIGINT 改为 VARCHAR(36)
-- 兼容 users.id 的 UUID 类型
-- ═══════════════════════════════════════════════════════════════

ALTER TABLE virtual_account
  ALTER COLUMN user_id TYPE VARCHAR(36) USING user_id::text;

ALTER TABLE virtual_position
  ALTER COLUMN user_id TYPE VARCHAR(36) USING user_id::text;

ALTER TABLE virtual_order
  ALTER COLUMN user_id TYPE VARCHAR(36) USING user_id::text;

ALTER TABLE virtual_transaction
  ALTER COLUMN user_id TYPE VARCHAR(36) USING user_id::text;

ALTER TABLE virtual_dca_plan
  ALTER COLUMN user_id TYPE VARCHAR(36) USING user_id::text;

-- 用现有 virtual_account.user_id (旧数字) 更新为真实 UUID
-- 执行前先确认 yanming 的 UUID：SELECT id FROM users WHERE username = 'yanming';
-- 然后替换下方的 '<yanming-uuid>'：
-- UPDATE virtual_account    SET user_id = '<yanming-uuid>' WHERE user_id = '1';
-- UPDATE virtual_position   SET user_id = '<yanming-uuid>' WHERE user_id = '1';
-- UPDATE virtual_order      SET user_id = '<yanming-uuid>' WHERE user_id = '1';
-- UPDATE virtual_transaction SET user_id = '<yanming-uuid>' WHERE user_id = '1';
-- UPDATE virtual_dca_plan   SET user_id = '<yanming-uuid>' WHERE user_id = '1';
