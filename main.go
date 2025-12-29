package main

import (
	_ "github.com/Bengkelin/bengkelin-service/docs"
	"github.com/Bengkelin/bengkelin-service/internal/api"
)

// @title Bengkelin API
// @version 1.0
// @description Bengkelin service API for connecting users with automotive repair shops (bengkels)
// @description
// @description This API provides endpoints for:
// @description - User authentication and management
// @description - Bengkel (automotive repair shop) management
// @description - Order and service management
// @description - Chat functionality
// @description - Health monitoring and metrics
// @description
// @description ## Authentication
// @description This API uses JWT (JSON Web Tokens) for authentication. Include the token in the Authorization header:
// @description `Authorization: Bearer <your-jwt-token>`
// @description
// @description ## Rate Limiting
// @description API endpoints are rate limited:
// @description - General endpoints: 100 requests per minute
// @description - Authentication endpoints: 10 requests per minute  
// @description - Login/Register endpoints: 5 requests per minute
// @description
// @description ## Error Handling
// @description The API returns structured error responses with appropriate HTTP status codes.
// @description All responses follow a consistent format with `status`, `message`, and optional `data` fields.

// @contact.name Bengkelin API Support
// @contact.url https://bengkelin.com/support
// @contact.email support@bengkelin.com

// @license.name MIT
// @license.url https://opensource.org/licenses/MIT

// @host localhost:8080
// @BasePath /api/v1

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Type "Bearer" followed by a space and JWT token.

func main() {
	api.Run("")
}