# Nakama RPC Function Implementation

This project implements a Nakama RPC function using Go, fulfilling the requirements specified by the task description. The function accepts a payload with type, version, and hash parameters, reads a corresponding file from the disk, saves data to a database, calculates the file content hash, and responds with type, version, hash, and content.

## Prerequisites

- Docker
- Docker Compose

## Setup

1. **Clone the repository:**

   ```bash
   git clone https://github.com/davittatenkohayrapetyan/nakama-rpc
   cd ./nakama-rpc
   ```

2. **Build the Docker image:**

   ```bash
   docker-compose build
   ```

3. **Run the application:**

   ```bash
   docker-compose up
   ```

## Files Description

### Dockerfile

This file defines the Docker image for running the Go application.

### Dockerfile.tests

This file defines the Docker image for running tests on the Go application.

### docker-compose.yml

This file sets up the Docker Compose environment to run the application and its dependencies, including Nakama, PostgreSQL, and PGAdmin.

### main.go

This is the main file containing the RPC function implementation. The function reads the specified file, saves data to the database, calculates the hash, and returns the appropriate response.

### main_e2e_test.go

This file contains end-to-end tests for the implemented RPC function, ensuring the correctness and robustness of the functionality.

## Running Tests

Tests are run automatically when the `<test>` service is started with Docker Compose. To run the tests, simply execute:

```bash
docker-compose up
```

## Accessing the Database using PGAdmin

1. **Run the PGAdmin Docker container** (assuming you have PGAdmin included in your `docker-compose.yml`):

   ```bash
   docker-compose up -d pgadmin
   ```

2. **Open PGAdmin in your web browser**:

   Navigate to `http://localhost:8080` (or the port specified in your `docker-compose.yml` for PGAdmin).

3. **Login to PGAdmin**:

   Use the credentials specified in your `docker-compose.yml`. Typically, you might have:
   - **Email**: `admin@admin.com`
   - **Password**: `admin`

4. **Add a New Server in PGAdmin**:

   - **Name**: Choose a name for your database server connection.
   - **Host name/address**: `postgres` (as specified in your `docker-compose.yml`).
   - **Port**: `5432` (default PostgreSQL port).
   - **Username**: `postgres` (as specified in your `docker-compose.yml`).
   - **Password**: `localdb` (as specified in your `docker-compose.yml`).

5. **Save and Connect**:

   Click "Save" to connect to your PostgreSQL database. You should now be able to see your database and manage it using PGAdmin.

## Logging into Nakama

1. **Run the Nakama Docker container** (assuming you have Nakama included in your `docker-compose.yml`):

   ```bash
   docker-compose up -d nakama
   ```

2. **Access Nakama Console**:

   Navigate to `http://localhost:7351` (or the port specified in your `docker-compose.yml` for Nakama).

3. **Login to Nakama Console**:

   Use the default credentials or those specified in your `docker-compose.yml`. Typically, you might have:
   - **Username**: `admin`
   - **Password**: `password`

4. **Manage Nakama**:

   Once logged in, you can manage your game server, view logs, configure settings, and monitor activity.

## Explanation and Thoughts

### Solution Explanation

The solution follows the task requirements closely:

1. **Payload Handling:** The RPC function accepts `type`, `version`, and `hash` parameters with default values.
2. **File Reading:** Reads the file from the disk based on the provided type and version.
3. **Database Interaction:** Saves relevant data to the database.
4. **Hash Calculation:** Calculates the file content hash and includes it in the response.
5. **Response Handling:** Constructs the response with the fields `type`, `version`, `hash`, and `content`. If hashes do not match, the content is set to null.

### Improvements

Given more time, I would consider the following improvements:

1. **Enhanced Error Handling:** Provide more granular error messages and handling mechanisms.
2. **Configuration Management:** Implement a configuration management system to handle environment-specific settings.
3. **Performance Optimization:** Optimize the file reading and hash calculation processes for large files.
4. **Scalability:** Improve the application to handle concurrent requests efficiently.
5. **Refactor Tests:** Refactor the existing tests to improve readability and maintainability.
6. **Additional Tests:** Add more tests to cover other edge cases and ensure robustness.
7. **Repository Pattern:** Use a separate repository file for accessing the database to improve code organization and separation of concerns.
8. **JSON Serialization/Deserialization:** Improve the implementation of JSON marshalling/unmarshalling for better performance and error handling.

## Thoughts on the Task

I found the task quite an interesting challenge. Although I don’t have prior experience with Nakama and Golang, I took help from AI to read through the Nakama documentation and find specific aspects required for the task implementation. While I did a quick run-through of the documentation, AI assistance was crucial in learning language features and implementing the solution in VSCode without an IDE.

Given more time to learn Golang, my coding style would certainly become more professional. Initially, I faced challenges building the code (Go plugin) locally and deploying it to the Nakama server due to my MacOS setup conflicting with Nakama's Linux-based environment, specifically with ELF headers. I eventually found a Go builder for Nakama that resolved this issue.

Since the RPC procedure was relatively simple, I decided to focus on an end-to-end testing approach to become familiar with interactions with Nakama. Consequently, the tests are built and run as a separate service in Docker Compose. The main challenges I faced, stemming from my first-time experience with Go, included JSON marshalling/unmarshalling logic and the serialization and deserialization of requests and responses.

## Conclusion

This project demonstrates the implementation of a Nakama RPC function using Go, covering all specified requirements. The provided Docker setup ensures easy deployment and testing of the application.
