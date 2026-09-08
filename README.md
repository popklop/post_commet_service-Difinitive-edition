# Post & Comments Service

GraphQL-сервис для публикации постов и иерархических комментариев к ним, аналогичный по духу комментариям на Habr / Reddit.

Проект реализован на **Go** и поддерживает:

* GraphQL API на базе `gqlgen`
* иерархические комментарии с неограниченной глубиной вложенности
* cursor-based pagination в стиле Relay
* два backend-хранилища: **PostgreSQL** и **in-memory**
* выбор хранилища при запуске без пересборки приложения
* GraphQL Subscriptions для real-time уведомлений о новых комментариях
* DataLoader для устранения N+1 запросов
* unit-тесты и интеграционные тесты PostgreSQL
* Docker / Docker Compose
* graceful shutdown

---

## Возможности

### Посты

* создание поста
* получение отдельного поста
* получение списка постов
* cursor-based пагинация
* включение / отключение комментирования автором поста

### Комментарии

* создание комментария
* комментарий может быть корневым или ответом на другой комментарий
* неограниченная глубина вложенности
* максимальная длина текста — **2000 символов**
* получение корневых комментариев поста
* получение дочерних комментариев
* cursor-based pagination

### Real-time подписки

Новые комментарии можно получать через GraphQL Subscription:

```graphql
subscription {
  commentAdded(postId: "<post-id>") {
    id
    text
    author
  }
}
```

После успешного создания комментария сервис асинхронно отправляет его подписчикам соответствующего поста.
  Построено это всё на базе буферизированных каналов, размер для тестовых запусков был выбран небольшой, 10, так что,
если данные будут приходить при полном канале, они, к сожалению потеряются, целостность данных в таких сценариях и брокер сообщений по типу RabbitMQ не реализован.

---

## Стек

* **Go**
* **GraphQL / gqlgen**
* **PostgreSQL**
* **database/sql**
* **graph-gophers/dataloader**
* **testify**
* **testcontainers-go**
* **Docker / Docker Compose**

---

## Архитектура

Проект разделён на несколько уровней:

```text
GraphQL Resolver
       │
       ▼
    Service
       │
       ▼
 Repository interface
      / \
     /   \
Memory   PostgreSQL
```

Дополнительно для GraphQL-запросов с вложенными комментариями используется DataLoader:

```text
GraphQL Resolver
       │
       ▼
   DataLoader
       │
       ▼
    Service
       │
       ▼
   Repository
       │
       ▼
 PostgreSQL / Memory
```

Основная идея разделения:

* **entity** — доменная модель и её инварианты
* **repository** — работа с хранилищем
* **service** — бизнес-логика
* **resolver** — GraphQL API
* **dataloader** — batching вложенных запросов
* **server** — HTTP/GraphQL server и middleware

---

## Структура проекта

```text
.
├── cmd/
│   └── server/
│       └── main.go
│
├── graph/
│   ├── generated/          # generated gqlgen code
│   ├── model/              # generated GraphQL models
│   ├── resolver/           # GraphQL resolvers
│   └── schema.graphqls     # GraphQL schema
│
├── internal/
│   ├── entity/
│   │   ├── post/
│   │   └── comment/
│   │
│   ├── repository/
│   │   ├── repository.go
│   │   ├── memory/
│   │   └── postgres/
│   │
│   ├── service/
│   │
│   ├── dataloader/
│   │
│   ├── pagination/
│   │
│   └── server/
│
├── migrations/
│   ├── 000001_init.up.sql
│   └── 000001_init.down.sql
│
├── Dockerfile
├── docker-compose.yml
├── go.mod
└── README.md
```

---

## Data model

Комментарии представлены через классическую adjacency-list модель:

```text
Comment
 ├── id
 ├── post_id
 └── parent_id
```

`parent_id = NULL` означает корневой комментарий.

Дочерние комментарии связаны с родителем через foreign key:

```sql
parent_id UUID REFERENCES comments(id)
```

Для идентификаторов используется **UUID**, а не автоинкрементный `SERIAL`.

Основные индексы:

```sql
CREATE INDEX idx_posts_created_at_id
    ON posts(created_at DESC, id DESC);

CREATE INDEX idx_comments_post_id_created_at_id
    ON comments(post_id, created_at DESC, id DESC);

CREATE INDEX idx_comments_parent_id_created_at_id
    ON comments(parent_id, created_at DESC, id DESC);
```

Составные индексы используются для эффективного поиска и стабильной сортировки при курсорной пагинации.

---

## GraphQL API

### Query

```graphql
type Query {
  posts(first: Int, after: String): PostConnection!
  post(id: ID!): Post
}
```

Пример:

```graphql
query {
  posts(first: 10) {
    edges {
      cursor
      node {
        id
        title
        content
        author
        commentsEnabled
        createdAt
      }
    }
    pageInfo {
      hasNextPage
      endCursor
    }
  }
}
```

Получение одного поста вместе с комментариями:

```graphql
query {
  post(id: "<post-id>") {
    id
    title
    content

    comments(first: 10) {
      edges {
        cursor
        node {
          id
          text
          author
          children(first: 10) {
            edges {
              node {
                id
                text
              }
            }
          }
        }
      }

      pageInfo {
        hasNextPage
        endCursor
      }
    }
  }
}
```

---

## Mutations

Создание поста:

```graphql
mutation {
  createPost(
    input: {
      title: "My post"
      content: "Hello world"
      author: "<author-uuid>"
    }
  ) {
    id
    title
  }
}
```

Запретить комментирование:

```graphql
mutation {
  updatePostCommentsEnabled(
    id: "<post-id>"
    enabled: false
  ) {
    id
    commentsEnabled
  }
}
```

Создание корневого комментария:

```graphql
mutation {
  createComment(
    input: {
      text: "Hello!"
      author: "<author-uuid>"
      postId: "<post-id>"
    }
  ) {
    id
    text
  }
}
```

Ответ на другой комментарий:

```graphql
mutation {
  createComment(
    input: {
      text: "Reply"
      author: "<author-uuid>"
      postId: "<post-id>"
      parentId: "<comment-id>"
    }
  ) {
    id
    text
    parentId
  }
}
```

---

## Pagination

Для списков используется cursor-based pagination:

```graphql
posts(first: 10, after: "<cursor>")
comments(first: 10, after: "<cursor>")
children(first: 10, after: "<cursor>")
```
Где:
first - Кол-во комментариев после after.

Вместо `OFFSET` используется стабильная сортировка:

```text
created_at DESC
id         DESC
```

На уровне SQL это позволяет использовать условие:

```sql
WHERE (created_at, id) < ($cursorCreatedAt, $cursorID)
```

и получать следующую страницу без необходимости пропускать большое количество строк через `OFFSET`.

---

## Решение N+1

Вложенные поля:

```graphql
Post.comments
Comment.children
```

загружаются через DataLoader.

Например, запрос:

```graphql
posts {
  comments {
    ...
  }
}
```

без batching мог бы привести к схеме:

```text
1 query → posts

N queries → comments for each post
```

DataLoader собирает идентификаторы за один execution cycle и передаёт их в batch-метод repository:

```text
multiple GraphQL resolvers
          │
          ▼
      DataLoader
          │
          ▼
   one batch request
          │
          ▼
        SQL
```

Для выборки нескольких родителей используется оконная функция:

```sql
ROW_NUMBER() OVER (
    PARTITION BY parent_id
    ORDER BY created_at DESC, id DESC
)
```

Это позволяет ограничивать количество загружаемых комментариев **для каждого родителя отдельно**.

Аналогичный подход используется для корневых комментариев нескольких постов.

---

## Валидация

Доменный слой проверяет основные инварианты комментария:

* текст не должен быть пустым
* длина текста — не более 2000 Unicode-символов
* `author` должен быть валидным UUID
* `post_id` должен быть валидным UUID

При создании вложенного комментария сервис дополнительно проверяет, что родительский комментарий принадлежит тому же посту.

---

## Storage

Хранилище выбирается через переменную окружения:

```bash
STORAGE_TYPE=memory
```

или:

```bash
STORAGE_TYPE=postgres
```

### PostgreSQL

Подключение задаётся через:

```bash
POSTGRES_DSN=postgres://postgres:pass@localhost:5432/postcommozon?sslmode=disable
```

### In-memory

Для локального запуска можно использовать:

```bash
STORAGE_TYPE=memory
```

Переключение backend'а не требует изменения бизнес-логики и пересборки отдельных компонентов приложения.

---

## Запуск через Docker Compose

```bash
docker compose up --build
```

После запуска будет доступен графкул плейграунд, откуда можно удобно отправлять запросы:

### GraphQL Playground

```text
http://localhost:8080/
```

### GraphQL endpoint

```text
http://localhost:8080/query
```

PostgreSQL запускается отдельным контейнером.

---

## Локальный запуск

Для запуска с in-memory storage:

```bash
export STORAGE_TYPE=memory
go run ./cmd/server
```

Для PostgreSQL:

```bash
export STORAGE_TYPE=postgres
export POSTGRES_DSN="postgres://postgres:pass@localhost:5432/postcommozon?sslmode=disable"

go run ./cmd/server
```

При запуске через Docker Compose миграции применяются автоматически при первом создании базы.
При локальном запуске без Docker миграции необходимо применить вручную.

---

## Graceful shutdown

При получении:

```text
SIGINT
SIGTERM
```

приложение:

1. прекращает приём новых HTTP-запросов;
2. выполняет `http.Server.Shutdown` с timeout;
3. закрывает соединение с PostgreSQL.

Shutdown timeout — 5 секунд.

---

## Testing

Unit-тесты:

```bash
go test ./...
```

Интеграционные тесты PostgreSQL:

```bash
go test -tags=integration ./...
```

Интеграционные тесты используют `testcontainers-go` и поднимают временный PostgreSQL-контейнер.

Проверяются, в частности:

* создание и получение постов
* создание и получение комментариев
* pagination
* cursor handling
* not-found cases
* ограничения длины комментария
* бизнес-правила
* PostgreSQL repository
* cascading delete

Для запуска интеграционных тестов необходим Docker.

---

## Известные ограничения

### DataLoader pagination
Так как при разработке, архитектура была выстроена без учёта даталоадеров и батчинга, а план их реализации был уже после сформированного скелета, не удалось сделать нормальную пагинацию в даталоадерах, было принято решение использовать компромиссное решение на фиксированное кол-во данных в окне,
исправление, и полноценная пагинация требует рефакторинга, в настоящем виде проект работает так:

Batch-загрузка комментариев использует фиксированное окно элементов на родителя, после чего GraphQL resolver применяет `first/after` к полученному набору.

Это позволяет устранить N+1 и одновременно ограничить объём batch-выборки, но при очень глубокой pagination одного окна может оказаться недостаточно.

---
