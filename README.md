# Chelbo — Cross-Platform Messenger / Кроссплатформенный мессенджер

<p align="center">
  <img src="screenshots/logo.png" alt="Chelbo Logo" width="120">
</p>

<p align="center">
  <strong>Russian alternative to Telegram with Web-first approach</strong><br>
  <strong>Российский аналог Telegram с Web-first подходом</strong>
</p>

<p align="center">
  <a href="#english">English</a> • 
  <a href="#russian">Русский</a>
</p>

---

<a name="english"></a>
# 🇬🇧 English

**Chelbo** is a modern Russian cross-platform messenger with an open architecture. It works directly in the browser and as a PWA without requiring installation. The project was developed as a graduation thesis at TOR Academy.

> **Live Demo:** [https://chelbo.ru](https://chelbo.ru)

---

## ✨ Features

### Core Features

| Feature | Description |
|---------|-------------|
| 💬 **Real-time messaging** | Instant message delivery via WebSocket |
| 👥 **Private & Group chats** | 1-on-1 conversations and group discussions |
| 📎 **File sharing** | Images (up to 10 MB) and documents (up to 50 MB) |
| 🎙️ **Voice messages** | Record and send voice messages directly in chat |
| 📹 **Video messages** | Record and share short videos |
| ↩️ **Message replies** | Quote and reply to any message (Telegram-style) |
| 📤 **Message forwarding** | Forward messages to other chats |
| ✓ **Read receipts** | See when messages are delivered and read |
| ⌨️ **Typing indicators** | See when someone is typing |
| 🔴 **Unread counter** | Badge showing number of unread messages |

### Advanced Features

| Feature | Description |
|---------|-------------|
| 📱 **PWA support** | Install on mobile without App Store |
| 💻 **Responsive design** | Works on desktop, tablet, and mobile |
| 🌙 **Dark theme** | Dark color scheme with red accent |
| 🤖 **AI assistant** | Built-in chatbot to answer questions |
| 🔍 **Message search** | Full-text search through messages |
| 🟢 **Online status** | See when users are online |

---

## 🛠 Technology Stack

### Backend

| Technology | Version | Purpose |
|------------|---------|---------|
| **Go** | 1.22+ | Primary programming language – high performance, lightweight goroutines |
| **Gorilla WebSocket** | 1.5.1 | Real-time bidirectional communication |
| **MySQL** | 8.0+ | Main database – ACID, JSON support, full-text search |
| **Redis** | 7.2+ | Caching, session storage, Pub/Sub for horizontal scaling |
| **JWT** | 5.2+ | Stateless authentication |
| **bcrypt** | - | Secure password hashing |
| **sqlx** | 1.3.5 | Database ORM extension |

### Frontend

| Technology | Version | Purpose |
|------------|---------|---------|
| **Vue 3** | 3.4+ | Progressive frontend framework with Composition API |
| **TypeScript** | 5.3+ | Type safety and better refactoring experience |
| **Pinia** | 2.1+ | Official Vue 3 state management |
| **Vite** | 5.0+ | Build tool with instant HMR |
| **Axios** | 1.6+ | HTTP client with interceptors |
| **Day.js** | 1.11+ | Lightweight date formatting |

### DevOps & Infrastructure

| Technology | Version | Purpose |
|------------|---------|---------|
| **Nginx** | 1.24+ | Reverse proxy, static file serving, WebSocket upgrade |
| **Certbot** | - | SSL/TLS certificate management |
| **Prometheus** | 2.50+ | Metrics collection |
| **Grafana** | 10.0+ | Metrics visualization |
| **GitHub Actions** | - | CI/CD automation |
| **systemd** | - | Process management |

---

## 🏗 Architecture Overview
