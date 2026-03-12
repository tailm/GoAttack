-- ============================================
-- GoAttack 漏洞扫描系统数据库初始化脚本 (PostgreSQL完整版)
-- 创建时间: 2026-03-10
-- 说明: 该脚本用于首次部署时初始化数据库结构
-- 此版本与MySQL版本完全一致
-- ============================================

-- 创建更新updated_at字段的函数
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ language 'plpgsql';

-- ============================================
-- 1. 用户表 (user)
-- 说明: 存储系统用户信息，包括管理员和普通用户
-- ============================================
CREATE TABLE IF NOT EXISTS "user" (
    id SERIAL PRIMARY KEY,
    username VARCHAR(50) NOT NULL UNIQUE,
    password VARCHAR(255) NOT NULL,
    role VARCHAR(20) DEFAULT 'user',
    avatar VARCHAR(500) DEFAULT '',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_username ON "user"(username);
CREATE INDEX IF NOT EXISTS idx_role ON "user"(role);

COMMENT ON TABLE "user" IS '用户表';
COMMENT ON COLUMN "user".id IS '用户ID';
COMMENT ON COLUMN "user".username IS '用户名，唯一';
COMMENT ON COLUMN "user".password IS '密码哈希值（bcrypt）';
COMMENT ON COLUMN "user".role IS '角色：admin/user';
COMMENT ON COLUMN "user".avatar IS '头像URL';
COMMENT ON COLUMN "user".created_at IS '创建时间';

-- ============================================
-- 2. 系统设置表 (system_settings)
-- 说明: 存储全局系统配置，包括扫描引擎配置、代理设置等
-- ============================================
CREATE TABLE IF NOT EXISTS system_settings (
    id SERIAL PRIMARY KEY,
    
    -- 扫描引擎配置
    network_card VARCHAR(100) DEFAULT '',
    concurrency INT DEFAULT 10,
    timeout INT DEFAULT 10,
    retries INT DEFAULT 2,
    
    -- 代理配置
    proxy_type VARCHAR(20) DEFAULT '',
    proxy_url VARCHAR(255) DEFAULT '',
    
    -- 反连平台配置
    reverse_dnslog_domain VARCHAR(255) DEFAULT '',
    reverse_dnslog_api VARCHAR(255) DEFAULT '',
    reverse_rmi_server VARCHAR(255) DEFAULT '',
    reverse_ldap_server VARCHAR(255) DEFAULT '',
    reverse_http_server VARCHAR(255) DEFAULT '',
    
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_system_settings_created_at ON system_settings(created_at);

COMMENT ON TABLE system_settings IS '系统设置表';
COMMENT ON COLUMN system_settings.id IS '设置ID';
COMMENT ON COLUMN system_settings.network_card IS '网卡接口名称';
COMMENT ON COLUMN system_settings.concurrency IS '并发数';
COMMENT ON COLUMN system_settings.timeout IS '超时时间（秒）';
COMMENT ON COLUMN system_settings.retries IS '重试次数';
COMMENT ON COLUMN system_settings.proxy_type IS '代理类型: http/socks5';
COMMENT ON COLUMN system_settings.proxy_url IS '代理地址';
COMMENT ON COLUMN system_settings.reverse_dnslog_domain IS 'DNSLog域名';
COMMENT ON COLUMN system_settings.reverse_dnslog_api IS 'DNSLog API';
COMMENT ON COLUMN system_settings.reverse_rmi_server IS 'RMI服务器';
COMMENT ON COLUMN system_settings.reverse_ldap_server IS 'LDAP服务器';
COMMENT ON COLUMN system_settings.reverse_http_server IS 'HTTP服务器';
COMMENT ON COLUMN system_settings.created_at IS '创建时间';
COMMENT ON COLUMN system_settings.updated_at IS '更新时间';

-- 创建触发器来更新 updated_at 字段
CREATE TRIGGER update_system_settings_updated_at BEFORE UPDATE
    ON system_settings FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- 插入默认系统设置
INSERT INTO system_settings (
    network_card, concurrency, timeout, retries,
    proxy_type, proxy_url,
    reverse_dnslog_domain, reverse_dnslog_api,
    reverse_rmi_server, reverse_ldap_server, reverse_http_server
) VALUES (
    '', 10, 10, 2,
    '', '',
    '', '',
    '', '', ''
)
ON CONFLICT (id) DO UPDATE SET
    network_card = EXCLUDED.network_card,
    concurrency = EXCLUDED.concurrency,
    timeout = EXCLUDED.timeout,
    retries = EXCLUDED.retries,
    proxy_type = EXCLUDED.proxy_type,
    proxy_url = EXCLUDED.proxy_url,
    reverse_dnslog_domain = EXCLUDED.reverse_dnslog_domain,
    reverse_dnslog_api = EXCLUDED.reverse_dnslog_api,
    reverse_rmi_server = EXCLUDED.reverse_rmi_server,
    reverse_ldap_server = EXCLUDED.reverse_ldap_server,
    reverse_http_server = EXCLUDED.reverse_http_server;

-- ============================================
-- 2.1 Tools config table (tools)
-- Description: API keys for search engines
-- ============================================
CREATE TABLE IF NOT EXISTS tools (
    id SERIAL PRIMARY KEY,
    name VARCHAR(50) NOT NULL UNIQUE,
    api_key VARCHAR(255) DEFAULT '',
    api_email VARCHAR(255) DEFAULT '',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_tools_name ON tools(name);

COMMENT ON TABLE tools IS 'Tools API config';
COMMENT ON COLUMN tools.id IS 'ID';
COMMENT ON COLUMN tools.name IS 'Engine name';
COMMENT ON COLUMN tools.api_key IS 'API Key';
COMMENT ON COLUMN tools.api_email IS 'API Email';
COMMENT ON COLUMN tools.created_at IS 'Created at';
COMMENT ON COLUMN tools.updated_at IS 'Updated at';

-- 创建触发器来更新 updated_at 字段
CREATE TRIGGER update_tools_updated_at BEFORE UPDATE
    ON tools FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

INSERT INTO tools (name, api_key, api_email) VALUES
    ('hunter', '', ''),
    ('fofa', '', ''),
    ('quake', '', '')
ON CONFLICT (name) DO UPDATE SET
    api_key = EXCLUDED.api_key,
    api_email = EXCLUDED.api_email;

-- 插入默认管理员用户
-- 用户名: admin
-- 密码: Qaz@123# (已使用 bcrypt 加密)
INSERT INTO "user" (
    username,
    password,
    role,
    avatar,
    created_at
) VALUES (
    'admin',
    '$2a$10$KwjLTl6X0Jnq/q2CyI6d0.9ucFz3BxNxcJI.wC55LS3b5VH13RPp2',
    'admin',
    'http://localhost:3000/uploads/avatars/admin_1768566268.jpg',
    '2026-01-13 11:35:59'
)
ON CONFLICT (username) DO NOTHING;

-- ============================================
-- 3. 任务表 (task)
-- 说明: 存储扫描任务信息和执行状态
-- ============================================
CREATE TABLE IF NOT EXISTS task (
    id SERIAL PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    target VARCHAR(255) NOT NULL,
    type VARCHAR(20) NOT NULL,
    status VARCHAR(20) DEFAULT 'pending',
    progress INT DEFAULT 0,
    creator VARCHAR(50) NOT NULL,
    
    -- 任务配置与结果
    description TEXT,
    options TEXT,
    
    -- 时间戳
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    started_at TIMESTAMP,
    completed_at TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_task_creator ON task(creator);
CREATE INDEX IF NOT EXISTS idx_task_status ON task(status);
CREATE INDEX IF NOT EXISTS idx_task_created_at ON task(created_at);

COMMENT ON TABLE task IS '扫描任务表';
COMMENT ON COLUMN task.id IS '任务ID';
COMMENT ON COLUMN task.name IS '任务名称';
COMMENT ON COLUMN task.target IS '扫描目标（IP/域名/URL/CIDR）';
COMMENT ON COLUMN task.type IS '扫描类型: alive/port/web/vuln';
COMMENT ON COLUMN task.status IS '任务状态: pending/running/completed/failed/stopped';
COMMENT ON COLUMN task.progress IS '进度百分比（0-100）';
COMMENT ON COLUMN task.creator IS '创建者用户名';
COMMENT ON COLUMN task.description IS '任务描述';
COMMENT ON COLUMN task.options IS '扫描选项（JSON格式）';
COMMENT ON COLUMN task.created_at IS '创建时间';
COMMENT ON COLUMN task.updated_at IS '更新时间';
COMMENT ON COLUMN task.started_at IS '开始时间';
COMMENT ON COLUMN task.completed_at IS '完成时间';

-- 创建触发器来更新 updated_at 字段
CREATE TRIGGER update_task_updated_at BEFORE UPDATE
    ON task FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- ============================================
-- 4. 漏洞表 (vulnerability)
-- 说明: 存储扫描发现的漏洞详细信息
-- ============================================
CREATE TABLE IF NOT EXISTS vulnerability (
    id SERIAL PRIMARY KEY,
    task_id INT NOT NULL,
    
    -- 目标信息
    target VARCHAR(500) NOT NULL,
    ip VARCHAR(45),
    port INT,
    service VARCHAR(50),
    
    -- 漏洞基本信息
    name VARCHAR(255) NOT NULL,
    description TEXT,
    severity VARCHAR(20) NOT NULL,
    type VARCHAR(50),
    
    -- 漏洞标识
    cve VARCHAR(255),
    cwe VARCHAR(255),
    cvss DECIMAL(3,1),
    
    -- 模板信息
    template_id VARCHAR(255),
    template_path VARCHAR(500),
    author VARCHAR(255),
    tags TEXT,
    reference TEXT,
    
    -- 证据信息
    evidence_request TEXT,
    evidence_response TEXT,
    matched_at VARCHAR(500),
    extracted_data TEXT,
    curl_command TEXT,
    
    -- 附加信息
    metadata JSONB,
    discovered_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    
    -- 外键约束
    CONSTRAINT fk_vulnerability_task FOREIGN KEY (task_id) REFERENCES task(id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_vulnerability_task_id ON vulnerability(task_id);
CREATE INDEX IF NOT EXISTS idx_vulnerability_severity ON vulnerability(severity);
CREATE INDEX IF NOT EXISTS idx_vulnerability_target ON vulnerability(target);
CREATE INDEX IF NOT EXISTS idx_vulnerability_cve ON vulnerability(cve);

COMMENT ON TABLE vulnerability IS '漏洞详情表';
COMMENT ON COLUMN vulnerability.id IS '漏洞ID';
COMMENT ON COLUMN vulnerability.task_id IS '关联任务ID';
COMMENT ON COLUMN vulnerability.target IS '漏洞目标URL';
COMMENT ON COLUMN vulnerability.ip IS '目标IP地址';
COMMENT ON COLUMN vulnerability.port IS '目标端口';
COMMENT ON COLUMN vulnerability.service IS '服务类型';
COMMENT ON COLUMN vulnerability.name IS '漏洞名称';
COMMENT ON COLUMN vulnerability.description IS '漏洞描述';
COMMENT ON COLUMN vulnerability.severity IS '严重程度: critical/high/medium/low/info';
COMMENT ON COLUMN vulnerability.type IS '漏洞类型';
COMMENT ON COLUMN vulnerability.cve IS 'CVE编号';
COMMENT ON COLUMN vulnerability.cwe IS 'CWE编号';
COMMENT ON COLUMN vulnerability.cvss IS 'CVSS评分';
COMMENT ON COLUMN vulnerability.template_id IS '检测模板ID';
COMMENT ON COLUMN vulnerability.template_path IS '模板路径';
COMMENT ON COLUMN vulnerability.author IS '模板作者';
COMMENT ON COLUMN vulnerability.tags IS '标签（JSON数组）';
COMMENT ON COLUMN vulnerability.reference IS '参考链接（JSON数组）';
COMMENT ON COLUMN vulnerability.evidence_request IS '请求内容';
COMMENT ON COLUMN vulnerability.evidence_response IS '响应内容';
COMMENT ON COLUMN vulnerability.matched_at IS '匹配位置';
COMMENT ON COLUMN vulnerability.extracted_data IS '提取的数据（JSON）';
COMMENT ON COLUMN vulnerability.curl_command IS 'CURL复现命令';
COMMENT ON COLUMN vulnerability.metadata IS '其他元数据';
COMMENT ON COLUMN vulnerability.discovered_at IS '发现时间';

-- ============================================
-- 5. 资产表 (asset)
-- 说明: 存储扫描发现资产信息
-- ============================================
CREATE TABLE IF NOT EXISTS asset (
    id BIGSERIAL PRIMARY KEY,
    value VARCHAR(255) NOT NULL,
    asset_type VARCHAR(20) NOT NULL,
    is_alive BOOLEAN DEFAULT FALSE,
    first_seen TIMESTAMP NOT NULL,
    last_seen TIMESTAMP NOT NULL,
    
    -- 唯一约束
    CONSTRAINT uk_value UNIQUE (value)
);

COMMENT ON TABLE asset IS '资产表';
COMMENT ON COLUMN asset.id IS '资产ID';
COMMENT ON COLUMN asset.value IS 'IP 或域名';
COMMENT ON COLUMN asset.asset_type IS 'ip / domain';
COMMENT ON COLUMN asset.is_alive IS '是否存活';
COMMENT ON COLUMN asset.first_seen IS '首次发现时间';
COMMENT ON COLUMN asset.last_seen IS '最后发现时间';

-- ============================================
-- 5. 资产扫描结果表 (asset_scan_result)
-- 说明: 存储扫描资产结果信息
-- ============================================
CREATE TABLE IF NOT EXISTS asset_scan_result (
    id BIGSERIAL PRIMARY KEY,
    task_id INT NOT NULL,
    asset_id BIGINT NOT NULL,
    scan_type VARCHAR(20) NOT NULL,
    status VARCHAR(20) NOT NULL,
    result JSONB,
    scanned_at TIMESTAMP NOT NULL,
    
    -- 外键约束
    CONSTRAINT fk_asset_scan_result_task FOREIGN KEY (task_id) REFERENCES task(id) ON DELETE CASCADE,
    CONSTRAINT fk_asset_scan_result_asset FOREIGN KEY (asset_id) REFERENCES asset(id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_asset_scan_result_task ON asset_scan_result(task_id);
CREATE INDEX IF NOT EXISTS idx_asset_scan_result_asset ON asset_scan_result(asset_id);
CREATE INDEX IF NOT EXISTS idx_asset_scan_result_task_asset ON asset_scan_result(task_id, asset_id);

COMMENT ON TABLE asset_scan_result IS '资产扫描结果表';
COMMENT ON COLUMN asset_scan_result.id IS '扫描结果ID';
COMMENT ON COLUMN asset_scan_result.task_id IS '任务ID';
COMMENT ON COLUMN asset_scan_result.asset_id IS '资产ID';
COMMENT ON COLUMN asset_scan_result.scan_type IS '扫描类型';
COMMENT ON COLUMN asset_scan_result.status IS '扫描状态';
COMMENT ON COLUMN asset_scan_result.result IS '扫描结果（JSON格式）';
COMMENT ON COLUMN asset_scan_result.scanned_at IS '扫描时间';

-- ============================================
-- 6. 端口资产表 (asset_port)
-- 说明: 存储端口扫描发现的开放端口及其服务指纹信息
-- ============================================
CREATE TABLE IF NOT EXISTS asset_port (
    id BIGSERIAL PRIMARY KEY,
    task_id INT NOT NULL,
    asset_id BIGINT,
    
    -- 目标信息
    ip VARCHAR(45) NOT NULL,
    port INT NOT NULL,
    protocol VARCHAR(10) DEFAULT 'tcp',
    state VARCHAR(20) DEFAULT 'open',
    
    -- 服务信息
    service_name VARCHAR(100),
    service_product VARCHAR(255),
    service_version VARCHAR(100),
    service_extra_info TEXT,
    service_hostname VARCHAR(255),
    service_os_type VARCHAR(100),
    service_device_type VARCHAR(100),
    service_confidence INT DEFAULT 0,
    
    -- 指纹信息
    banner TEXT,
    fingerprint_method VARCHAR(50),
    raw_response TEXT,
    
    -- CPE信息
    cpes JSONB,
    
    -- 脚本扫描结果
    scripts JSONB,
    
    -- 时间戳
    discovered_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    last_seen TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    
    -- 唯一约束
    CONSTRAINT uk_ip_port_task UNIQUE (ip, port, task_id),
    
    -- 外键约束
    CONSTRAINT fk_asset_port_task FOREIGN KEY (task_id) REFERENCES task(id) ON DELETE CASCADE,
    CONSTRAINT fk_asset_port_asset FOREIGN KEY (asset_id) REFERENCES asset(id) ON DELETE SET NULL
);

CREATE INDEX IF NOT EXISTS idx_asset_port_task_id ON asset_port(task_id);
CREATE INDEX IF NOT EXISTS idx_asset_port_asset_id ON asset_port(asset_id);
CREATE INDEX IF NOT EXISTS idx_asset_port_ip ON asset_port(ip);
CREATE INDEX IF NOT EXISTS idx_asset_port_port ON asset_port(port);
CREATE INDEX IF NOT EXISTS idx_asset_port_ip_port ON asset_port(ip, port);
CREATE INDEX IF NOT EXISTS idx_asset_port_service ON asset_port(service_name);
CREATE INDEX IF NOT EXISTS idx_asset_port_discovered_at ON asset_port(discovered_at);

COMMENT ON TABLE asset_port IS '端口资产表';
COMMENT ON COLUMN asset_port.id IS '端口资产ID';
COMMENT ON COLUMN asset_port.task_id IS '关联任务ID';
COMMENT ON COLUMN asset_port.asset_id IS '关联资产ID（可为空）';
COMMENT ON COLUMN asset_port.ip IS 'IP地址';
COMMENT ON COLUMN asset_port.port IS '端口号';
COMMENT ON COLUMN asset_port.protocol IS '协议类型：tcp/udp';
COMMENT ON COLUMN asset_port.state IS '端口状态：open/closed/filtered';
COMMENT ON COLUMN asset_port.service_name IS '服务名称';
COMMENT ON COLUMN asset_port.service_product IS '产品名称';
COMMENT ON COLUMN asset_port.service_version IS '服务版本';
COMMENT ON COLUMN asset_port.service_extra_info IS '额外信息';
COMMENT ON COLUMN asset_port.service_hostname IS '服务主机名';
COMMENT ON COLUMN asset_port.service_os_type IS '操作系统类型';
COMMENT ON COLUMN asset_port.service_device_type IS '设备类型';
COMMENT ON COLUMN asset_port.service_confidence IS '服务识别置信度（0-100）';
COMMENT ON COLUMN asset_port.banner IS 'Banner信息';
COMMENT ON COLUMN asset_port.fingerprint_method IS '识别方法：nmap-probes/banner/port-guess';
COMMENT ON COLUMN asset_port.raw_response IS '原始响应数据';
COMMENT ON COLUMN asset_port.cpes IS 'CPE列表（JSON数组）';
COMMENT ON COLUMN asset_port.scripts IS 'NSE脚本输出（JSON对象）';
COMMENT ON COLUMN asset_port.discovered_at IS '发现时间';
COMMENT ON COLUMN asset_port.last_seen IS '最后发现时间';

-- 创建触发器来更新 last_seen 字段
CREATE TRIGGER update_asset_port_last_seen BEFORE UPDATE
    ON asset_port FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- ============================================
-- 7. 仪表盘统计数据表 (dashboard)
-- 说明: 存储仪表盘各项统计数据，用于提升仪表盘加载性能
-- 注意: 此表设计为单行表，始终只保留一条记录
-- ============================================
CREATE TABLE IF NOT EXISTS dashboard (
    id SERIAL PRIMARY KEY,
    
    -- 核心统计数据
    total_assets INT DEFAULT 0,
    total_vulnerabilities INT DEFAULT 0,
    total_tasks INT DEFAULT 0,
    total_fingerprints INT DEFAULT 0,
    
    -- 漏洞严重程度统计
    critical_vulns INT DEFAULT 0,
    high_vulns INT DEFAULT 0,
    medium_vulns INT DEFAULT 0,
    low_vulns INT DEFAULT 0,
    info_vulns INT DEFAULT 0,
    
    -- 时间戳
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

COMMENT ON TABLE dashboard IS '仪表盘统计数据表';
COMMENT ON COLUMN dashboard.id IS '统计ID（始终为1）';
COMMENT ON COLUMN dashboard.total_assets IS '资产总数';
COMMENT ON COLUMN dashboard.total_vulnerabilities IS '漏洞总数';
COMMENT ON COLUMN dashboard.total_tasks IS '任务总数';
COMMENT ON COLUMN dashboard.total_fingerprints IS '指纹总数（已识别的服务）';
COMMENT ON COLUMN dashboard.critical_vulns IS '严重漏洞数';
COMMENT ON COLUMN dashboard.high_vulns IS '高危漏洞数';
COMMENT ON COLUMN dashboard.medium_vulns IS '中危漏洞数';
COMMENT ON COLUMN dashboard.low_vulns IS '低危漏洞数';
COMMENT ON COLUMN dashboard.info_vulns IS '信息级漏洞数';
COMMENT ON COLUMN dashboard.updated_at IS '最后更新时间';

-- 创建触发器来更新 updated_at 字段
CREATE TRIGGER update_dashboard_updated_at BEFORE UPDATE
    ON dashboard FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- 插入默认统计记录（初始值全为0）
INSERT INTO dashboard (
    id,
    total_assets, total_vulnerabilities, total_tasks, total_fingerprints,
    critical_vulns, high_vulns, medium_vulns, low_vulns, info_vulns
) VALUES (
    1,
    0, 0, 0, 0,
    0, 0, 0, 0, 0
)
ON CONFLICT (id) DO UPDATE SET
    total_assets = EXCLUDED.total_assets,
    total_vulnerabilities = EXCLUDED.total_vulnerabilities,
    total_tasks = EXCLUDED.total_tasks,
    total_fingerprints = EXCLUDED.total_fingerprints,
    critical_vulns = EXCLUDED.critical_vulns,
    high_vulns = EXCLUDED.high_vulns,
    medium_vulns = EXCLUDED.medium_vulns,
    low_vulns = EXCLUDED.low_vulns,
    info_vulns = EXCLUDED.info_vulns;

-- ============================================
-- 8. Web指纹资产表 (asset_web_fingerprints)
-- 说明: 存储Web指纹识别结果，使用wappalyzergo识别的技术栈信息
-- 注意: 此表可以从端口扫描结果自动触发，当发现HTTP/HTTPS服务时进行识别
-- ============================================
CREATE TABLE IF NOT EXISTS asset_web_fingerprints (
    id BIGSERIAL PRIMARY KEY,
    task_id INT NOT NULL,
    asset_id BIGINT,
    port_id BIGINT,
    
    -- 目标信息
    url VARCHAR(500) NOT NULL,
    ip VARCHAR(45) NOT NULL,
    port INT NOT NULL,
    protocol VARCHAR(10) DEFAULT 'http',
    
    -- 响应信息
    title VARCHAR(500),
    status_code INT,
    server VARCHAR(255),
    content_type VARCHAR(255),
    content_length BIGINT,
    response_time INT,
    
    -- 指纹信息
    technologies JSONB,
    frameworks JSONB,
    matched_rules JSONB,
    favicon_hash VARCHAR(100),
    
    -- 详细信息
    headers JSONB,
    meta_tags JSONB,
    cookies JSONB,
    
    -- 附加信息
    screenshot_path VARCHAR(500),
    raw_html_hash VARCHAR(64),
    
    -- 时间戳
    discovered_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    last_checked TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    
    -- 唯一约束
    CONSTRAINT uk_task_url UNIQUE (task_id, url),
    
    -- 外键约束
    CONSTRAINT fk_asset_web_fingerprints_task FOREIGN KEY (task_id) REFERENCES task(id) ON DELETE CASCADE,
    CONSTRAINT fk_asset_web_fingerprints_asset FOREIGN KEY (asset_id) REFERENCES asset(id) ON DELETE SET NULL,
    CONSTRAINT fk_asset_web_fingerprints_port FOREIGN KEY (port_id) REFERENCES asset_port(id) ON DELETE SET NULL
);

CREATE INDEX IF NOT EXISTS idx_asset_web_fingerprints_task_id ON asset_web_fingerprints(task_id);
CREATE INDEX IF NOT EXISTS idx_asset_web_fingerprints_asset_id ON asset_web_fingerprints(asset_id);
CREATE INDEX IF NOT EXISTS idx_asset_web_fingerprints_port_id ON asset_web_fingerprints(port_id);
CREATE INDEX IF NOT EXISTS idx_asset_web_fingerprints_url ON asset_web_fingerprints(url);
CREATE INDEX IF NOT EXISTS idx_asset_web_fingerprints_ip_port ON asset_web_fingerprints(ip, port);
CREATE INDEX IF NOT EXISTS idx_asset_web_fingerprints_status_code ON asset_web_fingerprints(status_code);
CREATE INDEX IF NOT EXISTS idx_asset_web_fingerprints_discovered_at ON asset_web_fingerprints(discovered_at);

COMMENT ON TABLE asset_web_fingerprints IS 'Web指纹资产表';
COMMENT ON COLUMN asset_web_fingerprints.id IS 'Web指纹ID';
COMMENT ON COLUMN asset_web_fingerprints.task_id IS '关联任务ID';
COMMENT ON COLUMN asset_web_fingerprints.asset_id IS '关联资产ID';
COMMENT ON COLUMN asset_web_fingerprints.port_id IS '关联端口ID（来自asset_port表）';
COMMENT ON COLUMN asset_web_fingerprints.url IS '完整URL';
COMMENT ON COLUMN asset_web_fingerprints.ip IS 'IP地址';
COMMENT ON COLUMN asset_web_fingerprints.port IS '端口号';
COMMENT ON COLUMN asset_web_fingerprints.protocol IS '协议类型：http/https';
COMMENT ON COLUMN asset_web_fingerprints.title IS '网页标题';
COMMENT ON COLUMN asset_web_fingerprints.status_code IS 'HTTP状态码';
COMMENT ON COLUMN asset_web_fingerprints.server IS 'Server响应头';
COMMENT ON COLUMN asset_web_fingerprints.content_type IS 'Content-Type';
COMMENT ON COLUMN asset_web_fingerprints.content_length IS '响应体大小（字节）';
COMMENT ON COLUMN asset_web_fingerprints.response_time IS '响应时间（毫秒）';
COMMENT ON COLUMN asset_web_fingerprints.technologies IS '识别到的技术栈列表 (Wappalyzer)，格式：["Nginx","PHP","WordPress"]';
COMMENT ON COLUMN asset_web_fingerprints.frameworks IS '识别到的应用框架列表 (GoAttack)，格式：["DVWA","RuoYi"]';
COMMENT ON COLUMN asset_web_fingerprints.matched_rules IS '匹配到的具体指纹规则信息';
COMMENT ON COLUMN asset_web_fingerprints.favicon_hash IS 'Favicon哈希值';
COMMENT ON COLUMN asset_web_fingerprints.headers IS 'HTTP响应头（JSON对象）';
COMMENT ON COLUMN asset_web_fingerprints.meta_tags IS 'Meta标签信息';
COMMENT ON COLUMN asset_web_fingerprints.cookies IS 'Set-Cookie信息';
COMMENT ON COLUMN asset_web_fingerprints.screenshot_path IS '截图路径（可选）';
COMMENT ON COLUMN asset_web_fingerprints.raw_html_hash IS 'HTML内容哈希（用于去重）';
COMMENT ON COLUMN asset_web_fingerprints.discovered_at IS '发现时间';
COMMENT ON COLUMN asset_web_fingerprints.last_checked IS '最后检查时间';

-- 创建触发器来更新 last_checked 字段
-- 创建触发器来更新 last_checked 字段
CREATE OR REPLACE FUNCTION update_last_checked_column()
RETURNS TRIGGER AS $
BEGIN
    NEW.last_checked = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$ language 'plpgsql';

CREATE TRIGGER update_asset_web_fingerprints_last_checked BEFORE UPDATE
    ON asset_web_fingerprints FOR EACH ROW EXECUTE FUNCTION update_last_checked_column();

-- ============================================
-- 9. POC模板表 (poc_template)
-- 说明: 存储从nuclei-templates扫描出来的POC模板信息
-- 注意: 此表用于POC管理功能，与漏洞扫描功能关联
-- ============================================
CREATE TABLE IF NOT EXISTS poc_template (
    id BIGSERIAL PRIMARY KEY,
    
    -- 模板基本信息
    template_id VARCHAR(255) NOT NULL,
    name VARCHAR(500) NOT NULL,
    description TEXT,
    author VARCHAR(255),
    
    -- 分类信息
    category VARCHAR(100),
    severity VARCHAR(20) NOT NULL,
    tags JSONB,
    
    -- CVE/CWE/CNVD信息
    cve_id VARCHAR(50),
    cnvd_id VARCHAR(50),
    cwe_id VARCHAR(50),
    cvss_score DECIMAL(3,1),
    cvss_metrics VARCHAR(200),
    
    -- 模板元数据
    protocol VARCHAR(50) DEFAULT 'http',
    max_request INT DEFAULT 1,
    reference JSONB,
    classification JSONB,
    metadata JSONB,
    
    -- 文件信息
    file_path VARCHAR(500) NOT NULL,
    file_hash VARCHAR(64),
    template_content TEXT,
    
    -- 状态信息
    is_active BOOLEAN DEFAULT TRUE,
    verified BOOLEAN DEFAULT FALSE,
    
    -- 时间戳
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    last_scanned_at TIMESTAMP,
    
    -- 唯一约束
    CONSTRAINT uk_template_id UNIQUE (template_id)
);

CREATE INDEX IF NOT EXISTS idx_poc_template_template_id ON poc_template(template_id);
CREATE INDEX IF NOT EXISTS idx_poc_template_category ON poc_template(category);
CREATE INDEX IF NOT EXISTS idx_poc_template_severity ON poc_template(severity);
CREATE INDEX IF NOT EXISTS idx_poc_template_cve_id ON poc_template(cve_id);
CREATE INDEX IF NOT EXISTS idx_poc_template_protocol ON poc_template(protocol);
CREATE INDEX IF NOT EXISTS idx_poc_template_is_active ON poc_template(is_active);
CREATE INDEX IF NOT EXISTS idx_poc_template_created_at ON poc_template(created_at);

COMMENT ON TABLE poc_template IS 'POC模板表';
COMMENT ON COLUMN poc_template.id IS 'POC模板ID';
COMMENT ON COLUMN poc_template.template_id IS '模板唯一标识（nuclei模板ID）';
COMMENT ON COLUMN poc_template.name IS 'POC名称';
COMMENT ON COLUMN poc_template.description IS 'POC描述';
COMMENT ON COLUMN poc_template.author IS '作者';
COMMENT ON COLUMN poc_template.category IS '分类：cves/vulnerabilities/exposures/misconfiguration等';
COMMENT ON COLUMN poc_template.severity IS '严重程度：critical/high/medium/low/info';
COMMENT ON COLUMN poc_template.tags IS '标签列表（JSON数组）';
COMMENT ON COLUMN poc_template.cve_id IS 'CVE编号（如有）';
COMMENT ON COLUMN poc_template.cnvd_id IS 'CNVD编号（如有）';
COMMENT ON COLUMN poc_template.cwe_id IS 'CWE编号（如有）';
COMMENT ON COLUMN poc_template.cvss_score IS 'CVSS评分';
COMMENT ON COLUMN poc_template.cvss_metrics IS 'CVSS向量';
COMMENT ON COLUMN poc_template.protocol IS '协议类型：http/network/dns/ssl等';
COMMENT ON COLUMN poc_template.max_request IS '最大请求数';
COMMENT ON COLUMN poc_template.reference IS '参考链接（JSON数组）';
COMMENT ON COLUMN poc_template.classification IS '分类信息（JSON对象）';
COMMENT ON COLUMN poc_template.metadata IS '其他元数据（JSON对象）';
COMMENT ON COLUMN poc_template.file_path IS '模板文件相对路径';
COMMENT ON COLUMN poc_template.file_hash IS '文件SHA256哈希值（用于检测变更）';
COMMENT ON COLUMN poc_template.template_content IS '模板YAML原始内容';
COMMENT ON COLUMN poc_template.is_active IS '是否启用该POC';
COMMENT ON COLUMN poc_template.verified IS '是否已验证';
COMMENT ON COLUMN poc_template.created_at IS '创建时间';
COMMENT ON COLUMN poc_template.updated_at IS '更新时间';
COMMENT ON COLUMN poc_template.last_scanned_at IS '最后扫描时间';

-- 创建触发器来更新 updated_at 字段
CREATE TRIGGER update_poc_template_updated_at BEFORE UPDATE
    ON poc_template FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- ============================================
-- 10. POC验证结果表 (poc_verify_result)
-- 说明: 存储POC验证的历史记录和详细结果
-- ============================================
CREATE TABLE IF NOT EXISTS poc_verify_result (
    id BIGSERIAL PRIMARY KEY,
    
    -- 验证基本信息
    target VARCHAR(500) NOT NULL,
    poc_id BIGINT NOT NULL,
    template_id VARCHAR(255) NOT NULL,
    template_name VARCHAR(500) NOT NULL,
    
    -- 验证结果
    matched BOOLEAN DEFAULT FALSE,
    severity VARCHAR(20),
    description TEXT,
    
    -- 请求和响应详情
    request TEXT,
    response TEXT,
    matched_at VARCHAR(500),
    
    -- 提取的数据
    extracted_data JSONB,
    
    -- 错误信息
    error TEXT,
    
    -- 执行信息
    verified_by VARCHAR(50),
    verified_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    
    -- 外键约束
    CONSTRAINT fk_poc_verify_result_poc FOREIGN KEY (poc_id) REFERENCES poc_template(id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_poc_verify_result_target ON poc_verify_result(target);
CREATE INDEX IF NOT EXISTS idx_poc_verify_result_poc_id ON poc_verify_result(poc_id);
CREATE INDEX IF NOT EXISTS idx_poc_verify_result_template_id ON poc_verify_result(template_id);
CREATE INDEX IF NOT EXISTS idx_poc_verify_result_matched ON poc_verify_result(matched);
CREATE INDEX IF NOT EXISTS idx_poc_verify_result_severity ON poc_verify_result(severity);
CREATE INDEX IF NOT EXISTS idx_poc_verify_result_verified_at ON poc_verify_result(verified_at);

COMMENT ON TABLE poc_verify_result IS 'POC验证结果表';
COMMENT ON COLUMN poc_verify_result.id IS '验证结果ID';
COMMENT ON COLUMN poc_verify_result.target IS '验证目标（IP/域名/URL）';
COMMENT ON COLUMN poc_verify_result.poc_id IS '关联的POC模板ID';
COMMENT ON COLUMN poc_verify_result.template_id IS '模板唯一标识（nuclei模板ID）';
COMMENT ON COLUMN poc_verify_result.template_name IS 'POC名称';
COMMENT ON COLUMN poc_verify_result.matched IS '是否匹配成功';
COMMENT ON COLUMN poc_verify_result.severity IS '严重程度：critical/high/medium/low/info';
COMMENT ON COLUMN poc_verify_result.description IS 'POC描述';
COMMENT ON COLUMN poc_verify_result.request IS '发送的请求包';
COMMENT ON COLUMN poc_verify_result.response IS '返回的响应包';
COMMENT ON COLUMN poc_verify_result.matched_at IS '匹配位置或URL';
COMMENT ON COLUMN poc_verify_result.extracted_data IS '提取的数据（JSON）';
COMMENT ON COLUMN poc_verify_result.error IS '错误信息（如果验证失败）';
COMMENT ON COLUMN poc_verify_result.verified_by IS '验证执行者用户名';
COMMENT ON COLUMN poc_verify_result.verified_at IS '验证时间';

-- ============================================
-- 11. 字典表 (dict)
-- 说明: 存储字典信息，支持默认字典与导入字典
-- ============================================
CREATE TABLE IF NOT EXISTS dict (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    type VARCHAR(20) DEFAULT 'preset',
    category VARCHAR(50) DEFAULT 'other',
    size BIGINT DEFAULT 0,
    lines_cnt BIGINT DEFAULT 0,
    path VARCHAR(500) NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    
    -- 唯一约束
    CONSTRAINT uk_name_type UNIQUE (name, type)
);

CREATE INDEX IF NOT EXISTS idx_dict_category ON dict(category);
CREATE INDEX IF NOT EXISTS idx_dict_path ON dict(path);
CREATE INDEX IF NOT EXISTS idx_dict_type ON dict(type);

COMMENT ON TABLE dict IS '字典表';
COMMENT ON COLUMN dict.id IS '字典ID';
COMMENT ON COLUMN dict.name IS '字典名称';
COMMENT ON COLUMN dict.type IS '类型：preset/custom';
COMMENT ON COLUMN dict.category IS '分类：password, directory, domain, fuzz等';
COMMENT ON COLUMN dict.size IS '文件大小(字节)';
COMMENT ON COLUMN dict.lines_cnt IS '文件行数';
COMMENT ON COLUMN dict.path IS '字典保存路径';
COMMENT ON COLUMN dict.created_at IS '创建时间';

-- ============================================
-- 12. 通知已读时间表 (notification_read_time)
-- 说明: 记录每个用户的漏洞通知已读时间戳
--       任何在该时间戳之后发现的漏洞视为"未读"
-- ============================================
CREATE TABLE IF NOT EXISTS notification_read_time (
    id SERIAL PRIMARY KEY,
    username VARCHAR(50) NOT NULL,
    last_read_at TIMESTAMP DEFAULT '2000-01-01 00:00:00',
    last_cleared_at TIMESTAMP DEFAULT '2000-01-01 00:00:00',
    
    -- 唯一约束
    CONSTRAINT uk_username UNIQUE (username)
);

CREATE INDEX IF NOT EXISTS idx_notification_read_time_username ON notification_read_time(username);

COMMENT ON TABLE notification_read_time IS '通知已读时间追踪表';
COMMENT ON COLUMN notification_read_time.id IS 'ID';
COMMENT ON COLUMN notification_read_time.username IS '用户名';
COMMENT ON COLUMN notification_read_time.last_read_at IS '上次已读时间';
COMMENT ON COLUMN notification_read_time.last_cleared_at IS '上次清空时间';

-- ============================================
-- 13. 插件表 (plugins)
-- 说明: 存储可执行工具插件的配置信息
-- ============================================
CREATE TABLE IF NOT EXISTS plugins (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) UNIQUE,
    version VARCHAR(50),
    type VARCHAR(50),
    enabled BOOLEAN DEFAULT TRUE,
    description TEXT,
    path VARCHAR(255),
    config TEXT,
    author VARCHAR(100) DEFAULT '',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_plugins_name ON plugins(name);

COMMENT ON TABLE plugins IS '插件表';
COMMENT ON COLUMN plugins.id IS '插件ID';
COMMENT ON COLUMN plugins.name IS '插件名称';
COMMENT ON COLUMN plugins.version IS '版本号';
COMMENT ON COLUMN plugins.type IS '类型';
COMMENT ON COLUMN plugins.enabled IS '是否启用';
COMMENT ON COLUMN plugins.description IS '描述';
COMMENT ON COLUMN plugins.path IS '可执行文件路径';
COMMENT ON COLUMN plugins.config IS '配置信息';
COMMENT ON COLUMN plugins.author IS '作者';
COMMENT ON COLUMN plugins.created_at IS '创建时间';
COMMENT ON COLUMN plugins.updated_at IS '更新时间';

-- 创建触发器来更新 updated_at 字段
CREATE TRIGGER update_plugins_updated_at BEFORE UPDATE
    ON plugins FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- ============================================
-- 修复插件表缺少author字段的问题
-- ============================================
ALTER TABLE plugins ADD COLUMN IF NOT EXISTS author VARCHAR(100) DEFAULT '';

EOF