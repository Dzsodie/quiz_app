# Quiz App

![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)
![Made with Go](https://img.shields.io/badge/Made%20with-Go-00ADD8?style=for-the-badge&logo=go)
![Swagger](https://img.shields.io/badge/Swagger-API%20Docs-green?style=for-the-badge&logo=swagger)
![Gorilla Toolkit](https://img.shields.io/badge/Gorilla-Toolkit-blue?style=for-the-badge&logo=go)
![Cobra CLI](https://img.shields.io/badge/Cobra-CLI-purple?style=for-the-badge&logo=go)
![Testify](https://img.shields.io/badge/Testify-Go%20Testing-orange?style=for-the-badge&logo=go)
![Zap Logging](https://img.shields.io/badge/Zap-Logging-blueviolet?style=for-the-badge&logo=go)
![GoAccess](https://img.shields.io/badge/GoAccess-Log%20Analyzer-darkgreen?style=for-the-badge&logo=go)


A Go-based quiz application that supports user registration, login, and quiz functionalities. This app is built with the Gorilla web toolkit and includes Swagger for easy API interaction and testing.

## Features

- User registration and authentication.
- CSV-based question loading.
- Quiz functionality with score tracking and statistics.
- Data persistence with in-Memory database with abstarction layer.
- Cobra CLI for user interaction.
- Gorilla/mux for URL router and dispatcher.
- Gorilla/sessions for managing session data across HTTP requests using cookies.
- Swagger for API testing and documentation.
- Postman collection for API testing.
- Testify for unit tests.
- Zap logging for logs.
- GoAccess for monitoring.

## Installation

1. First of all open your web browser and go to the official [Go downloads](https://go.dev/dl/) page.

2. Download the installer for your operating system. Follow the instructions.

3. Add Go to your system path in your `~/.profile`, `~/.bashrc` or `~/.zshrc` file.
    ```bash
    export PATH=$PATH:/usr/local/go/bin
    ```
4. Apply changes.
    ```bash
    source ~/.profile
    ```
5. Verify installation.
    ```bash
    go version
    ```
6. Clone the repository:
   ```bash
   git clone https://github.com/your-username/quiz_app.git
   ```
7. Navigate to the project directory:
    ```bash
    cd quiz_app
    ```
8. Install dependencies:
    ```bash
    go mod tidy
    ```
9. Run the application for using it with Postman only:
    ```bash
    go run main.go
    ```
10. Run the application in CLI game mode:
    ```bash
    go run main.go --cli
    ```

## Installation and Testing Locally with Docker

### Prerequisites
1. Before you begin, ensure you have the following installed:
    ![Docker])https://www.docker.com/products/docker-desktop/)
    Make sure to choose the version for your operating system from the Download dropdown menu on the site.

### Steps to Run the Application Locally
2. Clone this repository to your local machine:
    ```bash
    git clone https://github.com/Dzsodie/quiz_app.git
    cd quiz_app
    ```
3. Build the Docker image. Use the provided Dockerfile to build the image:
    ```bash
    docker build -t quiz_app .
    ```
4. Run the Docker container after the image is built, run the application in a container:
    ```bash
    docker run --rm -it quiz_app
    ```
5. Test the application interact with the CLI quiz application as it runs in the container.

### Additional Notes
6. Rebuilding the image: If you make changes to the source code, rebuild the Docker image to include the updates:
    ```bash
    docker build -t quiz_app .
    ```
7. Stopping the application: since the container is run interactively (-it), you can stop it by pressing Ctrl+C.

8. Debugging: To check logs or troubleshoot, you can run the container in detached mode and inspect it later:
    ```bash
    docker run -d --name quiz_app_container quiz_app
    docker logs quiz_app_container
    docker stop quiz_app_container
    ```

## Usage of CLI commands 

- start : Start the quiz.
- score : View user score and stats.
- exit : Quit the quiz app.

## Usage through REST API

1. Download ![Postman](https://www.postman.com/downloads/)
2. Import the Postman collection from the `quiz_app/postman_collection.json`
3. Register a user via the `/register` endpoint.
    Add `Content-Type : application/json` to the headers if it is missing.
    
    Example username and password payload.
    ```bash
    {
    "username": "testuser",
    "password": "Valid@123"
    }
    ```
4. Log in using the `/login` endpoint. Add `Content-Type : application/json` to the headers if it is missing. The same username and password should be added to the basic authentication.
5. Start a quiz with `/quiz/start`. The same username and password should be added to the basic authentication.
6. Get next question on `/quiz/next`. The same username and password should be added to the basic authentication.
7. Submit answers to questions using `/quiz/submit`.  Add `Content-Type : application/json` to the headers if it is missing. The same username and password should be added to the basic authentication.
    Example payload for answer.
    ```bash
    {
    "question_index": 0,
     "answer": 2
    }
    ```
8. Repeat steps 6 and 7 until you get the status Code `409 Gone` from the `/quiz/next` endpoint.
9. View results at `/quiz/results`. The same username and password should be added to the basic authentication.
10. Get statistics at `/quiz/stats`. The same username and password should be added to the basic authentication.
11. Check app health at `/health`. No authentication needed. Response should be similar to the following.
    ```bash
    {
    "in_memory_db": "OK",
    "sessions": "OK",
    "mutex": "Unlocked"
    }
    ```
+1. Cheat code: get the list of all the questions, answer options and correct answers loaded from the `/questions` endpoint.

## API Documentation

Swagger documentation is available at: http://localhost:8080/swagger/index.html after the application started successfully.

## Testing 

1. Run all of the unit tests with this command
    ```bash
    go build
    go test ./...
    ```
2. Run unit tests for specific folder
    ```bash
    go build
    go test ./internal/services
    ```
3. Test coverage can be seen with the following command
    ```bash
    go test -cover ./...

## Logging

Zap logging is used in the quiz app. 
- In the `utils/logger.go` the log format is configured.
- Log structure is defined in connection to GoAccess aggregation.
`[timestamp] [level] [logger] [caller] [message] [stacktrace/context]`

Here is a few lines of example log.
```bash
20/Jan/2025:02:56:40 +0100	DEBUG	middleware/auth_middleware.go:38	Session token validation{session_token_in_cookie 15 0 67aa6263b8f910dfeae5249bf14bf8b42cd4f5db4cf5ac2713694ac184e03014 <nil>} {stored_username 15 0 test1 <nil>} {exists 4 1  <nil>}

20/Jan/2025:02:56:40 +0100	INFO	services/quiz_service.go:40	Questions retrieved successfully{count 11 10  <nil>}

20/Jan/2025:02:56:40 +0100	WARN	utils/validator.go:19	Validation failed: question index out of range{questionIndex 11 10  <nil>}
```
- The logs are collected to the `logs/app.log` file. It makes sense to clear them from time to time as log rotation or clearing the logs when opening the app is not implemented yet. 

## Monitoring

1. For GoAccess monitoring get GoAccess.
    - Windows:
        - Download the [GoAccess](https://goaccess.io/) executable from the official website.
        - Follow the installation instructions for your Windows setup.

   - MacOS: 
    ```bash
    brew install goaccess
    ```
2. Verify installation.
    ```bash
    goaccess --version
    ```
3. Configure the log format according to your needs in the app and then in the goaccess.conf file at `/usr/local/Cellar/goaccess/1.9.3/etc/goaccess`.
    ```bash
    time-format %H:%M:%S
    date-format %d/%b/%Y
    log-format %d:%t %^ %^ %^ %m %^ %^ %U %^ %^ %^ {"username": "%u"}
    ```
4. Run the following command.
    ```bash
    goaccess quiz_app/logs/app.log -o report.html
    ```
5. Open the `report.html` in your favorite browser.
6. For real time monitoring use this command and open the report as mentioned before.

    ```bash
    tail -f quiz_app/logs/app.log | goaccess -o report.html --log-format=COMBINED
    ```

## Future enhancement

- Add logger rotation.
- Add encryption for data at rest and transition for sensitive info and personal data.
- Enhance the game play with randomized questions while keeping track of already chosen questions.
- Add database connection for persistency.
- Create an aesthetically pleasing frontend with UX/UI enhanced.
- Deploy to cloud or Heroku for easier distribution, monitoring and low cost 7/24 99.9% availability.

## License

This project is licensed under the MIT License. See the [LICENSE](LICENSE) file for details.

## Gopher  🐹

"A fun way to learn Go and improve your coding skills!"

![Golang Playground Guide](https://withcodeexample.com/wp-content/uploads/2025/01/golang-playground-guide-image.jpg)

---
