# 1. Overview

Restful API for managing task persistence.

---

# 2. Level of complexity

==BEGINNER==: this version is properly for developer technical skills in the use of the technical stack, it is the foundational part for building projects of more complexity.

---

# 3. Major Releases Timeline

## 0.2.0

- Released on September 2025
- Unit tested
- Spec coverage of 55%

## 0.1.5

- Special features such as rate limiting, caching using Redis

---

# 4. Architectural Diagram

---

# 5. Technical Specs

## 5.1 Architectural Pattern: Hexagonal

### Core Principles in This Project

- **Domain-Centric Design**: Business logic is isolated from external concerns
- **Ports**: Defined interfaces for interacting with the domain
- **Adapters**: Implementations that connect external systems to the domain through ports
- **Dependency Inversion**: Core domain depends on abstractions, not concrete implementations

### Architectural Layers

1. **Domain Layer** (Core Hexagon)

   - Contains business entities (Task models)
   - Domain services with core business logic
   - No dependencies on external frameworks or libraries

2. **Ports Layer** (Interfaces)

   - **Primary/Driving Ports**: Service interfaces exposing domain functionality
   - **Secondary/Driven Ports**: Repository interfaces defining storage requirements

3. **Adapters Layer**
   - **Primary/Driving Adapters**: HTTP handlers (Gin controllers) translating HTTP to domain calls
   - **Secondary/Driven Adapters**: Repository implementations (GORM), cache adapters (Redis)

### Flow of Control

External requests → Primary Adapters → Primary Ports → Domain Logic → Secondary Ports → Secondary Adapters → External systems

### Core Domain Features

- Task management (core domain logic)
- Status transitions (domain rules)
- Validation logic (domain constraints)

### Ports (Interfaces)

- `TaskServiceInterface`: Primary port for task operations
- `TaskRepositoryInterface`: Secondary port for task persistence
- `CacheInterface`: Secondary port for caching operations

### Adapters

- **HTTP Handlers**: Primary adapters converting HTTP to domain calls
- **GORM Repositories**: Secondary adapters implementing persistence ports
- **Redis Cache**: Secondary adapter implementing cache port
- **Rate Limiter**: Cross-cutting adapter for request throttling

## 5.2 Technology Stack with Hexagonal Context

- **Language**: Go 1.22.2 (Ideal for hexagonal due to interfaces and composition)
- **Web Framework**: Gin (Primary adapter for HTTP)
- **Database**: GORM (Secondary adapter for persistence)
- **Cache**: Redis (Secondary adapter for caching)
- **Rate Limiting**: Redis-based (Cross-cutting concern implemented as middleware)
- **Containerization**: Docker (Infrastructure concern, isolated from domain)

---

# 6. API Endpoints (Primary Adapters)

| Endpoint                | Method | Hexagonal Role                     | Cache Strategy                 | Rate Limit |
| ----------------------- | ------ | ---------------------------------- | ------------------------------ | ---------- |
| `/tasks/task`           | GET    | Primary adapter → TaskService port | 30s with ETag                  | 100/min    |
| `/tasks/task/list/page` | GET    | Primary adapter → TaskService port | 30s with ETag                  | 100/min    |
| `/tasks/task/:id`       | GET    | Primary adapter → TaskService port | 30s with ETag                  | 100/min    |
| `/tasks/task`           | POST   | Primary adapter → TaskService port | Invalidates list caches        | 30/min     |
| `/tasks/task/:id`       | PUT    | Primary adapter → TaskService port | Invalidates specific caches    | 30/min     |
| `/tasks/task/:id/done`  | PUT    | Primary adapter → TaskService port | Invalidates specific caches    | 30/min     |
| `/tasks/task/:id`       | DELETE | Primary adapter → TaskService port | Invalidates all related caches | 30/min     |

---

# 7. Request/Response Format (Domain Translation)

### Task DTO (Data Transfer Object)

```json
{
  "id": 1,
  "title": "Task title",
  "description": "Task description",
  "done": false,
  "owner": 1,
  "created_at": "2023-06-05T10:15:30Z",
  "updated_at": "2023-06-05T10:15:30Z"
}
```

### Success Response (Adapter Translation Layer)

```json
{
  "code": 200,
  "resultMessage": "SUCCESS",
  "data": {
    /* domain object converted to DTO */
  },
  "timestamp": 1686061234,
  "cacheTTL": 30
}
```

### Error Response (Adapter Translation Layer)

```json
{
  "code": 400,
  "resultMessage": "OPERATION_FAILED",
  "error": "Error description"
}
```

### Primary Adapters (HTTP Handlers)

- Translate HTTP requests to domain operations
- Call appropriate service port methods
- Transform domain objects to DTOs for response
- Handle HTTP-specific concerns (status codes, headers)

### Secondary Adapters (Repository Implementations)

- GORM implementation of repository interfaces
- Redis implementation of cache interfaces
- Isolate infrastructure details from domain logic

---

# 8. Special features:

## Cache Implementation in Hexagonal Context

- **Cache Port**: Defines caching operations as interfaces
- **Redis Adapter**: Implements cache port with Redis
- **Domain Event Listeners**: Trigger cache invalidation on domain events
- **Adapter-Specific Concerns**: TTL, serialization handled in adapter layer

## Rate Limiting in Hexagonal Context

- **Cross-cutting Concern**: Implemented as middleware (outside the hexagon)
- **Primary Adapter Extension**: Enhances HTTP handling without touching domain
- **Redis Adapter**: Secondary adapter for distributed rate limiting

## Testing Strategy for Hexagonal Architecture

- **Domain Tests**: Unit tests for core business logic
- **Port Tests**: Tests ensuring port contracts are fulfilled
- **Adapter Tests**: Tests for adapter implementations
- **Mock Ports**: For testing adapters in isolation
- **Integration Tests**: Test full flows through the hexagon

## Benefits of Hexagonal Architecture in This Project

1. **Testability**: Domain logic can be tested without infrastructure
2. **Maintainability**: Clear separation of concerns and dependencies
3. **Flexibility**: Ability to swap out adapters (e.g., change from Redis to another cache)
4. **Focus on Domain**: Business rules are centralized and explicit
5. **Technological Agnosticism**: Core business logic is independent of frameworks

---

# 9. Future Architectural Improvements

- **Domain Events**: Expand event-driven architecture for better decoupling
- **Anti-corruption Layer**: For integrating with external systems
- **Command Query Responsibility Segregation (CQRS)**: Separate read and write models
- **Bounded Contexts**: Define clear boundaries between different domain areas
- Testing strategy needs to be implemented
- Monitoring and observability are missing
- Documentation could be enhanced

---

# 10. Branches

1. ==Main==: it contains the latest deployed and published codebase, this one has been tested against unit, integration and end 2 end, also, there are special directories related to developer such as: devops (CI/CD pipelines), sshots (images for README file) and developer (diagrams, postman yaml files, documentation)

2. ==Stage==: target branch for test the execution of the CI/CD pipelines, includes the interaction with the CI tools and cloud providers, the use of this branch is suggested for QA and DevOps teams. Pre-release version management, this one should be the only one merged with main branch.

3. ==Unstable==: it containts the test codebase (unit, integration, end 2 end), it interacts with experimental and stage branches, must not merge with main.

4. ==Experimental==: alpha version of the codebase, all features are built here, it interacts with unstable and stage, must not be merged directly with main branch.

5. ==Refactor==: special feature requires by Experimental branch, the intention is to not affect the latest run version of the codebase contained in Experimental, if must be merged just with experimental branch.

---

# 11. Leftovers:

##### 1.edge cases:

- **Stale Cache After Service Restart**: Ensure Redis TTL values are reasonable to prevent very old data from being served
- **Cache Stampede**: When cache expires, multiple requests might try to rebuild it simultaneously - implement singleflight pattern
- **Cache Inconsistency**: If DB operations fail after cache invalidation, implement two-phase commit for critical operations
- **Cache Size Limits**: Add cache eviction policies to prevent memory issues with large datasets
- **Database Timeouts**: Add context with timeouts for all database operations
- **Partial Updates**: Ensure transactions are used for operations that modify multiple records
- **Race Conditions**: Implement optimistic locking for concurrent updates to the same record
- **Connection Pool Exhaustion**: Set reasonable pool sizes and handle connection limits
- **Malformed JSON**: Handle JSON parse errors gracefully
- **Invalid UTF-8**: Ensure proper encoding handling for international text
- **XSS Protection**: Sanitize inputs that might be displayed in UI
- **Excessively Large Inputs**: Limit request body sizes and pagination parameters
- **Token Expiration**: Gracefully handle expired tokens with clear error messages
- **Rate Limit Bypass Attempts**: Check for distributed attacks across multiple IPs
- **Permission Boundary Cases**: Verify edge cases where users might access resources they shouldn't
- **Graceful Service Degradation**: When Redis or other dependencies are down, degrade gracefully
- **Request Timeout Handling**: Add request-level timeouts to prevent hanging connections
- **Health Check Endpoints**: Skip rate limiting and authentication for health checks
- **Logging Overflow**: Protect against log flooding during error cascades
- **Deadlocks**: Ensure consistent lock acquisition order across your codebase
- **Fan-out Overload**: When spawning goroutines, use worker pools to limit concurrency
- **Context Propagation**: Ensure context cancellation propagates through all operations
- - **Handling Incomplete Downloads**: Include Content-Length headers and ETag validation
- **Browser Cache Inconsistencies**: Test cache headers across different browsers
- **Mobile Network Transitions**: Support resumable operations for mobile clients

##### 2. Observability & Monitoring

- Add structured logging for cache hits/misses
- Implement metrics for cache effectiveness and rate limit hits
- Add tracing for request flows through the system

##### 3. Error Handling Improvements

- Create consistent error response formats
- Add correlation IDs for error tracking
- Implement better validation error messages

##### 4. Performance Optimizations

- Add support for HTTP/2
- Implement database query optimization
- Consider adding background refresh for frequently accessed data

##### 5. Security Enhancements

- Add JWT validation
- Implement role-based access control
- Add request validation to prevent injection attacks

### References

[Original tutorial](https://youtu.be/QevhhM_QfbM?si=u-9zVUAnJWBmWJJY)
