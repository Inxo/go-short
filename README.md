# URL Shortener

URL Shortener is a simple web service that allows you to shorten long URLs into shorter and more convenient ones.

## Description

This service is written in Go programming language and uses SQLite database to store shortened URL addresses.

## Installation

To install and run the service, you will need Docker installed. Follow these steps:

1. Clone the repository:

   ```bash
   git clone https://github.com/your-username/url-shortener.git
   ```

2. Navigate to the project directory:
   ```bash
    cd url-shortener
    ```

3. Build the Docker image:

```bash
docker build -t my-url-shortener .
```

4. Run the Docker container:

```
docker run -p 8080:8080 -e BASE_URL=https://s.inxo.ru -v $(pwd)/database.db:/app/database.db my-url-shortener
```

Now your URL Shortener is up and running and accessible at `http://localhost:8080`.

## Usage

To shorten a long URL, send a POST request to `/shorten` with a `url` parameter containing your long URL. For example:

```bash
curl -X POST -d "url=https://example.com/very/long/url" http://localhost:8080/shorten
```

The service will return the shortened URL in the response.

To use the shortened URL, simply navigate to it in your browser or make a GET request to it.

## Additional Information

To configure the base URL, modify the `BASE_URL` environment variable in the Dockerfile.

## Author

Author: Dima Lario