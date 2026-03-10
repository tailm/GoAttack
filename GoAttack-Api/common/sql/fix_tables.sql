-- GoAttack PostgreSQL数据库表结构修复脚本
-- 修复缺失的表和列，使数据库结构与代码匹配

-- ============================================
-- 1. 创建dashboard表（用于仪表盘统计数据）
-- ============================================
CREATE TABLE IF NOT EXISTS dashboard (
    id SERIAL PRIMARY KEY,
    total_assets INT DEFAULT 0,
    total_vulnerabilities INT DEFAULT 0,
    total_tasks INT DEFAULT 0,
    total_fingerprints INT DEFAULT 0,
    critical_vulns INT DEFAULT 0,
    high_vulns INT DEFAULT 0,
    medium_vulns INT DEFAULT 0,
    low_vulns INT DEFAULT 0,
    info_vulns INT DEFAULT 0,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

COMMENT ON TABLE dashboard IS '仪表盘统计表';
COMMENT ON COLUMN dashboard.id IS '统计ID（固定为1）';
COMMENT ON COLUMN dashboard.total_assets IS '总资产数';
COMMENT ON COLUMN dashboard.total_vulnerabilities IS '总漏洞数';
COMMENT ON COLUMN dashboard.total_tasks IS '总任务数';
COMMENT ON COLUMN dashboard.total_fingerprints IS '总指纹数';
COMMENT ON COLUMN dashboard.critical_vulns IS '严重漏洞数';
COMMENT ON COLUMN dashboard.high_vulns IS '高危漏洞数';
COMMENT ON COLUMN dashboard.medium_vulns IS '中危漏洞数';
COMMENT ON COLUMN dashboard.low_vulns IS '低危漏洞数';
COMMENT ON COLUMN dashboard.info_vulns IS '信息漏洞数';
COMMENT ON COLUMN dashboard.updated_at IS '更新时间';

-- 插入初始数据
INSERT INTO dashboard (id, total_assets, total_vulnerabilities, total_tasks, total_fingerprints, critical_vulns, high_vulns, medium_vulns, low_vulns, info_vulns)
VALUES (1, 0, 0, 0, 0, 0, 0, 0, 0, 0)
ON CONFLICT (id) DO NOTHING;

-- ============================================
-- 2. 更新task表结构
-- ============================================
-- 添加缺失的列
DO $$ 
BEGIN
    -- 添加type列
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name='task' AND column_name='type') THEN
        ALTER TABLE task ADD COLUMN type VARCHAR(50) DEFAULT '';
    END IF;
    
    -- 添加creator列（如果created_by存在，重命名）
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name='task' AND column_name='creator') THEN
        IF EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name='task' AND column_name='created_by') THEN
            ALTER TABLE task RENAME COLUMN created_by TO creator;
        ELSE
            ALTER TABLE task ADD COLUMN creator VARCHAR(50) DEFAULT '';
        END IF;
    END IF;
    
    -- 添加options列
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name='task' AND column_name='options') THEN
        ALTER TABLE task ADD COLUMN options TEXT DEFAULT '';
    END IF;
    
    -- 添加updated_at列
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name='task' AND column_name='updated_at') THEN
        ALTER TABLE task ADD COLUMN updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP;
    END IF;
    
    -- 将finished_at重命名为completed_at
    IF EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name='task' AND column_name='finished_at') THEN
        ALTER TABLE task RENAME COLUMN finished_at TO completed_at;
    END IF;
    
    -- 如果scan_type存在，将其值复制到type列
    IF EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name='task' AND column_name='scan_type') THEN
        UPDATE task SET type = scan_type WHERE type = '' AND scan_type IS NOT NULL;
    END IF;
END $$;

-- 更新列注释
COMMENT ON COLUMN task.type IS '扫描类型：port/web/vuln等';
COMMENT ON COLUMN task.creator IS '创建者用户名';
COMMENT ON COLUMN task.options IS '扫描选项（JSON格式）';
COMMENT ON COLUMN task.updated_at IS '更新时间';
COMMENT ON COLUMN task.completed_at IS '完成时间';

-- ============================================
-- 3. 更新vulnerability表结构
-- ============================================
-- 添加缺失的列
DO $$ 
BEGIN
    -- 添加cve列
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name='vulnerability' AND column_name='cve') THEN
        ALTER TABLE vulnerability ADD COLUMN cve VARCHAR(255) DEFAULT '';
    END IF;
    
    -- 添加discovered_at列（如果created_at存在，重命名）
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name='vulnerability' AND column_name='discovered_at') THEN
        IF EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name='vulnerability' AND column_name='created_at') THEN
            ALTER TABLE vulnerability RENAME COLUMN created_at TO discovered_at;
        ELSE
            ALTER TABLE vulnerability ADD COLUMN discovered_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP;
        END IF;
    END IF;
    
    -- 添加target列
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name='vulnerability' AND column_name='target') THEN
        ALTER TABLE vulnerability ADD COLUMN target TEXT DEFAULT '';
    END IF;
    
    -- 添加ip列
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name='vulnerability' AND column_name='ip') THEN
        ALTER TABLE vulnerability ADD COLUMN ip VARCHAR(45) DEFAULT '';
    END IF;
    
    -- 添加port列
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name='vulnerability' AND column_name='port') THEN
        ALTER TABLE vulnerability ADD COLUMN port INT DEFAULT 0;
    END IF;
    
    -- 添加service列
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name='vulnerability' AND column_name='service') THEN
        ALTER TABLE vulnerability ADD COLUMN service VARCHAR(100) DEFAULT '';
    END IF;
    
    -- 添加cwe列
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name='vulnerability' AND column_name='cwe') THEN
        ALTER TABLE vulnerability ADD COLUMN cwe VARCHAR(255) DEFAULT '';
    END IF;
    
    -- 添加cvss列
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name='vulnerability' AND column_name='cvss') THEN
        ALTER TABLE vulnerability ADD COLUMN cvss DECIMAL(3,1) DEFAULT 0.0;
    END IF;
    
    -- 添加template_id列
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name='vulnerability' AND column_name='template_id') THEN
        ALTER TABLE vulnerability ADD COLUMN template_id VARCHAR(255) DEFAULT '';
    END IF;
    
    -- 添加template_path列
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name='vulnerability' AND column_name='template_path') THEN
        ALTER TABLE vulnerability ADD COLUMN template_path VARCHAR(500) DEFAULT '';
    END IF;
    
    -- 添加author列
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name='vulnerability' AND column_name='author') THEN
        ALTER TABLE vulnerability ADD COLUMN author VARCHAR(255) DEFAULT '';
    END IF;
    
    -- 添加tags列
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name='vulnerability' AND column_name='tags') THEN
        ALTER TABLE vulnerability ADD COLUMN tags TEXT DEFAULT '';
    END IF;
    
    -- 添加reference列
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name='vulnerability' AND column_name='reference') THEN
        ALTER TABLE vulnerability ADD COLUMN reference TEXT DEFAULT '';
    END IF;
    
    -- 添加metadata列
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name='vulnerability' AND column_name='metadata') THEN
        ALTER TABLE vulnerability ADD COLUMN metadata TEXT DEFAULT '';
    END IF;
    
    -- 添加status列
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name='vulnerability' AND column_name='status') THEN
        ALTER TABLE vulnerability ADD COLUMN status VARCHAR(20) DEFAULT 'new';
    END IF;
    
    -- 添加updated_at列
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name='vulnerability' AND column_name='updated_at') THEN
        ALTER TABLE vulnerability ADD COLUMN updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP;
    END IF;
END $$;

-- 更新列注释
COMMENT ON COLUMN vulnerability.cve IS 'CVE编号';
COMMENT ON COLUMN vulnerability.discovered_at IS '发现时间';
COMMENT ON COLUMN vulnerability.target IS '漏洞目标URL';
COMMENT ON COLUMN vulnerability.ip IS '目标IP地址';
COMMENT ON COLUMN vulnerability.port IS '目标端口';
COMMENT ON COLUMN vulnerability.service IS '服务类型';
COMMENT ON COLUMN vulnerability.cwe IS 'CWE编号';
COMMENT ON COLUMN vulnerability.cvss IS 'CVSS评分';
COMMENT ON COLUMN vulnerability.template_id IS '检测模板ID';
COMMENT ON COLUMN vulnerability.template_path IS '模板路径';
COMMENT ON COLUMN vulnerability.author IS '模板作者';
COMMENT ON COLUMN vulnerability.tags IS '标签（JSON数组）';
COMMENT ON COLUMN vulnerability.reference IS '参考链接（JSON数组）';
COMMENT ON COLUMN vulnerability.metadata IS '其他元数据（JSON格式）';
COMMENT ON COLUMN vulnerability.status IS '状态：new/confirmed/false_positive/fixed';
COMMENT ON COLUMN vulnerability.updated_at IS '更新时间';

-- ============================================
-- 4. 创建缺失的表
-- ============================================
-- 创建asset_scan_result表（如果不存在）
CREATE TABLE IF NOT EXISTS asset_scan_result (
    id SERIAL PRIMARY KEY,
    task_id INT NOT NULL,
    asset_id INT NOT NULL,
    scan_type VARCHAR(50) NOT NULL,
    result TEXT DEFAULT '',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (task_id) REFERENCES task(id) ON DELETE CASCADE,
    FOREIGN KEY (asset_id) REFERENCES asset(id) ON DELETE CASCADE
);

COMMENT ON TABLE asset_scan_result IS '资产扫描结果表';
COMMENT ON COLUMN asset_scan_result.id IS '结果ID';
COMMENT ON COLUMN asset_scan_result.task_id IS '关联的任务ID';
COMMENT ON COLUMN asset_scan_result.asset_id IS '关联的资产ID';
COMMENT ON COLUMN asset_scan_result.scan_type IS '扫描类型';
COMMENT ON COLUMN asset_scan_result.result IS '扫描结果（JSON格式）';
COMMENT ON COLUMN asset_scan_result.created_at IS '创建时间';

-- 创建索引
CREATE INDEX IF NOT EXISTS idx_asset_scan_result_task_id ON asset_scan_result(task_id);
CREATE INDEX IF NOT EXISTS idx_asset_scan_result_asset_id ON asset_scan_result(asset_id);

-- ============================================
-- 5. 更新现有数据
-- ============================================
-- 更新task表中的数据
UPDATE task SET type = 'full' WHERE type = '' AND scan_type = 'full';
UPDATE task SET type = 'quick' WHERE type = '' AND scan_type = 'quick';
UPDATE task SET type = 'custom' WHERE type = '' AND scan_type = 'custom';

-- 如果scan_type列存在，可以删除它
DO $$ 
BEGIN
    IF EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name='task' AND column_name='scan_type') THEN
        ALTER TABLE task DROP COLUMN scan_type;
    END IF;
END $$;

-- ============================================
-- 6. 创建触发器
-- ============================================
-- 为task表创建更新触发器
CREATE OR REPLACE FUNCTION update_task_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ language 'plpgsql';

DROP TRIGGER IF EXISTS update_task_updated_at ON task;
CREATE TRIGGER update_task_updated_at BEFORE UPDATE
    ON task FOR EACH ROW EXECUTE FUNCTION update_task_updated_at();

-- 为vulnerability表创建更新触发器
CREATE OR REPLACE FUNCTION update_vulnerability_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ language 'plpgsql';

DROP TRIGGER IF EXISTS update_vulnerability_updated_at ON vulnerability;
CREATE TRIGGER update_vulnerability_updated_at BEFORE UPDATE
    ON vulnerability FOR EACH ROW EXECUTE FUNCTION update_vulnerability_updated_at();

-- ============================================
-- 7. 输出修复结果
-- ============================================
SELECT '数据库表结构修复完成' as message;