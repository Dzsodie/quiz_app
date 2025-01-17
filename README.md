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
- Cobra CLI for UI.
- Gorilla Toolkit for session handling.
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

## Usage of CLI commands outside game mode

- quiz : Initiate the quiz application.
- start : Start the quiz.
- score : View user score and stats.
- exit : Quit the quiz app.

## Usage through REST API

- Import the Postman collection from the postman/quiz_app.postman_collection.json.
- Register a user via the `/register` endpoint.
    Add `Content-Type : application/json` to the headers if it is missing.
    
    Example username and password payload.
    ```bash
    {
    "username": "testuser",
    "password": "Valid@123"
    }
    ```
- Log in using the `/login` endpoint. Add `Content-Type : application/json` to the headers if it is missing. The same username and password should be added to the basic authentication.
- Start a quiz with `/quiz/start`. The same username and password should be added to the basic authentication.
- Get next question on `/quiz/next`. The same username and password should be added to the basic authentication.
- Submit answers to questions using `/quiz/submit`.  Add `Content-Type : application/json` to the headers if it is missing. The same username and password should be added to the basic authentication.
    Example payload for answer.
    ```bash
    {
    "question_index": 0,
     "answer": 2
    }
    ```
- View results at `/quiz/results`. The same username and password should be added to the basic authentication.
- Get statistics at `/quiz/stats`. The same username and password should be added to the basic authentication.

## API Documentation

Swagger documentation is available at: http://localhost:8080/swagger/index.html

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
    In the `utils/logger.go` the log format is configured and the logs are collected to the `logs/app.log` file.

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

- Enhance the game play with randomized questions with keeping track of already chosen questions.
- Adding database connection for persistency.
- Containerize with docker for portability.
- Deploying to cloud for easier distribution, monitoring and low cost 7/24 99.9% availability.

## License

This project is licensed under the MIT License. See the [LICENSE](LICENSE) file for details.

## Gopher  🐹

"A fun way to learn Go and improve your coding skills!"

![Golang Playground Guide](https://withcodeexample.com/wp-content/uploads/2025/01/golang-playground-guide-image.jpg)

---
