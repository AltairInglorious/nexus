# Nexus

Nexus is a Golang utility package develop by Adameus Technologies that helps to map your application models to the messaging system and integrate with a SurrealDB database. This package includes helper functions for handling NATS message brokering, error handling, efficient object pool management, and database interaction.

## Features

Create, Read operations on the Database: GeneralCreate and GeneralSelect helper functions provide abstraction over create and select operations with SurrealDB.

1. **Message Mapping**: MapperHandler function provides a way to map DB models to a message broker.

1. **NATS Subscription Handler**: Handle function provides a way to subscribe to a NATS event with error handling and response management.

1. **NATS Connection**: New function creates a new NATS connection with a given URL, name, and NKEY seed file.

1. **Dynamic Query Building**: UseFilter function helps to dynamically build SQL queries based on the provided filter struct.

## Documentation

Detailed in-line comments have been provided for each function in the codebase, which can be generated into docs using a tool like godoc. These docs explain the purpose of each function, the parameters they take, and what they return.

## Usage

```
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

## Error Handling

This package uses a sync pool to manage NATSError and NATSOK structs, minimizing allocations and GC pressure.

## Installation

Use the below command to add this package to your Go project:

`go get github.com/yourusername/go-nats-surrealdb-helper`

## Contribution

Please feel free to create issues for bugs and desired features. Pull requests are welcome!

## License

This project is licensed under the MIT License.
