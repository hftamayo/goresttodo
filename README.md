## Version History ##

### 0.0.2 ###
- Fiber as ORM framework
- Postgres as datalayer
- Improving function such as: data seeding and clean architecture

### 0.0.1 ###
- On memory data: JSON
- Architecture: Domain-Driven Design pattern and hexagonal architecture principles

In DDD, the codebase is structured around the business domain. The main components of DDD are:

- Entities (e.g., models): These are the business objects of the application.
- Repositories (e.g., repository.go, repoimpl.go): These handle the storage and retrieval of entities.
- Services (e.g., service.go): These implement business logic that doesn't naturally fit within entities.
- Interfaces (e.g., handler.go): These define how the outside world interacts with the domain.

Hexagonal Architecture principles, where the business logic (the "domain") is at the center of the design and the infrastructure and interface details are at the outer layers. This allows the domain logic to be independent of any external concerns.


### Useful commands ###

1. Initialize Go App: go mod init github.com/hftamayo/gotodo
2. Install fiber v2: go get -u github.com/gofiber/fiber/v2
3. Create frontend with yarn: yarn create vite client -- --template react-ts
4. Install dependencies: yarn add @mantine/hooks @mantine/core swr @primer/octicons-react
5. Run the project: go build main.go | go run main.go

### References ###
[Original tutorial](https://youtu.be/QevhhM_QfbM?si=u-9zVUAnJWBmWJJY)
