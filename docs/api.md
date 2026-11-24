# API REST

A resposta de sucesso usa `{ "data": ... }` ou um objeto documentado pelo endpoint. Erros retornam `{ "error": "..." }`. Listas públicas disponíveis: `GET /api/species`, `/api/indicators`, `/api/institutions`, `/api/sources`, `/api/observations` (somente aprovadas; resposta identifica explicitamente as demonstrações fictícias) e `/api/quiz`.

| Método e caminho | Uso |
|---|---|
| `GET /api/health` | Estado da API |
| `GET /api/species/{id}` | Detalhe por identificador |
| `POST /api/observations` | Enviar relato, sempre pendente |
| `POST /api/ecosystem/simulate` | Modelo conceitual com `{"sharks":0..100}` |
| `POST /api/quiz/evaluate` | Avalia `{ "answers": { "questionId": optionId } }` no servidor |
| `POST /api/auth/login` | Login; define cookies de sessão |
| `POST /api/auth/refresh` | Rotaciona refresh e emite novo access |
| `POST /api/auth/logout` | Revoga refresh; exige CSRF |
| `GET /api/auth/me` | Usuário atual |
| `GET /api/admin/dashboard` | Métricas do banco (ADMIN/EDITOR/MODERATOR) |
| `GET/POST /api/admin/species` | Listar/criar espécies (ADMIN/EDITOR) |
| `PUT/DELETE /api/admin/species/{id}` | Atualizar/arquivar espécie (ADMIN/EDITOR) |
| `GET/POST /api/admin/{indicators,institutions,sources}` | Listar/criar conteúdo editorial (ADMIN/EDITOR) |
| `PUT/DELETE /api/admin/{indicators,institutions,sources}/{id}` | Editar/remover conteúdo editorial (ADMIN/EDITOR) |
| `GET/POST /api/admin/quiz` e `PUT/DELETE /api/admin/quiz/{id}` | CRUD de perguntas e alternativas (ADMIN/EDITOR) |
| `GET/POST /api/admin/users` e `PUT/DELETE /api/admin/users/{id}` | Gestão de contas (exclusivo ADMIN); DELETE desativa conta |
| `GET /api/admin/observations` | Filtros por estado, município, categoria e status |
| `GET /api/admin/observations/{id}` | Detalhe da ocorrência |
| `PATCH /api/admin/observations/{id}/status` | Moderação e histórico |
| `GET /api/admin/audit-logs` | Auditoria (ADMIN) |

Requisições mutáveis autenticadas precisam de `X-CSRF-Token` igual ao cookie `mare_csrf`. O CORS só permite a origem definida em `CORS_ORIGIN`.

## Seeds e demonstrações

`make seed-science` é idempotente e não sobrescreve registros existentes. `make seed-demo` cria 40 ocorrências claramente marcadas como fictícias por `demonstration: true` e pelo texto do registro. O dashboard retorna contagens separadas para relatos da comunidade e demonstrações; os gráficos por estado informam as duas parcelas.

## Escopo ainda não implementado

Os módulos administrativos de espécies, indicadores, instituições, fontes, quiz e usuários possuem CRUD básico na interface e API. Permanecem pendentes paginação comum a todas as listas e uma tela detalhada de histórico de observações.
