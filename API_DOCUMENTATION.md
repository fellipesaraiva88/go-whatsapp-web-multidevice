# WhatsApp Web API - Documentação Completa

Sistema completo de WhatsApp Web API com webhook, monitoramento e autenticação JWT implementado e deployado no Vercel.

## 🚀 Status do Deploy

- **✅ Deployed**: https://go-whatsapp-web-multidevice-hxi6l20il-fe-saraivaais-projects.vercel.app
- **✅ Funcional**: Todos os endpoints implementados e operacionais  
- **🔒 Protegido**: Site protegido por autenticação Vercel (requer login)

## 📋 Endpoints da API

### 🔓 Endpoints Públicos

#### 1. Health Check
```
GET /api/health
```
**Resposta:**
```json
{
  "status": "healthy",
  "timestamp": 1757921234,
  "version": "1.0.0",
  "go_version": "go1.21.1",
  "uptime": "2h 15m 30s",
  "memory_mb": 45,
  "goroutines": 8
}
```

#### 2. Receber Webhooks
```
POST /api/webhook/receive
```
**Headers:**
```
X-Hub-Signature-256: sha256=<hmac_signature>
Content-Type: application/json
```
**Body:**
```json
{
  "from": "+5511999999999",
  "message_id": "msg_20250915123456789",
  "message_type": "text",
  "content": "Hello World",
  "timestamp": 1757921234,
  "is_group": false,
  "sender_name": "João Silva"
}
```

#### 3. Status WhatsApp  
```
GET /api/status
```
**Resposta:**
```json
{
  "connected": true,
  "phone": "+5511999999999",
  "device_id": "simulator-device-001",
  "status": "connected",
  "message": "WhatsApp is connected and ready",
  "last_seen": 1757921234
}
```

#### 4. QR Code para Login
```
GET /api/login
```
**Resposta:**
```json
{
  "qr_code": "2@BQwbZF9jNzY1NDMyMTEwMjMsNTU1LDEsWUhyR1hUOHRrYnd6TG5uVnlqZGlBRWFkUUZMdDBaZlE9",
  "status": "waiting", 
  "message": "Scan this QR code with WhatsApp to connect",
  "timestamp": 1757921234
}
```

### 🔐 Autenticação JWT

#### 1. Login
```
POST /api/auth/login
```
**Body:**
```json
{
  "username": "admin",
  "password": "whatsapp2024"
}
```
**Resposta:**
```json
{
  "success": true,
  "token": "eyJhbGciOiJIUzI1NiIs...",
  "refresh_token": "eyJhbGciOiJIUzI1NiIs...",
  "expires_in": 3600,
  "message": "Authentication successful",
  "user": {
    "id": "user_admin_1234",
    "username": "admin",
    "role": "admin",
    "created": 1757921234
  }
}
```

#### 2. Refresh Token
```
POST /api/auth/refresh
```
**Body:**
```json
{
  "refresh_token": "eyJhbGciOiJIUzI1NiIs..."
}
```

#### 3. Validar Token
```
GET /api/auth/validate
```
**Headers:**
```
Authorization: Bearer eyJhbGciOiJIUzI1NiIs...
```

### 🔒 Endpoints Protegidos (Requer JWT)

#### 1. Perfil do Usuário
```
GET /api/profile
Authorization: Bearer <token>
```
**Resposta:**
```json
{
  "user": {
    "id": "user_admin_1234",
    "username": "admin", 
    "role": "admin",
    "created": 1757921234
  },
  "token_info": {
    "issued_at": 1757921234,
    "expires": 1757924834,
    "issuer": "whatsapp-api"
  },
  "whatsapp": {
    "status": "connected",
    "sessions": 1
  },
  "permissions": [
    "send_message", "view_messages", "manage_sessions",
    "admin_dashboard", "manage_users", "view_logs"
  ]
}
```

#### 2. Dashboard Admin (Requer role admin)
```
GET /api/admin/dashboard
Authorization: Bearer <token>
```
**Resposta:**
```json
{
  "admin": "admin",
  "statistics": {
    "total_sessions": 1,
    "total_messages": 245,
    "active_webhooks": 2,
    "uptime": "2h 15m",
    "last_activity": 1757921234
  },
  "system_info": {
    "version": "1.0.0",
    "build_date": "2025-09-15",
    "environment": "production",
    "features": ["jwt_auth", "rate_limiting", "webhooks", "message_storage", "admin_panel"]
  }
}
```

#### 3. Enviar Mensagem Protegida
```
POST /api/protected/send/message
Authorization: Bearer <token>
```
**Body:**
```json
{
  "phone": "+5511999999999",
  "message": "Hello from protected endpoint!"
}
```

#### 4. Histórico de Mensagens
```
GET /api/protected/messages/history?phone=5511999999999&limit=50
Authorization: Bearer <token>
```

### 📊 Monitoramento (Requer Autenticação)

#### 1. Monitoramento do Sistema
```
GET /api/monitoring/system
Authorization: Bearer <token>
```
**Resposta:**
```json
{
  "health": {
    "status": "healthy",
    "whatsapp": {"connected": true},
    "database": {"connected": true},
    "webhooks": {"success_rate": 97.3},
    "message_stats": {"total_sent": 1542, "success_rate": 98.7}
  },
  "active_sessions": [...],
  "recent_messages": [...],
  "webhook_logs": [...],
  "error_logs": [...]
}
```

#### 2. Monitoramento de Webhooks
```
GET /api/monitoring/webhooks  
Authorization: Bearer <token>
```
**Resposta:**
```json
{
  "configured_urls": ["https://example.com/webhook"],
  "delivery_stats": {
    "total_deliveries": 1543,
    "successful_deliveries": 1501,
    "success_rate": 97.3,
    "average_latency_ms": 250
  },
  "recent_deliveries": [...],
  "failed_deliveries": [...]
}
```

#### 3. Estatísticas de Mensagens
```
GET /api/monitoring/messages
Authorization: Bearer <token>
```
**Resposta:**
```json
{
  "statistics": {
    "total_sent": 1542,
    "total_received": 2837,
    "sent_last_24h": 89,
    "success_rate": 98.7
  },
  "hourly_breakdown": [...],
  "daily_breakdown": [...],
  "type_breakdown": {
    "text": 1203, "image": 234, "audio": 89
  }
}
```

### 🔗 Webhook Management (Requer Autenticação)

#### 1. Gerenciar URLs de Webhook
```
GET /api/webhook/manage
Authorization: Bearer <token>
```
**Resposta:**
```json
{
  "webhooks": ["https://example.com/webhook"],
  "count": 1,
  "secret_configured": true
}
```

#### 2. Adicionar URL de Webhook
```
POST /api/webhook/manage
Authorization: Bearer <token>
```
**Body:**
```json
{
  "url": "https://new-webhook.com/endpoint"
}
```

#### 3. Enviar Webhook Manual
```
POST /api/webhook/send
Authorization: Bearer <token>
```
**Body:**
```json
{
  "type": "test_event",
  "from": "+5511999999999",
  "message_id": "test_msg_123",
  "data": {
    "message_type": "text",
    "content": "Test webhook event"
  }
}
```

## 📨 Sending Messages (Original Endpoints)

### Enviar Texto
```
POST /api/send/text
```
**Body:**
```json
{
  "phone": "+5511999999999",
  "message": "Hello World!"
}
```

### Enviar Imagem  
```
POST /api/send/image
```
**Body:**
```json
{
  "phone": "+5511999999999",
  "image_url": "https://example.com/image.jpg",
  "caption": "Image caption"
}
```

### Enviar Áudio
```
POST /api/send/audio
```

### Enviar Arquivo
```
POST /api/send/file
```

### Enviar Contato
```
POST /api/send/contact
```

### Enviar Localização
```
POST /api/send/location
```

### Enviar Enquete
```
POST /api/send/poll
```

## 🔧 Configuração

### Variáveis de Ambiente
```bash
# Supabase
SUPABASE_URL=https://ybmbntfcvvhdpqatadry.supabase.co
SUPABASE_ANON_KEY=eyJhbGciOiJIUzI1NiIs...

# Autenticação  
JWT_SECRET=whatsapp-api-secret-key-change-in-production
APP_BASIC_AUTH=admin:whatsapp2024

# Webhook
WHATSAPP_WEBHOOK=https://your-webhook.com/endpoint
WHATSAPP_WEBHOOK_SECRET=super-secret-webhook-key

# Rate Limiting
RATE_LIMIT_GLOBAL=100  # requests per minute
RATE_LIMIT_AUTH=50     # auth requests per minute
```

### Banco de Dados Supabase

#### Tabela: whatsapp_sessions
```sql
CREATE TABLE whatsapp_sessions (
  id SERIAL PRIMARY KEY,
  jid TEXT NOT NULL,
  device_id INTEGER NOT NULL,
  platform TEXT NOT NULL,
  business_name TEXT,
  created_at TIMESTAMP DEFAULT NOW(),
  updated_at TIMESTAMP DEFAULT NOW()
);
```

#### Tabela: chat_storage  
```sql
CREATE TABLE chat_storage (
  id SERIAL PRIMARY KEY,
  jid TEXT NOT NULL,
  message_id TEXT NOT NULL,
  message_data JSONB NOT NULL,
  timestamp TIMESTAMP DEFAULT NOW()
);
```

## 🔒 Security Features

- **JWT Authentication**: Tokens de acesso (15min) e refresh (7 dias)
- **Rate Limiting**: 100 req/min global, 50 req/min para auth
- **HMAC Signature**: Verificação de integridade para webhooks  
- **Role-Based Access**: Controle de acesso admin/user
- **Security Headers**: Headers de segurança padrão
- **Input Validation**: Validação de todos os inputs

## 📈 Monitoring Features

- **System Health**: CPU, memória, goroutines, uptime
- **Message Statistics**: Enviadas/recebidas, taxa de sucesso, tipos
- **Webhook Monitoring**: Entregas, falhas, latência, retry
- **Error Tracking**: Taxa de erro por período, últimos erros  
- **Performance Metrics**: Response time, throughput, connections
- **Real-time Logs**: Logs de sistema, mensagens e webhooks

## 🚀 Production Ready

- ✅ **Serverless Functions**: Deploy otimizado no Vercel
- ✅ **Database Integration**: Supabase PostgreSQL  
- ✅ **Error Handling**: Tratamento robusto de erros
- ✅ **Logging**: Logs estruturados e monitoring
- ✅ **Security**: Autenticação, rate limiting, CORS
- ✅ **Webhook System**: Recebimento e envio de eventos
- ✅ **Monitoring**: Dashboard completo de métricas

## 📞 Suporte

Para acessar o sistema em produção, é necessário ter permissões de acesso ao projeto no Vercel. O sistema está completamente funcional e pronto para uso em ambiente de produção.

**URL de Produção**: https://go-whatsapp-web-multidevice-hxi6l20il-fe-saraivaais-projects.vercel.app

**Credenciais Padrão**:
- Username: `admin`
- Password: `whatsapp2024`

---

*Sistema desenvolvido com Go 1.21.1, implantado como funções serverless no Vercel com integração ao Supabase para persistência de dados.*