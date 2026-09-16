-- ============================================================
-- 迁移脚本：数据字典增强
-- 日期：2026-09-16
--
-- 【适用场景】
--   升级**已存在**的数据库。全新部署直接执行 sql/init.sql 即可，无需本脚本。
--
-- 【本次改动】
--   1) sys_dict_data 增加复合唯一索引 uk_dict_type_value(dict_type, value)，
--      并删除被它覆盖的冗余单列索引 idx_dict_type
--   2) 新增「编辑」按钮权限 system:dict:edit（菜单 332）并授予超管
--   3) 预置三组框架自用字典：用户状态、支付订单状态、支付渠道
--
-- 【为什么加唯一索引】
--   同一类型下允许存在重复 value 时，前端按 value 取值回显哪一条 label
--   完全取决于查询顺序，属于随机行为；且「编辑」功能上线后，
--   两条记录互相覆盖的窗口会被放大。
--
-- 【可重复执行】
--   索引变更用 information_schema 条件判断，其余全部 INSERT IGNORE。
--
-- 【执行方式】
--   mysql -h127.0.0.1 -P3306 -ugin -p gin < sql/migrations/2026-09-16-dict-enhance.sql
--
-- ⚠️ 执行后需**重启服务**（或在后台改动一次角色菜单）以触发 SyncPoliciesFromRoleMenus，
--    让 system:dict:edit 写入 Casbin 策略，否则编辑接口仍会 403。
-- ============================================================

-- ------------------------------------------------------------
-- 0) 加索引前的体检：若下面查出记录，说明库里已存在同类型重复键值，
--    必须先手工处理（改 value 或删掉多余记录），否则第 1 步的 ALTER 会失败。
--    这里刻意不自动去重 —— 静默删数据的风险远大于让迁移报错。
-- ------------------------------------------------------------
SELECT `dict_type`, `value`, COUNT(*) AS 重复条数, GROUP_CONCAT(`id`) AS 涉及ID
  FROM `sys_dict_data`
 WHERE `deleted_at` IS NULL
 GROUP BY `dict_type`, `value`
HAVING COUNT(*) > 1;

-- ------------------------------------------------------------
-- 1) 唯一索引：uk_dict_type_value(dict_type, value)
-- ------------------------------------------------------------
SET @has_uk := (
  SELECT COUNT(*) FROM information_schema.statistics
   WHERE table_schema = DATABASE()
     AND table_name = 'sys_dict_data'
     AND index_name = 'uk_dict_type_value'
);
SET @sql := IF(@has_uk = 0,
  'ALTER TABLE `sys_dict_data` ADD UNIQUE KEY `uk_dict_type_value` (`dict_type`, `value`)',
  'SELECT ''uk_dict_type_value 已存在，跳过'' AS msg');
PREPARE stmt FROM @sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;

-- 2) 删除冗余单列索引：复合索引的最左前缀已覆盖 dict_type 查询，
--    再留一个单列索引只会白增写入成本。
SET @has_idx := (
  SELECT COUNT(*) FROM information_schema.statistics
   WHERE table_schema = DATABASE()
     AND table_name = 'sys_dict_data'
     AND index_name = 'idx_dict_type'
);
SET @sql := IF(@has_idx > 0,
  'ALTER TABLE `sys_dict_data` DROP INDEX `idx_dict_type`',
  'SELECT ''idx_dict_type 不存在，跳过'' AS msg');
PREPARE stmt FROM @sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;

-- ------------------------------------------------------------
-- 3) 「编辑」按钮权限（挂在菜单 8「数据字典」之下）
-- ------------------------------------------------------------
INSERT IGNORE INTO `sys_menu`
  (`id`, `parent_id`, `name`, `path`, `component`, `icon`, `title`, `type`, `permission`,
   `sort`, `visible`, `status`, `is_cache`, `is_external`, `create_by`, `update_by`)
VALUES
  (332, 8, 'DictEdit', '', '', '', '编辑', 2, 'system:dict:edit', 3, 1, 1, 1, 0, 1, 1);

INSERT IGNORE INTO `sys_role_menu` (`role_id`, `menu_id`) VALUES (1, 332);

-- ------------------------------------------------------------
-- 4) 字典种子数据
--    不指定 id：靠 uk_type / uk_dict_type_value 保证重复执行为空操作。
-- ------------------------------------------------------------
INSERT IGNORE INTO `sys_dict_type` (`name`, `type`, `status`, `create_by`, `update_by`, `created_at`, `updated_at`) VALUES
('用户状态', 'sys_user_status', 1, 1, 1, NOW(), NOW()),
('支付订单状态', 'sys_pay_order_status', 1, 1, 1, NOW(), NOW()),
('支付渠道', 'sys_pay_channel', 1, 1, 1, NOW(), NOW());

INSERT IGNORE INTO `sys_dict_data` (`dict_type`, `label`, `value`, `sort`, `list_class`, `status`, `create_by`, `update_by`, `created_at`, `updated_at`) VALUES
('sys_user_status', '正常', '1', 1, 'success', 1, 1, 1, NOW(), NOW()),
('sys_user_status', '停用', '0', 2, 'danger', 1, 1, 1, NOW(), NOW()),
('sys_pay_order_status', '待支付', '0', 1, 'info', 1, 1, 1, NOW(), NOW()),
('sys_pay_order_status', '已支付', '1', 2, 'success', 1, 1, 1, NOW(), NOW()),
('sys_pay_order_status', '已关闭', '2', 3, 'warning', 1, 1, 1, NOW(), NOW()),
('sys_pay_order_status', '已退款', '3', 4, 'danger', 1, 1, 1, NOW(), NOW()),
('sys_pay_order_status', '退款中', '4', 5, 'primary', 1, 1, 1, NOW(), NOW()),
('sys_pay_channel', '微信支付', 'wechat', 1, 'success', 1, 1, 1, NOW(), NOW()),
('sys_pay_channel', '支付宝', 'alipay', 2, 'primary', 1, 1, 1, NOW(), NOW());

-- ------------------------------------------------------------
-- 5) 校验
-- ------------------------------------------------------------
SELECT (SELECT COUNT(*) FROM `sys_dict_type`) AS 字典类型数,
       (SELECT COUNT(*) FROM `sys_dict_data`) AS 字典数据数,
       (SELECT COUNT(*) FROM information_schema.statistics
         WHERE table_schema = DATABASE() AND table_name = 'sys_dict_data'
           AND index_name = 'uk_dict_type_value') AS 唯一索引列数,
       (SELECT COUNT(*) FROM `sys_menu` WHERE `id` = 332) AS 编辑菜单数;
