# Chelbo — Cross-Platform Messenger / Кроссплатформенный мессенджер

[![Go Version](https://img.shields.io/badge/Go-1.22-blue.svg)](https://go.dev/)
[![Vue Version](https://img.shields.io/badge/Vue-3.4-green.svg)](https://vuejs.org/)
[![License](https://img.shields.io/badge/License-MIT-red.svg)](LICENSE)
[![PRs Welcome](https://img.shields.io/badge/PRs-welcome-brightgreen.svg)](CONTRIBUTING.md)

**[English](#english) | [Русский](#russian)**

---

<a name="english"></a>
# 🇬🇧 English

**Chelbo** is a modern Russian cross-platform messenger with an open architecture. It works directly in the browser and as a PWA without requiring installation. The project was developed as a graduation thesis at TOR Academy.

**Live Demo:** [https://chelbo.ru](https://chelbo.ru)

---

## 📋 Table of Contents

- [Features](#-features)
- [Tech Stack](#-tech-stack)
- [Architecture](#-architecture)
- [Installation](#-installation)
- [Configuration](#-configuration)
- [Running the Project](#-running-the-project)
- [API Documentation](#-api-documentation)
- [Deployment](#-deployment)
- [Roadmap](#-roadmap)
- [License](#-license)
- [Author](#-author)

---

## 🚀 Features

| Feature | Description | Status |
|---------|-------------|--------|
| User Authentication | Registration, login, logout via email | ✅ |
| Private Chats | 1-on-1 real-time messaging | ✅ |
| Group Chats | Create groups with multiple participants | ✅ |
| File Sharing | Send images (10 MB) and documents (50 MB) | ✅ |
| Voice Messages | Record and send voice messages | ✅ |
| Video Messages | Record and send short videos | ✅ |
| Message Replies | Quote and reply to any message | ✅ |
| Message Forwarding | Forward messages to other chats | ✅ |
| Read Receipts | See when messages are delivered and read | ✅ |
| Typing Indicators | See when someone is typing | ✅ |
| Unread Counter | Badge with number of unread messages | ✅ |
| PWA Support | Install on mobile without App Store | ✅ |
| Responsive Design | Works on desktop, tablet, and mobile | ✅ |
| Dark Theme | Dark color scheme with red accent | ✅ |
| AI Assistant | Built-in chatbot for questions | ✅ |

---

## 🛠 Tech Stack

### Backend

| Technology | Version | Purpose |
|------------|---------|---------|
| **Go** | 1.22+ | Main programming language |
| **Gorilla WebSocket** | 1.5.1 | Real-time communication |
| **MySQL** | 8.0+ | Primary database |
| **Redis** | 7.2+ | Caching & Pub/Sub |
| **JWT** | 5.2+ | Authentication |
| **bcrypt** | - | Password hashing |

### Frontend

| Technology | Version | Purpose |
|------------|---------|---------|
| **Vue 3** | 3.4+ | Frontend framework |
| **TypeScript** | 5.3+ | Type safety |
| **Pinia** | 2.1+ | State management |
| **Vite** | 5.0+ | Build tool |
| **Axios** | 1.6+ | HTTP client |

### DevOps

| Technology | Version | Purpose |
|------------|---------|---------|
| **Docker** | 24+ | Containerization |
| **Nginx** | 1.24+ | Reverse proxy |
| **Certbot** | - | SSL certificates |
| **GitHub Actions** | - | CI/CD |

---

## 🏗 Architecture
