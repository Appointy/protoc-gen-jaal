# protoc-gen-jaal - Develop relay compliant GraphQL servers

protoc-gen-jaal is a protoc plugin which is used to generate jaal APIs. The server built from these APIs is graphQL spec compliant as well as relay compliant. It also handles oneOf by registering it as a Union on the schema.

## Getting Started

Let's get started with a quick example.

```protobuf
syntax = "proto3";

package customer;

option go_package = "customerpb";

import "schema/schema.proto";

service Customers {
    // CreateCustomer creates new customer.
    rpc CreateCustomer (CreateCustomerRequest) returns (Customer) {
        option (graphql.schema) = {
            mutation : "createCustomer"
        };
    };

    // GetCustomer returns the customer by its unique user id.
    rpc GetCustomer (GetCustomerRequest) returns (Customer) {
        option (graphql.schema) = {
            query : "customer"
        };
    };
}

message CreateCustomerRequest {
    string email = 1;
    string first_name = 2;
    string last_name = 3;
}

message GetCustomerRequest {
    string id = 1;
}

message Customer {
    string id = 1;
    string email = 2;
    string first_name = 3;
    string last_name = 4;
}
```

protoc-gen-jaal uses the method option schema to determine whether to register the rpc as query or as mutation. If an rpc is not tagged, then it will not be registered on the schema. To generate relay compliant servers, protoc-gen-jaal generates the input and payload of each mutation with clientMutationId. The graphql schema of above example is as follows:

```GraphQl Schema
input CreateCustomerInput {
    clientMutationId: String
    email: String
    firstName: String
    lastName: String
}

type CreateCustomerPayload {
    clientMutationId: String!
    payload: Customer
}

type Customer {
    email: String!
    firstName: String!
    id: ID!
    lastName: String!
}

type Mutation {
    createCustomer(input: CreateCustomerInput): CreateCustomerPayload
}

type Query {
    customer(id: ID): Customer
}
```

### Installing

The installation of protoc-gen-jaal gen be done directly by running go get.

```
go get github.com/appointy/protoc-gen-jaal
```

### Usage

For a proto file customer.proto, the corresponding code is generated in customer.gq.go.

```
protoc \
  -I . \
  -I ${GOPATH}/src \
  -I ${GOPATH}/src/github.com/appointy/protoc-gen-jaal \
  --go_out=grpc=plugins:. \
  --jaal_out:. \
  customer.proto && goimports -w .
```

protoc-gen-jaal generates the code to register each message as input and payload. The payload is registered with the name of message. The input is registered with the name of message suffixed with "Input". protoc-gen-jaal implicitly registers field named id as graphQL ID.

## Available Options

The behaviour of protoc-gen-jaal can be modified using the following options:

### File Option

* file_skip : This option is used to skip the generation of gq file.

### Method Option

* schema : This option is used to tag an rpc as query or mutation.

### Message Options

* skip : This option is used to skip the registration of a message on the graphql schema.

* name : This option is used to change default name of message on graphql schema.

* type : This option is used to change go type of the message in the gq file.

### Field Options

* input_skip : This option is used to skip the registration of the field on input object.

* payload_skip : This option is used to skip the registration of the field on payload object.

* id : This option is used to expose the field as graphQL ID. Only string field can be tagged with this option.
