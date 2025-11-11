# API Endpoints Documentation

This document lists all API endpoints required to replace mock data in the front-end application.

## Base URL
```
/api/v1
```

## Authentication Endpoints

### 1. POST /auth/login
**Description:** Authenticate user and return user data with token

**Request Body:**
```json
{
  "email": "string",
  "password": "string"
}
```

**Response (200 OK):**
```json
{
  "user": {
    "id": "string",
    "email": "string",
    "name": "string",
    "phone": "string",
    "role": "admin" | "user",
    "createdAt": "string (ISO 8601)"
  },
  "token": "string (JWT or session token)"
}
```

**Error Response (401 Unauthorized):**
```json
{
  "error": "Invalid email or password"
}
```

---

### 2. POST /auth/register
**Description:** Register a new user

**Request Body:**
```json
{
  "email": "string",
  "password": "string",
  "name": "string",
  "phone": "string"
}
```

**Response (201 Created):**
```json
{
  "user": {
    "id": "string",
    "email": "string",
    "name": "string",
    "phone": "string",
    "role": "user",
    "createdAt": "string (ISO 8601)"
  },
  "token": "string (JWT or session token)"
}
```

**Error Response (400 Bad Request):**
```json
{
  "error": "Validation error",
  "details": {
    "email": "Email already exists",
    "password": "Password must be at least 6 characters"
  }
}
```

---

### 3. GET /auth/me
**Description:** Get current authenticated user

**Headers:**
```
Authorization: Bearer <token>
```

**Response (200 OK):**
```json
{
  "id": "string",
  "email": "string",
  "name": "string",
  "phone": "string",
  "role": "admin" | "user",
  "createdAt": "string (ISO 8601)"
}
```

**Error Response (401 Unauthorized):**
```json
{
  "error": "Unauthorized"
}
```

---

### 4. POST /auth/logout
**Description:** Logout current user (invalidate token/session)

**Headers:**
```
Authorization: Bearer <token>
```

**Response (200 OK):**
```json
{
  "message": "Logged out successfully"
}
```

---

## Reservation Endpoints

### 5. GET /reservations
**Description:** Get all reservations (admin sees all, users see only their own)

**Headers:**
```
Authorization: Bearer <token>
```

**Query Parameters:**
- `status` (optional): Filter by status (`pending`, `confirmed`, `cancelled`, `completed`)
- `date` (optional): Filter by date (YYYY-MM-DD)
- `search` (optional): Search by guest name, phone, or email

**Response (200 OK):**
```json
[
  {
    "id": "string",
    "userId": "string",
    "guestName": "string",
    "guestPhone": "string",
    "guestEmail": "string",
    "date": "string (YYYY-MM-DD)",
    "time": "string (HH:mm)",
    "guests": "number",
    "tableNumber": "string",
    "status": "pending" | "confirmed" | "cancelled" | "completed",
    "specialRequests": "string (optional)",
    "createdAt": "string (ISO 8601)"
  }
]
```

---

### 6. GET /reservations/:id
**Description:** Get a specific reservation by ID

**Headers:**
```
Authorization: Bearer <token>
```

**Response (200 OK):**
```json
{
  "id": "string",
  "userId": "string",
  "guestName": "string",
  "guestPhone": "string",
  "guestEmail": "string",
  "date": "string (YYYY-MM-DD)",
  "time": "string (HH:mm)",
  "guests": "number",
  "tableNumber": "string",
  "status": "pending" | "confirmed" | "cancelled" | "completed",
  "specialRequests": "string (optional)",
  "createdAt": "string (ISO 8601)"
}
```

**Error Response (404 Not Found):**
```json
{
  "error": "Reservation not found"
}
```

---

### 7. GET /reservations/user/:userId
**Description:** Get all reservations for a specific user

**Headers:**
```
Authorization: Bearer <token>
```

**Response (200 OK):**
```json
[
  {
    "id": "string",
    "userId": "string",
    "guestName": "string",
    "guestPhone": "string",
    "guestEmail": "string",
    "date": "string (YYYY-MM-DD)",
    "time": "string (HH:mm)",
    "guests": "number",
    "tableNumber": "string",
    "status": "pending" | "confirmed" | "cancelled" | "completed",
    "specialRequests": "string (optional)",
    "createdAt": "string (ISO 8601)"
  }
]
```

---

### 8. POST /reservations
**Description:** Create a new reservation

**Headers:**
```
Authorization: Bearer <token>
Content-Type: application/json
```

**Request Body:**
```json
{
  "guestName": "string",
  "guestPhone": "string",
  "guestEmail": "string",
  "date": "string (YYYY-MM-DD)",
  "time": "string (HH:mm)",
  "guests": "number",
  "tableNumber": "string",
  "specialRequests": "string (optional)"
}
```

**Response (201 Created):**
```json
{
  "id": "string",
  "userId": "string",
  "guestName": "string",
  "guestPhone": "string",
  "guestEmail": "string",
  "date": "string (YYYY-MM-DD)",
  "time": "string (HH:mm)",
  "guests": "number",
  "tableNumber": "string",
  "status": "pending",
  "specialRequests": "string (optional)",
  "createdAt": "string (ISO 8601)"
}
```

**Error Response (400 Bad Request):**
```json
{
  "error": "Validation error",
  "details": {
    "tableNumber": "Table not available at this time",
    "date": "Invalid date format"
  }
}
```

---

### 9. PATCH /reservations/:id
**Description:** Update a reservation

**Headers:**
```
Authorization: Bearer <token>
Content-Type: application/json
```

**Request Body:**
```json
{
  "guestName": "string (optional)",
  "guestPhone": "string (optional)",
  "guestEmail": "string (optional)",
  "date": "string (YYYY-MM-DD, optional)",
  "time": "string (HH:mm, optional)",
  "guests": "number (optional)",
  "tableNumber": "string (optional)",
  "specialRequests": "string (optional)"
}
```

**Response (200 OK):**
```json
{
  "id": "string",
  "userId": "string",
  "guestName": "string",
  "guestPhone": "string",
  "guestEmail": "string",
  "date": "string (YYYY-MM-DD)",
  "time": "string (HH:mm)",
  "guests": "number",
  "tableNumber": "string",
  "status": "pending" | "confirmed" | "cancelled" | "completed",
  "specialRequests": "string (optional)",
  "createdAt": "string (ISO 8601)"
}
```

**Error Response (404 Not Found):**
```json
{
  "error": "Reservation not found"
}
```

---

### 10. PATCH /reservations/:id/status
**Description:** Update reservation status

**Headers:**
```
Authorization: Bearer <token>
Content-Type: application/json
```

**Request Body:**
```json
{
  "status": "pending" | "confirmed" | "cancelled" | "completed"
}
```

**Response (200 OK):**
```json
{
  "id": "string",
  "userId": "string",
  "guestName": "string",
  "guestPhone": "string",
  "guestEmail": "string",
  "date": "string (YYYY-MM-DD)",
  "time": "string (HH:mm)",
  "guests": "number",
  "tableNumber": "string",
  "status": "pending" | "confirmed" | "cancelled" | "completed",
  "specialRequests": "string (optional)",
  "createdAt": "string (ISO 8601)"
}
```

---

### 11. DELETE /reservations/:id
**Description:** Delete a reservation

**Headers:**
```
Authorization: Bearer <token>
```

**Response (200 OK):**
```json
{
  "message": "Reservation deleted successfully"
}
```

**Error Response (404 Not Found):**
```json
{
  "error": "Reservation not found"
}
```

---

## Table Endpoints

### 12. GET /tables
**Description:** Get all tables

**Headers:**
```
Authorization: Bearer <token>
```

**Response (200 OK):**
```json
[
  {
    "id": "string",
    "number": "string",
    "capacity": "number",
    "isAvailable": "boolean",
    "location": "main" | "terrace" | "private"
  }
]
```

---

### 13. GET /tables/:id
**Description:** Get a specific table by ID

**Headers:**
```
Authorization: Bearer <token>
```

**Response (200 OK):**
```json
{
  "id": "string",
  "number": "string",
  "capacity": "number",
  "isAvailable": "boolean",
  "location": "main" | "terrace" | "private"
}
```

**Error Response (404 Not Found):**
```json
{
  "error": "Table not found"
}
```

---

### 14. GET /tables/available
**Description:** Get all available tables

**Headers:**
```
Authorization: Bearer <token>
```

**Query Parameters:**
- `date` (optional): Filter by date (YYYY-MM-DD)
- `time` (optional): Filter by time (HH:mm)
- `guests` (optional): Filter by minimum capacity

**Response (200 OK):**
```json
[
  {
    "id": "string",
    "number": "string",
    "capacity": "number",
    "isAvailable": "boolean",
    "location": "main" | "terrace" | "private"
  }
]
```

---

### 15. PATCH /tables/:id/availability
**Description:** Update table availability

**Headers:**
```
Authorization: Bearer <token>
Content-Type: application/json
```

**Request Body:**
```json
{
  "isAvailable": "boolean"
}
```

**Response (200 OK):**
```json
{
  "id": "string",
  "number": "string",
  "capacity": "number",
  "isAvailable": "boolean",
  "location": "main" | "terrace" | "private"
}
```

**Error Response (404 Not Found):**
```json
{
  "error": "Table not found"
}
```

---

## Reports Endpoints (Admin Only)

### 16. GET /reports/monthly
**Description:** Get list of all months with available statistics

**Headers:**
```
Authorization: Bearer <token>
```

**Response (200 OK):**
```json
[
  {
    "month": "string (YYYY-MM)",
    "totalReservations": "number",
    "completedReservations": "number",
    "cancelledReservations": "number",
    "revenue": "number"
  }
]
```

---

### 17. GET /reports/monthly/:month
**Description:** Get detailed monthly statistics for a specific month

**Headers:**
```
Authorization: Bearer <token>
```

**Path Parameters:**
- `month`: Month in format YYYY-MM (e.g., "2025-10")

**Response (200 OK):**
```json
{
  "month": "string (YYYY-MM)",
  "totalReservations": "number",
  "completedReservations": "number",
  "cancelledReservations": "number",
  "revenue": "number",
  "popularTables": [
    {
      "tableNumber": "string",
      "count": "number"
    }
  ],
  "peakHours": [
    {
      "hour": "string (HH:mm)",
      "count": "number"
    }
  ]
}
```

**Error Response (404 Not Found):**
```json
{
  "error": "Statistics for this month not found"
}
```

---

## User Endpoints

### 18. GET /users/:id
**Description:** Get user profile by ID

**Headers:**
```
Authorization: Bearer <token>
```

**Response (200 OK):**
```json
{
  "id": "string",
  "email": "string",
  "name": "string",
  "phone": "string",
  "role": "admin" | "user",
  "createdAt": "string (ISO 8601)"
}
```

**Error Response (404 Not Found):**
```json
{
  "error": "User not found"
}
```

---

### 19. PATCH /users/:id
**Description:** Update user profile

**Headers:**
```
Authorization: Bearer <token>
Content-Type: application/json
```

**Request Body:**
```json
{
  "name": "string (optional)",
  "phone": "string (optional)",
  "email": "string (optional)"
}
```

**Response (200 OK):**
```json
{
  "id": "string",
  "email": "string",
  "name": "string",
  "phone": "string",
  "role": "admin" | "user",
  "createdAt": "string (ISO 8601)"
}
```

**Error Response (400 Bad Request):**
```json
{
  "error": "Validation error",
  "details": {
    "email": "Email already exists"
  }
}
```

---

## Error Response Format

All endpoints should return errors in the following format:

```json
{
  "error": "string (error message)",
  "details": {
    "field": "error message for specific field"
  }
}
```

## Status Codes

- `200 OK` - Successful request
- `201 Created` - Resource created successfully
- `400 Bad Request` - Validation error or bad request
- `401 Unauthorized` - Authentication required or invalid token
- `403 Forbidden` - Insufficient permissions (e.g., non-admin accessing admin endpoints)
- `404 Not Found` - Resource not found
- `500 Internal Server Error` - Server error

## Authentication

Most endpoints require authentication via Bearer token in the Authorization header:
```
Authorization: Bearer <token>
```

The token should be obtained from the `/auth/login` or `/auth/register` endpoints.

## Authorization

- **Admin users** can access all endpoints and see all reservations
- **Regular users** can only:
  - View their own reservations
  - Create new reservations
  - Update/cancel their own reservations
  - View their own profile
  - Cannot access admin reports endpoints

## Notes

1. All dates should be in ISO 8601 format (YYYY-MM-DD for dates, HH:mm for times)
2. All timestamps should be in ISO 8601 format with timezone (e.g., "2025-11-05T10:30:00Z")
3. Table availability should be checked against existing reservations for the requested date/time
4. Revenue calculations should be based on completed reservations only
5. Popular tables and peak hours should be calculated from completed reservations
6. When creating a reservation, the server should automatically:
   - Set the `userId` to the authenticated user's ID
   - Set the `status` to "pending"
   - Set the `createdAt` timestamp
   - Validate that the table is available at the requested date/time

