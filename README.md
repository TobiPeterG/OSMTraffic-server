# Traffic Server

This project is a traffic server built using Go and Redis to provide traffic data for navigation applications, such as Organic Maps and OSMand. The server fetches traffic data from sources like Datex2 and caches it in Redis to ensure fast and efficient responses to client requests. It is designed to be scalable and efficient, using Docker for deployment.

## Features

- Fetches and caches traffic data. **MISSING**
- Supports Datex2 integration (or other traffic data APIs). **MISSING**
- Redis caching for better performance. **MISSING**
- Exposes traffic data via REST API. **MISSING**
- Scalable and containerized using Docker. **MISSING**

## Prerequisites

- **Go**: Version 1.23 or higher is recommended.
- **Redis**: A Redis instance to cache the traffic data.
- **Docker**: For containerization and running the server.

## Running the Server

You can run the traffic server locally or using Docker.

### Running Locally

1. Clone the repository:
   ```bash
   git clone https://github.com/TobiPeterG/OSMTraffic-server
   cd OSMTraffic-server
   ```
2. Install dependencies:
   ```bash
   go mod download
   ```
3. Build and run the server:
   ```bash
   go build -o traffic-server .
   ./traffic-server
   ```
4. Ensure that Redis is running locally on localhost:6379. You can start Redis with:
   ```bash
   redis-server
   ```

### Running with Docker

1. Build and run the Docker container:
   ```bash
   docker-compose up --build
   ```

The traffic server will be running on http://localhost:8080 and Redis on localhost:6379.

## Accessing the API
You can access traffic data through the following API endpoint:
- **GET /traffic**

Example:
   ```bash
   curl http://localhost:8080/traffic
   ```

## Traffic Data Sources
Currently, this server fetches traffic data from:

- **Datex2 API:** (Placeholder for the traffic data source).

You can integrate other traffic data sources by modifying the `GetTrafficFromDatex2()` function in main.go.

## Caching
The server uses Redis to cache traffic data to avoid overloading the data sources and ensure fast response times.

- The data is cached for 5 minutes by default.
- If data is unavailable in Redis (a cache miss), the server fetches the data from the source and stores it in Redis for future requests.

## Future Plans
- **Datex2 Integration:** Implement full Datex2 API integration.
- **Traffic Data Aggregation:** Aggregate traffic data from multiple sources (such as national traffic agencies).
- **User Reporting:** Add functionality to allow users to report traffic conditions.
- **Analytics and Statistics:** Track and report usage statistics for better monitoring.

## Contributing
Contributions are welcome! Please fork the repository, make changes, and submit a pull request. Ensure that your code follows best practices and includes tests where necessary.
