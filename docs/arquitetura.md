# Arquitetura

O projeto usa uma SPA React/TypeScript servida em desenvolvimento pelo Vite e uma API monolítica Go com `net/http`. O Vite encaminha `/api` para `localhost:8080`. O backend usa `database/sql` com SQLite (`modernc.org/sqlite`), migrations SQL versionadas e seeds JSON idempotentes. `make seed-science` populará conteúdo científico, enquanto `make seed-demo` inclui relatos fictícios separadamente e sob demanda.

## Dados

As tabelas cobrem usuários, sessões de refresh, espécies, indicadores, instituições, fontes, relacionamentos de fontes, ocorrências, histórico, quiz e auditoria. SQLite opera com foreign keys ativadas. Dados de cidadãos são armazenados sem coordenadas precisas.

## Sessões

O access JWT HS256 expira em 15 minutos e valida algoritmo, issuer e audience. O refresh aleatório expira em sete dias; somente seu SHA-256 é persistido e cada rotação revoga o anterior. Cookies são HttpOnly, SameSite e Secure quando `APP_ENV=production`. O cookie CSRF é enviado no header `X-CSRF-Token` para operações mutáveis. Papéis são conferidos no servidor em cada requisição protegida. Login tem limite simples de dez tentativas por endereço em cinco minutos.

## Execução e limitações

O processo aplica migrations pendentes ao iniciar; o histórico fica em `schema_migrations`. Defina `DB_PATH` dentro de `backend/data` para manter o arquivo local ignorado pelo Git. O rate limit é em memória por processo e não é compartilhado entre instâncias. O protótipo não inclui fluxo de recuperação de senha, gerenciamento completo de usuários nem formulários editoriais para todos os módulos.
