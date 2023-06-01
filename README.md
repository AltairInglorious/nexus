# Nexus

Nexus is a Golang utility package develop by Adameus Technologies that helps to map your application models to the messaging system and integrate with a SurrealDB database. This package includes helper functions for handling NATS message brokering, error handling, efficient object pool management, and database interaction.

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

// Create a new entity in the SurrealDB
entity, err := GeneralCreate(&DB, thing, data)

// Select an entity from the SurrealDB
entities, err := GeneralSelect(&DB, selectQuery)

// Use filters to modify an SQL query
query := UseFilter(filter, "SELECT * FROM table")
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
