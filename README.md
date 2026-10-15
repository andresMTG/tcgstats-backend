# TCGStats Backend

API backend do TCGStats, uma ferramenta para jogadores de Trading Card Games
(TCGs) registrarem seus decks, torneios, oponentes e partidas. O objetivo é
transformar esse histórico em informações que ajudem o jogador a entender seu
desempenho e tomar decisões melhores sobre o que praticar e jogar.

## O que a API oferece

- Cadastro e autenticação de usuários.
- Registro e consulta de decks, torneios, oponentes e partidas.
- Análises do desempenho por deck, oponente, fase e nível do torneio.
- Filtros de período nas análises para comparar recortes do histórico.

As análises usam as partidas associadas aos torneios do usuário autenticado.
A taxa de vitória (`win_rate`) é calculada como vitórias divididas pelo total
de partidas, em porcentagem. Empates fazem parte do total e também são
contabilizados separadamente.

## Tecnologias

- **Go 1.25.11** — linguagem e runtime do backend.
- **PostgreSQL** — banco de dados relacional.
- **GORM** — acesso ao banco e persistência dos modelos.
- **Chi** — roteamento HTTP da API.
- **Uber Fx** — injeção de dependências e composição dos módulos.
- **JWT** — autenticação das rotas protegidas.
- **bcrypt** — hash de senhas.
- **Redis** — sessões de refresh token com rotação e revogação.
- **golang-migrate** — execução de migrações do banco.
- **Testify** — assertions e mocks nos testes.

O Redis é usado para armazenar sessões de refresh token. A aplicação requer
Redis 6.2 ou superior para consumir e revogar tokens atomicamente com `GETDEL`.

Para iniciar a instância local definida em `docker-compose.yml`, pare qualquer
Redis que já esteja usando a porta `6379` e execute na pasta do backend:

```powershell
docker compose up -d redis
docker compose ps
```

Os dados são mantidos no volume Docker `redis-data` mesmo que o container seja
recriado. O serviço publica a porta somente em `127.0.0.1`, conforme o padrão
local de `REDIS_ADDR`.

## Autenticação e sessões

O login (`POST /api/users/login`) recebe `email` e `password` em JSON. Em caso
de sucesso, retorna `204 No Content` e grava dois cookies `HttpOnly`:

- `access_token`: JWT de acesso, válido por 15 minutos por padrão.
- `refresh_token`: JWT de renovação, válido por 30 dias por padrão e enviado
  apenas às rotas `/api/users`.

Os tokens não são incluídos no corpo da resposta. Cookies são `Secure` e
`SameSite=Lax` por padrão. O middleware aceita o cookie de acesso; também
mantém compatibilidade com clientes que enviem `Authorization: Bearer`.

- `POST /api/users/refresh` consome o refresh token atual e emite um novo par
  de cookies. Cada refresh token só pode ser usado uma vez.
- `POST /api/users/logout` revoga a sessão de refresh no Redis e expira os
  cookies no navegador.

Configuração de autenticação e Redis por variáveis de ambiente:

| Variável | Padrão | Descrição |
| --- | --- | --- |
| `JWT_SECRET` | — | Segredo obrigatório para assinar e validar JWTs. |
| `JWT_EXPIRATION_MINUTES` | `15` | Validade do JWT de acesso. |
| `JWT_REFRESH_EXPIRATION_DAYS` | `30` | Validade máxima do refresh token. |
| `AUTH_COOKIE_SECURE` | `true` | Use `true` em HTTPS; em desenvolvimento local HTTP, configure `false`. |
| `AUTH_COOKIE_SAME_SITE` | `lax` | Política SameSite: `lax`, `strict` ou `none`. `none` exige Secure. |
| `REDIS_ADDR` | `127.0.0.1:6379` | Endereço do Redis. |
| `REDIS_PASSWORD` | vazio | Senha do Redis, se configurada. |
| `REDIS_DB` | `0` | Índice do banco lógico Redis. |
| `CORS_ALLOWED_ORIGINS` | vazio | Origens exatas permitidas, separadas por vírgula. Ex.: `https://app.example.com`. |

Com `SameSite=Lax`, o frontend e a API devem ser servidos no mesmo site. Se
forem sites diferentes, será necessário avaliar `SameSite=None`, HTTPS,
credenciais CORS e proteção CSRF compatíveis com a implantação.

O backend habilita CORS com credenciais apenas para origens listadas em
`CORS_ALLOWED_ORIGINS`. A lista não aceita `*`, caminhos ou curingas; informe
somente as origens (esquema e host), por exemplo
`CORS_ALLOWED_ORIGINS=http://localhost:3000,https://app.example.com`. Sem uma
origem configurada, requisições sem `Origin` continuam funcionando, mas
requisições de navegador com `Origin` são recusadas. O frontend deve enviar
cookies usando `credentials: "include"`.

## Configuração do ambiente

Copie `.env.example` para `.env` antes de iniciar o backend:

```powershell
Copy-Item .env.example .env
```

Configure `DSN` e `JWT_SECRET` no arquivo. O exemplo usa PostgreSQL e Redis
locais; ajuste credenciais, host e portas conforme o ambiente. Gere um segredo
JWT próprio com um gerador criptograficamente seguro (por exemplo,
`openssl rand -base64 48`) e nunca reutilize o valor de exemplo em produção.
O arquivo `.env` é ignorado pelo Git.

Variáveis disponíveis:

| Variável | Padrão | Descrição |
| --- | --- | --- |
| `DSN` | — | String de conexão PostgreSQL (obrigatória). |
| `APP_NAME` | `tcgstats-backend` | Nome da aplicação. |
| `SERVER_PORT` | `8080` | Porta HTTP. |
| `SERVER_BASE_PATH` | `/api/` | Prefixo das rotas da API. |
| `PPROF_ENABLED` | `false` | Habilita profiling quando suportado pela aplicação. |
| `JWT_SECRET` | — | Segredo de assinatura JWT (obrigatório pelo serviço de auth). |
| `JWT_EXPIRATION_MINUTES` | `15` | Validade do token de acesso. |
| `JWT_REFRESH_EXPIRATION_DAYS` | `30` | Validade do refresh token. |
| `AUTH_COOKIE_SECURE` | `true` | Define o atributo Secure; use `false` apenas em desenvolvimento HTTP local. |
| `AUTH_COOKIE_SAME_SITE` | `lax` | Política SameSite: `lax`, `strict` ou `none`; `none` requer Secure. |
| `CORS_ALLOWED_ORIGINS` | vazio | Origens exatas separadas por vírgula, sem caminhos ou barra final. |
| `REDIS_ADDR` | `127.0.0.1:6379` | Endereço do Redis. |
| `REDIS_PASSWORD` | vazio | Senha do Redis, se configurada. |
| `REDIS_DB` | `0` | Índice do banco lógico do Redis. |
| `DB_MAX_OPEN_CONNS` | `25` | Limite de conexões abertas do pool PostgreSQL. |
| `DB_MAX_IDLE_CONNS` | `5` | Limite de conexões ociosas do pool PostgreSQL. |
| `DB_CONN_MAX_LIFETIME` | `5m` | Tempo máximo de vida de uma conexão PostgreSQL. |
| `DB_CONN_MAX_IDLE_TIME` | `1m` | Tempo máximo ocioso de uma conexão PostgreSQL. |

O modo de desenvolvimento do exemplo assume frontend em
`http://localhost:5173`, API em `http://localhost:8080`, PostgreSQL em
`localhost:5432` e Redis em `127.0.0.1:6379`. Em produção, configure variáveis
equivalentes pelo gerenciador de secrets/ambiente da plataforma, use HTTPS,
`AUTH_COOKIE_SECURE=true` e informe somente as origens reais do frontend.

## Endpoints de análise

Todas as rotas de análise exigem autenticação pelo cookie `access_token` ou,
para compatibilidade, pelo header bearer:

| Método | Endpoint | Agrupamento |
| --- | --- | --- |
| `GET` | `/api/analytics/decks` | Desempenho por deck |
| `GET` | `/api/analytics/opponents` | Desempenho por oponente |
| `GET` | `/api/analytics/phases` | Desempenho por fase do torneio |
| `GET` | `/api/analytics/tournament-scopes` | Desempenho por nível do torneio |

Todas aceitam os filtros opcionais inclusivos `from` e `to`, no formato
`YYYY-MM-DD`. Por exemplo:

```text
GET /api/analytics/decks?from=2026-01-01&to=2026-12-31
```

Cada item retornado inclui `matches`, `wins`, `draws`, `losses` e `win_rate`,
além do identificador e nome do deck/oponente ou da fase/nível correspondente.

## Desenvolvimento

Execute os testes com:

```sh
go test ./...
```

Os testes unitários de controllers e services usam mocks e não precisam de
uma instância do PostgreSQL.
