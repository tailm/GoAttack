-- GoAttack 漏洞扫描系统数据库初始化脚本 (PostgreSQL简化版本)
-- 创建时间: 2026-01-19
-- 说明: 该脚本用于首次部署时初始化数据库结构
-- PostgreSQL版本，已从MySQL语法转换

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

CREATE INDEX IF NOT EXISTS idx_user_username ON "user"(username);
CREATE INDEX IF NOT EXISTS idx_user_role ON "user"(role);

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
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ language 'plpgsql';

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
    name VARCHAR(255) NOT NULL,
    description TEXT DEFAULT '',
    target TEXT NOT NULL,
    scan_type VARCHAR(50) NOT NULL,
    status VARCHAR(20) DEFAULT 'pending',
    progress INT DEFAULT 0,
    created_by VARCHAR(50) NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    started_at TIMESTAMP,
    finished_at TIMESTAMP,
    result_summary TEXT DEFAULT '',
    error_message TEXT DEFAULT ''
);

CREATE INDEX IF NOT EXISTS idx_task_status ON task(status);
CREATE INDEX IF NOT EXISTS idx_task_created_by ON task(created_by);
CREATE INDEX IF NOT EXISTS idx_task_created_at ON task(created_at);

COMMENT ON TABLE task IS '任务表';
COMMENT ON COLUMN task.id IS '任务ID';
COMMENT ON COLUMN task.name IS '任务名称';
COMMENT ON COLUMN task.description IS '任务描述';
COMMENT ON COLUMN task.target IS '扫描目标（IP/域名/URL，多个用逗号分隔）';
COMMENT ON COLUMN task.scan_type IS '扫描类型：full/quick/custom';
COMMENT ON COLUMN task.status IS '任务状态：pending/running/completed/failed';
COMMENT ON COLUMN task.progress IS '进度百分比（0-100）';
COMMENT ON COLUMN task.created_by IS '创建者用户名';
COMMENT ON COLUMN task.created_at IS '创建时间';
COMMENT ON COLUMN task.started_at IS '开始时间';
COMMENT ON COLUMN task.finished_at IS '完成时间';
COMMENT ON COLUMN task.result_summary IS '结果摘要';
COMMENT ON COLUMN task.error_message IS '错误信息';

-- ============================================
-- 4. 资产表 (asset)
-- 说明: 存储扫描发现的资产信息
-- ============================================
CREATE TABLE IF NOT EXISTS asset (
    id SERIAL PRIMARY KEY,
    task_id INT NOT NULL,
    ip VARCHAR(45) NOT NULL,
    hostname VARCHAR(255) DEFAULT '',
    os VARCHAR(100) DEFAULT '',
    mac_address VARCHAR(17) DEFAULT '',
    vendor VARCHAR(100) DEFAULT '',
    status VARCHAR(20) DEFAULT 'unknown',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (task_id) REFERENCES task(id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_asset_task_id ON asset(task_id);
CREATE INDEX IF NOT EXISTS idx_asset_ip ON asset(ip);
CREATE INDEX IF NOT EXISTS idx_asset_status ON asset(status);

COMMENT ON TABLE asset IS '资产表';
COMMENT ON COLUMN asset.id IS '资产ID';
COMMENT ON COLUMN asset.task_id IS '关联的任务ID';
COMMENT ON COLUMN asset.ip IS 'IP地址';
COMMENT ON COLUMN asset.hostname IS '主机名';
COMMENT ON COLUMN asset.os IS '操作系统';
COMMENT ON COLUMN asset.mac_address IS 'MAC地址';
COMMENT ON COLUMN asset.vendor IS '设备厂商';
COMMENT ON COLUMN asset.status IS '状态：alive/dead/unknown';
COMMENT ON COLUMN asset.created_at IS '创建时间';

-- ============================================
-- 5. 端口表 (asset_port)
-- 说明: 存储资产开放的端口信息
-- ============================================
CREATE TABLE IF NOT EXISTS asset_port (
    id SERIAL PRIMARY KEY,
    asset_id INT NOT NULL,
    port INT NOT NULL,
    protocol VARCHAR(10) DEFAULT 'tcp',
    service VARCHAR(100) DEFAULT '',
    version VARCHAR(100) DEFAULT '',
    product VARCHAR(100) DEFAULT '',
    extra_info TEXT DEFAULT '',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (asset_id) REFERENCES asset(id) ON DELETE CASCADE,
    UNIQUE(asset_id, port, protocol)
);

CREATE INDEX IF NOT EXISTS idx_asset_port_asset_id ON asset_port(asset_id);
CREATE INDEX IF NOT EXISTS idx_asset_port_port ON asset_port(port);
CREATE INDEX IF NOT EXISTS idx_asset_port_service ON asset_port(service);

COMMENT ON TABLE asset_port IS '端口表';
COMMENT ON COLUMN asset_port.id IS '端口记录ID';
COMMENT ON COLUMN asset_port.asset_id IS '关联的资产ID';
COMMENT ON COLUMN asset_port.port IS '端口号';
COMMENT ON COLUMN asset_port.protocol IS '协议：tcp/udp';
COMMENT ON COLUMN asset_port.service IS '服务名称';
COMMENT ON COLUMN asset_port.version IS '服务版本';
COMMENT ON COLUMN asset_port.product IS '产品名称';
COMMENT ON COLUMN asset_port.extra_info IS '额外信息';
COMMENT ON COLUMN asset_port.created_at IS '创建时间';

-- ============================================
-- 6. Web指纹表 (web_fingerprint)
-- 说明: 存储Web应用的指纹信息
-- ============================================
CREATE TABLE IF NOT EXISTS web_fingerprint (
    id SERIAL PRIMARY KEY,
    asset_port_id INT NOT NULL,
    url TEXT NOT NULL,
    title VARCHAR(500) DEFAULT '',
    status_code INT DEFAULT 0,
    content_type VARCHAR(100) DEFAULT '',
    server VARCHAR(100) DEFAULT '',
    technologies TEXT DEFAULT '',
    headers TEXT DEFAULT '',
    body_hash VARCHAR(64) DEFAULT '',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (asset_port_id) REFERENCES asset_port(id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_web_fingerprint_asset_port_id ON web_fingerprint(asset_port_id);
CREATE INDEX IF NOT EXISTS idx_web_fingerprint_url ON web_fingerprint(url);
CREATE INDEX IF NOT EXISTS idx_web_fingerprint_title ON web_fingerprint(title);

COMMENT ON TABLE web_fingerprint IS 'Web指纹表';
COMMENT ON COLUMN web_fingerprint.id IS '指纹ID';
COMMENT ON COLUMN web_fingerprint.asset_port_id IS '关联的端口ID';
COMMENT ON COLUMN web_fingerprint.url IS 'URL地址';
COMMENT ON COLUMN web_fingerprint.title IS '页面标题';
COMMENT ON COLUMN web_fingerprint.status_code IS 'HTTP状态码';
COMMENT ON COLUMN web_fingerprint.content_type IS '内容类型';
COMMENT ON COLUMN web_fingerprint.server IS '服务器信息';
COMMENT ON COLUMN web_fingerprint.technologies IS '技术栈（JSON格式）';
COMMENT ON COLUMN web_fingerprint.headers IS 'HTTP头信息（JSON格式）';
COMMENT ON COLUMN web_fingerprint.body_hash IS '页面内容哈希';
COMMENT ON COLUMN web_fingerprint.created_at IS '创建时间';

-- ============================================
-- 7. 漏洞表 (vulnerability)
-- 说明: 存储发现的漏洞信息
-- ============================================
CREATE TABLE IF NOT EXISTS vulnerability (
    id SERIAL PRIMARY KEY,
    task_id INT NOT NULL,
    asset_id INT NOT NULL,
    asset_port_id INT,
    name VARCHAR(255) NOT NULL,
    severity VARCHAR(20) DEFAULT 'medium',
    description TEXT DEFAULT '',
    poc_content TEXT DEFAULT '',
    request TEXT DEFAULT '',
    response TEXT DEFAULT '',
    extra_info TEXT DEFAULT '',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (task_id) REFERENCES task(id) ON DELETE CASCADE,
    FOREIGN KEY (asset_id) REFERENCES asset(id) ON DELETE CASCADE,
    FOREIGN KEY (asset_port_id) REFERENCES asset_port(id) ON DELETE SET NULL
);

CREATE INDEX IF NOT EXISTS idx_vulnerability_task_id ON vulnerability(task_id);
CREATE INDEX IF NOT EXISTS idx_vulnerability_asset_id ON vulnerability(asset_id);
CREATE INDEX IF NOT EXISTS idx_vulnerability_severity ON vulnerability(severity);
CREATE INDEX IF NOT EXISTS idx_vulnerability_created_at ON vulnerability(created_at);

COMMENT ON TABLE vulnerability IS '漏洞表';
COMMENT ON COLUMN vulnerability.id IS '漏洞ID';
COMMENT ON COLUMN vulnerability.task_id IS '关联的任务ID';
COMMENT ON COLUMN vulnerability.asset_id IS '关联的资产ID';
COMMENT ON COLUMN vulnerability.asset_port_id IS '关联的端口ID（可为空）';
COMMENT ON COLUMN vulnerability.name IS '漏洞名称';
COMMENT ON COLUMN vulnerability.severity IS '严重程度：critical/high/medium/low/info';
COMMENT ON COLUMN vulnerability.description IS '漏洞描述';
COMMENT ON COLUMN vulnerability.poc_content IS 'POC内容';
COMMENT ON COLUMN vulnerability.request IS '请求内容';
COMMENT ON COLUMN vulnerability.response IS '响应内容';
COMMENT ON COLUMN vulnerability.extra_info IS '额外信息';
COMMENT ON COLUMN vulnerability.created_at IS '创建时间';

-- ============================================
-- 8. POC模板表 (poc_template)
-- 说明: 存储POC模板信息
-- ============================================
CREATE TABLE IF NOT EXISTS poc_template (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    author VARCHAR(100) DEFAULT '',
    severity VARCHAR(20) DEFAULT 'medium',
    description TEXT DEFAULT '',
    reference TEXT DEFAULT '',
    tags TEXT DEFAULT '',
    raw_content TEXT NOT NULL,
    enabled BOOLEAN DEFAULT true,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_poc_template_name ON poc_template(name);
CREATE INDEX IF NOT EXISTS idx_poc_template_severity ON poc_template(severity);
CREATE INDEX IF NOT EXISTS idx_poc_template_enabled ON poc_template(enabled);

COMMENT ON TABLE poc_template IS 'POC模板表';
COMMENT ON COLUMN poc_template.id IS '模板ID';
COMMENT ON COLUMN poc_template.name IS '模板名称';
COMMENT ON COLUMN poc_template.author IS '作者';
COMMENT ON COLUMN poc_template.severity IS '严重程度';
COMMENT ON COLUMN poc_template.description IS '描述';
COMMENT ON COLUMN poc_template.reference IS '参考链接';
COMMENT ON COLUMN poc_template.tags IS '标签（JSON格式）';
COMMENT ON COLUMN poc_template.raw_content IS '原始YAML内容';
COMMENT ON COLUMN poc_template.enabled IS '是否启用';
COMMENT ON COLUMN poc_template.created_at IS '创建时间';
COMMENT ON COLUMN poc_template.updated_at IS '更新时间';

-- 创建触发器来更新 updated_at 字段
CREATE TRIGGER update_poc_template_updated_at BEFORE UPDATE
    ON poc_template FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- ============================================
-- 9. 字典表 (dictionary)
-- 说明: 存储各种字典数据（密码、目录、子域名等）
-- ============================================
CREATE TABLE IF NOT EXISTS dictionary (
    id SERIAL PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    type VARCHAR(20) NOT NULL,
    content TEXT NOT NULL,
    description TEXT DEFAULT '',
    enabled BOOLEAN DEFAULT true,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(name, type)
);

CREATE INDEX IF NOT EXISTS idx_dictionary_type ON dictionary(type);
CREATE INDEX IF NOT EXISTS idx_dictionary_enabled ON dictionary(enabled);

COMMENT ON TABLE dictionary IS '字典表';
COMMENT ON COLUMN dictionary.id IS '字典ID';
COMMENT ON COLUMN dictionary.name IS '字典名称';
COMMENT ON COLUMN dictionary.type IS '字典类型：password/directory/domain等';
COMMENT ON COLUMN dictionary.content IS '字典内容（每行一个条目）';
COMMENT ON COLUMN dictionary.description IS '描述';
COMMENT ON COLUMN dictionary.enabled IS '是否启用';
COMMENT ON COLUMN dictionary.created_at IS '创建时间';
COMMENT ON COLUMN dictionary.updated_at IS '更新时间';

-- 创建触发器来更新 updated_at 字段
CREATE TRIGGER update_dictionary_updated_at BEFORE UPDATE
    ON dictionary FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- 插入简化的字典数据（避免单引号转义问题）
INSERT INTO dictionary (name, type, content, description) VALUES
    ('top1000-passwords', 'password', '123456
password
12345678
qwerty
123456789
12345
1234
111111
1234567
dragon
123123
baseball
abc123
football
monkey
letmein
shadow
master
666666
qwertyuiop
123321
mustang
1234567890
michael
654321
superman
1qaz2wsx
7777777
121212
000000
qazwsx
123qwe
killer
trustno1
jordan
jennifer
zxcvbnm
asdfgh
hunter
buster
soccer
harley
batman
andrew
tigger
sunshine
iloveyou
2000
charlie
robert
thomas
hockey
ranger
daniel
starwars
klaster
112233
george
computer
michelle
jessica
pepper
1111
zxcvbn
555555
11111111
131313
freedom
777777
pass
maggie
159753
aaaaaa
ginger
princess
joshua
cheese
amanda
summer
love
ashley
nicole
chelsea
biteme
matthew
access
yankees
987654321
dallas
austin
thunder
taylor
matrix
mobilemail
mom
monitor
monitoring
montana
moon
moscow', '常用密码字典（前100个）'),
    ('common-directories', 'directory', '/
/admin
/login
/logout
/register
/api
/api/v1
/api/v2
/docs
/swagger
/swagger-ui
/redoc
/health
/status
/metrics
/debug
/test
/demo
/example
/sample
/temp
/tmp
/backup
/backups
/backup.zip
/backup.tar
/backup.tar.gz
/backup.sql
/database
/db
/data
/files
/uploads
/downloads
/static
/assets
/images
/img
/css
/js
/fonts
/vendor
/node_modules
/bower_components
/.git
/.svn
/.env
/config
/configuration
/settings
/setup
/install', '常用目录字典（简化版）')
ON CONFLICT (name, type) DO UPDATE SET
    content = EXCLUDED.content,
    description = EXCLUDED.description;

-- ============================================
-- 10. 插件表 (plugin)
-- 说明: 存储插件信息
-- ============================================
CREATE TABLE IF NOT EXISTS plugin (
    id SERIAL PRIMARY KEY,
    name VARCHAR(100) NOT NULL UNIQUE,
    version VARCHAR(20) DEFAULT '',
    author VARCHAR(100) DEFAULT '',
    description TEXT DEFAULT '',
    enabled BOOLEAN DEFAULT true,
    config TEXT DEFAULT '',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_plugin_name ON plugin(name);
CREATE INDEX IF NOT EXISTS idx_plugin_enabled ON plugin(enabled);

COMMENT ON TABLE plugin IS '插件表';
COMMENT ON COLUMN plugin.id IS '插件ID';
COMMENT ON COLUMN plugin.name IS '插件名称';
COMMENT ON COLUMN plugin.version IS '版本';
COMMENT ON COLUMN plugin.author IS '作者';
COMMENT ON COLUMN plugin.description IS '描述';
COMMENT ON COLUMN plugin.enabled IS '是否启用';
COMMENT ON COLUMN plugin.config IS '配置（JSON格式）';
COMMENT ON COLUMN plugin.created_at IS '创建时间';
COMMENT ON COLUMN plugin.updated_at IS '更新时间';

-- 创建触发器来更新 updated_at 字段
CREATE TRIGGER update_plugin_updated_at BEFORE UPDATE
    ON plugin FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- ============================================
-- 11. 通知表 (notification)
-- 说明: 存储系统通知信息
-- ============================================
CREATE TABLE IF NOT EXISTS notification (
    id SERIAL PRIMARY KEY,
    title VARCHAR(255) NOT NULL,
    content TEXT NOT NULL,
    type VARCHAR(20) DEFAULT 'info',
    read BOOLEAN DEFAULT false,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_notification_type ON notification(type);
CREATE INDEX IF NOT EXISTS idx_notification_read ON notification(read);
CREATE INDEX IF NOT EXISTS idx_notification_created_at ON notification(created_at);

COMMENT ON TABLE notification IS '通知表';
COMMENT ON COLUMN notification.id IS '通知ID';
COMMENT ON COLUMN notification.title IS '通知标题';
COMMENT ON COLUMN notification.content IS '通知内容';
COMMENT ON COLUMN notification.type IS '通知类型：info/warning/error/success';
COMMENT ON COLUMN notification.read IS '是否已读';
COMMENT ON COLUMN notification.created_at IS '创建时间';

-- ============================================
-- 12. POC验证结果表 (poc_verify_result)
-- 说明: 存储POC验证结果
-- ============================================
CREATE TABLE IF NOT EXISTS poc_verify_result (
    id SERIAL PRIMARY KEY,
    task_id INT NOT NULL,
    poc_template_id INT NOT NULL,
    target TEXT NOT NULL,
    verified BOOLEAN DEFAULT false,
    request TEXT DEFAULT '',
    response TEXT DEFAULT '',
    extra_info TEXT DEFAULT '',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (task_id) REFERENCES task(id) ON DELETE CASCADE,
    FOREIGN KEY (poc_template_id) REFERENCES poc_template(id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_poc_verify_result_task_id ON poc_verify_result(task_id);
CREATE INDEX IF NOT EXISTS idx_poc_verify_result_poc_template_id ON poc_verify_result(poc_template_id);
CREATE INDEX IF NOT EXISTS idx_poc_verify_result_verified ON poc_verify_result(verified);

COMMENT ON TABLE poc_verify_result IS 'POC验证结果表';
COMMENT ON COLUMN poc_verify_result.id IS '结果ID';
COMMENT ON COLUMN poc_verify_result.task_id IS '关联的任务ID';
COMMENT ON COLUMN poc_verify_result.poc_template_id IS '关联的POC模板ID';
COMMENT ON COLUMN poc_verify_result.target IS '验证目标';
COMMENT ON COLUMN poc_verify_result.verified IS '是否验证成功';
COMMENT ON COLUMN poc_verify_result.request IS '请求内容';
COMMENT ON COLUMN poc_verify_result.response IS '响应内容';
COMMENT ON COLUMN poc_verify_result.extra_info IS '额外信息';
COMMENT ON COLUMN poc_verify_result.created_at IS '创建时间';

-- 创建通知读取时间表（用于autoMigrate函数）
CREATE TABLE IF NOT EXISTS notification_read_time (
    id SERIAL PRIMARY KEY,
    username VARCHAR(50) NOT NULL UNIQUE,
    last_read_at TIMESTAMP DEFAULT '2000-01-01 00:00:00',
    last_cleared_at TIMESTAMP DEFAULT '2000-01-01 00:00:00'
);

CREATE INDEX IF NOT EXISTS idx_notification_read_time_username ON notification_read_time(username);

COMMENT ON TABLE notification_read_time IS '通知读取时间表';
COMMENT ON COLUMN notification_read_time.id IS '记录ID';
COMMENT ON COLUMN notification_read_time.username IS '用户名';
COMMENT ON COLUMN notification_read_time.last_read_at IS '最后读取时间';
COMMENT ON COLUMN notification_read_time.last_cleared_at IS '最后清除时间';