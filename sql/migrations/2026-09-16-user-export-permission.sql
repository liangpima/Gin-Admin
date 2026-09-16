-- ============================================================
-- 迁移脚本：新增「用户导出」权限
-- 日期：2026-09-16
--
-- 【适用场景】
--   升级**已存在**的数据库。全新部署直接执行 sql/init.sql 即可，无需本脚本。
--
-- 【为什么需要本脚本】
--   init.sql 对 sys_menu / sys_role_menu 都是 INSERT IGNORE，
--   只对新装生效。已存在的库必须显式补这两条记录，
--   否则 GET /api/v1/system/user/export 会因「未配置访问权限」被拒绝（403）。
--
-- 【背景】
--   新增导出接口 GET /system/user/export（权限码 system:user:export）。
--   按规则13，新增路由必须登记权限码，且需在 sys_menu 补 type=2 的按钮记录，
--   否则该权限无法被分配给角色。
--
-- 【可重复执行】
--   全部为 INSERT IGNORE，重复执行无副作用。
--
-- 【执行方式】
--   mysql -h127.0.0.1 -P3306 -ugin -p gin < sql/migrations/2026-09-16-user-export-permission.sql
-- ============================================================

-- 1) 按钮权限记录（挂在菜单 2「管理员」之下）
INSERT IGNORE INTO `sys_menu`
  (`id`, `parent_id`, `name`, `path`, `component`, `icon`, `title`, `type`, `permission`,
   `sort`, `visible`, `status`, `is_cache`, `is_external`, `create_by`, `update_by`)
VALUES
  (103, 2, 'UserExport', '', '', '', '导出', 2, 'system:user:export', 4, 1, 1, 1, 0, 1, 1);

-- 2) 授予超级管理员（角色 id=1）
INSERT IGNORE INTO `sys_role_menu` (`role_id`, `menu_id`) VALUES (1, 103);

-- 3) 校验
SELECT m.`id`, m.`title`, m.`permission`,
       (SELECT COUNT(*) FROM `sys_role_menu` rm WHERE rm.menu_id = m.id) AS 已授权角色数
  FROM `sys_menu` m
 WHERE m.`id` = 103;

-- ⚠️ 执行后需**重启服务**（或在后台随意改动一次角色）以触发 SyncPoliciesFromRoleMenus，
--    让新权限码写入 Casbin 策略，否则改动不会生效。
