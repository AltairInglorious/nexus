# Nexus

Nexus is a Golang utility package develop by TechLead of Adameus Technologies that helps to map your application models to the messaging system and integrate with a SurrealDB database. This package includes helper functions for handling NATS message brokering, error handling, efficient object pool management, and database interaction.

## NATS + SurrealDB

The adoption of NATS as a message broker for microservices and SurrealDB as a database brings substantial advantages, paramount among which are superior performance and a robust solution for Service Discovery.

NATS stands out in the realm of message brokers due to its lightweight nature and blazingly fast performance. By offering high-speed, low-latency communication between services, NATS ensures your system remains responsive and efficient under high load, outperforming many alternative solutions. Moreover, NATS' publish-subscribe and request-reply models are intuitive and straightforward, simplifying inter-service communication.

Additionally, NATS efficiently addresses the Service Discovery problem often faced in microservice architectures. Since services need to discover and communicate with each other in a dynamic, distributed environment, NATS' intelligent service discovery mechanisms enable services to locate each other seamlessly. Thus, services can dynamically scale, deploy, and recover, thereby increasing the system's resilience and flexibility.

On the other hand, SurrealDB, a multi-model database, brings a unique blend of flexibility and power to your data layer. Unlike traditional databases, SurrealDB allows you to easily define complex data models thanks to its multi-model nature. Whether you're working with document, graph, key-value, or other data structures, SurrealDB lets you handle them all in one database, reducing the complexity of dealing with multiple disparate databases.

Moreover, SurrealDB provides efficient querying, reliable consistency, and horizontal scalability, making it an excellent choice for handling complex, large-scale data requirements. Its user-friendly nature further simplifies database management, allowing developers to focus more on the application logic and less on managing the data layer.

## Features

- Offers generic `GeneralCreate` and `GeneralSelect` functions to create and select objects from a SurrealDB database respectively.
- Provides a generic `UseFilter` function to generate dynamic SQL queries based on input struct fields and conditions.
- Offers a reusable `MapperHandler` function to handle incoming NATS messages and map them to database actions.
- Uses object pooling to efficiently handle and recycle NATS messages for lower GC pressure and higher performance.
- Provides the `SelectQuery` struct for flexible and dynamic query generation.

The `SelectQuery` struct is a convenient helper for generating SQL queries. It includes:

- `TableName` - The name of the database table.
- `Fields` - The list of fields to select. If empty, it selects all fields (\*).
- `Filter` - An optional filter for query conditions.

Here are few methods that are attached to the `SelectQuery` struct:

- `String()` - Converts the `SelectQuery` struct to an actual SQL string.
- `NewSelectAll(string)` - Returns a `SelectQuery` struct for selecting all fields from the provided table.
- `NewSelect(string, ...string)` - Returns a `SelectQuery` struct for selecting specific fields from the provided table.
- `WithFilter(any)` - Returns a `SelectQuery` struct with the provided filter.

## Documentation

Detailed in-line comments have been provided for each function in the codebase, which can be generated into docs using a tool like godoc. These docs explain the purpose of each function, the parameters they take, and what they return.

## Usage

### Basic Usage

```go
// Create a new NATS transport
transport, err := New(natsUrl, nkeyFile, name)

// Handle a NATS event
transport.Handle(event, MapperHandler(dbFn))

// Use filters to modify an SQL query
query := UseFilter(filter, "SELECT * FROM table")

// Create a new entity in the SurrealDB
entity, err := GeneralCreate(&DB, thing, data)

// Select an entity from the SurrealDB
entities, err := GeneralSelect(&DB, selectQuery)

// Update an existing entity in the SurrealDB
updatedEntity, err := GeneralUpdate(&DB, thing, id, updateData)

// Change an existing entity in the SurrealDB
changedEntity, err := GeneralChange(&DB, thing, id, changeData)

// Delete an entity from the SurrealDB
deletedEntity, err := GeneralDelete(&DB, thing, id)
```

### SelectQuery Usage

To use `SelectQuery`, create a new query using `NewSelectAll` or `NewSelect`. You can then optionally add a filter using `WithFilter`.

```go
// Create a select all query for the 'users' table
q := NewSelectAll("users")

// Create a select query for 'id' and 'name' fields in the 'users' table
q := NewSelect("users", "id", "name")

// Add a filter to the query
q = q.WithFilter(UserFilter{IsActive: true})
```

Use the `String` method to convert the `SelectQuery` struct to an actual SQL string:

```go
sql := q.String()
```

## Error Handling

This package uses a sync pool to manage NATSError and NATSOk structs, minimizing allocations and GC pressure.

## Installation

Use the below command to add this package to your Go project:

`go get github.com/yourusername/go-nats-surrealdb-helper`

## Contribution

Please feel free to create issues for bugs and desired features. Pull requests are welcome!

## License

This project is licensed under the MIT License.
