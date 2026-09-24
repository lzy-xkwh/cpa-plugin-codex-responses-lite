package main

// configWizardPage returns the HTML for the plugin management resource page
// served at /v0/resource/plugins/aq-codex-responses-lite/config-wizard.
// The page talks to the CPA management API from the browser (same origin as
// the admin UI) and stores the management key only in localStorage.
func configWizardPage() string {
	return wizardHTML
}

const wizardHTML = `<!DOCTYPE html>
<html lang="zh-CN">
<head>
<meta charset="utf-8">
<meta name="viewport" content="width=device-width, initial-scale=1">
<title>Codex Responses Lite 规则配置</title>
<style>
  :root { color-scheme: light dark; }
  * { box-sizing: border-box; }
  body { font-family: -apple-system, "PingFang SC", "Microsoft YaHei", sans-serif; margin: 0 auto; padding: 16px; max-width: 780px; }
  h2 { margin: 4px 0 12px; }
  h3 { margin: 20px 0 8px; }
  .muted { color: #888; font-size: 13px; }
  .card { border: 1px solid #8884; border-radius: 10px; padding: 14px; margin-bottom: 14px; }
  label { display: block; font-size: 13px; margin: 8px 0 4px; }
  input[type=text], input[type=password], textarea {
    width: 100%; padding: 7px 9px; border: 1px solid #8886; border-radius: 7px;
    font: inherit; background: transparent; color: inherit;
  }
  textarea { min-height: 72px; resize: vertical; font-family: ui-monospace, Menlo, Consolas, monospace; }
  .row { display: flex; gap: 8px; align-items: center; flex-wrap: wrap; }
  .rule { border: 1px dashed #8886; border-radius: 9px; padding: 10px; margin-bottom: 10px; }
  button {
    padding: 7px 14px; border-radius: 7px; border: 1px solid #8886; cursor: pointer;
    background: transparent; color: inherit; font: inherit;
  }
  button.primary { background: #2563eb; border-color: #2563eb; color: #fff; }
  button.danger { color: #dc2626; border-color: #dc262655; }
  #msg { margin-top: 10px; font-size: 14px; white-space: pre-wrap; }
  .ok { color: #16a34a; } .err { color: #dc2626; }
  .switch { display: flex; gap: 14px; align-items: center; margin: 6px 0; }
</style>
</head>
<body>
<h2>Codex Responses Lite 规则配置</h2>
<p class="muted">按 <code>provider_prefix/model</code> 精确（区分大小写）匹配，命中后为该请求启用 CPA 原生 Codex Responses Lite 链路，并阻止自动注入 <code>image_generation</code> 托管工具。</p>

<div class="card">
  <h3>连接</h3>
  <label>CPA 管理地址（默认当前站点）</label>
  <input type="text" id="base" placeholder="http://127.0.0.1:8317">
  <label>管理密钥（remote-management.secret-key 明文；仅保存在浏览器 localStorage）</label>
  <input type="password" id="key" placeholder="管理密钥">
  <div class="row" style="margin-top:8px">
    <button class="primary" onclick="loadCfg()">加载当前配置</button>
    <span class="muted">密钥与地址会自动记忆。</span>
  </div>
</div>

<div class="card">
  <div class="switch">
    <label style="margin:0"><input type="checkbox" id="enabled" checked> 启用插件</label>
    <span style="flex:1"></span>
    <label style="margin:0">优先级 <input type="text" id="priority" value="200" style="width:80px"></label>
  </div>
  <h3 style="margin-top:12px">匹配规则</h3>
  <div id="rules"></div>
  <div class="row">
    <button onclick="addRule()">+ 添加规则</button>
    <span style="flex:1"></span>
    <button class="primary" onclick="saveCfg()">保存到 CPA</button>
  </div>
</div>

<div id="msg"></div>

<script>
const ID = 'aq-codex-responses-lite';
const $ = id => document.getElementById(id);

function base() { return $('base').value.trim().replace(/\/+$/, ''); }
function authHeaders() {
  return { 'Authorization': 'Bearer ' + $('key').value.trim(), 'Content-Type': 'application/json' };
}
function note(text, cls) {
  $('msg').innerHTML = '<span class="' + (cls || '') + '">' + text + '</span>';
}
function persistConn() {
  localStorage.setItem('cqrl.base', $('base').value);
  localStorage.setItem('cqrl.key', $('key').value);
}
function restoreConn() {
  $('base').value = localStorage.getItem('cqrl.base') || window.location.origin;
  $('key').value = localStorage.getItem('cqrl.key') || '';
}

function addRule(prefix, models) {
  const div = document.createElement('div');
  div.className = 'rule';
  div.innerHTML =
    '<label>provider_prefix（单个路径段，区分大小写）</label>' +
    '<input type="text" class="prefix" placeholder="例如 opencode">' +
    '<label>models（每行一个上游模型名，不含前缀）</label>' +
    '<textarea class="models" placeholder="grok-4.5&#10;muse-spark-1.2-contributor"></textarea>' +
    '<div class="row" style="margin-top:6px"><span style="flex:1"></span>' +
    '<button class="danger" onclick="this.closest(\'.rule\').remove()">删除此规则</button></div>';
  div.querySelector('.prefix').value = prefix || '';
  div.querySelector('.models').value = (models || []).join('\n');
  $('rules').appendChild(div);
}

function collectCfg() {
  const rules = [];
  for (const div of document.querySelectorAll('#rules .rule')) {
    const prefix = div.querySelector('.prefix').value.trim();
    const models = div.querySelector('.models').value.split('\n')
      .map(s => s.trim()).filter(Boolean);
    if (!prefix && models.length === 0) continue; // 整行空白则忽略
    rules.push({ provider_prefix: prefix, models: models });
  }
  return {
    enabled: $('enabled').checked,
    priority: parseInt($('priority').value, 10) || 200,
    rules: rules
  };
}

async function loadCfg() {
  persistConn();
  try {
    const resp = await fetch(base() + '/v0/management/plugins/' + ID + '/config', { headers: authHeaders() });
    if (resp.status === 404) {
      $('rules').innerHTML = '';
      addRule('', []);
      note('当前未配置任何规则（插件处于空闲状态）。添加规则后保存即可。', 'ok');
      return;
    }
    if (!resp.ok) throw new Error(resp.status + ' ' + (await resp.text()));
    const cfg = await resp.json();
    if (typeof cfg.enabled === 'boolean') $('enabled').checked = cfg.enabled;
    if (cfg.priority !== undefined && cfg.priority !== null) $('priority').value = cfg.priority;
    $('rules').innerHTML = '';
    const rules = Array.isArray(cfg.rules) ? cfg.rules : [];
    if (rules.length === 0) addRule('', []);
    for (const r of rules) addRule(r.provider_prefix, Array.isArray(r.models) ? r.models : []);
    note('已加载当前配置。', 'ok');
  } catch (e) {
    note('加载失败：' + e.message, 'err');
  }
}

async function saveCfg() {
  persistConn();
  const cfg = collectCfg();
  for (let i = 0; i < cfg.rules.length; i++) {
    const r = cfg.rules[i];
    if (!r.provider_prefix || r.provider_prefix.includes('/')) {
      note('规则 ' + (i + 1) + ' 的 provider_prefix 不能为空且只能是单个路径段（不含 /）。', 'err');
      return;
    }
    if (r.models.length === 0) {
      note('规则 ' + (i + 1) + '（' + r.provider_prefix + '）至少需要一个模型名。', 'err');
      return;
    }
  }
  try {
    const resp = await fetch(base() + '/v0/management/plugins/' + ID + '/config', {
      method: 'PUT', headers: authHeaders(), body: JSON.stringify(cfg)
    });
    const text = await resp.text();
    if (!resp.ok) throw new Error(resp.status + ' ' + text);
    note('已保存并热加载。命中规则的请求（' +
      cfg.rules.map(r => r.provider_prefix + '/*').join(', ') +
      '）将启用 Responses Lite。', 'ok');
  } catch (e) {
    note('保存失败：' + e.message, 'err');
  }
}

restoreConn();
addRule('', []);
</script>
</body>
</html>`
