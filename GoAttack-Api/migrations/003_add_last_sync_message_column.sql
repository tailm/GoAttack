-- 添加last_sync_message列到intelligence_source表
-- 版本: 1.1
-- 创建时间: 2026-03-13
-- 描述: 修复intelligence_source表缺少last_sync_message列的问题

-- 1. 添加last_sync_message列
ALTER TABLE intelligence_source 
ADD COLUMN IF NOT EXISTS last_sync_message TEXT;

-- 2. 添加注释
COMMENT ON COLUMN intelligence_source.last_sync_message IS '最后一次同步消息（成功/失败信息）';

-- 3. 更新现有数据（如果有的话）
UPDATE intelligence_source 
SET last_sync_message = '初始化完成'
WHERE last_sync_message IS NULL AND last_sync_status = 'success';

UPDATE intelligence_source 
SET last_sync_message = '等待首次同步'
WHERE last_sync_message IS NULL;