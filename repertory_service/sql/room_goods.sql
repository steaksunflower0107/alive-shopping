CREATE TABLE `platform_room_goods` (
      `id` BIGINT(20) UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY COMMENT '主键',
      `create_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
      `create_by` VARCHAR(64) NOT NULL DEFAULT '' COMMENT '创建者',
      `update_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '更新时间',
      `update_by` VARCHAR(64) NOT NULL DEFAULT '' COMMENT '更新者',
      `version` SMALLINT(5) UNSIGNED NOT NULL DEFAULT '0' COMMENT '乐观锁版本号',
      `is_del` tinyint(4) UNSIGNED NOT NULL DEFAULT '0' COMMENT '是否删除：0正常1删除',
      `room_id` BIGINT(20) UNSIGNED NOT NULL DEFAULT '0' COMMENT '直播间/主播id',
      `goods_id` BIGINT(20) UNSIGNED NOT NULL DEFAULT '0' COMMENT '商品id',
      `weight` BIGINT(20) NOT NULL DEFAULT '1000' COMMENT '排序权重',
      `is_current` tinyint(4) UNSIGNED NOT NULL DEFAULT '0' COMMENT '是否当前讲解中：0不是1是',
      UNIQUE (room_id, goods_id),
      INDEX (is_del)
) ENGINE=INNODB DEFAULT CHARSET=utf8mb4 COMMENT = '直播间商品表';