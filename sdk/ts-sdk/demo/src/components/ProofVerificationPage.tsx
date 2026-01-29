import { useState } from 'react'
import { ProofVerificationClient } from '@vigilum/sdk'
import './ProofVerificationPage.css'

interface ProofState {
  status: 'idle' | 'generating' | 'submitting' | 'verifying' | 'success' | 'error'
  message: string
  challengeId?: string
  verificationScore?: number
  riskScore?: number
  isVerified?: boolean
  lastVerifiedAt?: string
  proofCount?: number
  verifiedProofCount?: number
}

export default function ProofVerificationPage() {
  const [userId, setUserId] = useState('')
  const [state, setState] = useState<ProofState>({ status: 'idle', message: '' })
  const [client] = useState(() => new ProofVerificationClient({ 
    baseUrl: 'http://localhost:8000/api/v1',
    fetcher: (...args) => fetch(...args),
  }))

  const handleGenerateChallenge = async () => {
    if (!userId) {
      setState({ status: 'error', message: 'Please enter a user ID' })
      return
    }

    setState({ status: 'generating', message: 'Generating challenge...' })
    try {
      const response = await client.generateChallenge(
        userId,
        '0x1234567890123456789012345678901234567890'
      )
      
      setState({
        status: 'idle',
        message: `Challenge generated: ${response.challenge_id}`,
        challengeId: response.challenge_id,
      })
    } catch (error) {
      setState({
        status: 'error',
        message: `Failed to generate challenge: ${error instanceof Error ? error.message : 'Unknown error'}`,
      })
    }
  }

  const handleVerifyProof = async () => {
    if (!state.challengeId) {
      setState({ status: 'error', message: 'No active challenge. Generate one first.' })
      return
    }

    setState({ status: 'submitting', message: 'Submitting proof...' })
    try {
      // Generate mock proof data for demo
      const proofData = new Uint8Array(64).fill(0)
      proofData[0] = Math.floor(Math.random() * 256)
      
      const response = await client.submitProof(
        state.challengeId,
        proofData,
        Math.floor(Math.random() * 100),
        Math.floor(Math.random() * 5000),
        `nonce_${Date.now()}`
      )

      setState({
        status: 'success',
        message: 'Proof verified successfully!',
        verificationScore: response.verification_score,
      })
    } catch (error) {
      setState({
        status: 'error',
        message: `Proof verification failed: ${error instanceof Error ? error.message : 'Unknown error'}`,
      })
    }
  }

  const handleGetScore = async () => {
    if (!userId) {
      setState({ status: 'error', message: 'Please enter a user ID' })
      return
    }

    setState({ status: 'verifying', message: 'Fetching verification score...' })
    try {
      // Call the firewall risk endpoint (which provides user verification data)
      const response = await fetch(
        `http://localhost:8000/api/v1/firewall/risk/${userId}?chain_id=1`,
        {
          method: 'GET',
          headers: { 'Content-Type': 'application/json' },
        }
      ).then(r => {
        if (!r.ok) throw new Error(`HTTP ${r.status}`);
        return r.json();
      });

      setState({
        status: 'success',
        message: 'Score retrieved successfully!',
        verificationScore: response.is_human ? 0.95 : 0.45,
        riskScore: response.risk_score,
        isVerified: response.is_human,
        lastVerifiedAt: response.last_proof_at ? new Date(response.last_proof_at).toLocaleString() : 'Never',
        proofCount: response.proof_count,
        verifiedProofCount: response.is_human ? response.proof_count : 0,
      })
    } catch (error) {
      setState({
        status: 'error',
        message: `Failed to fetch score: ${error instanceof Error ? error.message : 'Unknown error'}`,
      })
    }
  }

  return (
    <div className="container">
      <div className="card">
        <h1>VIGILUM Proof Verification Demo</h1>
        <p className="subtitle">Test the human proof verification system</p>

        <div className="form-group">
          <label htmlFor="userId">User ID:</label>
          <input
            id="userId"
            type="text"
            value={userId}
            onChange={(e) => setUserId(e.target.value)}
            placeholder="Enter user ID (e.g., user123)"
            className="input"
          />
        </div>

        <div className="button-group">
          <button
            onClick={handleGenerateChallenge}
            disabled={state.status !== 'idle'}
            className="button button-primary"
          >
            {state.status === 'generating' ? 'Generating...' : 'Generate Challenge'}
          </button>

          <button
            onClick={handleVerifyProof}
            disabled={!state.challengeId || state.status !== 'idle'}
            className="button button-secondary"
          >
            {state.status === 'submitting' ? 'Submitting...' : 'Verify Proof'}
          </button>

          <button
            onClick={handleGetScore}
            disabled={state.status !== 'idle'}
            className="button button-tertiary"
          >
            {state.status === 'verifying' ? 'Loading...' : 'Get Score'}
          </button>
        </div>

        {state.message && (
          <div className={`message message-${state.status}`}>
            {state.message}
          </div>
        )}

        {state.verificationScore !== undefined && (
          <div className="results">
            <h2>Verification Results</h2>
            <div className="result-item">
              <span>Verification Score:</span>
              <strong>{(state.verificationScore * 100).toFixed(1)}%</strong>
            </div>
            {state.riskScore !== undefined && (
              <div className="result-item">
                <span>Risk Score:</span>
                <strong>{state.riskScore}</strong>
              </div>
            )}
            {state.isVerified !== undefined && (
              <div className="result-item">
                <span>Verified:</span>
                <strong>{state.isVerified ? '✅ Yes' : '❌ No'}</strong>
              </div>
            )}
            {state.proofCount !== undefined && (
              <div className="result-item">
                <span>Total Proofs:</span>
                <strong>{state.proofCount}</strong>
              </div>
            )}
            {state.verifiedProofCount !== undefined && (
              <div className="result-item">
                <span>Verified Proofs:</span>
                <strong>{state.verifiedProofCount}</strong>
              </div>
            )}
            {state.lastVerifiedAt && (
              <div className="result-item">
                <span>Last Verified:</span>
                <strong>{state.lastVerifiedAt}</strong>
              </div>
            )}
          </div>
        )}
      </div>
    </div>
  )
}
