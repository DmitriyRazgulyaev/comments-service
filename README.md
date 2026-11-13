Создание поста с разрешенными комментариями
```go
mutation {
    createPost(input: {
        authorId: "1"
        title: "Заголовок"
        content: "содержание поста"
        commentsAllowed: true
    }) {
        id
        title
        createdAt
    }
}
```
Получение одного поста
```go
    query {
        post(id: "1") {
            id
            authorId
            title
            content
            commentsAllowed
            createdAt
        }
    }
```
Получение списка постов, максимум десять
```go
query {
    posts(first: 10) {
        nodes {
            id
            authorId
            title
            content
            commentsAllowed
            createdAt
        }

        pageInfo {
            hasNextPage
            endCursor
        }
    }
}
```
Получение следующей страницы постов (after указан для примера)
```go
query {
    posts(first: 10, after: "NQ") {
        nodes {
            id
            title
            content
        }

        pageInfo {
            hasNextPage
            endCursor
        }
    }
}
```
Получение поста с комментариями (лимит 10)
```go
query {
    post(id: "1") {
        id
        title
        content

        comments(first: 10) {
            nodes {
                id
                postId
                parentCommentId
                authorId
                content
                createdAt
            }

            pageInfo {
                hasNextPage
                endCursor
            }
        }
    }
}
```
Получение корневых комментариев под постом 1 (лимит 10)
```go
query {
        comments(
            postId: "1"
            first: 10
        ) {
        nodes {
			id
			postId
            parentCommentId
            authorId
            content
            createdAt
        }

        pageInfo {
            hasNextPage
            endCursor
        }
    }
}
```
Получение первых десяти ответов под комментарием 5 под постом 1

```go
 query {
        comments(
            postId: "1"
            parentCommentId: "5"
            first: 10
        ) {
            nodes {
                id
                content
                parentCommentId
            }
            pageInfo {
                hasNextPage
                endCursor
            }
        }
    }
```
Получение следующей страницы ответов (after указан для примера)
```go
 query {
        comments(
            postId: "1"
            parentCommentId: "1"
            first: 3
            after:"eyJDcmVhdGVkQXQiOiIyMDI2LTExLTEzVDExOjAxOjU3LjExNDQxOFoiLCJJRCI6OX0"
        ) {
            nodes {
                id
                content
                parentCommentId
            }
            pageInfo {
                hasNextPage
                endCursor
            }
        }
    }
```
Создание корневого комментария 
```go
mutation {
    createComment(input: {
        postId: "1"
        authorId: "2"
        content: "Содержание комментария"
    }) {
        id
        postId
        parentCommentId
        authorId
        content
        createdAt
    }
}
```
Создание ответа на комментарий 
```go
mutation {
    createComment(input: {
        postId: "1"
        parentCommentId: "5"
        authorId: "3"
        content: "Содержание комментария"
    }) {
        id
        postId
        parentCommentId
        authorId
        content
        createdAt
    }
}
```
Отключение комментариев
```go
mutation {
    setCommentsAllowed(
        postId: "1"
        authorId: "1"
        allowed: false
    )
}
```
Включение комментариев
```go
mutation {
    setCommentsAllowed(
        postId: "1"
        authorId: "1"
        allowed: true
    )
}
```
Подписка на пост для получения новых комментариев 
```go
subscription {
    commentAdded(postId: "1") {
        id
        postId
        parentCommentId
        authorId
        content
        createdAt
    }
}
```