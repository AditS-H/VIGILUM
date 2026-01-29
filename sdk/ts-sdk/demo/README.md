# VIGILUM Proof Verification Demo

A Vite + React + TypeScript demo application for testing the VIGILUM proof verification system.

## Setup

1. Install dependencies from the root `ts-sdk` directory:
   ```bash
   cd ..
   npm install
   cd demo
   npm install
   ```

2. Start the demo dev server:
   ```bash
   npm run dev
   ```

3. Build for production:
   ```bash
   npm run build
   ```

## Usage

1. Open http://localhost:5173 in your browser
2. Enter a user ID (e.g., `test-user-1`)
3. Click "Generate Challenge" to create a new proof challenge
4. Click "Verify Proof" to submit a proof
5. Click "Get Score" to view the user's verification score and risk assessment

## Features

- Generate ZK proof challenges
- Submit and verify proofs
- View verification scores and risk metrics
- Real-time proof status tracking
- Responsive UI design

## Backend Integration

The demo communicates with the VIGILUM backend API running on `http://localhost:8080`. 
Make sure the backend is running before testing.

### Expected Backend Endpoints

- `POST /api/v1/proofs/challenges` - Generate challenge
- `POST /api/v1/proofs/verify` - Submit proof
- `GET /api/v1/users/verification-score?user_id=<id>` - Get verification score

## Development

The demo uses:
- **Vite** for fast development and builds
- **React 18** for UI components
- **TypeScript** for type safety
- **CSS Grid/Flexbox** for responsive layouts
- **@vigilum/sdk** for backend communication

## Architecture

```
src/
├── main.tsx                    # Entry point
├── App.tsx                     # Main app component
├── index.css                   # Global styles
├── App.css                     # App styles
└── components/
    └── ProofVerificationPage/  # Main proof verification UI
        ├── ProofVerificationPage.tsx
        └── ProofVerificationPage.css
```

The `ProofVerificationPage` component handles:
- Challenge generation
- Proof submission with mock data
- Score retrieval and display
- Error handling and status messages
