-- 注意：原 DDL 时间戳列名为 created_at/updated_at，此处统一为 create_at/update_at 以对齐 BaseModel
CREATE TABLE IF NOT EXISTS `wf_graph_record` (
  `id` bigint(20) NOT NULL AUTO_INCREMENT COMMENT '唯一标识，主键ID',
  `graph_id` varchar(64) DEFAULT NULL COMMENT '流程图表id',
  `name` varchar(255) NOT NULL COMMENT '名称',
  `description` varchar(512) NOT NULL COMMENT '描述',
  `graph` text NOT NULL COMMENT '图节点信息',
  `version` int(11) NOT NULL COMMENT '版本号',
  `type` varchar(64) NOT NULL COMMENT '类型：svc_pipe:服务编排 flow_pipe:流程编排',
  `status` int(11) NOT NULL COMMENT '状态 1 草稿 2生效中  3 下线 4 删除',
  `create_by` varchar(255) NOT NULL DEFAULT '' COMMENT '创建用户',
  `update_by` varchar(255) NOT NULL DEFAULT '' COMMENT '最后更新用户',
  `is_delete` tinyint(4) NOT NULL DEFAULT '0' COMMENT '删除标识 0-未删除 1-已删除',
  `create_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `update_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  PRIMARY KEY (`id`),
  KEY `idx_graph_id_version` (`graph_id`, `version`),
  KEY `idx_type_status` (`type`, `status`)
) ENGINE=InnoDB AUTO_INCREMENT=1 DEFAULT CHARSET=utf8mb4 COMMENT='工作流配置记录表';
