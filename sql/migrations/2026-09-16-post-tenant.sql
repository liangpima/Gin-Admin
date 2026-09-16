-- ============================================================
-- 迁移脚本：sys_post 由「全局表」改为「租户内表」
-- 日期：2026-09-16
--
-- 【适用场景】
--   升级**已存在**的数据库。
--   全新部署请直接执行 sql/init.sql，无需本脚本。
--
-- 【为什么需要本脚本】
--   init.sql 使用 `CREATE TABLE IF NOT EXISTS`，对已存在的表**完全不改动**
--   （含列与索引）。因此新增列、索引调整必须显式执行 ALTER。
--
-- 【本脚本做什么】
--   1. 新增 `tenant_id` 列（默认 0）
--   2. 唯一索引 `uk_code(code)` → `uk_tenant_code(tenant_id, code)`，
--      使不同租户可以使用相同的岗位编码
--   3. 新增 `idx_tenant_id`
--
-- 【行为变化（务必知悉）】
--   历史数据会被置为 `tenant_id = 0`，即「平台级数据」。
--   由于 tenant_id=0 表示不过滤（见 common.TenantScope），
--   **平台级账号（如默认 admin）仍能看到全部岗位**；
--   而各租户账号从此只能看到属于自己租户的岗位，不再看到这批历史岗位。
--
--   若希望某个租户沿用历史岗位，请手工执行：
--     UPDATE sys_post SET tenant_id = <租户ID> WHERE tenant_id = 0;
--
-- 【可重复执行】
--   脚本通过查询 information_schema 判断，已应用的变更会跳过。
--
-- 【执行方式】
--   mysql -h<host> -P<port> -u<user> -p <dbname> < 2026-09-16-post-tenant.sql
-- ============================================================

DROP PROCEDURE IF EXISTS `_apply_post_tenant_migration`;

DELIMITER $$
CREATE PROCEDURE `_apply_post_tenant_migration`()
BEGIN
    -- 1. 新增 tenant_id。
    --    默认 0 = 平台级：平台账号仍可见历史岗位，租户账号隔离后不再可见。
    IF NOT EXISTS (SELECT 1 FROM information_schema.COLUMNS
                   WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'sys_post'
                     AND COLUMN_NAME = 'tenant_id') THEN
        ALTER TABLE `sys_post`
            ADD COLUMN `tenant_id` bigint unsigned DEFAULT 0 COMMENT '租户ID' AFTER `id`;
    END IF;

    -- 2. 唯一索引改为 (tenant_id, code)。
    --    原来是全局唯一，导致不同租户无法使用相同的岗位编码，
    --    而应用层的重名校验是按租户过滤的 —— 两边语义不一致，
    --    表现为校验通过却插入报 1062。
    --    先建新索引再删旧索引，避免中间出现「无唯一约束」的时间窗口。
    IF NOT EXISTS (SELECT 1 FROM information_schema.STATISTICS
                   WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'sys_post'
                     AND INDEX_NAME = 'uk_tenant_code') THEN
        ALTER TABLE `sys_post` ADD UNIQUE KEY `uk_tenant_code` (`tenant_id`, `code`);
    END IF;

    IF EXISTS (SELECT 1 FROM information_schema.STATISTICS
               WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'sys_post'
                 AND INDEX_NAME = 'uk_code') THEN
        ALTER TABLE `sys_post` DROP INDEX `uk_code`;
    END IF;

    -- 3. 租户过滤查询走 idx_tenant_id（与 sys_user / sys_role 的索引口径一致）
    IF NOT EXISTS (SELECT 1 FROM information_schema.STATISTICS
                   WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'sys_post'
                     AND INDEX_NAME = 'idx_tenant_id') THEN
        ALTER TABLE `sys_post` ADD INDEX `idx_tenant_id` (`tenant_id`);
    END IF;

    -- 4. 同步表注释（幂等，可重复执行）
    ALTER TABLE `sys_post` COMMENT = '岗位表（租户内数据）';
END$$

DELIMITER ;

CALL `_apply_post_tenant_migration`();
DROP PROCEDURE `_apply_post_tenant_migration`;

-- ============================================================
-- 后续步骤
-- ============================================================
-- 1. 确认历史岗位的租户归属（见上方「行为变化」），按需分配：
--      SELECT id, tenant_id, code, name FROM sys_post ORDER BY tenant_id, id;
--
-- 2. sys_user_post 是纯关联表（只有 user_id / post_id，无 tenant_id），本脚本无需改动。
--    隔离性由应用层保证：写入前校验岗位属于当前租户
--    （见 userService.normalizePostIDs）。
--
-- 3. 建议排查跨租户的历史绑定（用户与其岗位分属不同租户）：
--    这类绑定在应用层会失效（不报错，只是不再返回），如需清理请先确认再删除：
--      SELECT u.id AS user_id, u.tenant_id AS user_tenant, p.id AS post_id, p.tenant_id AS post_tenant
--        FROM sys_user_post ur
--        JOIN sys_user u ON u.id = ur.user_id
--        JOIN sys_post p ON p.id = ur.post_id
--       WHERE u.tenant_id <> p.tenant_id;
-- ============================================================
