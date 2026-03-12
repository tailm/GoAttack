-- 检测引擎和预警通知系统数据库迁移脚本
-- 版本: 1.0
-- 创建时间: 2026-03-12
-- 描述: 添加检测引擎和预警通知相关表结构

-- 1. 检测任务表
CREATE TABLE IF NOT EXISTS detection_tasks (
    id SERIAL PRIMARY KEY,
    name VARCHAR(200) NOT NULL,
    description TEXT,
    targets JSONB NOT NULL, -- 目标列表
    target_type VARCHAR(50) NOT NULL, -- 目标类型: host, network, web, api
    scan_type VARCHAR(50) NOT NULL, -- 扫描类型: full, quick, custom
    rules JSONB, -- 检测规则列表
    config JSONB, -- 任务配置
    status VARCHAR(50) DEFAULT 'pending', -- 状态: pending, running, completed, failed, stopped
    progress INTEGER DEFAULT 0, -- 进度百分比
    created_by VARCHAR(100),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    started_at TIMESTAMP,
    completed_at TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- 为检测任务表创建索引
CREATE INDEX IF NOT EXISTS idx_detection_tasks_status ON detection_tasks(status);
CREATE INDEX IF NOT EXISTS idx_detection_tasks_created_by ON detection_tasks(created_by);
CREATE INDEX IF NOT EXISTS idx_detection_tasks_created_at ON detection_tasks(created_at);
CREATE INDEX IF NOT EXISTS idx_detection_tasks_target_type ON detection_tasks(target_type);

-- 2. 检测结果表（扩展原有表）
CREATE TABLE IF NOT EXISTS detection_results (
    id SERIAL PRIMARY KEY,
    task_id INTEGER NOT NULL REFERENCES detection_tasks(id) ON DELETE CASCADE,
    target VARCHAR(500) NOT NULL,
    target_type VARCHAR(50) NOT NULL,
    rule_id INTEGER,
    rule_name VARCHAR(200),
    rule_type VARCHAR(50),
    severity VARCHAR(20) NOT NULL,
    status VARCHAR(50) NOT NULL, -- 状态: found, not_found, error
    evidence JSONB, -- 证据信息
    details TEXT, -- 详细信息
    confidence INTEGER CHECK (confidence >= 0 AND confidence <= 100),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- 为检测结果表创建索引
CREATE INDEX IF NOT EXISTS idx_detection_results_task ON detection_results(task_id);
CREATE INDEX IF NOT EXISTS idx_detection_results_target ON detection_results(target);
CREATE INDEX IF NOT EXISTS idx_detection_results_severity ON detection_results(severity);
CREATE INDEX IF NOT EXISTS idx_detection_results_status ON detection_results(status);
CREATE INDEX IF NOT EXISTS idx_detection_results_created_at ON detection_results(created_at);

-- 3. 检测规则表（扩展原有表）
CREATE TABLE IF NOT EXISTS detection_rules (
    id SERIAL PRIMARY KEY,
    name VARCHAR(200) NOT NULL,
    description TEXT,
    type VARCHAR(50) NOT NULL, -- 规则类型: port, service, version, cve, custom
    condition TEXT NOT NULL, -- 检测条件
    action VARCHAR(50) NOT NULL, -- 动作: alert, block, log, report
    severity VARCHAR(20) NOT NULL, -- 严重程度: critical, high, medium, low, info
    enabled BOOLEAN DEFAULT true,
    priority INTEGER DEFAULT 0,
    tags JSONB, -- 标签
    config JSONB, -- 规则配置
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- 为检测规则表创建索引
CREATE INDEX IF NOT EXISTS idx_detection_rules_type ON detection_rules(type);
CREATE INDEX IF NOT EXISTS idx_detection_rules_severity ON detection_rules(severity);
CREATE INDEX IF NOT EXISTS idx_detection_rules_enabled ON detection_rules(enabled);
CREATE INDEX IF NOT EXISTS idx_detection_rules_priority ON detection_rules(priority);

-- 4. 预警表
CREATE TABLE IF NOT EXISTS alerts (
    id SERIAL PRIMARY KEY,
    title VARCHAR(200) NOT NULL,
    description TEXT,
    type VARCHAR(50) NOT NULL, -- 预警类型: vulnerability, threat, system, custom
    severity VARCHAR(20) NOT NULL, -- 严重程度: critical, high, medium, low, info
    status VARCHAR(50) DEFAULT 'pending', -- 状态: pending, processing, resolved, closed
    source VARCHAR(100) NOT NULL, -- 来源: detection, intelligence, manual, system
    source_id VARCHAR(100), -- 来源ID
    target VARCHAR(500), -- 目标
    target_type VARCHAR(50), -- 目标类型: host, network, web, api
    details JSONB, -- 详细信息
    metadata JSONB, -- 元数据
    created_by VARCHAR(100),
    assigned_to VARCHAR(100),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    resolved_at TIMESTAMP,
    closed_at TIMESTAMP
);

-- 为预警表创建索引
CREATE INDEX IF NOT EXISTS idx_alerts_type ON alerts(type);
CREATE INDEX IF NOT EXISTS idx_alerts_severity ON alerts(severity);
CREATE INDEX IF NOT EXISTS idx_alerts_status ON alerts(status);
CREATE INDEX IF NOT EXISTS idx_alerts_source ON alerts(source);
CREATE INDEX IF NOT EXISTS idx_alerts_created_by ON alerts(created_by);
CREATE INDEX IF NOT EXISTS idx_alerts_assigned_to ON alerts(assigned_to);
CREATE INDEX IF NOT EXISTS idx_alerts_created_at ON alerts(created_at);

-- 5. 预警规则表
CREATE TABLE IF NOT EXISTS alert_rules (
    id SERIAL PRIMARY KEY,
    name VARCHAR(200) NOT NULL,
    description TEXT,
    type VARCHAR(50) NOT NULL, -- 规则类型: vulnerability, threat, system, custom
    severity VARCHAR(20) NOT NULL, -- 触发严重程度
    condition TEXT NOT NULL, -- 触发条件
    action VARCHAR(50) NOT NULL, -- 触发动作: email, webhook, sms, notification
    config JSONB, -- 动作配置
    enabled BOOLEAN DEFAULT true,
    priority INTEGER DEFAULT 0,
    tags JSONB, -- 标签
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- 为预警规则表创建索引
CREATE INDEX IF NOT EXISTS idx_alert_rules_type ON alert_rules(type);
CREATE INDEX IF NOT EXISTS idx_alert_rules_severity ON alert_rules(severity);
CREATE INDEX IF NOT EXISTS idx_alert_rules_enabled ON alert_rules(enabled);
CREATE INDEX IF NOT EXISTS idx_alert_rules_priority ON alert_rules(priority);

-- 6. 通知记录表
CREATE TABLE IF NOT EXISTS notifications (
    id SERIAL PRIMARY KEY,
    alert_id INTEGER NOT NULL REFERENCES alerts(id) ON DELETE CASCADE,
    type VARCHAR(50) NOT NULL, -- 通知类型: email, webhook, sms, notification
    status VARCHAR(50) DEFAULT 'pending', -- 状态: pending, sent, failed, delivered
    recipient VARCHAR(500), -- 接收者
    content JSONB, -- 通知内容
    error TEXT, -- 错误信息
    sent_at TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- 为通知记录表创建索引
CREATE INDEX IF NOT EXISTS idx_notifications_alert ON notifications(alert_id);
CREATE INDEX IF NOT EXISTS idx_notifications_type ON notifications(type);
CREATE INDEX IF NOT EXISTS idx_notifications_status ON notifications(status);
CREATE INDEX IF NOT EXISTS idx_notifications_created_at ON notifications(created_at);

-- 7. 检测统计表
CREATE TABLE IF NOT EXISTS detection_stats (
    id SERIAL PRIMARY KEY,
    date DATE NOT NULL,
    task_count INTEGER DEFAULT 0,
    completed_count INTEGER DEFAULT 0,
    failed_count INTEGER DEFAULT 0,
    running_count INTEGER DEFAULT 0,
    total_targets INTEGER DEFAULT 0,
    total_findings INTEGER DEFAULT 0,
    critical_findings INTEGER DEFAULT 0,
    high_findings INTEGER DEFAULT 0,
    medium_findings INTEGER DEFAULT 0,
    low_findings INTEGER DEFAULT 0,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(date)
);

-- 为检测统计表创建索引
CREATE INDEX IF NOT EXISTS idx_detection_stats_date ON detection_stats(date);

-- 8. 预警统计表
CREATE TABLE IF NOT EXISTS alert_stats (
    id SERIAL PRIMARY KEY,
    date DATE NOT NULL,
    total_alerts INTEGER DEFAULT 0,
    critical_alerts INTEGER DEFAULT 0,
    high_alerts INTEGER DEFAULT 0,
    medium_alerts INTEGER DEFAULT 0,
    low_alerts INTEGER DEFAULT 0,
    pending_alerts INTEGER DEFAULT 0,
    processing_alerts INTEGER DEFAULT 0,
    resolved_alerts INTEGER DEFAULT 0,
    closed_alerts INTEGER DEFAULT 0,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(date)
);

-- 为预警统计表创建索引
CREATE INDEX IF NOT EXISTS idx_alert_stats_date ON alert_stats(date);

-- 9. 创建视图：检测任务统计视图
CREATE OR REPLACE VIEW detection_task_stats AS
SELECT 
    DATE(created_at) as date,
    target_type,
    status,
    COUNT(*) as task_count,
    AVG(progress) as avg_progress,
    MIN(created_at) as first_task,
    MAX(created_at) as last_task
FROM detection_tasks
GROUP BY DATE(created_at), target_type, status
ORDER BY date DESC, target_type, status;

-- 10. 创建视图：预警统计视图
CREATE OR REPLACE VIEW alert_summary_stats AS
SELECT 
    DATE(created_at) as date,
    type,
    severity,
    status,
    COUNT(*) as alert_count,
    COUNT(CASE WHEN assigned_to IS NOT NULL THEN 1 END) as assigned_count,
    AVG(EXTRACT(EPOCH FROM (COALESCE(resolved_at, closed_at, CURRENT_TIMESTAMP) - created_at))) as avg_response_time
FROM alerts
GROUP BY DATE(created_at), type, severity, status
ORDER BY date DESC, type, severity, status;

-- 11. 创建函数：获取检测任务统计
CREATE OR REPLACE FUNCTION get_detection_task_stats(start_date DATE, end_date DATE)
RETURNS TABLE(
    date DATE,
    target_type VARCHAR(50),
    status VARCHAR(50),
    task_count BIGINT,
    avg_progress DECIMAL(5,2),
    first_task TIMESTAMP,
    last_task TIMESTAMP
) AS $$
BEGIN
    RETURN QUERY
    SELECT 
        DATE(dt.created_at) as date,
        dt.target_type,
        dt.status,
        COUNT(*) as task_count,
        AVG(dt.progress)::DECIMAL(5,2) as avg_progress,
        MIN(dt.created_at) as first_task,
        MAX(dt.created_at) as last_task
    FROM detection_tasks dt
    WHERE DATE(dt.created_at) BETWEEN start_date AND end_date
    GROUP BY DATE(dt.created_at), dt.target_type, dt.status
    ORDER BY date DESC, dt.target_type, dt.status;
END;
$$ LANGUAGE plpgsql;

-- 12. 创建函数：获取预警统计
CREATE OR REPLACE FUNCTION get_alert_stats(start_date DATE, end_date DATE)
RETURNS TABLE(
    date DATE,
    type VARCHAR(50),
    severity VARCHAR(20),
    status VARCHAR(50),
    alert_count BIGINT,
    assigned_count BIGINT,
    avg_response_time DECIMAL(10,2)
) AS $$
BEGIN
    RETURN QUERY
    SELECT 
        DATE(a.created_at) as date,
        a.type,
        a.severity,
        a.status,
        COUNT(*) as alert_count,
        COUNT(CASE WHEN a.assigned_to IS NOT NULL THEN 1 END) as assigned_count,
        AVG(EXTRACT(EPOCH FROM (COALESCE(a.resolved_at, a.closed_at, CURRENT_TIMESTAMP) - a.created_at)))::DECIMAL(10,2) as avg_response_time
    FROM alerts a
    WHERE DATE(a.created_at) BETWEEN start_date AND end_date
    GROUP BY DATE(a.created_at), a.type, a.severity, a.status
    ORDER BY date DESC, a.type, a.severity, a.status;
END;
$$ LANGUAGE plpgsql;

-- 13. 创建触发器：自动更新updated_at时间戳
CREATE TRIGGER update_detection_tasks_updated_at 
    BEFORE UPDATE ON detection_tasks 
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_detection_results_updated_at 
    BEFORE UPDATE ON detection_results 
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_detection_rules_updated_at 
    BEFORE UPDATE ON detection_rules 
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_alerts_updated_at 
    BEFORE UPDATE ON alerts 
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_alert_rules_updated_at 
    BEFORE UPDATE ON alert_rules 
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- 14. 创建索引：优化查询性能
CREATE INDEX IF NOT EXISTS idx_detection_tasks_targets ON detection_tasks USING GIN(targets);
CREATE INDEX IF NOT EXISTS idx_detection_tasks_rules ON detection_tasks USING GIN(rules);
CREATE INDEX IF NOT EXISTS idx_detection_tasks_config ON detection_tasks USING GIN(config);
CREATE INDEX IF NOT EXISTS idx_detection_results_evidence ON detection_results USING GIN(evidence);
CREATE INDEX IF NOT EXISTS idx_detection_rules_tags ON detection_rules USING GIN(tags);
CREATE INDEX IF NOT EXISTS idx_detection_rules_config ON detection_rules USING GIN(config);
CREATE INDEX IF NOT EXISTS idx_alerts_details ON alerts USING GIN(details);
CREATE INDEX IF NOT EXISTS idx_alerts_metadata ON alerts USING GIN(metadata);
CREATE INDEX IF NOT EXISTS idx_alert_rules_tags ON alert_rules USING GIN(tags);
CREATE INDEX IF NOT EXISTS idx_alert_rules_config ON alert_rules USING GIN(config);
CREATE INDEX IF NOT EXISTS idx_notifications_content ON notifications USING GIN(content);

-- 15. 添加注释
COMMENT ON TABLE detection_tasks IS '检测任务表，存储检测任务信息';
COMMENT ON TABLE detection_results IS '检测结果表，存储检测结果信息';
COMMENT ON TABLE detection_rules IS '检测规则表，存储检测规则信息';
COMMENT ON TABLE alerts IS '预警表，存储预警信息';
COMMENT ON TABLE alert_rules IS '预警规则表，存储预警规则信息';
COMMENT ON TABLE notifications IS '通知记录表，存储通知发送记录';
COMMENT ON TABLE detection_stats IS '检测统计表，存储检测统计信息';
COMMENT ON TABLE alert_stats IS '预警统计表，存储预警统计信息';
COMMENT ON VIEW detection_task_stats IS '检测任务统计视图';
COMMENT ON VIEW alert_summary_stats IS '预警统计视图';
COMMENT ON FUNCTION get_detection_task_stats IS '获取检测任务统计函数';
COMMENT ON FUNCTION get_alert_stats IS '获取预警统计函数';

-- 16. 插入默认检测规则（如果不存在）
INSERT INTO detection_rules (name, description, type, condition, action, severity, enabled, priority, tags, config)
SELECT * FROM (VALUES
    ('SSH弱密码检测', '检测SSH服务是否存在弱密码', 'service', 'service == ''ssh''', 'alert', 'high', true, 10, '["ssh", "authentication", "brute-force"]'::jsonb, '{"auto_enabled": true, "category": "default"}'::jsonb),
    ('MySQL未授权访问', '检测MySQL服务是否存在未授权访问漏洞', 'service', 'service == ''mysql''', 'alert', 'critical', true, 5, '["mysql", "database", "unauthorized"]'::jsonb, '{"auto_enabled": true, "category": "default"}'::jsonb),
    ('Redis未授权访问', '检测Redis服务是否存在未授权访问漏洞', 'service', 'service == ''redis''', 'alert', 'critical', true, 5, '["redis", "database", "unauthorized"]'::jsonb, '{"auto_enabled": true, "category": "default"}'::jsonb),
    ('FTP匿名登录', '检测FTP服务是否允许匿名登录', 'service', 'service == ''ftp''', 'alert', 'medium', true, 20, '["ftp", "anonymous", "file-transfer"]'::jsonb, '{"auto_enabled": true, "category": "default"}'::jsonb),
    ('HTTP目录遍历', '检测Web服务是否存在目录遍历漏洞', 'web', 'web.framework == ''Apache'' OR web.server == ''nginx''', 'alert', 'medium', true, 15, '["http", "web", "directory-traversal"]'::jsonb, '{"auto_enabled": true, "category": "default"}'::jsonb),
    ('Telnet服务开放', '检测Telnet服务是否开放（不安全协议）', 'port', 'port == 23', 'alert', 'high', true, 8, '["telnet", "insecure", "protocol"]'::jsonb, '{"auto_enabled": true, "category": "default"}'::jsonb),
    ('SMBv1协议检测', '检测SMBv1协议使用（存在永恒之蓝漏洞风险）', 'service', 'service == ''smb'' AND version < ''2.0''', 'alert', 'critical', true, 3, '["smb", "eternalblue", "windows"]'::jsonb, '{"auto_enabled": true, "category": "default"}'::jsonb),
    ('RDP服务开放', '检测RDP服务是否开放（远程桌面协议）', 'port', 'port == 3389', 'alert', 'medium', true, 12, '["rdp", "remote-desktop", "windows"]'::jsonb, '{"auto_enabled": true, "category": "default"}'::jsonb),
    ('Elasticsearch未授权访问', '检测Elasticsearch服务是否存在未授权访问漏洞', 'service', 'service == ''elasticsearch''', 'alert', 'critical', true, 6, '["elasticsearch", "search", "unauthorized"]'::jsonb, '{"auto_enabled": true, "category": "default"}'::jsonb),
    ('MongoDB未授权访问', '检测MongoDB服务是否存在未授权访问漏洞', 'service', 'service == ''mongodb''', 'alert', 'critical', true, 7, '["mongodb", "database", "unauthorized"]'::jsonb, '{"auto_enabled": true, "category": "default"}'::jsonb)
) AS v(name, description, type, condition, action, severity, enabled, priority, tags, config)
WHERE NOT EXISTS (SELECT 1 FROM detection_rules WHERE name = v.name);

-- 17. 插入默认预警规则（如果不存在）
INSERT INTO alert_rules (name, description, type, severity, condition, action, enabled, priority, tags, config)
SELECT * FROM (VALUES
    ('高危漏洞预警', '检测到高危漏洞时发送预警', 'vulnerability', 'critical', 'severity == ''critical''', 'notification', true, 1, '["vulnerability", "critical", "alert"]'::jsonb, '{"channels": ["system", "email"], "template": "高危漏洞预警: {{.Title}}"}'::jsonb),
    ('中危漏洞预警', '检测到中危漏洞时发送预警', 'vulnerability', 'high', 'severity == ''high''', 'notification', true, 2, '["vulnerability", "high", "alert"]'::jsonb, '{"channels": ["system"], "template": "中危漏洞预警: {{.Title}}"}'::jsonb),
    ('未授权访问预警', '检测到未授权访问时发送预警', 'threat', 'critical', 'type == ''threat'' AND severity == ''critical''', 'notification', true, 3, '["threat", "unauthorized", "alert"]'::jsonb, '{"channels": ["system", "email", "webhook"], "template": "未授权访问预警: {{.Title}}"}'::jsonb),
    ('端口扫描预警', '检测到端口扫描活动时发送预警', 'threat', 'medium', 'type == ''threat'' AND target_type == ''host''', 'notification', true, 4, '["threat", "port-scan", "alert"]'::jsonb, '{"channels": ["system"], "template": "端口扫描预警: {{.Title}}"}'::jsonb),
    ('系统异常预警', '系统出现异常时发送预警', 'system', 'high', 'type == ''system''', 'notification', true, 5, '["system", "error", "alert"]'::jsonb, '{"channels": ["system", "email"], "template": "系统异常预警: {{.Title}}"}'::jsonb),
    ('Web攻击预警', '检测到Web攻击时发送预警', 'threat', 'high', 'type == ''threat'' AND target_type == ''web''', 'notification', true, 6, '["threat", "web-attack", "alert"]'::jsonb, '{"channels": ["system", "webhook"], "template": "Web攻击预警: {{.Title}}"}'::jsonb),
    ('API安全预警', '检测到API安全问题时发送预警', 'threat', 'medium', 'type == ''threat'' AND target_type == ''api''', 'notification', true, 7, '["threat", "api-security", "alert"]'::jsonb, '{"channels": ["system"], "template": "API安全预警: {{.Title}}"}'::jsonb),
    ('网络攻击预警', '检测到网络攻击时发送预警', 'threat', 'critical', 'type == ''threat'' AND target_type == ''network''', 'notification', true, 8, '["threat", "network-attack", "alert"]'::jsonb, '{"channels": ["system", "email", "sms"], "template": "网络攻击预警: {{.Title}}"}'::jsonb),
    ('自定义规则预警', '自定义规则触发时发送预警', 'custom', 'medium', 'type == ''custom''', 'notification', true, 9, '["custom", "alert"]'::jsonb, '{"channels": ["system"], "template": "自定义规则预警: {{.Title}}"}'::jsonb),
    ('信息泄露预警', '检测到信息泄露时发送预警', 'threat', 'high', 'type == ''threat'' AND severity == ''high''', 'notification', true, 10, '["threat", "information-leak", "alert"]'::jsonb, '{"channels": ["system", "email"], "template": "信息泄露预警: {{.Title}}"}'::jsonb)
) AS v(name, description, type, severity, condition, action, enabled, priority, tags, config)
WHERE NOT EXISTS (SELECT 1 FROM alert_rules WHERE name = v.name);

-- 迁移完成
SELECT '检测引擎和预警通知表结构创建完成，共创建8张表和2个视图' as message;