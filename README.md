# Maré Viva

**Observatório de Tubarões do Nordeste** — uma plataforma acadêmica de conservação marinha, ciência cidadã e educação ambiental. “Proteger os tubarões é proteger os oceanos.” O projeto contribui com o **ODS 14 — Vida na Água**, articulando educação de qualidade (ODS 4), inovação (ODS 9) e consumo responsável (ODS 12).

## Problema e proposta

A desinformação sobre tubarões, a pressão pesqueira e a degradação de habitats dificultam a conservação. O Maré Viva reúne conteúdo científico com referências explícitas, pesquisa regional, visualizações educativas e relatos comunitários sujeitos à moderação. A plataforma é um protótipo acadêmico independente, sem vínculo institucional com as iniciativas citadas.

## Funcionalidades

- Portal público responsivo com catálogo de 23 espécies, busca, observatório de indicadores, conteúdo sobre o Nordeste, quiz e simulador ecológico executado pela API Go.
- 12 indicadores científicos com ano, unidade, abrangência, grupo taxonômico e fonte. Indicadores de tubarões, elasmobrânquios e conservação marinha geral são identificados separadamente.
- Ciência cidadã persistida em SQLite. Relatos reais começam como `PENDING`; somente aprovados aparecem publicamente e continuam identificados como não validados cientificamente.
- 40 registros sintéticos opcionais para demonstração, distribuídos pelos nove estados nordestinos, com marcação explícita de “DEMONSTRAÇÃO FICTÍCIA” na API e nas telas.
- Painel administrativo com JWT em cookies HttpOnly, dashboard, CRUD de espécies/indicadores/instituições/referências/quiz e usuários por papel, além de moderação e auditoria.
- 12 instituições e projetos referenciados e 30 questões de quiz com explicação e referência.

## Stack e arquitetura

- Frontend: React, Vite, TypeScript, Tailwind CSS, React Router, TanStack Query, React Hook Form, Zod, Lucide e Recharts.
- Backend: Go, `net/http`, `database/sql`, SQLite via `modernc.org/sqlite`, JWT e bcrypt.
- Dados: migrations versionadas em `backend/migrations/`; seeds estruturados e separados em `backend/seeds/`.
- A API é servida localmente em `:8080`; o Vite encaminha `/api` para o backend durante desenvolvimento. Com o `.env.example`, o banco usa `backend/data/mare-viva.db`; sem `DB_PATH`, o executável usa `backend/mare-viva.db`.

## Requisitos

Go 1.23 ou superior, Node.js 20.19 ou superior, npm, Make e Git. Não é necessário Docker nem serviço externo.

## Instalação e execução local

```sh
cp .env.example .env
# Edite JWT_SECRET com um segredo aleatório de pelo menos 32 caracteres.
# Configure ADMIN_EMAIL e ADMIN_PASSWORD para criar o primeiro administrador.
make setup
make migrate
make seed-science
make dev
```

Portal: <http://localhost:5173> · API: <http://localhost:8080/api/health> · Administração: <http://localhost:5173/admin/login>.

O administrador inicial só é criado quando `ADMIN_EMAIL` e `ADMIN_PASSWORD` estão configurados e não há administrador ativo. O processo não redefine senhas de contas existentes. Use credenciais próprias; `.env` e bancos locais são ignorados pelo Git.

### Seeds e persistência

- `make migrate`: aplica migrations sem remover ou recriar o banco existente.
- `make seed-science`: insere espécies, indicadores, instituições, fontes e quiz. É idempotente; não altera registros que já existem nem redefine usuários.
- `make seed-demo`: insere 40 ocorrências fictícias idempotentes. Rode somente quando quiser dados de demonstração; elas são claramente marcadas na interface.
- `make seed`: alias para `make seed-science`.

Para usar outro banco, defina `DB_PATH` no `.env`. Nunca exclua o arquivo SQLite para “resetar” dados sem antes fazer backup e decidir explicitamente o que deve ser preservado.

## Testes e qualidade

```sh
cd backend && go fmt ./... && go vet ./... && go test ./...
cd frontend && npm run lint && npm run build
cd .. && make build
```

Ou `make test` para os testes Go, lint e build do frontend. O projeto inclui testes HTTP com `httptest` e testes de autenticação, moderação, CRUD e simulador.

## Referências científicas

A base combina a lista nacional do MMA/ICMBio, avaliações de espécies ligadas à IUCN Red List, artigo revisado por pares sobre Recife, PAN Tubarões, relatório global IUCN SSC e indicadores da ONU/FAO. Os itens completos estão em `backend/seeds/science.json` e [docs/referencias.md](docs/referencias.md). Categorias nacionais e globais não são equivalentes. Dados agregados de tubarões e raias nunca são apresentados como exclusivos de tubarões.

## Limitações

O simulador é um modelo conceitual educativo, não uma previsão científica. Relatos cidadãos não são validação científica e a plataforma não os encaminha automaticamente a autoridades. Estudos locais e amostras não estimam populações estaduais ou nacionais. Nem todos os táxons têm fotografia específica local; o catálogo usa silhueta original quando a correspondência de imagem não foi confirmada. Dados já existentes são preservados pelo seed, então campos antigos de espécies podem permanecer como cadastrados até revisão editorial no painel.

## Licenças e imagens

Código e ilustrações originais deste protótipo são licenciados sob MIT (ver `LICENSE`). Fotografias de terceiros mantêm as licenças próprias indicadas em [docs/atribuicoes-de-imagens.md](docs/atribuicoes-de-imagens.md); cada obra deve ser redistribuída com sua atribuição e termos aplicáveis.
