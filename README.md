<p align="center">
  <img src="https://sutasuteam.github.io/Hunters/images/logo-hunters.png" width="220" alt="Hunters Logo">
</p>

<h1 align="center">Hunters</h1>

<p align="center">
A decentralized social platform built on the Canopy Network.
</p>

---

# Overview

Hunters is a decentralized social application powered by the **Canopy Network**.

The project demonstrates how a modern web application can integrate directly with Canopy while providing familiar Web2-style user experience.

Core integrations include:

- Canopy Wallet creation
- MetaMask authentication
- Token creation
- NFT minting
- On-chain profile management
- RPC integration
- Custom signer service

---

# Project Structure

```
canopy-main/
│
├── public/
├── routes/
├── utils/
├── plugin/
├── data/
├── server.js
├── package.json
└── .env
```

---

# Requirements

Before running Hunters, make sure you have installed:

- Go
- Node.js
- npm
- Canopy CLI
- Git

---

# Setup

## 1. Clone repository

```bash
git clone https://github.com/sutasuteam/Hunters.git

cd Hunters
```

---

## 2. Install dependencies

```bash
npm install
```

---

## 3. Configure Environment

Create `.env`

Example:

```env
PORT=3001

CANOPY_EXE=/path/to/canopy

CANOPY_PASSWORD=your_password

CANOPY_SIGNER_ADDRESS=xxxxxxxxxxxxxxxx

TEST_ADDRESS=

TEST_PUBLIC_KEY=

TEST_PRIVATE_KEY=
```

---

# Running Hunters

Hunters requires **three running processes**.

---

## Step 1 — Start Canopy

Start the Canopy blockchain normally.

Example

```bash
go run ./cmd/cli start
```

or

```bash
canopy start
```

---

## Step 2 — Enable Go Plugin

Open

```
config/config.json
```

Change

```json
"plugin": ""
```

to

```json
"plugin": "go"
```

Restart the Canopy node after saving.

---

## Step 3 — Start Hunters

Open another terminal

```bash
npm run dev
```

This starts:

- Express Backend
- Frontend
- Signer Service

---

## Step 4 — Open Hunters

Visit

```
http://127.0.0.1:8080/
```

---

# Architecture

```
                 MetaMask
                     │
                     ▼
             Hunters Frontend
                     │
                     ▼
             Express Backend
             (Node.js Server)
                     │
         ┌───────────┴───────────┐
         ▼                       ▼
  Canopy Signer             Local Database
         │                  (users/posts)
         ▼
     Canopy RPC
         │
         ▼
    Canopy Network
```

---

# Main Components

## Frontend

- HTML
- CSS
- JavaScript

---

## Backend

- Express
- Multer
- Local JSON Database

---

## Blockchain

- Canopy Network
- Canopy CLI
- RPC
- Signer Service

---

# Development Notes

Hunters currently stores application data locally using JSON files.

```
data/users.json
data/posts.json
data/minted-nfts.json
```

These can later be replaced by a persistent database.

---

# Technologies

- Node.js
- Express
- JavaScript
- HTML
- CSS
- Canopy Network
- MetaMask
- Ethers.js

---

# License

MIT