-- ============================================================
-- Gin-Admin 后台管理框架 - 数据库初始化脚本
-- 适用于: MySQL 5.7+ / 8.0+
-- 字符集: utf8mb4
-- ============================================================

-- 创建数据库（如已存在则跳过）
CREATE DATABASE IF NOT EXISTS `gin` DEFAULT CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci;
USE `gin`;

-- ============================================================
-- 一、系统管理表
-- ============================================================

-- 系统用户表
CREATE TABLE IF NOT EXISTS `sys_user` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `tenant_id` bigint unsigned DEFAULT 0 COMMENT '租户ID',
  `username` varchar(64) NOT NULL COMMENT '用户名',
  `password` varchar(128) NOT NULL COMMENT '密码',
  `nickname` varchar(64) DEFAULT '' COMMENT '昵称',
  `email` varchar(128) DEFAULT '' COMMENT '邮箱',
  `phone` varchar(16) DEFAULT '' COMMENT '手机号',
  `avatar` varchar(512) DEFAULT '' COMMENT '头像',
  `status` tinyint DEFAULT 1 COMMENT '状态 0停用 1正常',
  `dept_id` bigint unsigned DEFAULT 0 COMMENT '部门ID',
  `login_ip` varchar(128) DEFAULT '' COMMENT '最后登录IP',
  `login_time` datetime DEFAULT NULL COMMENT '最后登录时间',
  `create_by` bigint unsigned DEFAULT 0 COMMENT '创建者',
  `update_by` bigint unsigned DEFAULT 0 COMMENT '更新者',
  `created_at` datetime DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `updated_at` datetime DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  `deleted_at` datetime DEFAULT NULL COMMENT '删除时间',
  `remark` varchar(500) DEFAULT '' COMMENT '备注',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_username` (`username`),
  KEY `idx_tenant_id` (`tenant_id`),
  KEY `idx_dept_id` (`dept_id`),
  KEY `idx_deleted_at` (`deleted_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='系统用户表';

-- 系统角色表
CREATE TABLE IF NOT EXISTS `sys_role` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `tenant_id` bigint unsigned DEFAULT 0 COMMENT '租户ID',
  `name` varchar(64) NOT NULL COMMENT '角色名称',
  `code` varchar(64) NOT NULL COMMENT '角色编码',
  `sort` int DEFAULT 0 COMMENT '排序',
  `status` tinyint DEFAULT 1 COMMENT '状态',
  `data_scope` tinyint DEFAULT 1 COMMENT '数据权限范围',
  `create_by` bigint unsigned DEFAULT 0 COMMENT '创建者',
  `update_by` bigint unsigned DEFAULT 0 COMMENT '更新者',
  `created_at` datetime DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `updated_at` datetime DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  `deleted_at` datetime DEFAULT NULL COMMENT '删除时间',
  `remark` varchar(500) DEFAULT '' COMMENT '备注',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_code` (`code`),
  KEY `idx_tenant_id` (`tenant_id`),
  KEY `idx_deleted_at` (`deleted_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='系统角色表';

-- 系统菜单表
CREATE TABLE IF NOT EXISTS `sys_menu` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `parent_id` bigint unsigned DEFAULT 0 COMMENT '父菜单ID',
  `name` varchar(64) NOT NULL COMMENT '菜单名称',
  `path` varchar(200) DEFAULT '' COMMENT '路由地址',
  `component` varchar(200) DEFAULT '' COMMENT '组件路径',
  `redirect` varchar(200) DEFAULT '' COMMENT '重定向',
  `icon` varchar(64) DEFAULT '' COMMENT '图标',
  `title` varchar(64) DEFAULT '' COMMENT '标题',
  `type` tinyint NOT NULL COMMENT '类型 0目录 1菜单 2按钮',
  `permission` varchar(200) DEFAULT '' COMMENT '权限标识',
  `sort` int DEFAULT 0 COMMENT '排序',
  `visible` tinyint DEFAULT 1 COMMENT '是否可见',
  `status` tinyint DEFAULT 1 COMMENT '状态',
  `is_external` tinyint DEFAULT 0 COMMENT '是否外链',
  `is_cache` tinyint DEFAULT 1 COMMENT '是否缓存',
  `create_by` bigint unsigned DEFAULT 0 COMMENT '创建者',
  `update_by` bigint unsigned DEFAULT 0 COMMENT '更新者',
  `created_at` datetime DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `updated_at` datetime DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  `deleted_at` datetime DEFAULT NULL COMMENT '删除时间',
  `remark` varchar(500) DEFAULT '' COMMENT '备注',
  PRIMARY KEY (`id`),
  KEY `idx_parent_id` (`parent_id`),
  KEY `idx_deleted_at` (`deleted_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='系统菜单表';

-- 部门表
CREATE TABLE IF NOT EXISTS `sys_dept` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `parent_id` bigint unsigned DEFAULT 0 COMMENT '父部门ID',
  `name` varchar(64) NOT NULL COMMENT '部门名称',
  `sort` int DEFAULT 0 COMMENT '排序',
  `leader` varchar(64) DEFAULT '' COMMENT '负责人',
  `phone` varchar(16) DEFAULT '' COMMENT '联系电话',
  `email` varchar(128) DEFAULT '' COMMENT '邮箱',
  `status` tinyint DEFAULT 1 COMMENT '状态',
  `create_by` bigint unsigned DEFAULT 0 COMMENT '创建者',
  `update_by` bigint unsigned DEFAULT 0 COMMENT '更新者',
  `created_at` datetime DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `updated_at` datetime DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  `deleted_at` datetime DEFAULT NULL COMMENT '删除时间',
  `remark` varchar(500) DEFAULT '' COMMENT '备注',
  PRIMARY KEY (`id`),
  KEY `idx_parent_id` (`parent_id`),
  KEY `idx_deleted_at` (`deleted_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='部门表';

-- 岗位表
CREATE TABLE IF NOT EXISTS `sys_post` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `code` varchar(64) NOT NULL COMMENT '岗位编码',
  `name` varchar(64) NOT NULL COMMENT '岗位名称',
  `sort` int DEFAULT 0 COMMENT '排序',
  `status` tinyint DEFAULT 1 COMMENT '状态',
  `create_by` bigint unsigned DEFAULT 0 COMMENT '创建者',
  `update_by` bigint unsigned DEFAULT 0 COMMENT '更新者',
  `created_at` datetime DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `updated_at` datetime DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  `deleted_at` datetime DEFAULT NULL COMMENT '删除时间',
  `remark` varchar(500) DEFAULT '' COMMENT '备注',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_code` (`code`),
  KEY `idx_deleted_at` (`deleted_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='岗位表';

-- 系统配置表
CREATE TABLE IF NOT EXISTS `sys_config` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `name` varchar(128) NOT NULL COMMENT '参数名称',
  `config_key` varchar(191) NOT NULL COMMENT '参数键名',
  `value` text COMMENT '参数键值',
  `type` tinyint DEFAULT 1 COMMENT '系统内置 0是 1否',
  `create_by` bigint unsigned DEFAULT 0 COMMENT '创建者',
  `update_by` bigint unsigned DEFAULT 0 COMMENT '更新者',
  `created_at` datetime DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `updated_at` datetime DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  `deleted_at` datetime DEFAULT NULL COMMENT '删除时间',
  `remark` varchar(500) DEFAULT '' COMMENT '备注',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_config_key` (`config_key`),
  KEY `idx_deleted_at` (`deleted_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='系统配置表';

-- 字典类型表
CREATE TABLE IF NOT EXISTS `sys_dict_type` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `name` varchar(128) NOT NULL COMMENT '字典名称',
  `type` varchar(128) NOT NULL COMMENT '字典类型',
  `status` tinyint DEFAULT 1 COMMENT '状态',
  `create_by` bigint unsigned DEFAULT 0 COMMENT '创建者',
  `update_by` bigint unsigned DEFAULT 0 COMMENT '更新者',
  `created_at` datetime DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `updated_at` datetime DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  `deleted_at` datetime DEFAULT NULL COMMENT '删除时间',
  `remark` varchar(500) DEFAULT '' COMMENT '备注',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_type` (`type`),
  KEY `idx_deleted_at` (`deleted_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='字典类型表';

-- 字典数据表
CREATE TABLE IF NOT EXISTS `sys_dict_data` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `dict_type` varchar(128) NOT NULL COMMENT '字典类型',
  `label` varchar(128) NOT NULL COMMENT '字典标签',
  `value` varchar(128) NOT NULL COMMENT '字典键值',
  `sort` int DEFAULT 0 COMMENT '排序',
  `css_class` varchar(128) DEFAULT '' COMMENT '样式属性',
  `list_class` varchar(128) DEFAULT '' COMMENT '表格回显样式',
  `status` tinyint DEFAULT 1 COMMENT '状态',
  `create_by` bigint unsigned DEFAULT 0 COMMENT '创建者',
  `update_by` bigint unsigned DEFAULT 0 COMMENT '更新者',
  `created_at` datetime DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `updated_at` datetime DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  `deleted_at` datetime DEFAULT NULL COMMENT '删除时间',
  `remark` varchar(500) DEFAULT '' COMMENT '备注',
  PRIMARY KEY (`id`),
  KEY `idx_dict_type` (`dict_type`),
  KEY `idx_deleted_at` (`deleted_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='字典数据表';

-- 操作日志表
CREATE TABLE IF NOT EXISTS `sys_operation_log` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `tenant_id` bigint unsigned DEFAULT 0 COMMENT '租户ID',
  `title` varchar(64) DEFAULT '' COMMENT '模块标题',
  `action` varchar(64) DEFAULT '' COMMENT '操作类型',
  `method` varchar(200) DEFAULT '' COMMENT '请求方法',
  `request_method` varchar(10) DEFAULT '' COMMENT 'HTTP方法',
  `request_url` varchar(500) DEFAULT '' COMMENT '请求URL',
  `request_param` text COMMENT '请求参数',
  `response_result` text COMMENT '返回结果',
  `status` tinyint DEFAULT 1 COMMENT '状态 0失败 1成功',
  `error_msg` text COMMENT '错误消息',
  `ip` varchar(128) DEFAULT '' COMMENT '操作IP',
  `location` varchar(255) DEFAULT '' COMMENT '操作地点',
  `user_agent` varchar(500) DEFAULT '' COMMENT '浏览器UA',
  `operator_id` bigint unsigned DEFAULT 0 COMMENT '操作人ID',
  `operator_name` varchar(64) DEFAULT '' COMMENT '操作人名称',
  `cost_time` bigint DEFAULT 0 COMMENT '耗时(ms)',
  `created_at` datetime DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  PRIMARY KEY (`id`),
  KEY `idx_tenant_id` (`tenant_id`),
  KEY `idx_created_at` (`created_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='操作日志表';

-- 登录日志表
CREATE TABLE IF NOT EXISTS `sys_login_log` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `tenant_id` bigint unsigned DEFAULT 0 COMMENT '租户ID',
  `username` varchar(64) DEFAULT '' COMMENT '用户名',
  `ip` varchar(128) DEFAULT '' COMMENT '登录IP',
  `location` varchar(255) DEFAULT '' COMMENT '登录地点',
  `browser` varchar(128) DEFAULT '' COMMENT '浏览器',
  `os` varchar(128) DEFAULT '' COMMENT '操作系统',
  `status` tinyint DEFAULT 1 COMMENT '状态 0失败 1成功',
  `msg` varchar(255) DEFAULT '' COMMENT '消息',
  `login_time` datetime DEFAULT CURRENT_TIMESTAMP COMMENT '登录时间',
  PRIMARY KEY (`id`),
  KEY `idx_tenant_id` (`tenant_id`),
  KEY `idx_login_time` (`login_time`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='登录日志表';

-- 文件表
CREATE TABLE IF NOT EXISTS `sys_file` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `tenant_id` bigint unsigned DEFAULT 0 COMMENT '租户ID',
  `name` varchar(200) NOT NULL COMMENT '原始文件名',
  `storage_name` varchar(200) DEFAULT '' COMMENT '存储文件名',
  `path` varchar(500) DEFAULT '' COMMENT '文件路径',
  `url` varchar(500) DEFAULT '' COMMENT '访问URL',
  `size` bigint DEFAULT 0 COMMENT '文件大小',
  `mime_type` varchar(128) DEFAULT '' COMMENT 'MIME类型',
  `storage_type` tinyint DEFAULT 0 COMMENT '存储类型 0本地 1OSS',
  `create_by` bigint unsigned DEFAULT 0 COMMENT '创建者',
  `update_by` bigint unsigned DEFAULT 0 COMMENT '更新者',
  `created_at` datetime DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `updated_at` datetime DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  `deleted_at` datetime DEFAULT NULL COMMENT '删除时间',
  `remark` varchar(500) DEFAULT '' COMMENT '备注',
  PRIMARY KEY (`id`),
  KEY `idx_tenant_id` (`tenant_id`),
  KEY `idx_deleted_at` (`deleted_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='文件表';

-- 租户表
CREATE TABLE IF NOT EXISTS `sys_tenant` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `name` varchar(128) NOT NULL COMMENT '租户名称',
  `contact_name` varchar(64) DEFAULT '' COMMENT '联系人',
  `contact_phone` varchar(16) DEFAULT '' COMMENT '联系电话',
  `status` tinyint DEFAULT 1 COMMENT '状态',
  `expire_time` datetime DEFAULT NULL COMMENT '过期时间',
  `create_by` bigint unsigned DEFAULT 0 COMMENT '创建者',
  `update_by` bigint unsigned DEFAULT 0 COMMENT '更新者',
  `created_at` datetime DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `updated_at` datetime DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  `deleted_at` datetime DEFAULT NULL COMMENT '删除时间',
  `remark` varchar(500) DEFAULT '' COMMENT '备注',
  PRIMARY KEY (`id`),
  KEY `idx_deleted_at` (`deleted_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='租户表';

-- 协议管理表
CREATE TABLE IF NOT EXISTS `sys_agreement` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `title` varchar(128) NOT NULL COMMENT '标题',
  `content` longtext COMMENT '内容',
  `type` varchar(32) DEFAULT '' COMMENT '类型',
  `sort` int DEFAULT 0 COMMENT '排序',
  `status` tinyint DEFAULT 1 COMMENT '状态 0停用 1正常',
  `remark` varchar(500) DEFAULT '' COMMENT '备注',
  `create_by` bigint unsigned DEFAULT 0 COMMENT '创建者',
  `update_by` bigint unsigned DEFAULT 0 COMMENT '更新者',
  `created_at` datetime DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `updated_at` datetime DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  `deleted_at` datetime DEFAULT NULL COMMENT '删除时间',
  PRIMARY KEY (`id`),
  KEY `idx_type` (`type`),
  KEY `idx_deleted_at` (`deleted_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='协议管理表';

-- 关联表
CREATE TABLE IF NOT EXISTS `sys_user_role` (
  `user_id` bigint unsigned NOT NULL COMMENT '用户ID',
  `role_id` bigint unsigned NOT NULL COMMENT '角色ID',
  PRIMARY KEY (`user_id`, `role_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='用户角色关联表';

CREATE TABLE IF NOT EXISTS `sys_role_menu` (
  `role_id` bigint unsigned NOT NULL COMMENT '角色ID',
  `menu_id` bigint unsigned NOT NULL COMMENT '菜单ID',
  PRIMARY KEY (`role_id`, `menu_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='角色菜单关联表';

CREATE TABLE IF NOT EXISTS `sys_user_post` (
  `user_id` bigint unsigned NOT NULL COMMENT '用户ID',
  `post_id` bigint unsigned NOT NULL COMMENT '岗位ID',
  PRIMARY KEY (`user_id`, `post_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='用户岗位关联表';

-- ============================================================
-- 二、支付与会员表
-- ============================================================

-- 支付订单表
CREATE TABLE IF NOT EXISTS `pay_order` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `tenant_id` bigint unsigned DEFAULT 0 COMMENT '租户ID',
  `order_no` varchar(64) NOT NULL COMMENT '订单号',
  `trade_no` varchar(128) DEFAULT '' COMMENT '第三方交易号',
  `subject` varchar(200) NOT NULL COMMENT '订单标题',
  `body` varchar(500) DEFAULT '' COMMENT '订单描述',
  `amount` bigint NOT NULL COMMENT '金额(分)',
  `currency` varchar(10) DEFAULT 'CNY' COMMENT '币种',
  `channel` varchar(20) NOT NULL COMMENT '支付渠道 wechat/alipay',
  `status` tinyint DEFAULT 0 COMMENT '0待支付 1已支付 2已关闭 3已退款',
  `paid_at` datetime DEFAULT NULL COMMENT '支付时间',
  `refund_at` datetime DEFAULT NULL COMMENT '退款时间',
  `refund_amt` bigint DEFAULT 0 COMMENT '退款金额(分)',
  `open_id` varchar(128) DEFAULT '' COMMENT '用户OpenID',
  `notify_url` varchar(500) DEFAULT '' COMMENT '回调地址',
  `extra` text COMMENT '扩展信息',
  `pay_info` text COMMENT '支付参数',
  `raw_notify` text COMMENT '原始回调数据',
  `create_by` bigint unsigned DEFAULT 0 COMMENT '创建者',
  `update_by` bigint unsigned DEFAULT 0 COMMENT '更新者',
  `created_at` datetime DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `updated_at` datetime DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  `deleted_at` datetime DEFAULT NULL COMMENT '删除时间',
  `remark` varchar(500) DEFAULT '' COMMENT '备注',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_order_no` (`order_no`),
  KEY `idx_trade_no` (`trade_no`),
  KEY `idx_tenant_id` (`tenant_id`),
  KEY `idx_status` (`status`),
  KEY `idx_deleted_at` (`deleted_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='支付订单表';

-- 会员表
CREATE TABLE IF NOT EXISTS `pay_member` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `tenant_id` bigint unsigned DEFAULT 0 COMMENT '租户ID',
  `member_no` varchar(32) DEFAULT '' COMMENT '会员编号',
  `username` varchar(64) DEFAULT '' COMMENT '用户名',
  `nickname` varchar(64) DEFAULT '' COMMENT '昵称',
  `avatar` varchar(512) DEFAULT '' COMMENT '头像',
  `phone` varchar(20) DEFAULT '' COMMENT '手机号',
  `gender` tinyint DEFAULT 0 COMMENT '性别 0未知 1男 2女',
  `birthday` date DEFAULT NULL COMMENT '出生日期',
  `level_id` bigint unsigned DEFAULT 0 COMMENT '等级ID',
  `status` tinyint DEFAULT 1 COMMENT '状态 0停用 1正常',
  `points` bigint DEFAULT 0 COMMENT '积分',
  `wechat_openid` varchar(128) DEFAULT '' COMMENT '微信小程序openid',
  `register_time` datetime DEFAULT CURRENT_TIMESTAMP COMMENT '注册时间',
  `last_visit_time` datetime DEFAULT NULL COMMENT '最后访问时间',
  `create_by` bigint unsigned DEFAULT 0 COMMENT '创建者',
  `update_by` bigint unsigned DEFAULT 0 COMMENT '更新者',
  `created_at` datetime DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `updated_at` datetime DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  `deleted_at` datetime DEFAULT NULL COMMENT '删除时间',
  `remark` varchar(500) DEFAULT '' COMMENT '备注',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_phone` (`phone`),
  UNIQUE KEY `uk_member_no` (`member_no`),
  KEY `idx_tenant_id` (`tenant_id`),
  KEY `idx_level_id` (`level_id`),
  KEY `idx_wechat_openid` (`wechat_openid`),
  KEY `idx_deleted_at` (`deleted_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='会员表';

-- 会员等级表
CREATE TABLE IF NOT EXISTS `pay_member_level` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `tenant_id` bigint unsigned DEFAULT 0 COMMENT '租户ID',
  `name` varchar(64) NOT NULL COMMENT '等级名称',
  `min_points` bigint DEFAULT 0 COMMENT '最低积分',
  `discount` decimal(3,1) DEFAULT 10.0 COMMENT '折扣 10=不打折 8=八折',
  `icon` varchar(256) DEFAULT '' COMMENT '等级图标',
  `sort` int DEFAULT 0 COMMENT '排序',
  `status` tinyint DEFAULT 1 COMMENT '状态 0停用 1正常',
  `create_by` bigint unsigned DEFAULT 0 COMMENT '创建者',
  `update_by` bigint unsigned DEFAULT 0 COMMENT '更新者',
  `created_at` datetime DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `updated_at` datetime DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  `deleted_at` datetime DEFAULT NULL COMMENT '删除时间',
  `remark` varchar(500) DEFAULT '' COMMENT '备注',
  PRIMARY KEY (`id`),
  KEY `idx_tenant_id` (`tenant_id`),
  KEY `idx_deleted_at` (`deleted_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='会员等级表';

-- 会员标签表
CREATE TABLE IF NOT EXISTS `pay_member_tag` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `tenant_id` bigint unsigned DEFAULT 0 COMMENT '租户ID',
  `name` varchar(64) NOT NULL COMMENT '标签名称',
  `color` varchar(20) DEFAULT '#409eff' COMMENT '标签颜色',
  `sort` int DEFAULT 0 COMMENT '排序',
  `status` tinyint DEFAULT 1 COMMENT '状态 0停用 1正常',
  `create_by` bigint unsigned DEFAULT 0 COMMENT '创建者',
  `update_by` bigint unsigned DEFAULT 0 COMMENT '更新者',
  `created_at` datetime DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `updated_at` datetime DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  `deleted_at` datetime DEFAULT NULL COMMENT '删除时间',
  `remark` varchar(500) DEFAULT '' COMMENT '备注',
  PRIMARY KEY (`id`),
  KEY `idx_tenant_id` (`tenant_id`),
  KEY `idx_deleted_at` (`deleted_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='会员标签表';

-- 会员标签关联表
CREATE TABLE IF NOT EXISTS `pay_member_tag_rel` (
  `member_id` bigint unsigned NOT NULL,
  `tag_id` bigint unsigned NOT NULL,
  PRIMARY KEY (`member_id`,`tag_id`),
  KEY `idx_tag_id` (`tag_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='会员标签关联表';

-- 积分明细表
CREATE TABLE IF NOT EXISTS `pay_points_log` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `tenant_id` bigint unsigned DEFAULT 0 COMMENT '租户ID',
  `member_id` bigint unsigned NOT NULL COMMENT '会员ID',
  `points_change` bigint NOT NULL COMMENT '变更积分',
  `type` tinyint DEFAULT 1 COMMENT '类型 1获取 2消费',
  `source` varchar(64) DEFAULT '' COMMENT '来源',
  `order_no` varchar(64) DEFAULT '' COMMENT '关联订单号',
  `create_by` bigint unsigned DEFAULT 0 COMMENT '创建者',
  `update_by` bigint unsigned DEFAULT 0 COMMENT '更新者',
  `created_at` datetime DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `updated_at` datetime DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  `deleted_at` datetime DEFAULT NULL COMMENT '删除时间',
  `remark` varchar(500) DEFAULT '' COMMENT '备注',
  PRIMARY KEY (`id`),
  KEY `idx_tenant_id` (`tenant_id`),
  KEY `idx_member_id` (`member_id`),
  KEY `idx_deleted_at` (`deleted_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='积分明细表';

-- ============================================================
-- 三、初始数据
-- ============================================================

-- 默认部门
INSERT IGNORE INTO `sys_dept` (`id`, `parent_id`, `name`, `sort`, `status`) VALUES
(1, 0, 'Gin-Admin科技', 0, 1),
(2, 1, '研发部门', 1, 1),
(3, 1, '市场部门', 2, 1);

-- 默认角色
INSERT IGNORE INTO `sys_role` (`id`, `name`, `code`, `sort`, `status`, `data_scope`) VALUES
(1, '超级管理员', 'admin', 0, 1, 1),
(2, '普通角色', 'common', 1, 1, 2);

-- 默认管理员 (用户名: admin, 密码: admin123)
INSERT IGNORE INTO `sys_user` (`id`, `username`, `password`, `nickname`, `status`, `dept_id`, `create_by`, `update_by`) VALUES
(1, 'admin', '$2a$10$7JB720yubVSZvUI0rEqK/.VqGOZTH.ulu33dHOiBE8ByOhJIrdAu2', '超级管理员', 1, 1, 0, 0);

-- 管理员角色关联
INSERT IGNORE INTO `sys_user_role` (`user_id`, `role_id`) VALUES (1, 1);

-- 菜单数据
INSERT IGNORE INTO `sys_menu` (`id`, `parent_id`, `name`, `path`, `component`, `icon`, `title`, `type`, `permission`, `sort`, `visible`, `status`, `is_cache`, `is_external`, `create_by`, `update_by`) VALUES
-- 权限管理
(1, 0, 'System', '/system', '', 'Lock', '权限管理', 0, '', 2, 1, 1, 1, 0, 1, 1),
(2, 1, 'User', 'user', 'system/user/index', 'User', '管理员', 1, 'system:user:list', 1, 1, 1, 1, 0, 1, 1),
(3, 1, 'Role', 'role', 'system/role/index', 'UserFilled', '角色管理', 1, 'system:role:list', 2, 1, 1, 1, 0, 1, 1),
(4, 1, 'Menu', 'menu', 'system/menu/index', 'Menu', '菜单管理', 1, 'system:menu:list', 3, 1, 1, 1, 0, 1, 1),
(5, 1, 'Dept', 'dept', 'system/dept/index', 'OfficeBuilding', '部门管理', 1, 'system:dept:list', 4, 1, 1, 1, 0, 1, 1),
(6, 1, 'Post', 'post', 'system/post/index', 'Briefcase', '岗位管理', 1, 'system:post:list', 5, 0, 1, 1, 0, 1, 1),
(7, 1, 'Config', 'config', 'system/config/index', 'Tools', '参数管理', 1, 'system:config:list', 6, 1, 1, 1, 0, 1, 1),
(8, 1, 'Dict', 'dict', 'system/dict/index', 'Collection', '数据字典', 1, 'system:dict:list', 7, 1, 1, 1, 0, 1, 1),
(9, 1, 'Log', 'log', 'system/log/index', 'Document', '日志管理', 1, 'system:log:list', 8, 1, 1, 1, 0, 1, 1),
(15, 1, 'File', 'file', 'system/file/index', 'Upload', '附件管理', 1, 'system:file:list', 9, 1, 1, 1, 0, 1, 1),
-- 系统设置
(10, 0, 'Settings', '/settings', '', 'Tools', '系统设置', 0, '', 3, 1, 1, 1, 0, 1, 1),
(11, 10, 'SiteSettings', 'site', 'settings/site', 'Position', '网站设置', 1, 'system:settings:site', 1, 1, 1, 1, 0, 1, 1),
(12, 10, 'PaymentSettings', 'payment', 'settings/payment', 'Wallet', '支付设置', 1, 'system:settings:payment', 2, 1, 1, 1, 0, 1, 1),
(13, 10, 'OSSSettings', 'oss', 'settings/oss', 'FolderOpened', 'OSS存储设置', 1, 'system:settings:oss', 3, 1, 1, 1, 0, 1, 1),
(14, 10, 'SMSSettings', 'sms', 'settings/sms', 'Message', '短信设置', 1, 'system:settings:sms', 4, 1, 1, 1, 0, 1, 1),
(150, 10, 'Agreement', 'agreement', 'settings/agreement', 'Document', '协议管理', 1, 'system:agreement:list', 5, 1, 1, 1, 0, 1, 1),
-- 支付管理
(16, 0, 'Payment', '/payment', '', 'Wallet', '支付管理', 0, '', 4, 1, 1, 1, 0, 1, 1),
(17, 16, 'PayOrder', 'order', 'payment/index', 'Document', '订单管理', 1, 'payment:order:list', 1, 1, 1, 1, 0, 1, 1),
(18, 16, 'PayTest', 'create', 'payment/create', 'Position', '支付测试', 1, 'payment:order:create', 2, 1, 1, 1, 0, 1, 1),
-- 会员管理
(20, 0, 'Member', '/member', '', 'User', '会员管理', 0, '', 5, 1, 1, 1, 0, 1, 1),
(21, 20, 'MemberList', 'list', 'member/index', 'User', '会员列表', 1, 'member:list', 1, 1, 1, 1, 0, 1, 1),
(22, 20, 'MemberLevel', 'level', 'member/level', 'TrendCharts', '会员等级', 1, 'member:level:list', 2, 1, 1, 1, 0, 1, 1),
(23, 20, 'MemberTag', 'tag', 'member/tag', 'PriceTag', '会员标签', 1, 'member:tag:list', 3, 1, 1, 1, 0, 1, 1),
(24, 20, 'MemberPoints', 'points', 'member/points', 'Coin', '积分明细', 1, 'member:points:list', 4, 1, 1, 1, 0, 1, 1),
-- 按钮权限
(100, 2, 'UserAdd', '', '', '', '新增', 2, 'system:user:add', 1, 1, 1, 1, 0, 1, 1),
(101, 2, 'UserEdit', '', '', '', '编辑', 2, 'system:user:edit', 2, 1, 1, 1, 0, 1, 1),
(102, 2, 'UserDelete', '', '', '', '删除', 2, 'system:user:delete', 3, 1, 1, 1, 0, 1, 1),
(200, 3, 'RoleAdd', '', '', '', '新增', 2, 'system:role:add', 1, 1, 1, 1, 0, 1, 1),
(201, 3, 'RoleEdit', '', '', '', '编辑', 2, 'system:role:edit', 2, 1, 1, 1, 0, 1, 1),
(202, 3, 'RoleDelete', '', '', '', '删除', 2, 'system:role:delete', 3, 1, 1, 1, 0, 1, 1);

-- 超级管理员菜单权限
INSERT IGNORE INTO `sys_role_menu` (`role_id`, `menu_id`) VALUES
(1, 1), (1, 2), (1, 3), (1, 4), (1, 5), (1, 6), (1, 7), (1, 8), (1, 9), (1, 15),
(1, 10), (1, 11), (1, 12), (1, 13), (1, 14), (1, 150),
(1, 16), (1, 17), (1, 18),
(1, 20), (1, 21), (1, 22), (1, 23), (1, 24),
(1, 100), (1, 101), (1, 102), (1, 200), (1, 201), (1, 202);

-- 系统配置（站点设置、支付、OSS、短信）
INSERT IGNORE INTO `sys_config` (`name`, `config_key`, `value`, `type`, `create_by`, `update_by`, `created_at`, `updated_at`) VALUES
('网站名称', 'site.name', '', 1, 1, 1, NOW(), NOW()),
('网站标题', 'site.title', '', 1, 1, 1, NOW(), NOW()),
('Logo', 'site.logo', '', 1, 1, 1, NOW(), NOW()),
('网站描述', 'site.description', '', 1, 1, 1, NOW(), NOW()),
('版权信息', 'site.copyright', '', 1, 1, 1, NOW(), NOW()),
('ICP备案号', 'site.icp', '', 1, 1, 1, NOW(), NOW()),
('ICP备案链接', 'site.icpLink', '', 1, 1, 1, NOW(), NOW()),
('公安备案号', 'site.policeRecord', '', 1, 1, 1, NOW(), NOW()),
('公安备案链接', 'site.policeRecordLink', '', 1, 1, 1, NOW(), NOW()),
('联系电话', 'site.phone', '', 1, 1, 1, NOW(), NOW()),
('会员ID位数', 'site.memberIdDigits', '6', 1, 1, 1, NOW(), NOW()),
('微信AppID', 'pay.wechat_app_id', '', 1, 1, 1, NOW(), NOW()),
('微信商户号', 'pay.wechat_mch_id', '', 1, 1, 1, NOW(), NOW()),
('微信密钥', 'pay.wechat_key', '', 1, 1, 1, NOW(), NOW()),
('微信证书序列号', 'pay.wechat_serial_no', '', 1, 1, 1, NOW(), NOW()),
('微信PEM证书', 'pay.wechat_cert_pem', '', 1, 1, 1, NOW(), NOW()),
('微信证书密钥', 'pay.wechat_key_pem', '', 1, 1, 1, NOW(), NOW()),
('支付宝AppID', 'pay.alipay_app_id', '', 1, 1, 1, NOW(), NOW()),
('支付宝密钥', 'pay.alipay_key', '', 1, 1, 1, NOW(), NOW()),
('支付宝公钥', 'pay.alipay_public_key', '', 1, 1, 1, NOW(), NOW()),
('支付回调地址', 'pay.notify_url', '', 1, 1, 1, NOW(), NOW()),
('支付完成跳转', 'pay.return_url', '', 1, 1, 1, NOW(), NOW()),
('存储类型', 'oss.type', 'local', 1, 1, 1, NOW(), NOW()),
('Endpoint', 'oss.endpoint', '', 1, 1, 1, NOW(), NOW()),
('Bucket', 'oss.bucket', '', 1, 1, 1, NOW(), NOW()),
('AccessKey', 'oss.access_key', '', 1, 1, 1, NOW(), NOW()),
('SecretKey', 'oss.secret_key', '', 1, 1, 1, NOW(), NOW()),
('自定义域名', 'oss.domain', '', 1, 1, 1, NOW(), NOW()),
('短信状态', 'sms.status', '1', 1, 1, 1, NOW(), NOW()),
('短信接口', 'sms.provider', 'aliyun', 1, 1, 1, NOW(), NOW()),
('短信AccessKey ID', 'sms.access_key', '', 1, 1, 1, NOW(), NOW()),
('短信AccessKey Secret', 'sms.secret_key', '', 1, 1, 1, NOW(), NOW()),
('短信签名', 'sms.sign_name', '', 1, 1, 1, NOW(), NOW()),
('短信验证码模板', 'sms.tpl_verify_code', '', 1, 1, 1, NOW(), NOW()),
('短信验证码模板-启用', 'sms.tpl_verify_code_enabled', '1', 1, 1, 1, NOW(), NOW());
