import{_ as a,o as i,c as n,ae as p}from"./chunks/framework.Co1PSFSm.js";const c=JSON.parse('{"title":"Instalación y Setup (Administrador)","description":"","frontmatter":{},"headers":[],"relativePath":"admin/index.md","filePath":"admin/index.md"}'),e={name:"admin/index.md"};function l(t,s,h,k,r,d){return i(),n("div",null,[...s[0]||(s[0]=[p(`<h1 id="instalacion-y-setup-administrador" tabindex="-1">Instalación y Setup (Administrador) <a class="header-anchor" href="#instalacion-y-setup-administrador" aria-label="Permalink to &quot;Instalación y Setup (Administrador)&quot;">​</a></h1><h2 id="arquitectura" tabindex="-1">Arquitectura <a class="header-anchor" href="#arquitectura" aria-label="Permalink to &quot;Arquitectura&quot;">​</a></h2><div class="language- vp-adaptive-theme"><button title="Copy Code" class="copy"></button><span class="lang"></span><pre class="shiki shiki-themes github-light github-dark vp-code" tabindex="0"><code><span class="line"><span>┌─────────────────────────────────────────────┐</span></span>
<span class="line"><span>│              CONTROL PLANE                  │</span></span>
<span class="line"><span>│  api.hwperu.com (Go + Gin + PostgreSQL)     │</span></span>
<span class="line"><span>│  - Motor de Licencias SaaS                  │</span></span>
<span class="line"><span>│  - Policy Engine por Plan                   │</span></span>
<span class="line"><span>│  - Dispatcher de Alertas (Webhook)          │</span></span>
<span class="line"><span>│  - Endpoints REST para Agente y UI          │</span></span>
<span class="line"><span>└────────────────┬────────────────────────────┘</span></span>
<span class="line"><span>                 │ HTTPS / TLS 1.3</span></span>
<span class="line"><span>┌────────────────▼────────────────────────────┐</span></span>
<span class="line"><span>│              DATA PLANE (Cliente)           │</span></span>
<span class="line"><span>│  ghcr.io/hwperu/dbp-agent:prod (Alpine Go)  │</span></span>
<span class="line"><span>│  - Heartbeat cada 10 segundos               │</span></span>
<span class="line"><span>│  - Ejecuta backups según config del API     │</span></span>
<span class="line"><span>│  - Genera Fingerprint de Hardware (SHA256)  │</span></span>
<span class="line"><span>└────────────────┬────────────────────────────┘</span></span>
<span class="line"><span>                 │ S3 Protocol</span></span>
<span class="line"><span>┌────────────────▼────────────────────────────┐</span></span>
<span class="line"><span>│              STORAGE                        │</span></span>
<span class="line"><span>│  Wasabi S3 (o cualquier S3-compatible)      │</span></span>
<span class="line"><span>│  - Datos cifrados con Restic (AES-256)      │</span></span>
<span class="line"><span>│  - Retención configurable por plan          │</span></span>
<span class="line"><span>└─────────────────────────────────────────────┘</span></span></code></pre></div><h2 id="requisitos-del-servidor-api" tabindex="-1">Requisitos del Servidor API <a class="header-anchor" href="#requisitos-del-servidor-api" aria-label="Permalink to &quot;Requisitos del Servidor API&quot;">​</a></h2><ul><li>Servidor Linux con Docker + Docker Compose</li><li>Puerto 443 (HTTPS) abierto</li><li>Dominio apuntando al servidor (<code>api.hwperu.com</code>)</li><li>Cuenta GitHub con PAT (permisos <code>write:packages</code>, <code>read:packages</code>)</li></ul><h2 id="variables-de-entorno-env-en-el-servidor-api" tabindex="-1">Variables de Entorno (<code>.env</code> en el servidor API) <a class="header-anchor" href="#variables-de-entorno-env-en-el-servidor-api" aria-label="Permalink to &quot;Variables de Entorno (\`.env\` en el servidor API)&quot;">​</a></h2><div class="language-bash vp-adaptive-theme"><button title="Copy Code" class="copy"></button><span class="lang">bash</span><pre class="shiki shiki-themes github-light github-dark vp-code" tabindex="0"><code><span class="line"><span style="--shiki-light:#6A737D;--shiki-dark:#6A737D;"># Conexión a Base de Datos</span></span>
<span class="line"><span style="--shiki-light:#24292E;--shiki-dark:#E1E4E8;">DB_HOST</span><span style="--shiki-light:#D73A49;--shiki-dark:#F97583;">=</span><span style="--shiki-light:#032F62;--shiki-dark:#9ECBFF;">postgres</span></span>
<span class="line"><span style="--shiki-light:#24292E;--shiki-dark:#E1E4E8;">DB_USER</span><span style="--shiki-light:#D73A49;--shiki-dark:#F97583;">=</span><span style="--shiki-light:#032F62;--shiki-dark:#9ECBFF;">dbp_user</span></span>
<span class="line"><span style="--shiki-light:#24292E;--shiki-dark:#E1E4E8;">DB_PASS</span><span style="--shiki-light:#D73A49;--shiki-dark:#F97583;">=</span><span style="--shiki-light:#032F62;--shiki-dark:#9ECBFF;">contraseña_segura</span></span>
<span class="line"><span style="--shiki-light:#24292E;--shiki-dark:#E1E4E8;">DB_NAME</span><span style="--shiki-light:#D73A49;--shiki-dark:#F97583;">=</span><span style="--shiki-light:#032F62;--shiki-dark:#9ECBFF;">dbp_production</span></span>
<span class="line"><span style="--shiki-light:#24292E;--shiki-dark:#E1E4E8;">DB_PORT</span><span style="--shiki-light:#D73A49;--shiki-dark:#F97583;">=</span><span style="--shiki-light:#032F62;--shiki-dark:#9ECBFF;">5432</span></span>
<span class="line"></span>
<span class="line"><span style="--shiki-light:#6A737D;--shiki-dark:#6A737D;"># Seguridad</span></span>
<span class="line"><span style="--shiki-light:#24292E;--shiki-dark:#E1E4E8;">MASTER_ADMIN_TOKEN</span><span style="--shiki-light:#D73A49;--shiki-dark:#F97583;">=</span><span style="--shiki-light:#032F62;--shiki-dark:#9ECBFF;">token_maestro_para_whmcs</span></span>
<span class="line"><span style="--shiki-light:#24292E;--shiki-dark:#E1E4E8;">API_ADMIN_KEY</span><span style="--shiki-light:#D73A49;--shiki-dark:#F97583;">=</span><span style="--shiki-light:#032F62;--shiki-dark:#9ECBFF;">llave_admin_para_endpoints</span></span>
<span class="line"><span style="--shiki-light:#24292E;--shiki-dark:#E1E4E8;">DBP_ENCRYPTION_KEY</span><span style="--shiki-light:#D73A49;--shiki-dark:#F97583;">=</span><span style="--shiki-light:#032F62;--shiki-dark:#9ECBFF;">llave_aes_de_32_caracteres</span></span>
<span class="line"></span>
<span class="line"><span style="--shiki-light:#6A737D;--shiki-dark:#6A737D;"># Distribución Privada (GHCR)</span></span>
<span class="line"><span style="--shiki-light:#24292E;--shiki-dark:#E1E4E8;">GHCR_READ_PAT</span><span style="--shiki-light:#D73A49;--shiki-dark:#F97583;">=</span><span style="--shiki-light:#032F62;--shiki-dark:#9ECBFF;">ghp_token_de_lectura_de_github</span></span>
<span class="line"></span>
<span class="line"><span style="--shiki-light:#6A737D;--shiki-dark:#6A737D;"># Wasabi S3 Global (Fallback)</span></span>
<span class="line"><span style="--shiki-light:#24292E;--shiki-dark:#E1E4E8;">WASABI_ACCESS_KEY</span><span style="--shiki-light:#D73A49;--shiki-dark:#F97583;">=</span><span style="--shiki-light:#032F62;--shiki-dark:#9ECBFF;">...</span></span>
<span class="line"><span style="--shiki-light:#24292E;--shiki-dark:#E1E4E8;">WASABI_SECRET_KEY</span><span style="--shiki-light:#D73A49;--shiki-dark:#F97583;">=</span><span style="--shiki-light:#032F62;--shiki-dark:#9ECBFF;">...</span></span>
<span class="line"><span style="--shiki-light:#24292E;--shiki-dark:#E1E4E8;">WASABI_BUCKET</span><span style="--shiki-light:#D73A49;--shiki-dark:#F97583;">=</span><span style="--shiki-light:#032F62;--shiki-dark:#9ECBFF;">hwperu-backups</span></span>
<span class="line"><span style="--shiki-light:#24292E;--shiki-dark:#E1E4E8;">WASABI_REGION</span><span style="--shiki-light:#D73A49;--shiki-dark:#F97583;">=</span><span style="--shiki-light:#032F62;--shiki-dark:#9ECBFF;">us-east-1</span></span>
<span class="line"></span>
<span class="line"><span style="--shiki-light:#6A737D;--shiki-dark:#6A737D;"># Alertas (Webhook Maestro)</span></span>
<span class="line"><span style="--shiki-light:#24292E;--shiki-dark:#E1E4E8;">WEBHOOK_URL</span><span style="--shiki-light:#D73A49;--shiki-dark:#F97583;">=</span><span style="--shiki-light:#032F62;--shiki-dark:#9ECBFF;">https://tu-n8n.com/webhook/xyz</span></span></code></pre></div><h2 id="despliegue-de-la-api" tabindex="-1">Despliegue de la API <a class="header-anchor" href="#despliegue-de-la-api" aria-label="Permalink to &quot;Despliegue de la API&quot;">​</a></h2><div class="language-bash vp-adaptive-theme"><button title="Copy Code" class="copy"></button><span class="lang">bash</span><pre class="shiki shiki-themes github-light github-dark vp-code" tabindex="0"><code><span class="line"><span style="--shiki-light:#6A737D;--shiki-dark:#6A737D;"># 1. Clonar repositorio privado</span></span>
<span class="line"><span style="--shiki-light:#6F42C1;--shiki-dark:#B392F0;">git</span><span style="--shiki-light:#032F62;--shiki-dark:#9ECBFF;"> clone</span><span style="--shiki-light:#032F62;--shiki-dark:#9ECBFF;"> https://github.com/soporte-hostingweb/dockerbackupprov.git</span></span>
<span class="line"><span style="--shiki-light:#005CC5;--shiki-dark:#79B8FF;">cd</span><span style="--shiki-light:#032F62;--shiki-dark:#9ECBFF;"> dockerbackupprov</span></span>
<span class="line"></span>
<span class="line"><span style="--shiki-light:#6A737D;--shiki-dark:#6A737D;"># 2. Configurar variables de entorno</span></span>
<span class="line"><span style="--shiki-light:#6F42C1;--shiki-dark:#B392F0;">cp</span><span style="--shiki-light:#032F62;--shiki-dark:#9ECBFF;"> .env.example</span><span style="--shiki-light:#032F62;--shiki-dark:#9ECBFF;"> .env</span></span>
<span class="line"><span style="--shiki-light:#6F42C1;--shiki-dark:#B392F0;">nano</span><span style="--shiki-light:#032F62;--shiki-dark:#9ECBFF;"> .env</span></span>
<span class="line"></span>
<span class="line"><span style="--shiki-light:#6A737D;--shiki-dark:#6A737D;"># 3. Levantar servicios</span></span>
<span class="line"><span style="--shiki-light:#6F42C1;--shiki-dark:#B392F0;">docker</span><span style="--shiki-light:#032F62;--shiki-dark:#9ECBFF;"> compose</span><span style="--shiki-light:#032F62;--shiki-dark:#9ECBFF;"> up</span><span style="--shiki-light:#005CC5;--shiki-dark:#79B8FF;"> -d</span></span>
<span class="line"></span>
<span class="line"><span style="--shiki-light:#6A737D;--shiki-dark:#6A737D;"># 4. Verificar que está activo</span></span>
<span class="line"><span style="--shiki-light:#6F42C1;--shiki-dark:#B392F0;">curl</span><span style="--shiki-light:#032F62;--shiki-dark:#9ECBFF;"> https://api.hwperu.com/ping</span></span>
<span class="line"></span>
<span class="line"><span style="--shiki-light:#6A737D;--shiki-dark:#6A737D;"># 5. Subir imagen del agente a GHCR</span></span>
<span class="line"><span style="--shiki-light:#005CC5;--shiki-dark:#79B8FF;">echo</span><span style="--shiki-light:#032F62;--shiki-dark:#9ECBFF;"> &quot;TU_GITHUB_PAT&quot;</span><span style="--shiki-light:#D73A49;--shiki-dark:#F97583;"> |</span><span style="--shiki-light:#6F42C1;--shiki-dark:#B392F0;"> docker</span><span style="--shiki-light:#032F62;--shiki-dark:#9ECBFF;"> login</span><span style="--shiki-light:#032F62;--shiki-dark:#9ECBFF;"> ghcr.io</span><span style="--shiki-light:#005CC5;--shiki-dark:#79B8FF;"> -u</span><span style="--shiki-light:#032F62;--shiki-dark:#9ECBFF;"> hwperu</span><span style="--shiki-light:#005CC5;--shiki-dark:#79B8FF;"> --password-stdin</span></span>
<span class="line"><span style="--shiki-light:#6F42C1;--shiki-dark:#B392F0;">docker</span><span style="--shiki-light:#032F62;--shiki-dark:#9ECBFF;"> build</span><span style="--shiki-light:#005CC5;--shiki-dark:#79B8FF;"> -t</span><span style="--shiki-light:#032F62;--shiki-dark:#9ECBFF;"> ghcr.io/hwperu/dbp-agent:prod</span><span style="--shiki-light:#032F62;--shiki-dark:#9ECBFF;"> ./agent/</span></span>
<span class="line"><span style="--shiki-light:#6F42C1;--shiki-dark:#B392F0;">docker</span><span style="--shiki-light:#032F62;--shiki-dark:#9ECBFF;"> push</span><span style="--shiki-light:#032F62;--shiki-dark:#9ECBFF;"> ghcr.io/hwperu/dbp-agent:prod</span></span></code></pre></div>`,9)])])}const F=a(e,[["render",l]]);export{c as __pageData,F as default};
