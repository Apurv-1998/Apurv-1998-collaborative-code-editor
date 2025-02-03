# Collaborative Coding Editor

This project implements a collaborative coding editor using a microservices-based architecture. It provides real-time code editing, chat, code compilation via an external API (JDoodle), and robust room and session management. The application is built with a Golang backend and a ReactJS frontend and is designed with production readiness in mind.

---

## Overview

The system comprises several interconnected components:

1. **Authentication Service:**  
   - **Features:** User login, signup, and token management (access and refresh tokens).  
   - **Details:** Uses JWT for stateless authentication and supports role-based access (e.g., admin, user).

2. **Room Management Service:**  
   - **Features:**  
     - Room creation (admin-only).  
     - Invitation generation: Admins can create secure invitation tokens and optionally send them to specific email addresses.  
     - Active invitation retrieval: Students can view all active invitations addressed to their email.  
     - Room history: Both admins and participants can view the rooms they were part of or created.  
     - Room closure: The admin can close a room, which broadcasts a closure event to all connected participants.
   - **Data Flow:** Rooms are created with session metadata, and invitations are stored in the database for later retrieval.

3. **Collaboration Service:**  
   - **Features:**  
     - Real-time code editing and communication via WebSocket connections.  
     - Real-time presence: Admins and participants see who is in the room; join/leave events are broadcast.  
     - Chat functionality: Chat messages are exchanged in real time and persisted to the database; the sender’s username is displayed instead of just the user ID.
   - **Details:** Utilizes a hub/client model to manage WebSocket connections. Tokens are passed via query parameters or headers to authenticate WebSocket connections.

4. **Compiler Service:**  
   - **Features:** Code compilation and execution using the JDoodle API.  
   - **Details:**  
     - Supports multiple languages through a language drop-down (e.g., Python, JavaScript, Go, etc.).  
     - Compiler parameters such as language, version index, and standard input are configurable.  
     - Sensitive credentials for JDoodle (client ID and secret) are managed via environment variables.
     
5. **Session Management Service:**  
   - **Features:**  
     - Auto-save: Code is automatically saved at regular intervals and on events such as tab close or logout.  
     - Session export: The final code and audit trail (including chat messages, join/leave events, etc.) can be exported for later review.  
     - Audit logging: All significant events (e.g., edits, auto-save, join/leave, room closure) are logged and stored.
   - **Details:** Sessions and audit logs are stored in MongoDB for persistence and recovery.

6. **Frontend Application:**  
   - **Built With:** ReactJS (using Create React App) and Monaco Editor for code editing.  
   - **Features:**  
     - User interface for login, registration, dashboard, room creation, and room join.  
     - A real-time code editor with Monaco Editor for syntax highlighting, autocompletion, and an enhanced coding experience.  
     - Language drop-down to select the programming language for code compilation.  
     - Chat UI that displays messages with the sender’s username.  
     - Toast notifications for events such as code auto-save.  
     - Active invitations and room history pages for users to review their participation.
   - **Routing & Authentication:**  
     - Uses React Router v6 for navigation and protected routes.  
     - An AuthContext manages authentication state and token refresh.  
     - Logout functionality clears stored tokens and navigates the user back to the login page.

---

## Communication

- **REST API:**  
  Used for standard operations such as user authentication, room management (creation, invitation, room history, room closure), session management (auto-save, export), and code compilation.

- **WebSocket:**  
  Used for real-time code collaboration and chat. Clients connect to the collaboration endpoint with a valid JWT token (passed via a query parameter or header). The backend authenticates and then broadcasts messages (including real-time edits, chat messages, and join/leave notifications) to all connected clients.

---

## Deployment

- **Development Environment:**  
  - **Docker & Docker Compose:** The project uses Docker and Docker Compose to containerize and orchestrate the backend and frontend services during development.
  
- **Future Deployment:**  
  - **Kubernetes:** The architecture is designed for horizontal scalability, and Kubernetes can be used for production deployment.
  - **CI/CD:** Integration with CI/CD pipelines for automated testing and deployment is recommended.
  
- **Environment Management:**  
  Sensitive information (JWT secrets, JDoodle credentials, MongoDB URI, etc.) is stored in environment variables or a dedicated secrets manager in production.

---

## Data Flow

1. **User Authentication:**  
   - A user registers or logs in via the Authentication Service.
   - The server issues a JWT (with role, user_id, email, username, etc.) used for subsequent requests.
   
2. **Room Management:**  
   - An admin creates a room via the Room Management Service.
   - The admin can generate an invitation token, optionally associated with a specific email.
   - Active invitations are stored and can be retrieved by students.
   - Both admins and participants can view their room history.
   
3. **Real-Time Collaboration:**  
   - Users join a room and establish a WebSocket connection for real-time code editing and chat.
   - Join/leave events are broadcast to all connected clients, and presence is updated accordingly.
   
4. **Code Execution:**  
   - Admins can compile and run code via the Compiler Service.  
   - Code is sent to JDoodle and the output is returned and displayed in the UI.
   
5. **Session Management & Audit Logging:**  
   - Code is auto-saved at regular intervals and on specific events (e.g., logout, tab close, room expiration).
   - All significant events (e.g., edits, chat messages, join/leave, room closure) are logged as audit records.
   - Sessions and audit logs can be exported for later review.

---

## Future Considerations

- **Scalability:**  
  - The microservices architecture supports horizontal scaling of individual services (authentication, room management, collaboration, compiler, etc.).
  
- **Security:**  
  - Enhanced API security via encryption, rate limiting, and secure secrets management.
  - Implementation of robust CORS policies and secure WebSocket communication.
  
- **Monitoring & Analytics:**  
  - Logging and real-time analytics for performance monitoring and user behavior tracking.
  - Audit logging for security and compliance.
  
- **Extensibility:**  
  - Support for additional programming languages, debugging tools, and integrations (e.g., Git repositories, CI/CD pipelines).

---

## How to Run

1. **Backend Setup:**  
   - Configure environment variables (JWT secrets, JDoodle credentials, MongoDB URI, etc.) in a `.env` file.
   - Build and run the Golang backend using Docker Compose or directly on your machine.
   
2. **Frontend Setup:**  
   - Navigate to the `frontend/` directory.
   - Install dependencies with `npm install`.
   - Start the development server with `npm start`.
   
3. **Access the Application:**  
   - The frontend is available at [http://localhost:3000](http://localhost:3000) and communicates with the backend at [http://localhost:8080](http://localhost:8080).

---

## Contact

**Name:** Apurv Sirohi  
**Email:** [apurv.sirohi98@gmail.com](mailto:apurv.sirohi98@gmail.com)

---

This README provides a complete overview of the project’s architecture, data flow, deployment details, and future enhancements, incorporating all major components and improvements.

