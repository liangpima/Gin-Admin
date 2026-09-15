-- ============================================================
-- 迁移脚本：代码审查整改涉及的数据库结构变更
-- 日期：2026-09-15
--
-- 【适用场景】
--   升级**已存在**的数据库。
--   全新部署请直接执行 sql/init.sql，无需本脚本。
--
-- 【为什么需要本脚本】
--   init.sql 使用 `CREATE TABLE IF NOT EXISTS`，对已存在的表**完全不改动**
--   （已实测验证：索引与列都不会被更新）。因此索引调整与列删除必须显式执行 ALTER。
--
-- 【本脚本做什么】
--   仅处理**结构变更**（索引、列）。**数据变更**（新增权限码、新增配置项）请重跑
--   sql/init.sql —— 其中全部为 `INSERT IGNORE`，重复执行安全。
--
-- 【可重复执行】
--   脚本通过查询 information_schema 判断，已应用的变更会跳过。
--
-- 【执行方式】
--   mysql -h<host> -P<port> -u<user> -p <dbname> < 2026-09-15-schema.sql
-- ============================================================

DROP PROCEDURE IF EXISTS `_apply_review_migration`;

DELIMITER $$
CREATE PROCEDURE `_apply_review_migration`()
BEGIN
    -- ---------- 操作日志表 sys_operation_log ----------

    -- 删除 3 个从未写入数据的死列
    IF EXISTS (SELECT 1 FROM information_schema.COLUMNS
               WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'sys_operation_log'
                 AND COLUMN_NAME = 'method') THEN
        ALTER TABLE `sys_operation_log` DROP COLUMN `method`;
    END IF;

    IF EXISTS (SELECT 1 FROM information_schema.COLUMNS
               WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'sys_operation_log'
                 AND COLUMN_NAME = 'response_result') THEN
        ALTER TABLE `sys_operation_log` DROP COLUMN `response_result`;
    END IF;

    IF EXISTS (SELECT 1 FROM information_schema.COLUMNS
               WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'sys_operation_log'
                 AND COLUMN_NAME = 'location') THEN
        ALTER TABLE `sys_operation_log` DROP COLUMN `location`;
    END IF;

    -- 索引改为 (tenant_id, id)：匹配 "WHERE tenant_id=? ORDER BY id DESC LIMIT n" 的查询，
    -- 避免单列索引导致的 filesort
    IF EXISTS (SELECT 1 FROM information_schema.STATISTICS
               WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'sys_operation_log'
                 AND INDEX_NAME = 'idx_tenant_id' AND SEQ_IN_INDEX = 1)
       AND NOT EXISTS (SELECT 1 FROM information_schema.STATISTICS
                       WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'sys_operation_log'
                         AND INDEX_NAME = 'idx_tenant_id' AND SEQ_IN_INDEX = 2) THEN
        ALTER TABLE `sys_operation_log` DROP INDEX `idx_tenant_id`,
                                        ADD INDEX `idx_tenant_id` (`tenant_id`, `id`);
    END IF;

    -- ---------- 登录日志表 sys_login_log ----------

    IF EXISTS (SELECT 1 FROM information_schema.COLUMNS
               WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'sys_login_log'
                 AND COLUMN_NAME = 'location') THEN
        ALTER TABLE `sys_login_log` DROP COLUMN `location`;
    END IF;

    IF EXISTS (SELECT 1 FROM information_schema.STATISTICS
               WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'sys_login_log'
                 AND INDEX_NAME = 'idx_tenant_id' AND SEQ_IN_INDEX = 1)
       AND NOT EXISTS (SELECT 1 FROM information_schema.STATISTICS
                       WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'sys_login_log'
                         AND INDEX_NAME = 'idx_tenant_id' AND SEQ_IN_INDEX = 2) THEN
        ALTER TABLE `sys_login_log` DROP INDEX `idx_tenant_id`,
                                      ADD INDEX `idx_tenant_id` (`tenant_id`, `id`);
    END IF;

    -- ---------- 权限策略表 casbin_rule ----------
    -- 各列宽度必须 ≥ sys_menu.permission 的 200 字符，
    -- 否则长权限码写入会被截断，导致策略与预期不符。
    -- 表不存在时跳过（服务启动会自动建表，届时已是新宽度）。
    IF EXISTS (SELECT 1 FROM information_schema.TABLES
               WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'casbin_rule')
       AND EXISTS (SELECT 1 FROM information_schema.COLUMNS
                   WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'casbin_rule'
                     AND COLUMN_NAME = 'v2' AND CHARACTER_MAXIMUM_LENGTH < 200) THEN
        ALTER TABLE `casbin_rule`
            MODIFY `ptype` varchar(200) DEFAULT '',
            MODIFY `v0` varchar(200) DEFAULT '',
            MODIFY `v1` varchar(200) DEFAULT '',
            MODIFY `v2` varchar(200) DEFAULT '',
            MODIFY `v3` varchar(200) DEFAULT '',
            MODIFY `v4` varchar(200) DEFAULT '',
            MODIFY `v5` varchar(200) DEFAULT '';
    END IF;
END$$

DELIMITER ;

CALL `_apply_review_migration`();
DROP PROCEDURE `_apply_review_migration`;

-- ============================================================
-- 后续步骤（数据变更，请另行执行）
-- ============================================================
-- 1. 重跑 sql/init.sql 以补齐数据（全部为 INSERT IGNORE，可安全重复执行）：
--      - sys_menu 新增 30 条按钮权限（id 300~431）
--      - sys_role_menu 为 admin 角色补授权
--      - sys_config 新增 pay.wechat_apiv3_key（微信支付 APIv3 密钥，回调解密用）
--
-- 2. casbin_rule 表无需手工创建：服务启动时会自动 AutoMigrate。
--
-- 3. 升级后请检查：
--      - 系统配置中填入真实的 pay.wechat_apiv3_key（32 位），否则微信回调解密会报错
--      - 若使用非 admin 角色，需在「角色管理」中重新勾选菜单以授予新权限码
-- ============================================================
