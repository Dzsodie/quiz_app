# Quiz App

![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)
![Made with Go](https://img.shields.io/badge/Made%20with-Go-00ADD8?style=for-the-badge&logo=go)
![Swagger](https://img.shields.io/badge/Swagger-API%20Docs-green?style=for-the-badge&logo=swagger)
![Gorilla Toolkit](https://img.shields.io/badge/Gorilla-Toolkit-blue?style=for-the-badge&logo=go)
![Cobra CLI](https://img.shields.io/badge/Cobra-CLI-purple?style=for-the-badge&logo=go)

A Go-based quiz application that supports user registration, login, and quiz functionalities. This app is built with the Gorilla web toolkit and includes Swagger documentation for easy API interaction.

## Features
- User registration and authentication
- CSV-based question loading
- Quiz functionality with score tracking
- Swagger documentation for API testing

## Installation

1. Clone the repository:
   ```bash
   git clone https://github.com/your-username/quiz_app.git

2. Navigate to the project directory:
    ```bash
    cd quiz_app

3. Install dependencies:
    ```bash
    go mod tidy

4. Run the application:
    ```bash
    go run main.go

## API Documentation
Swagger documentation is available at: http://localhost:8080/swagger/index.html

## Usage
- Register a user via the /register endpoint.
- Log in using the /login endpoint.
- Start a quiz with /quiz/start.
- Submit answers to questions using /quiz/submit.
- View results at /quiz/results.

## License
This project is licensed under the MIT License. See the [LICENSE](LICENSE) file for details.

## Gopher  🐹
"A fun way to learn Go and improve your coding skills!"

---
