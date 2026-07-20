# GraphQL API

Схема находится в [`internal/transport/graph/schema`](../internal/transport/graph/schema).

## Эндпоинты

```text
HTTP | WebSocket: /query
Playground:       /        только при ENVIRONMENT=local
```

## Query

### Список постов

```graphql
query Posts($first: Int! = 20, $after: Cursor) {
  posts(first: $first, after: $after) {
    nodes {
      id
      authorId
      title
      body
      commentsEnabled
      createdAt
      updatedAt
      author {
        id
        name
      }
    }
    pageInfo {
      endCursor
      hasNextPage
    }
  }
}
```

### Пост по ID

```graphql
query Post($id: ID!) {
  post(id: $id) {
    id
    title
    body
    commentsEnabled
    author {
      id
      name
    }
    comments(first: 20) {
      nodes {
        id
        postId
        parentId
        body
        createdAt
      }
      pageInfo {
        endCursor
        hasNextPage
      }
    }
  }
}
```

`Post.comments` возвращает только корневые комментарии, то есть записи с `parentId = null`.

### Комментарий по ID

```graphql
query Comment($id: ID!) {
  comment(id: $id) {
    id
    postId
    parentId
    body
    createdAt
    author {
      id
      name
    }
    replies(first: 20) {
      nodes {
        id
        parentId
        body
      }
      pageInfo {
        endCursor
        hasNextPage
      }
    }
  }
}
```

### Явное получение страницы комментариев

`comments` позволяет получить корневую страницу либо непосредственных детей указанного комментария.

```graphql
query Comments($postId: ID!, $parentId: ID, $first: Int! = 20, $after: Cursor) {
  comments(postId: $postId, parentId: $parentId, first: $first, after: $after) {
    nodes {
      id
      postId
      parentId
      authorId
      body
      createdAt
    }
    pageInfo {
      endCursor
      hasNextPage
    }
  }
}
```

- `parentId: null` - корневые комментарии поста;
- `parentId: <id>` - непосредственные ответы на комментарий;
- каждый уровень дерева пагинируется независимо.

## Mutation

### Создание поста

```graphql
mutation CreatePost($input: CreatePostInput!) {
  createPost(input: $input) {
    id
    authorId
    title
    body
    commentsEnabled
    createdAt
  }
}
```

```json
{
  "input": {
    "authorId": "00000000-0000-0000-0000-000000000001",
    "title": "заголовок",
    "body": "текст",
    "commentsEnabled": true
  }
}
```

### Изменение возможности комментирования

```graphql
mutation SetPostCommentsEnabled($input: SetPostCommentsEnabledInput!) {
  setPostCommentsEnabled(input: $input) {
    id
    authorId
    commentsEnabled
    updatedAt
  }
}
```

```json
{
  "input": {
    "postId": "10000000-0000-0000-0000-000000000001",
    "authorId": "00000000-0000-0000-0000-000000000001",
    "enabled": false
  }
}
```

Если переданный `authorId` не совпадает с владельцем поста, возвращается `FORBIDDEN`.

### Создание корневого комментария

```graphql
mutation CreateComment($input: CreateCommentInput!) {
  createComment(input: $input) {
    id
    postId
    parentId
    authorId
    body
    createdAt
  }
}
```

```json
{
  "input": {
    "postId": "10000000-0000-0000-0000-000000000001",
    "authorId": "00000000-0000-0000-0000-000000000002",
    "body": "Новый комментарий"
  }
}
```

## Subscription

```graphql
subscription CommentCreated($postId: ID!) {
  commentCreated(postId: $postId) {
    id
    postId
    parentId
    authorId
    body
    createdAt
    author {
      id
      name
    }
  }
}
```

События передаются только после успешного сохранения комментария. Подписчик получает только сообщения, опубликованные во время активного соединения.

При потере соединения клиент должен повторить обычный query, чтобы синхронизировать состояние.

## Пагинация

Посты, корневые комментарии и ответы сортируются от новых к старым:

```text
order by created_at desc, id desc
```

Курсор для пагинации содержит `createdAt` и `id`, сериализованные в JSON и закодированные через base64.

Connections содержат `nodes`, курсор следующей страницы находится в `pageInfo.endCursor`.

## Ошибки

Публичная ошибка содержит безопасное сообщение и код в `extensions.code`. Для ошибки конкретного аргумента также возвращается `extensions.field`.

```json
{
  "errors": [
    {
      "message": "comments are disabled for this post",
      "extensions": {
        "code": "COMMENTS_DISABLED"
      }
    }
  ]
}
```

Существующие коды ошибок:

```text
INVALID_ARGUMENT
INVALID_CURSOR
INVALID_PAGE_SIZE
FORBIDDEN
POST_NOT_FOUND
USER_NOT_FOUND
COMMENT_NOT_FOUND
PARENT_COMMENT_NOT_FOUND
COMMENTS_DISABLED
POST_TITLE_EMPTY
POST_BODY_EMPTY
COMMENT_EMPTY
COMMENT_TOO_LONG
COMMENT_SELF_PARENT
REQUEST_CANCELLED
DEADLINE_EXCEEDED
INTERNAL
```
