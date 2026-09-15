-- ============================================================
-- 迁移脚本：系统设置菜单权限码对齐
-- 日期：2026-09-15
--
-- 【适用场景】
--   升级**已存在**的数据库。全新部署直接执行 sql/init.sql 即可，无需本脚本。
--
-- 【为什么需要本脚本】
--   init.sql 对 sys_menu 使用 `INSERT IGNORE`，对已存在的行**不会更新**。
--   而本次要改的是已有菜单行的 permission 值，属于 UPDATE，必须显式执行。
--
-- 【问题背景】
--   菜单 11~14（网站/支付/OSS/短信设置）原 permission 为 system:settings:*，
--   但这 4 个码**没有任何后端接口使用** —— 这些页面实际调用的是
--     GET /system/config/prefix  → system:config:list
--     PUT  /system/config/batch  → system:config:edit
--   结果是：给非 admin 角色勾选了设置菜单，打开页面仍然 403（读不到配置、也存不了）。
--
-- 【本脚本做什么】
--   1) 把 4 个设置菜单的 permission 改为 system:config:list
--   2) 为它们各补一个 type=2 按钮，权限码 system:config:edit（写权限单独授予）
--   3) 打印一条提示：需重启服务或改动一次角色以触发 Casbin 策略重建
--
-- 【可重复执行】
--   全部为幂等语句（UPDATE 定值 + INSERT IGNORE），重复执行无副作用。
--
-- 【执行方式】
--   mysql -h127.0.0.1 -P3306 -ugin -p gin < sql/migrations/2026-09-15-settings-menu-permission.sql
-- ============================================================

-- 1) 菜单权限码对齐到实际调用的接口
UPDATE `sys_menu`
   SET `permission` = 'system:config:list'
 WHERE `id` IN (11, 12, 13, 14);

-- 2) 补「保存」按钮（type=2），写权限与读权限分离
INSERT IGNORE INTO `sys_menu`
  (`id`, `parent_id`, `name`, `path`, `component`, `icon`, `title`, `type`, `permission`,
   `sort`, `visible`, `status`, `is_cache`, `is_external`, `create_by`, `update_by`)
VALUES
  (440, 11, 'SiteSettingsEdit',    '', '', '', '保存', 2, 'system:config:edit', 1, 1, 1, 1, 0, 1, 1),
  (441, 12, 'PaymentSettingsEdit', '', '', '', '保存', 2, 'system:config:edit', 1, 1, 1, 1, 0, 1, 1),
  (442, 13, 'OSSSettingsEdit',     '', '', '', '保存', 2, 'system:config:edit', 1, 1, 1, 1, 0, 1, 1),
  (443, 14, 'SMSSettingsEdit',     '', '', '', '保存', 2, 'system:config:edit', 1, 1, 1, 1, 0, 1, 1);

-- 3) 校验
SELECT `id`, `title`, `type`, `permission`
  FROM `sys_menu`
 WHERE `id` IN (11, 12, 13, 14, 440, 441, 442, 443)
 ORDER BY `id`;

-- ⚠️ 执行后需**重启服务**（或在后台随意改动一次角色）以触发 SyncPoliciesFromRoleMenus，
--    让新的权限码写入 Casbin 策略，否则改动不会生效。
