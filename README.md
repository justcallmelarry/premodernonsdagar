## Premodern Onsdagar

Premodern Onsdagar is a project for managing and displaying data related to Premodern Magic: The Gathering events. It includes features for tracking players, events, leaderboards, and decklists.

### Running the Project Locally

To run the project locally, follow these steps:

1. Clone the repository:
   ```bash
   git clone <repository-url>
   ```

2. Navigate to the project directory:
   ```bash
   cd premodernonsdagar
   ```

3. Run the service using Go:
   ```bash
   go run cmd/main/main.go
   ```

4. Open your browser and navigate to `http://localhost:8080` to view the application.

### Running the project in Production

For a production environment, you can use Docker Compose to manage the service. Here's an example `docker-compose.yml` configuration:

```yaml
version: '3.8'

services:
  premodernonsdagar:
    build:
      context: .
      dockerfile: Dockerfile
    container_name: premodernonsdagar
    build:
      context: .
      dockerfile: Dockerfile
    volumes:
      - ./files:/app/files
    ports:
      - "8080:8080"
```

Note: The volume is not strictly necessary, since it re-calculates most of the data on service start. But this might change at some point in the future, so it's safer to keep the data stored non-ephemerally.

To run the service in production:

1. Ensure Docker Compose is installed on your system.
2. Run the following command:
   ```bash
   docker-compose up --build -d
   ```
3. The service will be available at `http://localhost:8080`.

### Keeping Tailwind up to date
* Run `npm run tailwind` in a separate terminal while working with the frontend stuff (will be running continously).
* Run the service with `DEVENV=1 go run cmd/main/main.go` rather than without the env var.
* Templates are generated in `pages/html/`, but you can ignore those, they are just there to provide the css classes that are only specified in the Go part of the code.
* `static/tw.css` is being continously updated by tailwind, ensuring that everything looks as expected.

### Formatting Templates

To format the templates, use the following npm command:

1. Install the required npm packages:
   ```bash
   npm install
   ```

2. Run the formatting command:
   ```bash
   npm run format
   ```

### Tech Stack

Premodern Onsdagar is a fully Go-based service designed to manage and display data for Premodern Magic: The Gathering events. The backend is written entirely in Go, and the frontend uses Go templates for rendering dynamic content.

For frontend development, the project incorporates the following tools:

- **[Tailwind CSS](https://tailwindcss.com/)**: For utility-first styling and responsive design.
- **[Material Icons](https://fonts.google.com/icons)**: For consistent and visually appealing icons.
- **[Prettier](https://prettier.io/)**: For maintaining code formatting standards and ensuring clean, readable templates.
